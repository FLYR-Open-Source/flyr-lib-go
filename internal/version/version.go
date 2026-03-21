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

package version // import "github.com/FLYR-Open-Source/flyr-lib-go/internal/version"

import (
	"runtime/debug"
	"sync"
)

const modulePath = "github.com/FLYR-Open-Source/flyr-lib-go"

var (
	resolvedVersion = "unknown"
	once            sync.Once
)

// Version returns the version of the flyr-lib-go module.
// It is resolved once at runtime using debug.ReadBuildInfo().
// When the library is consumed as a dependency, the version is read from the
// module's build info. When running tests or binaries within this repo, it
// returns "unknown" because the main module version is "(devel)".
func Version() string {
	once.Do(func() {
		bi, ok := debug.ReadBuildInfo()
		if !ok {
			return
		}

		for _, dep := range bi.Deps {
			if dep.Path == modulePath {
				resolvedVersion = dep.Version
				return
			}
		}

		if bi.Main.Path == modulePath && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
			resolvedVersion = bi.Main.Version
		}
	})

	return resolvedVersion
}
