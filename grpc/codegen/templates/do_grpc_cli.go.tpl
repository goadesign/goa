func doGRPC(_, host string, _ int, _ bool) (goa.Endpoint, any, error) {
	conn, err := grpc.NewClient(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
    fmt.Fprintf(os.Stderr, "could not connect to gRPC server at %s: %v\n", host, err)
  }
{{- range .Services }}
	{{- if .Service.ClientInterceptors }}
		{{ .Service.VarName }}Interceptors := {{ $.InterceptorsPkg }}.New{{ .Service.StructName }}ClientInterceptors()
	{{- end }}
{{- end }}
	return cli.ParseEndpoint(
		conn,
{{- range .Services }}
	{{- if .Service.ClientInterceptors }}
		{{ .Service.VarName }}Interceptors,
	{{- end }}
{{- end }}
	)
}

{{ if eq .DefaultTransport.Type "grpc" }}
func grpcUsageCommands() string {
	return cli.UsageCommands()
}

func grpcUsageExamples() string {
	return cli.UsageExamples()
}
{{- end }}