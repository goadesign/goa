// This file verifies that one generated package owns the public names of every
// relocated declaration planned into it.
package codegen

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// TestGeneratedTypesRejectRelocatedNameCollision verifies that one generated
// package rejects distinct DSL names that produce the same exported Go name.
func TestGeneratedTypesRejectRelocatedNameCollision(t *testing.T) {
	var first, second expr.UserType
	root := RunDSL(t, func() {
		first = dsl.Type("foo-bar", func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.Attribute("first", dsl.String)
		})
		second = dsl.Type("foo_bar", func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.Attribute("second", dsl.String)
		})
	})

	generation := NewGeneration("generated.local/gen", []eval.Root{root})
	types := generation.GeneratedPackage("generated.local/gen/types")
	_, err := types.DeclareUserType(first)
	require.NoError(t, err)
	_, err = types.DeclareUserType(second)
	require.ErrorContains(t, err, "foo-bar")
	require.ErrorContains(t, err, "foo_bar")
	require.ErrorContains(t, err, "FooBar")
}

// TestGenerationOwnsPackageRecords verifies that one generation returns one
// stable package record and scope for each output path.
func TestGenerationOwnsPackageRecords(t *testing.T) {
	generation := NewGeneration("generated.local/gen", nil)

	first := generation.GeneratedPackage("generated.local/gen/types")
	second := generation.GeneratedPackage("generated.local/gen/types")
	other := generation.GeneratedPackage("generated.local/gen/other")
	require.Same(t, first, second)
	require.Panics(t, func() {
		first.Scope()
	})
	require.NoError(t, generation.Freeze())
	require.Same(t, first.Scope(), second.Scope())
	require.NotSame(t, first, other)
	require.NotSame(t, first.Scope(), other.Scope())
}

// TestGeneratedPackageUserTypes verifies that a generated package records one
// declaration per user type and that lookups do not reserve names.
func TestGeneratedPackageUserTypes(t *testing.T) {
	generation := NewGeneration("generated.local/gen", nil)
	types := generation.GeneratedPackage("generated.local/gen/types")
	widget := generatedUserType("Widget", "widget")
	missing := generatedUserType("Missing", "missing")

	_, err := types.UserType(missing)
	require.ErrorContains(t, err, "not declared")

	first, err := types.DeclareUserType(widget)
	require.NoError(t, err)
	require.Equal(t, &TypeDeclaration{
		Name:        "Widget",
		PackagePath: "generated.local/gen/types",
	}, first)
	second, err := types.DeclareUserType(widget)
	require.NoError(t, err)
	require.Same(t, first, second)

	lookedUp, err := types.UserType(widget)
	require.NoError(t, err)
	require.Same(t, first, lookedUp)
	declaredMissing, err := types.DeclareUserType(missing)
	require.NoError(t, err)
	require.Equal(t, "Missing", declaredMissing.Name)
	require.NoError(t, generation.Freeze())
	require.Equal(t, "Widget", types.Scope().GoTypeName(&expr.AttributeExpr{Type: widget}))
}

// TestGeneratedPackageExactUserTypesDoNotMerge verifies that structural
// equality does not weaken the exact-name contract for DSL declarations.
func TestGeneratedPackageExactUserTypesDoNotMerge(t *testing.T) {
	generation := NewGeneration("generated.local/gen", nil)
	types := generation.GeneratedPackage("generated.local/gen/types")
	first := generatedUserType("ValueText", "first")
	equivalent := generatedUserType("ValueText", "second")

	_, err := types.DeclareUserType(first)
	require.NoError(t, err)
	_, err = types.DeclareUserType(equivalent)
	require.ErrorContains(t, err, "ValueText")
	require.ErrorContains(t, err, "already declared")
}

// TestGeneratedPackageUserTypeCopiesShareDeclaration verifies that exact
// compiler copies use their declaration origin instead of one transient copy
// pointer as package identity.
func TestGeneratedPackageUserTypeCopiesShareDeclaration(t *testing.T) {
	generation := NewGeneration("generated.local/gen", nil)
	types := generation.GeneratedPackage("generated.local/gen/types")
	original := generatedUserType("ValueText", "value-text")
	copy := original.Dup(expr.DupAtt(original.Attribute()))

	first, err := types.DeclareUserType(original)
	require.NoError(t, err)
	second, err := types.DeclareUserType(copy)
	require.NoError(t, err)
	require.Same(t, first, second)
	require.NoError(t, generation.Freeze())

	lookedUp, err := types.UserType(copy)
	require.NoError(t, err)
	require.Same(t, first, lookedUp)
}

// TestGeneratedPackageDerivedTypesUseTypedSourceIdentity verifies that view
// declarations rebuilt in the render phase select the records planned from
// the same exact source declaration.
func TestGeneratedPackageDerivedTypesUseTypedSourceIdentity(t *testing.T) {
	generation := NewGeneration("generated.local/gen", nil)
	views := generation.GeneratedPackage("generated.local/gen/service/views")
	source := generatedUserType("Value", "value")
	copy := source.Dup(expr.DupAtt(source.Attribute()))
	projectedID := NewProjectedTypeID(source)
	viewedID := NewViewedResultTypeID(source)

	projected, err := views.DeclareDerivedType(projectedID, "ValueView")
	require.NoError(t, err)
	viewed, err := views.DeclareDerivedType(viewedID, "Value")
	require.NoError(t, err)
	require.NotSame(t, projected, viewed)
	require.NoError(t, generation.Freeze())

	projectedCopy, err := views.DerivedType(NewProjectedTypeID(copy))
	require.NoError(t, err)
	require.Same(t, projected, projectedCopy)
	viewedCopy, err := views.DerivedType(NewViewedResultTypeID(copy))
	require.NoError(t, err)
	require.Same(t, viewed, viewedCopy)
	require.Equal(t, "ValueView", projected.Name)
	require.Equal(t, "Value", viewed.Name)
}

// TestGeneratedPackageUnionBranchesShareDeclaration verifies that separately
// allocated copies of one structural union reuse their generated branch alias.
func TestGeneratedPackageUnionBranchesShareDeclaration(t *testing.T) {
	generation := NewGeneration("generated.local/gen", nil)
	types := generation.GeneratedPackage("generated.local/gen/types")
	firstUnion, firstAlias := generatedUnionWithBranch("Value", "text", "first", expr.String)
	secondUnion, secondAlias := generatedUnionWithBranch("Value", "text", "second", expr.String)

	_, err := types.DeclareUnion(firstUnion)
	require.NoError(t, err)
	firstDeclaration, err := types.DeclareUnionBranchType(firstUnion, "text", firstAlias)
	require.NoError(t, err)
	_, err = types.DeclareUnion(secondUnion)
	require.NoError(t, err)
	secondDeclaration, err := types.DeclareUnionBranchType(secondUnion, "text", secondAlias)
	require.NoError(t, err)
	require.Same(t, firstDeclaration, secondDeclaration)
	require.Empty(t, firstDeclaration.Name)

	require.NoError(t, generation.Freeze())
	require.Equal(t, "ValueText", firstDeclaration.Name)
	lookedUp, err := types.UnionBranchType(secondUnion, "text")
	require.NoError(t, err)
	require.Same(t, firstDeclaration, lookedUp)
}

// TestGeneratedPackageUnionBranchesAreIsolatedByUnion verifies that branch
// aliases from different emitted union definitions never collapse together.
func TestGeneratedPackageUnionBranchesAreIsolatedByUnion(t *testing.T) {
	generation := NewGeneration("generated.local/gen", nil)
	types := generation.GeneratedPackage("generated.local/gen/types")
	firstUnion, firstAlias := generatedUnionWithBranch("Value", "text", "first", expr.String)
	secondUnion, secondAlias := generatedUnionWithBranch("Value", "text", "second", expr.String)
	secondUnion.TypeKey = "kind"

	_, err := types.DeclareUnion(firstUnion)
	require.NoError(t, err)
	firstDeclaration, err := types.DeclareUnionBranchType(firstUnion, "text", firstAlias)
	require.NoError(t, err)
	_, err = types.DeclareUnion(secondUnion)
	require.NoError(t, err)
	secondDeclaration, err := types.DeclareUnionBranchType(secondUnion, "text", secondAlias)
	require.NoError(t, err)
	require.NotSame(t, firstDeclaration, secondDeclaration)

	require.NoError(t, generation.Freeze())
	require.ElementsMatch(t, []string{"ValueText", "ValueText2"}, []string{
		firstDeclaration.Name,
		secondDeclaration.Name,
	})
}

// TestGeneratedPackageUnionFamilyAvoidsExactTypeNames verifies that union
// constants and constructors use package-owned frozen names instead of
// colliding with exact DSL type declarations.
func TestGeneratedPackageUnionFamilyAvoidsExactTypeNames(t *testing.T) {
	generation := NewGeneration("generated.local/gen", nil)
	types := generation.GeneratedPackage("generated.local/gen/types")
	for _, name := range []string{"ValueKindText", "NewValueText"} {
		_, err := types.DeclareUserType(generatedUserType(name, name))
		require.NoError(t, err)
	}
	union, alias := generatedUnionWithBranch("Value", "text", "text", expr.String)
	_, err := types.DeclareUnion(union)
	require.NoError(t, err)
	aliasDeclaration, err := types.DeclareUnionBranchType(union, "text", alias)
	require.NoError(t, err)

	require.NoError(t, generation.Freeze())
	branch, err := types.UnionBranch(union, "text")
	require.NoError(t, err)
	require.Equal(t, "ValueKindText2", branch.KindConst)
	require.Equal(t, "NewValueText2", branch.Constructor)
	require.Same(t, aliasDeclaration, branch.Type)
}

// TestGeneratedPackageUnions verifies that emitted-definition identity makes
// equivalent unions idempotent while different unions with the same base name
// receive distinct declarations.
func TestGeneratedPackageUnions(t *testing.T) {
	generation := NewGeneration("generated.local/gen", nil)
	types := generation.GeneratedPackage("generated.local/gen/types")
	first := generatedUnion("Value", "type", "value")
	equivalent := generatedUnion("Value", "type", "value")
	different := generatedUnion("Value", "kind", "data")

	firstDeclaration, err := types.DeclareUnion(first)
	require.NoError(t, err)
	require.Equal(t, &UnionDeclaration{
		PackagePath: "generated.local/gen/types",
	}, firstDeclaration)
	equivalentDeclaration, err := types.DeclareUnion(equivalent)
	require.NoError(t, err)
	require.Same(t, firstDeclaration, equivalentDeclaration)

	differentDeclaration, err := types.DeclareUnion(different)
	require.NoError(t, err)
	require.Empty(t, differentDeclaration.Name)
	require.NotSame(t, firstDeclaration, differentDeclaration)

	lookedUp, err := types.Union(equivalent)
	require.NoError(t, err)
	require.Same(t, firstDeclaration, lookedUp)
	require.NoError(t, generation.Freeze())
	require.ElementsMatch(t, []string{"Value", "Value2"}, []string{
		firstDeclaration.Name,
		differentDeclaration.Name,
	})
	require.ElementsMatch(t, []string{"ValueKind", "Value2Kind"}, []string{
		firstDeclaration.KindName,
		differentDeclaration.KindName,
	})

	reversedGeneration := NewGeneration("generated.local/gen", nil)
	reversedTypes := reversedGeneration.GeneratedPackage("generated.local/gen/types")
	reversedDifferent, err := reversedTypes.DeclareUnion(generatedUnion("Value", "kind", "data"))
	require.NoError(t, err)
	reversedFirst, err := reversedTypes.DeclareUnion(generatedUnion("Value", "type", "value"))
	require.NoError(t, err)
	require.NoError(t, reversedGeneration.Freeze())
	require.Equal(t, firstDeclaration.Name, reversedFirst.Name)
	require.Equal(t, differentDeclaration.Name, reversedDifferent.Name)
}

// TestGeneratedPackageUserTypeWinsUnionNamesRegardlessOfOrder verifies that
// pending unions cannot take exact user-type or discriminator names based on
// traversal order.
func TestGeneratedPackageUserTypeWinsUnionNamesRegardlessOfOrder(t *testing.T) {
	for _, unionFirst := range []bool{true, false} {
		t.Run(fmt.Sprintf("union first %t", unionFirst), func(t *testing.T) {
			generation := NewGeneration("generated.local/gen", nil)
			types := generation.GeneratedPackage("generated.local/gen/types")
			userType := generatedUserType("Value", "value")
			kindUserType := generatedUserType("ValueKind", "value-kind")
			union := generatedUnion("Value", "type", "value")
			var (
				userDeclaration  *TypeDeclaration
				kindDeclaration  *TypeDeclaration
				unionDeclaration *UnionDeclaration
				err              error
			)
			if unionFirst {
				unionDeclaration, err = types.DeclareUnion(union)
				require.NoError(t, err)
				require.Panics(t, func() {
					types.Scope().GoTypeName(&expr.AttributeExpr{Type: union})
				})
				userDeclaration, err = types.DeclareUserType(userType)
				require.NoError(t, err)
				kindDeclaration, err = types.DeclareUserType(kindUserType)
				require.NoError(t, err)
			} else {
				userDeclaration, err = types.DeclareUserType(userType)
				require.NoError(t, err)
				kindDeclaration, err = types.DeclareUserType(kindUserType)
				require.NoError(t, err)
				require.Panics(t, func() {
					types.Scope().GoTypeName(&expr.AttributeExpr{Type: userType})
				})
				unionDeclaration, err = types.DeclareUnion(union)
				require.NoError(t, err)
			}

			require.Equal(t, "Value", userDeclaration.Name)
			require.Equal(t, "ValueKind", kindDeclaration.Name)
			require.Empty(t, unionDeclaration.Name)
			require.Empty(t, unionDeclaration.KindName)
			require.NoError(t, generation.Freeze())
			require.Equal(t, "Value", userDeclaration.Name)
			require.Equal(t, "ValueKind", kindDeclaration.Name)
			require.Equal(t, "Value2", unionDeclaration.Name)
			require.Equal(t, "Value2Kind", unionDeclaration.KindName)
			require.Equal(t, "Value", types.Scope().GoTypeName(&expr.AttributeExpr{Type: userType}))
			require.Equal(t, "ValueKind", types.Scope().GoTypeName(&expr.AttributeExpr{Type: kindUserType}))
			require.Equal(t, "Value2", types.Scope().GoTypeName(&expr.AttributeExpr{Type: union}))
		})
	}
}

// TestGeneratedPackageLookupAcrossFreeze verifies that freeze keeps existing
// declarations readable and rejects every later declaration attempt.
func TestGeneratedPackageLookupAcrossFreeze(t *testing.T) {
	generation := NewGeneration("generated.local/gen", nil)
	types := generation.GeneratedPackage("generated.local/gen/types")
	widget := generatedUserType("Widget", "widget")
	union := generatedUnion("Value", "type", "value")
	userDeclaration, err := types.DeclareUserType(widget)
	require.NoError(t, err)
	unionDeclaration, err := types.DeclareUnion(union)
	require.NoError(t, err)
	require.Empty(t, unionDeclaration.Name)

	require.NoError(t, generation.Freeze())
	lookedUpUser, err := types.UserType(widget)
	require.NoError(t, err)
	require.Same(t, userDeclaration, lookedUpUser)
	lookedUpUnion, err := types.Union(union)
	require.NoError(t, err)
	require.Same(t, unionDeclaration, lookedUpUnion)
	require.Equal(t, "Value", lookedUpUnion.Name)
	require.Equal(t, "Widget", types.Scope().GoTypeName(&expr.AttributeExpr{Type: widget}))
	require.Equal(t, "Value", types.Scope().GoTypeName(&expr.AttributeExpr{Type: union}))
	require.Panics(t, func() {
		types.Scope().Unique("Late")
	})
	require.Panics(t, func() {
		types.Scope().GoTypeName(&expr.AttributeExpr{Type: generatedUserType("Late", "late")})
	})

	_, err = types.DeclareUserType(widget)
	require.ErrorContains(t, err, "frozen")
	_, err = types.DeclareUnion(union)
	require.ErrorContains(t, err, "frozen")
}

// TestGenerationCatalogsAreIsolated verifies that standalone generation runs
// do not share declaration records or name reservations.
func TestGenerationCatalogsAreIsolated(t *testing.T) {
	firstGeneration := NewGeneration("generated.local/gen", nil)
	first := firstGeneration.GeneratedPackage("generated.local/gen/types")
	secondGeneration := NewGeneration("generated.local/gen", nil)
	second := secondGeneration.GeneratedPackage("generated.local/gen/types")
	firstUnion := generatedUnion("Value", "type", "value")
	secondUnion := generatedUnion("Value", "type", "value")

	firstDeclaration, err := first.DeclareUnion(firstUnion)
	require.NoError(t, err)
	secondDeclaration, err := second.DeclareUnion(secondUnion)
	require.NoError(t, err)
	require.NoError(t, firstGeneration.Freeze())
	require.NoError(t, secondGeneration.Freeze())
	require.Equal(t, "Value", firstDeclaration.Name)
	require.Equal(t, "Value", secondDeclaration.Name)
	require.NotSame(t, firstDeclaration, secondDeclaration)
	require.NotSame(t, first.Scope(), second.Scope())
}

// generatedUserType builds a distinct user type for catalog tests.
func generatedUserType(name, id string) expr.UserType {
	return generatedUserTypeOf(name, id, expr.String)
}

// generatedUserTypeOf builds a distinct user type with the supplied shape.
func generatedUserTypeOf(name, id string, dataType expr.DataType) expr.UserType {
	return &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{Type: dataType},
		TypeName:      name,
		UID:           id,
	}
}

// generatedUnion builds a union whose emitted identity includes the supplied
// JSON envelope keys.
func generatedUnion(name, typeKey, valueKey string) *expr.Union {
	return &expr.Union{
		TypeName: name,
		TypeKey:  typeKey,
		ValueKey: valueKey,
	}
}

// generatedUnionWithBranch builds a union with one generated branch alias.
func generatedUnionWithBranch(unionName, branchName, aliasID string, dataType expr.DataType) (*expr.Union, expr.UserType) {
	alias := generatedUserTypeOf(unionName+expr.Title(branchName), aliasID, dataType)
	return &expr.Union{
		TypeName: unionName,
		Values: []*expr.NamedAttributeExpr{{
			Name:      branchName,
			Attribute: &expr.AttributeExpr{Type: alias},
		}},
	}, alias
}
