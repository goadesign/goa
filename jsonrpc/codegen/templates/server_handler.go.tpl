{{- if and (not (isWebSocketEndpoint (index .Endpoints 0))) (not (hasMixedTransports)) }}
// ServeHTTP handles JSON-RPC requests.
func (s *{{ .ServerStruct }}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handleHTTP(w, r)
}
{{- end }}

{{- comment "handleHTTP handles JSON-RPC requests." }}
func (s *{{ .ServerStruct }}) handleHTTP(w http.ResponseWriter, r *http.Request) {
	// Peek at the first byte to determine request type
	bufReader := bufio.NewReader(r.Body)
	peek, err := bufReader.Peek(1)
	if err != nil && err != io.EOF {
		r.Body.Close()
		s.errhandler(r.Context(), w, fmt.Errorf("failed to read request body: %w", err))
		return
	}
	
	// Wrap the buffered reader with the original closer
	r.Body = struct {
		io.Reader
		io.Closer
	}{
		Reader: bufReader,
		Closer: r.Body,
	}
	defer func(r *http.Request) {
		if err := r.Body.Close(); err != nil {
			s.errhandler(r.Context(), w, fmt.Errorf("failed to close request body: %w", err))
		}
	}(r)
	
	// Route to appropriate handler
	if len(peek) > 0 && peek[0] == '[' {
		s.handleBatch(w, r)
		return
	}
	s.handleSingle(w, r)
}

// handleSingle handles a single JSON-RPC request.
func (s *Server) handleSingle(w http.ResponseWriter, r *http.Request) {
	var req jsonrpc.RawRequest
	if err := s.decoder(r).Decode(&req); err != nil {
		// JSON-RPC parse error with null id and generic message
		response := jsonrpc.MakeErrorResponse(nil, jsonrpc.ParseError, "Parse error", nil)
		if encErr := s.encoder(r.Context(), w).Encode(response); encErr != nil {
			s.errhandler(r.Context(), w, fmt.Errorf("failed to encode parse error response: %w", encErr))
		}
		return
	}
	s.processRequest(r.Context(), r, &req, w)
}

// handleBatch handles a batch of JSON-RPC requests.
func (s *Server) handleBatch(w http.ResponseWriter, r *http.Request) {
	var reqs []jsonrpc.RawRequest
	if err := s.decoder(r).Decode(&reqs); err != nil {
		// JSON-RPC parse error for batch with null id and generic message
		response := jsonrpc.MakeErrorResponse(nil, jsonrpc.ParseError, "Parse error", nil)
		if encErr := s.encoder(r.Context(), w).Encode(response); encErr != nil {
			s.errhandler(r.Context(), w, fmt.Errorf("failed to encode parse error response: %w", encErr))
		}
		return
	}
	
	// Write responses
	w.Header().Set("Content-Type", "application/json")
	writer := &batchWriter{Writer: w}
	
	for _, req := range reqs {
		// Process the request with batch writer
		s.processRequest(r.Context(), r, &req, writer)
	}
	
	// Close the batch array
	if writer.written {
		writer.Writer.Write([]byte{']'})
	}
}

// ProcessRequest processes a single JSON-RPC request.
func (s *Server) processRequest(ctx context.Context, r *http.Request, req *jsonrpc.RawRequest, w http.ResponseWriter) {
	if req.JSONRPC != "2.0" {
		s.encodeJSONRPCError(ctx, w, req, jsonrpc.InvalidRequest, "Invalid request", nil)
		return
	}

	if req.Method == "" {
		s.encodeJSONRPCError(ctx, w, req, jsonrpc.InvalidRequest, "Missing method field", nil)
		return
	}

	switch req.Method {
	{{- range .Endpoints }}
	case {{ printf "%q" .Method.Name }}:
		if err := s.{{ .Method.VarName }}(ctx, r, req, w); err != nil {
			s.errhandler(ctx, w, fmt.Errorf("handler error for %s: %w", {{ printf "%q" .Method.Name }}, err))
		}
	{{- end }}
	default:
		s.encodeJSONRPCError(ctx, w, req, jsonrpc.MethodNotFound, "Method not found", nil)
	}
}

// batchWriter is a helper type that implements http.ResponseWriter for writing multiple JSON-RPC responses
type batchWriter struct {
	io.Writer
	header http.Header
	statusCode int
	written bool
}

func (rb *batchWriter) Header() http.Header {
	if rb.header == nil {
		rb.header = make(http.Header)
	}
	return rb.header
}

func (rb *batchWriter) WriteHeader(statusCode int) {
	if rb.written {
		return
	}
	rb.statusCode = statusCode
}

func (rb *batchWriter) Write(data []byte) (int, error) {
	if !rb.written {
		rb.written = true
		rb.Writer.Write([]byte{'['})
	} else {
		rb.Writer.Write([]byte{','})
	}
	return rb.Writer.Write(data)
}
