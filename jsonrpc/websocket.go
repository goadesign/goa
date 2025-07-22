package jsonrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	goahttp "goa.design/goa/v3/http"
)

type (
	// IDProvider defines the interface for generating request IDs.
	// Implementations should provide unique identifiers for JSON-RPC requests.
	IDProvider interface {
		NextID() string
	}

	// ConnOption defines a configuration option for WebSocketConn.
	// Options are applied during connection creation to customize behavior.
	ConnOption func(*connConfig)

	// WebSocketConn manages a JSON-RPC 2.0 connection over WebSocket.
	// It handles request/response correlation, concurrent access, and connection lifecycle.
	// WebSocketConn is safe for concurrent use by multiple goroutines.
	WebSocketConn struct {
		ws *websocket.Conn

		encoder    func(io.Writer) goahttp.Encoder
		decoder    func(io.Reader) goahttp.Decoder
		idProvider IDProvider

		pending sync.Map
		send    chan []byte
		done    chan struct{}
	}

	atomicIDProvider struct {
		counter atomic.Uint64
	}

	connConfig struct {
		encoder        func(io.Writer) goahttp.Encoder
		decoder        func(io.Reader) goahttp.Decoder
		idProvider     IDProvider
		sendBufferSize int
	}
)

// WithEncoder returns a ConnOption that sets a custom JSON encoder.
// The encoder will be used for all JSON marshaling operations.
func WithEncoder(encoder func(io.Writer) goahttp.Encoder) ConnOption {
	return func(c *connConfig) {
		c.encoder = encoder
	}
}

// WithDecoder returns a ConnOption that sets a custom JSON decoder.
// The decoder will be used for all JSON unmarshaling operations.
func WithDecoder(decoder func(io.Reader) goahttp.Decoder) ConnOption {
	return func(c *connConfig) {
		c.decoder = decoder
	}
}

// WithSendBufferSize returns a ConnOption that sets the buffer size for the send channel.
// A larger buffer can improve performance under high load but uses more memory.
// The default buffer size is 256.
func WithSendBufferSize(size int) ConnOption {
	return func(c *connConfig) {
		c.sendBufferSize = size
	}
}

// WithIDProvider returns a ConnOption that sets a custom request ID provider.
// The provider will be used to generate unique identifiers for all requests.
func WithIDProvider(provider IDProvider) ConnOption {
	return func(c *connConfig) {
		c.idProvider = provider
	}
}

// NewConn creates a new JSON-RPC connection over the provided WebSocket.
// The connection automatically starts background goroutines to handle reading and writing.
// Options can be provided to customize JSON encoding, ID generation, and buffer sizes.
//
// The returned connection is ready for immediate use and will remain active until
// Close is called or the underlying WebSocket connection is terminated.
func NewConn(ws *websocket.Conn, opts ...ConnOption) *WebSocketConn {
	config := &connConfig{
		encoder:        standardEncoder,
		decoder:        standardDecoder,
		idProvider:     &atomicIDProvider{},
		sendBufferSize: 256,
	}

	for _, opt := range opts {
		opt(config)
	}

	c := &WebSocketConn{
		ws:         ws,
		encoder:    config.encoder,
		decoder:    config.decoder,
		idProvider: config.idProvider,
		send:       make(chan []byte, config.sendBufferSize),
		done:       make(chan struct{}),
	}

	go c.readPump()
	go c.writePump()

	return c
}

// Call performs a JSON-RPC 2.0 method call and waits for the response.
// The method blocks until a response is received, the context is canceled,
// or the connection is closed.
//
// If params is non-nil, it will be JSON-marshaled and included in the request.
// If result is non-nil and the response contains a result, it will be JSON-unmarshaled
// into result.
//
// Call returns an error if the request fails to send, the response contains an error,
// or JSON marshaling/unmarshaling fails.
func (c *WebSocketConn) Call(ctx context.Context, method string, params, result any) error {
	id := c.idProvider.NextID()

	req := RawRequest{
		JSONRPC: "2.0",
		Method:  method,
		ID:      &id,
	}

	if params != nil {
		var buf bytes.Buffer
		if err := c.encoder(&buf).Encode(params); err != nil {
			return fmt.Errorf("marshal params: %w", err)
		}
		req.Params = buf.Bytes()
	}

	var buf bytes.Buffer
	if err := c.encoder(&buf).Encode(req); err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}
	reqData := buf.Bytes()

	respChan := make(chan []byte, 1)
	c.pending.Store(id, respChan)
	defer c.pending.Delete(id)

	select {
	case c.send <- reqData:
	case <-ctx.Done():
		return ctx.Err()
	case <-c.done:
		return fmt.Errorf("connection closed")
	}

	select {
	case respData := <-respChan:
		var resp RawResponse
		if err := c.decoder(bytes.NewReader(respData)).Decode(&resp); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}

		if resp.Error != nil {
			return resp.Error
		}

		if result != nil && len(resp.Result) > 0 {
			return c.decoder(bytes.NewReader(resp.Result)).Decode(result)
		}

		return nil

	case <-ctx.Done():
		return ctx.Err()
	case <-c.done:
		return fmt.Errorf("connection closed")
	}
}

// Notify sends a JSON-RPC 2.0 notification (no response expected).
// Notifications are fire-and-forget messages that do not expect a response.
//
// If params is non-nil, it will be JSON-marshaled and included in the notification.
//
// Notify returns an error if the notification fails to send or JSON marshaling fails.
func (c *WebSocketConn) Notify(ctx context.Context, method string, params interface{}) error {
	req := Request{
		JSONRPC: "2.0",
		Method:  method,
	}

	if params != nil {
		var buf bytes.Buffer
		if err := c.encoder(&buf).Encode(params); err != nil {
			return fmt.Errorf("marshal params: %w", err)
		}
		req.Params = buf.Bytes()
	}

	var buf bytes.Buffer
	if err := c.encoder(&buf).Encode(req); err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	select {
	case c.send <- buf.Bytes():
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-c.done:
		return fmt.Errorf("connection closed")
	}
}

// Close gracefully closes the WebSocket connection.
// It sends a close frame to the peer and closes the underlying connection.
//
// After Close returns, no further operations should be performed on the connection.
func (c *WebSocketConn) Close() error {
	if err := c.ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
		return err
	}
	return c.ws.Close()
}

// Done returns a channel that is closed when the connection is closed.
// This can be used to detect when the connection has been terminated.
func (c *WebSocketConn) Done() <-chan struct{} {
	return c.done
}

func (p *atomicIDProvider) NextID() string {
	return strconv.FormatUint(p.counter.Add(1), 10)
}

func (c *WebSocketConn) readPump() {
	defer close(c.done)

	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			return
		}

		var msg struct {
			ID interface{} `json:"id"`
		}
		if err := c.decoder(bytes.NewReader(message)).Decode(&msg); err != nil {
			continue
		}

		if msg.ID == nil {
			continue
		}

		var id string
		switch v := msg.ID.(type) {
		case string:
			id = v
		case float64:
			id = strconv.FormatFloat(v, 'f', -1, 64)
		case int:
			id = strconv.Itoa(v)
		default:
			continue
		}

		if ch, ok := c.pending.Load(id); ok {
			if respChan, ok := ch.(chan<- []byte); ok {
				select {
				case respChan <- message:
				default:
				}
			}
		}
	}
}

func (c *WebSocketConn) writePump() {
	for {
		select {
		case message := <-c.send:
			if err := c.ws.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-c.done:
			return
		}
	}
}

// Default to standard json encoder/decoder.
func standardEncoder(w io.Writer) goahttp.Encoder { return json.NewEncoder(w) }
func standardDecoder(r io.Reader) goahttp.Decoder { return json.NewDecoder(r) }
