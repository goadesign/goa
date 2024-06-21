
	// Initialize gRPC server with the middleware.
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpcmdlwr.UnaryRequestID(),
			grpcmdlwr.UnaryServerLog(adapter),
		),
	{{- if needStream .Services }}
		grpc.ChainStreamInterceptor(
			grpcmdlwr.StreamRequestID(),
			grpcmdlwr.StreamServerLog(adapter),
		),
	{{- end }}
	)

	// Register the servers.
	{{- range .Services }}
	{{ .PkgName }}.Register{{ goify .Service.VarName true }}Server(srv, {{ .Service.VarName }}Server)
	{{- end }}

	for svc, info := range srv.GetServiceInfo() {
		for _, m := range info.Methods {
			logger.Printf("serving gRPC method %s", svc + "/" + m.Name)
		}
	}

	// Register the server reflection service on the server.
	// See https://grpc.github.io/grpc/core/md_doc_server-reflection.html.
	reflection.Register(srv)
