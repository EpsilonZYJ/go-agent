// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package execute

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"go-agent/internal/consts"
	"go-agent/internal/hooks"
	"go-agent/internal/logs"
	"go-agent/internal/model"
	"go-agent/internal/tool"
	"go-agent/internal/tool/permission"

	"github.com/anthropics/anthropic-sdk-go"
)

var printMu sync.Mutex

// serialTools 在执行过程中会直接向终端打印（子代理的流式输出、todo 列表）。
// 若放进并发批处理，它们的打印会与批内其他输出交错、并与自身 ">>> name" 头部
// 分离（头部在批处理结束后才打印）。因此这类工具必须在主 goroutine 串行执行，
// 先打印头部再执行，保证工具内输出连续。
var serialTools = map[string]bool{
	consts.ToolSubagent:  true,
	consts.ToolTodoWrite: true,
}

func batchExecution(
	toolUseList []anthropic.ContentBlockUnion,
	allowIndex []int,
	pRespToolExecutionResults *[]anthropic.ContentBlockParamUnion,
	pPrintableResults *[]model.ToolResult,
) {
	// 批处理allow
	var toolwg sync.WaitGroup

	// 并发执行自动批准的工具
	for _, allowidx := range allowIndex {
		toolwg.Add(1)
		go func(idx int, block anthropic.ContentBlockUnion) {
			defer toolwg.Done()
			result := model.ToolResult{}
			result.Name = fmt.Sprintf("\033[33m>>> %s\033[0m\n", block.Name)
			output, err := tool.Dispatch(block.Name, block.Input)
			if err != nil {
				logs.Warn("tool dispatch failed", "tool", block.Name, "err", err)
				(*pRespToolExecutionResults)[idx] = anthropic.NewToolResultBlock(block.ID, err.Error(), true)
				result.Result = fmt.Sprintf("\033[31mError: %s\033[0m\n", err.Error())
			} else {
				hooks.TriggerPostToolUse(block, output)
				(*pRespToolExecutionResults)[idx] = anthropic.NewToolResultBlock(block.ID, output, false)
				lines := strings.Split(output, "\n")
				lines = lines[:min(len(lines), consts.ToolMaxPrintOutputLines)]
				result.Result = fmt.Sprintf("\033[90m%s\033[0m\n", strings.Join(lines, "\n"))
			}
			(*pPrintableResults)[idx] = result
		}(allowidx, toolUseList[allowidx])
	}
	toolwg.Wait()
}

func toolExecutionWithoutAsk(
	toolUseList []anthropic.ContentBlockUnion,
	allowIndex []int,
	denyIndex []int,
	errIndex []int,
	denyErrMap map[int]string,
	errErrMap map[int]string,
) (
	respToolExecutionResults []anthropic.ContentBlockParamUnion,
	printableResults []model.ToolResult,
) {
	respToolExecutionResults = make([]anthropic.ContentBlockParamUnion, len(toolUseList))
	printableResults = make([]model.ToolResult, len(toolUseList))

	// 处理拒绝列表
	for _, denyidx := range denyIndex {
		respToolExecutionResults[denyidx] = anthropic.NewToolResultBlock(toolUseList[denyidx].ID, "Permission denied.", true)
		printableResults[denyidx] = model.ToolResult{
			Name:   fmt.Sprintf("\033[33m>>> %s\033[0m\n", toolUseList[denyidx].Name),
			Result: denyErrMap[denyidx],
		}
	}

	// 处理出错列表
	for _, erridx := range errIndex {
		respToolExecutionResults[erridx] = anthropic.NewToolResultBlock(toolUseList[erridx].ID, errErrMap[erridx], true)
		printableResults[erridx] = model.ToolResult{
			Name:   fmt.Sprintf("\033[33m>>> %s\033[0m\n", toolUseList[erridx].Name),
			Result: fmt.Sprintf("\033[31mError: %s\033[0m\n", errErrMap[erridx]),
		}
	}

	// 批处理允许列表
	batchExecution(toolUseList, allowIndex, &respToolExecutionResults, &printableResults)
	return
}

// runTool 执行单个工具并立即打印其结果，返回写入消息历史的 result block。
// 调用前需已打印 ">>> name" 头部，保证头部与输出连续。
func runTool(block anthropic.ContentBlockUnion) anthropic.ContentBlockParamUnion {
	output, err := tool.Dispatch(block.Name, block.Input)
	if err != nil {
		logs.Warn("tool dispatch failed", "tool", block.Name, "err", err)
		fmt.Printf("\033[31mError: %s\033[0m\n", err.Error())
		return anthropic.NewToolResultBlock(block.ID, err.Error(), true)
	}
	hooks.TriggerPostToolUse(block, output)
	lines := strings.Split(output, "\n")
	lines = lines[:min(len(lines), consts.ToolMaxPrintOutputLines)]
	fmt.Printf("\033[90m%s\033[0m\n", strings.Join(lines, "\n"))
	return anthropic.NewToolResultBlock(block.ID, output, false)
}

func ToolExecution(
	toolUseList []anthropic.ContentBlockUnion,
	allowIndex []int,
	denyIndex []int,
	askIndex []int,
	errIndex []int,
	denyErrMap map[int]string,
	errErrMap map[int]string,
	askReasonMap map[int]string,
) []anthropic.ContentBlockParamUnion {
	// 拆分 allow：静默工具并发批处理；打印类工具（task/todo_write）留到主 goroutine 串行执行。
	var batchAllow, serialIdx []int
	for _, idx := range allowIndex {
		if serialTools[toolUseList[idx].Name] {
			serialIdx = append(serialIdx, idx)
		} else {
			batchAllow = append(batchAllow, idx)
		}
	}

	respToolExecutionResults, printableResults := toolExecutionWithoutAsk(toolUseList, batchAllow, denyIndex, errIndex, denyErrMap, errErrMap)

	// 交互/串行列表 = ask + 打印类工具，按下标升序，在主 goroutine 依次处理。
	type interactive struct {
		idx   int
		isAsk bool
	}
	inter := make([]interactive, 0, len(askIndex)+len(serialIdx))
	for _, idx := range askIndex {
		inter = append(inter, interactive{idx, true})
	}
	for _, idx := range serialIdx {
		inter = append(inter, interactive{idx, false})
	}
	sort.Slice(inter, func(a, b int) bool { return inter[a].idx < inter[b].idx })

	cur := 0
	next := len(toolUseList)
	if len(inter) > 0 {
		next = inter[0].idx
	}
	for idx, toolUse := range toolUseList {
		if idx == next {
			if inter[cur].isAsk {
				// ask 路径：bash/write/edit 不会重入 printMu，持锁保证提示+读输入原子。
				printMu.Lock()
				decision := permission.AskUser(toolUse, askReasonMap[idx])
				if decision == consts.PermissionDeny {
					logs.Info("user denied tool execution", "tool", toolUse.Name)
					respToolExecutionResults[idx] = anthropic.NewToolResultBlock(toolUse.ID, "Permission denied.", true)
				} else {
					logs.Info("user allowed tool execution", "tool", toolUse.Name)
					respToolExecutionResults[idx] = runTool(toolUse)
				}
				printMu.Unlock()
			} else {
				// 打印类串行工具：先头部，再执行（其内部打印紧随其后），最后结果。
				// 此处不能持有 printMu：子代理内部可能再次进入 ask 分支请求同一把锁，
				// sync.Mutex 不可重入会死锁。批处理已结束，主 goroutine 是唯一打印者。
				fmt.Printf("\033[33m>>> %s\033[0m\n", toolUse.Name)
				respToolExecutionResults[idx] = runTool(toolUse)
			}
			cur++
			if cur < len(inter) {
				next = inter[cur].idx
			} else {
				next = len(toolUseList)
			}
		} else if idx < next {
			fmt.Printf("%s%s", printableResults[idx].Name, printableResults[idx].Result)
		}
	}
	return respToolExecutionResults
}
