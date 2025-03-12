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

package logger // import "github.com/FLYR-Open-Source/flyr-lib-go/logger"

import (
	"context"
	"log/slog"

	"github.com/FLYR-Open-Source/flyr-lib-go/internal/config"
	internalUtils "github.com/FLYR-Open-Source/flyr-lib-go/internal/utils"
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
