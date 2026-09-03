func doGRPC(ctx context.Context, _ string, host string, _ int, _ bool, stdout io.Writer) (err error) {
{{- if hasAnyInputStreams .Services }}
	switch flag.Arg(0) {
	{{- range .Services }}
		{{- if hasInputStreams . }}
	case {{ printf "%q" (kebab .Service.Name) }}:
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
	conn, err := grpc.NewClient(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("connect to gRPC server at %s: %w", host, err)
	}
	defer func() {
		err = errors.Join(err, conn.Close())
	}()
{{- range .Services }}
	{{- if .Service.ClientInterceptors }}
		{{ .Service.VarName }}Interceptors := {{ $.InterceptorsPkg }}.New{{ .Service.StructName }}ClientInterceptors()
	{{- end }}
{{- end }}
{{- if hasRunnable .Services }}
	endpoint, payload, err := {{ .CLIPkg }}.{{ .Parser.ParseEndpoint.Name }}(
{{- else }}
	_, _, err = {{ .CLIPkg }}.{{ .Parser.ParseEndpoint.Name }}(
{{- end }}
		conn,
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
	case {{ printf "%q" (kebab .Service.Name) }}:
		switch flag.Arg(1) {
		{{- range .Endpoints }}
			{{- if not (streamsInput .Method) }}
		case {{ printf "%q" (kebab .Method.Name) }}:
			{{- if streamsOutput .Method }}
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
	panic("parsed gRPC command has no generated result writer")
}

{{ if eq .DefaultTransport.Type "grpc" }}
func grpcUsageExamples() string {
	return {{ .CLIPkg }}.{{ .Parser.UsageExamples.Name }}()
}
{{- end }}
