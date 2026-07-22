// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package builtinTool

import (
	"go-agent/services"
)

func RegisterBuiltinTools(req *services.ChatRequest) error {
	if err := registerToolBash(req); err != nil {
		return err
	}
	if err := registerToolFileSystem(req); err != nil {
		return err
	}
	return nil
}
