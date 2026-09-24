// This file checks custom defaults across generated service, HTTP, gRPC and
// CLI code. Generated tests exchange values using the actual transport codecs.
package generator

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	d "goa.design/goa/v3/dsl"
)

// exchangeDuration has a different Go name from time.Duration but the same
// numeric JSON representation. The generated exchange must preserve its value.
type exchangeDuration int64

func TestGenerateCustomDefaultsExchange(t *testing.T) {
	registry := testRegistry(
		"gen",
		testGenerator(planServiceData, testServiceFiles),
		testGenerator(planTransportData, testTransportFiles),
	)
	codegen.RunDSL(t, func() {
		value := d.Type("CustomDefaults", func() {
			for index, field := range []struct {
				name      string
				primitive any
				value     any
			}{
				{"raw_string", d.String, json.RawMessage("true")},
				{"raw_bytes", d.Bytes, json.RawMessage("true")},
				{"plain_bytes", d.String, []byte("true")},
			} {
				d.Field(index+1, field.name, field.primitive, func() {
					d.Meta("struct:field:type", "json.RawMessage", "encoding/json")
					d.Default(field.value)
				})
			}
			d.Field(4, "duration", d.Int64, func() {
				d.Meta("struct:field:type", "time.Duration", "time")
				d.Default(exchangeDuration(7))
			})
		})
		d.Service("defaults", func() {
			d.Method("Exchange", func() {
				d.Payload(value)
				d.Result(value)
				d.HTTP(func() { d.POST("/defaults") })
				d.GRPC(func() {})
			})
		})
	})
	directory := filepath.Join(t.TempDir(), codegen.Gendir)
	writeGeneratedModule(t, directory, "generated.local/gen")
	_, err := generate(filepath.Dir(directory), "gen", false, registry)
	require.NoError(t, err)
	writeGeneratedContractTest(t, directory, filepath.Join("http", "defaults", "server"), customDefaultsHTTPTest)
	writeGeneratedContractTest(t, directory, filepath.Join("grpc", "defaults", "server"), customDefaultsGRPCTest)
	runGeneratedTests(t, directory)
}

const customDefaultsHTTPTest = `package server
import (
	"bytes"
	"encoding/json"
	"testing"
	genclient "generated.local/gen/http/defaults/client"
)
func TestCustomDefaultsHTTPExchange(t *testing.T) {
	var body ExchangeRequestBody
	if err := json.Unmarshal([]byte("{}"), &body); err != nil { t.Fatal(err) }
	payload := NewExchangeCustomDefaults(&body)
	if !bytes.Equal(payload.RawString, []byte("true")) ||
		!bytes.Equal(payload.RawBytes, []byte("true")) ||
		!bytes.Equal(payload.PlainBytes, []byte("true")) || payload.Duration != 7 {
		t.Fatalf("defaults changed: %#v", payload)
	}
	payload.RawString = json.RawMessage("false")
	payload.Duration = 9
	request := genclient.NewExchangeRequestBody(payload)
	encoded, err := json.Marshal(request)
	if err != nil { t.Fatal(err) }
	if err := json.Unmarshal(encoded, &body); err != nil { t.Fatal(err) }
	incoming := NewExchangeCustomDefaults(&body)
	if !bytes.Equal(incoming.RawString, payload.RawString) || incoming.Duration != payload.Duration {
		t.Fatalf("HTTP request changed: %#v", incoming)
	}
	response := NewExchangeResponseBody(incoming)
	encoded, err = json.Marshal(response)
	if err != nil { t.Fatal(err) }
	var received genclient.ExchangeResponseBody
	if err := json.Unmarshal(encoded, &received); err != nil { t.Fatal(err) }
	result := genclient.NewExchangeCustomDefaultsOK(&received)
	if !bytes.Equal(result.RawString, payload.RawString) ||
		!bytes.Equal(result.RawBytes, payload.RawBytes) ||
		!bytes.Equal(result.PlainBytes, payload.PlainBytes) || result.Duration != payload.Duration {
		t.Fatalf("HTTP exchange changed: %#v", result)
	}
}
`

const customDefaultsGRPCTest = `package server
import (
	"bytes"
	"testing"
	"google.golang.org/protobuf/proto"
	genclient "generated.local/gen/grpc/defaults/client"
	genpb "generated.local/gen/grpc/defaults/pb"
)
func TestCustomDefaultsGRPCExchange(t *testing.T) {
	payload := NewExchangePayload(&genpb.ExchangeRequest{})
	if !bytes.Equal(payload.RawString, []byte("true")) ||
		!bytes.Equal(payload.RawBytes, []byte("true")) ||
		!bytes.Equal(payload.PlainBytes, []byte("true")) || payload.Duration != 7 {
		t.Fatalf("defaults changed: %#v", payload)
	}
	payload.RawString = []byte("false")
	payload.Duration = 9
	request := genclient.NewProtoExchangeRequest(payload)
	encoded, err := proto.Marshal(request)
	if err != nil { t.Fatal(err) }
	var incoming genpb.ExchangeRequest
	if err := proto.Unmarshal(encoded, &incoming); err != nil { t.Fatal(err) }
	accepted := NewExchangePayload(&incoming)
	if !bytes.Equal(accepted.RawString, payload.RawString) || accepted.Duration != payload.Duration {
		t.Fatalf("protobuf request changed: %#v", accepted)
	}
	response := NewProtoExchangeResponse(accepted)
	encoded, err = proto.Marshal(response)
	if err != nil { t.Fatal(err) }
	var received genpb.ExchangeResponse
	if err := proto.Unmarshal(encoded, &received); err != nil { t.Fatal(err) }
	result := genclient.NewExchangeResult(&received)
	if !bytes.Equal(result.RawString, payload.RawString) ||
		!bytes.Equal(result.RawBytes, payload.RawBytes) ||
		!bytes.Equal(result.PlainBytes, payload.PlainBytes) || result.Duration != payload.Duration {
		t.Fatalf("protobuf exchange changed: %#v", result)
	}
}
`
