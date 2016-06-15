// +build !js

package goa

import (
	"time"

	"github.com/armon/go-metrics"
)

// metriks is the local instance of metrics.Metrics
var metriks *metrics.Metrics

// NewMetrics initializes goa's metrics instance with the supplied
// configuration and metrics sink
func NewMetrics(conf *metrics.Config, sink metrics.MetricSink) (err error) {
	metriks, err = metrics.NewGlobal(conf, sink)
	return
}

// AddSample adds a sample to an aggregated metric
// reporting count, min, max, mean, and std deviation
// Usage:
//     AddSample([]string{"my","namespace","key"}, 15.0)
func AddSample(key []string, val float32) {
	if metriks != nil {
		metriks.AddSample(key, val)
	}
}

// EmitKey emits a key/value pair
// Usage:
//     EmitKey([]string{"my","namespace","key"}, 15.0)
func EmitKey(key []string, val float32) {
	if metriks != nil {
		metriks.EmitKey(key, val)
	}
}

// IncrCounter increments the counter named by `key`
// Usage:
//     IncrCounter([]key{"my","namespace","counter"}, 1.0)
func IncrCounter(key []string, val float32) {
	if metriks != nil {
		metriks.IncrCounter(key, val)
	}
}

// MeasureSince creates a timing metric that records
// the duration of elapsed time since `start`
// Usage:
//     MeasureSince([]string{"my","namespace","action}, time.Now())
// Frequently used in a defer:
//    defer MeasureSince([]string{"my","namespace","action}, time.Now())
func MeasureSince(key []string, start time.Time) {
	if metriks != nil {
		metriks.MeasureSince(key, start)
	}
}

// SetGauge sets the named gauge to the specified value
// Usage:
//     SetGauge([]string{"my","namespace"}, 2.0)
func SetGauge(key []string, val float32) {
	if metriks != nil {
		metriks.SetGauge(key, val)
	}
}
