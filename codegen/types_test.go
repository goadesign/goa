package codegen

import (
	"testing"

	"goa.design/goa/expr"
)

func TestGoTypeDef(t *testing.T) {
	var (
		simpleArray = &expr.AttributeExpr{
			Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.Boolean}}}
		simpleMap = &expr.AttributeExpr{
			Type: &expr.Map{
				KeyType:  &expr.AttributeExpr{Type: expr.Int},
				ElemType: &expr.AttributeExpr{Type: expr.String},
			}}
		requiredObj = &expr.AttributeExpr{
			Type: &expr.Object{
				{"IntField", &expr.AttributeExpr{Type: expr.Int}},
				{"StringField", &expr.AttributeExpr{Type: expr.String}},
			},
			Validation: &expr.ValidationExpr{Required: []string{"IntField", "StringField"}}}
		defaultObj = &expr.AttributeExpr{
			Type: &expr.Object{
				{"IntField", &expr.AttributeExpr{Type: expr.Int, DefaultValue: 1}},
				{"StringField", &expr.AttributeExpr{Type: expr.String, DefaultValue: "foo"}},
			}}
		ut         = &expr.UserTypeExpr{AttributeExpr: &expr.AttributeExpr{Type: expr.Boolean}, TypeName: "UserType"}
		rt         = &expr.ResultTypeExpr{UserTypeExpr: &expr.UserTypeExpr{AttributeExpr: &expr.AttributeExpr{Type: expr.Boolean}, TypeName: "ResultType"}, Identifier: "application/vnd.goa.example", Views: nil}
		userType   = &expr.AttributeExpr{Type: ut}
		resultType = &expr.AttributeExpr{Type: rt}
		mixedObj   = &expr.AttributeExpr{
			Type: &expr.Object{
				{"IntField", &expr.AttributeExpr{Type: expr.Int}},
				{"ArrayField", simpleArray},
				{"MapField", simpleMap},
				{"UserTypeField", userType},
			},
			Validation: &expr.ValidationExpr{Required: []string{"IntField", "ArrayField", "MapField", "UserTypeField"}}}
	)
	cases := map[string]struct {
		att        *expr.AttributeExpr
		pointer    bool
		usedefault bool
		expected   string
	}{
		"BooleanKind": {&expr.AttributeExpr{Type: expr.Boolean}, false, true, "bool"},
		"IntKind":     {&expr.AttributeExpr{Type: expr.Int}, false, true, "int"},
		"Int32Kind":   {&expr.AttributeExpr{Type: expr.Int32}, false, true, "int32"},
		"Int64Kind":   {&expr.AttributeExpr{Type: expr.Int64}, false, true, "int64"},
		"UIntKind":    {&expr.AttributeExpr{Type: expr.UInt}, false, true, "uint"},
		"UInt32Kind":  {&expr.AttributeExpr{Type: expr.UInt32}, false, true, "uint32"},
		"UInt64Kind":  {&expr.AttributeExpr{Type: expr.UInt64}, false, true, "uint64"},
		"Float32Kind": {&expr.AttributeExpr{Type: expr.Float32}, false, true, "float32"},
		"Float64Kind": {&expr.AttributeExpr{Type: expr.Float64}, false, true, "float64"},
		"StringKind":  {&expr.AttributeExpr{Type: expr.String}, false, true, "string"},
		"BytesKind":   {&expr.AttributeExpr{Type: expr.Bytes}, false, true, "[]byte"},
		"AnyKind":     {&expr.AttributeExpr{Type: expr.Any}, false, true, "interface{}"},

		"Array":          {simpleArray, false, true, "[]bool"},
		"Map":            {simpleMap, false, true, "map[int]string"},
		"UserTypeExpr":   {userType, false, true, "UserType"},
		"ResultTypeExpr": {resultType, false, true, "ResultType"},

		"Object":          {requiredObj, false, true, "struct {\n\tIntField int\n\tStringField string\n}"},
		"ObjDefault":      {defaultObj, false, true, "struct {\n\tIntField int\n\tStringField string\n}"},
		"ObjDefaultNoDef": {defaultObj, false, false, "struct {\n\tIntField *int\n\tStringField *string\n}"},
		"ObjMixed":        {mixedObj, false, true, "struct {\n\tIntField int\n\tArrayField []bool\n\tMapField map[int]string\n\tUserTypeField UserType\n}"},
		"ObjMixedPointer": {mixedObj, true, true, "struct {\n\tIntField *int\n\tArrayField []bool\n\tMapField map[int]string\n\tUserTypeField *UserType\n}"},
	}

	for k, tc := range cases {
		scope := NewNameScope()
		actual := scope.GoTypeDef(tc.att, tc.pointer, tc.usedefault)
		if actual != tc.expected {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}

func TestGoNativeTypeName(t *testing.T) {
	cases := map[string]struct {
		dataType expr.DataType
		expected string
	}{
		"BooleanKind": {expr.Boolean, "bool"},
		"IntKind":     {expr.Int, "int"},
		"Int32Kind":   {expr.Int32, "int32"},
		"Int64Kind":   {expr.Int64, "int64"},
		"UIntKind":    {expr.UInt, "uint"},
		"UInt32Kind":  {expr.UInt32, "uint32"},
		"UInt64Kind":  {expr.UInt64, "uint64"},
		"Float32Kind": {expr.Float32, "float32"},
		"Float64Kind": {expr.Float64, "float64"},
		"StringKind":  {expr.String, "string"},
		"BytesKind":   {expr.Bytes, "[]byte"},
		"AnyKind":     {expr.Any, "interface{}"},
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
		"empty":                                             {"", false, ""},
		"first upper false":                                 {"blue_id", false, "blueID"},
		"first upper false normal identifier all lower":     {"blue", false, "blue"},
		"first upper false and UUID":                        {"blue_uuid", false, "blueUUID"},
		"first upper true":                                  {"blue_id", true, "BlueID"},
		"first upper true and UUID":                         {"blue_uuid", true, "BlueUUID"},
		"first upper true normal identifier all lower":      {"blue", true, "Blue"},
		"first upper false normal identifier":               {"Blue", false, "blue"},
		"first upper true normal identifier":                {"Blue", true, "Blue"},
		"invalid identifier":                                {"Blue%50", true, "Blue50"},
		"invalid identifier firstupper false":               {"Blue%50", false, "blue50"},
		"only UUID and firstupper false":                    {"UUID", false, "uuid"},
		"consecutives invalid identifiers firstupper false": {"[[fields___type]]", false, "fieldsType"},
		"consecutives invalid identifiers":                  {"[[fields___type]]", true, "FieldsType"},
		"invalid identifiers":                               {"[[", false, "val"},
		"middle upper firstupper false":                     {"MiddleUpper", false, "middleUpper"},
		"middle upper":                                      {"MiddleUpper", true, "MiddleUpper"},
	}

	for k, tc := range cases {
		actual := Goify(tc.str, tc.firstUpper)

		if actual != tc.expected {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}
