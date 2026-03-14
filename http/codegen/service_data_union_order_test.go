package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	cg "goa.design/goa/v3/codegen"
	svc "goa.design/goa/v3/codegen/service"
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

func TestUniqueHTTPUnionFieldNamesReservePreexistingNormalizedNames(t *testing.T) {
	names := uniqueHTTPUnionFieldNames([]*expr.NamedAttributeExpr{
		{Name: "Foo", Attribute: &expr.AttributeExpr{Type: expr.String}},
		{Name: "Foo!", Attribute: &expr.AttributeExpr{Type: expr.String}},
		{Name: "Foo2", Attribute: &expr.AttributeExpr{Type: expr.String}},
	})

	require.ElementsMatch(t, []string{"Foo", "Foo2", "Foo3"}, names)
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
	return &expr.Union{
		TypeName: typeName,
		Values:   values,
	}
}
