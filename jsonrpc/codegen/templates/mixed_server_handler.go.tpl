// ServeHTTP decodes one request and uses the response type designed for its method.
func (s *{{ .ServerStructDeclaration.Name }}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	acceptJSON := false
	acceptSSE := false
	acceptValues := r.Header.Values("Accept")
	if len(acceptValues) == 0 || len(acceptValues) == 1 && strings.TrimSpace(acceptValues[0]) == "" {
		acceptJSON = true
		acceptSSE = true
	} else {
		for _, header := range acceptValues {
			for _, value := range strings.Split(header, ",") {
				mediaType, params, err := mime.ParseMediaType(value)
				if err != nil {
					continue
				}
				quality := 1.0
				if value, ok := params["q"]; ok {
					quality, err = strconv.ParseFloat(value, 64)
					if err != nil {
						continue
					}
				}
				if quality <= 0 {
					continue
				}
				switch mediaType {
				case "*/*":
					acceptJSON = true
					acceptSSE = true
				case "application/json", "application/*":
					acceptJSON = true
				case "text/event-stream", "text/*":
					acceptSSE = true
				}
			}
		}
	}

	originalBody := r.Body
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
	r.Body = io.NopCloser(bufReader)

	// Request arrays always use ordinary JSON-RPC responses. Streaming methods
	// in an array receive one method error and are not called.
	if len(peek) > 0 && peek[0] == '[' {
		defer func() {
			if err := originalBody.Close(); err != nil {
				s.errhandler(r.Context(), w, fmt.Errorf("failed to close request body: %w", err))
			}
		}()
		if !acceptJSON {
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}
		s.handleBatch(w, r)
		return
	}

	// Decode the request once so the generated method switch below can choose
	// both the handler and its response type.
	var req jsonrpc.RawRequest
	if err := s.decoder(r).Decode(&req); err != nil {
		closeErr := originalBody.Close()
		s.errhandler(r.Context(), w, fmt.Errorf("failed to read request body: %w", errors.Join(err, closeErr)))
		switch {
		case acceptJSON:
			response := jsonrpc.MakeErrorResponse(nil, jsonrpc.ParseError, "Parse error", nil)
			if encErr := s.encoder(r.Context(), w).Encode(response); encErr != nil {
				s.errhandler(r.Context(), w, fmt.Errorf("failed to encode parse error response: %w", encErr))
			}
		case acceptSSE:
			stream := &{{ .SSEStream.Name }}{w: w, encoder: s.encoder}
			if sendErr := stream.sendError(r.Context(), nil, jsonrpc.ParseError, "Parse error", nil); sendErr != nil {
				s.errhandler(r.Context(), w, fmt.Errorf("write parse error event: %w", sendErr))
			}
		default:
			w.WriteHeader(http.StatusNotAcceptable)
		}
		return
	}
	defer func() {
		if err := originalBody.Close(); err != nil {
			s.errhandler(r.Context(), w, fmt.Errorf("failed to close request body: %w", err))
		}
	}()

	// Invalid and unknown requests do not have a designed response type. Use
	// JSON when the client accepts it, then events, or reject the response.
	if req.Invalid || req.JSONRPC != "2.0" || req.Method == "" {
		switch {
		case acceptJSON:
			s.processRequest(r.Context(), r, &req, w)
		case acceptSSE:
			s.processSSERequest(r.Context(), r, &req, w)
		default:
			w.WriteHeader(http.StatusNotAcceptable)
		}
		return
	}

	switch req.Method {
{{- range .Endpoints }}
	{{- if .SSE }}
	case {{ printf "%q" .Method.Name }}:
		if !acceptSSE {
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}
		s.processSSERequest(r.Context(), r, &req, w)
	{{- else }}
	case {{ printf "%q" .Method.Name }}:
		if !acceptJSON {
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}
		s.processRequest(r.Context(), r, &req, w)
	{{- end }}
{{- end }}
	default:
		switch {
		case acceptJSON:
			s.processRequest(r.Context(), r, &req, w)
		case acceptSSE:
			s.processSSERequest(r.Context(), r, &req, w)
		default:
			w.WriteHeader(http.StatusNotAcceptable)
		}
	}
}
