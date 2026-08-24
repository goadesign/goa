// This file checks that generated gRPC clients and servers reject an empty
// required OneOf and reject a selected branch whose value is nil.
package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	d "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
)

func TestGenerateGRPCRequiredUnionValidators(t *testing.T) {
	root := codegen.RunDSL(t, requiredGRPCUnionValidationDSL)
	plan := mustTestPlan(t, "generated.local/gen", []eval.Root{root}, planTransportData)
	files, err := testServiceFiles(plan)
	require.NoError(t, err)
	transportFiles, err := testTransportFiles(plan)
	require.NoError(t, err)
	files = append(files, transportFiles...)

	dir := t.TempDir()
	writeGeneratedModule(t, dir, "generated.local")
	for _, file := range files {
		_, err := file.Render(dir)
		require.NoError(t, err)
	}
	writeGRPCRequiredUnionValidationTest(t, dir)
	runGeneratedTests(t, dir)
}

// requiredGRPCUnionValidationDSL creates request and response unions with the
// same branches so both generated checks must enforce the same rules.
func requiredGRPCUnionValidationDSL() {
	d.API("required-union", func() {})
	token := d.Type("Token", d.String)
	detail := d.Type("Detail", func() {
		d.Field(1, "label", d.String)
		d.Required("label")
	})
	inactive := d.Type("Inactive", func() {})
	request := d.Type("RequestChoice", func() {
		d.OneOf("choice", func() {
			d.Field(1, "number", d.Int, func() { d.Minimum(1) })
			d.Field(2, "detail", detail)
			d.Field(3, "inactive", inactive)
			d.Field(4, "blob", d.Bytes)
			d.Field(5, "token", token)
		})
		d.Required("choice")
	})
	response := d.Type("ResponseChoice", func() {
		d.OneOf("choice", func() {
			d.Field(1, "number", d.Int, func() { d.Minimum(1) })
			d.Field(2, "detail", detail)
			d.Field(3, "inactive", inactive)
			d.Field(4, "blob", d.Bytes)
			d.Field(5, "token", token)
		})
		d.Required("choice")
	})
	d.Service("validation", func() {
		d.Method("Exchange", func() {
			d.Payload(request)
			d.Result(response)
			d.GRPC(func() {})
		})
	})
}

// writeGRPCRequiredUnionValidationTest adds a test which calls the generated
// server and client validation functions.
func writeGRPCRequiredUnionValidationTest(t *testing.T, moduleDir string) {
	t.Helper()
	dir := filepath.Join(moduleDir, "uniontest")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatalf("create required union validation package: %v", err)
	}
	const source = `package uniontest_test

import (
	"errors"
	"testing"

	goa "goa.design/goa/v3/pkg"
	genclient "generated.local/gen/grpc/validation/client"
	genpb "generated.local/gen/grpc/validation/pb"
	genserver "generated.local/gen/grpc/validation/server"
)

func TestServerRequestValidator(t *testing.T) {
	valid := []*genpb.ExchangeRequest{
		{Choice: &genpb.ExchangeRequest_Number{Number: 1}},
		{Choice: &genpb.ExchangeRequest_Detail{Detail: &genpb.Detail{Label: "ready"}}},
		{Choice: &genpb.ExchangeRequest_Inactive{Inactive: &genpb.Inactive{}}},
		{Choice: &genpb.ExchangeRequest_Blob{Blob: []byte{}}},
		{Choice: &genpb.ExchangeRequest_Token{Token: "ready"}},
	}
	for _, message := range valid {
		if err := genserver.ValidateExchangeRequest(message); err != nil {
			t.Errorf("valid request branch failed: %v", err)
		}
	}

	var nilNumber *genpb.ExchangeRequest_Number
	assertErrorName(t, genserver.ValidateExchangeRequest(&genpb.ExchangeRequest{Choice: &genpb.ExchangeRequest_Number{Number: 0}}), goa.InvalidRange)
	assertMissingField(t, genserver.ValidateExchangeRequest(&genpb.ExchangeRequest{}), "choice", "\"choice\" is missing from message")
	assertMissingField(t, genserver.ValidateExchangeRequest(&genpb.ExchangeRequest{Choice: nilNumber}), "number", "\"number\" is missing from message.choice")
	assertMissingField(t, genserver.ValidateExchangeRequest(&genpb.ExchangeRequest{Choice: &genpb.ExchangeRequest_Detail{}}), "detail", "\"detail\" is missing from message.choice")
	assertMissingField(t, genserver.ValidateExchangeRequest(&genpb.ExchangeRequest{Choice: &genpb.ExchangeRequest_Inactive{}}), "inactive", "\"inactive\" is missing from message.choice")
	assertMissingField(t, genserver.ValidateExchangeRequest(&genpb.ExchangeRequest{Choice: &genpb.ExchangeRequest_Blob{}}), "blob", "\"blob\" is missing from message.choice")
}

func TestClientResponseValidator(t *testing.T) {
	valid := []*genpb.ExchangeResponse{
		{Choice: &genpb.ExchangeResponse_Number{Number: 1}},
		{Choice: &genpb.ExchangeResponse_Detail{Detail: &genpb.Detail{Label: "ready"}}},
		{Choice: &genpb.ExchangeResponse_Inactive{Inactive: &genpb.Inactive{}}},
		{Choice: &genpb.ExchangeResponse_Blob{Blob: []byte{}}},
		{Choice: &genpb.ExchangeResponse_Token{Token: "ready"}},
	}
	for _, message := range valid {
		if err := genclient.ValidateExchangeResponse(message); err != nil {
			t.Errorf("valid response branch failed: %v", err)
		}
	}

	var nilDetail *genpb.ExchangeResponse_Detail
	assertErrorName(t, genclient.ValidateExchangeResponse(&genpb.ExchangeResponse{Choice: &genpb.ExchangeResponse_Number{Number: 0}}), goa.InvalidRange)
	assertMissingField(t, genclient.ValidateExchangeResponse(&genpb.ExchangeResponse{}), "choice", "\"choice\" is missing from message")
	assertMissingField(t, genclient.ValidateExchangeResponse(&genpb.ExchangeResponse{Choice: nilDetail}), "detail", "\"detail\" is missing from message.choice")
	assertMissingField(t, genclient.ValidateExchangeResponse(&genpb.ExchangeResponse{Choice: &genpb.ExchangeResponse_Detail{}}), "detail", "\"detail\" is missing from message.choice")
	assertMissingField(t, genclient.ValidateExchangeResponse(&genpb.ExchangeResponse{Choice: &genpb.ExchangeResponse_Inactive{}}), "inactive", "\"inactive\" is missing from message.choice")
	assertMissingField(t, genclient.ValidateExchangeResponse(&genpb.ExchangeResponse{Choice: &genpb.ExchangeResponse_Blob{}}), "blob", "\"blob\" is missing from message.choice")
}

// assertErrorName checks that generated validation returned the expected Goa error name.
func assertErrorName(t *testing.T, err error, name string) {
	t.Helper()
	if err == nil {
		t.Errorf("expected %q error", name)
		return
	}
	var serviceError *goa.ServiceError
	if !errors.As(err, &serviceError) {
		t.Errorf("expected Goa service error, got %T: %v", err, err)
		return
	}
	if serviceError.Name != name {
		t.Errorf("expected %q, got %q", name, serviceError.Name)
	}
}

// assertMissingField checks the error name, field, and message returned for a missing protobuf value.
func assertMissingField(t *testing.T, err error, field, message string) {
	t.Helper()
	if err == nil {
		t.Errorf("expected missing field %q", field)
		return
	}
	var serviceError *goa.ServiceError
	if !errors.As(err, &serviceError) {
		t.Errorf("expected Goa service error, got %T: %v", err, err)
		return
	}
	if serviceError.Name != goa.MissingField {
		t.Errorf("expected %q, got %q", goa.MissingField, serviceError.Name)
	}
	if serviceError.Field == nil || *serviceError.Field != field {
		t.Errorf("expected field %q, got %#v", field, serviceError.Field)
	}
	if serviceError.Message != message {
		t.Errorf("expected message %q, got %q", message, serviceError.Message)
	}
}
`
	if err := os.WriteFile(filepath.Join(dir, "required_union_validation_test.go"), []byte(source), 0o600); err != nil {
		t.Fatalf("write required union validation test: %v", err)
	}
}
