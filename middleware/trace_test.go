package middleware

import (
	"math"
	"regexp"
	"testing"
)

func TestNewTraceOptions(t *testing.T) {
	// valid sampling percentage
	{
		cases := map[string]struct{ Rate int }{
			"zero":  {0},
			"one":   {1},
			"fifty": {50},
			"100":   {100},
		}
		for k, c := range cases {
			m := NewTraceOptions(SamplingPercent(c.Rate))
			if m == nil {
				t.Errorf("SamplingPercent(%s): return nil", k)
			}
		}
	}

	// valid adaptive sampler tests
	{
		m := NewTraceOptions(MaxSamplingRate(2))
		if m == nil {
			t.Error("MaxSamplingRate(2): return nil")
		}
		m = NewTraceOptions(MaxSamplingRate(5), SampleSize(100))
		if m == nil {
			t.Error("MaxSamplingRate(5), SampleSize(100): return nil")
		}
	}

	// invalid sampling percentage
	{
		cases := map[string]struct{ SamplingPercentage int }{
			"negative":  {-1},
			"one-o-one": {101},
			"maxint":    {math.MaxInt64},
		}

		for k, c := range cases {
			func() {
				defer func() {
					r := recover()
					if r != "sampling rate must be between 0 and 100" {
						t.Errorf("SamplingPercent(%s): NewTraceOptions did *not* panic as expected: %v", k, r)
					}
				}()
				NewTraceOptions(SamplingPercent(c.SamplingPercentage))
			}()
		}
	}

	// invalid max sampling rate
	{
		cases := map[string]struct{ MaxSamplingRate int }{
			"negative": {-1},
			"zero":     {0},
		}
		for k, c := range cases {
			func() {
				defer func() {
					r := recover()
					if r != "max sampling rate must be greater than 0" {
						t.Errorf("MaxSamplingRate(%s): Trace did *not* panic as expected: %v", k, r)
					}
				}()
				NewTraceOptions(MaxSamplingRate(c.MaxSamplingRate))
			}()
		}
	}

	// invalid sample size
	{
		cases := map[string]struct{ SampleSize int }{
			"negative": {-1},
			"zero":     {0},
		}
		for k, c := range cases {
			func() {
				defer func() {
					r := recover()
					if r != "sample size must be greater than 0" {
						t.Errorf("SampleSize(%s): NewTraceOptions did *not* panic as expected: %v", k, r)
					}
				}()
				NewTraceOptions(SampleSize(c.SampleSize))
			}()
		}
	}

	// invalid discard
	{
		cases := map[string]struct{ Discard *regexp.Regexp }{
			"nil": {nil},
		}
		for k, c := range cases {
			func() {
				defer func() {
					r := recover()
					if r != "discard cannot be nil" {
						t.Errorf("DiscardFromTrace(%s): NewTraceOptions did *not* panic as expected: %v", k, r)
					}
				}()
				NewTraceOptions(DiscardFromTrace(c.Discard))
			}()
		}
	}
}
