package testdata

const PayloadWithNestedTypesClientTypeCode = `// NewMethodPayloadWithNestedTypesRequest builds the gRPC request type from the
// payload of the "MethodPayloadWithNestedTypes" endpoint of the
// "ServicePayloadWithNestedTypes" service.
func NewMethodPayloadWithNestedTypesRequest(payload *servicepayloadwithnestedtypes.MethodPayloadWithNestedTypesPayload) *service_payload_with_nested_typespb.MethodPayloadWithNestedTypesRequest {
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

const ResultWithCollectionClientTypeCode = `// NewMethodResultWithCollectionRequest builds the gRPC request type from the
// payload of the "MethodResultWithCollection" endpoint of the
// "ServiceResultWithCollection" service.
func NewMethodResultWithCollectionRequest() *service_result_with_collectionpb.MethodResultWithCollectionRequest {
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

const WithErrorsClientTypeCode = `// NewMethodUnaryRPCWithErrorsRequest builds the gRPC request type from the
// payload of the "MethodUnaryRPCWithErrors" endpoint of the
// "ServiceUnaryRPCWithErrors" service.
func NewMethodUnaryRPCWithErrorsRequest(payload string) *service_unary_rpc_with_errorspb.MethodUnaryRPCWithErrorsRequest {
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
		Name: message.Name,
	}
	if message.Description != "" {
		er.Description = &message.Description
	}
	return er
}

// NewMethodUnaryRPCWithErrorsBadRequestError builds the error type of the
// "MethodUnaryRPCWithErrors" endpoint of the "ServiceUnaryRPCWithErrors"
// service from the gRPC error response type.
func NewMethodUnaryRPCWithErrorsBadRequestError(message *service_unary_rpc_with_errorspb.MethodUnaryRPCWithErrorsBadRequestError) *serviceunaryrpcwitherrors.AnotherError {
	er := &serviceunaryrpcwitherrors.AnotherError{
		Name: message.Name,
	}
	if message.Description != "" {
		er.Description = &message.Description
	}
	return er
}

// NewMethodUnaryRPCWithErrorsCustomErrorError builds the error type of the
// "MethodUnaryRPCWithErrors" endpoint of the "ServiceUnaryRPCWithErrors"
// service from the gRPC error response type.
func NewMethodUnaryRPCWithErrorsCustomErrorError(message *service_unary_rpc_with_errorspb.MethodUnaryRPCWithErrorsCustomErrorError) *serviceunaryrpcwitherrors.ErrorType {
	er := &serviceunaryrpcwitherrors.ErrorType{}
	if message.A != "" {
		er.A = &message.A
	}
	return er
}

// ValidateMethodUnaryRPCWithErrorsInternalError runs the validations defined
// on MethodUnaryRPCWithErrorsInternalError.
func ValidateMethodUnaryRPCWithErrorsInternalError(message *service_unary_rpc_with_errorspb.MethodUnaryRPCWithErrorsInternalError) (err error) {
	if !(message.Name == "this" || message.Name == "that") {
		err = goa.MergeErrors(err, goa.InvalidEnumValueError("message.name", message.Name, []interface{}{"this", "that"}))
	}
	return
}

// ValidateMethodUnaryRPCWithErrorsBadRequestError runs the validations defined
// on MethodUnaryRPCWithErrorsBadRequestError.
func ValidateMethodUnaryRPCWithErrorsBadRequestError(message *service_unary_rpc_with_errorspb.MethodUnaryRPCWithErrorsBadRequestError) (err error) {
	if !(message.Name == "this" || message.Name == "that") {
		err = goa.MergeErrors(err, goa.InvalidEnumValueError("message.name", message.Name, []interface{}{"this", "that"}))
	}
	return
}
`
