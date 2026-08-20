// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package memory

import (
	"encoding/json"
	"fmt"
	"go-agent/internal/config"
	"go-agent/internal/llm"
	"go-agent/internal/logs"
	"go-agent/internal/utils"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

func ExtractMemories(message []anthropic.MessageParam) {
	var dialogueParts []string
	for _, msg := range message[max(0, len(message)-10):] {
		role := msg.Role
		content := utils.GetTextFromAnthropicMessageParam(msg)
		content = strings.TrimSpace(content)
		dialogueParts = append(dialogueParts, fmt.Sprintf("%s: %s", role, content))
	}
	dialogue := strings.Join(dialogueParts, "\n")
	dialogue = strings.TrimSpace(dialogue)

	existing := listMemoryFiles()
	var existingDesc string
	{
		if len(existing) <= 0 {
			existingDesc = "(none)"
		} else {
			var existingDescs []string
			for _, m := range existing {
				existingDescs = append(existingDescs, fmt.Sprintf("- %s: %s", m.Name, m.Description))
			}
			existingDesc = strings.Join(existingDescs, "\n")
		}
	}

	prompt := fmt.Sprintf(
		"Extract user preferences, constraints, or project facts from this dialogue.\n"+
			"Return a JSON array. Each item: {name, type, description, body}.\n"+
			"- name: short kebab-case identifier (e.g. 'user-preference-tabs')\n"+
			"- type: one of '%s' (user preference), '%s' (guidance), "+
			"'%s' (project fact), '%s' (external pointer)\n"+
			"- description: one-line summary for index lookup\n"+
			"- body: full detail in markdown\n"+
			"If nothing new or already covered by existing memories, return [].\n\n"+
			"Existing memories:\n%s\n\n"+
			"Dialogue:\n%s",
		MemTypeUser, MemTypeFeedback, MemTypeProject, MemTypeReference, existingDesc, utils.StringTruncateRunes(dialogue, 4000),
	)

	resp, err := llm.Call(
		anthropic.MessageNewParams{
			Messages: []anthropic.MessageParam{
				anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
			},
			MaxTokens: 800,
			Model:     config.Cfg.Model.Model,
		},
		0,
	)
	if err != nil {
		logs.Warn("ExtractMemories: llm call error", "error", err)
		return
	}
	text := utils.GetTextFromAnthropicMessageParam(resp.ToParam())
	text = strings.TrimSpace(text)
	logs.Info("ExtractMemories: llm response - ", text)
	match := memExtractionRe.FindString(text)
	if match == "" {
		logs.Warn("ExtractMemories: no valid JSON array found in llm response")
		return
	}
	var items []extractedMemory
	if err := json.Unmarshal([]byte(match), &items); err != nil {
		logs.Warn("ExtractMemories: failed to unmarshal JSON", "error", err)
		return
	}
	var count = 0
	for _, mem := range items {
		if mem.Description == "" || mem.Body == "" {
			logs.Warn("ExtractMemories: skipping memory with empty description or body", "memory", mem)
			continue
		}
		if mem.Name == "" {
			mem.Name = fmt.Sprintf("memory_%s", utils.NowTime())
		}
		if mem.Type == "" {
			mem.Type = MemTypeUser
		}
		if _, err := writeMemoryFile(mem.Name, mem.Type, mem.Description, mem.Body); err == nil {
			count++
		}
	}
	if count > 0 {
		fmt.Printf("\n\033[33m[Memory: extracted %d new memories]\033[0m\n", count)
	}
}
