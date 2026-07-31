// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package compact

import (
	"fmt"
	"slices"

	"go-agent/internal/consts"
	"go-agent/internal/logs"

	"github.com/anthropics/anthropic-sdk-go"
)

// SnipCompact 裁掉无关的旧对话
func SnipCompact(msgs []anthropic.MessageParam) []anthropic.MessageParam {
	if len(msgs) <= consts.MaxMessages {
		return msgs
	}
	var keepHead, keepTail int = 3, consts.MaxMessages - 3
	// 信息保留头部3条和尾部47条
	var headEnd, tailStart int = keepHead, len(msgs) - keepTail
	// 将头部的工具调用和结果完整保留
	if headEnd > 0 && messageHasToolUse(msgs[headEnd-1]) {
		for headEnd < len(msgs) && isToolResultMessage(msgs[headEnd]) {
			headEnd++
		}
	}
	// 保留尾部完整工具调用信息
	if (tailStart > 0 && tailStart < len(msgs)) &&
		isToolResultMessage(msgs[tailStart]) &&
		messageHasToolUse(msgs[tailStart-1]) {
		tailStart--
	}
	if headEnd >= tailStart {
		return msgs
	}
	snipped := tailStart - headEnd
	logs.Info("snip compact applied", "original", len(msgs), "snipped", snipped, "remaining", len(msgs)-snipped+1)
	return slices.Concat(
		msgs[:headEnd],
		[]anthropic.MessageParam{
			anthropic.NewUserMessage(
				anthropic.NewTextBlock(
					fmt.Sprintf("[snipped %d messages]", snipped),
				),
			),
		},
		msgs[tailStart:],
	)

}
