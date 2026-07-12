// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

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
