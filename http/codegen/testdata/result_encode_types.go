package testdata

var ResultWithEmbeddedCustomPkgTypeUnmarshalCode = `// unmarshalFooRequestBodyToFooFoo builds a value of type *foo.Foo from a value
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

var ResultWithEmbeddedCustomPkgTypeMarshalCode = `// marshalFooFooToFooResponseBody builds a value of type *FooResponseBody from
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
