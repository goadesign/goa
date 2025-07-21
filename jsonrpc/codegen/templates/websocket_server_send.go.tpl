{{ printf "Send streams JSON-RPC responses." | comment }}
func (s *{{ .Service.StructName }}Stream) Send(ctx context.Context, result any) error {
	switch actual := result.(type) {
{{- range .Endpoints }}
	{{- if .Result.Ref }}
	case {{ .Result.Ref }}:
		id := actual.ID
		actual.ID = ""
		return s.send(id, result)
	{{- end }}
{{- end }}
	default:
		return fmt.Errorf("unsupported response type: %T", result)
	}
}

{{ printf "SendError streams JSON-RPC errors." | comment }}
func (s *{{ .Service.StructName }}Stream) SendError(ctx context.Context, id string, err error) error {
	var en goa.GoaErrorNamer
	if !errors.As(err, &en) {
		return s.sendError(ctx, id, jsonrpc.InternalError, err.Error(), nil)
	}
	switch en.GoaErrorName() {
	{{- range allErrors . }}
	case {{ printf "%q" .Name }}:
		{{- with .Response}}
		return s.sendError(ctx, id, {{ .Code }}, err.Error(), err)
		{{- end }}
	{{- end }}
	default:
		return s.sendError(ctx, id, jsonrpc.InternalError, err.Error(), nil)
	}
}

{{ printf "send writes a JSON-RPC response to the websocket connection." | comment }}
func (s *{{ .Service.StructName }}Stream) send(id string, result any) error {
	return s.conn.WriteJSON(jsonrpc.MakeSuccessResponse(id, result))
}

{{ printf "sendError sends a JSON-RPC error response to the websocket connection." | comment }}
func (s *{{ .Service.StructName }}Stream) sendError(ctx context.Context, id string, code jsonrpc.Code, message string, data any) error {
	response := jsonrpc.MakeErrorResponse(id, code, "", message)
	if data != nil {
		response.Error.Message = message
		response.Error.Data = data
	}
	return s.conn.WriteJSON(response)
}
