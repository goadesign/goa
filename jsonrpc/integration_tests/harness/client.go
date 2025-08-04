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
	Method string `json:"method"`
	Params any    `json:"params,omitempty"`
	ID     any    `json:"id,omitempty"`
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
		wsDialer = websocket.DefaultDialer
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
	defer resp.Body.Close()

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
		"jsonrpc": "2.0",
		"method":  req.Method,
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
	defer resp.Body.Close()

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
	envelope := map[string]any{
		"jsonrpc": "2.0",
		"method":  req.Method,
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

	endpoint := c.baseURL.ResolveReference(&url.URL{Path: c.config.SSEPath})
	// Debug logging
	fmt.Printf("DEBUG: SSE endpoint: %s\n", endpoint.String())
	fmt.Printf("DEBUG: SSE request: %s\n", string(data))
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	// Read response body for debug
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read SSE response: %w", err)
	}
	fmt.Printf("DEBUG: SSE raw response: %q\n", string(body))
	
	// Parse SSE events
	events, err := c.parseSSEEvents(bytes.NewReader(body))
	fmt.Printf("DEBUG: SSE events received: %d\n", len(events))
	for i, event := range events {
		fmt.Printf("DEBUG: SSE event %d: %s\n", i, string(event))
	}
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
	
	conn, _, err := c.wsDialer.DialContext(ctx, wsURL.String(), headers)
	if err != nil {
		return fmt.Errorf("websocket dial failed: %w", err)
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
	envelope := map[string]any{
		"jsonrpc": "2.0",
		"method":  req.Method,
	}
	if req.Params != nil {
		envelope["params"] = req.Params
	}
	if req.ID != nil {
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
		return nil, err
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