package http

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"goa.design/goa"
)

type (
	// Upgrader is an HTTP connection that is able to upgrade to WebSocket
	// protocol.
	Upgrader interface {
		// Upgrade upgrades the HTTP connection to the WebSocket protocol.
		Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*websocket.Conn, error)
	}

	// Dialer creates a websocket connection to a given URL.
	Dialer interface {
		// DialContext creates a client connection to the websocket server.
		DialContext(ctx context.Context, url string, h http.Header) (*websocket.Conn, *http.Response, error)
	}

	// Streamer defines the streaming behavior using the WebSocket protocol.
	Streamer interface {
		goa.Contextualizer
		// SendMsg sends value v across the websocket connection. It returns any
		// error occurred during send.
		//
		// A client-side sender must send a nil message to signal the end of data
		// transmission.
		SendMsg(v interface{}) error
		// RecvMsg receives a value from the websocket connection into v.
		// It returns any error occurred during receive.
		//
		// A client-side receiver must return io.EOF error when the server closes
		// the stream normally (for example, using websocket.CloseNormalClosure).
		RecvMsg(v interface{}) error
		// Close closes the websocket connection.
		//
		// A server-side must always send a websocket.CloseMessage to the client
		// before actually closing the stream.
		//
		// A client-side stream must never send a websocket.CloseMessage to close
		// the connection if it must wait for the server response which is also
		// sent across the same stream. This is because server cannot use the same
		// connection to send a response after a websocket.CloseMessage is sent.
		Close() error
		// WithConn updates the underlying websocket connection and context.
		WithConn(*websocket.Conn, ConnConfigureFunc)
	}

	// ConnConfigureFunc is used to configure a websocket connection with
	// custom handlers.
	ConnConfigureFunc func(context.Context, *websocket.Conn) (context.Context, *websocket.Conn)

	// ServerStream is the default goa HTTP server-side streamer.
	ServerStream struct {
		*stream
	}

	// ClientStream is the default goa HTTP client-side streamer.
	ClientStream struct {
		*stream
	}

	// stream implements the streamer interface.
	stream struct {
		ctx  context.Context
		conn *websocket.Conn
	}
)

// NewServerStream returns a new server-side streamer.
func NewServerStream(ctx context.Context) *ServerStream {
	return &ServerStream{&stream{ctx: ctx}}
}

// NewClientStream returns a new server-side streamer.
func NewClientStream(ctx context.Context) *ClientStream {
	return &ClientStream{&stream{ctx: ctx}}
}

// SendMsg sends value v to the client. It returns any error occurred during
// sendint the message.
func (s *ServerStream) SendMsg(v interface{}) error {
	if s.conn == nil {
		return goa.Fault("error sending: server connection closed or not upgraded to WebSocket protocol")
	}
	return s.conn.WriteJSON(v)
}

// RecvMsg receives the value from the client into v. It returns any error
// occurred during receiving the message. It returns io.EOF error if the
// received message is nil.
func (s *ServerStream) RecvMsg(v interface{}) error {
	if s.conn == nil {
		return goa.Fault("error receiving: server connection closed or not upgraded to WebSocket protocol")
	}
	if err := s.conn.ReadJSON(v); err != nil {
		return err
	}
	return nil
}

// Close closes the websocket connection. It sends a websocket close message
// with the websocket.CloseNormalClosure code.
func (s *ServerStream) Close() error {
	if err := s.conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "server closing connection"),
		time.Now().Add(time.Second),
	); err != nil {
		return err
	}
	return s.conn.Close()
}

// SendMsg sends value v to the server. It returns any error occurred during
// sending the message.
func (c *ClientStream) SendMsg(v interface{}) error {
	if c.conn == nil {
		return goa.Fault("error sending: client connection closed or not upgraded to Websocket protocol")
	}
	return c.conn.WriteJSON(v)
}

// RecvMsg receives the value from the server into v. It returns any error
// occurred during receiving the message. It returns io.EOF error if the
// error is a websocket.CloseNormalClosure error.
func (c *ClientStream) RecvMsg(v interface{}) error {
	if c.conn == nil {
		return goa.Fault("error receiving: client connection closed or not upgraded to WebSocket protocol")
	}
	if err := c.conn.ReadJSON(v); err != nil {
		if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			return io.EOF
		}
		return err
	}
	return nil
}

// Close closes the websocket connection.
func (c *ClientStream) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (s *stream) Conn() *websocket.Conn {
	return s.conn
}

func (s *stream) Context() context.Context {
	return s.ctx
}

func (s *stream) SetContext(ctx context.Context) {
	s.ctx = ctx
}

func (s *stream) WithConn(conn *websocket.Conn, fn ConnConfigureFunc) {
	ctx := s.Context()
	if fn != nil {
		ctx, conn = fn(ctx, conn)
		s.SetContext(ctx)
	}
	s.conn = conn
}
