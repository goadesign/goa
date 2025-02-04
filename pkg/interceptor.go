package goa

import "context"

type (
	// InterceptorInfo contains information about the request shared between
	// all interceptors in the service chain. It provides access to the service name,
	// method name, the type of call the interceptor is handling (unary, streaming send,
	// or streaming recv), and the request payload.
	InterceptorInfo struct {
		// Name of service handling request
		Service string
		// Name of method handling request
		Method string
		// CallType is the type of call the interceptor is handling
		CallType InterceptorCallType
		// ReturnContext is the context returned by the interceptor for streaming calls
		ReturnContext context.Context
		// Payload of request
		RawPayload any
	}

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
