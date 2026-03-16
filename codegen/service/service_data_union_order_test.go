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

func TestBuildUnionTypeDataKindConstsUseUniqueFieldNames(t *testing.T) {
	union := makeUnionForOrderTest("normalized",
		"foo_bar",
		"foo-bar",
	)
	loc := &codegen.Location{RelImportPath: "gen/service"}

	data := buildUnionTypeData(union, codegen.NewNameScope(), loc)
	require.Len(t, data.Fields, 2)
	require.ElementsMatch(t,
		[]string{"FooBar", "FooBar2"},
		[]string{data.Fields[0].FieldName, data.Fields[1].FieldName},
	)
	require.ElementsMatch(t,
		[]string{"NormalizedKindFooBar", "NormalizedKindFooBar2"},
		[]string{data.Fields[0].KindConst, data.Fields[1].KindConst},
	)
	require.ElementsMatch(t,
		[]string{"foo_bar", "foo-bar"},
		[]string{data.Fields[0].TypeTag, data.Fields[1].TypeTag},
	)
}

func TestBuildViewUnionTypeDataKindConstsUseUniqueFieldNames(t *testing.T) {
	union := makeUnionForOrderTest("normalized",
		"foo_bar",
		"foo-bar",
	)
	loc := &codegen.Location{RelImportPath: "gen/views"}

	data := buildViewUnionTypeData(union, codegen.NewNameScope(), loc)
	require.Len(t, data.Fields, 2)
	require.ElementsMatch(t,
		[]string{"FooBar", "FooBar2"},
		[]string{data.Fields[0].FieldName, data.Fields[1].FieldName},
	)
	require.ElementsMatch(t,
		[]string{"NormalizedKindFooBar", "NormalizedKindFooBar2"},
		[]string{data.Fields[0].KindConst, data.Fields[1].KindConst},
	)
	require.ElementsMatch(t,
		[]string{"foo_bar", "foo-bar"},
		[]string{data.Fields[0].TypeTag, data.Fields[1].TypeTag},
	)
}

func TestUniqueUnionFieldNamesReservePreexistingNormalizedNames(t *testing.T) {
	names := codegen.UniqueUnionFieldNames([]*expr.NamedAttributeExpr{
		{Name: "Foo", Attribute: &expr.AttributeExpr{Type: expr.String}},
		{Name: "Foo!", Attribute: &expr.AttributeExpr{Type: expr.String}},
		{Name: "Foo2", Attribute: &expr.AttributeExpr{Type: expr.String}},
	})

	require.ElementsMatch(t, []string{"Foo", "Foo2", "Foo3"}, names)
}

func collectServiceUnionTypeNames(att *expr.AttributeExpr, loc *codegen.Location) map[string]string {
	scope := codegen.NewNameScope()
	seen := make(map[string]struct{})
	unionByHash := make(map[string]*UnionTypeData)
	collectUnionTypes(att, scope, loc, unionByHash, seen)

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
