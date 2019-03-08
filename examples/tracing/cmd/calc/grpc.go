package main

import (
	"context"
	"log"
	"net"
	"net/url"
	"sync"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	calcsvc "goa.design/goa/examples/basic/gen/calc"
	calcpb "goa.design/goa/examples/basic/gen/grpc/calc/pb"
	calcsvcsvr "goa.design/goa/examples/basic/gen/grpc/calc/server"
	grpcmdlwr "goa.design/goa/grpc/middleware"
	"goa.design/goa/grpc/middleware/xray"
	"goa.design/goa/middleware"
	"google.golang.org/grpc"
)

// handleGRPCServer starts configures and starts a gRPC server on the given
// URL. It shuts down the server if any error is received in the error channel.
func handleGRPCServer(ctx context.Context, u *url.URL, calcEndpoints *calcsvc.Endpoints, wg *sync.WaitGroup, errc chan error, logger *log.Logger, debug bool, daemon string) {

	// Setup goa log adapter.
	var (
		adapter middleware.Logger
	)
	{
		adapter = middleware.NewLogger(logger)
	}

	// Wrap the endpoints with the transport specific layers. The generated
	// server packages contains code generated from the design which maps
	// the service input and output data structures to gRPC requests and
	// responses.
	var (
		calcServer *calcsvcsvr.Server
	)
	{
		calcServer = calcsvcsvr.New(calcEndpoints, nil)
	}

	xm, err := xray.NewUnaryServer("calc", daemon)
	if err != nil {
		logger.Printf("[WARN] cannot connect to xray daemon %s: %s", daemon, err)
	}
	// Initialize gRPC server with the middleware.
	srv := grpc.NewServer(
		grpcmiddleware.WithUnaryServerChain(
			grpcmdlwr.UnaryRequestID(),
			grpcmdlwr.UnaryServerLog(adapter),
			// Mount the trace and X-Ray middleware. Order is very important.
			grpcmdlwr.UnaryServerTrace(),
			xm,
		),
	)

	// Register the servers.
	calcpb.RegisterCalcServer(srv, calcServer)

	for svc, info := range srv.GetServiceInfo() {
		for _, m := range info.Methods {
			logger.Printf("serving gRPC method %s", svc+"/"+m.Name)
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
			logger.Printf("gRPC server listening on %q", u.Host)
			errc <- srv.Serve(lis)
		}()

		<-ctx.Done()
		logger.Printf("shutting down gRPC server at %q", u.Host)
		srv.Stop()
	}()
}
