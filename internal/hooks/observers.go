// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package hooks

import (
	"fmt"
	"sync"

	"go-agent/internal/config"
	"go-agent/internal/consts"

	"github.com/anthropics/anthropic-sdk-go"
)

// printMu 串行化并发 PostToolUse 里的打印，避免多工具并发时输出交错。
var printMu sync.Mutex

// workdirHook: UserPromptSubmit —— 提示当前工作目录。
func workdirHook(query string) {
	fmt.Printf("\033[90m[HOOK] UserPromptSubmit: working in %s\033[0m\n", config.SysCfg.CurDir)
}

// logHook: PreToolUse —— 记录每次工具调用（单线程 collect 阶段触发，无需加锁）。
func logHook(block anthropic.ContentBlockUnion) {
	args := string(block.Input)
	if len([]rune(args)) > 60 {
		args = string([]rune(args)[:60]) + "..."
	}
	fmt.Printf("\033[90m[HOOK] %s(%s)\033[0m\n", block.Name, args)
}

// largeOutputHook: PostToolUse —— 工具输出过大时告警（并发触发，打印需加锁）。
func largeOutputHook(block anthropic.ContentBlockUnion, output string) {
	if len(output) <= consts.LargeOutputChars {
		return
	}
	printMu.Lock()
	defer printMu.Unlock()
	fmt.Printf("\033[33m[HOOK] ⚠ Large output from %s: %d chars\033[0m\n", block.Name, len(output))
}

// summaryHook: Stop —— 统计本会话的工具调用次数。
func summaryHook(messages []anthropic.MessageParam) {
	count := 0
	for _, m := range messages {
		for _, b := range m.Content {
			if b.OfToolResult != nil {
				count++
			}
		}
	}
	fmt.Printf("\033[90m[HOOK] Stop: session used %d tool calls\033[0m\n", count)
}

// RegisterBuiltinObservers 注册全部内置观察者 hook，在启动阶段调用一次。
func RegisterBuiltinObservers() {
	RegisterUserPromptSubmit(workdirHook)
	RegisterPreToolUse(logHook)
	RegisterPostToolUse(largeOutputHook)
	RegisterStop(summaryHook)
}
