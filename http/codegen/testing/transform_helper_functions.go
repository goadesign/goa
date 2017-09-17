package testing

var PayloadBodyUserInnerDefaultTransformCode1 = `// unmarshalInnerTypeRequestBodyToInnerType builds a value of type
// *servicebodyuserinnerdefault.InnerType from a value of type
// *InnerTypeRequestBody.
func unmarshalInnerTypeRequestBodyToInnerType(v *InnerTypeRequestBody) *servicebodyuserinnerdefault.InnerType {
	res := &servicebodyuserinnerdefault.InnerType{
		A: *v.A,
	}
	if v.B != nil {
		res.B = *v.B
	}

	return res
}
`

var PayloadBodyUserInnerDefaultTransformCode2 = `// unmarshalInnerTypeRequestBodyToInnerType builds a value of type
// *servicebodyuserinnerdefault.InnerType from a value of type
// *InnerTypeRequestBody.
func unmarshalInnerTypeRequestBodyToInnerType(v *InnerTypeRequestBody) *servicebodyuserinnerdefault.InnerType {
	res := &servicebodyuserinnerdefault.InnerType{
		A: *v.A,
	}
	if v.B != nil {
		res.B = *v.B
	}

	return res
}
`

var PayloadBodyInlineRecursiveUserTransformCode1 = `// unmarshalPayloadTypeRequestBodyToPayloadType builds a value of type
// *servicebodyinlinerecursiveuser.PayloadType from a value of type
// *PayloadTypeRequestBody.
func unmarshalPayloadTypeRequestBodyToPayloadType(v *PayloadTypeRequestBody) *servicebodyinlinerecursiveuser.PayloadType {
	res := &servicebodyinlinerecursiveuser.PayloadType{
		A: *v.A,
		B: v.B,
	}
	res.C = unmarshalPayloadTypeRequestBodyToPayloadType(v.C)

	return res
}
`

var PayloadBodyInlineRecursiveUserTransformCode2 = `// unmarshalPayloadTypeRequestBodyToPayloadType builds a value of type
// *servicebodyinlinerecursiveuser.PayloadType from a value of type
// *PayloadTypeRequestBody.
func unmarshalPayloadTypeRequestBodyToPayloadType(v *PayloadTypeRequestBody) *servicebodyinlinerecursiveuser.PayloadType {
	res := &servicebodyinlinerecursiveuser.PayloadType{
		A: *v.A,
		B: v.B,
	}
	res.C = unmarshalPayloadTypeRequestBodyToPayloadType(v.C)

	return res
}
`

var PayloadBodyUserInnerDefaultTransformCodeCLI1 = `// marshalInnerTypeRequestBodyToInnerType builds a value of type
// *servicebodyuserinnerdefault.InnerType from a value of type
// *InnerTypeRequestBody.
func marshalInnerTypeRequestBodyToInnerType(v *InnerTypeRequestBody) *servicebodyuserinnerdefault.InnerType {
	res := &servicebodyuserinnerdefault.InnerType{
		A: v.A,
		B: v.B,
	}

	return res
}
`

var PayloadBodyUserInnerDefaultTransformCodeCLI2 = `// marshalInnerTypeToInnerTypeRequestBody builds a value of type
// *InnerTypeRequestBody from a value of type
// *servicebodyuserinnerdefault.InnerType.
func marshalInnerTypeToInnerTypeRequestBody(v *servicebodyuserinnerdefault.InnerType) *InnerTypeRequestBody {
	res := &InnerTypeRequestBody{
		A: v.A,
		B: v.B,
	}

	return res
}
`

var PayloadBodyInlineRecursiveUserTransformCodeCLI1 = `// marshalPayloadTypeRequestBodyToPayloadType builds a value of type
// *servicebodyinlinerecursiveuser.PayloadType from a value of type
// *PayloadTypeRequestBody.
func marshalPayloadTypeRequestBodyToPayloadType(v *PayloadTypeRequestBody) *servicebodyinlinerecursiveuser.PayloadType {
	res := &servicebodyinlinerecursiveuser.PayloadType{
		A: v.A,
		B: v.B,
	}
	if v.C != nil {
		res.C = marshalPayloadTypeRequestBodyToPayloadType(v.C)
	}

	return res
}
`

var PayloadBodyInlineRecursiveUserTransformCodeCLI2 = `// marshalPayloadTypeToPayloadTypeRequestBody builds a value of type
// *PayloadTypeRequestBody from a value of type
// *servicebodyinlinerecursiveuser.PayloadType.
func marshalPayloadTypeToPayloadTypeRequestBody(v *servicebodyinlinerecursiveuser.PayloadType) *PayloadTypeRequestBody {
	res := &PayloadTypeRequestBody{
		A: v.A,
		B: v.B,
	}
	if v.C != nil {
		res.C = marshalPayloadTypeToPayloadTypeRequestBody(v.C)
	}

	return res
}
`
