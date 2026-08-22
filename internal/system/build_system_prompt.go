// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package system

import (
	"bytes"
	"encoding/json"
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

func marshalKey(ctx map[string]string) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(ctx)
	return buf.String()
}

func (c *promptCache) Get(ctx map[string]string) string {
	snapshot := make(map[string]string, len(ctx))
	for k, v := range ctx {
		snapshot[k] = v
	}

	key := marshalKey(snapshot)
	c.mu.Lock()
	defer c.mu.Unlock()

	if key == c.lastKey && c.lastPrompt != "" {
		fmt.Printf("  \033[90m[cache hit] system prompt unchanged\033[0m")
		return c.lastPrompt
	}
	c.lastKey = key
	c.lastPrompt = assembleSystemPrompt(snapshot)

	loaded := []string{"identity", "tools", "workspace"}
	if ctx["memories"] != "" {
		loaded = append(loaded, "memory")
	}
	fmt.Printf("  \033[32m[assembled] sections: %s\033[0m", strings.Join(loaded, ", "))
	return c.lastPrompt
}

func promptSections() map[string]string {
	return map[string]string{
		"identity":  "You are a coding agent. Act, don't explain.",
		"tools":     "Available tools: bash, read_file, write_file.",
		"workspace": fmt.Sprintf("Working directory: %s", config.Cfg.System.CurDir),
		"memory":    "Relevant memories are injected below when available.",
	}
}

func assembleSystemPrompt(context map[string]string) string {
	var sections []string

	ps := promptSections()
	sections = append(sections, ps["identity"])
	sections = append(sections, ps["tools"])
	sections = append(sections, ps["workspace"])

	memories := context["memories"]
	if memories != "" {
		sections = append(sections, fmt.Sprintf("Relevant memories:\n%s", memories))
	}

	return strings.Join(sections, "\n\n")
}

func BuildSystemPrompt() string {
	memIndex := memory.ReadMemoryIndex()
	var memoriesSection string
	if memIndex == "" {
		memoriesSection = ""
	} else {
		memoriesSection = fmt.Sprintf("\n\nMemories available:\n%s", memIndex)
	}
	return strings.TrimSpace(
		fmt.Sprintf(
			"You are a coding agent at %s. "+
				"%s"+
				"Relevant memories are injected below. Respect user preferences from memory.\n"+
				"When the user says 'remember' or expresses a clear preference, extract it as a memory. "+
				"Before starting any multi-step task, use todo_write to plan your steps. Update status as you go. "+
				"For complex sub-problems, use the task tool to spawn a subagent. "+
				"Skills available:\n%s\nUse load_skill to get full details when needed.",
			config.Cfg.System.CurDir, memoriesSection, skill.ListSkills(),
		),
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
