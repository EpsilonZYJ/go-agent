// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package permission

import (
	"fmt"
	"strings"

	"go-agent/internal/logs"
)

var denyList = []string{
	"rm -rf /", "sudo", "shutdown", "reboot", "mkfs", "dd if=", "> /dev/sda",
}

func check_deny_list(command string) error {
	for _, pattern := range denyList {
		if strings.Contains(command, pattern) {
			logs.Warn("deny list matched", "command", command, "pattern", pattern)
			return fmt.Errorf("Blocked: %s", pattern)
		}
	}
	return nil
}
