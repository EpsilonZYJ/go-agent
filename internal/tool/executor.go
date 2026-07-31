// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package tool

import (
	"encoding/json"
	"fmt"
	"sync"

	"go-agent/internal/logs"
)

type Executor func(input json.RawMessage) (string, error)

var (
	execMu  sync.RWMutex
	execMap = map[string]Executor{}
)

func RegisterExecutor(name string, exec Executor) {
	execMu.Lock()
	defer execMu.Unlock()
	execMap[name] = exec
}

func Dispatch(name string, input json.RawMessage) (string, error) {
	execMu.RLock()
	executor, ok := execMap[name]
	execMu.RUnlock()
	if !ok {
		logs.Warn("unknown tool dispatched", "tool", name)
		return "", fmt.Errorf("unknown tool: %s", name)
	}
	return executor(input)
}

func Wrap[T any](fn func(T) (string, error)) Executor {
	return func(input json.RawMessage) (string, error) {
		var args T
		if len(input) > 0 {
			if err := json.Unmarshal(input, &args); err != nil {
				logs.Warn("tool input unmarshal failed", "input", string(input), "err", err)
				return "", fmt.Errorf("invalid tool input: %s", string(input))
			}
		}
		return fn(args)
	}
}
