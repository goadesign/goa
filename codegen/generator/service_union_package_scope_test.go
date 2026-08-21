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
	registry := testRegistryFromGenfuncs([]testGenfunc{
		{Plan: planServiceData, Generate: testServiceFiles},
		{Plan: planTransportData, Generate: testTransportFiles},
	})

	root := func() {
		dsl.API("relocated union package names", func() {})

		firstInput := dsl.Type("FirstInput", func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.OneOf("Value", func() {
				dsl.Field(1, "record", func() {
					dsl.OneOf("Nested", func() {
						dsl.Field(1, "text", dsl.String)
					})
					dsl.Required("Nested")
				})
			})
			dsl.Required("Value")
		})
		secondInput := dsl.Type("SecondInput", func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.OneOf("Value", func() {
				dsl.Field(1, "record", func() {
					dsl.OneOf("Nested", func() {
						dsl.Field(1, "text", dsl.Int)
					})
					dsl.Required("Nested")
				})
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
			dsl.Method("ReadJSON", func() {
				dsl.Payload(firstInput)
				dsl.Result(dsl.String)
				dsl.JSONRPC(func() {})
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
			dsl.Method("ReadJSON", func() {
				dsl.Payload(secondInput)
				dsl.Result(dsl.String)
				dsl.JSONRPC(func() {})
			})
		})
	}
	codegen.RunDSL(t, root)

	dir := t.TempDir()
	genDir := filepath.Join(dir, codegen.Gendir)
	writeGeneratedModule(t, genDir, "gen")
	_, err := generate(dir, "gen", false, registry)
	require.NoError(t, err)
	for _, path := range []string{
		filepath.Join("types", "first_input.go"),
		filepath.Join("types", "second_input.go"),
		filepath.Join("http", "first", "server", "server.go"),
		filepath.Join("http", "second", "server", "server.go"),
		filepath.Join("grpc", "first", "server", "server.go"),
		filepath.Join("grpc", "second", "server", "server.go"),
		filepath.Join("jsonrpc", "first", "server", "server.go"),
		filepath.Join("jsonrpc", "second", "server", "server.go"),
	} {
		require.FileExists(t, filepath.Join(genDir, path))
	}
	runGeneratedTests(t, genDir)
}

// TestInheritedTransportErrorMappingsCompileWithMethodErrors verifies reusable
// HTTP and gRPC response policy binds to the equivalent error value declared by
// the endpoint method instead of retaining the API declaration object.
func TestInheritedTransportErrorMappingsCompileWithMethodErrors(t *testing.T) {
	registry := testRegistryFromGenfuncs([]testGenfunc{
		{Plan: planServiceData, Generate: testServiceFiles},
		{Plan: planTransportData, Generate: testTransportFiles},
	})

	codegen.RunDSL(t, func() {
		dsl.API("error policy", func() {
			dsl.Error("bad_request", dsl.String)
			dsl.HTTP(func() { dsl.Response(dsl.StatusBadRequest, "bad_request") })
			dsl.GRPC(func() { dsl.Response("bad_request", dsl.CodeInvalidArgument) })
		})
		dsl.Service("Values", func() {
			dsl.Method("Read", func() {
				dsl.Error("bad_request", dsl.String)
				dsl.HTTP(func() { dsl.GET("/values") })
				dsl.GRPC(func() {})
			})
		})
	})

	dir := t.TempDir()
	genDir := filepath.Join(dir, codegen.Gendir)
	writeGeneratedModule(t, genDir, "gen")
	_, err := generate(dir, "gen", false, registry)
	require.NoError(t, err)
	runGeneratedTests(t, genDir)
}

// TestNestedTransportMetadataOwnsRecursiveImports verifies conversion helpers
// import a custom field type nested inside a relocated service declaration.
func TestNestedTransportMetadataOwnsRecursiveImports(t *testing.T) {
	registry := testRegistryFromGenfuncs([]testGenfunc{
		{Plan: planServiceData, Generate: testServiceFiles},
		{Plan: planTransportData, Generate: testTransportFiles},
	})

	codegen.RunDSL(t, func() {
		outer := dsl.Type("Outer", func() {
			dsl.Meta("struct:pkg:path", "domain/outer")
			dsl.Field(1, "value", dsl.String, func() {
				dsl.Meta("struct:field:type", "custom.Value", "gen/custom/value", "custom")
			})
		})
		dsl.Service("Values", func() {
			dsl.Method("HTTP", func() {
				dsl.Payload(outer)
				dsl.HTTP(func() { dsl.POST("/values") })
			})
			dsl.Method("GRPC", func() {
				dsl.Payload(outer)
				dsl.GRPC(func() {})
			})
			dsl.Method("JSONRPC", func() {
				dsl.Payload(outer)
				dsl.JSONRPC(func() {})
			})
		})
	})

	dir := t.TempDir()
	genDir := filepath.Join(dir, codegen.Gendir)
	writeGeneratedModule(t, genDir, "gen")
	_, err := generate(dir, "gen", false, registry)
	require.NoError(t, err)
	writeStubPackage(t, filepath.Join(genDir, "custom", "value"), "custom")
	runGeneratedTests(t, genDir)
}

// TestTransportServiceImportsUseFrozenAliases verifies a service package whose
// natural name collides with a fixed runtime import is declared and referenced
// with the same generation-owned qualifier in every transport.
func TestTransportServiceImportsUseFrozenAliases(t *testing.T) {
	registry := testRegistryFromGenfuncs([]testGenfunc{
		{Plan: planServiceData, Generate: testServiceFiles},
		{Plan: planTransportData, Generate: testTransportFiles},
	})

	codegen.RunDSL(t, func() {
		dsl.Service("Goa", func() {
			for _, transport := range []string{"HTTP", "GRPC", "JSONRPC"} {
				dsl.Method(transport, func() {
					dsl.Payload(func() { dsl.Field(1, "value", dsl.String) })
					switch transport {
					case "HTTP":
						dsl.HTTP(func() { dsl.POST("/values") })
					case "GRPC":
						dsl.GRPC(func() {})
					case "JSONRPC":
						dsl.JSONRPC(func() {})
					}
				})
			}
		})
		dsl.Service("Goahttp", func() {
			dsl.Method("HTTP", func() {
				dsl.Payload(func() { dsl.Field(1, "value", dsl.String) })
				dsl.HTTP(func() { dsl.POST("/http-values") })
			})
		})
		dsl.Service("Goapb", func() {
			dsl.Method("GRPC", func() {
				dsl.Payload(func() { dsl.Field(1, "value", dsl.String) })
				dsl.GRPC(func() {})
			})
		})
	})

	dir := t.TempDir()
	genDir := filepath.Join(dir, codegen.Gendir)
	writeGeneratedModule(t, genDir, "gen")
	_, err := generate(dir, "gen", false, registry)
	require.NoError(t, err)
	runGeneratedTests(t, genDir)
}

// TestInheritedTransportErrorsOwnImports verifies API-level response policy
// imports the relocated effective error referenced by generated HTTP and gRPC
// encoders even though the method does not redeclare it.
func TestInheritedTransportErrorsOwnImports(t *testing.T) {
	registry := testRegistryFromGenfuncs([]testGenfunc{
		{Plan: planServiceData, Generate: testServiceFiles},
		{Plan: planTransportData, Generate: testTransportFiles},
	})

	codegen.RunDSL(t, func() {
		fault := dsl.Type("Fault", func() {
			dsl.Meta("struct:pkg:path", "domain/errors")
			dsl.Attribute("message", dsl.String)
		})
		dsl.API("error imports", func() {
			dsl.Error("fault", fault)
			dsl.HTTP(func() { dsl.Response(dsl.StatusBadRequest, "fault") })
			dsl.GRPC(func() { dsl.Response("fault", dsl.CodeInvalidArgument) })
		})
		dsl.Service("Values", func() {
			dsl.Method("Read", func() {
				dsl.HTTP(func() { dsl.GET("/values") })
				dsl.GRPC(func() {})
			})
		})
	})

	dir := t.TempDir()
	genDir := filepath.Join(dir, codegen.Gendir)
	writeGeneratedModule(t, genDir, "gen")
	_, err := generate(dir, "gen", false, registry)
	require.NoError(t, err)
	runGeneratedTests(t, genDir)
}

// TestServiceUnionGeneratedBranchShapesCompile verifies that generated branch
// aliases with one natural name but different primitive shapes remain distinct.
func TestServiceUnionGeneratedBranchShapesCompile(t *testing.T) {
	registry := testRegistryFromGenfuncs([]testGenfunc{{Plan: planServiceData, Generate: testServiceFiles}})

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
	_, err := generate(dir, "gen", false, registry)
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
	registry := testRegistryFromGenfuncs([]testGenfunc{{Plan: planServiceData, Generate: testServiceFiles}})

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
	_, err := generate(dir, "gen", false, registry)
	require.NoError(t, err)
	runGeneratedTests(t, genDir)
}

// TestServiceFilesOwnTheirImports verifies that imports used by one service do
// not leak into another service file generated from the same design root.
func TestServiceFilesOwnTheirImports(t *testing.T) {
	registry := testRegistryFromGenfuncs([]testGenfunc{
		{Plan: planServiceData, Generate: testServiceFiles},
		{Plan: planTransportData, Generate: testTransportFiles},
	})

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
	_, err := generate(dir, "gen", false, registry)
	require.NoError(t, err)
	runGeneratedTests(t, genDir)
}

// TestRawBodyStructsRemainInEndpointsPackage verifies relocated payload and
// result declarations never relocate the request/response wrappers consumed by
// the raw HTTP body path.
func TestRawBodyStructsRemainInEndpointsPackage(t *testing.T) {
	registry := testRegistryFromGenfuncs([]testGenfunc{
		{Plan: planServiceData, Generate: testServiceFiles},
		{Plan: planTransportData, Generate: testTransportFiles},
	})

	codegen.RunDSL(t, func() {
		upload := dsl.Type("Upload", func() {
			dsl.Meta("struct:pkg:path", "domain/types")
			dsl.Attribute("length", dsl.Int)
			dsl.Required("length")
		})
		download := dsl.Type("Download", func() {
			dsl.Meta("struct:pkg:path", "domain/types")
			dsl.Attribute("length", dsl.Int)
			dsl.Required("length")
		})
		dsl.Service("RawBodies", func() {
			dsl.Method("Upload", func() {
				dsl.Payload(upload)
				dsl.HTTP(func() {
					dsl.POST("/upload")
					dsl.Header("length:Content-Length")
					dsl.SkipRequestBodyEncodeDecode()
					dsl.Response(dsl.StatusNoContent)
				})
			})
			dsl.Method("Download", func() {
				dsl.Result(download)
				dsl.HTTP(func() {
					dsl.GET("/download")
					dsl.SkipResponseBodyEncodeDecode()
					dsl.Response(dsl.StatusOK, func() {
						dsl.Header("length:Content-Length")
					})
				})
			})
		})
	})

	dir := t.TempDir()
	genDir := filepath.Join(dir, codegen.Gendir)
	writeGeneratedModule(t, genDir, "gen")
	_, err := generate(dir, "gen", false, registry)
	require.NoError(t, err)

	clientSource, err := os.ReadFile(filepath.Join(genDir, "http", "raw_bodies", "client", "client.go"))
	require.NoError(t, err)
	require.Contains(t, string(clientSource), "&rawbodies.DownloadResponseData")
	require.NotContains(t, string(clientSource), "types.DownloadResponseData")
	codecSource, err := os.ReadFile(filepath.Join(genDir, "http", "raw_bodies", "client", "encode_decode.go"))
	require.NoError(t, err)
	require.Contains(t, string(codecSource), "*rawbodies.UploadRequestData")
	require.NotContains(t, string(codecSource), "types.UploadRequestData")
	runGeneratedTests(t, genDir)
}

// TestServiceReferencesUseImportPathAliases verifies that one service can
// reference generated packages with the same Go package name without emitting
// duplicate import aliases or ambiguous qualified references.
func TestServiceReferencesUseImportPathAliases(t *testing.T) {
	registry := testRegistryFromGenfuncs([]testGenfunc{{Plan: planServiceData, Generate: testServiceFiles}})

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
	_, err := generate(dir, "gen", false, registry)
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

// TestTransportReferencesUseImportPathAliases verifies HTTP, gRPC, and
// JSON-RPC files qualify two same-basename service packages with the aliases
// frozen by the shared generation.
func TestTransportReferencesUseImportPathAliases(t *testing.T) {
	registry := testRegistryFromGenfuncs([]testGenfunc{
		{Plan: planServiceData, Generate: testServiceFiles},
		{Plan: planTransportData, Generate: testTransportFiles},
	})

	codegen.RunDSL(t, func() {
		first := dsl.Type("First", func() {
			dsl.Meta("struct:pkg:path", "first/shared")
			dsl.Field(1, "value", dsl.String)
		})
		second := dsl.Type("Second", func() {
			dsl.Meta("struct:pkg:path", "second/shared")
			dsl.Field(1, "value", dsl.String)
		})
		dsl.Service("Values", func() {
			for _, method := range []struct {
				name    string
				path    string
				payload expr.UserType
			}{
				{"HTTPFirst", "/http/first", first},
				{"HTTPSecond", "/http/second", second},
			} {
				dsl.Method(method.name, func() {
					dsl.Payload(method.payload)
					dsl.HTTP(func() { dsl.POST(method.path) })
				})
			}
			for _, method := range []struct {
				name    string
				payload expr.UserType
			}{
				{"GRPCFirst", first},
				{"GRPCSecond", second},
			} {
				dsl.Method(method.name, func() {
					dsl.Payload(method.payload)
					dsl.GRPC(func() {})
				})
			}
			for _, method := range []struct {
				name    string
				payload expr.UserType
			}{
				{"JSONFirst", first},
				{"JSONSecond", second},
			} {
				dsl.Method(method.name, func() {
					dsl.Payload(method.payload)
					dsl.JSONRPC(func() {})
				})
			}
		})
	})

	dir := t.TempDir()
	genDir := filepath.Join(dir, codegen.Gendir)
	writeGeneratedModule(t, genDir, "gen")
	_, err := generate(dir, "gen", false, registry)
	require.NoError(t, err)
	for _, transport := range []string{"http", "grpc", "jsonrpc"} {
		source := generatedTreeSource(t, filepath.Join(genDir, transport, "values"))
		require.Contains(t, source, `shared "gen/first/shared"`)
		require.Contains(t, source, `shared2 "gen/second/shared"`)
		require.Contains(t, source, "shared.First")
		require.Contains(t, source, "shared2.Second")
	}
	runGeneratedTests(t, genDir)
}

// generatedTreeSource returns the concatenated Go source below root in path
// order so tests can assert file-owned imports without depending on which
// transport file contains a conversion helper.
func generatedTreeSource(t *testing.T, root string) string {
	t.Helper()
	var source strings.Builder
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		source.Write(data)
		return nil
	})
	require.NoError(t, err)
	return source.String()
}

// TestNamedUnionBranchImportsReferenceOnly verifies that unions.go does not
// expand a named branch definition and import packages used only where that
// named type itself is declared.
func TestNamedUnionBranchImportsReferenceOnly(t *testing.T) {
	registry := testRegistryFromGenfuncs([]testGenfunc{{Plan: planServiceData, Generate: testServiceFiles}})

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
	_, err := generate(dir, "gen", false, registry)
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
	registry := testRegistryFromGenfuncs([]testGenfunc{{Plan: planServiceData, Generate: testServiceFiles}})

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
		_, err := generate(dir, "gen", false, registry)
		require.NoError(t, err)
		content, err := os.ReadFile(filepath.Join(genDir, "values", "service.go"))
		require.NoError(t, err)
		require.Contains(t, string(content), "type UsePayload struct")
		require.NotContains(t, string(content), "type UsePayload2 struct")
		runGeneratedTests(t, genDir)
	})

	t.Run("local name collides", func(t *testing.T) {
		registry := testRegistryFromGenfuncs([]testGenfunc{
			{Plan: planServiceData, Generate: testServiceFiles},
			{Plan: planTransportData, Generate: testTransportFiles},
		})
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
		_, err := generate(dir, "gen", false, registry)
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
	registry := testRegistryFromGenfuncs([]testGenfunc{{Plan: planServiceData, Generate: testServiceFiles}})

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
	_, err := generate(dir, "gen", false, registry)
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
	generation := mustTestGeneration(t, "gen", []eval.Root{root})
	require.NoError(t, planTransportData(generation))
	require.NoError(t, generation.Freeze())
	files, err := testTransportFiles(generation)
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

// TestRelocatedStreamingUnionReferencesCompile verifies WebSocket and SSE
// files resolve relocated streaming declarations through the frozen service
// packages while their event and frame bodies remain transport-owned.
func TestRelocatedStreamingUnionReferencesCompile(t *testing.T) {
	registry := testRegistryFromGenfuncs([]testGenfunc{
		{Plan: planServiceData, Generate: testServiceFiles},
		{Plan: planTransportData, Generate: testTransportFiles},
	})

	codegen.RunDSL(t, func() {
		streamInput := relocatedStreamingType("StreamInput", "InputChoice", dsl.String)
		streamOutput := relocatedStreamingType("StreamOutput", "OutputChoice", dsl.Int)
		sseEvent := dsl.Type("SSEEvent", func() {
			dsl.Attribute("data", streamOutput)
			dsl.Attribute("id", dsl.String)
			dsl.Required("data", "id")
		})
		dsl.Service("HTTPStreams", func() {
			dsl.Method("Socket", func() {
				dsl.StreamingPayload(streamInput)
				dsl.StreamingResult(streamOutput)
				dsl.HTTP(func() { dsl.GET("/socket") })
			})
			dsl.Method("Events", func() {
				dsl.StreamingResult(sseEvent)
				dsl.HTTP(func() {
					dsl.GET("/events")
					dsl.ServerSentEvents("data", func() { dsl.SSEEventID("id") })
				})
			})
		})
		dsl.Service("JSONSockets", func() {
			dsl.Method("Socket", func() {
				dsl.StreamingPayload(streamInput)
				dsl.StreamingResult(streamOutput)
				dsl.JSONRPC(func() {})
			})
		})
		dsl.Service("JSONEvents", func() {
			dsl.Method("Events", func() {
				dsl.StreamingResult(sseEvent)
				dsl.JSONRPC(func() {
					dsl.ServerSentEvents("data", func() { dsl.SSEEventID("id") })
				})
			})
		})
	})

	dir := t.TempDir()
	genDir := filepath.Join(dir, codegen.Gendir)
	writeGeneratedModule(t, genDir, "gen")
	_, err := generate(dir, "gen", false, registry)
	require.NoError(t, err)
	for _, path := range []string{
		filepath.Join("http", "http_streams", "server", "websocket.go"),
		filepath.Join("http", "http_streams", "server", "sse.go"),
		filepath.Join("jsonrpc", "json_sockets", "server", "websocket.go"),
		filepath.Join("jsonrpc", "json_events", "server", "sse.go"),
	} {
		require.FileExists(t, filepath.Join(genDir, path))
	}
	runGeneratedTests(t, genDir)
}

// relocatedStreamingType builds an object with a nested union that is emitted
// in the shared streaming package used by the integration test.
func relocatedStreamingType(name, unionName string, value expr.DataType) expr.UserType {
	return dsl.Type(name, func() {
		dsl.Meta("struct:pkg:path", "stream/types")
		dsl.OneOf(unionName, func() {
			dsl.Attribute("value", value)
		})
		dsl.Required(unionName)
	})
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
	generation := mustTestGeneration(t, "goa.design/goa/example", roots)
	require.NoError(t, planServiceData(generation))
	require.NoError(t, generation.Freeze())
	files, err := testServiceFiles(generation)
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
	generation := mustTestGeneration(t, "generated.local/gen", []eval.Root{root})
	require.NoError(t, servicecodegen.Plan(root, generation))
	require.NoError(t, generation.Freeze())
	services, err := servicecodegen.NewServicesData(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
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

// TestServiceAndExamplesCompileWithImportQualifierCollisions verifies that
// fixed packages, generated service packages, generated views packages, and
// metadata packages share one path-owned qualifier mapping.
func TestServiceAndExamplesCompileWithImportQualifierCollisions(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		result := dsl.ResultType("application/vnd.value", func() {
			dsl.TypeName("Value")
			dsl.Attribute("custom", dsl.String, func() {
				dsl.Meta("struct:field:type", "valuesviews.Value", "generated.local/custom/views", "valuesviews")
			})
			dsl.View("default", func() {
				dsl.Attribute("custom")
			})
		})
		dsl.Service("Values", func() {
			dsl.Method("Read", func() {
				dsl.Payload(dsl.String, func() {
					dsl.Meta("struct:field:type", "values.Value", "generated.local/custom/values", "values")
				})
				dsl.Result(result)
			})
		})
		dsl.Service("Fmt", func() {
			dsl.Method("Read", func() {
				dsl.Payload(dsl.String, func() {
					dsl.Meta("struct:field:type", "strings.Value", "generated.local/custom/strings", "strings")
				})
			})
		})
	})
	generation := mustTestGeneration(t, "generated.local/gen", []eval.Root{root})
	require.NoError(t, servicecodegen.Plan(root, generation))
	require.NoError(t, generation.Freeze())
	services, err := servicecodegen.NewServicesData(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)

	dir := t.TempDir()
	files, err := testServiceFiles(generation)
	require.NoError(t, err)
	files = append(files, servicecodegen.ExampleServiceFiles(generation.GenPkg(), root, services)...)
	for _, file := range files {
		_, err := file.Render(dir)
		require.NoError(t, err)
	}
	writeGeneratedModule(t, dir, "generated.local")
	writeStubPackage(t, filepath.Join(dir, "custom", "strings"), "strings")
	writeStubPackage(t, filepath.Join(dir, "custom", "values"), "values")
	writeStubPackage(t, filepath.Join(dir, "custom", "views"), "valuesviews")
	runGeneratedTests(t, dir)
}

// TestFixedRuntimeAliasesCompileWithGoaAndLogServices verifies that generated
// service imports are suffixed when static interceptor templates require the
// goa and log qualifiers for their runtime packages.
func TestFixedRuntimeAliasesCompileWithGoaAndLogServices(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		interceptor := dsl.Interceptor("Trace")
		for _, name := range []string{"Goa", "Log"} {
			dsl.Service(name, func() {
				dsl.ServerInterceptor(interceptor)
				dsl.Method("Read", func() {})
			})
		}
	})
	generation := mustTestGeneration(t, "generated.local/gen", []eval.Root{root})
	require.NoError(t, servicecodegen.Plan(root, generation))
	require.NoError(t, generation.Freeze())
	services, err := servicecodegen.NewServicesData(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)

	dir := t.TempDir()
	files, err := testServiceFiles(generation)
	require.NoError(t, err)
	files = append(files, servicecodegen.ExampleInterceptorsFiles(generation.GenPkg(), root, services)...)
	for _, file := range files {
		_, err := file.Render(dir)
		require.NoError(t, err)
	}
	goaSource, err := os.ReadFile(filepath.Join(dir, "interceptors", "goa_server.go"))
	require.NoError(t, err)
	require.Contains(t, string(goaSource), `goa "goa.design/goa/v3/pkg"`)
	require.Contains(t, string(goaSource), `goa2 "generated.local/gen/goa"`)
	logSource, err := os.ReadFile(filepath.Join(dir, "interceptors", "log_server.go"))
	require.NoError(t, err)
	require.Contains(t, string(logSource), `"goa.design/clue/log"`)
	require.Contains(t, string(logSource), `log2 "generated.local/gen/log"`)
	writeGeneratedModule(t, dir, "generated.local")
	runGeneratedTests(t, dir)
}

// TestTransportStaticAliasesCompileWithHttpAndPathServices verifies transport
// imports retain their literal qualifiers beside conflicting service names.
func TestTransportStaticAliasesCompileWithHttpAndPathServices(t *testing.T) {
	registry := testRegistryFromGenfuncs([]testGenfunc{
		{Plan: planServiceData, Generate: testServiceFiles},
		{Plan: planTransportData, Generate: testTransportFiles},
	})

	codegen.RunDSL(t, func() {
		for _, name := range []string{"Http", "Path"} {
			dsl.Service(name, func() {
				dsl.Method("Read", func() {
					dsl.Payload(dsl.String)
					dsl.Result(dsl.String)
					dsl.HTTP(func() { dsl.POST("/" + strings.ToLower(name)) })
					dsl.GRPC(func() {})
				})
				dsl.Method("ReadJSON", func() {
					dsl.Payload(dsl.String)
					dsl.Result(dsl.String)
					dsl.JSONRPC(func() {})
				})
			})
		}
	})

	dir := t.TempDir()
	genDir := filepath.Join(dir, codegen.Gendir)
	writeGeneratedModule(t, genDir, "gen")
	_, err := generate(dir, "gen", false, registry)
	require.NoError(t, err)
	httpServers, err := filepath.Glob(filepath.Join(genDir, "http", "*", "server", "server.go"))
	require.NoError(t, err)
	require.Len(t, httpServers, 2)
	var httpSource strings.Builder
	for _, server := range httpServers {
		source, err := os.ReadFile(server)
		require.NoError(t, err)
		httpSource.Write(source)
	}
	require.Contains(t, httpSource.String(), `http_ "gen/http_"`)
	require.Contains(t, httpSource.String(), `path2 "gen/path"`)
	runGeneratedTests(t, genDir)
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
