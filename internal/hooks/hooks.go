// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

// Package hooks 提供可选观察者的挂载点（方案 b）：
// 权限等决策逻辑保留在独立子系统，不经过 hook；这里只有通知型 hook，
// 它们只观察、不改变主流程。注册集中在启动阶段，运行时只触发。
package hooks

import (
	"sync"

	"github.com/anthropics/anthropic-sdk-go"
)

// 通知型 hook：只观察，无返回值、不阻断流程。
type (
	UserPromptSubmitHook func(query string)
	PreToolUseHook       func(block anthropic.ContentBlockUnion)
	PostToolUseHook      func(block anthropic.ContentBlockUnion, output string)
	StopHook             func(messages []anthropic.MessageParam)
)

var (
	mu               sync.RWMutex
	userPromptHooks  []UserPromptSubmitHook
	preToolUseHooks  []PreToolUseHook
	postToolUseHooks []PostToolUseHook
	stopHooks        []StopHook
)

func RegisterUserPromptSubmit(h UserPromptSubmitHook) {
	mu.Lock()
	defer mu.Unlock()
	userPromptHooks = append(userPromptHooks, h)
}

func RegisterPreToolUse(h PreToolUseHook) {
	mu.Lock()
	defer mu.Unlock()
	preToolUseHooks = append(preToolUseHooks, h)
}

func RegisterPostToolUse(h PostToolUseHook) {
	mu.Lock()
	defer mu.Unlock()
	postToolUseHooks = append(postToolUseHooks, h)
}

func RegisterStop(h StopHook) {
	mu.Lock()
	defer mu.Unlock()
	stopHooks = append(stopHooks, h)
}

func TriggerUserPromptSubmit(query string) {
	mu.RLock()
	defer mu.RUnlock()
	for _, h := range userPromptHooks {
		h(query)
	}
}

func TriggerPreToolUse(block anthropic.ContentBlockUnion) {
	mu.RLock()
	defer mu.RUnlock()
	for _, h := range preToolUseHooks {
		h(block)
	}
}

// TriggerPostToolUse 会在并发执行工具的 goroutine 里被调用，
// RLock 允许多个 goroutine 同时持有，hook 自身需保证线程安全。
func TriggerPostToolUse(block anthropic.ContentBlockUnion, output string) {
	mu.RLock()
	defer mu.RUnlock()
	for _, h := range postToolUseHooks {
		h(block, output)
	}
}

func TriggerStop(messages []anthropic.MessageParam) {
	mu.RLock()
	defer mu.RUnlock()
	for _, h := range stopHooks {
		h(messages)
	}
}
