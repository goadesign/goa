
	// Wrap the endpoints with the transport specific layers. The generated
	// server packages contains code generated from the design which maps
	// the service input and output data structures to HTTP requests and
	// responses.
	var (
	{{- range .Services }}
		{{ .Service.VarName }}Server *{{.Service.PkgName}}svr.Server
	{{- end }}
	)
	{
		eh := errorHandler(logger)
	{{- if needStream .Services }}
		upgrader := &websocket.Upgrader{}
	{{- end }}
	{{- range $svc := .Services }}
		{{-  if .Endpoints }}
		{{ .Service.VarName }}Server = {{ .Service.PkgName }}svr.New({{ .Service.VarName }}Endpoints, mux, dec, enc, eh, nil{{ if hasWebSocket $svc }}, upgrader, nil{{ end }}{{ range .Endpoints }}{{ if .MultipartRequestDecoder }}, {{ $.APIPkg }}.{{ .MultipartRequestDecoder.FuncName }}{{ end }}{{ end }}{{ range .FileServers }}, nil{{ end }})
		{{-  else }}
		{{ .Service.VarName }}Server = {{ .Service.PkgName }}svr.New(nil, mux, dec, enc, eh, nil{{ range .FileServers }}, nil{{ end }})
		{{-  end }}
	{{- end }}
	{{- if .Services }}
		if debug {
			servers := goahttp.Servers{
				{{- range $svc := .Services }}
				{{ .Service.VarName }}Server,
				{{- end }}
			}
			servers.Use(httpmdlwr.Debug(mux, os.Stdout))
		}
	{{- end }}
	}
	// Configure the mux.
	{{- range .Services }}
		{{ .Service.PkgName }}svr.Mount(mux, {{ .Service.VarName }}Server)
	{{- end }}
