// This file verifies that one generated package owns the public names of every
// relocated declaration planned into it.
package codegen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

type (
	// testNameOrder supplies stable typed ordering facts to package-name tests.
	testNameOrder struct {
		value string
	}

	// alphaTestNameOrder and omegaTestNameOrder are unrelated order families
	// whose comparers reject foreign values.
	alphaTestNameOrder string
	omegaTestNameOrder string

	// unstableTestNameOrder is invalid because its value may change after
	// declaration collection.
	unstableTestNameOrder []string

	// indirectTestNameOrder is invalid because it contains pointer identity.
	indirectTestNameOrder struct {
		value *string
	}

	// testNameKey supplies the lookup key used by generated type formatters.
	testNameKey string
)

// ComparePackageName orders declarations from the same test family.
func (o testNameOrder) ComparePackageName(other PackageNameOrder) int {
	return strings.Compare(o.value, other.(testNameOrder).value)
}

// ComparePackageName orders declarations from the alpha test family.
func (o alphaTestNameOrder) ComparePackageName(other PackageNameOrder) int {
	return strings.Compare(string(o), string(other.(alphaTestNameOrder)))
}

// ComparePackageName orders declarations from the omega test family.
func (o omegaTestNameOrder) ComparePackageName(other PackageNameOrder) int {
	return strings.Compare(string(o), string(other.(omegaTestNameOrder)))
}

// ComparePackageName orders an invalid mutable test value.
func (o unstableTestNameOrder) ComparePackageName(other PackageNameOrder) int {
	return strings.Compare(strings.Join(o, "/"), strings.Join(other.(unstableTestNameOrder), "/"))
}

// ComparePackageName orders an invalid pointer-backed test value.
func (o indirectTestNameOrder) ComparePackageName(other PackageNameOrder) int {
	return strings.Compare(*o.value, *other.(indirectTestNameOrder).value)
}

// Hash returns the lookup key used by a generated package.
func (k testNameKey) Hash() string {
	return string(k)
}

// TestDeclareNameBindsLookupKeys checks that a plugin can declare a name and
// use the same final name through its normal typed lookup.
func TestDeclareNameBindsLookupKeys(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	pkg := mustClaimTestPackage(t, generation, "generated.local/gen/specs")
	key := testNameKey("request")
	declaration := NewPreferredName(NameType, "Request", ExportedName, testNameOrder{value: "request"})

	require.NoError(t, pkg.DeclareName(declaration, key))
	require.NoError(t, pkg.DeclareName(declaration, key))
	require.NoError(t, generation.Freeze())
	require.Equal(t, declaration.Name(), pkg.Scope().HashedUnique(key, "Ignored"))
}

// TestDeclareNameRejectsAnotherDeclarationForOneKey checks that one lookup
// key cannot select two package-level names.
func TestDeclareNameRejectsAnotherDeclarationForOneKey(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	pkg := mustClaimTestPackage(t, generation, "generated.local/gen/specs")
	key := testNameKey("request")
	first := NewPreferredName(NameType, "Request", ExportedName, testNameOrder{value: "first"})
	second := NewPreferredName(NameType, "Request", ExportedName, testNameOrder{value: "second"})

	require.NoError(t, pkg.DeclareName(first, key))
	err := pkg.DeclareName(second, key)
	require.ErrorContains(t, err, "lookup key")
}

// TestNameDeclarationOwnsOnePackageNamespace verifies that exact and preferred
// package symbols of every kind share one collision domain and one frozen name.
func TestNameDeclarationOwnsOnePackageNamespace(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	types := mustClaimTestPackage(t, generation, "generated.local/gen/types")
	exact := NewExactName(NameType, "Build")
	preferred := NewPreferredName(NameFunction, "Build", ExportedName, testNameOrder{value: "build"})

	require.NoError(t, types.DeclareName(exact))
	require.NoError(t, types.DeclareName(exact))
	require.NoError(t, types.DeclareName(preferred))
	require.Equal(t, NameType, exact.Kind())
	require.Panics(t, func() { exact.Name() })
	require.Panics(t, func() { preferred.Name() })

	require.NoError(t, generation.Freeze())
	require.Equal(t, "Build", exact.Name())
	require.Equal(t, "Build2", preferred.Name())
	require.Equal(t, "Build", exact.Name())

	for _, kind := range []PackageNameKind{NameType, NameFunction, NameConstant, NameVariable} {
		collisionGeneration := mustTestGeneration(t, "generated.local/gen", nil)
		pkg := mustClaimTestPackage(t, collisionGeneration, "generated.local/gen/types")
		require.NoError(t, pkg.DeclareName(NewExactName(NameType, "Shared")))
		err := pkg.DeclareName(NewExactName(kind, "Shared"))
		require.ErrorContains(t, err, "Shared")
	}
}

// TestDeclareGeneratedTypeUsesStablePackageNames verifies that plugins can
// declare generated types without claiming that they are authored Goa types.
func TestDeclareGeneratedTypeUsesStablePackageNames(t *testing.T) {
	declare := func(reverse bool) (string, string) {
		generation := mustTestGeneration(t, "generated.local/gen", nil)
		pkg := mustClaimTestPackage(t, generation, "generated.local/gen/types")
		firstOrder := testNameOrder{value: "first"}
		secondOrder := testNameOrder{value: "second"}
		var first, second *TypeDeclaration
		var err error
		if reverse {
			second, err = pkg.DeclareGeneratedType("Value", secondOrder)
			require.NoError(t, err)
			first, err = pkg.DeclareGeneratedType("Value", firstOrder)
		} else {
			first, err = pkg.DeclareGeneratedType("Value", firstOrder)
			require.NoError(t, err)
			second, err = pkg.DeclareGeneratedType("Value", secondOrder)
		}
		require.NoError(t, err)
		require.NoError(t, generation.Freeze())
		_, err = pkg.DeclareGeneratedType("Other", testNameOrder{value: "other"})
		require.ErrorContains(t, err, "frozen")
		return first.Name(), second.Name()
	}

	first, second := declare(false)
	reversedFirst, reversedSecond := declare(true)
	require.Equal(t, "Value", first)
	require.Equal(t, "Value2", second)
	require.Equal(t, first, reversedFirst)
	require.Equal(t, second, reversedSecond)
}

// TestGeneratedPackagePreservesExactGoNames checks that names produced by
// another Go generator are stored without changing their spelling.
func TestGeneratedPackagePreservesExactGoNames(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	types := mustClaimTestPackage(t, generation, "generated.local/gen/types")
	private := NewExactName(NameType, "api2HttpClient")
	handler := NewExactName(NameFunction, "_API2_HTTPHandler")
	require.NoError(t, types.DeclareName(private))
	require.NoError(t, types.DeclareName(handler))
	require.NoError(t, generation.Freeze())
	require.Equal(t, "api2HttpClient", private.Name())
	require.Equal(t, "_API2_HTTPHandler", handler.Name())
}

// TestGeneratedPackageRejectsInvalidExactGoName checks that an exact name
// must already be a valid Go identifier.
func TestGeneratedPackageRejectsInvalidExactGoName(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	types := mustClaimTestPackage(t, generation, "generated.local/gen/types")
	err := types.DeclareName(NewExactName(NameType, "not a name"))
	require.EqualError(t, err, `package name "not a name" is not a valid Go identifier`)
}

// TestDependentNameUsesFrozenBase verifies that companion declarations derive
// their spelling from the exact final name selected for their base declaration.
func TestDependentNameUsesFrozenBase(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	pkg := mustClaimTestPackage(t, generation, "generated.local/gen/types")
	require.NoError(t, pkg.DeclareName(NewExactName(NameType, "Result")))
	base := NewPreferredName(NameType, "Result", ExportedName, testNameOrder{value: "base"})
	require.NoError(t, pkg.DeclareName(base))
	validator, err := pkg.DeclareDependentName(
		NameFunction,
		base,
		"Validate",
		"",
		testNameOrder{value: "validator"},
	)
	require.NoError(t, err)

	require.NoError(t, generation.Freeze())
	require.Equal(t, "Result2", base.Name())
	require.Equal(t, "ValidateResult2", validator.Name())
}

// TestNameDeclarationRejectsUnownedPackageAccess verifies that only a package
// catalog can make internal declaration ownership available to typed records.
func TestNameDeclarationRejectsUnownedPackageAccess(t *testing.T) {
	declaration := NewExactName(NameType, "Value")
	require.Panics(t, func() { declaration.packagePath() })
}

// TestNameDeclarationRejectsEmptyPreferredName verifies that exact and
// suffixable declarations cannot mutate a package catalog without a Go name.
func TestNameDeclarationRejectsEmptyPreferredName(t *testing.T) {
	tests := []struct {
		name        string
		declaration *NameDeclaration
	}{
		{"exact", NewExactName(NameType, "")},
		{"preferred", NewPreferredName(NameFunction, "", ExportedName, testNameOrder{value: "empty"})},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			generation := mustTestGeneration(t, "generated.local/gen", nil)
			pkg := mustClaimTestPackage(t, generation, "generated.local/gen/types")

			err := pkg.DeclareName(test.declaration)
			require.ErrorContains(t, err, "name must not be empty")
			require.Nil(t, test.declaration.owner)
			require.Empty(t, pkg.names)
			require.Empty(t, pkg.exactNames)
		})
	}
}

// TestNameDeclarationRejectsInvalidKind verifies that direct and dependent
// declarations cannot mutate a package catalog with an unknown category.
func TestNameDeclarationRejectsInvalidKind(t *testing.T) {
	tests := []struct {
		name        string
		declaration func(*testing.T, *GeneratedPackage) *NameDeclaration
		wantNames   int
	}{
		{
			"exact",
			func(*testing.T, *GeneratedPackage) *NameDeclaration {
				return NewExactName(0, "Value")
			},
			0,
		},
		{
			"preferred",
			func(*testing.T, *GeneratedPackage) *NameDeclaration {
				return NewPreferredName(NameVariable+1, "Value", ExportedName, testNameOrder{value: "invalid"})
			},
			0,
		},
		{
			"dependent",
			func(t *testing.T, pkg *GeneratedPackage) *NameDeclaration {
				base := NewPreferredName(NameType, "Value", ExportedName, testNameOrder{value: "base"})
				require.NoError(t, pkg.DeclareName(base))
				return newDependentName(0, base, "New", "", testNameOrder{value: "dependent"})
			},
			1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			generation := mustTestGeneration(t, "generated.local/gen", nil)
			pkg := mustClaimTestPackage(t, generation, "generated.local/gen/types")
			declaration := test.declaration(t, pkg)

			err := pkg.DeclareName(declaration)
			require.ErrorContains(t, err, "invalid package name kind")
			require.Nil(t, declaration.owner)
			require.Len(t, pkg.names, test.wantNames)
		})
	}
}

// TestNameDeclarationPreferredOrder verifies that typed stable identity, not
// discovery order, decides suffix ownership and rejects an indistinguishable tie.
func TestNameDeclarationPreferredOrder(t *testing.T) {
	declare := func(reverse bool) (string, string) {
		generation := mustTestGeneration(t, "generated.local/gen", nil)
		pkg := mustClaimTestPackage(t, generation, "generated.local/gen/types")
		first := NewPreferredName(NameFunction, "Build", ExportedName, testNameOrder{value: "a"})
		second := NewPreferredName(NameConstant, "Build", ExportedName, testNameOrder{value: "b"})
		declarations := []*NameDeclaration{first, second}
		if reverse {
			declarations[0], declarations[1] = declarations[1], declarations[0]
		}
		for _, declaration := range declarations {
			require.NoError(t, pkg.DeclareName(declaration))
		}
		require.NoError(t, generation.Freeze())
		return first.Name(), second.Name()
	}

	first, second := declare(false)
	reversedFirst, reversedSecond := declare(true)
	require.Equal(t, "Build", first)
	require.Equal(t, "Build2", second)
	require.Equal(t, first, reversedFirst)
	require.Equal(t, second, reversedSecond)

	generation := mustTestGeneration(t, "generated.local/gen", nil)
	pkg := mustClaimTestPackage(t, generation, "generated.local/gen/types")
	require.NoError(t, pkg.DeclareName(NewPreferredName(
		NameFunction,
		"Build",
		ExportedName,
		testNameOrder{value: "same"},
	)))
	err := pkg.DeclareName(NewPreferredName(
		NameVariable,
		"Build",
		ExportedName,
		testNameOrder{value: "same"},
	))
	require.ErrorContains(t, err, "cannot deterministically order")
}

// TestPreferredNameVisibility verifies preferred declarations preserve their
// requested package visibility while sharing deterministic collision handling.
func TestPreferredNameVisibility(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	pkg := mustClaimTestPackage(t, generation, "generated.local/gen/types")
	exported := NewPreferredName(NameFunction, "build value", ExportedName, testNameOrder{value: "exported"})
	privateFirst := NewPreferredName(NameFunction, "Build Value", UnexportedName, testNameOrder{value: "private-a"})
	privateSecond := NewPreferredName(NameFunction, "build value", UnexportedName, testNameOrder{value: "private-b"})
	require.NoError(t, pkg.DeclareName(exported))
	require.NoError(t, pkg.DeclareName(privateSecond))
	require.NoError(t, pkg.DeclareName(privateFirst))
	require.NoError(t, generation.Freeze())
	require.Equal(t, "BuildValue", exported.Name())
	require.Equal(t, "buildValue", privateFirst.Name())
	require.Equal(t, "buildValue2", privateSecond.Name())
}

// TestNameDeclarationOrdersConcreteFamilies verifies that unrelated named
// order types never receive each other's values and remain discovery-order independent.
func TestNameDeclarationOrdersConcreteFamilies(t *testing.T) {
	declare := func(reverse bool) (string, string) {
		generation := mustTestGeneration(t, "generated.local/gen", nil)
		pkg := mustClaimTestPackage(t, generation, "generated.local/gen/types")
		alpha := NewPreferredName(NameFunction, "Build", ExportedName, alphaTestNameOrder("same"))
		omega := NewPreferredName(NameConstant, "Build", ExportedName, omegaTestNameOrder("same"))
		declarations := []*NameDeclaration{alpha, omega}
		if reverse {
			declarations[0], declarations[1] = declarations[1], declarations[0]
		}
		for _, declaration := range declarations {
			require.NoError(t, pkg.DeclareName(declaration))
		}
		require.NoError(t, generation.Freeze())
		return alpha.Name(), omega.Name()
	}

	alpha, omega := declare(false)
	reversedAlpha, reversedOmega := declare(true)
	require.Equal(t, "Build", alpha)
	require.Equal(t, "Build2", omega)
	require.Equal(t, alpha, reversedAlpha)
	require.Equal(t, omega, reversedOmega)
}

// TestNameDeclarationRejectsUnstableOrderTypes verifies that collection
// returns deterministic errors instead of accepting mutable or ambiguous order values.
func TestNameDeclarationRejectsUnstableOrderTypes(t *testing.T) {
	stable := testNameOrder{value: "stable"}
	tests := []struct {
		name  string
		order PackageNameOrder
	}{
		{"nil", nil},
		{"pointer", &stable},
		{"unnamed", struct{ PackageNameOrder }{PackageNameOrder: stable}},
		{"slice", unstableTestNameOrder{"mutable"}},
		{"pointer field", indirectTestNameOrder{value: new(string)}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			generation := mustTestGeneration(t, "generated.local/gen", nil)
			pkg := mustClaimTestPackage(t, generation, "generated.local/gen/types")
			declaration := NewPreferredName(NameFunction, "Build", ExportedName, test.order)
			err := pkg.DeclareName(declaration)
			require.ErrorContains(t, err, "stable concrete named value type")
		})
	}
}

// TestNameDeclarationRejectsDependentOrderTie verifies that dependency phase
// ordering cannot hide indistinguishable facts within one concrete family.
func TestNameDeclarationRejectsDependentOrderTie(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	pkg := mustClaimTestPackage(t, generation, "generated.local/gen/types")
	base := NewPreferredName(NameType, "Value", ExportedName, testNameOrder{value: "base"})
	first := newDependentName(NameFunction, base, "New", "First", testNameOrder{value: "same"})
	second := newDependentName(NameFunction, base, "New", "Second", testNameOrder{value: "same"})
	require.NoError(t, pkg.DeclareName(base))
	require.NoError(t, pkg.DeclareName(first))
	err := pkg.DeclareName(second)
	require.ErrorContains(t, err, "cannot deterministically order")
}

// TestNameDeclarationRejectsInvalidDependentOwners verifies that a dependent
// declaration derives its spelling only from a base already owned by its package.
func TestNameDeclarationRejectsInvalidDependentOwners(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	first := mustClaimTestPackage(t, generation, "generated.local/gen/first")
	second := mustClaimTestPackage(t, generation, "generated.local/gen/second")
	unowned := NewPreferredName(NameType, "Value", ExportedName, testNameOrder{value: "unowned"})
	dependent := newDependentName(NameFunction, unowned, "New", "", testNameOrder{value: "dependent"})
	err := first.DeclareName(dependent)
	require.ErrorContains(t, err, "base declaration is not owned")

	require.NoError(t, first.DeclareName(unowned))
	crossPackage := newDependentName(NameFunction, unowned, "New", "", testNameOrder{value: "cross-package"})
	err = second.DeclareName(crossPackage)
	require.ErrorContains(t, err, "base declaration belongs to generated package")
}

// TestNameDeclarationRejectsMultipleOwners verifies that one canonical name
// record cannot be rebound to another generated package.
func TestNameDeclarationRejectsMultipleOwners(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	declaration := NewExactName(NameType, "Value")
	require.NoError(t, mustClaimTestPackage(t, generation, "generated.local/gen/first").DeclareName(declaration))
	err := mustClaimTestPackage(t, generation, "generated.local/gen/second").DeclareName(declaration)
	require.ErrorContains(t, err, "already belongs")
}

// TestNameDeclarationRejectsSameImportAcrossGenerations verifies that
// canonical import spelling does not substitute for exact package ownership.
func TestNameDeclarationRejectsSameImportAcrossGenerations(t *testing.T) {
	declaration := NewExactName(NameType, "Value")
	first := mustClaimTestPackage(t, mustTestGeneration(t, "generated.local/gen", nil), "generated.local/gen/types")
	second := mustClaimTestPackage(t, mustTestGeneration(t, "generated.local/gen", nil), "generated.local/gen/types")
	require.NoError(t, first.DeclareName(declaration))
	err := second.DeclareName(declaration)
	require.ErrorContains(t, err, "already belongs")
}

// TestGenerationOwnsName checks declaration ownership before and after names
// are frozen without treating a matching package path as ownership.
func TestGenerationOwnsName(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	declaration := NewExactName(NameType, "Value")
	pkg := mustClaimTestPackage(t, generation, "generated.local/gen/types")
	require.NoError(t, pkg.DeclareName(declaration))

	foreignGeneration := mustTestGeneration(t, "generated.local/gen", nil)
	foreign := NewExactName(NameType, "Foreign")
	require.NoError(t, mustClaimTestPackage(t, foreignGeneration, "generated.local/gen/types").DeclareName(foreign))

	require.True(t, generation.OwnsName(declaration))
	require.False(t, generation.OwnsName(foreign))
	require.False(t, generation.OwnsName(NewExactName(NameType, "Unregistered")))
	require.False(t, generation.OwnsName(nil))
	require.True(t, pkg.OwnsName(declaration))
	require.False(t, pkg.OwnsName(foreign))
	require.False(t, pkg.OwnsName(nil))
	filePackage, ok := generation.PackageForFile("gen/types/value.go")
	require.True(t, ok)
	require.Same(t, pkg, filePackage)
	_, ok = generation.PackageForFile("gen/other/value.go")
	require.False(t, ok)

	require.NoError(t, generation.Freeze())
	require.True(t, generation.OwnsName(declaration))
}

// TestGeneratedOutputPathRejectsNormalizedCollisions verifies that equivalent
// import spellings cannot make two requested package identities share output.
func TestGeneratedOutputPathRejectsNormalizedCollisions(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/root/../gen", nil)
	first := mustClaimTestPackage(t, generation, "generated.local/gen/types")
	require.Equal(t, "generated.local/gen", generation.GenPkg())
	require.Equal(t, "generated.local/gen/types", first.ImportPath())
	require.Equal(t, "gen/types", first.OutputDirectory())
	_, err := generation.ClaimPackage("generated.local/gen/values/../types")
	require.EqualError(t, err,
		`generated package paths "generated.local/gen/types" and "generated.local/gen/values/../types" normalize to import path "generated.local/gen/types"`)
	require.Same(t, first, mustClaimTestPackage(t, generation, "generated.local/gen/types"))
	require.NoError(t, generation.Freeze())
}

// TestGeneratedOutputPathEmitsCanonicalImport verifies that a package keeps
// its exact planner claim for reuse while generated source sees a clean path.
func TestGeneratedOutputPathEmitsCanonicalImport(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	pkg := mustClaimTestPackage(t, generation, "generated.local/gen/values/../types")
	require.Same(t, pkg, mustClaimTestPackage(t, generation, "generated.local/gen/values/../types"))
	require.Equal(t, "generated.local/gen/types", pkg.ImportPath())
	require.Equal(t, "gen/types", pkg.OutputDirectory())
	require.NoError(t, generation.Freeze())
}

// TestGeneratedOutputPathRejectsLateNormalizedIdentity verifies that freeze
// does not let a new raw package identity reach an existing canonical package.
func TestGeneratedOutputPathRejectsLateNormalizedIdentity(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	first := mustClaimTestPackage(t, generation, "generated.local/gen/types")
	require.NoError(t, generation.Freeze())
	require.Same(t, first, generation.Package("generated.local/gen/types"))
	_, err := generation.ClaimPackage("generated.local/gen/types")
	require.ErrorContains(t, err, "cannot be claimed after generation freeze")
	_, err = generation.ClaimPackage("generated.local/gen/values/../types")
	require.ErrorContains(t, err, "cannot be claimed after generation freeze")
}

// TestGeneratedOutputPathRejectsPortableDirectoryCollision verifies that two
// import identities cannot claim the same directory on a case-insensitive host.
func TestGeneratedOutputPathRejectsPortableDirectoryCollision(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	first := mustClaimTestPackage(t, generation, "generated.local/gen/Foo")
	_, err := generation.ClaimPackage("generated.local/gen/foo")
	require.EqualError(t, err,
		`generated package paths "generated.local/gen/Foo" and "generated.local/gen/foo" resolve to output directory "gen/foo" on a case-insensitive filesystem`)
	require.Same(t, first, mustClaimTestPackage(t, generation, "generated.local/gen/Foo"))
}

// TestGeneratedOutputPathRejectsBackslashes verifies that invalid Go import
// separators are rejected instead of translated into another package identity.
func TestGeneratedOutputPathRejectsBackslashes(t *testing.T) {
	_, err := NewGeneration(`generated.local\gen`, nil)
	require.ErrorContains(t, err, "backslash")

	generation := mustTestGeneration(t, "generated.local/gen", nil)
	_, err = generation.ClaimPackage(`generated.local\gen\types`)
	require.ErrorContains(t, err, "contains a backslash")
}

// TestExplicitOutputPackageClaims verifies that packages outside GenPkg use
// the same canonical import and portable output ownership as ordinary claims.
func TestExplicitOutputPackageClaims(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	starter, err := generation.ClaimOutputPackage("generated.local", ".")
	require.NoError(t, err)
	reused, err := generation.ClaimOutputPackage("generated.local", "work/..")
	require.NoError(t, err)
	require.Same(t, starter, reused)
	require.Equal(t, "generated.local", starter.ImportPath())
	require.Equal(t, ".", starter.OutputDirectory())
	require.Same(t, starter, generation.Package("generated.local"))

	_, err = generation.ClaimOutputPackage("generated.local", "starter")
	require.ErrorContains(t, err, "already mapped")
	_, err = generation.ClaimOutputPackage("generated.local/../generated.local", ".")
	require.ErrorContains(t, err, "normalize to import path")
	require.NoError(t, generation.Freeze())
	_, err = generation.ClaimOutputPackage("generated.local/late", "late")
	require.ErrorContains(t, err, "after generation freeze")
}

// TestExplicitOutputPackageRejectsInvalidDirectories verifies that explicit
// output packages cannot escape the generation working directory or rely on
// host-specific path separators.
func TestExplicitOutputPackageRejectsInvalidDirectories(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	for _, directory := range []string{"../starter", "/starter", "C:/starter", `starter\service`} {
		_, err := generation.ClaimOutputPackage("generated.local/starter", directory)
		require.Error(t, err, directory)
	}
}

// TestExplicitOutputPackageSharesOrdinaryOwnership verifies that ordinary and
// explicit claims cannot assign one import path or portable directory twice.
func TestExplicitOutputPackageSharesOrdinaryOwnership(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	ordinary := mustClaimTestPackage(t, generation, "generated.local/gen/service")
	reused, err := generation.ClaimOutputPackage("generated.local/gen/service", ordinary.OutputDirectory())
	require.NoError(t, err)
	require.Same(t, ordinary, reused)

	_, err = generation.ClaimOutputPackage("generated.local/other", "gen/SERVICE")
	require.ErrorContains(t, err, "case-insensitive filesystem")
}

// TestGenerationRejectsImplicitLocalRoots verifies that only the exact local
// output sentinels are accepted as non-module generation roots.
func TestGenerationRejectsImplicitLocalRoots(t *testing.T) {
	for _, genpkg := range []string{"", "./", "//"} {
		_, err := NewGeneration(genpkg, nil)
		require.Error(t, err)
	}
	for _, genpkg := range []string{".", "/"} {
		_, err := NewGeneration(genpkg, nil)
		require.NoError(t, err)
	}
}

// TestGeneratedTypeFamiliesContainCanonicalNames verifies that existing type
// and union records expose the package-owned records used for rendering.
func TestGeneratedTypeFamiliesContainCanonicalNames(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	pkg := mustClaimTestPackage(t, generation, "generated.local/gen/types")
	user, err := pkg.DeclareUserType(generatedUserType("Widget", "widget"))
	require.NoError(t, err)
	union, alias := generatedUnionWithBranch("text")
	unionDeclaration, err := pkg.DeclareUnion(union)
	require.NoError(t, err)
	branchType, err := pkg.DeclareUnionBranchType(union, "text", alias)
	require.NoError(t, err)
	branch, err := pkg.UnionBranch(union, "text")
	require.NoError(t, err)

	require.Same(t, user.Declaration(), user.Declaration())
	require.Same(t, unionDeclaration.Declaration(), unionDeclaration.Declaration())
	require.Same(t, unionDeclaration.KindDeclaration(), unionDeclaration.KindDeclaration())
	require.Same(t, branchType.Declaration(), branchType.Declaration())
	require.Same(t, branch.KindDeclaration(), branch.KindDeclaration())
	require.Same(t, branch.ConstructorDeclaration(), branch.ConstructorDeclaration())
	require.Panics(t, func() { user.Declaration().Name() })
	require.Panics(t, func() { unionDeclaration.Declaration().Name() })

	require.NoError(t, generation.Freeze())
	require.Equal(t, user.Name(), user.Declaration().Name())
	require.Equal(t, unionDeclaration.Name(), unionDeclaration.Declaration().Name())
	require.Equal(t, unionDeclaration.KindName(), unionDeclaration.KindDeclaration().Name())
	require.Equal(t, branch.KindConst(), branch.KindDeclaration().Name())
	require.Equal(t, branch.Constructor(), branch.ConstructorDeclaration().Name())
}

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

	generation := mustTestGeneration(t, "generated.local/gen", []eval.Root{root})
	types := mustClaimTestPackage(t, generation, "generated.local/gen/types")
	_, err := types.DeclareUserType(first)
	require.NoError(t, err)
	_, err = types.DeclareUserType(second)
	require.ErrorContains(t, err, "foo_bar")
	require.ErrorContains(t, err, "FooBar")
	require.ErrorContains(t, err, "already declared by exact type")
}

// TestGenerationOwnsPackageRecords verifies that one generation returns one
// stable package record and scope for each output path.
func TestGenerationOwnsPackageRecords(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)

	first := mustClaimTestPackage(t, generation, "generated.local/gen/types")
	second := mustClaimTestPackage(t, generation, "generated.local/gen/types")
	other := mustClaimTestPackage(t, generation, "generated.local/gen/other")
	require.Same(t, first, second)
	require.Panics(t, func() {
		first.Scope()
	})
	require.NoError(t, generation.Freeze())
	require.Same(t, first.Scope(), second.Scope())
	require.NotSame(t, first, other)
	require.NotSame(t, first.Scope(), other.Scope())
}

// TestGenerationCopiesConstructionState verifies that callers cannot change
// root membership or the generated package path through constructor inputs or
// accessor results before or after freeze.
func TestGenerationCopiesConstructionState(t *testing.T) {
	first := RunDSL(t, func() {})
	second := RunDSL(t, func() {})
	roots := []eval.Root{first}
	generation := mustTestGeneration(t, "generated.local/gen", roots)

	roots[0] = second
	returnedRoots := generation.Roots()
	returnedRoots[0] = second
	require.Equal(t, "generated.local/gen", generation.GenPkg())
	require.True(t, generation.HasRoot(first))
	require.False(t, generation.HasRoot(second))

	require.NoError(t, generation.Freeze())
	roots[0] = nil
	returnedRoots = generation.Roots()
	returnedRoots[0] = second
	require.Equal(t, "generated.local/gen", generation.GenPkg())
	require.True(t, generation.HasRoot(first))
	require.False(t, generation.HasRoot(second))
}

// TestGeneratedPackageUserTypes verifies that a generated package records one
// declaration per user type and that lookups do not reserve names.
func TestGeneratedPackageUserTypes(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	types := mustClaimTestPackage(t, generation, "generated.local/gen/types")
	widget := generatedUserType("Widget", "widget")
	missing := generatedUserType("Missing", "missing")

	_, err := types.UserType(missing)
	require.ErrorContains(t, err, "not declared")

	first, err := types.DeclareUserType(widget)
	require.NoError(t, err)
	require.Panics(t, func() { first.Name() })
	require.Equal(t, "generated.local/gen/types", first.PackagePath())
	second, err := types.DeclareUserType(widget)
	require.NoError(t, err)
	require.Same(t, first, second)

	lookedUp, err := types.UserType(widget)
	require.NoError(t, err)
	require.Same(t, first, lookedUp)
	declaredMissing, err := types.DeclareUserType(missing)
	require.NoError(t, err)
	require.Panics(t, func() { declaredMissing.Name() })
	require.NoError(t, generation.Freeze())
	require.Equal(t, "Widget", first.Name())
	require.Equal(t, "Missing", declaredMissing.Name())
	require.Equal(t, "Widget", types.Scope().GoTypeName(&expr.AttributeExpr{Type: widget}))
}

// TestGeneratedPackageExactUserTypesDoNotMerge verifies that structural
// equality does not weaken the exact-name contract for DSL declarations.
func TestGeneratedPackageExactUserTypesDoNotMerge(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	types := mustClaimTestPackage(t, generation, "generated.local/gen/types")
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
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	types := mustClaimTestPackage(t, generation, "generated.local/gen/types")
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

// TestGeneratedPackageRepeatedUserTypesCompareCanonicalNames verifies that
// one origin may repeat an equivalent Go spelling but not change declarations.
func TestGeneratedPackageRepeatedUserTypesCompareCanonicalNames(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	types := mustClaimTestPackage(t, generation, "generated.local/gen/types")
	userType := generatedUserType("value-text", "value-text")
	expression := userType.(*expr.UserTypeExpr)
	first, err := types.DeclareUserType(userType)
	require.NoError(t, err)

	expression.TypeName = "value_text"
	second, err := types.DeclareUserType(userType)
	require.NoError(t, err)
	require.Same(t, first, second)

	expression.TypeName = "different"
	_, err = types.DeclareUserType(userType)
	require.ErrorContains(t, err, "cannot declare both")
}

// TestGeneratedPackageDerivedTypesUseTypedSourceIdentity verifies that view
// declarations rebuilt in the render phase select the records planned from
// the same exact source declaration.
func TestGeneratedPackageDerivedTypesUseTypedSourceIdentity(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	views := mustClaimTestPackage(t, generation, "generated.local/gen/service/views")
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
	require.Equal(t, "ValueView", projected.Name())
	require.Equal(t, "Value", viewed.Name())
}

// TestGeneratedPackageRepeatedDerivedTypesCompareCanonicalNames verifies that
// one typed identity accepts equivalent Go spellings but not another name.
func TestGeneratedPackageRepeatedDerivedTypesCompareCanonicalNames(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	views := mustClaimTestPackage(t, generation, "generated.local/gen/service/views")
	identity := NewProjectedTypeID(generatedUserType("Value", "value"))
	first, err := views.DeclareDerivedType(identity, "value-view")
	require.NoError(t, err)

	second, err := views.DeclareDerivedType(identity, "value_view")
	require.NoError(t, err)
	require.Same(t, first, second)

	_, err = views.DeclareDerivedType(identity, "different")
	require.ErrorContains(t, err, "cannot declare both")
}

// TestGeneratedPackageDerivedNamesIgnoreDeclarationOrder verifies that stable
// semantic source identifiers, not traversal order, decide suffix ownership.
func TestGeneratedPackageDerivedNamesIgnoreDeclarationOrder(t *testing.T) {
	declare := func(reverse bool) (string, string) {
		generation := mustTestGeneration(t, "generated.local/gen", nil)
		views := mustClaimTestPackage(t, generation, "generated.local/gen/service/views")
		first := generatedUserType("Value", "first")
		second := generatedUserType("Value", "second")
		ids := []DerivedTypeID{NewProjectedTypeID(first), NewProjectedTypeID(second)}
		if reverse {
			ids[0], ids[1] = ids[1], ids[0]
		}
		for _, identity := range ids {
			_, err := views.DeclareDerivedType(identity, "ValueView")
			require.NoError(t, err)
		}
		require.NoError(t, generation.Freeze())
		firstDeclaration, err := views.DerivedType(NewProjectedTypeID(first))
		require.NoError(t, err)
		secondDeclaration, err := views.DerivedType(NewProjectedTypeID(second))
		require.NoError(t, err)
		return firstDeclaration.Name(), secondDeclaration.Name()
	}

	first, second := declare(false)
	reversedFirst, reversedSecond := declare(true)
	require.Equal(t, first, reversedFirst)
	require.Equal(t, second, reversedSecond)
}

// TestGeneratedPackageRejectsAmbiguousDerivedOrder verifies that two distinct
// origins cannot rely on unstable expression shape to break an otherwise
// identical semantic ordering tuple.
func TestGeneratedPackageRejectsAmbiguousDerivedOrder(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	views := mustClaimTestPackage(t, generation, "generated.local/gen/service/views")
	first := generatedUserTypeOf("Value", "same", expr.String)
	second := generatedUserTypeOf("Value", "same", expr.Int)

	_, err := views.DeclareDerivedType(NewProjectedTypeID(first), "ValueView")
	require.NoError(t, err)
	_, err = views.DeclareDerivedType(NewProjectedTypeID(second), "ValueView")
	require.ErrorContains(t, err, "cannot deterministically order")
}

// TestGeneratedPackageUnionBranchesShareDeclaration verifies that copies of
// one authored OneOf reuse their generated branch alias.
func TestGeneratedPackageUnionBranchesShareDeclaration(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	types := mustClaimTestPackage(t, generation, "generated.local/gen/types")
	firstUnion, firstAlias := generatedUnionWithBranch("first")
	secondUnion := expr.DupAtt(firstUnion)
	secondAlias := secondUnion.Type.(*expr.Union).Values[0].Attribute.Type.(expr.UserType)

	_, err := types.DeclareUnion(firstUnion)
	require.NoError(t, err)
	firstDeclaration, err := types.DeclareUnionBranchType(firstUnion, "text", firstAlias)
	require.NoError(t, err)
	_, err = types.DeclareUnion(secondUnion)
	require.NoError(t, err)
	secondDeclaration, err := types.DeclareUnionBranchType(secondUnion, "text", secondAlias)
	require.NoError(t, err)
	require.Same(t, firstDeclaration, secondDeclaration)
	require.Panics(t, func() { firstDeclaration.Name() })

	require.NoError(t, generation.Freeze())
	require.Equal(t, "ValueBranchText", firstDeclaration.Name())
	lookedUp, err := types.UnionBranchType(secondUnion, "text")
	require.NoError(t, err)
	require.Same(t, firstDeclaration, lookedUp)
}

// TestGeneratedPackageUnionBranchesAreIsolatedByUnion verifies that branch
// aliases from separately named OneOf declarations never collapse together.
func TestGeneratedPackageUnionBranchesAreIsolatedByUnion(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	types := mustClaimTestPackage(t, generation, "generated.local/gen/types")
	firstUnion, firstAlias := generatedUnionWithBranch("first")
	secondUnion, secondAlias := generatedUnionWithBranch("second")
	secondUnion.Type.(*expr.Union).TypeName = "Other"
	secondAlias.(*expr.UserTypeExpr).TypeName = "OtherText"

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
	require.ElementsMatch(t, []string{"ValueBranchText", "OtherBranchText"}, []string{
		firstDeclaration.Name(),
		secondDeclaration.Name(),
	})
}

// TestGeneratedPackageUnionFamilyRejectsExactTypeNames verifies that no public
// OneOf companion silently receives a numeric suffix.
func TestGeneratedPackageUnionFamilyRejectsExactTypeNames(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	types := mustClaimTestPackage(t, generation, "generated.local/gen/types")
	for _, name := range []string{"ValueBranchText", "ValueKindText", "NewValueText"} {
		_, err := types.DeclareUserType(generatedUserType(name, name))
		require.NoError(t, err)
	}
	union, alias := generatedUnionWithBranch("text")
	_, err := types.DeclareUnion(union)
	require.ErrorContains(t, err, "set TypeName")
	require.ErrorContains(t, err, "ValueKindText")
	_, err = types.DeclareUnionBranchType(union, "text", alias)
	require.ErrorContains(t, err, "not declared")
}

// TestGeneratedPackageUnions verifies that authored identity makes copies
// idempotent and rejects an unrelated OneOf that asks for the same public name.
func TestGeneratedPackageUnions(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	types := mustClaimTestPackage(t, generation, "generated.local/gen/types")
	first := generatedUnion("type", "value")
	equivalent := expr.DupAtt(first)
	unrelated := generatedUnion("type", "value")
	different := generatedUnion("kind", "data")

	firstDeclaration, err := types.DeclareUnion(first)
	require.NoError(t, err)
	require.Panics(t, func() { firstDeclaration.Name() })
	require.Equal(t, "generated.local/gen/types", firstDeclaration.PackagePath())
	equivalentDeclaration, err := types.DeclareUnion(equivalent)
	require.NoError(t, err)
	require.Same(t, firstDeclaration, equivalentDeclaration)

	unrelatedDeclaration, err := types.DeclareUnion(unrelated)
	require.ErrorContains(t, err, "set TypeName")
	require.Nil(t, unrelatedDeclaration)

	differentDeclaration, err := types.DeclareUnion(different)
	require.ErrorContains(t, err, "set TypeName")
	require.Nil(t, differentDeclaration)

	lookedUp, err := types.Union(equivalent)
	require.NoError(t, err)
	require.Same(t, firstDeclaration, lookedUp)
	require.NoError(t, generation.Freeze())
	require.Equal(t, "Value", firstDeclaration.Name())
	require.Equal(t, "ValueKind", firstDeclaration.KindName())
}

// TestGeneratedPackageUnionAndUserTypeCollisionFailsRegardlessOfOrder verifies
// that traversal order never chooses a winner or creates a numeric suffix.
func TestGeneratedPackageUnionAndUserTypeCollisionFailsRegardlessOfOrder(t *testing.T) {
	for _, unionFirst := range []bool{true, false} {
		t.Run(fmt.Sprintf("union first %t", unionFirst), func(t *testing.T) {
			generation := mustTestGeneration(t, "generated.local/gen", nil)
			types := mustClaimTestPackage(t, generation, "generated.local/gen/types")
			userType := generatedUserType("Value", "value")
			kindUserType := generatedUserType("ValueKind", "value-kind")
			union := generatedUnion("type", "value")
			var (
				userDeclaration  *TypeDeclaration
				kindDeclaration  *TypeDeclaration
				unionDeclaration *UnionDeclaration
				err              error
			)
			if unionFirst {
				unionDeclaration, err = types.DeclareUnion(union)
				require.NoError(t, err)
				userDeclaration, err = types.DeclareUserType(userType)
				require.ErrorContains(t, err, "already declared")
			} else {
				userDeclaration, err = types.DeclareUserType(userType)
				require.NoError(t, err)
				kindDeclaration, err = types.DeclareUserType(kindUserType)
				require.NoError(t, err)
				require.Panics(t, func() {
					types.Scope().GoTypeName(&expr.AttributeExpr{Type: userType})
				})
				unionDeclaration, err = types.DeclareUnion(union)
				require.ErrorContains(t, err, "set TypeName")
			}
			if unionFirst {
				require.Nil(t, userDeclaration)
				require.Nil(t, kindDeclaration)
				require.NotNil(t, unionDeclaration)
			} else {
				require.NotNil(t, userDeclaration)
				require.NotNil(t, kindDeclaration)
				require.Nil(t, unionDeclaration)
			}
		})
	}
}

// TestGeneratedPackageLookupAcrossFreeze verifies that freeze keeps existing
// declarations readable and rejects every later declaration attempt.
func TestGeneratedPackageLookupAcrossFreeze(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	types := mustClaimTestPackage(t, generation, "generated.local/gen/types")
	widget := generatedUserType("Widget", "widget")
	union := generatedUnion("type", "value")
	userDeclaration, err := types.DeclareUserType(widget)
	require.NoError(t, err)
	unionDeclaration, err := types.DeclareUnion(union)
	require.NoError(t, err)
	require.Panics(t, func() { unionDeclaration.Name() })

	require.NoError(t, generation.Freeze())
	lookedUpUser, err := types.UserType(widget)
	require.NoError(t, err)
	require.Same(t, userDeclaration, lookedUpUser)
	lookedUpUnion, err := types.Union(union)
	require.NoError(t, err)
	require.Same(t, unionDeclaration, lookedUpUnion)
	require.Equal(t, "Value", lookedUpUnion.Name())
	require.Equal(t, "Widget", types.Scope().GoTypeName(&expr.AttributeExpr{Type: widget}))
	require.Equal(t, "Value", types.Scope().GoTypeName(union))
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

// TestGeneratedPackageRejectsConflictingOriginBindings verifies that exact
// and compiler-derived declarations cannot claim the same expression origin
// in either declaration order.
func TestGeneratedPackageRejectsConflictingOriginBindings(t *testing.T) {
	for _, derivedFirst := range []bool{false, true} {
		t.Run(fmt.Sprintf("derived first %t", derivedFirst), func(t *testing.T) {
			root := RunDSL(t, func() {
				dsl.Service("Values", func() {
					dsl.Method("Read", func() {
						dsl.Payload(func() {
							dsl.Attribute("value", dsl.String)
						})
					})
				})
			})
			generation := mustTestGeneration(t, "generated.local/gen", []eval.Root{root})
			types := mustClaimTestPackage(t, generation, "generated.local/gen/values")
			wrapper := root.Service("Values").Method("Read").Payload.Type.(expr.UserType)
			identity, ok := generation.NormalizedMethodType(wrapper)
			require.True(t, ok)

			if derivedFirst {
				_, _, err := types.DeclareMethodType(identity, wrapper)
				require.NoError(t, err)
				_, err = types.DeclareUserType(wrapper)
				require.ErrorContains(t, err, "already bound")
				return
			}

			_, err := types.DeclareUserType(wrapper)
			require.NoError(t, err)
			_, _, err = types.DeclareMethodType(identity, wrapper)
			require.ErrorContains(t, err, "already bound")
		})
	}
}

// TestGenerationOwnsNormalizedWrapper verifies that normalization
// and declaration planning share the same closed method-role identity.
func TestGenerationOwnsNormalizedWrapper(t *testing.T) {
	root := RunDSL(t, func() {
		dsl.Service("Values", func() {
			dsl.Method("Read", func() {
				dsl.Payload(func() {
					dsl.Attribute("value", dsl.String)
				})
			})
		})
	})
	generation := mustTestGeneration(t, "generated.local/gen", []eval.Root{root})
	wrapper := root.Service("Values").Method("Read").Payload.Type.(expr.UserType)
	identity, ok := generation.NormalizedMethodType(wrapper)

	require.True(t, ok)
	require.Equal(t, "ReadPayload", identity.Name())
	require.Equal(t, wrapper.ID(), identity.UID())
}

// TestGenerationAssignsExactMethodOwners verifies that every raw object role
// becomes a generated wrapper whose declaration and example identities agree.
func TestGenerationAssignsExactMethodOwners(t *testing.T) {
	service := &expr.ServiceExpr{Name: "Values"}
	method := &expr.MethodExpr{
		Name:             "Stream",
		Service:          service,
		Payload:          &expr.AttributeExpr{Type: &expr.Object{}},
		StreamingPayload: &expr.AttributeExpr{Type: &expr.Object{}},
		Result:           &expr.AttributeExpr{Type: &expr.Object{}},
		StreamingResult:  &expr.AttributeExpr{Type: &expr.Object{}},
	}
	service.Methods = []*expr.MethodExpr{method}
	root := &expr.RootExpr{Services: []*expr.ServiceExpr{service}}
	generation := mustTestGeneration(t, "generated.local/gen", []eval.Root{root})

	cases := []struct {
		name      string
		attribute *expr.AttributeExpr
		expected  expr.ExampleIdentity
	}{
		{"payload", method.Payload, expr.MethodPayloadExampleIdentity(method)},
		{"streaming payload", method.StreamingPayload, expr.MethodStreamingPayloadExampleIdentity(method)},
		{"result", method.Result, expr.MethodResultExampleIdentity(method)},
		{"streaming result", method.StreamingResult, expr.MethodStreamingResultExampleIdentity(method)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wrapper := tc.attribute.Type.(expr.UserType)
			exampleIdentity, generated := expr.GeneratedUserTypeExampleIdentity(wrapper)
			require.True(t, generated)
			require.Equal(t, tc.expected, exampleIdentity)
			declarationIdentity, normalized := generation.NormalizedMethodType(wrapper)
			require.True(t, normalized)
			require.Equal(t, wrapper.ID(), declarationIdentity.UID())
		})
	}
}

// TestMethodTypeIdentityPreservesRawOwner proves semantic wrapper identity does
// not collapse distinct DSL names that share one preferred Go spelling.
func TestMethodTypeIdentityPreservesRawOwner(t *testing.T) {
	firstMethod := &expr.MethodExpr{Name: "foo-bar", Service: &expr.ServiceExpr{Name: "Values"}}
	secondMethod := &expr.MethodExpr{Name: "foo_bar", Service: &expr.ServiceExpr{Name: "Values"}}
	cases := []struct {
		name   string
		kind   derivedTypeKind
		first  expr.ExampleIdentity
		second expr.ExampleIdentity
	}{
		{"payload", methodPayloadTypeKind, expr.MethodPayloadExampleIdentity(firstMethod), expr.MethodPayloadExampleIdentity(secondMethod)},
		{"streaming payload", methodStreamingPayloadTypeKind, expr.MethodStreamingPayloadExampleIdentity(firstMethod), expr.MethodStreamingPayloadExampleIdentity(secondMethod)},
		{"result", methodResultTypeKind, expr.MethodResultExampleIdentity(firstMethod), expr.MethodResultExampleIdentity(secondMethod)},
		{"streaming result", methodStreamingResultTypeKind, expr.MethodStreamingResultExampleIdentity(firstMethod), expr.MethodStreamingResultExampleIdentity(secondMethod)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			first := newMethodTypeIdentity("test api", firstMethod.Name, tc.kind, tc.first)
			second := newMethodTypeIdentity("test api", secondMethod.Name, tc.kind, tc.second)

			require.Equal(t, first.Name(), second.Name())
			require.NotEqual(t, first.UID(), second.UID())
			require.Equal(t, first.UID(), newMethodTypeIdentity("test api", firstMethod.Name, tc.kind, tc.first).UID())
		})
	}
}

// TestMethodTypeNamesUseAPIToBreakTies verifies two APIs can write equal
// service and method wrapper names to one package without using input order.
func TestMethodTypeNamesUseAPIToBreakTies(t *testing.T) {
	forwardFirst, forwardSecond := methodTypeNamesByAPI(t, false)
	reverseFirst, reverseSecond := methodTypeNamesByAPI(t, true)

	require.Equal(t, "ReadPayload", forwardFirst)
	require.Equal(t, "ReadPayload2", forwardSecond)
	require.Equal(t, forwardFirst, reverseFirst)
	require.Equal(t, forwardSecond, reverseSecond)
}

// TestGenerationPreservesRawMethodOwner proves synthesized wrappers retain
// the raw method identity even when their preferred generated names coincide.
func TestGenerationPreservesRawMethodOwner(t *testing.T) {
	root := RunDSL(t, func() {
		dsl.Service("Values", func() {
			dsl.Method("foo-bar", func() {
				dsl.Payload(func() {
					dsl.Attribute("first", dsl.String)
				})
			})
			dsl.Method("foo_bar", func() {
				dsl.Payload(func() {
					dsl.Attribute("second", dsl.String)
				})
			})
		})
	})
	generation := mustTestGeneration(t, "generated.local/gen", []eval.Root{root})
	first := root.Service("Values").Method("foo-bar").Payload.Type.(expr.UserType)
	second := root.Service("Values").Method("foo_bar").Payload.Type.(expr.UserType)
	firstIdentity, firstGenerated := generation.NormalizedMethodType(first)
	secondIdentity, secondGenerated := generation.NormalizedMethodType(second)

	require.True(t, firstGenerated)
	require.True(t, secondGenerated)
	require.Equal(t, first.Name(), second.Name())
	require.NotEqual(t, first.ID(), second.ID())
	require.Equal(t, first.ID(), firstIdentity.UID())
	require.Equal(t, second.ID(), secondIdentity.UID())
}

// TestGenerationOwnsNormalizedMethodProvenance proves only wrappers created by
// this generation are classified as compiler-owned method types. Authored text
// that equals a synthesized UID is still an authored declaration.
func TestGenerationOwnsNormalizedMethodProvenance(t *testing.T) {
	authoredMethod := &expr.MethodExpr{Name: "Authored", Service: &expr.ServiceExpr{Name: "Values"}}
	authoredUID := "generated:" + expr.MethodPayloadExampleIdentity(authoredMethod).Seed()
	root := RunDSL(t, func() {
		authored := dsl.Type(authoredUID, func() {
			dsl.Attribute("authored", dsl.String)
		})
		dsl.Service("Values", func() {
			dsl.Method("Authored", func() {
				dsl.Payload(authored)
			})
			dsl.Method("Raw", func() {
				dsl.Payload(func() {
					dsl.Attribute("raw", dsl.String)
				})
			})
		})
	})
	generation, err := NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)

	authored := root.Service("Values").Method("Authored").Payload.Type.(expr.UserType)
	_, ok := generation.NormalizedMethodType(authored)
	require.False(t, ok)
	raw := root.Service("Values").Method("Raw").Payload.Type.(expr.UserType)
	rawIdentity, ok := generation.NormalizedMethodType(raw)
	require.True(t, ok)
	require.Equal(t, raw.ID(), rawIdentity.UID())
}

// TestGenerationRecoversNormalizedMethodProvenance verifies that constructing
// another generation over the same evaluated root recognizes the exact typed
// wrapper instead of parsing its generated name or ID.
func TestGenerationRecoversNormalizedMethodProvenance(t *testing.T) {
	root := RunDSL(t, func() {
		dsl.Service("Values", func() {
			dsl.Method("Read", func() {
				dsl.Payload(func() {
					dsl.Attribute("value", dsl.String)
				})
			})
		})
	})
	first := mustTestGeneration(t, "generated.local/gen", []eval.Root{root})
	wrapper := root.Service("Values").Method("Read").Payload.Type.(expr.UserType)
	firstIdentity, ok := first.NormalizedMethodType(wrapper)
	require.True(t, ok)

	second := mustTestGeneration(t, "generated.local/gen", []eval.Root{root})
	secondIdentity, ok := second.NormalizedMethodType(wrapper)
	require.True(t, ok)
	require.Equal(t, firstIdentity.UID(), secondIdentity.UID())
}

// TestGenerationCatalogsAreIsolated verifies that standalone generation runs
// do not share declaration records or name reservations.
func TestGenerationCatalogsAreIsolated(t *testing.T) {
	firstGeneration := mustTestGeneration(t, "generated.local/gen", nil)
	first := mustClaimTestPackage(t, firstGeneration, "generated.local/gen/types")
	secondGeneration := mustTestGeneration(t, "generated.local/gen", nil)
	second := mustClaimTestPackage(t, secondGeneration, "generated.local/gen/types")
	firstUnion := generatedUnion("type", "value")
	secondUnion := generatedUnion("type", "value")

	firstDeclaration, err := first.DeclareUnion(firstUnion)
	require.NoError(t, err)
	secondDeclaration, err := second.DeclareUnion(secondUnion)
	require.NoError(t, err)
	require.NoError(t, firstGeneration.Freeze())
	require.NoError(t, secondGeneration.Freeze())
	require.Equal(t, "Value", firstDeclaration.Name())
	require.Equal(t, "Value", secondDeclaration.Name())
	require.NotSame(t, firstDeclaration, secondDeclaration)
	require.NotSame(t, first.Scope(), second.Scope())
}

// methodTypeNamesByAPI declares equal method wrappers in the requested order
// and returns the name assigned to each API.
func methodTypeNamesByAPI(t *testing.T, reverse bool) (string, string) {
	t.Helper()
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	generatedPackage := mustClaimTestPackage(t, generation, "generated.local/gen/shared")
	method := &expr.MethodExpr{Name: "Read", Service: &expr.ServiceExpr{Name: "Shared"}}
	example := expr.MethodPayloadExampleIdentity(method)
	firstIdentity, firstWrapper := testMethodTypeWrapper("first api", method.Name, example)
	secondIdentity, secondWrapper := testMethodTypeWrapper("second api", method.Name, example)
	var first, second *TypeDeclaration
	if reverse {
		second = declareTestMethodType(t, generatedPackage, secondIdentity, secondWrapper)
		first = declareTestMethodType(t, generatedPackage, firstIdentity, firstWrapper)
	} else {
		first = declareTestMethodType(t, generatedPackage, firstIdentity, firstWrapper)
		second = declareTestMethodType(t, generatedPackage, secondIdentity, secondWrapper)
	}
	require.NoError(t, generation.Freeze())
	return first.Name(), second.Name()
}

// testMethodTypeWrapper creates one generated method wrapper with an API name
// that is used only to order equal wrapper names.
func testMethodTypeWrapper(api, method string, example expr.ExampleIdentity) (MethodTypeIdentity, expr.UserType) {
	identity := newMethodTypeIdentity(api, method, methodPayloadTypeKind, example)
	wrapper := expr.NewGeneratedUserType(
		identity.Name(),
		&expr.AttributeExpr{Type: &expr.Object{
			{Name: "value", Attribute: &expr.AttributeExpr{Type: expr.String}},
		}},
		example,
	)
	return identity.bind(wrapper), wrapper
}

// declareTestMethodType submits one generated wrapper and returns its stored
// type declaration.
func declareTestMethodType(t *testing.T, generatedPackage *GeneratedPackage, identity MethodTypeIdentity, wrapper expr.UserType) *TypeDeclaration {
	t.Helper()
	declaration, _, err := generatedPackage.DeclareMethodType(identity, wrapper)
	require.NoError(t, err)
	return declaration
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
func generatedUnion(typeKey, valueKey string) *expr.AttributeExpr {
	return &expr.AttributeExpr{Type: &expr.Union{
		TypeName: "Value",
		TypeKey:  typeKey,
		ValueKey: valueKey,
	}}
}

// generatedUnionWithBranch builds a union with one generated branch alias.
func generatedUnionWithBranch(aliasID string) (*expr.AttributeExpr, expr.UserType) {
	alias := generatedUserTypeOf("ValueText", aliasID, expr.String)
	return &expr.AttributeExpr{Type: &expr.Union{
		TypeName: "Value",
		Values: []*expr.NamedAttributeExpr{{
			Name:      "text",
			Attribute: &expr.AttributeExpr{Type: alias},
		}},
	}}, alias
}

// mustTestGeneration creates a generation for tests whose package root is
// known to be valid.
func mustTestGeneration(t *testing.T, genpkg string, roots []eval.Root) *Generation {
	t.Helper()
	generation, err := NewGeneration(genpkg, roots)
	require.NoError(t, err)
	return generation
}

// mustClaimTestPackage claims a package for tests whose planner path is known
// to be valid and unique.
func mustClaimTestPackage(t *testing.T, generation *Generation, path string) *GeneratedPackage {
	t.Helper()
	generatedPackage, err := generation.ClaimPackage(path)
	require.NoError(t, err)
	return generatedPackage
}
