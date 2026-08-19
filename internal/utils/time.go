// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package utils

import "time"

func NowTime() string {
	return time.Now().Format("20060102150405")
}
