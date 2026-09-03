// This file verifies that gRPC import planning records the runtime packages
// used by each generated file before package names become final.
package codegen

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

// TestGRPCErrorAnyImportsArePlannedForTypeFiles checks that method errors
// containing Any reserve both packages used by their conversion code. A user
// package named fmt must then receive the same final name in imports and type
// references, while command builders stay free of structpb when payloads and
// results do not use Any.
func TestGRPCErrorAnyImportsArePlannedForTypeFiles(t *testing.T) {
	root := expr.RunDSL(t, func() {
		failure := dsl.Type("Failure", func() {
			dsl.Field(1, "details", dsl.Any)
			dsl.Field(2, "code", dsl.String, func() {
				dsl.Meta("struct:field:type", "fmt.Value", "example.com/fmt", "fmt")
			})
		})
		dsl.Service("Imports", func() {
			dsl.Method("Read", func() {
				dsl.Payload(func() {
					dsl.Field(1, "value", dsl.String)
				})
				dsl.Result(func() {
					dsl.Field(1, "value", dsl.String)
				})
				dsl.Error("failure", failure)
				dsl.GRPC(func() {
					dsl.Response("failure", dsl.CodeInvalidArgument)
				})
			})
		})
	})

	generation, services := grpcServicePlans(t, []*expr.RootExpr{root})
	plans, err := newPlans(
		generation,
		fixedProtobufToolResolver(),
		PlanInput{Root: root, Service: services[0]},
	)
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, services[0].Link())
	require.NoError(t, plans[0].Link())

	for _, files := range [][]*codegen.File{plans[0].ClientTypeFiles(), plans[0].ServerTypeFiles()} {
		header := sectionCode(t, files[0].SectionTemplates[0])
		require.Contains(t, header, `"fmt"`)
		require.Contains(t, header, `fmt2 "example.com/fmt"`)
		require.Contains(t, header, `"google.golang.org/protobuf/types/known/structpb"`)
	}
	for _, file := range plans[0].ClientCLIFiles() {
		header := sectionCode(t, file.SectionTemplates[0])
		require.NotContains(t, header, "structpb")
	}
}

// TestGRPCRelocatedCustomErrorServerCompiles checks that server handlers
// import the package of a custom error type that is generated outside the
// service package.
func TestGRPCRelocatedCustomErrorServerCompiles(t *testing.T) {
	root := expr.RunDSL(t, func() {
		detail := dsl.Type("Detail", func() {
			dsl.Field(1, "message", dsl.String)
			dsl.Meta("struct:pkg:path", "details")
		})
		failure := dsl.Type("Failure", func() {
			dsl.Field(1, "detail", detail)
			dsl.Required("detail")
			dsl.Meta("struct:pkg:path", "types")
		})
		dsl.Service("Imports", func() {
			dsl.Method("Read", func() {
				dsl.Error("failure", failure)
				dsl.GRPC(func() {
					dsl.Response("failure", dsl.CodeFailedPrecondition)
				})
			})
		})
	})

	generation, servicePlans := grpcServicePlans(t, []*expr.RootExpr{root})
	plans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlans[0]})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlans[0].Link())
	require.NoError(t, plans[0].Link())
	header := sectionCode(t, plans[0].ServerFiles()[0].SectionTemplates[0])
	require.Contains(t, header, `types "generated.local/gen/types"`)
	require.NotContains(t, header, `"generated.local/gen/details"`)
	compileProtobufMethodServer(t, plans[0], servicePlans)
}

// TestGRPCRelocatedPrimitiveErrorServerCompiles checks that a primitive error
// does not add a type import because its handler uses Goa's generic error
// response.
func TestGRPCRelocatedPrimitiveErrorServerCompiles(t *testing.T) {
	root := expr.RunDSL(t, func() {
		failure := dsl.Type("FailureCode", dsl.String)
		failure.Attribute().AddMeta("struct:pkg:path", "types")
		dsl.Service("Imports", func() {
			dsl.Method("Read", func() {
				dsl.Error("failure", failure)
				dsl.GRPC(func() {
					dsl.Response("failure", dsl.CodeFailedPrecondition)
				})
			})
		})
	})

	generation, servicePlans := grpcServicePlans(t, []*expr.RootExpr{root})
	plans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlans[0]})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlans[0].Link())
	require.NoError(t, plans[0].Link())
	header := sectionCode(t, plans[0].ServerFiles()[0].SectionTemplates[0])
	require.NotContains(t, header, `"generated.local/gen/types"`)
	compileProtobufMethodServer(t, plans[0], servicePlans)
}

// TestGRPCStandardImportsAreReservedOnlyWhenGeneratedFilesUseThem checks that
// a generated service package keeps its preferred fmt, strconv, or errors name
// until source in the same output package calls that standard package.
func TestGRPCStandardImportsAreReservedOnlyWhenGeneratedFilesUseThem(t *testing.T) {
	unused := grpcUnusedStandardImportCollisionRoot(t)
	for _, packageName := range []string{"fmt", "strconv", "errors"} {
		t.Run("unused "+packageName, func(t *testing.T) {
			importPath := "example.com/" + packageName
			require.Equal(t, packageName, grpcServerExternalImportName(t, unused, importPath))
		})
	}

	tests := []struct {
		name string
		root func(*testing.T) *expr.RootExpr
	}{
		{"fmt used by Any error conversion", grpcFmtCollisionRoot},
		{"strconv used by response metadata", grpcStrconvCollisionRoot},
		{"errors used by method error handling", grpcErrorsCollisionRoot},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := test.root(t)
			packageName := test.name[:strings.IndexByte(test.name, ' ')]
			importPath := "example.com/" + packageName
			require.Equal(t, packageName+"2", grpcServerExternalImportName(t, root, importPath))
		})
	}
}

// TestGRPCResponseMetadataAnyPlansFmt checks that server response metadata
// formatting reserves fmt before a user package with the same preferred name.
func TestGRPCResponseMetadataAnyPlansFmt(t *testing.T) {
	root := expr.RunDSL(t, func() {
		result := dsl.Type("Result", func() {
			dsl.Field(1, "dynamic", dsl.Any)
			dsl.Field(2, "label", dsl.String, func() {
				dsl.Meta("struct:field:type", "fmt.Label", "example.com/fmt", "fmt")
			})
		})
		dsl.Service("Metadata", func() {
			dsl.Method("Read", func() {
				dsl.Result(result)
				dsl.GRPC(func() {
					dsl.Response(func() {
						dsl.Headers(func() {
							dsl.Attribute("dynamic:x-dynamic")
						})
					})
				})
			})
		})
	})
	plans := linkedGRPCPlans(t, root)
	header := sectionCode(t, plans[0].ServerFiles()[1].SectionTemplates[0])
	require.Contains(t, header, `"fmt"`)
	require.Contains(t, header, `fmt2 "example.com/fmt"`)
}

// TestGRPCNumericMetadataAliasesPlanStrconv checks that metadata conversion
// follows the transport's primitive copy instead of the named service type.
func TestGRPCNumericMetadataAliasesPlanStrconv(t *testing.T) {
	root := expr.RunDSL(t, func() {
		count := dsl.Type("Count", dsl.Int)
		counts := dsl.Type("Counts", dsl.ArrayOf(count))
		dsl.Service("Metadata", func() {
			dsl.Method("Read", func() {
				dsl.Payload(func() {
					dsl.Field(1, "count", count)
					dsl.Field(2, "counts", counts)
				})
				dsl.Result(func() {
					dsl.Field(1, "count", count)
					dsl.Field(2, "counts", counts)
				})
				dsl.GRPC(func() {
					dsl.Metadata(func() {
						dsl.Attribute("count:x-count")
						dsl.Attribute("counts:x-counts")
					})
					dsl.Response(func() {
						dsl.Headers(func() {
							dsl.Attribute("count:x-count")
							dsl.Attribute("counts:x-counts")
						})
					})
				})
			})
		})
	})

	plans := linkedGRPCPlans(t, root)
	clientHeader := sectionCode(t, plans[0].ClientFiles()[1].SectionTemplates[0])
	serverHeader := sectionCode(t, plans[0].ServerFiles()[1].SectionTemplates[0])
	require.Contains(t, clientHeader, `"strconv"`)
	require.Contains(t, serverHeader, `"strconv"`)
}

// TestGRPCMetadataStringLengthValidationImports checks that each decoder
// codec reserves the UTF-8 package only when its own metadata needs it.
func TestGRPCMetadataStringLengthValidationImports(t *testing.T) {
	tests := []struct {
		name       string
		request    bool
		response   bool
		wantClient bool
		wantServer bool
	}{
		{name: "request metadata", request: true, wantServer: true},
		{name: "response metadata", response: true, wantClient: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plans := linkedGRPCPlans(t, grpcMetadataStringLengthRoot(t, test.request, test.response))
			clientHeader := sectionCode(t, plans[0].ClientFiles()[1].SectionTemplates[0])
			serverHeader := sectionCode(t, plans[0].ServerFiles()[1].SectionTemplates[0])
			require.Equal(t, test.wantClient, strings.Contains(clientHeader, `"unicode/utf8"`))
			require.Equal(t, test.wantServer, strings.Contains(serverHeader, `"unicode/utf8"`))
		})
	}
}

// TestGRPCMetadataStringLengthValidationFilesCompile checks complete generated
// client and server packages with request and response metadata validation.
func TestGRPCMetadataStringLengthValidationFilesCompile(t *testing.T) {
	root := grpcMetadataStringLengthRoot(t, true, true)
	generation, servicePlans := grpcServicePlans(t, []*expr.RootExpr{root})
	plans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlans[0]})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlans[0].Link())
	require.NoError(t, plans[0].Link())
	clientCodec := grpcCodecFile(t, plans[0].ClientFiles())
	serverCodec := grpcCodecFile(t, plans[0].ServerFiles())
	testutil.AssertGo(t, "testdata/golden/metadata_validation_client_codec.go.golden", renderedGRPCFile(t, clientCodec))
	testutil.AssertGo(t, "testdata/golden/metadata_validation_server_codec.go.golden", renderedGRPCFile(t, serverCodec))
	compileProtobufMethodServer(t, plans[0], servicePlans)
}

// TestGRPCCLIImportsFollowFlagCode checks that the command parser and payload
// builder reserve only the fixed packages their generated code names.
func TestGRPCCLIImportsFollowFlagCode(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.Service("Commands", func() {
			dsl.Method("Read", func() {
				dsl.Payload(func() {
					dsl.Field(1, "name", dsl.String, func() {
						dsl.MinLength(2)
					})
					dsl.Field(2, "count", dsl.Int)
					dsl.Required("name")
				})
				dsl.GRPC(func() {
					dsl.Metadata(func() {
						dsl.Attribute("name:x-name")
						dsl.Attribute("count:x-count")
					})
				})
			})
		})
	})

	plans := linkedGRPCPlans(t, root)
	files := plans[0].ClientCLIFiles()
	require.Len(t, files, 2)
	parserHeader := sectionCode(t, files[0].SectionTemplates[0])
	require.Contains(t, parserHeader, `"flag"`)
	require.Contains(t, parserHeader, `"fmt"`)
	require.Contains(t, parserHeader, `"os"`)
	require.Contains(t, parserHeader, `goa "goa.design/goa/v3/pkg"`)
	require.Contains(t, parserHeader, `grpc "google.golang.org/grpc"`)
	for _, absent := range []string{"context", "goagrpc", "strconv", "utf8", "protojson", "structpb"} {
		require.NotContains(t, parserHeader, absent)
	}

	builderHeader := sectionCode(t, files[1].SectionTemplates[0])
	require.Contains(t, builderHeader, `"fmt"`)
	require.Contains(t, builderHeader, `"strconv"`)
	require.Contains(t, builderHeader, `"unicode/utf8"`)
	require.Contains(t, builderHeader, `goa "goa.design/goa/v3/pkg"`)
	for _, absent := range []string{"encoding/json", "protojson", "structpb"} {
		require.NotContains(t, builderHeader, absent)
	}
}

// TestGRPCPayloadAnyPlansStructPBOnlyForBuilder checks that an Any payload
// adds protobuf conversion imports to the payload builder, not the parser.
func TestGRPCPayloadAnyPlansStructPBOnlyForBuilder(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.Service("Commands", func() {
			dsl.Method("Create", func() {
				dsl.Payload(func() {
					dsl.Field(1, "value", dsl.Any)
				})
				dsl.GRPC(func() {})
			})
		})
	})

	files := linkedGRPCPlans(t, root)[0].ClientCLIFiles()
	require.Len(t, files, 2)
	parserHeader := sectionCode(t, files[0].SectionTemplates[0])
	require.NotContains(t, parserHeader, "protojson")
	require.NotContains(t, parserHeader, "structpb")
	builderHeader := sectionCode(t, files[1].SectionTemplates[0])
	require.Contains(t, builderHeader, `"fmt"`)
	require.Contains(t, builderHeader, `"google.golang.org/protobuf/encoding/protojson"`)
	require.Contains(t, builderHeader, `"google.golang.org/protobuf/types/known/structpb"`)
}

// TestGRPCEmptyResultPlansServerProtobufImport checks that the generated
// empty response constructor imports its protobuf result type.
func TestGRPCEmptyResultPlansServerProtobufImport(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.Service("Values", func() {
			dsl.Method("Read", func() {
				dsl.GRPC(func() {})
			})
		})
	})

	plans := linkedGRPCPlans(t, root)
	file := plans[0].ServerTypeFiles()[0]
	header := sectionCode(t, file.SectionTemplates[0])
	require.Contains(t, header, `"generated.local/gen/grpc/values/pb"`)
}

// grpcServerExternalImportName returns the final name used for importPath by
// the generated gRPC server package.
func grpcServerExternalImportName(t *testing.T, root *expr.RootExpr, importPath string) string {
	t.Helper()
	generation, services := grpcServicePlans(t, []*expr.RootExpr{root})
	plans, err := newPlans(
		generation,
		fixedProtobufToolResolver(),
		PlanInput{Root: root, Service: services[0]},
	)
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, services[0].Link())
	require.NoError(t, plans[0].Link())
	servicePlan := plans[0].servicesPlan[0]
	serverPath := path.Join(generation.GenPkg(), "grpc", servicePlan.packages.pathName, "server")
	return generation.Package(serverPath).ImportName(importPath)
}

// grpcCodecFile returns the encoder and decoder file from one generated side.
func grpcCodecFile(t *testing.T, files []*codegen.File) *codegen.File {
	t.Helper()
	for _, file := range files {
		if strings.HasSuffix(file.Path, "encode_decode.go") {
			return file
		}
	}
	require.Fail(t, "generated gRPC codec file was not planned")
	return nil
}

// renderedGRPCFile writes one planned file and returns its complete source.
func renderedGRPCFile(t *testing.T, file *codegen.File) string {
	t.Helper()
	filePath, err := file.Render(t.TempDir())
	require.NoError(t, err)
	source, err := os.ReadFile(filePath)
	require.NoError(t, err)
	return string(source)
}

// grpcMetadataStringLengthRoot returns a service whose selected metadata
// fields validate their Unicode character count while decoding.
func grpcMetadataStringLengthRoot(t *testing.T, request, response bool) *expr.RootExpr {
	t.Helper()
	return expr.RunDSL(t, func() {
		dsl.Service("Metadata Validation", func() {
			dsl.Method("Check", func() {
				dsl.Payload(func() {
					dsl.Field(1, "request", dsl.String, func() {
						dsl.MinLength(2)
					})
					dsl.Required("request")
				})
				dsl.Result(func() {
					dsl.Field(1, "response", dsl.String, func() {
						dsl.MinLength(3)
					})
					dsl.Required("response")
				})
				dsl.GRPC(func() {
					if request {
						dsl.Metadata(func() {
							dsl.Attribute("request:x-request")
						})
					}
					if response {
						dsl.Response(func() {
							dsl.Headers(func() {
								dsl.Attribute("response:x-response")
							})
						})
					}
				})
			})
		})
	})
}

// linkedGRPCPlans creates and links the gRPC plans for root.
func linkedGRPCPlans(t *testing.T, root *expr.RootExpr) []*Plan {
	t.Helper()
	generation, services := grpcServicePlans(t, []*expr.RootExpr{root})
	plans, err := newPlans(
		generation,
		fixedProtobufToolResolver(),
		PlanInput{Root: root, Service: services[0]},
	)
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, services[0].Link())
	require.NoError(t, plans[0].Link())
	return plans
}

// grpcFmtCollisionRoot returns a service whose error conversion calls fmt.
func grpcFmtCollisionRoot(t *testing.T) *expr.RootExpr {
	t.Helper()
	return expr.RunDSL(t, func() {
		failure := dsl.Type("Failure", func() {
			dsl.Field(1, "details", dsl.Any)
			dsl.Field(2, "label", dsl.String, func() {
				dsl.Meta("struct:field:type", "fmt.Label", "example.com/fmt", "fmt")
			})
		})
		dsl.Service("Collision", func() {
			dsl.Method("Read", func() {
				dsl.Result(dsl.String)
				dsl.Error("failure", failure)
				dsl.GRPC(func() {
					dsl.Response("failure", dsl.CodeInvalidArgument)
				})
			})
		})
	})
}

// grpcStrconvCollisionRoot returns a service whose response metadata formats
// an integer with strconv.
func grpcStrconvCollisionRoot(t *testing.T) *expr.RootExpr {
	t.Helper()
	return expr.RunDSL(t, func() {
		dsl.Service("Collision", func() {
			dsl.Method("Read", func() {
				dsl.Result(func() {
					dsl.Field(1, "count", dsl.Int, func() {
						dsl.Meta("struct:field:type", "strconv.Count", "example.com/strconv", "strconv")
					})
				})
				dsl.GRPC(func() {
					dsl.Response(func() {
						dsl.Headers(func() {
							dsl.Attribute("count:x-count")
						})
					})
				})
			})
		})
	})
}

// grpcErrorsCollisionRoot returns a service whose server handler inspects a
// declared method error with the standard errors package.
func grpcErrorsCollisionRoot(t *testing.T) *expr.RootExpr {
	t.Helper()
	return expr.RunDSL(t, func() {
		failure := dsl.Type("Failure", func() {
			dsl.Field(1, "cause", dsl.String, func() {
				dsl.Meta("struct:field:type", "errors.Cause", "example.com/errors", "errors")
			})
		})
		dsl.Service("Collision", func() {
			dsl.Method("Read", func() {
				dsl.Result(dsl.String)
				dsl.Error("failure", failure)
				dsl.GRPC(func() {
					dsl.Response("failure", dsl.CodeInvalidArgument)
				})
			})
		})
	})
}

// grpcUnusedStandardImportCollisionRoot returns one service that references
// user packages named fmt, strconv, and errors without calling the standard
// packages of those names.
func grpcUnusedStandardImportCollisionRoot(t *testing.T) *expr.RootExpr {
	t.Helper()
	return expr.RunDSL(t, func() {
		dsl.Service("Collision", func() {
			dsl.Method("Read", func() {
				dsl.Result(func() {
					dsl.Field(1, "label", dsl.String, func() {
						dsl.Meta("struct:field:type", "fmt.Label", "example.com/fmt", "fmt")
					})
					dsl.Field(2, "count", dsl.Int, func() {
						dsl.Meta("struct:field:type", "strconv.Count", "example.com/strconv", "strconv")
					})
					dsl.Field(3, "cause", dsl.String, func() {
						dsl.Meta("struct:field:type", "errors.Cause", "example.com/errors", "errors")
					})
				})
				dsl.GRPC(func() {})
			})
		})
	})
}
