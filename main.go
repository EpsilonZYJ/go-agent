// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package main

import (
	"fmt"
	"os"
	"strings"

	"go-agent/internal/agent"
	"go-agent/internal/config"
	"go-agent/internal/consts"
	"go-agent/internal/hooks"
	"go-agent/internal/llm"
	"go-agent/internal/logs"
	"go-agent/internal/session"
	"go-agent/internal/skill"
	"go-agent/internal/tool/builtin"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

func InitAgent() error {
	var err error
	if err = config.Cfg.LoadConfig("gocode.json"); err != nil {
		logs.Info("config file not found, using environment variables")
		// 模型设置
		config.Cfg.Model.Model = os.Getenv("MODEL")
		config.Cfg.Model.MaxTokens = consts.MaxTokens
		config.Cfg.System.Url = os.Getenv("URL")
		config.Cfg.System.ApiKey = os.Getenv("API_KEY")
		if config.Cfg.Model.Model == "" || config.Cfg.System.Url == "" || config.Cfg.System.ApiKey == "" {
			return fmt.Errorf("environment variables not set")
		}
	}
	if config.Cfg.Model.Model == "" {
		return fmt.Errorf("model name not set")
	} else if config.Cfg.Model.MaxTokens == 0 {
		return fmt.Errorf("max tokens not set")
	} else if config.Cfg.System.ApiKey == "" {
		return fmt.Errorf("api key not set")
	} else if config.Cfg.System.Url == "" {
		return fmt.Errorf("url not set")
	}
	logs.Info("agent initializing",
		"model", config.Cfg.Model.Model,
		"url", config.Cfg.System.Url,
		"max_tokens", config.Cfg.Model.MaxTokens,
	)

	// 当前项目系统设置
	curDir, err := os.Getwd()
	if err != nil {
		logs.Warn("get working directory failed", "err", err)
		return fmt.Errorf("get current directory failed: %v", err)
	}

	config.Cfg.System.SetWorkDir(curDir)

	logs.Info("working directory", "dir", config.Cfg.System.CurDir)

	skill.ScanSkills(config.Cfg.System.SkillsDir)
	config.Cfg.System.SystemPrompt = config.BuildSystemPrompt()
	config.Cfg.System.SubSystemPrompt = config.BuildSubSystemPrompt()

	llm.Client = anthropic.NewClient(
		option.WithBaseURL(config.Cfg.System.Url),
		option.WithAPIKey(config.Cfg.System.ApiKey),
	)
	logs.Info("agent initialized successfully")
	return nil
}

func main() {
	err := InitAgent()
	if err != nil {
		logs.Error("init agent failed", "err", err)
		fmt.Println(err)
		os.Exit(consts.ExitEnvError)
	}
	req := llm.NewChatRequest(config.Cfg.Model.Model, config.Cfg.Model.MaxTokens, config.Cfg.System.SystemPrompt)
	if err := builtin.RegisterBuiltinTools(req); err != nil {
		logs.Error("register builtin tools failed", "err", err)
		fmt.Printf("register tools failed: %v\n", err)
		os.Exit(consts.ExitRegisterError)
	}
	hooks.RegisterBuiltinObservers()

	fmt.Println("Welcome to Go Agent! Type `/exit` to quit.")
	for {
		fmt.Printf("\033[36mUser >> \033[0m")

		query, err := session.ReadLine()
		if err != nil {
			logs.Error("read user input failed", "err", err)
			fmt.Println(err)
			os.Exit(consts.ExitInputError)
		}
		query = strings.TrimSpace(query)
		if query == "" {
			continue
		} else if query == "/exit" {
			logs.Info("user requested exit")
			fmt.Println("Bye!")
			break
		}
		hooks.TriggerUserPromptSubmit(query)
		req.AddUserContent(query)
		agent.AgentLoop(req)
	}
}
