package goa

type (
	// InterceptorInfo contains information about the request shared between
	// all interceptors in the service chain. It provides access to the service name,
	// method name, endpoint function, and request payload.
	InterceptorInfo struct {
		// Name of service handling request
		Service string
		// Name of method handling request
		Method string
		// Send is true if the request is a streaming Send
		Send bool
		// Recv is true if the request is a streaming Recv
		Recv bool
		// Payload of request
		RawPayload any
	}
)
