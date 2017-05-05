package middleware

import (
	"math/rand"
	"testing"
	"time"
)

func TestFixedSampler(t *testing.T) {
	// 0 %
	subject := NewFixedSampler(0)
	for i := 0; i < 10; i++ {
		if subject.Sample() {
			t.Errorf("%d: Sample() returned true for 0%%", i)
		}
	}

	// 100 %
	subject = NewFixedSampler(100)
	for i := 0; i < 10; i++ {
		if !subject.Sample() {
			t.Errorf("%d: Sample() returned false for 100%%", i)
		}
	}

	// 50 %
	rand.Seed(123) // set seed for reproducibility.
	trueCount := 0
	subject = NewFixedSampler(33)
	for i := 0; i < 100; i++ {
		if subject.Sample() {
			trueCount++
		}
	}
	if trueCount != 30 {
		t.Errorf("Unexpected trueCount: %d", trueCount)
	}

	// 66 %
	trueCount = 0
	subject = NewFixedSampler(66)
	for i := 0; i < 100; i++ {
		if subject.Sample() {
			trueCount++
		}
	}
	if trueCount != 67 {
		t.Errorf("Unexpected trueCount: %d", trueCount)
	}
}

func TestAdaptiveSampler(t *testing.T) {
	// initial sampling
	subject := NewAdaptiveSampler(1, 100)
	for i := 0; i < 99; i++ {
		if !subject.Sample() {
			t.Errorf("%d: Sample() returned false before reaching sample size", i)
		}
	}

	// change start time to 1s ago for a more predictable result.
	trueCount := 0
	rand.Seed(123) // set seed for reproducibility.
	now := time.Now()
	subject.(*adaptiveSampler).start = now.Add(-time.Second)
	for i := 99; i < 199; i++ {
		if subject.Sample() {
			trueCount++
		}
	}

	// sample rate should be 1/s
	if trueCount != 1 {
		t.Errorf("Unexpected trueCount: %d", trueCount)
	}

	// start time should be set to now after rate adjustment.
	if subject.(*adaptiveSampler).start.Before(now) {
		t.Errorf("start time was not updated: %v >= $v", subject.(*adaptiveSampler).start, now)
	}

	// simulate last 100 requests taking 10s.
	trueCount = 0
	subject.(*adaptiveSampler).start = time.Now().Add(-time.Second * 10)
	for i := 199; i < 299; i++ {
		if subject.Sample() {
			trueCount++
		}
	}

	// sample rate should be 10/s
	if trueCount != 10 {
		t.Errorf("Unexpected trueCount: %d", trueCount)
	}

	// simulate last 100 requests taking 100s.
	trueCount = 0
	subject.(*adaptiveSampler).start = time.Now().Add(-time.Second * 100)
	for i := 299; i < 399; i++ {
		if subject.Sample() {
			trueCount++
		}
	}

	// sampler should max out and sample all requests.
	if trueCount != 100 {
		t.Errorf("Unexpected trueCount: %d", trueCount)
	}
}
