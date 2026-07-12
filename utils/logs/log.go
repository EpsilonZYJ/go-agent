// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package logs

import (
	"log/slog"
	"os"
)

var logger *slog.Logger

func init() {
	debug := os.Getenv("LOG_LEVEL")
	level := slog.LevelInfo
	if debug == "debug" {
		level = slog.LevelDebug
	}
	logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
}

func Debug(msg string, args ...any) {
	logger.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	logger.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	logger.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	logger.Error(msg, args...)
}
