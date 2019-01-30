package main

import (
	"context"
	"net"
	"net/url"
	"sync"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	dividersvc "goa.design/goa/examples/error/gen/divider"
	"goa.design/goa/examples/error/gen/grpc/divider/pb"
	dividersvcsvr "goa.design/goa/examples/error/gen/grpc/divider/server"
	goagrpcmiddleware "goa.design/goa/grpc/middleware"
	"goa.design/goa/middleware"
	"google.golang.org/grpc"
)

// handleGRPCServer starts configures and starts a gRPC server on the given
// URL. It shuts down the server if any error is received in the error channel.
func handleGRPCServer(ctx context.Context, u *url.URL, dividerEndpoints *dividersvc.Endpoints, wg *sync.WaitGroup, errc chan error, logger middleware.Logger, debug bool) {

	// Wrap the endpoints with the transport specific layers. The generated
	// server packages contains code generated from the design which maps
	// the service input and output data structures to gRPC requests and
	// responses.
	var (
		dividerServer *dividersvcsvr.Server
	)
	{
		dividerServer = dividersvcsvr.New(dividerEndpoints, nil)
	}

	// Initialize gRPC server with the middleware.
	srv := grpc.NewServer(
		grpcmiddleware.WithUnaryServerChain(
			goagrpcmiddleware.UnaryRequestID(),
			goagrpcmiddleware.UnaryServerLog(logger),
		),
	)

	// Register the servers.
	pb.RegisterDividerServer(srv, dividerServer)

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
