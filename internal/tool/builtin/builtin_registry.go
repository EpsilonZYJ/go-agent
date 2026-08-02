// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package builtin

import (
	"go-agent/internal/llm"
	"go-agent/internal/logs"
)

// RegisterBuiltinTools registers all built-in tools (bash, file system, todo_write, subagent, load_skill, compact).
func RegisterBuiltinTools(req *llm.ChatRequest) error {
	if err := registerToolBash(req); err != nil {
		logs.Warn("register bash tool failed", "err", err)
		return err
	}
	if err := registerToolFileSystem(req); err != nil {
		logs.Warn("register filesystem tools failed", "err", err)
		return err
	}
	if err := registerToolTodoWrite(req); err != nil {
		logs.Warn("register todo_write tool failed", "err", err)
		return err
	}
	if err := registerToolSubagent(req); err != nil {
		logs.Warn("register subagent tool failed", "err", err)
		return err
	}
	if err := registerToolLoadSkill(req); err != nil {
		logs.Warn("register load_skill tool failed", "err", err)
		return err
	}
	if err := registerToolCompact(req); err != nil {
		logs.Warn("register compact tool failed", "err", err)
		return err
	}
	logs.Info("builtin tools registered", "count", len(req.Tools))
	return nil
}

// SubAgentRegisterBuiltinTools registers the built-in tools available to subagents (bash, file system, load_skill).
func SubAgentRegisterBuiltinTools(req *llm.ChatRequest) error {
	if err := registerToolBash(req); err != nil {
		logs.Warn("subagent register bash tool failed", "err", err)
		return err
	}
	if err := registerToolFileSystem(req); err != nil {
		logs.Warn("subagent register filesystem tools failed", "err", err)
		return err
	}
	if err := registerToolLoadSkill(req); err != nil {
		logs.Warn("subagent register load_skill tool failed", "err", err)
		return err
	}
	logs.Info("subagent builtin tools registered", "count", len(req.Tools))
	return nil
}
