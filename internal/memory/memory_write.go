// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package memory

import (
	"fmt"
	"go-agent/internal/config"
	"os"
	"path/filepath"
	"strings"
)

func writeMemoryFile(name string, memType MemType, description string, body string) (string, error) {
	name = strings.ToLower(name)
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "/", "-")
	filename := name + ".md"
	path := filepath.Join(config.SysCfg.MemoryDir, filename)
	file, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	_, err = file.WriteString(
		fmt.Sprintf("---\nname: %s\ndescription: %s\ntype: %s\n---\n\n%s\n", name, description, memType, body),
	)
	_ = rebuildIndex()
	return path, err
}
