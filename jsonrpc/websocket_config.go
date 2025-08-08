package jsonrpc

import (
	"context"
	"time"
)

type (
	// StreamErrorType represents different types of WebSocket stream errors
	StreamErrorType int

	// StreamErrorHandler allows users to handle stream errors
	StreamErrorHandler func(ctx context.Context, errorType StreamErrorType, err error, response *RawResponse)

	// StreamConfig contains configuration options for WebSocket streams
	StreamConfig struct {
		// Timeouts
		RequestTimeout    time.Duration // Timeout for individual requests (default: 30s)
		ConnectionTimeout time.Duration // Timeout for establishing connections (default: 10s)
		CloseTimeout      time.Duration // Timeout for graceful stream closure (default: 5s)

		// Buffer Sizes
		ResultChannelBuffer int // Buffer size for result channels (default: 1)
		WriteBufferSize     int // WebSocket write buffer size (default: 4096)
		ReadBufferSize      int // WebSocket read buffer size (default: 4096)

		// Retry Configuration
		MaxRetries       int           // Maximum number of connection retries (default: 3)
		RetryBackoffBase time.Duration // Base delay for exponential backoff (default: 1s)
		RetryBackoffMax  time.Duration // Maximum retry delay (default: 30s)

		// Advanced Options
		EnableCompression bool          // Enable WebSocket compression (default: false)
		PingInterval      time.Duration // Interval for sending ping frames (default: 30s)

		// Error Handling
		ErrorHandler StreamErrorHandler // Optional error handler for stream events (default: nil)
	}

	// StreamConfigOption is a function that modifies StreamConfig
	StreamConfigOption func(*StreamConfig)
)

const (
	StreamErrorConnection StreamErrorType = iota // WebSocket connection errors
	StreamErrorProtocol                          // Invalid JSON-RPC protocol
	StreamErrorParsing                           // Failed to parse/decode response
	StreamErrorOrphaned                          // Response with no matching request
	StreamErrorTimeout                           // Request timeout
	StreamErrorNotification                      // Server-initiated notification received
)

// WithRequestTimeout sets the timeout for individual requests
func WithRequestTimeout(timeout time.Duration) StreamConfigOption {
	return func(c *StreamConfig) {
		c.RequestTimeout = timeout
	}
}

// WithConnectionTimeout sets the timeout for establishing connections
func WithConnectionTimeout(timeout time.Duration) StreamConfigOption {
	return func(c *StreamConfig) {
		c.ConnectionTimeout = timeout
	}
}

// WithCloseTimeout sets the timeout for graceful stream closure
func WithCloseTimeout(timeout time.Duration) StreamConfigOption {
	return func(c *StreamConfig) {
		c.CloseTimeout = timeout
	}
}

// WithResultChannelBuffer sets the buffer size for result channels
func WithResultChannelBuffer(size int) StreamConfigOption {
	return func(c *StreamConfig) {
		c.ResultChannelBuffer = size
	}
}

// WithWebSocketBuffers sets both read and write buffer sizes
func WithWebSocketBuffers(readSize, writeSize int) StreamConfigOption {
	return func(c *StreamConfig) {
		c.ReadBufferSize = readSize
		c.WriteBufferSize = writeSize
	}
}

// WithRetryConfig sets retry behavior parameters
func WithRetryConfig(maxRetries int, baseDelay, maxDelay time.Duration) StreamConfigOption {
	return func(c *StreamConfig) {
		c.MaxRetries = maxRetries
		c.RetryBackoffBase = baseDelay
		c.RetryBackoffMax = maxDelay
	}
}

// WithCompression enables or disables WebSocket compression
func WithCompression(enabled bool) StreamConfigOption {
	return func(c *StreamConfig) {
		c.EnableCompression = enabled
	}
}

// WithPingInterval sets the interval for sending ping frames
func WithPingInterval(interval time.Duration) StreamConfigOption {
	return func(c *StreamConfig) {
		c.PingInterval = interval
	}
}

// WithErrorHandler sets the error handler for stream events
func WithErrorHandler(handler StreamErrorHandler) StreamConfigOption {
	return func(c *StreamConfig) {
		c.ErrorHandler = handler
	}
}

// NewStreamConfig creates a StreamConfig with the given options
func NewStreamConfig(opts ...StreamConfigOption) *StreamConfig {
	config := defaultStreamConfig()
	for _, opt := range opts {
		opt(config)
	}
	return config.Validate()
}

// defaultStreamConfig returns a StreamConfig with sensible production defaults
func defaultStreamConfig() *StreamConfig {
	return &StreamConfig{
		// Reasonable timeout defaults
		RequestTimeout:    30 * time.Second,
		ConnectionTimeout: 10 * time.Second,
		CloseTimeout:      5 * time.Second,

		// Conservative buffer sizes
		ResultChannelBuffer: 1,
		WriteBufferSize:     4096,
		ReadBufferSize:      4096,

		// Moderate retry behavior
		MaxRetries:       3,
		RetryBackoffBase: 1 * time.Second,
		RetryBackoffMax:  30 * time.Second,

		// Safe advanced defaults
		EnableCompression: false,
		PingInterval:      30 * time.Second,
	}
}

// Validate checks the configuration and applies constraints
func (c *StreamConfig) Validate() *StreamConfig {
	// Ensure positive timeouts
	if c.RequestTimeout <= 0 {
		c.RequestTimeout = 30 * time.Second
	}
	if c.ConnectionTimeout <= 0 {
		c.ConnectionTimeout = 10 * time.Second
	}
	if c.CloseTimeout <= 0 {
		c.CloseTimeout = 5 * time.Second
	}

	// Ensure reasonable buffer sizes
	if c.ResultChannelBuffer < 1 {
		c.ResultChannelBuffer = 1
	}
	if c.WriteBufferSize < 1024 {
		c.WriteBufferSize = 1024
	}
	if c.ReadBufferSize < 1024 {
		c.ReadBufferSize = 1024
	}

	// Ensure reasonable retry configuration
	if c.MaxRetries < 0 {
		c.MaxRetries = 0
	}
	if c.RetryBackoffBase <= 0 {
		c.RetryBackoffBase = 1 * time.Second
	}
	if c.RetryBackoffMax < c.RetryBackoffBase {
		c.RetryBackoffMax = c.RetryBackoffBase * 30
	}

	// Ensure reasonable ping interval
	if c.PingInterval <= 0 {
		c.PingInterval = 30 * time.Second
	}

	return c
}
