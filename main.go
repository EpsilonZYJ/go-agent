// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go-agent/internal/agent"
	"go-agent/internal/config"
	"go-agent/internal/consts"
	"go-agent/internal/hooks"
	"go-agent/internal/llm"
	"go-agent/internal/session"
	"go-agent/internal/skill"
	"go-agent/internal/tool/builtin"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

func InitAgent() error {
	var err error

	// 模型设置
	config.ModelCfg.Model = os.Getenv("MODEL")
	config.ModelCfg.MaxTokens = consts.MaxTokens
	config.SysCfg.Url = os.Getenv("URL")
	config.SysCfg.ApiKey = os.Getenv("API_KEY")
	if config.ModelCfg.Model == "" || config.SysCfg.Url == "" || config.SysCfg.ApiKey == "" {
		return fmt.Errorf("environment variables not set")
	}

	// 当前项目系统设置
	config.SysCfg.CurDir, err = os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory failed: %v", err)
	}
	config.SysCfg.SkillsDir = filepath.Join(config.SysCfg.CurDir, "skills")

	skill.ScanSkills(config.SysCfg.SkillsDir)
	config.SysCfg.SystemPrompt = config.BuildSystemPrompt()
	config.SysCfg.SubSystemPrompt = config.BuildSubSystemPrompt()

	llm.Client = anthropic.NewClient(
		option.WithBaseURL(config.SysCfg.Url),
		option.WithAPIKey(config.SysCfg.ApiKey),
	)
	return nil
}

func main() {
	err := InitAgent()
	if err != nil {
		fmt.Println(err)
		os.Exit(consts.ExitEnvError)
	}
	req := llm.NewChatRequest(config.ModelCfg.Model, config.ModelCfg.MaxTokens, config.SysCfg.SystemPrompt)
	if err := builtin.RegisterBuiltinTools(req); err != nil {
		fmt.Printf("register tools failed: %v\n", err)
		os.Exit(consts.ExitRegisterError)
	}
	hooks.RegisterBuiltinObservers()

	fmt.Println("Welcome to Go Agent! Type `/exit` to quit.")
	for {
		fmt.Printf("\033[36mUser >> \033[0m")

		query, err := session.ReadLine()
		if err != nil {
			fmt.Println(err)
			os.Exit(consts.ExitInputError)
		}
		query = strings.TrimSpace(query)
		if query == "" {
			continue
		} else if query == "/exit" {
			fmt.Println("Bye!")
			break
		}
		hooks.TriggerUserPromptSubmit(query)
		req.AddUserContent(query)
		agent.AgentLoop(req)
	}
}
