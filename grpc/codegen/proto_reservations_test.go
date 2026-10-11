// This file verifies that authored reservations reach generated request and
// response messages without requiring a custom protobuf message name.
package codegen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestProtoFilesPreserveMessageReservations(t *testing.T) {
	cases := []struct {
		name      string
		shape     string
		streaming bool
		custom    bool
	}{
		{name: "inline", shape: "inline"},
		{name: "named", shape: "named"},
		{name: "shared-named-type", shape: "shared"},
		{name: "result-type", shape: "result-type"},
		{name: "named-custom-name", shape: "custom-name"},
		{name: "inline-streaming", shape: "inline", streaming: true},
		{name: "named-streaming", shape: "named", streaming: true},
		{name: "result-type-streaming", shape: "result-type", streaming: true},
		{name: "inline-custom-message", shape: "inline", custom: true},
		{name: "named-custom-message", shape: "named", custom: true},
		{name: "named-custom-name-and-message", shape: "custom-name", custom: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := RunGRPCDSL(t, func() {
				messageReservationsDSL(c.shape, c.streaming, c.custom)
			})
			var messages []*expr.AttributeExpr
			for _, endpoint := range root.API.GRPC.Service("records").GRPCEndpoints {
				messages = append(messages, endpoint.Request, endpoint.Response.Message)
				if c.streaming {
					messages = append(messages, endpoint.StreamingRequest)
				}
			}
			numbers := []string{"5", "3"}
			names := []string{"retired_label", "retired_value"}
			if c.custom {
				numbers = append(numbers, "13")
				names = append(names, "retired_transport")
			}
			for _, message := range messages {
				require.ElementsMatch(t, numbers, message.Meta["rpc:reserved:number"])
				require.ElementsMatch(t, names, message.Meta["rpc:reserved:name"])
			}

			files := protoFiles(CreateGRPCServices(root))
			require.Len(t, files, 1)
			schema := sectionCode(t, files[0].SectionTemplates[1:]...)
			reservedNumbers := "reserved 3, 5;"
			reservedNames := `reserved "retired_label", "retired_value";`
			if c.custom {
				reservedNumbers = "reserved 3, 5, 13;"
				reservedNames = `reserved "retired_label", "retired_transport", "retired_value";`
			}
			require.Equal(t, len(messages), strings.Count(schema, reservedNumbers))
			require.Equal(t, len(messages), strings.Count(schema, reservedNames))
			require.Equal(t, len(messages), strings.Count(schema, "optional string current_value = 7;"))
			require.Equal(t, len(messages), strings.Count(schema, "optional string current_label = 11;"))
			if c.shape == "custom-name" {
				require.Contains(t, schema, "message InputRecord {")
				require.Contains(t, schema, "message OutputRecord {")
			}
			require.NoError(t, protoc(defaultProtocCmd, codegen.CreateTempFile(t, schema)))
		})
	}
}

func TestGRPCMessageReservationsStayIndependent(t *testing.T) {
	cases := []struct {
		name      string
		streaming bool
		response  bool
		custom    bool
	}{
		{name: "streaming", streaming: true},
		{name: "request-custom-name"},
		{name: "response-custom-name", response: true},
		{name: "request-explicit-message", custom: true},
		{name: "response-explicit-message", response: true, custom: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := RunGRPCDSL(t, func() {
				dsl.API("reservation-isolation", func() {
				})
				record := dsl.Type("Record", func() {
					if !c.streaming {
						dsl.Meta("struct:name:proto", "SharedRecord")
					}
					// Separate declarations leave room for append to reuse the slice,
					// reproducing metadata shared by two uses of the named type.
					dsl.Meta("rpc:reserved:number", "3")
					dsl.Meta("rpc:reserved:number", "5")
					dsl.Meta("rpc:reserved:number", "17")
					dsl.Meta("rpc:reserved:name", "retired_id")
					dsl.Meta("rpc:reserved:name", "retired_label")
					dsl.Meta("rpc:reserved:name", "retired_value")
					dsl.Field(7, "current_value", dsl.String)
				})
				dsl.Service("records", func() {
					for _, method := range []struct {
						name   string
						number string
					}{
						{name: "first", number: "19"},
						{name: "second", number: "23"},
					} {
						dsl.Method(method.name, func() {
							meta := func() {
								dsl.Meta("rpc:reserved:number", method.number)
								dsl.Meta("rpc:reserved:name", "retired_"+method.name)
							}
							switch {
							case c.streaming:
								dsl.StreamingPayload(record, meta)
							case c.response:
								dsl.Result(record, meta)
							default:
								dsl.Payload(record, meta)
							}
							dsl.GRPC(func() {
								if c.custom {
									message := func() {
										dsl.Meta("rpc:reserved:number", "29")
										dsl.Meta("rpc:reserved:name", "retired_transport")
										dsl.Meta("transport:note", "preserved")
										dsl.Attribute("current_value")
									}
									if c.response {
										dsl.Response(dsl.CodeOK, func() {
											dsl.Message(message)
										})
									} else {
										dsl.Message(message)
									}
								}
							})
						})
					}
				})
			})
			baseNumbers := []string{"3", "5", "17"}
			baseNames := []string{"retired_id", "retired_label", "retired_value"}
			source := root.UserType("Record").Attribute().Meta
			require.Equal(t, baseNumbers, source["rpc:reserved:number"])
			require.Equal(t, baseNames, source["rpc:reserved:name"])
			for index, endpoint := range root.API.GRPC.Service("records").GRPCEndpoints {
				message := endpoint.Request
				if c.streaming {
					message = endpoint.StreamingRequest
				} else if c.response {
					message = endpoint.Response.Message
				}
				numbers := append([]string{"3", "5", "17"}, []string{"19", "23"}[index])
				names := append([]string{"retired_id", "retired_label", "retired_value"}, "retired_"+endpoint.Name())
				if c.custom {
					numbers = append(numbers, "29")
					names = append(names, "retired_transport")
					require.Equal(t, []string{"preserved"}, message.Meta["transport:note"])
				}
				require.ElementsMatch(t, numbers, message.Meta["rpc:reserved:number"])
				require.ElementsMatch(t, names, message.Meta["rpc:reserved:name"])
				if !c.streaming {
					require.Equal(t, []string{"SharedRecord"}, message.Meta["struct:name:proto"])
				}
			}
		})
	}
}

// messageReservationsDSL defines removed and retained fields on each authored
// message so generation can prove that reservations never renumber live fields.
func messageReservationsDSL(shape string, streaming, custom bool) {
	dsl.API("message-reservations", func() {
	})
	fields := func() {
		dsl.Meta("rpc:reserved:number", "5", "3")
		dsl.Meta("rpc:reserved:name", "retired_label", "retired_value")
		dsl.Field(7, "current_value", dsl.String)
		dsl.Field(11, "current_label", dsl.String)
	}
	var input, output expr.DataType
	if shape != "inline" {
		input = dsl.Type("Input", func() {
			fields()
			if shape == "custom-name" {
				dsl.Meta("struct:name:proto", "InputRecord")
			}
		})
		if shape == "result-type" {
			output = dsl.ResultType("application/vnd.output", func() {
				dsl.Attributes(fields)
			})
		} else {
			output = dsl.Type("Output", func() {
				fields()
				if shape == "custom-name" {
					dsl.Meta("struct:name:proto", "OutputRecord")
				}
			})
		}
	}
	dsl.Service("records", func() {
		methods := []string{"exchange"}
		if shape == "shared" {
			methods = append(methods, "repeat")
		}
		for _, method := range methods {
			dsl.Method(method, func() {
				if shape == "inline" {
					dsl.Payload(fields)
					if streaming {
						dsl.StreamingPayload(fields)
						dsl.StreamingResult(fields)
					} else {
						dsl.Result(fields)
					}
				} else {
					dsl.Payload(input)
					if streaming {
						dsl.StreamingPayload(input)
						dsl.StreamingResult(output)
					} else {
						dsl.Result(output)
					}
				}
				dsl.GRPC(func() {
					if custom {
						message := func() {
							dsl.Meta("rpc:reserved:number", "13", "3")
							dsl.Meta("rpc:reserved:name", "retired_transport", "retired_value")
							dsl.Attribute("current_value")
						}
						dsl.Message(message)
						dsl.Response(dsl.CodeOK, func() {
							dsl.Message(message)
						})
					}
				})
			})
		}
	})
}
