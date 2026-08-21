
	// Wrap the endpoints with the transport specific layers. The generated
	// server packages contains code generated from the design which maps
	// the service input and output data structures to HTTP requests and
	// responses.
	var (
	{{- range .Services }}
		{{ .Service.VarName }}Server *{{ .ServerPkgName }}.Server
	{{- end }}
	{{- range .JSONRPCServices }}
		{{ .Service.VarName }}JSONRPCServer *{{ .ServerPkgName }}.Server
	{{- end }}
	)
	{
		eh := errorHandler(ctx)
	{{- if or (needDialer .Services) (needDialer .JSONRPCServices) }}
		upgrader := &websocket.Upgrader{}
	{{- end }}
	{{- range $svc := .Services }}
		{{-  if .Endpoints }}
		{{ .Service.VarName }}Server = {{ .ServerPkgName }}.New({{ .Service.VarName }}Endpoints, mux, dec, enc, eh, nil{{ if hasWebSocket $svc }}, upgrader, nil{{ end }}{{ range .Endpoints }}{{ if .MultipartRequestDecoder }}, {{ $.APIPkg }}.{{ .MultipartRequestDecoder.FuncName }}{{ end }}{{ end }}{{ range .FileServers }}, nil{{ end }})
		{{-  else }}
		{{ .Service.VarName }}Server = {{ .ServerPkgName }}.New(nil, mux, dec, enc, eh, nil{{ range .FileServers }}, nil{{ end }})
		{{-  end }}
	{{- end }}
	{{- range $svcData := .JSONRPCServices }}
		{{-  if .Endpoints }}
		{{- $svc := . }}
		{{ .Service.VarName }}JSONRPCServer = {{ .ServerPkgName }}.New({{ if hasWebSocket $svc }}{{ .Service.VarName }}Svc.HandleStream, {{ end }}{{ .Service.VarName }}Endpoints, mux, dec, enc, eh{{ if hasWebSocket $svc }}, upgrader, nil{{ end }})
		{{-  end }}
	{{- end }}
	}

	// Configure the mux.
	{{- range .Services }}
		{{ .ServerPkgName }}.Mount(mux, {{ .Service.VarName }}Server)
	{{- end }}
	{{- range .JSONRPCServices }}
		{{ .ServerPkgName }}.Mount(mux, {{ .Service.VarName }}JSONRPCServer)
	{{- end }}
