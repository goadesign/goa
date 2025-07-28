{{- range .Endpoints }}
	{{- if .Result.Ref }}
		{{- if .Payload.Ref }}
{{ printf "Send%s sends a JSON-RPC response for the %s method." .Method.VarName .Method.Name | comment }}
func (s *{{ lowerInitial $.Service.StructName }}Stream) Send{{ .Method.VarName }}(ctx context.Context, result {{ .Result.Ref }}) error {
			{{- if .Result.IDAttribute }}
	id := result.{{ .Result.IDAttribute }}
	result.{{ .Result.IDAttribute }} = ""
			{{- else }}
	id := ""
			{{- end }}
	return s.send(id, result)
}
		{{- else }}
{{ printf "Send%s sends a JSON-RPC notification for the %s method." .Method.VarName .Method.Name | comment }}
func (s *{{ lowerInitial $.Service.StructName }}Stream) Send{{ .Method.VarName }}(ctx context.Context, params {{ .Result.Ref }}) error {
	return s.conn.WriteJSON(jsonrpc.MakeNotification({{ printf "%q" .Method.Name }}, params))
}
		{{- end }}
	{{- end }}
{{- end }}

{{- if allErrors . }}
{{ printf "SendError streams JSON-RPC errors." | comment }}
func (s *{{ lowerInitial $.Service.StructName }}Stream) SendError(ctx context.Context, id string, err error) error {
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
{{- end }}

{{ printf "send writes a JSON-RPC response to the websocket connection." | comment }}
func (s *{{ lowerInitial $.Service.StructName }}Stream) send(id string, result any) error {
	return s.conn.WriteJSON(jsonrpc.MakeSuccessResponse(id, result))
}

{{ printf "sendError sends a JSON-RPC error response to the websocket connection." | comment }}
func (s *{{ lowerInitial $.Service.StructName }}Stream) sendError(ctx context.Context, id any, code jsonrpc.Code, message string, data any) error {
	response := jsonrpc.MakeErrorResponse(id, code, "", message)
	if data != nil {
		response.Error.Message = message
		response.Error.Data = data
	}
	return s.conn.WriteJSON(response)
}
