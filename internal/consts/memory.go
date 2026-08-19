// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package consts

// MaxRecentMessagesForRelevantSelect 在进行相关性选择时，最多保留的最近消息条数。超过此条数的消息将被忽略，以提高相关性选择的效率。
const MaxRecentMessagesForRelevantSelect = 3

// MaxRelevantMemoriesToSelect 相关性记忆时，最多选择的记忆条数。超过此条数的记忆将被忽略，以确保检索结果的简洁性和相关性。
const MaxRelevantMemoriesToSelect = 5

// MemoryConsolidateThreshold 记忆整合阈值，当记忆条数超过此阈值时，触发记忆整合操作，以优化记忆存储和检索效率。
const MemoryConsolidateThreshold = 10
