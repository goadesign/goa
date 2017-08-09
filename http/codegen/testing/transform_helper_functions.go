package testing

var PayloadBodyUserInnerDefaultTransformCode1 = `// innerTypeRequestBodyToInnerType builds a value of type
// *servicebodyuserinnerdefault.InnerType from a value of type
// *InnerTypeRequestBody.
func innerTypeRequestBodyToInnerType(v *InnerTypeRequestBody) *servicebodyuserinnerdefault.InnerType {
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
// *InnerTypeRequestBody from a value of type
// *servicebodyuserinnerdefault.InnerType.
func innerTypeToInnerTypeRequestBodyNoDefault(v *servicebodyuserinnerdefault.InnerType) *InnerTypeRequestBody {
	res := &InnerTypeRequestBody{
		A: v.A,
		B: v.B,
	}

	return res
}
`

var PayloadBodyInlineRecursiveUserTransformCode1 = `// payloadTypeRequestBodyToPayloadType builds a value of type
// *servicebodyinlinerecursiveuser.PayloadType from a value of type
// *PayloadTypeRequestBody.
func payloadTypeRequestBodyToPayloadType(v *PayloadTypeRequestBody) *servicebodyinlinerecursiveuser.PayloadType {
	res := &servicebodyinlinerecursiveuser.PayloadType{
		A: v.A,
		B: v.B,
	}
	res.C = payloadTypeRequestBodyToPayloadType(v.C)

	return res
}
`

var PayloadBodyInlineRecursiveUserTransformCode2 = `// payloadTypeToPayloadTypeRequestBodyNoDefault builds a value of type
// *PayloadTypeRequestBody from a value of type
// *servicebodyinlinerecursiveuser.PayloadType.
func payloadTypeToPayloadTypeRequestBodyNoDefault(v *servicebodyinlinerecursiveuser.PayloadType) *PayloadTypeRequestBody {
	res := &PayloadTypeRequestBody{
		A: v.A,
		B: v.B,
	}
	res.C = payloadTypeToPayloadTypeRequestBodyNoDefault(v.C)

	return res
}
`
