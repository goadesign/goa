// ServeHTTP handles JSON-RPC requests.
func (s *{{ .ServerStruct }}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
{{- if isWebSocketEndpoint (index .Endpoints 0) }}
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.errhandler(r.Context(), w, fmt.Errorf("failed to upgrade to WebSocket: %w", err))
		return
	}
	conn = s.configurer.ConfigFn(conn, cancel)
	defer conn.Close()

	stream := &{{ .Service.StructName }}Stream{
	{{- range .Endpoints }}
		{{ .Method.VarName }}: s.{{ .Method.VarName }},
	{{- end }}
		r: r,
		w: w,
		conn: conn,
		cancel: cancel,
	}
	s.Stream(ctx, stream)
{{- else }}
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
		s.errhandler(r.Context(), w, fmt.Errorf("failed to decode request: %w", err))
		return
	}
	s.processRequest(r.Context(), r, &req, w)
}

// handleBatch handles a batch of JSON-RPC requests.
func (s *Server) handleBatch(w http.ResponseWriter, r *http.Request) {
	var reqs []jsonrpc.RawRequest
	if err := s.decoder(r).Decode(&reqs); err != nil {
		s.errhandler(r.Context(), w, fmt.Errorf("failed to decode batch request: %w", err))
		return
	}
	for _, req := range reqs {
		s.processRequest(r.Context(), r, &req, w)
	}
}

// ProcessRequest processes a single JSON-RPC request.
func (s *Server) processRequest(ctx context.Context, r *http.Request, req *jsonrpc.RawRequest, w http.ResponseWriter) {
	if req.JSONRPC != "2.0" {
		s.encodeJSONRPCError(ctx, w, req, jsonrpc.InvalidRequest, fmt.Sprintf("Invalid JSON-RPC version, must be 2.0, got %q", req.JSONRPC), nil)
		return
	}

	if req.Method == "" {
		s.encodeJSONRPCError(ctx, w, req, jsonrpc.InvalidRequest, "Missing method field", nil)
		return
	}

	switch req.Method {
	{{- range .Endpoints }}
	case {{ printf "%q" .Method.Name }}:
		s.{{ .Method.VarName }}(ctx, r, req, w)
	{{- end }}
	default:
		s.encodeJSONRPCError(ctx, w, req, jsonrpc.MethodNotFound, fmt.Sprintf("Method %q not found", req.Method), nil)
	}
{{- end }}
}
