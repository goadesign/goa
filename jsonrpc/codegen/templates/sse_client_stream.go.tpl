type (
	{{ printf "%s reads results sent as server-sent events." .SSE.ClientInterfaceDeclaration.Name | comment }}
	{{ .SSE.ClientInterfaceDeclaration.Name }} interface {
		{{ .Method.ClientStream.RecvName }}() ({{ .Result.Ref }}, error)
		{{ .Method.ClientStream.RecvWithContextName }}(context.Context) ({{ .Result.Ref }}, error)
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

// parseSSEEvent reads one complete event from the response.
func (s *{{ .SSE.ClientStructDeclaration.Name }}) parseSSEEvent() (eventType string, data []byte, err error) {
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
func (s *{{ .SSE.ClientStructDeclaration.Name }}) {{ .Method.ClientStream.RecvName }}() ({{ .Result.Ref }}, error) {
	return s.{{ .Method.ClientStream.RecvWithContextName }}(context.Background())
}

{{ comment .Method.ClientStream.RecvWithContextDesc }}
func (s *{{ .SSE.ClientStructDeclaration.Name }}) {{ .Method.ClientStream.RecvWithContextName }}(_ context.Context) ({{ .Result.Ref }}, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	
	var zero {{ .Result.Ref }}
	
	if s.closed {
		return zero, io.EOF
	}
	
	for {
		eventType, data, err := s.parseSSEEvent()
		if err != nil {
			s.closed = true
			return zero, err
		}
		
		switch eventType {
		case "notification":
			// Parse JSON-RPC notification
			var notification struct {
				JSONRPC string          `json:"jsonrpc"`
				Method  string          `json:"method"`
				Params  json.RawMessage `json:"params"`
			}
			if err := json.Unmarshal(data, &notification); err != nil {
				return zero, fmt.Errorf("failed to parse notification: %w", err)
			}
			
			// Validate notification
			if notification.JSONRPC != "2.0" {
				return zero, fmt.Errorf("invalid JSON-RPC version: %s", notification.JSONRPC)
			}
			
			if notification.Method != {{ printf "%q" .Method.Name }} {
				// Skip notifications for other methods
				continue
			}
			
			// Decode the result from params
			{{- if .Method.Result }}
			result, err := s.decodeResult(notification.Params)
			if err != nil {
				return zero, fmt.Errorf("failed to decode result: %w", err)
			}
			return result, nil
			{{- else }}
			// Method has no result
			return zero, nil
			{{- end }}
			
		case "response":
			// Final response - parse and return
			var response jsonrpc.Response
			if err := json.Unmarshal(data, &response); err != nil {
				return zero, fmt.Errorf("failed to parse response: %w", err)
			}
			
			if response.Error != nil {
				return zero, response.Error
			}
			
			{{- if .Method.Result }}
			// Decode the final result
			if response.Result == nil {
				return zero, fmt.Errorf("missing result in response")
			}
			// Convert response.Result to json.RawMessage
			resultBytes, err := json.Marshal(response.Result)
			if err != nil {
				return zero, fmt.Errorf("failed to marshal result: %w", err)
			}
			result, err := s.decodeResult(json.RawMessage(resultBytes))
			if err != nil {
				return zero, fmt.Errorf("failed to decode final result: %w", err)
			}
			
			// Mark stream as closed after final response
			s.closed = true
			return result, nil
			{{- else }}
			// Method has no result
			s.closed = true
			return zero, nil
			{{- end }}
			
		case "error":
			// Error response
			var response jsonrpc.Response
			if err := json.Unmarshal(data, &response); err != nil {
				return zero, fmt.Errorf("failed to parse error response: %w", err)
			}
			
			s.closed = true
			if response.Error != nil {
				return zero, response.Error
			}
			return zero, fmt.Errorf("unexpected error response")
			
		default:
			// Ignore unknown event types
			continue
		}
	}
}

{{- if .Method.Result }}
// decodeResult passes one successful stream item to the decoder configured by NewClient.
func (s *{{ .SSE.ClientStructDeclaration.Name }}) decodeResult(data json.RawMessage) ({{ .Result.Ref }}, error) {
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
	var result {{ .Result.Ref }}
	if err := decoder.Decode(&result); err != nil {
		return result, err
	}
	
	return result, nil
	{{- end }}
}
{{- end }}

{{ comment "Close closes the stream." }}
func (s *{{ .SSE.ClientStructDeclaration.Name }}) Close() error {
    s.lock.Lock()
    defer s.lock.Unlock()
    
    if !s.closed {
        s.closed = true
        if s.resp != nil && s.resp.Body != nil {
            return s.resp.Body.Close()
        }
    }
    return nil
}
