package common_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"testing"

	"FlyrInc/flyr-lib-go/common-library/logs"
	"FlyrInc/flyr-lib-go/common-library/metrics"
	"FlyrInc/flyr-lib-go/common-library/trace"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace/noop"
)

// Test for logging functionality
func TestLogging(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf) // Capture log output to buffer

	// Initialize logger
	logs.InitLogger("test-service", "dev", "v1.0.0", "rx")

	// Log a sample message
	logs.Log("INFO", "Testing log entry", "common_test.go:24", nil, "trace-id-123", "span-id-123", "user-id-123", "request-id-123")

	// Read log output
	output := buf.String()

	// Check if output is valid JSON
	var logEntry map[string]interface{}
	err := json.Unmarshal([]byte(output), &logEntry)
	if err != nil {
		t.Fatalf("Log output is not valid JSON: %v", err)
	}

	// Check if required fields exist
	if logEntry["level"] != "INFO" || logEntry["message"] != "Testing log entry" {
		t.Errorf("Log output does not contain correct fields: %v", logEntry)
	}
	if logEntry["userId"] != "obfuscated-user-id" {
		t.Errorf("User ID was not obfuscated: %v", logEntry["userId"])
	}
}

// / Test InitMetrics function
func TestInitMetrics(t *testing.T) {
	meter := otel.GetMeterProvider().Meter("test-service")
	if meter == nil {
		t.Fatal("Failed to initialize the meter")
	}
}

// Test RecordMetric function
func TestRecordMetric(t *testing.T) {
	meter := otel.GetMeterProvider().Meter("test-service")
	metrics.RecordMetric(meter, "test_metric", 42.0)
}

// Test InitTracer with Noop Provider
func TestInitTracer(t *testing.T) {
	tracer := trace.InitTracer("test-service")
	if tracer == nil {
		t.Fatal("Failed to initialize the tracer")
	}
}

// Test StartSpan and EndSpan
func TestStartAndEndSpan(t *testing.T) {
	// Use a noop tracer for testing
	tracer := noop.NewTracerProvider().Tracer("test-service")

	// Start and end a span
	_, span := trace.StartSpan(context.Background(), tracer, "test-span")
	if span == nil {
		t.Fatal("Failed to start span")
	}

	// End the span
	trace.EndSpan(span)
}
