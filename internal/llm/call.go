// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package llm

import (
	"context"
	"errors"
	"net/http"
	"time"

	"go-agent/internal/consts"
	"go-agent/internal/errs"
	"go-agent/internal/logs"

	"github.com/anthropics/anthropic-sdk-go"
)

// RequestErrorKind 请求错误分类，调用方按类别决定处置策略。
type RequestErrorKind int

const (
	// TimeoutErr 请求超时（重试后仍超时）
	TimeoutErr RequestErrorKind = iota
	// PromptTooLongErr prompt 过长；调用方可压缩后重试
	PromptTooLongErr
	// NonRetryableErr 4xx 不可重试错误
	NonRetryableErr
	// RetryExhaustedErr 瞬时错误重试 maxTries 次后仍失败
	RetryExhaustedErr
)

// RequestError 分类后的请求错误。
type RequestError struct {
	Kind RequestErrorKind
	Err  error
}

func (e *RequestError) Error() string { return e.Err.Error() }

// Call 发起一次 LLM 请求：对瞬时错误（5xx/429/网络/超时）按 maxTries 退避重试；
// prompt_too_long 与不可重试 4xx 立即分类返回，不重试。
// 成功返回 *anthropic.Message；失败返回分类后的 *RequestError，由调用方决定策略。
func Call(params anthropic.MessageNewParams, maxTries int) (*anthropic.Message, *RequestError) {
	var lastErr error
	lastKind := RetryExhaustedErr
	for try := 0; try <= maxTries; try++ {
		ctx, cancel := context.WithTimeout(context.Background(), consts.RequestTimeout)
		resp, err := Client.Messages.New(ctx, params)
		cancel()
		if err == nil {
			return resp, nil
		}
		lastErr = err

		// 两类立即分类返回、不重试
		if errs.IsPromptTooLong(err) {
			return nil, &RequestError{Kind: PromptTooLongErr, Err: err}
		}
		if code := errs.AnthropicRequestErrorCode(err); code >= http.StatusBadRequest && code < http.StatusInternalServerError && code != http.StatusTooManyRequests {
			return nil, &RequestError{Kind: NonRetryableErr, Err: err}
		}

		// 其余视为瞬时错误，退避后重试
		if errors.Is(err, context.DeadlineExceeded) {
			lastKind = TimeoutErr
		} else {
			lastKind = RetryExhaustedErr
		}
		if try < maxTries {
			logs.Warn("llm request failed, retrying", "try", try+1, "maxTries", maxTries, "err", err)
			time.Sleep(time.Duration(try+1) * consts.RetryDelay)
		}
	}
	return nil, &RequestError{Kind: lastKind, Err: lastErr}
}
