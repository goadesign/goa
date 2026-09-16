package codegen

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/cli"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/grpc/codegen/testdata"
)

func TestClientCLIFiles(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"payload-with-validations", testdata.PayloadWithValidationsDSL},
		{"payload-with-message", testdata.PayloadWithMessageDSL},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := RunGRPCDSL(t, c.DSL)
			services := CreateGRPCServices(root)
			fs := clientCLIFiles(services)
			require.Greater(t, len(fs), 1, "expected at least 2 files")
			require.NotEmpty(t, fs[1].SectionTemplates)
			var buf bytes.Buffer
			for _, s := range fs[1].SectionTemplates {
				require.NoError(t, s.Write(&buf))
			}
			code := codegen.FormatTestCode(t, buf.String())
			testutil.AssertGo(t, "testdata/golden/client_cli_"+c.Name+".go.golden", code)
		})
	}
}

// TestClientCLIFlagPresenceGolden shows how generated gRPC commands preserve
// explicit empty metadata while applying every authored zero-valued default.
func TestClientCLIFlagPresenceGolden(t *testing.T) {
	root := RunGRPCDSL(t, func() {
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
				dsl.GRPC(func() {
					dsl.Metadata(func() {
						dsl.Attribute("required")
						dsl.Attribute("optional")
						dsl.Attribute("empty")
						dsl.Attribute("disabled")
						dsl.Attribute("zero")
					})
				})
			})
		})
	})
	services := CreateGRPCServices(root)
	files := clientCLIFiles(services)
	require.Len(t, files, 2)

	parser := codegen.SectionsCode(t, files[0].Section("parse-endpoint-grpc"))
	testutil.AssertGo(t, "testdata/golden/client_cli_flag-presence-parse.go.golden", parser)
	builder := codegen.SectionsCode(t, files[1].Section("cli-build-payload"))
	testutil.AssertGo(t, "testdata/golden/client_cli_flag-presence-build.go.golden", builder)
}

// TestReleasedGRPCNamesMatchDeclarations verifies released plugins can read
// final method and validation names without choosing those names themselves.
func TestReleasedGRPCNamesMatchDeclarations(t *testing.T) {
	root := RunGRPCDSL(t, testdata.PayloadWithValidationsDSL)
	services := CreateGRPCServices(root)
	files := clientCLIFiles(services)
	require.Greater(t, len(files), 1)

	build, ok := files[1].SectionTemplates[1].Data.(*cli.BuildFunctionData)
	require.True(t, ok)
	servicePlan := services.servicePlans[0]
	declaration := services.cliPlan.builders[servicePlan.expression.GRPCEndpoints[0]]
	require.NotNil(t, declaration)
	require.Equal(t, declaration.Name(), build.Name)

	service := services.GRPCServices["PayloadWithValidation"]
	require.NotNil(t, service)
	require.NotEmpty(t, service.Endpoints)
	endpoint := service.Endpoints[0]
	require.Equal(t, endpoint.ProtoMethodName, endpoint.ClientMethodName)

	validationRoot := RunGRPCDSL(t, testdata.ElemValidationDSL)
	validationService := CreateGRPCServices(validationRoot).GRPCServices["ServiceElemValidation"]
	require.NotNil(t, validationService)
	require.NotEmpty(t, validationService.Endpoints)
	request := validationService.Endpoints[0].Request
	require.NotNil(t, request)
	require.NotNil(t, request.ServerConvert)
	validation := request.ServerConvert.Validation
	require.NotNil(t, validation)
	require.NotNil(t, validation.Declaration)
	require.Equal(t, validation.Declaration.Name(), validation.Name)
}
