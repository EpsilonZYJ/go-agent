// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package tool

import (
	"go-agent/services"
)

func RegisterTools(req *services.ChatRequest) {
	registerToolBash(req)
	registerToolFileSystem(req)
}
