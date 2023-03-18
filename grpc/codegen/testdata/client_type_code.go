package testdata

const PayloadWithNestedTypesClientTypeCode = `// NewProtoMethodPayloadWithNestedTypesRequest builds the gRPC request type
// from the payload of the "MethodPayloadWithNestedTypes" endpoint of the
// "ServicePayloadWithNestedTypes" service.
func NewProtoMethodPayloadWithNestedTypesRequest(payload *servicepayloadwithnestedtypes.MethodPayloadWithNestedTypesPayload) *service_payload_with_nested_typespb.MethodPayloadWithNestedTypesRequest {
	message := &service_payload_with_nested_typespb.MethodPayloadWithNestedTypesRequest{}
	if payload.AParams != nil {
		message.AParams = svcServicepayloadwithnestedtypesAParamsToServicePayloadWithNestedTypespbAParams(payload.AParams)
	}
	if payload.BParams != nil {
		message.BParams = svcServicepayloadwithnestedtypesBParamsToServicePayloadWithNestedTypespbBParams(payload.BParams)
	}
	return message
}

// protobufServicePayloadWithNestedTypespbAParamsToServicepayloadwithnestedtypesAParams
// builds a value of type *servicepayloadwithnestedtypes.AParams from a value
// of type *service_payload_with_nested_typespb.AParams.
func protobufServicePayloadWithNestedTypespbAParamsToServicepayloadwithnestedtypesAParams(v *service_payload_with_nested_typespb.AParams) *servicepayloadwithnestedtypes.AParams {
	if v == nil {
		return nil
	}
	res := &servicepayloadwithnestedtypes.AParams{}
	if v.A != nil {
		res.A = make(map[string][]string, len(v.A))
		for key, val := range v.A {
			tk := key
			tv := make([]string, len(val.Field))
			for i, val := range val.Field {
				tv[i] = val
			}
			res.A[tk] = tv
		}
	}

	return res
}

// protobufServicePayloadWithNestedTypespbBParamsToServicepayloadwithnestedtypesBParams
// builds a value of type *servicepayloadwithnestedtypes.BParams from a value
// of type *service_payload_with_nested_typespb.BParams.
func protobufServicePayloadWithNestedTypespbBParamsToServicepayloadwithnestedtypesBParams(v *service_payload_with_nested_typespb.BParams) *servicepayloadwithnestedtypes.BParams {
	if v == nil {
		return nil
	}
	res := &servicepayloadwithnestedtypes.BParams{}
	if v.B != nil {
		res.B = make(map[string]string, len(v.B))
		for key, val := range v.B {
			tk := key
			tv := val
			res.B[tk] = tv
		}
	}

	return res
}

// svcServicepayloadwithnestedtypesAParamsToServicePayloadWithNestedTypespbAParams
// builds a value of type *service_payload_with_nested_typespb.AParams from a
// value of type *servicepayloadwithnestedtypes.AParams.
func svcServicepayloadwithnestedtypesAParamsToServicePayloadWithNestedTypespbAParams(v *servicepayloadwithnestedtypes.AParams) *service_payload_with_nested_typespb.AParams {
	if v == nil {
		return nil
	}
	res := &service_payload_with_nested_typespb.AParams{}
	if v.A != nil {
		res.A = make(map[string]*service_payload_with_nested_typespb.ArrayOfString, len(v.A))
		for key, val := range v.A {
			tk := key
			tv := &service_payload_with_nested_typespb.ArrayOfString{}
			tv.Field = make([]string, len(val))
			for i, val := range val {
				tv.Field[i] = val
			}
			res.A[tk] = tv
		}
	}

	return res
}

// svcServicepayloadwithnestedtypesBParamsToServicePayloadWithNestedTypespbBParams
// builds a value of type *service_payload_with_nested_typespb.BParams from a
// value of type *servicepayloadwithnestedtypes.BParams.
func svcServicepayloadwithnestedtypesBParamsToServicePayloadWithNestedTypespbBParams(v *servicepayloadwithnestedtypes.BParams) *service_payload_with_nested_typespb.BParams {
	if v == nil {
		return nil
	}
	res := &service_payload_with_nested_typespb.BParams{}
	if v.B != nil {
		res.B = make(map[string]string, len(v.B))
		for key, val := range v.B {
			tk := key
			tv := val
			res.B[tk] = tv
		}
	}

	return res
}
`

const PayloadWithMultipleUseTypesClientTypeCode = `// NewProtoDupePayload builds the gRPC request type from the payload of the
// "MethodPayloadDuplicateA" endpoint of the "ServicePayloadWithNestedTypes"
// service.
func NewProtoDupePayload(payload servicepayloadwithnestedtypes.DupePayload) *service_payload_with_nested_typespb.DupePayload {
	message := &service_payload_with_nested_typespb.DupePayload{}
	message.Field = string(payload)
	return message
}
`

const PayloadWithAliasTypeClientTypeCode = `// NewProtoMethodMessageUserTypeWithAliasRequest builds the gRPC request type
// from the payload of the "MethodMessageUserTypeWithAlias" endpoint of the
// "ServiceMessageUserTypeWithAlias" service.
func NewProtoMethodMessageUserTypeWithAliasRequest(payload *servicemessageusertypewithalias.PayloadAliasT) *service_message_user_type_with_aliaspb.MethodMessageUserTypeWithAliasRequest {
	message := &service_message_user_type_with_aliaspb.MethodMessageUserTypeWithAliasRequest{
		IntAliasField: int32(payload.IntAliasField),
	}
	if payload.OptionalIntAliasField != nil {
		optionalIntAliasField := int32(*payload.OptionalIntAliasField)
		message.OptionalIntAliasField = &optionalIntAliasField
	}
	return message
}

// NewMethodMessageUserTypeWithAliasResult builds the result type of the
// "MethodMessageUserTypeWithAlias" endpoint of the
// "ServiceMessageUserTypeWithAlias" service from the gRPC response type.
func NewMethodMessageUserTypeWithAliasResult(message *service_message_user_type_with_aliaspb.MethodMessageUserTypeWithAliasResponse) *servicemessageusertypewithalias.PayloadAliasT {
	result := &servicemessageusertypewithalias.PayloadAliasT{
		IntAliasField: servicemessageusertypewithalias.IntAlias(message.IntAliasField),
	}
	if message.OptionalIntAliasField != nil {
		optionalIntAliasField := servicemessageusertypewithalias.IntAlias(*message.OptionalIntAliasField)
		result.OptionalIntAliasField = &optionalIntAliasField
	}
	return result
}
`

const ResultWithAliasValidationClientTypeCode = `// NewProtoMethodResultWithAliasValidationRequest builds the gRPC request type
// from the payload of the "MethodResultWithAliasValidation" endpoint of the
// "ServiceResultWithAliasValidation" service.
func NewProtoMethodResultWithAliasValidationRequest() *service_result_with_alias_validationpb.MethodResultWithAliasValidationRequest {
	message := &service_result_with_alias_validationpb.MethodResultWithAliasValidationRequest{}
	return message
}

// NewMethodResultWithAliasValidationResult builds the result type of the
// "MethodResultWithAliasValidation" endpoint of the
// "ServiceResultWithAliasValidation" service from the gRPC response type.
func NewMethodResultWithAliasValidationResult(message *service_result_with_alias_validationpb.UUID) serviceresultwithaliasvalidation.UUID {
	result := serviceresultwithaliasvalidation.UUID(message.Field)
	return result
}

// ValidateUUID runs the validations defined on UUID.
func ValidateUUID(message *service_result_with_alias_validationpb.UUID) (err error) {
	err = goa.MergeErrors(err, goa.ValidateFormat("message.field", message.Field, goa.FormatUUID))
	return
}
`

const ResultWithCollectionClientTypeCode = `// NewProtoMethodResultWithCollectionRequest builds the gRPC request type from
// the payload of the "MethodResultWithCollection" endpoint of the
// "ServiceResultWithCollection" service.
func NewProtoMethodResultWithCollectionRequest() *service_result_with_collectionpb.MethodResultWithCollectionRequest {
	message := &service_result_with_collectionpb.MethodResultWithCollectionRequest{}
	return message
}

// NewMethodResultWithCollectionResult builds the result type of the
// "MethodResultWithCollection" endpoint of the "ServiceResultWithCollection"
// service from the gRPC response type.
func NewMethodResultWithCollectionResult(message *service_result_with_collectionpb.MethodResultWithCollectionResponse) *serviceresultwithcollection.MethodResultWithCollectionResult {
	result := &serviceresultwithcollection.MethodResultWithCollectionResult{}
	if message.Result != nil {
		result.Result = protobufServiceResultWithCollectionpbResultTToServiceresultwithcollectionResultT(message.Result)
	}
	return result
}

// svcServiceresultwithcollectionResultTToServiceResultWithCollectionpbResultT
// builds a value of type *service_result_with_collectionpb.ResultT from a
// value of type *serviceresultwithcollection.ResultT.
func svcServiceresultwithcollectionResultTToServiceResultWithCollectionpbResultT(v *serviceresultwithcollection.ResultT) *service_result_with_collectionpb.ResultT {
	if v == nil {
		return nil
	}
	res := &service_result_with_collectionpb.ResultT{}
	if v.CollectionField != nil {
		res.CollectionField = &service_result_with_collectionpb.RTCollection{}
		res.CollectionField.Field = make([]*service_result_with_collectionpb.RT, len(v.CollectionField))
		for i, val := range v.CollectionField {
			res.CollectionField.Field[i] = &service_result_with_collectionpb.RT{}
			if val.IntField != nil {
				intField := int32(*val.IntField)
				res.CollectionField.Field[i].IntField = &intField
			}
		}
	}

	return res
}

// protobufServiceResultWithCollectionpbResultTToServiceresultwithcollectionResultT
// builds a value of type *serviceresultwithcollection.ResultT from a value of
// type *service_result_with_collectionpb.ResultT.
func protobufServiceResultWithCollectionpbResultTToServiceresultwithcollectionResultT(v *service_result_with_collectionpb.ResultT) *serviceresultwithcollection.ResultT {
	if v == nil {
		return nil
	}
	res := &serviceresultwithcollection.ResultT{}
	if v.CollectionField != nil {
		res.CollectionField = make([]*serviceresultwithcollection.RT, len(v.CollectionField.Field))
		for i, val := range v.CollectionField.Field {
			res.CollectionField[i] = &serviceresultwithcollection.RT{}
			if val.IntField != nil {
				intField := int(*val.IntField)
				res.CollectionField[i].IntField = &intField
			}
		}
	}

	return res
}
`

const WithErrorsClientTypeCode = `// NewProtoMethodUnaryRPCWithErrorsRequest builds the gRPC request type from
// the payload of the "MethodUnaryRPCWithErrors" endpoint of the
// "ServiceUnaryRPCWithErrors" service.
func NewProtoMethodUnaryRPCWithErrorsRequest(payload string) *service_unary_rpc_with_errorspb.MethodUnaryRPCWithErrorsRequest {
	message := &service_unary_rpc_with_errorspb.MethodUnaryRPCWithErrorsRequest{}
	message.Field = payload
	return message
}

// NewMethodUnaryRPCWithErrorsResult builds the result type of the
// "MethodUnaryRPCWithErrors" endpoint of the "ServiceUnaryRPCWithErrors"
// service from the gRPC response type.
func NewMethodUnaryRPCWithErrorsResult(message *service_unary_rpc_with_errorspb.MethodUnaryRPCWithErrorsResponse) string {
	result := message.Field
	return result
}

// NewMethodUnaryRPCWithErrorsInternalError builds the error type of the
// "MethodUnaryRPCWithErrors" endpoint of the "ServiceUnaryRPCWithErrors"
// service from the gRPC error response type.
func NewMethodUnaryRPCWithErrorsInternalError(message *service_unary_rpc_with_errorspb.MethodUnaryRPCWithErrorsInternalError) *serviceunaryrpcwitherrors.AnotherError {
	er := &serviceunaryrpcwitherrors.AnotherError{
		Name:        message.Name,
		Description: message.Description,
	}
	return er
}

// NewMethodUnaryRPCWithErrorsBadRequestError builds the error type of the
// "MethodUnaryRPCWithErrors" endpoint of the "ServiceUnaryRPCWithErrors"
// service from the gRPC error response type.
func NewMethodUnaryRPCWithErrorsBadRequestError(message *service_unary_rpc_with_errorspb.MethodUnaryRPCWithErrorsBadRequestError) *serviceunaryrpcwitherrors.AnotherError {
	er := &serviceunaryrpcwitherrors.AnotherError{
		Name:        message.Name,
		Description: message.Description,
	}
	return er
}

// NewMethodUnaryRPCWithErrorsCustomErrorError builds the error type of the
// "MethodUnaryRPCWithErrors" endpoint of the "ServiceUnaryRPCWithErrors"
// service from the gRPC error response type.
func NewMethodUnaryRPCWithErrorsCustomErrorError(message *service_unary_rpc_with_errorspb.MethodUnaryRPCWithErrorsCustomErrorError) *serviceunaryrpcwitherrors.ErrorType {
	er := &serviceunaryrpcwitherrors.ErrorType{
		A: message.A,
	}
	return er
}

// ValidateMethodUnaryRPCWithErrorsInternalError runs the validations defined
// on MethodUnaryRPCWithErrorsInternalError.
func ValidateMethodUnaryRPCWithErrorsInternalError(errmsg *service_unary_rpc_with_errorspb.MethodUnaryRPCWithErrorsInternalError) (err error) {
	if !(errmsg.Name == "this" || errmsg.Name == "that") {
		err = goa.MergeErrors(err, goa.InvalidEnumValueError("errmsg.name", errmsg.Name, []any{"this", "that"}))
	}
	return
}

// ValidateMethodUnaryRPCWithErrorsBadRequestError runs the validations defined
// on MethodUnaryRPCWithErrorsBadRequestError.
func ValidateMethodUnaryRPCWithErrorsBadRequestError(errmsg *service_unary_rpc_with_errorspb.MethodUnaryRPCWithErrorsBadRequestError) (err error) {
	if !(errmsg.Name == "this" || errmsg.Name == "that") {
		err = goa.MergeErrors(err, goa.InvalidEnumValueError("errmsg.name", errmsg.Name, []any{"this", "that"}))
	}
	return
}
`

const BidirectionalStreamingRPCSameTypeClientTypeCode = `func NewMethodBidirectionalStreamingRPCSameTypeResponseUserType(v *service_bidirectional_streaming_rpc_same_typepb.MethodBidirectionalStreamingRPCSameTypeResponse) *servicebidirectionalstreamingrpcsametype.UserType {
	result := &servicebidirectionalstreamingrpcsametype.UserType{
		B: v.B,
	}
	if v.A != nil {
		a := int(*v.A)
		result.A = &a
	}
	return result
}

func NewProtoUserTypeMethodBidirectionalStreamingRPCSameTypeStreamingRequest(spayload *servicebidirectionalstreamingrpcsametype.UserType) *service_bidirectional_streaming_rpc_same_typepb.MethodBidirectionalStreamingRPCSameTypeStreamingRequest {
	v := &service_bidirectional_streaming_rpc_same_typepb.MethodBidirectionalStreamingRPCSameTypeStreamingRequest{
		B: spayload.B,
	}
	if spayload.A != nil {
		a := int32(*spayload.A)
		v.A = &a
	}
	return v
}
`

const StructMetaTypeTypeCode = `// NewProtoMethodRequest builds the gRPC request type from the payload of the
// "Method" endpoint of the "UsingMetaTypes" service.
func NewProtoMethodRequest(payload *usingmetatypes.MethodPayload) *using_meta_typespb.MethodRequest {
	message := &using_meta_typespb.MethodRequest{}
	a := int64(payload.A)
	message.A = &a
	b := int64(payload.B)
	message.B = &b
	if payload.D != nil {
		d := int64(*payload.D)
		message.D = &d
	}
	if payload.C != nil {
		message.C = make([]int64, len(payload.C))
		for i, val := range payload.C {
			message.C[i] = int64(val)
		}
	}
	return message
}

// NewMethodResult builds the result type of the "Method" endpoint of the
// "UsingMetaTypes" service from the gRPC response type.
func NewMethodResult(message *using_meta_typespb.MethodResponse) *usingmetatypes.MethodResult {
	result := &usingmetatypes.MethodResult{}
	if message.A != nil {
		result.A = flag.ErrorHandling(*message.A)
	}
	if message.B != nil {
		result.B = flag.ErrorHandling(*message.B)
	}
	if message.D != nil {
		d := flag.ErrorHandling(*message.D)
		result.D = &d
	}
	if message.A == nil {
		result.A = 1
	}
	if message.B == nil {
		result.B = 2
	}
	if message.C != nil {
		result.C = make([]time.Duration, len(message.C))
		for i, val := range message.C {
			result.C[i] = time.Duration(val)
		}
	}
	return result
}
`

const DefaultFieldsTypeCode = `// NewProtoMethodRequest builds the gRPC request type from the payload of the
// "Method" endpoint of the "DefaultFields" service.
func NewProtoMethodRequest(payload *defaultfields.MethodPayload) *default_fieldspb.MethodRequest {
	message := &default_fieldspb.MethodRequest{
		Req:  payload.Req,
		Opt:  payload.Opt,
		Def0: &payload.Def0,
		Def1: &payload.Def1,
		Def2: &payload.Def2,
		Reqs: payload.Reqs,
		Opts: payload.Opts,
		Defs: &payload.Defs,
		Defe: &payload.Defe,
		Rat:  payload.Rat,
		Flt:  payload.Flt,
		Flt0: &payload.Flt0,
		Flt1: &payload.Flt1,
	}
	return message
}
`
