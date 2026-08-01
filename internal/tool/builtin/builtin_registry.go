// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package builtin

import (
	"go-agent/internal/llm"
	"go-agent/internal/logs"
)

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
	logs.Info("builtin tools registered", "count", 6)
	return nil
}

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
	logs.Info("subagent builtin tools registered", "count", 3)
	return nil
}
