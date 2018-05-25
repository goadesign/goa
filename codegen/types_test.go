package codegen

import (
	"testing"

	"goa.design/goa/design"
)

func TestGoTypeDef(t *testing.T) {
	var (
		simpleArray = array(design.Boolean)
		simpleMap   = mapa(design.Int, design.String)
		requiredObj = require(object("IntField", design.Int, "StringField", design.String), "IntField", "StringField")
		defaultObj  = defaulta(object("IntField", design.Int, "StringField", design.String), "IntField", 1, "StringField", "foo")
		ut          = &design.UserTypeExpr{AttributeExpr: &design.AttributeExpr{Type: design.Boolean}, TypeName: "UserType"}
		rt          = &design.ResultTypeExpr{UserTypeExpr: &design.UserTypeExpr{AttributeExpr: &design.AttributeExpr{Type: design.Boolean}, TypeName: "ResultType"}, Identifier: "application/vnd.goa.example", Views: nil}
		userType    = &design.AttributeExpr{Type: ut}
		resultType  = &design.AttributeExpr{Type: rt}
		mixedObj    = require(object("IntField", design.Int, "ArrayField", simpleArray.Type, "MapField", simpleMap.Type, "UserTypeField", ut), "IntField", "ArrayField", "MapField", "UserTypeField")
	)
	cases := map[string]struct {
		att        *design.AttributeExpr
		usedefault bool
		expected   string
	}{
		"BooleanKind": {&design.AttributeExpr{Type: design.Boolean}, true, "bool"},
		"IntKind":     {&design.AttributeExpr{Type: design.Int}, true, "int"},
		"Int32Kind":   {&design.AttributeExpr{Type: design.Int32}, true, "int32"},
		"Int64Kind":   {&design.AttributeExpr{Type: design.Int64}, true, "int64"},
		"UIntKind":    {&design.AttributeExpr{Type: design.UInt}, true, "uint"},
		"UInt32Kind":  {&design.AttributeExpr{Type: design.UInt32}, true, "uint32"},
		"UInt64Kind":  {&design.AttributeExpr{Type: design.UInt64}, true, "uint64"},
		"Float32Kind": {&design.AttributeExpr{Type: design.Float32}, true, "float32"},
		"Float64Kind": {&design.AttributeExpr{Type: design.Float64}, true, "float64"},
		"StringKind":  {&design.AttributeExpr{Type: design.String}, true, "string"},
		"BytesKind":   {&design.AttributeExpr{Type: design.Bytes}, true, "[]byte"},
		"AnyKind":     {&design.AttributeExpr{Type: design.Any}, true, "interface{}"},

		"Array":          {simpleArray, true, "[]bool"},
		"Map":            {simpleMap, true, "map[int]string"},
		"UserTypeExpr":   {userType, true, "UserType"},
		"ResultTypeExpr": {resultType, true, "ResultType"},

		"Object":          {requiredObj, true, "struct {\n\tIntField int\n\tStringField string\n}"},
		"ObjDefault":      {defaultObj, true, "struct {\n\tIntField int\n\tStringField string\n}"},
		"ObjDefaultNoDef": {defaultObj, false, "struct {\n\tIntField *int\n\tStringField *string\n}"},
		"ObjMixed":        {mixedObj, true, "struct {\n\tIntField int\n\tArrayField []bool\n\tMapField map[int]string\n\tUserTypeField UserType\n}"},
	}

	for k, tc := range cases {
		scope := NewNameScope()
		actual := scope.GoTypeDef(tc.att, tc.usedefault)
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
