// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package builtinTool

import (
	"context"
	"errors"
	"fmt"
	"go-agent/common/consts"
	"go-agent/services"
	"go-agent/tool"
	"go-agent/utils/logs"
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
		logs.Debug("[executeCommand] bash timeout.")
		return "", fmt.Errorf("bash timed out after %s", consts.BashTimeout)
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func RunBash(command string) (string, error) {
	output, err := executeCommand(command)
	if err != nil {
		return "", err
	}
	return output, nil
}

func registerToolBash(req *services.ChatRequest) error {
	return tool.RegisterTool(req, consts.ToolBash, "Run a shell command", func(in command) (string, error) {
		return RunBash(in.Command)
	})
}
