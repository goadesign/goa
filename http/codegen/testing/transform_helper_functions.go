package testing

var PayloadBodyUserInnerDefaultTransformCode1 = `// innerTypeRequestBodyToInnerTypeSrcPtr builds a value of type
// *servicebodyuserinnerdefault.InnerType from a value of type
// *InnerTypeRequestBody.
func innerTypeRequestBodyToInnerTypeSrcPtr(v *InnerTypeRequestBody) *servicebodyuserinnerdefault.InnerType {
	res := &servicebodyuserinnerdefault.InnerType{
		A: v.A,
		B: v.B,
	}
	if v.A == nil {
		tmp := "defaulta"
		res.A = &tmp
	}
	if v.B == nil {
		tmp := "defaultb"
		res.B = &tmp
	}

	return res
}
`

var PayloadBodyUserInnerDefaultTransformCode2 = `// innerTypeToInnerTypeRequestBodyNoDefault builds a value of type
// *InnerTypeRequestBody from a value of type *InnerType.
func innerTypeToInnerTypeRequestBodyNoDefault(v *InnerType) *InnerTypeRequestBody {
	res := &InnerTypeRequestBody{
		A: v.A,
		B: v.B,
	}

	return res
}
`

var PayloadBodyUserInnerDefaultTransformCode3 = `// innerTypeToInnerTypeRequestBodyTgtPtr builds a value of type
// *InnerTypeRequestBody from a value of type *InnerType.
func innerTypeToInnerTypeRequestBodyTgtPtr(v *InnerType) *InnerTypeRequestBody {
	res := &InnerTypeRequestBody{
		A: &v.A,
		B: &v.B,
	}
	if v.A == nil {
		tmp := "defaulta"
		res.A = &tmp
	}
	if v.B == nil {
		tmp := "defaultb"
		res.B = &tmp
	}

	return res
}
`

var PayloadBodyInlineRecursiveUserTransformCode1 = `// payloadTypeRequestBodyToPayloadTypeSrcPtr builds a value of type
// *servicebodyinlinerecursiveuser.PayloadType from a value of type
// *PayloadTypeRequestBody.
func payloadTypeRequestBodyToPayloadTypeSrcPtr(v *PayloadTypeRequestBody) *servicebodyinlinerecursiveuser.PayloadType {
	res := &servicebodyinlinerecursiveuser.PayloadType{
		A: *v.A,
		B: v.B,
	}
	res.C = payloadTypeRequestBodyToPayloadTypeSrcPtr(v.C)

	return res
}
`

var PayloadBodyInlineRecursiveUserTransformCode2 = `// payloadTypeToPayloadTypeRequestBodyNoDefault builds a value of type
// *PayloadTypeRequestBody from a value of type *PayloadType.
func payloadTypeToPayloadTypeRequestBodyNoDefault(v *PayloadType) *PayloadTypeRequestBody {
	res := &PayloadTypeRequestBody{
		A: v.A,
		B: v.B,
	}
	res.C = payloadTypeToPayloadTypeRequestBodyNoDefault(v.C)

	return res
}
`

var PayloadBodyInlineRecursiveUserTransformCode3 = `// payloadTypeToPayloadTypeRequestBodyTgtPtr builds a value of type
// *PayloadTypeRequestBody from a value of type *PayloadType.
func payloadTypeToPayloadTypeRequestBodyTgtPtr(v *PayloadType) *PayloadTypeRequestBody {
	res := &PayloadTypeRequestBody{
		A: &v.A,
		B: v.B,
	}
	res.C = payloadTypeToPayloadTypeRequestBodyTgtPtr(v.C)

	return res
}
`
