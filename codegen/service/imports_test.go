// This file verifies that service import subsets and qualified references use
// one deterministic full-path alias binding across every render analysis.
package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

// TestImportAliasesIncludeExplicitAnalysisRoot verifies that plugin-local
// analysis sees imports from the root passed to NewServicesData even when that
// root is not listed in the generation's evaluated roots.
func TestImportAliasesIncludeExplicitAnalysisRoot(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		payload := dsl.Type("Payload", func() {
			dsl.Attribute("value", dsl.String, func() {
				dsl.Meta("struct:field:type", "shared.Value", "example.com/local/shared", "shared")
			})
		})
		dsl.Service("Values", func() {
			dsl.Method("Read", func() {
				dsl.Payload(payload)
			})
		})
	})
	generation := codegen.NewGeneration("generated.local/gen", nil)
	require.NoError(t, Plan(root, generation))
	require.NoError(t, generation.Freeze())

	services, err := NewServicesData(root, generation)
	require.NoError(t, err)
	aliases := services.aliases
	require.Equal(t, "shared", aliases.name("example.com/local/shared"))
	require.Equal(t, &codegen.ImportSpec{
		Name: "shared",
		Path: "example.com/local/shared",
	}, aliases.spec("example.com/local/shared"))
}

// TestImportAliasesReserveFixedJSON verifies that the union codec's
// encoding/json qualifier wins before a metadata package requests the same
// preferred name.
func TestImportAliasesReserveFixedJSON(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		payload := dsl.Type("Payload", func() {
			dsl.Attribute("value", dsl.String, func() {
				dsl.Meta("struct:field:type", "json.Value", "example.com/custom/json", "json")
			})
		})
		dsl.Service("Values", func() {
			dsl.Method("Read", func() {
				dsl.Payload(payload)
			})
		})
	})
	generation := codegen.NewGeneration("generated.local/gen", nil)

	aliases, err := newImportAliases(root, generation)
	require.NoError(t, err)
	require.Equal(t, "json", aliases.name("encoding/json"))
	require.Equal(t, "json2", aliases.name("example.com/custom/json"))
}

// TestUnionFieldReferencesUseFixedImportAliases verifies that the qualifier in
// a union field type and the import declaration come from the same frozen path
// binding when encoding/json already owns the preferred json name.
func TestUnionFieldReferencesUseFixedImportAliases(t *testing.T) {
	plan := &importAliasPlan{candidates: make(map[string]importAliasCandidate)}
	require.NoError(t, plan.addFixedImports())
	require.NoError(t, plan.add("generated.local/gen/values", "values", true, false))
	require.NoError(t, plan.add("example.com/custom/json", "json", true, false))
	aliases := plan.freeze()
	service := &expr.ServiceExpr{Name: "Values"}
	branch := &expr.AttributeExpr{Type: expr.String, Meta: expr.MetaExpr{
		"struct:field:type": {"json.Value", "example.com/custom/json", "json"},
	}}
	union := &expr.Union{
		TypeName: "Choice",
		Values: []*expr.NamedAttributeExpr{{
			Name:      "external",
			Attribute: branch,
		}},
	}
	generation := codegen.NewGeneration("generated.local/gen", nil)
	generatedPackage := generation.GeneratedPackage("generated.local/gen/values")
	_, err := generatedPackage.DeclareUnion(union)
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	declaration, err := generatedPackage.Union(union)
	require.NoError(t, err)
	data, err := buildUnionTypeData(
		union,
		declaration,
		newServiceResolver(generation, aliases, service, "generated.local/gen/values"),
		nil,
		false,
		func(branch *expr.NamedAttributeExpr) (*codegen.UnionBranchDeclaration, error) {
			return generatedPackage.UnionBranch(union, branch.Name)
		},
	)
	require.NoError(t, err)
	require.Equal(t, "json2.Value", data.Fields[0].FieldType)

	collector := newImportCollector(aliases, generation.GenPkg, "generated.local/gen/values")
	collector.collect(branch)
	header := codegen.Header(
		"Union types",
		"values",
		append([]*codegen.ImportSpec{codegen.SimpleImport("encoding/json")}, collector.imports()...),
	)
	var rendered strings.Builder
	require.NoError(t, header.Write(&rendered))
	require.Contains(t, rendered.String(), `"encoding/json"`)
	require.Contains(t, rendered.String(), `json2 "example.com/custom/json"`)
}
