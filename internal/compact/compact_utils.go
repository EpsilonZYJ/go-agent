// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package compact

import (
	"go-agent/internal/llm"
	"strings"
	"unicode/utf8"

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

func toolResultTextBytes(tr *anthropic.ToolResultBlockParam) int {
	n := 0
	for _, c := range tr.Content {
		if c.OfText != nil {
			n += len(c.OfText.Text)
		}
	}
	return n
}

func toolResultTextLen(tr *anthropic.ToolResultBlockParam) int {
	n := 0
	for _, c := range tr.Content {
		if c.OfText != nil {
			n += utf8.RuneCountInString(c.OfText.Text)
		}
	}
	return n
}

func toolResultText(tr *anthropic.ToolResultBlockParam) string {
	var b strings.Builder
	for _, c := range tr.Content {
		if c.OfText != nil {
			b.WriteString(c.OfText.Text)
		}
	}
	return b.String()
}
