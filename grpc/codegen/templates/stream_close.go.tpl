
func (s *{{ .VarName }}) Close() error {
{{- if eq .Type "client" }}
{{- if .Endpoint.Method.Result }}
	{{ comment "Close the send direction of the stream" }}
	return s.stream.CloseSend()
{{- else }}
	{{ comment "synchronize and report any server error" }}
	_, err := s.stream.CloseAndRecv()
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
