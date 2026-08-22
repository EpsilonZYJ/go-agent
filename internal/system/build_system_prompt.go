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
	c.mu.Lock()
	defer c.mu.Unlock()

	key := fmt.Sprintf("%s|%s|%v", ctx.Workspace, ctx.Memories, ctx.EnabledTools)

	if key == c.lastKey && c.lastPrompt != "" {
		fmt.Printf("  \033[90m[cache hit] system prompt unchanged\033[0m\n")
		return c.lastPrompt
	}
	c.lastKey = key

	prompt, loaded := assembleSystemPrompt(ctx)
	c.lastPrompt = prompt
	fmt.Printf("  \033[32m[assembled] sections: %s\033[0m\n", strings.Join(loaded, ", "))
	return c.lastPrompt
}

var promptSections = map[string]string{
	"identity":  "You are a coding agent. Act, don't explain.",
	"tools":     "Available tools: bash, read_file, write_file.",
	"workspace": "Working directory: {WORKDIR}",
	"workflow": "Before starting any multi-step task, use todo_write to plan your steps. Update status as you go. " +
		"For complex sub-problems, use the task tool to spawn a subagent.",
	"memory": "Relevant memories are injected below. Respect user preferences from memory.\n" +
		"When the user says 'remember' or expresses a clear preference, extract it as a memory.",
	"skills": "Skills available:\n{SKILLS}\nUse load_skill to get full details when needed.",
}

func assembleSystemPrompt(context promptContext) (string, []string) {
	type namedSection struct{ name, content string }
	sections := []namedSection{
		{"identity", promptSections["identity"]},
		{"tools", promptSections["tools"]},
		{"workspace", strings.ReplaceAll(promptSections["workspace"], "{WORKDIR}", context.Workspace)},
		{"workflow", promptSections["workflow"]},
		{"memory", promptSections["memory"]},
	}
	if context.Memories != "" {
		sections = append(sections, namedSection{"memory-content", fmt.Sprintf("Relevant memories:\n%s", context.Memories)})
	}
	sections = append(sections,
		namedSection{"skills", strings.ReplaceAll(promptSections["skills"], "{SKILLS}", skill.ListSkills())})

	names := make([]string, 0, len(sections))
	parts := make([]string, 0, len(sections))
	for _, s := range sections {
		names = append(names, s.name)
		parts = append(parts, s.content)
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n")), names
}

// deriveContext 从真实状态派生 prompt 上下文：工作目录与记忆索引。
func deriveContext() promptContext {
	return promptContext{
		Workspace: config.Cfg.System.CurDir,
		Memories:  memory.ReadMemoryIndex(),
	}
}

func BuildSystemPrompt() string {
	return pc.getSystemPrompt(deriveContext())
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
