// ServeHTTP handles JSON-RPC requests.
func (s *{{ .ServerStruct }}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
        // Peek at the first byte to determine request type
	bufReader := bufio.NewReader(r.Body)
	peek, err := bufReader.Peek(1)
	if err != nil && err != io.EOF {
		r.Body.Close()
		s.writeError(r.Context(), w, nil, jsonrpc.ParseError, fmt.Errorf("Failed to read request body: %w", err))
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
			s.writeError(r.Context(), w, nil, jsonrpc.InternalError, fmt.Errorf("Failed to close request body: %w", err))
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
	var req jsonrpc.Request
	if err := s.decoder(r.Body).Decode(&req); err != nil {
		s.writeError(r.Context(), w, nil, jsonrpc.ParseError, fmt.Errorf("Failed to decode request: %w", err))
		return
	}

	resp := s.processRequest(r.Context(), &req)
	if resp == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := s.encoder(w).Encode(resp); err != nil {
		s.writeError(r.Context(), w, req.ID, jsonrpc.InternalError, fmt.Errorf("Failed to encode response: %w", err))
	}
}

// handleBatch handles a batch of JSON-RPC requests.
func (s *Server) handleBatch(w http.ResponseWriter, r *http.Request) {
	var reqs []jsonrpc.Request
	if err := s.decoder(r.Body).Decode(&reqs); err != nil {
		s.writeError(r.Context(), w, nil, jsonrpc.ParseError, fmt.Errorf("Invalid JSON: %w", err))
		return
	}

	if len(reqs) == 0 {
		s.writeError(r.Context(), w, nil, jsonrpc.InvalidRequest, fmt.Errorf("Empty batch request"))
		return
	}

	responses := make([]jsonrpc.Response, 0, len(reqs))
	for _, req := range reqs {
		if resp := s.processRequest(r.Context(), &req); resp != nil {
			responses = append(responses, *resp)
		}
	}

	if err := s.encoder(w).Encode(responses); err != nil {
		s.writeError(r.Context(), w, nil, jsonrpc.InternalError, fmt.Errorf("Failed to encode batch response: %w", err))
	}
}

// ProcessRequest processes a single JSON-RPC request.
func (s *Server) processRequest(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
	if req.JSONRPC != "2.0" {
		return jsonrpc.MakeErrorResponse(req.ID, jsonrpc.InvalidRequest, fmt.Sprintf("Invalid JSON-RPC version, must be 2.0, got %q", req.JSONRPC), nil)
	}

	if req.Method == "" {
		return jsonrpc.MakeErrorResponse(req.ID, jsonrpc.InvalidRequest, "Missing method field", nil)
	}

	var resp *jsonrpc.Response
	switch req.Method {
        {{- range .Endpoints }}
        case {{ printf "%q" .Method.Name }}:
            resp = s.{{ .Method.VarName }}(ctx, req)
        {{- end }}
	default:
		if req.ID != nil {
			resp = jsonrpc.MakeErrorResponse(req.ID, jsonrpc.MethodNotFound, fmt.Sprintf("Method %q not found", req.Method), nil)
		}
	}

	return resp
}

// writeError writes a JSON-RPC error response.
func (s *{{ .ServerStruct }}) writeError(ctx context.Context, w http.ResponseWriter, reqID any, code jsonrpc.Code, err error) {
	resp := jsonrpc.MakeErrorResponse(reqID, code, err.Error(), nil)
	if err := s.encoder(w).Encode(resp); err != nil {
		s.errhandler(ctx, w, err)
	}
}
