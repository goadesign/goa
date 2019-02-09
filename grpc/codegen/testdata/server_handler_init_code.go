package testdata

const (
	UnaryRPCsServerHandlerInitCode = `// NewMethodUnaryRPCAHandler creates a gRPC handler which serves the
// "ServiceUnaryRPCs" service "MethodUnaryRPCA" endpoint.
func NewMethodUnaryRPCAHandler(endpoint goa.Endpoint, h goagrpc.UnaryHandler) goagrpc.UnaryHandler {
	if h == nil {
		h = goagrpc.NewUnaryHandler(endpoint, DecodeMethodUnaryRPCARequest, EncodeMethodUnaryRPCAResponse)
	}
	return h
}

// NewMethodUnaryRPCBHandler creates a gRPC handler which serves the
// "ServiceUnaryRPCs" service "MethodUnaryRPCB" endpoint.
func NewMethodUnaryRPCBHandler(endpoint goa.Endpoint, h goagrpc.UnaryHandler) goagrpc.UnaryHandler {
	if h == nil {
		h = goagrpc.NewUnaryHandler(endpoint, DecodeMethodUnaryRPCBRequest, EncodeMethodUnaryRPCBResponse)
	}
	return h
}
`

	UnaryRPCNoPayloadServerHandlerInitCode = `// NewMethodUnaryRPCNoPayloadHandler creates a gRPC handler which serves the
// "ServiceUnaryRPCNoPayload" service "MethodUnaryRPCNoPayload" endpoint.
func NewMethodUnaryRPCNoPayloadHandler(endpoint goa.Endpoint, h goagrpc.UnaryHandler) goagrpc.UnaryHandler {
	if h == nil {
		h = goagrpc.NewUnaryHandler(endpoint, nil, EncodeMethodUnaryRPCNoPayloadResponse)
	}
	return h
}
`

	UnaryRPCNoResultServerHandlerInitCode = `// NewMethodUnaryRPCNoResultHandler creates a gRPC handler which serves the
// "ServiceUnaryRPCNoResult" service "MethodUnaryRPCNoResult" endpoint.
func NewMethodUnaryRPCNoResultHandler(endpoint goa.Endpoint, h goagrpc.UnaryHandler) goagrpc.UnaryHandler {
	if h == nil {
		h = goagrpc.NewUnaryHandler(endpoint, DecodeMethodUnaryRPCNoResultRequest, EncodeMethodUnaryRPCNoResultResponse)
	}
	return h
}
`

	ServerStreamingRPCServerHandlerInitCode = `// NewMethodServerStreamingRPCHandler creates a gRPC handler which serves the
// "ServiceServerStreamingRPC" service "MethodServerStreamingRPC" endpoint.
func NewMethodServerStreamingRPCHandler(endpoint goa.Endpoint, h goagrpc.StreamHandler) goagrpc.StreamHandler {
	if h == nil {
		h = goagrpc.NewStreamHandler(endpoint, DecodeMethodServerStreamingRPCRequest)
	}
	return h
}
`

	ClientStreamingRPCServerHandlerInitCode = `// NewMethodClientStreamingRPCHandler creates a gRPC handler which serves the
// "ServiceClientStreamingRPC" service "MethodClientStreamingRPC" endpoint.
func NewMethodClientStreamingRPCHandler(endpoint goa.Endpoint, h goagrpc.StreamHandler) goagrpc.StreamHandler {
	if h == nil {
		h = goagrpc.NewStreamHandler(endpoint, nil)
	}
	return h
}
`

	ClientStreamingRPCWithPayloadServerHandlerInitCode = `// NewMethodClientStreamingRPCWithPayloadHandler creates a gRPC handler which
// serves the "ServiceClientStreamingRPCWithPayload" service
// "MethodClientStreamingRPCWithPayload" endpoint.
func NewMethodClientStreamingRPCWithPayloadHandler(endpoint goa.Endpoint, h goagrpc.StreamHandler) goagrpc.StreamHandler {
	if h == nil {
		h = goagrpc.NewStreamHandler(endpoint, DecodeMethodClientStreamingRPCWithPayloadRequest)
	}
	return h
}
`

	BidirectionalStreamingRPCServerHandlerInitCode = `// NewMethodBidirectionalStreamingRPCHandler creates a gRPC handler which
// serves the "ServiceBidirectionalStreamingRPC" service
// "MethodBidirectionalStreamingRPC" endpoint.
func NewMethodBidirectionalStreamingRPCHandler(endpoint goa.Endpoint, h goagrpc.StreamHandler) goagrpc.StreamHandler {
	if h == nil {
		h = goagrpc.NewStreamHandler(endpoint, nil)
	}
	return h
}
`

	BidirectionalStreamingRPCWithPayloadServerHandlerInitCode = `// NewMethodBidirectionalStreamingRPCWithPayloadHandler creates a gRPC handler
// which serves the "ServiceBidirectionalStreamingRPCWithPayload" service
// "MethodBidirectionalStreamingRPCWithPayload" endpoint.
func NewMethodBidirectionalStreamingRPCWithPayloadHandler(endpoint goa.Endpoint, h goagrpc.StreamHandler) goagrpc.StreamHandler {
	if h == nil {
		h = goagrpc.NewStreamHandler(endpoint, DecodeMethodBidirectionalStreamingRPCWithPayloadRequest)
	}
	return h
}
`
)
