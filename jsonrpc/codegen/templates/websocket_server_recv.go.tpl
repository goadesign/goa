{{ printf "Recv reads JSON-RPC requests from the %s service stream." .Service.Name | comment }}
func (s *{{ lowerInitial .Service.StructName }}Stream) Recv(ctx context.Context) error {
	var req jsonrpc.RawRequest
	if err := s.conn.ReadJSON(&req); err != nil {
		// Handle different types of errors gracefully
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			// Network/connection errors - terminate connection
			return err
		}
		
		// JSON parse errors - send Parse Error response and continue
		if err := s.sendError(ctx, nil, jsonrpc.ParseError, "Parse error", nil); err != nil {
			// If we can't send error response, connection is broken
			return fmt.Errorf("failed to send parse error: %w", err)
		}
		// Continue processing after sending parse error
		return nil
	}
	return s.processRequest(ctx, &req)
}

func (s *{{ lowerInitial .Service.StructName }}Stream) processRequest(ctx context.Context, req *jsonrpc.RawRequest) error {
	if req.JSONRPC != "2.0" {
		if req.ID != nil {
			return s.sendError(ctx, req.ID, jsonrpc.InvalidRequest, fmt.Sprintf("Invalid JSON-RPC version, must be 2.0, got %q", req.JSONRPC), nil)
		}
		return nil
	}

	if req.Method == "" {
		if req.ID != nil {
			return s.sendError(ctx, req.ID, jsonrpc.InvalidRequest, "Missing method field", nil)
		}
		return nil
	}

	switch req.Method {
	{{- range .Endpoints }}
		case {{ printf "%q" .Method.Name }}:
			res, err := s.{{ lowerInitial .Method.VarName }}(ctx, s.r, req)
			if err != nil {
				return fmt.Errorf("handler error for %s: %w", {{ printf "%q" .Method.Name }}, err)
			}
			if err := s.Send{{ .Method.VarName }}(ctx, res.({{ printf "*%s.%sResult" .ServicePkgName .Method.VarName }})); err != nil {
				return fmt.Errorf("send error for %s: %w", {{ printf "%q" .Method.Name }}, err)
			}
			return nil
	{{- end }}
	default:
		if req.ID != nil {
			return s.sendError(ctx, req.ID, jsonrpc.MethodNotFound, fmt.Sprintf("Method %q not found", req.Method), nil)
		}
		return nil
	}
}

