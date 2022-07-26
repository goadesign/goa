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
