{{ printf "%sClientStream implements the %s.%sClientStream interface using Server-Sent Events." .Method.VarName .ServicePkgName .Method.VarName | comment }}
type {{ .Method.VarName }}ClientStream struct {
	resp    *http.Response  // HTTP response object
	reader  *bufio.Reader   // Buffered reader for SSE parsing
	decoder func(*http.Response) goahttp.Decoder  // User-provided decoder
	closed  bool            // Whether the stream has been closed
	lock    sync.Mutex      // Mutex to protect state
}

// parseSSEEvent parses a single SSE event from the stream
func (s *{{ .Method.VarName }}ClientStream) parseSSEEvent() (eventType string, data []byte, err error) {
	var event strings.Builder
	var dataLines []string
	
	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF && len(dataLines) > 0 {
				// Process final event
				break
			}
			return "", nil, err
		}
		
		line = strings.TrimSuffix(line, "\n")
		line = strings.TrimSuffix(line, "\r")
		
		if line == "" {
			// Empty line marks end of event
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
		// Ignore other fields like id:, retry:
	}
	
	if len(dataLines) > 0 {
		data = []byte(strings.Join(dataLines, "\n"))
	}
	
	return event.String(), data, nil
}

{{ comment .Method.ClientStream.RecvDesc }}
func (s *{{ .Method.VarName }}ClientStream) {{ .Method.ClientStream.RecvName }}(ctx context.Context) ({{ .Result.Ref }}, error) {
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
				return zero, fmt.Errorf("JSON-RPC error %d: %s", response.Error.Code, response.Error.Message)
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
				return zero, fmt.Errorf("JSON-RPC error %d: %s", response.Error.Code, response.Error.Message)
			}
			return zero, fmt.Errorf("unexpected error response")
			
		default:
			// Ignore unknown event types
			continue
		}
	}
}

{{- if .Method.Result }}
// decodeResult decodes JSON-RPC result data using the user-provided decoder
func (s *{{ .Method.VarName }}ClientStream) decodeResult(data json.RawMessage) ({{ .Result.Ref }}, error) {
	// Create minimal HTTP response with raw JSON data for user's decoder
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(data)),
	}
	
	// Use the user-provided decoder to decode the result
	decoder := s.decoder(resp)
	var result {{ .Result.Ref }}
	if err := decoder.Decode(&result); err != nil {
		return result, err
	}
	
	return result, nil
}
{{- end }}

{{ comment "Close closes the stream." }}
func (s *{{ .Method.VarName }}ClientStream) Close() error {
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