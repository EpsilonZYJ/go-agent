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

// writeMemoryFile writes a memory entry to a Markdown file under the memory directory.
// The filename is derived from name by normalizing it (lowercase, trimmed, with spaces
// and slashes replaced by hyphens) and appending the .md extension. The file content is
// a Markdown document with frontmatter (name/description/type).
// After a successful write, it rebuilds the index and returns the absolute file path.
func writeMemoryFile(name string, memType MemType, description string, body string) (string, error) {
	name = strings.ToLower(name)
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "/", "-")
	filename := name + ".md"
	path := filepath.Join(config.Cfg.System.MemoryDir, filename)
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

// readMemoryFile reads the Markdown file with the given filename under the memory
// directory and returns its content. It returns an empty string if the file does not
// exist or cannot be read.
func readMemoryFile(filename string) string {
	path := filepath.Join(config.Cfg.System.MemoryDir, filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return ""
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(content)
}

// listMemoryFiles walks all .md memory files under the memory directory (skipping
// MEMORY.md), parses each file's frontmatter and body, and returns a list of Memory.
// When the name or type metadata is missing, the filename and MemTypeUser are used
// as defaults respectively.
func listMemoryFiles() []Memory {
	var result []Memory
	matches, err := filepath.Glob(filepath.Join(config.Cfg.System.MemoryDir, "*.md"))
	if err != nil {
		return result
	}
	for _, match := range matches {
		base := filepath.Base(match)
		if base == memoryIndexFilename {
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
