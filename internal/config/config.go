// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package config

import "path/filepath"

type SystemConfig struct {
	Url             string `json:"url"`
	ApiKey          string `json:"api_key"`
	SystemPrompt    string `json:"system_prompt"`
	SubSystemPrompt string `json:"sub_system_prompt"`
	CurDir          string `json:"cur_dir"`
	SkillsDir       string `json:"skills_dir"`
	ToolResultDir   string `json:"tool_result_dir"`
	TranscriptDir   string `json:"transcript_dir"`
	MemoryDir       string `json:"memory_dir"`
}

type ModelConfig struct {
	Model     string `json:"model"`
	MaxTokens int64  `json:"maxTokens"`
}

var SysCfg SystemConfig
var ModelCfg ModelConfig

func (cfg *SystemConfig) SetWorkDir(workdir string) {
	cfg.CurDir = workdir
	path := filepath.Join(workdir, ".gocode")
	cfg.SkillsDir = filepath.Join(cfg.CurDir, "skills")
	cfg.ToolResultDir = filepath.Join(path, ".task_outputs", "tool-results")
	cfg.TranscriptDir = filepath.Join(path, ".transcripts")
	cfg.MemoryDir = filepath.Join(path, ".memory")
}
