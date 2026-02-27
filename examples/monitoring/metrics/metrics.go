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

package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/FLYR-Open-Source/flyr-lib-go/monitoring/meter"
	"github.com/FLYR-Open-Source/flyr-lib-go/monitoring/meter/units"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	serviceName = "some-service"
)

// You don't need this part since it's automated in Kubernetes
func init() {
	_ = os.Setenv("OTEL_SERVICE_NAME", serviceName)
	// this is a flag for exporting the traces in stdout
	_ = os.Setenv("OTEL_EXPORTER_OTLP_TEST", "true")
	// set the log level to debug
	_ = os.Setenv("LOG_LEVEL", "debug")
	_ = os.Setenv("OTEL_RESOURCE_ATTRIBUTES", "k8s.container.name={some-container},k8s.deployment.name={some-deployment},k8s.deployment.uid={some-uid},k8s.namespace.name={some-namespace},k8s.node.name={some-node},k8s.pod.name={some-pod},k8s.pod.uid={some-uid},k8s.replicaset.name={some-replicaset},k8s.replicaset.uid={some-uid},service.instance.id={some-namespace}.{some-pod}.{some-container},service.version={some-version}")
}

// run this file to see the output
func main() {
	ctx := context.Background()

	meter.StartDefaultMeter(ctx)

	defer meter.ShutdownMeterProvider(ctx)

	counterMetrics(ctx)
	gaugeMetrics(ctx)
	histogramMetrics(ctx)
	updownCounters(ctx)
}

func counterMetrics(ctx context.Context) {
	// Initialise the Total MB Upload Meter
	totalMBUploadMeter, err := meter.FloatCounter("file.upload.total_data_mb", meter.MetricInput{
		Description: "This is a float counter",
		Unit:        units.Megabytes,
	})
	if err != nil {
		panic(err)
	}

	// Initialise the Total Requests Meter
	totalRequestsMeter, err := meter.IntCounter("http.server.total_requests", meter.MetricInput{
		Description: "Total number of HTTP requests processed",
		Unit:        units.Ratio,
	})
	if err != nil {
		panic(err)
	}

	// Simulate uploading a file of random size (e.g., 1.5 MB, 2.3 MB, etc.)
	contentLength := 4096
	fileSizeMB := 1.5 + float64(contentLength)/1024.0/1024.0 // Convert bytes to MB
	tags := []attribute.KeyValue{
		attribute.String("file_name", "example.txt"),
		attribute.String("file_type", "text/plain"),
	}
	totalMBUploadMeter.Add(ctx, fileSizeMB, metric.WithAttributes(tags...))

	// Simulate incrementing the total number of requests processed
	tags = []attribute.KeyValue{
		attribute.String("resource", "/my-endpoint"),
	}
	totalRequestsMeter.Add(ctx, 1, metric.WithAttributes(tags...))
}

func gaugeMetrics(ctx context.Context) {
	// Initialise the CPU Usage Meter
	cpuUsageMeter, err := meter.FloatGauge("cpu.usage.percent", meter.MetricInput{
		Description: "Current CPU usage as a percentage",
		Unit:        units.Percent,
	})
	if err != nil {
		panic(err)
	}

	// Initialise the Active HTTP Connections Meter
	activeHttpConnectionsMeter, err := meter.IntGauge("http.server.active_connections", meter.MetricInput{
		Description: "Current number of active HTTP connections",
		Unit:        units.Ratio,
	})
	if err != nil {
		panic(err)
	}

	// Simulate updating the CPU usage value (e.g., 25.0%, 30.5%, etc.)
	tags := []attribute.KeyValue{
		attribute.String("pod_name", "container-watcher-bh8p8"),
		attribute.String("region", "eu-west-1"),
	}
	cpuUsage := 25.0
	cpuUsageMeter.Record(ctx, cpuUsage, metric.WithAttributes(tags...))

	// Simulate updating the number of active HTTP connections
	tags = []attribute.KeyValue{
		attribute.String("resource", "/my-endpoint"),
		attribute.String("region", "eu-west-1"),
		attribute.String("zone", "a"),
	}
	activeHttpConnectionsMeter.Record(ctx, 10, metric.WithAttributes(tags...))
}

func histogramMetrics(ctx context.Context) {
	// Initialise the Request Duration Histogram
	requestDurationHistogram, err := meter.FloatHistogram("http.server.request_duration", meter.HistogramMetricInput{
		MetricInput: meter.MetricInput{
			Description: "Duration of HTTP requests in milliseconds",
			Unit:        units.Milliseconds,
		},
		ExplicitBucketBoundaries: meter.LATENCY_EXPLICIT_BUCKET_BOUNDARIES_IN_MS,
	})
	if err != nil {
		panic(err)
	}

	// Initialise the Request Size Histogram
	requestSizeHistogram, err := meter.IntHistogram(
		"http.server.request_size_bytes",
		meter.HistogramMetricInput{
			MetricInput: meter.MetricInput{
				Description: "Size of HTTP requests in bytes",
				Unit:        units.Bytes,
			},
		},
	)
	if err != nil {
		panic(err)
	}

	// Simulate recording the duration of an HTTP request (e.g., 1000 ms)
	tags := []attribute.KeyValue{
		attribute.String("resource", "/my-endpoint"),
	}
	duration := 1000.0 // (1 second)
	requestDurationHistogram.Record(ctx, duration, metric.WithAttributes(tags...))

	// Simulate recording the size of an HTTP request (e.g., 250 bytes)
	tags = []attribute.KeyValue{
		attribute.String("resource", "/my-endpoint"),
	}
	requestSize := 100 + time.Now().UnixNano()%400 // Random size between 100 and 500 bytes
	requestSizeHistogram.Record(ctx, requestSize, metric.WithAttributes(tags...))
}

func updownCounters(ctx context.Context) {
	// Initialise the Memory Usage Counter
	memoryUsageCounter, err := meter.FloatUpDownCounter(
		"memory.usage.mb", meter.MetricInput{
			Description: "Current memory usage in MB",
			Unit:        units.Megabytes,
		},
	)
	if err != nil {
		panic(err)
	}

	// Initialise the Active HTTP Connections Counter
	activeConnectionsCounter, err := meter.IntUpDownCounter(
		"http.server.active_connections",
		meter.MetricInput{
			Description: "Current number of active HTTP connections",
			Unit:        units.Ratio,
		},
	)
	if err != nil {
		panic(err)
	}

	// Simulate memory usage updates
	go func() {
		for {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			memoryUsageMB := float64(m.Alloc) / 1024.0 / 1024.0 // Convert bytes to MB

			// Update the memory usage counter
			memoryUsageCounter.Add(ctx, memoryUsageMB)

			fmt.Printf("Current memory usage: %.2f MB\n", memoryUsageMB)
			time.Sleep(100 * time.Millisecond)
		}
	}()

	go func() {
		i := 0

		a := 0
		b := 10

		for {
			fmt.Printf("Current active connections: %d\n", i)

			if i%2 == 0 {
				fmt.Printf("Incrementing active connections counter\n")
				activeConnectionsCounter.Add(ctx, 1)
			} else {
				fmt.Printf("Decrementing active connections counter\n")
				activeConnectionsCounter.Add(ctx, -1)
			}
			i = a + rand.Intn(b-a+1)
			time.Sleep(100 * time.Millisecond)
		}
	}()

	// Simulate some work to allow the counter to be updated
	time.Sleep(1 * time.Second)
}
