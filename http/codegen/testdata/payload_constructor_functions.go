package testdata

var PayloadQueryBoolConstructorCode = `// NewMethodQueryBoolPayload builds a ServiceQueryBool service MethodQueryBool
// endpoint payload.
func NewMethodQueryBoolPayload(q *bool) *servicequerybool.MethodQueryBoolPayload {
	return &servicequerybool.MethodQueryBoolPayload{
		Q: q,
	}
}
`

var PayloadQueryBoolValidateConstructorCode = `// NewMethodQueryBoolValidatePayload builds a ServiceQueryBoolValidate service
// MethodQueryBoolValidate endpoint payload.
func NewMethodQueryBoolValidatePayload(q bool) *servicequeryboolvalidate.MethodQueryBoolValidatePayload {
	return &servicequeryboolvalidate.MethodQueryBoolValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryIntConstructorCode = `// NewMethodQueryIntPayload builds a ServiceQueryInt service MethodQueryInt
// endpoint payload.
func NewMethodQueryIntPayload(q *int) *servicequeryint.MethodQueryIntPayload {
	return &servicequeryint.MethodQueryIntPayload{
		Q: q,
	}
}
`

var PayloadQueryIntValidateConstructorCode = `// NewMethodQueryIntValidatePayload builds a ServiceQueryIntValidate service
// MethodQueryIntValidate endpoint payload.
func NewMethodQueryIntValidatePayload(q int) *servicequeryintvalidate.MethodQueryIntValidatePayload {
	return &servicequeryintvalidate.MethodQueryIntValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryInt32ConstructorCode = `// NewMethodQueryInt32Payload builds a ServiceQueryInt32 service
// MethodQueryInt32 endpoint payload.
func NewMethodQueryInt32Payload(q *int32) *servicequeryint32.MethodQueryInt32Payload {
	return &servicequeryint32.MethodQueryInt32Payload{
		Q: q,
	}
}
`

var PayloadQueryInt32ValidateConstructorCode = `// NewMethodQueryInt32ValidatePayload builds a ServiceQueryInt32Validate
// service MethodQueryInt32Validate endpoint payload.
func NewMethodQueryInt32ValidatePayload(q int32) *servicequeryint32validate.MethodQueryInt32ValidatePayload {
	return &servicequeryint32validate.MethodQueryInt32ValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryInt64ConstructorCode = `// NewMethodQueryInt64Payload builds a ServiceQueryInt64 service
// MethodQueryInt64 endpoint payload.
func NewMethodQueryInt64Payload(q *int64) *servicequeryint64.MethodQueryInt64Payload {
	return &servicequeryint64.MethodQueryInt64Payload{
		Q: q,
	}
}
`

var PayloadQueryInt64ValidateConstructorCode = `// NewMethodQueryInt64ValidatePayload builds a ServiceQueryInt64Validate
// service MethodQueryInt64Validate endpoint payload.
func NewMethodQueryInt64ValidatePayload(q int64) *servicequeryint64validate.MethodQueryInt64ValidatePayload {
	return &servicequeryint64validate.MethodQueryInt64ValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryUIntConstructorCode = `// NewMethodQueryUIntPayload builds a ServiceQueryUInt service MethodQueryUInt
// endpoint payload.
func NewMethodQueryUIntPayload(q *uint) *servicequeryuint.MethodQueryUIntPayload {
	return &servicequeryuint.MethodQueryUIntPayload{
		Q: q,
	}
}
`

var PayloadQueryUIntValidateConstructorCode = `// NewMethodQueryUIntValidatePayload builds a ServiceQueryUIntValidate service
// MethodQueryUIntValidate endpoint payload.
func NewMethodQueryUIntValidatePayload(q uint) *servicequeryuintvalidate.MethodQueryUIntValidatePayload {
	return &servicequeryuintvalidate.MethodQueryUIntValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryUInt32ConstructorCode = `// NewMethodQueryUInt32Payload builds a ServiceQueryUInt32 service
// MethodQueryUInt32 endpoint payload.
func NewMethodQueryUInt32Payload(q *uint32) *servicequeryuint32.MethodQueryUInt32Payload {
	return &servicequeryuint32.MethodQueryUInt32Payload{
		Q: q,
	}
}
`

var PayloadQueryUInt32ValidateConstructorCode = `// NewMethodQueryUInt32ValidatePayload builds a ServiceQueryUInt32Validate
// service MethodQueryUInt32Validate endpoint payload.
func NewMethodQueryUInt32ValidatePayload(q uint32) *servicequeryuint32validate.MethodQueryUInt32ValidatePayload {
	return &servicequeryuint32validate.MethodQueryUInt32ValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryUInt64ConstructorCode = `// NewMethodQueryUInt64Payload builds a ServiceQueryUInt64 service
// MethodQueryUInt64 endpoint payload.
func NewMethodQueryUInt64Payload(q *uint64) *servicequeryuint64.MethodQueryUInt64Payload {
	return &servicequeryuint64.MethodQueryUInt64Payload{
		Q: q,
	}
}
`

var PayloadQueryUInt64ValidateConstructorCode = `// NewMethodQueryUInt64ValidatePayload builds a ServiceQueryUInt64Validate
// service MethodQueryUInt64Validate endpoint payload.
func NewMethodQueryUInt64ValidatePayload(q uint64) *servicequeryuint64validate.MethodQueryUInt64ValidatePayload {
	return &servicequeryuint64validate.MethodQueryUInt64ValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryFloat32ConstructorCode = `// NewMethodQueryFloat32Payload builds a ServiceQueryFloat32 service
// MethodQueryFloat32 endpoint payload.
func NewMethodQueryFloat32Payload(q *float32) *servicequeryfloat32.MethodQueryFloat32Payload {
	return &servicequeryfloat32.MethodQueryFloat32Payload{
		Q: q,
	}
}
`

var PayloadQueryFloat32ValidateConstructorCode = `// NewMethodQueryFloat32ValidatePayload builds a ServiceQueryFloat32Validate
// service MethodQueryFloat32Validate endpoint payload.
func NewMethodQueryFloat32ValidatePayload(q float32) *servicequeryfloat32validate.MethodQueryFloat32ValidatePayload {
	return &servicequeryfloat32validate.MethodQueryFloat32ValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryFloat64ConstructorCode = `// NewMethodQueryFloat64Payload builds a ServiceQueryFloat64 service
// MethodQueryFloat64 endpoint payload.
func NewMethodQueryFloat64Payload(q *float64) *servicequeryfloat64.MethodQueryFloat64Payload {
	return &servicequeryfloat64.MethodQueryFloat64Payload{
		Q: q,
	}
}
`

var PayloadQueryFloat64ValidateConstructorCode = `// NewMethodQueryFloat64ValidatePayload builds a ServiceQueryFloat64Validate
// service MethodQueryFloat64Validate endpoint payload.
func NewMethodQueryFloat64ValidatePayload(q float64) *servicequeryfloat64validate.MethodQueryFloat64ValidatePayload {
	return &servicequeryfloat64validate.MethodQueryFloat64ValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryStringConstructorCode = `// NewMethodQueryStringPayload builds a ServiceQueryString service
// MethodQueryString endpoint payload.
func NewMethodQueryStringPayload(q *string) *servicequerystring.MethodQueryStringPayload {
	return &servicequerystring.MethodQueryStringPayload{
		Q: q,
	}
}
`

var PayloadQueryStringValidateConstructorCode = `// NewMethodQueryStringValidatePayload builds a ServiceQueryStringValidate
// service MethodQueryStringValidate endpoint payload.
func NewMethodQueryStringValidatePayload(q string) *servicequerystringvalidate.MethodQueryStringValidatePayload {
	return &servicequerystringvalidate.MethodQueryStringValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryBytesConstructorCode = `// NewMethodQueryBytesPayload builds a ServiceQueryBytes service
// MethodQueryBytes endpoint payload.
func NewMethodQueryBytesPayload(q []byte) *servicequerybytes.MethodQueryBytesPayload {
	return &servicequerybytes.MethodQueryBytesPayload{
		Q: q,
	}
}
`

var PayloadQueryBytesValidateConstructorCode = `// NewMethodQueryBytesValidatePayload builds a ServiceQueryBytesValidate
// service MethodQueryBytesValidate endpoint payload.
func NewMethodQueryBytesValidatePayload(q []byte) *servicequerybytesvalidate.MethodQueryBytesValidatePayload {
	return &servicequerybytesvalidate.MethodQueryBytesValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryAnyConstructorCode = `// NewMethodQueryAnyPayload builds a ServiceQueryAny service MethodQueryAny
// endpoint payload.
func NewMethodQueryAnyPayload(q interface{}) *servicequeryany.MethodQueryAnyPayload {
	return &servicequeryany.MethodQueryAnyPayload{
		Q: q,
	}
}
`

var PayloadQueryAnyValidateConstructorCode = `// NewMethodQueryAnyValidatePayload builds a ServiceQueryAnyValidate service
// MethodQueryAnyValidate endpoint payload.
func NewMethodQueryAnyValidatePayload(q interface{}) *servicequeryanyvalidate.MethodQueryAnyValidatePayload {
	return &servicequeryanyvalidate.MethodQueryAnyValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryArrayBoolConstructorCode = `// NewMethodQueryArrayBoolPayload builds a ServiceQueryArrayBool service
// MethodQueryArrayBool endpoint payload.
func NewMethodQueryArrayBoolPayload(q []bool) *servicequeryarraybool.MethodQueryArrayBoolPayload {
	return &servicequeryarraybool.MethodQueryArrayBoolPayload{
		Q: q,
	}
}
`

var PayloadQueryArrayBoolValidateConstructorCode = `// NewMethodQueryArrayBoolValidatePayload builds a
// ServiceQueryArrayBoolValidate service MethodQueryArrayBoolValidate endpoint
// payload.
func NewMethodQueryArrayBoolValidatePayload(q []bool) *servicequeryarrayboolvalidate.MethodQueryArrayBoolValidatePayload {
	return &servicequeryarrayboolvalidate.MethodQueryArrayBoolValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryArrayIntConstructorCode = `// NewMethodQueryArrayIntPayload builds a ServiceQueryArrayInt service
// MethodQueryArrayInt endpoint payload.
func NewMethodQueryArrayIntPayload(q []int) *servicequeryarrayint.MethodQueryArrayIntPayload {
	return &servicequeryarrayint.MethodQueryArrayIntPayload{
		Q: q,
	}
}
`

var PayloadQueryArrayIntValidateConstructorCode = `// NewMethodQueryArrayIntValidatePayload builds a ServiceQueryArrayIntValidate
// service MethodQueryArrayIntValidate endpoint payload.
func NewMethodQueryArrayIntValidatePayload(q []int) *servicequeryarrayintvalidate.MethodQueryArrayIntValidatePayload {
	return &servicequeryarrayintvalidate.MethodQueryArrayIntValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryArrayInt32ConstructorCode = `// NewMethodQueryArrayInt32Payload builds a ServiceQueryArrayInt32 service
// MethodQueryArrayInt32 endpoint payload.
func NewMethodQueryArrayInt32Payload(q []int32) *servicequeryarrayint32.MethodQueryArrayInt32Payload {
	return &servicequeryarrayint32.MethodQueryArrayInt32Payload{
		Q: q,
	}
}
`

var PayloadQueryArrayInt32ValidateConstructorCode = `// NewMethodQueryArrayInt32ValidatePayload builds a
// ServiceQueryArrayInt32Validate service MethodQueryArrayInt32Validate
// endpoint payload.
func NewMethodQueryArrayInt32ValidatePayload(q []int32) *servicequeryarrayint32validate.MethodQueryArrayInt32ValidatePayload {
	return &servicequeryarrayint32validate.MethodQueryArrayInt32ValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryArrayInt64ConstructorCode = `// NewMethodQueryArrayInt64Payload builds a ServiceQueryArrayInt64 service
// MethodQueryArrayInt64 endpoint payload.
func NewMethodQueryArrayInt64Payload(q []int64) *servicequeryarrayint64.MethodQueryArrayInt64Payload {
	return &servicequeryarrayint64.MethodQueryArrayInt64Payload{
		Q: q,
	}
}
`

var PayloadQueryArrayInt64ValidateConstructorCode = `// NewMethodQueryArrayInt64ValidatePayload builds a
// ServiceQueryArrayInt64Validate service MethodQueryArrayInt64Validate
// endpoint payload.
func NewMethodQueryArrayInt64ValidatePayload(q []int64) *servicequeryarrayint64validate.MethodQueryArrayInt64ValidatePayload {
	return &servicequeryarrayint64validate.MethodQueryArrayInt64ValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryArrayUIntConstructorCode = `// NewMethodQueryArrayUIntPayload builds a ServiceQueryArrayUInt service
// MethodQueryArrayUInt endpoint payload.
func NewMethodQueryArrayUIntPayload(q []uint) *servicequeryarrayuint.MethodQueryArrayUIntPayload {
	return &servicequeryarrayuint.MethodQueryArrayUIntPayload{
		Q: q,
	}
}
`

var PayloadQueryArrayUIntValidateConstructorCode = `// NewMethodQueryArrayUIntValidatePayload builds a
// ServiceQueryArrayUIntValidate service MethodQueryArrayUIntValidate endpoint
// payload.
func NewMethodQueryArrayUIntValidatePayload(q []uint) *servicequeryarrayuintvalidate.MethodQueryArrayUIntValidatePayload {
	return &servicequeryarrayuintvalidate.MethodQueryArrayUIntValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryArrayUInt32ConstructorCode = `// NewMethodQueryArrayUInt32Payload builds a ServiceQueryArrayUInt32 service
// MethodQueryArrayUInt32 endpoint payload.
func NewMethodQueryArrayUInt32Payload(q []uint32) *servicequeryarrayuint32.MethodQueryArrayUInt32Payload {
	return &servicequeryarrayuint32.MethodQueryArrayUInt32Payload{
		Q: q,
	}
}
`

var PayloadQueryArrayUInt32ValidateConstructorCode = `// NewMethodQueryArrayUInt32ValidatePayload builds a
// ServiceQueryArrayUInt32Validate service MethodQueryArrayUInt32Validate
// endpoint payload.
func NewMethodQueryArrayUInt32ValidatePayload(q []uint32) *servicequeryarrayuint32validate.MethodQueryArrayUInt32ValidatePayload {
	return &servicequeryarrayuint32validate.MethodQueryArrayUInt32ValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryArrayUInt64ConstructorCode = `// NewMethodQueryArrayUInt64Payload builds a ServiceQueryArrayUInt64 service
// MethodQueryArrayUInt64 endpoint payload.
func NewMethodQueryArrayUInt64Payload(q []uint64) *servicequeryarrayuint64.MethodQueryArrayUInt64Payload {
	return &servicequeryarrayuint64.MethodQueryArrayUInt64Payload{
		Q: q,
	}
}
`

var PayloadQueryArrayUInt64ValidateConstructorCode = `// NewMethodQueryArrayUInt64ValidatePayload builds a
// ServiceQueryArrayUInt64Validate service MethodQueryArrayUInt64Validate
// endpoint payload.
func NewMethodQueryArrayUInt64ValidatePayload(q []uint64) *servicequeryarrayuint64validate.MethodQueryArrayUInt64ValidatePayload {
	return &servicequeryarrayuint64validate.MethodQueryArrayUInt64ValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryArrayFloat32ConstructorCode = `// NewMethodQueryArrayFloat32Payload builds a ServiceQueryArrayFloat32 service
// MethodQueryArrayFloat32 endpoint payload.
func NewMethodQueryArrayFloat32Payload(q []float32) *servicequeryarrayfloat32.MethodQueryArrayFloat32Payload {
	return &servicequeryarrayfloat32.MethodQueryArrayFloat32Payload{
		Q: q,
	}
}
`

var PayloadQueryArrayFloat32ValidateConstructorCode = `// NewMethodQueryArrayFloat32ValidatePayload builds a
// ServiceQueryArrayFloat32Validate service MethodQueryArrayFloat32Validate
// endpoint payload.
func NewMethodQueryArrayFloat32ValidatePayload(q []float32) *servicequeryarrayfloat32validate.MethodQueryArrayFloat32ValidatePayload {
	return &servicequeryarrayfloat32validate.MethodQueryArrayFloat32ValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryArrayFloat64ConstructorCode = `// NewMethodQueryArrayFloat64Payload builds a ServiceQueryArrayFloat64 service
// MethodQueryArrayFloat64 endpoint payload.
func NewMethodQueryArrayFloat64Payload(q []float64) *servicequeryarrayfloat64.MethodQueryArrayFloat64Payload {
	return &servicequeryarrayfloat64.MethodQueryArrayFloat64Payload{
		Q: q,
	}
}
`

var PayloadQueryArrayFloat64ValidateConstructorCode = `// NewMethodQueryArrayFloat64ValidatePayload builds a
// ServiceQueryArrayFloat64Validate service MethodQueryArrayFloat64Validate
// endpoint payload.
func NewMethodQueryArrayFloat64ValidatePayload(q []float64) *servicequeryarrayfloat64validate.MethodQueryArrayFloat64ValidatePayload {
	return &servicequeryarrayfloat64validate.MethodQueryArrayFloat64ValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryArrayStringConstructorCode = `// NewMethodQueryArrayStringPayload builds a ServiceQueryArrayString service
// MethodQueryArrayString endpoint payload.
func NewMethodQueryArrayStringPayload(q []string) *servicequeryarraystring.MethodQueryArrayStringPayload {
	return &servicequeryarraystring.MethodQueryArrayStringPayload{
		Q: q,
	}
}
`

var PayloadQueryArrayStringValidateConstructorCode = `// NewMethodQueryArrayStringValidatePayload builds a
// ServiceQueryArrayStringValidate service MethodQueryArrayStringValidate
// endpoint payload.
func NewMethodQueryArrayStringValidatePayload(q []string) *servicequeryarraystringvalidate.MethodQueryArrayStringValidatePayload {
	return &servicequeryarraystringvalidate.MethodQueryArrayStringValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryArrayBytesConstructorCode = `// NewMethodQueryArrayBytesPayload builds a ServiceQueryArrayBytes service
// MethodQueryArrayBytes endpoint payload.
func NewMethodQueryArrayBytesPayload(q [][]byte) *servicequeryarraybytes.MethodQueryArrayBytesPayload {
	return &servicequeryarraybytes.MethodQueryArrayBytesPayload{
		Q: q,
	}
}
`

var PayloadQueryArrayBytesValidateConstructorCode = `// NewMethodQueryArrayBytesValidatePayload builds a
// ServiceQueryArrayBytesValidate service MethodQueryArrayBytesValidate
// endpoint payload.
func NewMethodQueryArrayBytesValidatePayload(q [][]byte) *servicequeryarraybytesvalidate.MethodQueryArrayBytesValidatePayload {
	return &servicequeryarraybytesvalidate.MethodQueryArrayBytesValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryArrayAnyConstructorCode = `// NewMethodQueryArrayAnyPayload builds a ServiceQueryArrayAny service
// MethodQueryArrayAny endpoint payload.
func NewMethodQueryArrayAnyPayload(q []interface{}) *servicequeryarrayany.MethodQueryArrayAnyPayload {
	return &servicequeryarrayany.MethodQueryArrayAnyPayload{
		Q: q,
	}
}
`

var PayloadQueryArrayAnyValidateConstructorCode = `// NewMethodQueryArrayAnyValidatePayload builds a ServiceQueryArrayAnyValidate
// service MethodQueryArrayAnyValidate endpoint payload.
func NewMethodQueryArrayAnyValidatePayload(q []interface{}) *servicequeryarrayanyvalidate.MethodQueryArrayAnyValidatePayload {
	return &servicequeryarrayanyvalidate.MethodQueryArrayAnyValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryMapStringStringConstructorCode = `// NewMethodQueryMapStringStringPayload builds a ServiceQueryMapStringString
// service MethodQueryMapStringString endpoint payload.
func NewMethodQueryMapStringStringPayload(q map[string]string) *servicequerymapstringstring.MethodQueryMapStringStringPayload {
	return &servicequerymapstringstring.MethodQueryMapStringStringPayload{
		Q: q,
	}
}
`

var PayloadQueryMapStringStringValidateConstructorCode = `// NewMethodQueryMapStringStringValidatePayload builds a
// ServiceQueryMapStringStringValidate service
// MethodQueryMapStringStringValidate endpoint payload.
func NewMethodQueryMapStringStringValidatePayload(q map[string]string) *servicequerymapstringstringvalidate.MethodQueryMapStringStringValidatePayload {
	return &servicequerymapstringstringvalidate.MethodQueryMapStringStringValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryMapStringBoolConstructorCode = `// NewMethodQueryMapStringBoolPayload builds a ServiceQueryMapStringBool
// service MethodQueryMapStringBool endpoint payload.
func NewMethodQueryMapStringBoolPayload(q map[string]bool) *servicequerymapstringbool.MethodQueryMapStringBoolPayload {
	return &servicequerymapstringbool.MethodQueryMapStringBoolPayload{
		Q: q,
	}
}
`

var PayloadQueryMapStringBoolValidateConstructorCode = `// NewMethodQueryMapStringBoolValidatePayload builds a
// ServiceQueryMapStringBoolValidate service MethodQueryMapStringBoolValidate
// endpoint payload.
func NewMethodQueryMapStringBoolValidatePayload(q map[string]bool) *servicequerymapstringboolvalidate.MethodQueryMapStringBoolValidatePayload {
	return &servicequerymapstringboolvalidate.MethodQueryMapStringBoolValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryMapBoolStringConstructorCode = `// NewMethodQueryMapBoolStringPayload builds a ServiceQueryMapBoolString
// service MethodQueryMapBoolString endpoint payload.
func NewMethodQueryMapBoolStringPayload(q map[bool]string) *servicequerymapboolstring.MethodQueryMapBoolStringPayload {
	return &servicequerymapboolstring.MethodQueryMapBoolStringPayload{
		Q: q,
	}
}
`

var PayloadQueryMapBoolStringValidateConstructorCode = `// NewMethodQueryMapBoolStringValidatePayload builds a
// ServiceQueryMapBoolStringValidate service MethodQueryMapBoolStringValidate
// endpoint payload.
func NewMethodQueryMapBoolStringValidatePayload(q map[bool]string) *servicequerymapboolstringvalidate.MethodQueryMapBoolStringValidatePayload {
	return &servicequerymapboolstringvalidate.MethodQueryMapBoolStringValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryMapBoolBoolConstructorCode = `// NewMethodQueryMapBoolBoolPayload builds a ServiceQueryMapBoolBool service
// MethodQueryMapBoolBool endpoint payload.
func NewMethodQueryMapBoolBoolPayload(q map[bool]bool) *servicequerymapboolbool.MethodQueryMapBoolBoolPayload {
	return &servicequerymapboolbool.MethodQueryMapBoolBoolPayload{
		Q: q,
	}
}
`

var PayloadQueryMapBoolBoolValidateConstructorCode = `// NewMethodQueryMapBoolBoolValidatePayload builds a
// ServiceQueryMapBoolBoolValidate service MethodQueryMapBoolBoolValidate
// endpoint payload.
func NewMethodQueryMapBoolBoolValidatePayload(q map[bool]bool) *servicequerymapboolboolvalidate.MethodQueryMapBoolBoolValidatePayload {
	return &servicequerymapboolboolvalidate.MethodQueryMapBoolBoolValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryMapStringArrayStringConstructorCode = `// NewMethodQueryMapStringArrayStringPayload builds a
// ServiceQueryMapStringArrayString service MethodQueryMapStringArrayString
// endpoint payload.
func NewMethodQueryMapStringArrayStringPayload(q map[string][]string) *servicequerymapstringarraystring.MethodQueryMapStringArrayStringPayload {
	return &servicequerymapstringarraystring.MethodQueryMapStringArrayStringPayload{
		Q: q,
	}
}
`

var PayloadQueryMapStringArrayStringValidateConstructorCode = `// NewMethodQueryMapStringArrayStringValidatePayload builds a
// ServiceQueryMapStringArrayStringValidate service
// MethodQueryMapStringArrayStringValidate endpoint payload.
func NewMethodQueryMapStringArrayStringValidatePayload(q map[string][]string) *servicequerymapstringarraystringvalidate.MethodQueryMapStringArrayStringValidatePayload {
	return &servicequerymapstringarraystringvalidate.MethodQueryMapStringArrayStringValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryMapStringArrayBoolConstructorCode = `// NewMethodQueryMapStringArrayBoolPayload builds a
// ServiceQueryMapStringArrayBool service MethodQueryMapStringArrayBool
// endpoint payload.
func NewMethodQueryMapStringArrayBoolPayload(q map[string][]bool) *servicequerymapstringarraybool.MethodQueryMapStringArrayBoolPayload {
	return &servicequerymapstringarraybool.MethodQueryMapStringArrayBoolPayload{
		Q: q,
	}
}
`

var PayloadQueryMapStringArrayBoolValidateConstructorCode = `// NewMethodQueryMapStringArrayBoolValidatePayload builds a
// ServiceQueryMapStringArrayBoolValidate service
// MethodQueryMapStringArrayBoolValidate endpoint payload.
func NewMethodQueryMapStringArrayBoolValidatePayload(q map[string][]bool) *servicequerymapstringarrayboolvalidate.MethodQueryMapStringArrayBoolValidatePayload {
	return &servicequerymapstringarrayboolvalidate.MethodQueryMapStringArrayBoolValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryMapBoolArrayStringConstructorCode = `// NewMethodQueryMapBoolArrayStringPayload builds a
// ServiceQueryMapBoolArrayString service MethodQueryMapBoolArrayString
// endpoint payload.
func NewMethodQueryMapBoolArrayStringPayload(q map[bool][]string) *servicequerymapboolarraystring.MethodQueryMapBoolArrayStringPayload {
	return &servicequerymapboolarraystring.MethodQueryMapBoolArrayStringPayload{
		Q: q,
	}
}
`

var PayloadQueryMapBoolArrayStringValidateConstructorCode = `// NewMethodQueryMapBoolArrayStringValidatePayload builds a
// ServiceQueryMapBoolArrayStringValidate service
// MethodQueryMapBoolArrayStringValidate endpoint payload.
func NewMethodQueryMapBoolArrayStringValidatePayload(q map[bool][]string) *servicequerymapboolarraystringvalidate.MethodQueryMapBoolArrayStringValidatePayload {
	return &servicequerymapboolarraystringvalidate.MethodQueryMapBoolArrayStringValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryMapBoolArrayBoolConstructorCode = `// NewMethodQueryMapBoolArrayBoolPayload builds a ServiceQueryMapBoolArrayBool
// service MethodQueryMapBoolArrayBool endpoint payload.
func NewMethodQueryMapBoolArrayBoolPayload(q map[bool][]bool) *servicequerymapboolarraybool.MethodQueryMapBoolArrayBoolPayload {
	return &servicequerymapboolarraybool.MethodQueryMapBoolArrayBoolPayload{
		Q: q,
	}
}
`

var PayloadQueryMapBoolArrayBoolValidateConstructorCode = `// NewMethodQueryMapBoolArrayBoolValidatePayload builds a
// ServiceQueryMapBoolArrayBoolValidate service
// MethodQueryMapBoolArrayBoolValidate endpoint payload.
func NewMethodQueryMapBoolArrayBoolValidatePayload(q map[bool][]bool) *servicequerymapboolarrayboolvalidate.MethodQueryMapBoolArrayBoolValidatePayload {
	return &servicequerymapboolarrayboolvalidate.MethodQueryMapBoolArrayBoolValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryStringMappedConstructorCode = `// NewMethodQueryStringMappedPayload builds a ServiceQueryStringMapped service
// MethodQueryStringMapped endpoint payload.
func NewMethodQueryStringMappedPayload(query *string) *servicequerystringmapped.MethodQueryStringMappedPayload {
	return &servicequerystringmapped.MethodQueryStringMappedPayload{
		Query: query,
	}
}
`

var PayloadPathStringConstructorCode = `// NewMethodPathStringPayload builds a ServicePathString service
// MethodPathString endpoint payload.
func NewMethodPathStringPayload(p string) *servicepathstring.MethodPathStringPayload {
	return &servicepathstring.MethodPathStringPayload{
		P: &p,
	}
}
`

var PayloadPathStringValidateConstructorCode = `// NewMethodPathStringValidatePayload builds a ServicePathStringValidate
// service MethodPathStringValidate endpoint payload.
func NewMethodPathStringValidatePayload(p string) *servicepathstringvalidate.MethodPathStringValidatePayload {
	return &servicepathstringvalidate.MethodPathStringValidatePayload{
		P: p,
	}
}
`

var PayloadPathArrayStringConstructorCode = `// NewMethodPathArrayStringPayload builds a ServicePathArrayString service
// MethodPathArrayString endpoint payload.
func NewMethodPathArrayStringPayload(p []string) *servicepatharraystring.MethodPathArrayStringPayload {
	return &servicepatharraystring.MethodPathArrayStringPayload{
		P: p,
	}
}
`

var PayloadPathArrayStringValidateConstructorCode = `// NewMethodPathArrayStringValidatePayload builds a
// ServicePathArrayStringValidate service MethodPathArrayStringValidate
// endpoint payload.
func NewMethodPathArrayStringValidatePayload(p []string) *servicepatharraystringvalidate.MethodPathArrayStringValidatePayload {
	return &servicepatharraystringvalidate.MethodPathArrayStringValidatePayload{
		P: p,
	}
}
`

var PayloadHeaderStringConstructorCode = `// NewMethodHeaderStringPayload builds a ServiceHeaderString service
// MethodHeaderString endpoint payload.
func NewMethodHeaderStringPayload(h *string) *serviceheaderstring.MethodHeaderStringPayload {
	return &serviceheaderstring.MethodHeaderStringPayload{
		H: h,
	}
}
`

var PayloadHeaderStringValidateConstructorCode = `// NewMethodHeaderStringValidatePayload builds a ServiceHeaderStringValidate
// service MethodHeaderStringValidate endpoint payload.
func NewMethodHeaderStringValidatePayload(h *string) *serviceheaderstringvalidate.MethodHeaderStringValidatePayload {
	return &serviceheaderstringvalidate.MethodHeaderStringValidatePayload{
		H: h,
	}
}
`

var PayloadHeaderArrayStringConstructorCode = `// NewMethodHeaderArrayStringPayload builds a ServiceHeaderArrayString service
// MethodHeaderArrayString endpoint payload.
func NewMethodHeaderArrayStringPayload(h []string) *serviceheaderarraystring.MethodHeaderArrayStringPayload {
	return &serviceheaderarraystring.MethodHeaderArrayStringPayload{
		H: h,
	}
}
`

var PayloadHeaderArrayStringValidateConstructorCode = `// NewMethodHeaderArrayStringValidatePayload builds a
// ServiceHeaderArrayStringValidate service MethodHeaderArrayStringValidate
// endpoint payload.
func NewMethodHeaderArrayStringValidatePayload(h []string) *serviceheaderarraystringvalidate.MethodHeaderArrayStringValidatePayload {
	return &serviceheaderarraystringvalidate.MethodHeaderArrayStringValidatePayload{
		H: h,
	}
}
`

var PayloadBodyQueryObjectConstructorCode = `// NewMethodBodyQueryObjectPayload builds a ServiceBodyQueryObject service
// MethodBodyQueryObject endpoint payload.
func NewMethodBodyQueryObjectPayload(body *MethodBodyQueryObjectRequestBody, b *string) *servicebodyqueryobject.MethodBodyQueryObjectPayload {
	v := &servicebodyqueryobject.MethodBodyQueryObjectPayload{
		A: body.A,
	}
	v.B = b
	return v
}
`

var PayloadBodyQueryObjectValidateConstructorCode = `// NewMethodBodyQueryObjectValidatePayload builds a
// ServiceBodyQueryObjectValidate service MethodBodyQueryObjectValidate
// endpoint payload.
func NewMethodBodyQueryObjectValidatePayload(body *MethodBodyQueryObjectValidateRequestBody, b string) *servicebodyqueryobjectvalidate.MethodBodyQueryObjectValidatePayload {
	v := &servicebodyqueryobjectvalidate.MethodBodyQueryObjectValidatePayload{
		A: *body.A,
	}
	v.B = b
	return v
}
`

var PayloadBodyQueryUserConstructorCode = `// NewMethodBodyQueryUserPayloadType builds a ServiceBodyQueryUser service
// MethodBodyQueryUser endpoint payload.
func NewMethodBodyQueryUserPayloadType(body *MethodBodyQueryUserRequestBody, b *string) *servicebodyqueryuser.PayloadType {
	v := &servicebodyqueryuser.PayloadType{
		A: body.A,
	}
	v.B = b
	return v
}
`

var PayloadBodyQueryUserValidateConstructorCode = `// NewMethodBodyQueryUserValidatePayloadType builds a
// ServiceBodyQueryUserValidate service MethodBodyQueryUserValidate endpoint
// payload.
func NewMethodBodyQueryUserValidatePayloadType(body *MethodBodyQueryUserValidateRequestBody, b string) *servicebodyqueryuservalidate.PayloadType {
	v := &servicebodyqueryuservalidate.PayloadType{
		A: *body.A,
	}
	v.B = b
	return v
}
`

var PayloadBodyPathObjectConstructorCode = `// NewMethodBodyPathObjectPayload builds a ServiceBodyPathObject service
// MethodBodyPathObject endpoint payload.
func NewMethodBodyPathObjectPayload(body *MethodBodyPathObjectRequestBody, b string) *servicebodypathobject.MethodBodyPathObjectPayload {
	v := &servicebodypathobject.MethodBodyPathObjectPayload{
		A: body.A,
	}
	v.B = &b
	return v
}
`

var PayloadBodyPathObjectValidateConstructorCode = `// NewMethodBodyPathObjectValidatePayload builds a
// ServiceBodyPathObjectValidate service MethodBodyPathObjectValidate endpoint
// payload.
func NewMethodBodyPathObjectValidatePayload(body *MethodBodyPathObjectValidateRequestBody, b string) *servicebodypathobjectvalidate.MethodBodyPathObjectValidatePayload {
	v := &servicebodypathobjectvalidate.MethodBodyPathObjectValidatePayload{
		A: *body.A,
	}
	v.B = b
	return v
}
`

var PayloadBodyPathUserConstructorCode = `// NewMethodBodyPathUserPayloadType builds a ServiceBodyPathUser service
// MethodBodyPathUser endpoint payload.
func NewMethodBodyPathUserPayloadType(body *MethodBodyPathUserRequestBody, b string) *servicebodypathuser.PayloadType {
	v := &servicebodypathuser.PayloadType{
		A: body.A,
	}
	v.B = &b
	return v
}
`

var PayloadBodyPathUserValidateConstructorCode = `// NewMethodUserBodyPathValidatePayloadType builds a
// ServiceBodyPathUserValidate service MethodUserBodyPathValidate endpoint
// payload.
func NewMethodUserBodyPathValidatePayloadType(body *MethodUserBodyPathValidateRequestBody, b string) *servicebodypathuservalidate.PayloadType {
	v := &servicebodypathuservalidate.PayloadType{
		A: *body.A,
	}
	v.B = b
	return v
}
`

var PayloadBodyQueryPathObjectConstructorCode = `// NewMethodBodyQueryPathObjectPayload builds a ServiceBodyQueryPathObject
// service MethodBodyQueryPathObject endpoint payload.
func NewMethodBodyQueryPathObjectPayload(body *MethodBodyQueryPathObjectRequestBody, c string, b *string) *servicebodyquerypathobject.MethodBodyQueryPathObjectPayload {
	v := &servicebodyquerypathobject.MethodBodyQueryPathObjectPayload{
		A: body.A,
	}
	v.C = &c
	v.B = b
	return v
}
`

var PayloadBodyQueryPathObjectValidateConstructorCode = `// NewMethodBodyQueryPathObjectValidatePayload builds a
// ServiceBodyQueryPathObjectValidate service MethodBodyQueryPathObjectValidate
// endpoint payload.
func NewMethodBodyQueryPathObjectValidatePayload(body *MethodBodyQueryPathObjectValidateRequestBody, c string, b string) *servicebodyquerypathobjectvalidate.MethodBodyQueryPathObjectValidatePayload {
	v := &servicebodyquerypathobjectvalidate.MethodBodyQueryPathObjectValidatePayload{
		A: *body.A,
	}
	v.C = c
	v.B = b
	return v
}
`

var PayloadBodyQueryPathUserConstructorCode = `// NewMethodBodyQueryPathUserPayloadType builds a ServiceBodyQueryPathUser
// service MethodBodyQueryPathUser endpoint payload.
func NewMethodBodyQueryPathUserPayloadType(body *MethodBodyQueryPathUserRequestBody, c string, b *string) *servicebodyquerypathuser.PayloadType {
	v := &servicebodyquerypathuser.PayloadType{
		A: body.A,
	}
	v.C = &c
	v.B = b
	return v
}
`

var PayloadBodyQueryPathUserValidateConstructorCode = `// NewMethodBodyQueryPathUserValidatePayloadType builds a
// ServiceBodyQueryPathUserValidate service MethodBodyQueryPathUserValidate
// endpoint payload.
func NewMethodBodyQueryPathUserValidatePayloadType(body *MethodBodyQueryPathUserValidateRequestBody, c string, b string) *servicebodyquerypathuservalidate.PayloadType {
	v := &servicebodyquerypathuservalidate.PayloadType{
		A: *body.A,
	}
	v.C = c
	v.B = b
	return v
}
`

var PayloadBodyUserInnerConstructorCode = `// NewMethodBodyUserInnerPayloadType builds a ServiceBodyUserInner service
// MethodBodyUserInner endpoint payload.
func NewMethodBodyUserInnerPayloadType(body *MethodBodyUserInnerRequestBody) *servicebodyuserinner.PayloadType {
	v := &servicebodyuserinner.PayloadType{}
	if body.Inner != nil {
		v.Inner = unmarshalInnerTypeRequestBodyToServicebodyuserinnerInnerType(body.Inner)
	}
	return v
}
`

var PayloadBodyUserInnerDefaultConstructorCode = `// NewMethodBodyUserInnerDefaultPayloadType builds a
// ServiceBodyUserInnerDefault service MethodBodyUserInnerDefault endpoint
// payload.
func NewMethodBodyUserInnerDefaultPayloadType(body *MethodBodyUserInnerDefaultRequestBody) *servicebodyuserinnerdefault.PayloadType {
	v := &servicebodyuserinnerdefault.PayloadType{}
	if body.Inner != nil {
		v.Inner = unmarshalInnerTypeRequestBodyToServicebodyuserinnerdefaultInnerType(body.Inner)
	}
	return v
}
`

var PayloadBodyInlineArrayUserConstructorCode = `// NewMethodBodyInlineArrayUserElemType builds a ServiceBodyInlineArrayUser
// service MethodBodyInlineArrayUser endpoint payload.
func NewMethodBodyInlineArrayUserElemType(body []*ElemTypeRequestBody) []*servicebodyinlinearrayuser.ElemType {
	v := make([]*servicebodyinlinearrayuser.ElemType, len(body))
	for i, val := range body {
		v[i] = &servicebodyinlinearrayuser.ElemType{
			A: *val.A,
			B: val.B,
		}
	}
	return v
}
`

var PayloadBodyInlineMapUserConstructorCode = `// NewMethodBodyInlineMapUserMapKeyTypeElemType builds a
// ServiceBodyInlineMapUser service MethodBodyInlineMapUser endpoint payload.
func NewMethodBodyInlineMapUserMapKeyTypeElemType(body map[*KeyTypeRequestBody]*ElemTypeRequestBody) map[*servicebodyinlinemapuser.KeyType]*servicebodyinlinemapuser.ElemType {
	v := make(map[*servicebodyinlinemapuser.KeyType]*servicebodyinlinemapuser.ElemType, len(body))
	for key, val := range body {
		tk := &servicebodyinlinemapuser.KeyType{
			A: *key.A,
			B: key.B,
		}
		tv := &servicebodyinlinemapuser.ElemType{
			A: *val.A,
			B: val.B,
		}
		v[tk] = tv
	}
	return v
}
`

var PayloadBodyInlineRecursiveUserConstructorCode = `// NewMethodBodyInlineRecursiveUserPayloadType builds a
// ServiceBodyInlineRecursiveUser service MethodBodyInlineRecursiveUser
// endpoint payload.
func NewMethodBodyInlineRecursiveUserPayloadType(body *MethodBodyInlineRecursiveUserRequestBody, a string, b *string) *servicebodyinlinerecursiveuser.PayloadType {
	v := &servicebodyinlinerecursiveuser.PayloadType{}
	if body.C != nil {
		v.C = unmarshalPayloadTypeRequestBodyToServicebodyinlinerecursiveuserPayloadType(body.C)
	}
	v.A = a
	v.B = b
	return v
}
`
