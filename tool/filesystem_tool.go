package tool

import (
	"fmt"
	"go-agent/configs"
	"go-agent/model"
	"go-agent/services"
	"os"
	"path/filepath"
	"strings"
)

type globInput struct {
	Pattern string `json:"pattern"`
}

type readInput struct {
	Path  string `json:"path"`
	Limit int    `json:"limit"`
}

type writeInput struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type editInput struct {
	Path    string `json:"path"`
	OldText string `json:"old_text"`
	NewText string `json:"new_text"`
}

func SafePath(p string) (string, error) {
	workdir, err := filepath.Abs(configs.SysCfg.CurDir)
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(filepath.Join(workdir, p))
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(workdir, path)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes workspace: %s", path)
	}
	return path, nil
}

func RunRead(path string, limit int) (string, error) {
	path, err := SafePath(path)
	if err != nil {
		return "", fmt.Errorf("error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("error: %v", err)
	}

	lines := strings.Split(string(data), "\n")

	if limit >= 0 && len(lines) > limit {
		lines = append(
			lines[:limit],
			fmt.Sprintf("... (%d more lines)", len(lines)-limit),
		)
	}
	return strings.Join(lines, "\n"), nil
}

func RunWrite(path string, content string) (string, error) {
	path, err := SafePath(path)
	if err != nil {
		return "", fmt.Errorf("error: %v", err)
	}

	parent := filepath.Dir(path)
	err = os.MkdirAll(parent, 0755)
	if err != nil {
		return "", fmt.Errorf("error: %v", err)
	}

	err = os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("error: %v", err)
	}
	return fmt.Sprintf("Wrote %d bytes to %s", len(content), path), nil
}

func RunEdit(path string, oldtxt string, newtxt string) (string, error) {
	path, err := SafePath(path)
	if err != nil {
		return "", fmt.Errorf("error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("error: %v", err)
	}

	text := string(data)
	if !strings.Contains(text, oldtxt) {
		return "", fmt.Errorf("error: text not found in %s", path)
	}
	newContent := strings.Replace(text, oldtxt, newtxt, 1)

	err = os.WriteFile(path, []byte(newContent), 0644)
	if err != nil {
		return "", fmt.Errorf("error: %v", err)
	}
	return fmt.Sprintf("Edited: %s", path), nil
}

func RunGlob(pattern string) (string, error) {
	workdir := configs.SysCfg.CurDir
	matches, err := filepath.Glob(filepath.Join(workdir, pattern))
	if err != nil {
		return "", fmt.Errorf("error: %v", err)
	}
	results := make([]string, 0, len(matches))
	for _, match := range matches {
		rel, ok := toSafeRelative(workdir, match)
		if ok {
			results = append(results, rel)
		}
	}
	if len(results) == 0 {
		return "(no matches)", nil
	}
	return strings.Join(results, "\n"), nil
}

func registerToolFileSystem(req *services.ChatRequest) {
	req.AddTool(model.Tool{
		Name:        "read_file",
		Description: "Read file contents.",
		InputSchema: model.InputSchema{
			Type: "object",
			Properties: map[string]model.Property{
				"path": {
					Type:        "string",
					Description: "",
				},
				"limit": {
					Type:        "integer",
					Description: "read limited lines, -1 means no limit",
				},
			},
			Required: []string{"path", "limit"},
		},
	}.ToAnthropicTool())

	req.AddTool(model.Tool{
		Name:        "write_file",
		Description: "Write content to a file.",
		InputSchema: model.InputSchema{
			Type: "object",
			Properties: map[string]model.Property{
				"path": {
					Type:        "string",
					Description: "",
				},
				"content": {
					Type:        "string",
					Description: "",
				},
			},
			Required: []string{"path", "content"},
		},
	}.ToAnthropicTool())

	req.AddTool(model.Tool{
		Name:        "edit_file",
		Description: "Replace exact text in a file once.",
		InputSchema: model.InputSchema{
			Type: "object",
			Properties: map[string]model.Property{
				"path": {
					Type:        "string",
					Description: "",
				},
				"old_text": {
					Type:        "string",
					Description: "",
				},
				"new_text": {
					Type:        "string",
					Description: "",
				},
			},
			Required: []string{"path", "old_text", "new_text"},
		},
	}.ToAnthropicTool())

	req.AddTool(model.Tool{
		Name:        "glob",
		Description: "Find files matching a glob pattern.",
		InputSchema: model.InputSchema{
			Type: "object",
			Properties: map[string]model.Property{
				"pattern": {
					Type:        "string",
					Description: "",
				},
			},
			Required: []string{"pattern"},
		},
	}.ToAnthropicTool())
	RegisterExecutor("glob", Wrap(func(in globInput) (string, error) {
		return RunGlob(in.Pattern)
	}))
	RegisterExecutor("read_file", Wrap(func(in readInput) (string, error) {
		return RunRead(in.Path, in.Limit)
	}))
	RegisterExecutor("write_file", Wrap(func(in writeInput) (string, error) {
		return RunWrite(in.Path, in.Content)
	}))
	RegisterExecutor("edit_file", Wrap(func(in editInput) (string, error) {
		return RunEdit(in.Path, in.OldText, in.NewText)
	}))
}

func toSafeRelative(workdir string, match string) (string, bool) {
	resolved := match
	if p, err := filepath.EvalSymlinks(workdir); err == nil {
		resolved = p
	}

	absWork, err := filepath.Abs(workdir)
	if err != nil {
		return "", false
	}
	absMatch, err := filepath.Abs(resolved)
	if err != nil {
		return "", false
	}

	rel, err := filepath.Rel(absWork, absMatch)
	if err != nil {
		return "", false
	}

	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return filepath.ToSlash(rel), true
}
