// This file contains expected gRPC stream code used by generator tests.
package testdata

var ServerStreamingServerStructCode = `// MethodServerStreamingUserTypeRPCServerStream implements the
// serviceserverstreamingusertyperpc.MethodServerStreamingUserTypeRPCServerStream
// interface.
type MethodServerStreamingUserTypeRPCServerStream struct {
	stream service_server_streaming_user_type_rpcpb.ServiceServerStreamingUserTypeRPC_MethodServerStreamingUserTypeRPCServer
}
`

var ServerStreamingServerSendCode = `// Send streams instances of
// "service_server_streaming_user_type_rpcpb.MethodServerStreamingUserTypeRPCResponse"
// to the "MethodServerStreamingUserTypeRPC" endpoint gRPC stream.
func (s *MethodServerStreamingUserTypeRPCServerStream) Send(res *serviceserverstreamingusertyperpc.UserType) error {
	v := NewProtoMethodServerStreamingUserTypeRPCResponse(res)
	return s.stream.Send(v)
}

// SendWithContext streams instances of
// "service_server_streaming_user_type_rpcpb.MethodServerStreamingUserTypeRPCResponse"
// to the "MethodServerStreamingUserTypeRPC" endpoint gRPC stream with context.
func (s *MethodServerStreamingUserTypeRPCServerStream) SendWithContext(ctx context.Context, res *serviceserverstreamingusertyperpc.UserType) error {
	return s.Send(res)
}
`

var ServerStreamingServerCloseCode = `func (s *MethodServerStreamingUserTypeRPCServerStream) Close() error {
	// nothing to do here
	return nil
}
`

var ServerStreamingClientStructCode = `// MethodServerStreamingUserTypeRPCClientStream implements the
// serviceserverstreamingusertyperpc.MethodServerStreamingUserTypeRPCClientStream
// interface.
type MethodServerStreamingUserTypeRPCClientStream struct {
	stream service_server_streaming_user_type_rpcpb.ServiceServerStreamingUserTypeRPC_MethodServerStreamingUserTypeRPCClient
}
`

var ServerStreamingClientRecvCode = `// Recv reads instances of
// "service_server_streaming_user_type_rpcpb.MethodServerStreamingUserTypeRPCResponse"
// from the "MethodServerStreamingUserTypeRPC" endpoint gRPC stream.
func (s *MethodServerStreamingUserTypeRPCClientStream) Recv() (*serviceserverstreamingusertyperpc.UserType, error) {
	var res *serviceserverstreamingusertyperpc.UserType
	v, err := s.stream.Recv()
	if err != nil {
		return res, err
	}
	return NewMethodServerStreamingUserTypeRPCResponseUserType(v), nil
}

// RecvWithContext reads instances of
// "service_server_streaming_user_type_rpcpb.MethodServerStreamingUserTypeRPCResponse"
// from the "MethodServerStreamingUserTypeRPC" endpoint gRPC stream with
// context.
func (s *MethodServerStreamingUserTypeRPCClientStream) RecvWithContext(ctx context.Context) (*serviceserverstreamingusertyperpc.UserType, error) {
	return s.Recv()
}
`

var ServerStreamingResultWithViewsServerStructCode = `// MethodServerStreamingUserTypeRPCServerStream implements the
// serviceserverstreamingusertyperpc.MethodServerStreamingUserTypeRPCServerStream
// interface.
type MethodServerStreamingUserTypeRPCServerStream struct {
	stream service_server_streaming_user_type_rpcpb.ServiceServerStreamingUserTypeRPC_MethodServerStreamingUserTypeRPCServer
	view   string
	// sentView is the result view named in the response header. Later sends must
	// use the same view.
	sentView string
}
`

var ServerStreamingResultWithViewsServerSendCode = `// Send streams instances of
// "service_server_streaming_user_type_rpcpb.MethodServerStreamingUserTypeRPCResponse"
// to the "MethodServerStreamingUserTypeRPC" endpoint gRPC stream.
func (s *MethodServerStreamingUserTypeRPCServerStream) Send(res *serviceserverstreamingusertyperpc.ResultType) error {
	view := s.view
	if view == "" {
		view = "default"
	}
	if s.sentView != "" && view != s.sentView {
		return goa.InvalidEnumValueError("view", view, []any{s.sentView})
	}
	vres := serviceserverstreamingusertyperpc.NewViewedResultType(res, view)
	var v *service_server_streaming_user_type_rpcpb.MethodServerStreamingUserTypeRPCResponse
	switch view {
	case "tiny":
		v = NewProtoMethodServerStreamingUserTypeRPCResponseTiny(vres.Projected)
	case "default", "":
		v = NewProtoMethodServerStreamingUserTypeRPCResponse(vres.Projected)
	default:
		return goa.InvalidEnumValueError("view", view, []any{"tiny", "default"})
	}
	if s.sentView == "" {
		if err := s.stream.SetHeader(metadata.Pairs("goa-view", view)); err != nil {
			return err
		}
		s.sentView = view
	}
	return s.stream.Send(v)
}

// SendWithContext streams instances of
// "service_server_streaming_user_type_rpcpb.MethodServerStreamingUserTypeRPCResponse"
// to the "MethodServerStreamingUserTypeRPC" endpoint gRPC stream with context.
func (s *MethodServerStreamingUserTypeRPCServerStream) SendWithContext(ctx context.Context, res *serviceserverstreamingusertyperpc.ResultType) error {
	return s.Send(res)
}
`

var ServerStreamingResultWithViewsServerSetViewCode = `// SetView sets the view.
func (s *MethodServerStreamingUserTypeRPCServerStream) SetView(view string) {
	s.view = view
}
`

var ServerStreamingResultWithViewsClientStructCode = `// MethodServerStreamingUserTypeRPCClientStream implements the
// serviceserverstreamingusertyperpc.MethodServerStreamingUserTypeRPCClientStream
// interface.
type MethodServerStreamingUserTypeRPCClientStream struct {
	stream  service_server_streaming_user_type_rpcpb.ServiceServerStreamingUserTypeRPC_MethodServerStreamingUserTypeRPCClient
	view    string
	viewSet bool
}
`

var ServerStreamingResultWithViewsClientRecvCode = `// Recv reads instances of
// "service_server_streaming_user_type_rpcpb.MethodServerStreamingUserTypeRPCResponse"
// from the "MethodServerStreamingUserTypeRPC" endpoint gRPC stream.
func (s *MethodServerStreamingUserTypeRPCClientStream) Recv() (*serviceserverstreamingusertyperpc.ResultType, error) {
	var res *serviceserverstreamingusertyperpc.ResultType
	v, err := s.stream.Recv()
	if err != nil {
		return res, err
	}
	if !s.viewSet {
		hdr, err := s.stream.Header()
		if err != nil {
			return res, err
		}
		views := hdr.Get("goa-view")
		if len(views) == 0 {
			return res, goa.MissingFieldError("goa-view", "metadata")
		}
		s.view = views[0]
		s.viewSet = true
	}
	var proj *serviceserverstreamingusertyperpcviews.ResultTypeView
	switch s.view {
	case "tiny":
		if err := ValidateMethodServerStreamingUserTypeRPCResponseTiny(v); err != nil {
			return res, err
		}
		proj = NewMethodServerStreamingUserTypeRPCResponseResultTypeViewTiny(v)
	case "default", "":
		if err := ValidateMethodServerStreamingUserTypeRPCResponse(v); err != nil {
			return res, err
		}
		proj = NewMethodServerStreamingUserTypeRPCResponseResultTypeView(v)
	}
	vres := &serviceserverstreamingusertyperpcviews.ResultType{Projected: proj, View: s.view}
	if err := serviceserverstreamingusertyperpcviews.ValidateResultType(vres); err != nil {
		return nil, err
	}
	return serviceserverstreamingusertyperpc.NewResultType(vres), nil
}

// RecvWithContext reads instances of
// "service_server_streaming_user_type_rpcpb.MethodServerStreamingUserTypeRPCResponse"
// from the "MethodServerStreamingUserTypeRPC" endpoint gRPC stream with
// context.
func (s *MethodServerStreamingUserTypeRPCClientStream) RecvWithContext(ctx context.Context) (*serviceserverstreamingusertyperpc.ResultType, error) {
	return s.Recv()
}
`

var ServerStreamingResultWithViewsClientSetViewCode = `// SetView sets the view.
func (s *MethodServerStreamingUserTypeRPCClientStream) SetView(view string) {
	s.view = view
	s.viewSet = true
}
`

var ServerStreamingResultCollectionWithExplicitViewServerSendCode = `// Send streams instances of
// "service_server_streaming_result_type_collection_with_explicit_viewpb.ResultTypeCollection"
// to the "MethodServerStreamingResultTypeCollectionWithExplicitView" endpoint
// gRPC stream.
func (s *MethodServerStreamingResultTypeCollectionWithExplicitViewServerStream) Send(res serviceserverstreamingresulttypecollectionwithexplicitview.ResultTypeCollection) error {
	vres := serviceserverstreamingresulttypecollectionwithexplicitview.NewViewedResultTypeCollection(res, "tiny")
	v := NewProtoResultTypeCollection(vres.Projected)
	return s.stream.Send(v)
}

// SendWithContext streams instances of
// "service_server_streaming_result_type_collection_with_explicit_viewpb.ResultTypeCollection"
// to the "MethodServerStreamingResultTypeCollectionWithExplicitView" endpoint
// gRPC stream with context.
func (s *MethodServerStreamingResultTypeCollectionWithExplicitViewServerStream) SendWithContext(ctx context.Context, res serviceserverstreamingresulttypecollectionwithexplicitview.ResultTypeCollection) error {
	return s.Send(res)
}
`

var ServerStreamingResultCollectionWithExplicitViewClientRecvCode = `// Recv reads instances of
// "service_server_streaming_result_type_collection_with_explicit_viewpb.ResultTypeCollection"
// from the "MethodServerStreamingResultTypeCollectionWithExplicitView"
// endpoint gRPC stream.
func (s *MethodServerStreamingResultTypeCollectionWithExplicitViewClientStream) Recv() (serviceserverstreamingresulttypecollectionwithexplicitview.ResultTypeCollection, error) {
	var res serviceserverstreamingresulttypecollectionwithexplicitview.ResultTypeCollection
	v, err := s.stream.Recv()
	if err != nil {
		return res, err
	}
	proj := NewResultTypeCollectionResultTypeCollection(v)
	vres := serviceserverstreamingresulttypecollectionwithexplicitviewviews.ResultTypeCollection{Projected: proj, View: "tiny"}
	if err := serviceserverstreamingresulttypecollectionwithexplicitviewviews.ValidateResultTypeCollection(vres); err != nil {
		return nil, err
	}
	return serviceserverstreamingresulttypecollectionwithexplicitview.NewResultTypeCollection(vres), nil
}

// RecvWithContext reads instances of
// "service_server_streaming_result_type_collection_with_explicit_viewpb.ResultTypeCollection"
// from the "MethodServerStreamingResultTypeCollectionWithExplicitView"
// endpoint gRPC stream with context.
func (s *MethodServerStreamingResultTypeCollectionWithExplicitViewClientStream) RecvWithContext(ctx context.Context) (serviceserverstreamingresulttypecollectionwithexplicitview.ResultTypeCollection, error) {
	return s.Recv()
}
`

var ServerStreamingPrimitiveServerSendCode = `// Send streams instances of
// "service_server_streaming_rpcpb.MethodServerStreamingRPCResponse" to the
// "MethodServerStreamingRPC" endpoint gRPC stream.
func (s *MethodServerStreamingRPCServerStream) Send(res string) error {
	v := NewProtoMethodServerStreamingRPCResponse(res)
	return s.stream.Send(v)
}

// SendWithContext streams instances of
// "service_server_streaming_rpcpb.MethodServerStreamingRPCResponse" to the
// "MethodServerStreamingRPC" endpoint gRPC stream with context.
func (s *MethodServerStreamingRPCServerStream) SendWithContext(ctx context.Context, res string) error {
	return s.Send(res)
}
`

var ServerStreamingPrimitiveClientRecvCode = `// Recv reads instances of
// "service_server_streaming_rpcpb.MethodServerStreamingRPCResponse" from the
// "MethodServerStreamingRPC" endpoint gRPC stream.
func (s *MethodServerStreamingRPCClientStream) Recv() (string, error) {
	var res string
	v, err := s.stream.Recv()
	if err != nil {
		return res, err
	}
	if err = ValidateMethodServerStreamingRPCResponse(v); err != nil {
		return res, err
	}
	return NewMethodServerStreamingRPCResponseMethodServerStreamingRPCResponse(v), nil
}

// RecvWithContext reads instances of
// "service_server_streaming_rpcpb.MethodServerStreamingRPCResponse" from the
// "MethodServerStreamingRPC" endpoint gRPC stream with context.
func (s *MethodServerStreamingRPCClientStream) RecvWithContext(ctx context.Context) (string, error) {
	return s.Recv()
}
`

var ServerStreamingArrayServerSendCode = `// Send streams instances of
// "service_server_streaming_arraypb.MethodServerStreamingArrayResponse" to the
// "MethodServerStreamingArray" endpoint gRPC stream.
func (s *MethodServerStreamingArrayServerStream) Send(res []int) error {
	v := NewProtoMethodServerStreamingArrayResponse(res)
	return s.stream.Send(v)
}

// SendWithContext streams instances of
// "service_server_streaming_arraypb.MethodServerStreamingArrayResponse" to the
// "MethodServerStreamingArray" endpoint gRPC stream with context.
func (s *MethodServerStreamingArrayServerStream) SendWithContext(ctx context.Context, res []int) error {
	return s.Send(res)
}
`

var ServerStreamingArrayClientRecvCode = `// Recv reads instances of
// "service_server_streaming_arraypb.MethodServerStreamingArrayResponse" from
// the "MethodServerStreamingArray" endpoint gRPC stream.
func (s *MethodServerStreamingArrayClientStream) Recv() ([]int, error) {
	var res []int
	v, err := s.stream.Recv()
	if err != nil {
		return res, err
	}
	return NewMethodServerStreamingArrayResponseMethodServerStreamingArrayResponse(v), nil
}

// RecvWithContext reads instances of
// "service_server_streaming_arraypb.MethodServerStreamingArrayResponse" from
// the "MethodServerStreamingArray" endpoint gRPC stream with context.
func (s *MethodServerStreamingArrayClientStream) RecvWithContext(ctx context.Context) ([]int, error) {
	return s.Recv()
}
`

var ServerStreamingMapServerSendCode = `// Send streams instances of
// "service_server_streaming_mappb.MethodServerStreamingMapResponse" to the
// "MethodServerStreamingMap" endpoint gRPC stream.
func (s *MethodServerStreamingMapServerStream) Send(res map[string]*serviceserverstreamingmap.UserType) error {
	v := NewProtoMethodServerStreamingMapResponse(res)
	return s.stream.Send(v)
}

// SendWithContext streams instances of
// "service_server_streaming_mappb.MethodServerStreamingMapResponse" to the
// "MethodServerStreamingMap" endpoint gRPC stream with context.
func (s *MethodServerStreamingMapServerStream) SendWithContext(ctx context.Context, res map[string]*serviceserverstreamingmap.UserType) error {
	return s.Send(res)
}
`

var ServerStreamingMapClientRecvCode = `// Recv reads instances of
// "service_server_streaming_mappb.MethodServerStreamingMapResponse" from the
// "MethodServerStreamingMap" endpoint gRPC stream.
func (s *MethodServerStreamingMapClientStream) Recv() (map[string]*serviceserverstreamingmap.UserType, error) {
	var res map[string]*serviceserverstreamingmap.UserType
	v, err := s.stream.Recv()
	if err != nil {
		return res, err
	}
	return NewMethodServerStreamingMapResponseMethodServerStreamingMapResponse(v), nil
}

// RecvWithContext reads instances of
// "service_server_streaming_mappb.MethodServerStreamingMapResponse" from the
// "MethodServerStreamingMap" endpoint gRPC stream with context.
func (s *MethodServerStreamingMapClientStream) RecvWithContext(ctx context.Context) (map[string]*serviceserverstreamingmap.UserType, error) {
	return s.Recv()
}
`

var ServerStreamingServerRPCSharedResultRecvCode = `// Recv reads instances of
// "service_server_streaming_rpcpb.MethodServerStreamingRPCResponse" from the
// "MethodServerStreamingRPC" endpoint gRPC stream.
func (s *MethodServerStreamingRPCClientStream) Recv() (*serviceserverstreamingrpc.UserType, error) {
	var res *serviceserverstreamingrpc.UserType
	v, err := s.stream.Recv()
	if err != nil {
		return res, err
	}
	return NewMethodServerStreamingRPCResponseUserType(v), nil
}

// RecvWithContext reads instances of
// "service_server_streaming_rpcpb.MethodServerStreamingRPCResponse" from the
// "MethodServerStreamingRPC" endpoint gRPC stream with context.
func (s *MethodServerStreamingRPCClientStream) RecvWithContext(ctx context.Context) (*serviceserverstreamingrpc.UserType, error) {
	return s.Recv()
}

// Recv reads instances of
// "service_server_streaming_rpcpb.OtherMethodServerStreamingRPCResponse" from
// the "OtherMethodServerStreamingRPC" endpoint gRPC stream.
func (s *OtherMethodServerStreamingRPCClientStream) Recv() (*serviceserverstreamingrpc.UserType, error) {
	var res *serviceserverstreamingrpc.UserType
	v, err := s.stream.Recv()
	if err != nil {
		return res, err
	}
	return NewOtherMethodServerStreamingRPCResponseUserType(v), nil
}

// RecvWithContext reads instances of
// "service_server_streaming_rpcpb.OtherMethodServerStreamingRPCResponse" from
// the "OtherMethodServerStreamingRPC" endpoint gRPC stream with context.
func (s *OtherMethodServerStreamingRPCClientStream) RecvWithContext(ctx context.Context) (*serviceserverstreamingrpc.UserType, error) {
	return s.Recv()
}
`

var ClientStreamingServerStructCode = `// MethodClientStreamingRPCServerStream implements the
// serviceclientstreamingrpc.MethodClientStreamingRPCServerStream interface.
type MethodClientStreamingRPCServerStream struct {
	stream service_client_streaming_rpcpb.ServiceClientStreamingRPC_MethodClientStreamingRPCServer
}
`

var ClientStreamingServerSendCode = `// SendAndClose streams instances of
// "service_client_streaming_rpcpb.MethodClientStreamingRPCResponse" to the
// "MethodClientStreamingRPC" endpoint gRPC stream.
func (s *MethodClientStreamingRPCServerStream) SendAndClose(res string) error {
	v := NewProtoMethodClientStreamingRPCResponse(res)
	return s.stream.SendAndClose(v)
}

// SendAndCloseWithContext streams instances of
// "service_client_streaming_rpcpb.MethodClientStreamingRPCResponse" to the
// "MethodClientStreamingRPC" endpoint gRPC stream with context.
func (s *MethodClientStreamingRPCServerStream) SendAndCloseWithContext(ctx context.Context, res string) error {
	return s.SendAndClose(res)
}
`

var ClientStreamingServerRecvCode = `// Recv reads instances of
// "service_client_streaming_rpcpb.MethodClientStreamingRPCStreamingRequest"
// from the "MethodClientStreamingRPC" endpoint gRPC stream.
func (s *MethodClientStreamingRPCServerStream) Recv() (int, error) {
	var res int
	v, err := s.stream.Recv()
	if err != nil {
		return res, err
	}
	if err = ValidateMethodClientStreamingRPCStreamingRequest(v); err != nil {
		return res, err
	}
	return NewMethodClientStreamingRPCStreamingRequestMethodClientStreamingRPCStreamingRequest(v), nil
}

// RecvWithContext reads instances of
// "service_client_streaming_rpcpb.MethodClientStreamingRPCStreamingRequest"
// from the "MethodClientStreamingRPC" endpoint gRPC stream with context.
func (s *MethodClientStreamingRPCServerStream) RecvWithContext(ctx context.Context) (int, error) {
	return s.Recv()
}
`

var ClientStreamingClientStructCode = `// MethodClientStreamingRPCClientStream implements the
// serviceclientstreamingrpc.MethodClientStreamingRPCClientStream interface.
type MethodClientStreamingRPCClientStream struct {
	stream service_client_streaming_rpcpb.ServiceClientStreamingRPC_MethodClientStreamingRPCClient
}
`

var ClientStreamingClientSendCode = `// Send streams instances of
// "service_client_streaming_rpcpb.MethodClientStreamingRPCStreamingRequest" to
// the "MethodClientStreamingRPC" endpoint gRPC stream.
func (s *MethodClientStreamingRPCClientStream) Send(res int) error {
	v := NewProtoMethodClientStreamingRPCStreamingRequest(res)
	return s.stream.Send(v)
}

// SendWithContext streams instances of
// "service_client_streaming_rpcpb.MethodClientStreamingRPCStreamingRequest" to
// the "MethodClientStreamingRPC" endpoint gRPC stream with context.
func (s *MethodClientStreamingRPCClientStream) SendWithContext(ctx context.Context, res int) error {
	return s.Send(res)
}
`

var ClientStreamingClientRecvCode = `// CloseAndRecv reads instances of
// "service_client_streaming_rpcpb.MethodClientStreamingRPCResponse" from the
// "MethodClientStreamingRPC" endpoint gRPC stream.
func (s *MethodClientStreamingRPCClientStream) CloseAndRecv() (string, error) {
	var res string
	v, err := s.stream.CloseAndRecv()
	if err != nil {
		return res, err
	}
	if err = ValidateMethodClientStreamingRPCResponse(v); err != nil {
		return res, err
	}
	return NewMethodClientStreamingRPCResponseMethodClientStreamingRPCResponse(v), nil
}

// CloseAndRecvWithContext reads instances of
// "service_client_streaming_rpcpb.MethodClientStreamingRPCResponse" from the
// "MethodClientStreamingRPC" endpoint gRPC stream with context.
func (s *MethodClientStreamingRPCClientStream) CloseAndRecvWithContext(ctx context.Context) (string, error) {
	return s.CloseAndRecv()
}
`

var ClientStreamingServerNoResultCloseCode = `func (s *MethodClientStreamingNoResultServerStream) Close() error {
	// synchronize stream
	return s.stream.SendAndClose(&service_client_streaming_no_resultpb.MethodClientStreamingNoResultResponse{})
}
`

var ClientStreamingClientNoResultCloseCode = `func (s *MethodClientStreamingNoResultClientStream) Close() error {
	// synchronize and report any server error
	_, err := s.stream.CloseAndRecv()
	return err
}
`

var BidirectionalStreamingServerStructCode = `// MethodBidirectionalStreamingRPCServerStream implements the
// servicebidirectionalstreamingrpc.MethodBidirectionalStreamingRPCServerStream
// interface.
type MethodBidirectionalStreamingRPCServerStream struct {
	stream service_bidirectional_streaming_rpcpb.ServiceBidirectionalStreamingRPC_MethodBidirectionalStreamingRPCServer
}
`

var BidirectionalStreamingServerSendCode = `// Send streams instances of
// "service_bidirectional_streaming_rpcpb.MethodBidirectionalStreamingRPCResponse"
// to the "MethodBidirectionalStreamingRPC" endpoint gRPC stream.
func (s *MethodBidirectionalStreamingRPCServerStream) Send(res *servicebidirectionalstreamingrpc.ID) error {
	vres := servicebidirectionalstreamingrpc.NewViewedID(res, "default")
	v := NewProtoMethodBidirectionalStreamingRPCResponse(vres.Projected)
	return s.stream.Send(v)
}

// SendWithContext streams instances of
// "service_bidirectional_streaming_rpcpb.MethodBidirectionalStreamingRPCResponse"
// to the "MethodBidirectionalStreamingRPC" endpoint gRPC stream with context.
func (s *MethodBidirectionalStreamingRPCServerStream) SendWithContext(ctx context.Context, res *servicebidirectionalstreamingrpc.ID) error {
	return s.Send(res)
}
`

var BidirectionalStreamingServerRecvCode = `// Recv reads instances of
// "service_bidirectional_streaming_rpcpb.MethodBidirectionalStreamingRPCStreamingRequest"
// from the "MethodBidirectionalStreamingRPC" endpoint gRPC stream.
func (s *MethodBidirectionalStreamingRPCServerStream) Recv() (int, error) {
	var res int
	v, err := s.stream.Recv()
	if err != nil {
		return res, err
	}
	if err = ValidateMethodBidirectionalStreamingRPCStreamingRequest(v); err != nil {
		return res, err
	}
	return NewMethodBidirectionalStreamingRPCStreamingRequestMethodBidirectionalStreamingRPCStreamingRequest(v), nil
}

// RecvWithContext reads instances of
// "service_bidirectional_streaming_rpcpb.MethodBidirectionalStreamingRPCStreamingRequest"
// from the "MethodBidirectionalStreamingRPC" endpoint gRPC stream with context.
func (s *MethodBidirectionalStreamingRPCServerStream) RecvWithContext(ctx context.Context) (int, error) {
	return s.Recv()
}
`

var BidirectionalStreamingServerCloseCode = `func (s *MethodBidirectionalStreamingRPCServerStream) Close() error {
	// nothing to do here
	return nil
}
`

var BidirectionalStreamingClientStructCode = `// MethodBidirectionalStreamingRPCClientStream implements the
// servicebidirectionalstreamingrpc.MethodBidirectionalStreamingRPCClientStream
// interface.
type MethodBidirectionalStreamingRPCClientStream struct {
	stream service_bidirectional_streaming_rpcpb.ServiceBidirectionalStreamingRPC_MethodBidirectionalStreamingRPCClient
}
`

var BidirectionalStreamingClientSendCode = `// Send streams instances of
// "service_bidirectional_streaming_rpcpb.MethodBidirectionalStreamingRPCStreamingRequest"
// to the "MethodBidirectionalStreamingRPC" endpoint gRPC stream.
func (s *MethodBidirectionalStreamingRPCClientStream) Send(res int) error {
	v := NewProtoMethodBidirectionalStreamingRPCStreamingRequest(res)
	return s.stream.Send(v)
}

// SendWithContext streams instances of
// "service_bidirectional_streaming_rpcpb.MethodBidirectionalStreamingRPCStreamingRequest"
// to the "MethodBidirectionalStreamingRPC" endpoint gRPC stream with context.
func (s *MethodBidirectionalStreamingRPCClientStream) SendWithContext(ctx context.Context, res int) error {
	return s.Send(res)
}
`

var BidirectionalStreamingClientRecvCode = `// Recv reads instances of
// "service_bidirectional_streaming_rpcpb.MethodBidirectionalStreamingRPCResponse"
// from the "MethodBidirectionalStreamingRPC" endpoint gRPC stream.
func (s *MethodBidirectionalStreamingRPCClientStream) Recv() (*servicebidirectionalstreamingrpc.ID, error) {
	var res *servicebidirectionalstreamingrpc.ID
	v, err := s.stream.Recv()
	if err != nil {
		return res, err
	}
	proj := NewMethodBidirectionalStreamingRPCResponseIDView(v)
	vres := &servicebidirectionalstreamingrpcviews.ID{Projected: proj, View: "default"}
	if err := servicebidirectionalstreamingrpcviews.ValidateID(vres); err != nil {
		return nil, err
	}
	return servicebidirectionalstreamingrpc.NewID(vres), nil
}

// RecvWithContext reads instances of
// "service_bidirectional_streaming_rpcpb.MethodBidirectionalStreamingRPCResponse"
// from the "MethodBidirectionalStreamingRPC" endpoint gRPC stream with context.
func (s *MethodBidirectionalStreamingRPCClientStream) RecvWithContext(ctx context.Context) (*servicebidirectionalstreamingrpc.ID, error) {
	return s.Recv()
}
`

var BidirectionalStreamingClientCloseCode = `func (s *MethodBidirectionalStreamingRPCClientStream) Close() error {
	// Close the send direction of the stream
	return s.stream.CloseSend()
}
`
