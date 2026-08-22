{{ printf "Recv reads JSON-RPC requests from the %s service stream." .Service.Name | comment }}
func (s *{{ websocketServerStreamName }}) Recv(ctx context.Context) error {
	var req jsonrpc.RawRequest
	if err := s.conn.ReadJSON(&req); err != nil {
		// Return an unexpected connection close because no later request can be read.
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			return err
		}
		
		// Report every other read failure as a JSON-RPC parse error.
		if err := s.sendError(ctx, nil, jsonrpc.ParseError, "Parse error", nil); err != nil {
			// Return when the parse-error response cannot be written to the client.
			return fmt.Errorf("failed to send parse error: %w", err)
		}
		// The next Recv call reads the next request from this connection.
		return nil
	}
	return s.processRequest(ctx, &req)
}

func (s *{{ websocketServerStreamName }}) processRequest(ctx context.Context, req *jsonrpc.RawRequest) error {
	if req.JSONRPC != "2.0" {
		if req.HasID {
			return s.sendError(ctx, req.ID, jsonrpc.InvalidRequest, "Invalid request", nil)
		}
		return nil
	}

	if req.Method == "" {
		if req.HasID {
			return s.sendError(ctx, req.ID, jsonrpc.InvalidRequest, "Invalid request", nil)
		}
		return nil
	}

	switch req.Method {
	{{- range .Endpoints }}
		case {{ printf "%q" .Method.Name }}:
			{{- if and .Method.ServerStream (or (eq .Method.ServerStream.Kind 3) (eq .Method.ServerStream.Kind 4)) }}
			// Decode the request fields for this {{ if eq .Method.ServerStream.Kind 3 }}server-streaming{{ else }}bidirectional-streaming{{ end }} call.
			{{- if .Payload.Ref }}
			payload, err := s.{{ lowerInitial .Method.VarName }}(ctx, s.r, req)
			{{- else }}
			_, err := s.{{ lowerInitial .Method.VarName }}(ctx, s.r, req)
			{{- end }}
			if err != nil {
				return fmt.Errorf("handler error for %s: %w", {{ printf "%q" .Method.Name }}, err)
			}
			// Give the service a stream that writes responses on this connection
			// with the ID from this request.
			streamWrapper := &{{ websocketWrapperName .Method.Name }}{
				stream: s,
				requestID: req.ID,
				{{- if and .Method.ViewedResult (not .Method.ViewedResult.ViewName) }}
				view: {{ printf "%q" .Result.View }},
				{{- end }}
			}
			// Pass the decoded payload, when present, and this request's stream
			// to the service.
			endpointInput := &{{ .ServicePkgName }}.{{ .Method.ServerStream.EndpointStruct }}{
				{{- if .Payload.Ref }}
				Payload: payload.({{ .Payload.Ref }}),
				{{- end }}
				Stream: streamWrapper,
			}
			if _, err := s.{{ lowerInitial .Method.VarName }}Endpoint(ctx, endpointInput); err != nil {
				// Send the service error to callers that supplied a request ID.
				if req.HasID {
					if sendErr := streamWrapper.SendError(ctx, err); sendErr != nil {
						return fmt.Errorf("failed to send error response: %w", sendErr)
					}
					// The error response completes this request. The next Recv call
					// reads another request from the same connection.
					return nil
				}
				// Notifications have no response, so finish this request without
				// writing to the connection.
				return nil
			}
			return nil
			{{- else }}
			res, err := s.{{ lowerInitial .Method.VarName }}(ctx, s.r, req)
			if err != nil {
				// Send the call error only when the caller supplied a request ID.
				if req.HasID {
					if sendErr := s.SendError(ctx, req.ID, err); sendErr != nil {
						return fmt.Errorf("failed to send error response: %w", sendErr)
					}
				}
				return nil
			}
			// A notification has no request ID and receives no response.
			if req.HasID {
				if res == nil {
					return s.sendError(ctx, req.ID, jsonrpc.InternalError, "Internal error", nil)
				}
				if r, ok := res.({{ printf "*%s.%sResult" .ServicePkgName .Method.VarName }}); ok {
					if err := s.Send{{ .Method.VarName }}Response(ctx, req.ID, r); err != nil {
						return fmt.Errorf("send response error for %s: %w", {{ printf "%q" .Method.Name }}, err)
					}
				} else {
					return s.sendError(ctx, req.ID, jsonrpc.InternalError, "Internal error", nil)
				}
			}
			return nil
			{{- end }}
	{{- end }}
	default:
		if req.HasID {
			return s.sendError(ctx, req.ID, jsonrpc.MethodNotFound, "Method not found", nil)
		}
		return nil
	}
}
