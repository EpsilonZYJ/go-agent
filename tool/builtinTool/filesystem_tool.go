// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package builtinTool

import (
	"fmt"
	"go-agent/configs"
	"go-agent/services"
	"go-agent/tool"
	"os"
	"path/filepath"
	"strings"
)

type globInput struct {
	Pattern string `json:"pattern" jsonschema:"required" jsonschema_description:"if path is used, use relative path"`
}

type readInput struct {
	Path  string `json:"path" jsonschema:"required" jsonschema_description:"The relative path of the file you want to read."`
	Limit int    `json:"limit" jsonschema:"required" jsonschema_description:"Read limited lines, -1 means no limit"`
}

type writeInput struct {
	Path    string `json:"path" jsonschema:"required" jsonschema_description:"The relative path of the file you want to write."`
	Content string `json:"content" jsonschema:"required" jsonschema_description:"The content you want to write."`
}

type editInput struct {
	Path    string `json:"path" jsonschema:"required" jsonschema_description:"The relative path of the file you want to edit."`
	OldText string `json:"old_text" jsonschema:"required" jsonschema_description:"The origin text needed to be edited."`
	NewText string `json:"new_text" jsonschema:"required" jsonschema_description:"The new text to replace the origin."`
}

func SafePath(p string) (string, error) {
	var path string
	var workdir string

	workdir, err := filepath.Abs(configs.SysCfg.CurDir)
	if err != nil {
		return "", err
	}

	if filepath.IsAbs(p) {
		path = p
	} else {
		path, err = filepath.Abs(filepath.Join(workdir, p))
		if err != nil {
			return "", err
		}
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

func registerToolFileSystem(req *services.ChatRequest) error {
	if err := tool.RegisterTool(req, "read_file", "Read file contents.", func(in readInput) (string, error) {
		return RunRead(in.Path, in.Limit)
	}); err != nil {
		return err
	}
	if err := tool.RegisterTool(req, "write_file", "Write content to a file.", func(in writeInput) (string, error) {
		return RunWrite(in.Path, in.Content)
	}); err != nil {
		return err
	}
	if err := tool.RegisterTool(req, "edit_file", "Replace exact text in a file once.", func(in editInput) (string, error) {
		return RunEdit(in.Path, in.OldText, in.NewText)
	}); err != nil {
		return err
	}
	if err := tool.RegisterTool(req, "glob", "Find files matching a glob pattern.", func(in globInput) (string, error) {
		return RunGlob(in.Pattern)
	}); err != nil {
		return err
	}
	return nil
}

func toSafeRelative(workdir string, match string) (string, bool) {
	resolved := match
	if p, err := filepath.EvalSymlinks(match); err == nil {
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
