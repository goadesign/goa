// getConn returns the current connection or creates a new one
func (c *{{ .ClientStruct }}) getConn(ctx context.Context) (*jsonrpc.WebSocketConn, error) {
	c.connMu.RLock()
	conn := c.conn
	if conn != nil {
		select {
		case <-conn.Done():
			// Connection closed, need new one
		default:
			// Connection appears to be good
			c.connMu.RUnlock()
			return conn, nil
		}
	}
	c.connMu.RUnlock()
	
	// Create new connection
	c.connMu.Lock()
	defer c.connMu.Unlock()
	
	// Double-check after acquiring write lock
	if c.conn != nil {
		select {
		case <-c.conn.Done():
			// Still need new connection
		default:
			return c.conn, nil
		}
	}
	
	// Dial WebSocket
	url := c.scheme + "://" + c.host + "/"
	header := make(http.Header)
	
	ws, _, err := c.dialer.DialContext(ctx, url, header)
	if err != nil {
		return nil, goahttp.ErrRequestError("{{ .Service.Name }}", "connect", err)
	}
	
	if c.configfn != nil {
		ws = c.configfn(ws, nil)
	}
	
	// Create connection for JSON-RPC over WebSocket
	c.conn = jsonrpc.NewConn(ws)
	
	return c.conn, nil
}
