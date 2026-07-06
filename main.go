package main

import (
	"fmt"
	"os"
)

import (
	Const "go-agent/Const"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

var Model string
var URL string
var API_KEY string
var SYSTEM_PROMPT string
var CUR_DIR string

func InitAgent() error {
	var err error
	Model = os.Getenv("MODEL")
	URL = os.Getenv("URL")
	API_KEY = os.Getenv("API_KEY")
	CUR_DIR, err = os.Getwd()
	if err != nil {
		return fmt.Errorf("Get current directory failed: %v", err)
	}
	SYSTEM_PROMPT = fmt.Sprintf("You are a coding agent at %s. Use bash to solve tasks. Act, don't explain.", CUR_DIR)
	if Model == "" || URL == "" || API_KEY == "" {
		return fmt.Errorf("environment variables not set")
	}
	return nil
}

func main() {
	err := InitAgent()
	if err != nil {
		fmt.Println(err)
		os.Exit(Const.ExitEnvError)
	}
	for {

	}
}
