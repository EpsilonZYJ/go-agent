// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package utils

import (
	"strings"

	"go.yaml.in/yaml/v4"
)

// ParseFrontmatter extracts YAML frontmatter delimited by leading "---"
// markers from the given text. It returns the parsed metadata as a
// map[string]string along with the remaining body content (with surrounding
// whitespace trimmed). If the text has no frontmatter or the metadata cannot
// be parsed, an empty metadata map and the original text are returned.
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
