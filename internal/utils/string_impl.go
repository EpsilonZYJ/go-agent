// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package utils

func StringTruncateRunes(s string, maxlen int) string {
	return string([]rune(s)[:min(len([]rune(s)), maxlen)])
}
