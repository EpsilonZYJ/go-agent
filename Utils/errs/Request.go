package errs

import (
	"errors"
	"go-agent/common/consts"

	"github.com/anthropics/anthropic-sdk-go"
)

func AnthropicRequestErrorCode(err error) int {
	if apiErr, ok := errors.AsType[*anthropic.Error](err); ok {
		return apiErr.StatusCode
	} else {
		return consts.RequestUnknownError
	}
}
