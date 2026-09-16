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
                // started records whether the first stream line was read.
                started bool
		{{- if .SSE.IDField }}
		// lastEventID is the last valid id field read from the stream.
		lastEventID string
		// hasLastEventID distinguishes no id from an explicitly empty id.
		hasLastEventID bool
		{{- end }}
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
	for {
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
		var hasData bool
		event, hasData, err = s.processEvent(byts)
		if err != nil || hasData {
			return
		}
	}
}

// readEvent reads a single SSE event from the stream, respecting context
// cancellation. It accepts CR, LF, and CRLF line endings. A blank line ends an
// event; bytes after that line remain buffered for the next call.
func (s *{{ .SSE.ClientStructDeclaration.Name }}) readEvent(ctx context.Context) ([]byte, error) {
        const bufSize = 4096 // 4KB buffer size

        // Check for event in existing buffer
        event, ok := s.checkBuffer()
        if ok {
                return event, nil
        }

        eventData := event
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

		if n > 0 {
			eventData = append(eventData, buf[:n]...)
		}
		if end := s.eventEnd(eventData, errors.Is(err, io.EOF)); end >= 0 {
			s.lock.Lock()
			s.buffer = append(s.buffer[:0], eventData[end:]...)
			s.lock.Unlock()
			return eventData[:end], nil
		}

		// Discard any partial frame at EOF: an event that ends before its
		// blank-line delimiter was truncated by the transport.
		if errors.Is(err, io.EOF) {
			return nil, io.EOF
		}
        }
}

// checkBuffer returns the first complete event already read from the response.
// When no blank line is present, it returns a copy that readEvent can extend.
func (s *{{ .SSE.ClientStructDeclaration.Name }}) checkBuffer() ([]byte, bool) {
        s.lock.Lock()
        defer s.lock.Unlock()

        // Quick return if buffer is empty
        if len(s.buffer) == 0 {
                return nil, false
        }

	if end := s.eventEnd(s.buffer, false); end >= 0 {
		eventData := make([]byte, end)
		copy(eventData, s.buffer[:end])
		s.buffer = append(s.buffer[:0], s.buffer[end:]...)
		return eventData, true
	}

	// No complete event found, return a copy of the buffer contents: the
	// caller keeps accumulating into the returned slice while readEvent
	// refills s.buffer, so they must not share a backing array.
	eventData := make([]byte, len(s.buffer))
	copy(eventData, s.buffer)
	s.buffer = s.buffer[:0] // Clear buffer but keep capacity
	return eventData, false
}

// eventEnd returns the byte after the blank line that ends the first event.
// A CR at the end waits for the next byte unless the response has ended.
func (s *{{ .SSE.ClientStructDeclaration.Name }}) eventEnd(data []byte, atEOF bool) int {
	lineStart := 0
	for i := 0; i < len(data); {
		if data[i] != '\r' && data[i] != '\n' {
			i++
			continue
		}
		lineEnd := i + 1
		if data[i] == '\r' {
			if lineEnd == len(data) && !atEOF {
				return -1
			}
			if lineEnd < len(data) && data[lineEnd] == '\n' {
				lineEnd++
			}
		}
		if i == lineStart {
			return lineEnd
		}
		lineStart = lineEnd
		i = lineEnd
	}
	return -1
}

// eventLines splits one complete event using every line ending allowed by SSE.
func (s *{{ .SSE.ClientStructDeclaration.Name }}) eventLines(data []byte) [][]byte {
	var lines [][]byte
	for len(data) > 0 {
		lineEnd := 0
		for lineEnd < len(data) && data[lineEnd] != '\r' && data[lineEnd] != '\n' {
			lineEnd++
		}
		lines = append(lines, data[:lineEnd])
		if lineEnd == len(data) {
			break
		}
		next := lineEnd + 1
		if data[lineEnd] == '\r' && next < len(data) && data[next] == '\n' {
			next++
		}
		data = data[next:]
	}
	return lines
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
func (s *{{ .SSE.ClientStructDeclaration.Name }}) processEvent(eventData []byte) (event {{ .SSE.EventTypeRef }}, hasData bool, err error) {
	{{- if and .SSE.EventIsStruct (not .HasMixedResults) (not .Method.ViewedResult) (not .SSE.Response.ClientBody) }}
		event = new({{ .SSE.EventTypeName }})
	{{- end }}
	{{- if and (not .Method.ViewedResult) .SSE.Response.ClientBody }}
		{{- with .SSE.Response.ClientBody }}
	var body {{ if .Declaration }}{{ .Declaration.Name }}{{ else }}{{ .VarName }}{{ end }}
		{{- end }}
	{{- end }}
	{{- if .Method.ViewedResult }}
	var (
		{{- if .SSE.IDField }}
		idContent string
		hasID bool
		{{- end }}
		{{- if .SSE.EventField }}
		eventContent string
		hasEvent bool
		{{- end }}
		{{- if .SSE.RetryField }}
		retryContent string
		hasRetry bool
		{{- end }}
	)
	{{- end }}
        var dataLines []string
        for _, line := range s.eventLines(eventData) {
		if !s.started {
			line = bytes.TrimPrefix(line, []byte{0xef, 0xbb, 0xbf})
			s.started = true
		}
                if len(line) == 0 {
                        continue
                }
		field, value, _ := bytes.Cut(line, []byte(":"))
		if len(value) > 0 && value[0] == ' ' {
			value = value[1:]
		}
                if bytes.Equal(field, []byte("data")) {
                        dataLines = append(dataLines, string(value))
                        continue
                }
                {{- if .SSE.IDField }}
                if bytes.Equal(field, []byte("id")) {
			if !bytes.ContainsRune(value, '\x00') {
				s.lastEventID = string(value)
				s.hasLastEventID = true
			}
                        continue
                }
                {{- end }}
                {{- if .SSE.EventField }}
                if bytes.Equal(field, []byte("event")) {
			{{- if $.Method.ViewedResult }}
			eventContent = string(value)
			hasEvent = true
			{{- else if and $.SSE.Response.ClientBody $.SSE.ClientEventPointer }}
			eventContent := string(value)
			body.{{ .SSE.EventField }} = &eventContent
			{{- else }}
                        {{ if and (not $.Method.ViewedResult) $.SSE.Response.ClientBody }}body{{ else }}event{{ end }}.{{ .SSE.EventField }} = string(value)
			{{- end }}
                        continue
                }
                {{- end }}
                {{- if .SSE.RetryField }}
                if bytes.Equal(field, []byte("retry")) {
			// An invalid retry field does not replace an earlier valid field.
			validRetry := len(value) > 0
			for i := 0; validRetry && i < len(value); i++ {
				validRetry = value[i] >= '0' && value[i] <= '9'
			}
			if !validRetry {
				continue
			}
			{{- if $.Method.ViewedResult }}
			retryContent = string(value)
			hasRetry = true
			{{- else }}
			retryContent := string(value)
			{{- end }}
			{{- if $.Method.ViewedResult }}
			{{- else if $.HasMixedResults }}
			{{ template "partial_sse_parse" dict "Target" (printf "body.%s" .SSE.RetryField) "Source" "retryContent" "Encoding" .SSE.Retry "TypeRef" .SSE.Retry.ClientTypeRef "Named" false "Nullable" false "TargetPointer" .SSE.Retry.ClientPointer }}
			{{- else if and (not $.Method.ViewedResult) $.SSE.Response.ClientBody }}
			{{ template "partial_sse_parse" dict "Target" (printf "body.%s" .SSE.RetryField) "Source" "retryContent" "Encoding" .SSE.Retry "TypeRef" .SSE.Retry.ClientTypeRef "Named" false "Nullable" false "TargetPointer" .SSE.Retry.ClientPointer }}
			{{- else }}
			{{ template "partial_sse_parse" dict "Target" (printf "event.%s" .SSE.RetryField) "Source" "retryContent" "Encoding" .SSE.Retry "TypeRef" .SSE.Retry.TypeRef "Named" .SSE.Retry.Named "Nullable" false "TargetPointer" .SSE.Retry.Pointer }}
			{{- end }}
                        continue
		}
                {{- end }}
        }
	if len(dataLines) == 0 {
		return event, false, nil
	}
	{{- if .SSE.IDField }}
	if s.hasLastEventID {
		{{- if $.Method.ViewedResult }}
		idContent = s.lastEventID
		hasID = true
		{{- else if and $.SSE.Response.ClientBody $.SSE.ClientIDPointer }}
		idContent := s.lastEventID
		body.{{ .SSE.IDField }} = &idContent
		{{- else }}
		{{ if and (not $.Method.ViewedResult) $.SSE.Response.ClientBody }}body{{ else }}event{{ end }}.{{ .SSE.IDField }} = s.lastEventID
		{{- end }}
	}
	{{- end }}
	{{- if .Method.ViewedResult }}
		{{- if and .SSE.ID .SSE.ID.HasDefault }}
	if !hasID {
		idContent = {{ printf "%q" .SSE.ID.DefaultValue }}
		hasID = true
	}
		{{- end }}
		{{- if and .SSE.Event .SSE.Event.HasDefault }}
	if !hasEvent {
		eventContent = {{ printf "%q" .SSE.Event.DefaultValue }}
		hasEvent = true
	}
		{{- end }}
		{{- if and .SSE.Retry .SSE.Retry.HasDefault }}
	if !hasRetry {
		retryContent = {{ printf "%q" (printf "%v" .SSE.Retry.DefaultValue) }}
		hasRetry = true
	}
		{{- end }}
	{{- else if .SSE.Response.ClientBody }}
		{{- if and .SSE.ID .SSE.ID.HasDefault }}
			{{- if .SSE.ClientIDPointer }}
	if body.{{ .SSE.IDField }} == nil {
		value := {{ printf "%q" .SSE.ID.DefaultValue }}
		body.{{ .SSE.IDField }} = &value
	}
			{{- else }}
	body.{{ .SSE.IDField }} = {{ printf "%q" .SSE.ID.DefaultValue }}
			{{- end }}
		{{- end }}
		{{- if and .SSE.Event .SSE.Event.HasDefault }}
			{{- if .SSE.ClientEventPointer }}
	if body.{{ .SSE.EventField }} == nil {
		value := {{ printf "%q" .SSE.Event.DefaultValue }}
		body.{{ .SSE.EventField }} = &value
	}
			{{- else }}
	body.{{ .SSE.EventField }} = {{ printf "%q" .SSE.Event.DefaultValue }}
			{{- end }}
		{{- end }}
		{{- if and .SSE.Retry .SSE.Retry.HasDefault }}
			{{- if .SSE.Retry.ClientPointer }}
	if body.{{ .SSE.RetryField }} == nil {
		value := {{ printf "%#v" .SSE.Retry.DefaultValue }}
		body.{{ .SSE.RetryField }} = &value
	}
			{{- else }}
	body.{{ .SSE.RetryField }} = {{ printf "%#v" .SSE.Retry.DefaultValue }}
			{{- end }}
		{{- end }}
	{{- end }}
	{{- with .SSE.Response }}
		{{- if or .Headers .Cookies }}
	{{- template "sse_response_elements" $ }}
		{{- end }}
	{{- end }}
	{{- if .Method.ViewedResult }}
		{{- if .SSE.VariableView }}
	view := s.view
	switch view {
		{{- range .SSE.Response.ViewedRepresentations }}
	case {{ printf "%q" .View }}:
		{{- template "viewed_sse_client_result" dict "Endpoint" $ "Representation" . }}
		{{- end }}
	default:
		return event, false, goahttp.ErrValidationError("{{ .ServiceName }}", "{{ .Method.Name }}", goa.InvalidEnumValueError("view", view, []any{ {{ range .Method.ViewedResult.Views }}{{ printf "%q" .Name }}, {{ end }} }))
	}
		{{- else }}
	view := {{ printf "%q" .Method.ViewedResult.ViewName }}
			{{- range .SSE.Response.ViewedRepresentations }}
	{{- template "viewed_sse_client_result" dict "Endpoint" $ "Representation" . }}
			{{- end }}
		{{- end }}
	{{- else if .HasMixedResults }}
	{{- with .SSE.Response.ClientBody }}
        dataContent := strings.Join(dataLines, "\n")
                {{- if $.SSE.DataField }}
		{{ template "partial_sse_parse" dict "Target" (printf "body.%s" $.SSE.DataField) "Source" "dataContent" "Encoding" $.SSE.Data "TypeRef" $.SSE.Data.ClientTypeRef "Named" false "Nullable" $.SSE.Data.Pointer "TargetPointer" $.SSE.Data.ClientPointer }}
                {{- else if ssePrimitive $.SSE.Data }}
		{{ template "partial_sse_parse" dict "Target" "body" "Source" "dataContent" "Encoding" $.SSE.Data "TypeRef" $.SSE.Data.ClientTypeRef "Named" false "Nullable" $.SSE.Data.Pointer "TargetPointer" $.SSE.Data.ClientPointer }}
                {{- else }}
                respBody := &http.Response{
                        StatusCode: http.StatusOK,
                        Body:       io.NopCloser(bytes.NewReader([]byte(dataContent))),
                }
		if err = s.decoder(respBody).Decode(&body); err != nil {
			return event, false, goahttp.ErrDecodingError("{{ $.ServiceName }}", "{{ $.Method.Name }}", err)
		}
		{{- end }}
		{{- if and .ValidatorDeclaration .ValidationTarget }}
	err = {{ .ValidatorDeclaration.Name }}({{ .ValidationTarget }})
	if err != nil {
		return event, false, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", err)
	}
		{{- else if .ValidateRef }}
	{{ .ValidateRef }}
	if err != nil {
		return event, false, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", err)
	}
		{{- end }}
	{{ $.SSE.ClientEventCode }}
	return result, true, nil
	{{- else }}
	return event, true, nil
	{{- end }}
	{{- else }}
        dataContent := strings.Join(dataLines, "\n")
                {{- if .SSE.DataField }}
			{{- if .SSE.Response.ClientBody }}
		{{ template "partial_sse_parse" dict "Target" (printf "body.%s" .SSE.DataField) "Source" "dataContent" "Encoding" .SSE.Data "TypeRef" .SSE.Data.ClientTypeRef "Named" false "Nullable" .SSE.Data.Pointer "TargetPointer" .SSE.Data.ClientPointer }}
			{{- else }}
		{{ template "partial_sse_parse" dict "Target" (printf "event.%s" .SSE.DataField) "Source" "dataContent" "Encoding" .SSE.Data "TypeRef" .SSE.Data.TypeRef "Named" .SSE.Data.Named "Nullable" .SSE.Data.Pointer "TargetPointer" .SSE.Data.Pointer }}
			{{- end }}
                {{- else if .SSE.EventIsStruct }}
                // Decode the event data into the HTTP response type so required
                // primitive fields retain their presence for validation.
                respBody := &http.Response{
                        StatusCode: http.StatusOK,
                        Body:       io.NopCloser(bytes.NewReader([]byte(dataContent))),
                }
                err = s.decoder(respBody).Decode({{ if .SSE.Response.ClientBody }}&body{{ else }}event{{ end }})
                if err != nil {
                        return
                }
                {{- else }}
			{{- if .SSE.Response.ClientBody }}
		{{ template "partial_sse_parse" dict "Target" "body" "Source" "dataContent" "Encoding" .SSE.Data "TypeRef" .SSE.Data.ClientTypeRef "Named" false "Nullable" .SSE.Data.Pointer "TargetPointer" .SSE.Data.ClientPointer }}
			{{- else }}
		{{ template "partial_sse_parse" dict "Target" "event" "Source" "dataContent" "Encoding" .SSE.Data "TypeRef" .SSE.Data.TypeRef "Named" .SSE.Data.Named "Nullable" .SSE.Data.Pointer "TargetPointer" .SSE.Data.Pointer }}
			{{- end }}
		{{- end }}
		{{- with .SSE.Response.ClientBody }}
			{{- if and .ValidatorDeclaration .ValidationTarget }}
	err = {{ .ValidatorDeclaration.Name }}({{ .ValidationTarget }})
	if err != nil {
		return event, false, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", err)
	}
			{{- else if .ValidateRef }}
	{{ .ValidateRef }}
	if err != nil {
		return event, false, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", err)
	}
			{{- end }}
		{{- end }}
		{{- with .SSE.Response.ResultInit }}
	event = {{ .Declaration.Name }}({{ range .ClientArgs }}{{ .Ref }},{{ end }})
		{{- else }}
			{{- if .SSE.Response.ClientBody }}
	event = body
			{{- end }}
		{{- end }}
	{{- end }}
	{{- if not .HasMixedResults }}
	hasData = true
	return
	{{- end }}
}

{{- define "viewed_sse_client_result" }}
	{{- $endpoint := .Endpoint }}
		{{- with .Representation }}
			{{- if .ClientBody }}
			var body {{ if .ClientBody.Declaration }}{{ .ClientBody.Declaration.Name }}{{ else }}{{ .ClientBody.VarName }}{{ end }}
			{{- if $endpoint.SSE.IDField }}
			if hasID {
				{{- if $endpoint.SSE.ClientIDPointer }}
				body.{{ $endpoint.SSE.IDField }} = &idContent
				{{- else }}
				body.{{ $endpoint.SSE.IDField }} = idContent
				{{- end }}
			}
			{{- end }}
			{{- if $endpoint.SSE.EventField }}
			if hasEvent {
				{{- if $endpoint.SSE.ClientEventPointer }}
				body.{{ $endpoint.SSE.EventField }} = &eventContent
				{{- else }}
				body.{{ $endpoint.SSE.EventField }} = eventContent
				{{- end }}
			}
			{{- end }}
			{{- if $endpoint.SSE.RetryField }}
			if hasRetry {
			{{ template "partial_sse_parse" dict "Target" (printf "body.%s" $endpoint.SSE.RetryField) "Source" "retryContent" "Encoding" $endpoint.SSE.Retry "TypeRef" $endpoint.SSE.Retry.ClientTypeRef "Named" false "Nullable" false "TargetPointer" $endpoint.SSE.Retry.ClientPointer }}
			}
			{{- end }}
		dataContent := strings.Join(dataLines, "\n")
			{{- if ssePrimitive $endpoint.SSE.Data }}
				{{- if $endpoint.SSE.DataField }}
			{{ template "partial_sse_parse" dict "Target" (printf "body.%s" $endpoint.SSE.DataField) "Source" "dataContent" "Encoding" $endpoint.SSE.Data "TypeRef" $endpoint.SSE.Data.ClientTypeRef "Named" false "Nullable" $endpoint.SSE.Data.Pointer "TargetPointer" .ClientDataPointer }}
				{{- else }}
			{{ template "partial_sse_parse" dict "Target" "body" "Source" "dataContent" "Encoding" $endpoint.SSE.Data "TypeRef" $endpoint.SSE.Data.ClientTypeRef "Named" false "Nullable" $endpoint.SSE.Data.Pointer "TargetPointer" .ClientDataPointer }}
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
				return event, false, goahttp.ErrDecodingError("{{ $endpoint.ServiceName }}", "{{ $endpoint.Method.Name }}", err)
			}
			{{- end }}
			{{- end }}
			{{- if and .ClientBody.ValidatorDeclaration .ClientBody.ValidationTarget }}
			err = {{ .ClientBody.ValidatorDeclaration.Name }}({{ .ClientBody.ValidationTarget }})
			if err != nil {
				return event, false, goahttp.ErrValidationError("{{ $endpoint.ServiceName }}", "{{ $endpoint.Method.Name }}", err)
			}
			{{- else if .ClientBody.ValidateRef }}
			{{ .ClientBody.ValidateRef }}
			if err != nil {
				return event, false, goahttp.ErrValidationError("{{ $endpoint.ServiceName }}", "{{ $endpoint.Method.Name }}", err)
			}
			{{- end }}
			projected := {{ .ResultInit.Declaration.Name }}({{ range .ResultInit.ClientArgs }}{{ .Ref }}, {{ end }})
		viewed := {{ if not $endpoint.Method.ViewedResult.IsCollection }}&{{ end }}{{ $endpoint.Method.ViewedResult.ViewsPkg }}.{{ $endpoint.Method.ViewedResult.VarName }}{Projected: projected, View: view}
			if err = {{ $endpoint.Method.ViewedResult.ViewsPkg }}.{{ $endpoint.Method.ViewedResult.Validate.Declaration.Name }}(viewed); err != nil {
			return event, false, goahttp.ErrValidationError("{{ $endpoint.ServiceName }}", "{{ $endpoint.Method.Name }}", err)
			}
			result := {{ $endpoint.ServicePkgName }}.{{ $endpoint.Method.ViewedResult.ResultInit.Declaration.Name }}(viewed)
			return result, true, nil
	{{- end }}
{{- end }}

{{- define "sse_response_elements" }}
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
		return event, false, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", goa.MissingFieldError("{{ .Name }}", "header"))
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
		return event, false, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", goa.MissingFieldError("{{ .Name }}", "header"))
	}
					{{- end }}
				{{- else if .Slice }}
	{{ .VarName }}Raw := s.resp.Header["{{ .CanonicalName }}"]
					{{- if .Required }}
	if {{ .VarName }}Raw == nil {
		return event, false, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", goa.MissingFieldError("{{ .Name }}", "header"))
	}
					{{- end }}
	if {{ .VarName }}Raw != nil {
		{{- template "partial_element_slice_conversion" . }}
	}
				{{- else }}
	{{ .VarName }}Raw := s.resp.Header.Get("{{ .CanonicalName }}")
					{{- if .Required }}
	if {{ .VarName }}Raw == "" {
		return event, false, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", goa.MissingFieldError("{{ .Name }}", "header"))
	}
					{{- end }}
	if {{ .VarName }}Raw != "" {
		{{- template "partial_query_type_conversion" . }}
	}
				{{- end }}
				{{- if .Validate }}
	{{ .Validate }}
	if err != nil {
		return event, false, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", err)
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
		return event, false, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", goa.MissingFieldError("{{ .Name }}", "cookie"))
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
		return event, false, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", err)
	}
				{{- end }}
			{{- end }}
		{{- end }}
	{{- end }}
{{- end }}
