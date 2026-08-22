type (
	{{ printf "%s writes JSON-RPC messages as server-sent events." .Stream.Name | comment }}
	{{ .Stream.Name }} struct {
		// once writes the HTTP headers only for the first event.
		once sync.Once
		// w receives the HTTP headers and event bytes.
		w http.ResponseWriter
		// encoder turns one JSON-RPC message into bytes.
		encoder func(context.Context, http.ResponseWriter) goahttp.Encoder
	}

	{{ printf "%s stores an encoded event before any HTTP output is written." .Buffer.Name | comment }}
	{{ .Buffer.Name }} struct {
		bytes.Buffer
		header http.Header
	}
)

func (b *{{ .Buffer.Name }}) Header() http.Header {
	return b.header
}

func (b *{{ .Buffer.Name }}) WriteHeader(int) {}

// initSSEHeaders writes the response headers before the first event.
func (s *{{ .Stream.Name }}) initSSEHeaders() {
	s.once.Do(func() {
		header := s.w.Header()
		header.Set("Content-Type", "text/event-stream")
		header.Set("Cache-Control", "no-cache")
		header.Set("Connection", "keep-alive")
		header.Set("X-Accel-Buffering", "no")
		s.w.WriteHeader(http.StatusOK)
	})
}

// sendSSEEvent encodes one event before starting the response, then writes and
// flushes that complete event.
func (s *{{ .Stream.Name }}) sendSSEEvent(ctx context.Context, eventType string, value any) error {
	event := &{{ .Buffer.Name }}{header: make(http.Header)}
	if err := s.encoder(ctx, event).Encode(value); err != nil {
		return err
	}

	s.initSSEHeaders()
	if eventType != "" {
		if _, err := fmt.Fprintf(s.w, "event: %s\n", eventType); err != nil {
			return fmt.Errorf("write server-sent event name: %w", err)
		}
	}
	if _, err := s.w.Write([]byte("data: ")); err != nil {
		return fmt.Errorf("write server-sent event data label: %w", err)
	}
	if _, err := s.w.Write(event.Bytes()); err != nil {
		return fmt.Errorf("write server-sent event data: %w", err)
	}
	if _, err := s.w.Write([]byte("\n\n")); err != nil {
		return fmt.Errorf("finish server-sent event: %w", err)
	}
	if err := http.NewResponseController(s.w).Flush(); err != nil {
		return fmt.Errorf("flush server-sent event: %w", err)
	}
	return nil
}

// sendError writes one JSON-RPC error as a server-sent event.
func (s *{{ .Stream.Name }}) sendError(ctx context.Context, id any, code jsonrpc.Code, message string, data any) error {
	response := jsonrpc.MakeErrorResponse(id, code, message, data)
	return s.sendSSEEvent(ctx, "error", response)
}
