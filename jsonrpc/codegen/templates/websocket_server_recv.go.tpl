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
			{{- if and .Method.ServerStream (or (eq .Method.ServerStream.Kind 3) (eq .Method.ServerStream.Kind 4)) }}
			// {{ if eq .Method.ServerStream.Kind 3 }}Server{{ else }}Bidirectional{{ end }} streaming: decode payload and create stream wrapper
			payload, err := s.{{ lowerInitial .Method.VarName }}(ctx, s.r, req)
			if err != nil {
				return fmt.Errorf("handler error for %s: %w", {{ printf "%q" .Method.Name }}, err)
			}
			// Create wrapper that implements the method-specific stream interface
			streamWrapper := &{{ lowerInitial .Method.VarName }}StreamWrapper{
				stream: s,
				requestID: req.ID,
			}
			// Call the endpoint with payload and stream wrapper
			endpointInput := &{{ .ServicePkgName }}.{{ .Method.ServerStream.EndpointStruct }}{
				{{- if .Payload.Ref }}
				Payload: payload.({{ .Payload.Ref }}),
				{{- end }}
				Stream: streamWrapper,
			}
			if _, err := s.{{ lowerInitial .Method.VarName }}Endpoint(ctx, endpointInput); err != nil {
				// For streaming endpoints, send error as JSON-RPC error response
				if req.ID != nil {
					// Send error response to client
					if sendErr := streamWrapper.SendError(ctx, err); sendErr != nil {
						return fmt.Errorf("failed to send error response: %w", sendErr)
					}
					// Continue processing other requests
					return nil
				}
				// For notifications (no ID), just log and continue
				return nil
			}
			return nil
			{{- else }}
			res, err := s.{{ lowerInitial .Method.VarName }}(ctx, s.r, req)
			if err != nil {
				return fmt.Errorf("handler error for %s: %w", {{ printf "%q" .Method.Name }}, err)
			}
			if err := s.Send{{ .Method.VarName }}(ctx, res.({{ printf "*%s.%sResult" .ServicePkgName .Method.VarName }})); err != nil {
				return fmt.Errorf("send error for %s: %w", {{ printf "%q" .Method.Name }}, err)
			}
			return nil
			{{- end }}
	{{- end }}
	default:
		if req.ID != nil {
			return s.sendError(ctx, req.ID, jsonrpc.MethodNotFound, fmt.Sprintf("Method %q not found", req.Method), nil)
		}
		return nil
	}
}

