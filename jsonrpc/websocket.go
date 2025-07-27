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

	// NotificationHandler is called when a notification is received from the server.
	// Notifications are messages without an ID that don't expect a response.
	// The method parameter contains the notification method name, and params contains
	// the raw JSON parameters (if any). Use a JSON decoder to unmarshal params into
	// your desired type.
	NotificationHandler func(method string, params json.RawMessage)

	// WebSocketConn manages a JSON-RPC 2.0 connection over WebSocket.
	// It handles request/response correlation, concurrent access, and connection lifecycle.
	// WebSocketConn is safe for concurrent use by multiple goroutines.
	WebSocketConn struct {
		ws *websocket.Conn

		errorHandler        func(error)
		notificationHandler NotificationHandler
		encoder             func(io.Writer) goahttp.Encoder
		decoder             func(io.Reader) goahttp.Decoder
		idProvider          IDProvider

		pending           sync.Map
		send              chan []byte
		done              chan struct{}
		notificationQueue chan notificationJob
		workersDone       sync.WaitGroup
	}

	// ReadError is returned when the background goroutine fails to read from the
	// connection.
	ReadError struct {
		Err error
	}

	// WriteError is returned when the background goroutine fails to write to the
	// connection.
	WriteError struct {
		Err error
	}

	// DecodeError is returned when the background goroutine fails to decode a
	// JSON message received from the connection.
	DecodeError struct {
		Err error
	}

	// HandlerError is returned when a notification handler panics.
	HandlerError struct {
		Err error
	}

	atomicIDProvider struct {
		counter atomic.Uint64
	}

	connConfig struct {
		encoder                 func(io.Writer) goahttp.Encoder
		decoder                 func(io.Reader) goahttp.Decoder
		idProvider              IDProvider
		sendBufferSize          int
		errorHandler            func(error)
		notificationHandler     NotificationHandler
		notificationWorkerCount int
		notificationQueueSize   int
	}

	notificationJob struct {
		method string
		params json.RawMessage
	}
)

// WithErrorHandler returns a ConnOption that sets a custom error handler for the WebSocketConn.
// The provided handler will be invoked whenever an error occurs in the background
// goroutines responsible for reading from or writing to the WebSocket connection.
//
// The error passed to the handler will be of type ReadError, WriteError, DecodeError, or HandlerError,
// which wrap the underlying error. To determine the specific cause, use errors.Is or errors.As
// to inspect the wrapped error (for example, to check for *websocket.CloseError).
// Refer to the gorilla/websocket package documentation for possible error codes and types.
func WithErrorHandler(handler func(error)) ConnOption {
	return func(c *connConfig) {
		c.errorHandler = handler
	}
}

// WithNotificationHandler returns a ConnOption that sets a custom notification handler.
// The handler will be called when the connection receives a notification from the server
// (a JSON-RPC 2.0 message without an ID field).
//
// Notifications are fire-and-forget messages from the server that don't expect a response.
// Common use cases include server-sent events, status updates, or real-time data pushes.
//
// The notification handler is processed by a worker pool (default: 4 workers with queue
// size 100) to avoid blocking the connection's message processing. If the queue is full,
// notifications will be dropped and reported as HandlerError. If the handler panics, the
// panic will be recovered and reported as HandlerError via the error handler.
// Use WithNotificationWorkers to configure the worker pool size.
//
// Example:
//
//	handler := func(method string, params json.RawMessage) {
//	    switch method {
//	    case "server.notification":
//	        var data ServerNotification
//	        json.Unmarshal(params, &data)
//	        // handle the notification
//	    }
//	}
//	conn := NewConn(ws, WithNotificationHandler(handler))
func WithNotificationHandler(handler NotificationHandler) ConnOption {
	return func(c *connConfig) {
		c.notificationHandler = handler
	}
}

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

// WithNotificationWorkers returns a ConnOption that sets the number of worker goroutines
// for processing notifications and the size of the notification queue.
//
// Default values are 4 workers with a queue size of 100.
// Setting workerCount to 0 disables the worker pool and processes notifications synchronously.
func WithNotificationWorkers(workerCount, queueSize int) ConnOption {
	return func(c *connConfig) {
		c.notificationWorkerCount = workerCount
		c.notificationQueueSize = queueSize
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
		encoder:                 standardEncoder,
		decoder:                 standardDecoder,
		idProvider:              &atomicIDProvider{},
		sendBufferSize:          256,
		notificationWorkerCount: 4,
		notificationQueueSize:   100,
	}

	for _, opt := range opts {
		opt(config)
	}

	c := &WebSocketConn{
		ws:                  ws,
		errorHandler:        config.errorHandler,
		notificationHandler: config.notificationHandler,
		encoder:             config.encoder,
		decoder:             config.decoder,
		idProvider:          config.idProvider,
		send:                make(chan []byte, config.sendBufferSize),
		done:                make(chan struct{}),
	}

	// Initialize notification worker pool if handler is provided
	if config.notificationHandler != nil && config.notificationWorkerCount > 0 {
		c.notificationQueue = make(chan notificationJob, config.notificationQueueSize)
		c.startNotificationWorkers(config.notificationWorkerCount)
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
	// Close notification queue to shutdown workers
	if c.notificationQueue != nil {
		close(c.notificationQueue)
		c.workersDone.Wait()
	}

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

func (c *WebSocketConn) startNotificationWorkers(count int) {
	for i := 0; i < count; i++ {
		c.workersDone.Add(1)
		go c.notificationWorker()
	}
}

func (c *WebSocketConn) notificationWorker() {
	defer c.workersDone.Done()

	for job := range c.notificationQueue {
		func() {
			// Recover from panics in user notification handlers
			defer func() {
				if r := recover(); r != nil {
					c.handleError(HandlerError{Err: fmt.Errorf("notification handler panic: %v", r)})
				}
			}()
			c.notificationHandler(job.method, job.params)
		}()
	}
}

func (c *WebSocketConn) handleNotification(message []byte) {
	if c.notificationHandler == nil {
		return
	}

	var notification struct {
		Method string          `json:"method"`
		Params json.RawMessage `json:"params"`
	}
	if err := c.decoder(bytes.NewReader(message)).Decode(&notification); err != nil {
		c.handleError(DecodeError{Err: err})
		return
	}
	if notification.Method != "" {
		if c.notificationQueue != nil {
			// Use worker pool for notification handling
			select {
			case c.notificationQueue <- notificationJob{
				method: notification.Method,
				params: notification.Params,
			}:
			default:
				// Queue is full, drop notification and report error
				c.handleError(HandlerError{Err: fmt.Errorf("notification queue full, dropping notification: %s", notification.Method)})
			}
		} else {
			// No worker pool configured, handle synchronously (blocking)
			func() {
				defer func() {
					if r := recover(); r != nil {
						c.handleError(HandlerError{Err: fmt.Errorf("notification handler panic: %v", r)})
					}
				}()
				c.notificationHandler(notification.Method, notification.Params)
			}()
		}
	}
	return
}

func (c *WebSocketConn) readPump() {
	defer close(c.done)

	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			c.handleError(ReadError{Err: err})
			return
		}

		var msg struct {
			ID any `json:"id"`
		}
		if err := c.decoder(bytes.NewReader(message)).Decode(&msg); err != nil {
			c.handleError(DecodeError{Err: err})
			continue
		}

		if msg.ID == nil {
			c.handleNotification(message)
			continue
		}

		// This is a response - convert ID to string
		var id string
		switch v := msg.ID.(type) {
		case string:
			id = v
		case float64:
			id = strconv.FormatFloat(v, 'f', -1, 64)
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
		} else {
			c.handleError(fmt.Errorf("received response for unknown id %q", id))
		}
	}
}

func (c *WebSocketConn) writePump() {
	for {
		select {
		case message := <-c.send:
			if err := c.ws.WriteMessage(websocket.TextMessage, message); err != nil {
				c.handleError(WriteError{Err: err})
				return
			}
		case <-c.done:
			return
		}
	}
}

func (c *WebSocketConn) handleError(err error) {
	if c.errorHandler != nil {
		c.errorHandler(err)
	}
}

// Error returns the underlying error message.
func (e ReadError) Error() string { return e.Err.Error() }

// Error returns the underlying error message.
func (e WriteError) Error() string { return e.Err.Error() }

// Error returns the underlying error message.
func (e DecodeError) Error() string { return e.Err.Error() }

// Error returns the underlying error message.
func (e HandlerError) Error() string { return e.Err.Error() }

// Default to standard json encoder/decoder.
func standardEncoder(w io.Writer) goahttp.Encoder { return json.NewEncoder(w) }
func standardDecoder(r io.Reader) goahttp.Decoder { return json.NewDecoder(r) }
