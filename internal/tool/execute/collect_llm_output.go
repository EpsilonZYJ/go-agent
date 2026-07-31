// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package execute

import (
	"go-agent/internal/consts"
	"go-agent/internal/hooks"
	"go-agent/internal/logs"
	"go-agent/internal/tool/permission"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

func CollectLLMOutput(respContent []anthropic.ContentBlockUnion) (
	textOuts []strings.Builder,
	toolUseList []anthropic.ContentBlockUnion,
	allowIndex []int,
	denyIndex []int,
	askIndex []int,
	errIndex []int,
	denyErrMap map[int]string,
	errErrMap map[int]string,
	askReasonMap map[int]string,
) {
	toolUseList = []anthropic.ContentBlockUnion{}
	allowIndex = []int{}
	denyIndex = []int{}
	askIndex = []int{}
	errIndex = []int{}
	textOuts = []strings.Builder{}
	denyErrMap = map[int]string{}
	errErrMap = map[int]string{}
	askReasonMap = map[int]string{}
	var index int = 0

	for blockidx, b := range respContent {
		if b.Type == consts.Text && b.Text != "" {
			var tmp strings.Builder
			tmp.WriteString(b.Text)
			textOuts = append(textOuts, tmp)
		} else if b.Type == consts.ToolUse {
			// 通知型 PreToolUse hook（如日志），权限决策仍由 permission 子系统独立完成
			hooks.TriggerPreToolUse(b)
			checkPermission, err := permission.CheckPermission(b)
			toolUseList = append(toolUseList, b)
			switch checkPermission {
			case consts.PermissionAllow:
				allowIndex = append(allowIndex, index)
			case consts.PermissionDeny:
				denyIndex = append(denyIndex, index)
				denyErrMap[index] = err.Error()
			case consts.PermissionAskUser:
				askIndex = append(askIndex, index)
				askReasonMap[index] = err.Error()
			default:
				errIndex = append(errIndex, index)
				errErrMap[index] = err.Error()
			}
			index++
		}
		logs.Debug("block processed",
			"block", blockidx,
			"type", b.Type,
			"raw", b.RawJSON(),
		)
	}

	return
}
