package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"go-agent/Services"
	"go-agent/Tool"
	"os"
	"strings"
	"sync"

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

func AgentLoop(request *Services.ChatRequest) {
	for {
		// 创建请求
		resp, err := Client.Messages.New(
			context.TODO(),
			anthropic.MessageNewParams{
				MaxTokens: request.MaxTokens,
				Messages:  request.Messages,
				Model:     request.Model,
				System:    request.SystemPrompt,
				Tools:     request.Tools,
			},
		)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		request.Messages = append(request.Messages, resp.ToParam())

		// 收集输出和工具调用
		var textOut strings.Builder
		var toolUses []anthropic.ContentBlockUnion
		for _, b := range resp.Content {
			if b.Type == Const.Text && b.Text != "" {
				textOut.WriteString(b.Text)
			} else if b.Type == Const.ToolUse {
				toolUses = append(toolUses, b)
			}
		}
		if textOut.Len() > 0 {
			fmt.Println("\033[32mAgent: \n\n \033[0m" + textOut.String())
		}

		// 无工具调用，本轮结束
		if resp.StopReason != anthropic.StopReasonToolUse || len(toolUses) == 0 {
			return
		}

		results := make([]anthropic.ContentBlockParamUnion, len(toolUses))
		var toolwg sync.WaitGroup

		// 并发执行
		for i, block := range toolUses {
			toolwg.Add(1)
			go func(idx int, block anthropic.ContentBlockUnion) {
				defer toolwg.Done()
				switch block.Name {
				case "bash":
					var args Tool.Command
					if err := json.Unmarshal(block.Input, &args); err != nil {
						results[i] = anthropic.NewToolResultBlock(block.ID, "invalid tool input: "+err.Error(), true)
						return
					}
					output, err := Tool.RunBash(args.Command)
					if err != nil {
						results[i] = anthropic.NewToolResultBlock(block.ID, err.Error(), true)
						return
					}
					if strings.TrimSpace(output) == "" {
						output = ""
					}
					results[i] = anthropic.NewToolResultBlock(block.ID, output, false)
				default:
					results[i] = anthropic.NewToolResultBlock(block.ID, "unknown tool: "+block.Name, false)
				}

			}(i, block)

			toolwg.Wait()
			request.Messages = append(request.Messages, anthropic.NewUserMessage(results...))
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
	req := Services.NewChatRequest(ModelCfg.Model, ModelCfg.MaxTokens, SysCfg.SystemPrompt)

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
