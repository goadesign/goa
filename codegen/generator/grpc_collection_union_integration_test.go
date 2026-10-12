// These tests compile generated gRPC code and send collection union branches
// through protobuf serialization without changing their service representation.
package generator

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
)

func TestGRPCCollectionUnionRoundTrip(t *testing.T) {
	root := codegen.RunDSL(t, grpcCollectionUnionDSL)
	plan := mustTestPlan(t, "generated.local/gen", []eval.Root{root}, planTransportData)
	files, err := testServiceFiles(plan)
	require.NoError(t, err)
	transport, err := testTransportFiles(plan)
	require.NoError(t, err)
	files = append(files, transport...)
	directory := t.TempDir()
	writeGeneratedModule(t, directory, "generated.local")
	for _, file := range files {
		_, err := file.Render(directory)
		require.NoError(t, err)
	}
	writeGeneratedContractTest(t, directory, "checks", grpcCollectionUnionRoundTrip)
	runGeneratedTests(t, directory)
}

// grpcCollectionUnionDSL uses the same branches directly and through Extend.
// Named and inline collections must keep their service packages while each
// collection branch becomes one protobuf message containing the collection.
func grpcCollectionUnionDSL() {
	dsl.API("collection-unions", func() {})
	entry := dsl.Type("Entry", func() {
		dsl.Meta("struct:pkg:path", "shared")
		dsl.Field(1, "value", dsl.String, func() { dsl.MinLength(1) })
		dsl.Required("value")
	})
	labels := dsl.Type("Labels", dsl.MapOf(dsl.String, dsl.Bytes), func() {
		dsl.Meta("struct:pkg:path", "shared")
	})
	names := dsl.Type("Names", dsl.ArrayOf(dsl.String), func() {
		dsl.MinLength(1)
		dsl.Meta("struct:pkg:path", "shared")
	})
	nested := dsl.Type("Nested", func() {
		dsl.Meta("struct:pkg:path", "shared")
		dsl.OneOf("detail", func() {
			dsl.TypeName("Detail")
			dsl.Field(1, "numbers", dsl.ArrayOf(dsl.Int))
			dsl.Field(2, "text", dsl.String)
		})
		dsl.Required("detail")
	})
	direct := dsl.Type("Direct", func() {
		dsl.Meta("struct:pkg:path", "firsttypes")
		dsl.OneOf("choice", func() {
			dsl.TypeName("Selection")
			dsl.Field(1, "records", dsl.MapOf(dsl.String, entry))
			dsl.Field(2, "numbers", dsl.ArrayOf(dsl.Int))
			dsl.Field(3, "labels", labels)
			dsl.Field(4, "names", names)
			dsl.Field(5, "nested", nested)
		})
		dsl.Required("choice")
	})
	inherited := dsl.Type("Inherited", func() {
		dsl.Meta("struct:pkg:path", "secondtypes")
		dsl.Extend(direct)
	})
	packet := dsl.Type("Packet", func() {
		dsl.Field(1, "direct", direct)
		dsl.Field(2, "inherited", inherited)
		dsl.Required("direct", "inherited")
	})
	dsl.Service("sample", func() {
		dsl.Method("exchange", func() {
			dsl.Payload(packet)
			dsl.Result(packet)
			dsl.Error("fault", packet)
			dsl.GRPC(func() { dsl.Response("fault", dsl.CodeInvalidArgument) })
		})
	})
}

const grpcCollectionUnionRoundTrip = `package checks

import (
	"reflect"
	"testing"

	"google.golang.org/protobuf/proto"

	genfirst "generated.local/gen/firsttypes"
	genclient "generated.local/gen/grpc/sample/client"
	genpb "generated.local/gen/grpc/sample/pb"
	genserver "generated.local/gen/grpc/sample/server"
	gensample "generated.local/gen/sample"
	gensecond "generated.local/gen/secondtypes"
	genshared "generated.local/gen/shared"
)

func TestCollectionUnionRequestResponseAndErrors(t *testing.T) {
	tests := []struct {
		name string
		first genfirst.Selection
		second gensecond.Selection
	}{
		{"map objects", genfirst.NewSelectionRecords(map[string]*genshared.Entry{"one": {Value: "value"}}), gensecond.NewSelectionRecords(map[string]*genshared.Entry{"two": {Value: "other"}})},
		{"empty map", genfirst.NewSelectionRecords(map[string]*genshared.Entry{}), gensecond.NewSelectionRecords(map[string]*genshared.Entry{})},
		{"array primitives", genfirst.NewSelectionNumbers([]int{0, 1, -2}), gensecond.NewSelectionNumbers([]int{3, 4})},
		{"empty array", genfirst.NewSelectionNumbers([]int{}), gensecond.NewSelectionNumbers([]int{})},
		{"named maps", genfirst.NewSelectionLabels(genshared.Labels{"key": []byte("value")}), gensecond.NewSelectionLabels(genshared.Labels{"key": []byte("other")})},
		{"named arrays", genfirst.NewSelectionNames(genshared.Names{"first"}), gensecond.NewSelectionNames(genshared.Names{"second"})},
		{"nested union", genfirst.NewSelectionNested(&genshared.Nested{Detail: genshared.NewDetailNumbers([]int{1, 2})}), gensecond.NewSelectionNested(&genshared.Nested{Detail: genshared.NewDetailText("value")})},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := &gensample.Packet{Direct: &genfirst.Direct{Choice: test.first}, Inherited: &gensecond.Inherited{Choice: test.second}}
			request := genclient.NewProtoExchangeRequest(input)
			data, err := proto.Marshal(request)
			if err != nil {
				t.Fatal(err)
			}
			decoded := new(genpb.ExchangeRequest)
			if err := proto.Unmarshal(data, decoded); err != nil {
				t.Fatal(err)
			}
			if err := genserver.ValidateExchangeRequest(decoded); err != nil {
				t.Fatal(err)
			}
			payload := genserver.NewExchangePayload(decoded)
			response := genserver.NewProtoExchangeResponse(payload)
			data, err = proto.Marshal(response)
			if err != nil {
				t.Fatal(err)
			}
			decodedResponse := new(genpb.ExchangeResponse)
			if err := proto.Unmarshal(data, decodedResponse); err != nil {
				t.Fatal(err)
			}
			if err := genclient.ValidateExchangeResponse(decodedResponse); err != nil {
				t.Fatal(err)
			}
			output := genclient.NewExchangeResult(decodedResponse)
			if !reflect.DeepEqual(input, output) {
				t.Errorf("request and response changed collection values: got %#v want %#v", output, input)
			}
			fault := genserver.NewExchangeFaultError(input)
			data, err = proto.Marshal(fault)
			if err != nil {
				t.Fatal(err)
			}
			decodedFault := new(genpb.ExchangeFaultError)
			if err := proto.Unmarshal(data, decodedFault); err != nil {
				t.Fatal(err)
			}
			if err := genclient.ValidateExchangeFaultError(decodedFault); err != nil {
				t.Fatal(err)
			}
			errorOutput := genclient.NewExchangeFaultError(decodedFault)
			if !reflect.DeepEqual(input, errorOutput) {
				t.Errorf("error changed collection values: got %#v want %#v", errorOutput, input)
			}
		})
	}
}

func TestCollectionUnionRejectsMissingWrappersAndInvalidValues(t *testing.T) {
	tests := []struct {
		name string
		mutate func(*genpb.ExchangeRequest)
	}{
		{"missing branch", func(request *genpb.ExchangeRequest) {
			request.Direct.Choice = nil
		}},
		{"missing collection wrapper", func(request *genpb.ExchangeRequest) {
			request.Direct.Choice = &genpb.Direct_Records{}
		}},
		{"invalid map value", func(request *genpb.ExchangeRequest) {
			request.Direct.GetRecords().Field["key"].Value = nil
		}},
		{"invalid array length", func(request *genpb.ExchangeRequest) {
			request.Direct.Choice = &genpb.Direct_Names{Names: &genpb.Names{}}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := &gensample.Packet{
				Direct: &genfirst.Direct{Choice: genfirst.NewSelectionRecords(map[string]*genshared.Entry{"key": {Value: "value"}})},
				Inherited: &gensecond.Inherited{Choice: gensecond.NewSelectionNames(genshared.Names{"value"})},
			}
			request := genclient.NewProtoExchangeRequest(input)
			test.mutate(request)
			if err := genserver.ValidateExchangeRequest(request); err == nil {
				t.Error("invalid collection branch passed validation")
			}
		})
	}
}

`
