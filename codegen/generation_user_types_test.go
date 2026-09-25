// This file checks that plugins can discover original type declarations and
// reserve related names without finalizing or changing the current iteration.
package codegen

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func TestGenerationUserTypesOriginalOwners(t *testing.T) {
	for _, test := range []struct {
		name    string
		reverse bool
	}{
		{name: "forward"},
		{name: "reverse", reverse: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			generation := mustTestGeneration(t, "generated.local/gen", nil)
			apple := generatedUserType("apple", "apple")
			zebra := generatedUserType("Zebra", "zebra")
			otherApple := generatedUserType("apple", "apple")
			copy := apple.Dup(expr.DupAtt(apple.Attribute()))
			// The claim spelling sorts after b, but its actual package is a.
			claims := []string{
				"generated.local/gen/z/../a",
				"generated.local/gen/b",
				"generated.local/gen/c",
			}
			if test.reverse {
				slices.Reverse(claims)
			}
			for _, claim := range claims {
				mustClaimTestPackage(t, generation, claim)
			}
			a := generation.Package("generated.local/gen/a")
			b := generation.Package("generated.local/gen/b")
			c := generation.Package("generated.local/gen/c")
			registrations := []struct {
				pkg      *GeneratedPackage
				original expr.UserType
			}{
				{a, copy},
				{a, zebra},
				{b, apple},
				{c, otherApple},
			}
			if test.reverse {
				slices.Reverse(registrations)
			}
			for _, registration := range registrations {
				_, err := registration.pkg.DeclareUserType(registration.original)
				require.NoError(t, err)
			}
			aApple, err := a.DeclareUserType(apple)
			require.NoError(t, err)
			aCopy, err := a.DeclareUserType(copy)
			require.NoError(t, err)
			require.Same(t, aApple, aCopy)
			aZebra, err := a.UserType(zebra)
			require.NoError(t, err)
			bApple, err := b.UserType(apple)
			require.NoError(t, err)
			cApple, err := c.UserType(otherApple)
			require.NoError(t, err)

			wantTypes := []expr.UserType{apple, zebra, apple, otherApple}
			wantDeclarations := []*TypeDeclaration{aApple, aZebra, bApple, cApple}
			for _, frozen := range []bool{false, true} {
				if frozen {
					require.NoError(t, generation.Freeze())
				}
				var originals []expr.UserType
				var declarations []*TypeDeclaration
				for original, declaration := range generation.UserTypes() {
					originals = append(originals, original)
					declarations = append(declarations, declaration)
				}
				require.Len(t, originals, len(wantTypes))
				for i := range wantTypes {
					require.Same(t, wantTypes[i], originals[i])
					require.Same(t, wantDeclarations[i], declarations[i])
				}
			}
			require.Equal(t, "Apple", aApple.Name())
			require.Equal(t, "Zebra", aZebra.Name())
		})
	}
}

func TestGenerationUserTypesReserveDependentNames(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	pkg := mustClaimTestPackage(t, generation, "generated.local/gen/types")
	for _, name := range []string{"Value", "InspectValue"} {
		_, err := pkg.DeclareUserType(generatedUserType(name, name))
		require.NoError(t, err)
	}

	companions := make(map[string]*NameDeclaration)
	for original, declaration := range generation.UserTypes() {
		require.False(t, generation.Frozen())
		require.Panics(t, func() {
			declaration.Name()
		})
		owner := generation.Package(declaration.PackagePath())
		require.True(t, owner.OwnsName(declaration.Declaration()))
		companion, err := owner.DeclareDependentName(
			NameFunction, declaration.Declaration(), "Inspect", "",
			testNameOrder{value: original.Name()},
		)
		require.NoError(t, err)
		companions[original.Name()] = companion
	}
	require.Len(t, companions, 2)
	require.NoError(t, generation.Freeze())
	require.Equal(t, "InspectValue2", companions["Value"].Name())
	require.Equal(t, "InspectInspectValue", companions["InspectValue"].Name())
}

func TestGenerationUserTypesExcludeGeneratedDeclarations(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	pkg := mustClaimTestPackage(t, generation, "generated.local/gen/types")
	original := generatedUserType("Original", "original")
	authored, err := pkg.DeclareUserType(original)
	require.NoError(t, err)

	synthetic := generatedUserType("Transport", "transport")
	transport, err := pkg.DeclareGeneratedType("Transport", testNameOrder{value: "transport"})
	require.NoError(t, err)
	require.NoError(t, pkg.BindGeneratedType(synthetic, transport))
	_, err = pkg.DeclareDerivedType(NewProjectedTypeID(original), "OriginalView")
	require.NoError(t, err)
	_, err = pkg.DeclareDerivedType(NewViewedResultTypeID(original), "ViewedOriginal")
	require.NoError(t, err)
	union, branch := generatedUnionWithBranch("branch")
	_, err = pkg.DeclareUnion(union)
	require.NoError(t, err)
	_, err = pkg.DeclareUnionBranchType(union, "text", branch)
	require.NoError(t, err)

	for _, frozen := range []bool{false, true} {
		if frozen {
			require.NoError(t, generation.Freeze())
		}
		count := 0
		for got, declaration := range generation.UserTypes() {
			count++
			require.Same(t, original, got)
			require.Same(t, authored, declaration)
		}
		require.Equal(t, 1, count)
	}
}

func TestGenerationUserTypesSnapshotIteration(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	pkg := mustClaimTestPackage(t, generation, "generated.local/gen/types")
	first := generatedUserType("First", "first")
	second := generatedUserType("Second", "second")
	for _, original := range []expr.UserType{first, second} {
		_, err := pkg.DeclareUserType(original)
		require.NoError(t, err)
	}
	sequence := generation.UserTypes()
	var seen []expr.UserType
	for original := range sequence {
		seen = append(seen, original)
		if len(seen) == 1 {
			_, err := pkg.DeclareUserType(generatedUserType("Added", "added"))
			require.NoError(t, err)
			other := mustClaimTestPackage(t, generation, "generated.local/gen/other")
			_, err = other.DeclareUserType(first)
			require.NoError(t, err)
		}
	}
	require.Equal(t, []expr.UserType{first, second}, seen)

	var owners, names []string
	for original, declaration := range sequence {
		owners = append(owners, declaration.PackagePath())
		names = append(names, original.Name())
	}
	require.Equal(t, []string{
		"generated.local/gen/other",
		"generated.local/gen/types",
		"generated.local/gen/types",
		"generated.local/gen/types",
	}, owners)
	require.Equal(t, []string{"First", "Added", "First", "Second"}, names)
}

func TestGenerationUserTypesEarlyTermination(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	pkg := mustClaimTestPackage(t, generation, "generated.local/gen/types")
	for _, name := range []string{"First", "Second"} {
		_, err := pkg.DeclareUserType(generatedUserType(name, name))
		require.NoError(t, err)
	}
	sequence := generation.UserTypes()
	calls := 0
	sequence(func(expr.UserType, *TypeDeclaration) bool {
		calls++
		return false
	})
	require.Equal(t, 1, calls)

	var names []string
	for original := range sequence {
		names = append(names, original.Name())
	}
	require.Equal(t, []string{"First", "Second"}, names)
}

func TestGenerationUserTypesIsolation(t *testing.T) {
	first := mustTestGeneration(t, "generated.local/gen", nil)
	second := mustTestGeneration(t, "generated.local/gen", nil)
	original := generatedUserType("Value", "value")
	firstPackage := mustClaimTestPackage(t, first, "generated.local/gen/types")
	firstDeclaration, err := firstPackage.DeclareUserType(original)
	require.NoError(t, err)

	emptyCount := 0
	for range second.UserTypes() {
		emptyCount++
	}
	require.Zero(t, emptyCount)
	require.NoError(t, first.Freeze())
	secondPackage := mustClaimTestPackage(t, second, "generated.local/gen/types")
	secondDeclaration, err := secondPackage.DeclareUserType(original)
	require.NoError(t, err)
	require.NotSame(t, firstDeclaration, secondDeclaration)
	for _, generation := range []*Generation{first, second} {
		count := 0
		for got, declaration := range generation.UserTypes() {
			count++
			require.Same(t, original, got)
			require.True(t, generation.OwnsName(declaration.Declaration()))
			require.NotEqual(t, first.OwnsName(declaration.Declaration()), second.OwnsName(declaration.Declaration()))
		}
		require.Equal(t, 1, count)
	}
}
