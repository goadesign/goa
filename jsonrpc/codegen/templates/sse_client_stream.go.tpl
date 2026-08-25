{{- $checksNullID := false }}
{{- range .Errors }}
	{{- if or (eq .StatusCode "-32700") (eq .StatusCode "-32600") }}
		{{- $checksNullID = true }}
	{{- end }}
{{- end }}
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
		// requestID is the ID every final response must repeat.
		requestID string
		// lastEventID is the last valid id field read from the stream.
		lastEventID string
		// hasLastEventID distinguishes no id from an explicitly empty id.
		hasLastEventID bool
		// firstLine is true until the optional stream byte-order marker is read.
		firstLine bool
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

{{ printf "%s creates a stream that reads server-sent events from resp. Final responses must repeat requestID." .SSE.ClientInitDeclaration.Name | comment }}
func {{ .SSE.ClientInitDeclaration.Name }}(resp *http.Response, decoder func(*http.Response) goahttp.Decoder, requestID string) {{ .SSE.ClientInterfaceDeclaration.Name }} {
	return &{{ .SSE.ClientStructDeclaration.Name }}{
		resp:      resp,
		reader:    bufio.NewReader(resp.Body),
		decoder:   decoder,
		requestID: requestID,
		firstLine: true,
	}
}

// parseSSEEvent reads one complete event from the response. Ending ctx closes
// the response body so a blocked read returns.
func (s *{{ .SSE.ClientStructDeclaration.Name }}) parseSSEEvent(ctx context.Context) (id string, hasID bool, eventType string, hasEventType bool, retry string, hasRetry bool, data []byte, err error) {
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
			id = ""
			hasID = false
			eventType = ""
			hasEventType = false
			retry = ""
			hasRetry = false
			data = nil
			err = contextErr
		}
	}()
	var dataLines []string

	for {
		line, err := s.readSSELine()
		if err != nil {
			return "", false, "", false, "", false, nil, err
		}

		if line == "" {
			// A blank line ends the current event.
			if len(dataLines) > 0 {
				break
			}
			// An event without data is not delivered. Its event and retry fields
			// must not appear on the next delivered event.
			eventType = ""
			hasEventType = false
			retry = ""
			hasRetry = false
			continue
		}
		
		if strings.HasPrefix(line, ":") {
			continue
		}
		field, value, found := strings.Cut(line, ":")
		if !found {
			value = ""
		} else {
			value = strings.TrimPrefix(value, " ")
		}
		switch field {
		case "id":
			if !strings.ContainsRune(value, '\x00') {
				s.lastEventID = value
				s.hasLastEventID = true
			}
		case "event":
			eventType = value
			hasEventType = true
		case "retry":
			// An invalid retry field does not replace an earlier valid field.
			validRetry := len(value) > 0
			for index := 0; validRetry && index < len(value); index++ {
				validRetry = value[index] >= '0' && value[index] <= '9'
			}
			if !validRetry {
				continue
			}
			retry = value
			hasRetry = true
		case "data":
			dataLines = append(dataLines, value)
		}
	}
	
	if len(dataLines) > 0 {
		data = []byte(strings.Join(dataLines, "\n"))
	}
	
	return s.lastEventID, s.hasLastEventID, eventType, hasEventType, retry, hasRetry, data, nil
}

// readSSELine reads one line ending in CR, LF, or CRLF. It returns io.EOF for
// a final line that is not terminated, so the caller cannot dispatch an
// incomplete event.
func (s *{{ .SSE.ClientStructDeclaration.Name }}) readSSELine() (string, error) {
	var line strings.Builder
	for {
		character, err := s.reader.ReadByte()
		if err != nil {
			return "", err
		}
		switch character {
		case '\n':
			return s.finishSSELine(line.String()), nil
		case '\r':
			if next, err := s.reader.Peek(1); err == nil && next[0] == '\n' {
				if _, err := s.reader.Discard(1); err != nil {
					return "", err
				}
			}
			return s.finishSSELine(line.String()), nil
		default:
			line.WriteByte(character)
		}
	}
}

// finishSSELine removes the optional byte-order marker only at the start of
// the stream.
func (s *{{ .SSE.ClientStructDeclaration.Name }}) finishSSELine(line string) string {
	if !s.firstLine {
		return line
	}
	s.firstLine = false
	return strings.TrimPrefix(line, "\uFEFF")
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
		{{ if .SSE.IDField }}eventID, hasEventID{{ else }}_, _{{ end }}, {{ if .SSE.EventField }}eventType, hasEventType{{ else }}_, _{{ end }}, {{ if .SSE.RetryField }}eventRetry, hasEventRetry{{ else }}_, _{{ end }}, data, err := s.parseSSEEvent(ctx)
		if err != nil {
			return zero, s.endStream(err)
		}

		var message struct {
			Method json.RawMessage `json:"method"`
			Params json.RawMessage `json:"params"`
			Result json.RawMessage `json:"result"`
			Error  json.RawMessage `json:"error"`
			ID     json.RawMessage `json:"id"`
		}
		if err := json.Unmarshal(data, &message); err != nil {
			return zero, s.endStream(fmt.Errorf("failed to parse JSON-RPC message: %w", err))
		}
		if len(message.Method) > 0 {
			// Read the streamed service result from the notification parameters.
			if len(message.Result) > 0 || len(message.Error) > 0 || len(message.ID) > 0 {
				return zero, s.endStream(fmt.Errorf("invalid JSON-RPC notification"))
			}
			var notification jsonrpc.RawRequest
			if err := json.Unmarshal(data, &notification); err != nil {
				return zero, s.endStream(fmt.Errorf("failed to parse notification: %w", err))
			}
			if notification.Invalid {
				return zero, s.endStream(fmt.Errorf("invalid JSON-RPC notification"))
			}
			if notification.JSONRPC != "2.0" {
				return zero, s.endStream(fmt.Errorf("invalid JSON-RPC version: %s", notification.JSONRPC))
			}
			if notification.HasID {
				return zero, s.endStream(fmt.Errorf("notification contains an id"))
			}
			if !notification.HasMethod {
				return zero, s.endStream(fmt.Errorf("notification has no method"))
			}
			if notification.Method != {{ printf "%q" .Method.Name }} {
				return zero, s.endStream(fmt.Errorf("received notification for JSON-RPC method %q", notification.Method))
			}
			{{- if not .SSE.Params.AllowAbsent }}
			if len(notification.Params) == 0 {
				return zero, s.endStream(fmt.Errorf("notification has no params"))
			}
			{{- end }}
			{{- if and .SSE.Params .SSE.Params.Positional }}
			{{- if .SSE.Params.AllowAbsent }}
			var params json.RawMessage
			if len(notification.Params) > 0 {
				params, err = jsonrpc.SinglePositionalParam(notification.Params)
				if err != nil {
					return zero, s.endStream(err)
				}
			}
			{{- else }}
			params := notification.Params
			params, err = jsonrpc.SinglePositionalParam(params)
			if err != nil {
				return zero, s.endStream(err)
			}
			{{- end }}
			result, err := s.decodeResult(params{{ if .SSE.IDField }}, eventID, hasEventID{{ end }}{{ if .SSE.EventField }}, eventType, hasEventType{{ end }}{{ if .SSE.RetryField }}, eventRetry, hasEventRetry{{ end }})
			{{- else }}
			result, err := s.decodeResult(notification.Params{{ if .SSE.IDField }}, eventID, hasEventID{{ end }}{{ if .SSE.EventField }}, eventType, hasEventType{{ end }}{{ if .SSE.RetryField }}, eventRetry, hasEventRetry{{ end }})
			{{- end }}
			if err != nil {
				return zero, s.endStream(err)
			}
			return result, nil
		}

		if len(message.Params) > 0 {
			return zero, s.endStream(fmt.Errorf("JSON-RPC response contains params"))
		}
		{
			// A response completes the stream. Stream values arrive in the
			// notifications handled above.
			var response jsonrpc.RawResponse
			if err := json.Unmarshal(data, &response); err != nil {
				return zero, s.endStream(fmt.Errorf("failed to parse response: %w", err))
			}
			if err := response.Validate(s.requestID); err != nil {
				return zero, s.endStream(err)
			}
			if response.Error == nil && !bytes.Equal(bytes.TrimSpace(response.Result), []byte("null")) {
				return zero, s.endStream(fmt.Errorf("JSON-RPC stream completion result must be null"))
			}
			if response.Error != nil {
				serviceErr, decodeErr := s.decodeError(response.Error{{ if $checksNullID }}, response.ID == nil{{ end }})
				if decodeErr != nil {
					return zero, s.endStream(decodeErr)
				}
				return zero, s.endStream(serviceErr)
			}
			return zero, s.endStream(io.EOF)
		}
	}
}

// decodeError rebuilds one designed service error from its JSON-RPC code and data.
func (s *{{ .SSE.ClientStructDeclaration.Name }}) decodeError(response *jsonrpc.RawErrorResponse{{ if $checksNullID }}, nullID bool{{ end }}) (error, error) {
	{{- if .Errors }}
	serviceErrorName, serviceErrorBody, ok := jsonrpc.DecodeServiceErrorData(response.Data)
	if !ok {
		return response, nil
	}
	switch response.Code {
	{{- range .Errors }}
	case {{ .StatusCode }}:
		switch serviceErrorName {
		{{- range $mapped := .Errors }}
			{{- with $mapped.Response }}
		case {{ printf "%q" $mapped.Name }}:
			{{- if or (eq .Code -32700) (eq .Code -32600) }}
			if nullID {
				return nil, goahttp.ErrInvalidResponse("{{ $.ServiceName }}", "{{ $.Method.Name }}", http.StatusOK, "response id is null")
			}
			{{- end }}
			resp := &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(serviceErrorBody)),
			}
			{{- if not .ClientBody }}
			if !bytes.Equal(bytes.TrimSpace(serviceErrorBody), []byte("null")) {
				return response, nil
			}
			{{- end }}
			decoder := s.decoder
			{{- template "partial_single_response" (buildResponseData . $.ServiceName $.Method) }}
			{{- if .ResultInit }}
			return {{ .ResultInit.Declaration.Name }}({{ range .ResultInit.ClientArgs }}{{ .Ref }},{{ end }}), nil
			{{- else if .ClientBody }}
			return body, nil
			{{- else }}
			return nil, nil
			{{- end }}
			{{- end }}
		{{- end }}
		default:
			return response, nil
		}
	{{- end }}
	default:
		return response, nil
	}
	{{- else }}
	return response, nil
	{{- end }}
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
func (s *{{ .SSE.ClientStructDeclaration.Name }}) decodeResult(data json.RawMessage{{ if .SSE.IDField }}, eventID string, hasEventID bool{{ end }}{{ if .SSE.EventField }}, eventType string, hasEventType bool{{ end }}{{ if .SSE.RetryField }}, eventRetry string, hasEventRetry bool{{ end }}) ({{ .SSE.EventTypeRef }}, error) {
	{{- if .Method.ViewedResult }}
	// The HTTP 200 status tells the configured decoder that this stream item is
	// a successful JSON-RPC result. Streaming results cannot carry HTTP headers or cookies.
	resp := &http.Response{StatusCode: http.StatusOK, Header: make(http.Header)}
	return {{ viewedDecodeName .Method.Name }}(s.decoder, resp, data{{ if .SSE.IDField }}, eventID, hasEventID{{ end }}{{ if .SSE.EventField }}, eventType, hasEventType{{ end }}{{ if .SSE.RetryField }}, eventRetry, hasEventRetry{{ end }})
	{{- else }}
	{{- with .SSE.Response.ClientBody }}
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(data)),
	}
	decoder := s.decoder(resp)
	var (
		body {{ if .Declaration }}{{ .Declaration.Name }}{{ else }}{{ .VarName }}{{ end }}
		err error
	)
	{{- if $.SSE.DataField }}
	{{- if $.SSE.Params.AllowAbsent }}
	if len(data) > 0 {
	{{- end }}
	if err = decoder.Decode(&body.{{ $.SSE.DataField }}); err != nil {
	{{- else }}
	if err = decoder.Decode(&body); err != nil {
	{{- end }}
		var zero {{ $.SSE.EventTypeRef }}
		return zero, goahttp.ErrDecodingError("{{ $.ServiceName }}", "{{ $.Method.Name }}", err)
	}
	{{- if $.SSE.Params.AllowAbsent }}
	}
	{{- end }}
	{{- if $.SSE.IDField }}
	{{- if or $.SSE.ID.HasDefault (not $.SSE.ID.Pointer) }}
	if !hasEventID {
		{{- if $.SSE.ID.HasDefault }}
		value := {{ $.SSE.ID.ClientTypeRef }}({{ printf "%q" $.SSE.ID.DefaultValue }})
		body.{{ $.SSE.IDField }} = {{ if $.SSE.ClientIDPointer }}&{{ end }}value
		{{- else }}
		var zero {{ $.SSE.EventTypeRef }}
		return zero, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", fmt.Errorf("server-sent event has no id for result field {{ $.SSE.IDField }}"))
		{{- end }}
	} else {
	{{- else }}
	if hasEventID {
	{{- end }}
		value := {{ $.SSE.ID.ClientTypeRef }}(eventID)
		body.{{ $.SSE.IDField }} = {{ if $.SSE.ClientIDPointer }}&{{ end }}value
	}
	{{- end }}
	{{- if $.SSE.EventField }}
	{{- if or $.SSE.Event.HasDefault (not $.SSE.Event.Pointer) }}
	if !hasEventType {
		{{- if $.SSE.Event.HasDefault }}
		value := {{ $.SSE.Event.ClientTypeRef }}({{ printf "%q" $.SSE.Event.DefaultValue }})
		body.{{ $.SSE.EventField }} = {{ if $.SSE.ClientEventPointer }}&{{ end }}value
		{{- else }}
		var zero {{ $.SSE.EventTypeRef }}
		return zero, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", fmt.Errorf("server-sent event has no type for result field {{ $.SSE.EventField }}"))
		{{- end }}
	} else {
	{{- else }}
	if hasEventType {
	{{- end }}
		value := {{ $.SSE.Event.ClientTypeRef }}(eventType)
		body.{{ $.SSE.EventField }} = {{ if $.SSE.ClientEventPointer }}&{{ end }}value
	}
	{{- end }}
	{{- if $.SSE.RetryField }}
	{{- if or $.SSE.Retry.HasDefault (not $.SSE.Retry.Pointer) }}
	if !hasEventRetry {
		{{- if $.SSE.Retry.HasDefault }}
		value := {{ $.SSE.Retry.ClientTypeRef }}({{ printf "%v" $.SSE.Retry.DefaultValue }})
		body.{{ $.SSE.RetryField }} = {{ if $.SSE.Retry.ClientPointer }}&{{ end }}value
		{{- else }}
		var zero {{ $.SSE.EventTypeRef }}
		return zero, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", fmt.Errorf("server-sent event has no retry for result field {{ $.SSE.RetryField }}"))
		{{- end }}
	} else {
	{{- else }}
	if hasEventRetry {
	{{- end }}
		{{- if sseRetrySigned $.SSE.Retry }}
		parsed, parseErr := strconv.ParseInt(eventRetry, 10, {{ sseRetryBits $.SSE.Retry }})
		if parseErr != nil || parsed < 0 {
			var zero {{ $.SSE.EventTypeRef }}
			return zero, goahttp.ErrDecodingError("{{ $.ServiceName }}", "{{ $.Method.Name }}", fmt.Errorf("invalid server-sent event retry %q", eventRetry))
		}
		{{- else }}
		parsed, parseErr := strconv.ParseUint(eventRetry, 10, {{ sseRetryBits $.SSE.Retry }})
		if parseErr != nil {
			var zero {{ $.SSE.EventTypeRef }}
			return zero, goahttp.ErrDecodingError("{{ $.ServiceName }}", "{{ $.Method.Name }}", fmt.Errorf("invalid server-sent event retry %q", eventRetry))
		}
		{{- end }}
		value := {{ $.SSE.Retry.ClientTypeRef }}(parsed)
		body.{{ $.SSE.RetryField }} = {{ if $.SSE.Retry.ClientPointer }}&{{ end }}value
	}
	{{- end }}
	{{- if and .ValidatorDeclaration .ValidationTarget }}
	err = {{ .ValidatorDeclaration.Name }}({{ .ValidationTarget }})
	if err != nil {
		var zero {{ $.SSE.EventTypeRef }}
		return zero, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", err)
	}
	{{- else if .ValidateRef }}
	{{ .ValidateRef }}
	if err != nil {
		var zero {{ $.SSE.EventTypeRef }}
		return zero, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", err)
	}
	{{- end }}
	{{- if $.SSE.ClientEventCode }}
	{{ $.SSE.ClientEventCode }}
	{{- else if $.SSE.Response.ResultInit }}
	result := {{ $.SSE.Response.ResultInit.Declaration.Name }}({{ range $.SSE.Response.ResultInit.ClientArgs }}{{ .Ref }},{{ end }})
	{{- else }}
	result := body
	{{- end }}
	return result, nil
	{{- else }}
	var zero {{ .SSE.EventTypeRef }}
	return zero, nil
	{{- end }}
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

{{- define "viewed_sse_outer_fields" }}{{ end }}
