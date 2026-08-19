package utils

import "time"

func NowTime() string {
	return time.Now().Format("20060102150405")
}
