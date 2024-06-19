func doGRPC(scheme, host string, timeout int, debug bool) (goa.Endpoint, any, error) {
	conn, err := grpc.Dial(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
    fmt.Fprintf(os.Stderr, "could not connect to gRPC server at %s: %v\n", host, err)
  }
	return cli.ParseEndpoint(conn)
}

{{ if eq .DefaultTransport.Type "grpc" }}
func grpcUsageCommands() string {
	return cli.UsageCommands()
}

func grpcUsageExamples() string {
	return cli.UsageExamples()
}
{{- end }}