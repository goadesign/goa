package harness

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC *string `json:"jsonrpc,omitempty"` // Pointer to allow omitting the field entirely
	Method  string  `json:"method"`
	Params  any     `json:"params,omitempty"`
	ID      any     `json:"id,omitempty"`
}

// Default values
const (
	DefaultHTTPTimeout = 10 * time.Second
	DefaultJSONRPCPath = "/jsonrpc"
	DefaultSSEPath     = "/jsonrpc/sse"
)

// ClientConfig holds client configuration
type ClientConfig struct {
	// HTTPTimeout is the timeout for HTTP requests
	HTTPTimeout time.Duration
	// JSONRPCPath is the path for JSON-RPC HTTP endpoint
	JSONRPCPath string
	// SSEPath is the path for SSE endpoint
	SSEPath string
	// Headers are additional headers to send with requests
	Headers map[string]string
	// HTTPClient allows using a custom HTTP client
	HTTPClient *http.Client
}

// DefaultConfig returns default client configuration
func DefaultConfig() *ClientConfig {
	return &ClientConfig{
		HTTPTimeout: DefaultHTTPTimeout,
		JSONRPCPath: DefaultJSONRPCPath,
		SSEPath:     DefaultSSEPath,
		Headers:     make(map[string]string),
	}
}

// Client provides JSON-RPC client functionality for all transports
type Client struct {
	baseURL    *url.URL
	config     *ClientConfig
	httpClient *http.Client
}

// NewClient creates a new JSON-RPC client
func NewClient(baseURL string, config *ClientConfig) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	if config == nil {
		config = DefaultConfig()
	}

	// Create HTTP client if not provided
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: config.HTTPTimeout,
		}
	}

	return &Client{
		baseURL:    u,
		config:     config,
		httpClient: httpClient,
	}, nil
}

// CallHTTPRaw makes a raw HTTP call with the given body
func (c *Client) CallHTTPRaw(ctx context.Context, body []byte) (json.RawMessage, error) {
	endpoint := c.baseURL.ResolveReference(&url.URL{Path: c.config.JSONRPCPath})
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	for k, v := range c.config.Headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// For error responses, still return the body
	if resp.StatusCode == http.StatusBadRequest {
		return json.RawMessage(respBody), nil
	}

	// Check status code
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	// For notifications, we expect no response body
	if len(respBody) == 0 {
		return nil, nil
	}

	return json.RawMessage(respBody), nil
}

// CallHTTP makes a JSON-RPC call over HTTP
func (c *Client) CallHTTP(ctx context.Context, req JSONRPCRequest) (json.RawMessage, error) {
	// Build JSON-RPC request envelope
	envelope := map[string]any{
		"method": req.Method,
	}

	// Add jsonrpc field if provided, or default to "2.0"
	if req.JSONRPC != nil {
		if *req.JSONRPC != "" {
			envelope["jsonrpc"] = *req.JSONRPC
		}
		// If JSONRPC is explicitly set to empty string, omit the field
	} else {
		// Default behavior: include "jsonrpc": "2.0"
		envelope["jsonrpc"] = "2.0"
	}

	if req.Params != nil {
		envelope["params"] = req.Params
	}
	if req.ID != nil {
		envelope["id"] = req.ID
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := c.baseURL.ResolveReference(&url.URL{Path: c.config.JSONRPCPath})
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint.String(), bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	for k, v := range c.config.Headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	// For notifications, we expect no response body
	if len(body) == 0 {
		return nil, nil
	}

	return json.RawMessage(body), nil
}

// CallSSE makes a JSON-RPC call over SSE and returns all events
func (c *Client) CallSSE(ctx context.Context, req JSONRPCRequest) ([]json.RawMessage, error) {
	// Build JSON-RPC request envelope
	envelope := map[string]any{}
	// Set jsonrpc per request: nil -> default to "2.0"; non-nil empty -> omit; otherwise use value
	if req.JSONRPC != nil {
		if *req.JSONRPC != "" {
			envelope["jsonrpc"] = *req.JSONRPC
		}
	} else {
		envelope["jsonrpc"] = "2.0"
	}
	// Method
	envelope["method"] = req.Method
	if req.Params != nil {
		envelope["params"] = req.Params
	}
	if req.ID != nil {
		envelope["id"] = req.ID
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := c.baseURL.ResolveReference(&url.URL{Path: c.config.SSEPath})
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint.String(), bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	for k, v := range c.config.Headers {
		httpReq.Header.Set(k, v)
	}

	// Use a client without timeout for SSE
	sseClient := &http.Client{Transport: c.httpClient.Transport}
	resp, err := sseClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	// Read response body for debug
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read SSE response: %w", err)
	}

	// Parse SSE events
	events, err := c.parseSSEEvents(bytes.NewReader(body))
	return events, err
}

// parseSSEEvents parses Server-Sent Events from a reader
func (c *Client) parseSSEEvents(r io.Reader) ([]json.RawMessage, error) {
	var events []json.RawMessage
	scanner := bufio.NewScanner(r)

	var eventData strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			// Empty line signals end of event
			if eventData.Len() > 0 {
				events = append(events, json.RawMessage(eventData.String()))
				eventData.Reset()
			}
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if eventData.Len() > 0 {
				eventData.WriteString("\n")
			}
			eventData.WriteString(data)
		}
		// Ignore other SSE fields like event:, id:, retry:
	}

	// Handle last event if no trailing empty line
	if eventData.Len() > 0 {
		events = append(events, json.RawMessage(eventData.String()))
	}

	return events, scanner.Err()
}
