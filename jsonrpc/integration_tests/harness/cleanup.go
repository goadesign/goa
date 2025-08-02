package harness

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CleanupOrphanedTests removes test directories older than the specified duration
// from the runs directory. This prevents disk space issues from accumulating
// test artifacts over time.
//
// The function scans for directories matching the test run naming pattern
// (YYYYMMDD_HHMMSS_testname) and removes those created before the cutoff time.
// This is typically called during test suite initialization or as a periodic
// maintenance task.
func CleanupOrphanedTests(baseDir string, olderThan time.Duration) error {
	runsDir := filepath.Join(baseDir, "testdata", "runs")
	
	// Check if runs directory exists
	if _, err := os.Stat(runsDir); os.IsNotExist(err) {
		return nil // Nothing to clean
	}
	
	entries, err := os.ReadDir(runsDir)
	if err != nil {
		return fmt.Errorf("failed to read runs directory: %w", err)
	}
	
	cutoff := time.Now().Add(-olderThan)
	var cleaned int
	
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		
		// Parse timestamp from directory name
		// Format: YYYYMMDD_HHMMSS_testname
		parts := strings.Split(entry.Name(), "_")
		if len(parts) < 2 {
			continue
		}
		
		timestamp := parts[0] + parts[1]
		dirTime, err := time.Parse("20060102150405", timestamp)
		if err != nil {
			continue // Skip if can't parse timestamp
		}
		
		// Remove if older than cutoff
		if dirTime.Before(cutoff) {
			dirPath := filepath.Join(runsDir, entry.Name())
			if err := os.RemoveAll(dirPath); err != nil {
				fmt.Printf("Warning: failed to remove %s: %v\n", dirPath, err)
			} else {
				cleaned++
			}
		}
	}
	
	if cleaned > 0 {
		fmt.Printf("Cleaned up %d old test directories\n", cleaned)
	}
	
	return nil
}

// CleanupManager provides cleanup coordination across multiple tests
type CleanupManager struct {
	registered []func() error
}

// NewCleanupManager creates a new cleanup manager for coordinating cleanup
// operations across multiple test resources. The manager ensures cleanup
// functions are executed in LIFO order, which is important for proper
// resource deallocation (e.g., stopping processes before removing directories).
func NewCleanupManager() *CleanupManager {
	return &CleanupManager{
		registered: []func() error{},
	}
}

// Register adds a cleanup function to the manager's stack. Functions are
// executed in reverse order of registration (LIFO) during cleanup. This
// ensures dependencies are properly handled - resources created last are
// cleaned up first.
//
// The cleanup function should return an error if cleanup fails, though
// failures don't prevent other cleanup functions from running.
func (c *CleanupManager) Register(fn func() error) {
	c.registered = append(c.registered, fn)
}

// Cleanup executes all registered cleanup functions in reverse order of
// registration (LIFO). This ensures proper dependency ordering - resources
// created last are cleaned up first.
//
// Errors from individual cleanup functions are logged but don't prevent
// other functions from executing. This ensures maximum cleanup even if
// some operations fail.
func (c *CleanupManager) Cleanup() {
	// Execute in reverse order (LIFO)
	for i := len(c.registered) - 1; i >= 0; i-- {
		if err := c.registered[i](); err != nil {
			// Log but continue with other cleanups
			fmt.Printf("Cleanup error: %v\n", err)
		}
	}
}

// CleanupOnPanic ensures cleanup happens even if a panic occurs during test
// execution. This should be called with defer at the start of any function
// that allocates resources needing cleanup.
//
// The function recovers from the panic, executes the cleanup function, then
// re-panics to preserve the original error for debugging. This pattern ensures
// resources like processes and temporary files are cleaned up even during
// catastrophic test failures.
func CleanupOnPanic(cleanup func()) {
	if r := recover(); r != nil {
		cleanup()
		panic(r) // Re-panic after cleanup
	}
}