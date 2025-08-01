{{/*
websocket_client_conn.go.tpl generates WebSocket connection management methods for JSON-RPC clients.

This template provides connection lifecycle management including:
- Connection establishment with health checking
- Connection reuse and automatic reconnection
- Thread-safe connection access with read/write locking
- Proper cleanup on client close

Template variables:
- .ClientStruct: Name of the generated client struct
*/}}
// getConn returns the current WebSocket connection or creates a new one
func (c *{{ .ClientStruct }}) getConn(ctx context.Context) (*websocket.Conn, error) {
	c.connMu.RLock()
	conn := c.conn
	if conn != nil {
		// Check if connection is still alive
		if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(5*time.Second)); err == nil {
			c.connMu.RUnlock()
			return conn, nil
		}
		// Connection is dead, need new one
	}
	c.connMu.RUnlock()
	
	// Create new connection
	c.connMu.Lock()
	defer c.connMu.Unlock()
	
	// Double-check after acquiring write lock
	if c.conn != nil {
		if err := c.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(5*time.Second)); err == nil {
			return c.conn, nil
		}
		// Close the dead connection
		c.conn.Close()
	}
	
	// Convert scheme for WebSocket
	wsScheme := "ws"
	if c.scheme == "https" {
		wsScheme = "wss"
	}
	
	// Find the WebSocket path from the service endpoints
	{{- $found := false }}
	{{- range .Endpoints }}
		{{- range .Routes }}
			{{- if and (eq .Verb "GET") (ne .Path "/") (not $found) }}
	url := wsScheme + "://" + c.host + {{ printf "%q" .Path }}
				{{ $found = true }}
			{{- end }}
		{{- end }}
	{{- end }}
	{{- if not $found }}
	url := wsScheme + "://" + c.host
	{{- end }}
	
	ws, _, err := c.dialer.DialContext(ctx, url, nil)
	if err != nil {
		return nil, goahttp.ErrRequestError("{{ .Service.Name }}", "connect", err)
	}
	
	if c.configfn != nil {
		ws = c.configfn(ws, nil)
	}
	
	// Store the direct WebSocket connection
	c.conn = ws
	
	return c.conn, nil
}

// Close closes the WebSocket connection and marks the client as closed
func (c *{{ .ClientStruct }}) Close() error {
	if c.closed.Swap(true) {
		return nil // Already closed
	}

	c.connMu.Lock()
	defer c.connMu.Unlock()

	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}

// IsClosed returns true if the client connection has been closed
func (c *{{ .ClientStruct }}) IsClosed() bool {
	return c.closed.Load()
}
