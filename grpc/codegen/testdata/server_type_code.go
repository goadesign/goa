package testdata

const PayloadWithNestedTypesServerTypeCode = `// NewMethodPayloadWithNestedTypesPayload builds the payload of the
// "MethodPayloadWithNestedTypes" endpoint of the
// "ServicePayloadWithNestedTypes" service from the gRPC request type.
func NewMethodPayloadWithNestedTypesPayload(message *service_payload_with_nested_typespb.MethodPayloadWithNestedTypesRequest) *servicepayloadwithnestedtypes.MethodPayloadWithNestedTypesPayload {
	v := &servicepayloadwithnestedtypes.MethodPayloadWithNestedTypesPayload{}
	if message.AParams != nil {
		v.AParams = protobufServicePayloadWithNestedTypespbAParamsToServicepayloadwithnestedtypesAParams(message.AParams)
	}
	if message.BParams != nil {
		v.BParams = protobufServicePayloadWithNestedTypespbBParamsToServicepayloadwithnestedtypesBParams(message.BParams)
	}
	return v
}

// NewMethodPayloadWithNestedTypesResponse builds the gRPC response type from
// the result of the "MethodPayloadWithNestedTypes" endpoint of the
// "ServicePayloadWithNestedTypes" service.
func NewMethodPayloadWithNestedTypesResponse() *service_payload_with_nested_typespb.MethodPayloadWithNestedTypesResponse {
	message := &service_payload_with_nested_typespb.MethodPayloadWithNestedTypesResponse{}
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

const PayloadWithMultipleUseTypesServerTypeCode = `// NewMethodPayloadDuplicateAPayload builds the payload of the
// "MethodPayloadDuplicateA" endpoint of the "ServicePayloadWithNestedTypes"
// service from the gRPC request type.
func NewMethodPayloadDuplicateAPayload(message *service_payload_with_nested_typespb.DupePayload) servicepayloadwithnestedtypes.DupePayload {
	v := servicepayloadwithnestedtypes.DupePayload(message.Field)
	return v
}

// NewMethodPayloadDuplicateAResponse builds the gRPC response type from the
// result of the "MethodPayloadDuplicateA" endpoint of the
// "ServicePayloadWithNestedTypes" service.
func NewMethodPayloadDuplicateAResponse() *service_payload_with_nested_typespb.MethodPayloadDuplicateAResponse {
	message := &service_payload_with_nested_typespb.MethodPayloadDuplicateAResponse{}
	return message
}

// NewMethodPayloadDuplicateBPayload builds the payload of the
// "MethodPayloadDuplicateB" endpoint of the "ServicePayloadWithNestedTypes"
// service from the gRPC request type.
func NewMethodPayloadDuplicateBPayload(message *service_payload_with_nested_typespb.DupePayload) servicepayloadwithnestedtypes.DupePayload {
	v := servicepayloadwithnestedtypes.DupePayload(message.Field)
	return v
}

// NewMethodPayloadDuplicateBResponse builds the gRPC response type from the
// result of the "MethodPayloadDuplicateB" endpoint of the
// "ServicePayloadWithNestedTypes" service.
func NewMethodPayloadDuplicateBResponse() *service_payload_with_nested_typespb.MethodPayloadDuplicateBResponse {
	message := &service_payload_with_nested_typespb.MethodPayloadDuplicateBResponse{}
	return message
}
`

const PayloadWithAliasTypeServerTypeCode = `// NewMethodMessageUserTypeWithAliasPayload builds the payload of the
// "MethodMessageUserTypeWithAlias" endpoint of the
// "ServiceMessageUserTypeWithAlias" service from the gRPC request type.
func NewMethodMessageUserTypeWithAliasPayload(message *service_message_user_type_with_aliaspb.MethodMessageUserTypeWithAliasRequest) *servicemessageusertypewithalias.PayloadAliasT {
	v := &servicemessageusertypewithalias.PayloadAliasT{
		IntAliasField: servicemessageusertypewithalias.IntAlias(message.IntAliasField),
	}
	if message.OptionalIntAliasField != nil {
		optionalIntAliasFieldptr := servicemessageusertypewithalias.IntAlias(message.OptionalIntAliasField)
		v.OptionalIntAliasField = &optionalIntAliasFieldptr
	}
	return v
}

// NewMethodMessageUserTypeWithAliasResponse builds the gRPC response type from
// the result of the "MethodMessageUserTypeWithAlias" endpoint of the
// "ServiceMessageUserTypeWithAlias" service.
func NewMethodMessageUserTypeWithAliasResponse(result *servicemessageusertypewithalias.PayloadAliasT) *service_message_user_type_with_aliaspb.MethodMessageUserTypeWithAliasResponse {
	message := &service_message_user_type_with_aliaspb.MethodMessageUserTypeWithAliasResponse{
		IntAliasField: int(result.IntAliasField),
	}
	if result.OptionalIntAliasField != nil {
		message.OptionalIntAliasField = int(*result.OptionalIntAliasField)
	}
	return message
}
`

const ResultWithCollectionServerTypeCode = `// NewMethodResultWithCollectionResponse builds the gRPC response type from the
// result of the "MethodResultWithCollection" endpoint of the
// "ServiceResultWithCollection" service.
func NewMethodResultWithCollectionResponse(result *serviceresultwithcollection.MethodResultWithCollectionResult) *service_result_with_collectionpb.MethodResultWithCollectionResponse {
	message := &service_result_with_collectionpb.MethodResultWithCollectionResponse{}
	if result.Result != nil {
		message.Result = svcServiceresultwithcollectionResultTToServiceResultWithCollectionpbResultT(result.Result)
	}
	return message
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
				res.CollectionField.Field[i].IntField = int32(*val.IntField)
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
			if val.IntField != 0 {
				intFieldptr := int(val.IntField)
				res.CollectionField[i].IntField = &intFieldptr
			}
		}
	}

	return res
}
`

const PayloadWithMixedAttributesServerTypeCode = `// NewUnaryMethodPayload builds the payload of the "UnaryMethod" endpoint of
// the "ServicePayloadWithMixedAttributes" service from the gRPC request type.
func NewUnaryMethodPayload(message *service_payload_with_mixed_attributespb.UnaryMethodRequest) *servicepayloadwithmixedattributes.APayload {
	v := &servicepayloadwithmixedattributes.APayload{
		Required:        int(message.Required),
		Default:         int(message.Default),
		RequiredDefault: int(message.RequiredDefault),
	}
	if message.Optional != 0 {
		optionalptr := int(message.Optional)
		v.Optional = &optionalptr
	}
	if message.Default == 0 {
		v.Default = 100
	}
	return v
}

// NewUnaryMethodResponse builds the gRPC response type from the result of the
// "UnaryMethod" endpoint of the "ServicePayloadWithMixedAttributes" service.
func NewUnaryMethodResponse() *service_payload_with_mixed_attributespb.UnaryMethodResponse {
	message := &service_payload_with_mixed_attributespb.UnaryMethodResponse{}
	return message
}

// NewStreamingMethodResponse builds the gRPC response type from the result of
// the "StreamingMethod" endpoint of the "ServicePayloadWithMixedAttributes"
// service.
func NewStreamingMethodResponse() *service_payload_with_mixed_attributespb.StreamingMethodResponse {
	message := &service_payload_with_mixed_attributespb.StreamingMethodResponse{}
	return message
}

func NewAPayload(v *service_payload_with_mixed_attributespb.StreamingMethodStreamingRequest) *servicepayloadwithmixedattributes.APayload {
	spayload := &servicepayloadwithmixedattributes.APayload{
		Required:        int(v.Required),
		Default:         int(v.Default),
		RequiredDefault: int(v.RequiredDefault),
	}
	if v.Optional != 0 {
		optionalptr := int(v.Optional)
		spayload.Optional = &optionalptr
	}
	if v.Default == 0 {
		spayload.Default = 100
	}
	return spayload
}
`

const WithErrorsServerTypeCode = `// NewMethodUnaryRPCWithErrorsPayload builds the payload of the
// "MethodUnaryRPCWithErrors" endpoint of the "ServiceUnaryRPCWithErrors"
// service from the gRPC request type.
func NewMethodUnaryRPCWithErrorsPayload(message *service_unary_rpc_with_errorspb.MethodUnaryRPCWithErrorsRequest) string {
	v := message.Field
	return v
}

// NewMethodUnaryRPCWithErrorsResponse builds the gRPC response type from the
// result of the "MethodUnaryRPCWithErrors" endpoint of the
// "ServiceUnaryRPCWithErrors" service.
func NewMethodUnaryRPCWithErrorsResponse(result string) *service_unary_rpc_with_errorspb.MethodUnaryRPCWithErrorsResponse {
	message := &service_unary_rpc_with_errorspb.MethodUnaryRPCWithErrorsResponse{}
	message.Field = result
	return message
}

// NewMethodUnaryRPCWithErrorsInternalError builds the gRPC error response type
// from the error of the "MethodUnaryRPCWithErrors" endpoint of the
// "ServiceUnaryRPCWithErrors" service.
func NewMethodUnaryRPCWithErrorsInternalError(er *serviceunaryrpcwitherrors.AnotherError) *service_unary_rpc_with_errorspb.MethodUnaryRPCWithErrorsInternalError {
	message := &service_unary_rpc_with_errorspb.MethodUnaryRPCWithErrorsInternalError{
		Name: er.Name,
	}
	if er.Description != nil {
		message.Description = *er.Description
	}
	return message
}

// NewMethodUnaryRPCWithErrorsBadRequestError builds the gRPC error response
// type from the error of the "MethodUnaryRPCWithErrors" endpoint of the
// "ServiceUnaryRPCWithErrors" service.
func NewMethodUnaryRPCWithErrorsBadRequestError(er *serviceunaryrpcwitherrors.AnotherError) *service_unary_rpc_with_errorspb.MethodUnaryRPCWithErrorsBadRequestError {
	message := &service_unary_rpc_with_errorspb.MethodUnaryRPCWithErrorsBadRequestError{
		Name: er.Name,
	}
	if er.Description != nil {
		message.Description = *er.Description
	}
	return message
}

// NewMethodUnaryRPCWithErrorsCustomErrorError builds the gRPC error response
// type from the error of the "MethodUnaryRPCWithErrors" endpoint of the
// "ServiceUnaryRPCWithErrors" service.
func NewMethodUnaryRPCWithErrorsCustomErrorError(er *serviceunaryrpcwitherrors.ErrorType) *service_unary_rpc_with_errorspb.MethodUnaryRPCWithErrorsCustomErrorError {
	message := &service_unary_rpc_with_errorspb.MethodUnaryRPCWithErrorsCustomErrorError{}
	if er.A != nil {
		message.A = *er.A
	}
	return message
}
`

const ElemValidationServerTypesFile = `// NewMethodElemValidationPayload builds the payload of the
// "MethodElemValidation" endpoint of the "ServiceElemValidation" service from
// the gRPC request type.
func NewMethodElemValidationPayload(message *service_elem_validationpb.MethodElemValidationRequest) *serviceelemvalidation.ResultType {
	v := &serviceelemvalidation.ResultType{}
	if message.Foo != nil {
		v.Foo = make(map[string][]string, len(message.Foo))
		for key, val := range message.Foo {
			tk := key
			tv := make([]string, len(val.Field))
			for i, val := range val.Field {
				tv[i] = val
			}
			v.Foo[tk] = tv
		}
	}
	return v
}

// NewMethodElemValidationResponse builds the gRPC response type from the
// result of the "MethodElemValidation" endpoint of the "ServiceElemValidation"
// service.
func NewMethodElemValidationResponse() *service_elem_validationpb.MethodElemValidationResponse {
	message := &service_elem_validationpb.MethodElemValidationResponse{}
	return message
}

// ValidateMethodElemValidationRequest runs the validations defined on
// MethodElemValidationRequest.
func ValidateMethodElemValidationRequest(message *service_elem_validationpb.MethodElemValidationRequest) (err error) {
	for _, v := range message.Foo {
		if v != nil {
			if err2 := ValidateArrayOfString(v); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
	}
	return
}

// ValidateArrayOfString runs the validations defined on ArrayOfString.
func ValidateArrayOfString(message *service_elem_validationpb.ArrayOfString) (err error) {
	if len(message.Field) < 1 {
		err = goa.MergeErrors(err, goa.InvalidLengthError("message.field", message.Field, len(message.Field), 1, true))
	}
	return
}
`
