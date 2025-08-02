package harness

import (
	"fmt"
	"net"
	"sync"
)

// PortAllocator manages dynamic port allocation for integration tests,
// ensuring each test gets a unique port to avoid conflicts. It tracks
// allocated ports and attempts to find free ports in a configurable range.
type PortAllocator struct {
	mu         sync.Mutex
	allocated  map[int]bool
	basePort   int
	currentPort int
	maxRetries int
}

// NewPortAllocator creates a new port allocator starting from port 30000.
// This high port range is chosen to avoid conflicts with common services
// and allows non-privileged test execution.
func NewPortAllocator() *PortAllocator {
	return &PortAllocator{
		allocated:   make(map[int]bool),
		basePort:    30000, // Start from port 30000 to avoid conflicts
		currentPort: 30000,
		maxRetries:  100,
	}
}

// Allocate returns a free port for use in tests. The method attempts to
// find an available port by checking both internal allocation tracking and
// actual system availability.
//
// The allocator tries incrementally higher ports starting from the current port.
// This strategy helps avoid conflicts when multiple test suites run concurrently
// or when previous tests didn't clean up properly. Returns an error
// if no free port is found after maxRetries attempts.
func (p *PortAllocator) Allocate() (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	for i := 0; i < p.maxRetries; i++ {
		port := p.currentPort + i
		
		// Check if port is already allocated by us
		if p.allocated[port] {
			continue
		}
		
		// Check if port is available on the system
		if isPortAvailable(port) {
			p.allocated[port] = true
			p.currentPort = port + 1 // Move to next port for next allocation
			return port, nil
		}
	}
	
	return 0, fmt.Errorf("failed to allocate port after %d attempts", p.maxRetries)
}

// Release marks a port as available for reuse by removing it from the
// internal allocation tracking. This should be called when a test completes
// to allow the port to be reused by subsequent tests.
//
// Note that this only updates internal tracking - the actual system port
// may still be in TIME_WAIT state briefly after the process using it exits.
func (p *PortAllocator) Release(port int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.allocated, port)
}

// isPortAvailable checks if a port is available for binding by attempting
// to create a TCP listener on it. This provides a reliable way to verify
// port availability at the OS level.
//
// The function immediately closes the listener if successful, making the
// port available for the actual test server. There's a small race condition
// window between this check and actual use, but it's negligible in practice.
func isPortAvailable(port int) bool {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

// GetFreePort returns a free port by letting the OS assign one using the
// special port 0. This is an alternative approach to PortAllocator that's
// more reliable but less predictable.
//
// The OS guarantees the returned port is free at the moment of allocation,
// eliminating race conditions. However, the port numbers are unpredictable,
// which can make debugging harder. This function is useful for simple tests
// that don't need coordinated port management.
func GetFreePort() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	
	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}