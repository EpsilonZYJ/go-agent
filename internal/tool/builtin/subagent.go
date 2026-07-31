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
	logs.Info("subagent spawned", "description", description)
	fmt.Printf("\n\033[35m[Subagent spawned]\033[0m\n")
	req := llm.NewChatRequest(config.ModelCfg.Model, config.ModelCfg.MaxTokens, config.SysCfg.SubSystemPrompt)
	err := SubAgentRegisterBuiltinTools(req)
	if err != nil {
		logs.Warn("subagent tool registration failed", "err", err)
		return "", err
	}
	req.AddUserContent(description)
	var trials int = 0

	for loop := 0; loop < consts.SubAgentSafetyLimit; loop++ {
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
				logs.Debug("subagent request timeout",
					"loop", loop,
					"trial", trials,
				)
			}
			errCode := errs.AnthropicRequestErrorCode(err)
			if errCode >= http.StatusBadRequest && errCode < http.StatusInternalServerError && errCode != http.StatusTooManyRequests {
				logs.Warn("subagent non-retryable API error",
					"loop", loop,
					"errCode", errCode,
					"err", err,
				)
				cancel()
				return "", fmt.Errorf("An error occurred: %v\n", err)
			} else if trials >= consts.MaxRequestTries {
				logs.Warn("subagent max request tries exceeded",
					"loop", loop,
					"trials", trials,
					"err", err,
				)
				cancel()
				return "", fmt.Errorf("Max Request Tries: %d\n", trials)
			}
			trials++
			logs.Warn("subagent retrying request",
				"loop", loop,
				"trial", trials,
				"err", err,
			)
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
			logs.Debug("subagent turn finished", "loop", loop)
			break
		}

		results := execute.ToolExecution(toolUses, allowIndex, denyIndex, askIndex, errIndex, denyErrMap, errErrMap, askReasonMap)
		req.Messages = append(req.Messages, anthropic.NewUserMessage(results...))

	}

	// 如果走了完整 30 轮还没结束，标记警告
	if len(req.Messages) > 0 {
		lastMsg := req.Messages[len(req.Messages)-1]
		if lastMsg.Role == anthropic.MessageParamRoleUser {
			logs.Warn("subagent reached safety limit", "limit", consts.SubAgentSafetyLimit)
		}
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
			logs.Warn("subagent produced no text output")
			result = "Subagent stopped after 30 turns without final answer."
		}
	}
	logs.Info("subagent finished", "resultLen", len(result))
	fmt.Printf("\033[35m[Subagent done]\033[0m\n")
	return result, nil
}

func registerToolSubagent(req *llm.ChatRequest) error {
	return tool.RegisterTool(req, consts.ToolSubagent, "Launch a subagent to handle a complex subtask. Returns only the final conclusion.", func(in subagentInput) (string, error) {
		return RunSubagent(in.Description)
	})
}
