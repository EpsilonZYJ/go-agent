// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package utils

import (
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

func GetTextFromAnthropicMessageParam(msg anthropic.MessageParam) string {
	var contents []string
	for _, content := range msg.Content {
		if content.OfText != nil {
			contents = append(contents, content.OfText.Text)
		}
	}
	return strings.Join(contents, " ")
}
