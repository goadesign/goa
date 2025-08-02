package harness

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
)

// ClientConfig contains configuration for a test client
type ClientConfig struct {
	// SourceDir is the directory containing the generated client code
	SourceDir string

	// ServerURL is the URL of the server to connect to
	ServerURL string

	// Transport specifies the transport type (http, websocket, sse)
	Transport string
}

// ClientProcess represents a client for making requests with improved error handling
type ClientProcess struct {
	workDir    string
	config     ClientConfig
	logFile    *os.File
	httpClient *http.Client
	wsConn     *websocket.Conn
}

// NewClient creates a new client process with optimized timeouts for quick failure
func NewClient(workDir string, config ClientConfig) (*ClientProcess, error) {
	// Create log file
	logFile, err := os.Create(filepath.Join(workDir, "client.log"))
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	// Configure transport with aggressive timeouts for test scenarios
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   2 * time.Second, // Connection timeout
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   2 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConns:          10,
		IdleConnTimeout:       90 * time.Second,
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second, // Overall request timeout
	}

	return &ClientProcess{
		workDir:    workDir,
		config:     config,
		logFile:    logFile,
		httpClient: httpClient,
	}, nil
}

// CallJSONRPC makes a JSON-RPC request over HTTP transport
func (c *ClientProcess) CallJSONRPC(ctx context.Context, method string, params any) (json.RawMessage, error) {
	// Create request ID
	reqID := fmt.Sprintf("test-%d", time.Now().UnixNano())

	// Build JSON-RPC request
	request := map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
		"id":      reqID,
	}
	if params != nil {
		request["params"] = params
	}

	// Log request
	fmt.Fprintf(c.logFile, "[%s] Request: %s\n", time.Now().Format(time.RFC3339), method)
	if reqBytes, err := json.MarshalIndent(request, "", "  "); err == nil {
		fmt.Fprintln(c.logFile, string(reqBytes))
	}

	// Marshal request
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "POST", c.config.ServerURL+"/jsonrpc", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Make request with quick failure on connection errors
	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Check if it's a connection error for quick failure
		if netErr, ok := err.(net.Error); ok && (netErr.Timeout() || !netErr.Temporary()) {
			return nil, fmt.Errorf("connection failed immediately: %w", err)
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse response
	var response struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      string          `json:"id"`
		Result  json.RawMessage `json:"result"`
		Error   *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    any    `json:"data,omitempty"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Log response
	fmt.Fprintf(c.logFile, "[%s] Response:\n", time.Now().Format(time.RFC3339))
	if respBytes, err := json.MarshalIndent(response, "", "  "); err == nil {
		fmt.Fprintln(c.logFile, string(respBytes))
	}

	// Check for JSON-RPC error
	if response.Error != nil {
		return nil, fmt.Errorf("JSON-RPC error %d: %s", response.Error.Code, response.Error.Message)
	}

	// Verify response ID matches request ID
	if response.ID != reqID {
		return nil, fmt.Errorf("response ID mismatch: expected %s, got %s", reqID, response.ID)
	}

	return response.Result, nil
}

// CallHTTPBatch makes a batch JSON-RPC request over HTTP
func (c *ClientProcess) CallHTTPBatch(ctx context.Context, requests []Request) ([]json.RawMessage, error) {
	// Create request body as an array
	body, err := json.Marshal(requests)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch requests: %w", err)
	}

	// Log batch request
	fmt.Fprintf(c.logFile, "[%s] Batch Request:\n", time.Now().Format(time.RFC3339))
	if prettyJSON, err := json.MarshalIndent(requests, "", "  "); err == nil {
		fmt.Fprintln(c.logFile, string(prettyJSON))
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", c.config.ServerURL+"/jsonrpc", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create batch request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("batch request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Log raw response
	fmt.Fprintf(c.logFile, "[%s] Batch Response:\n%s\n", time.Now().Format(time.RFC3339), string(responseBody))

	// Parse as array of responses
	var responses []json.RawMessage
	if err := json.Unmarshal(responseBody, &responses); err != nil {
		// Try parsing as single response (server might not support batch)
		var singleResponse json.RawMessage
		if err := json.Unmarshal(responseBody, &singleResponse); err != nil {
			return nil, fmt.Errorf("failed to parse batch response: %w", err)
		}
		return []json.RawMessage{singleResponse}, nil
	}

	return responses, nil
}

// CallHTTP makes a JSON-RPC request and returns the full response for validation
func (c *ClientProcess) CallHTTP(ctx context.Context, method string, params any) (*Response, error) {
	// Create request ID
	reqID := fmt.Sprintf("test-%d", time.Now().UnixNano())

	// Build JSON-RPC request
	request := map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
		"id":      reqID,
	}
	if params != nil {
		request["params"] = params
	}

	// Log request
	fmt.Fprintf(c.logFile, "[%s] Request: %s\n", time.Now().Format(time.RFC3339), method)
	if reqBytes, err := json.MarshalIndent(request, "", "  "); err == nil {
		fmt.Fprintln(c.logFile, string(reqBytes))
	}

	// Marshal request
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "POST", c.config.ServerURL+"/jsonrpc", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Make request with quick failure on connection errors
	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Check if it's a connection error for quick failure
		if netErr, ok := err.(net.Error); ok && (netErr.Timeout() || !netErr.Temporary()) {
			return nil, fmt.Errorf("connection failed immediately: %w", err)
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse response
	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Log response
	fmt.Fprintf(c.logFile, "[%s] Response:\n", time.Now().Format(time.RFC3339))
	if respBytes, err := json.MarshalIndent(response, "", "  "); err == nil {
		fmt.Fprintln(c.logFile, string(respBytes))
	}

	return &response, nil
}

// SendNotification sends a JSON-RPC notification (no response expected)
func (c *ClientProcess) SendNotification(ctx context.Context, method string, params any) error {
	// Build JSON-RPC notification (no ID field)
	notification := map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
	}
	if params != nil {
		notification["params"] = params
	}

	// Log notification
	fmt.Fprintf(c.logFile, "[%s] Notification: %s\n", time.Now().Format(time.RFC3339), method)
	if notifBytes, err := json.MarshalIndent(notification, "", "  "); err == nil {
		fmt.Fprintln(c.logFile, string(notifBytes))
	}

	// Marshal notification
	body, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "POST", c.config.ServerURL+"/jsonrpc", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send notification
	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Check if it's a connection error
		if netErr, ok := err.(net.Error); ok && (netErr.Timeout() || !netErr.Temporary()) {
			return fmt.Errorf("connection failed immediately: %w", err)
		}
		return fmt.Errorf("notification failed: %w", err)
	}
	defer resp.Body.Close()

	// For notifications, we expect 200 OK with no body
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// ConnectWebSocket establishes a WebSocket connection
func (c *ClientProcess) ConnectWebSocket(ctx context.Context) error {
	// Create WebSocket dialer with timeout
	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	// Convert HTTP URL to WebSocket URL
	wsURL := c.config.ServerURL
	if len(wsURL) > 4 && wsURL[:4] == "http" {
		wsURL = "ws" + wsURL[4:]
	}

	// Connect
	conn, _, err := dialer.DialContext(ctx, wsURL+"/jsonrpc/ws", nil)
	if err != nil {
		return fmt.Errorf("failed to connect WebSocket: %w", err)
	}

	c.wsConn = conn
	return nil
}

// SendWebSocketMessage sends a message over WebSocket with timeout from context
func (c *ClientProcess) SendWebSocketMessage(ctx context.Context, message any) error {
	if c.wsConn == nil {
		return fmt.Errorf("WebSocket not connected")
	}

	// Set write deadline from context
	if deadline, ok := ctx.Deadline(); ok {
		if err := c.wsConn.SetWriteDeadline(deadline); err != nil {
			return fmt.Errorf("failed to set write deadline: %w", err)
		}
		defer c.wsConn.SetWriteDeadline(time.Time{})
	}

	// Log message
	fmt.Fprintf(c.logFile, "[%s] WebSocket Send:\n", time.Now().Format(time.RFC3339))
	if msgBytes, err := json.MarshalIndent(message, "", "  "); err == nil {
		fmt.Fprintln(c.logFile, string(msgBytes))
	}

	return c.wsConn.WriteJSON(message)
}

// ReceiveWebSocketMessage receives a message from WebSocket with timeout from context
func (c *ClientProcess) ReceiveWebSocketMessage(ctx context.Context) (any, error) {
	if c.wsConn == nil {
		return nil, fmt.Errorf("WebSocket not connected")
	}

	// Set read deadline from context
	if deadline, ok := ctx.Deadline(); ok {
		if err := c.wsConn.SetReadDeadline(deadline); err != nil {
			return nil, fmt.Errorf("failed to set read deadline: %w", err)
		}
		defer c.wsConn.SetReadDeadline(time.Time{})
	}

	var message any
	err := c.wsConn.ReadJSON(&message)
	if err != nil {
		return nil, err
	}

	// Log message
	fmt.Fprintf(c.logFile, "[%s] WebSocket Receive:\n", time.Now().Format(time.RFC3339))
	if msgBytes, err := json.MarshalIndent(message, "", "  "); err == nil {
		fmt.Fprintln(c.logFile, string(msgBytes))
	}

	return message, nil
}

// ConnectSSE establishes a Server-Sent Events connection
func (c *ClientProcess) ConnectSSE(ctx context.Context, path string, params any) (*SSEClient, error) {
	// Build URL with query parameters for GET request
	reqURL := c.config.ServerURL + path
	if params != nil {
		// Convert params to query parameters
		if paramMap, ok := params.(map[string]any); ok {
			values := url.Values{}
			for k, v := range paramMap {
				values.Add(k, fmt.Sprintf("%v", v))
			}
			if len(values) > 0 {
				reqURL += "?" + values.Encode()
			}
		}
	}

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSE request: %w", err)
	}
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect SSE: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return &SSEClient{
		reader: bufio.NewReader(resp.Body),
		closer: resp.Body,
		log:    c.logFile,
	}, nil
}

// Stop closes the client connections and cleans up resources
func (c *ClientProcess) Stop() error {
	// Close WebSocket if connected
	if c.wsConn != nil {
		c.wsConn.Close()
	}

	// Close log file
	if c.logFile != nil {
		c.logFile.Close()
	}

	return nil
}

// SSEClient handles Server-Sent Events
type SSEClient struct {
	reader *bufio.Reader
	closer io.Closer
	log    *os.File
}

// ReadEvent reads the next SSE event
func (s *SSEClient) ReadEvent() (*SSEEvent, error) {
	event := &SSEEvent{}

	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = line[:len(line)-1] // Remove newline

		if line == "" {
			// Empty line signals end of event
			if event.Data != "" {
				// Log event
				fmt.Fprintf(s.log, "[%s] SSE Event: %s\n", time.Now().Format(time.RFC3339), event.Data)
				return event, nil
			}
			continue
		}

		if len(line) > 5 && line[:5] == "data:" {
			event.Data = line[5:]
			if len(event.Data) > 0 && event.Data[0] == ' ' {
				event.Data = event.Data[1:]
			}
		} else if len(line) > 6 && line[:6] == "event:" {
			event.Event = line[6:]
			if len(event.Event) > 0 && event.Event[0] == ' ' {
				event.Event = event.Event[1:]
			}
		} else if len(line) > 3 && line[:3] == "id:" {
			event.ID = line[3:]
			if len(event.ID) > 0 && event.ID[0] == ' ' {
				event.ID = event.ID[1:]
			}
		}
	}
}

// Close closes the SSE connection
func (s *SSEClient) Close() error {
	return s.closer.Close()
}

// SSEEvent represents a Server-Sent Event
type SSEEvent struct {
	Event string
	Data  string
	ID    string
}
