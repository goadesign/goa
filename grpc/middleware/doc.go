// Package middleware contains gRPC server and client interceptors that wraps
// unary and streaming RPCs to provide additional functionality.
//
// This package contains the following middlewares:
//
//   * Logging server middleware for unary and streaming endpoints.
//   * Request ID server middleware for unary and streaming endpoints.
//   * Stream Canceler server middleware for canceling streaming requests.
//   * Tracing middleware for unary and streaming server and client.
//   * AWS X-Ray middleware for producing X-Ray segments for unary and streaming
//     client and server.
//
// Example to use the server middleware:
//
//    srv := grpc.NewServer(middleware.UnaryRequestID())
//
// Example to use the client middleware:
//
//    conn, err := grpc.Dial(host,
//        grpc.WithUnaryInterceptor(middleware.UnaryClientTrace()),
//        grpc.WithStreamInterceptor(middleware.StreamClientTrace()),
//    )
package middleware
