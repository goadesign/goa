package testdata

const PayloadUserTypeRequestEncoderCode = `// EncodeMethodMessageUserTypeWithNestedUserTypesRequest encodes requests sent
// to ServiceMessageUserTypeWithNestedUserTypes
// MethodMessageUserTypeWithNestedUserTypes endpoint.
func EncodeMethodMessageUserTypeWithNestedUserTypesRequest(ctx context.Context, v interface{}, md *metadata.MD) (interface{}, error) {
	payload, ok := v.(*servicemessageusertypewithnestedusertypes.UT)
	if !ok {
		return nil, goagrpc.ErrInvalidType("ServiceMessageUserTypeWithNestedUserTypes", "MethodMessageUserTypeWithNestedUserTypes", "*servicemessageusertypewithnestedusertypes.UT", v)
	}
	return NewMethodMessageUserTypeWithNestedUserTypesRequest(payload), nil
}
`

const PayloadArrayRequestEncoderCode = `// EncodeMethodUnaryRPCNoResultRequest encodes requests sent to
// ServiceUnaryRPCNoResult MethodUnaryRPCNoResult endpoint.
func EncodeMethodUnaryRPCNoResultRequest(ctx context.Context, v interface{}, md *metadata.MD) (interface{}, error) {
	payload, ok := v.([]string)
	if !ok {
		return nil, goagrpc.ErrInvalidType("ServiceUnaryRPCNoResult", "MethodUnaryRPCNoResult", "[]string", v)
	}
	return NewMethodUnaryRPCNoResultRequest(payload), nil
}
`

const PayloadMapRequestEncoderCode = `// EncodeMethodMessageMapRequest encodes requests sent to ServiceMessageMap
// MethodMessageMap endpoint.
func EncodeMethodMessageMapRequest(ctx context.Context, v interface{}, md *metadata.MD) (interface{}, error) {
	payload, ok := v.(map[int]*servicemessagemap.UT)
	if !ok {
		return nil, goagrpc.ErrInvalidType("ServiceMessageMap", "MethodMessageMap", "map[int]*servicemessagemap.UT", v)
	}
	return NewMethodMessageMapRequest(payload), nil
}
`

const PayloadPrimitiveRequestEncoderCode = `// EncodeMethodServerStreamingRPCRequest encodes requests sent to
// ServiceServerStreamingRPC MethodServerStreamingRPC endpoint.
func EncodeMethodServerStreamingRPCRequest(ctx context.Context, v interface{}, md *metadata.MD) (interface{}, error) {
	payload, ok := v.(int)
	if !ok {
		return nil, goagrpc.ErrInvalidType("ServiceServerStreamingRPC", "MethodServerStreamingRPC", "int", v)
	}
	return NewMethodServerStreamingRPCRequest(payload), nil
}
`

const PayloadPrimitiveWithStreamingPayloadRequestEncoderCode = `// EncodeMethodClientStreamingRPCWithPayloadRequest encodes requests sent to
// ServiceClientStreamingRPCWithPayload MethodClientStreamingRPCWithPayload
// endpoint.
func EncodeMethodClientStreamingRPCWithPayloadRequest(ctx context.Context, v interface{}, md *metadata.MD) (interface{}, error) {
	payload, ok := v.(int)
	if !ok {
		return nil, goagrpc.ErrInvalidType("ServiceClientStreamingRPCWithPayload", "MethodClientStreamingRPCWithPayload", "int", v)
	}
	(*md).Append("goa_payload", fmt.Sprintf("%v", payload))
	return nil, nil
}
`

const PayloadUserTypeWithStreamingPayloadRequestEncoderCode = `// EncodeMethodBidirectionalStreamingRPCWithPayloadRequest encodes requests
// sent to ServiceBidirectionalStreamingRPCWithPayload
// MethodBidirectionalStreamingRPCWithPayload endpoint.
func EncodeMethodBidirectionalStreamingRPCWithPayloadRequest(ctx context.Context, v interface{}, md *metadata.MD) (interface{}, error) {
	payload, ok := v.(*servicebidirectionalstreamingrpcwithpayload.Payload)
	if !ok {
		return nil, goagrpc.ErrInvalidType("ServiceBidirectionalStreamingRPCWithPayload", "MethodBidirectionalStreamingRPCWithPayload", "*servicebidirectionalstreamingrpcwithpayload.Payload", v)
	}
	if payload.A != nil {
		(*md).Append("a", fmt.Sprintf("%v", *payload.A))
	}
	if payload.B != nil {
		(*md).Append("b", *payload.B)
	}
	return nil, nil
}
`

const PayloadWithMetadataRequestEncoderCode = `// EncodeMethodMessageWithMetadataRequest encodes requests sent to
// ServiceMessageWithMetadata MethodMessageWithMetadata endpoint.
func EncodeMethodMessageWithMetadataRequest(ctx context.Context, v interface{}, md *metadata.MD) (interface{}, error) {
	payload, ok := v.(*servicemessagewithmetadata.RequestUT)
	if !ok {
		return nil, goagrpc.ErrInvalidType("ServiceMessageWithMetadata", "MethodMessageWithMetadata", "*servicemessagewithmetadata.RequestUT", v)
	}
	if payload.InMetadata != nil {
		(*md).Append("Authorization", fmt.Sprintf("%v", *payload.InMetadata))
	}
	return NewMethodMessageWithMetadataRequest(payload), nil
}
`

const PayloadWithSecurityAttrsRequestEncoderCode = `// EncodeMethodMessageWithSecurityRequest encodes requests sent to
// ServiceMessageWithSecurity MethodMessageWithSecurity endpoint.
func EncodeMethodMessageWithSecurityRequest(ctx context.Context, v interface{}, md *metadata.MD) (interface{}, error) {
	payload, ok := v.(*servicemessagewithsecurity.RequestUT)
	if !ok {
		return nil, goagrpc.ErrInvalidType("ServiceMessageWithSecurity", "MethodMessageWithSecurity", "*servicemessagewithsecurity.RequestUT", v)
	}
	if payload.Token != nil {
		(*md).Append("authorization", *payload.Token)
	}
	if payload.Key != nil {
		(*md).Append("authorization", *payload.Key)
	}
	if payload.Username != nil {
		(*md).Append("username", *payload.Username)
	}
	if payload.Password != nil {
		(*md).Append("password", *payload.Password)
	}
	return NewMethodMessageWithSecurityRequest(payload), nil
}
`
