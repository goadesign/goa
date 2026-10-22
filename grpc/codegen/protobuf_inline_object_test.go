// This file verifies that inline objects become complete protobuf messages before
// names are fixed, including when the same named type is used by several methods.
package codegen

import (
	"cmp"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

const inlineObjectRuntimeTest = `package server

import (
 "testing"
 genpb "generated.local/gen/grpc/records/pb"
)

func TestInlineRequirednessAndRoundTrip(t *testing.T) {
 if NewBasePayload(&genpb.BaseRequest{}).Session != nil {
  t.Error("base must preserve an omitted optional session")
 }
 if ValidateReadRequest(&genpb.ReadRequest{}) == nil {
  t.Error("derived request must reject missing required objects")
 }
 invalid := &genpb.ReadRequest{Session: &genpb.Session{}, Detail: &genpb.Detail{}}
 if ValidateReadRequest(invalid) == nil {
  t.Error("detail must reject an omitted required value")
 }
 value := "example"
 request := &genpb.ReadRequest{Session: &genpb.Session{}, Detail: &genpb.Detail{Value: &value}}
 if err := ValidateReadRequest(request); err != nil {
  t.Fatal(err)
 }
 result := NewProtoReadResponse(NewReadPayload(request))
 if result.Session == nil || result.Session.Reference != nil {
  t.Error("present session must preserve its omitted optional reference")
 }
 if result.Detail == nil || result.Detail.Value == nil || *result.Detail.Value != value {
  t.Error("nested value did not survive the conversion")
 }
}
`

const inlineCollectionRuntimeTest = `package server

import (
 "testing"
 genrecords "generated.local/gen/records"
 genpb "generated.local/gen/grpc/records/pb"
)

func TestNullableInlineCollections(t *testing.T) {
 t.Run("encode array", func(t *testing.T) {
  source := &genrecords.Envelope{Rows: []*struct{ Value *string }{nil}}
  result := NewProtoReadResponse(source)
  if len(result.Rows) != 1 || result.Rows[0] != nil { t.Error("nil array element was not preserved") }
 })
 t.Run("encode map", func(t *testing.T) {
  source := &genrecords.Envelope{Labels: map[string]*struct{ Value *string }{"empty": nil}}
  result := NewProtoReadResponse(source)
  value, exists := result.Labels["empty"]
  if !exists || value != nil { t.Error("nil map value was not preserved") }
 })
 t.Run("decode array", func(t *testing.T) {
  source := &genpb.ReadRequest{}
  source.Rows = append(source.Rows, nil)
  result := NewReadPayload(source)
  if len(result.Rows) != 1 || result.Rows[0] != nil { t.Error("nil array element was not preserved") }
 })
 t.Run("decode map", func(t *testing.T) {
  empty := NewProtoReadResponse(&genrecords.Envelope{Labels: map[string]*struct{ Value *string }{}})
  source := &genpb.ReadRequest{Labels: empty.Labels}
  source.Labels["empty"] = nil
  result := NewReadPayload(source)
  value, exists := result.Labels["empty"]
  if !exists || value != nil { t.Error("nil map value was not preserved") }
 })
}

func TestRequiredInlineArrayRejectsNil(t *testing.T) {
 value := "valid"
 source := &genpb.ReadRequest{Detail: &genpb.Detail{Value: &value}}
 source.RequiredRows = append(source.RequiredRows, nil)
 if ValidateReadRequest(source) == nil { t.Error("required array element must reject nil") }
}
`

func TestInlineObjectProtobufGeneration(t *testing.T) {
	var compiled []protobufCompileCase
	for _, name := range []string{"inline", "extended", "collections"} {
		extended := name == "extended"
		var previous []byte
		var previousMessages []byte
		for run := 0; run < 2; run++ {
			root := RunGRPCDSL(t, func() {
				base := dsl.Type("BaseAaa", func() {
					dsl.Field(1, "session", func() {
						dsl.Field(1, "reference", dsl.String)
					})
				})
				other := dsl.Type("Mmm", func() {
					dsl.Field(1, "session", func() {
						dsl.Field(1, "reference", dsl.String)
					})
				})
				node := dsl.Type("Node", func() {
					dsl.Field(1, "next", "Node")
					dsl.Field(2, "value", dsl.String)
				})
				messageName := "Envelope"
				if extended {
					messageName = "Zzz"
				}
				message := dsl.Type(messageName, func() {
					if extended {
						dsl.Extend(base)
						dsl.Required("session")
						dsl.Field(10, "other", other)
					}
					dsl.Field(2, "detail", func() {
						dsl.Meta("struct:name:proto", "Detail")
						dsl.Field(1, "value", dsl.String)
						dsl.Required("value")
					})
					dsl.Required("detail")
					if name == "collections" {
						dsl.Field(3, "rows", dsl.ArrayOf(&expr.Object{}, func() {
							dsl.Field(1, "value", dsl.String)
						}))
						dsl.Field(4, "labels", dsl.MapOf(dsl.String, &expr.Object{}, func() {
							dsl.Elem(func() {
								dsl.Field(1, "value", dsl.String)
							})
						}))
						dsl.Field(5, "matrix", dsl.ArrayOf(dsl.ArrayOf(&expr.Object{}, func() {
							dsl.Field(1, "value", dsl.String)
						})))
						dsl.Field(11, "required_rows", dsl.ArrayOfRequired(&expr.Object{}, func() {
							dsl.Field(1, "value", dsl.String)
						}))
						dsl.Field(6, "first", node)
						dsl.Field(7, "second", node)
						dsl.OneOf("choice", func() {
							dsl.Field(8, "text", dsl.String)
							dsl.Field(9, "object", func() {
								dsl.Field(1, "value", dsl.String)
							})
						})
					}
				})
				dsl.Service("records", func() {
					read := func() {
						dsl.Method("read", func() {
							dsl.Payload(message)
							dsl.Result(message)
							dsl.GRPC(func() {})
						})
					}
					baseMethod := func() {
						if extended {
							dsl.Method("base", func() {
								dsl.Payload(base)
								dsl.Result(base)
								dsl.GRPC(func() {})
							})
						}
					}
					if run == 0 {
						read()
						baseMethod()
					} else {
						baseMethod()
						read()
					}
				})
			})
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
			var source []byte
			var sessionMessage string
			for _, file := range files {
				path, err := file.Render(directory)
				require.NoError(t, err)
				if filepath.Ext(path) == ".proto" {
					source, err = os.ReadFile(path)
					require.NoError(t, err)
					descriptor := describeGeneratedProto(t, path)
					for _, message := range descriptor.GetMessageType() {
						if message.GetName() == "ReadRequest" && extended {
							qualified := fieldNamed(t, message, "session").GetTypeName()
							sessionMessage = qualified[strings.LastIndex(qualified, ".")+1:]
						}
					}
					descriptor.Service = nil
					slices.SortFunc(descriptor.MessageType, func(left, right *descriptorpb.DescriptorProto) int {
						return cmp.Compare(left.GetName(), right.GetName())
					})
					encoded, err := proto.MarshalOptions{Deterministic: true}.Marshal(descriptor)
					require.NoError(t, err)
					if run > 0 {
						require.Equal(t, previousMessages, encoded)
					}
					previousMessages = encoded
				}
			}
			require.NotEmpty(t, source)
			require.Equal(t, 1, strings.Count(string(source), "message Detail {"))
			if extended {
				require.Equal(t, 1, strings.Count(string(source), "message Session {"))
			}
			if run > 0 && !extended {
				require.Equal(t, previous, source)
			}
			previous = source
			if run == 1 {
				if extended {
					runtimeTest := filepath.Join(directory, "gen", "grpc", "records", "server", "inline_object_test.go")
					require.NoError(t, os.WriteFile(runtimeTest, []byte(strings.ReplaceAll(inlineObjectRuntimeTest, "genpb.Session", "genpb."+sessionMessage)), 0o600))
				}
				if name == "collections" {
					runtimeTest := filepath.Join(directory, "gen", "grpc", "records", "server", "inline_collection_test.go")
					require.NoError(t, os.WriteFile(runtimeTest, []byte(inlineCollectionRuntimeTest), 0o600))
				}
				compiled = append(compiled, protobufCompileCase{name: name, dir: directory})
			}
		}
	}
	compileProtobufDescriptorCases(t, compiled)
}

// TestInlineObjectMessagePreparation checks copied graph identity and authored
// field rules independently from client and server conversion rendering.
func TestInlineObjectMessagePreparation(t *testing.T) {
	for _, shape := range []string{"field", "array", "map", "union", "nested-array"} {
		t.Run(shape, func(t *testing.T) {
			inline := &expr.AttributeExpr{Type: &expr.Object{
				{Name: "value", Attribute: &expr.AttributeExpr{Type: expr.String, Meta: expr.MetaExpr{"rpc:tag": {"1"}}}},
			}, Meta: expr.MetaExpr{"rpc:tag": {"2"}, "struct:name:proto": {"Detail"}}, Validation: &expr.ValidationExpr{Required: []string{"value"}}}
			child := inline
			switch shape {
			case "array":
				child = &expr.AttributeExpr{Type: &expr.Array{ElemType: inline}}
			case "map":
				child = &expr.AttributeExpr{Type: &expr.Map{KeyType: &expr.AttributeExpr{Type: expr.String}, ElemType: inline}}
			case "union":
				child = &expr.AttributeExpr{Type: &expr.Union{TypeName: "Choice", Values: []*expr.NamedAttributeExpr{{Name: "detail", Attribute: inline}}}}
			case "nested-array":
				child = &expr.AttributeExpr{Type: &expr.Array{ElemType: &expr.AttributeExpr{Type: &expr.Array{ElemType: inline}}}}
			}
			named := &expr.UserTypeExpr{TypeName: "Node", AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{}}}
			*expr.AsObject(named) = expr.Object{
				{Name: "child", Attribute: child},
				{Name: "next", Attribute: &expr.AttributeExpr{Type: named}},
			}
			message := makeProtoBufMessage(&expr.AttributeExpr{Type: &expr.Object{
				{Name: "first", Attribute: &expr.AttributeExpr{Type: named}},
				{Name: "second", Attribute: &expr.AttributeExpr{Type: named}},
			}}, "Request", testGRPCMessageExampleIdentity(shape))
			first := expr.AsObject(message.Type).Attribute("first").Type.(expr.UserType)
			second := expr.AsObject(message.Type).Attribute("second").Type.(expr.UserType)
			require.Same(t, first, second)
			require.Same(t, named, first.Origin())
			body, ok := first.Attribute().Type.(*expr.Object)
			require.True(t, ok, "named type retains its direct object body")
			require.Same(t, first, body.Attribute("next").Type)
			prepared := body.Attribute("child")
			switch shape {
			case "array":
				prepared = expr.AsArray(prepared.Type).ElemType
			case "map":
				prepared = expr.AsMap(prepared.Type).ElemType
			case "union":
				prepared = expr.AsUnion(prepared.Type).Values[0].Attribute
			case "nested-array":
				prepared = expr.AsArray(unwrapAttr(expr.AsArray(prepared.Type).ElemType).Type).ElemType
			}
			generated, ok := prepared.Type.(expr.UserType)
			require.True(t, ok)
			require.Equal(t, inline.Meta, prepared.Meta)
			require.Equal(t, inline.Validation, generated.Attribute().Validation)
			require.Equal(t, []string{"1"}, expr.AsObject(generated).Attribute("value").Meta["rpc:tag"])
			_, anonymous := inline.Type.(*expr.Object)
			require.True(t, anonymous, "the input design remains unchanged")
		})
	}
}

// TestInlineObjectNamesIgnoreInheritedDiscoveryOrder reaches the same inline
// field through two named types, with a third type between their sorting names.
func TestInlineObjectNamesIgnoreInheritedDiscoveryOrder(t *testing.T) {
	baseID := expr.UserTypeExampleIdentity(&expr.UserTypeExpr{TypeName: "Aaa"}).Member("session").Seed()
	derivedID := expr.UserTypeExampleIdentity(&expr.UserTypeExpr{TypeName: "Zzz"}).Member("session").Seed()
	otherID := expr.UserTypeExampleIdentity(&expr.UserTypeExpr{TypeName: "Mmm"}).Member("session").Seed()
	require.NotEqual(t, cmp.Compare(baseID, otherID) < 0, cmp.Compare(derivedID, otherID) < 0,
		"unrelated source must sort between the two inherited copies")
	var previous map[string]string
	for _, reverse := range []bool{false, true} {
		root := RunGRPCDSL(t, func() {
			base := dsl.Type("Aaa", func() {
				dsl.Field(1, "session", func() {
					dsl.Field(1, "reference", dsl.String)
				})
			})
			derived := dsl.Type("Zzz", func() {
				dsl.Extend(base)
			})
			other := dsl.Type("Mmm", func() {
				dsl.Field(1, "session", func() {
					dsl.Field(1, "reference", dsl.String)
				})
			})
			dsl.Service("records", func() {
				methods := []string{"derived", "base"}
				if reverse {
					slices.Reverse(methods)
				}
				for _, method := range methods {
					dsl.Method(method, func() {
						dsl.Payload(func() {
							typ := derived
							if method == "base" {
								typ = base
							}
							dsl.Field(1, "value", typ)
							dsl.Field(2, "other", other)
						})
						dsl.GRPC(func() {})
					})
				}
			})
		})
		generation, services := grpcServicePlans(t, []*expr.RootExpr{root})
		plans, err := NewPlans(generation, PlanInput{Root: root, Service: services[0]})
		require.NoError(t, err)
		require.NoError(t, generation.Freeze())
		require.NoError(t, services[0].Link())
		require.NoError(t, plans[0].Link())
		references := make(map[string]string)
		for _, file := range plans[0].ProtoFiles() {
			path, err := file.Render(t.TempDir())
			require.NoError(t, err)
			if filepath.Ext(path) != ".proto" {
				continue
			}
			descriptor := describeGeneratedProto(t, path)
			for _, message := range descriptor.GetMessageType() {
				for _, field := range message.GetField() {
					references[message.GetName()+"."+field.GetName()] = field.GetTypeName()
				}
			}
		}
		require.NotEmpty(t, references)
		if reverse {
			require.Equal(t, previous, references)
		}
		previous = references
	}
}
