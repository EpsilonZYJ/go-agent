// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package errs

import (
	"errors"
	"go-agent/internal/consts"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

func AnthropicRequestErrorCode(err error) int {
	if apiErr, ok := errors.AsType[*anthropic.Error](err); ok {
		return apiErr.StatusCode
	} else {
		return consts.RequestUnknownError
	}
}

func IsPromptTooLong(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "prompt_too_long") ||
		strings.Contains(strings.ToLower(err.Error()), "prompt too long") ||
		strings.Contains(strings.ToLower(err.Error()), "too many tokens")
}
