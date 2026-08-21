// This file verifies that service import subsets and qualified references use
// one deterministic full-path alias binding across every render analysis.
package service

import (
	"go/format"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// TestPlanRejectsUnregisteredRoot verifies that service planning cannot create
// render state outside the roots owned by its generation.
func TestPlanRejectsUnregisteredRoot(t *testing.T) {
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
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	require.ErrorContains(t, Plan(root, generation), "does not belong")
	require.NoError(t, generation.Freeze())
	_, err := NewServicesData(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.ErrorContains(t, err, "does not belong")
}

// TestPlanUsesCopiedGenerationRoots verifies that mutating root slices outside
// the generation cannot change which service designs planning and rendering
// accept.
func TestPlanUsesCopiedGenerationRoots(t *testing.T) {
	first := codegen.RunDSL(t, func() {
		dsl.Service("First", func() {
			dsl.Method("Read", func() {})
		})
	})
	second := codegen.RunDSL(t, func() {
		dsl.Service("Second", func() {
			dsl.Method("Read", func() {})
		})
	})
	roots := []eval.Root{first}
	generation := mustTestGeneration(t, "generated.local/gen", roots)
	roots[0] = second
	returnedRoots := generation.Roots()
	returnedRoots[0] = second

	require.NoError(t, Plan(first, generation))
	require.ErrorContains(t, Plan(second, generation), "does not belong")
	require.NoError(t, generation.Freeze())
	roots[0] = nil
	returnedRoots = generation.Roots()
	returnedRoots[0] = second
	_, err := NewServicesData(first, generation, expr.NewExampleGenerator(first.API.RandomizerFactory))
	require.NoError(t, err)
	_, err = NewServicesData(second, generation, expr.NewExampleGenerator(second.API.RandomizerFactory))
	require.ErrorContains(t, err, "does not belong")
}

// TestImportAliasesUsePathAsIdentity verifies that generator-owned imports
// retain their canonical qualifier when metadata prefers another spelling for
// the same complete package path.
func TestImportAliasesUsePathAsIdentity(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	require.NoError(t, generation.RequireImport(codegen.SimpleImport("encoding/json")))
	require.NoError(t, generation.DeclareImport(codegen.NewImport("jason", "encoding/json")))
	require.NoError(t, generation.Freeze())

	aliases := &importAliases{generation: generation}
	require.Equal(t, "json", aliases.name("encoding/json"))
	require.Equal(t, "encoding/json", aliases.spec("encoding/json").Path)
}

// TestImportAliasPreferenceIsOrderIndependent verifies that two metadata
// spellings for one path produce the same frozen qualifier in either order.
func TestImportAliasPreferenceIsOrderIndependent(t *testing.T) {
	freeze := func(first, second string) string {
		generation := mustTestGeneration(t, "generated.local/gen", nil)
		require.NoError(t, generation.DeclareImport(codegen.NewImport(first, "example.com/value")))
		require.NoError(t, generation.DeclareImport(codegen.NewImport(second, "example.com/value")))
		require.NoError(t, generation.Freeze())
		return generation.ImportName("example.com/value")
	}

	require.Equal(t, freeze("alpha", "zeta"), freeze("zeta", "alpha"))
}

// TestRegisteredRootsShareImportAliases verifies that every root analysis and
// relocated declaration consumes the one mapping frozen for the generation.
func TestRegisteredRootsShareImportAliases(t *testing.T) {
	rootWithPreference := func(serviceName, typeName, preferred string) *expr.RootExpr {
		return codegen.RunDSL(t, func() {
			payload := dsl.Type(typeName, func() {
				dsl.Meta("struct:pkg:path", "types")
				dsl.Attribute("value", dsl.String, func() {
					dsl.Meta("struct:field:type", preferred+".Value", "example.com/shared/value", preferred)
				})
			})
			dsl.Service(serviceName, func() {
				dsl.Method("Read", func() {
					dsl.Payload(payload)
				})
			})
		})
	}
	firstRoot := rootWithPreference("First", "FirstPayload", "zeta")
	secondRoot := rootWithPreference("Second", "SecondPayload", "alpha")
	generation := mustTestGeneration(t, "generated.local/gen", []eval.Root{firstRoot, secondRoot})
	require.NoError(t, Plan(firstRoot, generation))
	require.NoError(t, Plan(secondRoot, generation))
	require.NoError(t, generation.Freeze())
	first, err := NewServicesData(firstRoot, generation, expr.NewExampleGenerator(firstRoot.API.RandomizerFactory))
	require.NoError(t, err)
	second, err := NewServicesData(secondRoot, generation, expr.NewExampleGenerator(secondRoot.API.RandomizerFactory))
	require.NoError(t, err)
	require.Equal(t, "alpha", first.aliases.name("example.com/shared/value"))
	require.Equal(t, first.aliases.name("example.com/shared/value"), second.aliases.name("example.com/shared/value"))

	files := Files(generation.GenPkg(), []*ServicesData{first, second})
	for _, name := range []string{"first_payload.go", "second_payload.go"} {
		file := findFile(files, path.Join("gen", "types", name))
		require.NotNil(t, file)
		code := renderSections(t, file.SectionTemplates)
		require.Contains(t, code, `alpha "example.com/shared/value"`)
		require.Contains(t, code, "alpha.Value")
	}
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
	generation := mustTestGeneration(t, "generated.local/gen", []eval.Root{root})
	require.NoError(t, Plan(root, generation))
	require.NoError(t, generation.Freeze())
	aliases, err := newImportAliases(root, generation)
	require.NoError(t, err)
	require.Equal(t, "json", aliases.name("encoding/json"))
	require.Equal(t, "json2", aliases.name("example.com/custom/json"))
}

// TestFixedTemplateAliasesBeatGeneratedPackages verifies that generated
// service paths cannot take qualifiers required by static Goa and log calls.
func TestFixedTemplateAliasesBeatGeneratedPackages(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		interceptor := dsl.Interceptor("Trace")
		for _, name := range []string{"Goa", "Log"} {
			dsl.Service(name, func() {
				dsl.ServerInterceptor(interceptor)
				dsl.Method("Read", func() {})
			})
		}
	})
	services := mustServicesData(t, root)
	require.Equal(t, "goa", services.aliases.name(codegen.GoaImport("").Path))
	require.Equal(t, "goa2", services.aliases.name(servicePackagePath(services.generation.GenPkg(), root.Service("Goa"))))
	require.Equal(t, "log", services.aliases.name("goa.design/clue/log"))
	require.Equal(t, "log2", services.aliases.name(servicePackagePath(services.generation.GenPkg(), root.Service("Log"))))
}

// TestDocumentedJSONMetadataUsesCanonicalAlias verifies that an alternate
// metadata spelling for encoding/json produces one import and one canonical
// qualifier in the generated service definition.
func TestDocumentedJSONMetadataUsesCanonicalAlias(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		payload := dsl.Type("Payload", func() {
			dsl.Attribute("raw", dsl.String, func() {
				dsl.Meta("struct:field:type", "jason.RawMessage", "encoding/json", "jason")
			})
		})
		dsl.Service("Values", func() {
			dsl.Method("Read", func() {
				dsl.Payload(payload)
			})
		})
	})
	services := mustServicesData(t, root)
	file := findFile(Files(services.generation.GenPkg(), []*ServicesData{services}), path.Join("gen", "values", "service.go"))
	require.NotNil(t, file)
	code := renderSections(t, file.SectionTemplates)
	require.Contains(t, code, "json.RawMessage")
	require.Equal(t, 1, strings.Count(code, `"encoding/json"`), code)
	require.NotContains(t, code, "jason.RawMessage")
}

// TestExampleServiceUsesCanonicalGeneratedPackageQualifier verifies that a
// metadata package cannot steal the qualifier reserved for a generated
// service package.
func TestExampleServiceUsesCanonicalGeneratedPackageQualifier(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		dsl.Service("Values", func() {
			dsl.Method("Read", func() {
				dsl.Payload(dsl.String, func() {
					dsl.Meta("struct:field:type", "values.Value", "example.com/custom/values", "values")
				})
			})
		})
	})
	services := mustServicesData(t, root)
	servicePath := servicePackagePath(services.generation.GenPkg(), root.Service("Values"))
	require.Equal(t, "values", services.aliases.name(servicePath))
	require.Equal(t, "values2", services.aliases.name("example.com/custom/values"))

	files := ExampleServiceFiles(services.generation.GenPkg(), root, services)
	require.Len(t, files, 1)
	code := renderSections(t, files[0].SectionTemplates)
	_, err := format.Source([]byte(code))
	require.NoError(t, err, code)
	require.Contains(t, code, "values.Service")
	require.Contains(t, code, "p values2.Value")
}

// TestExampleServiceReservesFixedQualifiers verifies that standard library
// and generated service imports retain their template qualifiers when service
// metadata or names request the same spelling.
func TestExampleServiceReservesFixedQualifiers(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		dsl.Service("Fmt", func() {
			dsl.Method("Read", func() {
				dsl.Payload(dsl.String, func() {
					dsl.Meta("struct:field:type", "strings.Value", "example.com/custom/strings", "strings")
				})
			})
		})
	})
	services := mustServicesData(t, root)
	servicePath := servicePackagePath(services.generation.GenPkg(), root.Service("Fmt"))
	servicePkg := services.aliases.name(servicePath)
	require.NotEqual(t, "fmt", servicePkg)
	require.Equal(t, "strings2", services.aliases.name("example.com/custom/strings"))

	files := ExampleServiceFiles(services.generation.GenPkg(), root, services)
	require.Len(t, files, 1)
	code := renderSections(t, files[0].SectionTemplates)
	_, err := format.Source([]byte(code))
	require.NoError(t, err, code)
	require.Contains(t, code, servicePkg+".Service")
	require.Contains(t, code, "p strings2.Value")
}

// TestServiceUsesCanonicalViewsQualifier verifies that design metadata cannot
// steal the qualifier reserved for the generated views package.
func TestServiceUsesCanonicalViewsQualifier(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		result := dsl.ResultType("application/vnd.value", func() {
			dsl.TypeName("Value")
			dsl.Attribute("custom", dsl.String, func() {
				dsl.Meta("struct:field:type", "valuesviews.Value", "example.com/custom/views", "valuesviews")
			})
			dsl.View("default", func() {
				dsl.Attribute("custom")
			})
		})
		dsl.Service("Values", func() {
			dsl.Method("Read", func() {
				dsl.Result(result)
			})
		})
	})
	services := mustServicesData(t, root)
	servicePath := servicePackagePath(services.generation.GenPkg(), root.Service("Values"))
	viewsPath := servicePath + "/views"
	require.Equal(t, "valuesviews", services.aliases.name(viewsPath))
	require.Equal(t, "valuesviews2", services.aliases.name("example.com/custom/views"))

	file := findFile(Files(services.generation.GenPkg(), []*ServicesData{services}), path.Join("gen", "values", "service.go"))
	require.NotNil(t, file)
	code := renderSections(t, file.SectionTemplates)
	_, err := format.Source([]byte(code))
	require.NoError(t, err, code)
	require.Contains(t, code, `valuesviews "`+viewsPath+`"`)
	require.Contains(t, code, `valuesviews2 "example.com/custom/views"`)
}

// TestUnionFieldReferencesUseFixedImportAliases verifies that the qualifier in
// a union field type and the import declaration come from the same frozen path
// binding when encoding/json already owns the preferred json name.
func TestUnionFieldReferencesUseFixedImportAliases(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	require.NoError(t, generation.RequireImport(codegen.SimpleImport("encoding/json")))
	require.NoError(t, generation.DeclareImport(codegen.NewImport("values", "generated.local/gen/values")))
	require.NoError(t, generation.DeclareImport(codegen.NewImport("json", "example.com/custom/json")))
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
	generatedPackage := mustClaimTestPackage(t, generation, "generated.local/gen/values")
	_, err := generatedPackage.DeclareUnion(union)
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	aliases := &importAliases{generation: generation}
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

	collector := newImportCollector(aliases, generation.GenPkg(), "generated.local/gen/values")
	collector.collectDefinition(branch)
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
