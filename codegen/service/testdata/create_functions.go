package testdata

var CreateStringCode = `// CreateFromStringT initializes t from the fields of v
func (t *StringType) CreateFromStringT(v *testdata.StringT) {
	temp := &StringType{
		String: &v.String,
	}
	*t = *temp
}
`

var CreateStringRequiredCode = `// CreateFromStringT initializes t from the fields of v
func (t *StringType) CreateFromStringT(v *testdata.StringT) {
	temp := &StringType{
		String: v.String,
	}
	*t = *temp
}
`

var CreateStringPointerCode = `// CreateFromStringPointerT initializes t from the fields of v
func (t *StringType) CreateFromStringPointerT(v *testdata.StringPointerT) {
	temp := &StringType{
		String: v.String,
	}
	*t = *temp
}
`

var CreateStringPointerRequiredCode = `// CreateFromStringPointerT initializes t from the fields of v
func (t *StringType) CreateFromStringPointerT(v *testdata.StringPointerT) {
	temp := &StringType{}
	if v.String != nil {
		temp.String = *v.String
	}
	*t = *temp
}
`

var CreateExternalNameCode = `// CreateFromExternalNameT initializes t from the fields of v
func (t *ExternalNameType) CreateFromExternalNameT(v *testdata.ExternalNameT) {
	temp := &ExternalNameType{
		String: &v.String,
	}
	*t = *temp
}
`

var CreateExternalNameRequiredCode = `// CreateFromExternalNameT initializes t from the fields of v
func (t *ExternalNameType) CreateFromExternalNameT(v *testdata.ExternalNameT) {
	temp := &ExternalNameType{
		String: v.String,
	}
	*t = *temp
}
`

var CreateExternalNamePointerCode = `// CreateFromExternalNamePointerT initializes t from the fields of v
func (t *ExternalNamePointerType) CreateFromExternalNamePointerT(v *testdata.ExternalNamePointerT) {
	temp := &ExternalNamePointerType{
		String: v.String,
	}
	*t = *temp
}
`

var CreateExternalNamePointerRequiredCode = `// CreateFromExternalNamePointerT initializes t from the fields of v
func (t *ExternalNamePointerType) CreateFromExternalNamePointerT(v *testdata.ExternalNamePointerT) {
	temp := &ExternalNamePointerType{}
	if v.String != nil {
		temp.String = *v.String
	}
	*t = *temp
}
`

var CreateArrayStringCode = `// CreateFromArrayStringT initializes t from the fields of v
func (t *ArrayStringType) CreateFromArrayStringT(v *testdata.ArrayStringT) {
	temp := &ArrayStringType{}
	if v.ArrayString != nil {
		temp.ArrayString = make([]string, len(v.ArrayString))
		for i, val := range v.ArrayString {
			temp.ArrayString[i] = val
		}
	}
	*t = *temp
}
`

var CreateArrayStringRequiredCode = `// CreateFromArrayStringT initializes t from the fields of v
func (t *ArrayStringType) CreateFromArrayStringT(v *testdata.ArrayStringT) {
	temp := &ArrayStringType{}
	if v.ArrayString != nil {
		temp.ArrayString = make([]string, len(v.ArrayString))
		for i, val := range v.ArrayString {
			temp.ArrayString[i] = val
		}
	}
	*t = *temp
}
`

var CreateObjectCode = `// CreateFromObjectT initializes t from the fields of v
func (t *ObjectType) CreateFromObjectT(v *testdata.ObjectT) {
	temp := &ObjectType{}
	if v.Object != nil {
		temp.Object = transformTestdataObjectFieldTToObjectField(v.Object)
	}
	*t = *temp
}
`

var CreateObjectRequiredCode = `// CreateFromObjectT initializes t from the fields of v
func (t *ObjectType) CreateFromObjectT(v *testdata.ObjectT) {
	temp := &ObjectType{}
	if v.Object != nil {
		temp.Object = transformTestdataObjectFieldTToObjectField(v.Object)
	}
	*t = *temp
}
`

var CreateObjectExtraCode = `// CreateFromObjectExtraT initializes t from the fields of v
func (t *ObjectType) CreateFromObjectExtraT(v *testdata.ObjectExtraT) {
	temp := &ObjectType{}
	if v.Object != nil {
		temp.Object = transformTestdataObjectFieldTToObjectField(v.Object)
	}
	*t = *temp
}
`

var CreateAliasConvert = `// Service service type conversion functions
//
// Command:
// $ goa

package service

import (
	aliasd "goa.design/goa/codegen/service/testdata/alias-external"
)

// CreateFromConvertModel initializes t from the fields of v
func (t *StringType) CreateFromConvertModel(v *aliasd.ConvertModel) {
	temp := &StringType{
		Bar: &v.Bar,
	}
	*t = *temp
}
`
