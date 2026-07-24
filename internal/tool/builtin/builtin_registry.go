// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package builtin

import (
	"go-agent/internal/llm"
)

func RegisterBuiltinTools(req *llm.ChatRequest) error {
	if err := registerToolBash(req); err != nil {
		return err
	}
	if err := registerToolFileSystem(req); err != nil {
		return err
	}
	return nil
}
