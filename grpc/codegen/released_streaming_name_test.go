// This file checks that one gRPC response conversion keeps its released public
// name when both the response encoder and a stream send method use it.
package codegen

import (
	"testing"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/grpc/codegen/testdata"
)

// TestReleasedStreamingResponseConstructorNames catches replacing an honest
// method response name with an internal type-based name when conversions merge.
func TestReleasedStreamingResponseConstructorNames(t *testing.T) {
	t.Run("caller selected view", func(t *testing.T) {
		root := RunGRPCDSL(t, testdata.ServerStreamingResultWithViewsDSL)
		services := CreateGRPCServices(root)
		types := serverTypeFiles(services)
		servers := serverFiles(services)

		sections := append(types[0].Section("server-type-init"), servers[1].Section("response-encoder")...)
		sections = append(sections, servers[0].Section("server-stream-send")...)
		code := codegen.SectionsCode(t, sections)
		testutil.AssertGo(t, "testdata/golden/released_streaming_response_constructors.go.golden", code)
	})

	t.Run("fixed collection view", func(t *testing.T) {
		root := RunGRPCDSL(t, testdata.ClientStreamingResultCollectionWithExplicitViewDSL)
		services := CreateGRPCServices(root)
		types := serverTypeFiles(services)
		servers := serverFiles(services)

		sections := append(types[0].Section("server-type-init"), servers[1].Section("response-encoder")...)
		sections = append(sections, servers[0].Section("server-stream-send")...)
		code := codegen.SectionsCode(t, sections)
		testutil.AssertGo(t, "testdata/golden/released_fixed_view_collection_constructor.go.golden", code)
	})
}
