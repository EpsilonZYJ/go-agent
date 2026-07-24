package files

import (
	"fmt"
	"go-agent/internal/config"
	"path/filepath"
	"strings"
)

func SafePath(p string) (string, error) {
	var path string
	var workdir string

	workdir, err := filepath.Abs(config.SysCfg.CurDir)
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

func PathIsSafe(p string) bool {
	var path string
	workdir, err := filepath.Abs(config.SysCfg.CurDir)
	if err != nil {
		return false
	}

	if filepath.IsAbs(p) {
		path = p
	} else {
		path, err = filepath.Abs(filepath.Join(workdir, p))
		if err != nil {
			return false
		}
	}
	rel, err := filepath.Rel(workdir, path)
	if err != nil {
		return false
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}
	return true
}
