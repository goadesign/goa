// +build js

package goa

import (
	"time"
)

// Not supported in gopherjs
func IncrCounter(key []string, val float32) {
	// Do nothing
}

// Not supported in gopherjs
func MeasureSince(key []string, start time.Time) {
	// Do nothing
}
