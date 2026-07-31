// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package permission

import (
	"fmt"
	"os"
	"strings"

	"go-agent/internal/consts"
	"go-agent/internal/logs"
	"go-agent/internal/session"
)

func askUser(tool_name string, args map[string]any, reason string) consts.PermissionCode {
	fmt.Printf("\n\033[33m⚠  %s\033[0m\n", reason)
	fmt.Printf("   Tool: %s(\n", tool_name)
	for k, v := range args {
		fmt.Println("         \t", k, ":", v)
	}
	fmt.Printf("         )")
	fmt.Printf("   Allow? [y/N] ")
	var tries int = 0
	choice, err := session.ReadLine()
	for err != nil {
		logs.Warn("user input read error during ask", "trial", tries, "err", err)
		fmt.Println(err)
		tries++
		if tries >= consts.IOMaxTries {
			logs.Error("max read tries exceeded during ask", "tries", tries)
			os.Exit(consts.ExitIOMaxTriesError)
		}
	}
	choice = strings.ToLower(strings.TrimSpace(choice))
	if choice == "y" || choice == "yes" {
		return consts.PermissionAllow
	}
	return consts.PermissionDeny
}
