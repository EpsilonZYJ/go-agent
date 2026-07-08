package main

import (
	"bufio"
	"context"
	"fmt"
	"go-agent/Model"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

import (
	Const "go-agent/Const"
)

type SystemConfig struct {
	Url          string `json:"url"`
	ApiKey       string `json:"api_key"`
	SystemPrompt string `json:"system_prompt"`
	CurDir       string `json:"cur_dir"`
}

type ModelConfig struct {
	Model     string `json:"model"`
	MaxTokens int64  `json:"maxTokens"`
}

var SysCfg SystemConfig
var ModelCfg ModelConfig

var Client anthropic.Client

func InitAgent() error {
	var err error
	ModelCfg.Model = os.Getenv("MODEL")
	ModelCfg.MaxTokens = 1024
	SysCfg.Url = os.Getenv("URL")
	SysCfg.ApiKey = os.Getenv("API_KEY")
	SysCfg.CurDir, err = os.Getwd()
	if err != nil {
		return fmt.Errorf("Get current directory failed: %v", err)
	}
	SysCfg.SystemPrompt = fmt.Sprintf("You are a coding agent at %s. Use bash to solve tasks. Act, don't explain.", SysCfg.CurDir)
	if ModelCfg.Model == "" || SysCfg.Url == "" || SysCfg.ApiKey == "" {
		return fmt.Errorf("environment variables not set")
	}

	Client = anthropic.NewClient(
		option.WithBaseURL(SysCfg.Url),
		option.WithAPIKey(SysCfg.ApiKey),
	)
	return nil
}

func AgentLoop(request *Model.ChatRequest) {
	for {
		resp, err := Client.Messages.New(
			context.TODO(),
			anthropic.MessageNewParams{
				MaxTokens: request.MaxTokens,
				Messages:  request.Messages,
				Model:     request.Model,
				System:    request.SystemPrompt,
			},
		)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

	}
}

func main() {
	err := InitAgent()
	if err != nil {
		fmt.Println(err)
		os.Exit(Const.ExitEnvError)
	}
	scanner := bufio.NewScanner(os.Stdin)
	req := Model.NewChatRequest(ModelCfg.Model, ModelCfg.MaxTokens, SysCfg.SystemPrompt)

	fmt.Println("Welcome to Go Agent! Type `/exit` to quit.")
	for {
		fmt.Printf("\033[36mUser >> \033[0m")
		scanner.Scan()
		query := scanner.Text()
		if query == "/exit" {
			fmt.Println("Bye!")
			break
		}
		req.AddUserContent(query)
		AgentLoop(req)
	}
}
