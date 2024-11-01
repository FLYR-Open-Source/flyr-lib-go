package tracer

import (
	"context"

	oteltrace "go.opentelemetry.io/otel/trace"
)

type Tracer struct {
	tracer oteltrace.Tracer
}

var defaultTracer *Tracer

func (t *Tracer) StartSpan(name string, kind SpanKind) (context.Context, *Span) {
	ctx := context.Background()

	if t.tracer == nil {
		return ctx, nil
	}

	ctx, span := t.tracer.Start(ctx, name, oteltrace.WithSpanKind(kind))
	return ctx, &Span{span: span}
}

func (t *Tracer) StartSpanFromContext(ctx context.Context, name string, kind SpanKind) (context.Context, *Span) {
	if t.tracer == nil {
		return ctx, nil
	}

	ctx, span := t.tracer.Start(ctx, name, oteltrace.WithSpanKind(kind))
	return ctx, &Span{span: span}
}
