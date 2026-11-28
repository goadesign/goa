// These tests compile and run generated gRPC conversions for a Goa type mapped
// to a protobuf message that Goa does not generate.
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

// TestExternalMessageFieldsKeepImplicitPresence checks that a type mapped to
// google.protobuf.Timestamp converts its plain int64 and int32 fields and
// round-trips zero values through protobuf bytes. The HTTP case checks that the
// same type keeps compiling in HTTP bodies, which still use pointers.
func TestExternalMessageFieldsKeepImplicitPresence(t *testing.T) {
	cases := []struct {
		Name   string
		Fields func()
		HTTP   bool
	}{
		{
			Name: "required",
			Fields: func() {
				d.Required("seconds", "nanos")
			},
		},
		{
			Name: "required with HTTP",
			Fields: func() {
				d.Required("seconds", "nanos")
			},
			HTTP: true,
		},
		{
			Name: "default",
			Fields: func() {
				d.Attribute("seconds", func() {
					d.Default(0)
				})
				d.Attribute("nanos", func() {
					d.Default(0)
				})
			},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := expr.RunDSL(t, externalMessageDSL(c.Fields, c.HTTP))
			generation, servicePlans := grpcServicePlans(t, []*expr.RootExpr{root})
			grpcPlans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlans[0]})
			require.NoError(t, err)
			var httpPlans []*httpcodegen.Plan
			if c.HTTP {
				httpPlans, err = httpcodegen.NewPlans(generation, httpcodegen.PlanInput{Root: root, Service: servicePlans[0]})
				require.NoError(t, err)
			}
			require.NoError(t, generation.Freeze())
			require.NoError(t, servicePlans[0].Link())
			require.NoError(t, grpcPlans[0].Link())
			for _, plan := range httpPlans {
				require.NoError(t, plan.Link())
			}
			files, err := service.Files(servicePlans...)
			require.NoError(t, err)
			files = append(files, grpcPlans[0].ServerFiles()...)
			files = append(files, grpcPlans[0].ClientFiles()...)
			files = append(files, grpcPlans[0].ServerTypeFiles()...)
			files = append(files, grpcPlans[0].ClientTypeFiles()...)
			files = append(files, grpcPlans[0].ProtoFiles()...)
			for _, plan := range httpPlans {
				files = append(files, plan.ServerTypeFiles()...)
				files = append(files, plan.ServerFiles()...)
				files = append(files, plan.ClientTypeFiles()...)
				files = append(files, plan.ClientFiles()...)
				files = append(files, plan.PathFiles()...)
			}
			directory := t.TempDir()
			writeProtobufDescriptorModule(t, directory)
			for _, file := range files {
				_, err := file.Render(directory)
				require.NoError(t, err)
			}
			source, err := format.Source([]byte(externalMessageRuntimeTest))
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(filepath.Join(directory, "clock_test.go"), source, 0o600))
			compileProtobufDescriptorModule(t, directory)
		})
	}
}

// externalMessageDSL maps a Goa timestamp type to google.protobuf.Timestamp in
// a request and a response. fields adds the rules for seconds and nanos, and
// withHTTP also exposes the method over HTTP.
func externalMessageDSL(fields func(), withHTTP bool) func() {
	return func() {
		timestamp := d.Type("Timestamp", func() {
			d.Field(1, "seconds", d.Int64, func() {
				d.Meta("struct:field:proto", "int64")
			})
			d.Field(2, "nanos", d.Int32, func() {
				d.Meta("struct:field:proto", "int32")
			})
			fields()
		})
		tick := func() {
			d.Field(1, "at", timestamp, func() {
				d.Meta("struct:field:proto", "google.protobuf.Timestamp", "google/protobuf/timestamp.proto",
					"Timestamp", "google.golang.org/protobuf/types/known/timestamppb")
			})
			d.Required("at")
		}
		d.Service("Clock", func() {
			d.Method("Tick", func() {
				d.Payload(tick)
				d.Result(tick)
				d.GRPC(func() {})
				if withHTTP {
					d.HTTP(func() {
						d.POST("/tick")
					})
				}
			})
		})
	}
}

const externalMessageRuntimeTest = `package clock_test

import (
 "testing"

 "github.com/stretchr/testify/require"
 genclock "generated.local/gen/clock"
 gengrpcclient "generated.local/gen/grpc/clock/client"
 gengrpcserver "generated.local/gen/grpc/clock/server"
 genpb "generated.local/gen/grpc/clock/pb"
 "google.golang.org/protobuf/proto"
 "google.golang.org/protobuf/types/known/timestamppb"
)

func TestTimestampRoundTrip(t *testing.T) {
 for _, at := range []*timestamppb.Timestamp{{Seconds: 1732480000, Nanos: 123}, {Seconds: 1732480000}, {}} {
  wire, err := proto.Marshal(&genpb.TickRequest{At: at})
  require.NoError(t, err)
  request := new(genpb.TickRequest)
  require.NoError(t, proto.Unmarshal(wire, request))
  require.NoError(t, gengrpcserver.ValidateTickRequest(request))
  payload := gengrpcserver.NewTickPayload(request)
  require.Equal(t, at.Seconds, payload.At.Seconds)
  require.Equal(t, at.Nanos, payload.At.Nanos)

  message := gengrpcclient.NewProtoTickRequest(payload)
  require.Equal(t, at.Seconds, message.At.Seconds)
  require.Equal(t, at.Nanos, message.At.Nanos)

  response := gengrpcserver.NewProtoTickResponse(&genclock.TickResult{At: payload.At})
  wire, err = proto.Marshal(response)
  require.NoError(t, err)
  decoded := new(genpb.TickResponse)
  require.NoError(t, proto.Unmarshal(wire, decoded))
  require.NoError(t, gengrpcclient.ValidateTickResponse(decoded))
  result := gengrpcclient.NewTickResult(decoded)
  require.Equal(t, at.Seconds, result.At.Seconds)
  require.Equal(t, at.Nanos, result.At.Nanos)
 }
}
`
