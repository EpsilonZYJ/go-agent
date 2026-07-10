package Errors

import (
	"errors"
	"go-agent/Const"

	"github.com/anthropics/anthropic-sdk-go"
)

func AnthropicRequestErrorCode(err error) int {
	var apiErr *anthropic.Error
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode
	} else {
		return Const.RequestUnknownError
	}
}
