package grpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type (
	// Invoker invokes a gRPC method. The request and response types
	// are goa types.
	Invoker interface {
		Invoke(ctx context.Context, req interface{}) (res interface{}, err error)
	}

	// RemoteFunc invokes a remote unary gRPC method.
	RemoteFunc func(ctx context.Context, reqpb interface{}, opts ...grpc.CallOption) (respb interface{}, err error)

	cliInvoker struct {
		encoder RequestEncoder
		decoder ResponseDecoder
		fn      RemoteFunc
	}

	// ClientError is an error returned by a gRPC service client.
	ClientError struct {
		// Name is a name for this class of errors.
		Name string
		// Message contains the specific error details.
		Message string
		// Service is the name of the service.
		Service string
		// Method is the name of the service method.
		Method string
		// Is the error temporary?
		Temporary bool
		// Is the error a timeout?
		Timeout bool
		// Is the error a server-side fault?
		Fault bool
	}
)

// NewInvoker returns an invoker to invoke gRPC methods.
func NewInvoker(fn RemoteFunc, enc RequestEncoder, dec ResponseDecoder) Invoker {
	return &cliInvoker{
		fn:      fn,
		encoder: enc,
		decoder: dec,
	}
}

// Invoke invokes the given remote gRPC client method.
func (d *cliInvoker) Invoke(ctx context.Context, req interface{}) (interface{}, error) {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}

	var (
		reqpb interface{}
		err   error
	)
	{
		if d.encoder != nil {
			// Encode gRPC request and outgoing metadata
			if reqpb, err = d.encoder(ctx, req, &md); err != nil {
				return nil, err
			}
		}
	}

	ctx = metadata.NewOutgoingContext(ctx, md)

	var (
		respb     interface{}
		hdr, trlr metadata.MD
	)
	{
		// Invoke remote method
		if respb, err = d.fn(ctx, reqpb, grpc.Header(&hdr), grpc.Trailer(&trlr)); err != nil {
			return nil, err
		}
	}

	var (
		res interface{}
	)
	{
		if d.decoder != nil {
			// Decode gRPC response and incoming header and trailer metadata
			if res, err = d.decoder(ctx, respb, hdr, trlr); err != nil {
				return nil, err
			}
		}
	}

	return res, nil
}

// ErrInvalidType is the error returned when the wrong type is given to a
// method function.
func ErrInvalidType(svc, m, expected string, actual interface{}) error {
	msg := fmt.Sprintf("invalid value expected %s, got %v", expected, actual)
	return &ClientError{Name: "invalid_type", Message: msg, Service: svc, Method: m}
}

// Error builds an error message.
func (c *ClientError) Error() string {
	return fmt.Sprintf("[%s %s]: %s", c.Service, c.Method, c.Message)
}
