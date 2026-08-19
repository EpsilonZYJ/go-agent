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

// rebuildIndex scans all .md memory files under the memory directory (skipping
// MEMORY.md), aggregates each memory's name, filename, and description, and writes
// an index.md index file. If the frontmatter lacks a description, the first line of
// the body (up to 80 characters) is used instead.
func rebuildIndex() error {
	matches, err := filepath.Glob(filepath.Join(config.Cfg.System.MemoryDir, "*.md"))
	if err != nil {
		return err
	}
	var lines []string
	for _, match := range matches {
		base := filepath.Base(match)
		if base == memoryIndexFilename {
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

// readMemoryIndex reads the content of the memory index file MEMORY.md and returns
// it. It returns an empty string if the file does not exist or cannot be read.
func readMemoryIndex() string {
	if _, err := os.Stat(filepath.Join(config.Cfg.System.MemoryDir, memoryIndexFilename)); errors.Is(err, os.ErrNotExist) {
		return ""
	}
	content, err := os.ReadFile(filepath.Join(config.Cfg.System.MemoryDir, memoryIndexFilename))
	if err != nil {
		return ""
	}
	return string(content)
}
