// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package system

import (
	"fmt"
	"strings"
	"sync"

	"go-agent/internal/config"
	"go-agent/internal/memory"

	"go-agent/internal/skill"
)

type promptCache struct {
	mu         sync.Mutex
	lastKey    string
	lastPrompt string
}

type promptContext struct {
	Workspace    string
	Memories     string
	EnabledTools []string
}

var pc promptCache

func (c *promptCache) getSystemPrompt(ctx promptContext) string {
	key := fmt.Sprintf("%s|%s|%v", ctx.Workspace, ctx.Memories, ctx.EnabledTools)

	if key == c.lastKey && c.lastPrompt != "" {
		fmt.Printf("  \033[90m[cache hit] system prompt unchanged\033[0m")
		return c.lastPrompt
	}
	c.lastKey = key
	c.lastPrompt = assembleSystemPrompt(ctx)

	loaded := []string{"identity", "tools", "workspace"}
	if ctx.Memories != "" {
		loaded = append(loaded, "memory")
	}
	fmt.Printf("  \033[32m[assembled] sections: %s\033[0m", strings.Join(loaded, ", "))
	return c.lastPrompt
}

var promptSections = map[string]string{
	"identity":  "You are a coding agent. Act, don't explain.",
	"tools":     "Available tools: bash, read_file, write_file.",
	"workspace": "Working directory: ",
	"memory":    "Relevant memories are injected below when available.",
	"skills":    "load_skill to get full details when needed. Skills available:\n",
}

func assembleSystemPrompt(context promptContext) string {
	var sections []string

	sections = append(sections, promptSections["identity"])
	sections = append(sections, promptSections["tools"])
	sections = append(sections, promptSections["workspace"]+config.Cfg.System.CurDir)

	memories := context.Memories
	if memories != "" {
		sections = append(sections, fmt.Sprintf("Relevant memories:\n%s", memories))
	}
	sections = append(sections, promptSections["skills"]+skill.ListSkills()+"\n")

	return strings.TrimSpace(strings.Join(sections, "\n\n"))
}

func UpdateContext(context map[string]string) map[string]string {
	memories := ""
	memIndex := memory.ReadMemoryIndex()
	if memIndex != "" {
		memories = memIndex
	}
	return map[string]string{
		"workspace": string(config.Cfg.System.CurDir),
		"memories":  memories,
	}
}

func BuildSystemPrompt() string {
	return pc.getSystemPrompt(
		promptContext{
			Workspace: config.Cfg.System.CurDir,
			Memories:  memory.ReadMemoryIndex(),
		},
	)
}

func BuildSubSystemPrompt() string {
	return strings.TrimSpace(
		fmt.Sprintf(
			"You are a coding agent at %s. "+
				"Complete the task you were given, then return a concise summary. "+
				"Do not delegate further. "+
				"Skills available:\n%s\nUse load_skill to get full details when needed. ",
			config.Cfg.System.CurDir, skill.ListSkills(),
		),
	)
}
