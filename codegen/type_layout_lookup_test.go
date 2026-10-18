// These tests keep repeated type lookups scoped to the resolver's retained graph.
package codegen

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func TestTypeLayoutLookupPreservesOccurrencesAndSubtrees(t *testing.T) {
	shared := &expr.AttributeExpr{Type: expr.String}
	child := &expr.AttributeExpr{Type: &expr.Object{{Name: "inside", Attribute: shared}}}
	root := &expr.AttributeExpr{Type: &expr.Object{
		{Name: "outside", Attribute: shared},
		{Name: "child", Attribute: child},
	}}
	plan, err := PlanGoType(root, GoTypePlanOptions{Owner: "example.test/service"})
	require.NoError(t, err)
	resolver := &goTypeLayoutAttributor{
		Attributor: NewAttributeScope(NewNameScope()),
		walk:       &goTypeLayoutWalk{visited: make(map[*GoTypePlan]struct{})},
		layout:     plan.Link("example.test/service", nil),
	}
	for range 2 {
		matches := resolver.findAll(shared)
		require.Len(t, matches, 2)
		require.Equal(t, "Outside", matches[0].Field(true))
		require.Equal(t, "Inside", matches[1].Field(true))
	}
	entered := resolver.Enter(child).(*goTypeLayoutAttributor)
	for range 2 {
		matches := entered.findAll(shared)
		require.Len(t, matches, 1)
		require.Equal(t, "Inside", matches[0].Field(true))
		require.Empty(t, entered.findAll(root))
	}
}

func TestTypeLayoutLookupPreservesRecursiveOrder(t *testing.T) {
	attribute := &expr.AttributeExpr{Type: expr.String}
	first := &GoTypePlan{occurrence: attribute, kind: GoPrimitive, primitive: "string"}
	second := &GoTypePlan{occurrence: attribute, kind: GoPrimitive, primitive: "string"}
	root := &GoTypePlan{fields: []*GoTypePlan{first, second}}
	first.value = root
	resolver := &goTypeLayoutAttributor{
		walk:   &goTypeLayoutWalk{visited: make(map[*GoTypePlan]struct{})},
		layout: root.Link("example.test/service", nil),
	}
	for range 2 {
		matches := resolver.findAll(attribute)
		require.Len(t, matches, 2)
		require.Same(t, first, matches[0].plan)
		require.Same(t, second, matches[1].plan)
	}
}

func TestTypeLayoutLookupStillRejectsAmbiguousReferences(t *testing.T) {
	attribute := &expr.AttributeExpr{Type: expr.String}
	root := &GoTypePlan{fields: []*GoTypePlan{
		{occurrence: attribute, kind: GoPrimitive, primitive: "string"},
		{occurrence: attribute, kind: GoPrimitive, primitive: "int"},
	}}
	resolver := &goTypeLayoutAttributor{
		walk:   &goTypeLayoutWalk{visited: make(map[*GoTypePlan]struct{})},
		layout: root.Link("example.test/service", nil),
	}
	for range 2 {
		require.PanicsWithValue(t, `Go type layout has different references for "string"`, func() {
			resolver.find(attribute)
		})
	}
}

func TestTypeLayoutLookupConcurrentReads(t *testing.T) {
	attribute := &expr.AttributeExpr{Type: expr.String}
	plan, err := PlanGoType(attribute, GoTypePlanOptions{Owner: "example.test/service"})
	require.NoError(t, err)
	resolver := &goTypeLayoutAttributor{
		walk:   &goTypeLayoutWalk{visited: make(map[*GoTypePlan]struct{})},
		layout: plan.Link("example.test/service", nil),
	}
	var workers sync.WaitGroup
	for range 16 {
		workers.Go(func() {
			for range 32 {
				matches := resolver.findAll(attribute)
				if len(matches) != 1 || matches[0].plan != plan {
					t.Error("concurrent lookup returned a different retained type")
				}
			}
		})
	}
	workers.Wait()
}

func TestTypeLayoutLookupReadsCurrentQualifier(t *testing.T) {
	attribute := &expr.AttributeExpr{Type: expr.String}
	plan := &GoTypePlan{occurrence: attribute, kind: GoNamed, owner: "example.test/types"}
	qualifier := "first"
	resolver := &goTypeLayoutAttributor{
		walk: &goTypeLayoutWalk{visited: make(map[*GoTypePlan]struct{})},
		layout: plan.Link("example.test/service", func(string) string {
			return qualifier
		}),
	}
	require.Equal(t, "first", resolver.findAll(attribute)[0].Package())
	qualifier = "final"
	require.Equal(t, "final", resolver.findAll(attribute)[0].Package())
}
