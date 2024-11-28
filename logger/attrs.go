package logger // import "github.com/FlyrInc/flyr-lib-go/logger"

import (
	"context"
	"log/slog"

	"github.com/FlyrInc/flyr-lib-go/internal/config"
	internalUtils "github.com/FlyrInc/flyr-lib-go/internal/utils"
)

const (
	// The depth of the caller in the stack trace
	callerDepth = 3
)

// Attribute is a wrapper around the log attributes
type Attribute struct {
	err               error
	metadata          []interface{}
	injectAttrsToSpan bool
}

// WithError sets the error in the attribute
func (a *Attribute) WithError(err error) *Attribute {
	a.err = err
	return a
}

// WithMetadata sets the metadata in the attribute
func (a *Attribute) WithMetadata(args ...interface{}) *Attribute {
	a.metadata = args
	return a
}

// WithOutInjectingAttrsToSpan disables injecting attributes to the span
func (a *Attribute) WithOutInjectingAttrsToSpan() *Attribute {
	a.injectAttrsToSpan = false
	return a
}

// Get returns the log attributes
func (a *Attribute) Get(ctx context.Context) []slog.Attr {
	metadata := slog.Group(config.LOG_METADATA_KEY, a.metadata...)
	if a.injectAttrsToSpan {
		injectAttrsToSpan(ctx, metadata)
	}

	caller := internalUtils.GetCallerName(callerDepth)
	callerAttrs := caller.LogAttributes()

	attrs := append(callerAttrs, metadata)

	var errorMessage slog.Attr
	if a.err != nil {
		errorMessage = slog.String(config.LOG_ERROR_KEY, a.err.Error())
		setErroredSpan(ctx, a.err) // Set spans as errored
		attrs = append(attrs, errorMessage)
	}

	return attrs
}

// NewAttribute creates a new attribute
func NewAttribute() *Attribute {
	return &Attribute{injectAttrsToSpan: true}
}
