package permission

import (
	"fmt"
	"strings"
)

var denyList = []string{
	"rm -rf /", "sudo", "shutdown", "reboot", "mkfs", "dd if=", "> /dev/sda",
}

func check_deny_list(command string) error {
	for _, pattern := range denyList {
		if strings.Contains(command, pattern) {
			return fmt.Errorf("Blocked: %s", pattern)
		}
	}
	return nil
}
