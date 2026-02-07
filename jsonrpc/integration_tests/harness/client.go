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

	"github.com/gorilla/websocket"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC *string `json:"jsonrpc,omitempty"` // Pointer to allow omitting the field entirely
	Method  string  `json:"method"`
	Params  any     `json:"params,omitempty"`
	ID      any     `json:"id,omitempty"`
	HasID   bool    `json:"-"` // true if the id key must be included even if null
}

// Default values
const (
	DefaultHTTPTimeout = 10 * time.Second
	DefaultJSONRPCPath = "/jsonrpc"
	DefaultSSEPath     = "/jsonrpc/sse"
	DefaultWSPath      = "/jsonrpc/ws"
)

// ClientConfig holds client configuration
type ClientConfig struct {
	// HTTPTimeout is the timeout for HTTP requests
	HTTPTimeout time.Duration
	// JSONRPCPath is the path for JSON-RPC HTTP endpoint
	JSONRPCPath string
	// SSEPath is the path for SSE endpoint
	SSEPath string
	// WSPath is the path for WebSocket endpoint
	WSPath string
	// Headers are additional headers to send with requests
	Headers map[string]string
	// HTTPClient allows using a custom HTTP client
	HTTPClient *http.Client
	// WSDialer allows using a custom WebSocket dialer
	WSDialer *websocket.Dialer
}

// DefaultConfig returns default client configuration
func DefaultConfig() *ClientConfig {
	return &ClientConfig{
		HTTPTimeout: DefaultHTTPTimeout,
		JSONRPCPath: DefaultJSONRPCPath,
		SSEPath:     DefaultSSEPath,
		WSPath:      DefaultWSPath,
		Headers:     make(map[string]string),
	}
}

// Client provides JSON-RPC client functionality for all transports
type Client struct {
	baseURL    *url.URL
	config     *ClientConfig
	httpClient *http.Client
	wsDialer   *websocket.Dialer
	wsConn     *websocket.Conn
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

	// Create WebSocket dialer if not provided
	wsDialer := config.WSDialer
	if wsDialer == nil {
		// Do not use proxies from environment variables for integration tests.
		// A developer shell may have HTTP(S)_PROXY set, which breaks localhost
		// WebSocket upgrades and causes "websocket: bad handshake" failures.
		//
		// Gorilla uses ProxyFromEnvironment in websocket.DefaultDialer; setting
		// Proxy to nil disables proxy use entirely.
		defaultDialer := *websocket.DefaultDialer
		defaultDialer.Proxy = nil
		wsDialer = &defaultDialer
	}

	return &Client{
		baseURL:    u,
		config:     config,
		httpClient: httpClient,
		wsDialer:   wsDialer,
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

// ConnectWebSocket establishes a WebSocket connection
func (c *Client) ConnectWebSocket(ctx context.Context) error {
	// Build WebSocket URL
	wsURL := *c.baseURL
	wsURL.Path = c.config.WSPath

	// Convert scheme
	switch wsURL.Scheme {
	case "http":
		wsURL.Scheme = "ws"
	case "https":
		wsURL.Scheme = "wss"
	default:
		// Keep as is (might already be ws/wss)
	}

	// Set headers
	headers := http.Header{}
	for k, v := range c.config.Headers {
		headers.Set(k, v)
	}

	conn, resp, err := c.wsDialer.DialContext(ctx, wsURL.String(), headers)
	if err != nil {
		if resp != nil {
			if resp.Body == nil {
				return fmt.Errorf("websocket dial failed (status %d): %w", resp.StatusCode, err)
			}
			body, readErr := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if readErr != nil {
				return fmt.Errorf("websocket dial failed (status %d): %w", resp.StatusCode, err)
			}
			return fmt.Errorf("websocket dial failed (status %d): %w: %s", resp.StatusCode, err, string(body))
		}
		return fmt.Errorf("websocket dial failed: %w", err)
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close() //nolint:errcheck
	}

	c.wsConn = conn
	return nil
}

// SendWebSocket sends a JSON-RPC request over WebSocket
func (c *Client) SendWebSocket(ctx context.Context, req JSONRPCRequest) error {
	if c.wsConn == nil {
		return fmt.Errorf("websocket not connected")
	}

	// Build JSON-RPC request envelope
	envelope := map[string]any{}
	// Allow tests to omit the method by passing "-" (treated as missing field)
	if req.Method != "-" && req.Method != "" {
		envelope["method"] = req.Method
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
	// Preserve explicit id presence, even if null
	if req.HasID {
		envelope["id"] = req.ID
	} else if req.ID != nil {
		envelope["id"] = req.ID
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Set write deadline from context
	if deadline, ok := ctx.Deadline(); ok {
		if err := c.wsConn.SetWriteDeadline(deadline); err != nil {
			return fmt.Errorf("failed to set write deadline: %w", err)
		}
	}

	return c.wsConn.WriteMessage(websocket.TextMessage, data)
}

// ReceiveWebSocket receives a message from WebSocket
func (c *Client) ReceiveWebSocket(ctx context.Context) (json.RawMessage, error) {
	if c.wsConn == nil {
		return nil, fmt.Errorf("websocket not connected")
	}

	// Set read deadline from context
	if deadline, ok := ctx.Deadline(); ok {
		if err := c.wsConn.SetReadDeadline(deadline); err != nil {
			return nil, fmt.Errorf("failed to set read deadline: %w", err)
		}
	}

	messageType, data, err := c.wsConn.ReadMessage()
	if err != nil {
		// Retry once on abnormal closure to tolerate immediate server close after response
		if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure) || strings.Contains(err.Error(), "unexpected EOF") {
			// Briefly wait and retry a single read within the same deadline window
			time.Sleep(10 * time.Millisecond)
			messageType, data, err = c.wsConn.ReadMessage()
		}
		if err != nil {
			return nil, err
		}
	}

	if messageType != websocket.TextMessage {
		return nil, fmt.Errorf("unexpected message type: %d", messageType)
	}

	return json.RawMessage(data), nil
}

// CloseWebSocket closes the WebSocket connection gracefully
func (c *Client) CloseWebSocket() error {
	if c.wsConn == nil {
		return nil
	}

	// Send close message
	deadline := time.Now().Add(5 * time.Second)
	closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	err := c.wsConn.WriteControl(websocket.CloseMessage, closeMsg, deadline)

	// Always close the connection
	closeErr := c.wsConn.Close()
	c.wsConn = nil

	// Ignore "broken pipe" errors on close - the server may have already closed
	if err != nil && strings.Contains(err.Error(), "broken pipe") {
		err = nil
	}
	if closeErr != nil && strings.Contains(closeErr.Error(), "broken pipe") {
		closeErr = nil
	}

	// Return the first error
	if err != nil {
		return err
	}
	return closeErr
}

// IsConnected returns true if WebSocket is connected
func (c *Client) IsConnected() bool {
	return c.wsConn != nil
}
