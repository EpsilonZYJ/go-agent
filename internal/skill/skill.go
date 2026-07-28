// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package skill

type Skill struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"content"`
}

var skillRegistry = map[string]Skill{}
