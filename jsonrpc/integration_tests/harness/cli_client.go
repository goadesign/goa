package harness

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	// Find the CLI binary
	candidates := []string{
		filepath.Join(workDir, "cmd", "test_api-cli", "test_api-cli"),
		filepath.Join(workDir, "cmd", "test-cli", "test-cli"),
		filepath.Join(workDir, "cmd", "api-cli", "api-cli"),
	}
	
	var cliPath string
	for _, path := range candidates {
		if _, err := exec.LookPath(path); err == nil {
			cliPath = path
			break
		}
	}
	
	if cliPath == "" {
		return nil, fmt.Errorf("CLI binary not found in %s", workDir)
	}
	
	return &CLIClient{
		cliPath:   cliPath,
		serverURL: serverURL,
	}, nil
}

// CallMethod invokes a service method via the CLI
func (c *CLIClient) CallMethod(ctx context.Context, service, method string, payload interface{}) (json.RawMessage, error) {
	// Build command arguments
	args := []string{
		service,
		method,
		"--url", c.serverURL,
	}
	
	cmd := exec.CommandContext(ctx, c.cliPath, args...)
	
	// Add payload if provided
	if payload != nil {
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		
		// The CLI expects payload as positional argument or via stdin
		// Let's use stdin for complex payloads
		cmd.Args = append(cmd.Args, "--payload", "-")
		cmd.Stdin = bytes.NewReader(payloadJSON)
	}
	
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
		return nil, fmt.Errorf("CLI command failed: %s", errMsg)
	}
	
	// Parse output - the CLI returns the result as JSON
	output := stdout.Bytes()
	if len(output) == 0 {
		return nil, nil
	}
	
	return json.RawMessage(output), nil
}

// CallJSONRPC makes a raw JSON-RPC call via the CLI
func (c *CLIClient) CallJSONRPC(ctx context.Context, request map[string]interface{}) (json.RawMessage, error) {
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