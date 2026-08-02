// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package builtin

import (
	"go-agent/internal/consts"
	"go-agent/internal/llm"
	"go-agent/internal/skill"
	"go-agent/internal/tool"
)

type skillName struct {
	Name string `json:"name" jsonschema:"required" jsonschema_description:"The name of the skill to load."`
}

// RunLoadSkill loads the skill with the given name and returns its full content.
func RunLoadSkill(name string) (string, error) {
	return skill.GetSkill(name)
}

// registerToolLoadSkill registers the load_skill tool with the request.
func registerToolLoadSkill(req *llm.ChatRequest) error {
	return tool.RegisterTool(req, consts.ToolLoadSkill, "Load the full content of a skill by name.", func(in skillName) (string, error) {
		return RunLoadSkill(in.Name)
	})
}
