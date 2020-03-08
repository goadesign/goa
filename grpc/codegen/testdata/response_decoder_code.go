package testdata

const ResultWithViewsResponseDecoderCode = `// DecodeMethodMessageResultTypeWithViewsResponse decodes responses from the
// ServiceMessageResultTypeWithViews MethodMessageResultTypeWithViews endpoint.
func DecodeMethodMessageResultTypeWithViewsResponse(ctx context.Context, v interface{}, hdr, trlr metadata.MD) (interface{}, error) {
	var view string
	{
		if vals := hdr.Get("goa-view"); len(vals) > 0 {
			view = vals[0]
		}
	}
	message, ok := v.(*service_message_result_type_with_viewspb.MethodMessageResultTypeWithViewsResponse)
	if !ok {
		return nil, goagrpc.ErrInvalidType("ServiceMessageResultTypeWithViews", "MethodMessageResultTypeWithViews", "*service_message_result_type_with_viewspb.MethodMessageResultTypeWithViewsResponse", v)
	}
	res := NewMethodMessageResultTypeWithViewsResult(message)
	vres := &servicemessageresulttypewithviewsviews.RT{Projected: res, View: view}
	if err := servicemessageresulttypewithviewsviews.ValidateRT(vres); err != nil {
		return nil, err
	}
	return servicemessageresulttypewithviews.NewRT(vres), nil
}
`

const ResultWithExplicitViewResponseDecoderCode = `// DecodeMethodMessageResultTypeWithExplicitViewResponse decodes responses from
// the ServiceMessageResultTypeWithExplicitView
// MethodMessageResultTypeWithExplicitView endpoint.
func DecodeMethodMessageResultTypeWithExplicitViewResponse(ctx context.Context, v interface{}, hdr, trlr metadata.MD) (interface{}, error) {
	var view string
	{
		if vals := hdr.Get("goa-view"); len(vals) > 0 {
			view = vals[0]
		}
	}
	message, ok := v.(*service_message_result_type_with_explicit_viewpb.MethodMessageResultTypeWithExplicitViewResponse)
	if !ok {
		return nil, goagrpc.ErrInvalidType("ServiceMessageResultTypeWithExplicitView", "MethodMessageResultTypeWithExplicitView", "*service_message_result_type_with_explicit_viewpb.MethodMessageResultTypeWithExplicitViewResponse", v)
	}
	res := NewMethodMessageResultTypeWithExplicitViewResult(message)
	vres := &servicemessageresulttypewithexplicitviewviews.RT{Projected: res, View: view}
	if err := servicemessageresulttypewithexplicitviewviews.ValidateRT(vres); err != nil {
		return nil, err
	}
	return servicemessageresulttypewithexplicitview.NewRT(vres), nil
}
`

const ResultArrayResponseDecoderCode = `// DecodeMethodMessageArrayResponse decodes responses from the
// ServiceMessageArray MethodMessageArray endpoint.
func DecodeMethodMessageArrayResponse(ctx context.Context, v interface{}, hdr, trlr metadata.MD) (interface{}, error) {
	message, ok := v.(*service_message_arraypb.MethodMessageArrayResponse)
	if !ok {
		return nil, goagrpc.ErrInvalidType("ServiceMessageArray", "MethodMessageArray", "*service_message_arraypb.MethodMessageArrayResponse", v)
	}
	res := NewMethodMessageArrayResult(message)
	return res, nil
}
`

const ResultPrimitiveResponseDecoderCode = `// DecodeMethodUnaryRPCNoPayloadResponse decodes responses from the
// ServiceUnaryRPCNoPayload MethodUnaryRPCNoPayload endpoint.
func DecodeMethodUnaryRPCNoPayloadResponse(ctx context.Context, v interface{}, hdr, trlr metadata.MD) (interface{}, error) {
	message, ok := v.(*service_unary_rpc_no_payloadpb.MethodUnaryRPCNoPayloadResponse)
	if !ok {
		return nil, goagrpc.ErrInvalidType("ServiceUnaryRPCNoPayload", "MethodUnaryRPCNoPayload", "*service_unary_rpc_no_payloadpb.MethodUnaryRPCNoPayloadResponse", v)
	}
	res := NewMethodUnaryRPCNoPayloadResult(message)
	return res, nil
}
`

const ResultWithMetadataResponseDecoderCode = `// DecodeMethodMessageWithMetadataResponse decodes responses from the
// ServiceMessageWithMetadata MethodMessageWithMetadata endpoint.
func DecodeMethodMessageWithMetadataResponse(ctx context.Context, v interface{}, hdr, trlr metadata.MD) (interface{}, error) {
	var (
		inHeader  *int
		inTrailer *bool
		err       error
	)
	{

		if vals := hdr.Get("Location"); len(vals) > 0 {
			inHeaderRaw = vals[0]

			v, err2 := strconv.ParseInt(inHeaderRaw, 10, strconv.IntSize)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError("inHeader", inHeaderRaw, "integer"))
			}
			pv := int(v)
			inHeader = &pv
		}

		if vals := trlr.Get("InTrailer"); len(vals) > 0 {
			inTrailerRaw = vals[0]

			v, err2 := strconv.ParseBool(inTrailerRaw)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError("inTrailer", inTrailerRaw, "boolean"))
			}
			inTrailer = &v
		}
	}
	if err != nil {
		return nil, err
	}
	message, ok := v.(*service_message_with_metadatapb.MethodMessageWithMetadataResponse)
	if !ok {
		return nil, goagrpc.ErrInvalidType("ServiceMessageWithMetadata", "MethodMessageWithMetadata", "*service_message_with_metadatapb.MethodMessageWithMetadataResponse", v)
	}
	res := NewMethodMessageWithMetadataResult(message, inHeader, inTrailer)
	return res, nil
}
`

const ResultWithValidateResponseDecoderCode = `// DecodeMethodMessageWithValidateResponse decodes responses from the
// ServiceMessageWithValidate MethodMessageWithValidate endpoint.
func DecodeMethodMessageWithValidateResponse(ctx context.Context, v interface{}, hdr, trlr metadata.MD) (interface{}, error) {
	var (
		inHeader  *int
		inTrailer *bool
		err       error
	)
	{

		if vals := hdr.Get("Location"); len(vals) > 0 {
			inHeaderRaw = vals[0]

			v, err2 := strconv.ParseInt(inHeaderRaw, 10, strconv.IntSize)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError("inHeader", inHeaderRaw, "integer"))
			}
			pv := int(v)
			inHeader = &pv
		}
		if inHeader != nil {
			if *inHeader < 1 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("inHeader", *inHeader, 1, true))
			}
		}

		if vals := trlr.Get("InTrailer"); len(vals) > 0 {
			inTrailerRaw = vals[0]

			v, err2 := strconv.ParseBool(inTrailerRaw)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError("inTrailer", inTrailerRaw, "boolean"))
			}
			inTrailer = &v
		}
		if inTrailer != nil {
			if !(*inTrailer == true) {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("inTrailer", *inTrailer, []interface{}{true}))
			}
		}
	}
	if err != nil {
		return nil, err
	}
	message, ok := v.(*service_message_with_validatepb.MethodMessageWithValidateResponse)
	if !ok {
		return nil, goagrpc.ErrInvalidType("ServiceMessageWithValidate", "MethodMessageWithValidate", "*service_message_with_validatepb.MethodMessageWithValidateResponse", v)
	}
	if err = ValidateMethodMessageWithValidateResponse(message); err != nil {
		return nil, err
	}
	res := NewMethodMessageWithValidateResult(message, inHeader, inTrailer)
	return res, nil
}
`

const ResultCollectionResponseDecoderCode = `// DecodeMethodMessageUserTypeWithNestedUserTypesResponse decodes responses
// from the ServiceMessageUserTypeWithNestedUserTypes
// MethodMessageUserTypeWithNestedUserTypes endpoint.
func DecodeMethodMessageUserTypeWithNestedUserTypesResponse(ctx context.Context, v interface{}, hdr, trlr metadata.MD) (interface{}, error) {
	var view string
	{
		if vals := hdr.Get("goa-view"); len(vals) > 0 {
			view = vals[0]
		}
	}
	message, ok := v.(*service_message_user_type_with_nested_user_typespb.RTCollection)
	if !ok {
		return nil, goagrpc.ErrInvalidType("ServiceMessageUserTypeWithNestedUserTypes", "MethodMessageUserTypeWithNestedUserTypes", "*service_message_user_type_with_nested_user_typespb.RTCollection", v)
	}
	res := NewMethodMessageUserTypeWithNestedUserTypesResult(message)
	vres := servicemessageusertypewithnestedusertypesviews.RTCollection{Projected: res, View: view}
	if err := servicemessageusertypewithnestedusertypesviews.ValidateRTCollection(vres); err != nil {
		return nil, err
	}
	return servicemessageusertypewithnestedusertypes.NewRTCollection(vres), nil
}
`

const ServerStreamingResponseDecoderCode = `// DecodeMethodServerStreamingUserTypeRPCResponse decodes responses from the
// ServiceServerStreamingUserTypeRPC MethodServerStreamingUserTypeRPC endpoint.
func DecodeMethodServerStreamingUserTypeRPCResponse(ctx context.Context, v interface{}, hdr, trlr metadata.MD) (interface{}, error) {
	return &MethodServerStreamingUserTypeRPCClientStream{
		stream: v.(service_server_streaming_user_type_rpcpb.ServiceServerStreamingUserTypeRPC_MethodServerStreamingUserTypeRPCClient),
	}, nil
}
`

const ServerStreamingResultWithViewsResponseDecoderCode = `// DecodeMethodServerStreamingUserTypeRPCResponse decodes responses from the
// ServiceServerStreamingUserTypeRPC MethodServerStreamingUserTypeRPC endpoint.
func DecodeMethodServerStreamingUserTypeRPCResponse(ctx context.Context, v interface{}, hdr, trlr metadata.MD) (interface{}, error) {
	var view string
	{
		if vals := hdr.Get("goa-view"); len(vals) > 0 {
			view = vals[0]
		}
	}
	return &MethodServerStreamingUserTypeRPCClientStream{
		stream: v.(service_server_streaming_user_type_rpcpb.ServiceServerStreamingUserTypeRPC_MethodServerStreamingUserTypeRPCClient),
		view:   view,
	}, nil
}
`

const ClientStreamingResponseDecoderCode = `// DecodeMethodClientStreamingRPCResponse decodes responses from the
// ServiceClientStreamingRPC MethodClientStreamingRPC endpoint.
func DecodeMethodClientStreamingRPCResponse(ctx context.Context, v interface{}, hdr, trlr metadata.MD) (interface{}, error) {
	return &MethodClientStreamingRPCClientStream{
		stream: v.(service_client_streaming_rpcpb.ServiceClientStreamingRPC_MethodClientStreamingRPCClient),
	}, nil
}
`

const BidirectionalStreamingResponseDecoderCode = `// DecodeMethodBidirectionalStreamingRPCResponse decodes responses from the
// ServiceBidirectionalStreamingRPC MethodBidirectionalStreamingRPC endpoint.
func DecodeMethodBidirectionalStreamingRPCResponse(ctx context.Context, v interface{}, hdr, trlr metadata.MD) (interface{}, error) {
	var view string
	{
		if vals := hdr.Get("goa-view"); len(vals) > 0 {
			view = vals[0]
		}
	}
	return &MethodBidirectionalStreamingRPCClientStream{
		stream: v.(service_bidirectional_streaming_rpcpb.ServiceBidirectionalStreamingRPC_MethodBidirectionalStreamingRPCClient),
		view:   view,
	}, nil
}
`
