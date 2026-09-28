// These tests run generated gRPC conversions and HTTP validators against the
// same required collections. Protobuf has no repeated-field or map presence;
// service construction still supplies required collections for JSON consumers.
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

// TestRequiredCollectionsPreserveTransportContracts runs generated wire,
// construction, and validation checks together.
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

// requiredCollectionsDSL keeps strict JSON and singular-field controls beside
// empty direct collections and recursive required/optional values.
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
	label := d.Type("Label", d.String)
	node := d.Type("CollectionNode", func() {
		d.Field(1, "items", d.ArrayOfRequired(item), "The node's required items.")
		d.Field(2, "labels", d.MapOf(d.String, d.String), "The node's required labels.")
		d.Field(3, "next", "CollectionNode", "An optional next node.")
		d.Field(4, "children", d.ArrayOf("CollectionNode"), "Optional child nodes.")
		d.Field(5, "byName", d.MapOf(d.String, "CollectionNode"), "Optional named nodes.")
		d.Required("items", "labels")
	})
	loose := d.Type("OptionalCollectionNode", func() {
		d.Field(1, "items", d.ArrayOfRequired(item), "The node's optional items.")
		d.Field(2, "labels", d.MapOf(d.String, d.String), "The node's optional labels.")
		d.Field(3, "next", "OptionalCollectionNode", "An optional next node.")
	})
	shape := d.Type("CollectionShape", func() {
		d.Field(1, "items", d.ArrayOfRequired(item), "Required items; empty is valid.")
		d.Field(2, "labels", d.MapOf(d.String, d.String), "Required labels; empty is valid.")
		d.Field(3, "optionalItems", d.ArrayOfRequired(item), "Optional direct items.")
		d.Field(4, "optionalLabels", d.MapOf(d.String, d.String), "Optional direct labels.")
		d.Field(9, "node", node, "A required node with required collections.")
		d.Field(10, "optionalNode", node, "An optional occurrence of the same node.")
		d.Field(11, "note", d.String, "An optional note, including empty text.")
		d.Field(14, "loose", loose, "A node whose collections remain optional.")
		d.Field(15, "names", d.ArrayOfRequired(label), "Required primitive alias elements.")
		d.Required("items", "labels", "node", "loose", "names")
	})
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
		d.Method("RoundTrip", func() {
			d.Payload(shape)
			d.Result(shape)
			d.GRPC(func() {})
		})
	})
}

const requiredCollectionsRuntimeTest = `package collections_test

import (
 "context"
 "encoding/json"
 "testing"

 "github.com/stretchr/testify/require"
 gencollections "generated.local/gen/collections"
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
  payload := gengrpcserver.NewExchangePayload(request)
  result := gengrpcclient.NewExchangeResult(response)
  require.NotNil(t, payload.Items)
  require.NotNil(t, payload.Labels)
  require.NotNil(t, result.Items)
  require.NotNil(t, result.Labels)
  require.Len(t, payload.Items, len(input.Items))
  require.Len(t, result.Items, len(input.Items))
  require.Equal(t, input.Labels, payload.Labels)
  require.Equal(t, input.Labels, result.Labels)
  for index, item := range input.Items {
   require.Equal(t, *item.Label, payload.Items[index].Label)
   require.Equal(t, *item.Label, result.Items[index].Label)
  }
  require.Equal(t, "good", payload.Detail.Label)
  require.Equal(t, "good", result.Detail.Label)
  require.False(t, payload.Enabled)
  require.False(t, result.Enabled)
  require.Zero(t, payload.Count)
  require.Zero(t, result.Count)
  require.Empty(t, payload.Text)
  require.Empty(t, result.Text)
 }
}

// TestConstructedCollectionsPreserveRecursion crosses actual wire
// bytes in both directions before comparing the complete service values.
func TestConstructedCollectionsPreserveRecursion(t *testing.T) {
 for _, populated := range []bool{false, true} {
  for _, optionalPresent := range []bool{false, true} {
   input := collectionShape(populated, optionalPresent)
   t.Run("request", func(t *testing.T) {
    data, err := proto.Marshal(gengrpcclient.NewProtoRoundTripRequest(input))
    require.NoError(t, err)
    message := new(genpb.RoundTripRequest)
    require.NoError(t, proto.Unmarshal(data, message))
    if !populated {
     require.Nil(t, message.Items)
     require.Nil(t, message.Labels)
     require.Nil(t, message.Names)
    }
    require.NoError(t, gengrpcserver.ValidateRoundTripRequest(message))
    result := gengrpcserver.NewRoundTripPayload(message)
    require.Equal(t, input, result)
   })
   t.Run("response", func(t *testing.T) {
    data, err := proto.Marshal(gengrpcserver.NewProtoRoundTripResponse(input))
    require.NoError(t, err)
    message := new(genpb.RoundTripResponse)
    require.NoError(t, proto.Unmarshal(data, message))
    if !populated {
     require.Nil(t, message.Items)
     require.Nil(t, message.Labels)
     require.Nil(t, message.Names)
    }
    require.NoError(t, gengrpcclient.ValidateRoundTripResponse(message))
    result := gengrpcclient.NewRoundTripResult(message)
    require.Equal(t, input, result)
   })
  }
 }
}

// TestMissingRequiredMessagesFailBeforeConstruction uses the full
// generated decoders so a missing message cannot reach an unguarded conversion.
func TestMissingRequiredMessagesFailBeforeConstruction(t *testing.T) {
 for _, field := range []string{"node", "loose"} {
  t.Run("request/"+field, func(t *testing.T) {
   input := gengrpcclient.NewProtoRoundTripRequest(collectionShape(false, false))
   switch field {
   case "node":
    input.Node = nil
   case "loose":
    input.Loose = nil
   }
   data, err := proto.Marshal(input)
   require.NoError(t, err)
   message := new(genpb.RoundTripRequest)
   require.NoError(t, proto.Unmarshal(data, message))
   result, err := gengrpcserver.DecodeRoundTripRequest(context.Background(), message, nil)
   require.Error(t, err)
   require.Nil(t, result)
  })
  t.Run("response/"+field, func(t *testing.T) {
   input := gengrpcserver.NewProtoRoundTripResponse(collectionShape(false, false))
   switch field {
   case "node":
    input.Node = nil
   case "loose":
    input.Loose = nil
   }
   data, err := proto.Marshal(input)
   require.NoError(t, err)
   message := new(genpb.RoundTripResponse)
   require.NoError(t, proto.Unmarshal(data, message))
   result, err := gengrpcclient.DecodeRoundTripResponse(context.Background(), message, nil, nil)
   require.Error(t, err)
   require.Nil(t, result)
  })
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
// collectionShape supplies empty required values and independently absent or
// present optional fields. Shared recursive nodes exercise reused helpers.
func collectionShape(populated, optionalPresent bool) *gencollections.CollectionShape {
 leaf := &gencollections.CollectionNode{
  Items: []*gencollections.Item{},
  Labels: map[string]string{},
 }
 result := &gencollections.CollectionShape{
  Items: []*gencollections.Item{},
  Labels: map[string]string{},
  Node: &gencollections.CollectionNode{
   Items: []*gencollections.Item{},
   Labels: map[string]string{},
   Next: leaf,
   Children: []*gencollections.CollectionNode{leaf},
   ByName: map[string]*gencollections.CollectionNode{"leaf": leaf},
  },
  Loose: &gencollections.OptionalCollectionNode{
   Next: &gencollections.OptionalCollectionNode{},
  },
  Names: []gencollections.Label{},
 }
 if populated {
  result.Items = []*gencollections.Item{{Label: "first"}, {Label: "second"}}
  result.Labels["key"] = "value"
  result.Names = []gencollections.Label{"one", "two"}
 }
 if optionalPresent {
  note := ""
  result.Note = &note
  result.OptionalItems = []*gencollections.Item{{Label: "optional"}}
  result.OptionalLabels = map[string]string{"optional": "value"}
  result.OptionalNode = leaf
 }
 return result
}

`
