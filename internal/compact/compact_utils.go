// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package compact

import (
	"go-agent/internal/llm"

	"github.com/anthropics/anthropic-sdk-go"
)

func estimateMessagesSize(req *llm.ChatRequest) int {
	return len(req.Messages)
}

func blockType() {

}

func messageHasToolUse(msg anthropic.MessageParam) bool {
	if msg.Role != anthropic.MessageParamRoleAssistant {
		return false
	}
	content := msg.Content
	for _, block := range content {
		if block.OfToolUse != nil {
			return true
		}
	}
	return false
}

func isToolResultMessage(msg anthropic.MessageParam) bool {
	if msg.Role != anthropic.MessageParamRoleUser {
		return false
	}
	content := msg.Content
	for _, block := range content {
		if block.OfToolResult != nil {
			return true
		}
	}
	return false
}
