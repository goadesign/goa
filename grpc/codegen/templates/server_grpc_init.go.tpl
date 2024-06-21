
	// Wrap the endpoints with the transport specific layers. The generated
	// server packages contains code generated from the design which maps
	// the service input and output data structures to gRPC requests and
	// responses.
	var (
	{{- range .Services }}
		{{ .Service.VarName }}Server *{{.Service.PkgName}}svr.Server
	{{- end }}
	)
	{
	{{- range .Services }}
		{{- if .Endpoints }}
		{{ .Service.VarName }}Server = {{ .Service.PkgName }}svr.New({{ .Service.VarName }}Endpoints{{ if .HasUnaryEndpoint }}, nil{{ end }}{{ if .HasStreamingEndpoint }}, nil{{ end }})
		{{-  else }}
		{{ .Service.VarName }}Server = {{ .Service.PkgName }}svr.New(nil{{ if .HasUnaryEndpoint }}, nil{{ end }}{{ if .HasStreamingEndpoint }}, nil{{ end }})
		{{-  end }}
	{{- end }}
	}
