package logger

import (
	"github.com/rs/zerolog"

	tracer "github.com/FlyrInc/flyr-lib-go/observability/tracer"
)

type TracingHook struct{}

const (
	DATADOG_TRACE_ID = "dd.trace_id"
	DATADOG_SPAN_ID  = "dd.span_id"
)

func (h TracingHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	traceID, spanID, foundSpan := tracer.ExtractTrace(e.GetCtx())
	if !foundSpan {
		return
	}

	e.Str(DATADOG_TRACE_ID, traceID)
	e.Str(DATADOG_SPAN_ID, spanID)
}
