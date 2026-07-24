// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package consts

import "time"

const (
	BashTimeout = time.Second * 120
)

const (
	ToolMaxPrintOutputLines = 5
)

const (
	ToolBash      = "bash"
	ToolReadFile  = "read_file"
	ToolWriteFile = "write_file"
	ToolEditFile  = "edit_file"
	ToolGlob      = "glob"
)

const (
	ToolExecuteBatch = 16
)
