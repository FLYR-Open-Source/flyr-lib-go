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
	"time"

	"github.com/FLYR-Open-Source/flyr-lib-go/monitoring/tracer"
	"go.opentelemetry.io/otel/attribute"
)

// Span attribute keys set by spanAttributeTrace. Named to match the http.client.* metric
// instruments in metrics.go, so the same connection-lifecycle data is queryable both as
// metrics (aggregated across requests) and as attributes on the individual request span
// (see OTEL_ENABLE_HTTP_CLIENT_SPAN_ATTRIBUTES).
const (
	attrGetConnAcquired = "http.client.getconn.acquired"
	attrGetConnDuration = "http.client.getconn.duration"
	attrGetConnReused   = "http.client.getconn.reused"
	attrGetConnWasIdle  = "http.client.getconn.was_idle"
	attrGetConnIdleTime = "http.client.getconn.idle_time"
	attrDNSDuration     = "http.client.dns.duration"
	attrConnectDuration = "http.client.connect.duration"
	attrTLSDuration     = "http.client.tls.duration"
	attrTTFBDuration    = "http.client.ttfb.duration"
)

// spanAttributeTrace returns an httptrace.ClientTrace that sets connection-lifecycle
// attributes directly on the span active in ctx (via tracer.GetSpanFromContext) as each
// phase completes, rather than as separate metric instruments (see clientMetrics). This
// lets a single trace show exactly where one specific request spent its time, which a
// metric aggregated across many requests cannot.
//
// Each attribute is set from within its own hook — not batched into a single call after the
// request finishes — because otelhttp.NewTransport ends its span before RoundTrip returns
// when the request fails (e.g. a connection acquisition that is canceled/times out), and
// span.SetAttributes is a silent no-op once a span has ended. Hooks run while RoundTrip is
// still in progress, before it can reach that span.End() call, so writing attributes there
// is the only way to reliably surface e.g. http.client.getconn.acquired=false — precisely
// the "stuck waiting for connection" failure mode this exists to make visible. GetConn
// optimistically sets acquired=false the instant it fires; GotConn overwrites it to true
// (plus duration/reused/was_idle) if the connection is actually obtained.
//
// If next is non-nil, its callbacks are invoked as well, so this can be composed with other
// traces (e.g. otelhttptrace.NewClientTrace) via otelhttp.WithClientTrace.
func spanAttributeTrace(ctx context.Context, next *httptrace.ClientTrace) *httptrace.ClientTrace {
	span := tracer.GetSpanFromContext(ctx)

	var (
		dnsStart, connectStart, tlsStart, getConnStart, gotConnAt time.Time
	)

	return &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			getConnStart = time.Now()
			span.SetAttributes(attribute.Bool(attrGetConnAcquired, false))
			if next != nil && next.GetConn != nil {
				next.GetConn(hostPort)
			}
		},
		DNSStart: func(info httptrace.DNSStartInfo) {
			dnsStart = time.Now()
			if next != nil && next.DNSStart != nil {
				next.DNSStart(info)
			}
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			if !dnsStart.IsZero() {
				span.SetAttributes(attribute.Float64(attrDNSDuration, time.Since(dnsStart).Seconds()))
			}
			if next != nil && next.DNSDone != nil {
				next.DNSDone(info)
			}
		},
		ConnectStart: func(network, addr string) {
			connectStart = time.Now()
			if next != nil && next.ConnectStart != nil {
				next.ConnectStart(network, addr)
			}
		},
		ConnectDone: func(network, addr string, err error) {
			if !connectStart.IsZero() {
				span.SetAttributes(attribute.Float64(attrConnectDuration, time.Since(connectStart).Seconds()))
			}
			if next != nil && next.ConnectDone != nil {
				next.ConnectDone(network, addr, err)
			}
		},
		TLSHandshakeStart: func() {
			tlsStart = time.Now()
			if next != nil && next.TLSHandshakeStart != nil {
				next.TLSHandshakeStart()
			}
		},
		TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
			if !tlsStart.IsZero() {
				span.SetAttributes(attribute.Float64(attrTLSDuration, time.Since(tlsStart).Seconds()))
			}
			if next != nil && next.TLSHandshakeDone != nil {
				next.TLSHandshakeDone(cs, err)
			}
		},
		GotConn: func(info httptrace.GotConnInfo) {
			gotConnAt = time.Now()
			if !getConnStart.IsZero() {
				span.SetAttributes(
					attribute.Bool(attrGetConnAcquired, true),
					attribute.Float64(attrGetConnDuration, gotConnAt.Sub(getConnStart).Seconds()),
					attribute.Bool(attrGetConnReused, info.Reused),
					attribute.Bool(attrGetConnWasIdle, info.WasIdle),
				)
				if info.WasIdle {
					span.SetAttributes(attribute.Float64(attrGetConnIdleTime, info.IdleTime.Seconds()))
				}
			}
			if next != nil && next.GotConn != nil {
				next.GotConn(info)
			}
		},
		GotFirstResponseByte: func() {
			if !gotConnAt.IsZero() {
				span.SetAttributes(attribute.Float64(attrTTFBDuration, time.Since(gotConnAt).Seconds()))
			}
			if next != nil && next.GotFirstResponseByte != nil {
				next.GotFirstResponseByte()
			}
		},
		PutIdleConn: func(err error) {
			if next != nil && next.PutIdleConn != nil {
				next.PutIdleConn(err)
			}
		},
	}
}
