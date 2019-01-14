package testdata

var PayloadBodyUserInnerDefaultTransformCode1 = `// unmarshalInnerTypeRequestBodyToServicebodyuserinnerdefaultInnerType builds a
// value of type *servicebodyuserinnerdefault.InnerType from a value of type
// *InnerTypeRequestBody.
func unmarshalInnerTypeRequestBodyToServicebodyuserinnerdefaultInnerType(v *InnerTypeRequestBody) *servicebodyuserinnerdefault.InnerType {
	if v == nil {
		return nil
	}
	res := &servicebodyuserinnerdefault.InnerType{
		A: *v.A,
	}
	if v.B != nil {
		res.B = *v.B
	}
	if v.A == nil {
		res.A = "defaulta"
	}
	if v.B == nil {
		res.B = "defaultb"
	}

	return res
}
`

var PayloadBodyUserInnerDefaultTransformCode2 = `// unmarshalInnerTypeRequestBodyToServicebodyuserinnerdefaultInnerType builds a
// value of type *servicebodyuserinnerdefault.InnerType from a value of type
// *InnerTypeRequestBody.
func unmarshalInnerTypeRequestBodyToServicebodyuserinnerdefaultInnerType(v *InnerTypeRequestBody) *servicebodyuserinnerdefault.InnerType {
	if v == nil {
		return nil
	}
	res := &servicebodyuserinnerdefault.InnerType{
		A: *v.A,
	}
	if v.B != nil {
		res.B = *v.B
	}
	if v.A == nil {
		res.A = "defaulta"
	}
	if v.B == nil {
		res.B = "defaultb"
	}

	return res
}
`

var PayloadBodyInlineRecursiveUserTransformCode1 = `// unmarshalPayloadTypeRequestBodyToServicebodyinlinerecursiveuserPayloadType
// builds a value of type *servicebodyinlinerecursiveuser.PayloadType from a
// value of type *PayloadTypeRequestBody.
func unmarshalPayloadTypeRequestBodyToServicebodyinlinerecursiveuserPayloadType(v *PayloadTypeRequestBody) *servicebodyinlinerecursiveuser.PayloadType {
	res := &servicebodyinlinerecursiveuser.PayloadType{
		A: *v.A,
		B: v.B,
	}
	if v.C != nil {
		res.C = unmarshalPayloadTypeRequestBodyToServicebodyinlinerecursiveuserPayloadType(v.C)
	}

	return res
}
`

var PayloadBodyInlineRecursiveUserTransformCode2 = `// unmarshalPayloadTypeRequestBodyToServicebodyinlinerecursiveuserPayloadType
// builds a value of type *servicebodyinlinerecursiveuser.PayloadType from a
// value of type *PayloadTypeRequestBody.
func unmarshalPayloadTypeRequestBodyToServicebodyinlinerecursiveuserPayloadType(v *PayloadTypeRequestBody) *servicebodyinlinerecursiveuser.PayloadType {
	res := &servicebodyinlinerecursiveuser.PayloadType{
		A: *v.A,
		B: v.B,
	}
	if v.C != nil {
		res.C = unmarshalPayloadTypeRequestBodyToServicebodyinlinerecursiveuserPayloadType(v.C)
	}

	return res
}
`

var PayloadBodyUserInnerDefaultTransformCodeCLI1 = `// marshalInnerTypeRequestBodyToServicebodyuserinnerdefaultInnerType builds a
// value of type *servicebodyuserinnerdefault.InnerType from a value of type
// *InnerTypeRequestBody.
func marshalInnerTypeRequestBodyToServicebodyuserinnerdefaultInnerType(v *InnerTypeRequestBody) *servicebodyuserinnerdefault.InnerType {
	if v == nil {
		return nil
	}
	res := &servicebodyuserinnerdefault.InnerType{
		A: v.A,
		B: v.B,
	}

	return res
}
`

var PayloadBodyUserInnerDefaultTransformCodeCLI2 = `// marshalServicebodyuserinnerdefaultInnerTypeToInnerTypeRequestBody builds a
// value of type *InnerTypeRequestBody from a value of type
// *servicebodyuserinnerdefault.InnerType.
func marshalServicebodyuserinnerdefaultInnerTypeToInnerTypeRequestBody(v *servicebodyuserinnerdefault.InnerType) *InnerTypeRequestBody {
	if v == nil {
		return nil
	}
	res := &InnerTypeRequestBody{
		A: v.A,
		B: v.B,
	}

	return res
}
`

var PayloadBodyInlineRecursiveUserTransformCodeCLI1 = `// marshalPayloadTypeRequestBodyToServicebodyinlinerecursiveuserPayloadType
// builds a value of type *servicebodyinlinerecursiveuser.PayloadType from a
// value of type *PayloadTypeRequestBody.
func marshalPayloadTypeRequestBodyToServicebodyinlinerecursiveuserPayloadType(v *PayloadTypeRequestBody) *servicebodyinlinerecursiveuser.PayloadType {
	res := &servicebodyinlinerecursiveuser.PayloadType{
		A: v.A,
		B: v.B,
	}
	res.C = marshalPayloadTypeRequestBodyToServicebodyinlinerecursiveuserPayloadType(v.C)

	return res
}
`

var PayloadBodyInlineRecursiveUserTransformCodeCLI2 = `// marshalServicebodyinlinerecursiveuserPayloadTypeToPayloadTypeRequestBody
// builds a value of type *PayloadTypeRequestBody from a value of type
// *servicebodyinlinerecursiveuser.PayloadType.
func marshalServicebodyinlinerecursiveuserPayloadTypeToPayloadTypeRequestBody(v *servicebodyinlinerecursiveuser.PayloadType) *PayloadTypeRequestBody {
	res := &PayloadTypeRequestBody{
		A: v.A,
		B: v.B,
	}
	res.C = marshalServicebodyinlinerecursiveuserPayloadTypeToPayloadTypeRequestBody(v.C)

	return res
}
`
