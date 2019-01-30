package testdata

const NoServerServerHandleCode = `// handleGRPCServer starts configures and starts a gRPC server on the given
// URL. It shuts down the server if any error is received in the error channel.
func handleGRPCServer(ctx context.Context, u *url.URL, serviceEndpoints *service.Endpoints, wg *sync.WaitGroup, errc chan error, logger middleware.Logger, debug bool) {

	// Wrap the endpoints with the transport specific layers. The generated
	// server packages contains code generated from the design which maps
	// the service input and output data structures to gRPC requests and
	// responses.
	var (
		serviceServer *servicesvr.Server
	)
	{
		serviceServer = servicesvr.New(serviceEndpoints, nil)
	}

	// Initialize gRPC server with the middleware.
	srv := grpc.NewServer(
		grpcmiddleware.WithUnaryServerChain(
			goagrpcmiddleware.UnaryRequestID(),
			goagrpcmiddleware.UnaryServerLog(logger),
		),
	)

	// Register the servers.
	pb.RegisterServiceServer(srv, serviceServer)

	for svc, info := range srv.GetServiceInfo() {
		for _, m := range info.Methods {
			logger.Log("msg", "serving gRPC method", "method", svc+"/"+m.Name)
		}
	}

	(*wg).Add(1)
	go func() {
		defer (*wg).Done()

		// Start gRPC server in a separate goroutine.
		go func() {
			lis, err := net.Listen("tcp", u.Host)
			if err != nil {
				errc <- err
			}
			logger.Log("msg", "gRPC server listening", "host", u.Host)
			errc <- srv.Serve(lis)
		}()

		select {
		case <-ctx.Done():
			logger.Log("msg", "shutting down gRPC server", "host", u.Host)
			srv.Stop()
			return
		}
	}()
}
`

const ServerHostingServiceSubsetServerHandleCode = `// handleGRPCServer starts configures and starts a gRPC server on the given
// URL. It shuts down the server if any error is received in the error channel.
func handleGRPCServer(ctx context.Context, u *url.URL, serviceEndpoints *service.Endpoints, wg *sync.WaitGroup, errc chan error, logger middleware.Logger, debug bool) {

	// Wrap the endpoints with the transport specific layers. The generated
	// server packages contains code generated from the design which maps
	// the service input and output data structures to gRPC requests and
	// responses.
	var (
		serviceServer *servicesvr.Server
	)
	{
		serviceServer = servicesvr.New(serviceEndpoints, nil)
	}

	// Initialize gRPC server with the middleware.
	srv := grpc.NewServer(
		grpcmiddleware.WithUnaryServerChain(
			goagrpcmiddleware.UnaryRequestID(),
			goagrpcmiddleware.UnaryServerLog(logger),
		),
	)

	// Register the servers.
	pb.RegisterServiceServer(srv, serviceServer)

	for svc, info := range srv.GetServiceInfo() {
		for _, m := range info.Methods {
			logger.Log("msg", "serving gRPC method", "method", svc+"/"+m.Name)
		}
	}

	(*wg).Add(1)
	go func() {
		defer (*wg).Done()

		// Start gRPC server in a separate goroutine.
		go func() {
			lis, err := net.Listen("tcp", u.Host)
			if err != nil {
				errc <- err
			}
			logger.Log("msg", "gRPC server listening", "host", u.Host)
			errc <- srv.Serve(lis)
		}()

		select {
		case <-ctx.Done():
			logger.Log("msg", "shutting down gRPC server", "host", u.Host)
			srv.Stop()
			return
		}
	}()
}
`

const ServerHostingMultipleServicesServerHandleCode = `// handleGRPCServer starts configures and starts a gRPC server on the given
// URL. It shuts down the server if any error is received in the error channel.
func handleGRPCServer(ctx context.Context, u *url.URL, serviceEndpoints *service.Endpoints, anotherServiceEndpoints *anotherservice.Endpoints, wg *sync.WaitGroup, errc chan error, logger middleware.Logger, debug bool) {

	// Wrap the endpoints with the transport specific layers. The generated
	// server packages contains code generated from the design which maps
	// the service input and output data structures to gRPC requests and
	// responses.
	var (
		serviceServer        *servicesvr.Server
		anotherServiceServer *anotherservicesvr.Server
	)
	{
		serviceServer = servicesvr.New(serviceEndpoints, nil)
		anotherServiceServer = anotherservicesvr.New(anotherServiceEndpoints, nil)
	}

	// Initialize gRPC server with the middleware.
	srv := grpc.NewServer(
		grpcmiddleware.WithUnaryServerChain(
			goagrpcmiddleware.UnaryRequestID(),
			goagrpcmiddleware.UnaryServerLog(logger),
		),
	)

	// Register the servers.
	pb.RegisterServiceServer(srv, serviceServer)
	pb.RegisterAnotherServiceServer(srv, anotherServiceServer)

	for svc, info := range srv.GetServiceInfo() {
		for _, m := range info.Methods {
			logger.Log("msg", "serving gRPC method", "method", svc+"/"+m.Name)
		}
	}

	(*wg).Add(1)
	go func() {
		defer (*wg).Done()

		// Start gRPC server in a separate goroutine.
		go func() {
			lis, err := net.Listen("tcp", u.Host)
			if err != nil {
				errc <- err
			}
			logger.Log("msg", "gRPC server listening", "host", u.Host)
			errc <- srv.Serve(lis)
		}()

		select {
		case <-ctx.Done():
			logger.Log("msg", "shutting down gRPC server", "host", u.Host)
			srv.Stop()
			return
		}
	}()
}
`

const ExampleCLICode = `func doGRPC(scheme, host string, timeout int, debug bool) (goa.Endpoint, interface{}, error) {
	conn, err := grpc.Dial(host, grpc.WithInsecure())
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("could not connect to gRPC server at %s: %v", host, err))
	}
	return cli.ParseEndpoint(conn)
}

func grpcUsageCommands() string {
	return cli.UsageCommands()
}

func grpcUsageExamples() string {
	return cli.UsageExamples()
}
`
