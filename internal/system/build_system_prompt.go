// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package system

import (
	"fmt"
	"go-agent/internal/config"
	"go-agent/internal/memory"
	"strings"

	"go-agent/internal/skill"
)

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
				"When the user says 'remember' or expresses a clear preference, extract it as a memory."+
				"Before starting any multi-step task, use todo_write to plan your steps. Update status as you go. "+
				"For complex sub-problems, use the task tool to spawn a subagent. "+
				"Skills available:\n%s\nUse load_skill to get full details when needed.",
			memoriesSection, config.Cfg.System.CurDir, skill.ListSkills(),
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
