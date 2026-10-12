// This file compiles generated protobuf services whose method names match
// request or response messages, so name lookup cannot select a method instead.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

func TestProtoRPCReferencesResolveMessages(t *testing.T) {
	cases := []struct {
		name          string
		kind          expr.StreamKind
		pkg           string
		externalField bool
		requestName   string
	}{
		{name: "unary", kind: expr.NoStreamKind, requestName: "GetRunRequest"},
		{name: "server-stream", kind: expr.ServerStreamKind, requestName: "GetRunRequest"},
		{name: "client-stream", kind: expr.ClientStreamKind, requestName: "GetRunStreamingRequest"},
		{name: "bidirectional-stream", kind: expr.BidirectionalStreamKind, requestName: "GetRunStreamingRequest"},
		{name: "custom-package", kind: expr.NoStreamKind, pkg: "example.sample", requestName: "GetRunRequest"},
		{name: "external-field", kind: expr.NoStreamKind, externalField: true, requestName: "GetRunRequest"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := RunGRPCDSL(t, func() {
				request := dsl.Type("GetRunRequest", func() {
					dsl.Field(1, "value", dsl.String)
					if c.externalField {
						dsl.Field(2, "created_at", dsl.String, func() {
							dsl.Meta("struct:field:proto", "google.protobuf.Timestamp",
								"google/protobuf/timestamp.proto", "Timestamp",
								"google.golang.org/protobuf/types/known/timestamppb")
						})
					}
				})
				response := dsl.Type("GetRunResponse", func() {
					dsl.Field(1, "value", dsl.String)
				})
				dsl.Service("Sample", func() {
					dsl.GRPC(func() {
						if c.pkg != "" {
							dsl.Package(c.pkg)
						}
					})
					dsl.Method("get_run", func() {
						if c.kind == expr.ClientStreamKind || c.kind == expr.BidirectionalStreamKind {
							dsl.StreamingPayload(request)
						} else {
							dsl.Payload(request)
						}
						if c.kind == expr.ServerStreamKind || c.kind == expr.BidirectionalStreamKind {
							dsl.StreamingResult(response)
						} else {
							dsl.Result(response)
						}
						dsl.GRPC(func() {})
					})
					dsl.Method(c.requestName, func() {
						dsl.GRPC(func() {})
					})
					dsl.Method("get_run_response", func() {
						dsl.GRPC(func() {})
					})
				})
			})
			files := protoFiles(CreateGRPCServices(root))
			require.Len(t, files, 1)
			file, err := files[0].Render(t.TempDir())
			require.NoError(t, err)
			descriptor := describeGeneratedProto(t, file)
			require.Len(t, descriptor.GetService(), 1)
			methods := descriptor.GetService()[0].GetMethod()
			require.Len(t, methods, 3)
			method := methods[0]
			require.Equal(t, "GetRun", method.GetName())
			require.Equal(t, "."+descriptor.GetPackage()+"."+c.requestName, method.GetInputType())
			require.Equal(t, "."+descriptor.GetPackage()+".GetRunResponse", method.GetOutputType())
			require.Equal(t, c.kind == expr.ClientStreamKind || c.kind == expr.BidirectionalStreamKind, method.GetClientStreaming())
			require.Equal(t, c.kind == expr.ServerStreamKind || c.kind == expr.BidirectionalStreamKind, method.GetServerStreaming())
			require.Equal(t, c.requestName, methods[1].GetName())
			require.Equal(t, "GetRunResponse", methods[2].GetName())
		})
	}
}
