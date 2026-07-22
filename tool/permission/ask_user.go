package permission

import (
	"bufio"
	"fmt"
	"go-agent/common/consts"
	"os"
	"strings"
)

func askUser(tool_name string, args map[string]string, reason string, scanner *bufio.Scanner) consts.PermissionCode {
	fmt.Printf("\n\033[33m⚠  %s\033[0m\n", reason)
	fmt.Printf("   Tool: %s(%v)", tool_name, args)
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
