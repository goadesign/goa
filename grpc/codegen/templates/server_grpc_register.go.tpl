
	// Create interceptor which sets up the logger in each request context.
	chain := grpc.ChainUnaryInterceptor(log.UnaryServerInterceptor(ctx))
	if dbg {
		// Log request and response content if debug logs are enabled.
		chain = grpc.ChainUnaryInterceptor(log.UnaryServerInterceptor(ctx), debug.UnaryServerInterceptor())
	}
	{{- if needStream .Services}}
	streamchain := grpc.ChainStreamInterceptor(log.StreamServerInterceptor(ctx))
	if dbg {
		streamchain = grpc.ChainStreamInterceptor(log.StreamServerInterceptor(ctx), debug.StreamServerInterceptor())
	}
	{{- end }}

	// Initialize gRPC server
	srv := grpc.NewServer(chain{{ if needStream .Services }}, streamchain{{ end }})

	// Register the servers.
	{{- range .Services }}
	{{ .PkgName }}.Register{{ goify .Service.VarName true }}Server(srv, {{ .Service.VarName }}Server)
	{{- end }}

	for svc, info := range srv.GetServiceInfo() {
		for _, m := range info.Methods {
			log.Printf(ctx, "serving gRPC method %s", svc + "/" + m.Name)
		}
	}

	// Register the server reflection service on the server.
	// See https://grpc.github.io/grpc/core/md_doc_server-reflection.html.
	reflection.Register(srv)
