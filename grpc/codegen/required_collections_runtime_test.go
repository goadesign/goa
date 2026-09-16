// These tests run generated gRPC and HTTP validators against the same required
// collections. Protobuf has no collection presence; JSON still requires fields.
package codegen

import (
	"go/format"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen/service"
	d "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

func TestRequiredCollectionsPreserveTransportContracts(t *testing.T) {
	root := expr.RunDSL(t, requiredCollectionsDSL)
	generation, servicePlans := grpcServicePlans(t, []*expr.RootExpr{root})
	grpcPlans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlans[0]})
	require.NoError(t, err)
	httpPlans, err := httpcodegen.NewPlans(generation, httpcodegen.PlanInput{Root: root, Service: servicePlans[0]})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlans[0].Link())
	require.NoError(t, grpcPlans[0].Link())
	require.NoError(t, httpPlans[0].Link())
	files, err := service.Files(servicePlans...)
	require.NoError(t, err)
	files = append(files, grpcPlans[0].ServerFiles()...)
	files = append(files, grpcPlans[0].ClientFiles()...)
	files = append(files, grpcPlans[0].ServerTypeFiles()...)
	files = append(files, grpcPlans[0].ClientTypeFiles()...)
	files = append(files, grpcPlans[0].ProtoFiles()...)
	files = append(files, httpPlans[0].ServerTypeFiles()...)
	files = append(files, httpPlans[0].ServerFiles()...)
	files = append(files, httpPlans[0].PathFiles()...)
	directory := t.TempDir()
	writeProtobufDescriptorModule(t, directory)
	for _, file := range files {
		_, err := file.Render(directory)
		require.NoError(t, err)
	}
	source, err := format.Source([]byte(requiredCollectionsRuntimeTest))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(directory, "collections_test.go"), source, 0o600))
	compileProtobufDescriptorModule(t, directory)
}

func requiredCollectionsDSL() {
	item := d.Type("Item", func() {
		d.Field(1, "label", d.String)
		d.Required("label")
	})
	fields := func() {
		d.Field(1, "items", d.ArrayOfRequired(item))
		d.Field(2, "labels", d.MapOf(d.String, d.String))
		d.Field(3, "detail", item)
		d.Field(4, "enabled", d.Boolean)
		d.Field(5, "count", d.Int)
		d.Field(6, "text", d.String)
		d.Required("items", "labels", "detail", "enabled", "count", "text")
	}
	bounded := func() {
		d.Field(1, "items", d.ArrayOf(d.String), func() { d.MinLength(1) })
		d.Field(2, "labels", d.MapOf(d.String, d.String), func() { d.MinLength(1) })
		d.Required("items", "labels")
	}
	d.Service("Collections", func() {
		d.Method("Exchange", func() {
			d.Payload(fields)
			d.Result(fields)
			d.GRPC(func() {})
			d.HTTP(func() { d.POST("/exchange") })
		})
		d.Method("Bounded", func() {
			d.Payload(bounded)
			d.Result(bounded)
			d.GRPC(func() {})
		})
	})
}

const requiredCollectionsRuntimeTest = `package collections_test

import (
 "encoding/json"
 "testing"

 gengrpcclient "generated.local/gen/grpc/collections/client"
 gengrpcserver "generated.local/gen/grpc/collections/server"
 genpb "generated.local/gen/grpc/collections/pb"
 genhttpserver "generated.local/gen/http/collections/server"
 "google.golang.org/protobuf/proto"
)

func TestEmptyAndPopulatedCollectionsRoundTrip(t *testing.T) {
 for _, populated := range []bool{false, true} {
  input := &genpb.ExchangeRequest{
   Items: []*genpb.Item{}, Labels: map[string]string{},
   Detail: &genpb.Item{Label: proto.String("good")},
   Enabled: proto.Bool(false), Count: proto.Int32(0), Text: proto.String(""),
  }
  if populated {
   input.Items = []*genpb.Item{{Label: proto.String("item")}}
   input.Labels["key"] = "value"
  }
  wire, err := proto.Marshal(input)
  if err != nil { t.Fatal(err) }
  request := new(genpb.ExchangeRequest)
  response := new(genpb.ExchangeResponse)
  if err := proto.Unmarshal(wire, request); err != nil { t.Fatal(err) }
  if err := proto.Unmarshal(wire, response); err != nil { t.Fatal(err) }
  if !populated && (request.Items != nil || request.Labels != nil) {
   t.Fatal("test must exercise protobuf's loss of empty collection presence")
  }
  if err := gengrpcserver.ValidateExchangeRequest(request); err != nil { t.Fatal(err) }
  if err := gengrpcclient.ValidateExchangeResponse(response); err != nil { t.Fatal(err) }
 }
}

func TestRequiredNonCollectionFieldsAndElementsRemainValidated(t *testing.T) {
 for _, field := range []string{"detail", "enabled", "count", "text", "item"} {
  t.Run(field, func(t *testing.T) {
   request := &genpb.ExchangeRequest{
    Detail: &genpb.Item{Label: proto.String("good")},
    Enabled: proto.Bool(false), Count: proto.Int32(0), Text: proto.String(""),
   }
   switch field {
   case "detail": request.Detail = nil
   case "enabled": request.Enabled = nil
   case "count": request.Count = nil
   case "text": request.Text = nil
   case "item": request.Items = []*genpb.Item{{}}
   }
   wire, err := proto.Marshal(request)
   if err != nil { t.Fatal(err) }
   response := new(genpb.ExchangeResponse)
   if err := proto.Unmarshal(wire, response); err != nil { t.Fatal(err) }
   if err := gengrpcserver.ValidateExchangeRequest(request); err == nil { t.Fatal("server accepted missing required " + field) }
   if err := gengrpcclient.ValidateExchangeResponse(response); err == nil { t.Fatal("client accepted missing required " + field) }
  })
 }
 request := &genpb.ExchangeRequest{Items: []*genpb.Item{nil}, Detail: &genpb.Item{Label: proto.String("good")}, Enabled: proto.Bool(false), Count: proto.Int32(0), Text: proto.String("")}
 if err := gengrpcserver.ValidateExchangeRequest(request); err == nil { t.Fatal("accepted nil required element") }
}

func TestCollectionMinimumLengthsRemainValidated(t *testing.T) {
 for _, field := range []string{"items", "labels"} {
  request := &genpb.BoundedRequest{Items: []string{"one"}, Labels: map[string]string{"key":"value"}}
  if field == "items" { request.Items = nil } else { request.Labels = nil }
  wire, err := proto.Marshal(request)
  if err != nil { t.Fatal(err) }
  response := new(genpb.BoundedResponse)
  if err := proto.Unmarshal(wire, response); err != nil { t.Fatal(err) }
  if err := gengrpcserver.ValidateBoundedRequest(request); err == nil { t.Fatal("server ignored minimum length for " + field) }
  if err := gengrpcclient.ValidateBoundedResponse(response); err == nil { t.Fatal("client ignored minimum length for " + field) }
 }
}

func TestJSONStillRequiresPresentNonNullCollections(t *testing.T) {
 for _, test := range []struct{ name, fields string; valid bool }{
  {"empty", "\"items\":[],\"labels\":{}", true},
  {"populated", "\"items\":[{\"label\":\"item\"}],\"labels\":{\"key\":\"value\"}", true},
  {"missing items", "\"labels\":{}", false},
  {"missing labels", "\"items\":[]", false},
  {"null items", "\"items\":null,\"labels\":{}", false},
  {"null labels", "\"items\":[],\"labels\":null", false},
 } {
  t.Run(test.name, func(t *testing.T) {
   data := "{" + test.fields + ",\"detail\":{\"label\":\"good\"},\"enabled\":false,\"count\":0,\"text\":\"\"}"
   var body genhttpserver.ExchangeRequestBody
   if err := json.Unmarshal([]byte(data), &body); err != nil { t.Fatal(err) }
   err := genhttpserver.ValidateExchangeRequestBody(&body)
   if (err == nil) != test.valid { t.Fatalf("valid=%v, error=%v", test.valid, err) }
  })
 }
}
`
