// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package main

import (
	"bufio"
	"fmt"
	"go-agent/internal/agent"
	"go-agent/internal/config"
	"go-agent/internal/consts"
	"go-agent/internal/hooks"
	"go-agent/internal/llm"
	"go-agent/internal/tool/builtin"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

func InitAgent() error {
	var err error
	config.ModelCfg.Model = os.Getenv("MODEL")
	config.ModelCfg.MaxTokens = consts.MaxTokens
	config.SysCfg.Url = os.Getenv("URL")
	config.SysCfg.ApiKey = os.Getenv("API_KEY")
	config.SysCfg.CurDir, err = os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory failed: %v", err)
	}
	config.SysCfg.SystemPrompt = fmt.Sprintf("You are a coding agent at %s. Before starting any multi-step task, use todo_write to plan your steps. Update status as you go.", config.SysCfg.CurDir)
	if config.ModelCfg.Model == "" || config.SysCfg.Url == "" || config.SysCfg.ApiKey == "" {
		return fmt.Errorf("environment variables not set")
	}

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
	scanner := bufio.NewScanner(os.Stdin)
	req := llm.NewChatRequest(config.ModelCfg.Model, config.ModelCfg.MaxTokens, config.SysCfg.SystemPrompt)
	if err := builtin.RegisterBuiltinTools(req); err != nil {
		fmt.Printf("register tools failed: %v\n", err)
		os.Exit(consts.ExitRegisterError)
	}
	hooks.RegisterBuiltinObservers()

	fmt.Println("Welcome to Go Agent! Type `/exit` to quit.")
	for {
		fmt.Printf("\033[36mUser >> \033[0m")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Println(err)
			}
			os.Exit(consts.ExitInputError)
		}
		query := strings.TrimSpace(scanner.Text())
		if query == "" {
			continue
		} else if query == "/exit" {
			fmt.Println("Bye!")
			break
		}
		hooks.TriggerUserPromptSubmit(query)
		req.AddUserContent(query)
		agent.AgentLoop(req, scanner)
	}
}
