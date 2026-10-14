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

package logger

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureStdout redirects os.Stdout while fn runs and returns everything written to it.
// NewJSONLogHandler binds os.Stdout at construction time, so handlers must be created inside fn.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	require.NoError(t, err)

	original := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = original })

	fn()

	os.Stdout = original
	require.NoError(t, w.Close())
	out, err := io.ReadAll(r)
	require.NoError(t, err)
	require.NoError(t, r.Close())

	return string(out)
}

func TestNewJSONLogHandler(t *testing.T) {
	t.Run("Writes JSON to stdout with message key and UTC time", func(t *testing.T) {
		before := time.Now().UTC().Add(-time.Second)

		out := captureStdout(t, func() {
			h := NewJSONLogHandler(slog.LevelInfo)
			record := slog.NewRecord(time.Now(), slog.LevelInfo, "hello world", 0)
			record.AddAttrs(slog.String("foo", "bar"))
			require.NoError(t, h.Handle(context.Background(), record))
		})

		lines := strings.Split(strings.TrimSpace(out), "\n")
		require.Len(t, lines, 1)

		var got map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(lines[0]), &got))

		assert.Equal(t, "hello world", got["message"])
		assert.NotContains(t, got, "msg")
		assert.Equal(t, "INFO", got["level"])
		assert.Equal(t, "bar", got["foo"])
		assert.NotContains(t, got, "source")

		ts, ok := got["time"].(string)
		require.True(t, ok)
		parsed, err := time.Parse(time.RFC3339Nano, ts)
		require.NoError(t, err)
		assert.True(t, strings.HasSuffix(ts, "Z"), "time must be UTC, got %s", ts)
		assert.False(t, parsed.Before(before))
	})

	tests := []struct {
		name         string
		handlerLevel slog.Level
		logLevel     slog.Level
		wantEnabled  bool
	}{
		{name: "debug disabled at info level", handlerLevel: slog.LevelInfo, logLevel: slog.LevelDebug, wantEnabled: false},
		{name: "info enabled at info level", handlerLevel: slog.LevelInfo, logLevel: slog.LevelInfo, wantEnabled: true},
		{name: "warn enabled at info level", handlerLevel: slog.LevelInfo, logLevel: slog.LevelWarn, wantEnabled: true},
		{name: "info disabled at error level", handlerLevel: slog.LevelError, logLevel: slog.LevelInfo, wantEnabled: false},
		{name: "debug enabled at debug level", handlerLevel: slog.LevelDebug, logLevel: slog.LevelDebug, wantEnabled: true},
	}

	for _, tt := range tests {
		t.Run("Respects level: "+tt.name, func(t *testing.T) {
			var enabled bool
			captureStdout(t, func() {
				enabled = NewJSONLogHandler(tt.handlerLevel).Enabled(context.Background(), tt.logLevel)
			})

			assert.Equal(t, tt.wantEnabled, enabled)
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected slog.Level
	}{
		{
			name:     "Debug level",
			input:    "debug",
			expected: slog.LevelDebug,
		},
		{
			name:     "Info level",
			input:    "info",
			expected: slog.LevelInfo,
		},
		{
			name:     "Warn level",
			input:    "warn",
			expected: slog.LevelWarn,
		},
		{
			name:     "Error level",
			input:    "error",
			expected: slog.LevelError,
		},
		{
			name:     "Uppercase Debug level",
			input:    "DEBUG",
			expected: slog.LevelDebug,
		},
		{
			name:     "Mixed case Info level",
			input:    "InFo",
			expected: slog.LevelInfo,
		},
		{
			name:     "Invalid level, default to Info",
			input:    "invalid",
			expected: slog.LevelInfo,
		},
		{
			name:     "Empty string, default to Info",
			input:    "",
			expected: slog.LevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseLogLevel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
