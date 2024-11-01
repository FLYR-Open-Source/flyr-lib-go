package tracer

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func PropagateHttpRequest(ctx context.Context, name string, headers map[string][]string) {
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(headers))
}
