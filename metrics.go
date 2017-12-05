// +build !js,!appengine

package goa

import (
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/armon/go-metrics"
)

const (
	allMatcher      string = "*/*"
	allReplacement  string = "all"
	normalizedToken string = "_"
)

var (
	// metriks contains current collector
	metriks Collector

	// metriksMu is mutex for metriks variable
	metriksMu sync.Mutex

	// invalidCharactersRE is the invert match of validCharactersRE
	invalidCharactersRE = regexp.MustCompile(`[\*/]`)

	// Taken from https://github.com/prometheus/client_golang/blob/66058aac3a83021948e5fb12f1f408ff556b9037/prometheus/desc.go
	validCharactersRE = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_:]*$`)
)

// Collector is the interface used for collecting metrics.
type Collector interface {
	AddSample(key []string, val float32)
	EmitKey(key []string, val float32)
	IncrCounter(key []string, val float32)
	MeasureSince(key []string, start time.Time)
	SetGauge(key []string, val float32)
}

func init() {
	SetMetrics(NewNoOpCollector())
}

// newNoOpCollecter implements Collector, but provides no collection.
type noOpCollecter struct{}

func (*noOpCollecter) AddSample(key []string, val float32)        {}
func (*noOpCollecter) EmitKey(key []string, val float32)          {}
func (*noOpCollecter) IncrCounter(key []string, val float32)      {}
func (*noOpCollecter) MeasureSince(key []string, start time.Time) {}
func (*noOpCollecter) SetGauge(key []string, val float32)         {}

// NewNoOpCollector returns a Collector that does no collection.
func NewNoOpCollector() Collector {
	return &noOpCollecter{}
}

// NewNoOpSink returns a NOOP sink.
func NewNoOpSink() metrics.MetricSink {
	return &NoOpSink{}
}

// NoOpSink default NOOP metrics recorder
type NoOpSink struct{}

func (*NoOpSink) SetGauge(key []string, val float32)                                      {}
func (*NoOpSink) SetGaugeWithLabels(key []string, val float32, labels []metrics.Label)    {}
func (*NoOpSink) EmitKey(key []string, val float32)                                       {}
func (*NoOpSink) IncrCounter(key []string, val float32)                                   {}
func (*NoOpSink) IncrCounterWithLabels(key []string, val float32, labels []metrics.Label) {}
func (*NoOpSink) AddSample(key []string, val float32)                                     {}
func (*NoOpSink) AddSampleWithLabels(key []string, val float32, labels []metrics.Label)   {}

// NewMetrics initializes goa's metrics instance with the supplied
// configuration and metrics sink
// This method is deprecated and SetMetrics should be used instead.
func NewMetrics(conf *metrics.Config, sink metrics.MetricSink) (err error) {
	m, err := metrics.NewGlobal(conf, sink)
	SetMetrics(m)

	return nil
}

// SetMetrics initializes goa's metrics instance with the supplied metrics adapter interface.
func SetMetrics(m Collector) {
	metriksMu.Lock()
	metriks = m
	metriksMu.Unlock()
}

// GetMetrics returns goa's metrics collector adapter interface.
func GetMetrics() Collector {
	metriksMu.Lock()
	m := metriks
	metriksMu.Unlock()
	return m
}

// AddSample adds a sample to an aggregated metric
// reporting count, min, max, mean, and std deviation
// Usage:
//     AddSample([]string{"my","namespace","key"}, 15.0)
func AddSample(key []string, val float32) {
	normalizeKeys(key)

	GetMetrics().AddSample(key, val)
}

// EmitKey emits a key/value pair
// Usage:
//     EmitKey([]string{"my","namespace","key"}, 15.0)
func EmitKey(key []string, val float32) {
	normalizeKeys(key)

	GetMetrics().EmitKey(key, val)
}

// IncrCounter increments the counter named by `key`
// Usage:
//     IncrCounter([]key{"my","namespace","counter"}, 1.0)
func IncrCounter(key []string, val float32) {
	normalizeKeys(key)

	GetMetrics().IncrCounter(key, val)
}

// MeasureSince creates a timing metric that records
// the duration of elapsed time since `start`
// Usage:
//     MeasureSince([]string{"my","namespace","action"}, time.Now())
// Frequently used in a defer:
//    defer MeasureSince([]string{"my","namespace","action"}, time.Now())
func MeasureSince(key []string, start time.Time) {
	normalizeKeys(key)

	GetMetrics().MeasureSince(key, start)
}

// SetGauge sets the named gauge to the specified value
// Usage:
//     SetGauge([]string{"my","namespace"}, 2.0)
func SetGauge(key []string, val float32) {
	normalizeKeys(key)

	GetMetrics().SetGauge(key, val)
}

// This function is used to make metric names safe for all metric services. Specifically, prometheus does
// not support * or / in metric names.
func normalizeKeys(key []string) {
	for i, k := range key {
		if !validCharactersRE.MatchString(k) {
			// first replace */* with all
			k = strings.Replace(k, allMatcher, allReplacement, -1)

			// now replace all other invalid characters with a safe one.
			key[i] = invalidCharactersRE.ReplaceAllString(k, normalizedToken)
		}
	}
}
