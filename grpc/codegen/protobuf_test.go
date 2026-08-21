// This file verifies protobuf wire shaping, naming, JSON options, wrappers,
// and recursion follow generated-package declaration ownership.
package codegen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
		"Any", expr.Any, "google.protobuf.Value",
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
		"Any", expr.Any, "*structpb.Value",
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

func TestMakeProtoBufMessageMarksWrappers(t *testing.T) {
	cases := []struct {
		Name string
		Type func() expr.DataType
		// ElemWrapped indicates that the wrapped field is an array whose
		// element is itself a synthesized wrapper message.
		ElemWrapped bool
		// FieldKind is the expected kind of the attribute wrapped by the
		// top-level message.
		FieldKind expr.Kind
	}{{
		Name:      "primitive",
		Type:      func() expr.DataType { return expr.Int },
		FieldKind: expr.IntKind,
	}, {
		Name: "array",
		Type: func() expr.DataType {
			return &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.Int}}
		},
		FieldKind: expr.ArrayKind,
	}, {
		Name: "map",
		Type: func() expr.DataType {
			return &expr.Map{
				KeyType:  &expr.AttributeExpr{Type: expr.String},
				ElemType: &expr.AttributeExpr{Type: expr.Int},
			}
		},
		FieldKind: expr.MapKind,
	}, {
		Name: "array-of-array",
		Type: func() expr.DataType {
			return &expr.Array{ElemType: &expr.AttributeExpr{
				Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: expr.Int}},
			}}
		},
		ElemWrapped: true,
		FieldKind:   expr.ArrayKind,
	}}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			att := makeProtoBufMessage(
				&expr.AttributeExpr{Type: c.Type()},
				"Message",
				testGRPCMessageExampleIdentity(c.Name),
			)
			require.True(t, isWrappedAttr(att), "expected message to be marked as a wrapper")
			field := unwrapAttr(att)
			assert.Equal(t, c.FieldKind, field.Type.Kind(), "unexpected wrapped field kind")
			if c.ElemWrapped {
				elem := expr.AsArray(field.Type).ElemType
				require.True(t, isWrappedAttr(elem), "expected nested array element to be marked as a wrapper")
				inner := unwrapAttr(elem)
				assert.Equal(t, expr.ArrayKind, inner.Type.Kind(), "unexpected nested wrapped field kind")
			}
		})
	}
}

func TestMakeProtoBufMessageDistinguishesEqualUIDOrigins(t *testing.T) {
	first := protobufArrayTraversalType("First", "shared")
	second := protobufArrayTraversalType("Second", "shared")
	body := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "first", Attribute: &expr.AttributeExpr{Type: first}},
		{Name: "second", Attribute: &expr.AttributeExpr{Type: second}},
	}}

	message := makeProtoBufMessage(body, "Request", testGRPCMessageExampleIdentity("equal-UID-origins"))
	object := expr.AsObject(message.Type.(expr.UserType).Attribute().Type)
	wireFirst := object.Attribute("first").Type.(expr.UserType)
	wireSecond := object.Attribute("second").Type.(expr.UserType)
	require.True(t, isWrappedAttr(&expr.AttributeExpr{Type: wireFirst}))
	require.True(t, isWrappedAttr(&expr.AttributeExpr{Type: wireSecond}))
}

func TestMakeProtoBufMessageDistinguishesNormalizedMethodNames(t *testing.T) {
	service := &expr.ServiceExpr{Name: "Values"}
	dashedMethod := &expr.MethodExpr{Name: "foo-bar", Service: service}
	underscoreMethod := &expr.MethodExpr{Name: "foo_bar", Service: service}
	dashedOwner := expr.GRPCRequestMessageExampleIdentity(dashedMethod)
	underscoreOwner := expr.GRPCRequestMessageExampleIdentity(underscoreMethod)
	dashed := makeProtoBufMessage(
		&expr.AttributeExpr{Type: &expr.Object{
			{Name: "dashed", Attribute: &expr.AttributeExpr{Type: expr.String}},
		}},
		"FooBarRequest",
		dashedOwner,
	)
	underscore := makeProtoBufMessage(
		&expr.AttributeExpr{Type: &expr.Object{
			{Name: "underscore", Attribute: &expr.AttributeExpr{Type: expr.String}},
		}},
		"FooBarRequest",
		underscoreOwner,
	)
	require.NotEqual(t, dashed.Type.(expr.UserType).ID(), underscore.Type.(expr.UserType).ID())

	cases := []struct {
		name        string
		first       *expr.AttributeExpr
		firstOwner  expr.ExampleIdentity
		firstField  string
		second      *expr.AttributeExpr
		secondOwner expr.ExampleIdentity
		secondField string
	}{
		{
			name:        "dashed then underscore",
			first:       dashed,
			firstOwner:  dashedOwner,
			firstField:  "dashed",
			second:      underscore,
			secondOwner: underscoreOwner,
			secondField: "underscore",
		},
		{
			name:        "underscore then dashed",
			first:       underscore,
			firstOwner:  underscoreOwner,
			firstField:  "underscore",
			second:      dashed,
			secondOwner: dashedOwner,
			secondField: "dashed",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			generator := expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("test"))
			first := test.first.Example(generator.At(test.firstOwner)).(map[string]any)
			second := test.second.Example(generator.At(test.secondOwner)).(map[string]any)

			require.Contains(t, first, test.firstField)
			require.NotContains(t, first, test.secondField)
			require.Contains(t, second, test.secondField)
			require.NotContains(t, second, test.firstField)
		})
	}
}

func TestMakeProtoBufMessageSharesAuthoredCollectionWrapperIdentity(t *testing.T) {
	arrayAlias := &expr.UserTypeExpr{
		TypeName: "Strings",
		UID:      "strings",
		AttributeExpr: &expr.AttributeExpr{Type: &expr.Array{
			ElemType: &expr.AttributeExpr{Type: expr.String},
		}},
	}
	mapAlias := &expr.UserTypeExpr{
		TypeName: "Labels",
		UID:      "labels",
		AttributeExpr: &expr.AttributeExpr{Type: &expr.Map{
			KeyType:  &expr.AttributeExpr{Type: expr.String},
			ElemType: &expr.AttributeExpr{Type: expr.Int},
		}},
	}
	owner := testGRPCMessageExampleIdentity("shared-collection-aliases")
	build := func(fields []string) *expr.AttributeExpr {
		attributes := make(expr.Object, len(fields))
		for index, name := range fields {
			typ := expr.UserType(arrayAlias)
			if name == "map_a" || name == "map_b" {
				typ = mapAlias
			}
			attributes[index] = &expr.NamedAttributeExpr{
				Name:      name,
				Attribute: &expr.AttributeExpr{Type: typ},
			}
		}
		return makeProtoBufMessage(
			&expr.AttributeExpr{Type: &attributes},
			"SharedCollectionsRequest",
			owner,
		)
	}
	forward := build([]string{"array_a", "map_a", "array_b", "map_b"})
	reverse := build([]string{"map_b", "array_b", "map_a", "array_a"})
	example := func(message *expr.AttributeExpr) map[string]any {
		generator := expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("test"))
		return message.Example(generator.At(owner)).(map[string]any)
	}

	forwardExample := example(forward)
	reverseExample := example(reverse)
	require.Equal(t, forwardExample, reverseExample)
	require.Equal(t, forwardExample["array_a"], forwardExample["array_b"])
	require.Equal(t, forwardExample["map_a"], forwardExample["map_b"])
}

// protobufArrayTraversalType builds an authored array declaration that protobuf
// conversion must wrap in a message.
func protobufArrayTraversalType(name, uid string) *expr.UserTypeExpr {
	return &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{Type: &expr.Array{
			ElemType: &expr.AttributeExpr{Type: expr.String},
		}},
		TypeName: name,
		UID:      uid,
	}
}

func TestUnwrapAttrPanicsOnNonWrapper(t *testing.T) {
	cases := []struct {
		Name string
		Att  *expr.AttributeExpr
	}{{
		Name: "primitive",
		Att:  &expr.AttributeExpr{Type: expr.Int},
	}, {
		Name: "object",
		Att: &expr.AttributeExpr{Type: &expr.Object{
			{Name: "id", Attribute: &expr.AttributeExpr{Type: expr.Int}},
		}},
	}, {
		// A user type with an attribute literally named "field" but no
		// marker: only the marker set at wrapping time identifies wrappers.
		Name: "unmarked user type with field attribute",
		Att: &expr.AttributeExpr{Type: &expr.UserTypeExpr{
			TypeName: "NotAWrapper",
			AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{
				{Name: "field", Attribute: &expr.AttributeExpr{Type: expr.Int}},
			}},
		}},
	}}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			assert.False(t, isWrappedAttr(c.Att))
			require.Panics(t, func() { unwrapAttr(c.Att) })
		})
	}
}
