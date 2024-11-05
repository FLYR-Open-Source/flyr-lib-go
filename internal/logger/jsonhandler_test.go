package logger

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
			result := parseLogLevel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
