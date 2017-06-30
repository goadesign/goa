package testing

var PayloadQueryBoolConstructorCode = `// NewMethodQueryBoolPayload instantiates and validates the ServiceQueryBool
// service MethodQueryBool endpoint payload.
func NewMethodQueryBoolPayload(q *bool) *servicequerybool.MethodQueryBoolPayload {
	p := servicequerybool.MethodQueryBoolPayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryBoolValidateConstructorCode = `// NewMethodQueryBoolValidatePayload instantiates and validates the
// ServiceQueryBoolValidate service MethodQueryBoolValidate endpoint payload.
func NewMethodQueryBoolValidatePayload(q bool) *servicequeryboolvalidate.MethodQueryBoolValidatePayload {
	p := servicequeryboolvalidate.MethodQueryBoolValidatePayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryIntConstructorCode = `// NewMethodQueryIntPayload instantiates and validates the ServiceQueryInt
// service MethodQueryInt endpoint payload.
func NewMethodQueryIntPayload(q *int) *servicequeryint.MethodQueryIntPayload {
	p := servicequeryint.MethodQueryIntPayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryIntValidateConstructorCode = `// NewMethodQueryIntValidatePayload instantiates and validates the
// ServiceQueryIntValidate service MethodQueryIntValidate endpoint payload.
func NewMethodQueryIntValidatePayload(q int) *servicequeryintvalidate.MethodQueryIntValidatePayload {
	p := servicequeryintvalidate.MethodQueryIntValidatePayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryInt32ConstructorCode = `// NewMethodQueryInt32Payload instantiates and validates the ServiceQueryInt32
// service MethodQueryInt32 endpoint payload.
func NewMethodQueryInt32Payload(q *int32) *servicequeryint32.MethodQueryInt32Payload {
	p := servicequeryint32.MethodQueryInt32Payload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryInt32ValidateConstructorCode = `// NewMethodQueryInt32ValidatePayload instantiates and validates the
// ServiceQueryInt32Validate service MethodQueryInt32Validate endpoint payload.
func NewMethodQueryInt32ValidatePayload(q int32) *servicequeryint32validate.MethodQueryInt32ValidatePayload {
	p := servicequeryint32validate.MethodQueryInt32ValidatePayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryInt64ConstructorCode = `// NewMethodQueryInt64Payload instantiates and validates the ServiceQueryInt64
// service MethodQueryInt64 endpoint payload.
func NewMethodQueryInt64Payload(q *int64) *servicequeryint64.MethodQueryInt64Payload {
	p := servicequeryint64.MethodQueryInt64Payload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryInt64ValidateConstructorCode = `// NewMethodQueryInt64ValidatePayload instantiates and validates the
// ServiceQueryInt64Validate service MethodQueryInt64Validate endpoint payload.
func NewMethodQueryInt64ValidatePayload(q int64) *servicequeryint64validate.MethodQueryInt64ValidatePayload {
	p := servicequeryint64validate.MethodQueryInt64ValidatePayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryUIntConstructorCode = `// NewMethodQueryUIntPayload instantiates and validates the ServiceQueryUInt
// service MethodQueryUInt endpoint payload.
func NewMethodQueryUIntPayload(q *uint) *servicequeryuint.MethodQueryUIntPayload {
	p := servicequeryuint.MethodQueryUIntPayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryUIntValidateConstructorCode = `// NewMethodQueryUIntValidatePayload instantiates and validates the
// ServiceQueryUIntValidate service MethodQueryUIntValidate endpoint payload.
func NewMethodQueryUIntValidatePayload(q uint) *servicequeryuintvalidate.MethodQueryUIntValidatePayload {
	p := servicequeryuintvalidate.MethodQueryUIntValidatePayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryUInt32ConstructorCode = `// NewMethodQueryUInt32Payload instantiates and validates the
// ServiceQueryUInt32 service MethodQueryUInt32 endpoint payload.
func NewMethodQueryUInt32Payload(q *uint32) *servicequeryuint32.MethodQueryUInt32Payload {
	p := servicequeryuint32.MethodQueryUInt32Payload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryUInt32ValidateConstructorCode = `// NewMethodQueryUInt32ValidatePayload instantiates and validates the
// ServiceQueryUInt32Validate service MethodQueryUInt32Validate endpoint
// payload.
func NewMethodQueryUInt32ValidatePayload(q uint32) *servicequeryuint32validate.MethodQueryUInt32ValidatePayload {
	p := servicequeryuint32validate.MethodQueryUInt32ValidatePayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryUInt64ConstructorCode = `// NewMethodQueryUInt64Payload instantiates and validates the
// ServiceQueryUInt64 service MethodQueryUInt64 endpoint payload.
func NewMethodQueryUInt64Payload(q *uint64) *servicequeryuint64.MethodQueryUInt64Payload {
	p := servicequeryuint64.MethodQueryUInt64Payload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryUInt64ValidateConstructorCode = `// NewMethodQueryUInt64ValidatePayload instantiates and validates the
// ServiceQueryUInt64Validate service MethodQueryUInt64Validate endpoint
// payload.
func NewMethodQueryUInt64ValidatePayload(q uint64) *servicequeryuint64validate.MethodQueryUInt64ValidatePayload {
	p := servicequeryuint64validate.MethodQueryUInt64ValidatePayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryFloat32ConstructorCode = `// NewMethodQueryFloat32Payload instantiates and validates the
// ServiceQueryFloat32 service MethodQueryFloat32 endpoint payload.
func NewMethodQueryFloat32Payload(q *float32) *servicequeryfloat32.MethodQueryFloat32Payload {
	p := servicequeryfloat32.MethodQueryFloat32Payload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryFloat32ValidateConstructorCode = `// NewMethodQueryFloat32ValidatePayload instantiates and validates the
// ServiceQueryFloat32Validate service MethodQueryFloat32Validate endpoint
// payload.
func NewMethodQueryFloat32ValidatePayload(q float32) *servicequeryfloat32validate.MethodQueryFloat32ValidatePayload {
	p := servicequeryfloat32validate.MethodQueryFloat32ValidatePayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryFloat64ConstructorCode = `// NewMethodQueryFloat64Payload instantiates and validates the
// ServiceQueryFloat64 service MethodQueryFloat64 endpoint payload.
func NewMethodQueryFloat64Payload(q *float64) *servicequeryfloat64.MethodQueryFloat64Payload {
	p := servicequeryfloat64.MethodQueryFloat64Payload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryFloat64ValidateConstructorCode = `// NewMethodQueryFloat64ValidatePayload instantiates and validates the
// ServiceQueryFloat64Validate service MethodQueryFloat64Validate endpoint
// payload.
func NewMethodQueryFloat64ValidatePayload(q float64) *servicequeryfloat64validate.MethodQueryFloat64ValidatePayload {
	p := servicequeryfloat64validate.MethodQueryFloat64ValidatePayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryStringConstructorCode = `// NewMethodQueryStringPayload instantiates and validates the
// ServiceQueryString service MethodQueryString endpoint payload.
func NewMethodQueryStringPayload(q *string) *servicequerystring.MethodQueryStringPayload {
	p := servicequerystring.MethodQueryStringPayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryStringValidateConstructorCode = `// NewMethodQueryStringValidatePayload instantiates and validates the
// ServiceQueryStringValidate service MethodQueryStringValidate endpoint
// payload.
func NewMethodQueryStringValidatePayload(q string) *servicequerystringvalidate.MethodQueryStringValidatePayload {
	p := servicequerystringvalidate.MethodQueryStringValidatePayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryBytesConstructorCode = `// NewMethodQueryBytesPayload instantiates and validates the ServiceQueryBytes
// service MethodQueryBytes endpoint payload.
func NewMethodQueryBytesPayload(q []byte) *servicequerybytes.MethodQueryBytesPayload {
	p := servicequerybytes.MethodQueryBytesPayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryBytesValidateConstructorCode = `// NewMethodQueryBytesValidatePayload instantiates and validates the
// ServiceQueryBytesValidate service MethodQueryBytesValidate endpoint payload.
func NewMethodQueryBytesValidatePayload(q []byte) *servicequerybytesvalidate.MethodQueryBytesValidatePayload {
	p := servicequerybytesvalidate.MethodQueryBytesValidatePayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryAnyConstructorCode = `// NewMethodQueryAnyPayload instantiates and validates the ServiceQueryAny
// service MethodQueryAny endpoint payload.
func NewMethodQueryAnyPayload(q interface{}) *servicequeryany.MethodQueryAnyPayload {
	p := servicequeryany.MethodQueryAnyPayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryAnyValidateConstructorCode = `// NewMethodQueryAnyValidatePayload instantiates and validates the
// ServiceQueryAnyValidate service MethodQueryAnyValidate endpoint payload.
func NewMethodQueryAnyValidatePayload(q interface{}) *servicequeryanyvalidate.MethodQueryAnyValidatePayload {
	p := servicequeryanyvalidate.MethodQueryAnyValidatePayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryArrayBoolConstructorCode = `// NewMethodQueryArrayBoolPayload instantiates and validates the
// ServiceQueryArrayBool service MethodQueryArrayBool endpoint payload.
func NewMethodQueryArrayBoolPayload(q []bool) *servicequeryarraybool.MethodQueryArrayBoolPayload {
	p := servicequeryarraybool.MethodQueryArrayBoolPayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryArrayBoolValidateConstructorCode = `// NewMethodQueryArrayBoolValidatePayload instantiates and validates the
// ServiceQueryArrayBoolValidate service MethodQueryArrayBoolValidate endpoint
// payload.
func NewMethodQueryArrayBoolValidatePayload(q []bool) *servicequeryarrayboolvalidate.MethodQueryArrayBoolValidatePayload {
	p := servicequeryarrayboolvalidate.MethodQueryArrayBoolValidatePayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryArrayIntConstructorCode = `// NewMethodQueryArrayIntPayload instantiates and validates the
// ServiceQueryArrayInt service MethodQueryArrayInt endpoint payload.
func NewMethodQueryArrayIntPayload(q []int) *servicequeryarrayint.MethodQueryArrayIntPayload {
	p := servicequeryarrayint.MethodQueryArrayIntPayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryArrayIntValidateConstructorCode = `// NewMethodQueryArrayIntValidatePayload instantiates and validates the
// ServiceQueryArrayIntValidate service MethodQueryArrayIntValidate endpoint
// payload.
func NewMethodQueryArrayIntValidatePayload(q []int) *servicequeryarrayintvalidate.MethodQueryArrayIntValidatePayload {
	p := servicequeryarrayintvalidate.MethodQueryArrayIntValidatePayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryArrayInt32ConstructorCode = `// NewMethodQueryArrayInt32Payload instantiates and validates the
// ServiceQueryArrayInt32 service MethodQueryArrayInt32 endpoint payload.
func NewMethodQueryArrayInt32Payload(q []int32) *servicequeryarrayint32.MethodQueryArrayInt32Payload {
	p := servicequeryarrayint32.MethodQueryArrayInt32Payload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryArrayInt32ValidateConstructorCode = `// NewMethodQueryArrayInt32ValidatePayload instantiates and validates the
// ServiceQueryArrayInt32Validate service MethodQueryArrayInt32Validate
// endpoint payload.
func NewMethodQueryArrayInt32ValidatePayload(q []int32) *servicequeryarrayint32validate.MethodQueryArrayInt32ValidatePayload {
	p := servicequeryarrayint32validate.MethodQueryArrayInt32ValidatePayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryArrayInt64ConstructorCode = `// NewMethodQueryArrayInt64Payload instantiates and validates the
// ServiceQueryArrayInt64 service MethodQueryArrayInt64 endpoint payload.
func NewMethodQueryArrayInt64Payload(q []int64) *servicequeryarrayint64.MethodQueryArrayInt64Payload {
	p := servicequeryarrayint64.MethodQueryArrayInt64Payload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryArrayInt64ValidateConstructorCode = `// NewMethodQueryArrayInt64ValidatePayload instantiates and validates the
// ServiceQueryArrayInt64Validate service MethodQueryArrayInt64Validate
// endpoint payload.
func NewMethodQueryArrayInt64ValidatePayload(q []int64) *servicequeryarrayint64validate.MethodQueryArrayInt64ValidatePayload {
	p := servicequeryarrayint64validate.MethodQueryArrayInt64ValidatePayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryArrayUIntConstructorCode = `// NewMethodQueryArrayUIntPayload instantiates and validates the
// ServiceQueryArrayUInt service MethodQueryArrayUInt endpoint payload.
func NewMethodQueryArrayUIntPayload(q []uint) *servicequeryarrayuint.MethodQueryArrayUIntPayload {
	p := servicequeryarrayuint.MethodQueryArrayUIntPayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryArrayUIntValidateConstructorCode = `// NewMethodQueryArrayUIntValidatePayload instantiates and validates the
// ServiceQueryArrayUIntValidate service MethodQueryArrayUIntValidate endpoint
// payload.
func NewMethodQueryArrayUIntValidatePayload(q []uint) *servicequeryarrayuintvalidate.MethodQueryArrayUIntValidatePayload {
	p := servicequeryarrayuintvalidate.MethodQueryArrayUIntValidatePayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryArrayUInt32ConstructorCode = `// NewMethodQueryArrayUInt32Payload instantiates and validates the
// ServiceQueryArrayUInt32 service MethodQueryArrayUInt32 endpoint payload.
func NewMethodQueryArrayUInt32Payload(q []uint32) *servicequeryarrayuint32.MethodQueryArrayUInt32Payload {
	p := servicequeryarrayuint32.MethodQueryArrayUInt32Payload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryArrayUInt32ValidateConstructorCode = `// NewMethodQueryArrayUInt32ValidatePayload instantiates and validates the
// ServiceQueryArrayUInt32Validate service MethodQueryArrayUInt32Validate
// endpoint payload.
func NewMethodQueryArrayUInt32ValidatePayload(q []uint32) *servicequeryarrayuint32validate.MethodQueryArrayUInt32ValidatePayload {
	p := servicequeryarrayuint32validate.MethodQueryArrayUInt32ValidatePayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryArrayUInt64ConstructorCode = `// NewMethodQueryArrayUInt64Payload instantiates and validates the
// ServiceQueryArrayUInt64 service MethodQueryArrayUInt64 endpoint payload.
func NewMethodQueryArrayUInt64Payload(q []uint64) *servicequeryarrayuint64.MethodQueryArrayUInt64Payload {
	p := servicequeryarrayuint64.MethodQueryArrayUInt64Payload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryArrayUInt64ValidateConstructorCode = `// NewMethodQueryArrayUInt64ValidatePayload instantiates and validates the
// ServiceQueryArrayUInt64Validate service MethodQueryArrayUInt64Validate
// endpoint payload.
func NewMethodQueryArrayUInt64ValidatePayload(q []uint64) *servicequeryarrayuint64validate.MethodQueryArrayUInt64ValidatePayload {
	p := servicequeryarrayuint64validate.MethodQueryArrayUInt64ValidatePayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryArrayFloat32ConstructorCode = `// NewMethodQueryArrayFloat32Payload instantiates and validates the
// ServiceQueryArrayFloat32 service MethodQueryArrayFloat32 endpoint payload.
func NewMethodQueryArrayFloat32Payload(q []float32) *servicequeryarrayfloat32.MethodQueryArrayFloat32Payload {
	p := servicequeryarrayfloat32.MethodQueryArrayFloat32Payload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryArrayFloat32ValidateConstructorCode = `// NewMethodQueryArrayFloat32ValidatePayload instantiates and validates the
// ServiceQueryArrayFloat32Validate service MethodQueryArrayFloat32Validate
// endpoint payload.
func NewMethodQueryArrayFloat32ValidatePayload(q []float32) *servicequeryarrayfloat32validate.MethodQueryArrayFloat32ValidatePayload {
	p := servicequeryarrayfloat32validate.MethodQueryArrayFloat32ValidatePayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryArrayFloat64ConstructorCode = `// NewMethodQueryArrayFloat64Payload instantiates and validates the
// ServiceQueryArrayFloat64 service MethodQueryArrayFloat64 endpoint payload.
func NewMethodQueryArrayFloat64Payload(q []float64) *servicequeryarrayfloat64.MethodQueryArrayFloat64Payload {
	p := servicequeryarrayfloat64.MethodQueryArrayFloat64Payload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryArrayFloat64ValidateConstructorCode = `// NewMethodQueryArrayFloat64ValidatePayload instantiates and validates the
// ServiceQueryArrayFloat64Validate service MethodQueryArrayFloat64Validate
// endpoint payload.
func NewMethodQueryArrayFloat64ValidatePayload(q []float64) *servicequeryarrayfloat64validate.MethodQueryArrayFloat64ValidatePayload {
	p := servicequeryarrayfloat64validate.MethodQueryArrayFloat64ValidatePayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryArrayStringConstructorCode = `// NewMethodQueryArrayStringPayload instantiates and validates the
// ServiceQueryArrayString service MethodQueryArrayString endpoint payload.
func NewMethodQueryArrayStringPayload(q []string) *servicequeryarraystring.MethodQueryArrayStringPayload {
	p := servicequeryarraystring.MethodQueryArrayStringPayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryArrayStringValidateConstructorCode = `// NewMethodQueryArrayStringValidatePayload instantiates and validates the
// ServiceQueryArrayStringValidate service MethodQueryArrayStringValidate
// endpoint payload.
func NewMethodQueryArrayStringValidatePayload(q []string) *servicequeryarraystringvalidate.MethodQueryArrayStringValidatePayload {
	p := servicequeryarraystringvalidate.MethodQueryArrayStringValidatePayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryArrayBytesConstructorCode = `// NewMethodQueryArrayBytesPayload instantiates and validates the
// ServiceQueryArrayBytes service MethodQueryArrayBytes endpoint payload.
func NewMethodQueryArrayBytesPayload(q [][]byte) *servicequeryarraybytes.MethodQueryArrayBytesPayload {
	p := servicequeryarraybytes.MethodQueryArrayBytesPayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryArrayBytesValidateConstructorCode = `// NewMethodQueryArrayBytesValidatePayload instantiates and validates the
// ServiceQueryArrayBytesValidate service MethodQueryArrayBytesValidate
// endpoint payload.
func NewMethodQueryArrayBytesValidatePayload(q [][]byte) *servicequeryarraybytesvalidate.MethodQueryArrayBytesValidatePayload {
	p := servicequeryarraybytesvalidate.MethodQueryArrayBytesValidatePayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryArrayAnyConstructorCode = `// NewMethodQueryArrayAnyPayload instantiates and validates the
// ServiceQueryArrayAny service MethodQueryArrayAny endpoint payload.
func NewMethodQueryArrayAnyPayload(q []interface{}) *servicequeryarrayany.MethodQueryArrayAnyPayload {
	p := servicequeryarrayany.MethodQueryArrayAnyPayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryArrayAnyValidateConstructorCode = `// NewMethodQueryArrayAnyValidatePayload instantiates and validates the
// ServiceQueryArrayAnyValidate service MethodQueryArrayAnyValidate endpoint
// payload.
func NewMethodQueryArrayAnyValidatePayload(q []interface{}) *servicequeryarrayanyvalidate.MethodQueryArrayAnyValidatePayload {
	p := servicequeryarrayanyvalidate.MethodQueryArrayAnyValidatePayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryMapStringStringConstructorCode = `// NewMethodQueryMapStringStringPayload instantiates and validates the
// ServiceQueryMapStringString service MethodQueryMapStringString endpoint
// payload.
func NewMethodQueryMapStringStringPayload(q map[string]string) *servicequerymapstringstring.MethodQueryMapStringStringPayload {
	p := servicequerymapstringstring.MethodQueryMapStringStringPayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryMapStringStringValidateConstructorCode = `// NewMethodQueryMapStringStringValidatePayload instantiates and validates the
// ServiceQueryMapStringStringValidate service
// MethodQueryMapStringStringValidate endpoint payload.
func NewMethodQueryMapStringStringValidatePayload(q map[string]string) *servicequerymapstringstringvalidate.MethodQueryMapStringStringValidatePayload {
	p := servicequerymapstringstringvalidate.MethodQueryMapStringStringValidatePayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryMapStringBoolConstructorCode = `// NewMethodQueryMapStringBoolPayload instantiates and validates the
// ServiceQueryMapStringBool service MethodQueryMapStringBool endpoint payload.
func NewMethodQueryMapStringBoolPayload(q map[string]bool) *servicequerymapstringbool.MethodQueryMapStringBoolPayload {
	p := servicequerymapstringbool.MethodQueryMapStringBoolPayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryMapStringBoolValidateConstructorCode = `// NewMethodQueryMapStringBoolValidatePayload instantiates and validates the
// ServiceQueryMapStringBoolValidate service MethodQueryMapStringBoolValidate
// endpoint payload.
func NewMethodQueryMapStringBoolValidatePayload(q map[string]bool) *servicequerymapstringboolvalidate.MethodQueryMapStringBoolValidatePayload {
	p := servicequerymapstringboolvalidate.MethodQueryMapStringBoolValidatePayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryMapBoolStringConstructorCode = `// NewMethodQueryMapBoolStringPayload instantiates and validates the
// ServiceQueryMapBoolString service MethodQueryMapBoolString endpoint payload.
func NewMethodQueryMapBoolStringPayload(q map[bool]string) *servicequerymapboolstring.MethodQueryMapBoolStringPayload {
	p := servicequerymapboolstring.MethodQueryMapBoolStringPayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryMapBoolStringValidateConstructorCode = `// NewMethodQueryMapBoolStringValidatePayload instantiates and validates the
// ServiceQueryMapBoolStringValidate service MethodQueryMapBoolStringValidate
// endpoint payload.
func NewMethodQueryMapBoolStringValidatePayload(q map[bool]string) *servicequerymapboolstringvalidate.MethodQueryMapBoolStringValidatePayload {
	p := servicequerymapboolstringvalidate.MethodQueryMapBoolStringValidatePayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryMapBoolBoolConstructorCode = `// NewMethodQueryMapBoolBoolPayload instantiates and validates the
// ServiceQueryMapBoolBool service MethodQueryMapBoolBool endpoint payload.
func NewMethodQueryMapBoolBoolPayload(q map[bool]bool) *servicequerymapboolbool.MethodQueryMapBoolBoolPayload {
	p := servicequerymapboolbool.MethodQueryMapBoolBoolPayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryMapBoolBoolValidateConstructorCode = `// NewMethodQueryMapBoolBoolValidatePayload instantiates and validates the
// ServiceQueryMapBoolBoolValidate service MethodQueryMapBoolBoolValidate
// endpoint payload.
func NewMethodQueryMapBoolBoolValidatePayload(q map[bool]bool) *servicequerymapboolboolvalidate.MethodQueryMapBoolBoolValidatePayload {
	p := servicequerymapboolboolvalidate.MethodQueryMapBoolBoolValidatePayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryMapStringArrayStringConstructorCode = `// NewMethodQueryMapStringArrayStringPayload instantiates and validates the
// ServiceQueryMapStringArrayString service MethodQueryMapStringArrayString
// endpoint payload.
func NewMethodQueryMapStringArrayStringPayload(q map[string][]string) *servicequerymapstringarraystring.MethodQueryMapStringArrayStringPayload {
	p := servicequerymapstringarraystring.MethodQueryMapStringArrayStringPayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryMapStringArrayStringValidateConstructorCode = `// NewMethodQueryMapStringArrayStringValidatePayload instantiates and validates
// the ServiceQueryMapStringArrayStringValidate service
// MethodQueryMapStringArrayStringValidate endpoint payload.
func NewMethodQueryMapStringArrayStringValidatePayload(q map[string][]string) *servicequerymapstringarraystringvalidate.MethodQueryMapStringArrayStringValidatePayload {
	p := servicequerymapstringarraystringvalidate.MethodQueryMapStringArrayStringValidatePayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryMapStringArrayBoolConstructorCode = `// NewMethodQueryMapStringArrayBoolPayload instantiates and validates the
// ServiceQueryMapStringArrayBool service MethodQueryMapStringArrayBool
// endpoint payload.
func NewMethodQueryMapStringArrayBoolPayload(q map[string][]bool) *servicequerymapstringarraybool.MethodQueryMapStringArrayBoolPayload {
	p := servicequerymapstringarraybool.MethodQueryMapStringArrayBoolPayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryMapStringArrayBoolValidateConstructorCode = `// NewMethodQueryMapStringArrayBoolValidatePayload instantiates and validates
// the ServiceQueryMapStringArrayBoolValidate service
// MethodQueryMapStringArrayBoolValidate endpoint payload.
func NewMethodQueryMapStringArrayBoolValidatePayload(q map[string][]bool) *servicequerymapstringarrayboolvalidate.MethodQueryMapStringArrayBoolValidatePayload {
	p := servicequerymapstringarrayboolvalidate.MethodQueryMapStringArrayBoolValidatePayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryMapBoolArrayStringConstructorCode = `// NewMethodQueryMapBoolArrayStringPayload instantiates and validates the
// ServiceQueryMapBoolArrayString service MethodQueryMapBoolArrayString
// endpoint payload.
func NewMethodQueryMapBoolArrayStringPayload(q map[bool][]string) *servicequerymapboolarraystring.MethodQueryMapBoolArrayStringPayload {
	p := servicequerymapboolarraystring.MethodQueryMapBoolArrayStringPayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryMapBoolArrayStringValidateConstructorCode = `// NewMethodQueryMapBoolArrayStringValidatePayload instantiates and validates
// the ServiceQueryMapBoolArrayStringValidate service
// MethodQueryMapBoolArrayStringValidate endpoint payload.
func NewMethodQueryMapBoolArrayStringValidatePayload(q map[bool][]string) *servicequerymapboolarraystringvalidate.MethodQueryMapBoolArrayStringValidatePayload {
	p := servicequerymapboolarraystringvalidate.MethodQueryMapBoolArrayStringValidatePayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryMapBoolArrayBoolConstructorCode = `// NewMethodQueryMapBoolArrayBoolPayload instantiates and validates the
// ServiceQueryMapBoolArrayBool service MethodQueryMapBoolArrayBool endpoint
// payload.
func NewMethodQueryMapBoolArrayBoolPayload(q map[bool][]bool) *servicequerymapboolarraybool.MethodQueryMapBoolArrayBoolPayload {
	p := servicequerymapboolarraybool.MethodQueryMapBoolArrayBoolPayload{
		Q: q,
	}
	return &p
}
`

var PayloadQueryMapBoolArrayBoolValidateConstructorCode = `// NewMethodQueryMapBoolArrayBoolValidatePayload instantiates and validates the
// ServiceQueryMapBoolArrayBoolValidate service
// MethodQueryMapBoolArrayBoolValidate endpoint payload.
func NewMethodQueryMapBoolArrayBoolValidatePayload(q map[bool][]bool) *servicequerymapboolarrayboolvalidate.MethodQueryMapBoolArrayBoolValidatePayload {
	p := servicequerymapboolarrayboolvalidate.MethodQueryMapBoolArrayBoolValidatePayload{
		Q: q,
	}
	return &p
}
`

var PayloadPathStringConstructorCode = `// NewMethodPathStringPayload instantiates and validates the ServicePathString
// service MethodPathString endpoint payload.
func NewMethodPathStringPayload(p string) *servicepathstring.MethodPathStringPayload {
	p := servicepathstring.MethodPathStringPayload{
		P: &p,
	}
	return &p
}
`

var PayloadPathStringValidateConstructorCode = `// NewMethodPathStringValidatePayload instantiates and validates the
// ServicePathStringValidate service MethodPathStringValidate endpoint payload.
func NewMethodPathStringValidatePayload(p string) *servicepathstringvalidate.MethodPathStringValidatePayload {
	p := servicepathstringvalidate.MethodPathStringValidatePayload{
		P: p,
	}
	return &p
}
`

var PayloadPathArrayStringConstructorCode = `// NewMethodPathArrayStringPayload instantiates and validates the
// ServicePathArrayString service MethodPathArrayString endpoint payload.
func NewMethodPathArrayStringPayload(p []string) *servicepatharraystring.MethodPathArrayStringPayload {
	p := servicepatharraystring.MethodPathArrayStringPayload{
		P: p,
	}
	return &p
}
`

var PayloadPathArrayStringValidateConstructorCode = `// NewMethodPathArrayStringValidatePayload instantiates and validates the
// ServicePathArrayStringValidate service MethodPathArrayStringValidate
// endpoint payload.
func NewMethodPathArrayStringValidatePayload(p []string) *servicepatharraystringvalidate.MethodPathArrayStringValidatePayload {
	p := servicepatharraystringvalidate.MethodPathArrayStringValidatePayload{
		P: p,
	}
	return &p
}
`

var PayloadHeaderStringConstructorCode = `// NewMethodHeaderStringPayload instantiates and validates the
// ServiceHeaderString service MethodHeaderString endpoint payload.
func NewMethodHeaderStringPayload(h *string) *serviceheaderstring.MethodHeaderStringPayload {
	p := serviceheaderstring.MethodHeaderStringPayload{
		H: h,
	}
	return &p
}
`

var PayloadHeaderStringValidateConstructorCode = `// NewMethodHeaderStringValidatePayload instantiates and validates the
// ServiceHeaderStringValidate service MethodHeaderStringValidate endpoint
// payload.
func NewMethodHeaderStringValidatePayload(h *string) *serviceheaderstringvalidate.MethodHeaderStringValidatePayload {
	p := serviceheaderstringvalidate.MethodHeaderStringValidatePayload{
		H: h,
	}
	return &p
}
`

var PayloadHeaderArrayStringConstructorCode = `// NewMethodHeaderArrayStringPayload instantiates and validates the
// ServiceHeaderArrayString service MethodHeaderArrayString endpoint payload.
func NewMethodHeaderArrayStringPayload(h []string) *serviceheaderarraystring.MethodHeaderArrayStringPayload {
	p := serviceheaderarraystring.MethodHeaderArrayStringPayload{
		H: h,
	}
	return &p
}
`

var PayloadHeaderArrayStringValidateConstructorCode = `// NewMethodHeaderArrayStringValidatePayload instantiates and validates the
// ServiceHeaderArrayStringValidate service MethodHeaderArrayStringValidate
// endpoint payload.
func NewMethodHeaderArrayStringValidatePayload(h []string) *serviceheaderarraystringvalidate.MethodHeaderArrayStringValidatePayload {
	p := serviceheaderarraystringvalidate.MethodHeaderArrayStringValidatePayload{
		H: h,
	}
	return &p
}
`

var PayloadBodyQueryObjectConstructorCode = `// NewMethodBodyQueryObjectPayload instantiates and validates the
// ServiceBodyQueryObject service MethodBodyQueryObject endpoint payload.
func NewMethodBodyQueryObjectPayload(body *MethodBodyQueryObjectServerRequestBody, b *string) *servicebodyqueryobject.MethodBodyQueryObjectPayload {
	p := servicebodyqueryobject.MethodBodyQueryObjectPayload{
		A: body.A,
		B: b,
	}
	return &p
}
`

var PayloadBodyQueryObjectValidateConstructorCode = `// NewMethodBodyQueryObjectValidatePayload instantiates and validates the
// ServiceBodyQueryObjectValidate service MethodBodyQueryObjectValidate
// endpoint payload.
func NewMethodBodyQueryObjectValidatePayload(body *MethodBodyQueryObjectValidateServerRequestBody, b string) *servicebodyqueryobjectvalidate.MethodBodyQueryObjectValidatePayload {
	p := servicebodyqueryobjectvalidate.MethodBodyQueryObjectValidatePayload{
		A: *body.A,
		B: b,
	}
	return &p
}
`

var PayloadBodyQueryUserConstructorCode = `// NewPayloadType instantiates and validates the ServiceBodyQueryUser service
// MethodBodyQueryUser endpoint payload.
func NewPayloadType(body *MethodBodyQueryUserServerRequestBody, b *string) *servicebodyqueryuser.PayloadType {
	p := servicebodyqueryuser.PayloadType{
		A: body.A,
		B: b,
	}
	return &p
}
`

var PayloadBodyQueryUserValidateConstructorCode = `// NewPayloadType instantiates and validates the ServiceBodyQueryUserValidate
// service MethodBodyQueryUserValidate endpoint payload.
func NewPayloadType(body *MethodBodyQueryUserValidateServerRequestBody, b string) *servicebodyqueryuservalidate.PayloadType {
	p := servicebodyqueryuservalidate.PayloadType{
		A: *body.A,
		B: b,
	}
	return &p
}
`

var PayloadBodyPathObjectConstructorCode = `// NewMethodBodyPathObjectPayload instantiates and validates the
// ServiceBodyPathObject service MethodBodyPathObject endpoint payload.
func NewMethodBodyPathObjectPayload(body *MethodBodyPathObjectServerRequestBody, b string) *servicebodypathobject.MethodBodyPathObjectPayload {
	p := servicebodypathobject.MethodBodyPathObjectPayload{
		A: body.A,
		B: &b,
	}
	return &p
}
`

var PayloadBodyPathObjectValidateConstructorCode = `// NewMethodBodyPathObjectValidatePayload instantiates and validates the
// ServiceBodyPathObjectValidate service MethodBodyPathObjectValidate endpoint
// payload.
func NewMethodBodyPathObjectValidatePayload(body *MethodBodyPathObjectValidateServerRequestBody, b string) *servicebodypathobjectvalidate.MethodBodyPathObjectValidatePayload {
	p := servicebodypathobjectvalidate.MethodBodyPathObjectValidatePayload{
		A: *body.A,
		B: b,
	}
	return &p
}
`

var PayloadBodyPathUserConstructorCode = `// NewPayloadType instantiates and validates the ServiceBodyPathUser service
// MethodBodyPathUser endpoint payload.
func NewPayloadType(body *MethodBodyPathUserServerRequestBody, b string) *servicebodypathuser.PayloadType {
	p := servicebodypathuser.PayloadType{
		A: body.A,
		B: &b,
	}
	return &p
}
`

var PayloadBodyPathUserValidateConstructorCode = `// NewPayloadType instantiates and validates the ServiceBodyPathUserValidate
// service MethodUserBodyPathValidate endpoint payload.
func NewPayloadType(body *MethodUserBodyPathValidateServerRequestBody, b string) *servicebodypathuservalidate.PayloadType {
	p := servicebodypathuservalidate.PayloadType{
		A: *body.A,
		B: b,
	}
	return &p
}
`

var PayloadBodyQueryPathObjectConstructorCode = `// NewMethodBodyQueryPathObjectPayload instantiates and validates the
// ServiceBodyQueryPathObject service MethodBodyQueryPathObject endpoint
// payload.
func NewMethodBodyQueryPathObjectPayload(body *MethodBodyQueryPathObjectServerRequestBody, b *string, c string) *servicebodyquerypathobject.MethodBodyQueryPathObjectPayload {
	p := servicebodyquerypathobject.MethodBodyQueryPathObjectPayload{
		A: body.A,
		B: b,
		C: &c,
	}
	return &p
}
`

var PayloadBodyQueryPathObjectValidateConstructorCode = `// NewMethodBodyQueryPathObjectValidatePayload instantiates and validates the
// ServiceBodyQueryPathObjectValidate service MethodBodyQueryPathObjectValidate
// endpoint payload.
func NewMethodBodyQueryPathObjectValidatePayload(body *MethodBodyQueryPathObjectValidateServerRequestBody, b string, c string) *servicebodyquerypathobjectvalidate.MethodBodyQueryPathObjectValidatePayload {
	p := servicebodyquerypathobjectvalidate.MethodBodyQueryPathObjectValidatePayload{
		A: *body.A,
		B: b,
		C: c,
	}
	return &p
}
`

var PayloadBodyQueryPathUserConstructorCode = `// NewPayloadType instantiates and validates the ServiceBodyQueryPathUser
// service MethodBodyQueryPathUser endpoint payload.
func NewPayloadType(body *MethodBodyQueryPathUserServerRequestBody, b *string, c string) *servicebodyquerypathuser.PayloadType {
	p := servicebodyquerypathuser.PayloadType{
		A: body.A,
		B: b,
		C: &c,
	}
	return &p
}
`

var PayloadBodyQueryPathUserValidateConstructorCode = `// NewPayloadType instantiates and validates the
// ServiceBodyQueryPathUserValidate service MethodBodyQueryPathUserValidate
// endpoint payload.
func NewPayloadType(body *MethodBodyQueryPathUserValidateServerRequestBody, b string, c string) *servicebodyquerypathuservalidate.PayloadType {
	p := servicebodyquerypathuservalidate.PayloadType{
		A: *body.A,
		B: b,
		C: c,
	}
	return &p
}
`
