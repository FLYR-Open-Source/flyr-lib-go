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

package middleware // import "github.com/FLYR-Open-Source/flyr-lib-go/monitoring/middleware"

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func newGinRouter(t *testing.T) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(OtelGinMiddleware())
	r.GET("/users/:id", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.POST("/users", func(c *gin.Context) { c.Status(http.StatusCreated) })
	r.GET("/bad", func(c *gin.Context) { c.Status(http.StatusBadRequest) })
	r.GET("/boom", func(c *gin.Context) { c.String(http.StatusInternalServerError, "boom") })
	r.GET("/unavailable", func(c *gin.Context) { c.Status(http.StatusServiceUnavailable) })
	r.GET("/errors", func(c *gin.Context) {
		_ = c.Error(errors.New("first failure"))
		_ = c.Error(errors.New("second failure"))
		c.Status(http.StatusBadRequest)
	})
	return r
}

func TestOtelGinMiddleware_Span(t *testing.T) {
	tests := []struct {
		name            string
		method          string
		target          string
		wantName        string
		wantRoute       string
		wantStatus      int64
		wantCode        codes.Code
		wantDescription string
		wantErrorAttr   string
		wantGinErrors   []string
	}{
		{
			name:       "templated route 200",
			method:     http.MethodGet,
			target:     "/users/42",
			wantName:   "/users/:id",
			wantRoute:  "/users/:id",
			wantStatus: http.StatusOK,
			wantCode:   codes.Unset,
		},
		{
			name:       "post 201",
			method:     http.MethodPost,
			target:     "/users",
			wantName:   "/users",
			wantRoute:  "/users",
			wantStatus: http.StatusCreated,
			wantCode:   codes.Unset,
		},
		{
			name:       "client error does not set error status",
			method:     http.MethodGet,
			target:     "/bad",
			wantName:   "/bad",
			wantRoute:  "/bad",
			wantStatus: http.StatusBadRequest,
			wantCode:   codes.Unset,
		},
		{
			name:            "500 sets error status and attribute",
			method:          http.MethodGet,
			target:          "/boom",
			wantName:        "/boom",
			wantRoute:       "/boom",
			wantStatus:      http.StatusInternalServerError,
			wantCode:        codes.Error,
			wantDescription: "The request has failed with status: Internal Server Error",
			wantErrorAttr:   "500: Internal Server Error",
		},
		{
			name:            "503 sets error status and attribute",
			method:          http.MethodGet,
			target:          "/unavailable",
			wantName:        "/unavailable",
			wantRoute:       "/unavailable",
			wantStatus:      http.StatusServiceUnavailable,
			wantCode:        codes.Error,
			wantDescription: "The request has failed with status: Service Unavailable",
			wantErrorAttr:   "503: Service Unavailable",
		},
		{
			name:          "gin errors are recorded",
			method:        http.MethodGet,
			target:        "/errors",
			wantName:      "/errors",
			wantRoute:     "/errors",
			wantStatus:    http.StatusBadRequest,
			wantCode:      codes.Unset,
			wantGinErrors: []string{"first failure", "second failure"},
		},
		{
			name:       "unknown route",
			method:     http.MethodGet,
			target:     "/nope",
			wantName:   "HTTP GET route not found",
			wantRoute:  "",
			wantStatus: http.StatusNotFound,
			wantCode:   codes.Unset,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := setupTracing(t)
			router := newGinRouter(t)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.target, http.NoBody)
			router.ServeHTTP(rec, req)
			assert.Equal(t, int(tt.wantStatus), rec.Code)

			span := onlySpan(t, recorder)
			assert.Equal(t, tt.wantName, span.Name())
			assert.Equal(t, oteltrace.SpanKindServer, span.SpanKind())
			assert.Equal(t, ScopeName, span.InstrumentationScope().Name)

			attrs := attrMap(span.Attributes())
			assert.Equal(t, tt.method, attrs["http.method"].AsString())
			assert.Equal(t, "http", attrs["http.scheme"].AsString())
			assert.Equal(t, testService, attrs["net.host.name"].AsString())
			assert.Equal(t, "http", attrs["net.protocol.name"].AsString())
			assert.Equal(t, tt.wantRoute, attrs["http.route"].AsString())
			assert.Equal(t, tt.wantStatus, attrs["http.status_code"].AsInt64())

			assert.Equal(t, tt.wantCode, span.Status().Code)
			assert.Equal(t, tt.wantDescription, span.Status().Description)

			errAttr, hasErr := attrs["Error"]
			if tt.wantErrorAttr == "" {
				assert.False(t, hasErr)
			} else {
				require.True(t, hasErr)
				assert.Equal(t, tt.wantErrorAttr, errAttr.AsString())
			}

			ginErrs, hasGinErrs := attrs["gin.errors"]
			if len(tt.wantGinErrors) == 0 {
				assert.False(t, hasGinErrs)
			} else {
				require.True(t, hasGinErrs)
				for _, msg := range tt.wantGinErrors {
					assert.Contains(t, ginErrs.AsString(), msg)
				}
			}
		})
	}
}

func TestOtelGinMiddleware_NoIncomingTraceCreatesRoot(t *testing.T) {
	recorder := setupTracing(t)
	router := newGinRouter(t)

	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/users/1", http.NoBody))

	span := onlySpan(t, recorder)
	assert.False(t, span.Parent().IsValid())
	assert.True(t, span.SpanContext().TraceID().IsValid())
}

func TestOtelGinMiddleware_PropagatesIncomingTraceparent(t *testing.T) {
	recorder := setupTracing(t)
	gin.SetMode(gin.TestMode)

	var (
		handlerSpan oteltrace.SpanContext
		tracerValue any
		hasTracer   bool
	)
	r := gin.New()
	r.Use(OtelGinMiddleware())
	r.GET("/users/:id", func(c *gin.Context) {
		handlerSpan = oteltrace.SpanContextFromContext(c.Request.Context())
		tracerValue, hasTracer = c.Get(tracerKey)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/users/7", http.NoBody)
	req.Header.Set("traceparent", testTraceparent)
	r.ServeHTTP(httptest.NewRecorder(), req)

	span := onlySpan(t, recorder)
	assert.Equal(t, testTraceID, span.SpanContext().TraceID().String())
	assert.Equal(t, testTraceID, span.Parent().TraceID().String())
	assert.Equal(t, testParentSpan, span.Parent().SpanID().String())
	assert.True(t, span.Parent().IsRemote())
	assert.NotEqual(t, testParentSpan, span.SpanContext().SpanID().String())

	// The span is visible to the handler, and the tracer is stored on the gin context.
	assert.Equal(t, span.SpanContext().SpanID(), handlerSpan.SpanID())
	require.True(t, hasTracer)
	assert.Implements(t, (*oteltrace.Tracer)(nil), tracerValue)
}

func TestOtelGinMiddleware_RestoresRequestContext(t *testing.T) {
	setupTracing(t)
	gin.SetMode(gin.TestMode)

	var afterSpan oteltrace.SpanContext
	r := gin.New()
	// Registered before the tracing middleware, so it observes the request after it returns.
	r.Use(func(c *gin.Context) {
		c.Next()
		afterSpan = oteltrace.SpanContextFromContext(c.Request.Context())
	})
	r.Use(OtelGinMiddleware())
	r.GET("/users/:id", func(c *gin.Context) { c.Status(http.StatusOK) })

	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/users/1", http.NoBody))

	assert.False(t, afterSpan.IsValid())
}

func TestOtelGinMiddleware_HostFallsBackToRequestHost(t *testing.T) {
	recorder := setupTracing(t)
	t.Setenv("OTEL_SERVICE_NAME", "")
	router := newGinRouter(t)

	req := httptest.NewRequest(http.MethodGet, "http://example.test:8080/users/1", http.NoBody)
	router.ServeHTTP(httptest.NewRecorder(), req)

	attrs := attrMap(onlySpan(t, recorder).Attributes())
	assert.Equal(t, "example.test", attrs["net.host.name"].AsString())
	assert.Equal(t, int64(8080), attrs["net.host.port"].AsInt64())
}
