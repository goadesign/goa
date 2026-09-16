{{- if hasAnyInputStreams .Services }}
	switch flag.Arg(0) {
	{{- range .Services }}
		{{- if hasInputStreams . }}
	case {{ printf "%q" (kebab .Service.PathName) }}:
		switch flag.Arg(1) {
		{{- range .Endpoints }}
			{{- if streamsInput .Method }}
		case {{ printf "%q" (kebab .Method.Name) }}:
			return errors.New({{ printf "%q" (printf "example client does not support streamed input for service %q method %q" .ServiceName .Method.Name) }})
			{{- end }}
		{{- end }}
		}
		{{- end }}
	{{- end }}
	}
{{- end }}
{{- if hasRunnable .Services }}
	endpoint, payload, err := {{ .CLIPkg }}.{{ .Parser.ParseEndpoint.Name }}(
{{- else }}
	_, _, err := {{ .CLIPkg }}.{{ .Parser.ParseEndpoint.Name }}(
{{- end }}
		scheme,
		host,
		doer,
		goahttp.RequestEncoder,
		goahttp.ResponseDecoder,
		debug,
{{- if needDialer .Services }}
		dialer,
	{{- range $svc := .Services }}
		{{- if hasWebSocket $svc }}
		nil,
		{{- end }}
	{{- end }}
{{- end }}
{{- range .Services }}
	{{- range .Endpoints }}
		{{- if .MultipartRequestDecoder }}
		{{ $.APIPkg }}.{{ .MultipartRequestEncoder.FuncDeclaration.Name }},
		{{- end }}
	{{- end }}
{{- end }}
{{- range .Services }}
	{{- if .Service.ClientInterceptors }}
		{{ .Service.VarName }}Interceptors,
	{{- end }}
{{- end }}
	)
	if err != nil {
		return fmt.Errorf("parse endpoint: %w", err)
	}

{{ if hasRunnable .Services }}
	switch flag.Arg(0) {
	{{- range .Services }}
		{{- if hasRunnableService . }}
	case {{ printf "%q" (kebab .Service.PathName) }}:
		switch flag.Arg(1) {
		{{- range .Endpoints }}
			{{- if not (streamsInput .Method) }}
		case {{ printf "%q" (kebab .Method.Name) }}:
			{{- if and (streamsOutput .Method) (not .HasMixedResults) }}
			data, err := endpoint(ctx, payload)
			if err != nil {
				return err
			}
			stream := data.({{ .ServicePkgName }}.{{ .Method.ClientStream.Interface }})
			return writeStreamResults(ctx, stdout, stream.{{ .Method.ClientStream.RecvWithContextName }})
			{{- else }}
			return writeEndpointResult(ctx, stdout, endpoint, payload)
			{{- end }}
			{{- end }}
		{{- end }}
		}
		{{- end }}
	{{- end }}
	}
	{{- end }}
	panic({{ printf "%q" (printf "parsed %s command has no generated result writer" .Transport) }})
}
