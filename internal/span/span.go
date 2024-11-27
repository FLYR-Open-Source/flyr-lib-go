package span // import "github.com/FlyrInc/flyr-lib-go/internal/span"

import (
	otelcodes "go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// Span is a wrapper around oteltrace.Span
//
// It is used to provide a more convenient API with extra functionality
type Span struct {
	oteltrace.Span
}

// EndWithError ends the span by updating the status to Error and recording the error
func (s *Span) EndWithError(err error) {
	if err != nil {
		s.SetStatus(otelcodes.Error, err.Error())
		s.RecordError(err)
	}
	s.End()
}

// EndSuccessfully ends the span by updating the status to Ok
func (s *Span) EndSuccessfully() {
	s.SetStatus(otelcodes.Ok, "")
	s.End()
}
