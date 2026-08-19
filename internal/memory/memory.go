// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package memory

import "regexp"

type MemType string

type Memory struct {
	Filename    string  `json:"filename"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Type        MemType `json:"type"`
	Body        string  `json:"body"`
}

type extractedMemory struct {
	Name        string  `json:"name"`
	Type        MemType `json:"type"`
	Description string  `json:"description"`
	Body        string  `json:"body"`
}

const (
	MemTypeUser      MemType = "user"
	MemTypeFeedback  MemType = "feedback"
	MemTypeProject   MemType = "project"
	MemTypeReference MemType = "reference"
)

const memoryIndexFilename = "MEMORY.md"

var memSelectionRe = regexp.MustCompile(`(?s)\[.*?\]`)
var memExtractionRe = regexp.MustCompile(`(?s)\[.*\]`)
