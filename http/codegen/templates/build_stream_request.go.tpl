// {{ printf "%s creates a streaming endpoint request payload from the method payload and the path to the file to be streamed" .BuildStreamPayload | comment }}
func {{ .BuildStreamPayload }}({{ if .Payload.Ref }}payload any, {{ end }}fpath string) (*{{ requestStructPkg .Method .ServicePkgName }}.{{ .Method.RequestStruct }}, error) {
	f, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}
	return &{{ requestStructPkg .Method .ServicePkgName }}.{{ .Method.RequestStruct }}{
		{{- if .Payload.Ref }}
		Payload: payload.({{ .Payload.Ref }}),
		{{- end }}
		Body: f,
	}, nil
}
