// This file verifies deterministic HTTP wire union identity and confirms that
// detached wire expressions do not retain service package ownership.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestCollectHTTPUnionTypesDeterministicAcrossObjectOrder(t *testing.T) {
	sourceFromSignals := makeUnionForOrderTest("source",
		"physical_point",
		"synthetic_series",
	)
	sourceFromInputs := makeUnionForOrderTest("source",
		"time_series",
		"energy_rates",
	)

	forward := &expr.AttributeExpr{
		Type: &expr.Object{
			{
				Name: "alpha",
				Attribute: &expr.AttributeExpr{
					Type: sourceFromSignals,
				},
			},
			{
				Name: "beta",
				Attribute: &expr.AttributeExpr{
					Type: sourceFromInputs,
				},
			},
		},
	}
	reverse := &expr.AttributeExpr{
		Type: &expr.Object{
			{
				Name: "beta",
				Attribute: &expr.AttributeExpr{
					Type: sourceFromInputs,
				},
			},
			{
				Name: "alpha",
				Attribute: &expr.AttributeExpr{
					Type: sourceFromSignals,
				},
			},
		},
	}

	forwardNames := collectHTTPUnionTypeNames(forward)
	reverseNames := collectHTTPUnionTypeNames(reverse)

	require.Len(t, forwardNames, 2)
	require.Equal(t, forwardNames, reverseNames)
}

func TestCollectHTTPUnionTypesReusesSameShapedDeclarationsAndReferences(t *testing.T) {
	first := makeUnionForOrderTest("Value", "bool", "number")
	second := makeUnionForOrderTest("Value", "bool", "number")
	bodies := &expr.AttributeExpr{
		Type: &expr.Object{
			{
				Name:      "first",
				Attribute: &expr.AttributeExpr{Type: first},
			},
			{
				Name:      "second",
				Attribute: &expr.AttributeExpr{Type: second},
			},
		},
	}

	catalog := newWireTypeCatalog()
	catalog.collect(bodies, wireAttribute, wireTypePolicy{}, "")
	catalog.Freeze()
	catalog.applyNames(bodies, wireAttribute, wireTypePolicy{})

	emitted := make([]string, 0, len(catalog.unions))
	for _, union := range catalog.unionTypes() {
		emitted = append(emitted, union.Name)
	}
	references := []string{
		catalog.scope.GoTypeName(&expr.AttributeExpr{Type: first}),
		catalog.scope.GoTypeName(&expr.AttributeExpr{Type: second}),
	}
	require.Equal(t, []string{"Value"}, emitted)
	require.Equal(t, []string{"Value", "Value"}, references)
}

func TestHTTPServiceDataReusesSameShapedMethodBodyUnions(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.Service("values", func() {
			dsl.Method("first", func() {
				dsl.Payload(func() {
					dsl.OneOf("Value", sameShapedValueUnionDSL)
				})
				dsl.HTTP(func() {
					dsl.POST("/first")
				})
			})
			dsl.Method("second", func() {
				dsl.Payload(func() {
					dsl.OneOf("Value", sameShapedValueUnionDSL)
				})
				dsl.HTTP(func() {
					dsl.POST("/second")
				})
			})
		})
	})

	data := CreateHTTPServices(root).Get("values")
	require.NotNil(t, data)
	for _, catalog := range []*wireTypeCatalog{data.serverWireTypes, data.clientWireTypes} {
		unions := catalog.unionTypes()
		emitted := make([]string, len(unions))
		for i, union := range unions {
			emitted[i] = union.Name
		}
		require.Equal(t, []string{"Value", "Value2"}, emitted)
	}
	require.Contains(t, data.Endpoint("first").Payload.Request.ServerBody.Def, "Value *Value ")
	require.Contains(t, data.Endpoint("second").Payload.Request.ServerBody.Def, "Value *Value2 ")
}

func TestMakeHTTPTypeRemovesServicePackageOwnershipFromWireCopy(t *testing.T) {
	nested := &expr.UserTypeExpr{
		TypeName: "Nested",
		AttributeExpr: &expr.AttributeExpr{
			Type: &expr.Object{
				{Name: "choice", Attribute: &expr.AttributeExpr{Type: makeUnionForOrderTest("Choice", "text", "number")}},
			},
			Meta: expr.MetaExpr{"struct:pkg:path": {"service/types"}},
		},
	}
	outer := &expr.UserTypeExpr{
		TypeName: "Envelope",
		AttributeExpr: &expr.AttributeExpr{
			Type: &expr.Object{
				{Name: "nested", Attribute: &expr.AttributeExpr{Type: nested}},
			},
			Meta: expr.MetaExpr{"struct:pkg:path": {"service/types"}},
		},
	}

	wire := makeHTTPType(&expr.AttributeExpr{Type: outer})
	wireOuter := wire.Type.(expr.UserType)
	wireNested := expr.AsObject(wireOuter.Attribute().Type).Attribute("nested").Type.(expr.UserType)

	require.NotContains(t, wireOuter.Attribute().Meta, "struct:pkg:path")
	require.NotContains(t, wireNested.Attribute().Meta, "struct:pkg:path")
	require.Contains(t, outer.Attribute().Meta, "struct:pkg:path")
	require.Contains(t, nested.Attribute().Meta, "struct:pkg:path")
}

func TestStreamingHTTPTypeRemovesServicePackageOwnershipFromWireCopy(t *testing.T) {
	nested := &expr.UserTypeExpr{
		TypeName: "Nested",
		AttributeExpr: &expr.AttributeExpr{
			Type: &expr.Object{
				{Name: "value", Attribute: &expr.AttributeExpr{Type: expr.String}},
			},
			Meta: expr.MetaExpr{"struct:pkg:path": {"service/types"}},
		},
	}
	outer := &expr.UserTypeExpr{
		TypeName: "Envelope",
		AttributeExpr: &expr.AttributeExpr{
			Type: &expr.Object{
				{Name: "nested", Attribute: &expr.AttributeExpr{Type: nested}},
			},
			Meta: expr.MetaExpr{"struct:pkg:path": {"service/types"}},
		},
	}
	body := &expr.AttributeExpr{Type: outer}
	endpoint := &expr.HTTPEndpointExpr{StreamingBody: body}

	wire := new(shapedBodies).streaming(endpoint)
	wireOuter := wire.Type.(expr.UserType)
	wireNested := expr.AsObject(wireOuter.Attribute().Type).Attribute("nested").Type.(expr.UserType)

	require.NotContains(t, wireOuter.Attribute().Meta, "struct:pkg:path")
	require.NotContains(t, wireNested.Attribute().Meta, "struct:pkg:path")
	require.Contains(t, outer.Attribute().Meta, "struct:pkg:path")
	require.Contains(t, nested.Attribute().Meta, "struct:pkg:path")
}

func sameShapedValueUnionDSL() {
	dsl.Attribute("bool", dsl.Boolean)
	dsl.Attribute("number", dsl.Float64)
}

func collectHTTPUnionTypeNames(att *expr.AttributeExpr) map[string]string {
	catalog := newWireTypeCatalog()
	catalog.collect(att, wireAttribute, wireTypePolicy{}, "")
	catalog.Freeze()

	names := make(map[string]string, len(catalog.unions))
	for _, record := range catalog.unions {
		names[record.identity.definition.Hash()] = record.data.Name
	}
	return names
}

func makeUnionForOrderTest(typeName string, variants ...string) *expr.Union {
	values := make([]*expr.NamedAttributeExpr, len(variants))
	for i, variant := range variants {
		values[i] = &expr.NamedAttributeExpr{
			Name: variant,
			Attribute: &expr.AttributeExpr{
				Type: expr.String,
			},
		}
	}
	return &expr.Union{TypeName: typeName, Values: values}
}
