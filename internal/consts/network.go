// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package consts

import "time"

const (
	RequestTimeout = time.Second * 90
	RetryDelay     = time.Millisecond * 500
)
const MaxRequestTries = 3

const (
	RequestUnknownError = -1
)
