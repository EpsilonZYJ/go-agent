// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package model

import "github.com/anthropics/anthropic-sdk-go"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content,omitempty"`
}

func (msg Message) ToAnthropicMessage() anthropic.MessageParam {
	switch msg.Role {
	case "user":
		return anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content))
	case "assistant":
		return anthropic.NewAssistantMessage(anthropic.NewTextBlock(msg.Content))
	default:
		return anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content))
	}
}
