package testdata

const ResultWithViewsResponseEncoderCode = `// EncodeMethodMessageUserTypeWithNestedUserTypesResponse encodes responses
// from the "ServiceMessageUserTypeWithNestedUserTypes" service
// "MethodMessageUserTypeWithNestedUserTypes" endpoint.
func EncodeMethodMessageUserTypeWithNestedUserTypesResponse(ctx context.Context, v interface{}, hdr, trlr *metadata.MD) (interface{}, error) {
	vres, ok := v.(*servicemessageusertypewithnestedusertypesviews.RecursiveT)
	if !ok {
		return nil, goagrpc.ErrInvalidType("ServiceMessageUserTypeWithNestedUserTypes", "MethodMessageUserTypeWithNestedUserTypes", "*servicemessageusertypewithnestedusertypesviews.RecursiveT", v)
	}
	result := vres.Projected
	(*hdr).Append("goa-view", vres.View)
	resp := NewMethodMessageUserTypeWithNestedUserTypesResponse(result)
	return resp, nil
}
`

const ResultArrayResponseEncoderCode = `// EncodeMethodMessageArrayResponse encodes responses from the
// "ServiceMessageArray" service "MethodMessageArray" endpoint.
func EncodeMethodMessageArrayResponse(ctx context.Context, v interface{}, hdr, trlr *metadata.MD) (interface{}, error) {
	result, ok := v.([]*servicemessagearray.UT)
	if !ok {
		return nil, goagrpc.ErrInvalidType("ServiceMessageArray", "MethodMessageArray", "[]*servicemessagearray.UT", v)
	}
	resp := NewMethodMessageArrayResponse(result)
	return resp, nil
}
`

const ResultPrimitiveResponseEncoderCode = `// EncodeMethodUnaryRPCNoPayloadResponse encodes responses from the
// "ServiceUnaryRPCNoPayload" service "MethodUnaryRPCNoPayload" endpoint.
func EncodeMethodUnaryRPCNoPayloadResponse(ctx context.Context, v interface{}, hdr, trlr *metadata.MD) (interface{}, error) {
	result, ok := v.(string)
	if !ok {
		return nil, goagrpc.ErrInvalidType("ServiceUnaryRPCNoPayload", "MethodUnaryRPCNoPayload", "string", v)
	}
	resp := NewMethodUnaryRPCNoPayloadResponse(result)
	return resp, nil
}
`

const ResultWithMetadataResponseEncoderCode = `// EncodeMethodMessageWithMetadataResponse encodes responses from the
// "ServiceMessageWithMetadata" service "MethodMessageWithMetadata" endpoint.
func EncodeMethodMessageWithMetadataResponse(ctx context.Context, v interface{}, hdr, trlr *metadata.MD) (interface{}, error) {
	result, ok := v.(*servicemessagewithmetadata.ResponseUT)
	if !ok {
		return nil, goagrpc.ErrInvalidType("ServiceMessageWithMetadata", "MethodMessageWithMetadata", "*servicemessagewithmetadata.ResponseUT", v)
	}
	resp := NewMethodMessageWithMetadataResponse(result)

	if res.InHeader != nil {
		(*hdr).Append("Location", fmt.Sprintf("%v", *p.InHeader))
	}

	if res.InTrailer != nil {
		(*trlr).Append("InTrailer", fmt.Sprintf("%v", *p.InTrailer))
	}
	return resp, nil
}
`

const ResultCollectionResponseEncoderCode = `// EncodeMethodMessageUserTypeWithNestedUserTypesResponse encodes responses
// from the "ServiceMessageUserTypeWithNestedUserTypes" service
// "MethodMessageUserTypeWithNestedUserTypes" endpoint.
func EncodeMethodMessageUserTypeWithNestedUserTypesResponse(ctx context.Context, v interface{}, hdr, trlr *metadata.MD) (interface{}, error) {
	vres, ok := v.(servicemessageusertypewithnestedusertypesviews.RTCollection)
	if !ok {
		return nil, goagrpc.ErrInvalidType("ServiceMessageUserTypeWithNestedUserTypes", "MethodMessageUserTypeWithNestedUserTypes", "servicemessageusertypewithnestedusertypesviews.RTCollection", v)
	}
	result := vres.Projected
	(*hdr).Append("goa-view", vres.View)
	resp := NewRTCollection(result)
	return resp, nil
}
`
