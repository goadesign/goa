package testdata

var ConvertStringCode = `// ConvertToStringT creates an instance of StringT initialized from t.
func (t *StringType) ConvertToStringT() *testdata.StringT {
	v := &testdata.StringT{}
	if t.String != nil {
		v.String = *t.String
	}
	return v
}
`

var ConvertStringRequiredCode = `// ConvertToStringT creates an instance of StringT initialized from t.
func (t *StringType) ConvertToStringT() *testdata.StringT {
	v := &testdata.StringT{
		String: t.String,
	}
	return v
}
`

var ConvertStringPointerCode = `// ConvertToStringPointerT creates an instance of StringPointerT initialized
// from t.
func (t *StringPointerType) ConvertToStringPointerT() *testdata.StringPointerT {
	v := &testdata.StringPointerT{
		String: t.String,
	}
	return v
}
`

var ConvertStringPointerRequiredCode = `// ConvertToStringPointerT creates an instance of StringPointerT initialized
// from t.
func (t *StringPointerType) ConvertToStringPointerT() *testdata.StringPointerT {
	v := &testdata.StringPointerT{
		String: &t.String,
	}
	return v
}
`

var ConvertExternalNameCode = `// ConvertToExternalNameT creates an instance of ExternalNameT initialized from
// t.
func (t *ExternalNameType) ConvertToExternalNameT() *testdata.ExternalNameT {
	v := &testdata.ExternalNameT{}
	if t.String != nil {
		v.String = *t.String
	}
	return v
}
`

var ConvertExternalNameRequiredCode = `// ConvertToExternalNameT creates an instance of ExternalNameT initialized from
// t.
func (t *ExternalNameType) ConvertToExternalNameT() *testdata.ExternalNameT {
	v := &testdata.ExternalNameT{
		String: t.String,
	}
	return v
}
`

var ConvertExternalNamePointerCode = `// ConvertToExternalNamePointerT creates an instance of ExternalNamePointerT
// initialized from t.
func (t *ExternalNamePointerType) ConvertToExternalNamePointerT() *testdata.ExternalNamePointerT {
	v := &testdata.ExternalNamePointerT{
		String: t.String,
	}
	return v
}
`

var ConvertExternalNamePointerRequiredCode = `// ConvertToExternalNamePointerT creates an instance of ExternalNamePointerT
// initialized from t.
func (t *ExternalNamePointerType) ConvertToExternalNamePointerT() *testdata.ExternalNamePointerT {
	v := &testdata.ExternalNamePointerT{
		String: &t.String,
	}
	return v
}
`

var ConvertArrayStringCode = `// ConvertToArrayStringT creates an instance of ArrayStringT initialized from t.
func (t *ArrayStringType) ConvertToArrayStringT() *testdata.ArrayStringT {
	v := &testdata.ArrayStringT{}
	if t.ArrayString != nil {
		v.ArrayString = make([]string, len(t.ArrayString))
		for i, val := range t.ArrayString {
			v.ArrayString[i] = val
		}
	}
	return v
}
`

var ConvertArrayStringRequiredCode = `// ConvertToArrayStringT creates an instance of ArrayStringT initialized from t.
func (t *ArrayStringType) ConvertToArrayStringT() *testdata.ArrayStringT {
	v := &testdata.ArrayStringT{}
	v.ArrayString = make([]string, len(t.ArrayString))
	for i, val := range t.ArrayString {
		v.ArrayString[i] = val
	}
	return v
}
`

var ConvertObjectCode = `// ConvertToObjectT creates an instance of ObjectT initialized from t.
func (t *ObjectType) ConvertToObjectT() *testdata.ObjectT {
	v := &testdata.ObjectT{}
	v.Object = transformObjectFieldToTestdataObjectFieldT(t.Object)
	return v
}
`

var ConvertObjectRequiredCode = `// ConvertToObjectT creates an instance of ObjectT initialized from t.
func (t *ObjectType) ConvertToObjectT() *testdata.ObjectT {
	v := &testdata.ObjectT{}
	v.Object = transformObjectFieldToTestdataObjectFieldT(t.Object)
	return v
}
`

var ConvertObjectHelperCode = `// transformObjectFieldToTestdataObjectFieldT builds a value of type
// *testdata.ObjectFieldT from a value of type *ObjectField.
func transformObjectFieldToTestdataObjectFieldT(v *ObjectField) *testdata.ObjectFieldT {
	res := &testdata.ObjectFieldT{
		Bytes: v.Bytes,
	}
	if v.Bool != nil {
		res.Bool = *v.Bool
	}
	if v.Int != nil {
		res.Int = *v.Int
	}
	if v.Int32 != nil {
		res.Int32 = *v.Int32
	}
	if v.Int64 != nil {
		res.Int64 = *v.Int64
	}
	if v.UInt != nil {
		res.UInt = *v.UInt
	}
	if v.UInt32 != nil {
		res.UInt32 = *v.UInt32
	}
	if v.UInt64 != nil {
		res.UInt64 = *v.UInt64
	}
	if v.Float32 != nil {
		res.Float32 = *v.Float32
	}
	if v.Float64 != nil {
		res.Float64 = *v.Float64
	}
	if v.String != nil {
		res.String = *v.String
	}
	if v.Array != nil {
		res.Array = make([]bool, len(v.Array))
		for i, val := range v.Array {
			res.Array[i] = val
		}
	}
	if v.Map != nil {
		res.Map = make(map[string]bool, len(v.Map))
		for key, val := range v.Map {
			tk := key
			tv := val
			res.Map[tk] = tv
		}
	}

	return res
}
`

var ConvertObjectRequiredHelperCode = `// transformObjectFieldToTestdataObjectFieldT builds a value of type
// *testdata.ObjectFieldT from a value of type *ObjectField.
func transformObjectFieldToTestdataObjectFieldT(v *ObjectField) *testdata.ObjectFieldT {
	res := &testdata.ObjectFieldT{
		Bool:    v.Bool,
		Int:     v.Int,
		Int32:   v.Int32,
		Int64:   v.Int64,
		UInt:    v.UInt,
		UInt32:  v.UInt32,
		UInt64:  v.UInt64,
		Float32: v.Float32,
		Float64: v.Float64,
		Bytes:   v.Bytes,
		String:  v.String,
	}
	res.Array = make([]bool, len(v.Array))
	for i, val := range v.Array {
		res.Array[i] = val
	}
	res.Map = make(map[string]bool, len(v.Map))
	for key, val := range v.Map {
		tk := key
		tv := val
		res.Map[tk] = tv
	}

	return res
}
`

var CreateExternalConvert = `// Service service type conversion functions
//
// Command:
// $ goa

package service

import (
	external "goa.design/goa/codegen/service/testdata/external"
)

// CreateFromConvertModel initializes t from the fields of v
func (t *StringType) CreateFromConvertModel(v *external.ConvertModel) {
	temp := &StringType{
		Foo: &v.Foo,
	}
	*t = *temp
}
`
