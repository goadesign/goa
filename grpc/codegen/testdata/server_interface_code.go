package testdata

const UnaryRPCsServerInterfaceCode = `// MethodUnaryRPCA implements the "MethodUnaryRPCA" method in
// pb.ServiceUnaryRPCsServer interface.
func (s *Server) MethodUnaryRPCA(ctx context.Context, message *pb.MethodUnaryRPCARequest) (*pb.MethodUnaryRPCAResponse, error) {
	resp, err := s.MethodUnaryRPCAH.Handle(ctx, message)
	if err != nil {
		return nil, goagrpc.EncodeError(err)
	}
	return resp.(*pb.MethodUnaryRPCAResponse), nil
}

// MethodUnaryRPCB implements the "MethodUnaryRPCB" method in
// pb.ServiceUnaryRPCsServer interface.
func (s *Server) MethodUnaryRPCB(ctx context.Context, message *pb.MethodUnaryRPCBRequest) (*pb.MethodUnaryRPCBResponse, error) {
	resp, err := s.MethodUnaryRPCBH.Handle(ctx, message)
	if err != nil {
		return nil, goagrpc.EncodeError(err)
	}
	return resp.(*pb.MethodUnaryRPCBResponse), nil
}
`

const UnaryRPCNoPayloadServerInterfaceCode = `// MethodUnaryRPCNoPayload implements the "MethodUnaryRPCNoPayload" method in
// pb.ServiceUnaryRPCNoPayloadServer interface.
func (s *Server) MethodUnaryRPCNoPayload(ctx context.Context, message *pb.MethodUnaryRPCNoPayloadRequest) (*pb.MethodUnaryRPCNoPayloadResponse, error) {
	resp, err := s.MethodUnaryRPCNoPayloadH.Handle(ctx, message)
	if err != nil {
		return nil, goagrpc.EncodeError(err)
	}
	return resp.(*pb.MethodUnaryRPCNoPayloadResponse), nil
}
`

const UnaryRPCNoResultServerInterfaceCode = `// MethodUnaryRPCNoResult implements the "MethodUnaryRPCNoResult" method in
// pb.ServiceUnaryRPCNoResultServer interface.
func (s *Server) MethodUnaryRPCNoResult(ctx context.Context, message *pb.MethodUnaryRPCNoResultRequest) (*pb.MethodUnaryRPCNoResultResponse, error) {
	resp, err := s.MethodUnaryRPCNoResultH.Handle(ctx, message)
	if err != nil {
		return nil, goagrpc.EncodeError(err)
	}
	return resp.(*pb.MethodUnaryRPCNoResultResponse), nil
}
`

const UnaryRPCWithErrorsServerInterfaceCode = `// MethodUnaryRPCWithErrors implements the "MethodUnaryRPCWithErrors" method in
// pb.ServiceUnaryRPCWithErrorsServer interface.
func (s *Server) MethodUnaryRPCWithErrors(ctx context.Context, message *pb.MethodUnaryRPCWithErrorsRequest) (*pb.MethodUnaryRPCWithErrorsResponse, error) {
	resp, err := s.MethodUnaryRPCWithErrorsH.Handle(ctx, message)
	if err != nil {
		if en, ok := err.(ErrorNamer); ok {
			switch en.ErrorName() {
			case "timeout":
				return nil, goagrpc.NewStatusError(codes.Canceled, err)
			case "internal":
				return nil, goagrpc.NewStatusError(codes.Unknown, err)
			case "bad_request":
				return nil, goagrpc.NewStatusError(codes.InvalidArgument, err)
			}
		}
		return nil, goagrpc.EncodeError(err)
	}
	return resp.(*pb.MethodUnaryRPCWithErrorsResponse), nil
}
`

const ServerStreamingRPCServerInterfaceCode = `// MethodServerStreamingRPC implements the "MethodServerStreamingRPC" method in
// pb.ServiceServerStreamingRPCServer interface.
func (s *Server) MethodServerStreamingRPC(message *pb.MethodServerStreamingRPCRequest, stream pb.ServiceServerStreamingRPC_MethodServerStreamingRPCServer) error {
	p, err := s.MethodServerStreamingRPCH.Decode(stream.Context(), message)
	if err != nil {
		return goagrpc.EncodeError(err)
	}
	ep := &serviceserverstreamingrpc.MethodServerStreamingRPCEndpointInput{
		Stream:  &MethodServerStreamingRPCServerStream{stream: stream},
		Payload: p.(int),
	}
	err = s.MethodServerStreamingRPCH.Handle(stream.Context(), ep)
	if err != nil {
		return goagrpc.EncodeError(err)
	}
	return nil
}
`

const ClientStreamingRPCServerInterfaceCode = `// MethodClientStreamingRPC implements the "MethodClientStreamingRPC" method in
// pb.ServiceClientStreamingRPCServer interface.
func (s *Server) MethodClientStreamingRPC(stream pb.ServiceClientStreamingRPC_MethodClientStreamingRPCServer) error {
	p, err := s.MethodClientStreamingRPCH.Decode(stream.Context(), nil)
	if err != nil {
		return goagrpc.EncodeError(err)
	}
	ep := &serviceclientstreamingrpc.MethodClientStreamingRPCEndpointInput{
		Stream: &MethodClientStreamingRPCServerStream{stream: stream},
	}
	err = s.MethodClientStreamingRPCH.Handle(stream.Context(), ep)
	if err != nil {
		return goagrpc.EncodeError(err)
	}
	return nil
}
`

const ClientStreamingRPCWithPayloadServerInterfaceCode = `// MethodClientStreamingRPCWithPayload implements the
// "MethodClientStreamingRPCWithPayload" method in
// pb.ServiceClientStreamingRPCWithPayloadServer interface.
func (s *Server) MethodClientStreamingRPCWithPayload(stream pb.ServiceClientStreamingRPCWithPayload_MethodClientStreamingRPCWithPayloadServer) error {
	p, err := s.MethodClientStreamingRPCWithPayloadH.Decode(stream.Context(), nil)
	if err != nil {
		return goagrpc.EncodeError(err)
	}
	ep := &serviceclientstreamingrpcwithpayload.MethodClientStreamingRPCWithPayloadEndpointInput{
		Stream:  &MethodClientStreamingRPCWithPayloadServerStream{stream: stream},
		Payload: p.(int),
	}
	err = s.MethodClientStreamingRPCWithPayloadH.Handle(stream.Context(), ep)
	if err != nil {
		return goagrpc.EncodeError(err)
	}
	return nil
}
`

const BidirectionalStreamingRPCServerInterfaceCode = `// MethodBidirectionalStreamingRPC implements the
// "MethodBidirectionalStreamingRPC" method in
// pb.ServiceBidirectionalStreamingRPCServer interface.
func (s *Server) MethodBidirectionalStreamingRPC(stream pb.ServiceBidirectionalStreamingRPC_MethodBidirectionalStreamingRPCServer) error {
	p, err := s.MethodBidirectionalStreamingRPCH.Decode(stream.Context(), nil)
	if err != nil {
		return goagrpc.EncodeError(err)
	}
	ep := &servicebidirectionalstreamingrpc.MethodBidirectionalStreamingRPCEndpointInput{
		Stream: &MethodBidirectionalStreamingRPCServerStream{stream: stream},
	}
	err = s.MethodBidirectionalStreamingRPCH.Handle(stream.Context(), ep)
	if err != nil {
		return goagrpc.EncodeError(err)
	}
	return nil
}
`

const BidirectionalStreamingRPCWithPayloadServerInterfaceCode = `// MethodBidirectionalStreamingRPCWithPayload implements the
// "MethodBidirectionalStreamingRPCWithPayload" method in
// pb.ServiceBidirectionalStreamingRPCWithPayloadServer interface.
func (s *Server) MethodBidirectionalStreamingRPCWithPayload(stream pb.ServiceBidirectionalStreamingRPCWithPayload_MethodBidirectionalStreamingRPCWithPayloadServer) error {
	p, err := s.MethodBidirectionalStreamingRPCWithPayloadH.Decode(stream.Context(), nil)
	if err != nil {
		return goagrpc.EncodeError(err)
	}
	ep := &servicebidirectionalstreamingrpcwithpayload.MethodBidirectionalStreamingRPCWithPayloadEndpointInput{
		Stream:  &MethodBidirectionalStreamingRPCWithPayloadServerStream{stream: stream},
		Payload: p.(*servicebidirectionalstreamingrpcwithpayload.Payload),
	}
	err = s.MethodBidirectionalStreamingRPCWithPayloadH.Handle(stream.Context(), ep)
	if err != nil {
		return goagrpc.EncodeError(err)
	}
	return nil
}
`

const BidirectionalStreamingRPCWithErrorsServerInterfaceCode = `// MethodBidirectionalStreamingRPCWithErrors implements the
// "MethodBidirectionalStreamingRPCWithErrors" method in
// pb.ServiceBidirectionalStreamingRPCWithErrorsServer interface.
func (s *Server) MethodBidirectionalStreamingRPCWithErrors(stream pb.ServiceBidirectionalStreamingRPCWithErrors_MethodBidirectionalStreamingRPCWithErrorsServer) error {
	p, err := s.MethodBidirectionalStreamingRPCWithErrorsH.Decode(stream.Context(), nil)
	if err != nil {
		if en, ok := err.(ErrorNamer); ok {
			switch en.ErrorName() {
			case "timeout":
				return goagrpc.NewStatusError(codes.Canceled, err)
			case "internal":
				return goagrpc.NewStatusError(codes.Unknown, err)
			case "bad_request":
				return goagrpc.NewStatusError(codes.InvalidArgument, err)
			}
		}
		return goagrpc.EncodeError(err)
	}
	ep := &servicebidirectionalstreamingrpcwitherrors.MethodBidirectionalStreamingRPCWithErrorsEndpointInput{
		Stream: &MethodBidirectionalStreamingRPCWithErrorsServerStream{stream: stream},
	}
	err = s.MethodBidirectionalStreamingRPCWithErrorsH.Handle(stream.Context(), ep)
	if err != nil {
		if en, ok := err.(ErrorNamer); ok {
			switch en.ErrorName() {
			case "timeout":
				return goagrpc.NewStatusError(codes.Canceled, err)
			case "internal":
				return goagrpc.NewStatusError(codes.Unknown, err)
			case "bad_request":
				return goagrpc.NewStatusError(codes.InvalidArgument, err)
			}
		}
		return goagrpc.EncodeError(err)
	}
	return nil
}
`
