// This file verifies that generated gRPC metadata codecs convert between
// native header values and relocated service aliases in both directions.
package generator

import (
	"os"
	"path/filepath"
	"testing"

	"goa.design/goa/v3/codegen"
	d "goa.design/goa/v3/dsl"
)

func TestGenerateGRPCMetadataAliasesCompile(t *testing.T) {
	registry := testRegistry(
		"gen",
		testGenerator(planServiceData, Service),
		testGenerator(planTransportData, Transport),
	)

	_ = codegen.RunDSL(t, func() {
		d.API("metadata", func() {})
		value := d.Type("Value", d.Int, func() {
			d.Enum(1, 2)
			d.Meta("struct:pkg:path", "shared/types")
		})
		values := d.Type("Values", d.ArrayOf(value), func() {
			d.Meta("struct:pkg:path", "shared/types")
		})
		payload := d.Type("Payload", func() {
			d.Field(1, "required_values", values)
			d.Field(2, "optional_value", value)
			d.Field(3, "anonymous_values", d.ArrayOf(value))
			d.Field(4, "optional_values", values)
			d.Required("required_values", "anonymous_values")
		})
		result := d.Type("Result", func() {
			d.Field(1, "header_values", values)
			d.Field(2, "trailer_value", value)
			d.Field(3, "optional_header_values", values)
			d.Required("header_values")
		})
		d.Service("Metadata", func() {
			d.Method("Exchange", func() {
				d.Payload(payload)
				d.Result(result)
				d.GRPC(func() {
					d.Metadata(func() {
						d.Attribute("required_values")
						d.Attribute("optional_value")
						d.Attribute("anonymous_values")
						d.Attribute("optional_values")
					})
					d.Response(func() {
						d.Headers(func() {
							d.Attribute("header_values")
							d.Attribute("optional_header_values")
						})
						d.Trailers(func() { d.Attribute("trailer_value") })
					})
				})
			})
		})
	})

	dir := t.TempDir()
	genDir := filepath.Join(dir, codegen.Gendir)
	writeGeneratedModule(t, genDir, "gen")
	if _, err := generate(dir, "gen", false, registry); err != nil {
		t.Fatalf("generate gRPC metadata module: %v", err)
	}
	writeGRPCMetadataRoundTripTest(t, genDir)
	runGeneratedTests(t, genDir)
}

// writeGRPCMetadataRoundTripTest adds consumer code outside the generated
// packages that exercises both metadata directions through their public API.
func writeGRPCMetadataRoundTripTest(t *testing.T, moduleDir string) {
	t.Helper()
	dir := filepath.Join(moduleDir, "roundtrip")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatalf("create metadata round-trip package: %v", err)
	}
	const source = `package roundtrip_test

import (
	"context"
	"testing"

	genclient "gen/grpc/metadata/client"
	genserver "gen/grpc/metadata/server"
	genmetadata "gen/metadata"
	gentypes "gen/shared/types"
	"google.golang.org/grpc/metadata"
)

func TestMetadataRoundTrip(t *testing.T) {
	ctx := context.Background()
	optional := gentypes.Value(2)
	payload := &genmetadata.Payload{
		RequiredValues: gentypes.Values{1, 2},
		OptionalValue:  &optional,
		AnonymousValues: []gentypes.Value{2, 1},
		OptionalValues: gentypes.Values{1},
	}
	requestMetadata := metadata.MD{}
	message, err := genclient.EncodeExchangeRequest(ctx, payload, &requestMetadata)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := genserver.DecodeExchangeRequest(ctx, message, requestMetadata)
	if err != nil {
		t.Fatal(err)
	}
	gotPayload := decoded.(*genmetadata.Payload)
	if len(gotPayload.RequiredValues) != 2 || gotPayload.RequiredValues[1] != 2 || *gotPayload.OptionalValue != 2 || len(gotPayload.AnonymousValues) != 2 || gotPayload.AnonymousValues[1] != 1 || len(gotPayload.OptionalValues) != 1 || gotPayload.OptionalValues[0] != 1 {
		t.Fatalf("unexpected payload: %#v", gotPayload)
	}
	for _, optionalValues := range []gentypes.Values{nil, {}} {
		payload.OptionalValues = optionalValues
		requestMetadata = metadata.MD{}
		message, err = genclient.EncodeExchangeRequest(ctx, payload, &requestMetadata)
		if err != nil {
			t.Fatal(err)
		}
		decoded, err = genserver.DecodeExchangeRequest(ctx, message, requestMetadata)
		if err != nil {
			t.Fatal(err)
		}
		if got := decoded.(*genmetadata.Payload); len(got.OptionalValues) != 0 {
			t.Fatalf("unexpected absent optional values: %#v", got.OptionalValues)
		}
	}

	trailer := gentypes.Value(1)
	result := &genmetadata.Result{
		HeaderValues: gentypes.Values{2, 1},
		TrailerValue: &trailer,
		OptionalHeaderValues: gentypes.Values{1},
	}
	headers, trailers := metadata.MD{}, metadata.MD{}
	response, err := genserver.EncodeExchangeResponse(ctx, result, &headers, &trailers)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err = genclient.DecodeExchangeResponse(ctx, response, headers, trailers)
	if err != nil {
		t.Fatal(err)
	}
	gotResult := decoded.(*genmetadata.Result)
	if len(gotResult.HeaderValues) != 2 || gotResult.HeaderValues[0] != 2 || *gotResult.TrailerValue != 1 || len(gotResult.OptionalHeaderValues) != 1 || gotResult.OptionalHeaderValues[0] != 1 {
		t.Fatalf("unexpected result: %#v", gotResult)
	}
}
`
	if err := os.WriteFile(filepath.Join(dir, "roundtrip_test.go"), []byte(source), 0o600); err != nil {
		t.Fatalf("write metadata round-trip test: %v", err)
	}
}
