// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package builtin

import (
	"context"
	"errors"
	"fmt"
	"go-agent/internal/consts"
	"go-agent/internal/llm"
	"go-agent/internal/logs"
	"go-agent/internal/tool"
	"os/exec"
	"strings"
)

type command struct {
	Command string `json:"command" jsonschema:"required" jsonschema_description:"The shell command to execute."`
}

func executeCommand(command string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), consts.BashTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	output, err := cmd.CombinedOutput()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		logs.Debug("bash timeout",
			"command", command,
			"timeout", consts.BashTimeout,
		)
		return "", fmt.Errorf("bash timed out after %s", consts.BashTimeout)
	}
	if err != nil {
		logs.Warn("bash command failed", "command", command, "err", err)
		return "", err
	}
	logs.Debug("bash command succeeded", "command", command, "outputLen", len(output))
	return strings.TrimSpace(string(output)), nil
}

func RunBash(command string) (string, error) {
	output, err := executeCommand(command)
	if err != nil {
		return "", err
	}
	return output, nil
}

func registerToolBash(req *llm.ChatRequest) error {
	return tool.RegisterTool(req, consts.ToolBash, "Run a shell command", func(in command) (string, error) {
		return RunBash(in.Command)
	})
}
