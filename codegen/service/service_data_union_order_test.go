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
	service := &expr.ServiceExpr{Name: "test"}
	generation := codegen.NewGeneration("generated.local/gen", nil)
	generatedPackage := generation.GeneratedPackage(
		generatedPackagePath(generation.GenPkg, service, loc),
	)
	object := att.Type.(*expr.Object)
	for _, named := range *object {
		_, err := generatedPackage.DeclareUnion(named.Attribute.Type.(*expr.Union))
		if err != nil {
			panic(err)
		}
	}
	if err := generation.Freeze(); err != nil {
		panic(err)
	}
	services := &ServicesData{
		generation: generation,
		packages:   make(map[string]*generatedPackageData),
	}
	seen := make(map[string]struct{})
	unionByHash := make(map[unionDataKey]*UnionTypeData)
	if err := services.collectUnionTypes(att, service, codegen.NewNameScope(), loc, unionByHash, seen, false); err != nil {
		panic(err)
	}

	names := make(map[string]string, len(unionByHash))
	for key, data := range unionByHash {
		names[string(key.identity)] = data.Name
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
