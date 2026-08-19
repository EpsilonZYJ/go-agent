package memory

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"go-agent/internal/config"
	"go-agent/internal/consts"
	"go-agent/internal/llm"
	"go-agent/internal/utils"

	"github.com/anthropics/anthropic-sdk-go"
)

func strContainsAny(s string, keywords []string) bool {
	for _, kw := range keywords {
		if strings.Contains(s, kw) {
			return true
		}
	}
	return false
}

func selectRelevantMemories(messages []anthropic.MessageParam, maxItems int) []string {
	files := listMemoryFiles()
	if len(files) <= 0 {
		return []string{}
	}

	// collect recent user text for context
	var recentContext []string
	slices.Reverse(messages)
	for _, msg := range messages {
		if msg.Role == anthropic.MessageParamRoleUser {
			toAppend := utils.GetTextFromAnthropicMessageParam(msg)
			if toAppend != "" {
				recentContext = append(recentContext, toAppend)
			} else {
				continue
			}
			if len(recentContext) >= consts.MaxRecentMessagesForRelevantSelect {
				break
			}
		}
	}
	slices.Reverse(recentContext)
	recent := strings.Join(recentContext, " ")
	recent = strings.TrimSpace(recent)
	recent = utils.StringTruncateRunes(recent, 2000)
	if len(recent) <= 0 {
		return []string{}
	}
	var catalogLines []string
	for idx, f := range files {
		catalogLines = append(catalogLines, fmt.Sprintf("%d: %s — %s", idx, f.Name, f.Description))
	}
	catalog := strings.Join(catalogLines, " ")
	prompt := fmt.Sprintf(
		"Given the recent conversation and the memory catalog below, "+
			"select the indices of memories that are clearly relevant. "+
			"Return ONLY a JSON array of integers, e.g. [0, 3]. "+
			"If none are relevant, return [].\n\n"+
			"Recent conversation:\n%s\n\n"+
			"Memory catalog:\n%s",
		recent, catalog,
	)
	resp, err := llm.Call(
		anthropic.MessageNewParams{
			Messages: []anthropic.MessageParam{
				anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
			},
			MaxTokens: 200,
			Model:     config.Cfg.Model.Model,
		},
		0,
	)
	if err == nil {
		var text string
		text = utils.GetTextFromAnthropicMessageParam(resp.ToParam())
		text = strings.TrimSpace(text)
		match := memSelectionRe.FindString(text)
		if match != "" {
			var indices []any
			if err := json.Unmarshal([]byte(match), &indices); err != nil {
				return []string{}
			}
			selected := make([]string, 0, maxItems)
			for _, idx := range indices {
				f, ok := idx.(float64)
				if !ok {
					continue
				}
				i := int(f)
				if f != float64(i) {
					continue
				}
				if i >= 0 && i < len(files) {
					selected = append(selected, files[i].Filename)
					if len(selected) >= maxItems {
						break
					}
				}
			}
			return selected
		}
	}

	// fallback, keyword matching on name + description
	var keywords []string
	for _, rc := range strings.Fields(strings.ToLower(recent)) {
		if utf8.RuneCountInString(rc) > 3 {
			keywords = append(keywords, rc)
		}
	}
	var selected []string
	for _, f := range files {
		text := strings.ToLower(f.Name + " " + f.Description)
		if strContainsAny(text, selected) {
			selected = append(selected, f.Filename)
			if len(selected) >= maxItems {
				break
			}
		}
	}
	return selected
}

func loadMemories(messages []anthropic.MessageParam) string {
	selectedFiles := selectRelevantMemories(messages, consts.MaxRelevantMemoriesToSelect)
	if len(selectedFiles) <= 0 {
		return ""
	}
	parts := []string{"<relevant_memories>"}
	for _, filename := range selectedFiles {
		content := readMemoryFile(filename)
		if content != "" {
			parts = append(parts, content)
		}
	}
	parts = append(parts, "</relevant_memories>")
	return strings.Join(parts, "\n\n")
}
