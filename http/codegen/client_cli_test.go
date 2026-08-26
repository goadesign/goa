// This file verifies HTTP client CLI generation consumes stable, non-empty
// examples for body, parameter, header, cookie, array, and map flags.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/cli"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestClientCLIFiles(t *testing.T) {
	cases := []struct {
		Name         string
		DSL          func()
		FileIndex    int
		SectionIndex int
	}{
		{"no-payload-parse", testdata.MultiNoPayloadDSL, 0, 3},
		{"simple-parse", testdata.MultiSimpleDSL, 0, 3},
		{"multi-parse", testdata.MultiDSL, 0, 3},
		{"multi-required-payload", testdata.MultiRequiredPayloadDSL, 0, 3},
		{"skip-request-body-encode-decode", testdata.SkipRequestBodyEncodeDecodeDSL, 0, 3},
		{"streaming-parse", testdata.StreamingMultipleServicesDSL, 0, 3},
		{"simple-build", testdata.MultiSimpleDSL, 1, 1},
		{"multi-build", testdata.MultiDSL, 1, 1},
		{"bool-build", testdata.PayloadQueryBoolDSL, 1, 1},
		{"uint32-build", testdata.PayloadQueryUInt32DSL, 1, 1},
		{"uint64-build", testdata.PayloadQueryUIntDSL, 1, 1},
		{"string-build", testdata.PayloadQueryStringDSL, 1, 1},
		{"string-required-build", testdata.PayloadQueryStringValidateDSL, 1, 1},
		{"string-default-build", testdata.PayloadQueryStringDefaultDSL, 1, 1},
		{"path-string-default-build", testdata.PayloadPathStringDefaultDSL, 1, 1},
		{"body-query-path-object-build", testdata.PayloadBodyQueryPathObjectDSL, 1, 1},
		{"body-nested-user-build", testdata.PayloadBodyNestedUserDSL, 1, 1},
		{"param-validation-build", testdata.ParamValidateDSL, 1, 1},
		{"payload-primitive-type", testdata.PayloadBodyPrimitiveBoolValidateDSL, 0, 3},
		{"payload-array-primitive-type", testdata.PayloadBodyPrimitiveArrayStringValidateDSL, 0, 3},
		{"payload-array-user-type", testdata.PayloadBodyInlineArrayUserDSL, 1, 1},
		{"payload-map-user-type", testdata.PayloadBodyInlineMapUserDSL, 1, 1},
		{"payload-object-type", testdata.PayloadBodyInlineObjectDSL, 1, 1},
		{"payload-object-default-type", testdata.PayloadBodyInlineObjectDefaultDSL, 1, 1},
		{"map-query", testdata.PayloadMapQueryPrimitiveArrayDSL, 0, 3},
		{"map-query-object", testdata.PayloadMapQueryObjectDSL, 1, 1},
		{"empty-body-build", testdata.PayloadBodyPrimitiveFieldEmptyDSL, 1, 1},
		{"with-params-and-headers-dsl", testdata.WithParamsAndHeadersBlockDSL, 1, 1},
		{"body-custom-name", testdata.PayloadBodyCustomNameDSL, 1, 1},
		{"path-custom-name", testdata.PayloadPathCustomNameDSL, 1, 1},
		{"query-custom-name", testdata.PayloadQueryCustomNameDSL, 1, 1},
		{"header-custom-name", testdata.PayloadHeaderCustomNameDSL, 1, 1},
		{"cookie-custom-name", testdata.PayloadCookieCustomNameDSL, 1, 1},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := expr.RunDSL(t, c.DSL)
			plan := linkedHTTPPlanForRoot(t, root)
			fs := plan.ClientCLIFiles()
			sections := fs[c.FileIndex].SectionTemplates
			code := codegen.SectionCode(t, sections[c.SectionIndex])
			testutil.AssertGo(t, "testdata/golden/client_cli_"+c.Name+".go.golden", code)
		})
	}
}

// TestClientCLIFlagPresenceGolden shows how generated HTTP commands preserve
// explicit empty values while applying every authored zero-valued default.
func TestClientCLIFlagPresenceGolden(t *testing.T) {
	root := expr.RunDSL(t, func() {
		dsl.Service("FlagPresence", func() {
			dsl.Method("check", func() {
				dsl.Payload(func() {
					dsl.Field(1, "required", dsl.String)
					dsl.Field(2, "optional", dsl.String)
					dsl.Field(3, "empty", dsl.String, func() {
						dsl.Default("")
					})
					dsl.Field(4, "disabled", dsl.Boolean, func() {
						dsl.Default(false)
					})
					dsl.Field(5, "zero", dsl.Int, func() {
						dsl.Default(0)
					})
					dsl.Required("required")
				})
				dsl.HTTP(func() {
					dsl.GET("/")
					dsl.Param("required")
					dsl.Param("optional")
					dsl.Param("empty")
					dsl.Param("disabled")
					dsl.Param("zero")
				})
			})
		})
	})
	files := linkedHTTPPlanForRoot(t, root).ClientCLIFiles()
	require.Len(t, files, 2)

	parser := codegen.SectionsCode(t, files[0].Section("parse-endpoint"))
	testutil.AssertGo(t, "testdata/golden/client_cli_flag-presence-parse.go.golden", parser)
	builder := codegen.SectionsCode(t, files[1].Section("cli-build-payload"))
	testutil.AssertGo(t, "testdata/golden/client_cli_flag-presence-build.go.golden", builder)
}

// TestJSONRPCRequestIDCLIFlagPresence checks that a defaulted request ID uses
// an ordinary string flag while an optional ID keeps omission distinct from an
// explicitly empty string.
func TestJSONRPCRequestIDCLIFlagPresence(t *testing.T) {
	tests := []struct {
		name          string
		defaultID     bool
		wantPresence  bool
		wantParamType string
	}{
		{"defaulted", true, false, "string"},
		{"optional", false, true, "*string"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := expr.RunDSL(t, func() {
				dsl.Service("Requests", func() {
					dsl.JSONRPC(func() {
						dsl.POST("/rpc")
					})
					dsl.Method("send", func() {
						dsl.Payload(func() {
							if test.defaultID {
								dsl.ID("id", dsl.String, func() {
									dsl.Default("default-id")
								})
								return
							}
							dsl.ID("id", dsl.String)
						})
						dsl.JSONRPC(func() {})
					})
				})
			})
			plan, generation, servicePlan := plannedHTTPPlan(t, root, true)
			parser := plan.cliParsers[root.API.Servers[0].Name]
			require.NotNil(t, parser)
			require.Equal(t, test.wantPresence, parser.Declarations.PresenceFlagType != nil)

			require.NoError(t, generation.Freeze())
			require.NoError(t, servicePlan.Link())
			require.NoError(t, plan.Link())

			service := plan.services.Get("Requests")
			require.NotNil(t, service)
			command := buildSubcommandData(service, service.Endpoint("send"))
			require.Len(t, command.Flags, 1)
			require.Equal(t, test.defaultID, command.Flags[0].HasDefault)
			require.Equal(t, test.wantPresence, command.Flags[0].TracksPresence)
			require.NotNil(t, command.BuildFunction)
			require.Equal(t, []string{test.wantParamType}, command.BuildFunction.FormalParamTypes)
		})
	}
}

// TestClientCLIBuildNameMatchesDeclaration verifies released plugins can read
// the final payload builder name without choosing that name themselves.
func TestClientCLIBuildNameMatchesDeclaration(t *testing.T) {
	root := expr.RunDSL(t, testdata.MultiSimpleDSL)
	plan := linkedHTTPPlanForRoot(t, root)
	files := plan.ClientCLIFiles()
	require.Greater(t, len(files), 1)

	build, ok := files[1].SectionTemplates[1].Data.(*cli.BuildFunctionData)
	require.True(t, ok)
	endpoint := plan.services.Get("ServiceMultiSimple1").Endpoint("MethodMultiSimplePayload")
	require.Equal(t, endpoint.CLIPayloadDeclaration.Name(), build.Name)
}

// TestClientCLITransportNamesMatchDeclarations checks the released multipart
// and stream helper names exposed to plugin templates.
func TestClientCLITransportNamesMatchDeclarations(t *testing.T) {
	plan := linkedHTTPPlanForRoot(t, releasedHTTPNamesRoot(t))
	service := plan.services.Get("Names")

	multipart := buildSubcommandData(service, service.Endpoint("Multipart"))
	require.NotNil(t, multipart.MultipartFuncDeclaration)
	require.Equal(t, multipart.MultipartFuncDeclaration.Name(), multipart.MultipartFuncName)

	stream := buildSubcommandData(service, service.Endpoint("Raw"))
	require.NotNil(t, stream.BuildStreamPayloadDeclaration)
	require.Equal(t, stream.BuildStreamPayloadDeclaration.Name(), stream.BuildStreamPayload)
}

func TestEmptyBodyCLIUsesPayloadFieldExample(t *testing.T) {
	root := expr.RunDSL(t, testdata.PayloadBodyPrimitiveFieldEmptyDSL)
	plan := linkedHTTPPlanForRoot(t, root)
	endpoint := plan.services.Get("ServiceBodyPrimitiveArrayUser").Endpoints[0]
	require.NotNil(t, endpoint.Payload.Request.PayloadInit)
	require.Len(t, endpoint.Payload.Request.PayloadInit.ClientArgs, 1)
	example := endpoint.Payload.Request.PayloadInit.ClientArgs[0].Example

	require.IsType(t, []string{}, example)
	require.NotEmpty(t, example)
}

// TestClientCLINestedBodyWithoutValidationEmitsNoChecks verifies that a body
// with no top-level checks does not add validation to the payload builder.
func TestClientCLINestedBodyWithoutValidationEmitsNoChecks(t *testing.T) {
	root := expr.RunDSL(t, testdata.PayloadBodyUserInnerDSL)
	plan := linkedHTTPPlanForRoot(t, root)
	files := plan.ClientCLIFiles()
	require.NotEmpty(t, files)
	service := plan.services.Get("ServiceBodyUserInner")
	endpoint := service.Endpoints[0]
	require.NotNil(t, endpoint.Payload.Request.PayloadInit)
	require.Len(t, endpoint.Payload.Request.PayloadInit.ClientArgs, 1)
	arg := endpoint.Payload.Request.PayloadInit.ClientArgs[0]
	require.Empty(t, arg.Validate)
	require.NotNil(t, arg.CLIPlan)
	_, builder := buildFlags(service, endpoint)
	require.NotNil(t, builder)
	require.Len(t, builder.Fields, 1)
	require.NotContains(t, builder.Fields[0].Init, "goa.")
}
