
	// Wrap the endpoints with the transport specific layers. The generated
	// server packages contains code generated from the design which maps
	// the service input and output data structures to HTTP requests and
	// responses.
	var (
	{{- range .Services }}
		{{ .Service.VarName }}Server *{{ .ServerPkgName }}.{{ .ServerStructDeclaration.Name }}
	{{- end }}
	{{- range .JSONRPCServices }}
		{{ .Service.VarName }}JSONRPCServer *{{ .ServerPkgName }}.{{ .ServerStructDeclaration.Name }}
	{{- end }}
	)
	{
		eh := errorHandler(ctx)
	{{- if or (needDialer .Services) (needDialer .JSONRPCServices) }}
		upgrader := &websocket.Upgrader{}
	{{- end }}
	{{- range $svc := .Services }}
		{{-  if .Endpoints }}
		{{ .Service.VarName }}Server = {{ .ServerPkgName }}.{{ .ServerInitDeclaration.Name }}({{ .Service.VarName }}Endpoints, mux, dec, enc, eh, nil{{ if hasWebSocket $svc }}, upgrader, nil{{ end }}{{ range .Endpoints }}{{ if .MultipartRequestDecoder }}, {{ $.APIPkg }}.{{ .MultipartRequestDecoder.FuncDeclaration.Name }}{{ end }}{{ end }}{{ range .FileServers }}, nil{{ end }})
		{{-  else }}
		{{ .Service.VarName }}Server = {{ .ServerPkgName }}.{{ .ServerInitDeclaration.Name }}(nil, mux, dec, enc, eh, nil{{ range .FileServers }}, nil{{ end }})
		{{-  end }}
	{{- end }}
	{{- range $svcData := .JSONRPCServices }}
		{{-  if .Endpoints }}
		{{- $svc := . }}
		{{ .Service.VarName }}JSONRPCServer = {{ .ServerPkgName }}.{{ .ServerInitDeclaration.Name }}({{ if hasWebSocket $svc }}{{ .Service.VarName }}Svc.HandleStream, {{ end }}{{ .Service.VarName }}Endpoints, mux, dec, enc, eh{{ if hasWebSocket $svc }}, upgrader, nil{{ end }})
		{{-  end }}
	{{- end }}
	}

	// Configure the mux.
	{{- range .Services }}
		{{ .ServerPkgName }}.{{ .MountServerDeclaration.Name }}(mux, {{ .Service.VarName }}Server)
	{{- end }}
	{{- range .JSONRPCServices }}
		{{ .ServerPkgName }}.{{ .MountServerDeclaration.Name }}(mux, {{ .Service.VarName }}JSONRPCServer)
	{{- end }}
