{{- range .Endpoints }}
	{{- if .Result.Ref }}
{{ printf "Send%sNotification sends a JSON-RPC notification for the %s method." .Method.VarName .Method.Name | comment }}
func (s *{{ lowerInitial $.Service.StructName }}Stream) Send{{ .Method.VarName }}Notification(ctx context.Context, result {{ .Result.Ref }}) error {
	{{- if and .Result (index .Result.Responses 0).ServerBody (index (index .Result.Responses 0).ServerBody 0).Init }}
	body := {{ (index (index .Result.Responses 0).ServerBody 0).Init.Name }}(result)
	{{- else }}
	body := result
	{{- end }}
	return s.conn.WriteJSON(jsonrpc.MakeNotification({{ printf "%q" .Method.Name }}, body))
}

{{ printf "Send%sResponse sends a JSON-RPC response for the %s method." .Method.VarName .Method.Name | comment }}
func (s *{{ lowerInitial $.Service.StructName }}Stream) Send{{ .Method.VarName }}Response(ctx context.Context, id any, result {{ .Result.Ref }}) error {
	{{- if and .Result (index .Result.Responses 0).ServerBody (index (index .Result.Responses 0).ServerBody 0).Init }}
	body := {{ (index (index .Result.Responses 0).ServerBody 0).Init.Name }}(result)
	{{- else }}
	body := result
	{{- end }}
	return s.conn.WriteJSON(jsonrpc.MakeSuccessResponse(id, body))
}
	{{- end }}
{{- end }}


{{ printf "SendError streams JSON-RPC errors." | comment }}
func (s *{{ lowerInitial $.Service.StructName }}Stream) SendError(ctx context.Context, id any, err error) error {
	{{- if allErrors . }}
	var en goa.GoaErrorNamer
	if !errors.As(err, &en) {
		code := jsonrpc.InternalError
		if _, ok := err.(*goa.ServiceError); ok {
			code = jsonrpc.InvalidParams
		}
		return s.sendError(ctx, id, code, err.Error(), nil)
	}
	switch en.GoaErrorName() {
	{{- range allErrors . }}
	case {{ printf "%q" .Name }}:
		{{- with .Response}}
		return s.sendError(ctx, id, {{ .Code }}, err.Error(), err)
		{{- end }}
	{{- end }}
	default:
		code := jsonrpc.InternalError
		if _, ok := err.(*goa.ServiceError); ok {
			code = jsonrpc.InvalidParams
		}
		return s.sendError(ctx, id, code, err.Error(), nil)
	}
	{{- else }}
	// No custom errors defined - check if it's a validation error, otherwise use internal error
	code := jsonrpc.InternalError
	if _, ok := err.(*goa.ServiceError); ok {
		code = jsonrpc.InvalidParams
	}
	return s.sendError(ctx, id, code, err.Error(), nil)
	{{- end }}
}

{{ printf "send writes a JSON-RPC response to the websocket connection." | comment }}
func (s *{{ lowerInitial $.Service.StructName }}Stream) send(id any, method string, result any) error {
	// If there's no ID, send as a notification instead of a response
	// A JSON-RPC result with no ID is invalid per the spec
	if id == nil || id == "" {
		return s.conn.WriteJSON(jsonrpc.MakeNotification(method, result))
	}
	return s.conn.WriteJSON(jsonrpc.MakeSuccessResponse(id, result))
}

{{ printf "sendError sends a JSON-RPC error response to the websocket connection." | comment }}
func (s *{{ lowerInitial $.Service.StructName }}Stream) sendError(ctx context.Context, id any, code jsonrpc.Code, message string, data any) error {
	response := jsonrpc.MakeErrorResponse(id, code, message, data)
	return s.conn.WriteJSON(response)
}
