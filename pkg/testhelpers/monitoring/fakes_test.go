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

package monitoring

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// restoreGlobals resets the global tracer provider mutated by GetFakeTracer/GetFakeSpan.
func restoreGlobals(t *testing.T) {
	t.Helper()
	prev := otel.GetTracerProvider()
	t.Cleanup(func() { otel.SetTracerProvider(prev) })
}

// newRecordedSpan returns a FakeSpan wrapping a real SDK span, plus the recorder that sees what reaches it.
func newRecordedSpan(t *testing.T) (*FakeSpan, *tracetest.SpanRecorder) {
	t.Helper()
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	t.Cleanup(func() { assert.NoError(t, tp.Shutdown(context.Background())) })

	_, span := tp.Tracer("test").Start(context.Background(), "op")
	return &FakeSpan{Span: span}, recorder
}

func TestFakeSpan_AddEvent(t *testing.T) {
	fs, recorder := newRecordedSpan(t)
	attrs := oteltrace.WithAttributes(attribute.String("k", "v"))

	fs.AddEvent("evt", attrs)
	fs.End()

	assert.Len(t, fs.FakeEvents, 1)

	ended := recorder.Ended()
	require.Len(t, ended, 1)
	events := ended[0].Events()
	require.Len(t, events, 1)
	assert.Equal(t, "evt", events[0].Name)
	assert.Contains(t, events[0].Attributes, attribute.String("k", "v"))
}

func TestFakeSpan_AddLink(t *testing.T) {
	fs, recorder := newRecordedSpan(t)
	linked := oteltrace.NewSpanContext(oteltrace.SpanContextConfig{
		TraceID: oteltrace.TraceID{1},
		SpanID:  oteltrace.SpanID{2},
	})
	link := oteltrace.Link{SpanContext: linked}

	fs.AddLink(link)
	fs.End()

	assert.Equal(t, link, fs.FakeLink)

	ended := recorder.Ended()
	require.Len(t, ended, 1)
	links := ended[0].Links()
	require.Len(t, links, 1)
	assert.Equal(t, linked.TraceID(), links[0].SpanContext.TraceID())
	assert.Equal(t, linked.SpanID(), links[0].SpanContext.SpanID())
}

func TestFakeSpan_RecordError(t *testing.T) {
	fs, recorder := newRecordedSpan(t)
	boom := errors.New("boom")
	opt := oteltrace.WithAttributes(attribute.String("where", "here"))

	fs.RecordError(boom, opt)
	fs.End()

	require.ErrorIs(t, fs.FakeRecordedError.Error, boom)
	assert.Len(t, fs.FakeRecordedError.Options, 1)

	ended := recorder.Ended()
	require.Len(t, ended, 1)
	events := ended[0].Events()
	require.Len(t, events, 1)
	assert.Equal(t, "exception", events[0].Name)
	assert.Contains(t, events[0].Attributes, attribute.String("exception.message", "boom"))
	assert.Contains(t, events[0].Attributes, attribute.String("where", "here"))
}

func TestFakeSpan_SpanContext(t *testing.T) {
	t.Run("delegates to the wrapped span by default", func(t *testing.T) {
		fs, _ := newRecordedSpan(t)

		sc := fs.SpanContext()

		assert.True(t, sc.IsValid())
		assert.Equal(t, fs.Span.SpanContext(), sc)
	})

	t.Run("fake span context takes precedence", func(t *testing.T) {
		fs, _ := newRecordedSpan(t)
		fake := oteltrace.NewSpanContext(oteltrace.SpanContextConfig{
			TraceID: oteltrace.TraceID{9},
			SpanID:  oteltrace.SpanID{8},
		})
		fs.FakeSpanContext = &fake

		assert.Equal(t, fake, fs.SpanContext())
		assert.NotEqual(t, fs.Span.SpanContext(), fs.SpanContext())
	})
}

func TestFakeSpan_SetStatus(t *testing.T) {
	fs, recorder := newRecordedSpan(t)

	fs.SetStatus(otelcodes.Error, "it failed")
	fs.End()

	assert.Equal(t, FakeStatus{Code: otelcodes.Error, Description: "it failed"}, fs.FakeStatus)

	ended := recorder.Ended()
	require.Len(t, ended, 1)
	assert.Equal(t, otelcodes.Error, ended[0].Status().Code)
	assert.Equal(t, "it failed", ended[0].Status().Description)
}

func TestFakeSpan_SetAttributes(t *testing.T) {
	fs, recorder := newRecordedSpan(t)

	fs.SetAttributes(attribute.String("a", "1"))
	fs.SetAttributes(attribute.Int("b", 2), attribute.Bool("c", true))
	fs.End()

	assert.Equal(t, []attribute.KeyValue{
		attribute.String("a", "1"),
		attribute.Int("b", 2),
		attribute.Bool("c", true),
	}, fs.FakeAttributes)

	ended := recorder.Ended()
	require.Len(t, ended, 1)
	assert.Contains(t, ended[0].Attributes(), attribute.String("a", "1"))
	assert.Contains(t, ended[0].Attributes(), attribute.Int("b", 2))
	assert.Contains(t, ended[0].Attributes(), attribute.Bool("c", true))
}

func TestFakeSpan_EndAndIsRecording(t *testing.T) {
	fs, recorder := newRecordedSpan(t)

	assert.True(t, fs.IsRecording())
	assert.Empty(t, recorder.Ended())

	fs.End()

	assert.False(t, fs.IsRecording())
	assert.Len(t, recorder.Ended(), 1)
}

func TestFakeSpan_IsRecordingWithNoopSpan(t *testing.T) {
	_, span := noop.NewTracerProvider().Tracer("noop").Start(context.Background(), "op")
	fs := &FakeSpan{Span: span}

	assert.False(t, fs.IsRecording())
	assert.NotPanics(t, func() {
		fs.AddEvent("e")
		fs.SetAttributes(attribute.String("k", "v"))
		fs.End()
	})
}

func TestFakeTracer_Start(t *testing.T) {
	restoreGlobals(t)
	tp, tracer := GetFakeTracer()
	t.Cleanup(func() { assert.NoError(t, tp.Shutdown(context.Background())) })

	ctx, span := tracer.Start(context.Background(), "my-op")

	fs, ok := span.(*FakeSpan)
	require.True(t, ok, "Start must return a *FakeSpan, got %T", span)
	assert.True(t, span.SpanContext().IsValid())
	assert.True(t, span.IsRecording())

	// The returned context carries the underlying span, so the trace IDs line up.
	assert.Equal(t, span.SpanContext().TraceID(), oteltrace.SpanContextFromContext(ctx).TraceID())

	// The fake span records what is done to it.
	span.SetAttributes(attribute.String("k", "v"))
	span.SetStatus(otelcodes.Ok, "fine")
	assert.Equal(t, []attribute.KeyValue{attribute.String("k", "v")}, fs.FakeAttributes)
	assert.Equal(t, otelcodes.Ok, fs.FakeStatus.Code)

	span.End()
	assert.False(t, span.IsRecording())
}

func TestFakeTracer_StartCreatesChildInSameTrace(t *testing.T) {
	restoreGlobals(t)
	tp, tracer := GetFakeTracer()
	t.Cleanup(func() { assert.NoError(t, tp.Shutdown(context.Background())) })

	parentCtx, parent := tracer.Start(context.Background(), "parent")
	_, child := tracer.Start(parentCtx, "child")

	assert.Equal(t, parent.SpanContext().TraceID(), child.SpanContext().TraceID())
	assert.NotEqual(t, parent.SpanContext().SpanID(), child.SpanContext().SpanID())
}

func TestGetFakeTracer(t *testing.T) {
	restoreGlobals(t)

	tp, tracer := GetFakeTracer()
	t.Cleanup(func() { assert.NoError(t, tp.Shutdown(context.Background())) })

	require.NotNil(t, tp)
	require.NotNil(t, tracer.Tracer)
	// Installed as the global provider and propagator.
	assert.Same(t, tp, otel.GetTracerProvider())
	assert.ElementsMatch(t, []string{"traceparent", "tracestate", "baggage"}, otel.GetTextMapPropagator().Fields())
}

func TestGetFakeSpan(t *testing.T) {
	restoreGlobals(t)

	ctx, fs := GetFakeSpan(context.Background())

	require.NotNil(t, fs.Span)
	assert.True(t, fs.IsRecording())
	assert.True(t, fs.SpanContext().IsValid())
	assert.Equal(t, fs.SpanContext().TraceID(), oteltrace.SpanContextFromContext(ctx).TraceID())

	fs.SetAttributes(attribute.String("k", "v"))
	assert.Equal(t, []attribute.KeyValue{attribute.String("k", "v")}, fs.FakeAttributes)

	fs.End()
	assert.False(t, fs.IsRecording())
}

func TestOverrideContextValue(t *testing.T) {
	_, span := noop.NewTracerProvider().Tracer("noop").Start(context.Background(), "op")
	parent, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	ctx, fs := overrideContextValue(parent, span)

	assert.Equal(t, span, fs.Span)
	stored, ok := ctx.Value(currentSpanKey).(FakeSpan)
	require.True(t, ok, "context must hold the FakeSpan under currentSpanKey")
	assert.Equal(t, fs, stored)
	// Parent context behaviour (deadline) is preserved.
	_, hasDeadline := ctx.Deadline()
	assert.True(t, hasDeadline)
	assert.Nil(t, parent.Value(currentSpanKey))
}
