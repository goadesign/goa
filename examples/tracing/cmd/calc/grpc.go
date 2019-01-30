package main

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"sync"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	calcsvc "goa.design/goa/examples/calc/gen/calc"
	"goa.design/goa/examples/calc/gen/grpc/calc/pb"
	calcsvcsvr "goa.design/goa/examples/calc/gen/grpc/calc/server"
	goagrpcmiddleware "goa.design/goa/grpc/middleware"
	"goa.design/goa/grpc/middleware/xray"
	"goa.design/goa/middleware"
	"google.golang.org/grpc"
)

// handleGRPCServer starts configures and starts a gRPC server on the given
// URL. It shuts down the server if any error is received in the error channel.
func handleGRPCServer(ctx context.Context, u *url.URL, calcEndpoints *calcsvc.Endpoints, wg *sync.WaitGroup, errc chan error, logger middleware.Logger, debug bool, daemon string) {

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
		logger.Log("error", "cannot connect to xray daemon", "daemon", daemon, "err", err)
	}
	// Initialize gRPC server with the middleware.
	srv := grpc.NewServer(
		grpcmiddleware.WithUnaryServerChain(
			goagrpcmiddleware.UnaryRequestID(),
			goagrpcmiddleware.UnaryServerLog(logger),
			// Mount the trace and X-Ray middleware. Order is very important.
			goagrpcmiddleware.UnaryServerTrace(),
			xm,
		),
	)

	// Register the servers.
	pb.RegisterCalcServer(srv, calcServer)

	(*wg).Add(1)
	go func() {
		defer (*wg).Done()

		// Start gRPC server in a separate goroutine.
		go func() {
			lis, err := net.Listen("tcp", u.Host)
			if err != nil {
				errc <- err
			}
			logger.Log("msg", fmt.Sprintf("gRPC server listening on %q", u.Host))
			errc <- srv.Serve(lis)
		}()

		select {
		case <-ctx.Done():
			logger.Log("msg", fmt.Sprintf("shutting down gRPC server at %q", u.Host))
			srv.Stop()
			return
		}
	}()
}
