// These tests generate recursive unions through both transports, including
// declarations placed in a package shared by the service and its callers.
package generator

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	d "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
)

func TestGenerateRecursiveUnion(t *testing.T) {
	locations := []string{"", "types"}
	cases := make([]generatedTestDir, 0, len(locations))
	for _, location := range locations {
		name := location
		if name == "" {
			name = "local"
		}
		registry := testRegistry("gen",
			testGenerator(planServiceData, testServiceFiles),
			testGenerator(planTransportData, testTransportFiles),
			testGenerator(planOpenAPIData, testOpenAPIFiles))
		root := codegen.RunDSL(t, func() { recursiveUnionDSL(location) })
		first, err := sources("generated.local/gen", []eval.Root{root}, registry)
		require.NoError(t, err)
		root = codegen.RunDSL(t, func() { recursiveUnionDSL(location) })
		second, err := sources("generated.local/gen", []eval.Root{root}, registry)
		require.NoError(t, err)
		require.Equal(t, first, second)
		directory := t.TempDir()
		writeGeneratedModule(t, directory, "generated.local")
		codegen.RunDSL(t, func() { recursiveUnionDSL(location) })
		_, err = generate(directory, "gen", false, registry)
		require.NoError(t, err)
		writeGeneratedContractTest(t, directory, "gen/http/values/server", strings.ReplaceAll(recursiveUnionHTTPTest, "DOCUMENT", strconv.Quote(recursiveUnionDocument)))
		writeGeneratedContractTest(t, directory, "checks", strings.ReplaceAll(recursiveUnionGRPCTest, "DOCUMENT", strconv.Quote(recursiveUnionDocument)))
		cases = append(cases, generatedTestDir{name: name, dir: directory})
	}
	runGeneratedTestDirs(t, cases)
}

func recursiveUnionDSL(location string) {
	d.API("recursive-values", func() {})
	null := d.Type("Null", func() {})
	entry := d.Type("Entry", func() {
		d.Field(1, "value", "JSONValue")
		d.Required("value")
	})
	value := d.Type("JSONValue", func() {
		if location != "" {
			d.Meta("struct:pkg:path", location)
		}
		d.OneOf("value", func() {
			d.Field(1, "text", d.String)
			d.Field(2, "number", d.String)
			d.Field(3, "boolean", d.Boolean)
			d.Field(4, "array", d.ArrayOf("JSONValue"))
			d.Field(5, "object", d.MapOf(d.String, "JSONValue"))
			d.Field(6, "null", null)
			d.Field(7, "entry", entry)
			d.Field(8, "inline", func() {
				d.Field(1, "value", "JSONValue")
				d.Required("value")
			})
		})
		d.Required("value")
	})
	packet := d.Type("Packet", func() {
		d.Field(1, "outputs", d.MapOf(d.String, value))
		d.Required("outputs")
	})
	d.Service("values", func() {
		d.Method("exchange", func() {
			d.Payload(packet)
			d.Result(packet)
			d.HTTP(func() {
				d.POST("/values")
				d.Response(d.StatusOK)
			})
			d.GRPC(func() {})
		})
		d.Method("read", func() {
			d.Payload(d.String)
			d.Result(packet)
			d.HTTP(func() {
				d.POST("/read")
				d.Response(d.StatusOK)
			})
			d.GRPC(func() {})
		})
	})
}

const recursiveUnionDocument = `{
	"outputs": {
		"nested": {
			"value": {
				"type": "object",
				"value": {
					"unrecognized-key": {
						"value": {
							"type": "array",
							"value": [
								{
									"value": {
										"type": "null",
										"value": {}
									}
								},
								{
									"value": {
										"type": "array",
										"value": []
									}
								},
								{
									"value": {
										"type": "object",
										"value": {}
									}
								},
								{
									"value": {
										"type": "boolean",
										"value": false
									}
								},
								{
									"value": {
										"type": "text",
										"value": ""
									}
								},
								{
									"value": {
										"type": "number",
										"value": "9007199254740993123456789"
									}
								},
								{
									"value": {
										"type": "number",
										"value": "1.2300e+45"
									}
								},
								{
									"value": {
										"type": "entry",
										"value": {
											"value": {
												"value": {
													"type": "text",
													"value": "through mutual reference"
												}
											}
										}
									}
								},
								{
									"value": {
										"type": "inline",
										"value": {
											"value": {
												"value": {
													"type": "text",
													"value": "through inline reference"
												}
											}
										}
									}
								}
							]
						}
					}
				}
			}
		}
	}
}`

const recursiveUnionHTTPTest = `package server

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestRecursiveUnionHTTP(t *testing.T) {
	const document = DOCUMENT
	var body ExchangeRequestBody
	if err := json.Unmarshal([]byte(document), &body); err != nil {
		t.Fatal(err)
	}
	if err := ValidateExchangeRequestBody(&body); err != nil {
		t.Fatal(err)
	}
	value := NewExchangePacket(&body)
	for _, response := range []any{NewExchangeResponseBody(value), NewReadResponseBody(value)} {
		encoded, err := json.Marshal(response)
		if err != nil {
			t.Fatal(err)
		}
		var want, got any
		if err := json.Unmarshal([]byte(document), &want); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(encoded, &got); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(want, got) {
			t.Errorf("recursive response changed values: %s", encoded)
		}
	}
}
`

const recursiveUnionGRPCTest = `package checks

import (
	"encoding/json"
	"reflect"
	"testing"

	"google.golang.org/protobuf/proto"

	genclient "generated.local/gen/grpc/values/client"
	genpb "generated.local/gen/grpc/values/pb"
	genserver "generated.local/gen/grpc/values/server"
	genvalues "generated.local/gen/values"
)

func TestRecursiveUnionGRPC(t *testing.T) {
	const document = DOCUMENT
	var input genvalues.Packet
	if err := json.Unmarshal([]byte(document), &input); err != nil {
		t.Fatal(err)
	}
	request := genclient.NewProtoExchangeRequest(&input)
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
	if !reflect.DeepEqual(&input, output) {
		t.Errorf("protobuf round trip changed recursive values: got %#v want %#v", output, &input)
	}
}
`
