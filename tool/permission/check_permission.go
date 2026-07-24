package permission

import (
	"bufio"
	"encoding/json"
	"fmt"
	"go-agent/common/consts"

	"github.com/anthropics/anthropic-sdk-go"
)

func CheckPermission(block anthropic.ContentBlockUnion) (consts.PermissionCode, error) {
	var raw map[string]string
	err := json.Unmarshal(block.Input, &raw)
	if err != nil {
		return consts.PermissionInputInvalid, fmt.Errorf("cannot parse input")
	}
	if block.Name == consts.ToolBash {
		command, ok := raw["command"]
		if !ok {
			return consts.PermissionInputInvalid, fmt.Errorf("tool bash command not found")
		}
		err = check_deny_list(command)
		if err != nil {
			return consts.PermissionDeny, fmt.Errorf("\n\033[31m⛔ %s\033[0m\n", err.Error())
		}
	}
	err = check_rules(block.Name, raw)
	if err != nil {
		return consts.PermissionAskUser, err
	}
	return consts.PermissionAllow, nil
}

func AskUser(block anthropic.ContentBlockUnion, scanner *bufio.Scanner, reason string) consts.PermissionCode {
	var raw map[string]string
	err := json.Unmarshal(block.Input, &raw)
	if err != nil {
		return consts.PermissionInputInvalid
	}
	decision := askUser(block.Name, raw, reason, scanner)
	return decision
}
