// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package compact

import (
	"encoding/json"
	"fmt"
	"go-agent/internal/config"
	"go-agent/internal/consts"
	"go-agent/internal/llm"
	"go-agent/internal/logs"
	"go-agent/internal/utils"
	"os"
	"path/filepath"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

func writeTranscript(msgs []anthropic.MessageParam) string {
	if err := os.MkdirAll(config.Cfg.System.TranscriptDir, 0755); err != nil {
		logs.Error("llm summary dir mkdir failed", "dir", config.Cfg.System.TranscriptDir, "error", err)
	}
	path := filepath.Join(config.Cfg.System.TranscriptDir, fmt.Sprintf("transcript_%s.jsonl", utils.NowTime()))
	var b strings.Builder
	for _, msg := range msgs {
		line, _ := json.Marshal(msg)
		b.Write(line)
		b.WriteByte('\n')
	}
	err := os.WriteFile(path, []byte(b.String()), 0644)
	if err != nil {
		logs.Error("write transcript failed", "error", err)
	}
	return path
}

func summarizeHistory(msgs []anthropic.MessageParam) string {
	conversation, err := json.Marshal(msgs)
	if err != nil {
		logs.Error("json marshal failed", "error", err)
	}
	content := string([]rune(string(conversation))[:min(consts.SummarizeHistoryMaxChars, len([]rune(string(conversation))))])
	prompt := "Summarize this coding-agent conversation so work can continue.\n" +
		"Output EXACTLY these sections, in this order:\n" +
		"CURRENT GOAL: <quote the user's MOST RECENT explicit instruction VERBATIM, one line>\n" +
		"KEY FINDINGS / DECISIONS: ...\n" +
		"FILES READ / CHANGED: ...\n" +
		"REMAINING WORK: ...\n" +
		"USER CONSTRAINTS: ...\n" +
		"Be compact but concrete. NEVER invent a new goal; CURRENT GOAL must be the user's actual last request, not your inference.\n\n" + content

	resp, rerr := llm.Call(
		anthropic.MessageNewParams{
			Messages: []anthropic.MessageParam{
				anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
			},
			Model:     config.Cfg.Model.Model,
			MaxTokens: config.Cfg.Model.MaxTokens,
		},
		0, // 单次调用，不重试
	)
	if rerr != nil {
		logs.Error("summarize request failed", "kind", rerr.Kind, "err", rerr.Err)
		return "(summarize failed, empty summary)"
	}

	rawSummary := resp.Content
	var summary string = ""
	for _, block := range rawSummary {
		if block.Type == consts.Text {
			summary = summary + block.Text
		}
	}
	summary = strings.TrimSpace(summary)
	if summary == "" {
		return "(empty summary)"
	}
	return summary
}

// CompactHistory LLM 进行总结压缩。为避免目标漂移，最近 CompactKeepTail 条消息
func CompactHistory(msgs []anthropic.MessageParam) []anthropic.MessageParam {
	transcriptPath := writeTranscript(msgs)
	logs.Info("transcript saved:", "path", transcriptPath)
	fmt.Printf("[transcript saved: %s]\n", transcriptPath)

	// 历史太短就不切尾部，直接整体总结（保持原有行为）
	if len(msgs) <= consts.CompactKeepTail {
		summary := summarizeHistory(msgs)
		return []anthropic.MessageParam{
			anthropic.NewUserMessage(
				anthropic.NewTextBlock(fmt.Sprintf("[Compacted]\n\n%s", summary)),
			),
		}
	}

	// 最近 CompactKeepTail 条原样保留；对齐到不拆分 tool_use/tool_result 对
	tailStart := len(msgs) - consts.CompactKeepTail
	if isToolResultMessage(msgs[tailStart]) && messageHasToolUse(msgs[tailStart-1]) {
		tailStart--
	}
	summary := summarizeHistory(msgs[:tailStart])
	head := []anthropic.MessageParam{
		anthropic.NewUserMessage(
			anthropic.NewTextBlock(fmt.Sprintf("[Compacted]\n\n%s", summary)),
		),
	}
	result := append(head, msgs[tailStart:]...)
	if idx, instr := LastUserInstruction(msgs); idx >= 0 && idx < tailStart {
		result = append(result, anthropic.NewUserMessage(
			anthropic.NewTextBlock(fmt.Sprintf(""+
				"[system-note] The above history has been compressed. The user’s latest and current sole objective is: \n%s\n"+
				"Please continue completing it; if it is unrelated to the topic before compression, please set aside the old task and directly address it.",
				instr,
			)),
		))
	}
	return result
}

// ReactiveCompact API 出错时进行压缩
func ReactiveCompact(msgs []anthropic.MessageParam) []anthropic.MessageParam {
	transcriptPath := writeTranscript(msgs)
	logs.Info("transcript saved:", "path", transcriptPath)
	fmt.Printf("[transcript saved: %s]\n", transcriptPath)
	summary := summarizeHistory(msgs)
	tailStart := max(0, len(msgs)-consts.CompactKeepTail)
	if (tailStart > 0 && tailStart < len(msgs)) &&
		isToolResultMessage(msgs[tailStart]) &&
		messageHasToolUse(msgs[tailStart-1]) {
		tailStart--
	}
	head := []anthropic.MessageParam{
		anthropic.NewUserMessage(
			anthropic.NewTextBlock(
				fmt.Sprintf("[Reactive compact]\n\n%s", summary),
			),
		),
	}
	result := append(head, msgs[tailStart:]...)
	if idx, instr := LastUserInstruction(msgs); idx >= 0 && idx < tailStart {
		result = append(result, anthropic.NewUserMessage(
			anthropic.NewTextBlock(fmt.Sprintf(""+
				"[system-note] The above history has been compressed. The user’s latest and current sole objective is: \n%s\n"+
				"Please continue completing it; if it is unrelated to the topic before compression, please set aside the old task and directly address it.",
				instr,
			)),
		))
	}
	return result
}
