package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/FlyrInc/flyr-lib-go/internal/span"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// setErroredSpan sets the span as errored if it is recording.
//
// It marks the current span (if found) as errored and records the given error.
func setErroredSpan(ctx context.Context, err error) {
	span := span.GetSpanFromContext(ctx)

	if span.IsRecording() {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
	}
}

// injectAttrsToSpan injects the given attributes to the current span if it is recording.
//
// The attributes are converted to from slog.Attr to Go types (primitives or not) and then added to the span.
func injectAttrsToSpan(ctx context.Context, attr slog.Attr) {
	span := span.GetSpanFromContext(ctx)

	if !span.IsRecording() {
		return
	}

	value, err := valueToJSONString(attr.Value)
	if err != nil {
		return // if it's null just ignore it and swallow the error
	}

	span.SetAttributes(attribute.String(attr.Key, value))
}

// valueToJSONString converts a slog.Value to its JSON string representation.
//
// This function takes a slog.Value and extracts its underlying data based on its kind.
// It supports various kinds, including Any, Bool, Int64, Uint64, Float64, String,
// Duration, Time, Group, and LogValuer. For Group kinds, it constructs a map of
// attributes. If the value cannot be processed due to an unsupported kind, it returns
// an error. After extracting the data, it marshals it into a JSON string using the
// encoding/json package.
//
// It returns a JSON string representation of the slog.Value
// and an error if the conversion fails.
func valueToJSONString(value slog.Value) (string, error) {
	var data interface{}

	// Determine the kind of slog.Value and extract data accordingly
	switch value.Kind() {
	case slog.KindAny:
		data = value.Any()

	case slog.KindBool:
		data = value.Bool()

	case slog.KindInt64:
		data = value.Int64()

	case slog.KindUint64:
		data = value.Uint64()

	case slog.KindFloat64:
		data = value.Float64()

	case slog.KindString:
		data = value.String()

	case slog.KindDuration:
		data = value.Duration().String()

	case slog.KindTime:
		data = value.Time()

	case slog.KindGroup:
		// Groups are collections of attributes, convert to a map
		groupData := make(map[string]interface{})
		for _, attr := range value.Group() {
			groupData[attr.Key] = attr.Value.Any()
		}
		data = groupData

	case slog.KindLogValuer:
		// For dynamic values, resolve them and then process recursively
		return valueToJSONString(value.LogValuer().LogValue())

	default:
		// Return an error for unsupported kinds
		return "", fmt.Errorf("unsupported slog.Value kind: %v", value.Kind())
	}

	// Marshal the extracted data to JSON
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal slog.Value to JSON: %w", err)
	}

	return string(jsonBytes), nil
}
