package tracer

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type Span struct {
	span oteltrace.Span
}

type SpanKind = oteltrace.SpanKind
type Code = codes.Code
type KeyValue = attribute.KeyValue

const (
	// SpanKindUnspecified is an unspecified SpanKind and is not a valid
	// SpanKind. SpanKindUnspecified should be replaced with SpanKindInternal
	// if it is received.
	SpanKindUnspecified SpanKind = 0
	// SpanKindInternal is a SpanKind for a Span that represents an internal
	// operation within an application.
	SpanKindInternal SpanKind = 1
	// SpanKindServer is a SpanKind for a Span that represents the operation
	// of handling a request from a client.
	SpanKindServer SpanKind = 2
	// SpanKindClient is a SpanKind for a Span that represents the operation
	// of client making a request to a server.
	SpanKindClient SpanKind = 3
	// SpanKindProducer is a SpanKind for a Span that represents the operation
	// of a producer sending a message to a message broker. Unlike
	// SpanKindClient and SpanKindServer, there is often no direct
	// relationship between this kind of Span and a SpanKindConsumer kind. A
	// SpanKindProducer Span will end once the message is accepted by the
	// message broker which might not overlap with the processing of that
	// message.
	SpanKindProducer SpanKind = 4
	// SpanKindConsumer is a SpanKind for a Span that represents the operation
	// of a consumer receiving a message from a message broker. Like
	// SpanKindProducer Spans, there is often no direct relationship between
	// this Span and the Span that produced the message.
	SpanKindConsumer SpanKind = 5
)

const (
	// Unset is the default status code.
	Unset Code = 0

	// Error indicates the operation contains an error.
	//
	// NOTE: The error code in OTLP is 2.
	// The value of this enum is only relevant to the internals
	// of the Go SDK.
	Error Code = 1

	// Ok indicates operation has been validated by an Application developers
	// or Operator to have completed successfully, or contain no error.
	//
	// NOTE: The Ok code in OTLP is 1.
	// The value of this enum is only relevant to the internals
	// of the Go SDK.
	Ok Code = 2

	maxCode = 3
)

func StartSpanFromContext(ctx context.Context, name string, kind SpanKind) (context.Context, *Span) {
	if defaultTracer == nil {
		return ctx, nil
	}

	return defaultTracer.StartSpanFromContext(ctx, name, kind)
}

func StartSpan(name string, kind SpanKind) (context.Context, *Span) {
	if defaultTracer == nil {
		return context.TODO(), nil
	}

	return defaultTracer.StartSpan(name, kind)
}

func (s *Span) SetAttributes(kv ...KeyValue) {
	if s.span.IsRecording() {
		s.span.SetAttributes(kv...)
	}
}

func (s *Span) SetError(err error) {
	if s.span.IsRecording() {
		s.span.RecordError(err)
	}
}

func (s *Span) SetStatus(code Code, message string) {
	if s.span.IsRecording() {
		s.span.SetStatus(code, message)
	}
}

func (s *Span) EndWithError(err error, setSpanErrored bool) {
	if err != nil {
		if setSpanErrored {
			s.SetStatus(Error, err.Error())
		}
		s.SetError(err)
		return
	}
	s.span.End()
}

func (s *Span) End() {
	s.span.SetStatus(Ok, "")
	s.span.End()
}
