package testing

var PayloadQueryBoolConstructorCode = `// NewEndpointQueryBoolPayload instantiates and validates the ServiceQueryBool
// service EndpointQueryBool endpoint server request body.
func NewEndpointQueryBoolPayload(q *bool) (*EndpointQueryBoolPayload, error) {
	p := EndpointQueryBoolPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryBoolValidateConstructorCode = `// NewEndpointQueryBoolValidatePayload instantiates and validates the
// ServiceQueryBoolValidate service EndpointQueryBoolValidate endpoint server
// request body.
func NewEndpointQueryBoolValidatePayload(q bool) (*EndpointQueryBoolValidatePayload, error) {
	p := EndpointQueryBoolValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryIntConstructorCode = `// NewEndpointQueryIntPayload instantiates and validates the ServiceQueryInt
// service EndpointQueryInt endpoint server request body.
func NewEndpointQueryIntPayload(q *int) (*EndpointQueryIntPayload, error) {
	p := EndpointQueryIntPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryIntValidateConstructorCode = `// NewEndpointQueryIntValidatePayload instantiates and validates the
// ServiceQueryIntValidate service EndpointQueryIntValidate endpoint server
// request body.
func NewEndpointQueryIntValidatePayload(q int) (*EndpointQueryIntValidatePayload, error) {
	p := EndpointQueryIntValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryInt32ConstructorCode = `// NewEndpointQueryInt32Payload instantiates and validates the
// ServiceQueryInt32 service EndpointQueryInt32 endpoint server request body.
func NewEndpointQueryInt32Payload(q *int32) (*EndpointQueryInt32Payload, error) {
	p := EndpointQueryInt32Payload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryInt32ValidateConstructorCode = `// NewEndpointQueryInt32ValidatePayload instantiates and validates the
// ServiceQueryInt32Validate service EndpointQueryInt32Validate endpoint server
// request body.
func NewEndpointQueryInt32ValidatePayload(q int32) (*EndpointQueryInt32ValidatePayload, error) {
	p := EndpointQueryInt32ValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryInt64ConstructorCode = `// NewEndpointQueryInt64Payload instantiates and validates the
// ServiceQueryInt64 service EndpointQueryInt64 endpoint server request body.
func NewEndpointQueryInt64Payload(q *int64) (*EndpointQueryInt64Payload, error) {
	p := EndpointQueryInt64Payload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryInt64ValidateConstructorCode = `// NewEndpointQueryInt64ValidatePayload instantiates and validates the
// ServiceQueryInt64Validate service EndpointQueryInt64Validate endpoint server
// request body.
func NewEndpointQueryInt64ValidatePayload(q int64) (*EndpointQueryInt64ValidatePayload, error) {
	p := EndpointQueryInt64ValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryUIntConstructorCode = `// NewEndpointQueryUIntPayload instantiates and validates the ServiceQueryUInt
// service EndpointQueryUInt endpoint server request body.
func NewEndpointQueryUIntPayload(q *uint) (*EndpointQueryUIntPayload, error) {
	p := EndpointQueryUIntPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryUIntValidateConstructorCode = `// NewEndpointQueryUIntValidatePayload instantiates and validates the
// ServiceQueryUIntValidate service EndpointQueryUIntValidate endpoint server
// request body.
func NewEndpointQueryUIntValidatePayload(q uint) (*EndpointQueryUIntValidatePayload, error) {
	p := EndpointQueryUIntValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryUInt32ConstructorCode = `// NewEndpointQueryUInt32Payload instantiates and validates the
// ServiceQueryUInt32 service EndpointQueryUInt32 endpoint server request body.
func NewEndpointQueryUInt32Payload(q *uint32) (*EndpointQueryUInt32Payload, error) {
	p := EndpointQueryUInt32Payload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryUInt32ValidateConstructorCode = `// NewEndpointQueryUInt32ValidatePayload instantiates and validates the
// ServiceQueryUInt32Validate service EndpointQueryUInt32Validate endpoint
// server request body.
func NewEndpointQueryUInt32ValidatePayload(q uint32) (*EndpointQueryUInt32ValidatePayload, error) {
	p := EndpointQueryUInt32ValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryUInt64ConstructorCode = `// NewEndpointQueryUInt64Payload instantiates and validates the
// ServiceQueryUInt64 service EndpointQueryUInt64 endpoint server request body.
func NewEndpointQueryUInt64Payload(q *uint64) (*EndpointQueryUInt64Payload, error) {
	p := EndpointQueryUInt64Payload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryUInt64ValidateConstructorCode = `// NewEndpointQueryUInt64ValidatePayload instantiates and validates the
// ServiceQueryUInt64Validate service EndpointQueryUInt64Validate endpoint
// server request body.
func NewEndpointQueryUInt64ValidatePayload(q uint64) (*EndpointQueryUInt64ValidatePayload, error) {
	p := EndpointQueryUInt64ValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryFloat32ConstructorCode = `// NewEndpointQueryFloat32Payload instantiates and validates the
// ServiceQueryFloat32 service EndpointQueryFloat32 endpoint server request
// body.
func NewEndpointQueryFloat32Payload(q *float32) (*EndpointQueryFloat32Payload, error) {
	p := EndpointQueryFloat32Payload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryFloat32ValidateConstructorCode = `// NewEndpointQueryFloat32ValidatePayload instantiates and validates the
// ServiceQueryFloat32Validate service EndpointQueryFloat32Validate endpoint
// server request body.
func NewEndpointQueryFloat32ValidatePayload(q float32) (*EndpointQueryFloat32ValidatePayload, error) {
	p := EndpointQueryFloat32ValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryFloat64ConstructorCode = `// NewEndpointQueryFloat64Payload instantiates and validates the
// ServiceQueryFloat64 service EndpointQueryFloat64 endpoint server request
// body.
func NewEndpointQueryFloat64Payload(q *float64) (*EndpointQueryFloat64Payload, error) {
	p := EndpointQueryFloat64Payload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryFloat64ValidateConstructorCode = `// NewEndpointQueryFloat64ValidatePayload instantiates and validates the
// ServiceQueryFloat64Validate service EndpointQueryFloat64Validate endpoint
// server request body.
func NewEndpointQueryFloat64ValidatePayload(q float64) (*EndpointQueryFloat64ValidatePayload, error) {
	p := EndpointQueryFloat64ValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryStringConstructorCode = `// NewEndpointQueryStringPayload instantiates and validates the
// ServiceQueryString service EndpointQueryString endpoint server request body.
func NewEndpointQueryStringPayload(q *string) (*EndpointQueryStringPayload, error) {
	p := EndpointQueryStringPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryStringValidateConstructorCode = `// NewEndpointQueryStringValidatePayload instantiates and validates the
// ServiceQueryStringValidate service EndpointQueryStringValidate endpoint
// server request body.
func NewEndpointQueryStringValidatePayload(q string) (*EndpointQueryStringValidatePayload, error) {
	p := EndpointQueryStringValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryBytesConstructorCode = `// NewEndpointQueryBytesPayload instantiates and validates the
// ServiceQueryBytes service EndpointQueryBytes endpoint server request body.
func NewEndpointQueryBytesPayload(q []byte) (*EndpointQueryBytesPayload, error) {
	p := EndpointQueryBytesPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryBytesValidateConstructorCode = `// NewEndpointQueryBytesValidatePayload instantiates and validates the
// ServiceQueryBytesValidate service EndpointQueryBytesValidate endpoint server
// request body.
func NewEndpointQueryBytesValidatePayload(q []byte) (*EndpointQueryBytesValidatePayload, error) {
	p := EndpointQueryBytesValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryAnyConstructorCode = `// NewEndpointQueryAnyPayload instantiates and validates the ServiceQueryAny
// service EndpointQueryAny endpoint server request body.
func NewEndpointQueryAnyPayload(q interface{}) (*EndpointQueryAnyPayload, error) {
	p := EndpointQueryAnyPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryAnyValidateConstructorCode = `// NewEndpointQueryAnyValidatePayload instantiates and validates the
// ServiceQueryAnyValidate service EndpointQueryAnyValidate endpoint server
// request body.
func NewEndpointQueryAnyValidatePayload(q interface{}) (*EndpointQueryAnyValidatePayload, error) {
	p := EndpointQueryAnyValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayBoolConstructorCode = `// NewEndpointQueryArrayBoolPayload instantiates and validates the
// ServiceQueryArrayBool service EndpointQueryArrayBool endpoint server request
// body.
func NewEndpointQueryArrayBoolPayload(q []bool) (*EndpointQueryArrayBoolPayload, error) {
	p := EndpointQueryArrayBoolPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayBoolValidateConstructorCode = `// NewEndpointQueryArrayBoolValidatePayload instantiates and validates the
// ServiceQueryArrayBoolValidate service EndpointQueryArrayBoolValidate
// endpoint server request body.
func NewEndpointQueryArrayBoolValidatePayload(q []bool) (*EndpointQueryArrayBoolValidatePayload, error) {
	p := EndpointQueryArrayBoolValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayIntConstructorCode = `// NewEndpointQueryArrayIntPayload instantiates and validates the
// ServiceQueryArrayInt service EndpointQueryArrayInt endpoint server request
// body.
func NewEndpointQueryArrayIntPayload(q []int) (*EndpointQueryArrayIntPayload, error) {
	p := EndpointQueryArrayIntPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayIntValidateConstructorCode = `// NewEndpointQueryArrayIntValidatePayload instantiates and validates the
// ServiceQueryArrayIntValidate service EndpointQueryArrayIntValidate endpoint
// server request body.
func NewEndpointQueryArrayIntValidatePayload(q []int) (*EndpointQueryArrayIntValidatePayload, error) {
	p := EndpointQueryArrayIntValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayInt32ConstructorCode = `// NewEndpointQueryArrayInt32Payload instantiates and validates the
// ServiceQueryArrayInt32 service EndpointQueryArrayInt32 endpoint server
// request body.
func NewEndpointQueryArrayInt32Payload(q []int32) (*EndpointQueryArrayInt32Payload, error) {
	p := EndpointQueryArrayInt32Payload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayInt32ValidateConstructorCode = `// NewEndpointQueryArrayInt32ValidatePayload instantiates and validates the
// ServiceQueryArrayInt32Validate service EndpointQueryArrayInt32Validate
// endpoint server request body.
func NewEndpointQueryArrayInt32ValidatePayload(q []int32) (*EndpointQueryArrayInt32ValidatePayload, error) {
	p := EndpointQueryArrayInt32ValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayInt64ConstructorCode = `// NewEndpointQueryArrayInt64Payload instantiates and validates the
// ServiceQueryArrayInt64 service EndpointQueryArrayInt64 endpoint server
// request body.
func NewEndpointQueryArrayInt64Payload(q []int64) (*EndpointQueryArrayInt64Payload, error) {
	p := EndpointQueryArrayInt64Payload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayInt64ValidateConstructorCode = `// NewEndpointQueryArrayInt64ValidatePayload instantiates and validates the
// ServiceQueryArrayInt64Validate service EndpointQueryArrayInt64Validate
// endpoint server request body.
func NewEndpointQueryArrayInt64ValidatePayload(q []int64) (*EndpointQueryArrayInt64ValidatePayload, error) {
	p := EndpointQueryArrayInt64ValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayUIntConstructorCode = `// NewEndpointQueryArrayUIntPayload instantiates and validates the
// ServiceQueryArrayUInt service EndpointQueryArrayUInt endpoint server request
// body.
func NewEndpointQueryArrayUIntPayload(q []uint) (*EndpointQueryArrayUIntPayload, error) {
	p := EndpointQueryArrayUIntPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayUIntValidateConstructorCode = `// NewEndpointQueryArrayUIntValidatePayload instantiates and validates the
// ServiceQueryArrayUIntValidate service EndpointQueryArrayUIntValidate
// endpoint server request body.
func NewEndpointQueryArrayUIntValidatePayload(q []uint) (*EndpointQueryArrayUIntValidatePayload, error) {
	p := EndpointQueryArrayUIntValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayUInt32ConstructorCode = `// NewEndpointQueryArrayUInt32Payload instantiates and validates the
// ServiceQueryArrayUInt32 service EndpointQueryArrayUInt32 endpoint server
// request body.
func NewEndpointQueryArrayUInt32Payload(q []uint32) (*EndpointQueryArrayUInt32Payload, error) {
	p := EndpointQueryArrayUInt32Payload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayUInt32ValidateConstructorCode = `// NewEndpointQueryArrayUInt32ValidatePayload instantiates and validates the
// ServiceQueryArrayUInt32Validate service EndpointQueryArrayUInt32Validate
// endpoint server request body.
func NewEndpointQueryArrayUInt32ValidatePayload(q []uint32) (*EndpointQueryArrayUInt32ValidatePayload, error) {
	p := EndpointQueryArrayUInt32ValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayUInt64ConstructorCode = `// NewEndpointQueryArrayUInt64Payload instantiates and validates the
// ServiceQueryArrayUInt64 service EndpointQueryArrayUInt64 endpoint server
// request body.
func NewEndpointQueryArrayUInt64Payload(q []uint64) (*EndpointQueryArrayUInt64Payload, error) {
	p := EndpointQueryArrayUInt64Payload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayUInt64ValidateConstructorCode = `// NewEndpointQueryArrayUInt64ValidatePayload instantiates and validates the
// ServiceQueryArrayUInt64Validate service EndpointQueryArrayUInt64Validate
// endpoint server request body.
func NewEndpointQueryArrayUInt64ValidatePayload(q []uint64) (*EndpointQueryArrayUInt64ValidatePayload, error) {
	p := EndpointQueryArrayUInt64ValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayFloat32ConstructorCode = `// NewEndpointQueryArrayFloat32Payload instantiates and validates the
// ServiceQueryArrayFloat32 service EndpointQueryArrayFloat32 endpoint server
// request body.
func NewEndpointQueryArrayFloat32Payload(q []float32) (*EndpointQueryArrayFloat32Payload, error) {
	p := EndpointQueryArrayFloat32Payload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayFloat32ValidateConstructorCode = `// NewEndpointQueryArrayFloat32ValidatePayload instantiates and validates the
// ServiceQueryArrayFloat32Validate service EndpointQueryArrayFloat32Validate
// endpoint server request body.
func NewEndpointQueryArrayFloat32ValidatePayload(q []float32) (*EndpointQueryArrayFloat32ValidatePayload, error) {
	p := EndpointQueryArrayFloat32ValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayFloat64ConstructorCode = `// NewEndpointQueryArrayFloat64Payload instantiates and validates the
// ServiceQueryArrayFloat64 service EndpointQueryArrayFloat64 endpoint server
// request body.
func NewEndpointQueryArrayFloat64Payload(q []float64) (*EndpointQueryArrayFloat64Payload, error) {
	p := EndpointQueryArrayFloat64Payload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayFloat64ValidateConstructorCode = `// NewEndpointQueryArrayFloat64ValidatePayload instantiates and validates the
// ServiceQueryArrayFloat64Validate service EndpointQueryArrayFloat64Validate
// endpoint server request body.
func NewEndpointQueryArrayFloat64ValidatePayload(q []float64) (*EndpointQueryArrayFloat64ValidatePayload, error) {
	p := EndpointQueryArrayFloat64ValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayStringConstructorCode = `// NewEndpointQueryArrayStringPayload instantiates and validates the
// ServiceQueryArrayString service EndpointQueryArrayString endpoint server
// request body.
func NewEndpointQueryArrayStringPayload(q []string) (*EndpointQueryArrayStringPayload, error) {
	p := EndpointQueryArrayStringPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayStringValidateConstructorCode = `// NewEndpointQueryArrayStringValidatePayload instantiates and validates the
// ServiceQueryArrayStringValidate service EndpointQueryArrayStringValidate
// endpoint server request body.
func NewEndpointQueryArrayStringValidatePayload(q []string) (*EndpointQueryArrayStringValidatePayload, error) {
	p := EndpointQueryArrayStringValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayBytesConstructorCode = `// NewEndpointQueryArrayBytesPayload instantiates and validates the
// ServiceQueryArrayBytes service EndpointQueryArrayBytes endpoint server
// request body.
func NewEndpointQueryArrayBytesPayload(q [][]byte) (*EndpointQueryArrayBytesPayload, error) {
	p := EndpointQueryArrayBytesPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayBytesValidateConstructorCode = `// NewEndpointQueryArrayBytesValidatePayload instantiates and validates the
// ServiceQueryArrayBytesValidate service EndpointQueryArrayBytesValidate
// endpoint server request body.
func NewEndpointQueryArrayBytesValidatePayload(q [][]byte) (*EndpointQueryArrayBytesValidatePayload, error) {
	p := EndpointQueryArrayBytesValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayAnyConstructorCode = `// NewEndpointQueryArrayAnyPayload instantiates and validates the
// ServiceQueryArrayAny service EndpointQueryArrayAny endpoint server request
// body.
func NewEndpointQueryArrayAnyPayload(q []interface{}) (*EndpointQueryArrayAnyPayload, error) {
	p := EndpointQueryArrayAnyPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryArrayAnyValidateConstructorCode = `// NewEndpointQueryArrayAnyValidatePayload instantiates and validates the
// ServiceQueryArrayAnyValidate service EndpointQueryArrayAnyValidate endpoint
// server request body.
func NewEndpointQueryArrayAnyValidatePayload(q []interface{}) (*EndpointQueryArrayAnyValidatePayload, error) {
	p := EndpointQueryArrayAnyValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryMapStringStringConstructorCode = `// NewEndpointQueryMapStringStringPayload instantiates and validates the
// ServiceQueryMapStringString service EndpointQueryMapStringString endpoint
// server request body.
func NewEndpointQueryMapStringStringPayload(q map[string]string) (*EndpointQueryMapStringStringPayload, error) {
	p := EndpointQueryMapStringStringPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryMapStringStringValidateConstructorCode = `// NewEndpointQueryMapStringStringValidatePayload instantiates and validates
// the ServiceQueryMapStringStringValidate service
// EndpointQueryMapStringStringValidate endpoint server request body.
func NewEndpointQueryMapStringStringValidatePayload(q map[string]string) (*EndpointQueryMapStringStringValidatePayload, error) {
	p := EndpointQueryMapStringStringValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryMapStringBoolConstructorCode = `// NewEndpointQueryMapStringBoolPayload instantiates and validates the
// ServiceQueryMapStringBool service EndpointQueryMapStringBool endpoint server
// request body.
func NewEndpointQueryMapStringBoolPayload(q map[string]bool) (*EndpointQueryMapStringBoolPayload, error) {
	p := EndpointQueryMapStringBoolPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryMapStringBoolValidateConstructorCode = `// NewEndpointQueryMapStringBoolValidatePayload instantiates and validates the
// ServiceQueryMapStringBoolValidate service EndpointQueryMapStringBoolValidate
// endpoint server request body.
func NewEndpointQueryMapStringBoolValidatePayload(q map[string]bool) (*EndpointQueryMapStringBoolValidatePayload, error) {
	p := EndpointQueryMapStringBoolValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryMapBoolStringConstructorCode = `// NewEndpointQueryMapBoolStringPayload instantiates and validates the
// ServiceQueryMapBoolString service EndpointQueryMapBoolString endpoint server
// request body.
func NewEndpointQueryMapBoolStringPayload(q map[bool]string) (*EndpointQueryMapBoolStringPayload, error) {
	p := EndpointQueryMapBoolStringPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryMapBoolStringValidateConstructorCode = `// NewEndpointQueryMapBoolStringValidatePayload instantiates and validates the
// ServiceQueryMapBoolStringValidate service EndpointQueryMapBoolStringValidate
// endpoint server request body.
func NewEndpointQueryMapBoolStringValidatePayload(q map[bool]string) (*EndpointQueryMapBoolStringValidatePayload, error) {
	p := EndpointQueryMapBoolStringValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryMapBoolBoolConstructorCode = `// NewEndpointQueryMapBoolBoolPayload instantiates and validates the
// ServiceQueryMapBoolBool service EndpointQueryMapBoolBool endpoint server
// request body.
func NewEndpointQueryMapBoolBoolPayload(q map[bool]bool) (*EndpointQueryMapBoolBoolPayload, error) {
	p := EndpointQueryMapBoolBoolPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryMapBoolBoolValidateConstructorCode = `// NewEndpointQueryMapBoolBoolValidatePayload instantiates and validates the
// ServiceQueryMapBoolBoolValidate service EndpointQueryMapBoolBoolValidate
// endpoint server request body.
func NewEndpointQueryMapBoolBoolValidatePayload(q map[bool]bool) (*EndpointQueryMapBoolBoolValidatePayload, error) {
	p := EndpointQueryMapBoolBoolValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryMapStringArrayStringConstructorCode = `// NewEndpointQueryMapStringArrayStringPayload instantiates and validates the
// ServiceQueryMapStringArrayString service EndpointQueryMapStringArrayString
// endpoint server request body.
func NewEndpointQueryMapStringArrayStringPayload(q map[string][]string) (*EndpointQueryMapStringArrayStringPayload, error) {
	p := EndpointQueryMapStringArrayStringPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryMapStringArrayStringValidateConstructorCode = `// NewEndpointQueryMapStringArrayStringValidatePayload instantiates and
// validates the ServiceQueryMapStringArrayStringValidate service
// EndpointQueryMapStringArrayStringValidate endpoint server request body.
func NewEndpointQueryMapStringArrayStringValidatePayload(q map[string][]string) (*EndpointQueryMapStringArrayStringValidatePayload, error) {
	p := EndpointQueryMapStringArrayStringValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryMapStringArrayBoolConstructorCode = `// NewEndpointQueryMapStringArrayBoolPayload instantiates and validates the
// ServiceQueryMapStringArrayBool service EndpointQueryMapStringArrayBool
// endpoint server request body.
func NewEndpointQueryMapStringArrayBoolPayload(q map[string][]bool) (*EndpointQueryMapStringArrayBoolPayload, error) {
	p := EndpointQueryMapStringArrayBoolPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryMapStringArrayBoolValidateConstructorCode = `// NewEndpointQueryMapStringArrayBoolValidatePayload instantiates and validates
// the ServiceQueryMapStringArrayBoolValidate service
// EndpointQueryMapStringArrayBoolValidate endpoint server request body.
func NewEndpointQueryMapStringArrayBoolValidatePayload(q map[string][]bool) (*EndpointQueryMapStringArrayBoolValidatePayload, error) {
	p := EndpointQueryMapStringArrayBoolValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryMapBoolArrayStringConstructorCode = `// NewEndpointQueryMapBoolArrayStringPayload instantiates and validates the
// ServiceQueryMapBoolArrayString service EndpointQueryMapBoolArrayString
// endpoint server request body.
func NewEndpointQueryMapBoolArrayStringPayload(q map[bool][]string) (*EndpointQueryMapBoolArrayStringPayload, error) {
	p := EndpointQueryMapBoolArrayStringPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryMapBoolArrayStringValidateConstructorCode = `// NewEndpointQueryMapBoolArrayStringValidatePayload instantiates and validates
// the ServiceQueryMapBoolArrayStringValidate service
// EndpointQueryMapBoolArrayStringValidate endpoint server request body.
func NewEndpointQueryMapBoolArrayStringValidatePayload(q map[bool][]string) (*EndpointQueryMapBoolArrayStringValidatePayload, error) {
	p := EndpointQueryMapBoolArrayStringValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryMapBoolArrayBoolConstructorCode = `// NewEndpointQueryMapBoolArrayBoolPayload instantiates and validates the
// ServiceQueryMapBoolArrayBool service EndpointQueryMapBoolArrayBool endpoint
// server request body.
func NewEndpointQueryMapBoolArrayBoolPayload(q map[bool][]bool) (*EndpointQueryMapBoolArrayBoolPayload, error) {
	p := EndpointQueryMapBoolArrayBoolPayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadQueryMapBoolArrayBoolValidateConstructorCode = `// NewEndpointQueryMapBoolArrayBoolValidatePayload instantiates and validates
// the ServiceQueryMapBoolArrayBoolValidate service
// EndpointQueryMapBoolArrayBoolValidate endpoint server request body.
func NewEndpointQueryMapBoolArrayBoolValidatePayload(q map[bool][]bool) (*EndpointQueryMapBoolArrayBoolValidatePayload, error) {
	p := EndpointQueryMapBoolArrayBoolValidatePayload{
		Q: q,
	}
	return &p, nil
}
`

var PayloadPathStringConstructorCode = `// NewEndpointPathStringPayload instantiates and validates the
// ServicePathString service EndpointPathString endpoint server request body.
func NewEndpointPathStringPayload(p string) (*EndpointPathStringPayload, error) {
	p := EndpointPathStringPayload{
		P: p,
	}
	return &p, nil
}
`

var PayloadPathStringValidateConstructorCode = `// NewEndpointPathStringValidatePayload instantiates and validates the
// ServicePathStringValidate service EndpointPathStringValidate endpoint server
// request body.
func NewEndpointPathStringValidatePayload(p string) (*EndpointPathStringValidatePayload, error) {
	p := EndpointPathStringValidatePayload{
		P: p,
	}
	return &p, nil
}
`

var PayloadPathArrayStringConstructorCode = `// NewEndpointPathArrayStringPayload instantiates and validates the
// ServicePathArrayString service EndpointPathArrayString endpoint server
// request body.
func NewEndpointPathArrayStringPayload(p []string) (*EndpointPathArrayStringPayload, error) {
	p := EndpointPathArrayStringPayload{
		P: p,
	}
	return &p, nil
}
`

var PayloadPathArrayStringValidateConstructorCode = `// NewEndpointPathArrayStringValidatePayload instantiates and validates the
// ServicePathArrayStringValidate service EndpointPathArrayStringValidate
// endpoint server request body.
func NewEndpointPathArrayStringValidatePayload(p []string) (*EndpointPathArrayStringValidatePayload, error) {
	p := EndpointPathArrayStringValidatePayload{
		P: p,
	}
	return &p, nil
}
`

var PayloadHeaderStringConstructorCode = `// NewEndpointHeaderStringPayload instantiates and validates the
// ServiceHeaderString service EndpointHeaderString endpoint server request
// body.
func NewEndpointHeaderStringPayload(h *string) (*EndpointHeaderStringPayload, error) {
	p := EndpointHeaderStringPayload{
		H: h,
	}
	return &p, nil
}
`

var PayloadHeaderStringValidateConstructorCode = `// NewEndpointHeaderStringValidatePayload instantiates and validates the
// ServiceHeaderStringValidate service EndpointHeaderStringValidate endpoint
// server request body.
func NewEndpointHeaderStringValidatePayload(h *string) (*EndpointHeaderStringValidatePayload, error) {
	p := EndpointHeaderStringValidatePayload{
		H: h,
	}
	return &p, nil
}
`

var PayloadHeaderArrayStringConstructorCode = `// NewEndpointHeaderArrayStringPayload instantiates and validates the
// ServiceHeaderArrayString service EndpointHeaderArrayString endpoint server
// request body.
func NewEndpointHeaderArrayStringPayload(h []string) (*EndpointHeaderArrayStringPayload, error) {
	p := EndpointHeaderArrayStringPayload{
		H: h,
	}
	return &p, nil
}
`

var PayloadHeaderArrayStringValidateConstructorCode = `// NewEndpointHeaderArrayStringValidatePayload instantiates and validates the
// ServiceHeaderArrayStringValidate service EndpointHeaderArrayStringValidate
// endpoint server request body.
func NewEndpointHeaderArrayStringValidatePayload(h []string) (*EndpointHeaderArrayStringValidatePayload, error) {
	p := EndpointHeaderArrayStringValidatePayload{
		H: h,
	}
	return &p, nil
}
`

var PayloadBodyQueryObjectConstructorCode = `// NewEndpointBodyQueryObjectPayload instantiates and validates the
// ServiceBodyQueryObject service EndpointBodyQueryObject endpoint server
// request body.
func NewEndpointBodyQueryObjectPayload(body *EndpointBodyQueryObjectServerRequestBody, b *string) (*EndpointBodyQueryObjectPayload, error) {
	p := EndpointBodyQueryObjectPayload{
		A: body.A,
		B: b,
	}
	return &p, nil
}
`

var PayloadBodyQueryObjectValidateConstructorCode = `// NewEndpointBodyQueryObjectValidatePayload instantiates and validates the
// ServiceBodyQueryObjectValidate service EndpointBodyQueryObjectValidate
// endpoint server request body.
func NewEndpointBodyQueryObjectValidatePayload(body *EndpointBodyQueryObjectValidateServerRequestBody, b string) (*EndpointBodyQueryObjectValidatePayload, error) {
	p := EndpointBodyQueryObjectValidatePayload{
		A: body.A,
		B: b,
	}
	return &p, nil
}
`

var PayloadBodyQueryUserConstructorCode = `// NewPayloadType instantiates and validates the ServiceBodyQueryUser service
// EndpointBodyQueryUser endpoint server request body.
func NewPayloadType(body *EndpointBodyQueryUserServerRequestBody, b *string) (*PayloadType, error) {
	p := PayloadType{
		A: body.A,
		B: b,
	}
	return &p, nil
}
`

var PayloadBodyQueryUserValidateConstructorCode = `// NewPayloadType instantiates and validates the ServiceBodyQueryUserValidate
// service EndpointBodyQueryUserValidate endpoint server request body.
func NewPayloadType(body *EndpointBodyQueryUserValidateServerRequestBody, b string) (*PayloadType, error) {
	p := PayloadType{
		A: body.A,
		B: b,
	}
	return &p, nil
}
`

var PayloadBodyPathObjectConstructorCode = `// NewEndpointBodyPathObjectPayload instantiates and validates the
// ServiceBodyPathObject service EndpointBodyPathObject endpoint server request
// body.
func NewEndpointBodyPathObjectPayload(body *EndpointBodyPathObjectServerRequestBody, b string) (*EndpointBodyPathObjectPayload, error) {
	p := EndpointBodyPathObjectPayload{
		A: body.A,
		B: b,
	}
	return &p, nil
}
`

var PayloadBodyPathObjectValidateConstructorCode = `// NewEndpointBodyPathObjectValidatePayload instantiates and validates the
// ServiceBodyPathObjectValidate service EndpointBodyPathObjectValidate
// endpoint server request body.
func NewEndpointBodyPathObjectValidatePayload(body *EndpointBodyPathObjectValidateServerRequestBody, b string) (*EndpointBodyPathObjectValidatePayload, error) {
	p := EndpointBodyPathObjectValidatePayload{
		A: body.A,
		B: b,
	}
	return &p, nil
}
`

var PayloadBodyPathUserConstructorCode = `// NewPayloadType instantiates and validates the ServiceBodyPathUser service
// EndpointBodyPathUser endpoint server request body.
func NewPayloadType(body *EndpointBodyPathUserServerRequestBody, b string) (*PayloadType, error) {
	p := PayloadType{
		A: body.A,
		B: b,
	}
	return &p, nil
}
`

var PayloadBodyPathUserValidateConstructorCode = `// NewPayloadType instantiates and validates the ServiceBodyPathUserValidate
// service EndpointUserBodyPathValidate endpoint server request body.
func NewPayloadType(body *EndpointUserBodyPathValidateServerRequestBody, b string) (*PayloadType, error) {
	p := PayloadType{
		A: body.A,
		B: b,
	}
	return &p, nil
}
`

var PayloadBodyQueryPathObjectConstructorCode = `// NewEndpointBodyQueryPathObjectPayload instantiates and validates the
// ServiceBodyQueryPathObject service EndpointBodyQueryPathObject endpoint
// server request body.
func NewEndpointBodyQueryPathObjectPayload(body *EndpointBodyQueryPathObjectServerRequestBody, b *string, c string) (*EndpointBodyQueryPathObjectPayload, error) {
	p := EndpointBodyQueryPathObjectPayload{
		A: body.A,
		B: b,
		C: c,
	}
	return &p, nil
}
`

var PayloadBodyQueryPathObjectValidateConstructorCode = `// NewEndpointBodyQueryPathObjectValidatePayload instantiates and validates the
// ServiceBodyQueryPathObjectValidate service
// EndpointBodyQueryPathObjectValidate endpoint server request body.
func NewEndpointBodyQueryPathObjectValidatePayload(body *EndpointBodyQueryPathObjectValidateServerRequestBody, b string, c string) (*EndpointBodyQueryPathObjectValidatePayload, error) {
	p := EndpointBodyQueryPathObjectValidatePayload{
		A: body.A,
		B: b,
		C: c,
	}
	return &p, nil
}
`

var PayloadBodyQueryPathUserConstructorCode = `// NewPayloadType instantiates and validates the ServiceBodyQueryPathUser
// service EndpointBodyQueryPathUser endpoint server request body.
func NewPayloadType(body *EndpointBodyQueryPathUserServerRequestBody, b *string, c string) (*PayloadType, error) {
	p := PayloadType{
		A: body.A,
		B: b,
		C: c,
	}
	return &p, nil
}
`

var PayloadBodyQueryPathUserValidateConstructorCode = `// NewPayloadType instantiates and validates the
// ServiceBodyQueryPathUserValidate service EndpointBodyQueryPathUserValidate
// endpoint server request body.
func NewPayloadType(body *EndpointBodyQueryPathUserValidateServerRequestBody, b string, c string) (*PayloadType, error) {
	p := PayloadType{
		A: body.A,
		B: b,
		C: c,
	}
	return &p, nil
}
`
