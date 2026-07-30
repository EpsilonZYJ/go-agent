// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package compact

import (
	"go-agent/internal/consts"

	"github.com/anthropics/anthropic-sdk-go"
)

type toolResult struct {
	MsgIdx int
	BlkIdx int
	Block  anthropic.ContentBlockParamUnion
}

func collectToolResults(msgs []anthropic.MessageParam) []toolResult {
	var blocks []toolResult
	for mi, msg := range msgs {
		if msg.Role != anthropic.MessageParamRoleUser {
			continue
		}
		for bi, block := range msg.Content {
			if block.OfToolResult != nil {
				blocks = append(blocks, toolResult{BlkIdx: bi, MsgIdx: mi, Block: block})
			}
		}
	}
	return blocks
}

// MicroCompact 旧工具结果占位替换，只保留最近3条tool_result都完整内容，更旧的替换为一行占位符
func MicroCompact(msgs []anthropic.MessageParam) []anthropic.MessageParam {
	toolResults := collectToolResults(msgs)
	if len(toolResults) <= consts.KeepRecent {
		return msgs
	}
	for _, r := range toolResults[:len(toolResults)-consts.KeepRecent] {
		tr := r.Block.OfToolResult
		if toolResultTextLen(tr) > consts.ToolResultMaxLen {
			tr.Content = []anthropic.ToolResultBlockParamContentUnion{
				{
					OfText: &anthropic.TextBlockParam{
						Text: "[Earlier tool result compacted. Re-run if needed.]",
					},
				},
			}
		}
	}
	return msgs
}
