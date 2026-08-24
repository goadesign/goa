// This file proves generated HTTP definitions and their callers use the same
// package names after another generator claims the preferred spelling.
package codegen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/codegentest"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestHTTPPlannedNamesSurvivePackageCollisions(t *testing.T) {
	root := expr.RunDSL(t, func() {
		child := dsl.Type("ChildPayload", func() {
			dsl.Attribute("value", dsl.String, func() {
				dsl.Pattern("value")
			})
			dsl.Required("value")
		})
		dsl.Service("Names", func() {
			dsl.Method("Complete", func() {
				dsl.Payload(func() {
					dsl.Attribute("child", child)
					dsl.Required("child")
				})
				dsl.HTTP(func() {
					dsl.POST("/complete")
				})
			})
			dsl.Method("Socket", func() {
				dsl.StreamingPayload(dsl.String)
				dsl.StreamingResult(dsl.String)
				dsl.HTTP(func() {
					dsl.GET("/socket")
				})
			})
		})
	})
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	clientPackage, err := generation.ClaimPackage("generated.local/gen/http/names/client")
	require.NoError(t, err)
	serverPackage, err := generation.ClaimPackage("generated.local/gen/http/names/server")
	require.NoError(t, err)
	for _, declaration := range []*codegen.NameDeclaration{
		codegen.NewExactName(codegen.NameFunction, "BuildCompleteRequest"),
	} {
		require.NoError(t, clientPackage.DeclareName(declaration))
	}
	for _, declaration := range []*codegen.NameDeclaration{
		codegen.NewExactName(codegen.NameType, "ChildPayloadRequestBody"),
		codegen.NewExactName(codegen.NameFunction, "ValidateChildPayloadRequestBody"),
		codegen.NewExactName(codegen.NameFunction, "validateChildPayloadRequestBody"),
		codegen.NewExactName(codegen.NameType, "SocketServerStream"),
	} {
		require.NoError(t, serverPackage.DeclareName(declaration))
	}
	plans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, plans[0].Link())

	serviceData := plans[0].services.Get("Names")
	complete := serviceData.Endpoint("Complete")
	require.Equal(t, "BuildCompleteRequest2", complete.RequestInit.Declaration.Name())
	require.Equal(t, complete.RequestInit.Declaration.Name(), complete.RequestInit.Name)
	childData := releasedTypeData(t, serviceData, func(data *TypeData) bool {
		return data.Declaration != nil && strings.HasPrefix(data.Declaration.Name(), "ChildPayload")
	})
	require.Equal(t, "ChildPayloadRequestBody2", childData.Declaration.Name())
	require.Equal(t, "ValidateChildPayloadRequestBody2", childData.ValidatorDeclaration.Name())
	require.Equal(t, "validateChildPayloadRequestBody2", childData.NestedValidatorDeclaration.Name())
	require.Equal(t, childData.Declaration.Name(), childData.VarName)
	require.Equal(t, childData.ValidatorDeclaration.Name(), childData.ValidatorName)
	require.Equal(t, childData.NestedValidatorDeclaration.Name(), childData.NestedValidatorName)
	socket := serviceData.Endpoint("Socket").ServerWebSocket
	require.Equal(t, "SocketServerStream2", socket.VarDeclaration.Name())
	require.Equal(t, socket.VarDeclaration.Name(), socket.VarName)

	var source strings.Builder
	for _, selection := range []struct {
		files   []*codegen.File
		file    string
		section string
		match   func(any) bool
	}{
		{plans[0].ClientFiles(), "encode_decode.go", "request-builder", func(data any) bool {
			endpoint, ok := data.(*EndpointData)
			return ok && endpoint.Method.Name == "Complete"
		}},
		{plans[0].ClientFiles(), "client.go", "client-endpoint-init", func(data any) bool {
			endpoint, ok := data.(*EndpointData)
			return ok && endpoint.Method.Name == "Complete"
		}},
		{plans[0].ServerTypeFiles(), "types.go", "server-body-attributes", func(data any) bool {
			body, ok := data.(*TypeData)
			return ok && body.Declaration == childData.Declaration
		}},
		{plans[0].ServerTypeFiles(), "types.go", "server-validate", func(data any) bool {
			body, ok := data.(*TypeData)
			return ok && body.Declaration == complete.Payload.Request.ServerBody.Declaration
		}},
		{plans[0].ServerTypeFiles(), "types.go", "server-validate", func(data any) bool {
			body, ok := data.(*TypeData)
			return ok && body.Declaration == childData.Declaration
		}},
		{plans[0].ServerFiles(), "encode_decode.go", "request-decoder", func(data any) bool {
			endpoint, ok := data.(*EndpointData)
			return ok && endpoint.Method.Name == "Complete"
		}},
		{plans[0].ServerFiles(), "websocket.go", "server-websocket-struct-type", func(data any) bool {
			stream, ok := data.(*WebSocketData)
			return ok && stream.VarDeclaration == socket.VarDeclaration
		}},
		{plans[0].ServerFiles(), "server.go", "server-handler-init", func(data any) bool {
			endpoint, ok := data.(*EndpointData)
			return ok && endpoint.Method.Name == "Socket"
		}},
	} {
		sections := codegentest.Sections(selection.files, selection.file, selection.section)
		matched := false
		for _, section := range sections {
			if selection.match(section.Data) {
				source.WriteString(codegen.SectionCode(t, section))
				source.WriteString("\n")
				matched = true
				break
			}
		}
		require.True(t, matched, "missing %s section in %s", selection.section, selection.file)
	}
	testutil.AssertGo(t, "testdata/golden/planned_name_collisions.go.golden", strings.TrimSpace(source.String())+"\n")
}

// TestHTTPUnionPlannedNamesRejectPackageCollisions checks that an authored
// union name never receives a moving numeric suffix.
func TestHTTPUnionPlannedNamesRejectPackageCollisions(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.Service("Names", func() {
			dsl.Method("Choose", func() {
				dsl.Payload(func() {
					dsl.OneOf("choice", func() {
						dsl.Attribute("text", dsl.String)
						dsl.Attribute("count", dsl.Int)
					})
					dsl.Required("choice")
				})
				dsl.HTTP(func() {
					dsl.POST("/choose")
				})
			})
		})
	})
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	serverPackage, err := generation.ClaimPackage("generated.local/gen/http/names/server")
	require.NoError(t, err)
	require.NoError(t, serverPackage.DeclareName(codegen.NewExactName(codegen.NameType, "ChoiceRequestBody")))
	_, err = NewPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.ErrorContains(t, err, `declare HTTP OneOf "choice" for RequestBody`)
	require.ErrorContains(t, err, `cannot declare exact type "ChoiceRequestBody"`)
	require.ErrorContains(t, err, "set TypeName on the OneOf to a unique name")
}

// TestJSONRPCValidatorPlannedNamesSurvivePackageCollisions checks that the
// JSON-RPC body validator definition and decoder call use the same planned
// declaration after the preferred names are already taken.
func TestJSONRPCValidatorPlannedNamesSurvivePackageCollisions(t *testing.T) {
	root := expr.RunDSL(t, func() {
		payload := dsl.Type("ChoosePayload", func() {
			dsl.Attribute("value", dsl.String, func() {
				dsl.MinLength(2)
			})
			dsl.Required("value")
		})
		dsl.Service("Names", func() {
			dsl.Method("Choose", func() {
				dsl.Payload(payload)
				dsl.JSONRPC(func() {
				})
			})
		})
	})
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	serverPackage, err := generation.ClaimPackage("generated.local/gen/jsonrpc/names/server")
	require.NoError(t, err)
	for _, declaration := range []*codegen.NameDeclaration{
		codegen.NewExactName(codegen.NameType, "ChooseRequestBody"),
		codegen.NewExactName(codegen.NameFunction, "ValidateChooseRequestBody"),
		codegen.NewExactName(codegen.NameFunction, "DecodeChooseRequest"),
	} {
		require.NoError(t, serverPackage.DeclareName(declaration))
	}
	plans, err := NewJSONRPCPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, plans[0].Link())

	serviceData := plans[0].services.Get("Names")
	body := serviceData.Endpoint("Choose").Payload.Request.ServerBody
	require.NotEqual(t, "ChooseRequestBody", body.Declaration.Name())
	require.NotEqual(t, "ValidateChooseRequestBody", body.ValidatorDeclaration.Name())
	snapshot, ok := plans[0].JSONRPCService("Names")
	require.True(t, ok)
	require.NotEqual(t, "DecodeChooseRequest", snapshot.Endpoints[0].RequestDecoderDeclaration.Name())

	var source strings.Builder
	for _, selection := range []struct {
		files   []*codegen.File
		file    string
		section string
		match   func(any) bool
	}{
		{plans[0].ServerTypeFiles(), "types.go", "request-body-type-decl", func(data any) bool {
			candidate, ok := data.(*TypeData)
			return ok && candidate.Declaration == body.Declaration
		}},
		{plans[0].ServerTypeFiles(), "types.go", "server-validate", func(data any) bool {
			candidate, ok := data.(*TypeData)
			return ok && candidate.Declaration == body.Declaration
		}},
		{[]*codegen.File{snapshot.ServerCodecFile()}, "encode_decode.go", "request-decoder", func(data any) bool {
			endpoint := plannedJSONRPCEndpoint(data)
			return endpoint != nil && endpoint.Method.Name == "Choose"
		}},
	} {
		sections := codegentest.Sections(selection.files, selection.file, selection.section)
		matched := false
		for _, section := range sections {
			if selection.match(section.Data) {
				source.WriteString(codegen.SectionCode(t, section))
				source.WriteString("\n")
				matched = true
				break
			}
		}
		require.True(t, matched, "missing %s section in %s", selection.section, selection.file)
	}
	testutil.AssertGo(t, "testdata/golden/planned_jsonrpc_validator_collisions.go.golden", strings.TrimSpace(source.String())+"\n")
}

// plannedJSONRPCEndpoint returns the copied endpoint stored in a JSON-RPC
// request codec section.
func plannedJSONRPCEndpoint(data any) *JSONRPCEndpointSnapshot {
	switch actual := data.(type) {
	case *jsonRPCRequestCodecData:
		return actual.JSONRPCEndpointSnapshot
	case *JSONRPCEndpointSnapshot:
		return actual
	default:
		return nil
	}
}
