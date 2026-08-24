{{/*
client_sse.go.tpl writes the HTTP client stream for one SSE endpoint. The plan
provides the exact data and retry types used to rebuild each service result.
*/ -}}
// {{ .SSE.ClientInterfaceDeclaration.Name }} is the interface for reading Server-Sent Events.
type {{ .SSE.ClientInterfaceDeclaration.Name }} interface {
    // {{ .Method.ClientStream.RecvName }} reads and returns the next event from the SSE stream.
    {{ .Method.ClientStream.RecvName }}() ({{ .SSE.EventTypeRef }}, error)
    // {{ .Method.ClientStream.RecvWithContextName }} reads and returns the next event from the SSE stream with context.
    {{ .Method.ClientStream.RecvWithContextName }}(context.Context) ({{ .SSE.EventTypeRef }}, error)
    // Close closes the SSE stream and releases resources.
    Close() error
}

type (
        // {{ .SSE.ClientStructDeclaration.Name }} implements the {{ .SSE.ClientInterfaceDeclaration.Name }} interface.
        {{ .SSE.ClientStructDeclaration.Name }} struct {
                resp *http.Response
                decoder func(*http.Response) goahttp.Decoder
                buffer []byte // Buffer for unprocessed data
                lock sync.Mutex
                closed bool
		{{- if .SSE.VariableView }}
		view string
		{{- end }}
        }
)

// {{ .SSE.ClientStructDeclaration.Name }} implements the {{ .SSE.ClientInterfaceDeclaration.Name }} interface.
var _ {{ .SSE.ClientInterfaceDeclaration.Name }} = (*{{ .SSE.ClientStructDeclaration.Name }})(nil)

// {{ .SSE.ClientStructDeclaration.Name }} implements the service client stream
// interface so the generated endpoint client can return it directly.
var _ {{ .ServicePkgName }}.{{ .Method.ClientStream.Interface }} = (*{{ .SSE.ClientStructDeclaration.Name }})(nil)

// {{ .SSE.ClientInitDeclaration.Name }} creates a new {{ .SSE.ClientInterfaceDeclaration.Name }}.
func {{ .SSE.ClientInitDeclaration.Name }}(resp *http.Response, decoder func(*http.Response) goahttp.Decoder) {{ .SSE.ClientInterfaceDeclaration.Name }} {
        return &{{ .SSE.ClientStructDeclaration.Name }}{
                resp: resp,
                decoder: decoder,
                buffer: make([]byte, 0, 4096), // Pre-allocate buffer
		{{- if .SSE.VariableView }}
		view: resp.Header.Get("goa-view"),
		{{- end }}
        }
}

// {{ .Method.ClientStream.RecvName }} reads and returns the next event from the SSE stream.
func (s *{{ .SSE.ClientStructDeclaration.Name }}) {{ .Method.ClientStream.RecvName }}() ({{ .SSE.EventTypeRef }}, error) {
        return s.{{ .Method.ClientStream.RecvWithContextName }}(context.Background())
}

// {{ .Method.ClientStream.RecvWithContextName }} reads and returns the next event from the SSE stream, respecting context cancellation.
func (s *{{ .SSE.ClientStructDeclaration.Name }}) {{ .Method.ClientStream.RecvWithContextName }}(ctx context.Context) (event {{ .SSE.EventTypeRef }}, err error) {
        var byts []byte
        byts, err = s.readEvent(ctx)
        if err != nil {
                if errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
                        // Clean up on EOF or context cancellation. io.EOF
                        // propagates to the caller to signal end of stream.
                        s.Close()
                }
                return
        }
        return s.processEvent(byts)
}

// readEvent reads a single SSE event from the stream, respecting context
// cancellation.  It first checks the internal buffer for a complete event
// (delimited by double newlines). If no complete event is found, it reads from
// the HTTP response body until it either finds an event boundary, reaches EOF,
// or encounters an error. Any data after the event boundary is saved in the
// buffer for the next call.
func (s *{{ .SSE.ClientStructDeclaration.Name }}) readEvent(ctx context.Context) ([]byte, error) {
        const bufSize = 4096 // 4KB buffer size

        // Check for event in existing buffer
        event, ok := s.checkBuffer()
        if ok {
                return event, nil
        }

        // Initialize with any data from buffer
        eventData := event
        wasNewline := len(eventData) > 0 && eventData[len(eventData)-1] == '\n'
        buf := make([]byte, bufSize)

	// Read data in chunks until we find an event or hit EOF. A stream that
	// ends mid-event (before the blank-line delimiter) discards the partial
	// frame, per the SSE specification.
	for {
		// Check if context is done
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// Continue processing
		}

		// Check if stream is closed
		s.lock.Lock()
		if s.closed {
			s.lock.Unlock()
			return nil, io.EOF
		}

                // Read next chunk
                n, err := s.resp.Body.Read(buf)
                s.lock.Unlock()

                // Handle read errors
                if err != nil && err != io.EOF {
                        return nil, err
                }

                // Process data if we got any
                if n > 0 {
                        // Look for event boundary in this chunk
                        for i := 0; i < n; i++ {
                                b := buf[i]
                                eventData = append(eventData, b)

                                // Check for double newlines (event boundary)
                                if b == '\n' && wasNewline {
                                        // Save any remaining data for next read
                                        if i+1 < n {
                                                s.lock.Lock()
                                                s.buffer = append(s.buffer[:0], buf[i+1:n]...)
                                                s.lock.Unlock()
                                        }
                                        return eventData, nil
                                }

                                // Update newline tracking
                                wasNewline = (b == '\n')
                        }
                }

		// Discard any partial frame at EOF: an event that ends before its
		// blank-line delimiter was truncated by the transport.
		if errors.Is(err, io.EOF) {
			return nil, io.EOF
		}
        }
}

// checkBuffer examines the internal buffer for a complete SSE event (delimited
// by double newlines).  It returns two values: the event data (or all buffer
// contents if no complete event is found), and a boolean indicating whether a
// complete event was found. If a complete event is found, any remaining data
// after the event is kept in the buffer for the next call.
func (s *{{ .SSE.ClientStructDeclaration.Name }}) checkBuffer() ([]byte, bool) {
        s.lock.Lock()
        defer s.lock.Unlock()

        // Quick return if buffer is empty
        if len(s.buffer) == 0 {
                return nil, false
        }

	// Look for double newline in buffer
	for i := 0; i < len(s.buffer)-1; i++ {
		if s.buffer[i] == '\n' && s.buffer[i+1] == '\n' {
			// Found complete event. Copy it out: compacting the buffer
			// below would otherwise overwrite the returned bytes, since
			// both slices share the same backing array.
			eventEnd := i + 2 // Include both newlines
			eventData := make([]byte, eventEnd)
			copy(eventData, s.buffer[:eventEnd])

			// Save remaining data for next time
			if eventEnd < len(s.buffer) {
				s.buffer = append(s.buffer[:0], s.buffer[eventEnd:]...)
			} else {
				s.buffer = s.buffer[:0]
			}

			return eventData, true
		}
	}

	// No complete event found, return a copy of the buffer contents: the
	// caller keeps accumulating into the returned slice while readEvent
	// refills s.buffer, so they must not share a backing array.
	eventData := make([]byte, len(s.buffer))
	copy(eventData, s.buffer)
	s.buffer = s.buffer[:0] // Clear buffer but keep capacity
	return eventData, false
}

// Close closes the SSE stream and releases any associated resources.
func (s *{{ .SSE.ClientStructDeclaration.Name }}) Close() error {
        s.lock.Lock()
        defer s.lock.Unlock()
        if s.closed {
                return nil
        }
        s.closed = true
        return s.resp.Body.Close()
}

// processEvent processes a raw SSE event into the expected type
func (s *{{ .SSE.ClientStructDeclaration.Name }}) processEvent(eventData []byte) (event {{ .SSE.EventTypeRef }}, err error) {
	{{- if and .SSE.EventIsStruct (not .HasMixedResults) }}
		event = new({{ .SSE.EventTypeName }})
	{{- end }}
	{{- if .HasMixedResults }}
		{{- with .SSE.Response.ClientBody }}
	var body {{ if .Declaration }}{{ .Declaration.Name }}{{ else }}{{ .VarName }}{{ end }}
		{{- end }}
	{{- end }}
        var dataLines []string
        for _, line := range bytes.Split(eventData, []byte("\n")) {
                if len(line) == 0 {
                        continue
                }
                if bytes.HasPrefix(line, []byte("data:")) {
                        dataLines = append(dataLines, s.trimHeader(line[len("data:"):]))
                        continue
                }
                {{- if .SSE.IDField }}
                if bytes.HasPrefix(line, []byte("id:")) {
			{{- if and $.HasMixedResults $.SSE.ClientIDPointer }}
			idContent := s.trimHeader(line[len("id:"):])
			body.{{ .SSE.IDField }} = &idContent
			{{- else }}
                        {{ if $.HasMixedResults }}body{{ else }}event{{ end }}.{{ .SSE.IDField }} = s.trimHeader(line[len("id:"):])
			{{- end }}
                        continue
                }
                {{- end }}
                {{- if .SSE.EventField }}
                if bytes.HasPrefix(line, []byte("event:")) {
			{{- if and $.HasMixedResults $.SSE.ClientEventPointer }}
			eventContent := s.trimHeader(line[len("event:"):])
			body.{{ .SSE.EventField }} = &eventContent
			{{- else }}
                        {{ if $.HasMixedResults }}body{{ else }}event{{ end }}.{{ .SSE.EventField }} = s.trimHeader(line[len("event:"):])
			{{- end }}
                        continue
                }
                {{- end }}
                {{- if .SSE.RetryField }}
                if bytes.HasPrefix(line, []byte("retry:")) {
			retryContent := s.trimHeader(line[len("retry:"):])
			{{- if $.HasMixedResults }}
			{{ template "partial_sse_parse" dict "Target" (printf "body.%s" .SSE.RetryField) "Source" "retryContent" "Encoding" .SSE.Retry "Nullable" false "TargetPointer" .SSE.Retry.ClientPointer }}
			{{- else }}
			{{ template "partial_sse_parse" dict "Target" (printf "event.%s" .SSE.RetryField) "Source" "retryContent" "Encoding" .SSE.Retry "Nullable" false "TargetPointer" .SSE.Retry.Pointer }}
			{{- end }}
                        continue
                }
                {{- end }}
        }
	{{- if .Method.ViewedResult }}
	{{- template "viewed_sse_response_elements" . }}
		{{- if .SSE.VariableView }}
	view := s.view
	switch view {
		{{- range .SSE.Response.ViewedRepresentations }}
	case {{ printf "%q" .View }}:
		{{- template "viewed_sse_client_result" dict "Endpoint" $ "Representation" . }}
		{{- end }}
	default:
		return event, goahttp.ErrValidationError("{{ .ServiceName }}", "{{ .Method.Name }}", goa.InvalidEnumValueError("view", view, []any{ {{ range .Method.ViewedResult.Views }}{{ printf "%q" .Name }}, {{ end }} }))
	}
		{{- else }}
	view := {{ printf "%q" .Method.ViewedResult.ViewName }}
			{{- range .SSE.Response.ViewedRepresentations }}
	{{- template "viewed_sse_client_result" dict "Endpoint" $ "Representation" . }}
			{{- end }}
		{{- end }}
	{{- else if .HasMixedResults }}
	{{- with .SSE.Response.ClientBody }}
        if len(dataLines) > 0 {
                dataContent := strings.Join(dataLines, "\n")
                {{- if $.SSE.DataField }}
		{{ template "partial_sse_parse" dict "Target" (printf "body.%s" $.SSE.DataField) "Source" "dataContent" "Encoding" $.SSE.Data "Nullable" $.SSE.Data.Pointer "TargetPointer" $.SSE.Data.ClientPointer }}
                {{- else if ssePrimitive $.SSE.Data }}
		{{ template "partial_sse_parse" dict "Target" "body" "Source" "dataContent" "Encoding" $.SSE.Data "Nullable" $.SSE.Data.Pointer "TargetPointer" $.SSE.Data.Pointer }}
                {{- else }}
                respBody := &http.Response{
                        StatusCode: http.StatusOK,
                        Body:       io.NopCloser(bytes.NewReader([]byte(dataContent))),
                }
		if err = s.decoder(respBody).Decode(&body); err != nil {
			return event, goahttp.ErrDecodingError("{{ $.ServiceName }}", "{{ $.Method.Name }}", err)
		}
                {{- end }}
        }
		{{- if and .ValidatorDeclaration .ValidationTarget }}
	err = {{ .ValidatorDeclaration.Name }}({{ .ValidationTarget }})
	if err != nil {
		return event, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", err)
	}
		{{- else if .ValidateRef }}
	{{ .ValidateRef }}
	if err != nil {
		return event, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", err)
	}
		{{- end }}
	{{ $.SSE.ClientEventCode }}
	return result, nil
	{{- else }}
	return event, nil
	{{- end }}
	{{- else }}
        if len(dataLines) > 0 {
                dataContent := strings.Join(dataLines, "\n")
                {{- if .SSE.DataField }}
		{{ template "partial_sse_parse" dict "Target" (printf "event.%s" .SSE.DataField) "Source" "dataContent" "Encoding" .SSE.Data "Nullable" .SSE.Data.Pointer "TargetPointer" .SSE.Data.Pointer }}
                {{- else if .SSE.EventIsStruct }}
                // Decode the event data into the result value returned by Recv.
                respBody := &http.Response{
                        StatusCode: http.StatusOK,
                        Body:       io.NopCloser(bytes.NewReader([]byte(dataContent))),
                }
                err = s.decoder(respBody).Decode(event)
                if err != nil {
                        return
                }
                {{- else }}
		{{ template "partial_sse_parse" dict "Target" "event" "Source" "dataContent" "Encoding" .SSE.Data "Nullable" .SSE.Data.Pointer "TargetPointer" .SSE.Data.Pointer }}
                {{- end }}
        }
	{{- end }}
	{{- if not .HasMixedResults }}
	return
	{{- end }}
}

{{- define "viewed_sse_client_result" }}
	{{- $endpoint := .Endpoint }}
	{{- with .Representation }}
		{{- if .ClientBody }}
		var body {{ if .ClientBody.Declaration }}{{ .ClientBody.Declaration.Name }}{{ else }}{{ .ClientBody.VarName }}{{ end }}
		{{- if $endpoint.SSE.IDField }}
		body.{{ $endpoint.SSE.IDField }} = event.{{ $endpoint.SSE.IDField }}
		{{- end }}
		{{- if $endpoint.SSE.EventField }}
		body.{{ $endpoint.SSE.EventField }} = event.{{ $endpoint.SSE.EventField }}
		{{- end }}
		if len(dataLines) > 0 {
			dataContent := strings.Join(dataLines, "\n")
			{{- if ssePrimitive $endpoint.SSE.Data }}
				{{- if $endpoint.SSE.DataField }}
			{{ template "partial_sse_parse" dict "Target" (printf "body.%s" $endpoint.SSE.DataField) "Source" "dataContent" "Encoding" $endpoint.SSE.Data "Nullable" $endpoint.SSE.Data.Pointer "TargetPointer" .ClientDataPointer }}
				{{- else }}
			{{ template "partial_sse_parse" dict "Target" "body" "Source" "dataContent" "Encoding" $endpoint.SSE.Data "Nullable" $endpoint.SSE.Data.Pointer "TargetPointer" .ClientDataPointer }}
				{{- end }}
			{{- else }}
			respBody := &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(dataContent))),
			}
				{{- if $endpoint.SSE.DataField }}
			if err = s.decoder(respBody).Decode(&body.{{ $endpoint.SSE.DataField }}); err != nil {
				{{- else }}
			if err = s.decoder(respBody).Decode(&body); err != nil {
				{{- end }}
				return event, goahttp.ErrDecodingError("{{ $endpoint.ServiceName }}", "{{ $endpoint.Method.Name }}", err)
			}
			{{- end }}
		}
		{{- end }}
		projected := {{ .ResultInit.Declaration.Name }}({{ range .ResultInit.ClientArgs }}{{ .Ref }}, {{ end }})
		viewed := {{ if not $endpoint.Method.ViewedResult.IsCollection }}&{{ end }}{{ $endpoint.Method.ViewedResult.ViewsPkg }}.{{ $endpoint.Method.ViewedResult.VarName }}{Projected: projected, View: view}
			if err = {{ $endpoint.Method.ViewedResult.ViewsPkg }}.{{ $endpoint.Method.ViewedResult.Validate.Declaration.Name }}(viewed); err != nil {
			return event, goahttp.ErrValidationError("{{ $endpoint.ServiceName }}", "{{ $endpoint.Method.Name }}", err)
		}
		result := {{ $endpoint.ServicePkgName }}.{{ $endpoint.Method.ViewedResult.ResultInit.Declaration.Name }}(viewed)
		{{- if $endpoint.SSE.IDField }}
		result.{{ $endpoint.SSE.IDField }} = event.{{ $endpoint.SSE.IDField }}
		{{- end }}
		{{- if $endpoint.SSE.EventField }}
		result.{{ $endpoint.SSE.EventField }} = event.{{ $endpoint.SSE.EventField }}
		{{- end }}
		return result, nil
	{{- end }}
{{- end }}

{{- define "viewed_sse_response_elements" }}
	{{- with .SSE.Response }}
		{{- if .Headers }}
	var (
			{{- range .Headers }}
		{{ .VarName }} {{ .TypeRef }}
			{{- end }}
	)
			{{- range .Headers }}
				{{- if (or (eq .Type.Name "string") (eq .Type.Name "any")) }}
	{{ .VarName }}Raw := s.resp.Header.Get("{{ .CanonicalName }}")
					{{- if .Required }}
	if {{ .VarName }}Raw == "" {
		return event, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", goa.MissingFieldError("{{ .Name }}", "header"))
	}
	{{ .VarName }} = {{ if and (eq .Type.Name "string") .Pointer }}&{{ end }}{{ .VarName }}Raw
					{{- else }}
	if {{ .VarName }}Raw != "" {
		{{ .VarName }} = {{ if and (eq .Type.Name "string") .Pointer }}&{{ end }}{{ .VarName }}Raw
	}
					{{- end }}
				{{- else if .StringSlice }}
	{{ .VarName }} = s.resp.Header["{{ .CanonicalName }}"]
					{{- if .Required }}
	if {{ .VarName }} == nil {
		return event, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", goa.MissingFieldError("{{ .Name }}", "header"))
	}
					{{- end }}
				{{- else if .Slice }}
	{{ .VarName }}Raw := s.resp.Header["{{ .CanonicalName }}"]
					{{- if .Required }}
	if {{ .VarName }}Raw == nil {
		return event, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", goa.MissingFieldError("{{ .Name }}", "header"))
	}
					{{- end }}
	if {{ .VarName }}Raw != nil {
		{{- template "partial_element_slice_conversion" . }}
	}
				{{- else }}
	{{ .VarName }}Raw := s.resp.Header.Get("{{ .CanonicalName }}")
					{{- if .Required }}
	if {{ .VarName }}Raw == "" {
		return event, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", goa.MissingFieldError("{{ .Name }}", "header"))
	}
					{{- end }}
	if {{ .VarName }}Raw != "" {
		{{- template "partial_query_type_conversion" . }}
	}
				{{- end }}
				{{- if .Validate }}
	{{ .Validate }}
	if err != nil {
		return event, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", err)
	}
				{{- end }}
			{{- end }}
		{{- end }}
		{{- if .Cookies }}
	var (
			{{- range .Cookies }}
		{{ .VarName }} {{ .TypeRef }}
		{{ .VarName }}Raw string
			{{- end }}
	)
	for _, cookie := range s.resp.Cookies() {
		switch cookie.Name {
			{{- range .Cookies }}
		case {{ printf "%q" .HTTPName }}:
			{{ .VarName }}Raw = cookie.Value
			{{- end }}
		}
	}
			{{- range .Cookies }}
				{{- if .Required }}
	if {{ .VarName }}Raw == "" {
		return event, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", goa.MissingFieldError("{{ .Name }}", "cookie"))
	}
				{{- end }}
	if {{ .VarName }}Raw != "" {
				{{- if (or (eq .Type.Name "string") (eq .Type.Name "any")) }}
		{{ .VarName }} = {{ if and (eq .Type.Name "string") .Pointer }}&{{ end }}{{ .VarName }}Raw
				{{- else }}
		{{- template "partial_query_type_conversion" . }}
				{{- end }}
	}
				{{- if .Validate }}
	{{ .Validate }}
	if err != nil {
		return event, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", err)
	}
				{{- end }}
			{{- end }}
		{{- end }}
	{{- end }}
{{- end }}

// trimHeader removes the optional space after an SSE field name.
func (s *{{ .SSE.ClientStructDeclaration.Name }}) trimHeader(data []byte) string {
        if len(data) > 0 && data[0] == ' ' {
                data = data[1:]
        }
        return string(data)
}
