// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package config

import (
	"fmt"

	"go-agent/internal/skill"
)

func BuildSystem() string {
	catalog := skill.ListSkills()
	return fmt.Sprintf(
		"You are a coding agent at %s."+
			"Before starting any multi-step task, use todo_write to plan your steps. Update status as you go. "+
			"For complex sub-problems, use the task tool to spawn a subagent."+
			"Skills available:\n%s\nUse load_skill to get full details when needed.",
		SysCfg.CurDir, catalog,
	)
}
