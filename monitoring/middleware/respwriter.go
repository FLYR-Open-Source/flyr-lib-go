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

package middleware // import "github.com/FLYR-Open-Source/flyr-lib-go/monitoring/middleware"

import "net/http"

type customResponseWriter struct {
	responseWriter http.ResponseWriter
	StatusCode     int
}

func extendResponseWriter(w http.ResponseWriter) *customResponseWriter {
	return &customResponseWriter{w, 0}
}

func (w *customResponseWriter) Write(b []byte) (int, error) {
	return w.responseWriter.Write(b)
}

func (w *customResponseWriter) Header() http.Header {
	return w.responseWriter.Header()
}

func (w *customResponseWriter) WriteHeader(statusCode int) {
	// receive status code from this method
	w.StatusCode = statusCode
	w.responseWriter.WriteHeader(statusCode)
}

func (w *customResponseWriter) Done() {
	// if the `w.WriteHeader` wasn't called, set status code to 200 OK
	if w.StatusCode == 0 {
		w.StatusCode = http.StatusOK
	}
}
