package tool

import (
	"context"
	"errors"
	"fmt"
	"go-agent/common/consts"
	"go-agent/model"
	"go-agent/services"
	"go-agent/utils/logs"
	"os/exec"
	"strings"
)

type command struct {
	Command string `json:"command"`
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

var dangerous = []string{"rm -rf /", "sudo", "shutdown", "reboot", "> /dev/"}

func RunBash(command string) (string, error) {
	for _, d := range dangerous {
		if strings.Contains(command, d) {
			return "", fmt.Errorf("dangerous command:%s, contains %s", command, d)
		}
	}
	output, err := executeCommand(command)
	if err != nil {
		return "", err
	}
	return output, nil
}

func registerToolBash(req *services.ChatRequest) {
	req.AddTool(model.Tool{
		Name:        "bash",
		Description: "Run a shell command",
		InputSchema: model.InputSchema{
			Type: "object",
			Properties: map[string]model.Property{
				"command": {
					Type:        "string",
					Description: "",
				},
			},
			Required: []string{"command"},
		},
	}.ToAnthropicTool())
	RegisterExecutor("bash", Wrap(func(in command) (string, error) {
		return RunBash(in.Command)
	}))
}
