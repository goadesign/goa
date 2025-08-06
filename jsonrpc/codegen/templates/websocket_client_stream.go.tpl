{{/*
websocket_client_stream.go.tpl generates JSON-RPC WebSocket streaming client implementations.

This template creates stream types that handle direct WebSocket connections for JSON-RPC
streaming endpoints, providing:
- Direct WebSocket transport without intermediate wrappers
- Dual ID correlation (user payload ID + JSON-RPC request ID)
- Comprehensive error handling with user-configurable error handlers
- Generated decoder integration for consistent response parsing
- Thread-safe operations with proper lifecycle management

Template variables:
- .VarName: Name of the generated stream struct
- .Endpoint.Method.Name: Name of the endpoint method
- .SendName/.SendTypeRef: Send method name and payload type (if stream accepts input)
- .RecvName/.RecvTypeRef: Receive method name and result type (if stream produces output)
- .Endpoint.ServiceVarName: Service name for JSON-RPC method naming

The template handles three streaming patterns:
1. Client streaming (send-only): $hasSend && !$hasRecv
2. Server streaming (recv-only): !$hasSend && $hasRecv  
3. Bidirectional streaming: $isBidirectional ($hasSend && $hasRecv)
*/}}
{{ printf "%s implements the %s client stream with direct WebSocket handling." .VarName .Endpoint.Method.Name | comment }}
{{- $hasRecv := and .RecvName .RecvTypeRef }}
{{- $hasSend := .SendName }}
{{- $isBidirectional := and $hasSend $hasRecv }}
type {{ .VarName }} struct {
	// Direct WebSocket transport
	ws          *websocket.Conn
	writeMu     sync.Mutex              // Serialize WebSocket writes
	
	// JSON-RPC correlation  
	pending     sync.Map                // map[jsonrpcID]*{{ .VarName }}PendingRequest
	idGenerator atomic.Uint64           // JSON-RPC request ID generator
	
	// Lifecycle management
	ctx         context.Context
	cancel      context.CancelFunc
	done        chan struct{}           // Signals stream closure
	closeOnce   sync.Once
	
	// Error handling
	errorOnce   sync.Once
	lastError   atomic.Value            // Last error encountered
	
	// Stream configuration
	config *jsonrpc.StreamConfig // Stream configuration options
	{{- if $hasRecv }}
	decoder        func(*http.Response) (any, error) // Pre-computed decoder for responses
	{{- end }}
}


// Stream-specific types for {{ .VarName }}
type {{ .VarName }}PendingRequest struct {
	userID      string                  // User-provided payload ID
	resultChan  chan {{ .VarName }}StreamResult    // Buffered result delivery
	timeout     *time.Timer             // Request timeout handling
}

type {{ .VarName }}StreamResult struct {
{{- if $hasRecv }}
	result      {{ .RecvTypeRef }}
{{- end }}
	err         error
}

{{- if $hasSend }}
{{ printf "%s sends streaming data to the %s endpoint with dual ID correlation." .SendName .Endpoint.Method.Name | comment }}
func (s *{{ .VarName }}) {{ .SendName }}(v {{ .SendTypeRef }}) error {
	return s.{{ .SendName }}WithContext(s.ctx, v)
}

{{ printf "%sWithContext sends streaming data to the %s endpoint with context." .SendName .Endpoint.Method.Name | comment }}
func (s *{{ .VarName }}) {{ .SendName }}WithContext(ctx context.Context, v {{ .SendTypeRef }}) error {
	// Check for stream-level errors first
	if err := s.getError(); err != nil {
		return err
	}
	
{{- if $isBidirectional }}
	// Honor user-provided ID or generate one
	userID := ""
{{- if .SendTypeRef }}
	{{- if .Endpoint.Payload }}
	// Honor user-provided ID if it exists in the payload
	userID = s.generateUserID()
	{{- end }}
{{- else }}
	userID = s.generateUserID()
{{- end }}
	
	// Generate JSON-RPC protocol ID
	jsonrpcID := strconv.FormatUint(s.idGenerator.Add(1), 10)
	// Create pending request tracking for bidirectional streaming
	pending := &{{ .VarName }}PendingRequest{
		userID:     userID,
		resultChan: make(chan {{ .VarName }}StreamResult, s.config.ResultChannelBuffer),
		timeout:    time.NewTimer(s.config.RequestTimeout),
	}
	
	s.pending.Store(jsonrpcID, pending)
	
	// Construct JSON-RPC request
	request := &jsonrpc.Request{
		JSONRPC: "2.0",
		Method:  "{{ .Endpoint.Method.Name }}",
		Params:  v,
		ID:      &jsonrpcID,
	}
{{- else }}
	// For payload-only streaming, use notification (fire-and-forget)
	request := &jsonrpc.Request{
		JSONRPC: "2.0",
		Method:  "{{ .Endpoint.Method.Name }}",
		Params:  v,
		// No ID field for notifications
	}
{{- end }}
	
	// Send with write protection
	s.writeMu.Lock()
	err := s.ws.WriteJSON(request)
	s.writeMu.Unlock()
	
	if err != nil {
{{- if $isBidirectional }}
		s.pending.Delete(jsonrpcID)
		pending.timeout.Stop()
{{- end }}
		s.setError(err)
		// Report connection errors
		s.handleError(jsonrpc.StreamErrorConnection, err, nil)
		return fmt.Errorf("failed to send request: %w", err)
	}
	
	return nil
}
{{- end }}

{{- if $hasRecv }}
{{ printf "%s receives streaming data from the %s endpoint." .RecvName .Endpoint.Method.Name | comment }}
func (s *{{ .VarName }}) {{ .RecvName }}() ({{ .RecvTypeRef }}, error) {
	return s.{{ .RecvName }}WithContext(s.ctx)
}

{{ printf "%sWithContext receives streaming data from the %s endpoint with context." .RecvName .Endpoint.Method.Name | comment }}
func (s *{{ .VarName }}) {{ .RecvName }}WithContext(ctx context.Context) ({{ .RecvTypeRef }}, error) {
	// Check for stream-level errors first
	if err := s.getError(); err != nil {
		return nil, err
	}
	
{{- if $isBidirectional }}
	// Find the oldest pending request (FIFO ordering)
	var oldestPending *{{ .VarName }}PendingRequest
	var oldestKey string
	
	s.pending.Range(func(key, value any) bool {
		pending := value.(*{{ .VarName }}PendingRequest)
		if oldestPending == nil {
			oldestPending = pending
			oldestKey = key.(string)
		}
		return false // Take first one for FIFO
	})
	
	if oldestPending == nil {
		return nil, fmt.Errorf("no pending requests - call {{ .SendName }}() first")
	}
	
	// Wait for result with context cancellation
	select {
	case result := <-oldestPending.resultChan:
		s.pending.Delete(oldestKey)
		oldestPending.timeout.Stop()
		return result.result, result.err
		
	case <-oldestPending.timeout.C:
		s.pending.Delete(oldestKey)
		timeoutErr := fmt.Errorf("request timeout after %v", s.config.RequestTimeout)
		// Report timeout errors
		s.handleError(jsonrpc.StreamErrorTimeout, timeoutErr, nil)
		return nil, timeoutErr
		
	case <-ctx.Done():
		return nil, ctx.Err()
		
	case <-s.done:
		if err := s.getError(); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("stream closed")
	}
{{- else }}
	// For result-only streaming, make direct call
	jsonrpcID := strconv.FormatUint(s.idGenerator.Add(1), 10)
	
	request := &jsonrpc.Request{
		JSONRPC: "2.0",
		Method:  "{{ .Endpoint.Method.Name }}",
		Params:  nil,
		ID:      &jsonrpcID,
	}
	
	// Create result channel for this request
	resultChan := make(chan {{ .VarName }}StreamResult, s.config.ResultChannelBuffer)
	pending := &{{ .VarName }}PendingRequest{
		userID:     jsonrpcID,
		resultChan: resultChan,
		timeout:    time.NewTimer(s.config.RequestTimeout),
	}
	
	s.pending.Store(jsonrpcID, pending)
	defer func() {
		s.pending.Delete(jsonrpcID)
		pending.timeout.Stop()
	}()
	
	// Send request
	s.writeMu.Lock()
	err := s.ws.WriteJSON(request)
	s.writeMu.Unlock()
	
	if err != nil {
		s.setError(err)
		// Report connection errors
		s.handleError(jsonrpc.StreamErrorConnection, err, nil)
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	
	// Wait for response
	select {
	case result := <-resultChan:
		return result.result, result.err
	case <-pending.timeout.C:
		timeoutErr := fmt.Errorf("request timeout after %v", s.config.RequestTimeout)
		// Report timeout errors
		s.handleError(jsonrpc.StreamErrorTimeout, timeoutErr, nil)
		return nil, timeoutErr
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-s.done:
		if err := s.getError(); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("stream closed")
	}
{{- end }}
}
{{- end }}

// responseHandler processes incoming WebSocket messages in a background goroutine
func (s *{{ .VarName }}) responseHandler() {
	defer close(s.done)
	
	for {
		select {
		case <-s.ctx.Done():
			s.cleanupPendingRequests(s.ctx.Err())
			return
		default:
			var response jsonrpc.RawResponse
			if err := s.ws.ReadJSON(&response); err != nil {
				connectionErr := fmt.Errorf("failed to read response: %w", err)
				s.setError(connectionErr)
				
				// Report connection errors
				s.handleError(jsonrpc.StreamErrorConnection, connectionErr, nil)
				
				s.cleanupPendingRequests(connectionErr)
				return
			}
			
			s.handleResponse(&response)
		}
	}
}

func (s *{{ .VarName }}) handleResponse(response *jsonrpc.RawResponse) {
	if response.ID == "" {
		// This is a server-initiated notification
		// For now, just report it as an event via the error handler
		// In the future, we could add a dedicated notification handler
		if s.config.ErrorHandler != nil {
			s.config.ErrorHandler(s.ctx, jsonrpc.StreamErrorNotification, 
				fmt.Errorf("received server notification"), response)
		}
		return
	}
	
	jsonrpcID := response.ID
	pendingInterface, exists := s.pending.LoadAndDelete(jsonrpcID)
	if !exists {
		// Orphaned response - report to error handler
		s.handleError(jsonrpc.StreamErrorOrphaned, fmt.Errorf("received response for unknown ID: %s", jsonrpcID), response)
		return
	}
	
	pending := pendingInterface.(*{{ .VarName }}PendingRequest)
	pending.timeout.Stop()
	
	var result {{ .VarName }}StreamResult
	
	if response.Error != nil {
		result.err = response.Error
		// Report protocol-level JSON-RPC errors
		s.handleError(jsonrpc.StreamErrorProtocol, response.Error, response)
	} else {
{{- if $hasRecv }}
		// Use generated decoder for consistent response parsing
		parsedResult, err := s.decodeResponse(response.Result)
		if err != nil {
			result.err = fmt.Errorf("failed to decode response: %w", err)
			// Report parsing errors
			s.handleError(jsonrpc.StreamErrorParsing, err, response)
		} else {
			{{- if .Endpoint.Result.IDAttribute }}
			// Set the ID from the JSON-RPC envelope into the result
			if parsedResult.{{ .Endpoint.Result.IDAttribute }} == "" {
				parsedResult.{{ .Endpoint.Result.IDAttribute }} = response.ID
			}
			{{- end }}
			result.result = parsedResult
		}
{{- end }}
	}
	
	// Non-blocking send to result channel
	select {
	case pending.resultChan <- result:
	default:
		// Channel full - should not happen with buffer size 1
	}
}

// Helper methods
func (s *{{ .VarName }}) generateUserID() string {
	return fmt.Sprintf("user-%d-%d", time.Now().UnixNano(), s.idGenerator.Load())
}

// handleError calls the user-provided error handler if available
func (s *{{ .VarName }}) handleError(errorType jsonrpc.StreamErrorType, err error, response *jsonrpc.RawResponse) {
	if s.config.ErrorHandler != nil {
		s.config.ErrorHandler(s.ctx, errorType, err, response)
	}
}


{{- if $hasRecv }}
// decodeResponse decodes JSON-RPC response data using the user-provided decoder
func (s *{{ .VarName }}) decodeResponse(data json.RawMessage) ({{ .RecvTypeRef }}, error) {
	// For WebSocket, we need to inject a dummy ID into the result data
	// because the decoder expects it, but it actually comes from the envelope
	
	// First decode to check what we have
	var temp map[string]json.RawMessage
	if err := json.Unmarshal(data, &temp); err != nil {
		return nil, fmt.Errorf("failed to pre-decode response: %w", err)
	}
	
	// If there's no ID field, inject a dummy one for the decoder
	if _, hasID := temp["id"]; !hasID {
		temp["id"] = json.RawMessage(`""`) // Empty string as placeholder
	}
	
	// Re-encode with the ID field
	modifiedData, err := json.Marshal(temp)
	if err != nil {
		return nil, fmt.Errorf("failed to re-encode response: %w", err)
	}
	
	// Create a minimal JSON-RPC response wrapper for the decoder
	wrappedResponse := jsonrpc.RawResponse{
		JSONRPC: "2.0",
		Result:  modifiedData,
	}
	
	// Marshal it back to JSON 
	wrappedJSON, err := json.Marshal(wrappedResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to wrap response: %w", err)
	}
	
	// Create minimal HTTP response with the wrapped JSON for the decoder
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(wrappedJSON)),
	}
	
	// Use the pre-computed decoder function (contains user's decoder + validation logic)
	decodedResult, err := s.decoder(resp)
	if err != nil {
		return nil, err
	}
	
	// Type assert to the expected result type
	if result, ok := decodedResult.({{ .RecvTypeRef }}); ok {
		return result, nil
	}
	
	return nil, fmt.Errorf("unexpected response type: %T", decodedResult)
}
{{- end }}

func (s *{{ .VarName }}) setError(err error) {
	s.errorOnce.Do(func() {
		s.lastError.Store(err)
		s.cancel() // Cancel context to signal error state
	})
}

func (s *{{ .VarName }}) getError() error {
	if err, ok := s.lastError.Load().(error); ok {
		return err
	}
	return nil
}

func (s *{{ .VarName }}) cleanupPendingRequests(err error) {
	s.pending.Range(func(key, value any) bool {
		pending := value.(*{{ .VarName }}PendingRequest)
		pending.timeout.Stop()
		
		select {
		case pending.resultChan <- {{ .VarName }}StreamResult{err: err}:
		default:
		}
		
		s.pending.Delete(key)
		return true
	})
}

{{ printf "Close closes the stream and cleans up resources." | comment }}
func (s *{{ .VarName }}) Close() error {
	var err error
	s.closeOnce.Do(func() {
		s.cancel()
		
		// Wait for response handler to finish
		select {
		case <-s.done:
		case <-time.After(s.config.CloseTimeout):
			// Force close if handler doesn't respond
		}
		
		// Clean up any remaining pending requests
		s.cleanupPendingRequests(fmt.Errorf("stream closed"))
		
		// Close the WebSocket connection
		if s.ws != nil {
			err = s.ws.Close()
		}
	})
	return err
}
