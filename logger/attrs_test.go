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
	"errors"

	"testing"

	"github.com/FLYR-Open-Source/flyr-lib-go/internal/config"
	"github.com/stretchr/testify/assert"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func TestGetAttributes(t *testing.T) {
	args := []interface{}{"key1", "value1", "key2", "value2"}

	// if that test fails, it means the depth of the caller is different,
	// therefore the caller information is not being retrieved correctly
	t.Run("With correct code details", func(t *testing.T) {
		attrs := NewAttribute().
			WithMetadata(args...).
			Get(context.Background())

		assert.GreaterOrEqual(t, len(attrs), 4)

		codePath := attrs[0]
		assert.Equal(t, string(semconv.CodeFilepathKey), codePath.Key)
		assert.Contains(t, codePath.Value.String(), "src/testing/testing.go")

		codeLine := attrs[1]
		assert.Equal(t, string(semconv.CodeLineNumberKey), codeLine.Key)
		assert.Positive(t, codeLine.Value.Int64())

		codeFunc := attrs[2]
		assert.Equal(t, string(semconv.CodeFunctionKey), codeFunc.Key)
		assert.Contains(t, codeFunc.Value.String(), "tRunner")

		codeNs := attrs[3]
		assert.Equal(t, string(semconv.CodeNamespaceKey), codeNs.Key)
		assert.Contains(t, codeNs.Value.String(), "testing")
	})

	t.Run("With metadata", func(t *testing.T) {
		attrs := NewAttribute().
			WithMetadata(args...).
			Get(context.Background())

		assert.Len(t, attrs, 5)

		metadata := attrs[4]
		assert.Equal(t, config.LOG_METADATA_KEY, metadata.Key)
		assert.Equal(t, "[key1=value1 key2=value2]", metadata.Value.String())
	})

	t.Run("With an error", func(t *testing.T) {
		err := errors.New("test error")
		attrs := NewAttribute().
			WithMetadata(args...).
			WithError(err).
			Get(context.Background())

		assert.Len(t, attrs, 6)

		errorMessage := attrs[5]
		assert.Equal(t, config.LOG_ERROR_KEY, errorMessage.Key)
		assert.Equal(t, err.Error(), errorMessage.Value.String())
	})

	t.Run("Without extra metadata", func(t *testing.T) {
		attrs := NewAttribute().Get(context.Background())

		assert.Len(t, attrs, 5)

		metadata := attrs[4]
		assert.Equal(t, config.LOG_METADATA_KEY, metadata.Key)
		assert.Equal(t, "[]", metadata.Value.String())
	})
}
