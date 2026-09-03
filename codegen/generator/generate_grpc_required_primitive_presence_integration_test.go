// This file checks that generated gRPC messages preserve required primitive
// presence across protobuf encoding and command-line JSON.
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

// TestGenerateGRPCRequiredPrimitivePresence proves that generated client,
// server, and command-line validators accept explicit zero values and reject
// omitted values.
func TestGenerateGRPCRequiredPrimitivePresence(t *testing.T) {
	root := codegen.RunDSL(t, requiredGRPCPrimitivePresenceDSL)
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
	writeGRPCRequiredPrimitivePresenceTest(t, dir)
	runGeneratedTests(t, dir)
}

// requiredGRPCPrimitivePresenceDSL defines object and top-level alias values
// used on both sides of one generated gRPC service.
func requiredGRPCPrimitivePresenceDSL() {
	d.API("required-primitive-presence", func() {})
	stringAlias := d.Type("StringAlias", d.String)
	bytesAlias := d.Type("BytesAlias", d.Bytes)
	anyAlias := d.Type("AnyAlias", d.Any)
	defaults := d.Type("PresenceDefaults", func() {
		d.Field(1, "bytes", d.Bytes, func() {
			d.Default([]byte("fallback"))
		})
		d.Field(2, "value", d.Any, func() {
			d.Default("fallback")
		})
	})
	message := d.Type("PresenceMessage", func() {
		d.Field(1, "boolean", d.Boolean)
		d.Field(2, "integer", d.Int)
		d.Field(3, "text", d.String)
		d.Field(4, "bytes", d.Bytes)
		d.Field(5, "text_alias", stringAlias)
		d.Field(6, "bytes_alias", bytesAlias)
		d.Field(7, "any_alias", anyAlias)
		d.Required("boolean", "integer", "text", "bytes", "text_alias", "bytes_alias", "any_alias")
	})
	d.Service("presence", func() {
		d.Method("Exchange", func() {
			d.Payload(message)
			d.Result(message)
			d.GRPC(func() {})
		})
		d.Method("Echo", func() {
			d.Payload(stringAlias)
			d.Result(stringAlias)
			d.GRPC(func() {})
		})
		d.Method("Plain", func() {
			d.Payload(d.String)
			d.Result(d.String)
			d.GRPC(func() {})
		})
		d.Method("Defaults", func() {
			d.Payload(defaults)
			d.GRPC(func() {})
		})
	})
}

// writeGRPCRequiredPrimitivePresenceTest adds tests that use the generated
// protobuf types, codecs, command-line builders, and validation functions
// together.
func writeGRPCRequiredPrimitivePresenceTest(t *testing.T, moduleDir string) {
	t.Helper()
	dir := filepath.Join(moduleDir, "presencetest")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatalf("create required primitive presence package: %v", err)
	}
	const source = `package presencetest_test

import (
	"bytes"
	"errors"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	goa "goa.design/goa/v3/pkg"
	genclient "generated.local/gen/grpc/presence/client"
	genpb "generated.local/gen/grpc/presence/pb"
	genserver "generated.local/gen/grpc/presence/server"
	genpresence "generated.local/gen/presence"
)

func TestServicePrimitiveFieldsRemainValues(t *testing.T) {
	message := genpresence.PresenceMessage{
		Boolean:    false,
		Integer:    0,
		Text:       "",
		Bytes:      []byte{},
		TextAlias:  "",
		BytesAlias: []byte{},
		AnyAlias:   nil,
	}
	if message.Boolean || message.Integer != 0 || message.Text != "" {
		t.Errorf("unexpected service primitive values: %#v", message)
	}
}

func TestRequiredRequestPrimitivePresence(t *testing.T) {
	message := roundTrip(t, validRequest())
	if err := genserver.ValidateExchangeRequest(message); err != nil {
		t.Errorf("explicit request zero values failed validation: %v", err)
	}

	missing := validRequest()
	missing.Boolean = nil
	assertMissingField(t, genserver.ValidateExchangeRequest(missing), "boolean")
	missing = validRequest()
	missing.Integer = nil
	assertMissingField(t, genserver.ValidateExchangeRequest(missing), "integer")
	missing = validRequest()
	missing.Text = nil
	assertMissingField(t, genserver.ValidateExchangeRequest(missing), "text")
	missing = validRequest()
	missing.Bytes_ = nil
	assertMissingField(t, genserver.ValidateExchangeRequest(missing), "bytes")
	missing = validRequest()
	missing.TextAlias = nil
	assertMissingField(t, genserver.ValidateExchangeRequest(missing), "text_alias")
	missing = validRequest()
	missing.BytesAlias = nil
	assertMissingField(t, genserver.ValidateExchangeRequest(missing), "bytes_alias")
	missing = validRequest()
	missing.AnyAlias = nil
	assertMissingField(t, genserver.ValidateExchangeRequest(missing), "any_alias")
}

func TestRequiredResponsePrimitivePresence(t *testing.T) {
	message := roundTrip(t, validResponse())
	if err := genclient.ValidateExchangeResponse(message); err != nil {
		t.Errorf("explicit response zero values failed validation: %v", err)
	}

	missing := validResponse()
	missing.Boolean = nil
	assertMissingField(t, genclient.ValidateExchangeResponse(missing), "boolean")
	missing = validResponse()
	missing.Integer = nil
	assertMissingField(t, genclient.ValidateExchangeResponse(missing), "integer")
	missing = validResponse()
	missing.Text = nil
	assertMissingField(t, genclient.ValidateExchangeResponse(missing), "text")
	missing = validResponse()
	missing.Bytes_ = nil
	assertMissingField(t, genclient.ValidateExchangeResponse(missing), "bytes")
	missing = validResponse()
	missing.TextAlias = nil
	assertMissingField(t, genclient.ValidateExchangeResponse(missing), "text_alias")
	missing = validResponse()
	missing.BytesAlias = nil
	assertMissingField(t, genclient.ValidateExchangeResponse(missing), "bytes_alias")
	missing = validResponse()
	missing.AnyAlias = nil
	assertMissingField(t, genclient.ValidateExchangeResponse(missing), "any_alias")
}

func TestRequiredTopLevelPrimitiveAliasPresence(t *testing.T) {
	empty := ""
	request := roundTrip(t, &genpb.StringAlias{Field: &empty})
	if err := genserver.ValidateStringAlias(request); err != nil {
		t.Errorf("explicit empty request alias failed validation: %v", err)
	}
	assertMissingField(t, genserver.ValidateStringAlias(&genpb.StringAlias{}), "field")

	response := roundTrip(t, &genpb.StringAlias{Field: &empty})
	if err := genclient.ValidateStringAlias(response); err != nil {
		t.Errorf("explicit empty response alias failed validation: %v", err)
	}
	assertMissingField(t, genclient.ValidateStringAlias(&genpb.StringAlias{}), "field")
}

func TestRequiredTopLevelPrimitivePresence(t *testing.T) {
	empty := ""
	request := roundTrip(t, &genpb.PlainRequest{Field: &empty})
	if err := genserver.ValidatePlainRequest(request); err != nil {
		t.Errorf("explicit empty request failed validation: %v", err)
	}
	assertMissingField(t, genserver.ValidatePlainRequest(&genpb.PlainRequest{}), "field")

	response := roundTrip(t, &genpb.PlainResponse{Field: &empty})
	if err := genclient.ValidatePlainResponse(response); err != nil {
		t.Errorf("explicit empty response failed validation: %v", err)
	}
	assertMissingField(t, genclient.ValidatePlainResponse(&genpb.PlainResponse{}), "field")

	explicitEmpty := "{\"field\":\"\"}"
	value, err := genclient.BuildPlainPayload(&explicitEmpty)
	if err != nil {
		t.Errorf("explicit empty CLI value failed validation: %v", err)
	}
	if value != "" {
		t.Errorf("expected empty CLI value, got %q", value)
	}
	missing := "{}"
	_, err = genclient.BuildPlainPayload(&missing)
	assertMissingField(t, err, "field")
}

func TestRequiredPrimitivePresenceFromCLIJSON(t *testing.T) {
	valid := "{\"boolean\":false,\"integer\":0,\"text\":\"\",\"bytes\":\"\"," +
		"\"textAlias\":\"\",\"bytesAlias\":\"\",\"anyAlias\":null}"
	payload, err := genclient.BuildExchangePayload(&valid)
	if err != nil {
		t.Errorf("explicit CLI zero values failed validation: %v", err)
	}
	if payload == nil {
		t.Error("explicit CLI zero values returned a nil payload")
	}

	missingBoolean := "{\"integer\":0,\"text\":\"\",\"bytes\":\"\"," +
		"\"textAlias\":\"\",\"bytesAlias\":\"\",\"anyAlias\":null}"
	_, err = genclient.BuildExchangePayload(&missingBoolean)
	assertMissingField(t, err, "boolean")

	explicitEmpty := "{\"field\":\"\"}"
	alias, err := genclient.BuildEchoPayload(&explicitEmpty)
	if err != nil {
		t.Errorf("explicit empty CLI alias failed validation: %v", err)
	}
	if alias != "" {
		t.Errorf("expected empty CLI alias, got %q", alias)
	}
	missing := "{}"
	_, err = genclient.BuildEchoPayload(&missing)
	assertMissingField(t, err, "field")
}

func TestOptionalNilableDefaultsUseWirePresence(t *testing.T) {
	omitted := genserver.NewDefaultsPayload(&genpb.DefaultsRequest{})
	if !bytes.Equal(omitted.Bytes, []byte("fallback")) {
		t.Errorf("omitted bytes got %#v", omitted.Bytes)
	}
	if omitted.Value != "fallback" {
		t.Errorf("omitted Any got %#v", omitted.Value)
	}

	present := genserver.NewDefaultsPayload(&genpb.DefaultsRequest{
		Bytes_: []byte{},
		Value:  structpb.NewNullValue(),
	})
	if present.Bytes == nil || len(present.Bytes) != 0 {
		t.Errorf("explicit empty bytes got %#v", present.Bytes)
	}
	if present.Value != nil {
		t.Errorf("explicit protobuf null got %#v", present.Value)
	}
}

func validRequest() *genpb.ExchangeRequest {
	boolean := false
	integer := int32(0)
	text := ""
	textAlias := ""
	return &genpb.ExchangeRequest{
		Boolean:    &boolean,
		Integer:    &integer,
		Text:       &text,
		Bytes_:     []byte{},
		TextAlias:  &textAlias,
		BytesAlias: []byte{},
		AnyAlias:   structpb.NewNullValue(),
	}
}

func validResponse() *genpb.ExchangeResponse {
	boolean := false
	integer := int32(0)
	text := ""
	textAlias := ""
	return &genpb.ExchangeResponse{
		Boolean:    &boolean,
		Integer:    &integer,
		Text:       &text,
		Bytes_:     []byte{},
		TextAlias:  &textAlias,
		BytesAlias: []byte{},
		AnyAlias:   structpb.NewNullValue(),
	}
}

func roundTrip[T proto.Message](t *testing.T, source T) T {
	t.Helper()
	encoded, err := proto.Marshal(source)
	if err != nil {
		t.Fatalf("marshal protobuf message: %v", err)
	}
	result := source.ProtoReflect().Type().New().Interface().(T)
	if err := proto.Unmarshal(encoded, result); err != nil {
		t.Fatalf("unmarshal protobuf message: %v", err)
	}
	return result
}

func assertMissingField(t *testing.T, err error, field string) {
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
}
`
	if err := os.WriteFile(filepath.Join(dir, "required_primitive_presence_test.go"), []byte(source), 0o600); err != nil {
		t.Fatalf("write required primitive presence test: %v", err)
	}
}
