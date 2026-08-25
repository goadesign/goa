package codegen

import (
	"strings"
	"testing"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

func TestProtobufify(t *testing.T) {
	cases := []struct {
		Name       string
		String     string
		FirstUpper bool
		Acronym    bool
		Expected   string
	}{{
		"AllLower", "lower", false, false, "lower",
	}, {
		"AllLowerFirstUpper", "lower", true, false, "Lower",
	}, {
		"AllUpper", "UPPER", false, false, "uPPER",
	}, {
		"AllUpperFirstUpper", "UPPER", true, false, "UPPER",
	}, {
		"StartUpperThenLower", "Upper", false, false, "upper",
	}, {
		"StartUpperThenLowerFirstUpper", "Upper", true, false, "Upper",
	}, {
		"StartsWithUnderscore", "_foo", false, false, "foo",
	}, {
		"EndsWithUnderscore", "foo_", false, false, "foo",
	}, {
		"ContainsUnderscore", "foo_bar", false, false, "fooBar",
	}, {
		"StartsWithDigits", "123foo", false, false, "123Foo",
	}, {
		"EndsWithDigits", "foo123", false, false, "foo123",
	}, {
		"ContainsDigits", "foo123bar", false, false, "foo123Bar",
	}, {
		"ContainsIgnoredAcronym", "foo_jwt", false, false, "fooJwt",
	}, {
		"ContainsAcronym", "foo_jwt", false, true, "fooJWT",
	}}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := protoBufify(c.String, c.FirstUpper, c.Acronym)
			if got != c.Expected {
				t.Errorf("got %q, expected %q", got, c.Expected)
			}
		})
	}
}

func TestProtoNativeType(t *testing.T) {
	cases := []struct {
		Name     string
		DataType expr.DataType
		Expected string
	}{{
		"Boolean", expr.Boolean, "bool",
	}, {
		"Int", expr.Int, "sint32",
	}, {
		"Int32", expr.Int32, "sint32",
	}, {
		"Int64", expr.Int64, "sint64",
	}, {
		"UInt", expr.UInt, "uint32",
	}, {
		"UInt32", expr.UInt32, "uint32",
	}, {
		"UInt64", expr.UInt64, "uint64",
	}, {
		"Float32", expr.Float32, "float",
	}, {
		"Float64", expr.Float64, "double",
	}, {
		"String", expr.String, "string",
	}, {
		"Bytes", expr.Bytes, "bytes",
	}, {
		"Any", expr.Any, "google.protobuf.Any",
	}}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := protoNativeType(c.DataType)
			if got != c.Expected {
				t.Errorf("got %q, expected %q", got, c.Expected)
			}
		})
	}
}

func TestProtoBufNativeGoTypeName(t *testing.T) {
	cases := []struct {
		Name     string
		DataType expr.DataType
		Expected string
	}{{
		"Boolean", expr.Boolean, "bool",
	}, {
		"Int", expr.Int, "int32",
	}, {
		"Int32", expr.Int32, "int32",
	}, {
		"Int64", expr.Int64, "int64",
	}, {
		"UInt", expr.UInt, "uint32",
	}, {
		"UInt32", expr.UInt32, "uint32",
	}, {
		"UInt64", expr.UInt64, "uint64",
	}, {
		"Float32", expr.Float32, "float32",
	}, {
		"Float64", expr.Float64, "float64",
	}, {
		"String", expr.String, "string",
	}, {
		"Bytes", expr.Bytes, "[]byte",
	}, {
		"Any", expr.Any, "*anypb.Any",
	}}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := protoBufNativeGoTypeName(c.DataType)
			if got != c.Expected {
				t.Errorf("got %q, expected %q", got, c.Expected)
			}
		})
	}
}

func TestHasAnyType(t *testing.T) {
	cases := []struct {
		Name     string
		Attr     *expr.AttributeExpr
		Expected bool
	}{{
		"NoAnyType", &expr.AttributeExpr{Type: expr.String}, false,
	}, {
		"DirectAnyType", &expr.AttributeExpr{Type: expr.Any}, true,
	}, {
		"ArrayWithAnyType", &expr.AttributeExpr{
			Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.Any}},
		}, true,
	}, {
		"MapWithAnyKeyType", &expr.AttributeExpr{
			Type: &expr.Map{
				KeyType:  &expr.AttributeExpr{Type: expr.Any},
				ElemType: &expr.AttributeExpr{Type: expr.String},
			},
		}, true,
	}, {
		"MapWithAnyElemType", &expr.AttributeExpr{
			Type: &expr.Map{
				KeyType:  &expr.AttributeExpr{Type: expr.String},
				ElemType: &expr.AttributeExpr{Type: expr.Any},
			},
		}, true,
	}, {
		"ObjectWithAnyField", &expr.AttributeExpr{
			Type: &expr.Object{
				&expr.NamedAttributeExpr{
					Name:      "field",
					Attribute: &expr.AttributeExpr{Type: expr.Any},
				},
			},
		}, true,
	}}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := hasAnyType(c.Attr)
			if got != c.Expected {
				t.Errorf("got %t, expected %t", got, c.Expected)
			}
		})
	}
}

func TestProtoBufMessageDefJSONNameOption(t *testing.T) {
	attr := &expr.AttributeExpr{
		Type: &expr.Object{
			&expr.NamedAttributeExpr{
				Name: "value",
				Attribute: &expr.AttributeExpr{
					Type: expr.String,
					Meta: expr.MetaExpr{
						"rpc:tag":        []string{"1"},
						"proto:tag:json": []string{"customValue"},
					},
				},
			},
		},
	}
	sd := &ServiceData{Scope: codegen.NewNameScope()}
	def := protoBufMessageDef(attr, sd)
	if !strings.Contains(def, `json_name = "customValue"`) {
		t.Fatalf("expected json_name option, got %q", def)
	}
}

func TestProtoBufMessageDefJSONNameOptionOneOf(t *testing.T) {
	attr := &expr.AttributeExpr{
		Type: &expr.Union{
			TypeName: "Animal",
			Values: []*expr.NamedAttributeExpr{
				{
					Name: "Cat",
					Attribute: &expr.AttributeExpr{
						Type: expr.String,
						Meta: expr.MetaExpr{
							"rpc:tag":        []string{"1"},
							"proto:tag:json": []string{"cat"},
						},
					},
				},
			},
		},
	}
	sd := &ServiceData{Scope: codegen.NewNameScope()}
	def := protoBufMessageDef(attr, sd)
	if !strings.Contains(def, `json_name = "cat"`) {
		t.Fatalf("expected json_name option in oneof, got %q", def)
	}
}

func TestProtoBufMessageDefReservations(t *testing.T) {
	attr := &expr.AttributeExpr{
		Type: &expr.Object{
			&expr.NamedAttributeExpr{
				Name: "id",
				Attribute: &expr.AttributeExpr{
					Type: expr.String,
					Meta: expr.MetaExpr{"rpc:tag": []string{"1"}},
				},
			},
		},
		Meta: expr.MetaExpr{
			"rpc:reserved:number": []string{"20", "3", "15"},
			"rpc:reserved:name":   []string{"linked_control_point_id", "deployment_id"},
		},
	}
	sd := &ServiceData{Scope: codegen.NewNameScope()}

	def := protoBufMessageDef(attr, sd)

	if !strings.Contains(def, "reserved 3, 15, 20;") {
		t.Fatalf("expected sorted reserved numbers, got %q", def)
	}
	if !strings.Contains(def, `reserved "deployment_id", "linked_control_point_id";`) {
		t.Fatalf("expected sorted reserved names, got %q", def)
	}
}
