package logger

import (
	"encoding/json"
	"testing"
	"time"

	"log/slog"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValueToJSONString(t *testing.T) {
	tests := []struct {
		name    string
		value   slog.Value
		want    string
		wantErr bool
	}{
		{
			name:  "Bool true",
			value: slog.BoolValue(true),
			want:  "true",
		},
		{
			name:  "Int64",
			value: slog.Int64Value(42),
			want:  "42",
		},
		{
			name:  "Uint64",
			value: slog.Uint64Value(42),
			want:  "42",
		},
		{
			name:  "Float64",
			value: slog.Float64Value(3.14),
			want:  "3.14",
		},
		{
			name:  "String",
			value: slog.StringValue("hello"),
			want:  `"hello"`,
		},
		{
			name:  "Duration",
			value: slog.DurationValue(2 * time.Second),
			want:  `"2s"`,
		},
		{
			name:  "Time",
			value: slog.TimeValue(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
			want:  `"2023-01-01T00:00:00Z"`,
		},
		{
			name: "Group",
			value: slog.GroupValue(
				slog.String("key1", "value1"),
				slog.Int64("key2", 123),
			),
			want: `{"key1":"value1","key2":123}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := valueToJSONString(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("valueToJSONString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("valueToJSONString() = %v, want %v", got, tt.want)
			}

			// If we expect a valid JSON output, verify it is valid JSON.
			if !tt.wantErr {
				var jsonCheck interface{}
				if err := json.Unmarshal([]byte(got), &jsonCheck); err != nil {
					t.Errorf("result is not valid JSON: %v", err)
				}
			}
		})
	}
}

func TestConvertToDatadogTags(t *testing.T) {
	type myStruct struct {
		Name string
		Age  int
	}

	group := slog.Group(
		"my_items",
		slog.Any("response_body", myStruct{Name: "joe", Age: 30}),
		slog.Int64("id", 10),
		slog.String("name", "test"),
		slog.Bool("is_active", true),
		slog.Duration("duration", 10*time.Second),
		slog.Float64("amount", 10.5),
	)
	value, err := valueToJSONString(group.Value)
	require.NoError(t, err)

	attributesResult := make(map[string]string)
	convertToDatadogTags(group.Key, value, attributesResult)

	v, ok := attributesResult["my_items.response_body.Name"]
	assert.True(t, ok)
	assert.Equal(t, "joe", v)

	v, ok = attributesResult["my_items.response_body.Age"]
	assert.True(t, ok)
	assert.Equal(t, "30", v)

	v, ok = attributesResult["my_items.id"]
	assert.True(t, ok)
	assert.Equal(t, "10", v)

	v, ok = attributesResult["my_items.name"]
	assert.True(t, ok)
	assert.Equal(t, "test", v)

	v, ok = attributesResult["my_items.is_active"]
	assert.True(t, ok)
	assert.Equal(t, "true", v)

	v, ok = attributesResult["my_items.duration"]
	assert.True(t, ok)
	assert.Equal(t, "1e+10", v)

	v, ok = attributesResult["my_items.amount"]
	assert.True(t, ok)
	assert.Equal(t, "10.5", v)
}
