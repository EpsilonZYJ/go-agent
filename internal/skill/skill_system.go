// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package skill

import (
	"fmt"
	"go-agent/internal/utils"
	"os"
	"path/filepath"
	"strings"

	"go-agent/internal/logs"
)

func ScanSkills(skillDir string) {
	entries, err := os.ReadDir(skillDir)
	if err != nil {
		logs.Warn("skills directory not found", "dir", skillDir, "err", err)
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(skillDir, entry.Name(), "SKILL.md")
		raw, err := os.ReadFile(path)
		if err != nil {
			logs.Warn("skill file read failed", "path", path, "err", err)
			continue
		}
		content := string(raw)
		metadata, body := utils.ParseFrontmatter(content)
		name := metadata["name"]
		if name == "" {
			name = entry.Name()
		}
		desc := metadata["description"]
		if desc == "" {
			firstline := strings.SplitN(body, "\n", 2)[0]
			desc = strings.TrimSpace(strings.TrimLeft(firstline, "#"))
		}
		skillRegistry[name] = Skill{
			Name:        name,
			Description: desc,
			Content:     content,
		}
		logs.Info("skill loaded", "name", name)
	}
	logs.Info("skills scan complete", "count", len(skillRegistry))
}

func ListSkills() string {
	if len(skillRegistry) == 0 {
		return "(no skills found)"
	}
	skillDesc := make([]string, len(skillRegistry))
	var i int = 0
	for _, s := range skillRegistry {
		skillDesc[i] = fmt.Sprintf("- **%s**: %s", s.Name, s.Description)
		i++
	}
	return strings.Join(skillDesc, "\n")
}

func GetSkill(name string) (string, error) {
	skillToFind, ok := skillRegistry[name]
	if !ok {
		logs.Warn("skill not found", "name", name)
		return "", fmt.Errorf("Skill not found: %s", name)
	}
	return skillToFind.Content, nil
}
