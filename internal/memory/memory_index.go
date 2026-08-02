// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package memory

import (
	"errors"
	"fmt"
	"go-agent/internal/config"
	"go-agent/internal/utils"
	"os"
	"path/filepath"
	"strings"
)

func rebuildIndex() error {
	matches, err := filepath.Glob(filepath.Join(config.Cfg.System.MemoryDir, "*.md"))
	if err != nil {
		return err
	}
	var lines []string
	for _, match := range matches {
		base := filepath.Base(match)
		if base == "MEMORY.md" {
			continue
		}
		raw, err := os.ReadFile(match)
		if err != nil {
			return err
		}
		meta, body := utils.ParseFrontmatter(string(raw))
		name := meta["name"]
		if name == "" {
			name = strings.TrimSuffix(base, ".md")
		}
		desc := meta["description"]
		if desc == "" {
			firstline := strings.SplitN(body, "\n", 2)[0]
			r := []rune(strings.TrimSpace(firstline))
			if len(r) > 80 {
				r = r[:80]
			}
			desc = string(r)
		}
		lines = append(lines, fmt.Sprintf("- [%s](%s) — %s", name, base, desc))
	}
	content := ""
	if len(lines) > 0 {
		content = strings.Join(lines, "\n") + "\n"
	}
	return os.WriteFile(filepath.Join(config.Cfg.System.MemoryDir, "index.md"), []byte(content), 0644)
}

func readMemoryIndex() string {
	if _, err := os.Stat(filepath.Join(config.Cfg.System.MemoryDir, "MEMORY.md")); errors.Is(err, os.ErrNotExist) {
		return ""
	}
	content, err := os.ReadFile(filepath.Join(config.Cfg.System.MemoryDir, "MEMORY.md"))
	if err != nil {
		return ""
	}
	return string(content)
}
