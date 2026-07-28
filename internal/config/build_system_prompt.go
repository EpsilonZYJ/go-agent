// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package config

import (
	"fmt"
	"strings"

	"go-agent/internal/skill"
)

func BuildSystemPrompt() string {
	return strings.TrimSpace(
		fmt.Sprintf(
			"You are a coding agent at %s."+
				"Before starting any multi-step task, use todo_write to plan your steps. Update status as you go. "+
				"For complex sub-problems, use the task tool to spawn a subagent. "+
				"Skills available:\n%s\nUse load_skill to get full details when needed.",
			SysCfg.CurDir, skill.ListSkills(),
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
			SysCfg.CurDir, skill.ListSkills(),
		),
	)
}
