package permission

import (
	"fmt"
	"go-agent/utils/baseImpl"
	"go-agent/utils/files"
	"strings"
)

type checkPassFunc func(map[string]any) bool
type rule struct {
	Tools   []string
	Check   checkPassFunc
	Message string
}

var bashCheckCommand = []string{
	"rm ", "> /etc/", "chmod 777",
}

var permissionRules = []rule{
	{
		Tools:   []string{"write_file", "edit_file"},
		Check:   writeOutsideWorkspaceCheck,
		Message: "Writing outside workspace",
	},
	{
		Tools:   []string{"bash"},
		Check:   bashDestructiveCheck,
		Message: "Potentially destructive command",
	},
}

func checkRules(tool_name string, args map[string]any) error {
	for _, r := range permissionRules {
		if baseImpl.ListContains(r.Tools, tool_name) {
			if !r.Check(args) {
				return nil
			} else {
				return fmt.Errorf("%s", r.Message)
			}
		} else {
			continue
		}
	}
	return nil
}

func writeOutsideWorkspaceCheck(args map[string]any) bool {
	relPath, ok := args["path"].(string)
	if !ok {
		return false
	}
	return files.PathIsSafe(relPath)
}

func bashDestructiveCheck(args map[string]any) bool {
	command, ok := args["command"].(string)
	if !ok {
		return false
	}
	for _, pattern := range bashCheckCommand {
		if strings.Contains(command, pattern) {
			return false
		}
	}
	return true
}
