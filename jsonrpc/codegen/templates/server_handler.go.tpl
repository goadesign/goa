{{- if not (hasMixedTransports) }}
// ServeHTTP handles JSON-RPC requests.
func (s *{{ .ServerStructDeclaration.Name }}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handleHTTP(w, r)
}

{{ comment "handleHTTP reads one JSON-RPC request object or one array of requests." }}
func (s *{{ .ServerStructDeclaration.Name }}) handleHTTP(w http.ResponseWriter, r *http.Request) {
	originalBody := r.Body

	// Find the first JSON byte so leading whitespace does not change whether the
	// body is decoded as one request or an array.
	bufReader := bufio.NewReader(originalBody)
	var peek []byte
	for {
		var err error
		peek, err = bufReader.Peek(1)
		if err != nil && err != io.EOF {
			closeErr := originalBody.Close()
			s.errhandler(r.Context(), w, fmt.Errorf("failed to read request body: %w", errors.Join(err, closeErr)))
			return
		}
		if len(peek) == 0 || (peek[0] != ' ' && peek[0] != '\t' && peek[0] != '\n' && peek[0] != '\r') {
			break
		}
		if _, err := bufReader.Discard(1); err != nil {
			closeErr := originalBody.Close()
			s.errhandler(r.Context(), w, fmt.Errorf("failed to read request body: %w", errors.Join(err, closeErr)))
			return
		}
	}
	
	// The generated handler owns the original body. Decoders receive a wrapper
	// whose Close method cannot close it a second time.
	r.Body = io.NopCloser(bufReader)
	defer func() {
		if err := originalBody.Close(); err != nil {
			s.errhandler(r.Context(), w, fmt.Errorf("failed to close request body: %w", err))
		}
	}()
	
	// A leading '[' starts an array of requests.
	if len(peek) > 0 && peek[0] == '[' {
		s.handleBatch(w, r)
		return
	}
	s.handleSingle(w, r)
}

// handleSingle decodes and runs one JSON-RPC request.
func (s *{{ .ServerStructDeclaration.Name }}) handleSingle(w http.ResponseWriter, r *http.Request) {
	var req jsonrpc.RawRequest
	if err := s.decoder(r).Decode(&req); err != nil {
		// A request that cannot be decoded receives the JSON-RPC parse error.
		response := jsonrpc.MakeErrorResponse(nil, jsonrpc.ParseError, "Parse error", nil)
		if encErr := s.encoder(r.Context(), w).Encode(response); encErr != nil {
			s.errhandler(r.Context(), w, fmt.Errorf("failed to encode parse error response: %w", encErr))
		}
		return
	}
	s.processRequest(r.Context(), r, &req, w)
}
{{- end }}

// handleBatch handles an array of JSON-RPC values and writes the required responses.
func (s *{{ .ServerStructDeclaration.Name }}) handleBatch(w http.ResponseWriter, r *http.Request) {
	var reqs []jsonrpc.RawRequest
	if err := s.decoder(r).Decode(&reqs); err != nil {
		// An array that cannot be decoded receives the JSON-RPC parse error.
		response := jsonrpc.MakeErrorResponse(nil, jsonrpc.ParseError, "Parse error", nil)
		if encErr := s.encoder(r.Context(), w).Encode(response); encErr != nil {
			s.errhandler(r.Context(), w, fmt.Errorf("failed to encode parse error response: %w", encErr))
		}
		return
	}
	if len(reqs) == 0 {
		// JSON-RPC defines an empty request array as one invalid request.
		response := jsonrpc.MakeErrorResponse(nil, jsonrpc.InvalidRequest, "Invalid request", nil)
		if err := s.encoder(r.Context(), w).Encode(response); err != nil {
			s.errhandler(r.Context(), w, fmt.Errorf("failed to encode invalid request response: %w", err))
		}
		return
	}
	
	// Write every response into one JSON array.
	w.Header().Set("Content-Type", "application/json")
	writer := &{{ .BatchWriter.Name }}{Writer: w}
	
	for _, req := range reqs {
		// The writer inserts the array separators around each response.
		s.processRequest(r.Context(), r, &req, writer)
	}
	
	// Write the closing bracket only when at least one request produced a response.
	if writer.written {
		if _, err := writer.Writer.Write([]byte{']'}); err != nil {
			s.errhandler(r.Context(), w, fmt.Errorf("failed to close JSON-RPC batch response: %w", err))
		}
	}
}

// processRequest validates the JSON-RPC version and method, then calls the matching handler.
func (s *{{ .ServerStructDeclaration.Name }}) processRequest(ctx context.Context, r *http.Request, req *jsonrpc.RawRequest, w http.ResponseWriter) {
	if req.Invalid || req.JSONRPC != "2.0" {
		s.encodeJSONRPCError(ctx, w, req, jsonrpc.InvalidRequest, "Invalid request", nil)
		return
	}

	if !req.HasMethod {
		s.encodeJSONRPCError(ctx, w, req, jsonrpc.InvalidRequest, "Missing method field", nil)
		return
	}

	switch req.Method {
	{{- range .Endpoints }}
	{{- if not .SSE }}
	case {{ printf "%q" .Method.Name }}:
		{{- if .IsJSONRPCNotification }}
		if req.HasID {
			s.encodeJSONRPCError(ctx, w, req, jsonrpc.InvalidRequest, "Invalid request", nil)
			return
		}
		{{- else }}
		if !req.HasID {
			s.reportRejectedNotification(ctx, req)
			return
		}
		if req.ID == nil {
			s.encodeJSONRPCError(ctx, w, req, jsonrpc.InvalidRequest, "Invalid request", nil)
			return
		}
		{{- end }}
		if err := s.{{ .Method.VarName }}(ctx, r, req, w); err != nil {
			s.errhandler(ctx, w, fmt.Errorf("handler error for %s: %w", {{ printf "%q" .Method.Name }}, err))
		}
	{{- else }}
	case {{ printf "%q" .Method.Name }}:
		{{- if .IsJSONRPCNotification }}
		if req.HasID {
			s.encodeJSONRPCError(ctx, w, req, jsonrpc.InvalidRequest, "Invalid request", nil)
			return
		}
		{{- else }}
		if !req.HasID {
			s.reportRejectedNotification(ctx, req)
			return
		}
		if req.ID == nil {
			s.encodeJSONRPCError(ctx, w, req, jsonrpc.InvalidRequest, "Invalid request", nil)
			return
		}
		{{- end }}
		s.encodeJSONRPCError(ctx, w, req, jsonrpc.MethodNotFound, "Method is not available in a batch request", nil)
	{{- end }}
	{{- end }}
	default:
		if req.HasID {
			s.encodeJSONRPCError(ctx, w, req, jsonrpc.MethodNotFound, "Method not found", nil)
		}
	}
}

{{ printf "%s inserts JSON array separators around responses from one request array." .BatchWriter.Name | comment }}
type {{ .BatchWriter.Name }} struct {
	io.Writer
	header http.Header
	statusCode int
	written bool
}

func (rb *{{ .BatchWriter.Name }}) Header() http.Header {
	if rb.header == nil {
		rb.header = make(http.Header)
	}
	return rb.header
}

func (rb *{{ .BatchWriter.Name }}) WriteHeader(statusCode int) {
	if rb.written {
		return
	}
	rb.statusCode = statusCode
}

func (rb *{{ .BatchWriter.Name }}) Write(data []byte) (int, error) {
	separator := byte(',')
	if !rb.written {
		separator = '['
	}
	if _, err := rb.Writer.Write([]byte{separator}); err != nil {
		return 0, err
	}
	rb.written = true
	return rb.Writer.Write(data)
}
