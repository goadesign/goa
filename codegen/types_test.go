package codegen

import (
	"testing"

	"goa.design/goa.v2/design"
)

func TestGoTypeDef(t *testing.T) {
	cases := map[string]struct {
		dataType design.DataType
		expected string
	}{
		"BooleanKind": {design.Boolean, "bool"},
		"IntKind":     {design.Int, "int"},
		"Int32Kind":   {design.Int32, "int32"},
		"Int64Kind":   {design.Int64, "int64"},
		"UIntKind":    {design.UInt, "uint"},
		"UInt32Kind":  {design.UInt32, "uint32"},
		"UInt64Kind":  {design.UInt64, "uint64"},
		"Float32Kind": {design.Float32, "float32"},
		"Float64Kind": {design.Float64, "float64"},
		"StringKind":  {design.String, "string"},
		"BytesKind":   {design.Bytes, "[]byte"},
		"AnyKind":     {design.Any, "interface{}"},

		"Array":         {&design.Array{&design.AttributeExpr{Type: design.Boolean}}, "[]bool"},
		"Map":           {&design.Map{KeyType: &design.AttributeExpr{Type: design.Int}, ElemType: &design.AttributeExpr{Type: design.String}}, "map[int]string"},
		"Object":        {design.Object{"IntField": &design.AttributeExpr{Type: design.Int}, "StringField": &design.AttributeExpr{Type: design.String}}, "struct {\n\tIntField *int\n\tStringField *string\n}"},
		"UserTypeExpr":  {&design.UserTypeExpr{AttributeExpr: &design.AttributeExpr{Type: design.Boolean}, TypeName: "UserType"}, "UserType"},
		"MediaTypeExpr": {&design.MediaTypeExpr{UserTypeExpr: &design.UserTypeExpr{AttributeExpr: &design.AttributeExpr{Type: design.Boolean}, TypeName: "MediaType"}, Identifier: "application/vnd.goa.example", Views: nil}, "MediaType"},
	}

	for k, tc := range cases {
		scope := NewNameScope()
		actual := scope.GoTypeDef(&design.AttributeExpr{Type: tc.dataType})
		if actual != tc.expected {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}

func TestGoNativeTypeName(t *testing.T) {
	cases := map[string]struct {
		dataType design.DataType
		expected string
	}{
		"BooleanKind": {design.Boolean, "bool"},
		"IntKind":     {design.Int, "int"},
		"Int32Kind":   {design.Int32, "int32"},
		"Int64Kind":   {design.Int64, "int64"},
		"UIntKind":    {design.UInt, "uint"},
		"UInt32Kind":  {design.UInt32, "uint32"},
		"UInt64Kind":  {design.UInt64, "uint64"},
		"Float32Kind": {design.Float32, "float32"},
		"Float64Kind": {design.Float64, "float64"},
		"StringKind":  {design.String, "string"},
		"BytesKind":   {design.Bytes, "[]byte"},
		"AnyKind":     {design.Any, "interface{}"},
	}

	for k, tc := range cases {
		actual := GoNativeTypeName(tc.dataType)
		if actual != tc.expected {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}

func TestGoify(t *testing.T) {
	cases := map[string]struct {
		str        string
		firstUpper bool
		expected   string
	}{
		"with first upper false":                                 {"blue_id", false, "blueID"},
		"with first upper false normal identifier all lower":     {"blue", false, "blue"},
		"with first upper false and UUID":                        {"blue_uuid", false, "blueUUID"},
		"with first upper true":                                  {"blue_id", true, "BlueID"},
		"with first upper true and UUID":                         {"blue_uuid", true, "BlueUUID"},
		"with first upper true normal identifier all lower":      {"blue", true, "Blue"},
		"with first upper false normal identifier":               {"Blue", false, "blue"},
		"with first upper true normal identifier":                {"Blue", true, "Blue"},
		"with invalid identifier":                                {"Blue%50", true, "Blue50"},
		"with invalid identifier firstupper false":               {"Blue%50", false, "blue50"},
		"with only UUID and firstupper false":                    {"UUID", false, "uuid"},
		"with consecutives invalid identifiers firstupper false": {"[[fields___type]]", false, "fieldsType"},
		"with consecutives invalid identifiers":                  {"[[fields___type]]", true, "FieldsType"},
		"with all invalid identifiers":                           {"[[", false, ""},
	}

	for k, tc := range cases {
		actual := Goify(tc.str, tc.firstUpper)

		if actual != tc.expected {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}
