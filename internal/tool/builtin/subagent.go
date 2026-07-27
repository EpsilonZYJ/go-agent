// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package builtin

import (
	"context"
	"errors"
	"fmt"
	"go-agent/internal/config"
	"go-agent/internal/consts"
	"go-agent/internal/errs"
	"go-agent/internal/llm"
	"go-agent/internal/logs"
	"go-agent/internal/tool"
	"go-agent/internal/tool/execute"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
)

type subagentInput struct {
	Description string `json:"description" jsonschema:"required" jsonschema_description:"description"`
}

func extractText(content []anthropic.ContentBlockParamUnion) string {
	var parts []string
	for _, b := range content {
		if b.OfText != nil && b.OfText.Text != "" {
			parts = append(parts, b.OfText.Text)
		}
	}
	return strings.Join(parts, "\n")
}

func RunSubagent(description string) (string, error) {
	fmt.Printf("\n\033[35m[Subagent spawned]\033[0m\n")
	req := llm.NewChatRequest(config.ModelCfg.Model, config.ModelCfg.MaxTokens, config.SysCfg.SubSystemPrompt)
	err := SubAgentRegisterBuiltinTools(req)
	if err != nil {
		return "", err
	}
	req.AddUserContent(description)
	var trials int = 0

	for _ = range consts.SubAgentSafetyLimit {
		ctx, cancel := context.WithTimeout(context.Background(), consts.RequestTimeout)
		resp, err := llm.Client.Messages.New(
			ctx,
			anthropic.MessageNewParams{
				MaxTokens: req.MaxTokens,
				Messages:  req.Messages,
				Model:     req.Model,
				System:    req.SystemPrompt,
				Tools:     req.Tools,
			},
		)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
				logs.Debug("[SubAgentLoop] Request timeout.")
			}
			errCode := errs.AnthropicRequestErrorCode(err)
			if errCode >= http.StatusBadRequest && errCode < http.StatusInternalServerError && errCode != http.StatusTooManyRequests {
				cancel()
				return "", fmt.Errorf("An error occurred: %v\n", err)
			} else if trials >= consts.MaxRequestTries {
				cancel()
				return "", fmt.Errorf("Max Request Tries: %d\n", trials)
			}
			trials++
			time.Sleep(time.Duration(trials) * consts.RetryDelay)
			cancel()
			fmt.Printf("Error: %v\n", err)
			continue
		}
		cancel()

		trials = 0
		req.Messages = append(req.Messages, resp.ToParam())

		var toolUses []anthropic.ContentBlockUnion
		_, toolUses, allowIndex, denyIndex, askIndex, errIndex, denyErrMap, errErrMap, askReasonMap := execute.CollectLLMOutput(resp.Content)

		// 无工具调用，本轮结束
		if resp.StopReason != anthropic.StopReasonToolUse || len(toolUses) == 0 {
			break
		}

		results := execute.ToolExecution(toolUses, allowIndex, denyIndex, askIndex, errIndex, denyErrMap, errErrMap, askReasonMap)
		req.Messages = append(req.Messages, anthropic.NewUserMessage(results...))

	}

	var result string
	if len(req.Messages) > 0 {
		result = extractText(req.Messages[len(req.Messages)-1].Content)
	}
	if result == "" {
		slices.Reverse(req.Messages)
		for _, msg := range req.Messages {
			if msg.Role == anthropic.MessageParamRoleAssistant {
				result = extractText(msg.Content)
				if result != "" {
					break
				}
			}
		}
		if result == "" {
			result = "Subagent stopped after 30 turns without final answer."
		}
	}
	fmt.Printf("\033[35m[Subagent done]\033[0m\n")
	return result, nil
}

func registerToolSubagent(req *llm.ChatRequest) error {
	return tool.RegisterTool(req, consts.ToolSubagent, "Launch a subagent to handle a complex subtask. Returns only the final conclusion.", func(in subagentInput) (string, error) {
		return RunSubagent(in.Description)
	})
}
