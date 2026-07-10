package Services

import (
	"fmt"
	"go-agent/Model"
	"go-agent/Utils/logs"
	"sync"

	"github.com/anthropics/anthropic-sdk-go"
)

type ChatRequest struct {
	Model        string                     `json:"model"`
	SystemPrompt []anthropic.TextBlockParam `json:"system_prompt,omitempty"`
	Messages     []anthropic.MessageParam   `json:"message"`
	MaxTokens    int64                      `json:"maxTokens"`
	Tools        []anthropic.ToolUnionParam `json:"tool_list,omitempty"`
	toolMap      map[string]anthropic.ToolUnionParam
	mu           sync.RWMutex
}

func (req *ChatRequest) GetTool(name string) anthropic.ToolUnionParam {
	req.mu.RLock()
	defer req.mu.RUnlock()
	return req.toolMap[name]
}

func NewChatRequest(model string, maxTokens int64, systemPrompt string) *ChatRequest {
	return &ChatRequest{
		Model:        model,
		SystemPrompt: NewSystemBlocks(systemPrompt),
		Messages:     []anthropic.MessageParam{},
		MaxTokens:    maxTokens,
		Tools:        []anthropic.ToolUnionParam{},
		toolMap:      map[string]anthropic.ToolUnionParam{},
	}
}

func NewSystemBlocks(prompt string) []anthropic.TextBlockParam {
	return []anthropic.TextBlockParam{{Text: prompt}}
}

func (req *ChatRequest) AddMessages(message []Model.Message) error {
	if req.Messages == nil {
		req.Messages = []anthropic.MessageParam{}
	}
	for _, msg := range message {
		switch msg.Role {
		case "user":
			req.Messages = append(req.Messages,
				anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content)),
			)
		case "assistant":
			req.Messages = append(req.Messages,
				anthropic.NewAssistantMessage(anthropic.NewTextBlock(msg.Content)),
			)
		default:
			return fmt.Errorf("unknown role %s", msg.Role)
		}
	}
	return nil
}

func (req *ChatRequest) AddTool(name string, tool anthropic.ToolUnionParam) {
	req.mu.Lock()
	defer req.mu.Unlock()
	req.toolMap[name] = tool
	req.Tools = append(req.Tools, tool)
}

func (req *ChatRequest) AddTools(names []string, tools []anthropic.ToolUnionParam) {
	if len(names) != len(tools) {
		logs.Error("The number of tools does not match the number of tool names")
	}
	req.mu.Lock()
	defer req.mu.Unlock()
	for idx, name := range names {
		req.toolMap[name] = tools[idx]
		req.Tools = append(req.Tools, tools[idx])
	}
}

func (req *ChatRequest) AddUserContent(userContent string) {
	if req.Messages == nil {
		req.Messages = []anthropic.MessageParam{}
	}
	req.Messages = append(req.Messages,
		anthropic.NewUserMessage(anthropic.NewTextBlock(userContent)),
	)
}

func (req *ChatRequest) AddAssistantContent(asContent string) {
	if req.Messages == nil {
		req.Messages = []anthropic.MessageParam{}
	}
	req.Messages = append(req.Messages,
		anthropic.NewAssistantMessage(anthropic.NewTextBlock(asContent)),
	)
}
