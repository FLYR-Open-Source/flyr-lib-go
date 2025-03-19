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

package units

import (
	"fmt"
	"testing"
)

func TestUnitToString(t *testing.T) {
	testCases := []struct {
		unit     Unit
		expected string
	}{
		// Time
		{Days, "d"},
		{Hours, "h"},
		{Minutes, "min"},
		{Seconds, "s"},
		{Milliseconds, "ms"},
		{Microseconds, "us"},
		{Nanoseconds, "ns"},

		// Bytes
		{Bytes, "By"},
		{Kibibytes, "KiBy"},
		{Mebibytes, "MiBy"},
		{Gibibytes, "GiBy"},
		{Tebibytes, "TiBy"},
		{Kilobytes, "KBy"},
		{Megabytes, "MBy"},
		{Gigabytes, "GBy"},
		{Terabytes, "TBy"},

		// SI Units
		{Meters, "m"},
		{Volts, "V"},
		{Amperes, "A"},
		{Joules, "J"},
		{Watts, "W"},
		{Grams, "g"},

		// Misc
		{Celsius, "Cel"},
		{Hertz, "Hz"},
		{Ratio, "1"},
		{Percent, "%"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Unit %s", tc.unit), func(t *testing.T) {
			if tc.unit.String() != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, tc.unit.String())
			}
		})
	}

}

func TestGenerateUnit(t *testing.T) {
	tests := []struct {
		name     string
		quantity string
		expected Unit
	}{
		{
			name:     "valid quantity",
			quantity: "time",
			expected: Unit("{time}"),
		},
		{
			name:     "empty quantity",
			quantity: "",
			expected: Unit("{}"),
		},
		{
			name:     "numeric quantity",
			quantity: "123",
			expected: Unit("{123}"),
		},
		{
			name:     "special characters in quantity",
			quantity: "bytes/second",
			expected: Unit("{bytes/second}"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateUnit(tt.quantity)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}
