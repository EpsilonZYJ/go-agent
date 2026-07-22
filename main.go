// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"go-agent/common/consts"
	"go-agent/configs"
	"go-agent/model"
	"go-agent/services"
	"go-agent/tool"
	"go-agent/tool/builtinTool"
	"go-agent/utils/errs"
	"go-agent/utils/logs"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

var Client anthropic.Client

func InitAgent() error {
	var err error
	configs.ModelCfg.Model = os.Getenv("MODEL")
	configs.ModelCfg.MaxTokens = consts.MaxTokens
	configs.SysCfg.Url = os.Getenv("URL")
	configs.SysCfg.ApiKey = os.Getenv("API_KEY")
	configs.SysCfg.CurDir, err = os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory failed: %v", err)
	}
	configs.SysCfg.SystemPrompt = fmt.Sprintf("You are a coding agent at %s. Use bash to solve tasks. Act, and explain.", configs.SysCfg.CurDir)
	if configs.ModelCfg.Model == "" || configs.SysCfg.Url == "" || configs.SysCfg.ApiKey == "" {
		return fmt.Errorf("environment variables not set")
	}

	Client = anthropic.NewClient(
		option.WithBaseURL(configs.SysCfg.Url),
		option.WithAPIKey(configs.SysCfg.ApiKey),
	)
	return nil
}

func AgentLoop(request *services.ChatRequest, scanner *bufio.Scanner) {
	var trials int = 0
	for {
		// 创建请求
		ctx, cancel := context.WithTimeout(context.Background(), consts.RequestTimeout)
		resp, err := Client.Messages.New(
			ctx,
			anthropic.MessageNewParams{
				MaxTokens: request.MaxTokens,
				Messages:  request.Messages,
				Model:     request.Model,
				System:    request.SystemPrompt,
				Tools:     request.Tools,
			},
		)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
				logs.Debug("[AgentLoop] Request timeout.")
			}
			errCode := errs.AnthropicRequestErrorCode(err)
			if errCode >= http.StatusBadRequest && errCode < http.StatusInternalServerError && errCode != http.StatusTooManyRequests {
				fmt.Printf("An error occurred: %v\n", err)
				cancel()
				return
			} else if trials >= consts.MaxRequestTries {
				fmt.Printf("Max Request Tries: %d\n", trials)
				cancel()
				return
			}
			trials++
			time.Sleep(time.Duration(trials) * consts.RetryDelay)
			cancel()
			fmt.Printf("Error: %v\n", err)
			continue
		}
		cancel()

		trials = 0
		request.Messages = append(request.Messages, resp.ToParam())

		// 收集输出和工具调用
		var toolUses []anthropic.ContentBlockUnion
		var textOuts []strings.Builder
		for blkidx, b := range resp.Content {
			if b.Type == consts.Text && b.Text != "" {
				var tmp strings.Builder
				tmp.WriteString(b.Text)
				textOuts = append(textOuts, tmp)
			} else if b.Type == consts.ToolUse {
				// TODO: 收集权限检查情况，对拒绝的提前进行拒绝
				// permission.CheckPermission(b)
				toolUses = append(toolUses, b)
			}
			logs.Debug(
				"[AgentLoop] ",
				"block=", blkidx,
				"type=", b.Type,
				"raw=", b.RawJSON(),
				"\n", "",
			)
		}

		PrintAgentOutput(textOuts)
		// 无工具调用，本轮结束
		if resp.StopReason != anthropic.StopReasonToolUse || len(toolUses) == 0 {
			return
		}

		results := make([]anthropic.ContentBlockParamUnion, len(toolUses))
		execResults := make([]model.ToolResult, len(toolUses))
		var toolwg sync.WaitGroup

		// 并发执行
		for i, block := range toolUses {
			toolwg.Add(1)
			go func(idx int, block anthropic.ContentBlockUnion) {
				defer toolwg.Done()
				var result = model.ToolResult{}
				result.Name = fmt.Sprintf("\033[33m>>> %s\033[0m\n", block.Name)
				output, err := tool.Dispatch(block.Name, block.Input)
				if err != nil {
					results[idx] = anthropic.NewToolResultBlock(block.ID, err.Error(), true)
					result.Result = fmt.Sprintf("\033[31mError: %s\033[0m\n", err.Error())
				} else {
					results[idx] = anthropic.NewToolResultBlock(block.ID, output, false)
					lines := strings.Split(output, "\n")
					lines = lines[:min(len(lines), consts.ToolMaxPrintOutputLines)]
					result.Result = fmt.Sprintf("\033[90m%s\033[0m\n", strings.Join(lines, "\n"))
				}
				execResults[i] = result
			}(i, block)

		}
		toolwg.Wait()
		for _, r := range execResults {
			fmt.Printf("%s%s", r.Name, r.Result)
		}
		request.Messages = append(request.Messages, anthropic.NewUserMessage(results...))
	}
}

func PrintAgentOutput(textOuts []strings.Builder) {
	for _, textOut := range textOuts {
		if textOut.Len() > 0 {
			fmt.Println("\033[32mAgent: \n\n \033[0m" + textOut.String())
		}
	}
	fmt.Println()
}

func main() {
	err := InitAgent()
	if err != nil {
		fmt.Println(err)
		os.Exit(consts.ExitEnvError)
	}
	scanner := bufio.NewScanner(os.Stdin)
	req := services.NewChatRequest(configs.ModelCfg.Model, configs.ModelCfg.MaxTokens, configs.SysCfg.SystemPrompt)
	if err := builtinTool.RegisterBuiltinTools(req); err != nil {
		fmt.Printf("register tools failed: %v\n", err)
		os.Exit(consts.ExitRegisterError)
	}

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
		req.AddUserContent(query)
		AgentLoop(req, scanner)
	}
}
