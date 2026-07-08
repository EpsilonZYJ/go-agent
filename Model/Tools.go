package Model

import (
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
	"github.com/anthropics/anthropic-sdk-go/shared/constant"
)

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"input_schema"`
}

type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

func (tool Tool) ToAnthropicTool() anthropic.ToolParam {
	return anthropic.ToolParam{
		Name:        tool.Name,
		Description: param.NewOpt(tool.Description),
		InputSchema: anthropic.ToolInputSchemaParam{
			Type:       constant.Object(tool.InputSchema.Type),
			Required:   tool.InputSchema.Required,
			Properties: tool.InputSchema.Properties,
		},
	}
}
