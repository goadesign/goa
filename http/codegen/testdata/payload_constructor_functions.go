package testdata

var PayloadQueryBoolConstructorCode = `// NewMethodQueryBoolMethodQueryBoolPayload builds a ServiceQueryBool service
// MethodQueryBool endpoint payload.
func NewMethodQueryBoolMethodQueryBoolPayload(q *bool) *servicequerybool.MethodQueryBoolPayload {
	return &servicequerybool.MethodQueryBoolPayload{
		Q: q,
	}
}
`

var PayloadQueryBoolValidateConstructorCode = `// NewMethodQueryBoolValidateMethodQueryBoolValidatePayload builds a
// ServiceQueryBoolValidate service MethodQueryBoolValidate endpoint payload.
func NewMethodQueryBoolValidateMethodQueryBoolValidatePayload(q bool) *servicequeryboolvalidate.MethodQueryBoolValidatePayload {
	return &servicequeryboolvalidate.MethodQueryBoolValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryIntConstructorCode = `// NewMethodQueryIntMethodQueryIntPayload builds a ServiceQueryInt service
// MethodQueryInt endpoint payload.
func NewMethodQueryIntMethodQueryIntPayload(q *int) *servicequeryint.MethodQueryIntPayload {
	return &servicequeryint.MethodQueryIntPayload{
		Q: q,
	}
}
`

var PayloadQueryIntValidateConstructorCode = `// NewMethodQueryIntValidateMethodQueryIntValidatePayload builds a
// ServiceQueryIntValidate service MethodQueryIntValidate endpoint payload.
func NewMethodQueryIntValidateMethodQueryIntValidatePayload(q int) *servicequeryintvalidate.MethodQueryIntValidatePayload {
	return &servicequeryintvalidate.MethodQueryIntValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryInt32ConstructorCode = `// NewMethodQueryInt32MethodQueryInt32Payload builds a ServiceQueryInt32
// service MethodQueryInt32 endpoint payload.
func NewMethodQueryInt32MethodQueryInt32Payload(q *int32) *servicequeryint32.MethodQueryInt32Payload {
	return &servicequeryint32.MethodQueryInt32Payload{
		Q: q,
	}
}
`

var PayloadQueryInt32ValidateConstructorCode = `// NewMethodQueryInt32ValidateMethodQueryInt32ValidatePayload builds a
// ServiceQueryInt32Validate service MethodQueryInt32Validate endpoint payload.
func NewMethodQueryInt32ValidateMethodQueryInt32ValidatePayload(q int32) *servicequeryint32validate.MethodQueryInt32ValidatePayload {
	return &servicequeryint32validate.MethodQueryInt32ValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryInt64ConstructorCode = `// NewMethodQueryInt64MethodQueryInt64Payload builds a ServiceQueryInt64
// service MethodQueryInt64 endpoint payload.
func NewMethodQueryInt64MethodQueryInt64Payload(q *int64) *servicequeryint64.MethodQueryInt64Payload {
	return &servicequeryint64.MethodQueryInt64Payload{
		Q: q,
	}
}
`

var PayloadQueryInt64ValidateConstructorCode = `// NewMethodQueryInt64ValidateMethodQueryInt64ValidatePayload builds a
// ServiceQueryInt64Validate service MethodQueryInt64Validate endpoint payload.
func NewMethodQueryInt64ValidateMethodQueryInt64ValidatePayload(q int64) *servicequeryint64validate.MethodQueryInt64ValidatePayload {
	return &servicequeryint64validate.MethodQueryInt64ValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryUIntConstructorCode = `// NewMethodQueryUIntMethodQueryUIntPayload builds a ServiceQueryUInt service
// MethodQueryUInt endpoint payload.
func NewMethodQueryUIntMethodQueryUIntPayload(q *uint) *servicequeryuint.MethodQueryUIntPayload {
	return &servicequeryuint.MethodQueryUIntPayload{
		Q: q,
	}
}
`

var PayloadQueryUIntValidateConstructorCode = `// NewMethodQueryUIntValidateMethodQueryUIntValidatePayload builds a
// ServiceQueryUIntValidate service MethodQueryUIntValidate endpoint payload.
func NewMethodQueryUIntValidateMethodQueryUIntValidatePayload(q uint) *servicequeryuintvalidate.MethodQueryUIntValidatePayload {
	return &servicequeryuintvalidate.MethodQueryUIntValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryUInt32ConstructorCode = `// NewMethodQueryUInt32MethodQueryUInt32Payload builds a ServiceQueryUInt32
// service MethodQueryUInt32 endpoint payload.
func NewMethodQueryUInt32MethodQueryUInt32Payload(q *uint32) *servicequeryuint32.MethodQueryUInt32Payload {
	return &servicequeryuint32.MethodQueryUInt32Payload{
		Q: q,
	}
}
`

var PayloadQueryUInt32ValidateConstructorCode = `// NewMethodQueryUInt32ValidateMethodQueryUInt32ValidatePayload builds a
// ServiceQueryUInt32Validate service MethodQueryUInt32Validate endpoint
// payload.
func NewMethodQueryUInt32ValidateMethodQueryUInt32ValidatePayload(q uint32) *servicequeryuint32validate.MethodQueryUInt32ValidatePayload {
	return &servicequeryuint32validate.MethodQueryUInt32ValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryUInt64ConstructorCode = `// NewMethodQueryUInt64MethodQueryUInt64Payload builds a ServiceQueryUInt64
// service MethodQueryUInt64 endpoint payload.
func NewMethodQueryUInt64MethodQueryUInt64Payload(q *uint64) *servicequeryuint64.MethodQueryUInt64Payload {
	return &servicequeryuint64.MethodQueryUInt64Payload{
		Q: q,
	}
}
`

var PayloadQueryUInt64ValidateConstructorCode = `// NewMethodQueryUInt64ValidateMethodQueryUInt64ValidatePayload builds a
// ServiceQueryUInt64Validate service MethodQueryUInt64Validate endpoint
// payload.
func NewMethodQueryUInt64ValidateMethodQueryUInt64ValidatePayload(q uint64) *servicequeryuint64validate.MethodQueryUInt64ValidatePayload {
	return &servicequeryuint64validate.MethodQueryUInt64ValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryFloat32ConstructorCode = `// NewMethodQueryFloat32MethodQueryFloat32Payload builds a ServiceQueryFloat32
// service MethodQueryFloat32 endpoint payload.
func NewMethodQueryFloat32MethodQueryFloat32Payload(q *float32) *servicequeryfloat32.MethodQueryFloat32Payload {
	return &servicequeryfloat32.MethodQueryFloat32Payload{
		Q: q,
	}
}
`

var PayloadQueryFloat32ValidateConstructorCode = `// NewMethodQueryFloat32ValidateMethodQueryFloat32ValidatePayload builds a
// ServiceQueryFloat32Validate service MethodQueryFloat32Validate endpoint
// payload.
func NewMethodQueryFloat32ValidateMethodQueryFloat32ValidatePayload(q float32) *servicequeryfloat32validate.MethodQueryFloat32ValidatePayload {
	return &servicequeryfloat32validate.MethodQueryFloat32ValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryFloat64ConstructorCode = `// NewMethodQueryFloat64MethodQueryFloat64Payload builds a ServiceQueryFloat64
// service MethodQueryFloat64 endpoint payload.
func NewMethodQueryFloat64MethodQueryFloat64Payload(q *float64) *servicequeryfloat64.MethodQueryFloat64Payload {
	return &servicequeryfloat64.MethodQueryFloat64Payload{
		Q: q,
	}
}
`

var PayloadQueryFloat64ValidateConstructorCode = `// NewMethodQueryFloat64ValidateMethodQueryFloat64ValidatePayload builds a
// ServiceQueryFloat64Validate service MethodQueryFloat64Validate endpoint
// payload.
func NewMethodQueryFloat64ValidateMethodQueryFloat64ValidatePayload(q float64) *servicequeryfloat64validate.MethodQueryFloat64ValidatePayload {
	return &servicequeryfloat64validate.MethodQueryFloat64ValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryStringConstructorCode = `// NewMethodQueryStringMethodQueryStringPayload builds a ServiceQueryString
// service MethodQueryString endpoint payload.
func NewMethodQueryStringMethodQueryStringPayload(q *string) *servicequerystring.MethodQueryStringPayload {
	return &servicequerystring.MethodQueryStringPayload{
		Q: q,
	}
}
`

var PayloadQueryStringValidateConstructorCode = `// NewMethodQueryStringValidateMethodQueryStringValidatePayload builds a
// ServiceQueryStringValidate service MethodQueryStringValidate endpoint
// payload.
func NewMethodQueryStringValidateMethodQueryStringValidatePayload(q string) *servicequerystringvalidate.MethodQueryStringValidatePayload {
	return &servicequerystringvalidate.MethodQueryStringValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryBytesConstructorCode = `// NewMethodQueryBytesMethodQueryBytesPayload builds a ServiceQueryBytes
// service MethodQueryBytes endpoint payload.
func NewMethodQueryBytesMethodQueryBytesPayload(q []byte) *servicequerybytes.MethodQueryBytesPayload {
	return &servicequerybytes.MethodQueryBytesPayload{
		Q: q,
	}
}
`

var PayloadQueryBytesValidateConstructorCode = `// NewMethodQueryBytesValidateMethodQueryBytesValidatePayload builds a
// ServiceQueryBytesValidate service MethodQueryBytesValidate endpoint payload.
func NewMethodQueryBytesValidateMethodQueryBytesValidatePayload(q []byte) *servicequerybytesvalidate.MethodQueryBytesValidatePayload {
	return &servicequerybytesvalidate.MethodQueryBytesValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryAnyConstructorCode = `// NewMethodQueryAnyMethodQueryAnyPayload builds a ServiceQueryAny service
// MethodQueryAny endpoint payload.
func NewMethodQueryAnyMethodQueryAnyPayload(q interface{}) *servicequeryany.MethodQueryAnyPayload {
	return &servicequeryany.MethodQueryAnyPayload{
		Q: q,
	}
}
`

var PayloadQueryAnyValidateConstructorCode = `// NewMethodQueryAnyValidateMethodQueryAnyValidatePayload builds a
// ServiceQueryAnyValidate service MethodQueryAnyValidate endpoint payload.
func NewMethodQueryAnyValidateMethodQueryAnyValidatePayload(q interface{}) *servicequeryanyvalidate.MethodQueryAnyValidatePayload {
	return &servicequeryanyvalidate.MethodQueryAnyValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryArrayBoolConstructorCode = `// NewMethodQueryArrayBoolMethodQueryArrayBoolPayload builds a
// ServiceQueryArrayBool service MethodQueryArrayBool endpoint payload.
func NewMethodQueryArrayBoolMethodQueryArrayBoolPayload(q []bool) *servicequeryarraybool.MethodQueryArrayBoolPayload {
	return &servicequeryarraybool.MethodQueryArrayBoolPayload{
		Q: q,
	}
}
`

var PayloadQueryArrayBoolValidateConstructorCode = `// NewMethodQueryArrayBoolValidateMethodQueryArrayBoolValidatePayload builds a
// ServiceQueryArrayBoolValidate service MethodQueryArrayBoolValidate endpoint
// payload.
func NewMethodQueryArrayBoolValidateMethodQueryArrayBoolValidatePayload(q []bool) *servicequeryarrayboolvalidate.MethodQueryArrayBoolValidatePayload {
	return &servicequeryarrayboolvalidate.MethodQueryArrayBoolValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryArrayIntConstructorCode = `// NewMethodQueryArrayIntMethodQueryArrayIntPayload builds a
// ServiceQueryArrayInt service MethodQueryArrayInt endpoint payload.
func NewMethodQueryArrayIntMethodQueryArrayIntPayload(q []int) *servicequeryarrayint.MethodQueryArrayIntPayload {
	return &servicequeryarrayint.MethodQueryArrayIntPayload{
		Q: q,
	}
}
`

var PayloadQueryArrayIntValidateConstructorCode = `// NewMethodQueryArrayIntValidateMethodQueryArrayIntValidatePayload builds a
// ServiceQueryArrayIntValidate service MethodQueryArrayIntValidate endpoint
// payload.
func NewMethodQueryArrayIntValidateMethodQueryArrayIntValidatePayload(q []int) *servicequeryarrayintvalidate.MethodQueryArrayIntValidatePayload {
	return &servicequeryarrayintvalidate.MethodQueryArrayIntValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryArrayInt32ConstructorCode = `// NewMethodQueryArrayInt32MethodQueryArrayInt32Payload builds a
// ServiceQueryArrayInt32 service MethodQueryArrayInt32 endpoint payload.
func NewMethodQueryArrayInt32MethodQueryArrayInt32Payload(q []int32) *servicequeryarrayint32.MethodQueryArrayInt32Payload {
	return &servicequeryarrayint32.MethodQueryArrayInt32Payload{
		Q: q,
	}
}
`

var PayloadQueryArrayInt32ValidateConstructorCode = `// NewMethodQueryArrayInt32ValidateMethodQueryArrayInt32ValidatePayload builds
// a ServiceQueryArrayInt32Validate service MethodQueryArrayInt32Validate
// endpoint payload.
func NewMethodQueryArrayInt32ValidateMethodQueryArrayInt32ValidatePayload(q []int32) *servicequeryarrayint32validate.MethodQueryArrayInt32ValidatePayload {
	return &servicequeryarrayint32validate.MethodQueryArrayInt32ValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryArrayInt64ConstructorCode = `// NewMethodQueryArrayInt64MethodQueryArrayInt64Payload builds a
// ServiceQueryArrayInt64 service MethodQueryArrayInt64 endpoint payload.
func NewMethodQueryArrayInt64MethodQueryArrayInt64Payload(q []int64) *servicequeryarrayint64.MethodQueryArrayInt64Payload {
	return &servicequeryarrayint64.MethodQueryArrayInt64Payload{
		Q: q,
	}
}
`

var PayloadQueryArrayInt64ValidateConstructorCode = `// NewMethodQueryArrayInt64ValidateMethodQueryArrayInt64ValidatePayload builds
// a ServiceQueryArrayInt64Validate service MethodQueryArrayInt64Validate
// endpoint payload.
func NewMethodQueryArrayInt64ValidateMethodQueryArrayInt64ValidatePayload(q []int64) *servicequeryarrayint64validate.MethodQueryArrayInt64ValidatePayload {
	return &servicequeryarrayint64validate.MethodQueryArrayInt64ValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryArrayUIntConstructorCode = `// NewMethodQueryArrayUIntMethodQueryArrayUIntPayload builds a
// ServiceQueryArrayUInt service MethodQueryArrayUInt endpoint payload.
func NewMethodQueryArrayUIntMethodQueryArrayUIntPayload(q []uint) *servicequeryarrayuint.MethodQueryArrayUIntPayload {
	return &servicequeryarrayuint.MethodQueryArrayUIntPayload{
		Q: q,
	}
}
`

var PayloadQueryArrayUIntValidateConstructorCode = `// NewMethodQueryArrayUIntValidateMethodQueryArrayUIntValidatePayload builds a
// ServiceQueryArrayUIntValidate service MethodQueryArrayUIntValidate endpoint
// payload.
func NewMethodQueryArrayUIntValidateMethodQueryArrayUIntValidatePayload(q []uint) *servicequeryarrayuintvalidate.MethodQueryArrayUIntValidatePayload {
	return &servicequeryarrayuintvalidate.MethodQueryArrayUIntValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryArrayUInt32ConstructorCode = `// NewMethodQueryArrayUInt32MethodQueryArrayUInt32Payload builds a
// ServiceQueryArrayUInt32 service MethodQueryArrayUInt32 endpoint payload.
func NewMethodQueryArrayUInt32MethodQueryArrayUInt32Payload(q []uint32) *servicequeryarrayuint32.MethodQueryArrayUInt32Payload {
	return &servicequeryarrayuint32.MethodQueryArrayUInt32Payload{
		Q: q,
	}
}
`

var PayloadQueryArrayUInt32ValidateConstructorCode = `// NewMethodQueryArrayUInt32ValidateMethodQueryArrayUInt32ValidatePayload
// builds a ServiceQueryArrayUInt32Validate service
// MethodQueryArrayUInt32Validate endpoint payload.
func NewMethodQueryArrayUInt32ValidateMethodQueryArrayUInt32ValidatePayload(q []uint32) *servicequeryarrayuint32validate.MethodQueryArrayUInt32ValidatePayload {
	return &servicequeryarrayuint32validate.MethodQueryArrayUInt32ValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryArrayUInt64ConstructorCode = `// NewMethodQueryArrayUInt64MethodQueryArrayUInt64Payload builds a
// ServiceQueryArrayUInt64 service MethodQueryArrayUInt64 endpoint payload.
func NewMethodQueryArrayUInt64MethodQueryArrayUInt64Payload(q []uint64) *servicequeryarrayuint64.MethodQueryArrayUInt64Payload {
	return &servicequeryarrayuint64.MethodQueryArrayUInt64Payload{
		Q: q,
	}
}
`

var PayloadQueryArrayUInt64ValidateConstructorCode = `// NewMethodQueryArrayUInt64ValidateMethodQueryArrayUInt64ValidatePayload
// builds a ServiceQueryArrayUInt64Validate service
// MethodQueryArrayUInt64Validate endpoint payload.
func NewMethodQueryArrayUInt64ValidateMethodQueryArrayUInt64ValidatePayload(q []uint64) *servicequeryarrayuint64validate.MethodQueryArrayUInt64ValidatePayload {
	return &servicequeryarrayuint64validate.MethodQueryArrayUInt64ValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryArrayFloat32ConstructorCode = `// NewMethodQueryArrayFloat32MethodQueryArrayFloat32Payload builds a
// ServiceQueryArrayFloat32 service MethodQueryArrayFloat32 endpoint payload.
func NewMethodQueryArrayFloat32MethodQueryArrayFloat32Payload(q []float32) *servicequeryarrayfloat32.MethodQueryArrayFloat32Payload {
	return &servicequeryarrayfloat32.MethodQueryArrayFloat32Payload{
		Q: q,
	}
}
`

var PayloadQueryArrayFloat32ValidateConstructorCode = `// NewMethodQueryArrayFloat32ValidateMethodQueryArrayFloat32ValidatePayload
// builds a ServiceQueryArrayFloat32Validate service
// MethodQueryArrayFloat32Validate endpoint payload.
func NewMethodQueryArrayFloat32ValidateMethodQueryArrayFloat32ValidatePayload(q []float32) *servicequeryarrayfloat32validate.MethodQueryArrayFloat32ValidatePayload {
	return &servicequeryarrayfloat32validate.MethodQueryArrayFloat32ValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryArrayFloat64ConstructorCode = `// NewMethodQueryArrayFloat64MethodQueryArrayFloat64Payload builds a
// ServiceQueryArrayFloat64 service MethodQueryArrayFloat64 endpoint payload.
func NewMethodQueryArrayFloat64MethodQueryArrayFloat64Payload(q []float64) *servicequeryarrayfloat64.MethodQueryArrayFloat64Payload {
	return &servicequeryarrayfloat64.MethodQueryArrayFloat64Payload{
		Q: q,
	}
}
`

var PayloadQueryArrayFloat64ValidateConstructorCode = `// NewMethodQueryArrayFloat64ValidateMethodQueryArrayFloat64ValidatePayload
// builds a ServiceQueryArrayFloat64Validate service
// MethodQueryArrayFloat64Validate endpoint payload.
func NewMethodQueryArrayFloat64ValidateMethodQueryArrayFloat64ValidatePayload(q []float64) *servicequeryarrayfloat64validate.MethodQueryArrayFloat64ValidatePayload {
	return &servicequeryarrayfloat64validate.MethodQueryArrayFloat64ValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryArrayStringConstructorCode = `// NewMethodQueryArrayStringMethodQueryArrayStringPayload builds a
// ServiceQueryArrayString service MethodQueryArrayString endpoint payload.
func NewMethodQueryArrayStringMethodQueryArrayStringPayload(q []string) *servicequeryarraystring.MethodQueryArrayStringPayload {
	return &servicequeryarraystring.MethodQueryArrayStringPayload{
		Q: q,
	}
}
`

var PayloadQueryArrayStringValidateConstructorCode = `// NewMethodQueryArrayStringValidateMethodQueryArrayStringValidatePayload
// builds a ServiceQueryArrayStringValidate service
// MethodQueryArrayStringValidate endpoint payload.
func NewMethodQueryArrayStringValidateMethodQueryArrayStringValidatePayload(q []string) *servicequeryarraystringvalidate.MethodQueryArrayStringValidatePayload {
	return &servicequeryarraystringvalidate.MethodQueryArrayStringValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryArrayBytesConstructorCode = `// NewMethodQueryArrayBytesMethodQueryArrayBytesPayload builds a
// ServiceQueryArrayBytes service MethodQueryArrayBytes endpoint payload.
func NewMethodQueryArrayBytesMethodQueryArrayBytesPayload(q [][]byte) *servicequeryarraybytes.MethodQueryArrayBytesPayload {
	return &servicequeryarraybytes.MethodQueryArrayBytesPayload{
		Q: q,
	}
}
`

var PayloadQueryArrayBytesValidateConstructorCode = `// NewMethodQueryArrayBytesValidateMethodQueryArrayBytesValidatePayload builds
// a ServiceQueryArrayBytesValidate service MethodQueryArrayBytesValidate
// endpoint payload.
func NewMethodQueryArrayBytesValidateMethodQueryArrayBytesValidatePayload(q [][]byte) *servicequeryarraybytesvalidate.MethodQueryArrayBytesValidatePayload {
	return &servicequeryarraybytesvalidate.MethodQueryArrayBytesValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryArrayAnyConstructorCode = `// NewMethodQueryArrayAnyMethodQueryArrayAnyPayload builds a
// ServiceQueryArrayAny service MethodQueryArrayAny endpoint payload.
func NewMethodQueryArrayAnyMethodQueryArrayAnyPayload(q []interface{}) *servicequeryarrayany.MethodQueryArrayAnyPayload {
	return &servicequeryarrayany.MethodQueryArrayAnyPayload{
		Q: q,
	}
}
`

var PayloadQueryArrayAnyValidateConstructorCode = `// NewMethodQueryArrayAnyValidateMethodQueryArrayAnyValidatePayload builds a
// ServiceQueryArrayAnyValidate service MethodQueryArrayAnyValidate endpoint
// payload.
func NewMethodQueryArrayAnyValidateMethodQueryArrayAnyValidatePayload(q []interface{}) *servicequeryarrayanyvalidate.MethodQueryArrayAnyValidatePayload {
	return &servicequeryarrayanyvalidate.MethodQueryArrayAnyValidatePayload{
		Q: q,
	}
}
`

var PayloadQueryStringMappedConstructorCode = `// NewMethodQueryStringMappedMethodQueryStringMappedPayload builds a
// ServiceQueryStringMapped service MethodQueryStringMapped endpoint payload.
func NewMethodQueryStringMappedMethodQueryStringMappedPayload(query *string) *servicequerystringmapped.MethodQueryStringMappedPayload {
	return &servicequerystringmapped.MethodQueryStringMappedPayload{
		Query: query,
	}
}
`

var PayloadPathStringConstructorCode = `// NewMethodPathStringMethodPathStringPayload builds a ServicePathString
// service MethodPathString endpoint payload.
func NewMethodPathStringMethodPathStringPayload(p string) *servicepathstring.MethodPathStringPayload {
	return &servicepathstring.MethodPathStringPayload{
		P: &p,
	}
}
`

var PayloadPathStringValidateConstructorCode = `// NewMethodPathStringValidateMethodPathStringValidatePayload builds a
// ServicePathStringValidate service MethodPathStringValidate endpoint payload.
func NewMethodPathStringValidateMethodPathStringValidatePayload(p string) *servicepathstringvalidate.MethodPathStringValidatePayload {
	return &servicepathstringvalidate.MethodPathStringValidatePayload{
		P: p,
	}
}
`

var PayloadPathArrayStringConstructorCode = `// NewMethodPathArrayStringMethodPathArrayStringPayload builds a
// ServicePathArrayString service MethodPathArrayString endpoint payload.
func NewMethodPathArrayStringMethodPathArrayStringPayload(p []string) *servicepatharraystring.MethodPathArrayStringPayload {
	return &servicepatharraystring.MethodPathArrayStringPayload{
		P: p,
	}
}
`

var PayloadPathArrayStringValidateConstructorCode = `// NewMethodPathArrayStringValidateMethodPathArrayStringValidatePayload builds
// a ServicePathArrayStringValidate service MethodPathArrayStringValidate
// endpoint payload.
func NewMethodPathArrayStringValidateMethodPathArrayStringValidatePayload(p []string) *servicepatharraystringvalidate.MethodPathArrayStringValidatePayload {
	return &servicepatharraystringvalidate.MethodPathArrayStringValidatePayload{
		P: p,
	}
}
`

var PayloadHeaderStringConstructorCode = `// NewMethodHeaderStringMethodHeaderStringPayload builds a ServiceHeaderString
// service MethodHeaderString endpoint payload.
func NewMethodHeaderStringMethodHeaderStringPayload(h *string) *serviceheaderstring.MethodHeaderStringPayload {
	return &serviceheaderstring.MethodHeaderStringPayload{
		H: h,
	}
}
`

var PayloadHeaderStringValidateConstructorCode = `// NewMethodHeaderStringValidateMethodHeaderStringValidatePayload builds a
// ServiceHeaderStringValidate service MethodHeaderStringValidate endpoint
// payload.
func NewMethodHeaderStringValidateMethodHeaderStringValidatePayload(h *string) *serviceheaderstringvalidate.MethodHeaderStringValidatePayload {
	return &serviceheaderstringvalidate.MethodHeaderStringValidatePayload{
		H: h,
	}
}
`

var PayloadHeaderArrayStringConstructorCode = `// NewMethodHeaderArrayStringMethodHeaderArrayStringPayload builds a
// ServiceHeaderArrayString service MethodHeaderArrayString endpoint payload.
func NewMethodHeaderArrayStringMethodHeaderArrayStringPayload(h []string) *serviceheaderarraystring.MethodHeaderArrayStringPayload {
	return &serviceheaderarraystring.MethodHeaderArrayStringPayload{
		H: h,
	}
}
`

var PayloadHeaderArrayStringValidateConstructorCode = `// NewMethodHeaderArrayStringValidateMethodHeaderArrayStringValidatePayload
// builds a ServiceHeaderArrayStringValidate service
// MethodHeaderArrayStringValidate endpoint payload.
func NewMethodHeaderArrayStringValidateMethodHeaderArrayStringValidatePayload(h []string) *serviceheaderarraystringvalidate.MethodHeaderArrayStringValidatePayload {
	return &serviceheaderarraystringvalidate.MethodHeaderArrayStringValidatePayload{
		H: h,
	}
}
`

var PayloadBodyQueryObjectConstructorCode = `// NewMethodBodyQueryObjectMethodBodyQueryObjectPayload builds a
// ServiceBodyQueryObject service MethodBodyQueryObject endpoint payload.
func NewMethodBodyQueryObjectMethodBodyQueryObjectPayload(body *MethodBodyQueryObjectRequestBody, b *string) *servicebodyqueryobject.MethodBodyQueryObjectPayload {
	v := &servicebodyqueryobject.MethodBodyQueryObjectPayload{
		A: body.A,
	}
	v.B = b
	return v
}
`

var PayloadBodyQueryObjectValidateConstructorCode = `// NewMethodBodyQueryObjectValidateMethodBodyQueryObjectValidatePayload builds
// a ServiceBodyQueryObjectValidate service MethodBodyQueryObjectValidate
// endpoint payload.
func NewMethodBodyQueryObjectValidateMethodBodyQueryObjectValidatePayload(body *MethodBodyQueryObjectValidateRequestBody, b string) *servicebodyqueryobjectvalidate.MethodBodyQueryObjectValidatePayload {
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

var PayloadBodyPathObjectConstructorCode = `// NewMethodBodyPathObjectMethodBodyPathObjectPayload builds a
// ServiceBodyPathObject service MethodBodyPathObject endpoint payload.
func NewMethodBodyPathObjectMethodBodyPathObjectPayload(body *MethodBodyPathObjectRequestBody, b string) *servicebodypathobject.MethodBodyPathObjectPayload {
	v := &servicebodypathobject.MethodBodyPathObjectPayload{
		A: body.A,
	}
	v.B = &b
	return v
}
`

var PayloadBodyPathObjectValidateConstructorCode = `// NewMethodBodyPathObjectValidateMethodBodyPathObjectValidatePayload builds a
// ServiceBodyPathObjectValidate service MethodBodyPathObjectValidate endpoint
// payload.
func NewMethodBodyPathObjectValidateMethodBodyPathObjectValidatePayload(body *MethodBodyPathObjectValidateRequestBody, b string) *servicebodypathobjectvalidate.MethodBodyPathObjectValidatePayload {
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

var PayloadBodyQueryPathObjectConstructorCode = `// NewMethodBodyQueryPathObjectMethodBodyQueryPathObjectPayload builds a
// ServiceBodyQueryPathObject service MethodBodyQueryPathObject endpoint
// payload.
func NewMethodBodyQueryPathObjectMethodBodyQueryPathObjectPayload(body *MethodBodyQueryPathObjectRequestBody, c string, b *string) *servicebodyquerypathobject.MethodBodyQueryPathObjectPayload {
	v := &servicebodyquerypathobject.MethodBodyQueryPathObjectPayload{
		A: body.A,
	}
	v.C = &c
	v.B = b
	return v
}
`

var PayloadBodyQueryPathObjectValidateConstructorCode = `// NewMethodBodyQueryPathObjectValidateMethodBodyQueryPathObjectValidatePayload
// builds a ServiceBodyQueryPathObjectValidate service
// MethodBodyQueryPathObjectValidate endpoint payload.
func NewMethodBodyQueryPathObjectValidateMethodBodyQueryPathObjectValidatePayload(body *MethodBodyQueryPathObjectValidateRequestBody, c string, b string) *servicebodyquerypathobjectvalidate.MethodBodyQueryPathObjectValidatePayload {
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
	v.Inner = unmarshalInnerTypeRequestBodyToInnerType(body.Inner)
	return v
}
`

var PayloadBodyUserInnerDefaultConstructorCode = `// NewMethodBodyUserInnerDefaultPayloadType builds a
// ServiceBodyUserInnerDefault service MethodBodyUserInnerDefault endpoint
// payload.
func NewMethodBodyUserInnerDefaultPayloadType(body *MethodBodyUserInnerDefaultRequestBody) *servicebodyuserinnerdefault.PayloadType {
	v := &servicebodyuserinnerdefault.PayloadType{}
	v.Inner = unmarshalInnerTypeRequestBodyToInnerType(body.Inner)
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
	v.C = unmarshalPayloadTypeRequestBodyToPayloadType(body.C)
	v.A = a
	v.B = b
	return v
}
`
