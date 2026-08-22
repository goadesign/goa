{{/*
websocket_client_conn.go.tpl generates the WebSocket shared by every JSON-RPC
method on one client. One function reads every server message. Each request
gets a unique ID, a timeout, and a function that receives its result or error.
The response ID selects that function. Only one caller writes to the socket at
a time.
*/}}
type (
	// {{ .WebSocketConnection.Name }} keeps the socket, the next request ID, and the
	// functions called when waiting requests receive a result or error.
	{{ .WebSocketConnection.Name }} struct {
		ws *websocket.Conn

		writeMu sync.Mutex
		nextID atomic.Uint64

		stateMu sync.Mutex
		pending map[string]*{{ .WebSocketPendingRequest.Name }}
		err     error
		done    chan struct{}

		closeOnce sync.Once
		closeErr  error

		ctx    context.Context
		config *jsonrpc.StreamConfig
	}

	// {{ .WebSocketRequestOwner.Name }} identifies one method stream. closed records that
	// Close was called so the stream cannot accept another request.
	{{ .WebSocketRequestOwner.Name }} struct {
		closed atomic.Bool
	}

	// {{ .WebSocketPendingRequest.Name }} stores one request's context, timer, and function
	// that receives its result or error.
	{{ .WebSocketPendingRequest.Name }} struct {
		owner    *{{ .WebSocketRequestOwner.Name }}
		ctx      context.Context
		timer    *time.Timer
		complete func(context.Context, *jsonrpc.RawResponse, error)
	}

	// {{ .WebSocketMessage.Name }} keeps enough of an incoming JSON-RPC message to tell a
	// server notification from a response, including an explicit null ID.
	{{ .WebSocketMessage.Name }} struct {
		JSONRPC string                    `json:"jsonrpc"`
		Method  string                    `json:"method,omitempty"`
		Params  json.RawMessage           `json:"params,omitempty"`
		Result  json.RawMessage           `json:"result,omitempty"`
		Error   *jsonrpc.RawErrorResponse `json:"error,omitempty"`
		ID      json.RawMessage           `json:"id"`
	}
)

// {{ .WebSocketClosedError.Name }} is returned after Close prevents a method
// stream from sending or receiving more messages.
var {{ .WebSocketClosedError.Name }} = errors.New("JSON-RPC WebSocket method stream is closed")

// {{ .NewWebSocketConnection.Name }} returns a connection that uses config for timeouts and
// error reporting. It also starts readResponses, the only function that reads
// from ws. Socket errors use a context that remains valid after the method that
// opened the socket returns.
func {{ .NewWebSocketConnection.Name }}(ws *websocket.Conn, config *jsonrpc.StreamConfig) *{{ .WebSocketConnection.Name }} {
	conn := &{{ .WebSocketConnection.Name }}{
		ws:      ws,
		pending: make(map[string]*{{ .WebSocketPendingRequest.Name }}),
		done:    make(chan struct{}),
		ctx:     context.Background(),
		config:  config,
	}
	go conn.readResponses()
	return conn
}

// getConn returns the open WebSocket connection. If no socket exists, it uses
// ctx to open one for all waiting callers. After the socket fails, later calls
// return that error instead of opening a new socket while earlier requests are
// still waiting for results.
func (c *{{ .ClientStructDeclaration.Name }}) getConn(ctx context.Context) (*{{ .WebSocketConnection.Name }}, error) {
	for {
		c.connMu.Lock()
		if c.closed.Load() {
			c.connMu.Unlock()
			return nil, fmt.Errorf("JSON-RPC WebSocket client is closed")
		}
		if c.conn != nil {
			conn := c.conn
			c.connMu.Unlock()
			if err := conn.terminalError(); err != nil {
				return nil, err
			}
			return conn, nil
		}
		if c.connecting != nil {
			connecting := c.connecting
			c.connMu.Unlock()
			select {
			case <-connecting:
				continue
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		connecting := make(chan struct{})
		c.connecting = connecting
		c.connMu.Unlock()
		return c.connect(ctx, connecting)
	}
}

// connect uses ctx to open and configure one socket. It returns that socket to
// the caller and makes it available to other getConn callers, unless Close ran
// while the dial was in progress.
func (c *{{ .ClientStructDeclaration.Name }}) connect(ctx context.Context, connecting chan struct{}) (*{{ .WebSocketConnection.Name }}, error) {

	wsScheme := "ws"
	if c.scheme == "https" {
		wsScheme = "wss"
	}

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
		c.finishConnect(connecting, nil)
		return nil, goahttp.ErrRequestError("{{ .Service.Name }}", "connect", err)
	}
	if c.configfn != nil {
		ws = c.configfn(ws, nil)
	}

	conn := {{ .NewWebSocketConnection.Name }}(ws, c.streamConfig)
	if c.finishConnect(connecting, conn) {
		return conn, nil
	}
	err = fmt.Errorf("JSON-RPC WebSocket client closed while connecting")
	if closeErr := conn.close(); closeErr != nil {
		err = fmt.Errorf("%w; close new connection: %v", err, closeErr)
	}
	return nil, err
}

// finishConnect stores conn unless the client closed while the dial was in
// progress. It wakes every getConn caller waiting for the dial and returns
// whether conn was stored.
func (c *{{ .ClientStructDeclaration.Name }}) finishConnect(connecting chan struct{}, conn *{{ .WebSocketConnection.Name }}) bool {
	c.connMu.Lock()
	accepted := !c.closed.Load() && conn != nil
	if accepted {
		c.conn = conn
	}
	c.connecting = nil
	c.connMu.Unlock()
	close(connecting)
	return accepted
}

// sendRequest stores the function passed in complete, writes request, and
// returns its new ID. It returns an error without writing if ctx is canceled,
// the method stream is closed, or the socket has failed. Its timer calls
// complete with a timeout error even if the caller never calls Recv.
func (c *{{ .WebSocketConnection.Name }}) sendRequest(ctx context.Context, request *jsonrpc.Request, owner *{{ .WebSocketRequestOwner.Name }}, complete func(context.Context, *jsonrpc.RawResponse, error)) (string, error) {
	id := strconv.FormatUint(c.nextID.Add(1), 10)
	request.ID = id
	pending := &{{ .WebSocketPendingRequest.Name }}{
		owner:    owner,
		ctx:      ctx,
		complete: complete,
	}

	c.writeMu.Lock()
	select {
	case <-ctx.Done():
		err := ctx.Err()
		if owner.closed.Load() {
			err = {{ .WebSocketClosedError.Name }}
		}
		c.writeMu.Unlock()
		return "", err
	default:
	}
	c.stateMu.Lock()
	switch {
	case c.err != nil:
		err := c.err
		c.stateMu.Unlock()
		c.writeMu.Unlock()
		return "", err
	case owner.closed.Load():
		c.stateMu.Unlock()
		c.writeMu.Unlock()
		return "", {{ .WebSocketClosedError.Name }}
	}
	c.pending[id] = pending
	pending.timer = time.AfterFunc(c.config.RequestTimeout, func() {
		c.timeoutRequest(id)
	})
	c.stateMu.Unlock()
	err := c.ws.WriteJSON(request)
	c.writeMu.Unlock()
	if err == nil {
		return id, nil
	}

	c.removeRequest(id)
	err = fmt.Errorf("failed to write JSON-RPC WebSocket request: %w", err)
	c.fail(err)
	return "", err
}

// sendNotification waits for the current socket write to finish and then
// writes request without an ID. It returns an error if ctx is canceled, the
// method stream is closed, or the socket write fails.
func (c *{{ .WebSocketConnection.Name }}) sendNotification(ctx context.Context, request *jsonrpc.Request, owner *{{ .WebSocketRequestOwner.Name }}) error {
	c.writeMu.Lock()
	select {
	case <-ctx.Done():
		err := ctx.Err()
		if owner.closed.Load() {
			err = {{ .WebSocketClosedError.Name }}
		}
		c.writeMu.Unlock()
		return err
	default:
	}
	c.stateMu.Lock()
	switch {
	case c.err != nil:
		err := c.err
		c.stateMu.Unlock()
		c.writeMu.Unlock()
		return err
	case owner.closed.Load():
		c.stateMu.Unlock()
		c.writeMu.Unlock()
		return {{ .WebSocketClosedError.Name }}
	}
	c.stateMu.Unlock()
	err := c.ws.WriteJSON(request)
	c.writeMu.Unlock()
	if err != nil {
		err = fmt.Errorf("failed to write JSON-RPC WebSocket notification: %w", err)
		c.fail(err)
		return err
	}
	return nil
}

// readResponses reads every message from the shared WebSocket. No other
// function reads from that socket. It reports server notifications and uses
// each response ID to find the function that receives the request result.
func (c *{{ .WebSocketConnection.Name }}) readResponses() {
	for {
		var message {{ .WebSocketMessage.Name }}
		if err := c.ws.ReadJSON(&message); err != nil {
			c.fail(fmt.Errorf("failed to read JSON-RPC WebSocket message: %w", err))
			return
		}
		if message.Method != "" {
			c.handleIncomingMethod(&message)
			continue
		}
		c.handleIncomingResponse(&message)
	}
}

// handleIncomingMethod reports the name of a server notification. A message
// with an ID, including null, is a server request that this client does not
// support.
func (c *{{ .WebSocketConnection.Name }}) handleIncomingMethod(message *{{ .WebSocketMessage.Name }}) {
	if len(message.ID) > 0 {
		c.handleError(c.ctx, jsonrpc.StreamErrorProtocol, fmt.Errorf("received unsupported JSON-RPC WebSocket server request %q", message.Method), nil)
		return
	}
	c.handleError(c.ctx, jsonrpc.StreamErrorNotification, fmt.Errorf("received JSON-RPC WebSocket notification %q", message.Method), nil)
}

// handleIncomingResponse checks the response ID and passes the unchanged
// result or error to the function stored under that ID.
func (c *{{ .WebSocketConnection.Name }}) handleIncomingResponse(message *{{ .WebSocketMessage.Name }}) {
	response := &jsonrpc.RawResponse{
		JSONRPC: message.JSONRPC,
		Result:  message.Result,
		Error:   message.Error,
	}
	if len(message.ID) == 0 {
		c.handleError(c.ctx, jsonrpc.StreamErrorProtocol, fmt.Errorf("received JSON-RPC WebSocket response without an ID"), response)
		return
	}
	if string(message.ID) == "null" {
		c.handleError(c.ctx, jsonrpc.StreamErrorProtocol, fmt.Errorf("received JSON-RPC WebSocket response with a null ID"), response)
		return
	}
	if err := json.Unmarshal(message.ID, &response.ID); err != nil {
		c.handleError(c.ctx, jsonrpc.StreamErrorParsing, fmt.Errorf("decode JSON-RPC WebSocket response ID: %w", err), response)
		return
	}
	id := jsonrpc.IDToString(response.ID)
	pending := c.removeRequest(id)
	if pending == nil {
		c.handleError(c.ctx, jsonrpc.StreamErrorOrphaned, fmt.Errorf("received JSON-RPC WebSocket response for unknown ID %q", id), response)
		return
	}
	pending.complete(pending.ctx, response, nil)
}

// timeoutRequest removes the request with id, reports its timeout, and calls
// the function waiting for its result even if the method stream has not called
// Recv. It does nothing if a response, cancellation, or close already removed
// the request.
func (c *{{ .WebSocketConnection.Name }}) timeoutRequest(id string) {
	c.stateMu.Lock()
	pending := c.pending[id]
	if pending != nil {
		delete(c.pending, id)
	}
	c.stateMu.Unlock()
	if pending == nil {
		return
	}
	err := fmt.Errorf("JSON-RPC WebSocket request timed out after %v", c.config.RequestTimeout)
	c.handleError(pending.ctx, jsonrpc.StreamErrorTimeout, err, nil)
	pending.complete(pending.ctx, nil, err)
}

// removeRequest removes one request and stops its timer. It returns nil if a
// response, timeout, cancellation, or close already removed the request.
func (c *{{ .WebSocketConnection.Name }}) removeRequest(id string) *{{ .WebSocketPendingRequest.Name }} {
	c.stateMu.Lock()
	pending := c.pending[id]
	if pending != nil {
		delete(c.pending, id)
		pending.timer.Stop()
	}
	c.stateMu.Unlock()
	return pending
}

// cancelRequest removes the request with id and passes err to the function
// waiting for its result. It returns whether it found and canceled the request.
func (c *{{ .WebSocketConnection.Name }}) cancelRequest(id string, err error) bool {
	pending := c.removeRequest(id)
	if pending == nil {
		return false
	}
	pending.complete(pending.ctx, nil, err)
	return true
}

// closeOwner marks owner closed and passes {{ .WebSocketClosedError.Name }} to
// every request sent by that method stream. Requests from other method streams
// continue on the same socket.
func (c *{{ .WebSocketConnection.Name }}) closeOwner(owner *{{ .WebSocketRequestOwner.Name }}) {
	c.stateMu.Lock()
	if owner.closed.Swap(true) {
		c.stateMu.Unlock()
		return
	}
	var canceled []*{{ .WebSocketPendingRequest.Name }}
	for id, pending := range c.pending {
		if pending.owner == owner {
			delete(c.pending, id)
			pending.timer.Stop()
			canceled = append(canceled, pending)
		}
	}
	c.stateMu.Unlock()
	for _, pending := range canceled {
		pending.complete(pending.ctx, nil, {{ .WebSocketClosedError.Name }})
	}
}

// terminalError returns the error that closed the connection.
func (c *{{ .WebSocketConnection.Name }}) terminalError() error {
	c.stateMu.Lock()
	defer c.stateMu.Unlock()
	return c.err
}

// fail records the first socket error, closes the socket to interrupt a blocked
// read or write, reports the error, and passes it to every function waiting for
// a request result. Later calls do nothing because the first failure already
// closed the socket.
func (c *{{ .WebSocketConnection.Name }}) fail(err error) {
	pending, ended := c.beginEnd(err)
	if !ended {
		return
	}
	if closeErr := c.closeSocket(); closeErr != nil {
		err = fmt.Errorf("%w; close JSON-RPC WebSocket: %v", err, closeErr)
	}
	c.finishEnd(err)
	c.handleError(c.ctx, jsonrpc.StreamErrorConnection, err, nil)
	for _, request := range pending {
		request.complete(request.ctx, nil, err)
	}
}

// beginEnd records err and returns every request that was waiting for a
// response. Its boolean result is false if another call already recorded the
// connection error. The caller closes the socket and calls the waiting request
// functions after the shared request map is unlocked.
func (c *{{ .WebSocketConnection.Name }}) beginEnd(err error) ([]*{{ .WebSocketPendingRequest.Name }}, bool) {
	c.stateMu.Lock()
	defer c.stateMu.Unlock()
	if c.err != nil {
		return nil, false
	}
	c.err = err
	pending := make([]*{{ .WebSocketPendingRequest.Name }}, 0, len(c.pending))
	for id, request := range c.pending {
		delete(c.pending, id)
		request.timer.Stop()
		pending = append(pending, request)
	}
	return pending, true
}

// finishEnd replaces the connection error with err and closes done so Recv
// calls know that the socket has closed.
func (c *{{ .WebSocketConnection.Name }}) finishEnd(err error) {
	c.stateMu.Lock()
	c.err = err
	close(c.done)
	c.stateMu.Unlock()
}

// closeSocket closes ws exactly once and returns the socket close error. It
// does not wait for a current WriteJSON call to finish, because closing the
// network connection must interrupt that blocked write.
func (c *{{ .WebSocketConnection.Name }}) closeSocket() error {
	c.closeOnce.Do(func() {
		c.closeErr = c.ws.Close()
	})
	return c.closeErr
}

// handleError passes errorType, err, and response to the configured error
// function. Request errors use the request context and socket errors use the
// connection context. Callers must finish changing the shared connection state
// first because user code may call back into the client.
func (c *{{ .WebSocketConnection.Name }}) handleError(ctx context.Context, errorType jsonrpc.StreamErrorType, err error, response *jsonrpc.RawResponse) {
	if c.config.ErrorHandler != nil {
		c.config.ErrorHandler(ctx, errorType, err, response)
	}
}

// close closes the shared socket, passes a connection-closed error to every
// function waiting for a request result, and returns the socket close error.
func (c *{{ .WebSocketConnection.Name }}) close() error {
	err := fmt.Errorf("JSON-RPC WebSocket connection closed")
	pending, ended := c.beginEnd(err)
	if !ended {
		return c.closeSocket()
	}
	closeErr := c.closeSocket()
	c.finishEnd(err)
	for _, request := range pending {
		request.complete(request.ctx, nil, err)
	}
	return closeErr
}

// Close rejects future client calls, closes the shared WebSocket, and returns
// the socket close error.
func (c *{{ .ClientStructDeclaration.Name }}) Close() error {
	if c.closed.Swap(true) {
		return nil
	}
	c.connMu.Lock()
	conn := c.conn
	c.conn = nil
	c.connMu.Unlock()
	if conn == nil {
		return nil
	}
	return conn.close()
}

// IsClosed reports whether Close has closed this client.
func (c *{{ .ClientStructDeclaration.Name }}) IsClosed() bool {
	return c.closed.Load()
}
