package testdata

const UnaryRPCsServerInterfaceCode = `// MethodUnaryRPCA implements the "MethodUnaryRPCA" method in
// service_unary_rp_cspb.ServiceUnaryRPCsServer interface.
func (s *Server) MethodUnaryRPCA(ctx context.Context, message *service_unary_rp_cspb.MethodUnaryRPCARequest) (*service_unary_rp_cspb.MethodUnaryRPCAResponse, error) {
	ctx = context.WithValue(ctx, goa.MethodKey, "MethodUnaryRPCA")
	ctx = context.WithValue(ctx, goa.ServiceKey, "ServiceUnaryRPCs")
	resp, err := s.MethodUnaryRPCAH.Handle(ctx, message)
	if err != nil {
		return nil, goagrpc.EncodeError(err)
	}
	return resp.(*service_unary_rp_cspb.MethodUnaryRPCAResponse), nil
}

// MethodUnaryRPCB implements the "MethodUnaryRPCB" method in
// service_unary_rp_cspb.ServiceUnaryRPCsServer interface.
func (s *Server) MethodUnaryRPCB(ctx context.Context, message *service_unary_rp_cspb.MethodUnaryRPCBRequest) (*service_unary_rp_cspb.MethodUnaryRPCBResponse, error) {
	ctx = context.WithValue(ctx, goa.MethodKey, "MethodUnaryRPCB")
	ctx = context.WithValue(ctx, goa.ServiceKey, "ServiceUnaryRPCs")
	resp, err := s.MethodUnaryRPCBH.Handle(ctx, message)
	if err != nil {
		return nil, goagrpc.EncodeError(err)
	}
	return resp.(*service_unary_rp_cspb.MethodUnaryRPCBResponse), nil
}
`

const UnaryRPCNoPayloadServerInterfaceCode = `// MethodUnaryRPCNoPayload implements the "MethodUnaryRPCNoPayload" method in
// service_unary_rpc_no_payloadpb.ServiceUnaryRPCNoPayloadServer interface.
func (s *Server) MethodUnaryRPCNoPayload(ctx context.Context, message *service_unary_rpc_no_payloadpb.MethodUnaryRPCNoPayloadRequest) (*service_unary_rpc_no_payloadpb.MethodUnaryRPCNoPayloadResponse, error) {
	ctx = context.WithValue(ctx, goa.MethodKey, "MethodUnaryRPCNoPayload")
	ctx = context.WithValue(ctx, goa.ServiceKey, "ServiceUnaryRPCNoPayload")
	resp, err := s.MethodUnaryRPCNoPayloadH.Handle(ctx, message)
	if err != nil {
		return nil, goagrpc.EncodeError(err)
	}
	return resp.(*service_unary_rpc_no_payloadpb.MethodUnaryRPCNoPayloadResponse), nil
}
`

const UnaryRPCNoResultServerInterfaceCode = `// MethodUnaryRPCNoResult implements the "MethodUnaryRPCNoResult" method in
// service_unary_rpc_no_resultpb.ServiceUnaryRPCNoResultServer interface.
func (s *Server) MethodUnaryRPCNoResult(ctx context.Context, message *service_unary_rpc_no_resultpb.MethodUnaryRPCNoResultRequest) (*service_unary_rpc_no_resultpb.MethodUnaryRPCNoResultResponse, error) {
	ctx = context.WithValue(ctx, goa.MethodKey, "MethodUnaryRPCNoResult")
	ctx = context.WithValue(ctx, goa.ServiceKey, "ServiceUnaryRPCNoResult")
	resp, err := s.MethodUnaryRPCNoResultH.Handle(ctx, message)
	if err != nil {
		return nil, goagrpc.EncodeError(err)
	}
	return resp.(*service_unary_rpc_no_resultpb.MethodUnaryRPCNoResultResponse), nil
}
`

const UnaryRPCWithErrorsServerInterfaceCode = `// MethodUnaryRPCWithErrors implements the "MethodUnaryRPCWithErrors" method in
// service_unary_rpc_with_errorspb.ServiceUnaryRPCWithErrorsServer interface.
func (s *Server) MethodUnaryRPCWithErrors(ctx context.Context, message *service_unary_rpc_with_errorspb.MethodUnaryRPCWithErrorsRequest) (*service_unary_rpc_with_errorspb.MethodUnaryRPCWithErrorsResponse, error) {
	ctx = context.WithValue(ctx, goa.MethodKey, "MethodUnaryRPCWithErrors")
	ctx = context.WithValue(ctx, goa.ServiceKey, "ServiceUnaryRPCWithErrors")
	resp, err := s.MethodUnaryRPCWithErrorsH.Handle(ctx, message)
	if err != nil {
		var en ErrorNamer
		if errors.As(err, &en) {
			switch en.ErrorName() {
			case "timeout":
				return nil, goagrpc.NewStatusError(codes.Canceled, err, goagrpc.NewErrorResponse(err))
			case "internal":
				var er *serviceunaryrpcwitherrors.AnotherError
				errors.As(err, &er)
				return nil, goagrpc.NewStatusError(codes.Unknown, err, NewMethodUnaryRPCWithErrorsInternalError(er))
			case "bad_request":
				var er *serviceunaryrpcwitherrors.AnotherError
				errors.As(err, &er)
				return nil, goagrpc.NewStatusError(codes.InvalidArgument, err, NewMethodUnaryRPCWithErrorsBadRequestError(er))
			case "custom_error":
				var er *serviceunaryrpcwitherrors.ErrorType
				errors.As(err, &er)
				return nil, goagrpc.NewStatusError(codes.Unknown, err, NewMethodUnaryRPCWithErrorsCustomErrorError(er))
			}
		}
		return nil, goagrpc.EncodeError(err)
	}
	return resp.(*service_unary_rpc_with_errorspb.MethodUnaryRPCWithErrorsResponse), nil
}
`

const UnaryRPCWithOverridingErrorsServerInterfaceCode = `// MethodUnaryRPCWithOverridingErrors implements the
// "MethodUnaryRPCWithOverridingErrors" method in
// service_unary_rpc_with_overriding_errorspb.ServiceUnaryRPCWithOverridingErrorsServer
// interface.
func (s *Server) MethodUnaryRPCWithOverridingErrors(ctx context.Context, message *service_unary_rpc_with_overriding_errorspb.MethodUnaryRPCWithOverridingErrorsRequest) (*service_unary_rpc_with_overriding_errorspb.MethodUnaryRPCWithOverridingErrorsResponse, error) {
	ctx = context.WithValue(ctx, goa.MethodKey, "MethodUnaryRPCWithOverridingErrors")
	ctx = context.WithValue(ctx, goa.ServiceKey, "ServiceUnaryRPCWithOverridingErrors")
	resp, err := s.MethodUnaryRPCWithOverridingErrorsH.Handle(ctx, message)
	if err != nil {
		var en ErrorNamer
		if errors.As(err, &en) {
			switch en.ErrorName() {
			case "overridden":
				return nil, goagrpc.NewStatusError(codes.Unknown, err, goagrpc.NewErrorResponse(err))
			case "internal":
				return nil, goagrpc.NewStatusError(codes.Unknown, err, goagrpc.NewErrorResponse(err))
			}
		}
		return nil, goagrpc.EncodeError(err)
	}
	return resp.(*service_unary_rpc_with_overriding_errorspb.MethodUnaryRPCWithOverridingErrorsResponse), nil
}
`

const ServerStreamingRPCServerInterfaceCode = `// MethodServerStreamingRPC implements the "MethodServerStreamingRPC" method in
// service_server_streaming_rpcpb.ServiceServerStreamingRPCServer interface.
func (s *Server) MethodServerStreamingRPC(message *service_server_streaming_rpcpb.MethodServerStreamingRPCRequest, stream service_server_streaming_rpcpb.ServiceServerStreamingRPC_MethodServerStreamingRPCServer) error {
	ctx := stream.Context()
	ctx = context.WithValue(ctx, goa.MethodKey, "MethodServerStreamingRPC")
	ctx = context.WithValue(ctx, goa.ServiceKey, "ServiceServerStreamingRPC")
	p, err := s.MethodServerStreamingRPCH.Decode(ctx, message)
	if err != nil {
		return goagrpc.EncodeError(err)
	}
	ep := &serviceserverstreamingrpc.MethodServerStreamingRPCEndpointInput{
		Stream:  &MethodServerStreamingRPCServerStream{stream: stream},
		Payload: p.(int),
	}
	err = s.MethodServerStreamingRPCH.Handle(ctx, ep)
	if err != nil {
		return goagrpc.EncodeError(err)
	}
	return nil
}
`

const ClientStreamingRPCServerInterfaceCode = `// MethodClientStreamingRPC implements the "MethodClientStreamingRPC" method in
// service_client_streaming_rpcpb.ServiceClientStreamingRPCServer interface.
func (s *Server) MethodClientStreamingRPC(stream service_client_streaming_rpcpb.ServiceClientStreamingRPC_MethodClientStreamingRPCServer) error {
	ctx := stream.Context()
	ctx = context.WithValue(ctx, goa.MethodKey, "MethodClientStreamingRPC")
	ctx = context.WithValue(ctx, goa.ServiceKey, "ServiceClientStreamingRPC")
	_, err := s.MethodClientStreamingRPCH.Decode(ctx, nil)
	if err != nil {
		return goagrpc.EncodeError(err)
	}
	ep := &serviceclientstreamingrpc.MethodClientStreamingRPCEndpointInput{
		Stream: &MethodClientStreamingRPCServerStream{stream: stream},
	}
	err = s.MethodClientStreamingRPCH.Handle(ctx, ep)
	if err != nil {
		return goagrpc.EncodeError(err)
	}
	return nil
}
`

const ClientStreamingRPCWithPayloadServerInterfaceCode = `// MethodClientStreamingRPCWithPayload implements the
// "MethodClientStreamingRPCWithPayload" method in
// service_client_streaming_rpc_with_payloadpb.ServiceClientStreamingRPCWithPayloadServer
// interface.
func (s *Server) MethodClientStreamingRPCWithPayload(stream service_client_streaming_rpc_with_payloadpb.ServiceClientStreamingRPCWithPayload_MethodClientStreamingRPCWithPayloadServer) error {
	ctx := stream.Context()
	ctx = context.WithValue(ctx, goa.MethodKey, "MethodClientStreamingRPCWithPayload")
	ctx = context.WithValue(ctx, goa.ServiceKey, "ServiceClientStreamingRPCWithPayload")
	p, err := s.MethodClientStreamingRPCWithPayloadH.Decode(ctx, nil)
	if err != nil {
		return goagrpc.EncodeError(err)
	}
	ep := &serviceclientstreamingrpcwithpayload.MethodClientStreamingRPCWithPayloadEndpointInput{
		Stream:  &MethodClientStreamingRPCWithPayloadServerStream{stream: stream},
		Payload: p.(int),
	}
	err = s.MethodClientStreamingRPCWithPayloadH.Handle(ctx, ep)
	if err != nil {
		return goagrpc.EncodeError(err)
	}
	return nil
}
`

const BidirectionalStreamingRPCServerInterfaceCode = `// MethodBidirectionalStreamingRPC implements the
// "MethodBidirectionalStreamingRPC" method in
// service_bidirectional_streaming_rpcpb.ServiceBidirectionalStreamingRPCServer
// interface.
func (s *Server) MethodBidirectionalStreamingRPC(stream service_bidirectional_streaming_rpcpb.ServiceBidirectionalStreamingRPC_MethodBidirectionalStreamingRPCServer) error {
	ctx := stream.Context()
	ctx = context.WithValue(ctx, goa.MethodKey, "MethodBidirectionalStreamingRPC")
	ctx = context.WithValue(ctx, goa.ServiceKey, "ServiceBidirectionalStreamingRPC")
	_, err := s.MethodBidirectionalStreamingRPCH.Decode(ctx, nil)
	if err != nil {
		return goagrpc.EncodeError(err)
	}
	ep := &servicebidirectionalstreamingrpc.MethodBidirectionalStreamingRPCEndpointInput{
		Stream: &MethodBidirectionalStreamingRPCServerStream{stream: stream},
	}
	err = s.MethodBidirectionalStreamingRPCH.Handle(ctx, ep)
	if err != nil {
		return goagrpc.EncodeError(err)
	}
	return nil
}
`

const BidirectionalStreamingRPCWithPayloadServerInterfaceCode = `// MethodBidirectionalStreamingRPCWithPayload implements the
// "MethodBidirectionalStreamingRPCWithPayload" method in
// service_bidirectional_streaming_rpc_with_payloadpb.ServiceBidirectionalStreamingRPCWithPayloadServer
// interface.
func (s *Server) MethodBidirectionalStreamingRPCWithPayload(stream service_bidirectional_streaming_rpc_with_payloadpb.ServiceBidirectionalStreamingRPCWithPayload_MethodBidirectionalStreamingRPCWithPayloadServer) error {
	ctx := stream.Context()
	ctx = context.WithValue(ctx, goa.MethodKey, "MethodBidirectionalStreamingRPCWithPayload")
	ctx = context.WithValue(ctx, goa.ServiceKey, "ServiceBidirectionalStreamingRPCWithPayload")
	p, err := s.MethodBidirectionalStreamingRPCWithPayloadH.Decode(ctx, nil)
	if err != nil {
		return goagrpc.EncodeError(err)
	}
	ep := &servicebidirectionalstreamingrpcwithpayload.MethodBidirectionalStreamingRPCWithPayloadEndpointInput{
		Stream:  &MethodBidirectionalStreamingRPCWithPayloadServerStream{stream: stream},
		Payload: p.(*servicebidirectionalstreamingrpcwithpayload.Payload),
	}
	err = s.MethodBidirectionalStreamingRPCWithPayloadH.Handle(ctx, ep)
	if err != nil {
		return goagrpc.EncodeError(err)
	}
	return nil
}
`

const BidirectionalStreamingRPCWithErrorsServerInterfaceCode = `// MethodBidirectionalStreamingRPCWithErrors implements the
// "MethodBidirectionalStreamingRPCWithErrors" method in
// service_bidirectional_streaming_rpc_with_errorspb.ServiceBidirectionalStreamingRPCWithErrorsServer
// interface.
func (s *Server) MethodBidirectionalStreamingRPCWithErrors(stream service_bidirectional_streaming_rpc_with_errorspb.ServiceBidirectionalStreamingRPCWithErrors_MethodBidirectionalStreamingRPCWithErrorsServer) error {
	ctx := stream.Context()
	ctx = context.WithValue(ctx, goa.MethodKey, "MethodBidirectionalStreamingRPCWithErrors")
	ctx = context.WithValue(ctx, goa.ServiceKey, "ServiceBidirectionalStreamingRPCWithErrors")
	_, err := s.MethodBidirectionalStreamingRPCWithErrorsH.Decode(ctx, nil)
	if err != nil {
		var en ErrorNamer
		if errors.As(err, &en) {
			switch en.ErrorName() {
			case "timeout":
				return goagrpc.NewStatusError(codes.Canceled, err, goagrpc.NewErrorResponse(err))
			case "internal":
				return goagrpc.NewStatusError(codes.Unknown, err, goagrpc.NewErrorResponse(err))
			case "bad_request":
				return goagrpc.NewStatusError(codes.InvalidArgument, err, goagrpc.NewErrorResponse(err))
			}
		}
		return goagrpc.EncodeError(err)
	}
	ep := &servicebidirectionalstreamingrpcwitherrors.MethodBidirectionalStreamingRPCWithErrorsEndpointInput{
		Stream: &MethodBidirectionalStreamingRPCWithErrorsServerStream{stream: stream},
	}
	err = s.MethodBidirectionalStreamingRPCWithErrorsH.Handle(ctx, ep)
	if err != nil {
		var en ErrorNamer
		if errors.As(err, &en) {
			switch en.ErrorName() {
			case "timeout":
				return goagrpc.NewStatusError(codes.Canceled, err, goagrpc.NewErrorResponse(err))
			case "internal":
				return goagrpc.NewStatusError(codes.Unknown, err, goagrpc.NewErrorResponse(err))
			case "bad_request":
				return goagrpc.NewStatusError(codes.InvalidArgument, err, goagrpc.NewErrorResponse(err))
			}
		}
		return goagrpc.EncodeError(err)
	}
	return nil
}
`
