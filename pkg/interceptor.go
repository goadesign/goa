package goa

import "context"

type (
	// InterceptorInfo contains information about the request shared between
	// all interceptors in the service chain. It provides access to the service name,
	// method name, the type of call the interceptor is handling (unary, streaming send,
	// or streaming recv), and the request payload.
	InterceptorInfo interface {
		// Service returns the name of the service handling the request.
		Service() string
		// Method returns the name of the method handling the request.
		Method() string
		// CallType returns the type of call the interceptor is handling.
		CallType() InterceptorCallType
		// RawPayload returns the raw payload of the request.
		RawPayload() any
	}

	// InterceptorEndpoint is the endpoint function interface that interceptors call. It differs from the
	// Endpoint function interface in that it also returns a context that allows streaming interceptors to
	// return a context to the SendWithContext or RecvWithContext methods.
	InterceptorEndpoint func(ctx context.Context, request any) (response any, returnCtx context.Context, err error)

	// InterceptorCallType is the type of call the interceptor is handling
	InterceptorCallType int
)

const (
	// InterceptorUnary indicates the interceptor is handling a unary call
	InterceptorUnary InterceptorCallType = iota
	// InterceptorStreamingSend indicates the interceptor is handling a streaming Send
	InterceptorStreamingSend
	// InterceptorStreamingRecv indicates the interceptor is handling a streaming Recv
	InterceptorStreamingRecv
)
