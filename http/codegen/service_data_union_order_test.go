package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	cg "goa.design/goa/v3/codegen"
	svc "goa.design/goa/v3/codegen/service"
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

	scope := cg.NewNameScope()
	unions := make(map[string]*svc.UnionTypeData)
	collectHTTPUnionTypes(bodies, scope, unions, make(map[string]struct{}))

	emitted := make([]string, 0, len(unions))
	for _, union := range unions {
		emitted = append(emitted, union.Name)
	}
	references := []string{
		scope.GoTypeName(&expr.AttributeExpr{Type: first}),
		scope.GoTypeName(&expr.AttributeExpr{Type: second}),
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
	emitted := make([]string, len(data.UnionTypes))
	for i, union := range data.UnionTypes {
		emitted[i] = union.Name
	}
	require.Equal(t, []string{"Value"}, emitted)
	require.Contains(t, data.Endpoint("first").Payload.Request.ServerBody.Def, "Value *Value ")
	require.Contains(t, data.Endpoint("second").Payload.Request.ServerBody.Def, "Value *Value ")
}

func sameShapedValueUnionDSL() {
	dsl.Attribute("bool", dsl.Boolean)
	dsl.Attribute("number", dsl.Float64)
}

func collectHTTPUnionTypeNames(att *expr.AttributeExpr) map[string]string {
	scope := cg.NewNameScope()
	seen := make(map[string]struct{})
	unionByHash := make(map[string]*svc.UnionTypeData)
	collectHTTPUnionTypes(att, scope, unionByHash, seen)

	names := make(map[string]string, len(unionByHash))
	for hash, data := range unionByHash {
		names[hash] = data.Name
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
