// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package consts

import "time"

const (
	BashTimeout = time.Second * 120
)

const (
	ToolMaxPrintOutputLines = 5
)

// LargeOutputChars 工具输出字符数超过该值时，由 PostToolUse 的大输出观察者告警
const LargeOutputChars = 100000

const (
	ToolBash      = "bash"
	ToolReadFile  = "read_file"
	ToolWriteFile = "write_file"
	ToolEditFile  = "edit_file"
	ToolGlob      = "glob"
	ToolTodoWrite = "todo_write"
	ToolSubagent  = "task"
	ToolLoadSkill = "load_skill"
)

const (
	ToolExecuteBatch = 16
)

const TodoReminderRounds = 3

// SubAgentSafetyLimit subagent最大执行轮次
const SubAgentSafetyLimit = 30
