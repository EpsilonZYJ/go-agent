// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package agent

import (
	"fmt"
	"go-agent/internal/compact"
	"go-agent/internal/utils"
	"strings"
	"sync"
	"time"

	"go-agent/internal/consts"
	"go-agent/internal/hooks"
	"go-agent/internal/llm"
	"go-agent/internal/logs"
	"go-agent/internal/tool/execute"

	"github.com/anthropics/anthropic-sdk-go"
)

var (
	roundSinceTodo int = 0
	syncRST        sync.RWMutex
)

func RoundSinceTodoSetZero() {
	syncRST.Lock()
	defer syncRST.Unlock()
	roundSinceTodo = 0
}

func GetRoundSinceTodo() int {
	syncRST.RLock()
	defer syncRST.RUnlock()
	return roundSinceTodo
}

func IncreaseRoundSinceTodoByOne() {
	syncRST.Lock()
	defer syncRST.Unlock()
	roundSinceTodo++
}

func compactToolUseID(toolUses []anthropic.ContentBlockUnion) (string, bool) {
	for _, tu := range toolUses {
		if tu.Name == consts.ToolContextCompact {
			return tu.ID, true
		}
	}
	return "", false
}

func AgentLoop(request *llm.ChatRequest) {
	var reactiveTrials int = 0 // 压缩重试次数（网络重试已由 llm.Call 内部处理）

	startTime := time.Now()
	for loop := 0; ; loop++ {

		request.Messages = compact.ToolResultBudget(request.Messages) // L3: persist large results
		request.Messages = compact.SnipCompact(request.Messages)      // L1: trim middle
		request.Messages = compact.MicroCompact(request.Messages)     // L2: old result placeholders

		if compact.EstimateMessagesSize(request) > consts.ContextLimit {
			fmt.Printf("[auto compact]\n")
			request.Messages = compact.CompactHistory(request.Messages)
		}

		// 添加todo reminder
		if GetRoundSinceTodo() >= consts.TodoReminderRounds && len(request.Messages) != 0 {
			request.AddUserContent("<reminder>Update your todos.</reminder>")
			RoundSinceTodoSetZero()
			logs.Debug("todo reminder injected")
		}

		// 创建请求：瞬时错误由 llm.Call 内部退避重试
		resp, rerr := llm.Call(
			anthropic.MessageNewParams{
				MaxTokens: request.MaxTokens,
				Messages:  request.Messages,
				Model:     request.Model,
				System:    request.SystemPrompt,
				Tools:     request.Tools,
			},
			consts.MaxRequestTries,
		)
		if rerr != nil {
			if rerr.Kind == llm.PromptTooLongErr {
				if reactiveTrials < consts.MaxReactiveTrials {
					fmt.Printf("[reactive compact]\n")
					logs.Debug("prompt too long")
					request.Messages = compact.ReactiveCompact(request.Messages)
					reactiveTrials++
					continue
				}
				logs.Debug("reactive compact hit max limit")
				return
			}
			logs.Error("request failed",
				"loop", loop,
				"kind", rerr.Kind,
				"err", rerr.Err,
			)
			fmt.Printf("An error occurred: %v\n", rerr.Err)
			return
		}

		reactiveTrials = 0 // 仅在成功后重置
		request.Messages = append(request.Messages, resp.ToParam())

		logs.Debug("response received",
			"loop", loop,
			"stopReason", resp.StopReason,
			"contentBlocks", len(resp.Content),
		)

		// 收集输出和工具调用
		var toolUses []anthropic.ContentBlockUnion
		var textOuts []strings.Builder
		textOuts, toolUses, allowIndex, denyIndex, askIndex, errIndex, denyErrMap, errErrMap, askReasonMap := execute.CollectLLMOutput(resp.Content)
		IncreaseRoundSinceTodoByOne()

		utils.PrintAgentOutput(textOuts)
		// 无工具调用，本轮结束
		if resp.StopReason != anthropic.StopReasonToolUse || len(toolUses) == 0 {
			logs.Info("agent turn finished",
				"loops", loop+1,
				"duration", time.Since(startTime),
			)
			hooks.TriggerStop(request.Messages)
			return
		}

		if compactID, ok := compactToolUseID(toolUses); ok {
			fmt.Printf("[compact]\n")
			logs.Info("compact tool invoked, compacting history")
			last := request.Messages[len(request.Messages)-1]
			head := compact.CompactHistory(request.Messages[:len(request.Messages)-1])
			request.Messages = append(head, last)
			request.Messages = append(request.Messages, anthropic.NewUserMessage(
				anthropic.NewToolResultBlock(compactID, "[Compacted. Conversation history has been summarized.]", false),
			))
			continue
		}

		logs.Info("executing tools",
			"loop", loop,
			"toolCount", len(toolUses),
			"allowCount", len(allowIndex),
			"denyCount", len(denyIndex),
			"askCount", len(askIndex),
		)
		results := execute.ToolExecution(toolUses, allowIndex, denyIndex, askIndex, errIndex, denyErrMap, errErrMap, askReasonMap)
		// 本轮若调用了 todo_write（无论 allow/ask 路径），重置提醒计数
		for _, tu := range toolUses {
			if tu.Name == consts.ToolTodoWrite {
				RoundSinceTodoSetZero()
				logs.Debug("todo_write detected, resetting round counter")
				break
			}
		}
		request.Messages = append(request.Messages, anthropic.NewUserMessage(results...))
	}
}
