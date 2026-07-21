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

package http // import "github.com/FLYR-Open-Source/flyr-lib-go/monitoring/http"

import (
	"context"
	"crypto/tls"
	"net/http/httptrace"
	"sync"
	"time"

	"github.com/FLYR-Open-Source/flyr-lib-go/internal/version"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// meterName identifies the Meter used for HTTP client connection-lifecycle metrics.
const meterName = "github.com/FLYR-Open-Source/flyr-lib-go/monitoring/http"

// durationBucketBoundaries are the explicit bucket boundaries (in seconds) used for all
// duration histograms in this package. Heavily skewed towards sub-second granularity, since
// DNS/connect/TLS/get-connection phases are typically much faster than a full request.
var durationBucketBoundaries = []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}

// clientMetrics holds the instruments used to record HTTP client connection-lifecycle
// metrics: where time is spent acquiring a connection (DNS resolution, TCP connect, TLS
// handshake, waiting for/reusing a pooled connection) and time-to-first-byte. These are
// not covered by otelhttp.NewTransport's own metrics (http.client.request.duration and
// http.client.request.body.size), which only measure the request as a whole.
//
// The instruments are created from a metric.MeterProvider (see newClientMetrics), which by
// default is the globally registered go.opentelemetry.io/otel MeterProvider
// (otel.GetMeterProvider()) — matching how otelhttp.NewTransport itself sources its meter
// when no MeterProvider is explicitly configured. Passing WithMeterProvider to
// SetHttpTransport/NewHttpClient makes both otelhttp's own metrics and these use the same
// provider.
type clientMetrics struct {
	getConnDuration   metric.Float64Histogram
	getConnUnacquired metric.Int64Counter
	dnsDuration       metric.Float64Histogram
	connectDuration   metric.Float64Histogram
	tlsDuration       metric.Float64Histogram
	ttfbDuration      metric.Float64Histogram
	putIdleConnErrors metric.Int64Counter

	// getConnAttrs holds the four possible attribute combinations for the
	// http.client.getconn.duration histogram, indexed as [reused][wasIdle], precomputed once
	// at instrument creation. GotConn fires on every request, and building the attribute set
	// there with metric.WithAttributes would allocate per request; metric.WithAttributeSet
	// over a precomputed set does not (it is the documented choice for hot paths).
	getConnAttrs [2][2]metric.MeasurementOption
}

// boolIndex converts a bool to an index into the getConnAttrs table.
func boolIndex(b bool) int {
	if b {
		return 1
	}
	return 0
}

// newClientMetrics creates the instruments for HTTP client connection-lifecycle metrics
// using the given MeterProvider.
//
// It returns an error if any of the instruments could not be created.
func newClientMetrics(provider metric.MeterProvider) (*clientMetrics, error) {
	mt := provider.Meter(meterName, metric.WithInstrumentationVersion(version.Version()))

	var err error
	m := &clientMetrics{}

	if m.getConnDuration, err = mt.Float64Histogram("http.client.getconn.duration",
		metric.WithDescription("Time spent acquiring a connection from the pool (GetConn to GotConn)"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(durationBucketBoundaries...),
	); err != nil {
		return nil, err
	}
	if m.getConnUnacquired, err = mt.Int64Counter("http.client.getconn.unacquired",
		metric.WithDescription("Number of requests where a connection was requested (GetConn) but never acquired (GotConn never fired) before the request ended. Signals the client is stuck waiting for a connection, e.g. blocking pool exhaustion or a request cancellation/timeout during acquisition"),
		metric.WithUnit("{request}"),
	); err != nil {
		return nil, err
	}
	if m.dnsDuration, err = mt.Float64Histogram("http.client.dns.duration",
		metric.WithDescription("Time spent on DNS resolution for new connections"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(durationBucketBoundaries...),
	); err != nil {
		return nil, err
	}
	if m.connectDuration, err = mt.Float64Histogram("http.client.connect.duration",
		metric.WithDescription("Time spent establishing the TCP connection for new connections"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(durationBucketBoundaries...),
	); err != nil {
		return nil, err
	}
	if m.tlsDuration, err = mt.Float64Histogram("http.client.tls.duration",
		metric.WithDescription("Time spent on the TLS handshake for new connections"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(durationBucketBoundaries...),
	); err != nil {
		return nil, err
	}
	if m.ttfbDuration, err = mt.Float64Histogram("http.client.ttfb.duration",
		metric.WithDescription("Time from obtaining a connection to receiving the first response byte"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(durationBucketBoundaries...),
	); err != nil {
		return nil, err
	}
	if m.putIdleConnErrors, err = mt.Int64Counter("http.client.putidleconn.errors",
		metric.WithDescription("Number of connections that could not be returned to the idle pool after use. A rising rate indicates connection pool churn"),
		metric.WithUnit("{error}"),
	); err != nil {
		return nil, err
	}

	for _, reused := range []bool{false, true} {
		for _, wasIdle := range []bool{false, true} {
			m.getConnAttrs[boolIndex(reused)][boolIndex(wasIdle)] = metric.WithAttributeSet(attribute.NewSet(
				attribute.Bool("reused", reused),
				attribute.Bool("was_idle", wasIdle),
			))
		}
	}

	return m, nil
}

// recordUnacquired records a connection acquisition that started (GetConn) but never
// completed (GotConn) before the request ended.
func (m *clientMetrics) recordUnacquired(ctx context.Context) {
	m.getConnUnacquired.Add(ctx, 1)
}

// getConnState tracks the timings for a single connection acquisition attempt, and is reset
// on each GetConn so that a retried request reusing the same trace context doesn't leak a
// previous attempt's timestamps into the new attempt's readings. wrapRoundTripper (see
// http.go) uses it after the request finishes to detect a connection that was requested but
// never acquired (see acquired below) and record the http.client.getconn.unacquired metric.
type getConnState struct {
	mu sync.Mutex
	getConnStateData
}

// getConnStateData holds the fields that reset() clears. Kept separate from mu so reset can
// overwrite the data wholesale (*s = getConnStateData{}) without touching (and so
// invalidating, mid-critical-section) the mutex guarding it.
type getConnStateData struct {
	getConnStart time.Time
	dnsStart     time.Time
	dnsDone      time.Time
	connectStart time.Time
	connectDone  time.Time
	tlsStart     time.Time
	tlsDone      time.Time
	gotConnAt    time.Time
	ttfbAt       time.Time
	reused       bool
	wasIdle      bool
	idleTime     time.Duration
}

// reset clears every data field, leaving mu untouched. Called from GetConn, with the caller
// re-setting getConnStart immediately after under the same lock: a retry reuses the same
// request context/trace, so stages that don't recur on this attempt (e.g. no dial on a
// reused connection) must not carry over a previous attempt's stale timestamp.
func (s *getConnState) reset() {
	s.getConnStateData = getConnStateData{}
}

// acquired reports whether GetConn fired for this request and, if so, whether GotConn also
// fired. The second return value distinguishes "no connection was ever requested" (e.g. the
// request never got as far as dialing) from "a connection was requested but never
// acquired" (e.g. blocking pool exhaustion, or the request was canceled/timed out while
// waiting) — the exact failure mode a bare "Client.Timeout exceeded while awaiting headers"
// error cannot otherwise be attributed to.
func (s *getConnState) acquired() (started, acquired bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return !s.getConnStart.IsZero(), !s.getConnStart.IsZero() && !s.gotConnAt.IsZero()
}

// newClientMetricTrace returns an httptrace.ClientTrace that records connection-lifecycle
// metrics via m into state (see getConnState), which wrapRoundTripper (in http.go) later
// reads to detect an unacquired connection. If next is non-nil, its callbacks are invoked as
// well, so this can be composed with other traces (e.g. otelhttptrace.NewClientTrace,
// spanAttributeTrace) via otelhttp.WithClientTrace.
func newClientMetricTrace(ctx context.Context, state *getConnState, m *clientMetrics, next *httptrace.ClientTrace) *httptrace.ClientTrace {
	return &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			state.mu.Lock()
			state.reset()
			state.getConnStart = time.Now()
			state.mu.Unlock()
			if next != nil && next.GetConn != nil {
				next.GetConn(hostPort)
			}
		},
		DNSStart: func(info httptrace.DNSStartInfo) {
			state.mu.Lock()
			state.dnsStart = time.Now()
			state.mu.Unlock()
			if next != nil && next.DNSStart != nil {
				next.DNSStart(info)
			}
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			state.mu.Lock()
			state.dnsDone = time.Now()
			dnsStart := state.dnsStart
			state.mu.Unlock()
			if !dnsStart.IsZero() {
				m.dnsDuration.Record(ctx, state.dnsDone.Sub(dnsStart).Seconds())
			}
			if next != nil && next.DNSDone != nil {
				next.DNSDone(info)
			}
		},
		ConnectStart: func(network, addr string) {
			state.mu.Lock()
			state.connectStart = time.Now()
			state.mu.Unlock()
			if next != nil && next.ConnectStart != nil {
				next.ConnectStart(network, addr)
			}
		},
		ConnectDone: func(network, addr string, err error) {
			state.mu.Lock()
			state.connectDone = time.Now()
			connectStart := state.connectStart
			state.mu.Unlock()
			if !connectStart.IsZero() {
				m.connectDuration.Record(ctx, state.connectDone.Sub(connectStart).Seconds())
			}
			if next != nil && next.ConnectDone != nil {
				next.ConnectDone(network, addr, err)
			}
		},
		TLSHandshakeStart: func() {
			state.mu.Lock()
			state.tlsStart = time.Now()
			state.mu.Unlock()
			if next != nil && next.TLSHandshakeStart != nil {
				next.TLSHandshakeStart()
			}
		},
		TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
			state.mu.Lock()
			state.tlsDone = time.Now()
			tlsStart := state.tlsStart
			state.mu.Unlock()
			if !tlsStart.IsZero() {
				m.tlsDuration.Record(ctx, state.tlsDone.Sub(tlsStart).Seconds())
			}
			if next != nil && next.TLSHandshakeDone != nil {
				next.TLSHandshakeDone(cs, err)
			}
		},
		GotConn: func(info httptrace.GotConnInfo) {
			state.mu.Lock()
			state.gotConnAt = time.Now()
			state.reused = info.Reused
			state.wasIdle = info.WasIdle
			state.idleTime = info.IdleTime
			getConnStart := state.getConnStart
			state.mu.Unlock()
			if !getConnStart.IsZero() {
				m.getConnDuration.Record(ctx, state.gotConnAt.Sub(getConnStart).Seconds(),
					m.getConnAttrs[boolIndex(info.Reused)][boolIndex(info.WasIdle)],
				)
			}
			if next != nil && next.GotConn != nil {
				next.GotConn(info)
			}
		},
		GotFirstResponseByte: func() {
			state.mu.Lock()
			state.ttfbAt = time.Now()
			gotConnAt := state.gotConnAt
			state.mu.Unlock()
			if !gotConnAt.IsZero() {
				m.ttfbDuration.Record(ctx, state.ttfbAt.Sub(gotConnAt).Seconds())
			}
			if next != nil && next.GotFirstResponseByte != nil {
				next.GotFirstResponseByte()
			}
		},
		PutIdleConn: func(err error) {
			if err != nil {
				m.putIdleConnErrors.Add(ctx, 1)
			}
			if next != nil && next.PutIdleConn != nil {
				next.PutIdleConn(err)
			}
		},
	}
}
