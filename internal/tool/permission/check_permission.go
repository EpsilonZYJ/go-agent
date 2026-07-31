// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package permission

import (
	"encoding/json"
	"fmt"

	"go-agent/internal/consts"
	"go-agent/internal/logs"

	"github.com/anthropics/anthropic-sdk-go"
)

func CheckPermission(block anthropic.ContentBlockUnion) (consts.PermissionCode, error) {
	var raw map[string]any
	err := json.Unmarshal(block.Input, &raw)
	if err != nil {
		logs.Warn("cannot parse tool input for permission check", "tool", block.Name, "err", err)
		return consts.PermissionInputInvalid, fmt.Errorf("cannot parse input")
	}
	if block.Name == consts.ToolBash {
		command, ok := raw["command"].(string)
		if !ok {
			return consts.PermissionInputInvalid, fmt.Errorf("tool bash command not found")
		}
		err = check_deny_list(command)
		if err != nil {
			return consts.PermissionDeny, fmt.Errorf("\n\033[31m⛔ %s\033[0m\n", err.Error())
		}
	}
	err = checkRules(block.Name, raw)
	if err != nil {
		logs.Info("rule triggered ask user", "tool", block.Name, "reason", err.Error())
		return consts.PermissionAskUser, err
	}
	return consts.PermissionAllow, nil
}

func AskUser(block anthropic.ContentBlockUnion, reason string) consts.PermissionCode {
	var raw map[string]any
	err := json.Unmarshal(block.Input, &raw)
	if err != nil {
		return consts.PermissionInputInvalid
	}
	decision := askUser(block.Name, raw, reason)
	return decision
}
