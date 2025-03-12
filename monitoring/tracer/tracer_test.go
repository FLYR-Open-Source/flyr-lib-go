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

package tracer

import (
	"context"
	"testing"

	testhelpers "github.com/FLYR-Open-Source/flyr-lib-go/pkg/testhelpers/monitoring"
	"github.com/stretchr/testify/assert"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// if that test fails, it means the depth of the caller is different,
// therefore the caller information is not being retrieved correctly
func TestStartSpan(t *testing.T) {
	pc, fakeTracer := testhelpers.GetFakeTracer()
	//nolint:errcheck
	defer pc.Shutdown(context.Background())

	tracer := Tracer{tracer: fakeTracer}

	_, span := tracer.StartSpan(context.Background(), "test-span", oteltrace.SpanKindInternal)
	defer span.End()

	testSpan := span.Span.(*testhelpers.FakeSpan)
	assert.Equal(t, string(semconv.CodeFilepathKey), string(testSpan.FakeAttributes[0].Key))
	assert.Contains(t, testSpan.FakeAttributes[0].Value.AsString(), "src/testing/testing.go")

	assert.Equal(t, string(semconv.CodeLineNumberKey), string(testSpan.FakeAttributes[1].Key))
	assert.Positive(t, testSpan.FakeAttributes[1].Value.AsInt64())

	assert.Equal(t, string(semconv.CodeFunctionKey), string(testSpan.FakeAttributes[2].Key))
	assert.Contains(t, testSpan.FakeAttributes[2].Value.AsString(), "tRunner")

	assert.Equal(t, string(semconv.CodeNamespaceKey), string(testSpan.FakeAttributes[3].Key))
	assert.Contains(t, testSpan.FakeAttributes[3].Value.AsString(), "testing")
}
