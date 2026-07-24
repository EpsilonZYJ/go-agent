package permission

import (
	"bufio"
	"fmt"
	"go-agent/common/consts"
	"os"
	"strings"
)

func askUser(tool_name string, args map[string]any, reason string, scanner *bufio.Scanner) consts.PermissionCode {
	fmt.Printf("\n\033[33m⚠  %s\033[0m\n", reason)
	fmt.Printf("   Tool: %s(\n", tool_name)
	for k, v := range args {
		fmt.Println("         \t", k, ":", v)
	}
	fmt.Printf("         )")
	fmt.Printf("   Allow? [y/N] ")
	var tries int = 0
	for !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			fmt.Println(err)
			tries++
		}
		if tries >= consts.IOMaxTries {
			os.Exit(consts.IOMaxTries)
		}
	}
	choice := strings.ToLower(strings.TrimSpace(scanner.Text()))
	if choice == "y" || choice == "yes" {
		return consts.PermissionAllow
	}
	return consts.PermissionDeny
}
