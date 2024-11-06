package utils

import (
	"crypto/tls"
	"testing"

	"net/http"

	"github.com/stretchr/testify/assert"
	otelattribute "go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
)

func TestMethodMetric(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected otelattribute.KeyValue
	}{
		// Standard HTTP methods
		{"HTTP GET Method", "GET", semconv.HTTPMethodKey.String("GET")},
		{"HTTP POST Method", "POST", semconv.HTTPMethodKey.String("POST")},
		{"HTTP PUT Method", "PUT", semconv.HTTPMethodKey.String("PUT")},
		{"HTTP DELETE Method", "DELETE", semconv.HTTPMethodKey.String("DELETE")},
		{"HTTP PATCH Method", "PATCH", semconv.HTTPMethodKey.String("PATCH")},
		{"HTTP OPTIONS Method", "OPTIONS", semconv.HTTPMethodKey.String("OPTIONS")},
		{"HTTP HEAD Method", "HEAD", semconv.HTTPMethodKey.String("HEAD")},
		{"HTTP TRACE Method", "TRACE", semconv.HTTPMethodKey.String("TRACE")},
		{"HTTP CONNECT Method", "CONNECT", semconv.HTTPMethodKey.String("CONNECT")},

		// Lowercase input, should be converted to uppercase
		{"Lowercase GET Method", "get", semconv.HTTPMethodKey.String("GET")},
		{"Mixed Case POST Method", "PoSt", semconv.HTTPMethodKey.String("POST")},

		// Non-standard method
		{"Non-standard Method", "FOOBAR", semconv.HTTPMethodKey.String("_OTHER")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := methodMetric(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestScheme(t *testing.T) {
	tests := []struct {
		name     string
		https    bool
		expected otelattribute.KeyValue
	}{
		{"HTTPS scheme", true, semconv.HTTPSchemeHTTPS},
		{"HTTP scheme", false, semconv.HTTPSchemeHTTP},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scheme(tt.https)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSplitHostPort(t *testing.T) {
	tests := []struct {
		name     string
		hostport string
		wantHost string
		wantPort int
	}{
		{
			name:     "IPv4 address with port",
			hostport: "127.0.0.1:8080",
			wantHost: "127.0.0.1",
			wantPort: 8080,
		},
		{
			name:     "IPv6 address with port",
			hostport: "[2001:db8::1]:9090",
			wantHost: "2001:db8::1",
			wantPort: 9090,
		},
		{
			name:     "IPv4 address without port",
			hostport: "127.0.0.1",
			wantHost: "127.0.0.1",
			wantPort: -1,
		},
		{
			name:     "IPv6 address without port",
			hostport: "[2001:db8::1]",
			wantHost: "2001:db8::1",
			wantPort: -1,
		},
		{
			name:     "Invalid format with missing closing bracket",
			hostport: "[2001:db8::1:8080",
			wantHost: "",
			wantPort: -1,
		},
		{
			name:     "Invalid format with non-numeric port",
			hostport: "127.0.0.1:abcd",
			wantHost: "127.0.0.1",
			wantPort: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, port := splitHostPort(tt.hostport)
			if host != tt.wantHost || port != tt.wantPort {
				t.Errorf("splitHostPort(%q) = %q, %d; want %q, %d", tt.hostport, host, port, tt.wantHost, tt.wantPort)
			}
		})
	}
}

func TestRequiredHTTPPort(t *testing.T) {
	tests := []struct {
		name     string
		https    bool
		port     int
		expected int
	}{
		{
			name:     "HTTPS with custom port",
			https:    true,
			port:     8443,
			expected: 8443,
		},
		{
			name:     "HTTPS with default port",
			https:    true,
			port:     443,
			expected: -1,
		},
		{
			name:     "HTTPS with invalid port",
			https:    true,
			port:     -1,
			expected: -1,
		},
		{
			name:     "HTTP with custom port",
			https:    false,
			port:     8080,
			expected: 8080,
		},
		{
			name:     "HTTP with default port",
			https:    false,
			port:     80,
			expected: -1,
		},
		{
			name:     "HTTP with invalid port",
			https:    false,
			port:     -1,
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := requiredHTTPPort(tt.https, tt.port)
			if result != tt.expected {
				t.Errorf("requiredHTTPPort(%v, %d) = %d; want %d", tt.https, tt.port, result, tt.expected)
			}
		})
	}
}

func TestNetProtocol(t *testing.T) {
	tests := []struct {
		name     string
		proto    string
		wantName string
		wantVer  string
	}{
		{
			name:     "Valid protocol with version",
			proto:    "HTTP/1.1",
			wantName: "http",
			wantVer:  "1.1",
		},
		{
			name:     "Valid protocol without version",
			proto:    "HTTP",
			wantName: "http",
			wantVer:  "",
		},
		{
			name:     "Mixed case protocol with version",
			proto:    "HtTp/2",
			wantName: "http",
			wantVer:  "2",
		},
		{
			name:     "Empty string",
			proto:    "",
			wantName: "",
			wantVer:  "",
		},
		{
			name:     "Only slash",
			proto:    "/",
			wantName: "",
			wantVer:  "",
		},
		{
			name:     "Slash at the end",
			proto:    "HTTP/",
			wantName: "http",
			wantVer:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotVer := netProtocol(tt.proto)
			if gotName != tt.wantName || gotVer != tt.wantVer {
				t.Errorf("netProtocol(%q) = %q, %q; want %q, %q", tt.proto, gotName, gotVer, tt.wantName, tt.wantVer)
			}
		})
	}
}

func TestHTTPServerStatus(t *testing.T) {
	tests := []struct {
		name         string
		code         int
		errorMessage string
		wantCode     otelcodes.Code
		wantMessage  string
	}{
		{
			name:         "Valid 500-level status code",
			code:         500,
			errorMessage: "Internal Server Error",
			wantCode:     otelcodes.Error,
			wantMessage:  "Internal Server Error",
		},
		{
			name:         "Valid 400-level status code",
			code:         404,
			errorMessage: "Not Found",
			wantCode:     otelcodes.Unset,
			wantMessage:  "",
		},
		{
			name:         "Valid 200-level status code",
			code:         200,
			errorMessage: "OK",
			wantCode:     otelcodes.Unset,
			wantMessage:  "",
		},
		{
			name:         "Invalid status code below 100",
			code:         99,
			errorMessage: "Invalid Status",
			wantCode:     otelcodes.Error,
			wantMessage:  "Invalid HTTP status code 99",
		},
		{
			name:         "Invalid status code above 599",
			code:         600,
			errorMessage: "Invalid Status",
			wantCode:     otelcodes.Error,
			wantMessage:  "Invalid HTTP status code 600",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCode, gotMessage := HTTPServerStatus(tt.code, tt.errorMessage)
			if gotCode != tt.wantCode || gotMessage != tt.wantMessage {
				t.Errorf("HTTPServerStatus(%d, %q) = %v, %q; want %v, %q",
					tt.code, tt.errorMessage, gotCode, gotMessage, tt.wantCode, tt.wantMessage)
			}
		})
	}
}

func TestServerRequestMetrics(t *testing.T) {
	tests := []struct {
		name     string
		server   string
		req      *http.Request
		wantAttr []otelattribute.KeyValue
	}{
		{
			name:   "HTTP request with server name",
			server: "example.com:8080",
			req: &http.Request{
				Method: "GET",
				Host:   "example.com",
				Proto:  "HTTP/1.1",
			},
			wantAttr: []otelattribute.KeyValue{
				semconv.HTTPMethodKey.String("GET"),
				semconv.HTTPSchemeHTTP,
				semconv.NetHostNameKey.String("example.com"),
				semconv.NetHostPortKey.Int(8080),
				semconv.NetProtocolNameKey.String("http"),
				semconv.NetProtocolVersionKey.String("1.1"),
			},
		},
		{
			name:   "HTTPS request with default port",
			server: "secure.example.com",
			req: &http.Request{
				Method: "POST",
				Host:   "secure.example.com",
				Proto:  "HTTP/2",
				TLS:    &tls.ConnectionState{}, // indicates HTTPS
			},
			wantAttr: []otelattribute.KeyValue{
				semconv.HTTPMethodKey.String("POST"),
				semconv.HTTPSchemeHTTPS,
				semconv.NetHostNameKey.String("secure.example.com"),
				semconv.NetProtocolNameKey.String("http"),
				semconv.NetProtocolVersionKey.String("2"),
			},
		},
		{
			name:   "HTTP request without server name, with port in Host header",
			server: "",
			req: &http.Request{
				Method: "PUT",
				Host:   "localhost:3000",
				Proto:  "HTTP/1.0",
			},
			wantAttr: []otelattribute.KeyValue{
				semconv.HTTPMethodKey.String("PUT"),
				semconv.HTTPSchemeHTTP,
				semconv.NetHostNameKey.String("localhost"),
				semconv.NetHostPortKey.Int(3000),
				semconv.NetProtocolNameKey.String("http"),
				semconv.NetProtocolVersionKey.String("1.0"),
			},
		},
		{
			name:   "HTTPS request with server name and no port",
			server: "secure.local",
			req: &http.Request{
				Method: "DELETE",
				Host:   "secure.local",
				Proto:  "HTTP/1.1",
				TLS:    &tls.ConnectionState{}, // indicates HTTPS
			},
			wantAttr: []otelattribute.KeyValue{
				semconv.HTTPMethodKey.String("DELETE"),
				semconv.HTTPSchemeHTTPS,
				semconv.NetHostNameKey.String("secure.local"),
				semconv.NetProtocolNameKey.String("http"),
				semconv.NetProtocolVersionKey.String("1.1"),
			},
		},
		{
			name:   "Request with empty server and invalid Host",
			server: "",
			req: &http.Request{
				Method: "PATCH",
				Host:   "invalid:host",
				Proto:  "HTTP/1.1",
			},
			wantAttr: []otelattribute.KeyValue{
				semconv.HTTPMethodKey.String("PATCH"),
				semconv.HTTPSchemeHTTP,
				semconv.NetHostNameKey.String("invalid"),
				semconv.NetProtocolNameKey.String("http"),
				semconv.NetProtocolVersionKey.String("1.1"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAttr := ServerRequestMetrics(tt.server, tt.req)
			if len(gotAttr) != len(tt.wantAttr) {
				t.Errorf("ServerRequestMetrics() got %d attributes, want %d", len(gotAttr), len(tt.wantAttr))
				return
			}
			for i, got := range gotAttr {
				want := tt.wantAttr[i]
				if got != want {
					t.Errorf("ServerRequestMetrics() attribute %d = %v, want %v", i, got, want)
				}
			}
		})
	}
}
