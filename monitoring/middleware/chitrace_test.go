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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func newChiRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(OtelChiMiddleware())
	r.Get("/users/{id}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.Post("/users", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	r.Get("/bad", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
	r.Get("/boom", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})
	r.Get("/unavailable", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	return r
}

func TestOtelChiMiddleware_Span(t *testing.T) {
	tests := []struct {
		name          string
		method        string
		target        string
		wantName      string
		wantRoute     string
		wantStatus    int64
		wantErrorAttr string
	}{
		{
			name:       "templated route 200",
			method:     http.MethodGet,
			target:     "/users/42",
			wantName:   "GET /users/42",
			wantRoute:  "/users/{id}",
			wantStatus: http.StatusOK,
		},
		{
			name:       "post 201",
			method:     http.MethodPost,
			target:     "/users",
			wantName:   "POST /users",
			wantRoute:  "/users",
			wantStatus: http.StatusCreated,
		},
		{
			name:       "client error is not flagged",
			method:     http.MethodGet,
			target:     "/bad",
			wantName:   "GET /bad",
			wantRoute:  "/bad",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:          "500 sets error attribute",
			method:        http.MethodGet,
			target:        "/boom",
			wantName:      "GET /boom",
			wantRoute:     "/boom",
			wantStatus:    http.StatusInternalServerError,
			wantErrorAttr: "500: Internal Server Error",
		},
		{
			name:          "503 sets error attribute",
			method:        http.MethodGet,
			target:        "/unavailable",
			wantName:      "GET /unavailable",
			wantRoute:     "/unavailable",
			wantStatus:    http.StatusServiceUnavailable,
			wantErrorAttr: "503: Service Unavailable",
		},
		{
			name:       "unknown route is 404 with empty route",
			method:     http.MethodGet,
			target:     "/nope",
			wantName:   "GET /nope",
			wantRoute:  "",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := setupTracing(t)
			handler := newChiRouter()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.target, http.NoBody)
			handler.ServeHTTP(rec, req)
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

			errAttr, hasErr := attrs["Error"]
			if tt.wantErrorAttr == "" {
				assert.False(t, hasErr)
			} else {
				require.True(t, hasErr)
				assert.Equal(t, tt.wantErrorAttr, errAttr.AsString())
			}

			// The chi middleware does not set a span status.
			assert.Equal(t, codes.Unset, span.Status().Code)
		})
	}
}

func TestOtelChiMiddleware_NoIncomingTraceCreatesRoot(t *testing.T) {
	recorder := setupTracing(t)
	handler := newChiRouter()

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/users/1", http.NoBody))

	span := onlySpan(t, recorder)
	assert.False(t, span.Parent().IsValid())
	assert.True(t, span.SpanContext().TraceID().IsValid())
}

func TestOtelChiMiddleware_PropagatesIncomingTraceparent(t *testing.T) {
	recorder := setupTracing(t)

	var handlerSpan oteltrace.SpanContext
	r := chi.NewRouter()
	r.Use(OtelChiMiddleware())
	r.Get("/users/{id}", func(w http.ResponseWriter, req *http.Request) {
		handlerSpan = oteltrace.SpanContextFromContext(req.Context())
		w.WriteHeader(http.StatusOK)
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

	// The span is made available to downstream handlers through the request context.
	assert.Equal(t, span.SpanContext().SpanID(), handlerSpan.SpanID())
	assert.Equal(t, span.SpanContext().TraceID(), handlerSpan.TraceID())
}

func TestOtelChiMiddleware_HostFallsBackToRequestHost(t *testing.T) {
	recorder := setupTracing(t)
	t.Setenv("OTEL_SERVICE_NAME", "")
	// Re-parse the config with an empty service name.
	handler := newChiRouter()

	req := httptest.NewRequest(http.MethodGet, "http://example.test:8080/users/1", http.NoBody)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	attrs := attrMap(onlySpan(t, recorder).Attributes())
	assert.Equal(t, "example.test", attrs["net.host.name"].AsString())
	assert.Equal(t, int64(8080), attrs["net.host.port"].AsInt64())
}
