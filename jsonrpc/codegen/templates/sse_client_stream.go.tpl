type (
	{{ printf "%s reads results sent as server-sent events." .SSE.ClientInterfaceDeclaration.Name | comment }}
	{{ .SSE.ClientInterfaceDeclaration.Name }} interface {
		{{ .Method.ClientStream.RecvName }}() ({{ .SSE.EventTypeRef }}, error)
		{{ .Method.ClientStream.RecvWithContextName }}(context.Context) ({{ .SSE.EventTypeRef }}, error)
		Close() error
	}

	{{ printf "%s reads and decodes events for %s." .SSE.ClientStructDeclaration.Name .Method.Name | comment }}
	{{ .SSE.ClientStructDeclaration.Name }} struct {
		// resp is the open server response.
		resp *http.Response
		// reader reads one line at a time from resp.
		reader *bufio.Reader
		// decoder converts each result into its service type.
		decoder func(*http.Response) goahttp.Decoder
		// closed records whether Close was called or the response ended.
		closed bool
		// closeOnce ensures the response body is closed only once.
		closeOnce sync.Once
		// closeErr stores the response body close error.
		closeErr error
		// lock prevents two calls from reading or closing the response at once.
		lock sync.Mutex
	}
)

{{ printf "%s creates a stream that reads server-sent events from resp." .SSE.ClientInitDeclaration.Name | comment }}
func {{ .SSE.ClientInitDeclaration.Name }}(resp *http.Response, decoder func(*http.Response) goahttp.Decoder) {{ .SSE.ClientInterfaceDeclaration.Name }} {
	return &{{ .SSE.ClientStructDeclaration.Name }}{
		resp:    resp,
		reader:  bufio.NewReader(resp.Body),
		decoder: decoder,
	}
}

// parseSSEEvent reads one complete event from the response. Ending ctx closes
// the response body so a blocked read returns.
func (s *{{ .SSE.ClientStructDeclaration.Name }}) parseSSEEvent(ctx context.Context) (eventType string, data []byte, err error) {
	closeResult := make(chan struct{}, 1)
	stopClose := context.AfterFunc(ctx, func() {
		s.closeBody()
		closeResult <- struct{}{}
	})
	defer func() {
		if stopClose() {
			return
		}
		<-closeResult
		if contextErr := ctx.Err(); contextErr != nil {
			eventType = ""
			data = nil
			err = contextErr
		}
	}()
	var event strings.Builder
	var dataLines []string
	
	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF && len(dataLines) > 0 {
				// Return the last event even when the response has no final blank line.
				break
			}
			return "", nil, err
		}
		
		line = strings.TrimSuffix(line, "\n")
		line = strings.TrimSuffix(line, "\r")
		
		if line == "" {
			// A blank line ends the current event.
			if len(dataLines) > 0 {
				break
			}
			continue
		}
		
		if strings.HasPrefix(line, "event:") {
			event.WriteString(strings.TrimSpace(line[6:]))
		} else if strings.HasPrefix(line, "data:") {
			dataLines = append(dataLines, strings.TrimSpace(line[5:]))
		}
		// This client does not use the id and retry fields.
	}
	
	if len(dataLines) > 0 {
		data = []byte(strings.Join(dataLines, "\n"))
	}
	
	return event.String(), data, nil
}

{{ comment .Method.ClientStream.RecvDesc }}
func (s *{{ .SSE.ClientStructDeclaration.Name }}) {{ .Method.ClientStream.RecvName }}() ({{ .SSE.EventTypeRef }}, error) {
	return s.{{ .Method.ClientStream.RecvWithContextName }}(context.Background())
}

{{ comment .Method.ClientStream.RecvWithContextDesc }}
func (s *{{ .SSE.ClientStructDeclaration.Name }}) {{ .Method.ClientStream.RecvWithContextName }}(ctx context.Context) ({{ .SSE.EventTypeRef }}, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	
	var zero {{ .SSE.EventTypeRef }}
	
	if s.closed {
		return zero, io.EOF
	}
	
	for {
		eventType, data, err := s.parseSSEEvent(ctx)
		if err != nil {
			return zero, s.endStream(err)
		}
		
		switch eventType {
		case "notification":
			// Read the streamed service result from the notification parameters.
			var notification struct {
				JSONRPC string          `json:"jsonrpc"`
				Method  string          `json:"method"`
				Params  json.RawMessage `json:"params"`
			}
			if err := json.Unmarshal(data, &notification); err != nil {
				return zero, s.endStream(fmt.Errorf("failed to parse notification: %w", err))
			}
			
			if notification.JSONRPC != "2.0" {
				return zero, s.endStream(fmt.Errorf("invalid JSON-RPC version: %s", notification.JSONRPC))
			}
			
			if notification.Method != {{ printf "%q" .Method.Name }} {
				return zero, s.endStream(fmt.Errorf("received notification for JSON-RPC method %q", notification.Method))
			}
			
			result, err := s.decodeResult(notification.Params)
			if err != nil {
				return zero, s.endStream(fmt.Errorf("failed to decode result: %w", err))
			}
			return result, nil
			
		case "response":
			// A successful response completes the stream. Stream values arrive in
			// the notifications handled above.
			var response jsonrpc.Response
			if err := json.Unmarshal(data, &response); err != nil {
				return zero, s.endStream(fmt.Errorf("failed to parse response: %w", err))
			}
			
			if response.Error != nil {
				return zero, s.endStream(response.Error)
			}
			return zero, s.endStream(io.EOF)
			
		case "error":
			// A JSON-RPC error completes the stream.
			var response jsonrpc.Response
			if err := json.Unmarshal(data, &response); err != nil {
				return zero, s.endStream(fmt.Errorf("failed to parse error response: %w", err))
			}
			if response.Error != nil {
				return zero, s.endStream(response.Error)
			}
			return zero, s.endStream(fmt.Errorf("JSON-RPC error event did not contain an error"))
			
		default:
			return zero, s.endStream(fmt.Errorf("unsupported server-sent event type %q", eventType))
		}
	}
}

// closeBody closes the HTTP response body once and returns its close error.
func (s *{{ .SSE.ClientStructDeclaration.Name }}) closeBody() error {
	s.closeOnce.Do(func() {
		s.closeErr = s.resp.Body.Close()
	})
	return s.closeErr
}

// endStream marks the stream closed and preserves both the receive error and
// any error returned while closing the HTTP response body.
func (s *{{ .SSE.ClientStructDeclaration.Name }}) endStream(err error) error {
	s.closed = true
	if closeErr := s.closeBody(); closeErr != nil {
		return errors.Join(err, closeErr)
	}
	return err
}

// decodeResult passes one successful stream item to the decoder configured by NewClient.
func (s *{{ .SSE.ClientStructDeclaration.Name }}) decodeResult(data json.RawMessage) ({{ .SSE.EventTypeRef }}, error) {
	{{- if .Method.ViewedResult }}
	// The HTTP 200 status tells the configured decoder that this stream item is
	// a successful JSON-RPC result. Streaming results cannot carry HTTP headers or cookies.
	resp := &http.Response{StatusCode: http.StatusOK, Header: make(http.Header)}
	return {{ viewedDecodeName .Method.Name }}(s.decoder, resp, data)
	{{- else }}
	// Give the configured decoder the successful result bytes as an HTTP response body.
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(data)),
	}
	
	decoder := s.decoder(resp)
	var result {{ .SSE.EventTypeRef }}
	if err := decoder.Decode(&result); err != nil {
		return result, err
	}
	
	return result, nil
	{{- end }}
}

{{ comment "Close closes the stream." }}
func (s *{{ .SSE.ClientStructDeclaration.Name }}) Close() error {
    s.lock.Lock()
    defer s.lock.Unlock()
    
	if !s.closed {
		s.closed = true
		return s.closeBody()
	}
    return nil
}
