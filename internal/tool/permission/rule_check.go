package permission

import (
	"fmt"
	"go-agent/internal/baseImpl"
	"go-agent/internal/files"
	"strings"
)

type checkPassFunc func(map[string]any) bool
type rule struct {
	Tools     []string
	CheckPass checkPassFunc
	Message   string
}

var bashCheckCommand = []string{
	"rm ", "> /etc/", "chmod 777",
}

var permissionRules = []rule{
	{
		Tools:     []string{"write_file", "edit_file"},
		CheckPass: writeNotOutsideWorkspace,
		Message:   "Writing outside workspace",
	},
	{
		Tools:     []string{"bash"},
		CheckPass: bashIsNotDestructive,
		Message:   "Potentially destructive command",
	},
}

func checkRules(tool_name string, args map[string]any) error {
	for _, r := range permissionRules {
		if baseImpl.ListContains(r.Tools, tool_name) {
			if r.CheckPass(args) {
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

func writeNotOutsideWorkspace(args map[string]any) bool {
	relPath, ok := args["path"].(string)
	if !ok {
		return false
	}
	return files.PathIsSafe(relPath)
}

func bashIsNotDestructive(args map[string]any) bool {
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
