package testing

var ConvertStringCode = `// ConvertToStringT creates an instance of StringT initialized from t.
func (t *StringType) ConvertToStringT() *testing.StringT {
	var v *testing.StringT
	v := &testing.StringT{}
	if t.String != nil {
		v.String = *t.String
	}
	return v
}
`

var ConvertStringRequiredCode = `// ConvertToStringT creates an instance of StringT initialized from t.
func (t *StringType) ConvertToStringT() *testing.StringT {
	var v *testing.StringT
	v := &testing.StringT{
		String: t.String,
	}
	return v
}
`

var ConvertStringPointerCode = `// ConvertToStringPointerT creates an instance of StringPointerT initialized
// from t.
func (t *StringPointerType) ConvertToStringPointerT() *testing.StringPointerT {
	var v *testing.StringPointerT
	v := &testing.StringPointerT{
		String: t.String,
	}
	return v
}
`

var ConvertStringPointerRequiredCode = `// ConvertToStringPointerT creates an instance of StringPointerT initialized
// from t.
func (t *StringPointerType) ConvertToStringPointerT() *testing.StringPointerT {
	var v *testing.StringPointerT
	v := &testing.StringPointerT{
		String: &t.String,
	}
	return v
}
`

var ConvertArrayStringCode = `// ConvertToArrayStringT creates an instance of ArrayStringT initialized from t.
func (t *ArrayStringType) ConvertToArrayStringT() *testing.ArrayStringT {
	var v *testing.ArrayStringT
	v := &testing.ArrayStringT{}
	if t.ArrayString != nil {
		v.ArrayString = make([]string, len(t.ArrayString))
		for j, val := range t.ArrayString {
			v.ArrayString[j] = val
		}
	}
	return v
}
`

var ConvertArrayStringRequiredCode = `// ConvertToArrayStringT creates an instance of ArrayStringT initialized from t.
func (t *ArrayStringType) ConvertToArrayStringT() *testing.ArrayStringT {
	var v *testing.ArrayStringT
	v := &testing.ArrayStringT{}
	if t.ArrayString != nil {
		v.ArrayString = make([]string, len(t.ArrayString))
		for j, val := range t.ArrayString {
			v.ArrayString[j] = val
		}
	}
	return v
}
`

var ConvertObjectCode = `// ConvertToObjectT creates an instance of ObjectT initialized from t.
func (t *ObjectType) ConvertToObjectT() *testing.ObjectT {
	var v *testing.ObjectT
	v := &testing.ObjectT{}
	if t.Object != nil {
		v.Object = marshalObjectFieldToObjectFieldT(t.Object)
	}
	return v
}
`

var ConvertObjectRequiredCode = `// ConvertToObjectT creates an instance of ObjectT initialized from t.
func (t *ObjectType) ConvertToObjectT() *testing.ObjectT {
	var v *testing.ObjectT
	v := &testing.ObjectT{}
	if t.Object != nil {
		v.Object = marshalObjectFieldToObjectFieldT(t.Object)
	}
	return v
}
`

var ConvertObjectHelperCode = `// marshalObjectFieldToObjectFieldT builds a value of type
// *testing.ObjectFieldT from a value of type *ObjectField.
func marshalObjectFieldToObjectFieldT(v *ObjectField) *testing.ObjectFieldT {
	res := &testing.ObjectFieldT{
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
		for j, val := range v.Array {
			res.Array[j] = val
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

var ConvertObjectRequiredHelperCode = `// marshalObjectFieldToObjectFieldT builds a value of type
// *testing.ObjectFieldT from a value of type *ObjectField.
func marshalObjectFieldToObjectFieldT(v *ObjectField) *testing.ObjectFieldT {
	res := &testing.ObjectFieldT{
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
	if v.Array != nil {
		res.Array = make([]bool, len(v.Array))
		for j, val := range v.Array {
			res.Array[j] = val
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
