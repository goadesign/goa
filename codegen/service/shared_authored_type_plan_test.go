// This file verifies that service planning recognizes authored union children
// across design roots before assigning their generated package declarations.
package service

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// TestNewPlansShareAuthoredUnionChildAcrossRoots checks that a shared recursive
// child keeps its authored declaration while an anonymous primitive branch
// receives a separate union-owned declaration, regardless of root order.
func TestNewPlansShareAuthoredUnionChildAcrossRoots(t *testing.T) {
	for _, test := range []struct {
		name    string
		reverse bool
	}{
		{name: "forward roots"},
		{name: "reverse roots", reverse: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			var child, container expr.UserType
			first := codegen.RunDSL(t, func() {
				child = dsl.Type("Child", func() {
					dsl.Meta("struct:pkg:path", "shared/types")
					dsl.Attribute("value", dsl.String)
					dsl.Attribute("next", "Child")
					dsl.Required("value")
				})
				container = dsl.Type("Container", func() {
					dsl.Meta("struct:pkg:path", "shared/types")
					dsl.OneOf("selection", func() {
						dsl.TypeName("Selection")
						dsl.Attribute("child", child)
						dsl.Attribute("text", dsl.String)
					})
				})
				dsl.Service("Alpha", func() {
					dsl.Method("Use", func() {
						dsl.Payload(container)
					})
				})
			})
			second := codegen.RunDSL(t, func() {
				dsl.Service("Beta", func() {
					dsl.Method("Use", func() {
						dsl.Payload(container)
					})
				})
			})
			roots := []*expr.RootExpr{first, second}
			if test.reverse {
				slices.Reverse(roots)
			}
			evaluated := make([]eval.Root, len(roots))
			inputs := make([]PlanInput, len(roots))
			for i, root := range roots {
				evaluated[i] = root
				inputs[i] = PlanInput{Root: root, Examples: expr.NewExampleGenerator(root.API.RandomizerFactory)}
			}
			generation := mustTestGeneration(t, "generated.local/gen", evaluated)
			_, err := NewPlans(generation, inputs...)
			require.NoError(t, err)

			owner := generation.Package("generated.local/gen/shared/types")
			childDeclaration, err := owner.UserType(child)
			require.NoError(t, err)
			var textDeclaration *codegen.TypeDeclaration
			for _, root := range roots {
				selection := root.Services[0].Methods[0].Payload.Find("selection")
				require.NotNil(t, selection)
				union := expr.AsUnion(selection.Type)
				require.NotNil(t, union)
				require.Len(t, union.Values, 2)
				require.Equal(t, "child", union.Values[0].Name)
				require.Equal(t, "text", union.Values[1].Name)

				childType, ok := union.Values[0].Attribute.Type.(expr.UserType)
				require.True(t, ok)
				require.Same(t, child.Origin(), childType.Origin())
				actualChild, err := owner.Type(childType)
				require.NoError(t, err)
				require.Same(t, childDeclaration, actualChild)
				childBranch, err := owner.UnionBranch(selection, "child")
				require.NoError(t, err)
				_, generated := childBranch.Type()
				require.False(t, generated)

				textType, ok := union.Values[1].Attribute.Type.(expr.UserType)
				require.True(t, ok)
				branchDeclaration, err := owner.UnionBranchType(selection, "text")
				require.NoError(t, err)
				require.NotSame(t, childDeclaration, branchDeclaration)
				actualText, err := owner.Type(textType)
				require.NoError(t, err)
				require.Same(t, branchDeclaration, actualText)
				_, err = owner.UserType(textType)
				require.Error(t, err)
				if textDeclaration == nil {
					textDeclaration = branchDeclaration
				} else {
					require.Same(t, textDeclaration, branchDeclaration)
				}
			}
		})
	}
}
