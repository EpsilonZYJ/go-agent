package toolExecute

import (
	"bufio"
	"fmt"
	"go-agent/common/consts"
	"go-agent/model"
	"go-agent/tool"
	"go-agent/tool/permission"
	"strings"
	"sync"

	"github.com/anthropics/anthropic-sdk-go"
)

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
			var result = model.ToolResult{}
			result.Name = fmt.Sprintf("\033[33m>>> %s\033[0m\n", block.Name)
			output, err := tool.Dispatch(block.Name, block.Input)
			if err != nil {
				(*pRespToolExecutionResults)[idx] = anthropic.NewToolResultBlock(block.ID, err.Error(), true)
				result.Result = fmt.Sprintf("\033[31mError: %s\033[0m\n", err.Error())
			} else {
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

	// 批处理允许列表和询问列表
	batchExecution(toolUseList, allowIndex, &respToolExecutionResults, &printableResults)
	return
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
	scanner *bufio.Scanner,
) []anthropic.ContentBlockParamUnion {
	respToolExecutionResults, printableResults := toolExecutionWithoutAsk(toolUseList, allowIndex, denyIndex, errIndex, denyErrMap, errErrMap)
	var curAskIdx int = 0 // 当前askIndex的下标
	var curIdx int
	if len(askIndex) > 0 {
		curIdx = askIndex[curAskIdx]
	} else {
		curIdx = len(toolUseList)
	}
	for idx, toolUse := range toolUseList {
		if idx == curIdx {
			decision := permission.AskUser(toolUse, scanner, askReasonMap[curIdx])
			if decision == consts.PermissionDeny {
				respToolExecutionResults[curIdx] = anthropic.NewToolResultBlock(toolUseList[curIdx].ID, "Permission denied.", true)
			} else {
				output, err := tool.Dispatch(toolUseList[curIdx].Name, toolUseList[curIdx].Input)
				if err != nil {
					respToolExecutionResults[curIdx] = anthropic.NewToolResultBlock(toolUseList[curIdx].ID, err.Error(), true)
					fmt.Printf("\033[31mError: %s\033[0m\n", err.Error())
				} else {
					respToolExecutionResults[curIdx] = anthropic.NewToolResultBlock(toolUseList[curIdx].ID, output, false)
					lines := strings.Split(output, "\n")
					lines = lines[:min(len(lines), consts.ToolMaxPrintOutputLines)]
					fmt.Printf("\033[90m%s\033[0m\n", strings.Join(lines, "\n"))
				}
			}
			if curAskIdx == len(askIndex) {
				curAskIdx++
				curIdx = len(toolUseList)
			} else {
				curAskIdx++
				curIdx = askIndex[curAskIdx]
			}
		} else if idx < curIdx {
			fmt.Printf("%s%s", printableResults[idx].Name, printableResults[idx].Result)
		}
	}
	return respToolExecutionResults
}
