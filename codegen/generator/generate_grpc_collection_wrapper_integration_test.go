// This file checks that generated gRPC conversions preserve protobuf's
// collection rules when arrays appear as map values.
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

// TestGenerateGRPCCollectionMapValues runs generated request and response
// conversions with nil and empty protobuf wrappers around repeated fields.
func TestGenerateGRPCCollectionMapValues(t *testing.T) {
	root := codegen.RunDSL(t, grpcCollectionMapValuesDSL)
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
	writeGRPCCollectionMapValuesTest(t, dir)
	runGeneratedTests(t, dir)
}

// grpcCollectionMapValuesDSL defines an optional map whose values are arrays.
// It has no length rule, so protobuf nil and empty arrays are both valid.
func grpcCollectionMapValuesDSL() {
	d.API("collection-map-values", func() {})
	collection := d.Type("Collection", func() {
		d.Field(1, "values", d.MapOf(d.String, d.ArrayOf(d.String)))
	})
	constrained := d.Type("ConstrainedCollection", func() {
		d.Field(1, "values", d.MapOf(d.String, d.ArrayOf(d.String, func() {
			d.Pattern("^ok$")
		})), func() {
			d.Elem(func() {
				d.MinLength(1)
			})
		})
	})
	d.Service("collections", func() {
		d.Method("Exchange", func() {
			d.Payload(collection)
			d.Result(collection)
			d.GRPC(func() {})
		})
		d.Method("Check", func() {
			d.Payload(constrained)
			d.GRPC(func() {})
		})
	})
}

// writeGRPCCollectionMapValuesTest writes tests that use the generated
// protobuf types and conversion functions together.
func writeGRPCCollectionMapValuesTest(t *testing.T, moduleDir string) {
	t.Helper()
	dir := filepath.Join(moduleDir, "collectionmaptest")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatalf("create collection map test package: %v", err)
	}
	const source = `package collectionmaptest_test

import (
	"errors"
	"testing"

	goa "goa.design/goa/v3/pkg"
	genclient "generated.local/gen/grpc/collections/client"
	genpb "generated.local/gen/grpc/collections/pb"
	genserver "generated.local/gen/grpc/collections/server"
	gencollections "generated.local/gen/collections"
)

func TestRequestMapValuesAcceptNilAndEmptyWrappers(t *testing.T) {
	message := &genpb.ExchangeRequest{Values: map[string]*genpb.ArrayOfString{
		"nil":   nil,
		"empty": {},
	}}
	payload := genserver.NewExchangePayload(message)
	assertEmptyMapValues(t, payload.Values)
}

func TestResponseMapValuesRoundTripAsEmptyCollections(t *testing.T) {
	result := &gencollections.Collection{Values: map[string][]string{
		"nil":   nil,
		"empty": {},
	}}
	message := genserver.NewProtoExchangeResponse(result)
	converted := genclient.NewExchangeResult(message)
	assertEmptyMapValues(t, converted.Values)
}

func TestNilAndEmptyWrappersKeepAuthoredLengthRules(t *testing.T) {
	tests := map[string]*genpb.ArrayOfString{
		"nil":   nil,
		"empty": {},
	}
	var messages []string
	for name, wrapper := range tests {
		t.Run(name, func(t *testing.T) {
			message := &genpb.CheckRequest{Values: map[string]*genpb.ArrayOfString{"value": wrapper}}
			err := genserver.ValidateCheckRequest(message)
			if err == nil {
				t.Error("empty map value passed its minimum length rule")
				return
			}
			assertServiceErrorName(t, err, goa.InvalidLength)
			messages = append(messages, err.Error())
		})
	}
	if len(messages) == 2 && messages[0] != messages[1] {
		t.Errorf("nil and empty wrappers returned different errors: %q and %q", messages[0], messages[1])
	}
}

func TestPresentWrapperKeepsAuthoredItemRules(t *testing.T) {
	message := &genpb.CheckRequest{Values: map[string]*genpb.ArrayOfString{
		"value": {Field: []string{"not-ok"}},
	}}
	if err := genserver.ValidateCheckRequest(message); err == nil {
		t.Error("invalid repeated item passed its pattern rule")
	} else {
		assertServiceErrorName(t, err, goa.InvalidPattern)
	}
}

func assertServiceErrorName(t *testing.T, err error, name string) {
	t.Helper()
	var serviceError *goa.ServiceError
	if !errors.As(err, &serviceError) {
		t.Fatalf("expected service error, got %T: %v", err, err)
	}
	if serviceError.Name != name {
		t.Errorf("expected %q, got %q: %v", name, serviceError.Name, err)
	}
}

func assertEmptyMapValues(t *testing.T, values map[string][]string) {
	t.Helper()
	for _, key := range []string{"nil", "empty"} {
		value, ok := values[key]
		if !ok {
			t.Errorf("converted map is missing %q", key)
			continue
		}
		if len(value) != 0 {
			t.Errorf("converted map value %q has length %d", key, len(value))
		}
	}
}
`
	if err := os.WriteFile(filepath.Join(dir, "collection_map_test.go"), []byte(source), 0o600); err != nil {
		t.Fatalf("write collection map test: %v", err)
	}
}
