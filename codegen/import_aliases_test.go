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
		generation := NewGeneration("generated.local/gen", nil)
		declare := []func() error{
			func() error {
				return generation.RequireImport(NewImport("goa", "goa.design/goa/v3/pkg"))
			},
			func() error {
				return generation.ReserveGeneratedImport(NewImport("goa", "generated.local/gen/goa"))
			},
			func() error {
				return generation.DeclareImport(NewImport("goa", "example.com/custom/goa"))
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
			"fixed":     generation.ImportName("goa.design/goa/v3/pkg"),
			"generated": generation.ImportName("generated.local/gen/goa"),
			"metadata":  generation.ImportName("example.com/custom/goa"),
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
		generation := NewGeneration("generated.local/gen", nil)
		declare := []func() error{
			func() error {
				return generation.RequireImport(NewImport("json", "encoding/json"))
			},
			func() error {
				return generation.ReserveGeneratedImport(NewImport("jason", "encoding/json"))
			},
			func() error {
				return generation.DeclareImport(NewImport("jsonp", "encoding/json"))
			},
		}
		if reverse {
			declare[0], declare[2] = declare[2], declare[0]
		}
		for _, register := range declare {
			require.NoError(t, register())
		}
		require.NoError(t, generation.Freeze())
		return generation.ImportName("encoding/json")
	}

	require.Equal(t, "json", freeze(false))
	require.Equal(t, "json", freeze(true))
}

// TestGeneratedImportPreferenceIsOrderIndependent verifies that repeated
// generated-package preferences for one path use deterministic spelling.
func TestGeneratedImportPreferenceIsOrderIndependent(t *testing.T) {
	freeze := func(first, second string) string {
		generation := NewGeneration("generated.local/gen", nil)
		require.NoError(t, generation.ReserveGeneratedImport(NewImport(first, "generated.local/gen/value")))
		require.NoError(t, generation.ReserveGeneratedImport(NewImport(second, "generated.local/gen/value")))
		require.NoError(t, generation.Freeze())
		return generation.ImportName("generated.local/gen/value")
	}

	require.Equal(t, freeze("alpha", "zeta"), freeze("zeta", "alpha"))
}

// TestImportAliasRejectsIncompatibleFixedRequirements verifies that static
// templates cannot request two different mandatory spellings for one path.
func TestImportAliasRejectsIncompatibleFixedRequirements(t *testing.T) {
	generation := NewGeneration("generated.local/gen", nil)
	require.NoError(t, generation.RequireImport(NewImport("json", "encoding/json")))
	require.ErrorContains(
		t,
		generation.RequireImport(NewImport("jason", "encoding/json")),
		"requires qualifier",
	)
}

// TestImportAliasRejectsFixedQualifierCollision verifies that two static
// packages cannot both require the same qualifier.
func TestImportAliasRejectsFixedQualifierCollision(t *testing.T) {
	generation := NewGeneration("generated.local/gen", nil)
	require.NoError(t, generation.RequireImport(NewImport("runtime", "example.com/first")))
	require.NoError(t, generation.RequireImport(NewImport("runtime", "example.com/second")))
	require.ErrorContains(t, generation.Freeze(), "required by both")
}
