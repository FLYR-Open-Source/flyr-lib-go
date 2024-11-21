package tracer

import (
	"context"

	internaltracer "github.com/FlyrInc/flyr-lib-go/internal/traceprovider"
)

// ShutdownTracerProvider gracefully shuts down the global TracerProvider.
func Shutdown(ctx context.Context) error {
	return internaltracer.ShutdownTracerProvider(ctx)
}
