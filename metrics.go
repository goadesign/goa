// +build !js

package goa

import (
	"regexp"
	"strings"
	"time"

	"github.com/armon/go-metrics"
)

const (
	allMatcher      string = "*/*"
	allReplacement  string = "all"
	normalizedToken string = "_"
)

var (
	// metriks is the local instance of metrics.Metrics
	metriks *metrics.Metrics

	// used for normalizing names by matching '*' and '/' so they can be replaced.
	invalidCharactersRE = regexp.MustCompile(`[\*/]`)
)

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
		normalizeKeys(key)
		metriks.AddSample(key, val)
	}
}

// EmitKey emits a key/value pair
// Usage:
//     EmitKey([]string{"my","namespace","key"}, 15.0)
func EmitKey(key []string, val float32) {
	if metriks != nil {
		normalizeKeys(key)
		metriks.EmitKey(key, val)
	}
}

// IncrCounter increments the counter named by `key`
// Usage:
//     IncrCounter([]key{"my","namespace","counter"}, 1.0)
func IncrCounter(key []string, val float32) {
	if metriks != nil {
		normalizeKeys(key)
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
		normalizeKeys(key)
		metriks.MeasureSince(key, start)
	}
}

// SetGauge sets the named gauge to the specified value
// Usage:
//     SetGauge([]string{"my","namespace"}, 2.0)
func SetGauge(key []string, val float32) {
	if metriks != nil {
		normalizeKeys(key)
		metriks.SetGauge(key, val)
	}
}

// takes in the list of keys and
func normalizeKeys(key []string) {
	if key != nil {
		for i, k := range(key) {
			// first replace */* with all
			k = strings.Replace(k, allMatcher, allReplacement, -1)

			// now replace all other invalid characters with a safe one.
			key[i] = invalidCharactersRE.ReplaceAllString(k, normalizedToken)
		}
	}
}
