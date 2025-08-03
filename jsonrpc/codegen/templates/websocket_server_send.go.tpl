{{- range .Endpoints }}
	{{- if .Result.Ref }}
		{{- if .Payload.Ref }}
{{ printf "Send%s sends a JSON-RPC response for the %s method." .Method.VarName .Method.Name | comment }}
func (s *{{ lowerInitial $.Service.StructName }}Stream) Send{{ .Method.VarName }}(ctx context.Context, result {{ .Result.Ref }}) error {
			{{- if .Result.IDAttribute }}
				{{- if .Result.IDAttributeRequired }}
	id := result.{{ .Result.IDAttribute }}
	result.{{ .Result.IDAttribute }} = ""
				{{- else }}
	var id any
	if result.{{ .Result.IDAttribute }} != nil {
		id = *result.{{ .Result.IDAttribute }}
		result.{{ .Result.IDAttribute }} = nil
	} else {
		id = ""
	}
				{{- end }}
	return s.send(id, result)
			{{- else }}
	return s.send("", result)
			{{- end }}
}
		{{- else }}
{{ printf "Send%s sends a JSON-RPC notification for the %s method." .Method.VarName .Method.Name | comment }}
func (s *{{ lowerInitial $.Service.StructName }}Stream) Send{{ .Method.VarName }}(ctx context.Context, params {{ .Result.Ref }}) error {
	return s.conn.WriteJSON(jsonrpc.MakeNotification({{ printf "%q" .Method.Name }}, params))
}
		{{- end }}
	{{- end }}
{{- end }}

{{- $hasResults := false }}
{{- range .Endpoints }}
	{{- if .Result.Ref }}
		{{- $hasResults = true }}
	{{- end }}
{{- end }}

{{- if $hasResults }}
{{ printf "Send sends an event to the client." | comment }}
func (s *{{ lowerInitial $.Service.StructName }}Stream) Send(event {{ $.Service.PkgName }}.Event) error {
	switch v := event.(type) {
{{- range .Endpoints }}
	{{- if .Result.Ref }}
	case {{ .Result.Ref }}:
		return s.Send{{ .Method.VarName }}(context.Background(), v)
	{{- end }}
{{- end }}
	default:
		return fmt.Errorf("unknown event type: %T", event)
	}
}
{{- end }}

{{ printf "SendError streams JSON-RPC errors." | comment }}
func (s *{{ lowerInitial $.Service.StructName }}Stream) SendError(ctx context.Context, id string, err error) error {
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
