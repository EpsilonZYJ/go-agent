package Tool

import (
	"context"
	"errors"
	"fmt"
	"go-agent/Const"
	"go-agent/Model"
	"go-agent/Services"
	"os/exec"
	"strings"
)

type Command struct {
	Command string `json:"command"`
}

func executeCommand(command string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), Const.BashTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	output, err := cmd.CombinedOutput()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return "", fmt.Errorf("bash timed out after %s", Const.BashTimeout)
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

func registerToolBash(req *Services.ChatRequest) {
	req.AddTool(Model.Tool{
		Name:        "bash",
		Description: "Run a shell command",
		InputSchema: Model.InputSchema{
			Type: "object",
			Properties: map[string]Model.Property{
				"command": {
					Type:        "string",
					Description: "",
				},
			},
			Required: []string{"command"},
		},
	}.ToAnthropicTool())
}
