package framework

import (
	"os"
	"regexp"
	"strconv"
	"time"
)

// RunnerOption configures the test runner
type RunnerOption func(*RunnerConfig)

// WithParallel enables or disables parallel test execution
func WithParallel(parallel bool) RunnerOption {
	return func(c *RunnerConfig) {
		c.Parallel = parallel
	}
}

// WithFilter sets a regex filter for test scenarios
func WithFilter(pattern string) RunnerOption {
	return func(c *RunnerConfig) {
		c.Filter = pattern
	}
}

// WithServerURL overrides the server URL
func WithServerURL(url string) RunnerOption {
	return func(c *RunnerConfig) {
		c.ServerURL = url
	}
}

// WithKeepGenerated keeps generated code after tests
func WithKeepGenerated(keep bool) RunnerOption {
	return func(c *RunnerConfig) {
		c.KeepGenerated = keep
	}
}

// WithSkipCodeGen skips code generation (assumes it's already done)
func WithSkipCodeGen(skip bool) RunnerOption {
	return func(c *RunnerConfig) {
		c.SkipCodeGen = skip
	}
}

// WithTimeout sets the default timeout for test operations
func WithTimeout(d time.Duration) RunnerOption {
	return func(c *RunnerConfig) {
		c.DefaultTimeout = d
	}
}

// WithDebug enables debug output
func WithDebug(debug bool) RunnerOption {
	return func(c *RunnerConfig) {
		c.Debug = debug
	}
}

// LoadRunnerConfigFromEnv loads configuration from environment variables
func LoadRunnerConfigFromEnv() RunnerConfig {
	config := RunnerConfig{
		DefaultTimeout: 30 * time.Second,
	}
	
	// JSONRPC_TEST_PARALLEL controls parallel execution
	if val := os.Getenv("JSONRPC_TEST_PARALLEL"); val != "" {
		config.Parallel, _ = strconv.ParseBool(val)
	}
	
	// FILTER or JSONRPC_TEST_FILTER for test filtering
	if val := os.Getenv("FILTER"); val != "" {
		config.Filter = val
	} else if val := os.Getenv("JSONRPC_TEST_FILTER"); val != "" {
		config.Filter = val
	}
	
	// JSONRPC_TEST_TIMEOUT for operation timeout
	if val := os.Getenv("JSONRPC_TEST_TIMEOUT"); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			config.DefaultTimeout = d
		}
	}
	
	// KEEP_GENERATED to retain generated code
	config.KeepGenerated = os.Getenv("KEEP_GENERATED") != ""
	
	// JSONRPC_TEST_DEBUG for debug output
	if val := os.Getenv("JSONRPC_TEST_DEBUG"); val != "" {
		config.Debug, _ = strconv.ParseBool(val)
	}
	
	// JSONRPC_TEST_SERVER_URL to use existing server
	if val := os.Getenv("JSONRPC_TEST_SERVER_URL"); val != "" {
		config.ServerURL = val
		config.SkipCodeGen = true // Don't generate if using external server
	}
	
	return config
}

// ApplyOptions applies runner options to a config
func ApplyOptions(config *RunnerConfig, opts ...RunnerOption) {
	for _, opt := range opts {
		opt(config)
	}
}

// ExecutorOption configures test execution
type ExecutorOption func(*executorConfig)

type executorConfig struct {
	WebSocketTimeout time.Duration
	Debug            bool
}

// WithWebSocketTimeout sets the WebSocket operation timeout
func WithWebSocketTimeout(d time.Duration) ExecutorOption {
	return func(c *executorConfig) {
		c.WebSocketTimeout = d
	}
}

// WithExecutorDebug enables debug output for executor
func WithExecutorDebug(debug bool) ExecutorOption {
	return func(c *executorConfig) {
		c.Debug = debug
	}
}

// Update RunnerConfig to include new fields
type RunnerConfigExt struct {
	RunnerConfig
	DefaultTimeout time.Duration
	Debug          bool
}

// ValidateFilter validates and compiles the filter pattern
func ValidateFilter(pattern string) (*regexp.Regexp, error) {
	if pattern == "" {
		return nil, nil
	}
	return regexp.Compile(pattern)
}