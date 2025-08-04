package framework

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
	"goa.design/goa/v3/jsonrpc/integration_tests/harness"
)

// RunnerConfig holds runner configuration
type RunnerConfig struct {
	// Parallel runs tests in parallel
	Parallel bool
	// Filter is a regex to filter scenarios by name
	Filter string
	// ServerURL overrides the server URL
	ServerURL string
	// SkipCodeGen skips code generation (assumes it's already done)
	SkipCodeGen bool
	// KeepGenerated keeps generated code after tests
	KeepGenerated bool
	// DefaultTimeout is the default timeout for operations
	DefaultTimeout time.Duration
	// Debug enables debug output
	Debug bool
}

// Runner executes JSON-RPC test scenarios
type Runner struct {
	config          Config
	runnerConfig    RunnerConfig
	// generator field removed - created on demand
	testDir         string
	servers         map[string]*harness.Server
	filterPattern   *regexp.Regexp
	executorFactory func(string) *Executor
}

// NewRunner creates a new test runner
func NewRunner(scenariosPath string, opts ...RunnerOption) (*Runner, error) {
	// Start with environment config
	runnerConfig := LoadRunnerConfigFromEnv()
	
	// Apply options
	ApplyOptions(&runnerConfig, opts...)
	
	data, err := os.ReadFile(scenariosPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read scenarios: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse scenarios: %w", err)
	}

	runner := &Runner{
		config:       config,
		runnerConfig: runnerConfig,
		servers:      make(map[string]*harness.Server),
	}

	// Compile filter pattern if provided
	if runnerConfig.Filter != "" {
		pattern, err := regexp.Compile(runnerConfig.Filter)
		if err != nil {
			return nil, fmt.Errorf("invalid filter pattern: %w", err)
		}
		runner.filterPattern = pattern
	}

	return runner, nil
}

// SetExecutorFactory sets a custom executor factory
func (r *Runner) SetExecutorFactory(factory func(string) *Executor) {
	r.executorFactory = factory
}

// TestDir returns the test directory path
func (r *Runner) TestDir() string {
	return r.testDir
}

// Run executes all test scenarios
func (r *Runner) Run(t *testing.T) {
	// Setup test directory
	if r.runnerConfig.KeepGenerated {
		tempDir, err := os.MkdirTemp("", "jsonrpc-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		r.testDir = tempDir
		t.Logf("Generated code in: %s", tempDir)
	} else {
		r.testDir = t.TempDir()
	}

	// Generate code if needed
	if !r.runnerConfig.SkipCodeGen {
		if err := r.generateCode(t); err != nil {
			t.Fatalf("Failed to generate code: %v", err)
		}
	}

	// Start servers
	if err := r.startServers(t); err != nil {
		t.Fatalf("Failed to start servers: %v", err)
	}
	defer r.stopServers()

	// Count scenarios to run
	scenarios := r.filterScenarios()
	if len(scenarios) == 0 {
		t.Skip("No scenarios match filter")
	}

	t.Logf("Running %d scenarios", len(scenarios))

	// Execute scenarios
	for _, scenario := range scenarios {
		scenario := scenario // capture for parallel tests
		
		t.Run(scenario.Name, func(t *testing.T) {
			if r.runnerConfig.Parallel {
				t.Parallel()
			}
			r.runScenario(t, scenario)
		})
	}
}

// filterScenarios returns scenarios that match the filter
func (r *Runner) filterScenarios() []Scenario {
	if r.filterPattern == nil {
		return r.config.Scenarios
	}

	var filtered []Scenario
	for _, scenario := range r.config.Scenarios {
		if r.filterPattern.MatchString(scenario.Name) {
			filtered = append(filtered, scenario)
		}
	}
	return filtered
}

// generateCode generates DSL and implementation for all methods
func (r *Runner) generateCode(t *testing.T) error {
	t.Helper()
	
	// Collect all unique methods
	methods, err := r.collectMethods()
	if err != nil {
		return err
	}

	// Use generator with templates
	generator := NewGenerator(r.testDir, methods)
	t.Logf("Generating code for %d methods", len(methods))
	if err := generator.Generate(); err != nil {
		t.Logf("Failed to generate code: %v", err)
		return err
	}
	return nil
}

// collectMethods gets all unique methods from scenarios
func (r *Runner) collectMethods() (map[string]MethodInfo, error) {
	methods := make(map[string]MethodInfo)

	for _, scenario := range r.config.Scenarios {
		// Handle batch requests
		if len(scenario.Batch) > 0 {
			for _, req := range scenario.Batch {
				method := req.GetMethod("")
				if method == "" {
					continue // Skip notifications without methods
				}
				
				info, err := ParseMethod(method)
				if err != nil {
					// Skip invalid methods (used for testing errors)
					continue
				}
				
				methods[method] = info
			}
			continue
		}
		
		// Handle single requests
		method := scenario.Request.GetMethod(scenario.Method)
		if method == "" {
			// Skip scenarios without methods (e.g., raw requests)
			continue
		}
		
		info, err := ParseMethod(method)
		if err != nil {
			// Skip invalid methods (used for testing errors)
			continue
		}
		
		methods[method] = info
	}

	return methods, nil
}

// startServers starts test servers for all transports
func (r *Runner) startServers(t *testing.T) error {
	t.Helper()
	
	ctx := context.Background()

	// Start a single server that handles all transports
	server, err := harness.StartServer(ctx, r.testDir, 0)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	r.servers["main"] = server
	
	// Update base URL if not set
	if r.runnerConfig.ServerURL == "" {
		r.runnerConfig.ServerURL = server.URL()
	}
	
	t.Logf("Server running at: %s", r.runnerConfig.ServerURL)
	return nil
}

// stopServers stops all test servers
func (r *Runner) stopServers() {
	for name, server := range r.servers {
		if err := server.Stop(); err != nil {
			// Log but don't fail tests
			fmt.Printf("Failed to stop server %q: %v\n", name, err)
		}
	}
	r.servers = make(map[string]*harness.Server)
}

// runScenario executes a single test scenario
func (r *Runner) runScenario(t *testing.T, scenario Scenario) {
	t.Helper()
	
	// Get server URL
	serverURL := r.runnerConfig.ServerURL
	if serverURL == "" && r.config.Settings.BaseURL != "" {
		serverURL = r.config.Settings.BaseURL
	}
	if serverURL == "" {
		t.Fatal("No server URL configured")
	}

	// Create executor with timeout from settings
	opts := []ExecutorOption{}
	if r.config.Settings.Timeout > 0 {
		opts = append(opts, WithWebSocketTimeout(r.config.Settings.Timeout))
	}
	
	var executor *Executor
	if r.executorFactory != nil {
		executor = r.executorFactory(serverURL)
	} else {
		executor = NewExecutor(serverURL, opts...)
	}
	executor.Execute(t, scenario)
}


// Cleanup performs any necessary cleanup
func (r *Runner) Cleanup() {
	r.stopServers()
	
	if !r.runnerConfig.KeepGenerated && r.testDir != "" {
		os.RemoveAll(r.testDir)
	}
}