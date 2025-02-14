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

package logger // import "github.com/FlyrInc/flyr-lib-go/logger"

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

	attributesResult := make(map[string]string)
	convertToDatadogTags(attr.Key, value, attributesResult)
	for k, v := range attributesResult {
		span.SetAttributes(attribute.String(k, v))
	}
}

// convertToDatadogTags converts a JSON string to a map of key-value pairs.
func convertToDatadogTags(prefix string, jsonString string, result map[string]string) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonString), &data); err != nil {
		return // if it's null just ignore it and swallow the error
	}

	flattenJSON(data, prefix, result)
}

// flattenJSON flattens a JSON object into a map of key-value pairs.
//
// The span attributes' key must have the format `key1.key2.key3` to be properly flattened.
func flattenJSON(data map[string]interface{}, prefix string, result map[string]string) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case map[string]interface{}:
			flattenJSON(v, fullKey, result)
		default:
			result[fullKey] = fmt.Sprintf("%v", value)
		}
	}
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
