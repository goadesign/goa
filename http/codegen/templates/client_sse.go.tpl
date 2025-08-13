// {{ .Method.VarName }}ClientStream is the interface for reading Server-Sent Events.
type {{ .Method.VarName }}ClientStream interface {
    // Recv reads and returns the next event from the SSE stream.
    Recv(context.Context) ({{ .SSE.EventTypeRef }}, error)
    // Close closes the SSE stream and releases resources.
    Close() error
}

type (
        // {{ .Method.VarName }}StreamImpl implements the {{ .Method.VarName }}ClientStream interface.
        {{ .Method.VarName }}StreamImpl struct {
                resp *http.Response
                decoder func(*http.Response) goahttp.Decoder
                buffer []byte // Buffer for unprocessed data
                lock sync.Mutex
                closed bool
        }
)

// {{ .Method.VarName }}StreamImpl implements the {{ .Method.VarName }}ClientStream interface.
var _ {{ .Method.VarName }}ClientStream = (*{{ .Method.VarName }}StreamImpl)(nil)

// New{{ .Method.VarName }}Stream creates a new {{ .Method.VarName }}ClientStream.
func New{{ .Method.VarName }}Stream(resp *http.Response, decoder func(*http.Response) goahttp.Decoder) {{ .Method.VarName }}ClientStream {
        return &{{ .Method.VarName }}StreamImpl{
                resp: resp,
                decoder: decoder,
                buffer: make([]byte, 0, 4096), // Pre-allocate buffer
        }
}

// Recv reads and returns the next event from the SSE stream, respecting context cancellation.
func (s *{{ .Method.VarName }}StreamImpl) Recv(ctx context.Context) (event {{ .SSE.EventTypeRef }}, err error) {
        var byts []byte
        byts, err = s.readEvent(ctx)
        if err != nil {
                if errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
                        // Clean up on EOF or context cancellation
                        s.Close()
                        if errors.Is(err, io.EOF) {
                                err = nil
                        }
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
func (s *{{ .Method.VarName }}StreamImpl) readEvent(ctx context.Context) ([]byte, error) {
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

        // Read data in chunks until we find an event or hit EOF
        for {
                // Check if context is done
                select {
                case <-ctx.Done():
                        if len(eventData) > 0 {
                                return eventData, nil
                        }
                        return nil, ctx.Err()
                default:
                        // Continue processing
                }

                // Check if stream is closed
                s.lock.Lock()
                if s.closed {
                        s.lock.Unlock()
                        if len(eventData) > 0 {
                                return eventData, nil
                        }
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

                // Return partial data at EOF
                if errors.Is(err, io.EOF) {
                        if len(eventData) > 0 {
                                return eventData, nil
                        }
                        return nil, io.EOF
                }
        }
}

// checkBuffer examines the internal buffer for a complete SSE event (delimited
// by double newlines).  It returns two values: the event data (or all buffer
// contents if no complete event is found), and a boolean indicating whether a
// complete event was found. If a complete event is found, any remaining data
// after the event is kept in the buffer for the next call.
func (s *{{ .Method.VarName }}StreamImpl) checkBuffer() ([]byte, bool) {
        s.lock.Lock()
        defer s.lock.Unlock()

        // Quick return if buffer is empty
        if len(s.buffer) == 0 {
                return nil, false
        }

        // Look for double newline in buffer
        for i := 0; i < len(s.buffer)-1; i++ {
                if s.buffer[i] == '\n' && s.buffer[i+1] == '\n' {
                        // Found complete event
                        eventEnd := i + 2 // Include both newlines
                        eventData := s.buffer[:eventEnd]

                        // Save remaining data for next time
                        if eventEnd < len(s.buffer) {
                                s.buffer = append(s.buffer[:0], s.buffer[eventEnd:]...)
                        } else {
                                s.buffer = s.buffer[:0]
                        }

                        return eventData, true
                }
        }

        // No complete event found, return buffer contents
        eventData := s.buffer
        s.buffer = s.buffer[:0] // Clear buffer but keep capacity
        return eventData, false
}

// Close closes the SSE stream and releases any associated resources.
func (s *{{ .Method.VarName }}StreamImpl) Close() error {
        s.lock.Lock()
        defer s.lock.Unlock()
        if s.closed {
                return nil
        }
        s.closed = true
        return s.resp.Body.Close()
}

// processEvent processes a raw SSE event into the expected type
func (s *{{ .Method.VarName }}StreamImpl) processEvent(eventData []byte) (event {{ .SSE.EventTypeRef }}, err error) {
        {{- if .SSE.EventIsStruct }}
        event = &{{ .SSE.EventTypeName }}{}
        {{- end }}
        {{- if .SSE.IDField }}
        var id string
        {{- end }}
        {{- if .SSE.EventField }}
        var eventType string
        {{- end }}
        {{- if .SSE.RetryField }}
        var retry int
        {{- end }}
        var dataLines []string
        for _, line := range bytes.Split(eventData, []byte("\n")) {
                if len(line) == 0 {
                        continue
                }
                if bytes.HasPrefix(line, []byte("data:")) {
                        dataLines = append(dataLines, s.trimHeader(len("data:"), line))
                        continue
                }
                {{- if .SSE.IDField }}
                if bytes.HasPrefix(line, []byte("id:")) {
                        event.{{ .SSE.IDField }} = s.trimHeader(len("id:"), line)
                        continue
                }
                {{- end }}
                {{- if .SSE.EventField }}
                if bytes.HasPrefix(line, []byte("event:")) {
                        event.{{ .SSE.EventField }} = s.trimHeader(len("event:"), line)
                        continue
                }
                {{- end }}
                {{- if .SSE.RetryField }}
                if bytes.HasPrefix(line, []byte("retry:")) {
                        event.{{ .SSE.RetryField }} = s.trimHeader(len("retry:"), line)
                        continue
                }
                {{- end }}
        }
        if len(dataLines) > 0 {
                dataContent := strings.Join(dataLines, "\n")
                {{- if .SSE.DataField }}
                {{ template "partial_sse_parse" dict "Target" (printf "event.%s" .SSE.DataField) "TypeRef" .SSE.DataFieldTypeRef }}
                {{- else }}
                {{ template "partial_sse_parse" dict "Target" "event" "TypeRef" .SSE.EventTypeRef }}
                {{- end }}
        }
        return
}

// trimHeader removes the header prefix and optional leading space
func (s *{{ .Method.VarName }}StreamImpl) trimHeader(size int, data []byte) string {
        if len(data) < size {
                return string(data)
        }
        data = data[size:]
        if len(data) > 0 && data[0] == ' ' {
                data = data[1:]
        }
        return string(data)
}
