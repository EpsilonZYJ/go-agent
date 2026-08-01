// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package builtin

import (
	"go-agent/internal/consts"
	"go-agent/internal/llm"
	"go-agent/internal/tool"
)

type compactInput struct {
	// Focus is an optional hint about what the summary should emphasize
	// (e.g. the current task, decisions made, pending work). Leave empty to
	// summarize the entire conversation.
	Focus string `json:"focus" jsonschema_description:"Optional hint for what the summary should focus on, such as the current task, decisions, or pending work. Leave empty for a general summary."`
}

func registerToolCompact(req *llm.ChatRequest) error {
	return tool.RegisterTool(
		req, consts.ToolContextCompact,
		"Summarize earlier conversation to free context space. Any other tool execution will be discarded in this round.",
		func(in compactInput) (string, error) {
			return "[Compacted. Conversation history has been summarized.]", nil
		},
	)
}
