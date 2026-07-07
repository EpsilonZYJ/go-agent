package main

import (
	"bufio"
	"fmt"
	"go-agent/Model"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

import (
	Const "go-agent/Const"
)

type Config struct {
	Model        string `json:"model"`
	Url          string `json:"url"`
	ApiKey       string `json:"api_key"`
	SystemPrompt string `json:"system_prompt"`
	CurDir       string `json:"cur_dir"`
}

var Cfg Config

var Client anthropic.Client

func InitAgent() error {
	var err error
	Cfg.Model = os.Getenv("MODEL")
	Cfg.Url = os.Getenv("URL")
	Cfg.ApiKey = os.Getenv("API_KEY")
	Cfg.CurDir, err = os.Getwd()
	if err != nil {
		return fmt.Errorf("Get current directory failed: %v", err)
	}
	Cfg.SystemPrompt = fmt.Sprintf("You are a coding agent at %s. Use bash to solve tasks. Act, don't explain.", Cfg.CurDir)
	if Cfg.Model == "" || Cfg.Url == "" || Cfg.ApiKey == "" {
		return fmt.Errorf("environment variables not set")
	}

	Client = anthropic.NewClient(
		option.WithBaseURL(Cfg.Url),
		option.WithAPIKey(Cfg.ApiKey),
	)
	return nil
}

func AgentLoop(messages []Model.Message) {
	for {

	}
}

func main() {
	err := InitAgent()
	if err != nil {
		fmt.Println(err)
		os.Exit(Const.ExitEnvError)
	}
	scanner := bufio.NewScanner(os.Stdin)
	var history []Model.Message

	fmt.Println("Welcome to Go Agent! Type `/exit` to quit.`")
	for {
		fmt.Printf("\033[36mYou >> \033[0m")
		scanner.Scan()
		query := scanner.Text()
		if query == "/exit" {
			fmt.Println("Bye!")
			break
		}
		history = append(history, Model.Message{
			Role:    "user",
			Content: query,
		})
		AgentLoop(history)
	}
}
