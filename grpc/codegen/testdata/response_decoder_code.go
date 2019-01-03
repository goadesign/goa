package testdata

const ResultWithViewsResponseDecoderCode = `// DecodeMethodMessageUserTypeWithNestedUserTypesResponse decodes responses
// from the ServiceMessageUserTypeWithNestedUserTypes
// MethodMessageUserTypeWithNestedUserTypes endpoint.
func DecodeMethodMessageUserTypeWithNestedUserTypesResponse(ctx context.Context, v interface{}, hdr, trlr metadata.MD) (interface{}, error) {
	var view string
	{
		if vals := hdr.Get("goa-view"); len(vals) > 0 {
			view = vals[0]
		}
	}
	message, ok := v.(*pb.MethodMessageUserTypeWithNestedUserTypesResponse)
	if !ok {
		return nil, goagrpc.ErrInvalidType("ServiceMessageUserTypeWithNestedUserTypes", "MethodMessageUserTypeWithNestedUserTypes", "*pb.MethodMessageUserTypeWithNestedUserTypesResponse", v)
	}
	res := NewRecursiveTView(message)
	vres := &servicemessageusertypewithnestedusertypesviews.RecursiveT{Projected: res}
	vres.View = view
	return servicemessageusertypewithnestedusertypes.NewRecursiveT(vres), nil
}
`

const ResultArrayResponseDecoderCode = `// DecodeMethodMessageArrayResponse decodes responses from the
// ServiceMessageArray MethodMessageArray endpoint.
func DecodeMethodMessageArrayResponse(ctx context.Context, v interface{}, hdr, trlr metadata.MD) (interface{}, error) {
	message, ok := v.(*pb.MethodMessageArrayResponse)
	if !ok {
		return nil, goagrpc.ErrInvalidType("ServiceMessageArray", "MethodMessageArray", "*pb.MethodMessageArrayResponse", v)
	}
	res := NewMethodMessageArrayResponse(message)
	return res, nil
}
`

const ResultPrimitiveResponseDecoderCode = `// DecodeMethodUnaryRPCNoPayloadResponse decodes responses from the
// ServiceUnaryRPCNoPayload MethodUnaryRPCNoPayload endpoint.
func DecodeMethodUnaryRPCNoPayloadResponse(ctx context.Context, v interface{}, hdr, trlr metadata.MD) (interface{}, error) {
	message, ok := v.(*pb.MethodUnaryRPCNoPayloadResponse)
	if !ok {
		return nil, goagrpc.ErrInvalidType("ServiceUnaryRPCNoPayload", "MethodUnaryRPCNoPayload", "*pb.MethodUnaryRPCNoPayloadResponse", v)
	}
	res := NewMethodUnaryRPCNoPayloadResponse(message)
	return res, nil
}
`

const ResultWithMetadataResponseDecoderCode = `// DecodeMethodMessageWithMetadataResponse decodes responses from the
// ServiceMessageWithMetadata MethodMessageWithMetadata endpoint.
func DecodeMethodMessageWithMetadataResponse(ctx context.Context, v interface{}, hdr, trlr metadata.MD) (interface{}, error) {
	var (
		inHeader  int
		inTrailer bool
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
	message, ok := v.(*pb.MethodMessageWithMetadataResponse)
	if !ok {
		return nil, goagrpc.ErrInvalidType("ServiceMessageWithMetadata", "MethodMessageWithMetadata", "*pb.MethodMessageWithMetadataResponse", v)
	}
	res := NewResponseUT(message, inHeader, inTrailer)
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
	message, ok := v.(*pb.RTCollection)
	if !ok {
		return nil, goagrpc.ErrInvalidType("ServiceMessageUserTypeWithNestedUserTypes", "MethodMessageUserTypeWithNestedUserTypes", "*pb.RTCollection", v)
	}
	res := NewRTCollection(message)
	vres := servicemessageusertypewithnestedusertypesviews.RTCollection{Projected: res}
	vres.View = view
	return servicemessageusertypewithnestedusertypes.NewRTCollection(vres), nil
}
`
