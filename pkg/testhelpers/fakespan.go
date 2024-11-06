package testhelpers

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
