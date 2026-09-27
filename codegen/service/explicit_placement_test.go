// These tests preserve explicit declaration owners while inferring unlocated
// descendants. Actual imports, rather than reservations, determine cycles.
package service

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestExplicitPlacementOwnsDescendants(t *testing.T) {
	for _, forced := range []bool{false, true} {
		for _, ambiguous := range []bool{false, true} {
			for _, reverse := range []bool{false, true} {
				var leaf, inner, outer expr.UserType
				first := codegen.RunDSL(t, func() {
					dsl.API("first", func() {})
					leaf = dsl.Type("Leaf", func() {
						dsl.Attribute("value", dsl.String)
						dsl.Attribute("next", "Leaf")
						dsl.Required("value")
					})
					inner = dsl.Type("Inner", func() {
						dsl.Meta("struct:pkg:path", "right/types")
						dsl.Attribute("leaf", leaf)
						dsl.Attribute("leaves", dsl.ArrayOf(leaf))
						dsl.Attribute("by_name", dsl.MapOf(dsl.String, leaf))
					})
					outer = dsl.Type("Outer", func() {
						dsl.Meta("struct:pkg:path", "left/types")
						if forced {
							dsl.Meta("type:generate:force")
						}
						dsl.Attribute("inner", inner)
					})
					dsl.Service("catalog", func() {
						if !forced {
							dsl.Method("exchange", func() {
								dsl.Payload(outer)
								dsl.Result(outer)
							})
						}
					})
				})
				second := codegen.RunDSL(t, func() {
					dsl.API("second", func() {})
					dsl.Type("Other", func() {
						if ambiguous {
							dsl.Meta("struct:pkg:path", "third/types")
						} else {
							dsl.Meta("struct:pkg:path", "right/types")
						}
						dsl.Attribute("leaf", leaf)
					})
				})
				roots := []*expr.RootExpr{first, second}
				if reverse {
					slices.Reverse(roots)
				}
				generation, inputs := inheritedPlacementInputs(t, roots)
				plans, err := NewPlans(generation, inputs...)
				if ambiguous {
					require.ErrorContains(t, err, `type "Leaf" requires incompatible packages`)
					require.Nil(t, plans)
					continue
				}
				require.NoError(t, err)
				found := make(map[expr.UserType]bool)
				for original, declaration := range generation.UserTypes() {
					switch original.Origin() {
					case leaf.Origin(), inner.Origin():
						require.Equal(t, "generated.local/gen/right/types", declaration.PackagePath())
						found[original.Origin()] = true
					case outer.Origin():
						require.Equal(t, "generated.local/gen/left/types", declaration.PackagePath())
						found[original.Origin()] = true
					}
				}
				require.Len(t, found, 3)
				require.NoError(t, generation.Freeze())
				for _, plan := range plans {
					require.NoError(t, plan.Link())
				}
				files, err := Files(plans...)
				require.NoError(t, err)
				compileGeneratedServiceFilesWith(t, files, nil)
				require.NotContains(t, leaf.Attribute().Meta, "struct:pkg:path")
				require.Equal(t, []string{"right/types"}, inner.Attribute().Meta["struct:pkg:path"])
			}
		}
	}
}

func TestExplicitPlacementConflictingOriginCopies(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		var original, copied expr.UserType
		first := codegen.RunDSL(t, func() {
			dsl.API("first", func() {})
			original = dsl.Type("Value", func() {
				dsl.Meta("struct:pkg:path", "first/types")
				dsl.Attribute("value", dsl.String)
			})
			dsl.Service("alpha", func() {
				dsl.Method("read", func() {
					dsl.Payload(original)
				})
			})
		})
		// Compiler copies preserve Origin. Only the synthetic copy's metadata
		// is edited to test contradictory explicit claims at the planner input.
		attribute := expr.DupAtt(original.Attribute())
		attribute.Meta["struct:pkg:path"] = []string{"second/types"}
		copied = original.Dup(attribute)
		second := codegen.RunDSL(t, func() {
			dsl.API("second", func() {})
			dsl.Service("beta", func() {
				dsl.Method("read", func() {
					dsl.Payload(copied)
				})
			})
		})
		roots := []*expr.RootExpr{first, second}
		if reverse {
			slices.Reverse(roots)
		}
		generation, inputs := inheritedPlacementInputs(t, roots)
		plans, err := NewPlans(generation, inputs...)
		require.ErrorContains(t, err, "explicit type package placement failed")
		require.ErrorContains(t, err, `type "Value" requires incompatible packages`)
		require.Nil(t, plans)
		require.Equal(t, []string{"first/types"}, original.Attribute().Meta["struct:pkg:path"])
		require.Equal(t, []string{"second/types"}, copied.Attribute().Meta["struct:pkg:path"])
	}
}

func TestExplicitPlacementUnannotatedOriginCopy(t *testing.T) {
	var original expr.UserType
	first := codegen.RunDSL(t, func() {
		dsl.API("first", func() {})
		original = dsl.Type("Value", func() {
			dsl.Meta("struct:pkg:path", "left/types")
			dsl.Attribute("value", dsl.String)
		})
	})
	attribute := expr.DupAtt(original.Attribute())
	delete(attribute.Meta, "struct:pkg:path")
	copied := original.Dup(attribute)
	second := codegen.RunDSL(t, func() {
		dsl.API("second", func() {})
		outer := dsl.Type("Outer", func() {
			dsl.Meta("struct:pkg:path", "right/types")
			dsl.Attribute("value", copied)
		})
		dsl.Service("catalog", func() {
			dsl.Method("read", func() {
				dsl.Payload(outer)
			})
		})
	})
	generation, plans := inheritedPlacementPlans(t, []*expr.RootExpr{second, first})
	declaration, err := generation.Package("generated.local/gen/left/types").Type(original)
	require.NoError(t, err)
	retainedAttribute, layout, err := plans[0].UserTypeLayout(copied, declaration)
	require.NoError(t, err)
	require.Same(t, original.Origin(), retainedAttribute.Type)
	require.NotSame(t, copied, retainedAttribute.Type)
	require.True(t, layout.MatchesOccurrence(retainedAttribute))
	require.Equal(t, "generated.local/gen/left/types", layout.Owner())
	otherAttribute, otherLayout, err := plans[1].UserTypeLayout(original, declaration)
	require.NoError(t, err)
	require.Same(t, retainedAttribute, otherAttribute)
	require.Same(t, layout, otherLayout)
	require.NotContains(t, copied.Attribute().Meta, "struct:pkg:path")
}

func TestRequiredPackageImportCycles(t *testing.T) {
	for _, samePackage := range []bool{false, true} {
		for _, forced := range []bool{false, true} {
			root := codegen.RunDSL(t, func() {
				first := dsl.Type("First", func() {
					dsl.Meta("struct:pkg:path", "first/types")
					if forced {
						dsl.Meta("type:generate:force")
					}
					dsl.Attribute("second", "Second")
				})
				dsl.Type("Second", func() {
					if samePackage {
						dsl.Meta("struct:pkg:path", "first/types")
					} else {
						dsl.Meta("struct:pkg:path", "second/types")
					}
					dsl.Attribute("first", first)
				})
				dsl.Service("catalog", func() {
					if !forced {
						dsl.Method("read", func() {
							dsl.Payload(first)
						})
					}
				})
			})
			generation, inputs := inheritedPlacementInputs(t, []*expr.RootExpr{root})
			plans, err := NewPlans(generation, inputs...)
			if samePackage {
				require.NoError(t, err)
				require.Len(t, plans, 1)
			} else {
				require.ErrorContains(t, err, "required generated package import cycle")
				require.Nil(t, plans)
				require.False(t, generation.Frozen())
			}
		}
	}
}

func TestPackageReservationsDoNotCreateDeclarationCycles(t *testing.T) {
	var inner, outer expr.UserType
	root := codegen.RunDSL(t, func() {
		inner = dsl.Type("Inner", dsl.String, func() {
			dsl.Meta("struct:pkg:path", "right/types")
		})
		outer = dsl.Type("Outer", func() {
			dsl.Meta("struct:pkg:path", "left/types")
			dsl.Meta("type:generate:force")
			dsl.Attribute("inner", inner)
		})
		dsl.Service("catalog", func() {})
	})
	generation, plans := inheritedPlacementPlans(t, []*expr.RootExpr{root})
	declaration, err := generation.Package("generated.local/gen/left/types").Type(outer)
	require.NoError(t, err)
	attribute, layout, err := plans[0].UserTypeLayout(outer, declaration)
	require.NoError(t, err)
	require.True(t, layout.MatchesOccurrence(attribute))
	// A possible conversion reserves Outer in right, but no emitted right-side
	// declaration refers to Outer. Only the actual left-to-right edge exists.
	reservations := codegen.NewGeneratedImportPlan(generation.Package("generated.local/gen/right/types"))
	require.NoError(t, reservations.AddCompleteType(layout))
	require.Contains(t, reservations.Paths(), "generated.local/gen/left/types")
	require.NoError(t, validateRequiredPackageImports([]*rootFacts{plans[0].facts}))
	require.NoError(t, generation.Freeze())
	require.NoError(t, plans[0].Link())
	files, err := Files(plans...)
	require.NoError(t, err)
	compileGeneratedServiceFilesWith(t, files, nil)
}
