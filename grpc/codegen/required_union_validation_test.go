// This file checks the validation functions generated for gRPC server requests
// and client responses which contain a required OneOf.
package codegen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/testutil"
	d "goa.design/goa/v3/dsl"
)

func TestRequiredUnionValidationUsesCompleteProtobufBranches(t *testing.T) {
	root := RunGRPCDSL(t, requiredUnionValidationDSL)
	services := CreateGRPCServices(root)
	service := services.Get("UnionValidation")
	for _, expected := range []struct {
		side validateKind
		name string
	}{
		{validateServer, "validatetest_20_api_UnionValidation_Detail_At_detail"},
		{validateClient, "validatetest_20_api_UnionValidation_Detail_At_detail"},
	} {
		var nested *protobufValidationRecord
		for _, validation := range service.protobuf.validators {
			if validation.side == expected.side &&
				validation.source.path == "choice.detail" {
				nested = validation
				break
			}
		}
		require.NotNil(t, nested)
		require.Equal(t, expected.name, nested.declaration.Name())
		require.Contains(t, nested.data.Def, `MissingFieldError("label", "detail")`)
	}

	for _, test := range []struct {
		name        string
		files       []*codegen.File
		sectionName string
		function    string
		golden      string
	}{
		{
			name:        "server",
			files:       serverTypeFiles(services),
			sectionName: "server-validate",
			function:    "ValidateExchangeRequest",
			golden:      "testdata/golden/server_types_server-required-union-validation.go.golden",
		},
		{
			name:        "client",
			files:       clientTypeFiles(services),
			sectionName: "client-validate",
			function:    "ValidateExchangeResponse",
			golden:      "testdata/golden/client_types_client-required-union-validation.go.golden",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			require.Len(t, test.files, 1)
			sections := test.files[0].Section(test.sectionName)
			require.NotEmpty(t, sections)
			testutil.AssertGo(t, test.golden, validationSection(t, sections, test.function))
		})
	}
}

// validationSection returns the generated validator named function.
func validationSection(t *testing.T, sections []*codegen.SectionTemplate, function string) string {
	t.Helper()
	for _, section := range sections {
		code := codegen.SectionCode(t, section)
		if strings.Contains(code, "func "+function+"(") {
			return code
		}
	}
	t.Errorf("missing generated validator %s", function)
	return ""
}

// requiredUnionValidationDSL creates branches whose values include scalars,
// messages, an empty message, a byte slice, a named string, and Any.
func requiredUnionValidationDSL() {
	token := d.Type("Token", d.String)
	detail := d.Type("Detail", func() {
		d.Field(1, "label", d.String)
		d.Required("label")
	})
	inactive := d.Type("Inactive", func() {})
	request := d.Type("RequestChoice", func() {
		d.OneOf("choice", func() {
			d.TypeName("RequestChoiceValue")
			d.Field(1, "number", d.Int, func() { d.Minimum(1) })
			d.Field(2, "detail", detail)
			d.Field(3, "inactive", inactive)
			d.Field(4, "blob", d.Bytes)
			d.Field(5, "token", token)
			d.Field(6, "metadata", d.Any)
		})
		d.Required("choice")
	})
	response := d.Type("ResponseChoice", func() {
		d.OneOf("choice", func() {
			d.TypeName("ResponseChoiceValue")
			d.Field(1, "number", d.Int, func() { d.Minimum(1) })
			d.Field(2, "detail", detail)
			d.Field(3, "inactive", inactive)
			d.Field(4, "blob", d.Bytes)
			d.Field(5, "token", token)
			d.Field(6, "metadata", d.Any)
		})
		d.Required("choice")
	})
	d.Service("UnionValidation", func() {
		d.Method("Exchange", func() {
			d.Payload(request)
			d.Result(response)
			d.GRPC(func() {})
		})
	})
}
