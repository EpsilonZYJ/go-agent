// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go-agent/internal/config"
)

func scanSkills() {
	entries, err := os.ReadDir(config.SysCfg.SkillsDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(config.SysCfg.SkillsDir, entry.Name(), "SKILL.md")
		raw, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		content := string(raw)
		metadata, body := parseFrontmatter(content)
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
	}
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
