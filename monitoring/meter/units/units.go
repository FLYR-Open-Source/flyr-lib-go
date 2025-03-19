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

package units // import "github.com/FLYR-Open-Source/flyr-lib-go/monitoring/meter/units"

// The unit should be defined using the appropriate [UCUM](https://ucum.org) case-sensitive code.
type Unit string

func (u Unit) String() string {
	return string(u)
}

const (
	// Time
	Days         Unit = "d"
	Hours        Unit = "h"
	Minutes      Unit = "min"
	Seconds      Unit = "s"
	Milliseconds Unit = "ms"
	Microseconds Unit = "us"
	Nanoseconds  Unit = "ns"

	// Bytes
	Bytes     Unit = "By"
	Kibibytes Unit = "KiBy"
	Mebibytes Unit = "MiBy"
	Gibibytes Unit = "GiBy"
	Tebibytes Unit = "TiBy"
	Kilobytes Unit = "KBy"
	Megabytes Unit = "MBy"
	Gigabytes Unit = "GBy"
	Terabytes Unit = "TBy"

	// SI Units
	Meters  Unit = "m"
	Volts   Unit = "V"
	Amperes Unit = "A"
	Joules  Unit = "J"
	Watts   Unit = "W"
	Grams   Unit = "g"

	// Misc
	Celsius Unit = "Cel"
	Hertz   Unit = "Hz"
	Ratio   Unit = "1"
	Percent Unit = "%"
)

// GenerateUnit creates a new unit with the given quantity, wrapped in curly braces.
//
// All non-units that use curly braces to annotate a quantity need to match the grammatical number of the quantity it represent.
// For example if measuring the number of individual requests to a process the unit would be {request}, not {requests}.
func GenerateUnit(quantity string) Unit {
	return Unit("{" + quantity + "}")
}
