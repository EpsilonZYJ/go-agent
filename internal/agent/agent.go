// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package agent

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"go-agent/internal/consts"
	"go-agent/internal/errs"
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

func AgentLoop(request *llm.ChatRequest, scanner *bufio.Scanner) {
	var trials int = 0
	for {
		// 添加todo reminder
		if GetRoundSinceTodo() >= consts.TodoReminderRounds && len(request.Messages) != 0 {
			request.AddUserContent("<reminder>Update your todos.</reminder>")
			roundSinceTodo = 0
		}

		// 创建请求
		ctx, cancel := context.WithTimeout(context.Background(), consts.RequestTimeout)
		resp, err := llm.Client.Messages.New(
			ctx,
			anthropic.MessageNewParams{
				MaxTokens: request.MaxTokens,
				Messages:  request.Messages,
				Model:     request.Model,
				System:    request.SystemPrompt,
				Tools:     request.Tools,
			},
		)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
				logs.Debug("[AgentLoop] Request timeout.")
			}
			errCode := errs.AnthropicRequestErrorCode(err)
			if errCode >= http.StatusBadRequest && errCode < http.StatusInternalServerError && errCode != http.StatusTooManyRequests {
				fmt.Printf("An error occurred: %v\n", err)
				cancel()
				return
			} else if trials >= consts.MaxRequestTries {
				fmt.Printf("Max Request Tries: %d\n", trials)
				cancel()
				return
			}
			trials++
			time.Sleep(time.Duration(trials) * consts.RetryDelay)
			cancel()
			fmt.Printf("Error: %v\n", err)
			continue
		}
		cancel()

		trials = 0
		request.Messages = append(request.Messages, resp.ToParam())

		// 收集输出和工具调用
		var toolUses []anthropic.ContentBlockUnion
		var textOuts []strings.Builder
		textOuts, toolUses, allowIndex, denyIndex, askIndex, errIndex, denyErrMap, errErrMap, askReasonMap := execute.CollectLLMOutput(resp.Content)
		IncreaseRoundSinceTodoByOne()

		PrintAgentOutput(textOuts)
		// 无工具调用，本轮结束
		if resp.StopReason != anthropic.StopReasonToolUse || len(toolUses) == 0 {
			return
		}

		results := execute.ToolExecution(toolUses, allowIndex, denyIndex, askIndex, errIndex, denyErrMap, errErrMap, askReasonMap, scanner)
		request.Messages = append(request.Messages, anthropic.NewUserMessage(results...))
	}
}

func PrintAgentOutput(textOuts []strings.Builder) {
	for _, textOut := range textOuts {
		if textOut.Len() > 0 {
			fmt.Println("\033[32mAgent: \n\n \033[0m" + textOut.String())
		}
	}
	fmt.Println()
}
