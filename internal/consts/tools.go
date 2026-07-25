// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package consts

import "time"

const (
	BashTimeout = time.Second * 120
)

const (
	ToolMaxPrintOutputLines = 5
)

// 工具输出字符数超过该值时，由 PostToolUse 的大输出观察者告警
const LargeOutputChars = 100000

const (
	ToolBash      = "bash"
	ToolReadFile  = "read_file"
	ToolWriteFile = "write_file"
	ToolEditFile  = "edit_file"
	ToolGlob      = "glob"
	ToolTodoWrite = "todo_write"
)

const (
	ToolExecuteBatch = 16
)

const TodoReminderRounds = 3
