package compact

import (
	"fmt"
	"go-agent/internal/config"
	"go-agent/internal/consts"
	"go-agent/internal/logs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"unicode/utf8"

	"github.com/anthropics/anthropic-sdk-go"
)

func persistLargeOutput(toolUseID string, output string) string {
	if utf8.RuneCountInString(output) <= consts.PersistThreshold {
		return output
	}
	err := os.MkdirAll(config.SysCfg.ToolResultDir, 0755)
	if err != nil {
		logs.Error("tool result dir mkdir failed", "dir", config.SysCfg.ToolResultDir, "err", err)
	}
	path := filepath.Join(config.SysCfg.ToolResultDir, toolUseID+".txt")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.WriteFile(path, []byte(output), 0644)
		if err != nil {
			logs.Error("persisted output write failed", "path", path, "err", err)
		} else {
			logs.Info("tool output persisted to disk", "path", path, "bytes", len(output))
		}
	}
	preview := output
	if len(preview) > consts.ToolResultPreview {
		preview = preview[:consts.ToolResultPreview]
	}
	return fmt.Sprintf("<persisted-output>\nFull output: %s\nPreview:\n%s\n</persisted-output>", path, preview)

}

// ToolResultBudget 压缩工具较大输出进行落盘处理
func ToolResultBudget(msgs []anthropic.MessageParam) []anthropic.MessageParam {
	var last anthropic.MessageParam
	if len(msgs) != 0 {
		last = msgs[len(msgs)-1]
	} else {
		return msgs
	}
	if last.Role != anthropic.MessageParamRoleUser {
		return msgs
	}
	var blocks []struct {
		Index int
		Block anthropic.ContentBlockParamUnion
	}
	for idx, block := range last.Content {
		if block.OfToolResult != nil {
			blocks = append(blocks, struct {
				Index int
				Block anthropic.ContentBlockParamUnion
			}{Index: idx, Block: block})
		}
	}
	var total = 0
	for _, block := range blocks {
		total += toolResultTextBytes(block.Block.OfToolResult)
	}
	if total <= consts.ToolResultsMaxBytes {
		return msgs
	}
	logs.Info("tool result budget compact triggered", "totalBytes", total, "limit", consts.ToolResultsMaxBytes)
	ranked := slices.Clone(blocks)
	sort.Slice(ranked, func(i, j int) bool {
		return toolResultTextBytes(ranked[i].Block.OfToolResult) >
			toolResultTextBytes(ranked[j].Block.OfToolResult)
	})
	for _, block := range ranked {
		if total <= consts.ToolResultsMaxBytes {
			break
		}
		if toolResultTextBytes(block.Block.OfToolResult) <= consts.PersistThreshold {
			continue
		}
		tid := block.Block.OfToolResult.ToolUseID
		content := toolResultText(block.Block.OfToolResult)
		ref := persistLargeOutput(tid, content)
		block.Block.OfToolResult.Content = []anthropic.ToolResultBlockParamContentUnion{
			{OfText: &anthropic.TextBlockParam{Text: ref}},
		}
		total = 0
		for _, block := range blocks {
			total += toolResultTextBytes(block.Block.OfToolResult)
		}
	}
	return msgs
}
