// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package memory

import (
	"fmt"
	"go-agent/internal/config"
	"go-agent/internal/logs"
	"go-agent/internal/utils"
	"os"
	"path/filepath"
	"strings"
)

func writeMemoryFile(name string, memType MemType, description string, body string) (string, error) {
	name = strings.ToLower(name)
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "/", "-")
	filename := name + ".md"
	path := filepath.Join(config.SysCfg.MemoryDir, filename)
	file, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	_, err = file.WriteString(
		fmt.Sprintf("---\nname: %s\ndescription: %s\ntype: %s\n---\n\n%s\n", name, description, memType, body),
	)
	_ = rebuildIndex()
	return path, err
}

func readMemoryFile(filename string) string {
	path := filepath.Join(config.SysCfg.MemoryDir, filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return ""
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(content)
}

func listMemoryFiles() []Memory {
	var result []Memory
	matches, err := filepath.Glob(filepath.Join(config.SysCfg.MemoryDir, "*.md"))
	if err != nil {
		return result
	}
	for _, match := range matches {
		base := filepath.Base(match)
		if base == "MEMORY.md" {
			continue
		}
		raw, err := os.ReadFile(match)
		if err != nil {
			logs.Warn("read memory file failed", "path", match, "err", err)
			continue
		}
		meta, body := utils.ParseFrontmatter(string(raw))
		if meta["name"] == "" {
			meta["name"] = strings.TrimSuffix(base, ".md")
		}
		if meta["type"] == "" {
			meta["type"] = string(MemTypeUser)
		}
		result = append(result, Memory{
			Filename:    base,
			Name:        meta["name"],
			Description: meta["description"],
			Type:        MemType(meta["type"]),
			Body:        body,
		})
	}
	return result
}
