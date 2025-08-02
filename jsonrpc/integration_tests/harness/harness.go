package harness

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// TestHarness orchestrates integration test execution, managing the complete
// lifecycle of test resources including process management, port allocation,
// and cleanup. It ensures all resources are properly released even if the
// test panics or the process is interrupted.
//
// A typical test creates a harness, generates code from DSL, starts a server,
// executes client requests, and validates responses. The harness automatically
// handles cleanup when the test completes or fails.
type TestHarness struct {
	t           testing.TB
	baseDir     string
	servers     map[string]*ServerProcess
	clients     map[string]*ClientProcess
	cleanup     []func() error
	cleanupOnce sync.Once
	mu          sync.Mutex

	// Port management
	portAllocator *PortAllocator

	// Cleanup tracking
	cleanupDone chan struct{}

	// DSL loader for loading DSL files
	dslLoader *DSLLoader

	// Code cache for reusing generated code
	codeCache *CodeCache
}

// New creates a new test harness for the given test. The harness automatically
// registers cleanup handlers that will run when the test completes, ensuring
// all temporary files and processes are properly cleaned up.
//
// The harness creates an isolated directory for test artifacts and sets up
// signal handlers to clean up resources if the process is interrupted.
func New(t testing.TB) *TestHarness {
	baseDir := createTestDir(t)

	// Default DSL directory is relative to the test directory
	dslDir := filepath.Join(filepath.Dir(baseDir), "..", "testdata", "dsls")

	// Create code cache
	codeCache, err := NewCodeCache(baseDir)
	if err != nil {
		t.Fatalf("Failed to create code cache: %v", err)
	}

	h := &TestHarness{
		t:             t,
		baseDir:       baseDir,
		servers:       make(map[string]*ServerProcess),
		clients:       make(map[string]*ClientProcess),
		cleanup:       []func() error{},
		cleanupDone:   make(chan struct{}),
		portAllocator: NewPortAllocator(),
		dslLoader:     NewDSLLoader(dslDir),
		codeCache:     codeCache,
	}

	// Register cleanup immediately (unless debugging)
	if os.Getenv("KEEP_ARTIFACTS") != "1" {
		t.Cleanup(h.Cleanup)
	}

	// Also handle signals for cleanup
	h.registerSignalHandlers()

	// Add base directory cleanup
	h.addCleanup(func() error {
		return os.RemoveAll(baseDir)
	})

	return h
}

// BaseDir returns the base directory for this test run
func (h *TestHarness) BaseDir() string {
	return h.baseDir
}

// AllocatePort returns a free port for use in tests
func (h *TestHarness) AllocatePort() (int, error) {
	// Since tests run sequentially, use OS-allocated ports
	return GetFreePort()
}

// StartServer compiles the generated server code and starts it as a subprocess.
// The server is assigned a dynamic port (or uses the port specified in config)
// and is tracked by the harness for automatic cleanup.
//
// The method waits for the server to be ready before returning, using the
// ReadyString in the config to detect when startup is complete. If the server
// fails to start within the timeout, an error is returned and the process is
// terminated.
func (h *TestHarness) StartServer(ctx context.Context, name string, config ServerConfig) (*ServerProcess, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Check if server already exists
	if srv, exists := h.servers[name]; exists {
		return srv, fmt.Errorf("server %s already running", name)
	}

	// Apply any service implementations before compiling
	if len(config.ServiceImplementations) > 0 {
		// The service implementation files are in the parent of cmd/[server]
		// config.SourceDir is like generated/sse_primitive_result/cmd/test
		// We need generated/sse_primitive_result
		genDir := filepath.Dir(filepath.Dir(config.SourceDir))
		for _, impl := range config.ServiceImplementations {
			if err := h.InjectServiceImplementation(genDir, impl.ServiceName, impl.MethodName, impl.Implementation); err != nil {
				return nil, fmt.Errorf("failed to inject implementation: %w", err)
			}
		}
	}

	// Create server directory
	serverDir := filepath.Join(h.baseDir, "servers", name)
	if err := os.MkdirAll(serverDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create server directory: %w", err)
	}

	// Start server
	srv, err := StartServer(ctx, serverDir, config)
	if err != nil {
		return nil, fmt.Errorf("failed to start server %s: %w", name, err)
	}

	// Track server
	h.servers[name] = srv

	// Add cleanup (use locked version since we already hold the mutex)
	h.addCleanupLocked(func() error {
		return srv.Stop()
	})

	return srv, nil
}

// StartClient creates a client process for testing
func (h *TestHarness) StartClient(name string, config ClientConfig) (*ClientProcess, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Create client directory
	clientDir := filepath.Join(h.baseDir, "clients", name)
	if err := os.MkdirAll(clientDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create client directory: %w", err)
	}

	// Create client
	client, err := NewClient(clientDir, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create client %s: %w", name, err)
	}

	// Track client
	h.clients[name] = client

	return client, nil
}

// GenerateCode generates server and client code from a DSL string in an isolated
// directory. The DSL string should define a complete Goa service with JSON-RPC
// endpoints.
//
// The generated code includes both server implementation and client libraries,
// ready for compilation. The method returns the absolute path to the generated
// code directory.
func (h *TestHarness) GenerateCode(ctx context.Context, name string, dslCode string) (string, error) {
	debug := os.Getenv("DEBUG_TESTS") == "1"
	if debug {
		fmt.Printf("[HARNESS] GenerateCode called for %s at %s\n", name, time.Now().Format("15:04:05.000"))
		defer func(start time.Time) {
			fmt.Printf("[HARNESS] GenerateCode completed for %s in %v\n", name, time.Since(start))
		}(time.Now())
	}

	// Check cache first
	if cachedDir, ok := h.codeCache.Get(dslCode); ok {
		h.t.Logf("Using cached generated code for %s", name)
		return cachedDir, nil
	}

	genDir := filepath.Join(h.baseDir, "generated", name)

	// Get absolute path first to avoid confusion
	absGenDir, err := filepath.Abs(genDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	if err := os.MkdirAll(absGenDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create generation directory: %w", err)
	}

	// Generate code using the DSL string
	if err := GenerateFromDSL(ctx, absGenDir, dslCode); err != nil {
		return "", fmt.Errorf("code generation failed: %w", err)
	}

	// No automatic injection - tests will provide implementations

	// Cache the generated code
	if err := h.codeCache.Put(dslCode, absGenDir); err != nil {
		// Log but don't fail on cache errors
		h.t.Logf("Failed to cache generated code: %v", err)
	}

	return absGenDir, nil
}

// GenerateCodeFromFile generates code from a DSL file
func (h *TestHarness) GenerateCodeFromFile(ctx context.Context, name string, dslFile string) (string, error) {
	// Load DSL from file
	dslCode, err := h.dslLoader.Load(dslFile)
	if err != nil {
		return "", fmt.Errorf("failed to load DSL file: %w", err)
	}

	return h.GenerateCode(ctx, name, dslCode)
}

// Cleanup performs all cleanup operations with a timeout
func (h *TestHarness) Cleanup() {
	h.cleanupOnce.Do(func() {
		// Use a goroutine with timeout to ensure cleanup doesn't hang
		done := make(chan struct{})
		go func() {
			h.mu.Lock()
			cleanupFuncs := h.cleanup
			h.mu.Unlock()

			// Execute cleanup in reverse order
			for i := len(cleanupFuncs) - 1; i >= 0; i-- {
				if err := cleanupFuncs[i](); err != nil {
					h.t.Logf("cleanup error: %v", err)
				}
			}
			close(done)
		}()

		// Wait for cleanup with timeout
		select {
		case <-done:
			// Cleanup completed
		case <-time.After(1 * time.Second):
			h.t.Logf("cleanup timeout - forcing completion")
		}

		// Signal cleanup done
		close(h.cleanupDone)
	})
}

// addCleanup registers a cleanup function
func (h *TestHarness) addCleanup(fn func() error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cleanup = append(h.cleanup, fn)
}

// addCleanupLocked registers a cleanup function when the mutex is already held
func (h *TestHarness) addCleanupLocked(fn func() error) {
	h.cleanup = append(h.cleanup, fn)
}

// registerSignalHandlers sets up signal handlers for cleanup
func (h *TestHarness) registerSignalHandlers() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		select {
		case <-sigChan:
			h.t.Logf("received interrupt signal, cleaning up...")
			h.Cleanup()
			os.Exit(1)
		case <-h.cleanupDone:
			// Normal cleanup completed
		}
	}()
}

// createTestDir creates a unique test directory
func createTestDir(t testing.TB) string {
	// Create a unique directory for this test run
	timestamp := time.Now().Format("20060102_150405")
	testName := sanitizeTestName(t.Name())

	// Get the integration test root directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	// Find the integration_tests directory
	integrationRoot := cwd
	for !strings.HasSuffix(integrationRoot, "integration_tests") {
		parent := filepath.Dir(integrationRoot)
		if parent == integrationRoot {
			// Reached root without finding integration_tests
			t.Fatalf("could not find integration_tests directory from %s", cwd)
		}
		integrationRoot = parent
	}

	baseDir := filepath.Join(
		integrationRoot,
		"tests",
		"testdata",
		"runs",
		fmt.Sprintf("%s_%s", timestamp, testName),
	)

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	return baseDir
}

// InjectServiceImplementation replaces a generated service method implementation
// with a test-specific implementation provided by the test.
func (h *TestHarness) InjectServiceImplementation(genDir, serviceName, methodName, implementation string) error {
	// The generated service implementation file has a predictable path
	serviceFile := filepath.Join(genDir, serviceName+".go")

	// Read the file
	content, err := os.ReadFile(serviceFile)
	if err != nil {
		return fmt.Errorf("failed to read service file: %w", err)
	}

	// We're looking for the generated method implementation to replace it
	// The pattern in generated code is:
	// // MethodName implements methodname.
	// func (s *servicenamesrvc) MethodName(...) ... {
	//     log.Printf(ctx, "servicename.methodname")
	//     return
	// }

	contentStr := string(content)

	// Find the method by looking for the log.Printf line which is unique
	// The pattern in generated code can be either "servicename.methodname" or "servicename_.methodname"
	logPattern := fmt.Sprintf(`log.Printf(ctx, "%s.%s")`, serviceName, methodName)
	logIdx := strings.Index(contentStr, logPattern)
	if logIdx == -1 {
		// Try with underscore in the log pattern (e.g., "errors_.test_error")
		if strings.HasSuffix(serviceName, "_") {
			logPattern = fmt.Sprintf(`log.Printf(ctx, "%s.%s")`, serviceName, methodName)
		} else {
			logPattern = fmt.Sprintf(`log.Printf(ctx, "%s_.%s")`, serviceName, methodName)
		}
		logIdx = strings.Index(contentStr, logPattern)
		if logIdx == -1 {
			// Try without underscore in service name
			logPattern = fmt.Sprintf(`log.Printf(ctx, "%s.%s")`, strings.TrimSuffix(serviceName, "_"), methodName)
			logIdx = strings.Index(contentStr, logPattern)
			if logIdx == -1 {
				return fmt.Errorf("could not find log statement for %s.%s", serviceName, methodName)
			}
		}
	}

	// Find the start of the method by searching backwards for "func"
	funcStart := strings.LastIndex(contentStr[:logIdx], "func")
	if funcStart == -1 {
		return fmt.Errorf("could not find function start for %s.%s", serviceName, methodName)
	}

	// Find the comment before the function
	commentEnd := funcStart - 1
	for commentEnd > 0 && (contentStr[commentEnd] == '\n' || contentStr[commentEnd] == '\t' || contentStr[commentEnd] == ' ') {
		commentEnd--
	}
	commentStart := strings.LastIndex(contentStr[:commentEnd], "\n") + 1
	if commentStart == 0 {
		commentStart = 0
	}

	// Find the end of the method by finding the matching closing brace
	// First, find the opening brace
	braceIdx := strings.Index(contentStr[funcStart:], "{")
	if braceIdx == -1 {
		return fmt.Errorf("could not find opening brace for %s.%s", serviceName, methodName)
	}
	braceIdx += funcStart

	// Count braces to find the matching closing brace
	braceCount := 1
	endIdx := braceIdx + 1
	inString := false
	escaped := false

	for endIdx < len(contentStr) && braceCount > 0 {
		ch := contentStr[endIdx]

		// Handle string literals to avoid counting braces inside strings
		if ch == '\\' && !escaped {
			escaped = true
			endIdx++
			continue
		}

		if ch == '"' && !escaped {
			inString = !inString
		}

		if !inString && !escaped {
			if ch == '{' {
				braceCount++
			} else if ch == '}' {
				braceCount--
			}
		}

		escaped = false
		endIdx++
	}

	if braceCount != 0 {
		return fmt.Errorf("could not find matching closing brace for %s.%s", serviceName, methodName)
	}

	// Replace the entire method (including comment)
	newContent := contentStr[:commentStart] + implementation + contentStr[endIdx:]

	// Add required imports if they're used in the implementation
	if strings.Contains(implementation, "fmt.") && !strings.Contains(newContent, `"fmt"`) {
		newContent = strings.Replace(newContent, "import (", "import (\n\t\"fmt\"", 1)
	}
	if strings.Contains(implementation, "time.") && !strings.Contains(newContent, `"time"`) {
		newContent = strings.Replace(newContent, "import (", "import (\n\t\"time\"", 1)
	}
	if strings.Contains(implementation, "goa.") && !strings.Contains(newContent, `goa "goa.design/goa/v3/pkg"`) {
		newContent = strings.Replace(newContent, "import (", "import (\n\tgoa \"goa.design/goa/v3/pkg\"", 1)
	}
	if strings.Contains(implementation, "io.EOF") && !strings.Contains(newContent, `"io"`) {
		newContent = strings.Replace(newContent, "import (", "import (\n\t\"io\"", 1)
	}

	// Write back the modified content
	if err := os.WriteFile(serviceFile, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	return nil
}

// sanitizeTestName makes a test name safe for use as a directory name
func sanitizeTestName(name string) string {
	// Replace problematic characters
	sanitized := name
	for _, char := range []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", " "} {
		sanitized = replaceAll(sanitized, char, "_")
	}

	// Limit length
	if len(sanitized) > 50 {
		sanitized = sanitized[:50]
	}

	return sanitized
}

// replaceAll replaces all occurrences of old with new in s
func replaceAll(s, old, new string) string {
	result := s
	for {
		index := indexOf(result, old)
		if index == -1 {
			break
		}
		result = result[:index] + new + result[index+len(old):]
	}
	return result
}

// indexOf returns the index of substr in s, or -1 if not found
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
