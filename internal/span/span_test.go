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

package span

import (
	"context"
	"errors"
	"testing"

	testhelpers "github.com/FLYR-Open-Source/flyr-lib-go/pkg/testhelpers/monitoring"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"
)

func TestEndWithError(t *testing.T) {
	t.Run("With Error", func(t *testing.T) {
		_, fakeSpan := testhelpers.GetFakeSpan(context.Background())
		defer fakeSpan.End()
		err := errors.New("test error")

		span := Span{Span: &fakeSpan}
		span.EndWithError(err)

		require.ErrorIs(t, err, fakeSpan.FakeRecordedError.Error)
		assert.Equal(t, codes.Error, fakeSpan.FakeStatus.Code)
		assert.Equal(t, fakeSpan.FakeStatus.Description, err.Error())
	})

	t.Run("Without Error", func(t *testing.T) {
		_, fakeSpan := testhelpers.GetFakeSpan(context.Background())
		defer fakeSpan.End()

		span := Span{Span: &fakeSpan}
		span.EndWithError(nil)

		require.NoError(t, fakeSpan.FakeRecordedError.Error)
		assert.Equal(t, codes.Unset, fakeSpan.FakeStatus.Code)
		assert.Equal(t, "", fakeSpan.FakeStatus.Description)
	})
}

func TestEndSuccessfully(t *testing.T) {
	_, fakeSpan := testhelpers.GetFakeSpan(context.Background())
	defer fakeSpan.End()

	span := Span{Span: &fakeSpan}
	span.EndSuccessfully()

	require.NoError(t, fakeSpan.FakeRecordedError.Error)
	assert.Equal(t, codes.Ok, fakeSpan.FakeStatus.Code)
	assert.Equal(t, "", fakeSpan.FakeStatus.Description)
}
