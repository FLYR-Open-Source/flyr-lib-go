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

package monitoring // import "github.com/FLYR-Open-Source/flyr-lib-go/pkg/testhelpers/monitoring"

import (
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type FakeStatus struct {
	Code        otelcodes.Code
	Description string
}

type FakeRecordedError struct {
	Error   error
	Options []oteltrace.EventOption
}

type FakeSpan struct {
	oteltrace.Span

	FakeSpanContext *oteltrace.SpanContext

	FakeEvents        []oteltrace.EventOption
	FakeLink          oteltrace.Link
	FakeAttributes    []attribute.KeyValue
	FakeStatus        FakeStatus
	FakeRecordedError FakeRecordedError
}

func (fs *FakeSpan) AddEvent(name string, options ...oteltrace.EventOption) {
	fs.FakeEvents = append(fs.FakeEvents, options...)
	fs.Span.AddEvent(name, options...)
}

func (fs *FakeSpan) AddLink(link oteltrace.Link) {
	fs.FakeLink = link
	fs.Span.AddLink(link)
}

func (fs *FakeSpan) RecordError(err error, options ...oteltrace.EventOption) {
	fs.FakeRecordedError = FakeRecordedError{
		Error:   err,
		Options: options,
	}
	fs.Span.RecordError(err, options...)
}

func (fs *FakeSpan) SpanContext() oteltrace.SpanContext {
	if fs.FakeSpanContext != nil {
		return *fs.FakeSpanContext
	}
	return fs.Span.SpanContext()
}

func (fs *FakeSpan) SetStatus(code otelcodes.Code, description string) {
	fs.FakeStatus = FakeStatus{
		Code:        code,
		Description: description,
	}
	fs.Span.SetStatus(code, description)
}

func (fs *FakeSpan) SetAttributes(kv ...attribute.KeyValue) {
	fs.FakeAttributes = append(fs.FakeAttributes, kv...)
	fs.Span.SetAttributes(kv...)
}

func (fs *FakeSpan) End(options ...oteltrace.SpanEndOption) {
	fs.Span.End(options...)
}

func (fs *FakeSpan) IsRecording() bool {
	return fs.Span.IsRecording()
}
