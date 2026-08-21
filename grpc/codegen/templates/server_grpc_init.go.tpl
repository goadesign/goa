
	// Wrap the endpoints with the transport specific layers. The generated
	// server packages contains code generated from the design which maps
	// the service input and output data structures to gRPC requests and
	// responses.
	var (
	{{- range .Services }}
		{{ .Service.VarName }}Server *{{ .ServerPkgName }}.Server
	{{- end }}
	)
	{
	{{- range .Services }}
		{{- if .Endpoints }}
		{{ .Service.VarName }}Server = {{ .ServerPkgName }}.New({{ .Service.VarName }}Endpoints{{ if .HasUnaryEndpoint }}, nil{{ end }}{{ if .HasStreamingEndpoint }}, nil{{ end }})
		{{-  else }}
		{{ .Service.VarName }}Server = {{ .ServerPkgName }}.New(nil{{ if .HasUnaryEndpoint }}, nil{{ end }}{{ if .HasStreamingEndpoint }}, nil{{ end }})
		{{-  end }}
	{{- end }}
	}
