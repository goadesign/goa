package testdata

const PayloadUserTypeRequestDecoderCode = `// DecodeMethodMessageUserTypeWithNestedUserTypesRequest decodes requests sent
// to "ServiceMessageUserTypeWithNestedUserTypes" service
// "MethodMessageUserTypeWithNestedUserTypes" endpoint.
func DecodeMethodMessageUserTypeWithNestedUserTypesRequest(ctx context.Context, v interface{}, md metadata.MD) (interface{}, error) {
	var (
		message *pb.MethodMessageUserTypeWithNestedUserTypesRequest
		ok      bool
		err     error
	)
	{
		if message, ok = v.(*pb.MethodMessageUserTypeWithNestedUserTypesRequest); !ok {
			return nil, goagrpc.ErrInvalidType("ServiceMessageUserTypeWithNestedUserTypes", "MethodMessageUserTypeWithNestedUserTypes", "*pb.MethodMessageUserTypeWithNestedUserTypesRequest", v)
		}
	}
	var (
		payload *servicemessageusertypewithnestedusertypes.UT
	)
	{
		payload = NewUT(message)
	}
	return payload, err
}
`

const PayloadArrayRequestDecoderCode = `// DecodeMethodUnaryRPCNoResultRequest decodes requests sent to
// "ServiceUnaryRPCNoResult" service "MethodUnaryRPCNoResult" endpoint.
func DecodeMethodUnaryRPCNoResultRequest(ctx context.Context, v interface{}, md metadata.MD) (interface{}, error) {
	var (
		message *pb.MethodUnaryRPCNoResultRequest
		ok      bool
		err     error
	)
	{
		if message, ok = v.(*pb.MethodUnaryRPCNoResultRequest); !ok {
			return nil, goagrpc.ErrInvalidType("ServiceUnaryRPCNoResult", "MethodUnaryRPCNoResult", "*pb.MethodUnaryRPCNoResultRequest", v)
		}
	}
	var (
		payload []string
	)
	{
		payload = NewMethodUnaryRPCNoResultRequest(message)
	}
	return payload, err
}
`

const PayloadMapRequestDecoderCode = `// DecodeMethodMessageMapRequest decodes requests sent to "ServiceMessageMap"
// service "MethodMessageMap" endpoint.
func DecodeMethodMessageMapRequest(ctx context.Context, v interface{}, md metadata.MD) (interface{}, error) {
	var (
		message *pb.MethodMessageMapRequest
		ok      bool
		err     error
	)
	{
		if message, ok = v.(*pb.MethodMessageMapRequest); !ok {
			return nil, goagrpc.ErrInvalidType("ServiceMessageMap", "MethodMessageMap", "*pb.MethodMessageMapRequest", v)
		}
	}
	var (
		payload map[int]*servicemessagemap.UT
	)
	{
		payload = NewMethodMessageMapRequest(message)
	}
	return payload, err
}
`

const PayloadPrimitiveRequestDecoderCode = `// DecodeMethodServerStreamingRPCRequest decodes requests sent to
// "ServiceServerStreamingRPC" service "MethodServerStreamingRPC" endpoint.
func DecodeMethodServerStreamingRPCRequest(ctx context.Context, v interface{}, md metadata.MD) (interface{}, error) {
	var (
		message *pb.MethodServerStreamingRPCRequest
		ok      bool
		err     error
	)
	{
		if message, ok = v.(*pb.MethodServerStreamingRPCRequest); !ok {
			return nil, goagrpc.ErrInvalidType("ServiceServerStreamingRPC", "MethodServerStreamingRPC", "*pb.MethodServerStreamingRPCRequest", v)
		}
	}
	var (
		payload int
	)
	{
		payload = NewMethodServerStreamingRPCRequest(message)
	}
	return payload, err
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
			goaPayloadRaw = vals[0]

			v, err2 := strconv.ParseInt(goaPayloadRaw, 10, strconv.IntSize)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError("goaPayload", goaPayloadRaw, "integer"))
			}
			goaPayload = int(v)
		}
	}
	var (
		payload int
	)
	{
		payload = goaPayload
	}
	return payload, err
}
`

const PayloadUserTypeWithStreamingPayloadRequestDecoderCode = `// DecodeMethodBidirectionalStreamingRPCWithPayloadRequest decodes requests
// sent to "ServiceBidirectionalStreamingRPCWithPayload" service
// "MethodBidirectionalStreamingRPCWithPayload" endpoint.
func DecodeMethodBidirectionalStreamingRPCWithPayloadRequest(ctx context.Context, v interface{}, md metadata.MD) (interface{}, error) {
	var (
		a   int
		b   string
		err error
	)
	{
		if vals := md.Get("a"); len(vals) > 0 {
			aRaw = vals[0]

			v, err2 := strconv.ParseInt(aRaw, 10, strconv.IntSize)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError("a", aRaw, "integer"))
			}
			pv := int(v)
			a = &pv
		}
		if vals := md.Get("b"); len(vals) > 0 {
			b = vals[0]
		}
	}
	var (
		payload *servicebidirectionalstreamingrpcwithpayload.Payload
	)
	{
		payload = NewPayload(a, b)
	}
	return payload, err
}
`

const PayloadWithMetadataRequestDecoderCode = `// DecodeMethodMessageWithMetadataRequest decodes requests sent to
// "ServiceMessageWithMetadata" service "MethodMessageWithMetadata" endpoint.
func DecodeMethodMessageWithMetadataRequest(ctx context.Context, v interface{}, md metadata.MD) (interface{}, error) {
	var (
		inMetadata int
		err        error
	)
	{
		if vals := md.Get("Authorization"); len(vals) > 0 {
			inMetadataRaw = vals[0]

			v, err2 := strconv.ParseInt(inMetadataRaw, 10, strconv.IntSize)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError("inMetadata", inMetadataRaw, "integer"))
			}
			pv := int(v)
			inMetadata = &pv
		}
	}
	var (
		message *pb.MethodMessageWithMetadataRequest
		ok      bool
	)
	{
		if message, ok = v.(*pb.MethodMessageWithMetadataRequest); !ok {
			return nil, goagrpc.ErrInvalidType("ServiceMessageWithMetadata", "MethodMessageWithMetadata", "*pb.MethodMessageWithMetadataRequest", v)
		}
	}
	var (
		payload *servicemessagewithmetadata.RequestUT
	)
	{
		payload = NewRequestUT(message, inMetadata)
	}
	return payload, err
}
`

const PayloadWithSecurityAttrsRequestDecoderCode = `// DecodeMethodMessageWithSecurityRequest decodes requests sent to
// "ServiceMessageWithSecurity" service "MethodMessageWithSecurity" endpoint.
func DecodeMethodMessageWithSecurityRequest(ctx context.Context, v interface{}, md metadata.MD) (interface{}, error) {
	var (
		token    string
		key      string
		username string
		password string
		err      error
	)
	{
		if vals := md.Get("authorization"); len(vals) > 0 {
			token = vals[0]
		}
		if vals := md.Get("authorization"); len(vals) > 0 {
			key = vals[0]
		}
		if vals := md.Get("username"); len(vals) > 0 {
			username = vals[0]
		}
		if vals := md.Get("password"); len(vals) > 0 {
			password = vals[0]
		}
	}
	var (
		message *pb.MethodMessageWithSecurityRequest
		ok      bool
	)
	{
		if message, ok = v.(*pb.MethodMessageWithSecurityRequest); !ok {
			return nil, goagrpc.ErrInvalidType("ServiceMessageWithSecurity", "MethodMessageWithSecurity", "*pb.MethodMessageWithSecurityRequest", v)
		}
	}
	var (
		payload *servicemessagewithsecurity.RequestUT
	)
	{
		payload = NewRequestUT(message, token, key, username, password)
		if payload.Token != nil {
			if strings.Contains(*payload.Token, " ") {
				// Remove authorization scheme prefix (e.g. "Bearer")
				cred := strings.SplitN(*payload.Token, " ", 2)[1]
				payload.Token = &cred
			}
		}
	}
	return payload, err
}
`
