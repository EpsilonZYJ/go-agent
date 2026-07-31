// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package consts

// MaxReactiveTrials Reactive compact最大重试次数
const MaxReactiveTrials = 1

const ContextLimit int = 50000

// KeepRecent 保留的最近的旧工具结果占位
const KeepRecent int = 3

// PersistThreshold 单个工具结果的最大字节数目
const PersistThreshold int = 30000

// ToolResultsMaxBytes 所有工具结果的最大字节数目
const ToolResultsMaxBytes = 200000

// MaxMessages Message最大数目
const MaxMessages int = 50

// ToolResultMaxLen 压缩时保留的工具执行结果的最大长度
const ToolResultMaxLen int = 120

// ToolResultPreview 过大的工具执行结果预览最大长度
const ToolResultPreview = 2000

// SummarizeHistoryMaxChars 历史对话总结历史记录最大字数
const SummarizeHistoryMaxChars = 80000
