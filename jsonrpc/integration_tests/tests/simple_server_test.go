package tests

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"goa.design/goa/v3/jsonrpc/integration_tests/harness"
)

func TestSimpleServerStartup(t *testing.T) {
	h := harness.New(t)

	// Simple DSL
	simpleDSLCode := `	API("test", func() {
		Title("Test API")
	})
	
	Service("test", func() {
		JSONRPC(func() {
			POST("/jsonrpc")
		})
		Method("ping", func() {
			Result(String)
			JSONRPC(func() {
			})
		})
	})`

	// Generate code
	genDir, err := h.GenerateCode(context.Background(), "simple_server", simpleDSLCode)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	// Allocate port
	port, err := h.AllocatePort()
	if err != nil {
		t.Fatalf("Failed to allocate port: %v", err)
	}

	// Start server - the server is in cmd/test/
	serverConfig := harness.ServerConfig{
		SourceDir:      genDir + "/cmd/test",
		Port:           port,
		StartupTimeout: 2 * time.Second,
		ReadyString:    "HTTP server listening",
	}

	server, err := h.StartServer(context.Background(), "simple_server", serverConfig)
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Try to access the server with HTTP
	resp, err := http.Get(server.URL() + "/")
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != 404 {
		t.Fatalf("Expected 404 (Not Found) for GET on root path, got %d", resp.StatusCode)
	}

	// Now try the JSON-RPC endpoint with an undefined method
	client := &http.Client{Timeout: 5 * time.Second}

	jsonReq := `{"jsonrpc":"2.0","method":"undefined_method","id":1}`
	resp2, err := client.Post(server.URL()+"/jsonrpc", "application/json",
		strings.NewReader(jsonReq))
	if err != nil {
		t.Fatalf("Failed to call JSON-RPC: %v", err)
	}
	defer resp2.Body.Close() //nolint:errcheck

	body, _ := io.ReadAll(resp2.Body)

	// We expect a method not found error since "undefined_method" is not defined in the DSL
	if !strings.Contains(string(body), "-32601") {
		t.Fatalf("Expected method not found error, got: %s", body)
	}
}
