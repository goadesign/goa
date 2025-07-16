// ServeHTTP handles JSON-RPC requests.
func (s *{{ .ServerStruct }}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	resp := s.processRequest(r.Context(), r, &req)
	if resp == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := s.encoder(r.Context(), w).Encode(resp); err != nil {
		s.errhandler(r.Context(), w, fmt.Errorf("failed to encode response: %w", err))
	}
}

// handleBatch handles a batch of JSON-RPC requests.
func (s *Server) handleBatch(w http.ResponseWriter, r *http.Request) {
	var reqs []jsonrpc.RawRequest
	if err := s.decoder(r).Decode(&reqs); err != nil {
		s.errhandler(r.Context(), w, fmt.Errorf("failed to decode batch request: %w", err))
		return
	}

	resps := make([]jsonrpc.Response, 0, len(reqs))
	for _, req := range reqs {
		if resp := s.processRequest(r.Context(), r, &req); resp != nil {
			resps = append(resps, *resp)
		}
	}

	if err := s.encoder(r.Context(), w).Encode(resps); err != nil {
		s.errhandler(r.Context(), w, fmt.Errorf("failed to encode batch response: %w", err))
	}
}

// ProcessRequest processes a single JSON-RPC request.
func (s *Server) processRequest(ctx context.Context, r *http.Request, req *jsonrpc.RawRequest) *jsonrpc.Response {
	if req.JSONRPC != "2.0" {
		if req.ID != nil {
			return jsonrpc.MakeErrorResponse(*req.ID, jsonrpc.InvalidRequest, "", fmt.Sprintf("Invalid JSON-RPC version, must be 2.0, got %q", req.JSONRPC))
		}
		return nil
	}

	if req.Method == "" {
		if req.ID != nil {
			return jsonrpc.MakeErrorResponse(*req.ID, jsonrpc.InvalidRequest, "", "Missing method field")
		}
		return nil
	}

	var resp *jsonrpc.Response
	switch req.Method {
        {{- range .Endpoints }}
        case {{ printf "%q" .Method.Name }}:
            resp = s.{{ .Method.VarName }}(ctx, r, req)
        {{- end }}
	default:
		if req.ID != nil {
			return jsonrpc.MakeErrorResponse(*req.ID, jsonrpc.MethodNotFound, "", fmt.Sprintf("Method %q not found", req.Method))
		}
		return nil
	}

	return resp
}
