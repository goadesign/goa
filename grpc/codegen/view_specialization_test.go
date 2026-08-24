// This file checks that generated gRPC code keeps design-selected views in
// source and reads transport metadata only when the caller selects the view.
package codegen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/grpc/codegen/testdata"
)

func TestUnaryViewedResultSpecialization(t *testing.T) {
	t.Run("missing or conflicting headers keep design selected view", func(t *testing.T) {
		root := RunGRPCDSL(t, testdata.MessageResultTypeWithExplicitViewDSL)
		services := CreateGRPCServices(root)
		clientFiles := clientFiles(services)
		serverFiles := serverFiles(services)
		require.Len(t, clientFiles, 2)
		require.Len(t, serverFiles, 2)

		decoder := codegen.SectionsCode(t, clientFiles[1].Section("response-decoder"))
		assert.NotContains(t, decoder, `hdr.Get("goa-view")`)
		assert.Contains(t, decoder, `View: "tiny"`)
		assert.NotContains(t, decoder, "switch")
		encoder := codegen.SectionsCode(t, serverFiles[1].Section("response-encoder"))
		assert.Contains(t, encoder, `Append("goa-view", "tiny")`)
		assert.NotContains(t, encoder, `Append("goa-view", vres.View)`)
		testutil.AssertGo(t, "testdata/golden/viewed_result_fixed_response_encoder.go.golden", encoder)
		response := services.Get("ServiceMessageResultTypeWithExplicitView").Endpoints[0].Response
		require.Len(t, response.ClientConverts, 1)
		require.Equal(t, "tiny", response.ClientConverts[0].View)
		require.Same(t, response.ClientConverts[0].Convert, response.ClientConvert)
	})

	t.Run("caller selected view", func(t *testing.T) {
		root := RunGRPCDSL(t, testdata.MessageResultTypeWithViewsDSL)
		services := CreateGRPCServices(root)
		clientFiles := clientFiles(services)
		serverFiles := serverFiles(services)
		require.Len(t, clientFiles, 2)
		require.Len(t, serverFiles, 2)

		decoder := codegen.SectionsCode(t, clientFiles[1].Section("response-decoder"))
		assert.Contains(t, decoder, `hdr.Get("goa-view")`)
		assert.Contains(t, decoder, "View: view")
		assert.Contains(t, decoder, "switch view")
		assert.Contains(t, decoder, "NewMethodMessageResultTypeWithViewsResultTiny(message)")
		assert.Contains(t, decoder, "NewMethodMessageResultTypeWithViewsResult(message)")
		encoder := codegen.SectionsCode(t, serverFiles[1].Section("response-encoder"))
		assert.Contains(t, encoder, `Append("goa-view", vres.View)`)
		assert.Contains(t, encoder, `return nil, goa.InvalidEnumValueError("view", vres.View`)
		testutil.AssertGo(t, "testdata/golden/viewed_result_dynamic_response_encoder.go.golden", encoder)
		response := services.Get("ServiceMessageResultTypeWithViews").Endpoints[0].Response
		require.Len(t, response.ServerConverts, 2)
		require.Equal(t, "default", response.ServerConverts[1].View)
		require.Same(t, response.ServerConverts[1].Convert, response.ServerConvert)
		require.Len(t, response.ClientConverts, 2)
		require.Equal(t, "default", response.ClientConverts[1].View)
		require.Same(t, response.ClientConverts[1].Convert, response.ClientConvert)
	})
}

func TestStreamingViewedResultSpecialization(t *testing.T) {
	t.Run("design selected view", func(t *testing.T) {
		root := RunGRPCDSL(t, testdata.ServerStreamingResultCollectionWithExplicitViewDSL)
		services := CreateGRPCServices(root)
		serverFiles := serverFiles(services)
		clientFiles := clientFiles(services)
		require.Len(t, serverFiles, 2)
		require.Len(t, clientFiles, 2)

		serverStruct := codegen.SectionsCode(t, serverFiles[0].Section("server-stream-struct-type"))
		clientStruct := codegen.SectionsCode(t, clientFiles[0].Section("client-stream-struct-type"))
		assert.NotContains(t, serverStruct, "\n\tview")
		assert.NotContains(t, clientStruct, "\n\tview")
		stream := services.Get("ServiceServerStreamingResultTypeCollectionWithExplicitView").Endpoints[0].ClientStream
		require.Len(t, stream.RecvConverts, 1)
		require.Equal(t, "tiny", stream.RecvConverts[0].View)
		require.Same(t, stream.RecvConverts[0].Convert, stream.RecvConvert)
	})

	t.Run("caller selected view", func(t *testing.T) {
		root := RunGRPCDSL(t, testdata.ServerStreamingResultWithViewsDSL)
		services := CreateGRPCServices(root)
		serverFiles := serverFiles(services)
		clientFiles := clientFiles(services)
		require.Len(t, serverFiles, 2)
		require.Len(t, clientFiles, 2)

		serverStruct := codegen.SectionsCode(t, serverFiles[0].Section("server-stream-struct-type"))
		clientStruct := codegen.SectionsCode(t, clientFiles[0].Section("client-stream-struct-type"))
		assert.Contains(t, serverStruct, "\n\tview")
		assert.Contains(t, clientStruct, "\n\tview")
		assert.Contains(t, serverStruct, "sentView string")
		assert.Contains(t, clientStruct, "viewSet bool")
		send := codegen.SectionsCode(t, serverFiles[0].Section("server-stream-send"))
		assert.Contains(t, send, `if view == "" {`)
		assert.Contains(t, send, `view = "default"`)
		assert.Contains(t, send, `if s.sentView != "" && view != s.sentView`)
		assert.Contains(t, send, `SetHeader(metadata.Pairs("goa-view", view))`)
		assert.Contains(t, send, `return goa.InvalidEnumValueError("view", view`)
		require.Less(t,
			strings.Index(send, `InvalidEnumValueError("view", view`),
			strings.Index(send, `SetHeader(metadata.Pairs("goa-view", view))`),
		)
		testutil.AssertGo(t, "testdata/golden/viewed_result_dynamic_stream_send.go.golden", send)
		recv := codegen.SectionsCode(t, clientFiles[0].Section("client-stream-recv"))
		assert.Contains(t, recv, `s.stream.Header()`)
		assert.Contains(t, recv, `goa.MissingFieldError("goa-view", "metadata")`)
		assert.Contains(t, recv, "switch s.view")
		assert.Contains(t, recv, "NewMethodServerStreamingUserTypeRPCResponseResultTypeViewTiny(v)")
		assert.Contains(t, recv, "NewMethodServerStreamingUserTypeRPCResponseResultTypeView(v)")
		stream := services.Get("ServiceServerStreamingUserTypeRPC").Endpoints[0].ServerStream
		require.Len(t, stream.SendConverts, 2)
		require.Equal(t, "default", stream.SendConverts[1].View)
		require.Same(t, stream.SendConverts[1].Convert, stream.SendConvert)
		clientStream := services.Get("ServiceServerStreamingUserTypeRPC").Endpoints[0].ClientStream
		require.Len(t, clientStream.RecvConverts, 2)
		require.Equal(t, "default", clientStream.RecvConverts[1].View)
		require.Same(t, clientStream.RecvConverts[1].Convert, clientStream.RecvConvert)
	})
}
