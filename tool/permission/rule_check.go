package permission

import (
	"fmt"
	"go-agent/utils/baseImpl"
)

type checkPassFunc func(map[string]string) bool
type rule struct {
	Tools   []string
	Check   checkPassFunc
	Message string
}

// TODO: add rules
var permissionRules = []rule{}

func check_rules(tool_name string, args map[string]string) error {
	for _, r := range permissionRules {
		if baseImpl.ListContains(r.Tools, tool_name) {
			if r.Check(args) {
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
