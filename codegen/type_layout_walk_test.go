// These tests preserve pointer-identity lookups while parent and child resolvers
// reuse temporary traversal storage.
package codegen

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func TestTypeLayoutWalkSharesStorageWithoutSharingScope(t *testing.T) {
	shared := &expr.AttributeExpr{Type: expr.String}
	child := &expr.AttributeExpr{Type: &expr.Object{{Name: "inside", Attribute: shared}}}
	root := &expr.AttributeExpr{Type: &expr.Object{{Name: "outside", Attribute: shared}, {Name: "child", Attribute: child}}}
	plan, err := PlanGoType(root, GoTypePlanOptions{Owner: "example.test/service", Policy: GoLayoutPolicy{SumType: true}})
	require.NoError(t, err)
	context, err := NewAttributeContext(false, false, false, "", NewNameScope()).WithGoTypeLayout(plan.Link("example.test/service", nil))
	require.NoError(t, err)
	parent := context.Scope.(*goTypeLayoutAttributor)
	entered := parent.Enter(child).(*goTypeLayoutAttributor)
	require.Same(t, parent.walk, entered.walk)
	original := parent.findAll(shared)
	var workers sync.WaitGroup
	for range 16 {
		workers.Go(func() {
			for range 32 {
				if len(parent.findAll(shared)) != 2 || len(entered.findAll(shared)) != 1 {
					t.Error("parent or child returned another scope's matches")
				}
			}
		})
	}
	workers.Wait()
	require.Len(t, original, 2)
	require.Equal(t, "Outside", original[0].Field(true))
	require.Equal(t, "Inside", original[1].Field(true))
	require.Empty(t, parent.walk.visited)
}

func TestTypeLayoutWalkPreservesCopiedValidationOccurrences(t *testing.T) {
	const owner = "example.test/service"
	node := &expr.UserTypeExpr{TypeName: "Node"}
	element := &expr.AttributeExpr{Type: node}
	node.AttributeExpr = &expr.AttributeExpr{Type: &expr.Object{
		{Name: "name", Attribute: &expr.AttributeExpr{Type: expr.String, Validation: &expr.ValidationExpr{Pattern: "[a-z]+"}}},
		{Name: "children", Attribute: &expr.AttributeExpr{Type: &expr.Array{ElemType: element}}},
	}}
	root := &expr.AttributeExpr{Type: node}
	generation, err := NewGeneration("example.test", nil)
	require.NoError(t, err)
	declaration := declareGoTypeTestUserType(t, generation, owner, node)
	validator := NewExactName(NameFunction, "ValidateNode")
	require.NoError(t, generation.Package(owner).DeclareName(validator))
	layout, err := PlanGoType(root, GoTypePlanOptions{
		Owner:            owner,
		Policy:           GoLayoutPolicy{SumType: true},
		RetainNamedValue: true,
		Bind:             goTypeTestBinder(map[expr.DataType]GoTypeBinding{node: {Owner: owner, Type: declaration}}),
	})
	require.NoError(t, err)
	original := layout.PlansForOccurrence(element)
	require.Len(t, original, 1)
	calls := 0
	_, err = NewValidationPlan(root, layout, ValidationPlanOptions{
		Required: true,
		Bind: func(request ValidatorBindingRequest) (*NameDeclaration, error) {
			calls++
			require.Same(t, element, request.Attribute)
			matches := request.Layout.PlansForOccurrence(element)
			require.Len(t, matches, 2)
			require.Same(t, request.Layout, matches[0])
			require.Same(t, original[0], matches[1])
			context := &AttributeContext{Scope: NewAttributeScope(NewNameScope())}
			linked, linkErr := context.WithGoTypeLayout(request.Layout.Link(owner, nil))
			require.NoError(t, linkErr)
			resolver := linked.Scope.(*goTypeLayoutAttributor)
			for range 2 {
				actual := resolver.findAll(element)
				require.Len(t, actual, 2)
				require.Same(t, matches[0], actual[0].plan)
				require.Same(t, matches[1], actual[1].plan)
				require.Empty(t, resolver.walk.visited)
			}
			return validator, nil
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, calls)
}
