
	(*wg).Add(1)
	go func() {
		defer (*wg).Done()

		{{ comment "Start gRPC server in a separate goroutine." }}
		go func() {
			lis, err := net.Listen("tcp", u.Host)
			if err != nil {
				errc <- err
			}
			logger.Printf("gRPC server listening on %q", u.Host)
			errc <- srv.Serve(lis)
		}()

		<-ctx.Done()
		logger.Printf("shutting down gRPC server at %q", u.Host)
		srv.Stop()
  }()
}
