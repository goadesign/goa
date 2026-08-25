{{- if or .SSEStream .NoOutputWriter }}
type (
	{{- if .SSEStream }}
	{{ printf "%s writes JSON-RPC messages as server-sent events." .SSEStream.Name | comment }}
	{{ .SSEStream.Name }} struct {
		// once writes the HTTP headers only for the first event.
		once sync.Once
		// w receives the HTTP headers and event bytes.
		w http.ResponseWriter
		// encoder turns one JSON-RPC message into bytes.
		encoder func(context.Context, http.ResponseWriter) goahttp.Encoder
	}

	{{ printf "%s stores an encoded event before any HTTP output is written." .SSEBuffer.Name | comment }}
	{{ .SSEBuffer.Name }} struct {
		bytes.Buffer
		header http.Header
	}
	{{- end }}

	{{- if .NoOutputWriter }}
	{{ printf "%s accepts notification output without storing or sending it." .NoOutputWriter.Name | comment }}
	{{ .NoOutputWriter.Name }} struct {
		header http.Header
	}
	{{- end }}
)
{{- end }}

{{- if .SSEStream }}
// Header returns the headers written while the event is being encoded.
func (b *{{ .SSEBuffer.Name }}) Header() http.Header {
	return b.header
}

// WriteHeader leaves the response status for the real HTTP response writer.
func (b *{{ .SSEBuffer.Name }}) WriteHeader(int) {
}
{{- end }}

{{- if .NoOutputWriter }}
// Header returns the private headers written while a notification is encoded.
func (w *{{ .NoOutputWriter.Name }}) Header() http.Header {
	return w.header
}

// Write accepts notification bytes without storing or sending them.
func (w *{{ .NoOutputWriter.Name }}) Write(data []byte) (int, error) {
	return len(data), nil
}

// WriteHeader accepts a notification status without sending it.
func (w *{{ .NoOutputWriter.Name }}) WriteHeader(int) {
}

// Flush completes successfully because notification bytes are not sent.
func (w *{{ .NoOutputWriter.Name }}) Flush() {
}
{{- end }}

{{- if .SSEStream }}
// initSSEHeaders writes the response headers before the first event.
func (s *{{ .SSEStream.Name }}) initSSEHeaders() {
	s.once.Do(func() {
		header := s.w.Header()
		header.Set("Content-Type", "text/event-stream")
		header.Set("Cache-Control", "no-cache")
		header.Set("Connection", "keep-alive")
		header.Set("X-Accel-Buffering", "no")
		s.w.WriteHeader(http.StatusOK)
	})
}

// sendSSEEvent encodes one complete JSON-RPC message before starting the
// response, then writes its mapped SSE fields and data.
func (s *{{ .SSEStream.Name }}) sendSSEEvent(ctx context.Context, value any, id, eventType, retry *string) error {
	if id != nil {
		for _, character := range *id {
			if character == 0 || character == '\r' || character == '\n' {
				return fmt.Errorf("server-sent event id contains a forbidden character")
			}
		}
	}
	if eventType != nil {
		for _, character := range *eventType {
			if character == '\r' || character == '\n' {
				return fmt.Errorf("server-sent event name contains a line break")
			}
		}
	}
	if retry != nil {
		if *retry == "" {
			return fmt.Errorf("server-sent event retry must contain only decimal digits")
		}
		for index := 0; index < len(*retry); index++ {
			if (*retry)[index] < '0' || (*retry)[index] > '9' {
				return fmt.Errorf("server-sent event retry must contain only decimal digits")
			}
		}
	}

	event := &{{ .SSEBuffer.Name }}{header: make(http.Header)}
	if err := s.encoder(ctx, event).Encode(value); err != nil {
		return err
	}

	s.initSSEHeaders()
	if id != nil {
		if _, err := fmt.Fprintf(s.w, "id: %s\n", *id); err != nil {
			return fmt.Errorf("write server-sent event id: %w", err)
		}
	}
	if eventType != nil {
		if _, err := fmt.Fprintf(s.w, "event: %s\n", *eventType); err != nil {
			return fmt.Errorf("write server-sent event name: %w", err)
		}
	}
	if retry != nil {
		if _, err := fmt.Fprintf(s.w, "retry: %s\n", *retry); err != nil {
			return fmt.Errorf("write server-sent event retry: %w", err)
		}
	}
	data := bytes.ReplaceAll(event.Bytes(), []byte("\r\n"), []byte("\n"))
	data = bytes.ReplaceAll(data, []byte("\r"), []byte("\n"))
	data = bytes.TrimSuffix(data, []byte("\n"))
	for _, line := range bytes.Split(data, []byte("\n")) {
		if _, err := s.w.Write([]byte("data: ")); err != nil {
			return fmt.Errorf("write server-sent event data label: %w", err)
		}
		if _, err := s.w.Write(line); err != nil {
			return fmt.Errorf("write server-sent event data: %w", err)
		}
		if _, err := s.w.Write([]byte("\n")); err != nil {
			return fmt.Errorf("finish server-sent event data line: %w", err)
		}
	}
	if _, err := s.w.Write([]byte("\n")); err != nil {
		return fmt.Errorf("finish server-sent event: %w", err)
	}
	if err := http.NewResponseController(s.w).Flush(); err != nil {
		return fmt.Errorf("flush server-sent event: %w", err)
	}
	return nil
}

// sendError writes one JSON-RPC error as a server-sent event.
func (s *{{ .SSEStream.Name }}) sendError(ctx context.Context, id any, code jsonrpc.Code, message string, data any) error {
	response := jsonrpc.MakeErrorResponse(id, code, message, data)
	return s.sendSSEEvent(ctx, response, nil, nil, nil)
}
{{- end }}
