// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config 封装系统配置与模型配置。
type Config struct {
	System SystemConfig `json:"system"`
	Model  ModelConfig  `json:"model"`
}

// SystemConfig 封装系统相关元数据。
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

// ModelConfig 封装模型相关元数据。
type ModelConfig struct {
	Model     string `json:"model_name"`
	MaxTokens int64  `json:"max_tokens"`
}

var Cfg Config

// LoadConfig load configuration from file
func (cfg *Config) LoadConfig(relativePath string) error {
	raw, err := os.ReadFile(relativePath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(raw, cfg)
	if err != nil {
		return err
	}
	return nil
}

func (cfg *SystemConfig) SetWorkDir(workdir string) {
	cfg.CurDir = workdir
	path := filepath.Join(workdir, ".gocode")
	cfg.SkillsDir = filepath.Join(cfg.CurDir, "skills")
	cfg.ToolResultDir = filepath.Join(path, ".task_outputs", "tool-results")
	cfg.TranscriptDir = filepath.Join(path, ".transcripts")
	cfg.MemoryDir = filepath.Join(path, ".memory")
}
