package codegen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/grpc/codegen/testdata"
)

type (
	// sectionExpectation holds the expected code for a section in a file.
	sectionExpectation struct {
		// Name is the section name.
		Name string
		// Code is the expected section code.
		Code *string
	}

	// testCase holds a test case.
	testCase struct {
		// Name is the name of the test case.
		Name string
		// DSL is the DSL to execute (test input).
		DSL func()
		// Sections is the expected code (test output).
		Sections []*sectionExpectation
	}
)

func TestStreaming(t *testing.T) {
	cases := []*testCase{
		// streaming result
		{"server-streaming", testdata.ServerStreamingUserTypeDSL, []*sectionExpectation{
			{"server-stream-struct-type", &testdata.ServerStreamingServerStructCode},
			{"server-stream-send", &testdata.ServerStreamingServerSendCode},
			{"server-stream-close", &testdata.ServerStreamingServerCloseCode},
			{"server-stream-set-view", nil},
			{"client-stream-struct-type", &testdata.ServerStreamingClientStructCode},
			{"client-stream-recv", &testdata.ServerStreamingClientRecvCode},
		}},
		{"server-streaming-result-with-views", testdata.ServerStreamingResultWithViewsDSL, []*sectionExpectation{
			{"server-stream-struct-type", &testdata.ServerStreamingResultWithViewsServerStructCode},
			{"server-stream-send", &testdata.ServerStreamingResultWithViewsServerSendCode},
			{"server-stream-set-view", &testdata.ServerStreamingResultWithViewsServerSetViewCode},
			{"client-stream-struct-type", &testdata.ServerStreamingResultWithViewsClientStructCode},
			{"client-stream-recv", &testdata.ServerStreamingResultWithViewsClientRecvCode},
			{"client-stream-set-view", &testdata.ServerStreamingResultWithViewsClientSetViewCode},
		}},
		{"server-streaming-result-collection-with-explicit-views", testdata.ServerStreamingResultCollectionWithExplicitViewDSL, []*sectionExpectation{
			{"server-stream-send", &testdata.ServerStreamingResultCollectionWithExplicitViewServerSendCode},
			{"server-stream-set-view", nil},
			{"client-stream-recv", &testdata.ServerStreamingResultCollectionWithExplicitViewClientRecvCode},
			{"client-stream-set-view", nil},
		}},
		{"server-streaming-primitive", testdata.ServerStreamingRPCDSL, []*sectionExpectation{
			{"server-stream-send", &testdata.ServerStreamingPrimitiveServerSendCode},
			{"client-stream-recv", &testdata.ServerStreamingPrimitiveClientRecvCode},
		}},
		{"server-streaming-array", testdata.ServerStreamingArrayDSL, []*sectionExpectation{
			{"server-stream-send", &testdata.ServerStreamingArrayServerSendCode},
			{"client-stream-recv", &testdata.ServerStreamingArrayClientRecvCode},
		}},
		{"server-streaming-map", testdata.ServerStreamingMapDSL, []*sectionExpectation{
			{"server-stream-send", &testdata.ServerStreamingMapServerSendCode},
			{"client-stream-recv", &testdata.ServerStreamingMapClientRecvCode},
		}},
		{"server-streaming-shared-result", testdata.ServerStreamingSharedResultRPCDSL, []*sectionExpectation{
			{"client-stream-recv", &testdata.ServerStreamingServerRPCSharedResultRecvCode},
		}},

		// streaming payload

		{"client-streaming", testdata.ClientStreamingRPCDSL, []*sectionExpectation{
			{"server-stream-struct-type", &testdata.ClientStreamingServerStructCode},
			{"server-stream-send", &testdata.ClientStreamingServerSendCode},
			{"server-stream-recv", &testdata.ClientStreamingServerRecvCode},
			{"client-stream-struct-type", &testdata.ClientStreamingClientStructCode},
			{"client-stream-send", &testdata.ClientStreamingClientSendCode},
			{"client-stream-recv", &testdata.ClientStreamingClientRecvCode},
		}},
		{"client-streaming-no-result", testdata.ClientStreamingNoResultDSL, []*sectionExpectation{
			{"server-stream-send", nil},
			{"server-stream-close", &testdata.ClientStreamingServerNoResultCloseCode},
			{"client-stream-recv", nil},
			{"client-stream-close", &testdata.ClientStreamingClientNoResultCloseCode},
		}},

		// bidirectional streaming

		{"bidirectional-streaming", testdata.BidirectionalStreamingRPCDSL, []*sectionExpectation{
			{"server-stream-struct-type", &testdata.BidirectionalStreamingServerStructCode},
			{"server-stream-send", &testdata.BidirectionalStreamingServerSendCode},
			{"server-stream-recv", &testdata.BidirectionalStreamingServerRecvCode},
			{"server-stream-close", &testdata.BidirectionalStreamingServerCloseCode},
			{"client-stream-struct-type", &testdata.BidirectionalStreamingClientStructCode},
			{"client-stream-send", &testdata.BidirectionalStreamingClientSendCode},
			{"client-stream-recv", &testdata.BidirectionalStreamingClientRecvCode},
			{"client-stream-close", &testdata.BidirectionalStreamingClientCloseCode},
		}},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := RunGRPCDSL(t, c.DSL)
			services := CreateGRPCServices(root)
			serverfs := ServerFiles("", services)
			if len(serverfs) < 2 {
				t.Fatalf("got %d server files, expected 2", len(serverfs))
			}
			clientfs := ClientFiles("", services)
			if len(clientfs) < 2 {
				t.Fatalf("got %d client files, expected 2", len(clientfs))
			}
			for _, s := range c.Sections {
				var (
					path     string
					sections []*codegen.SectionTemplate
				)
				if strings.HasPrefix(s.Name, "server-") {
					sections = serverfs[0].Section(s.Name)
					path = serverfs[0].Path
				} else {
					sections = clientfs[0].Section(s.Name)
					path = clientfs[0].Path
				}
				seclen := len(sections)
				code := make([]string, 0, seclen)
				for _, section := range sections {
					code = append(code, codegen.SectionCode(t, section))
				}
				switch {
				case seclen == 0 && s.Code == nil:
					// Test passed: no section found and no expected section code
				case seclen == 0 && s.Code != nil:
					// Test failed: no section found, but expected section code
					t.Errorf("invalid code for %s: got 0 %s sections, expected at least one", path, s.Name)
				case seclen > 0 && s.Code == nil:
					// Test failed: section exists in file, but no code expected.
					t.Errorf("invalid code for %s: got %d %s sections, expected 0.\n%s", path, seclen, s.Name, code)
				default:
					gen := strings.Join(code, "\n")
					assert.Equal(t, *s.Code, gen, "invalid code for %s %s section", path, s.Name)
				}
			}
		})
	}
}

func TestStreamingPayloadEnvelopeWithUnionPayload(t *testing.T) {
	root := RunGRPCDSL(t, testdata.ClientStreamingRPCWithUnionPayloadDSL)
	services := CreateGRPCServices(root)

	clientfs := ClientFiles("", services)
	require.Len(t, clientfs, 2)
	serverfs := ServerFiles("", services)
	require.Len(t, serverfs, 2)
	protofs := ProtoFiles("", services)
	require.Len(t, protofs, 1)

	requestEncoder := codegen.SectionsCode(t, clientfs[1].Section("request-encoder"))
	assert.Contains(t, requestEncoder, "InitialPayload")
	assert.Contains(t, requestEncoder, "MethodClientStreamingRPCWithUnionPayloadStreamingRequest")

	clientSend := codegen.SectionsCode(t, clientfs[0].Section("client-stream-send"))
	assert.Contains(t, clientSend, "StreamItem")
	assert.Contains(t, clientSend, "UploadChunk")

	serverInterface := codegen.SectionsCode(t, serverfs[0].Section("server-grpc-interface"))
	assert.Contains(t, serverInterface, "message, err := stream.Recv()")
	assert.Contains(t, serverInterface, "Decode(ctx, reqpb)")

	requestDecoder := codegen.SectionsCode(t, serverfs[1].Section("request-decoder"))
	assert.Contains(t, requestDecoder, "InitialPayload")
	assert.Contains(t, requestDecoder, "stream_item")
	assert.Contains(t, requestDecoder, "NewMethodClientStreamingRPCWithUnionPayloadPayload(message)")

	proto := sectionCode(t, protofs[0].SectionTemplates[1:]...)
	assert.Contains(t, proto, "message MethodClientStreamingRPCWithUnionPayloadStreamingRequest")
	assert.Contains(t, proto, "oneof body")
	assert.Contains(t, proto, "MethodClientStreamingRPCWithUnionPayloadRequest initial_payload")
	assert.Contains(t, proto, "MethodClientStreamingRPCWithUnionPayloadStreamItem stream_item")

	fpath := codegen.CreateTempFile(t, proto)
	assert.NoError(t, protoc(defaultProtocCmd, fpath, nil))
}

func TestStreamingPayloadLegacyCompat(t *testing.T) {
	root := RunGRPCDSL(t, testdata.BidirectionalStreamingRPCWithPayloadLegacyCompatDSL)
	services := CreateGRPCServices(root)

	serverfs := ServerFiles("", services)
	require.Len(t, serverfs, 2)
	clientfs := ClientFiles("", services)
	require.Len(t, clientfs, 2)

	// The server stream tracks the protocol spoken by the client.
	structCode := codegen.SectionsCode(t, serverfs[0].Section("server-stream-struct-type"))
	assert.Contains(t, structCode, "legacy bool")

	// The server stream reads raw stream item frames from legacy clients.
	serverRecv := codegen.SectionsCode(t, serverfs[0].Section("server-stream-recv"))
	assert.Contains(t, serverRecv, "if s.legacy {")
	assert.Contains(t, serverRecv, "RecvMsg")

	// The handler only waits for an initial payload frame from envelope clients.
	serverInterface := codegen.SectionsCode(t, serverfs[0].Section("server-grpc-interface"))
	assert.Contains(t, serverInterface, "goagrpc.UsesStreamEnvelope(ctx)")
	assert.Contains(t, serverInterface, "legacy:")

	// The request decoder dispatches to a legacy decoder that reads the
	// payload from request metadata.
	requestDecoder := codegen.SectionsCode(t, serverfs[1].Section("request-decoder"))
	assert.Contains(t, requestDecoder, "LegacyRequest(ctx, md)")
	assert.Contains(t, requestDecoder, `md.Get("a")`)
	assert.Contains(t, requestDecoder, "PayloadFromMetadata(")

	// Generated clients declare the envelope protocol in request metadata.
	requestEncoder := codegen.SectionsCode(t, clientfs[1].Section("request-encoder"))
	assert.Contains(t, requestEncoder, "goagrpc.StreamProtocolMetadataKey")

	// The wire contract for envelope clients is unchanged.
	protofs := ProtoFiles("", services)
	require.Len(t, protofs, 1)
	proto := sectionCode(t, protofs[0].SectionTemplates[1:]...)
	assert.Contains(t, proto, "oneof body")
	assert.Contains(t, proto, "initial_payload")
}
