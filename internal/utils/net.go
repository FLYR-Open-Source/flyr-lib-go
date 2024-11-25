package utils

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
)

// methodMetric generates an attribute representing the HTTP method for metrics.
//
// This function checks the method matches one of the standard HTTP methods. If the method
// is not one of the recognized methods (CONNECT, DELETE, GET, HEAD, OPTIONS,
// PATCH, POST, PUT, TRACE), it assigns "_OTHER" to the method. It then returns
// the corresponding attribute.KeyValue for the HTTP method, which can be used
// for logging or metrics collection.
func methodMetric(method string) attribute.KeyValue {
	method = strings.ToUpper(method)
	switch method {
	case http.MethodConnect, http.MethodDelete, http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPatch, http.MethodPost, http.MethodPut, http.MethodTrace:
	default:
		method = "_OTHER"
	}
	return semconv.HTTPMethodKey.String(method)
}

// scheme returns the appropriate HTTP scheme attribute based on whether HTTPS is enabled.
//
// This method checks the https parameter and returns the corresponding HTTP scheme
// as an attribute.KeyValue. If https is true, it returns HTTPSchemeHTTPS; otherwise,
// it returns HTTPSchemeHTTP. This is useful for tagging requests with the correct
// protocol scheme.
func scheme(https bool) attribute.KeyValue { // nolint:revive
	if https {
		return semconv.HTTPSchemeHTTPS
	}
	return semconv.HTTPSchemeHTTP
}

// splitHostPort splits a network address hostport of the form "host",
// "host%zone", "[host]", "[host%zone], "host:port", "host%zone:port",
// "[host]:port", "[host%zone]:port", or ":port" into host or host%zone and
// port.
//
// An empty host is returned if it is not provided or unparsable. A negative
// port is returned if it is not provided or unparsable.
func splitHostPort(hostport string) (host string, port int) {
	port = -1

	if strings.HasPrefix(hostport, "[") {
		addrEnd := strings.LastIndex(hostport, "]")
		if addrEnd < 0 {
			// Invalid hostport.
			return
		}
		if i := strings.LastIndex(hostport[addrEnd:], ":"); i < 0 {
			host = hostport[1:addrEnd]
			return
		}
	} else {
		if i := strings.LastIndex(hostport, ":"); i < 0 {
			host = hostport
			return
		}
	}

	host, pStr, err := net.SplitHostPort(hostport)
	if err != nil {
		return
	}

	p, err := strconv.ParseUint(pStr, 10, 16)
	if err != nil {
		return
	}
	return host, int(p) // nolint: gosec  // Bitsize checked to be 16 above.
}

// requiredHTTPPort determines the required HTTP or HTTPS port based on the protocol and provided port.
//
// This function checks whether HTTPS is enabled and validates the provided port number.
// If HTTPS is true, it returns the specified port if it is greater than 0 and not the default HTTPS port (443).
// If HTTPS is false, it returns the specified port if it is greater than 0 and not the default HTTP port (80).
// If the specified port does not meet these criteria, it returns -1 to indicate no custom port is required.
//
// It returns the specified port if it meets the criteria, or -1 if the default port should be used.
func requiredHTTPPort(https bool, port int) int { // nolint:revive
	if https {
		if port > 0 && port != 443 {
			return port
		}
	} else {
		if port > 0 && port != 80 {
			return port
		}
	}
	return -1
}

// netProtocol parses a network protocol string and extracts its name and version.
//
// This function splits the protocol string (formatted as "name/version") into
// two components: the protocol name and its version. The protocol name is
// converted to lowercase for consistency. If the protocol string does not
// contain a "/", the version will be an empty string.
//
// Returns:
//   - name: the protocol name in lowercase.
//   - version: the protocol version as a string.
func netProtocol(proto string) (name string, version string) {
	name, version, _ = strings.Cut(proto, "/")
	name = strings.ToLower(name)
	return name, version
}

// HTTPServerStatus returns a span status code and message for an HTTP status code
// value returned by a server. Status codes in the 400-499 range are not
// returned as errors.
func HTTPServerStatus(code int) (codes.Code, string) {
	if code < 100 || code >= 600 {
		return codes.Error, fmt.Sprintf("Invalid HTTP status code %d", code)
	}
	if code >= 500 {
		return codes.Error, "The request has failed with status: " + http.StatusText(code)
	}
	return codes.Unset, ""
}

// ServerRequestMetrics returns metric attributes for an HTTP request received
// by a server.
//
// The server must be the primary server name if it is known. For example this
// would be the ServerName directive
// (https://httpd.apache.org/docs/2.4/mod/core.html#servername) for an Apache
// server, and the server_name directive
// (http://nginx.org/en/docs/http/ngx_http_core_module.html#server_name) for an
// nginx server. More generically, the primary server name would be the host
// header value that matches the default virtual host of an HTTP server. It
// should include the host identifier and if a port is used to route to the
// server that port identifier should be included as an appropriate port
// suffix.
//
// If the primary server name is not known, server should be an empty string.
// The req Host will be used to determine the server instead.
//
// The following attributes are always returned: "http.method", "http.scheme",
// "net.host.name". The following attributes are returned if they related
// values are defined in req: "net.host.port".
func ServerRequestMetrics(server string, req *http.Request) []attribute.KeyValue {
	/* The following semantic conventions are returned if present:
	http.scheme             string
	http.route              string
	http.method             string
	http.status_code        int
	net.host.name           string
	net.host.port           int
	net.protocol.name       string Note: not set if the value is "http".
	net.protocol.version    string
	*/

	n := 3 // Method, scheme, and host name.
	var host string
	var p int
	if server == "" {
		host, p = splitHostPort(req.Host)
	} else {
		// Prioritize the primary server name.
		host, p = splitHostPort(server)
		if p < 0 {
			_, p = splitHostPort(req.Host)
		}
	}
	hostPort := requiredHTTPPort(req.TLS != nil, p)
	if hostPort > 0 {
		n++
	}
	protoName, protoVersion := netProtocol(req.Proto)
	if protoName != "" {
		n++
	}
	if protoVersion != "" {
		n++
	}

	attrs := make([]attribute.KeyValue, 0, n)

	attrs = append(attrs, methodMetric(req.Method))
	attrs = append(attrs, scheme(req.TLS != nil))
	attrs = append(attrs, semconv.NetHostNameKey.String(host))

	if hostPort > 0 {
		attrs = append(attrs, semconv.NetHostPortKey.Int(hostPort))
	}
	if protoName != "" {
		attrs = append(attrs, semconv.NetProtocolNameKey.String(protoName))
	}
	if protoVersion != "" {
		attrs = append(attrs, semconv.NetProtocolVersionKey.String(protoVersion))
	}

	return attrs
}
