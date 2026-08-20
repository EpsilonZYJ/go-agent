// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package compact

import (
	"encoding/json"
	"go-agent/internal/llm"
	"strings"
	"unicode/utf8"

	"github.com/anthropics/anthropic-sdk-go"
)

func EstimateMessagesSize(req *llm.ChatRequest) int {
	b, err := json.Marshal(req.Messages)
	if err != nil {
		return 0
	}
	return len([]rune(string(b)))
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

func isSyntheticUserText(s string) bool {
	s = strings.TrimSpace(s)
	switch {
	case strings.HasPrefix(s, "<reminder>"),
		strings.HasPrefix(s, "[snipped"),
		strings.HasPrefix(s, "[Compacted]"),
		strings.HasPrefix(s, "[Reactive compact]"):
		return true
	}
	return false
}

func LastUserInstruction(msgs []anthropic.MessageParam) (int, string) {
	for i := len(msgs) - 1; i >= 0; i-- {
		m := msgs[i]
		if m.Role != anthropic.MessageParamRoleUser || isToolResultMessage(m) {
			continue
		}
		for _, b := range m.Content {
			if b.OfText == nil {
				continue
			}
			t := strings.TrimSpace(b.OfText.Text)
			if t == "" || isSyntheticUserText(t) {
				continue
			}
			return i, t
		}
	}
	return -1, ""
}
