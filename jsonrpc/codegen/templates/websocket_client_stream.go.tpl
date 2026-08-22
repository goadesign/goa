{{/*
This file writes one JSON-RPC method stream. One shared connection reads and
writes the socket and assigns request IDs. It finds the function waiting for
each response. This stream turns response fields into service results, returns
them in send order, and handles cancellation.
*/}}
{{- $hasRecv := and .RecvName .RecvTypeRef }}
{{- $hasSend := .SendName }}
{{- $isBidirectional := and $hasSend $hasRecv }}
{{- $pendingType := .Pending.Name }}
{{- $resultType := .Result.Name }}
{{ printf "%s implements the %s client stream." .VarDeclaration.Name .Endpoint.Method.Name | comment }}
type (
	{{ .VarDeclaration.Name }} struct {
		conn  *{{ .Connection.Name }}
		owner *{{ .RequestOwner.Name }}

		ctx       context.Context
		cancel    context.CancelFunc
		closeOnce sync.Once

		{{- if $hasRecv }}
		decoder func(*http.Response) goahttp.Decoder
		{{- end }}
		{{- if $isBidirectional }}

		sendMu       sync.Mutex
		pendingMu    sync.Mutex
		pending      []*{{ $pendingType }}
		pendingReady chan struct{}
		{{- end }}
	}

	{{- if $hasRecv }}
	// {{ $pendingType }} stores the channel that receives the result or error
	// for one request. The shared connection starts and stops its timer.
	{{ $pendingType }} struct {
		id         string
		resultChan chan {{ $resultType }}
	}

	// {{ $resultType }} contains the decoded result or error returned by one
	// request.
	{{ $resultType }} struct {
		result {{ .RecvTypeRef }}
		err    error
	}
	{{- end }}
)

{{- if $hasSend }}
{{ comment .SendDesc }}
func (s *{{ .VarDeclaration.Name }}) {{ .SendName }}(v {{ .SendTypeRef }}) error {
	return s.{{ .SendName }}WithContext(s.ctx, v)
}

{{ comment .SendWithContextDesc }}
func (s *{{ .VarDeclaration.Name }}) {{ .SendName }}WithContext(ctx context.Context, v {{ .SendTypeRef }}) error {
	request := &jsonrpc.Request{
		JSONRPC: "2.0",
		Method:  "{{ .Endpoint.Method.Name }}",
		Params:  v,
	}
	{{- if $isBidirectional }}
	pending := &{{ $pendingType }}{
		resultChan: make(chan {{ $resultType }}, 1),
	}

	s.sendMu.Lock()
	id, err := s.conn.sendRequest(ctx, request, s.owner, func(ctx context.Context, response *jsonrpc.RawResponse, err error) {
		s.completeResponse(ctx, pending, response, err)
	})
	if err == nil {
		pending.id = id
		s.enqueuePending(pending)
	}
	s.sendMu.Unlock()
	if err != nil {
		return err
	}
	return nil
	{{- else }}
	return s.conn.sendNotification(ctx, request, s.owner)
	{{- end }}
}
{{- end }}

{{- if $hasRecv }}
{{ comment .RecvDesc }}
func (s *{{ .VarDeclaration.Name }}) {{ .RecvName }}() ({{ .RecvTypeRef }}, error) {
	return s.{{ .RecvName }}WithContext(s.ctx)
}

{{ comment .RecvWithContextDesc }}
func (s *{{ .VarDeclaration.Name }}) {{ .RecvName }}WithContext(ctx context.Context) ({{ .RecvTypeRef }}, error) {
	{{- if $isBidirectional }}
	pending, err := s.nextPending(ctx)
	if err != nil {
		var zero {{ .RecvTypeRef }}
		return zero, err
	}
	return s.awaitPending(ctx, pending)
	{{- else }}
	request := &jsonrpc.Request{
		JSONRPC: "2.0",
		Method:  "{{ .Endpoint.Method.Name }}",
	}
	pending := &{{ $pendingType }}{
		resultChan: make(chan {{ $resultType }}, 1),
	}
	id, err := s.conn.sendRequest(ctx, request, s.owner, func(ctx context.Context, response *jsonrpc.RawResponse, err error) {
		s.completeResponse(ctx, pending, response, err)
	})
	if err != nil {
		var zero {{ .RecvTypeRef }}
		return zero, err
	}
	pending.id = id
	return s.awaitPending(ctx, pending)
	{{- end }}
}

// awaitPending waits for pending to receive a result or error. It returns the
// closed-stream error if Close runs, even when another cancellation is ready.
func (s *{{ .VarDeclaration.Name }}) awaitPending(ctx context.Context, pending *{{ $pendingType }}) ({{ .RecvTypeRef }}, error) {
	for {
		if s.owner.closed.Load() {
			var zero {{ .RecvTypeRef }}
			return zero, {{ .ClosedError.Name }}
		}
		select {
		case result := <-pending.resultChan:
			return result.result, s.methodStreamError(result.err)
		case <-ctx.Done():
			err := s.methodStreamError(ctx.Err())
			s.conn.cancelRequest(pending.id, err)
			var zero {{ .RecvTypeRef }}
			return zero, err
		case <-s.ctx.Done():
			err := s.methodStreamError(s.ctx.Err())
			s.conn.cancelRequest(pending.id, err)
			var zero {{ .RecvTypeRef }}
			return zero, err
		case <-s.conn.done:
			var zero {{ .RecvTypeRef }}
			return zero, s.methodStreamError(s.conn.terminalError())
		}
	}
}

// completeResponse turns response into this method's service result, or uses
// err when the request failed, and sends it to the Recv call waiting for pending.
func (s *{{ .VarDeclaration.Name }}) completeResponse(ctx context.Context, pending *{{ $pendingType }}, response *jsonrpc.RawResponse, err error) {
	var result {{ $resultType }}
	switch {
	case err != nil:
		result.err = err
	case response.Error != nil:
		result.err = response.Error
		s.conn.handleError(ctx, jsonrpc.StreamErrorProtocol, response.Error, response)
	default:
		parsedResult, decodeErr := s.decodeResponse(response.Result)
		if decodeErr != nil {
			result.err = fmt.Errorf("failed to decode JSON-RPC WebSocket response: %w", decodeErr)
			s.conn.handleError(ctx, jsonrpc.StreamErrorParsing, result.err, response)
		} else {
			{{- if .Endpoint.Result.IDAttribute }}
			{{- if .Endpoint.Result.IDAttributeRequired }}
			if parsedResult.{{ .Endpoint.Result.IDAttribute }} == "" {
				parsedResult.{{ .Endpoint.Result.IDAttribute }} = jsonrpc.IDToString(response.ID)
			}
			{{- else }}
			if parsedResult.{{ .Endpoint.Result.IDAttribute }} == nil || *parsedResult.{{ .Endpoint.Result.IDAttribute }} == "" {
				id := jsonrpc.IDToString(response.ID)
				parsedResult.{{ .Endpoint.Result.IDAttribute }} = &id
			}
			{{- end }}
			{{- end }}
			result.result = parsedResult
		}
	}
	pending.resultChan <- result
}

{{- if $isBidirectional }}
// enqueuePending adds pending to the requests waiting for Recv. It keeps their
// send order even when the server responds in a different order.
func (s *{{ .VarDeclaration.Name }}) enqueuePending(pending *{{ $pendingType }}) {
	s.pendingMu.Lock()
	s.pending = append(s.pending, pending)
	if s.owner.closed.Load() {
		s.pending = s.pending[:len(s.pending)-1]
		s.pendingMu.Unlock()
		return
	}
	s.pendingMu.Unlock()
	select {
	case s.pendingReady <- struct{}{}:
	default:
	}
}

// nextPending returns the first request sent by this method stream that has not
// yet been passed to Recv. It returns an error if the caller cancels, Close
// runs, or the socket fails first.
func (s *{{ .VarDeclaration.Name }}) nextPending(ctx context.Context) (*{{ $pendingType }}, error) {
	for {
		if s.owner.closed.Load() {
			return nil, {{ .ClosedError.Name }}
		}
		s.pendingMu.Lock()
		if len(s.pending) > 0 {
			if s.owner.closed.Load() {
				s.pendingMu.Unlock()
				return nil, {{ .ClosedError.Name }}
			}
			pending := s.pending[0]
			s.pending = s.pending[1:]
			s.pendingMu.Unlock()
			return pending, nil
		}
		s.pendingMu.Unlock()
		select {
		case <-s.pendingReady:
		case <-ctx.Done():
			return nil, s.methodStreamError(ctx.Err())
		case <-s.ctx.Done():
			return nil, s.methodStreamError(s.ctx.Err())
		case <-s.conn.done:
			return nil, s.methodStreamError(s.conn.terminalError())
		}
	}
}
{{- end }}

// methodStreamError returns the closed-stream error if Close has run.
// Otherwise it returns the supplied err unchanged.
func (s *{{ .VarDeclaration.Name }}) methodStreamError(err error) error {
	if s.owner.closed.Load() {
		return {{ .ClosedError.Name }}
	}
	return err
}

// decodeResponse reads data using this method's response format and returns the
// service result.
func (s *{{ .VarDeclaration.Name }}) decodeResponse(data json.RawMessage) ({{ .RecvTypeRef }}, error) {
	{{- if .Endpoint.Method.ViewedResult }}
	// The HTTP 200 status tells the configured decoder that this stream item is
	// a successful JSON-RPC result. Streaming results cannot carry HTTP headers or cookies.
	resp := &http.Response{StatusCode: http.StatusOK, Header: make(http.Header)}
	return {{ viewedDecodeName .Endpoint.Method.Name }}(s.decoder, resp, data)
	{{- else }}
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(data)),
	}
	dec := s.decoder(resp)
	var out {{ .RecvTypeRef }}
	if err := dec.Decode(&out); err != nil {
		var zero {{ .RecvTypeRef }}
		return zero, err
	}
	return out, nil
	{{- end }}
}
{{- end }}

{{ printf "Close closes the %s method stream without closing the WebSocket shared by other methods." .Endpoint.Method.Name | comment }}
func (s *{{ .VarDeclaration.Name }}) Close() error {
	s.closeOnce.Do(func() {
		s.conn.closeOwner(s.owner)
		{{- if $isBidirectional }}
		s.pendingMu.Lock()
		s.pending = nil
		s.pendingMu.Unlock()
		{{- end }}
		s.cancel()
	})
	return nil
}
