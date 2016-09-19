package goa_test

import (
	"time"

	"github.com/goadesign/goa"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// mock metrics
type mockMetrics struct{}

func (m mockMetrics) SetGauge(key []string, val float32)         {}
func (m mockMetrics) EmitKey(key []string, val float32)          {}
func (m mockMetrics) IncrCounter(key []string, val float32)      {}
func (m mockMetrics) AddSample(key []string, val float32)        {}
func (m mockMetrics) MeasureSince(key []string, start time.Time) {}

var _ = Describe("Metrics", func() {
	var keys = [6]string{}
	var metrics goa.Metrics

	BeforeEach(func() {
		metrics = goa.MetricDiscarder{}
		keys = [6]string{
			"foo_bar_*/*",
			"foo_*_baz",
			"foo/baz",
			"foo/bar/baz",
			"foo/bar*_*/*",
			"//foo/bar*",
		}
	})

	Describe("Add sample", func() {
		Context("With invalid characters in key", func() {
			It("should replace invalid characters with normalized characters", func() {
				goa.SetMetrics(metrics)
				goa.AddSample(keys[:], 3.14)
				Ω(keys).Should(ConsistOf([]string{
					"foo_bar_all",
					"foo___baz",
					"foo_baz",
					"foo_bar_baz",
					"foo_bar__all",
					"__foo_bar_",
				}))
			})
		})
	})

	Describe("Emit key", func() {
		Context("With invalid characters in key", func() {
			It("should replace invalid characters with normalized characters", func() {
				goa.SetMetrics(metrics)
				goa.EmitKey(keys[:], 3.14)
				Ω(keys).Should(ConsistOf([]string{
					"foo_bar_all",
					"foo___baz",
					"foo_baz",
					"foo_bar_baz",
					"foo_bar__all",
					"__foo_bar_",
				}))
			})
		})
	})

	Describe("Increment Counter", func() {
		Context("With invalid characters in key", func() {
			It("should replace invalid characters with normalized characters", func() {
				goa.SetMetrics(metrics)
				goa.IncrCounter(keys[:], 3.14)
				Ω(keys).Should(ConsistOf([]string{
					"foo_bar_all",
					"foo___baz",
					"foo_baz",
					"foo_bar_baz",
					"foo_bar__all",
					"__foo_bar_",
				}))
			})
		})
	})

	Describe("Measure since", func() {
		Context("With invalid characters in key", func() {
			It("should replace invalid characters with normalized characters", func() {
				goa.SetMetrics(metrics)
				goa.MeasureSince(keys[:], time.Time{})
				Ω(keys).Should(ConsistOf([]string{
					"foo_bar_all",
					"foo___baz",
					"foo_baz",
					"foo_bar_baz",
					"foo_bar__all",
					"__foo_bar_",
				}))
			})
		})
	})

	Describe("Set gauge", func() {
		Context("With invalid characters in key", func() {
			It("should replace invalid characters with normalized characters", func() {
				goa.SetMetrics(metrics)
				goa.SetGauge(keys[:], 3.14)
				Ω(keys).Should(ConsistOf([]string{
					"foo_bar_all",
					"foo___baz",
					"foo_baz",
					"foo_bar_baz",
					"foo_bar__all",
					"__foo_bar_",
				}))
			})
		})
	})
})
