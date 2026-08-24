// This file verifies that service import subsets and qualified references use
// one deterministic full-path alias binding across every render analysis.
package service

import (
	"go/format"
	"path"
	"path/filepath"
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
	require.ErrorContains(t, planTestServices(root, generation), "does not belong")
	require.NoError(t, generation.Freeze())
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

	firstPlan, err := NewPlan(first, generation, expr.NewExampleGenerator(first.API.RandomizerFactory))
	require.NoError(t, err)
	require.ErrorContains(t, planTestServices(second, generation), "does not belong")
	require.NoError(t, generation.Freeze())
	roots[0] = nil
	returnedRoots = generation.Roots()
	returnedRoots[0] = second
	require.NoError(t, firstPlan.Link())
	require.NotNil(t, firstPlan.Services().Get("First"))
}

// TestFileImportsAreRetainedBeforeFreeze verifies that rendering uses the
// exact package paths selected with the file contribution, not a later walk
// over mutable service-analysis slices.
func TestFileImportsAreRetainedBeforeFreeze(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		payload := dsl.Type("Payload", func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.Attribute("value", dsl.String)
		})
		dsl.Service("Values", func() {
			dsl.Method("Read", func() {
				dsl.Payload(payload)
			})
		})
	})
	generation := mustTestGeneration(t, "generated.local/gen", []eval.Root{root})
	plan, err := NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	plan.facts.services[0].referenceAttributes = nil
	require.NoError(t, generation.Freeze())
	require.NoError(t, plan.Link())

	file := endpointFile(plan, plan.facts.services[0])
	code := renderSections(t, file.SectionTemplates)
	require.Contains(t, code, `"generated.local/gen/types"`)
}

// TestImportAliasesUsePathAsIdentity verifies that generator-owned imports
// retain their canonical qualifier when metadata prefers another spelling for
// the same complete package path.
func TestImportAliasesUsePathAsIdentity(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	pkg := mustClaimTestPackage(t, generation, "generated.local/gen/values")
	require.NoError(t, pkg.RequireImport(codegen.SimpleImport("encoding/json")))
	require.NoError(t, pkg.DeclareImport(codegen.NewImport("jason", "encoding/json")))
	require.NoError(t, generation.Freeze())

	aliases := &importAliases{generation: generation}
	require.Equal(t, "json", aliases.name(pkg.ImportPath(), "encoding/json"))
	require.Equal(t, "encoding/json", aliases.spec(pkg.ImportPath(), "encoding/json").Path)
}

// TestImportAliasPreferenceIsOrderIndependent verifies that two metadata
// spellings for one path produce the same frozen qualifier in either order.
func TestImportAliasPreferenceIsOrderIndependent(t *testing.T) {
	freeze := func(first, second string) string {
		generation := mustTestGeneration(t, "generated.local/gen", nil)
		pkg := mustClaimTestPackage(t, generation, "generated.local/gen/values")
		require.NoError(t, pkg.DeclareImport(codegen.NewImport(first, "example.com/value")))
		require.NoError(t, pkg.DeclareImport(codegen.NewImport(second, "example.com/value")))
		require.NoError(t, generation.Freeze())
		return pkg.ImportName("example.com/value")
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
	plans, err := NewPlans(
		generation,
		PlanInput{Root: firstRoot, Examples: expr.NewExampleGenerator(firstRoot.API.RandomizerFactory)},
		PlanInput{Root: secondRoot, Examples: expr.NewExampleGenerator(secondRoot.API.RandomizerFactory)},
	)
	require.NoError(t, err)
	firstPlan, secondPlan := plans[0], plans[1]
	require.NoError(t, generation.Freeze())
	require.NoError(t, firstPlan.Link())
	require.NoError(t, secondPlan.Link())
	first := firstPlan.Services()
	second := secondPlan.Services()
	const outputPackage = "generated.local/gen/types"
	require.Equal(t, "alpha", first.aliases.name(outputPackage, "example.com/shared/value"))
	require.Equal(t, first.aliases.name(outputPackage, "example.com/shared/value"), second.aliases.name(outputPackage, "example.com/shared/value"))

	files := mustServiceFiles(t, firstPlan, secondPlan)
	for _, name := range []string{"first_payload.go", "second_payload.go"} {
		file := findFile(files, filepath.Join("gen", "types", name))
		require.NotNil(t, file)
		code := renderSections(t, file.SectionTemplates)
		require.Contains(t, code, `alpha "example.com/shared/value"`)
		require.Contains(t, code, "alpha.Value")
	}
}

// TestEmittedUnionReservesFixedJSON verifies that an emitted union codec's
// encoding/json qualifier wins before a metadata package requests the same
// preferred name. Files without a union do not reserve this runtime import.
func TestEmittedUnionReservesFixedJSON(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		payload := dsl.Type("Payload", func() {
			dsl.OneOf("choice", func() {
				dsl.Attribute("text", dsl.String)
			})
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
	require.NoError(t, planTestServices(root, generation))
	require.NoError(t, generation.Freeze())
	aliases, err := newImportAliases(root, generation)
	require.NoError(t, err)
	const outputPackage = "generated.local/gen/values"
	require.Equal(t, "json", aliases.name(outputPackage, "encoding/json"))
	require.Equal(t, "json2", aliases.name(outputPackage, "example.com/custom/json"))
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
	plan := mustServicePlan(t, root)
	services := plan.Services()
	outputPackage := path.Join(path.Dir(services.generation.GenPkg()), "interceptors")
	require.Equal(t, "goa", services.aliases.name(outputPackage, codegen.GoaImport("").Path))
	require.Equal(t, "goa2", services.aliases.name(outputPackage, servicePackagePath(services.generation.GenPkg(), root.Service("Goa"))))
	require.Equal(t, "log", services.aliases.name(outputPackage, "goa.design/clue/log"))
	require.Equal(t, "log2", services.aliases.name(outputPackage, servicePackagePath(services.generation.GenPkg(), root.Service("Log"))))
}

// TestMetadataImportKeepsItsPreferredAlias verifies that an import used only
// by design metadata is not renamed by an unused runtime package.
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
	plan := mustServicePlan(t, root)
	file := findFile(mustServiceFiles(t, plan), filepath.Join("gen", "values", "service.go"))
	require.NotNil(t, file)
	code := renderSections(t, file.SectionTemplates)
	require.Contains(t, code, "jason.RawMessage")
	require.Equal(t, 1, strings.Count(code, `"encoding/json"`), code)
	require.NotContains(t, code, "json.RawMessage")
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
	plan := mustServicePlan(t, root)
	services := plan.Services()
	servicePath := servicePackagePath(services.generation.GenPkg(), root.Service("Values"))
	outputPackage := path.Dir(services.generation.GenPkg())
	require.Equal(t, "values", services.aliases.name(outputPackage, servicePath))
	require.Equal(t, "values2", services.aliases.name(outputPackage, "example.com/custom/values"))

	files := ExampleServiceFiles(plan)
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
		result := dsl.Type("Result", func() {
			dsl.Attribute("length", dsl.Int)
		})
		dsl.Service("Fmt", func() {
			dsl.Method("Read", func() {
				dsl.Payload(dsl.String, func() {
					dsl.Meta("struct:field:type", "strings.Value", "example.com/custom/strings", "strings")
				})
				dsl.Result(result)
				dsl.HTTP(func() {
					dsl.GET("/")
					dsl.SkipResponseBodyEncodeDecode()
					dsl.Response(dsl.StatusOK, func() {
						dsl.Header("length:Content-Length")
					})
				})
			})
		})
	})
	plan := mustServicePlan(t, root)
	services := plan.Services()
	servicePath := servicePackagePath(services.generation.GenPkg(), root.Service("Fmt"))
	outputPackage := path.Dir(services.generation.GenPkg())
	servicePkg := services.aliases.name(outputPackage, servicePath)
	require.NotEqual(t, "fmt", servicePkg)
	require.Equal(t, "strings2", services.aliases.name(outputPackage, "example.com/custom/strings"))

	files := ExampleServiceFiles(plan)
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
	plan := mustServicePlan(t, root)
	services := plan.Services()
	servicePath := servicePackagePath(services.generation.GenPkg(), root.Service("Values"))
	viewsPath := servicePath + "/views"
	require.Equal(t, "valuesviews", services.aliases.name(servicePath, viewsPath))
	require.Equal(t, "valuesviews2", services.aliases.name(servicePath, "example.com/custom/views"))

	file := findFile(mustServiceFiles(t, plan), filepath.Join("gen", "values", "service.go"))
	require.NotNil(t, file)
	code := renderSections(t, file.SectionTemplates)
	_, err := format.Source([]byte(code))
	require.NoError(t, err, code)
	require.Contains(t, code, `valuesviews "`+viewsPath+`"`)
	require.Contains(t, code, `valuesviews2 "example.com/custom/views"`)
}

// TestViewValidationReservesOnlyUsedImports verifies that validation
// without string-length checks leaves the utf8 package name available to a
// field type supplied by the design.
func TestViewValidationReservesOnlyUsedImports(t *testing.T) {
	const customUTF8 = "example.com/custom/utf8"
	root := codegen.RunDSL(t, func() {
		result := dsl.ResultType("application/vnd.value", func() {
			dsl.Attribute("value", dsl.String, func() {
				dsl.Meta("struct:field:type", "utf8.Value", customUTF8, "utf8")
			})
			dsl.Attribute("name", dsl.String, func() {
				dsl.Pattern("^[a-z]+$")
			})
			dsl.View("default", func() {
				dsl.Attribute("value")
				dsl.Attribute("name")
			})
		})
		dsl.Service("Values", func() {
			dsl.Method("Read", func() {
				dsl.Result(result)
			})
		})
	})
	plan := mustServicePlan(t, root)
	outputPackage := servicePackagePath(plan.Services().generation.GenPkg(), root.Service("Values")) + "/views"
	require.Equal(t, "utf8", plan.Services().aliases.name(outputPackage, customUTF8))

	file := findFile(mustServiceFiles(t, plan), filepath.Join("gen", "values", "views", "view.go"))
	require.NotNil(t, file)
	code := renderSections(t, file.SectionTemplates)
	require.Contains(t, code, `"`+customUTF8+`"`)
	require.NotContains(t, code, `"unicode/utf8"`)
}

// TestUnionFieldReferencesUseFixedImportAliases verifies that the qualifier in
// a union field type and the import declaration come from the same frozen path
// binding when encoding/json already owns the preferred json name.
func TestUnionFieldReferencesUseFixedImportAliases(t *testing.T) {
	generation := mustTestGeneration(t, "generated.local/gen", nil)
	generatedPackage := mustClaimTestPackage(t, generation, "generated.local/gen/values")
	require.NoError(t, generatedPackage.RequireImport(codegen.SimpleImport("encoding/json")))
	require.NoError(t, generatedPackage.DeclareImport(codegen.NewImport("values", "generated.local/gen/values")))
	require.NoError(t, generatedPackage.DeclareImport(codegen.NewImport("json", "example.com/custom/json")))
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
	unionAttribute := &expr.AttributeExpr{Type: union}
	declaration, err := generatedPackage.DeclareUnion(unionAttribute)
	require.NoError(t, err)
	facts := &unionFacts{
		attribute:   unionAttribute,
		union:       union,
		identity:    codegen.NewUnionDeclarationID(unionAttribute),
		typeKey:     union.GetTypeKey(),
		valueKey:    union.GetValueKey(),
		declaration: declaration,
	}
	require.NoError(t, planUnionRenderFacts(facts, nil, generatedPackage))
	require.NoError(t, generation.Freeze())
	aliases := &importAliases{generation: generation}
	data := buildRetainedUnionTypeData(facts, aliases)
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
