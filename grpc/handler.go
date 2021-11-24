package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	goa "goa.design/goa/v3/pkg"
)

// Inspired from https://github.com/go-kit/kit/blob/1c17eccf28596f5a2c59314f7923ca66301b90e4/transport/grpc/server.go

type (
	// UnaryHandler handles a unary RPC. The request and response types are
	// protocol buffer message types.
	UnaryHandler interface {
		// Handle handles a unary RPC.
		//
		// It takes a protocol buffer message type and returns a
		// protocol buffer message type and any error when executing the
		// RPC.
		Handle(ctx context.Context, reqpb interface{}) (respb interface{}, err error)
	}

	// StreamHandler handles a streaming RPC. The stream may be client-side,
	// server-side, or bidirectional.
	StreamHandler interface {
		// Handle handles a streaming RPC.
		//
		// input contains the endpoint payload (if any) and generated
		// endpoint stream.
		Handle(ctx context.Context, input interface{}) (err error)
		// Decode decodes the protocol buffer message and metadata to
		// the service type. For client-side and bidirectional streams,
		// the message is nil.
		Decode(ctx context.Context, reqpb interface{}) (req interface{}, err error)
	}

	unaryHandler struct {
		endpoint goa.Endpoint
		decoder  RequestDecoder
		encoder  ResponseEncoder
	}

	streamHandler struct {
		endpoint goa.Endpoint
		decoder  RequestDecoder
	}
)

// NewUnaryHandler returns a handler to handle unary gRPC endpoints.
func NewUnaryHandler(e goa.Endpoint, dec RequestDecoder, enc ResponseEncoder) UnaryHandler {
	return &unaryHandler{
		endpoint: e,
		decoder:  dec,
		encoder:  enc,
	}
}

// NewStreamHandler returns a handler to handle streaming gRPC endpoints.
func NewStreamHandler(e goa.Endpoint, dec RequestDecoder) StreamHandler {
	return &streamHandler{
		endpoint: e,
		decoder:  dec,
	}
}

// Handle serves a gRPC request.
func (h *unaryHandler) Handle(ctx context.Context, reqpb interface{}) (interface{}, error) {
	var (
		req interface{}
		err error
	)
	{
		if h.decoder != nil {
			// Decode gRPC request message and incoming metadata
			md, _ := metadata.FromIncomingContext(ctx)
			if req, err = h.decoder(ctx, reqpb, md); err != nil {
				if _, ok := err.(*goa.ServiceError); ok {
					return nil, err
				}
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		}
	}

	var (
		resp interface{}
	)
	{
		// Invoke goa endpoint
		if resp, err = h.endpoint(ctx, req); err != nil {
			return nil, err
		}
	}

	var (
		respb interface{}

		hdr  = metadata.MD{}
		trlr = metadata.MD{}
	)
	{
		if h.encoder != nil {
			// Encode gRPC response
			if respb, err = h.encoder(ctx, resp, &hdr, &trlr); err != nil {
				if _, ok := err.(*goa.ServiceError); ok {
					return nil, err
				}
				return nil, status.Error(codes.Unknown, err.Error())
			}
		}
	}

	// Encode gRPC headers
	if len(hdr) > 0 {
		if err := grpc.SendHeader(ctx, hdr); err != nil {
			return nil, status.Error(codes.Unknown, err.Error())
		}
	}

	// Encode gRPC trailers
	if len(trlr) > 0 {
		if err := grpc.SetTrailer(ctx, trlr); err != nil {
			return nil, status.Error(codes.Unknown, err.Error())
		}
	}

	return respb, err
}

// Decode decodes the request message and incoming metadata into goa type.
func (h *streamHandler) Decode(ctx context.Context, reqpb interface{}) (interface{}, error) {
	var (
		req interface{}
		err error
	)
	{
		if h.decoder != nil {
			md, _ := metadata.FromIncomingContext(ctx)
			if req, err = h.decoder(ctx, reqpb, md); err != nil {
				if _, ok := err.(*goa.ServiceError); ok {
					return nil, err
				}
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		}
	}
	return req, nil
}

// Handle serves a gRPC request.
func (h *streamHandler) Handle(ctx context.Context, stream interface{}) error {
	_, err := h.endpoint(ctx, stream)
	return err
}
