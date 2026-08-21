package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

func TestCollectUnionTypesDeterministicAcrossObjectOrder(t *testing.T) {
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

	loc := &codegen.Location{
		RelImportPath: "gen/service",
	}
	forwardNames := collectServiceUnionTypeNames(forward, loc)
	reverseNames := collectServiceUnionTypeNames(reverse, loc)

	require.Len(t, forwardNames, 2)
	require.Equal(t, forwardNames, reverseNames)
}

func collectServiceUnionTypeNames(att *expr.AttributeExpr, loc *codegen.Location) map[string]string {
	scope := codegen.NewNameScope()
	scopes := &serviceNameScopes{
		local: scope,
		packages: &packageScopes{
			scopes: make(map[string]*codegen.NameScope),
		},
	}
	seen := make(map[string]struct{})
	unionByHash := make(map[string]*UnionTypeData)
	collectUnionTypes(att, scopes, loc, unionByHash, seen, false)

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
