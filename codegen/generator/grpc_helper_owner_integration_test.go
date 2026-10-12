// These tests compile and run gRPC conversions whose inline union branches
// inherit different packages from their enclosing shared service types.
package generator

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
)

func TestGRPCInheritedUnionOwnersRoundTrip(t *testing.T) {
	root := codegen.RunDSL(t, grpcInheritedUnionOwnersDSL)
	plan := mustTestPlan(t, "generated.local/gen", []eval.Root{root}, planTransportData)
	files, err := testServiceFiles(plan)
	require.NoError(t, err)
	transportFiles, err := testTransportFiles(plan)
	require.NoError(t, err)
	files = append(files, transportFiles...)

	directory := t.TempDir()
	writeGeneratedModule(t, directory, "generated.local")
	for _, file := range files {
		_, err := file.Render(directory)
		require.NoError(t, err)
	}
	writeGeneratedContractTest(t, directory, "checks", grpcInheritedUnionRoundTrip)
	runGeneratedTests(t, directory)
}

// grpcInheritedUnionOwnersDSL shares an inline branch through Extend while
// placing each enclosing value in its own package. Request, response, and error
// conversions must retain that package for every generated branch reference.
func grpcInheritedUnionOwnersDSL() {
	dsl.API("inherited-owner", func() {})
	first := dsl.Type("First", func() {
		dsl.Meta("struct:pkg:path", "firsttypes")
		dsl.OneOf("choice", func() {
			dsl.TypeName("Selection")
			dsl.Field(1, "entries", func() {
				dsl.Field(1, "values", dsl.MapOf(dsl.String, dsl.String))
				dsl.Required("values")
			})
			dsl.Field(2, "text", dsl.String)
		})
		dsl.Required("choice")
	})
	second := dsl.Type("Second", func() {
		dsl.Meta("struct:pkg:path", "secondtypes")
		dsl.Extend(first)
	})
	packet := dsl.Type("Packet", func() {
		dsl.Field(1, "first", first)
		dsl.Field(2, "second", second)
		dsl.Required("first", "second")
	})
	dsl.Service("sample", func() {
		dsl.Method("exchange", func() {
			dsl.Payload(packet)
			dsl.Result(packet)
			dsl.Error("fault", packet)
			dsl.GRPC(func() {
				dsl.Response("fault", dsl.CodeInvalidArgument)
			})
		})
	})
}

const grpcInheritedUnionRoundTrip = `package checks

import (
	"reflect"
	"testing"

	genfirst "generated.local/gen/firsttypes"
	genclient "generated.local/gen/grpc/sample/client"
	genserver "generated.local/gen/grpc/sample/server"
	gensample "generated.local/gen/sample"
	gensecond "generated.local/gen/secondtypes"
)

func TestRequestResponseAndErrorsKeepInheritedBranchOwners(t *testing.T) {
	for _, text := range []bool{false, true} {
		input := &gensample.Packet{
			First: &genfirst.First{Choice: genfirst.NewSelectionEntries(&genfirst.SelectionBranchEntries{Values: map[string]string{"first": "value"}})},
			Second: &gensecond.Second{Choice: gensecond.NewSelectionEntries(&gensecond.SelectionBranchEntries{Values: map[string]string{"second": "value"}})},
		}
		if text {
			input.First.Choice = genfirst.NewSelectionText("first")
			input.Second.Choice = gensecond.NewSelectionText("second")
		}
		request := genclient.NewProtoExchangeRequest(input)
		if err := genserver.ValidateExchangeRequest(request); err != nil {
			t.Fatalf("validate request: %v", err)
		}
		payload := genserver.NewExchangePayload(request)
		response := genserver.NewProtoExchangeResponse(payload)
		if err := genclient.ValidateExchangeResponse(response); err != nil {
			t.Fatalf("validate response: %v", err)
		}
		output := genclient.NewExchangeResult(response)
		if !reflect.DeepEqual(input, output) {
			t.Errorf("round trip changed the selected values: input=%#v output=%#v", input, output)
		}
		fault := genserver.NewExchangeFaultError(input)
		errorOutput := genclient.NewExchangeFaultError(fault)
		if !reflect.DeepEqual(input, errorOutput) {
			t.Errorf("error round trip changed the selected values: input=%#v output=%#v", input, errorOutput)
		}
	}
}
`
