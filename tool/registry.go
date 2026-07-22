package tool

import (
	"encoding/json"
	"go-agent/services"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
	"github.com/invopop/jsonschema"
)

func RegisterTool[T any](req *services.ChatRequest, name, description string, fn func(T) (string, error)) error {
	var zero T
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties:  false,
		RequiredFromJSONSchemaTags: true,
		DoNotReference:             true,
	}
	schemaJSON, err := json.Marshal(reflector.Reflect(zero))
	if err != nil {
		return err
	}
	var inputSchema anthropic.ToolInputSchemaParam
	if err := inputSchema.UnmarshalJSON(schemaJSON); err != nil {
		return err
	}
	req.AddTool(anthropic.ToolUnionParam{
		OfTool: &anthropic.ToolParam{
			Name:        name,
			Description: param.NewOpt(description),
			InputSchema: inputSchema,
		},
	})
	RegisterExecutor(name, Wrap(fn))
	return nil
}
