// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package utils

import (
	"strings"

	"go.yaml.in/yaml/v4"
)

func ParseFrontmatter(text string) (map[string]string, string) {
	if !strings.HasPrefix(text, "---") {
		return map[string]string{}, text
	}
	parts := strings.SplitN(text, "---", 3)
	if len(parts) < 3 {
		return map[string]string{}, text
	}
	var metadata map[string]string
	if err := yaml.Unmarshal([]byte(parts[1]), &metadata); err != nil {
		metadata = map[string]string{}
	}
	return metadata, strings.TrimSpace(parts[2])
}
