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
