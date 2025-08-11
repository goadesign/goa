package harness

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Server manages a test server process
type Server struct {
	cmd     *exec.Cmd
	port    int
	logFile *os.File
	workDir string
}

// StartServer starts a test server
func StartServer(ctx context.Context, workDir string, port int) (*Server, error) {
	// Find available port if not specified
	if port == 0 {
		listener, err := net.Listen("tcp", ":0")
		if err != nil {
			return nil, fmt.Errorf("failed to find free port: %w", err)
		}
		port = listener.Addr().(*net.TCPAddr).Port
		listener.Close() //nolint:errcheck
	}

	// Create log file
	logPath := filepath.Join(workDir, fmt.Sprintf("server-%d.log", port))
	logFile, err := os.Create(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	// Build server command - look for the generated server
	// Try multiple possible locations
	var serverPath string
	candidates := []string{
		filepath.Join(workDir, "cmd", "test_api", "main.go"),
		filepath.Join(workDir, "cmd", "test", "main.go"),
		filepath.Join(workDir, "cmd", "api", "main.go"),
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			serverPath = path
			break
		}
	}

	if serverPath == "" {
		return nil, fmt.Errorf("server main.go not found in any expected location")
	}

	// Ensure dependencies are downloaded before running
	downloadCmd := exec.Command("go", "mod", "download")
	downloadCmd.Dir = workDir
	downloadCmd.Env = append(os.Environ(),
		"GO111MODULE=on",
		"GOWORK=off",
	)
	if output, err := downloadCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("go mod download failed: %w\nOutput: %s", err, output)
	}

	// Run all files in the package so helpers generated alongside main.go are included
	serverDir := filepath.Dir(serverPath)
	cmd := exec.CommandContext(ctx, "go", "run", ".", "--http-port", fmt.Sprintf("%d", port))
	cmd.Dir = serverDir
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Env = append(os.Environ(),
		"GO111MODULE=on",
		"GOWORK=off",
	)

	// Start server
	if err := cmd.Start(); err != nil {
		logFile.Close() //nolint:errcheck
		return nil, fmt.Errorf("failed to start server: %w", err)
	}

	server := &Server{
		cmd:     cmd,
		port:    port,
		logFile: logFile,
		workDir: workDir,
	}

	// Wait for server to be ready
	if err := server.waitForReady(ctx); err != nil {
		// Read log file for diagnostics
		logFile.Seek(0, 0) //nolint:errcheck
		logContent, _ := bufio.NewReader(logFile).ReadString('\x00')
		server.Stop() //nolint:errcheck
		return nil, fmt.Errorf("%w\nServer log:\n%s", err, logContent)
	}

	return server, nil
}

// URL returns the server's base URL
func (s *Server) URL() string {
	return fmt.Sprintf("http://localhost:%d", s.port)
}

// Stop stops the server
func (s *Server) Stop() error {
	if s.cmd != nil && s.cmd.Process != nil {
		s.cmd.Process.Kill() //nolint:errcheck
		s.cmd.Wait()         //nolint:errcheck
	}

	if s.logFile != nil {
		s.logFile.Close() //nolint:errcheck
	}

	return nil
}

// waitForReady waits for the server to be ready to accept connections
func (s *Server) waitForReady(ctx context.Context) error {
	// Watch log for ready message
	logScanner := bufio.NewScanner(s.logFile)
	readyChan := make(chan bool)
	errChan := make(chan error)

	go func() {
		for logScanner.Scan() {
			line := logScanner.Text()
			if strings.Contains(line, "HTTP server listening") ||
				strings.Contains(line, fmt.Sprintf(":%d", s.port)) {
				readyChan <- true
				return
			}
			if strings.Contains(line, "error") || strings.Contains(line, "failed") {
				errChan <- fmt.Errorf("server error: %s", line)
				return
			}
		}
		if err := logScanner.Err(); err != nil {
			errChan <- fmt.Errorf("log scan error: %w", err)
		}
	}()

	// Also try connecting
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", s.port))
				if err == nil {
					conn.Close() //nolint:errcheck
					readyChan <- true
					return
				}
			}
		}
	}()

	// Wait for ready or timeout
	select {
	case <-readyChan:
		return nil
	case err := <-errChan:
		return err
	case <-time.After(10 * time.Second):
		return fmt.Errorf("server failed to start within 10 seconds")
	case <-ctx.Done():
		return ctx.Err()
	}
}
