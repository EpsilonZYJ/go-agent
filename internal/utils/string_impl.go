package utils

func StringTruncateRunes(s string, maxlen int) string {
	return string([]rune(s)[:min(len([]rune(s)), maxlen)])
}
