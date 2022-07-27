package testdata

var EmbeddedCustomPkgTypeUnmarshalCode = `// unmarshalFooRequestBodyToFooFoo builds a value of type *foo.Foo from a value
// of type *FooRequestBody.
func unmarshalFooRequestBodyToFooFoo(v *FooRequestBody) *foo.Foo {
	if v == nil {
		return nil
	}
	res := &foo.Foo{
		Bar: v.Bar,
	}

	return res
}
`

var EmbeddedCustomPkgTypeMarshalCode = `// marshalFooFooToFooResponseBody builds a value of type *FooResponseBody from
// a value of type *foo.Foo.
func marshalFooFooToFooResponseBody(v *foo.Foo) *FooResponseBody {
	if v == nil {
		return nil
	}
	res := &FooResponseBody{
		Bar: v.Bar,
	}

	return res
}
`

var ArrayAliasExtendedUnmarshalCode = `// unmarshalResultTypeRequestBodyToFooserviceResultType builds a value of type
// *fooservice.ResultType from a value of type *ResultTypeRequestBody.
func unmarshalResultTypeRequestBodyToFooserviceResultType(v *ResultTypeRequestBody) *fooservice.ResultType {
	res := &fooservice.ResultType{}
	if v.Foo != nil {
		foo := fooservice.Foo(*v.Foo)
		res.Foo = &foo
	}

	return res
}
`

var ArrayAliasExtendedMarshalCode = `// marshalFooserviceResultTypeToResultTypeResponse builds a value of type
// *ResultTypeResponse from a value of type *fooservice.ResultType.
func marshalFooserviceResultTypeToResultTypeResponse(v *fooservice.ResultType) *ResultTypeResponse {
	res := &ResultTypeResponse{}
	if v.Foo != nil {
		foo := string(*v.Foo)
		res.Foo = &foo
	}

	return res
}
`

var ExtensionWithAliasUnmarshalExtensionCode = `// unmarshalExtensionRequestBodyToFooserviceExtension builds a value of type
// *fooservice.Extension from a value of type *ExtensionRequestBody.
func unmarshalExtensionRequestBodyToFooserviceExtension(v *ExtensionRequestBody) *fooservice.Extension {
	if v == nil {
		return nil
	}
	res := &fooservice.Extension{}
	if v.Bar != nil {
		res.Bar = unmarshalBarRequestBodyToFooserviceBar(v.Bar)
	}

	return res
}
`

var ExtensionWithAliasUnmarshalBarCode = `// unmarshalBarRequestBodyToFooserviceBar builds a value of type
// *fooservice.Bar from a value of type *BarRequestBody.
func unmarshalBarRequestBodyToFooserviceBar(v *BarRequestBody) *fooservice.Bar {
	if v == nil {
		return nil
	}
	res := &fooservice.Bar{
		Bar: *v.Bar,
	}

	return res
}
`

var ExtensionWithAliasMarshalResultCode = `// marshalFooserviceResultTypeToResultTypeResponse builds a value of type
// *ResultTypeResponse from a value of type *fooservice.ResultType.
func marshalFooserviceResultTypeToResultTypeResponse(v *fooservice.ResultType) *ResultTypeResponse {
	res := &ResultTypeResponse{}
	if v.Extension != nil {
		res.Extension = marshalFooserviceExtensionToExtensionResponse(v.Extension)
	}

	return res
}
`

var ExtensionWithAliasMarshalExtensionCode = `// marshalFooserviceExtensionToExtensionResponse builds a value of type
// *ExtensionResponse from a value of type *fooservice.Extension.
func marshalFooserviceExtensionToExtensionResponse(v *fooservice.Extension) *ExtensionResponse {
	if v == nil {
		return nil
	}
	res := &ExtensionResponse{}
	if v.Bar != nil {
		res.Bar = marshalFooserviceBarToBarResponse(v.Bar)
	}

	return res
}
`

var ExtensionWithAliasMarshalBarCode = `// marshalFooserviceBarToBarResponse builds a value of type *BarResponse from a
// value of type *fooservice.Bar.
func marshalFooserviceBarToBarResponse(v *fooservice.Bar) *BarResponse {
	if v == nil {
		return nil
	}
	res := &BarResponse{
		Bar: v.Bar,
	}

	return res
}
`
