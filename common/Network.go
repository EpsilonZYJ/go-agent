package common

import "time"

const (
	RequestTimeout = time.Second * 90
	RetryDelay     = time.Millisecond * 500
)
const MaxRequestTries = 3

const (
	RequestUnknownError = -1
)
