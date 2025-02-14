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

package utils

import (
	"encoding/json"
	"testing"

	"log/slog"

	"github.com/FlyrInc/flyr-lib-go/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
)

func TestSplitFunctionName(t *testing.T) {
	tests := []struct {
		input        string
		expectedNs   string
		expectedFunc string
	}{
		{"main.main", "main", "main"},
		{"main", "", "main"},
	}

	for _, tt := range tests {
		ns, functionName := splitFunctionName(tt.input)
		assert.Equal(t, tt.expectedNs, ns)
		assert.Equal(t, tt.expectedFunc, functionName)
	}
}

func TestCallerString(t *testing.T) {
	caller := Caller{
		FilePath:     "/path/to/file.go",
		LineNumber:   42,
		FunctionName: "TestFunction",
		Namespace:    "example",
	}

	expected := "/path/to/file.go:42 (TestFunction)"
	assert.Equal(t, expected, caller.String())
}

func TestCallerMarshalJSON(t *testing.T) {
	caller := Caller{
		FilePath:     "/path/to/file.go",
		LineNumber:   42,
		FunctionName: "TestFunction",
		Namespace:    "example",
	}

	expectedJSON := map[string]interface{}{
		config.FILE_PATH:             "/path/to/file.go",
		config.LINE_NUMBER:           float64(42),
		config.FUNCTION_NAME:         "TestFunction",
		config.FUNCTION_PACKAGE_NAME: "example",
	}

	data, err := caller.MarshalJSON()
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, expectedJSON, result)
}

func TestCallerLogAttributes(t *testing.T) {
	caller := Caller{
		FilePath:     "/path/to/file.go",
		LineNumber:   42,
		FunctionName: "TestFunction",
		Namespace:    "example",
	}

	expected := []slog.Attr{
		slog.String(config.FILE_PATH, "/path/to/file.go"),
		slog.Int(config.LINE_NUMBER, 42),
		slog.String(config.FUNCTION_NAME, "TestFunction"),
		slog.String(config.FUNCTION_PACKAGE_NAME, "example"),
	}

	logAttrs := caller.LogAttributes()

	assert.Equal(t, expected, logAttrs)
}

func TestCallerSpanAttributes(t *testing.T) {
	caller := Caller{
		FilePath:     "/path/to/file.go",
		LineNumber:   42,
		FunctionName: "TestFunction",
		Namespace:    "example",
	}

	expected := []attribute.KeyValue{
		attribute.String(config.FILE_PATH, "/path/to/file.go"),
		attribute.Int(config.LINE_NUMBER, 42),
		attribute.String(config.FUNCTION_NAME, "TestFunction"),
		attribute.String(config.FUNCTION_PACKAGE_NAME, "example"),
	}

	spanAttrs := caller.SpanAttributes()

	assert.Equal(t, expected, spanAttrs)
}
