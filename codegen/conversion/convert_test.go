package conversion

import (
	"testing"

	"goa.design/goa/design"
	"goa.design/goa/dsl"
)

func TestCompatible(t *testing.T) {
	var obj = &design.UserTypeExpr{
		AttributeExpr: &design.AttributeExpr{
			Type: &design.Object{
				{"foo", &design.AttributeExpr{Type: design.String}},
				{"bar", &design.AttributeExpr{Type: design.Int}},
				{"baz", &design.AttributeExpr{Type: design.Boolean, Metadata: design.MetadataExpr{"struct.field.external": []string{"Mapped"}}}},
			},
		},
		TypeName: "TestType",
	}
	type objT struct {
		Foo    string
		Bar    int
		Mapped bool
	}
	type objT2 struct {
		Foo    string
		Bar    string
		Mapped bool
	}
	type objT3 struct {
		Foo string
		Bar int
	}
	cases := []struct {
		Name        string
		From        design.DataType
		To          interface{}
		ExpectedErr string
	}{
		{"bool", design.Boolean, false, ""},
		{"int", design.Int, 0, ""},
		{"int32", design.Int32, int32(0), ""},
		{"int64", design.Int64, int64(0), ""},
		{"uint", design.UInt, uint(0), ""},
		{"uint32", design.UInt32, uint32(0), ""},
		{"uint64", design.UInt64, uint64(0), ""},
		{"float32", design.Float32, float32(0.0), ""},
		{"float64", design.Float64, 0.0, ""},
		{"string", design.String, "", ""},
		{"bytes", design.Bytes, []byte(""), ""},
		{"array", dsl.ArrayOf(design.String), []string{""}, ""},
		{"map", dsl.MapOf(design.String, design.String), map[string]string{"": ""}, ""},
		{"object", obj, objT{}, ""},
		{"array-object", dsl.ArrayOf(obj), []objT{objT{}}, ""},

		{"invalid-primitive", design.String, 0, "types don't match: type of <value> is int but type of corresponding attribute is string"},
		{"empty-array", dsl.ArrayOf(design.String), []string{}, "slice <value> must contain exactly one item"},
		{"invalid-array", dsl.ArrayOf(design.String), []int{0}, "types don't match: type of <value>[0] is int but type of corresponding attribute is string"},
		{"empty-map", dsl.MapOf(design.String, design.String), map[string]string{}, "map <value> must contain exactly one key"},
		{"invalid-map-key", dsl.MapOf(design.String, design.String), map[int]string{0: ""}, "types don't match: type of <value>.key is int but type of corresponding attribute is string"},
		{"invalid-map-val", dsl.MapOf(design.String, design.String), map[string]int{"": 0}, "types don't match: type of <value>.value is int but type of corresponding attribute is string"},
		{"invalid-obj", obj, "", "types don't match: <value> is not a struct"},
		{"invalid-obj-2", obj, objT2{}, "types don't match: type of <value>.Bar is string but type of corresponding attribute is int"},
		{"invalid-obj-3", obj, objT3{}, "types don't match: could not find field \"Mapped\" of external type conversion.objT3 matching attribute \"baz\" of type \"TestType\""},
		{"invalid-array-object", dsl.ArrayOf(obj), []objT2{objT2{}}, "types don't match: type of <value>[0].Bar is string but type of corresponding attribute is int"},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			err := Compatible(c.From, c.To)
			if err == nil {
				if c.ExpectedErr != "" {
					t.Errorf("got no error, expected %q", c.ExpectedErr)
				}
			} else {
				if c.ExpectedErr == "" {
					t.Errorf("got error %q, expected none", err)
				} else {
					if err.Error() != c.ExpectedErr {
						t.Errorf("got error %q, expected %q", err, c.ExpectedErr)
					}
				}
			}
		})
	}
}
