package harness

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CLIClient wraps the generated CLI client for testing
type CLIClient struct {
	cliPath   string
	serverURL string
}

// NewCLIClient creates a new CLI client wrapper
func NewCLIClient(workDir, serverURL string) (*CLIClient, error) {
	// Find the CLI source directory
	candidates := []string{
		filepath.Join(workDir, "cmd", "test_api-cli"),
		filepath.Join(workDir, "cmd", "test-cli"),
		filepath.Join(workDir, "cmd", "api-cli"),
	}

	var cliPath string
	for _, path := range candidates {
		mainFile := filepath.Join(path, "main.go")
		if _, err := os.Stat(mainFile); err == nil {
			cliPath = path
			break
		}
	}

	if cliPath == "" {
		return nil, fmt.Errorf("CLI source not found in %s", workDir)
	}

	return &CLIClient{
		cliPath:   cliPath,
		serverURL: serverURL,
	}, nil
}

// CallMethod invokes a service method via the CLI
func (c *CLIClient) CallMethod(ctx context.Context, service, method string, payload any) (json.RawMessage, error) {
	// Convert method name from snake_case to kebab-case for CLI
	cliMethod := strings.ReplaceAll(method, "_", "-")

	// Build command arguments - use go run to execute the CLI
	// URL must come before service and method for proper flag parsing
	args := []string{
		"run", ".",
		"-url", c.serverURL,
		"-verbose",
		service,
		cliMethod,
	}

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = c.cliPath

	// Add payload if provided
	// Note: Generate methods don't take payload but the scenario might still
	// have params: {} which results in an empty map
	if payload != nil {
		// Skip empty maps for methods without payloads (like generate_*)
		if m, ok := payload.(map[string]any); ok && len(m) == 0 && strings.Contains(method, "generate_") {
			// Don't add any payload for empty maps on generate methods
		} else if str, ok := payload.(string); ok {
			// Check if payload is a simple string - use -p flag
			args = append(args, "-p", str)
		} else {
			// For any other payload, use --body flag
			payloadJSON, err := json.Marshal(payload)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal payload: %w", err)
			}

			args = append(args, "--body", string(payloadJSON))
		}
	}
	// If payload is nil, don't add any body argument - let the CLI handle it

	// Create command with all args
	cmd = exec.CommandContext(ctx, "go", args...)
	cmd.Dir = c.cliPath

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run command
	err := cmd.Run()

	// Check for errors
	if err != nil {
		errMsg := stderr.String()
		if errMsg == "" {
			errMsg = err.Error()
		}
		// Include both stdout and stderr for debugging
		return nil, fmt.Errorf("CLI command failed: %s\nStderr: %s\nStdout: %s", errMsg, stderr.String(), stdout.String())
	}

	// Parse verbose output from stderr to get the raw JSON-RPC response
	verboseOutput := stderr.String()
	lines := strings.Split(verboseOutput, "\n")

	// Find the JSON-RPC response - it's the last line starting with {
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "{") {
			var resp struct {
				Result json.RawMessage `json:"result"`
				Error  *struct {
					Code    int    `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			}

			if err := json.Unmarshal([]byte(line), &resp); err == nil && resp.Result != nil {
				return resp.Result, nil
			}
		}
	}

	// Fallback to stdout
	output := stdout.Bytes()
	if len(output) == 0 {
		return nil, nil
	}
	return json.RawMessage(output), nil
}

// CallJSONRPC makes a raw JSON-RPC call via the CLI
func (c *CLIClient) CallJSONRPC(ctx context.Context, request map[string]any) (json.RawMessage, error) {
	// For JSON-RPC, we need to use the jsonrpc command if available
	// Otherwise fall back to method call

	method, ok := request["method"].(string)
	if !ok {
		return nil, fmt.Errorf("no method in request")
	}

	// Extract service and method from JSON-RPC method name
	// Assuming format: service.method or just method
	parts := strings.Split(method, ".")
	service := "test" // default service
	methodName := method

	if len(parts) == 2 {
		service = parts[0]
		methodName = parts[1]
	}

	// Use the CLI to call the method
	return c.CallMethod(ctx, service, methodName, request["params"])
}

// CanHandle returns true if the CLI can handle this method
func (c *CLIClient) CanHandle(method string, params any) bool {
	// CLI can handle unary HTTP methods but not SSE streams.
	if strings.Contains(method, "_sse") {
		return false
	}

	// CLI doesn't handle notification methods (no response expected)
	if strings.Contains(method, "_notify") {
		return false
	}

	// CLI can handle methods with payloads
	return true
}
