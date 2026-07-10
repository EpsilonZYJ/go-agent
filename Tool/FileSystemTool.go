package Tool

import (
	"fmt"
	"go-agent/Model"
	"go-agent/Services"
	"go-agent/configs"
	"os"
	"path/filepath"
	"strings"
)

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

func RunRead(path string, limit int) string {
	path, err := SafePath(path)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	lines := strings.Split(string(data), "\n")

	if limit >= 0 && len(lines) > limit {
		lines = append(
			lines[:limit],
			fmt.Sprintf("... (%d more lines)", len(lines)-limit),
		)
	}
	return strings.Join(lines, "\n")
}

func RunWrite(path string, content string) string {
	path, err := SafePath(path)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	parent := filepath.Dir(path)
	err = os.MkdirAll(parent, 0755)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	err = os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return fmt.Sprintf("Wrote %d bytes to %s", len(content), path)
}

func RunEdit(path string, oldtxt string, newtxt string) string {
	path, err := SafePath(path)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	text := string(data)
	if !strings.Contains(text, oldtxt) {
		return fmt.Sprintf("Error: text not found in %s", path)
	}
	newContent := strings.Replace(text, oldtxt, newtxt, 1)

	err = os.WriteFile(path, []byte(newContent), 0644)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return fmt.Sprintf("Edited: %s", path)
}

func RunGlob(pattern string) string {
	workdir := configs.SysCfg.CurDir
	matches, err := filepath.Glob(filepath.Join(workdir, pattern))
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	results := make([]string, 0, len(matches))
	for _, match := range matches {
		rel, ok := toSafeRelative(workdir, match)
		if ok {
			results = append(results, rel)
		}
	}
	if len(results) == 0 {
		return "(no matches)"
	}
	return strings.Join(results, "\n")
}

func registerToolFileSystem(req *Services.ChatRequest) {
	req.AddTool("read_file", Model.Tool{
		Name:        "read_file",
		Description: "Read file contents.",
		InputSchema: Model.InputSchema{
			Type: "object",
			Properties: map[string]Model.Property{
				"path": {
					Type:        "string",
					Description: "",
				},
				"limit": {
					Type:        "integer",
					Description: "",
				},
			},
			Required: []string{"path"},
		},
	}.ToAnthropicTool())

	req.AddTool("write_file", Model.Tool{
		Name:        "write_file",
		Description: "Write content to a file.",
		InputSchema: Model.InputSchema{
			Type: "object",
			Properties: map[string]Model.Property{
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

	req.AddTool("edit_file", Model.Tool{
		Name:        "edit_file",
		Description: "Replace exact text in a file once.",
		InputSchema: Model.InputSchema{
			Type: "object",
			Properties: map[string]Model.Property{
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

	req.AddTool("glob", Model.Tool{
		Name:        "glob",
		Description: "Find files matching a glob pattern.",
		InputSchema: Model.InputSchema{
			Type: "object",
			Properties: map[string]Model.Property{
				"pattern": {
					Type:        "string",
					Description: "",
				},
			},
			Required: []string{"pattern"},
		},
	}.ToAnthropicTool())
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
