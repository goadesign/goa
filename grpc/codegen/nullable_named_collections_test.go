// These tests run generated gRPC conversions with nullable named object entries.
// They check in-memory positions separately from protobuf wire behavior.
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
)

// TestNullableNamedCollectionsPreserveConversionContracts generates both gRPC
// sides and runs nullable conversion, required validation, and wire checks.
func TestNullableNamedCollectionsPreserveConversionContracts(t *testing.T) {
	root := RunGRPCDSL(t, nullableNamedCollectionsDSL)
	generation, servicePlans := grpcServicePlans(t, []*expr.RootExpr{root})
	plans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlans[0]})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlans[0].Link())
	require.NoError(t, plans[0].Link())
	files, err := service.Files(servicePlans...)
	require.NoError(t, err)
	files = append(files, plans[0].ServerFiles()...)
	files = append(files, plans[0].ClientFiles()...)
	files = append(files, plans[0].ServerTypeFiles()...)
	files = append(files, plans[0].ClientTypeFiles()...)
	files = append(files, plans[0].ProtoFiles()...)
	directory := t.TempDir()
	writeProtobufDescriptorModule(t, directory)
	for _, file := range files {
		_, err := file.Render(directory)
		require.NoError(t, err)
	}
	source, err := format.Source([]byte(nullableNamedCollectionsRuntimeTest))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(directory, "collections_test.go"), source, 0o600))
	compileProtobufDescriptorModule(t, directory)
}

// nullableNamedCollectionsDSL uses named objects in nullable arrays and maps,
// beside arrays whose validators must reject null entries.
func nullableNamedCollectionsDSL() {
	entry := d.Type("Entry", func() {
		d.Field(1, "label", d.String)
	})
	choice := d.Type("ChoiceEntry", func() {
		d.OneOf("choice", func() {
			d.TypeName("ChoiceValue")
			d.Field(1, "text", d.String)
		})
		d.Required("choice")
	})
	envelope := d.Type("Envelope", func() {
		d.Field(1, "loose", d.ArrayOf(entry))
		d.Field(2, "strict", d.ArrayOfRequired(entry))
		d.Field(3, "loose_choices", d.ArrayOf(choice))
		d.Field(4, "strict_choices", d.ArrayOfRequired(choice))
		d.Field(5, "entries", d.MapOf(d.String, entry))
		d.Field(6, "choices", d.MapOf(d.String, choice))
	})
	d.Service("records", func() {
		d.Method("send", func() {
			d.Payload(envelope)
			d.Result(envelope)
			d.GRPC(func() {})
		})
	})
}

const nullableNamedCollectionsRuntimeTest = `package collections_test

import (
 "testing"

 "github.com/stretchr/testify/require"
 genclient "generated.local/gen/grpc/records/client"
 genpb "generated.local/gen/grpc/records/pb"
 genserver "generated.local/gen/grpc/records/server"
 genrecords "generated.local/gen/records"
 "google.golang.org/protobuf/encoding/protojson"
 "google.golang.org/protobuf/proto"
)

func TestNullableNamedEntriesKeepPositions(t *testing.T) {
 source := &genrecords.Envelope{
  Loose: []*genrecords.Entry{{Label: proto.String("first")}, nil, {Label: proto.String("last")}},
  LooseChoices: []*genrecords.ChoiceEntry{
   {Choice: genrecords.NewChoiceValueText("first")},
   nil,
   {Choice: genrecords.NewChoiceValueText("last")},
  },
 }
 t.Run("request", func(t *testing.T) {
  message := genclient.NewProtoSendRequest(source)
  require.Len(t, message.Loose, 3)
  require.Nil(t, message.Loose[1])
  require.Len(t, message.LooseChoices, 3)
  require.Nil(t, message.LooseChoices[1])
  require.NoError(t, genserver.ValidateSendRequest(message))
  require.Equal(t, source, genserver.NewSendPayload(message))
 })
 t.Run("response", func(t *testing.T) {
  message := genserver.NewProtoSendResponse(source)
  require.Len(t, message.Loose, 3)
  require.Nil(t, message.Loose[1])
  require.Len(t, message.LooseChoices, 3)
  require.Nil(t, message.LooseChoices[1])
  require.NoError(t, genclient.ValidateSendResponse(message))
  require.Equal(t, source, genclient.NewSendResult(message))
 })
}

func TestNullableNamedMapValuesKeepKeys(t *testing.T) {
 source := &genrecords.Envelope{
  Entries: map[string]*genrecords.Entry{
   "first": {Label: proto.String("first")}, "missing": nil, "last": {Label: proto.String("last")}, "empty": {},
  },
  Choices: map[string]*genrecords.ChoiceEntry{
   "first": {Choice: genrecords.NewChoiceValueText("first")}, "missing": nil, "last": {Choice: genrecords.NewChoiceValueText("last")},
  },
 }
 t.Run("request", func(t *testing.T) {
  message := genclient.NewProtoSendRequest(source)
  require.Len(t, message.Entries, 4)
  require.Contains(t, message.Entries, "missing")
  require.Nil(t, message.Entries["missing"])
  require.NotNil(t, message.Entries["empty"])
  require.Len(t, message.Choices, 3)
  require.Contains(t, message.Choices, "missing")
  require.Nil(t, message.Choices["missing"])
  require.NoError(t, genserver.ValidateSendRequest(message))
  require.Equal(t, source, genserver.NewSendPayload(message))
 })
 t.Run("response", func(t *testing.T) {
  message := genserver.NewProtoSendResponse(source)
  require.Len(t, message.Entries, 4)
  require.Contains(t, message.Entries, "missing")
  require.Nil(t, message.Entries["missing"])
  require.NotNil(t, message.Entries["empty"])
  require.Len(t, message.Choices, 3)
  require.Contains(t, message.Choices, "missing")
  require.Nil(t, message.Choices["missing"])
  require.NoError(t, genclient.ValidateSendResponse(message))
  require.Equal(t, source, genclient.NewSendResult(message))
 })
}

func TestNamedMapValuesOnTheWire(t *testing.T) {
 for _, choice := range []bool{false, true} {
  before := new(genpb.SendRequest)
  if choice {
   before.Choices = map[string]*genpb.ChoiceEntry{"missing": nil}
  } else {
   before.Entries = map[string]*genpb.Entry{"missing": nil, "next": {Label: proto.String("next")}}
  }
  require.NoError(t, genserver.ValidateSendRequest(before))
  data, err := proto.Marshal(before)
  require.NoError(t, err)
  after := new(genpb.SendRequest)
  require.NoError(t, proto.Unmarshal(data, after))
  response := new(genpb.SendResponse)
  require.NoError(t, proto.Unmarshal(data, response))
  if choice {
   require.Contains(t, after.Choices, "missing")
   require.NotNil(t, after.Choices["missing"])
   require.Error(t, genserver.ValidateSendRequest(after))
   require.Error(t, genclient.ValidateSendResponse(response))
  } else {
   require.NoError(t, genserver.ValidateSendRequest(after))
   require.NoError(t, genclient.ValidateSendResponse(response))
   result := genserver.NewSendPayload(after)
   require.Contains(t, result.Entries, "missing")
   require.NotNil(t, result.Entries["missing"])
   require.Nil(t, result.Entries["missing"].Label)
   require.Equal(t, "next", *result.Entries["next"].Label)
  }
 }
 require.Error(t, protojson.Unmarshal([]byte("{\"entries\":{\"missing\":null}}"), new(genpb.SendRequest)))
}

func TestPresentEmptyNamedEntriesRemainPresent(t *testing.T) {
 source := &genrecords.Envelope{Loose: []*genrecords.Entry{{}}, Strict: []*genrecords.Entry{{}}}
 request := genclient.NewProtoSendRequest(source)
 require.NotNil(t, request.Loose[0])
 require.NotNil(t, request.Strict[0])
 require.NoError(t, genserver.ValidateSendRequest(request))
 require.Equal(t, source, genserver.NewSendPayload(request))
 data, err := proto.Marshal(request)
 require.NoError(t, err)
 decodedRequest := new(genpb.SendRequest)
 require.NoError(t, proto.Unmarshal(data, decodedRequest))
 require.NoError(t, genserver.ValidateSendRequest(decodedRequest))
 require.Equal(t, source, genserver.NewSendPayload(decodedRequest))
 response := genserver.NewProtoSendResponse(source)
 require.NotNil(t, response.Loose[0])
 require.NotNil(t, response.Strict[0])
 require.NoError(t, genclient.ValidateSendResponse(response))
 require.Equal(t, source, genclient.NewSendResult(response))
 data, err = proto.Marshal(response)
 require.NoError(t, err)
 decodedResponse := new(genpb.SendResponse)
 require.NoError(t, proto.Unmarshal(data, decodedResponse))
 require.NoError(t, genclient.ValidateSendResponse(decodedResponse))
 require.Equal(t, source, genclient.NewSendResult(decodedResponse))
}

func TestRequiredNamedEntriesRejectNil(t *testing.T) {
 for _, choice := range []bool{false, true} {
  request := new(genpb.SendRequest)
  response := new(genpb.SendResponse)
  if choice {
   request.StrictChoices = []*genpb.ChoiceEntry{nil}
   response.StrictChoices = []*genpb.ChoiceEntry{nil}
  } else {
   request.Strict = []*genpb.Entry{nil}
   response.Strict = []*genpb.Entry{nil}
  }
  require.Error(t, genserver.ValidateSendRequest(request))
  require.Error(t, genclient.ValidateSendResponse(response))
 }
}

// TestNamedEntriesOnTheWire verifies protobuf turns a nil repeated message into
// a present empty message, whose required fields still need validation.
func TestNamedEntriesOnTheWire(t *testing.T) {
 for _, test := range []struct {
  name string
  required, choice, beforeValid, afterValid bool
 }{
  {name: "nullable object", beforeValid: true, afterValid: true},
  {name: "required object", required: true, afterValid: true},
  {name: "nullable union", choice: true, beforeValid: true},
  {name: "required union", required: true, choice: true},
 } {
  t.Run(test.name, func(t *testing.T) {
   before := new(genpb.SendRequest)
   switch {
   case test.required && test.choice:
    before.StrictChoices = []*genpb.ChoiceEntry{nil}
   case test.choice:
    before.LooseChoices = []*genpb.ChoiceEntry{nil}
   case test.required:
    before.Strict = []*genpb.Entry{nil, {Label: proto.String("next")}}
   default:
    before.Loose = []*genpb.Entry{nil, {Label: proto.String("next")}}
   }
   require.Equal(t, test.beforeValid, genserver.ValidateSendRequest(before) == nil)
   data, err := proto.Marshal(before)
   require.NoError(t, err)
   after := new(genpb.SendRequest)
   require.NoError(t, proto.Unmarshal(data, after))
   require.Equal(t, test.afterValid, genserver.ValidateSendRequest(after) == nil)
   response := new(genpb.SendResponse)
   require.NoError(t, proto.Unmarshal(data, response))
   require.Equal(t, test.afterValid, genclient.ValidateSendResponse(response) == nil)
   if test.afterValid {
    result := genserver.NewSendPayload(after)
    entries := result.Loose
    if test.required {
     entries = result.Strict
    }
    require.Len(t, entries, 2)
    require.NotNil(t, entries[0])
    require.Nil(t, entries[0].Label)
    require.Equal(t, "next", *entries[1].Label)
   }
  })
 }
 require.Error(t, protojson.Unmarshal([]byte("{\"loose\":[null]}"), new(genpb.SendRequest)))
}
`
