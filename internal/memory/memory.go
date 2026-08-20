// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package memory

import "regexp"

// MemType classifies what a memory is about.
type MemType string

// Memory is a single memory entry parsed from a markdown file in the memory dir.
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

// memoryIndexFilename is the human-readable index of all memories.
const memoryIndexFilename = "MEMORY.md"

// memSelectionRe matches the first JSON array (non-greedy) in an LLM response.
var memSelectionRe = regexp.MustCompile(`(?s)\[.*?\]`)

// memExtractionRe matches a JSON array (greedy, to the last ']') in an LLM response.
var memExtractionRe = regexp.MustCompile(`(?s)\[.*\]`)
