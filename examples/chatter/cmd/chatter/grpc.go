package main

import (
	"context"
	"net"
	"net/url"
	"sync"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	chattersvc "goa.design/goa/examples/chatter/gen/chatter"
	"goa.design/goa/examples/chatter/gen/grpc/chatter/pb"
	chattersvcsvr "goa.design/goa/examples/chatter/gen/grpc/chatter/server"
	goagrpcmiddleware "goa.design/goa/grpc/middleware"
	"goa.design/goa/middleware"
	"google.golang.org/grpc"
)

// handleGRPCServer starts configures and starts a gRPC server on the given
// URL. It shuts down the server if any error is received in the error channel.
func handleGRPCServer(ctx context.Context, u *url.URL, chatterEndpoints *chattersvc.Endpoints, wg *sync.WaitGroup, errc chan error, logger middleware.Logger, debug bool) {

	// Wrap the endpoints with the transport specific layers. The generated
	// server packages contains code generated from the design which maps
	// the service input and output data structures to gRPC requests and
	// responses.
	var (
		chatterServer *chattersvcsvr.Server
	)
	{
		chatterServer = chattersvcsvr.New(chatterEndpoints, nil, nil)
	}

	// Initialize gRPC server with the middleware.
	srv := grpc.NewServer(
		grpcmiddleware.WithUnaryServerChain(
			goagrpcmiddleware.UnaryRequestID(),
			goagrpcmiddleware.UnaryServerLog(logger),
		),
		grpcmiddleware.WithStreamServerChain(
			goagrpcmiddleware.StreamRequestID(),
			goagrpcmiddleware.StreamServerLog(logger),
		),
	)

	// Register the servers.
	pb.RegisterChatterServer(srv, chatterServer)

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
