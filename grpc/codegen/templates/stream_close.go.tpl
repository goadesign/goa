
func (s *{{ .Declaration.Name }}) Close() error {
{{- if eq .Type "client" }}
{{- if .Endpoint.Method.Result }}
	{{ comment "Close the send direction of the stream" }}
	return s.stream.CloseSend()
{{- else }}
	{{ comment "synchronize and report any server error" }}
	_, err := s.stream.CloseAndRecv()
	if err != nil {
		{{- if .Endpoint.Errors }}
		resp := goagrpc.DecodeError(err)
		switch message := resp.(type) {
		{{- range .Endpoint.Errors }}
			{{- if .Response.ClientConvert }}
		case {{ .Response.ClientConvert.SrcRef }}:
			{{- if .Response.ClientConvert.Validation }}
			if err := {{ .Response.ClientConvert.Validation.Declaration.Name }}(message); err != nil {
				return err
			}
			{{- end }}
			return {{ .Response.ClientConvert.Init.Declaration.Name }}({{ range .Response.ClientConvert.Init.Args }}{{ .Name }}, {{ end }})
			{{- end }}
		{{- end }}
		case *goapb.ErrorResponse:
			return goagrpc.NewServiceError(message)
		default:
			if ctxErr := goagrpc.ContextError(s.ctx, err); ctxErr != nil {
				return ctxErr
			}
			return err
		}
		{{- else }}
		if ctxErr := goagrpc.ContextError(s.ctx, err); ctxErr != nil {
			return ctxErr
		}
		return err
		{{- end }}
	}
	return err
{{- end }}
{{- else }}
{{- if .Endpoint.Method.Result }}
	{{ comment "nothing to do here" }}
	return nil
{{- else }}
	{{ comment "synchronize stream" }}
	return s.stream.SendAndClose(&{{ .Endpoint.Response.ServerConvert.TgtName }}{})
{{- end }}
{{- end }}
}
