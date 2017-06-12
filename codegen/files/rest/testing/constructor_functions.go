package testing

var PayloadQueryBoolConstructorCode = `// NewMethodQueryBoolPayload instantiates and validates the ServiceQueryBool
// service MethodQueryBool endpoint server request body.
func NewMethodQueryBoolPayload(q *bool) (*MethodQueryBoolPayload, error) {
	p := MethodQueryBoolPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryBoolValidateConstructorCode = `// NewMethodQueryBoolValidatePayload instantiates and validates the
// ServiceQueryBoolValidate service MethodQueryBoolValidate endpoint server
// request body.
func NewMethodQueryBoolValidatePayload(q bool) (*MethodQueryBoolValidatePayload, error) {
	p := MethodQueryBoolValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryIntConstructorCode = `// NewMethodQueryIntPayload instantiates and validates the ServiceQueryInt
// service MethodQueryInt endpoint server request body.
func NewMethodQueryIntPayload(q *int) (*MethodQueryIntPayload, error) {
	p := MethodQueryIntPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryIntValidateConstructorCode = `// NewMethodQueryIntValidatePayload instantiates and validates the
// ServiceQueryIntValidate service MethodQueryIntValidate endpoint server
// request body.
func NewMethodQueryIntValidatePayload(q int) (*MethodQueryIntValidatePayload, error) {
	p := MethodQueryIntValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryInt32ConstructorCode = `// NewMethodQueryInt32Payload instantiates and validates the ServiceQueryInt32
// service MethodQueryInt32 endpoint server request body.
func NewMethodQueryInt32Payload(q *int32) (*MethodQueryInt32Payload, error) {
	p := MethodQueryInt32Payload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryInt32ValidateConstructorCode = `// NewMethodQueryInt32ValidatePayload instantiates and validates the
// ServiceQueryInt32Validate service MethodQueryInt32Validate endpoint server
// request body.
func NewMethodQueryInt32ValidatePayload(q int32) (*MethodQueryInt32ValidatePayload, error) {
	p := MethodQueryInt32ValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryInt64ConstructorCode = `// NewMethodQueryInt64Payload instantiates and validates the ServiceQueryInt64
// service MethodQueryInt64 endpoint server request body.
func NewMethodQueryInt64Payload(q *int64) (*MethodQueryInt64Payload, error) {
	p := MethodQueryInt64Payload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryInt64ValidateConstructorCode = `// NewMethodQueryInt64ValidatePayload instantiates and validates the
// ServiceQueryInt64Validate service MethodQueryInt64Validate endpoint server
// request body.
func NewMethodQueryInt64ValidatePayload(q int64) (*MethodQueryInt64ValidatePayload, error) {
	p := MethodQueryInt64ValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryUIntConstructorCode = `// NewMethodQueryUIntPayload instantiates and validates the ServiceQueryUInt
// service MethodQueryUInt endpoint server request body.
func NewMethodQueryUIntPayload(q *uint) (*MethodQueryUIntPayload, error) {
	p := MethodQueryUIntPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryUIntValidateConstructorCode = `// NewMethodQueryUIntValidatePayload instantiates and validates the
// ServiceQueryUIntValidate service MethodQueryUIntValidate endpoint server
// request body.
func NewMethodQueryUIntValidatePayload(q uint) (*MethodQueryUIntValidatePayload, error) {
	p := MethodQueryUIntValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryUInt32ConstructorCode = `// NewMethodQueryUInt32Payload instantiates and validates the
// ServiceQueryUInt32 service MethodQueryUInt32 endpoint server request body.
func NewMethodQueryUInt32Payload(q *uint32) (*MethodQueryUInt32Payload, error) {
	p := MethodQueryUInt32Payload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryUInt32ValidateConstructorCode = `// NewMethodQueryUInt32ValidatePayload instantiates and validates the
// ServiceQueryUInt32Validate service MethodQueryUInt32Validate endpoint server
// request body.
func NewMethodQueryUInt32ValidatePayload(q uint32) (*MethodQueryUInt32ValidatePayload, error) {
	p := MethodQueryUInt32ValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryUInt64ConstructorCode = `// NewMethodQueryUInt64Payload instantiates and validates the
// ServiceQueryUInt64 service MethodQueryUInt64 endpoint server request body.
func NewMethodQueryUInt64Payload(q *uint64) (*MethodQueryUInt64Payload, error) {
	p := MethodQueryUInt64Payload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryUInt64ValidateConstructorCode = `// NewMethodQueryUInt64ValidatePayload instantiates and validates the
// ServiceQueryUInt64Validate service MethodQueryUInt64Validate endpoint server
// request body.
func NewMethodQueryUInt64ValidatePayload(q uint64) (*MethodQueryUInt64ValidatePayload, error) {
	p := MethodQueryUInt64ValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryFloat32ConstructorCode = `// NewMethodQueryFloat32Payload instantiates and validates the
// ServiceQueryFloat32 service MethodQueryFloat32 endpoint server request body.
func NewMethodQueryFloat32Payload(q *float32) (*MethodQueryFloat32Payload, error) {
	p := MethodQueryFloat32Payload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryFloat32ValidateConstructorCode = `// NewMethodQueryFloat32ValidatePayload instantiates and validates the
// ServiceQueryFloat32Validate service MethodQueryFloat32Validate endpoint
// server request body.
func NewMethodQueryFloat32ValidatePayload(q float32) (*MethodQueryFloat32ValidatePayload, error) {
	p := MethodQueryFloat32ValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryFloat64ConstructorCode = `// NewMethodQueryFloat64Payload instantiates and validates the
// ServiceQueryFloat64 service MethodQueryFloat64 endpoint server request body.
func NewMethodQueryFloat64Payload(q *float64) (*MethodQueryFloat64Payload, error) {
	p := MethodQueryFloat64Payload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryFloat64ValidateConstructorCode = `// NewMethodQueryFloat64ValidatePayload instantiates and validates the
// ServiceQueryFloat64Validate service MethodQueryFloat64Validate endpoint
// server request body.
func NewMethodQueryFloat64ValidatePayload(q float64) (*MethodQueryFloat64ValidatePayload, error) {
	p := MethodQueryFloat64ValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryStringConstructorCode = `// NewMethodQueryStringPayload instantiates and validates the
// ServiceQueryString service MethodQueryString endpoint server request body.
func NewMethodQueryStringPayload(q *string) (*MethodQueryStringPayload, error) {
	p := MethodQueryStringPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryStringValidateConstructorCode = `// NewMethodQueryStringValidatePayload instantiates and validates the
// ServiceQueryStringValidate service MethodQueryStringValidate endpoint server
// request body.
func NewMethodQueryStringValidatePayload(q string) (*MethodQueryStringValidatePayload, error) {
	p := MethodQueryStringValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryBytesConstructorCode = `// NewMethodQueryBytesPayload instantiates and validates the ServiceQueryBytes
// service MethodQueryBytes endpoint server request body.
func NewMethodQueryBytesPayload(q []byte) (*MethodQueryBytesPayload, error) {
	p := MethodQueryBytesPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryBytesValidateConstructorCode = `// NewMethodQueryBytesValidatePayload instantiates and validates the
// ServiceQueryBytesValidate service MethodQueryBytesValidate endpoint server
// request body.
func NewMethodQueryBytesValidatePayload(q []byte) (*MethodQueryBytesValidatePayload, error) {
	p := MethodQueryBytesValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryAnyConstructorCode = `// NewMethodQueryAnyPayload instantiates and validates the ServiceQueryAny
// service MethodQueryAny endpoint server request body.
func NewMethodQueryAnyPayload(q interface{}) (*MethodQueryAnyPayload, error) {
	p := MethodQueryAnyPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryAnyValidateConstructorCode = `// NewMethodQueryAnyValidatePayload instantiates and validates the
// ServiceQueryAnyValidate service MethodQueryAnyValidate endpoint server
// request body.
func NewMethodQueryAnyValidatePayload(q interface{}) (*MethodQueryAnyValidatePayload, error) {
	p := MethodQueryAnyValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayBoolConstructorCode = `// NewMethodQueryArrayBoolPayload instantiates and validates the
// ServiceQueryArrayBool service MethodQueryArrayBool endpoint server request
// body.
func NewMethodQueryArrayBoolPayload(q []bool) (*MethodQueryArrayBoolPayload, error) {
	p := MethodQueryArrayBoolPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayBoolValidateConstructorCode = `// NewMethodQueryArrayBoolValidatePayload instantiates and validates the
// ServiceQueryArrayBoolValidate service MethodQueryArrayBoolValidate endpoint
// server request body.
func NewMethodQueryArrayBoolValidatePayload(q []bool) (*MethodQueryArrayBoolValidatePayload, error) {
	p := MethodQueryArrayBoolValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayIntConstructorCode = `// NewMethodQueryArrayIntPayload instantiates and validates the
// ServiceQueryArrayInt service MethodQueryArrayInt endpoint server request
// body.
func NewMethodQueryArrayIntPayload(q []int) (*MethodQueryArrayIntPayload, error) {
	p := MethodQueryArrayIntPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayIntValidateConstructorCode = `// NewMethodQueryArrayIntValidatePayload instantiates and validates the
// ServiceQueryArrayIntValidate service MethodQueryArrayIntValidate endpoint
// server request body.
func NewMethodQueryArrayIntValidatePayload(q []int) (*MethodQueryArrayIntValidatePayload, error) {
	p := MethodQueryArrayIntValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayInt32ConstructorCode = `// NewMethodQueryArrayInt32Payload instantiates and validates the
// ServiceQueryArrayInt32 service MethodQueryArrayInt32 endpoint server request
// body.
func NewMethodQueryArrayInt32Payload(q []int32) (*MethodQueryArrayInt32Payload, error) {
	p := MethodQueryArrayInt32Payload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayInt32ValidateConstructorCode = `// NewMethodQueryArrayInt32ValidatePayload instantiates and validates the
// ServiceQueryArrayInt32Validate service MethodQueryArrayInt32Validate
// endpoint server request body.
func NewMethodQueryArrayInt32ValidatePayload(q []int32) (*MethodQueryArrayInt32ValidatePayload, error) {
	p := MethodQueryArrayInt32ValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayInt64ConstructorCode = `// NewMethodQueryArrayInt64Payload instantiates and validates the
// ServiceQueryArrayInt64 service MethodQueryArrayInt64 endpoint server request
// body.
func NewMethodQueryArrayInt64Payload(q []int64) (*MethodQueryArrayInt64Payload, error) {
	p := MethodQueryArrayInt64Payload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayInt64ValidateConstructorCode = `// NewMethodQueryArrayInt64ValidatePayload instantiates and validates the
// ServiceQueryArrayInt64Validate service MethodQueryArrayInt64Validate
// endpoint server request body.
func NewMethodQueryArrayInt64ValidatePayload(q []int64) (*MethodQueryArrayInt64ValidatePayload, error) {
	p := MethodQueryArrayInt64ValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayUIntConstructorCode = `// NewMethodQueryArrayUIntPayload instantiates and validates the
// ServiceQueryArrayUInt service MethodQueryArrayUInt endpoint server request
// body.
func NewMethodQueryArrayUIntPayload(q []uint) (*MethodQueryArrayUIntPayload, error) {
	p := MethodQueryArrayUIntPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayUIntValidateConstructorCode = `// NewMethodQueryArrayUIntValidatePayload instantiates and validates the
// ServiceQueryArrayUIntValidate service MethodQueryArrayUIntValidate endpoint
// server request body.
func NewMethodQueryArrayUIntValidatePayload(q []uint) (*MethodQueryArrayUIntValidatePayload, error) {
	p := MethodQueryArrayUIntValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayUInt32ConstructorCode = `// NewMethodQueryArrayUInt32Payload instantiates and validates the
// ServiceQueryArrayUInt32 service MethodQueryArrayUInt32 endpoint server
// request body.
func NewMethodQueryArrayUInt32Payload(q []uint32) (*MethodQueryArrayUInt32Payload, error) {
	p := MethodQueryArrayUInt32Payload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayUInt32ValidateConstructorCode = `// NewMethodQueryArrayUInt32ValidatePayload instantiates and validates the
// ServiceQueryArrayUInt32Validate service MethodQueryArrayUInt32Validate
// endpoint server request body.
func NewMethodQueryArrayUInt32ValidatePayload(q []uint32) (*MethodQueryArrayUInt32ValidatePayload, error) {
	p := MethodQueryArrayUInt32ValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayUInt64ConstructorCode = `// NewMethodQueryArrayUInt64Payload instantiates and validates the
// ServiceQueryArrayUInt64 service MethodQueryArrayUInt64 endpoint server
// request body.
func NewMethodQueryArrayUInt64Payload(q []uint64) (*MethodQueryArrayUInt64Payload, error) {
	p := MethodQueryArrayUInt64Payload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayUInt64ValidateConstructorCode = `// NewMethodQueryArrayUInt64ValidatePayload instantiates and validates the
// ServiceQueryArrayUInt64Validate service MethodQueryArrayUInt64Validate
// endpoint server request body.
func NewMethodQueryArrayUInt64ValidatePayload(q []uint64) (*MethodQueryArrayUInt64ValidatePayload, error) {
	p := MethodQueryArrayUInt64ValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayFloat32ConstructorCode = `// NewMethodQueryArrayFloat32Payload instantiates and validates the
// ServiceQueryArrayFloat32 service MethodQueryArrayFloat32 endpoint server
// request body.
func NewMethodQueryArrayFloat32Payload(q []float32) (*MethodQueryArrayFloat32Payload, error) {
	p := MethodQueryArrayFloat32Payload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayFloat32ValidateConstructorCode = `// NewMethodQueryArrayFloat32ValidatePayload instantiates and validates the
// ServiceQueryArrayFloat32Validate service MethodQueryArrayFloat32Validate
// endpoint server request body.
func NewMethodQueryArrayFloat32ValidatePayload(q []float32) (*MethodQueryArrayFloat32ValidatePayload, error) {
	p := MethodQueryArrayFloat32ValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayFloat64ConstructorCode = `// NewMethodQueryArrayFloat64Payload instantiates and validates the
// ServiceQueryArrayFloat64 service MethodQueryArrayFloat64 endpoint server
// request body.
func NewMethodQueryArrayFloat64Payload(q []float64) (*MethodQueryArrayFloat64Payload, error) {
	p := MethodQueryArrayFloat64Payload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayFloat64ValidateConstructorCode = `// NewMethodQueryArrayFloat64ValidatePayload instantiates and validates the
// ServiceQueryArrayFloat64Validate service MethodQueryArrayFloat64Validate
// endpoint server request body.
func NewMethodQueryArrayFloat64ValidatePayload(q []float64) (*MethodQueryArrayFloat64ValidatePayload, error) {
	p := MethodQueryArrayFloat64ValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayStringConstructorCode = `// NewMethodQueryArrayStringPayload instantiates and validates the
// ServiceQueryArrayString service MethodQueryArrayString endpoint server
// request body.
func NewMethodQueryArrayStringPayload(q []string) (*MethodQueryArrayStringPayload, error) {
	p := MethodQueryArrayStringPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayStringValidateConstructorCode = `// NewMethodQueryArrayStringValidatePayload instantiates and validates the
// ServiceQueryArrayStringValidate service MethodQueryArrayStringValidate
// endpoint server request body.
func NewMethodQueryArrayStringValidatePayload(q []string) (*MethodQueryArrayStringValidatePayload, error) {
	p := MethodQueryArrayStringValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayBytesConstructorCode = `// NewMethodQueryArrayBytesPayload instantiates and validates the
// ServiceQueryArrayBytes service MethodQueryArrayBytes endpoint server request
// body.
func NewMethodQueryArrayBytesPayload(q [][]byte) (*MethodQueryArrayBytesPayload, error) {
	p := MethodQueryArrayBytesPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayBytesValidateConstructorCode = `// NewMethodQueryArrayBytesValidatePayload instantiates and validates the
// ServiceQueryArrayBytesValidate service MethodQueryArrayBytesValidate
// endpoint server request body.
func NewMethodQueryArrayBytesValidatePayload(q [][]byte) (*MethodQueryArrayBytesValidatePayload, error) {
	p := MethodQueryArrayBytesValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayAnyConstructorCode = `// NewMethodQueryArrayAnyPayload instantiates and validates the
// ServiceQueryArrayAny service MethodQueryArrayAny endpoint server request
// body.
func NewMethodQueryArrayAnyPayload(q []interface{}) (*MethodQueryArrayAnyPayload, error) {
	p := MethodQueryArrayAnyPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayAnyValidateConstructorCode = `// NewMethodQueryArrayAnyValidatePayload instantiates and validates the
// ServiceQueryArrayAnyValidate service MethodQueryArrayAnyValidate endpoint
// server request body.
func NewMethodQueryArrayAnyValidatePayload(q []interface{}) (*MethodQueryArrayAnyValidatePayload, error) {
	p := MethodQueryArrayAnyValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryMapStringStringConstructorCode = `// NewMethodQueryMapStringStringPayload instantiates and validates the
// ServiceQueryMapStringString service MethodQueryMapStringString endpoint
// server request body.
func NewMethodQueryMapStringStringPayload(q map[string]string) (*MethodQueryMapStringStringPayload, error) {
	p := MethodQueryMapStringStringPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryMapStringStringValidateConstructorCode = `// NewMethodQueryMapStringStringValidatePayload instantiates and validates the
// ServiceQueryMapStringStringValidate service
// MethodQueryMapStringStringValidate endpoint server request body.
func NewMethodQueryMapStringStringValidatePayload(q map[string]string) (*MethodQueryMapStringStringValidatePayload, error) {
	p := MethodQueryMapStringStringValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryMapStringBoolConstructorCode = `// NewMethodQueryMapStringBoolPayload instantiates and validates the
// ServiceQueryMapStringBool service MethodQueryMapStringBool endpoint server
// request body.
func NewMethodQueryMapStringBoolPayload(q map[string]bool) (*MethodQueryMapStringBoolPayload, error) {
	p := MethodQueryMapStringBoolPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryMapStringBoolValidateConstructorCode = `// NewMethodQueryMapStringBoolValidatePayload instantiates and validates the
// ServiceQueryMapStringBoolValidate service MethodQueryMapStringBoolValidate
// endpoint server request body.
func NewMethodQueryMapStringBoolValidatePayload(q map[string]bool) (*MethodQueryMapStringBoolValidatePayload, error) {
	p := MethodQueryMapStringBoolValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryMapBoolStringConstructorCode = `// NewMethodQueryMapBoolStringPayload instantiates and validates the
// ServiceQueryMapBoolString service MethodQueryMapBoolString endpoint server
// request body.
func NewMethodQueryMapBoolStringPayload(q map[bool]string) (*MethodQueryMapBoolStringPayload, error) {
	p := MethodQueryMapBoolStringPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryMapBoolStringValidateConstructorCode = `// NewMethodQueryMapBoolStringValidatePayload instantiates and validates the
// ServiceQueryMapBoolStringValidate service MethodQueryMapBoolStringValidate
// endpoint server request body.
func NewMethodQueryMapBoolStringValidatePayload(q map[bool]string) (*MethodQueryMapBoolStringValidatePayload, error) {
	p := MethodQueryMapBoolStringValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryMapBoolBoolConstructorCode = `// NewMethodQueryMapBoolBoolPayload instantiates and validates the
// ServiceQueryMapBoolBool service MethodQueryMapBoolBool endpoint server
// request body.
func NewMethodQueryMapBoolBoolPayload(q map[bool]bool) (*MethodQueryMapBoolBoolPayload, error) {
	p := MethodQueryMapBoolBoolPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryMapBoolBoolValidateConstructorCode = `// NewMethodQueryMapBoolBoolValidatePayload instantiates and validates the
// ServiceQueryMapBoolBoolValidate service MethodQueryMapBoolBoolValidate
// endpoint server request body.
func NewMethodQueryMapBoolBoolValidatePayload(q map[bool]bool) (*MethodQueryMapBoolBoolValidatePayload, error) {
	p := MethodQueryMapBoolBoolValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryMapStringArrayStringConstructorCode = `// NewMethodQueryMapStringArrayStringPayload instantiates and validates the
// ServiceQueryMapStringArrayString service MethodQueryMapStringArrayString
// endpoint server request body.
func NewMethodQueryMapStringArrayStringPayload(q map[string][]string) (*MethodQueryMapStringArrayStringPayload, error) {
	p := MethodQueryMapStringArrayStringPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryMapStringArrayStringValidateConstructorCode = `// NewMethodQueryMapStringArrayStringValidatePayload instantiates and validates
// the ServiceQueryMapStringArrayStringValidate service
// MethodQueryMapStringArrayStringValidate endpoint server request body.
func NewMethodQueryMapStringArrayStringValidatePayload(q map[string][]string) (*MethodQueryMapStringArrayStringValidatePayload, error) {
	p := MethodQueryMapStringArrayStringValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryMapStringArrayBoolConstructorCode = `// NewMethodQueryMapStringArrayBoolPayload instantiates and validates the
// ServiceQueryMapStringArrayBool service MethodQueryMapStringArrayBool
// endpoint server request body.
func NewMethodQueryMapStringArrayBoolPayload(q map[string][]bool) (*MethodQueryMapStringArrayBoolPayload, error) {
	p := MethodQueryMapStringArrayBoolPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryMapStringArrayBoolValidateConstructorCode = `// NewMethodQueryMapStringArrayBoolValidatePayload instantiates and validates
// the ServiceQueryMapStringArrayBoolValidate service
// MethodQueryMapStringArrayBoolValidate endpoint server request body.
func NewMethodQueryMapStringArrayBoolValidatePayload(q map[string][]bool) (*MethodQueryMapStringArrayBoolValidatePayload, error) {
	p := MethodQueryMapStringArrayBoolValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryMapBoolArrayStringConstructorCode = `// NewMethodQueryMapBoolArrayStringPayload instantiates and validates the
// ServiceQueryMapBoolArrayString service MethodQueryMapBoolArrayString
// endpoint server request body.
func NewMethodQueryMapBoolArrayStringPayload(q map[bool][]string) (*MethodQueryMapBoolArrayStringPayload, error) {
	p := MethodQueryMapBoolArrayStringPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryMapBoolArrayStringValidateConstructorCode = `// NewMethodQueryMapBoolArrayStringValidatePayload instantiates and validates
// the ServiceQueryMapBoolArrayStringValidate service
// MethodQueryMapBoolArrayStringValidate endpoint server request body.
func NewMethodQueryMapBoolArrayStringValidatePayload(q map[bool][]string) (*MethodQueryMapBoolArrayStringValidatePayload, error) {
	p := MethodQueryMapBoolArrayStringValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryMapBoolArrayBoolConstructorCode = `// NewMethodQueryMapBoolArrayBoolPayload instantiates and validates the
// ServiceQueryMapBoolArrayBool service MethodQueryMapBoolArrayBool endpoint
// server request body.
func NewMethodQueryMapBoolArrayBoolPayload(q map[bool][]bool) (*MethodQueryMapBoolArrayBoolPayload, error) {
	p := MethodQueryMapBoolArrayBoolPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryMapBoolArrayBoolValidateConstructorCode = `// NewMethodQueryMapBoolArrayBoolValidatePayload instantiates and validates the
// ServiceQueryMapBoolArrayBoolValidate service
// MethodQueryMapBoolArrayBoolValidate endpoint server request body.
func NewMethodQueryMapBoolArrayBoolValidatePayload(q map[bool][]bool) (*MethodQueryMapBoolArrayBoolValidatePayload, error) {
	p := MethodQueryMapBoolArrayBoolValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadPathStringConstructorCode = `// NewMethodPathStringPayload instantiates and validates the ServicePathString
// service MethodPathString endpoint server request body.
func NewMethodPathStringPayload(p string) (*MethodPathStringPayload, error) {
	p := MethodPathStringPayload{
		P: p,
	}
	return &p, nil
}
`

var PayloadPathStringValidateConstructorCode = `// NewMethodPathStringValidatePayload instantiates and validates the
// ServicePathStringValidate service MethodPathStringValidate endpoint server
// request body.
func NewMethodPathStringValidatePayload(p string) (*MethodPathStringValidatePayload, error) {
	p := MethodPathStringValidatePayload{
		P: p,
	}
	return &p, nil
}
`

var PayloadPathArrayStringConstructorCode = `// NewMethodPathArrayStringPayload instantiates and validates the
// ServicePathArrayString service MethodPathArrayString endpoint server request
// body.
func NewMethodPathArrayStringPayload(p []string) (*MethodPathArrayStringPayload, error) {
	p := MethodPathArrayStringPayload{
		P: p,
	}
	return &p, nil
}
`

var PayloadPathArrayStringValidateConstructorCode = `// NewMethodPathArrayStringValidatePayload instantiates and validates the
// ServicePathArrayStringValidate service MethodPathArrayStringValidate
// endpoint server request body.
func NewMethodPathArrayStringValidatePayload(p []string) (*MethodPathArrayStringValidatePayload, error) {
	p := MethodPathArrayStringValidatePayload{
		P: p,
	}
	return &p, nil
}
`

var PayloadHeaderStringConstructorCode = `// NewMethodHeaderStringPayload instantiates and validates the
// ServiceHeaderString service MethodHeaderString endpoint server request body.
func NewMethodHeaderStringPayload(h *string) (*MethodHeaderStringPayload, error) {
	p := MethodHeaderStringPayload{
		H: h,
	}
	return &p, nil
}
`

var PayloadHeaderStringValidateConstructorCode = `// NewMethodHeaderStringValidatePayload instantiates and validates the
// ServiceHeaderStringValidate service MethodHeaderStringValidate endpoint
// server request body.
func NewMethodHeaderStringValidatePayload(h *string) (*MethodHeaderStringValidatePayload, error) {
	p := MethodHeaderStringValidatePayload{
		H: h,
	}
	return &p, nil
}
`

var PayloadHeaderArrayStringConstructorCode = `// NewMethodHeaderArrayStringPayload instantiates and validates the
// ServiceHeaderArrayString service MethodHeaderArrayString endpoint server
// request body.
func NewMethodHeaderArrayStringPayload(h []string) (*MethodHeaderArrayStringPayload, error) {
	p := MethodHeaderArrayStringPayload{
		H: h,
	}
	return &p, nil
}
`

var PayloadHeaderArrayStringValidateConstructorCode = `// NewMethodHeaderArrayStringValidatePayload instantiates and validates the
// ServiceHeaderArrayStringValidate service MethodHeaderArrayStringValidate
// endpoint server request body.
func NewMethodHeaderArrayStringValidatePayload(h []string) (*MethodHeaderArrayStringValidatePayload, error) {
	p := MethodHeaderArrayStringValidatePayload{
		H: h,
	}
	return &p, nil
}
`

var PayloadBodyQueryObjectConstructorCode = `// NewMethodBodyQueryObjectPayload instantiates and validates the
// ServiceBodyQueryObject service MethodBodyQueryObject endpoint server request
// body.
func NewMethodBodyQueryObjectPayload(body *MethodBodyQueryObjectServerRequestBody, b *string) (*MethodBodyQueryObjectPayload, error) {
	p := MethodBodyQueryObjectPayload{
		A: body.A,
		B: b,
	}
	return &p, nil
}
`

var PayloadBodyQueryObjectValidateConstructorCode = `// NewMethodBodyQueryObjectValidatePayload instantiates and validates the
// ServiceBodyQueryObjectValidate service MethodBodyQueryObjectValidate
// endpoint server request body.
func NewMethodBodyQueryObjectValidatePayload(body *MethodBodyQueryObjectValidateServerRequestBody, b string) (*MethodBodyQueryObjectValidatePayload, error) {
	p := MethodBodyQueryObjectValidatePayload{
		A: body.A,
		B: b,
	}
	return &p, nil
}
`

var PayloadBodyQueryUserConstructorCode = `// NewPayloadType instantiates and validates the ServiceBodyQueryUser service
// MethodBodyQueryUser endpoint server request body.
func NewPayloadType(body *MethodBodyQueryUserServerRequestBody, b *string) (*PayloadType, error) {
	p := PayloadType{
		A: body.A,
		B: b,
	}
	return &p, nil
}
`

var PayloadBodyQueryUserValidateConstructorCode = `// NewPayloadType instantiates and validates the ServiceBodyQueryUserValidate
// service MethodBodyQueryUserValidate endpoint server request body.
func NewPayloadType(body *MethodBodyQueryUserValidateServerRequestBody, b string) (*PayloadType, error) {
	p := PayloadType{
		A: body.A,
		B: b,
	}
	return &p, nil
}
`

var PayloadBodyPathObjectConstructorCode = `// NewMethodBodyPathObjectPayload instantiates and validates the
// ServiceBodyPathObject service MethodBodyPathObject endpoint server request
// body.
func NewMethodBodyPathObjectPayload(body *MethodBodyPathObjectServerRequestBody, b string) (*MethodBodyPathObjectPayload, error) {
	p := MethodBodyPathObjectPayload{
		A: body.A,
		B: b,
	}
	return &p, nil
}
`

var PayloadBodyPathObjectValidateConstructorCode = `// NewMethodBodyPathObjectValidatePayload instantiates and validates the
// ServiceBodyPathObjectValidate service MethodBodyPathObjectValidate endpoint
// server request body.
func NewMethodBodyPathObjectValidatePayload(body *MethodBodyPathObjectValidateServerRequestBody, b string) (*MethodBodyPathObjectValidatePayload, error) {
	p := MethodBodyPathObjectValidatePayload{
		A: body.A,
		B: b,
	}
	return &p, nil
}
`

var PayloadBodyPathUserConstructorCode = `// NewPayloadType instantiates and validates the ServiceBodyPathUser service
// MethodBodyPathUser endpoint server request body.
func NewPayloadType(body *MethodBodyPathUserServerRequestBody, b string) (*PayloadType, error) {
	p := PayloadType{
		A: body.A,
		B: b,
	}
	return &p, nil
}
`

var PayloadBodyPathUserValidateConstructorCode = `// NewPayloadType instantiates and validates the ServiceBodyPathUserValidate
// service MethodUserBodyPathValidate endpoint server request body.
func NewPayloadType(body *MethodUserBodyPathValidateServerRequestBody, b string) (*PayloadType, error) {
	p := PayloadType{
		A: body.A,
		B: b,
	}
	return &p, nil
}
`

var PayloadBodyQueryPathObjectConstructorCode = `// NewMethodBodyQueryPathObjectPayload instantiates and validates the
// ServiceBodyQueryPathObject service MethodBodyQueryPathObject endpoint server
// request body.
func NewMethodBodyQueryPathObjectPayload(body *MethodBodyQueryPathObjectServerRequestBody, b *string, c string) (*MethodBodyQueryPathObjectPayload, error) {
	p := MethodBodyQueryPathObjectPayload{
		A: body.A,
		B: b,
		C: c,
	}
	return &p, nil
}
`

var PayloadBodyQueryPathObjectValidateConstructorCode = `// NewMethodBodyQueryPathObjectValidatePayload instantiates and validates the
// ServiceBodyQueryPathObjectValidate service MethodBodyQueryPathObjectValidate
// endpoint server request body.
func NewMethodBodyQueryPathObjectValidatePayload(body *MethodBodyQueryPathObjectValidateServerRequestBody, b string, c string) (*MethodBodyQueryPathObjectValidatePayload, error) {
	p := MethodBodyQueryPathObjectValidatePayload{
		A: body.A,
		B: b,
		C: c,
	}
	return &p, nil
}
`

var PayloadBodyQueryPathUserConstructorCode = `// NewPayloadType instantiates and validates the ServiceBodyQueryPathUser
// service MethodBodyQueryPathUser endpoint server request body.
func NewPayloadType(body *MethodBodyQueryPathUserServerRequestBody, b *string, c string) (*PayloadType, error) {
	p := PayloadType{
		A: body.A,
		B: b,
		C: c,
	}
	return &p, nil
}
`

var PayloadBodyQueryPathUserValidateConstructorCode = `// NewPayloadType instantiates and validates the
// ServiceBodyQueryPathUserValidate service MethodBodyQueryPathUserValidate
// endpoint server request body.
func NewPayloadType(body *MethodBodyQueryPathUserValidateServerRequestBody, b string, c string) (*PayloadType, error) {
	p := PayloadType{
		A: body.A,
		B: b,
		C: c,
	}
	return &p, nil
}
`
