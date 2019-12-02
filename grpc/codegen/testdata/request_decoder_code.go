package testdata

const PayloadUserTypeRequestDecoderCode = `// DecodeMethodMessageUserTypeWithNestedUserTypesRequest decodes requests sent
// to "ServiceMessageUserTypeWithNestedUserTypes" service
// "MethodMessageUserTypeWithNestedUserTypes" endpoint.
func DecodeMethodMessageUserTypeWithNestedUserTypesRequest(ctx context.Context, v interface{}, md metadata.MD) (interface{}, error) {
	var (
		message *service_message_user_type_with_nested_user_typespb.MethodMessageUserTypeWithNestedUserTypesRequest
		ok      bool
	)
	{
		if message, ok = v.(*service_message_user_type_with_nested_user_typespb.MethodMessageUserTypeWithNestedUserTypesRequest); !ok {
			return nil, goagrpc.ErrInvalidType("ServiceMessageUserTypeWithNestedUserTypes", "MethodMessageUserTypeWithNestedUserTypes", "*service_message_user_type_with_nested_user_typespb.MethodMessageUserTypeWithNestedUserTypesRequest", v)
		}
	}
	var payload *servicemessageusertypewithnestedusertypes.UT
	{
		payload = NewMethodMessageUserTypeWithNestedUserTypesPayload(message)
	}
	return payload, nil
}
`

const PayloadArrayRequestDecoderCode = `// DecodeMethodUnaryRPCNoResultRequest decodes requests sent to
// "ServiceUnaryRPCNoResult" service "MethodUnaryRPCNoResult" endpoint.
func DecodeMethodUnaryRPCNoResultRequest(ctx context.Context, v interface{}, md metadata.MD) (interface{}, error) {
	var (
		message *service_unary_rpc_no_resultpb.MethodUnaryRPCNoResultRequest
		ok      bool
	)
	{
		if message, ok = v.(*service_unary_rpc_no_resultpb.MethodUnaryRPCNoResultRequest); !ok {
			return nil, goagrpc.ErrInvalidType("ServiceUnaryRPCNoResult", "MethodUnaryRPCNoResult", "*service_unary_rpc_no_resultpb.MethodUnaryRPCNoResultRequest", v)
		}
	}
	var payload []string
	{
		payload = NewMethodUnaryRPCNoResultPayload(message)
	}
	return payload, nil
}
`

const PayloadMapRequestDecoderCode = `// DecodeMethodMessageMapRequest decodes requests sent to "ServiceMessageMap"
// service "MethodMessageMap" endpoint.
func DecodeMethodMessageMapRequest(ctx context.Context, v interface{}, md metadata.MD) (interface{}, error) {
	var (
		message *service_message_mappb.MethodMessageMapRequest
		ok      bool
	)
	{
		if message, ok = v.(*service_message_mappb.MethodMessageMapRequest); !ok {
			return nil, goagrpc.ErrInvalidType("ServiceMessageMap", "MethodMessageMap", "*service_message_mappb.MethodMessageMapRequest", v)
		}
	}
	var payload map[int]*servicemessagemap.UT
	{
		payload = NewMethodMessageMapPayload(message)
	}
	return payload, nil
}
`

const PayloadPrimitiveRequestDecoderCode = `// DecodeMethodServerStreamingRPCRequest decodes requests sent to
// "ServiceServerStreamingRPC" service "MethodServerStreamingRPC" endpoint.
func DecodeMethodServerStreamingRPCRequest(ctx context.Context, v interface{}, md metadata.MD) (interface{}, error) {
	var (
		message *service_server_streaming_rpcpb.MethodServerStreamingRPCRequest
		ok      bool
	)
	{
		if message, ok = v.(*service_server_streaming_rpcpb.MethodServerStreamingRPCRequest); !ok {
			return nil, goagrpc.ErrInvalidType("ServiceServerStreamingRPC", "MethodServerStreamingRPC", "*service_server_streaming_rpcpb.MethodServerStreamingRPCRequest", v)
		}
	}
	var payload int
	{
		payload = NewMethodServerStreamingRPCPayload(message)
	}
	return payload, nil
}
`

const PayloadPrimitiveWithStreamingPayloadRequestDecoderCode = `// DecodeMethodClientStreamingRPCWithPayloadRequest decodes requests sent to
// "ServiceClientStreamingRPCWithPayload" service
// "MethodClientStreamingRPCWithPayload" endpoint.
func DecodeMethodClientStreamingRPCWithPayloadRequest(ctx context.Context, v interface{}, md metadata.MD) (interface{}, error) {
	var (
		goaPayload int
		err        error
	)
	{
		if vals := md.Get("goa_payload"); len(vals) == 0 {
			err = goa.MergeErrors(err, goa.MissingFieldError("goa_payload", "metadata"))
		} else {
			goaPayloadRaw := vals[0]

			v, err2 := strconv.ParseInt(goaPayloadRaw, 10, strconv.IntSize)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError("goaPayload", goaPayloadRaw, "integer"))
			}
			goaPayload = int(v)
		}
	}
	if err != nil {
		return nil, err
	}
	var payload int
	{
		payload = goaPayload
	}
	return payload, nil
}
`

const PayloadUserTypeWithStreamingPayloadRequestDecoderCode = `// DecodeMethodBidirectionalStreamingRPCWithPayloadRequest decodes requests
// sent to "ServiceBidirectionalStreamingRPCWithPayload" service
// "MethodBidirectionalStreamingRPCWithPayload" endpoint.
func DecodeMethodBidirectionalStreamingRPCWithPayloadRequest(ctx context.Context, v interface{}, md metadata.MD) (interface{}, error) {
	var (
		a   *int
		b   *string
		err error
	)
	{
		if vals := md.Get("a"); len(vals) > 0 {
			aRaw := vals[0]

			v, err2 := strconv.ParseInt(aRaw, 10, strconv.IntSize)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError("a", aRaw, "integer"))
			}
			pv := int(v)
			a = &pv
		}
		if vals := md.Get("b"); len(vals) > 0 {
			b = &vals[0]
		}
	}
	if err != nil {
		return nil, err
	}
	var payload *servicebidirectionalstreamingrpcwithpayload.Payload
	{
		payload = NewMethodBidirectionalStreamingRPCWithPayloadPayload(a, b)
	}
	return payload, nil
}
`

const PayloadWithMetadataRequestDecoderCode = `// DecodeMethodMessageWithMetadataRequest decodes requests sent to
// "ServiceMessageWithMetadata" service "MethodMessageWithMetadata" endpoint.
func DecodeMethodMessageWithMetadataRequest(ctx context.Context, v interface{}, md metadata.MD) (interface{}, error) {
	var (
		inMetadata *int
		err        error
	)
	{
		if vals := md.Get("Authorization"); len(vals) > 0 {
			inMetadataRaw := vals[0]

			v, err2 := strconv.ParseInt(inMetadataRaw, 10, strconv.IntSize)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError("inMetadata", inMetadataRaw, "integer"))
			}
			pv := int(v)
			inMetadata = &pv
		}
	}
	if err != nil {
		return nil, err
	}
	var (
		message *service_message_with_metadatapb.MethodMessageWithMetadataRequest
		ok      bool
	)
	{
		if message, ok = v.(*service_message_with_metadatapb.MethodMessageWithMetadataRequest); !ok {
			return nil, goagrpc.ErrInvalidType("ServiceMessageWithMetadata", "MethodMessageWithMetadata", "*service_message_with_metadatapb.MethodMessageWithMetadataRequest", v)
		}
	}
	var payload *servicemessagewithmetadata.RequestUT
	{
		payload = NewMethodMessageWithMetadataPayload(message, inMetadata)
	}
	return payload, nil
}
`

const PayloadWithValidateRequestDecoderCode = `// DecodeMethodMessageWithValidateRequest decodes requests sent to
// "ServiceMessageWithValidate" service "MethodMessageWithValidate" endpoint.
func DecodeMethodMessageWithValidateRequest(ctx context.Context, v interface{}, md metadata.MD) (interface{}, error) {
	var (
		inMetadata *int
		err        error
	)
	{
		if vals := md.Get("Authorization"); len(vals) > 0 {
			inMetadataRaw := vals[0]

			v, err2 := strconv.ParseInt(inMetadataRaw, 10, strconv.IntSize)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError("inMetadata", inMetadataRaw, "integer"))
			}
			pv := int(v)
			inMetadata = &pv
		}
		if inMetadata != nil {
			if *inMetadata > 100 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("inMetadata", *inMetadata, 100, false))
			}
		}
	}
	if err != nil {
		return nil, err
	}
	var (
		message *service_message_with_validatepb.MethodMessageWithValidateRequest
		ok      bool
	)
	{
		if message, ok = v.(*service_message_with_validatepb.MethodMessageWithValidateRequest); !ok {
			return nil, goagrpc.ErrInvalidType("ServiceMessageWithValidate", "MethodMessageWithValidate", "*service_message_with_validatepb.MethodMessageWithValidateRequest", v)
		}
		if err = ValidateMethodMessageWithValidateRequest(message); err != nil {
			return nil, err
		}
	}
	var payload *servicemessagewithvalidate.RequestUT
	{
		payload = NewMethodMessageWithValidatePayload(message, inMetadata)
	}
	return payload, nil
}
`

const PayloadWithSecurityAttrsRequestDecoderCode = `// DecodeMethodMessageWithSecurityRequest decodes requests sent to
// "ServiceMessageWithSecurity" service "MethodMessageWithSecurity" endpoint.
func DecodeMethodMessageWithSecurityRequest(ctx context.Context, v interface{}, md metadata.MD) (interface{}, error) {
	var (
		token    *string
		key      *string
		username *string
		password *string
		err      error
	)
	{
		if vals := md.Get("authorization"); len(vals) > 0 {
			token = &vals[0]
		}
		if vals := md.Get("authorization"); len(vals) > 0 {
			key = &vals[0]
		}
		if vals := md.Get("username"); len(vals) > 0 {
			username = &vals[0]
		}
		if vals := md.Get("password"); len(vals) > 0 {
			password = &vals[0]
		}
	}
	if err != nil {
		return nil, err
	}
	var (
		message *service_message_with_securitypb.MethodMessageWithSecurityRequest
		ok      bool
	)
	{
		if message, ok = v.(*service_message_with_securitypb.MethodMessageWithSecurityRequest); !ok {
			return nil, goagrpc.ErrInvalidType("ServiceMessageWithSecurity", "MethodMessageWithSecurity", "*service_message_with_securitypb.MethodMessageWithSecurityRequest", v)
		}
	}
	var payload *servicemessagewithsecurity.RequestUT
	{
		payload = NewMethodMessageWithSecurityPayload(message, token, key, username, password)
		if payload.Token != nil {
			if strings.Contains(*payload.Token, " ") {
				// Remove authorization scheme prefix (e.g. "Bearer")
				cred := strings.SplitN(*payload.Token, " ", 2)[1]
				payload.Token = &cred
			}
		}
		if payload.Key != nil {
			if strings.Contains(*payload.Key, " ") {
				// Remove authorization scheme prefix (e.g. "Bearer")
				cred := strings.SplitN(*payload.Key, " ", 2)[1]
				payload.Key = &cred
			}
		}
	}
	return payload, nil
}
`
