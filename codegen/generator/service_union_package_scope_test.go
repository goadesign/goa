// This file verifies that service generation allocates relocated union symbols
// once across every design root that contributes to a generated Go package.
package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	servicecodegen "goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// TestRelocatedUnionPackageNamesCompile verifies that two services and their
// HTTP and gRPC transports compile against distinct unions in one shared package.
func TestRelocatedUnionPackageNamesCompile(t *testing.T) {
	t.Cleanup(func() { Generators = generators })
	Generators = func(_ string) ([]Genfunc, error) {
		return []Genfunc{
			{Plan: planServiceData, Generate: Service},
			{Plan: planServiceData, Generate: Transport},
		}, nil
	}

	root := func() {
		dsl.API("relocated union package names", func() {})

		firstInput := dsl.Type("FirstInput", func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.OneOf("Value", func() {
				dsl.Field(1, "text", dsl.String)
			})
			dsl.Required("Value")
		})
		secondInput := dsl.Type("SecondInput", func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.OneOf("Value", func() {
				dsl.Field(1, "number", dsl.Int)
			})
			dsl.Required("Value")
		})
		dsl.Service("First", func() {
			dsl.Method("Read", func() {
				dsl.Payload(firstInput)
				dsl.Result(dsl.String)
				dsl.HTTP(func() {
					dsl.POST("/first")
					dsl.Response(200)
				})
				dsl.GRPC(func() {})
			})
		})
		dsl.Service("Second", func() {
			dsl.Method("Read", func() {
				dsl.Payload(secondInput)
				dsl.Result(dsl.String)
				dsl.HTTP(func() {
					dsl.POST("/second")
					dsl.Response(200)
				})
				dsl.GRPC(func() {})
			})
		})
	}
	codegen.RunDSL(t, root)

	dir := t.TempDir()
	genDir := filepath.Join(dir, codegen.Gendir)
	writeGeneratedModule(t, genDir, "gen")
	_, err := Generate(dir, "gen", false)
	require.NoError(t, err)
	for _, path := range []string{
		filepath.Join("types", "first_input.go"),
		filepath.Join("types", "second_input.go"),
		filepath.Join("http", "first", "server", "server.go"),
		filepath.Join("http", "second", "server", "server.go"),
		filepath.Join("grpc", "first", "server", "server.go"),
		filepath.Join("grpc", "second", "server", "server.go"),
	} {
		require.FileExists(t, filepath.Join(genDir, path))
	}
	runGeneratedTests(t, genDir)
}

// TestServiceUnionGeneratedBranchShapesCompile verifies that generated branch
// aliases with one natural name but different primitive shapes remain distinct.
func TestServiceUnionGeneratedBranchShapesCompile(t *testing.T) {
	t.Cleanup(func() { Generators = generators })
	Generators = func(_ string) ([]Genfunc, error) {
		return []Genfunc{{Plan: planServiceData, Generate: Service}}, nil
	}

	codegen.RunDSL(t, func() {
		first := dsl.Type("FirstValue", func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.OneOf("Value", func() {
				dsl.Attribute("text", dsl.String)
			})
		})
		second := dsl.Type("SecondValue", func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.OneOf("Value", func() {
				dsl.Attribute("text", dsl.Int)
			})
		})
		dsl.Service("Values", func() {
			dsl.Method("Read", func() {
				dsl.Payload(first)
				dsl.Result(second)
			})
		})
	})

	dir := t.TempDir()
	genDir := filepath.Join(dir, codegen.Gendir)
	writeGeneratedModule(t, genDir, "gen")
	_, err := Generate(dir, "gen", false)
	require.NoError(t, err)
	unionSource, err := os.ReadFile(filepath.Join(genDir, "types", "unions.go"))
	require.NoError(t, err)
	require.Contains(t, string(unionSource), "type Value struct")
	require.Contains(t, string(unionSource), "type Value2 struct")
	runGeneratedTests(t, genDir)
}

// TestServiceUnionFamilyNamesAvoidExactDeclarations verifies that union
// constants and constructors cannot collide with exact DSL type names.
func TestServiceUnionFamilyNamesAvoidExactDeclarations(t *testing.T) {
	t.Cleanup(func() { Generators = generators })
	Generators = func(_ string) ([]Genfunc, error) {
		return []Genfunc{{Plan: planServiceData, Generate: Service}}, nil
	}

	codegen.RunDSL(t, func() {
		kind := dsl.Type("ValueKindText", dsl.String)
		constructor := dsl.Type("NewValueText", dsl.String)
		payload := dsl.Type("Payload", func() {
			dsl.Attribute("kind", kind)
			dsl.Attribute("constructor", constructor)
			dsl.OneOf("Value", func() {
				dsl.Attribute("text", dsl.String)
			})
		})
		dsl.Service("Values", func() {
			dsl.Method("Read", func() {
				dsl.Payload(payload)
			})
		})
	})

	dir := t.TempDir()
	genDir := filepath.Join(dir, codegen.Gendir)
	writeGeneratedModule(t, genDir, "gen")
	_, err := Generate(dir, "gen", false)
	require.NoError(t, err)
	runGeneratedTests(t, genDir)
}

// TestServiceFilesOwnTheirImports verifies that imports used by one service do
// not leak into another service file generated from the same design root.
func TestServiceFilesOwnTheirImports(t *testing.T) {
	t.Cleanup(func() { Generators = generators })
	Generators = func(_ string) ([]Genfunc, error) {
		return []Genfunc{
			{Plan: planServiceData, Generate: Service},
			{Plan: planServiceData, Generate: Transport},
		}, nil
	}

	codegen.RunDSL(t, func() {
		dsl.API("file-owned imports", func() {})
		firstInput := dsl.Type("FirstInput", func() {
			dsl.Meta("struct:pkg:path", "first/shared")
			dsl.Field(1, "value", dsl.String)
		})
		secondInput := dsl.Type("SecondInput", func() {
			dsl.Meta("struct:pkg:path", "second/shared")
			dsl.Field(1, "value", dsl.String)
		})
		dsl.Service("First", func() {
			dsl.Method("Read", func() {
				dsl.Payload(firstInput)
				dsl.HTTP(func() {
					dsl.POST("/first")
					dsl.Response(204)
				})
				dsl.GRPC(func() {})
			})
		})
		dsl.Service("Second", func() {
			dsl.Method("Read", func() {
				dsl.Payload(secondInput)
				dsl.HTTP(func() {
					dsl.POST("/second")
					dsl.Response(204)
				})
				dsl.GRPC(func() {})
			})
		})
	})

	dir := t.TempDir()
	genDir := filepath.Join(dir, codegen.Gendir)
	writeGeneratedModule(t, genDir, "gen")
	_, err := Generate(dir, "gen", false)
	require.NoError(t, err)
	runGeneratedTests(t, genDir)
}

// TestServiceReferencesUseImportPathAliases verifies that one service can
// reference generated packages with the same Go package name without emitting
// duplicate import aliases or ambiguous qualified references.
func TestServiceReferencesUseImportPathAliases(t *testing.T) {
	t.Cleanup(func() { Generators = generators })
	Generators = func(_ string) ([]Genfunc, error) {
		return []Genfunc{{Plan: planServiceData, Generate: Service}}, nil
	}

	codegen.RunDSL(t, func() {
		dsl.API("path-owned aliases", func() {})
		first := dsl.Type("First", func() {
			dsl.Meta("struct:pkg:path", "first/shared")
			dsl.Attribute("value", dsl.String)
		})
		second := dsl.Type("Second", func() {
			dsl.Meta("struct:pkg:path", "second/shared")
			dsl.Attribute("value", dsl.String)
		})
		dsl.Service("Values", func() {
			dsl.Method("First", func() {
				dsl.Payload(first)
			})
			dsl.Method("Second", func() {
				dsl.Payload(second)
			})
		})
	})

	dir := t.TempDir()
	genDir := filepath.Join(dir, codegen.Gendir)
	writeGeneratedModule(t, genDir, "gen")
	_, err := Generate(dir, "gen", false)
	require.NoError(t, err)
	content, err := os.ReadFile(filepath.Join(genDir, "values", "service.go"))
	require.NoError(t, err)
	code := string(content)
	require.Contains(t, code, `shared "gen/first/shared"`)
	require.Contains(t, code, `shared2 "gen/second/shared"`)
	require.Contains(t, code, `*shared.First`)
	require.Contains(t, code, `*shared2.Second`)
	runGeneratedTests(t, genDir)
}

// TestNamedUnionBranchImportsReferenceOnly verifies that unions.go does not
// expand a named branch definition and import packages used only where that
// named type itself is declared.
func TestNamedUnionBranchImportsReferenceOnly(t *testing.T) {
	t.Cleanup(func() { Generators = generators })
	Generators = func(_ string) ([]Genfunc, error) {
		return []Genfunc{{Plan: planServiceData, Generate: Service}}, nil
	}

	codegen.RunDSL(t, func() {
		dsl.API("named branch imports", func() {})
		value := dsl.Type("Value", func() {
			dsl.OneOf("choice", func() {
				dsl.Attribute("external", dsl.String, func() {
					dsl.Meta("struct:field:type", "json.Value", "gen/custom/json", "json")
				})
			})
		})
		dsl.Service("Values", func() {
			dsl.Method("Read", func() {
				dsl.Result(value)
			})
		})
	})

	dir := t.TempDir()
	genDir := filepath.Join(dir, codegen.Gendir)
	writeGeneratedModule(t, genDir, "gen")
	_, err := Generate(dir, "gen", false)
	require.NoError(t, err)
	writeStubPackage(t, filepath.Join(genDir, "custom", "json"), "json")
	content, err := os.ReadFile(filepath.Join(genDir, "values", "unions.go"))
	require.NoError(t, err)
	code := string(content)
	require.Contains(t, code, `"encoding/json"`)
	require.NotContains(t, code, `"gen/custom/json"`)
	runGeneratedTests(t, genDir)
}

// TestNormalizedMethodTypesUseServicePackageNames verifies that raw method
// object wrappers collide only with declarations emitted in the same service
// package, never with a nested declaration relocated elsewhere.
func TestNormalizedMethodTypesUseServicePackageNames(t *testing.T) {
	t.Cleanup(func() { Generators = generators })
	Generators = func(_ string) ([]Genfunc, error) {
		return []Genfunc{{Plan: planServiceData, Generate: Service}}, nil
	}

	t.Run("relocated name does not collide", func(t *testing.T) {
		codegen.RunDSL(t, func() {
			dsl.API("relocated wrapper names", func() {})
			relocated := dsl.Type("UsePayload", func() {
				dsl.Meta("struct:pkg:path", "types")
				dsl.Field(1, "value", dsl.String)
			})
			dsl.Service("Values", func() {
				dsl.Method("Other", func() {
					dsl.Payload(func() {
						dsl.Field(1, "nested", relocated)
					})
					dsl.HTTP(func() {
						dsl.POST("/other")
						dsl.Response(204)
					})
					dsl.GRPC(func() {})
				})
				dsl.Method("Use", func() {
					dsl.Payload(func() {
						dsl.Field(1, "value", dsl.String)
					})
					dsl.HTTP(func() {
						dsl.POST("/use")
						dsl.Response(204)
					})
					dsl.GRPC(func() {})
				})
			})
		})

		dir := t.TempDir()
		genDir := filepath.Join(dir, codegen.Gendir)
		writeGeneratedModule(t, genDir, "gen")
		_, err := Generate(dir, "gen", false)
		require.NoError(t, err)
		content, err := os.ReadFile(filepath.Join(genDir, "values", "service.go"))
		require.NoError(t, err)
		require.Contains(t, string(content), "type UsePayload struct")
		require.NotContains(t, string(content), "type UsePayload2 struct")
		runGeneratedTests(t, genDir)
	})

	t.Run("local name collides", func(t *testing.T) {
		Generators = func(_ string) ([]Genfunc, error) {
			return []Genfunc{
				{Plan: planServiceData, Generate: Service},
				{Plan: planServiceData, Generate: Transport},
			}, nil
		}
		codegen.RunDSL(t, func() {
			dsl.API("local wrapper names", func() {})
			local := dsl.Type("UsePayload", func() {
				dsl.Field(1, "existing", dsl.String)
			})
			dsl.Service("Values", func() {
				dsl.Method("Existing", func() {
					dsl.Payload(local)
					dsl.HTTP(func() {
						dsl.POST("/existing")
						dsl.Response(204)
					})
					dsl.GRPC(func() {})
				})
				dsl.Method("Use", func() {
					dsl.Payload(func() {
						dsl.Field(1, "value", dsl.String)
					})
					dsl.HTTP(func() {
						dsl.POST("/use")
						dsl.Response(204)
					})
					dsl.GRPC(func() {})
				})
			})
		})

		dir := t.TempDir()
		genDir := filepath.Join(dir, codegen.Gendir)
		writeGeneratedModule(t, genDir, "gen")
		_, err := Generate(dir, "gen", false)
		require.NoError(t, err)
		content, err := os.ReadFile(filepath.Join(genDir, "values", "service.go"))
		require.NoError(t, err)
		code := string(content)
		require.Contains(t, code, "type UsePayload struct")
		require.Contains(t, code, "type UsePayload2 struct")
		runGeneratedTests(t, genDir)
	})
}

// TestNestedRelocatedDeclarationsOwnTheirImports verifies that metadata imports
// used by two relocated declarations stay in their respective declaration
// files and do not leak into the service file that references their package.
func TestNestedRelocatedDeclarationsOwnTheirImports(t *testing.T) {
	t.Cleanup(func() { Generators = generators })
	Generators = func(_ string) ([]Genfunc, error) {
		return []Genfunc{{Plan: planServiceData, Generate: Service}}, nil
	}

	codegen.RunDSL(t, func() {
		dsl.API("nested file-owned imports", func() {})
		outer := dsl.Type("Outer", func() {
			dsl.Meta("struct:pkg:path", "models")
			dsl.Attribute("value", dsl.String, func() {
				dsl.Meta("struct:field:type", "shared.Value", "gen/custom/first/shared", "shared")
			})
		})
		inner := dsl.Type("Inner", func() {
			dsl.Meta("struct:pkg:path", "models")
			dsl.Attribute("value", dsl.String, func() {
				dsl.Meta("struct:field:type", "shared.Value", "gen/custom/second/shared", "shared")
			})
		})
		dsl.Service("Nested", func() {
			dsl.Method("Outer", func() {
				dsl.Payload(outer)
			})
			dsl.Method("Inner", func() {
				dsl.Payload(inner)
			})
		})
	})

	dir := t.TempDir()
	genDir := filepath.Join(dir, codegen.Gendir)
	writeGeneratedModule(t, genDir, "gen")
	_, err := Generate(dir, "gen", false)
	require.NoError(t, err)
	writeStubPackage(t, filepath.Join(genDir, "custom", "first", "shared"), "shared")
	writeStubPackage(t, filepath.Join(genDir, "custom", "second", "shared"), "shared")
	runGeneratedTests(t, genDir)
}

// TestTransportSectionsOwnTheirImports verifies that a streaming transport
// file imports only the declarations used by the streaming endpoints it
// renders, even when another endpoint uses the same package name elsewhere.
func TestTransportSectionsOwnTheirImports(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		dsl.API("transport section imports", func() {})
		streamMessage := dsl.Type("StreamMessage", func() {
			dsl.Meta("struct:pkg:path", "stream/shared")
			dsl.Attribute("value", dsl.String)
		})
		request := dsl.Type("Request", func() {
			dsl.Meta("struct:pkg:path", "request/shared")
			dsl.Attribute("value", dsl.String)
		})
		dsl.Service("Messages", func() {
			dsl.Method("Watch", func() {
				dsl.StreamingResult(streamMessage)
				dsl.HTTP(func() {
					dsl.GET("/watch")
					dsl.Response(200)
				})
			})
			dsl.Method("Create", func() {
				dsl.Payload(request)
				dsl.HTTP(func() {
					dsl.POST("/messages")
					dsl.Response(204)
				})
			})
		})
	})
	generation := codegen.NewGeneration("gen", []eval.Root{root})
	require.NoError(t, planServiceData(generation))
	require.NoError(t, generation.Freeze())
	files, err := Transport(generation)
	require.NoError(t, err)

	var header strings.Builder
	for _, file := range files {
		if filepath.ToSlash(file.Path) != "gen/http/messages/server/websocket.go" {
			continue
		}
		require.NoError(t, file.SectionTemplates[0].Write(&header))
		break
	}
	require.NotEmpty(t, header.String())
	require.Contains(t, header.String(), `"gen/stream/shared"`)
	require.NotContains(t, header.String(), `"gen/request/shared"`)
}

// writeStubPackage creates the external package referenced by struct:field:type
// metadata inside the generated module used by the integration test.
func writeStubPackage(t *testing.T, dir, packageName string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, 0o750))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "value.go"),
		[]byte("package "+packageName+"\n\ntype Value string\n"),
		0o600,
	))
}

func TestServiceRelocatedUnionNamesSpanDesignRoots(t *testing.T) {
	roots := []eval.Root{
		codegen.RunDSL(t, unusedRelocatedValueRoot()),
		codegen.RunDSL(t, relocatedUnionRoot("ZExistingValue", "FirstService")),
		codegen.RunDSL(t, relocatedUnionRoot("MExistingValue", "SecondService")),
		codegen.RunDSL(t, relocatedUnionRoot("AAddedValue", "ThirdService")),
		codegen.RunDSL(t, relocatedDifferentUnionRoot()),
		codegen.RunDSL(t, relocatedTopLevelValueRoot()),
	}
	generation := codegen.NewGeneration("goa.design/goa/example", roots)
	require.NoError(t, planServiceData(generation))
	require.NoError(t, generation.Freeze())
	files, err := Service(generation)
	require.NoError(t, err)

	var generated strings.Builder
	for _, file := range files {
		if !strings.Contains(file.Path, filepath.Join("gen", "types")) {
			continue
		}
		for _, section := range file.SectionTemplates {
			require.NoError(t, section.Write(&generated))
		}
	}
	code := generated.String()
	require.Equal(t, 1, strings.Count(code, "type Value struct {"), code)
	require.Contains(t, code, "type Value struct {\n\tText string", code)
	require.Equal(t, 1, strings.Count(code, "type Value2 struct {"), code)
	require.Equal(t, 1, strings.Count(code, "type Value3 struct {"), code)
	require.Equal(t, 1, strings.Count(code, "type Value2Kind string"), code)
	require.Equal(t, 1, strings.Count(code, "type Value3Kind string"), code)
	require.Equal(t,
		[]string{"Value2", "Value2", "Value2", "Value3"},
		[]string{
			unionFieldType(code, "ZExistingValue"),
			unionFieldType(code, "MExistingValue"),
			unionFieldType(code, "AAddedValue"),
			unionFieldType(code, "DifferentValue"),
		},
	)
}

// TestServiceRelocatedUnionOwnerCompilesAcrossGeneration verifies a complete
// root analysis emits one shared relocated union for every referencing service.
func TestServiceRelocatedUnionOwnerCompilesAcrossGeneration(t *testing.T) {
	root := codegen.RunDSL(t, sharedRelocatedUnionRoot())
	generation := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, servicecodegen.Plan(root, generation))
	require.NoError(t, generation.Freeze())
	services, err := servicecodegen.NewServicesData(root, generation)
	require.NoError(t, err)
	files := servicecodegen.Files("generated.local/gen", []*servicecodegen.ServicesData{services})
	dir := t.TempDir()
	for _, file := range files {
		_, err := file.Render(dir)
		require.NoError(t, err)
	}
	writeGeneratedModule(t, dir, "generated.local")
	runGeneratedTests(t, dir)
}

// unusedRelocatedValueRoot declares a relocated type that no service reaches
// and does not force generation. It must not reserve a generated package name.
func unusedRelocatedValueRoot() func() {
	return func() {
		dsl.Type("Value", func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.Attribute("unused", dsl.String)
		})
	}
}

// sharedRelocatedUnionRoot declares the same relocated union from two services
// so one generation-wide render emits its definition exactly once.
func sharedRelocatedUnionRoot() func() {
	return func() {
		first := relocatedValueType("FirstValue")
		second := relocatedValueType("SecondValue")
		dsl.Service("FirstService", func() {
			dsl.Method("Read", func() {
				dsl.Payload(first)
			})
		})
		dsl.Service("SecondService", func() {
			dsl.Method("Read", func() {
				dsl.Payload(second)
			})
		})
	}
}

// relocatedValueType defines one force-generated owner of the shared Value
// union in the generated types package.
func relocatedValueType(name string) expr.UserType {
	return dsl.Type(name, func() {
		dsl.Meta("struct:pkg:path", "types")
		dsl.Meta("type:generate:force")
		dsl.OneOf("Value", func() {
			dsl.Attribute("Bool", dsl.Boolean)
			dsl.Attribute("Enum", dsl.String)
			dsl.Attribute("Number", dsl.Float64)
		})
	})
}

// relocatedTopLevelValueRoot declares an emitted top-level Value after the
// union definitions, so it receives the next available package-wide name.
func relocatedTopLevelValueRoot() func() {
	return func() {
		value := dsl.Type("Value", func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.Attribute("text", dsl.String)
			dsl.Required("text")
		})
		dsl.Service("FifthService", func() {
			dsl.Method("Read", func() {
				dsl.Payload(value)
			})
		})
	}
}

// relocatedDifferentUnionRoot defines a union with the same natural name but
// a different branch shape, so it must receive the next package-wide name.
func relocatedDifferentUnionRoot() func() {
	return func() {
		value := dsl.Type("DifferentValue", func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.Meta("type:generate:force")
			dsl.OneOf("Value", func() {
				dsl.Attribute("Bool", dsl.Boolean)
				dsl.Attribute("Text", dsl.String)
			})
		})
		dsl.Service("FourthService", func() {
			dsl.Method("Read", func() {
				dsl.Payload(value)
			})
		})
	}
}

// relocatedUnionRoot defines one independently declared union with the same
// name and branch shape as the declarations in the other design roots.
func relocatedUnionRoot(typeName, serviceName string) func() {
	return func() {
		value := dsl.Type(typeName, func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.Meta("type:generate:force")
			dsl.OneOf("Value", func() {
				dsl.Attribute("Bool", dsl.Boolean)
				dsl.Attribute("Enum", dsl.String)
				dsl.Attribute("Number", dsl.Float64)
			})
		})
		dsl.Service(serviceName, func() {
			dsl.Method("Read", func() {
				dsl.Payload(value)
			})
		})
	}
}

// unionFieldType returns the generated union type referenced by owner.
func unionFieldType(code, owner string) string {
	prefix := "type " + owner + " struct {\n\tValue "
	start := strings.Index(code, prefix)
	if start == -1 {
		return ""
	}
	start += len(prefix)
	end := strings.IndexByte(code[start:], '\n')
	if end == -1 {
		return ""
	}
	return code[start : start+end]
}
