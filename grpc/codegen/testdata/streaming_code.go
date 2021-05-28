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
	v := NewMethodServerStreamingUserTypeRPCResponse(res)
	return s.stream.Send(v)
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
	return NewUserType(v), nil
}
`

var ServerStreamingResultWithViewsServerStructCode = `// MethodServerStreamingUserTypeRPCServerStream implements the
// serviceserverstreamingusertyperpc.MethodServerStreamingUserTypeRPCServerStream
// interface.
type MethodServerStreamingUserTypeRPCServerStream struct {
	stream service_server_streaming_user_type_rpcpb.ServiceServerStreamingUserTypeRPC_MethodServerStreamingUserTypeRPCServer
	view   string
}
`

var ServerStreamingResultWithViewsServerSendCode = `// Send streams instances of
// "service_server_streaming_user_type_rpcpb.MethodServerStreamingUserTypeRPCResponse"
// to the "MethodServerStreamingUserTypeRPC" endpoint gRPC stream.
func (s *MethodServerStreamingUserTypeRPCServerStream) Send(res *serviceserverstreamingusertyperpc.ResultType) error {
	vres := serviceserverstreamingusertyperpc.NewViewedResultType(res, s.view)
	v := NewMethodServerStreamingUserTypeRPCResponse(vres.Projected)
	return s.stream.Send(v)
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
	stream service_server_streaming_user_type_rpcpb.ServiceServerStreamingUserTypeRPC_MethodServerStreamingUserTypeRPCClient
	view   string
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
	proj := NewResultTypeView(v)
	vres := &serviceserverstreamingusertyperpcviews.ResultType{Projected: proj, View: s.view}
	if err := serviceserverstreamingusertyperpcviews.ValidateResultType(vres); err != nil {
		return nil, err
	}
	return serviceserverstreamingusertyperpc.NewResultType(vres), nil
}
`

var ServerStreamingResultWithViewsClientSetViewCode = `// SetView sets the view.
func (s *MethodServerStreamingUserTypeRPCClientStream) SetView(view string) {
	s.view = view
}
`

var ServerStreamingResultCollectionWithExplicitViewServerSendCode = `// Send streams instances of
// "service_server_streaming_result_type_collection_with_explicit_viewpb.ResultTypeCollection"
// to the "MethodServerStreamingResultTypeCollectionWithExplicitView" endpoint
// gRPC stream.
func (s *MethodServerStreamingResultTypeCollectionWithExplicitViewServerStream) Send(res serviceserverstreamingresulttypecollectionwithexplicitview.ResultTypeCollection) error {
	vres := serviceserverstreamingresulttypecollectionwithexplicitview.NewViewedResultTypeCollection(res, "tiny")
	v := NewResultTypeCollection(vres.Projected)
	return s.stream.Send(v)
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
	proj := NewResultTypeCollection(v)
	vres := serviceserverstreamingresulttypecollectionwithexplicitviewviews.ResultTypeCollection{Projected: proj, View: "tiny"}
	if err := serviceserverstreamingresulttypecollectionwithexplicitviewviews.ValidateResultTypeCollection(vres); err != nil {
		return nil, err
	}
	return serviceserverstreamingresulttypecollectionwithexplicitview.NewResultTypeCollection(vres), nil
}
`

var ServerStreamingPrimitiveServerSendCode = `// Send streams instances of
// "service_server_streaming_rpcpb.MethodServerStreamingRPCResponse" to the
// "MethodServerStreamingRPC" endpoint gRPC stream.
func (s *MethodServerStreamingRPCServerStream) Send(res string) error {
	v := NewMethodServerStreamingRPCResponse(res)
	return s.stream.Send(v)
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
	return NewMethodServerStreamingRPCResponse(v), nil
}
`

var ServerStreamingArrayServerSendCode = `// Send streams instances of
// "service_server_streaming_arraypb.MethodServerStreamingArrayResponse" to the
// "MethodServerStreamingArray" endpoint gRPC stream.
func (s *MethodServerStreamingArrayServerStream) Send(res []int) error {
	v := NewMethodServerStreamingArrayResponse(res)
	return s.stream.Send(v)
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
	return NewMethodServerStreamingArrayResponse(v), nil
}
`

var ServerStreamingMapServerSendCode = `// Send streams instances of
// "service_server_streaming_mappb.MethodServerStreamingMapResponse" to the
// "MethodServerStreamingMap" endpoint gRPC stream.
func (s *MethodServerStreamingMapServerStream) Send(res map[string]*serviceserverstreamingmap.UserType) error {
	v := NewMethodServerStreamingMapResponse(res)
	return s.stream.Send(v)
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
	return NewMethodServerStreamingMapResponse(v), nil
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
	v := NewMethodClientStreamingRPCResponse(res)
	return s.stream.SendAndClose(v)
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
	return NewMethodClientStreamingRPCStreamingRequest(v), nil
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
	v := NewMethodClientStreamingRPCStreamingRequest(res)
	return s.stream.Send(v)
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
	return NewMethodClientStreamingRPCResponse(v), nil
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
	view   string
}
`

var BidirectionalStreamingServerSendCode = `// Send streams instances of
// "service_bidirectional_streaming_rpcpb.MethodBidirectionalStreamingRPCResponse"
// to the "MethodBidirectionalStreamingRPC" endpoint gRPC stream.
func (s *MethodBidirectionalStreamingRPCServerStream) Send(res *servicebidirectionalstreamingrpc.ID) error {
	vres := servicebidirectionalstreamingrpc.NewViewedID(res, "default")
	v := NewMethodBidirectionalStreamingRPCResponse(vres.Projected)
	return s.stream.Send(v)
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
	return NewMethodBidirectionalStreamingRPCStreamingRequest(v), nil
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
	view   string
}
`

var BidirectionalStreamingClientSendCode = `// Send streams instances of
// "service_bidirectional_streaming_rpcpb.MethodBidirectionalStreamingRPCStreamingRequest"
// to the "MethodBidirectionalStreamingRPC" endpoint gRPC stream.
func (s *MethodBidirectionalStreamingRPCClientStream) Send(res int) error {
	v := NewMethodBidirectionalStreamingRPCStreamingRequest(res)
	return s.stream.Send(v)
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
	proj := NewIDView(v)
	vres := &servicebidirectionalstreamingrpcviews.ID{Projected: proj, View: "default"}
	if err := servicebidirectionalstreamingrpcviews.ValidateID(vres); err != nil {
		return nil, err
	}
	return servicebidirectionalstreamingrpc.NewID(vres), nil
}
`

var BidirectionalStreamingClientCloseCode = `func (s *MethodBidirectionalStreamingRPCClientStream) Close() error {
	// Close the send direction of the stream
	return s.stream.CloseSend()
}
`
