package Services

import (
	"fmt"
	"go-agent/Model"

	"github.com/anthropics/anthropic-sdk-go"
)

type ChatRequest struct {
	Model        string                     `json:"model"`
	SystemPrompt []anthropic.TextBlockParam `json:"system_prompt,omitempty"`
	Messages     []anthropic.MessageParam   `json:"message"`
	MaxTokens    int64                      `json:"maxTokens"`
	Tools        []anthropic.ToolUnionParam `json:"tools"`
}

func NewChatRequest(model string, maxTokens int64, systemPrompt string) *ChatRequest {
	return &ChatRequest{
		Model:        model,
		SystemPrompt: NewSystemBlocks(systemPrompt),
		Messages:     []anthropic.MessageParam{},
		MaxTokens:    maxTokens,
		Tools:        []anthropic.ToolUnionParam{},
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

func (req *ChatRequest) AddTool(tool anthropic.ToolUnionParam) {
	req.Tools = append(req.Tools, tool)
}

func (req *ChatRequest) AddTools(tools []anthropic.ToolUnionParam) {
	req.Tools = append(req.Tools, tools...)
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
