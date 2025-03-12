// MIT License
//
// Copyright (c) 2025 FLYR, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package logger // import "github.com/FLYR-Open-Source/flyr-lib-go/internal/logger"

import (
	"log/slog"
	"os"
	"strings"
)

// NewJSONLogHandler creates a new JSON log handler with custom configurations.
//
// This function initializes a slog.Handler that outputs logs in JSON format
// to standard output. The handler is setting the log level according to the provided configuration,
// and replacing certain attributes using the replaceAttributes function for
// custom formatting.
func NewJSONLogHandler(level slog.Level) slog.Handler {
	return slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{
			AddSource:   false,
			Level:       level,
			ReplaceAttr: replaceAttributes,
		})
}

// parseLogLevel converts a string to a slog.Level.
//
// If the string is not a valid log level, it defaults to info.
func ParseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
