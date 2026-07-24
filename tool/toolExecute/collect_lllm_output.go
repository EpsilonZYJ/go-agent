package toolExecute

import (
	"go-agent/common/consts"
	"go-agent/tool/permission"
	"go-agent/utils/logs"
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
			// TODO: 收集权限检查情况，对拒绝的提前进行拒绝
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
		logs.Debug(
			"[AgentLoop] ",
			"block=", blockidx,
			"type=", b.Type,
			"raw=", b.RawJSON(),
			"\n", "",
		)
	}

	return
}
