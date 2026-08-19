package consts

// MaxRecentMessagesForRelevantSelect 在进行相关性选择时，最多保留的最近消息条数。超过此条数的消息将被忽略，以提高相关性选择的效率。
const MaxRecentMessagesForRelevantSelect = 3

// MaxRelevantMemoriesToSelect 相关性记忆时，最多选择的记忆条数。超过此条数的记忆将被忽略，以确保检索结果的简洁性和相关性。
const MaxRelevantMemoriesToSelect = 5
