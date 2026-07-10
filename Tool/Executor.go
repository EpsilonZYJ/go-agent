package Tool

import (
	"encoding/json"
	"fmt"
	"sync"
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
		return "", fmt.Errorf("unknown tool: %s", name)
	}
	return executor(input)
}

func Wrap[T any](fn func(T) (string, error)) Executor {
	return func(input json.RawMessage) (string, error) {
		var args T
		if len(input) > 0 {
			if err := json.Unmarshal(input, &args); err != nil {
				return "", fmt.Errorf("invalid tool input: %s", string(input))
			}
		}
		return fn(args)
	}
}
