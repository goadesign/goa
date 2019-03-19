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

// ValidateMethodPayloadWithNestedTypesRequest runs the validations defined on
// MethodPayloadWithNestedTypesRequest.
func ValidateMethodPayloadWithNestedTypesRequest(message *service_payload_with_nested_typespb.MethodPayloadWithNestedTypesRequest) (err error) {
	if message.AParams != nil {
		if err2 := ValidateAParams(message.AParams); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// ValidateAParams runs the validations defined on AParams.
func ValidateAParams(message *service_payload_with_nested_typespb.AParams) (err error) {

	return
}

// ValidateArrayOfString runs the validations defined on ArrayOfString.
func ValidateArrayOfString(message *service_payload_with_nested_typespb.ArrayOfString) (err error) {

	return
}

// ValidateBParams runs the validations defined on BParams.
func ValidateBParams(message *service_payload_with_nested_typespb.BParams) (err error) {

	return
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
			intFieldptr := int(val.IntField)
			res.CollectionField[i].IntField = &intFieldptr
		}
	}

	return res
}
`
