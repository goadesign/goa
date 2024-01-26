package middleware

import (
	"crypto/rand"
	"math/big"
	"sync"
	"sync/atomic"
	"time"
)

type (
	// Sampler is an interface for computing when a sample falls within a range.
	Sampler interface {
		// Sample returns true if the caller should sample now.
		Sample() bool
	}

	// randomizer is an interface for generating random numbers.
	randomizer interface {
		// Int64 returns a random int64 in the range [0, bound).
		Int64(bound int64) int64
	}

	// randomGenerator is a randomizer that uses crypto/rand.
	randomGenerator struct {
	}

	adaptiveSampler struct {
		sync.Mutex
		lastRate        int64
		maxSamplingRate int
		sampleSize      uint32
		start           time.Time
		counter         uint32
	}

	fixedSampler int
)

const (
	// adaptive upper bound has granularity in case caller becomes extremely busy.
	adaptiveUpperBoundInt   = int64(10000)
	adaptiveUpperBoundFloat = float64(adaptiveUpperBoundInt)
)

var rnd randomizer = newRandomizer()

// NewAdaptiveSampler computes the interval for sampling for tracing middleware.
// it can also be used by non-web go routines to trace internal API calls.
//
// maxSamplingRate is the desired maximum sampling rate in requests per second.
//
// sampleSize sets the number of requests between two adjustments of the
// sampling rate when MaxSamplingRate is set. the sample rate cannot be adjusted
// until the sample size is reached at least once.
func NewAdaptiveSampler(maxSamplingRate, sampleSize int) Sampler {
	if maxSamplingRate <= 0 {
		panic("maxSamplingRate must be greater than 0")
	}
	if sampleSize <= 0 {
		panic("sample size must be greater than 0")
	}
	return &adaptiveSampler{
		lastRate:        adaptiveUpperBoundInt, // samples all until initial count reaches sample size
		maxSamplingRate: maxSamplingRate,
		sampleSize:      uint32(sampleSize),
		start:           time.Now(),
	}
}

// NewFixedSampler sets the tracing sampling rate as a percentage value.
func NewFixedSampler(samplingPercent int) Sampler {
	if samplingPercent < 0 || samplingPercent > 100 {
		panic("samplingPercent must be between 0 and 100")
	}
	return fixedSampler(samplingPercent)
}

// Sample implementation for adaptive rate
func (s *adaptiveSampler) Sample() bool {
	// adjust sampling rate whenever sample size is reached.
	var currentRate int64
	if atomic.AddUint32(&s.counter, 1) == s.sampleSize { // exact match prevents
		atomic.StoreUint32(&s.counter, 0) // race is ok
		s.Lock()
		{
			d := time.Since(s.start).Seconds()
			r := float64(s.sampleSize) / d
			currentRate = int64((float64(s.maxSamplingRate) * adaptiveUpperBoundFloat) / r)
			if currentRate > adaptiveUpperBoundInt {
				currentRate = adaptiveUpperBoundInt
			} else if currentRate < 1 {
				currentRate = 1
			}
			s.start = time.Now()
		}
		s.Unlock()
		atomic.StoreInt64(&s.lastRate, int64(currentRate))
	} else {
		currentRate = int64(atomic.LoadInt64(&s.lastRate))
	}

	// currentRate is never zero.
	return currentRate == adaptiveUpperBoundInt || rnd.Int64(adaptiveUpperBoundInt) < currentRate
}

// Sample implementation for fixed percentage
func (s fixedSampler) Sample() bool {
	samplingPercent := int64(s)
	return samplingPercent > 0 && (samplingPercent == 100 || rnd.Int64(100) < samplingPercent)
}

// newRandomizer returns a randomizer.
func newRandomizer() randomizer {
	return &randomGenerator{}
}

// Int64 returns a random int64 in the range [0, bound).
func (r *randomGenerator) Int64(bound int64) int64 {
	// can we ignore the error?
	rnd, _ := rand.Int(rand.Reader, big.NewInt(bound))
	return rnd.Int64()
}
