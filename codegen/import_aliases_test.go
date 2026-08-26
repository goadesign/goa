// This file verifies that one generation assigns import qualifiers by explicit
// ownership priority rather than declaration order or import-path sorting.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestImportAliasPrioritiesIgnoreRegistrationOrder verifies that static
// template imports keep required qualifiers ahead of generated packages and
// design metadata regardless of planning order.
func TestImportAliasPrioritiesIgnoreRegistrationOrder(t *testing.T) {
	freeze := func(reverse bool) map[string]string {
		generation := mustTestGeneration(t, "generated.local/gen", nil)
		pkg, err := generation.ClaimPackage("generated.local/gen/service")
		require.NoError(t, err)
		declare := []func() error{
			func() error {
				return pkg.RequireImport(NewImport("goa", "goa.design/goa/v3/pkg"))
			},
			func() error {
				return pkg.ReserveGeneratedImport(NewImport("goa", "generated.local/gen/goa"))
			},
			func() error {
				return pkg.DeclareImport(NewImport("goa", "example.com/custom/goa"))
			},
		}
		if reverse {
			declare[0], declare[2] = declare[2], declare[0]
		}
		for _, register := range declare {
			require.NoError(t, register())
		}
		require.NoError(t, generation.Freeze())
		return map[string]string{
			"fixed":     pkg.ImportName("goa.design/goa/v3/pkg"),
			"generated": pkg.ImportName("generated.local/gen/goa"),
			"metadata":  pkg.ImportName("example.com/custom/goa"),
		}
	}

	want := map[string]string{
		"fixed":     "goa",
		"generated": "goa2",
		"metadata":  "goa3",
	}
	require.Equal(t, want, freeze(false))
	require.Equal(t, want, freeze(true))
}

// TestImportAliasHighestPriorityWinsPerPath verifies that one complete import
// path has one identity and uses its highest-priority requested spelling.
func TestImportAliasHighestPriorityWinsPerPath(t *testing.T) {
	freeze := func(reverse bool) string {
		generation := mustTestGeneration(t, "generated.local/gen", nil)
		pkg, err := generation.ClaimPackage("generated.local/gen/service")
		require.NoError(t, err)
		declare := []func() error{
			func() error {
				return pkg.RequireImport(NewImport("json", "encoding/json"))
			},
			func() error {
				return pkg.ReserveGeneratedImport(NewImport("jason", "encoding/json"))
			},
			func() error {
				return pkg.DeclareImport(NewImport("jsonp", "encoding/json"))
			},
		}
		if reverse {
			declare[0], declare[2] = declare[2], declare[0]
		}
		for _, register := range declare {
			require.NoError(t, register())
		}
		require.NoError(t, generation.Freeze())
		return pkg.ImportName("encoding/json")
	}

	require.Equal(t, "json", freeze(false))
	require.Equal(t, "json", freeze(true))
}

// TestGeneratedImportPreferenceIsOrderIndependent verifies that repeated
// generated-package preferences for one path use deterministic spelling.
func TestGeneratedImportPreferenceIsOrderIndependent(t *testing.T) {
	freeze := func(first, second string) string {
		generation := mustTestGeneration(t, "generated.local/gen", nil)
		pkg, err := generation.ClaimPackage("generated.local/gen/service")
		require.NoError(t, err)
		require.NoError(t, pkg.ReserveGeneratedImport(NewImport(first, "generated.local/gen/value")))
		require.NoError(t, pkg.ReserveGeneratedImport(NewImport(second, "generated.local/gen/value")))
		require.NoError(t, generation.Freeze())
		return pkg.ImportName("generated.local/gen/value")
	}

	require.Equal(t, freeze("alpha", "zeta"), freeze("zeta", "alpha"))
}

// TestImportAliasRejectsIncompatibleFixedRequirements verifies that static
// templates cannot request two different mandatory spellings for one path.
func TestImportAliasRejectsIncompatibleFixedRequirements(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	pkg, err := generation.ClaimPackage("generated.local/gen/service")
	require.NoError(t, err)
	require.NoError(t, pkg.RequireImport(NewImport("json", "encoding/json")))
	require.ErrorContains(
		t,
		pkg.RequireImport(NewImport("jason", "encoding/json")),
		"requires qualifier",
	)
}

// TestImportAliasRejectsFixedQualifierCollision verifies that two static
// packages cannot both require the same qualifier.
func TestImportAliasRejectsFixedQualifierCollision(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	pkg, err := generation.ClaimPackage("generated.local/gen/service")
	require.NoError(t, err)
	require.NoError(t, pkg.RequireImport(NewImport("runtime", "example.com/first")))
	require.NoError(t, pkg.RequireImport(NewImport("runtime", "example.com/second")))
	require.ErrorContains(t, generation.Freeze(), "required by both")
}

// TestImportAliasesAreIndependentAcrossOutputPackages verifies that packages
// which never compile together may use the same natural import name.
func TestImportAliasesAreIndependentAcrossOutputPackages(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	httpPackage, err := generation.ClaimPackage("generated.local/gen/http/cli/calc")
	require.NoError(t, err)
	grpcPackage, err := generation.ClaimPackage("generated.local/gen/grpc/cli/calc")
	require.NoError(t, err)
	require.NoError(t, httpPackage.ReserveGeneratedImport(NewImport(
		"calcc",
		"generated.local/gen/http/calc/client",
	)))
	require.NoError(t, grpcPackage.ReserveGeneratedImport(NewImport(
		"calcc",
		"generated.local/gen/grpc/calc/client",
	)))
	require.NoError(t, generation.Freeze())
	require.Equal(t, "calcc", httpPackage.ImportName("generated.local/gen/http/calc/client"))
	require.Equal(t, "calcc", grpcPackage.ImportName("generated.local/gen/grpc/calc/client"))
}

// TestImportAliasSharesPackageDeclarationScope verifies that an import and a
// package declaration cannot receive the same Go name.
func TestImportAliasSharesPackageDeclarationScope(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	pkg := mustClaimTestPackage(t, generation, "generated.local/gen/service")
	require.NoError(t, pkg.ReserveGeneratedImport(NewImport(
		"inspectPayloadFieldDescs",
		"example.com/inspect-payload-field-descs",
	)))
	declaration := NewPreferredName(
		NameVariable,
		"inspectPayloadFieldDescs",
		UnexportedName,
		testNameOrder{value: "inspect-payload-field-descs"},
	)
	require.NoError(t, pkg.DeclareName(declaration))

	require.NoError(t, generation.Freeze())
	require.Equal(t, "inspectPayloadFieldDescs", pkg.ImportName("example.com/inspect-payload-field-descs"))
	require.Equal(t, "inspectPayloadFieldDescs2", declaration.Name())
}

// TestGeneratedImportYieldsToExactPackageDeclaration verifies that a package
// declaration which must keep its spelling is chosen before a renamable import.
func TestGeneratedImportYieldsToExactPackageDeclaration(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	pkg := mustClaimTestPackage(t, generation, "generated.local/gen/service")
	require.NoError(t, pkg.ReserveGeneratedImport(NewImport("widget", "example.com/widget")))
	declaration := NewExactName(NameVariable, "widget")
	require.NoError(t, pkg.DeclareName(declaration))

	require.NoError(t, generation.Freeze())
	require.Equal(t, "widget", declaration.Name())
	require.Equal(t, "widget2", pkg.ImportName("example.com/widget"))
}
