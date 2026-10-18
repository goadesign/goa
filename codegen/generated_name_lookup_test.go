// This file checks that faster order lookup preserves declaration rejection,
// lookup bindings, discovery order, and separation between generated packages.
package codegen

import (
	"cmp"
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

type (
	// lookupTestNameOrder compares rank while allowing distinct representations
	// of the same ordering facts.
	lookupTestNameOrder struct {
		rank           int
		representation string
	}

	// lookupTestFloatOrder accepts immutable floating-point ordering facts,
	// including NaN values whose Go equality differs from comparison equality.
	lookupTestFloatOrder float64
)

func (o lookupTestNameOrder) ComparePackageName(other PackageNameOrder) int {
	return cmp.Compare(o.rank, other.(lookupTestNameOrder).rank)
}

func (o lookupTestFloatOrder) ComparePackageName(other PackageNameOrder) int {
	return cmp.Compare(o, other.(lookupTestFloatOrder))
}

func TestDeclareNameRejectsComparatorEquivalentRepresentations(t *testing.T) {
	tests := []struct {
		name   string
		first  PackageNameOrder
		second PackageNameOrder
	}{
		{"unused facts", lookupTestNameOrder{3, "first"}, lookupTestNameOrder{3, "second"}},
		{"NaN", lookupTestFloatOrder(math.NaN()), lookupTestFloatOrder(math.Float64frombits(0x7ff8000000000002))},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, reverse := range []bool{false, true} {
				pkg := newGeneratedPackage("generated.local/gen/types", "generated.local/gen/types", "types")
				orders := []PackageNameOrder{test.first, test.second}
				if reverse {
					orders[0], orders[1] = orders[1], orders[0]
				}
				require.NoError(t, pkg.DeclareName(NewPreferredName(NameFunction, "First", ExportedName, orders[0])))
				rejected := NewPreferredName(NameVariable, "Second", ExportedName, orders[1])
				require.EqualError(t, pkg.DeclareName(rejected),
					`generated package "generated.local/gen/types" cannot deterministically order preferred variable "Second"`)
				require.Nil(t, rejected.owner)
				require.Len(t, pkg.names, 1)
			}
		})
	}
}

func TestDeclareNameRejectsTiesAcrossDependentAndIndependentNames(t *testing.T) {
	for _, firstDependent := range []bool{false, true} {
		for _, secondDependent := range []bool{false, true} {
			t.Run(fmt.Sprintf("dependent=%t,%t", firstDependent, secondDependent), func(t *testing.T) {
				pkg := newGeneratedPackage("generated.local/gen/types", "generated.local/gen/types", "types")
				base := NewExactName(NameType, "Base")
				require.NoError(t, pkg.DeclareName(base))
				first := NewPreferredName(NameFunction, "First", ExportedName, lookupTestNameOrder{3, "first"})
				second := NewPreferredName(NameConstant, "Second", ExportedName, lookupTestNameOrder{3, "second"})
				if firstDependent {
					first = newDependentName(NameFunction, base, "New", "First", first.order)
				}
				if secondDependent {
					second = newDependentName(NameConstant, base, "New", "Second", second.order)
				}
				lower := NewPreferredName(NameType, "Lower", ExportedName, lookupTestNameOrder{1, "lower"})
				upper := NewPreferredName(NameType, "Upper", ExportedName, lookupTestNameOrder{5, "upper"})
				require.NoError(t, pkg.DeclareName(upper))
				require.NoError(t, pkg.DeclareName(lower))
				require.NoError(t, pkg.DeclareName(first))
				require.ErrorContains(t, pkg.DeclareName(second), "cannot deterministically order")
				require.Nil(t, second.owner)
				require.Equal(t, []*NameDeclaration{base, upper, lower, first}, pkg.names)
				require.NoError(t, pkg.freeze())
				require.Equal(t, "Upper", upper.Name())
				require.Equal(t, "Lower", lower.Name())
			})
		}
	}
}

func TestDeclareNameRepeatedDeclarationAddsLookupKeys(t *testing.T) {
	pkg := newGeneratedPackage("generated.local/gen/types", "generated.local/gen/types", "types")
	declaration := NewPreferredName(NameFunction, "Build", ExportedName, lookupTestNameOrder{1, "first"})
	first, second := testNameKey("first"), testNameKey("second")
	require.NoError(t, pkg.DeclareName(declaration, first))
	require.NoError(t, pkg.DeclareName(declaration, second, first))
	require.Len(t, pkg.names, 1)
	require.NoError(t, pkg.freeze())
	require.Equal(t, "Build", pkg.Scope().HashedUnique(first, "Ignored"))
	require.Equal(t, "Build", pkg.Scope().HashedUnique(second, "Ignored"))
	require.Nil(t, pkg.nameOrders)
	require.ErrorContains(t, pkg.DeclareName(declaration), "frozen")
}

func TestDeclareNameRejectedInputDoesNotReserveOrder(t *testing.T) {
	tests := []struct {
		name string
		fail func(*testing.T, *GeneratedPackage, PackageNameOrder) error
		want string
	}{
		{
			"kind",
			func(t *testing.T, pkg *GeneratedPackage, order PackageNameOrder) error {
				return pkg.DeclareName(NewPreferredName(0, "Invalid", ExportedName, order))
			},
			"invalid package name kind",
		},
		{
			"lookup key",
			func(t *testing.T, pkg *GeneratedPackage, order PackageNameOrder) error {
				key := testNameKey("reserved")
				require.NoError(t, pkg.DeclareName(NewExactName(NameType, "Existing"), key))
				return pkg.DeclareName(NewPreferredName(NameFunction, "Invalid", ExportedName, order), key)
			},
			"lookup key",
		},
		{
			"unowned base",
			func(t *testing.T, pkg *GeneratedPackage, order PackageNameOrder) error {
				base := NewExactName(NameType, "Base")
				return pkg.DeclareName(newDependentName(NameFunction, base, "New", "", order))
			},
			"base declaration is not owned",
		},
		{
			"foreign owner",
			func(t *testing.T, pkg *GeneratedPackage, order PackageNameOrder) error {
				foreign := newGeneratedPackage("generated.local/gen/foreign", "generated.local/gen/foreign", "foreign")
				declaration := NewPreferredName(NameFunction, "Invalid", ExportedName, order)
				require.NoError(t, foreign.DeclareName(declaration))
				return pkg.DeclareName(declaration)
			},
			"already belongs",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pkg := newGeneratedPackage("generated.local/gen/types", "generated.local/gen/types", "types")
			order := lookupTestNameOrder{1, "same"}
			require.ErrorContains(t, test.fail(t, pkg, order), test.want)
			require.Empty(t, pkg.nameOrders)
			valid := NewPreferredName(NameFunction, "Valid", ExportedName, order)
			require.NoError(t, pkg.DeclareName(valid))
			require.NoError(t, pkg.freeze())
			require.Equal(t, "Valid", valid.Name())
		})
	}
}

// TestDeclareNameOrderLookupSurvivesFailedFreeze checks that an import conflict
// leaves collection open and equal ordering facts still return the same error.
func TestDeclareNameOrderLookupSurvivesFailedFreeze(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	pkg := mustClaimTestPackage(t, generation, "generated.local/gen/types")
	first := NewPreferredName(NameFunction, "First", ExportedName, lookupTestNameOrder{1, "first"})
	require.NoError(t, pkg.DeclareName(first))
	require.NoError(t, pkg.RequireImport(NewImport("runtime", "example.com/first")))
	require.NoError(t, pkg.RequireImport(NewImport("runtime", "example.com/second")))

	require.ErrorContains(t, generation.Freeze(), "required by both")
	require.False(t, pkg.frozen)
	second := NewPreferredName(NameVariable, "Second", ExportedName, lookupTestNameOrder{1, "second"})
	require.EqualError(t, pkg.DeclareName(second),
		`generated package "generated.local/gen/types" cannot deterministically order preferred variable "Second"`)
	require.Nil(t, second.owner)
	require.Equal(t, []*NameDeclaration{first}, pkg.names)
}

func TestDeclareNameOrderLookupIsPackageAndRunLocal(t *testing.T) {
	for range 2 {
		generation := mustTestGeneration(t, "generated.local/gen", nil)
		for _, path := range []string{"generated.local/gen/first", "generated.local/gen/second"} {
			pkg := mustClaimTestPackage(t, generation, path)
			first := NewPreferredName(NameFunction, "Build", ExportedName, lookupTestNameOrder{1, "same"})
			otherFamily := NewPreferredName(NameConstant, "Build", ExportedName, alphaTestNameOrder("same"))
			require.NoError(t, pkg.DeclareName(first))
			require.NoError(t, pkg.DeclareName(otherFamily))
		}
		require.NoError(t, generation.Freeze())
		for _, path := range []string{"generated.local/gen/first", "generated.local/gen/second"} {
			pkg := generation.Package(path)
			require.Equal(t, "Build2", pkg.names[0].Name())
			require.Equal(t, "Build", pkg.names[1].Name())
			require.Nil(t, pkg.nameOrders)
		}
	}
}
