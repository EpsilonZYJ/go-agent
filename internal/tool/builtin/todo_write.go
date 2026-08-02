// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package builtin

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"go-agent/internal/consts"
	"go-agent/internal/llm"
	"go-agent/internal/logs"
	"go-agent/internal/tool"
)

type TodoStatus string

const (
	Pending    TodoStatus = "pending"
	InProgress TodoStatus = "in_progress"
	Completed  TodoStatus = "completed"
)

type Todo struct {
	Content string     `json:"content" jsonschema:"required"`
	Status  TodoStatus `json:"status" jsonschema:"required,enum=pending,enum=in_progress,enum=completed"`
}

// todoWriteInput is the input for todo_write: {"todos": [...]}, corresponding to the todos array object.
type todoWriteInput struct {
	Todos []Todo `json:"todos" jsonschema:"required"`
}

var (
	todoMu       sync.RWMutex
	currentTodos = []Todo{}
)

// validateTodos validates the todos list, ensuring each item has a content and a valid status.
func validateTodos(todos []Todo) ([]Todo, error) {
	for i, t := range todos {
		if t.Content == "" || t.Status == "" {
			return nil, fmt.Errorf("Error: todos[%d] missing 'content' or 'status'", i)
		}
		switch t.Status {
		case Pending, InProgress, Completed:
		default:
			return nil, fmt.Errorf("Error: todos[%d] has invalid status '%s'", i, t.Status)
		}
	}
	return todos, nil
}

// normalizeTodos normalizes various forms of todos input (object, array, JSON string, etc.) into []Todo.
func normalizeTodos(todos any) ([]Todo, error) {
	if s, ok := todos.(string); ok {
		var parsed any
		if err := json.Unmarshal([]byte(s), &parsed); err != nil {
			return nil, fmt.Errorf("Error: todos must be a list or JSON array string")
		}
		todos = parsed
	}

	var rawList []any
	switch v := todos.(type) {
	case []Todo:
		return validateTodos(v)
	case []any:
		rawList = v
	case []map[string]any:
		rawList = make([]any, len(v))
		for i, m := range v {
			rawList[i] = m
		}
	default:
		// Try a JSON round-trip to handle map[string]interface{} slices, etc.
		b, err := json.Marshal(todos)
		if err != nil {
			return nil, fmt.Errorf("Error: todos must be a list")
		}
		if err := json.Unmarshal(b, &rawList); err != nil {
			return nil, fmt.Errorf("Error: todos must be a list")
		}
	}

	result := make([]Todo, 0, len(rawList))
	for i, item := range rawList {
		m, ok := item.(map[string]any)
		if !ok {
			b, err := json.Marshal(item)
			if err != nil {
				return nil, fmt.Errorf("Error: todos[%d] must be an object", i)
			}
			var tmp map[string]any
			if err := json.Unmarshal(b, &tmp); err != nil {
				return nil, fmt.Errorf("Error: todos[%d] must be an object", i)
			}
			m = tmp
		}

		content, _ := m["content"].(string)
		status, _ := m["status"].(string)
		if content == "" || status == "" {
			// Field is missing or has the wrong type.
			if _, hasC := m["content"]; !hasC {
				return nil, fmt.Errorf("Error: todos[%d] missing 'content' or 'status'", i)
			}
			if _, hasS := m["status"]; !hasS {
				return nil, fmt.Errorf("Error: todos[%d] missing 'content' or 'status'", i)
			}
			// content/status exist but are not strings.
			return nil, fmt.Errorf("Error: todos[%d] missing 'content' or 'status'", i)
		}

		switch status {
		case string(Pending), string(InProgress), string(Completed):
		default:
			return nil, fmt.Errorf("Error: todos[%d] has invalid status '%s'", i, status)
		}
		result = append(result, Todo{Content: content, Status: TodoStatus(status)})
	}
	return result, nil
}

// RunTodoWrite validates and normalizes the todos input, saves the list, and returns the saved task count.
func RunTodoWrite(todos any) (string, error) {
	normalizedTodos, err := normalizeTodos(todos)
	if err != nil {
		logs.Warn("todo_write parse failed", "err", err)
		return "", err
	}

	todoMu.Lock()
	currentTodos = normalizedTodos
	todoMu.Unlock()

	lines := []string{"\033[33m## Current Tasks\033[0m"}
	icons := map[string]string{
		"pending":     " ",
		"in_progress": "\033[36m▸\033[0m",
		"completed":   "\033[32m✓\033[0m",
	}

	for _, t := range currentTodos {
		icon := icons[string(t.Status)]
		lines = append(lines, fmt.Sprintf("  [%s] %s", icon, t.Content))
	}
	fmt.Printf("%s\n", strings.Join(lines, "\n"))
	logs.Info("todos updated", "count", len(currentTodos))
	return fmt.Sprintf("Updated %d tasks", len(currentTodos)), nil
}

// registerToolTodoWrite registers the todo_write tool with the request.
func registerToolTodoWrite(req *llm.ChatRequest) error {
	return tool.RegisterTool(req, consts.ToolTodoWrite, "Create and manage a task list for your current coding session.", func(in todoWriteInput) (string, error) {
		return RunTodoWrite(in.Todos)
	})
}
