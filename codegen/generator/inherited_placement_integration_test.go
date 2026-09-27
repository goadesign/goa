// These tests compile shared declarations through supported HTTP and gRPC shapes,
// including direct uses of inherited children from two evaluated design roots.
package generator

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service/testdata/inherited-placement/schema"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestGenerateInheritedPlacementAcrossTransports(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		var values schema.Types
		var packet expr.UserType
		first := codegen.RunDSL(t, func() {
			dsl.API("first", func() {})
			values = schema.Define()
			packet = dsl.Type("Packet", func() {
				dsl.Meta("struct:pkg:path", "shared/types")
				dsl.Field(1, "entries", dsl.ArrayOf(values.Entry))
				dsl.Required("entries")
			})
			inheritedTransportService("alpha", packet, values.Child, false)
		})
		second := codegen.RunDSL(t, func() {
			dsl.API("second", func() {})
			dsl.Type("AnotherRoot", func() {
				dsl.Meta("struct:pkg:path", "shared/types")
				dsl.Attribute("child", values.Child)
			})
			inheritedTransportService("beta", packet, values.Child, false)
		})
		roots := []eval.Root{first, second}
		if reverse {
			roots[0], roots[1] = roots[1], roots[0]
		}
		plan := mustTestPlan(t, "generated.local/gen", roots, planTransportData)
		files, err := testServiceFiles(plan)
		require.NoError(t, err)
		transports, err := testTransportFiles(plan)
		require.NoError(t, err)
		files, err = mergeFilesByPath(append(files, transports...))
		require.NoError(t, err)
		directory := t.TempDir()
		writeGeneratedModule(t, directory, "generated.local")
		for _, file := range files {
			_, err := file.Render(directory)
			require.NoError(t, err)
		}
		writeGeneratedContractTest(t, directory, filepath.Join("gen", "http", "alpha", "server"), inheritedHTTPExchange)
		writeGeneratedContractTest(t, directory, filepath.Join("gen", "grpc", "alpha", "server"), inheritedGRPCExchange)
		runGeneratedTests(t, directory)
		for _, value := range []expr.UserType{values.Text, values.Child, values.Nested, values.Entry} {
			require.NotContains(t, value.Attribute().Meta, "struct:pkg:path")
		}
	}
}

func TestGenerateExplicitPlacementAcrossTransports(t *testing.T) {
	codegen.RunDSL(t, func() {
		child := dsl.Type("Child", func() {
			dsl.Field(1, "value", dsl.String)
			dsl.Field(2, "next", "Child")
			dsl.Required("value")
		})
		inner := dsl.Type("Inner", func() {
			dsl.Meta("struct:pkg:path", "right/types")
			dsl.Field(1, "child", child)
			dsl.Field(2, "children", dsl.ArrayOf(child))
			dsl.Required("child")
		})
		outer := dsl.Type("Outer", func() {
			dsl.Meta("struct:pkg:path", "left/types")
			dsl.Field(1, "inner", inner)
			dsl.Required("inner")
		})
		inheritedTransportService("alpha", outer, child, true)
		inheritedTransportService("beta", outer, child, true)
	})
	registry := testRegistry("gen", testGenerator(planServiceData, testServiceFiles), testGenerator(planTransportData, testTransportFiles))
	directory := filepath.Join(t.TempDir(), codegen.Gendir)
	writeGeneratedModule(t, directory, "generated.local/gen")
	_, err := generate(filepath.Dir(directory), "gen", false, registry)
	require.NoError(t, err)
	writeGeneratedContractTest(t, directory, filepath.Join("http", "alpha", "server"), explicitHTTPExchange)
	runGeneratedTests(t, directory)
}

const explicitHTTPExchange = `package server
import (
    "encoding/json"
    "reflect"
    "testing"
    genclient "generated.local/gen/http/alpha/client"
    left "generated.local/gen/left/types"
    right "generated.local/gen/right/types"
)
func TestExplicitOwnerExchange(t *testing.T) {
    child := &right.Child{Value: "child", Next: &right.Child{Value: "next"}}
    input := &left.Outer{Inner: &right.Inner{Child: child, Children: []*right.Child{child}}}
    data, err := json.Marshal(genclient.NewExchangeRequestBody(input))
    if err != nil { t.Fatal(err) }
    var body ExchangeRequestBody
    if err := json.Unmarshal(data, &body); err != nil { t.Fatal(err) }
    result := NewExchangeOuter(&body)
    if !reflect.DeepEqual(input, result) { t.Fatalf("explicit owner graph changed: %#v", result) }
}
`

func inheritedTransportService(name string, packet, child expr.UserType, grpcExchange bool) {
	dsl.Service(name, func() {
		for _, entry := range []struct {
			method string
			value  expr.UserType
			grpc   bool
		}{{"Exchange", packet, grpcExchange}, {"Child", child, true}} {
			dsl.Method(entry.method, func() {
				dsl.Payload(entry.value)
				dsl.Result(entry.value)
				dsl.HTTP(func() { dsl.POST("/" + name + "/" + entry.method) })
				if entry.grpc {
					dsl.GRPC(func() {})
				}
			})
		}
	})
}

const inheritedHTTPExchange = `package server
import (
    "encoding/json"
    "reflect"
    "testing"
    genclient "generated.local/gen/http/alpha/client"
)
func TestInheritedHTTPGraph(t *testing.T) {
    bodies := []string{
        "{\"entries\":[{\"choice\":{\"type\":\"text\",\"value\":\"\"}}]}",
        "{\"entries\":[{\"choice\":{\"type\":\"child\",\"value\":{\"value\":\"child\",\"next\":{\"value\":\"next\"},\"leaf\":{\"value\":\"leaf\"}}}}]}",
        "{\"entries\":[{\"choice\":{\"type\":\"words\",\"value\":[\"word\"]}}]}",
        "{\"entries\":[{\"choice\":{\"type\":\"table\",\"value\":{\"key\":[\"word\"]}}}]}",
    }
    for _, input := range bodies {
        var body ExchangeRequestBody
        if err := json.Unmarshal([]byte(input), &body); err != nil { t.Fatal(err) }
        payload := NewExchangePacket(&body)
        request, err := json.Marshal(genclient.NewExchangeRequestBody(payload))
        if err != nil { t.Fatal(err) }
        var accepted ExchangeRequestBody
        if err := json.Unmarshal(request, &accepted); err != nil { t.Fatal(err) }
        response, err := json.Marshal(NewExchangeResponseBody(NewExchangePacket(&accepted)))
        if err != nil { t.Fatal(err) }
        var received genclient.ExchangeResponseBody
        if err := json.Unmarshal(response, &received); err != nil { t.Fatal(err) }
        result := genclient.NewExchangePacketOK(&received)
        if !reflect.DeepEqual(payload, result) { t.Fatalf("HTTP graph changed: %#v", result) }
    }
}
`

const inheritedGRPCExchange = `package server
import (
    "reflect"
    "testing"
    "google.golang.org/protobuf/proto"
    genclient "generated.local/gen/grpc/alpha/client"
    genpb "generated.local/gen/grpc/alpha/pb"
    shared "generated.local/gen/shared/types"
)
func TestInheritedGRPCGraph(t *testing.T) {
    input := &shared.Child{Value: "child", Next: &shared.Child{Value: "next"}, Leaf: &shared.NestedValue{Value: "leaf"}}
    request := genclient.NewProtoChildRequest(input)
    data, err := proto.Marshal(request)
    if err != nil { t.Fatal(err) }
    var accepted genpb.ChildRequest
    if err := proto.Unmarshal(data, &accepted); err != nil { t.Fatal(err) }
    payload := NewChildPayload(&accepted)
    data, err = proto.Marshal(NewProtoChildResponse(payload))
    if err != nil { t.Fatal(err) }
    var received genpb.ChildResponse
    if err := proto.Unmarshal(data, &received); err != nil { t.Fatal(err) }
    result := genclient.NewChildResult(&received)
    if !reflect.DeepEqual(input, result) { t.Fatalf("gRPC graph changed: %#v", result) }
}
`
