package harness

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ServiceImplementation describes a service method implementation to inject
type ServiceImplementation struct {
	ServiceName    string // e.g., "events"
	MethodName     string // e.g., "subscribe"
	Implementation string // The complete method implementation
}

// ServerConfig contains configuration for starting a test server
type ServerConfig struct {
	// SourceDir is the directory containing the generated server code
	SourceDir string
	
	// Port is the port to listen on (0 for dynamic allocation)
	Port int
	
	// StartupTimeout is how long to wait for the server to start
	StartupTimeout time.Duration
	
	// ReadyString is the log output that indicates the server is ready
	ReadyString string
	
	// Env contains additional environment variables
	Env map[string]string
	
	// ServiceImplementations contains test implementations to inject
	ServiceImplementations []ServiceImplementation
}

// ServerProcess represents a running server process with improved error handling
type ServerProcess struct {
	cmd        *exec.Cmd
	port       int
	logFile    *os.File
	ctx        context.Context
	cancel     context.CancelFunc
	ready      chan struct{}
	failed     chan error
	readyOnce  sync.Once
	stopOnce   sync.Once
	mu         sync.Mutex
	stopped    bool
}

// StartServer compiles the server code from the specified source directory and
// starts it as a managed subprocess with improved error detection.
func StartServer(ctx context.Context, workDir string, config ServerConfig) (*ServerProcess, error) {
	debug := os.Getenv("DEBUG_TESTS") == "1"
	if debug {
		fmt.Printf("[HARNESS] StartServer called for %s on port %d at %s\n", 
			config.SourceDir, config.Port, time.Now().Format("15:04:05.000"))
	}

	// Set defaults - use aggressive timeouts for quick failure
	if config.StartupTimeout == 0 {
		config.StartupTimeout = 2 * time.Second
	}
	if config.ReadyString == "" {
		config.ReadyString = "listening"
	}

	// Create a build context with timeout
	buildCtx, buildCancel := context.WithTimeout(ctx, 30*time.Second)
	defer buildCancel()

	// Build the server
	binaryPath := filepath.Join(workDir, "server")
	if err := buildBinary(buildCtx, config.SourceDir, binaryPath); err != nil {
		return nil, fmt.Errorf("failed to build server: %w", err)
	}

	// Create log file
	logFile, err := os.Create(filepath.Join(workDir, "server.log"))
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	// Create server context
	serverCtx, serverCancel := context.WithCancel(ctx)

	// Create command
	args := []string{"-http-port", fmt.Sprintf("%d", config.Port)}
	cmd := exec.CommandContext(serverCtx, binaryPath, args...)
	cmd.Dir = workDir

	// Set environment
	cmd.Env = os.Environ()
	for k, v := range config.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	if config.Port > 0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("PORT=%d", config.Port))
	}

	// Capture output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		serverCancel()
		logFile.Close()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		serverCancel()
		logFile.Close()
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the server
	if err := cmd.Start(); err != nil {
		serverCancel()
		logFile.Close()
		return nil, fmt.Errorf("failed to start server: %w", err)
	}

	srv := &ServerProcess{
		cmd:     cmd,
		port:    config.Port,
		logFile: logFile,
		ctx:     serverCtx,
		cancel:  serverCancel,
		ready:   make(chan struct{}),
		failed:  make(chan error, 1),
	}

	// Monitor output with better error detection
	go srv.monitorOutput(stdout, stderr, config.ReadyString)

	// Monitor process exit
	go srv.monitorExit()

	// Wait for server to be ready or fail
	startupCtx, startupCancel := context.WithTimeout(ctx, config.StartupTimeout)
	defer startupCancel()

	if debug {
		fmt.Printf("[HARNESS] Waiting for server to be ready (timeout: %v)...\n", config.StartupTimeout)
	}

	select {
	case <-srv.ready:
		// Give the server a moment to fully bind to the port
		time.Sleep(100 * time.Millisecond)
		if debug {
			fmt.Printf("[HARNESS] Server ready on port %d\n", config.Port)
		}
		return srv, nil

	case err := <-srv.failed:
		// Server failed to start
		srv.cleanup()
		logContent, _ := os.ReadFile(logFile.Name())
		if debug {
			fmt.Printf("[HARNESS] Server failed to start: %v\n", err)
		}
		return nil, fmt.Errorf("server failed: %w\nServer output:\n%s", err, string(logContent))

	case <-startupCtx.Done():
		// Startup timeout
		srv.Stop()
		logContent, _ := os.ReadFile(logFile.Name())
		if debug {
			fmt.Printf("[HARNESS] Server startup timeout after %v\n", config.StartupTimeout)
		}
		return nil, fmt.Errorf("server startup timeout after %v\nServer output:\n%s", 
			config.StartupTimeout, string(logContent))

	case <-ctx.Done():
		// Parent context canceled
		srv.Stop()
		if debug {
			fmt.Printf("[HARNESS] Server startup canceled: %v\n", ctx.Err())
		}
		return nil, fmt.Errorf("server startup canceled: %w", ctx.Err())
	}
}

// monitorOutput monitors server output with better ready detection and error handling
func (s *ServerProcess) monitorOutput(stdout, stderr io.Reader, readyString string) {
	var wg sync.WaitGroup
	wg.Add(2)

	// Monitor stdout
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Fprintln(s.logFile, line)
			
			// Check for ready string
			if strings.Contains(line, readyString) {
				s.signalReady()
			}
			
			// Extract port if dynamically allocated
			if s.port == 0 && strings.Contains(line, "listening") {
				if port := extractPort(line); port > 0 {
					s.mu.Lock()
					s.port = port
					s.mu.Unlock()
				}
			}
		}
	}()

	// Monitor stderr
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Fprintln(s.logFile, "[ERROR] "+line)
			
			// Detect common startup failures for quick failure
			if isStartupError(line) {
				s.failed <- fmt.Errorf("startup error: %s", line)
			}
		}
	}()

	wg.Wait()
}

// monitorExit monitors process exit and signals failure if it exits unexpectedly
func (s *ServerProcess) monitorExit() {
	err := s.cmd.Wait()
	
	s.mu.Lock()
	stopped := s.stopped
	s.mu.Unlock()
	
	if !stopped {
		// Process exited unexpectedly
		if err != nil {
			s.failed <- fmt.Errorf("process exited with error: %w", err)
		} else {
			s.failed <- fmt.Errorf("process exited unexpectedly")
		}
	}
}

// signalReady signals that the server is ready
func (s *ServerProcess) signalReady() {
	s.readyOnce.Do(func() {
		close(s.ready)
	})
}

// Port returns the port the server is listening on
func (s *ServerProcess) Port() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.port
}

// URL returns the base URL for the server
func (s *ServerProcess) URL() string {
	return fmt.Sprintf("http://localhost:%d", s.Port())
}

// Stop stops the server process immediately
func (s *ServerProcess) Stop() error {
	var err error
	s.stopOnce.Do(func() {
		s.mu.Lock()
		s.stopped = true
		s.mu.Unlock()
		
		// Cancel context first
		s.cancel()
		
		// Kill process immediately - don't wait
		if s.cmd.Process != nil {
			if killErr := s.cmd.Process.Kill(); killErr != nil {
				if !strings.Contains(killErr.Error(), "process already") {
					err = killErr
				}
			}
		}
		
		// Clean up resources immediately without waiting
		s.cleanup()
	})
	return err
}

// cleanup closes resources
func (s *ServerProcess) cleanup() {
	if s.logFile != nil {
		s.logFile.Close()
	}
}

// extractPort extracts port number from log line
func extractPort(line string) int {
	// Look for patterns like ":8080" or "port 8080"
	patterns := []string{
		`:(\d+)`,
		`port\s+(\d+)`,
		`listening\s+on\s+.*:(\d+)`,
	}
	
	for _, pattern := range patterns {
		if matches := regexp.MustCompile(pattern).FindStringSubmatch(line); len(matches) > 1 {
			if port, err := strconv.Atoi(matches[1]); err == nil {
				return port
			}
		}
	}
	
	return 0
}

// isStartupError checks if a log line indicates a startup error
func isStartupError(line string) bool {
	errorPatterns := []string{
		"bind: address already in use",
		"permission denied",
		"no such file or directory",
		"panic:",
		"fatal:",
		"cannot",
		"failed to",
		"error:",
	}
	
	lowerLine := strings.ToLower(line)
	for _, pattern := range errorPatterns {
		if strings.Contains(lowerLine, pattern) {
			return true
		}
	}
	
	return false
}