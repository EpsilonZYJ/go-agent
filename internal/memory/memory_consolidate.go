// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package memory

import (
	"encoding/json"
	"fmt"
	"go-agent/internal/config"
	"go-agent/internal/consts"
	"go-agent/internal/llm"
	"go-agent/internal/logs"
	"go-agent/internal/utils"
	"os"
	"path/filepath"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

func ConsolidateMemory() {
	files := listMemoryFiles()
	if len(files) < consts.MemoryConsolidateThreshold {
		return
	}

	var catalog string
	{
		var catalogs []string
		for _, file := range files {
			catalogs = append(catalogs, fmt.Sprintf("## %s\nname: %s\ndescription: %s\n%s",
				file.Filename, file.Name, file.Description, file.Body,
			))
		}
		catalog = strings.Join(catalogs, "\n\n")
	}

	prompt := fmt.Sprintf(
		"Consolidate the following memory files. Rules:\n"+
			"1. Merge duplicates into one\n"+
			"2. Remove outdated/contradicted memories\n"+
			"3. Keep the total under 30 memories\n"+
			"4. Preserve important user preferences above all\n"+
			"Return a JSON array. Each item: {name, type, description, body}.\n\n"+
			"%s",
		utils.StringTruncateRunes(catalog, 16000),
	)

	resp, err := llm.Call(
		anthropic.MessageNewParams{
			Messages: []anthropic.MessageParam{
				anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
			},
			MaxTokens: 3000,
			Model:     config.Cfg.Model.Model,
		},
		0,
	)
	if err != nil {
		return
	}
	text := utils.GetTextFromAnthropicMessageParam(resp.ToParam())
	text = strings.TrimSpace(text)
	match := memExtractionRe.FindString(text)
	if match == "" {
		return
	}
	var items []extractedMemory
	if err := json.Unmarshal([]byte(match), &items); err != nil {
		return
	}
	matches, globErr := filepath.Glob(filepath.Join(config.Cfg.System.MemoryDir, "*.md"))
	if globErr != nil {
		return
	}
	for _, m := range matches {
		if filepath.Base(m) == memoryIndexFilename {
			continue
		}
		if err := os.Remove(m); err != nil {
			logs.Warn("remove memory file failed", "path", m, "err", err)
		}
	}
	for _, mem := range items {
		if mem.Description == "" || mem.Body == "" {
			continue
		}
		if mem.Name == "" {
			mem.Name = fmt.Sprintf("memory_%s", utils.NowTime())
		}
		if mem.Type == "" {
			mem.Type = MemTypeUser
		}
		_, err := writeMemoryFile(mem.Name, mem.Type, mem.Description, mem.Body)
		if err == nil {
			continue
		}
	}
	addNotice(fmt.Sprintf("[Memory: consolidated %d → %d memories]", len(files), len(items)))
}
