package codegen

import (
	"testing"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
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

func TestServerStreaming(t *testing.T) {
	cases := []*testCase{
		{"mixed-endpoints", testdata.StreamingResultDSL, []*sectionExpectation{
			{"server-websocket-conn-configurer-struct", &testdata.MixedEndpointsConnConfigurerStructCode},
			{"server-websocket-conn-configurer-struct-init", &testdata.MixedEndpointsConnConfigurerInitCode},
		}},

		// streaming result
		{"streaming-result", testdata.StreamingResultDSL, []*sectionExpectation{
			{"server-handler-init", &testdata.StreamingResultServerHandlerInitCode},
			{"server-websocket-send", &testdata.StreamingResultServerStreamSendCode},
			{"server-websocket-close", &testdata.StreamingResultServerStreamCloseCode},
			{"server-websocket-set-view", nil},
		}},
		{"streaming-result-with-views", testdata.StreamingResultWithViewsDSL, []*sectionExpectation{
			{"server-websocket-send", &testdata.StreamingResultWithViewsServerStreamSendCode},
			{"server-websocket-close", &testdata.StreamingResultWithViewsServerStreamCloseCode},
			{"server-websocket-set-view", &testdata.StreamingResultWithViewsServerStreamSetViewCode},
		}},
		{"streaming-result-with-explicit-view", testdata.StreamingResultWithExplicitViewDSL, []*sectionExpectation{
			{"server-websocket-send", &testdata.StreamingResultWithExplicitViewServerStreamSendCode},
			{"server-websocket-set-view", nil},
		}},
		{"streaming-result-collection-with-views", testdata.StreamingResultCollectionWithViewsDSL, []*sectionExpectation{
			{"server-websocket-send", &testdata.StreamingResultCollectionWithViewsServerStreamSendCode},
			{"server-websocket-set-view", &testdata.StreamingResultCollectionWithViewsServerStreamSetViewCode},
		}},
		{"streaming-result-collection-with-explicit-view", testdata.StreamingResultCollectionWithExplicitViewDSL, []*sectionExpectation{
			{"server-websocket-send", &testdata.StreamingResultCollectionWithExplicitViewServerStreamSendCode},
			{"server-websocket-set-view", nil},
		}},
		{"streaming-result-primitive", testdata.StreamingResultPrimitiveDSL, []*sectionExpectation{
			{"server-websocket-send", &testdata.StreamingResultPrimitiveServerStreamSendCode},
			{"server-websocket-set-view", nil},
		}},
		{"streaming-result-primitive-array", testdata.StreamingResultPrimitiveArrayDSL, []*sectionExpectation{
			{"server-websocket-send", &testdata.StreamingResultPrimitiveArrayServerStreamSendCode},
		}},
		{"streaming-result-primitive-map", testdata.StreamingResultPrimitiveMapDSL, []*sectionExpectation{
			{"server-websocket-send", &testdata.StreamingResultPrimitiveMapServerStreamSendCode},
		}},
		{"streaming-result-user-type-array", testdata.StreamingResultUserTypeArrayDSL, []*sectionExpectation{
			{"server-websocket-send", &testdata.StreamingResultUserTypeArrayServerStreamSendCode},
		}},
		{"streaming-result-user-type-map", testdata.StreamingResultUserTypeMapDSL, []*sectionExpectation{
			{"server-websocket-send", &testdata.StreamingResultUserTypeMapServerStreamSendCode},
		}},
		{"streaming-result-no-payload", testdata.StreamingResultNoPayloadDSL, []*sectionExpectation{
			{"server-handler-init", &testdata.StreamingResultNoPayloadServerHandlerInitCode},
		}},

		// streaming payload

		{"streaming-payload", testdata.StreamingPayloadDSL, []*sectionExpectation{
			{"server-handler-init", &testdata.StreamingPayloadServerHandlerInitCode},
			{"server-websocket-send", &testdata.StreamingPayloadServerStreamSendCode},
			{"server-websocket-recv", &testdata.StreamingPayloadServerStreamRecvCode},
			{"server-websocket-close", nil},
			{"server-websocket-set-view", nil},
		}},
		{"streaming-payload-no-payload", testdata.StreamingPayloadNoPayloadDSL, []*sectionExpectation{
			{"server-handler-init", &testdata.StreamingPayloadNoPayloadServerHandlerInitCode},
			{"server-websocket-close", nil},
		}},
		{"streaming-payload-no-result", testdata.StreamingPayloadNoResultDSL, []*sectionExpectation{
			{"server-websocket-send", nil},
			{"server-websocket-recv", &testdata.StreamingPayloadNoResultServerStreamRecvCode},
			{"server-websocket-close", &testdata.StreamingPayloadNoResultServerStreamCloseCode},
			{"server-websocket-set-view", nil},
		}},
		{"streaming-payload-result-with-views", testdata.StreamingPayloadResultWithViewsDSL, []*sectionExpectation{
			{"server-websocket-send", &testdata.StreamingPayloadResultWithViewsServerStreamSendCode},
			{"server-websocket-recv", &testdata.StreamingPayloadResultWithViewsServerStreamRecvCode},
			{"server-websocket-close", nil},
			{"server-websocket-set-view", &testdata.StreamingPayloadResultWithViewsServerStreamSetViewCode},
		}},
		{"streaming-payload-result-with-explicit-view", testdata.StreamingPayloadResultWithExplicitViewDSL, []*sectionExpectation{
			{"server-websocket-send", &testdata.StreamingPayloadResultWithExplicitViewServerStreamSendCode},
			{"server-websocket-recv", &testdata.StreamingPayloadResultWithExplicitViewServerStreamRecvCode},
			{"server-websocket-set-view", nil},
		}},
		{"streaming-payload-result-collection-with-views", testdata.StreamingPayloadResultCollectionWithViewsDSL, []*sectionExpectation{
			{"server-websocket-send", &testdata.StreamingPayloadResultCollectionWithViewsServerStreamSendCode},
			{"server-websocket-recv", &testdata.StreamingPayloadResultCollectionWithViewsServerStreamRecvCode},
			{"server-websocket-set-view", &testdata.StreamingPayloadResultCollectionWithViewsServerStreamSetViewCode},
		}},
		{"streaming-payload-result-collection-with-explicit-view", testdata.StreamingPayloadResultCollectionWithExplicitViewDSL, []*sectionExpectation{
			{"server-websocket-send", &testdata.StreamingPayloadResultCollectionWithExplicitViewServerStreamSendCode},
			{"server-websocket-recv", &testdata.StreamingPayloadResultCollectionWithExplicitViewServerStreamRecvCode},
			{"server-websocket-set-view", nil},
		}},
		{"streaming-payload-primitive", testdata.StreamingPayloadPrimitiveDSL, []*sectionExpectation{
			{"server-websocket-send", &testdata.StreamingPayloadPrimitiveServerStreamSendCode},
			{"server-websocket-recv", &testdata.StreamingPayloadPrimitiveServerStreamRecvCode},
			{"server-websocket-set-view", nil},
		}},
		{"streaming-payload-primitive-array", testdata.StreamingPayloadPrimitiveArrayDSL, []*sectionExpectation{
			{"server-websocket-send", &testdata.StreamingPayloadPrimitiveArrayServerStreamSendCode},
			{"server-websocket-recv", &testdata.StreamingPayloadPrimitiveArrayServerStreamRecvCode},
		}},
		{"streaming-payload-primitive-map", testdata.StreamingPayloadPrimitiveMapDSL, []*sectionExpectation{
			{"server-websocket-send", &testdata.StreamingPayloadPrimitiveMapServerStreamSendCode},
			{"server-websocket-recv", &testdata.StreamingPayloadPrimitiveMapServerStreamRecvCode},
		}},
		{"streaming-payload-user-type-array", testdata.StreamingPayloadUserTypeArrayDSL, []*sectionExpectation{
			{"server-websocket-send", &testdata.StreamingPayloadUserTypeArrayServerStreamSendCode},
			{"server-websocket-recv", &testdata.StreamingPayloadUserTypeArrayServerStreamRecvCode},
		}},
		{"streaming-payload-user-type-map", testdata.StreamingPayloadUserTypeMapDSL, []*sectionExpectation{
			{"server-websocket-send", &testdata.StreamingPayloadUserTypeMapServerStreamSendCode},
			{"server-websocket-recv", &testdata.StreamingPayloadUserTypeMapServerStreamRecvCode},
		}},

		// bidirectional streaming

		{"bidirectional-streaming", testdata.BidirectionalStreamingDSL, []*sectionExpectation{
			{"server-handler-init", &testdata.BidirectionalStreamingServerHandlerInitCode},
			{"server-websocket-send", &testdata.BidirectionalStreamingServerStreamSendCode},
			{"server-websocket-recv", &testdata.BidirectionalStreamingServerStreamRecvCode},
			{"server-websocket-close", &testdata.BidirectionalStreamingServerStreamCloseCode},
			{"server-websocket-set-view", nil},
		}},
		{"bidirectional-streaming-no-payload", testdata.BidirectionalStreamingNoPayloadDSL, []*sectionExpectation{
			{"server-handler-init", &testdata.BidirectionalStreamingNoPayloadServerHandlerInitCode},
			{"server-websocket-close", &testdata.BidirectionalStreamingNoPayloadServerStreamCloseCode},
		}},
		{"bidirectional-streaming-result-with-views", testdata.BidirectionalStreamingResultWithViewsDSL, []*sectionExpectation{
			{"server-websocket-send", &testdata.BidirectionalStreamingResultWithViewsServerStreamSendCode},
			{"server-websocket-recv", &testdata.BidirectionalStreamingResultWithViewsServerStreamRecvCode},
			{"server-websocket-close", &testdata.BidirectionalStreamingResultWithViewsServerStreamCloseCode},
			{"server-websocket-set-view", &testdata.BidirectionalStreamingResultWithViewsServerStreamSetViewCode},
		}},
		{"bidirectional-streaming-result-with-explicit-view", testdata.BidirectionalStreamingResultWithExplicitViewDSL, []*sectionExpectation{
			{"server-websocket-send", &testdata.BidirectionalStreamingResultWithExplicitViewServerStreamSendCode},
			{"server-websocket-recv", &testdata.BidirectionalStreamingResultWithExplicitViewServerStreamRecvCode},
			{"server-websocket-set-view", nil},
		}},
		{"bidirectional-streaming-result-collection-with-views", testdata.BidirectionalStreamingResultCollectionWithViewsDSL, []*sectionExpectation{
			{"server-websocket-send", &testdata.BidirectionalStreamingResultCollectionWithViewsServerStreamSendCode},
			{"server-websocket-recv", &testdata.BidirectionalStreamingResultCollectionWithViewsServerStreamRecvCode},
			{"server-websocket-set-view", &testdata.BidirectionalStreamingResultCollectionWithViewsServerStreamSetViewCode},
		}},
		{"bidirectional-streaming-result-collection-with-explicit-view", testdata.BidirectionalStreamingResultCollectionWithExplicitViewDSL, []*sectionExpectation{
			{"server-websocket-send", &testdata.BidirectionalStreamingResultCollectionWithExplicitViewServerStreamSendCode},
			{"server-websocket-recv", &testdata.BidirectionalStreamingResultCollectionWithExplicitViewServerStreamRecvCode},
			{"server-websocket-set-view", nil},
		}},
		{"bidirectional-streaming-primitive", testdata.BidirectionalStreamingPrimitiveDSL, []*sectionExpectation{
			{"server-websocket-send", &testdata.BidirectionalStreamingPrimitiveServerStreamSendCode},
			{"server-websocket-recv", &testdata.BidirectionalStreamingPrimitiveServerStreamRecvCode},
			{"server-websocket-set-view", nil},
		}},
		{"bidirectional-streaming-primitive-array", testdata.BidirectionalStreamingPrimitiveArrayDSL, []*sectionExpectation{
			{"server-websocket-send", &testdata.BidirectionalStreamingPrimitiveArrayServerStreamSendCode},
			{"server-websocket-recv", &testdata.BidirectionalStreamingPrimitiveArrayServerStreamRecvCode},
		}},
		{"bidirectional-streaming-primitive-map", testdata.BidirectionalStreamingPrimitiveMapDSL, []*sectionExpectation{
			{"server-websocket-send", &testdata.BidirectionalStreamingPrimitiveMapServerStreamSendCode},
			{"server-websocket-recv", &testdata.BidirectionalStreamingPrimitiveMapServerStreamRecvCode},
		}},
		{"bidirectional-streaming-user-type-array", testdata.BidirectionalStreamingUserTypeArrayDSL, []*sectionExpectation{
			{"server-websocket-send", &testdata.BidirectionalStreamingUserTypeArrayServerStreamSendCode},
			{"server-websocket-recv", &testdata.BidirectionalStreamingUserTypeArrayServerStreamRecvCode},
		}},
		{"bidirectional-streaming-user-type-map", testdata.BidirectionalStreamingUserTypeMapDSL, []*sectionExpectation{
			{"server-websocket-send", &testdata.BidirectionalStreamingUserTypeMapServerStreamSendCode},
			{"server-websocket-recv", &testdata.BidirectionalStreamingUserTypeMapServerStreamRecvCode},
		}},
	}

	filesFn := func() []*codegen.File { return ServerFiles("", expr.Root) }
	runTests(t, cases, filesFn)
}

func TestClientStreaming(t *testing.T) {
	cases := []*testCase{
		{"mixed-endpoints", testdata.StreamingResultDSL, []*sectionExpectation{
			{"client-websocket-conn-configurer-struct", &testdata.MixedEndpointsConnConfigurerStructCode},
			{"client-websocket-conn-configurer-struct-init", &testdata.MixedEndpointsConnConfigurerInitCode},
		}},

		// streaming result
		{"streaming-result", testdata.StreamingResultDSL, []*sectionExpectation{
			{"client-endpoint-init", &testdata.StreamingResultClientEndpointCode},
			{"client-websocket-recv", &testdata.StreamingResultClientStreamRecvCode},
			{"client-websocket-close", nil},
			{"client-websocket-set-view", nil},
		}},
		{"streaming-result-with-views", testdata.StreamingResultWithViewsDSL, []*sectionExpectation{
			{"client-endpoint-init", &testdata.StreamingResultWithViewsClientEndpointCode},
			{"client-websocket-recv", &testdata.StreamingResultWithViewsClientStreamRecvCode},
			{"client-websocket-close", nil},
			{"client-websocket-set-view", &testdata.StreamingResultWithViewsClientStreamSetViewCode},
		}},
		{"streaming-result-with-explicit-view", testdata.StreamingResultWithExplicitViewDSL, []*sectionExpectation{
			{"client-endpoint-init", &testdata.StreamingResultWithExplicitViewClientEndpointCode},
			{"client-websocket-recv", &testdata.StreamingResultWithExplicitViewClientStreamRecvCode},
			{"client-websocket-set-view", nil},
		}},
		{"streaming-result-collection-with-views", testdata.StreamingResultCollectionWithViewsDSL, []*sectionExpectation{
			{"client-websocket-recv", &testdata.StreamingResultCollectionWithViewsClientStreamRecvCode},
			{"client-websocket-set-view", &testdata.StreamingResultCollectionWithViewsClientStreamSetViewCode},
		}},
		{"streaming-result-collection-with-explicit-view", testdata.StreamingResultCollectionWithExplicitViewDSL, []*sectionExpectation{
			{"client-endpoint-init", &testdata.StreamingResultCollectionWithExplicitViewClientEndpointCode},
			{"client-websocket-recv", &testdata.StreamingResultCollectionWithExplicitViewClientStreamRecvCode},
			{"client-websocket-set-view", nil},
		}},
		{"streaming-result-primitive", testdata.StreamingResultPrimitiveDSL, []*sectionExpectation{
			{"client-websocket-recv", &testdata.StreamingResultPrimitiveClientStreamRecvCode},
			{"client-websocket-set-view", nil},
		}},
		{"streaming-result-primitive-array", testdata.StreamingResultPrimitiveArrayDSL, []*sectionExpectation{
			{"client-websocket-recv", &testdata.StreamingResultPrimitiveArrayClientStreamRecvCode},
		}},
		{"streaming-result-primitive-map", testdata.StreamingResultPrimitiveMapDSL, []*sectionExpectation{
			{"client-websocket-recv", &testdata.StreamingResultPrimitiveMapClientStreamRecvCode},
		}},
		{"streaming-result-user-type-array", testdata.StreamingResultUserTypeArrayDSL, []*sectionExpectation{
			{"client-websocket-recv", &testdata.StreamingResultUserTypeArrayClientStreamRecvCode},
		}},
		{"streaming-result-user-type-map", testdata.StreamingResultUserTypeMapDSL, []*sectionExpectation{
			{"client-websocket-recv", &testdata.StreamingResultUserTypeMapClientStreamRecvCode},
		}},
		{"streaming-result-no-payload", testdata.StreamingResultNoPayloadDSL, []*sectionExpectation{
			{"client-endpoint-init", &testdata.StreamingResultNoPayloadClientEndpointCode},
		}},

		// streaming payload

		{"streaming-payload", testdata.StreamingPayloadDSL, []*sectionExpectation{
			{"client-endpoint-init", &testdata.StreamingPayloadClientEndpointCode},
			{"client-websocket-send", &testdata.StreamingPayloadClientStreamSendCode},
			{"client-websocket-recv", &testdata.StreamingPayloadClientStreamRecvCode},
			{"client-websocket-close", nil},
			{"client-websocket-set-view", nil},
		}},
		{"streaming-payload-no-payload", testdata.StreamingPayloadNoPayloadDSL, []*sectionExpectation{
			{"client-endpoint-init", &testdata.StreamingPayloadNoPayloadClientEndpointCode},
			{"client-websocket-send", &testdata.StreamingPayloadNoPayloadClientStreamSendCode},
			{"client-websocket-recv", &testdata.StreamingPayloadNoPayloadClientStreamRecvCode},
			{"client-websocket-close", nil},
		}},
		{"streaming-payload-no-result", testdata.StreamingPayloadNoResultDSL, []*sectionExpectation{
			{"client-websocket-send", &testdata.StreamingPayloadNoResultClientStreamSendCode},
			{"client-websocket-recv", nil},
			{"client-websocket-close", &testdata.StreamingPayloadNoResultClientStreamCloseCode},
			{"client-websocket-set-view", nil},
		}},
		{"streaming-payload-result-with-views", testdata.StreamingPayloadResultWithViewsDSL, []*sectionExpectation{
			{"client-websocket-send", &testdata.StreamingPayloadResultWithViewsClientStreamSendCode},
			{"client-websocket-recv", &testdata.StreamingPayloadResultWithViewsClientStreamRecvCode},
			{"client-websocket-close", nil},
			{"client-websocket-set-view", &testdata.StreamingPayloadResultWithViewsClientStreamSetViewCode},
		}},
		{"streaming-payload-result-with-explicit-view", testdata.StreamingPayloadResultWithExplicitViewDSL, []*sectionExpectation{
			{"client-websocket-send", &testdata.StreamingPayloadResultWithExplicitViewClientStreamSendCode},
			{"client-websocket-recv", &testdata.StreamingPayloadResultWithExplicitViewClientStreamRecvCode},
			{"client-websocket-set-view", nil},
		}},
		{"streaming-payload-result-collection-with-views", testdata.StreamingPayloadResultCollectionWithViewsDSL, []*sectionExpectation{
			{"client-websocket-send", &testdata.StreamingPayloadResultCollectionWithViewsClientStreamSendCode},
			{"client-websocket-recv", &testdata.StreamingPayloadResultCollectionWithViewsClientStreamRecvCode},
			{"client-websocket-set-view", &testdata.StreamingPayloadResultCollectionWithViewsClientStreamSetViewCode},
		}},
		{"streaming-payload-result-collection-with-explicit-view", testdata.StreamingPayloadResultCollectionWithExplicitViewDSL, []*sectionExpectation{
			{"client-websocket-send", &testdata.StreamingPayloadResultCollectionWithExplicitViewClientStreamSendCode},
			{"client-websocket-recv", &testdata.StreamingPayloadResultCollectionWithExplicitViewClientStreamRecvCode},
			{"client-websocket-set-view", nil},
		}},
		{"streaming-payload-primitive", testdata.StreamingPayloadPrimitiveDSL, []*sectionExpectation{
			{"client-websocket-send", &testdata.StreamingPayloadPrimitiveClientStreamSendCode},
			{"client-websocket-recv", &testdata.StreamingPayloadPrimitiveClientStreamRecvCode},
			{"client-websocket-set-view", nil},
		}},
		{"streaming-payload-primitive-array", testdata.StreamingPayloadPrimitiveArrayDSL, []*sectionExpectation{
			{"client-websocket-send", &testdata.StreamingPayloadPrimitiveArrayClientStreamSendCode},
			{"client-websocket-recv", &testdata.StreamingPayloadPrimitiveArrayClientStreamRecvCode},
		}},
		{"streaming-payload-primitive-map", testdata.StreamingPayloadPrimitiveMapDSL, []*sectionExpectation{
			{"client-websocket-send", &testdata.StreamingPayloadPrimitiveMapClientStreamSendCode},
			{"client-websocket-recv", &testdata.StreamingPayloadPrimitiveMapClientStreamRecvCode},
		}},
		{"streaming-payload-user-type-array", testdata.StreamingPayloadUserTypeArrayDSL, []*sectionExpectation{
			{"client-websocket-send", &testdata.StreamingPayloadUserTypeArrayClientStreamSendCode},
			{"client-websocket-recv", &testdata.StreamingPayloadUserTypeArrayClientStreamRecvCode},
		}},
		{"streaming-payload-user-type-map", testdata.StreamingPayloadUserTypeMapDSL, []*sectionExpectation{
			{"client-websocket-send", &testdata.StreamingPayloadUserTypeMapClientStreamSendCode},
			{"client-websocket-recv", &testdata.StreamingPayloadUserTypeMapClientStreamRecvCode},
		}},

		// bidirectional streaming

		{"bidirectional-streaming", testdata.BidirectionalStreamingDSL, []*sectionExpectation{
			{"client-endpoint-init", &testdata.BidirectionalStreamingClientEndpointCode},
			{"client-websocket-send", &testdata.BidirectionalStreamingClientStreamSendCode},
			{"client-websocket-recv", &testdata.BidirectionalStreamingClientStreamRecvCode},
			{"client-websocket-close", &testdata.BidirectionalStreamingClientStreamCloseCode},
			{"client-websocket-set-view", nil},
		}},
		{"bidirectional-streaming-no-payload", testdata.BidirectionalStreamingNoPayloadDSL, []*sectionExpectation{
			{"client-endpoint-init", &testdata.BidirectionalStreamingNoPayloadClientEndpointCode},
			{"client-websocket-send", &testdata.BidirectionalStreamingNoPayloadClientStreamSendCode},
			{"client-websocket-recv", &testdata.BidirectionalStreamingNoPayloadClientStreamRecvCode},
			{"client-websocket-close", &testdata.BidirectionalStreamingNoPayloadClientStreamCloseCode},
		}},
		{"bidirectional-streaming-result-with-views", testdata.BidirectionalStreamingResultWithViewsDSL, []*sectionExpectation{
			{"client-websocket-send", &testdata.BidirectionalStreamingResultWithViewsClientStreamSendCode},
			{"client-websocket-recv", &testdata.BidirectionalStreamingResultWithViewsClientStreamRecvCode},
			{"client-websocket-close", &testdata.BidirectionalStreamingResultWithViewsClientStreamCloseCode},
			{"client-websocket-set-view", &testdata.BidirectionalStreamingResultWithViewsClientStreamSetViewCode},
		}},
		{"bidirectional-streaming-result-with-explicit-view", testdata.BidirectionalStreamingResultWithExplicitViewDSL, []*sectionExpectation{
			{"client-websocket-send", &testdata.BidirectionalStreamingResultWithExplicitViewClientStreamSendCode},
			{"client-websocket-recv", &testdata.BidirectionalStreamingResultWithExplicitViewClientStreamRecvCode},
			{"client-websocket-set-view", nil},
		}},
		{"bidirectional-streaming-result-collection-with-views", testdata.BidirectionalStreamingResultCollectionWithViewsDSL, []*sectionExpectation{
			{"client-websocket-send", &testdata.BidirectionalStreamingResultCollectionWithViewsClientStreamSendCode},
			{"client-websocket-recv", &testdata.BidirectionalStreamingResultCollectionWithViewsClientStreamRecvCode},
			{"client-websocket-set-view", &testdata.BidirectionalStreamingResultCollectionWithViewsClientStreamSetViewCode},
		}},
		{"bidirectional-streaming-result-collection-with-explicit-view", testdata.BidirectionalStreamingResultCollectionWithExplicitViewDSL, []*sectionExpectation{
			{"client-websocket-send", &testdata.BidirectionalStreamingResultCollectionWithExplicitViewClientStreamSendCode},
			{"client-websocket-recv", &testdata.BidirectionalStreamingResultCollectionWithExplicitViewClientStreamRecvCode},
			{"client-websocket-set-view", nil},
		}},
		{"bidirectional-streaming-primitive", testdata.BidirectionalStreamingPrimitiveDSL, []*sectionExpectation{
			{"client-websocket-send", &testdata.BidirectionalStreamingPrimitiveClientStreamSendCode},
			{"client-websocket-recv", &testdata.BidirectionalStreamingPrimitiveClientStreamRecvCode},
			{"client-websocket-set-view", nil},
		}},
		{"bidirectional-streaming-primitive-array", testdata.BidirectionalStreamingPrimitiveArrayDSL, []*sectionExpectation{
			{"client-websocket-send", &testdata.BidirectionalStreamingPrimitiveArrayClientStreamSendCode},
			{"client-websocket-recv", &testdata.BidirectionalStreamingPrimitiveArrayClientStreamRecvCode},
		}},
		{"bidirectional-streaming-primitive-map", testdata.BidirectionalStreamingPrimitiveMapDSL, []*sectionExpectation{
			{"client-websocket-send", &testdata.BidirectionalStreamingPrimitiveMapClientStreamSendCode},
			{"client-websocket-recv", &testdata.BidirectionalStreamingPrimitiveMapClientStreamRecvCode},
		}},
		{"bidirectional-streaming-user-type-array", testdata.BidirectionalStreamingUserTypeArrayDSL, []*sectionExpectation{
			{"client-websocket-send", &testdata.BidirectionalStreamingUserTypeArrayClientStreamSendCode},
			{"client-websocket-recv", &testdata.BidirectionalStreamingUserTypeArrayClientStreamRecvCode},
		}},
		{"bidirectional-streaming-user-type-map", testdata.BidirectionalStreamingUserTypeMapDSL, []*sectionExpectation{
			{"client-websocket-send", &testdata.BidirectionalStreamingUserTypeMapClientStreamSendCode},
			{"client-websocket-recv", &testdata.BidirectionalStreamingUserTypeMapClientStreamRecvCode},
		}},
	}
	filesFn := func() []*codegen.File { return ClientFiles("", expr.Root) }
	runTests(t, cases, filesFn)
}

func runTests(t *testing.T, cases []*testCase, filesFn func() []*codegen.File) {
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunHTTPDSL(t, c.DSL)
			fs := filesFn()
			if len(fs) < 2 {
				t.Fatalf("got %d files, expected 2", len(fs))
			}
			for _, s := range c.Sections {
				var code string
				var f *codegen.File
				if s.Name == "server-handler-init" || s.Name == "client-endpoint-init" {
					// server.go || client.go
					f = fs[0]
				} else {
					// websocket.go
					f = fs[1]
				}
				sections := f.Section(s.Name)
				seclen := len(sections)
				if seclen > 0 {
					code = codegen.SectionCode(t, sections[0])
				}
				switch {
				case seclen == 0 && s.Code == nil:
					// Test passed: no section found and no expected section code
				case seclen == 0 && s.Code != nil:
					// Test failed: no section found, but expected section code
					t.Errorf("invalid code for %s: got 0 %s sections, expected at least one", f.Path, s.Name)
				case seclen > 0 && s.Code == nil:
					// Test failed: section exists in file, but no code expected.
					t.Errorf("invalid code for %s: got %d %s sections, expected 0.\n%s", f.Path, seclen, s.Name, code)
				default:
					if code != *s.Code {
						t.Errorf("invalid code for %s %s section, got:\n%s\ngot vs. expected:\n%s", f.Path, s.Name, code, codegen.Diff(t, code, *s.Code))
					}
				}
			}
		})
	}
}
