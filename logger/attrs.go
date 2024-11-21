package logger // import "github.com/FlyrInc/flyr-lib-go/logger"

import (
	"context"
	"log/slog"

	"github.com/FlyrInc/flyr-lib-go/config"
	internalUtils "github.com/FlyrInc/flyr-lib-go/internal/utils"
)

const (
	// The depth of the caller in the stack trace
	callerDepth = 3
)

type Attribute struct {
	ctx               context.Context
	err               error
	metadata          []interface{}
	injectAttrsToSpan bool
}

func (a *Attribute) WithError(err error) *Attribute {
	a.err = err
	return a
}

func (a *Attribute) WithMetadata(args ...interface{}) *Attribute {
	a.metadata = args
	return a
}

func (a *Attribute) WithOutInjectingAttrsToSpan() *Attribute {
	a.injectAttrsToSpan = false
	return a
}

func (a *Attribute) Get() []slog.Attr {
	metadata := slog.Group(config.LOG_METADATA_KEY, a.metadata...)
	if a.injectAttrsToSpan {
		injectAttrsToSpan(a.ctx, metadata)
	}

	caller := internalUtils.GetCallerName(callerDepth)
	callerAttrs := caller.LogAttributes()

	attrs := append(callerAttrs, metadata)

	var errorMessage slog.Attr
	if a.err != nil {
		errorMessage = slog.String(config.LOG_ERROR_KEY, a.err.Error())
		setErroredSpan(a.ctx, a.err) // Set spans as errored
		attrs = append(attrs, errorMessage)
	}

	return attrs
}

func NewAttribute(ctx context.Context) *Attribute {
	return &Attribute{ctx: ctx, injectAttrsToSpan: true}
}
