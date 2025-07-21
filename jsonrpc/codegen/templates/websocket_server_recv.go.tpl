{{ printf "Recv reads JSON-RPC requests from the %s service stream." .Service.Name | comment }}
func (s *{{ .Service.StructName }}Stream) Recv(ctx context.Context) error {
	var req jsonrpc.RawRequest
	if err := s.conn.ReadJSON(&req); err != nil {
		return err
	}
	return s.processRequest(ctx, &req)
}

func (s *{{ .Service.StructName }}Stream) processRequest(ctx context.Context, req *jsonrpc.RawRequest) error {
	if req.JSONRPC != "2.0" {
		if req.ID != nil {
			return s.sendError(ctx, *req.ID, jsonrpc.InvalidRequest, fmt.Sprintf("Invalid JSON-RPC version, must be 2.0, got %q", req.JSONRPC), nil)
		}
		return nil
	}

	if req.Method == "" {
		if req.ID != nil {
			return s.sendError(ctx, *req.ID, jsonrpc.InvalidRequest, "Missing method field", nil)
		}
		return nil
	}

	switch req.Method {
	{{- range .Endpoints }}
		case {{ printf "%q" .Method.Name }}:
			return s.{{ .Method.VarName }}(ctx, s.r, req)
	{{- end }}
	default:
		if req.ID != nil {
			return s.sendError(ctx, *req.ID, jsonrpc.MethodNotFound, fmt.Sprintf("Method %q not found", req.Method), nil)
		}
		return nil
	}
}

