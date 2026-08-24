// This file verifies HTTP client CLI generation consumes stable, non-empty
// examples for body, parameter, header, cookie, array, and map flags.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/cli"
	"goa.design/goa/v3/codegen/testutil"
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
		{"body-query-path-object-build", testdata.PayloadBodyQueryPathObjectDSL, 1, 1},
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
