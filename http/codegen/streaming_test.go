package codegen

import (
	"testing"

	"goa.design/goa/codegen"
	"goa.design/goa/expr"
	"goa.design/goa/http/codegen/testdata"
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
			{"server-stream-conn-configurer-struct", &testdata.MixedEndpointsConnConfigurerStructCode},
			{"server-stream-conn-configurer-struct-init", &testdata.MixedEndpointsConnConfigurerInitCode},
		}},

		// streaming result
		{"streaming-result", testdata.StreamingResultDSL, []*sectionExpectation{
			{"server-handler-init", &testdata.StreamingResultServerHandlerInitCode},
			{"server-stream-send", &testdata.StreamingResultServerStreamSendCode},
			{"server-stream-close", &testdata.StreamingResultServerStreamCloseCode},
			{"server-stream-set-view", nil},
		}},
		{"streaming-result-with-views", testdata.StreamingResultWithViewsDSL, []*sectionExpectation{
			{"server-stream-send", &testdata.StreamingResultWithViewsServerStreamSendCode},
			{"server-stream-close", &testdata.StreamingResultWithViewsServerStreamCloseCode},
			{"server-stream-set-view", &testdata.StreamingResultWithViewsServerStreamSetViewCode},
		}},
		{"streaming-result-with-explicit-view", testdata.StreamingResultWithExplicitViewDSL, []*sectionExpectation{
			{"server-stream-send", &testdata.StreamingResultWithExplicitViewServerStreamSendCode},
			{"server-stream-set-view", nil},
		}},
		{"streaming-result-collection-with-views", testdata.StreamingResultCollectionWithViewsDSL, []*sectionExpectation{
			{"server-stream-send", &testdata.StreamingResultCollectionWithViewsServerStreamSendCode},
			{"server-stream-set-view", &testdata.StreamingResultCollectionWithViewsServerStreamSetViewCode},
		}},
		{"streaming-result-collection-with-explicit-view", testdata.StreamingResultCollectionWithExplicitViewDSL, []*sectionExpectation{
			{"server-stream-send", &testdata.StreamingResultCollectionWithExplicitViewServerStreamSendCode},
			{"server-stream-set-view", nil},
		}},
		{"streaming-result-primitive", testdata.StreamingResultPrimitiveDSL, []*sectionExpectation{
			{"server-stream-send", &testdata.StreamingResultPrimitiveServerStreamSendCode},
			{"server-stream-set-view", nil},
		}},
		{"streaming-result-primitive-array", testdata.StreamingResultPrimitiveArrayDSL, []*sectionExpectation{
			{"server-stream-send", &testdata.StreamingResultPrimitiveArrayServerStreamSendCode},
		}},
		{"streaming-result-primitive-map", testdata.StreamingResultPrimitiveMapDSL, []*sectionExpectation{
			{"server-stream-send", &testdata.StreamingResultPrimitiveMapServerStreamSendCode},
		}},
		{"streaming-result-user-type-array", testdata.StreamingResultUserTypeArrayDSL, []*sectionExpectation{
			{"server-stream-send", &testdata.StreamingResultUserTypeArrayServerStreamSendCode},
		}},
		{"streaming-result-user-type-map", testdata.StreamingResultUserTypeMapDSL, []*sectionExpectation{
			{"server-stream-send", &testdata.StreamingResultUserTypeMapServerStreamSendCode},
		}},
		{"streaming-result-no-payload", testdata.StreamingResultNoPayloadDSL, []*sectionExpectation{
			{"server-handler-init", &testdata.StreamingResultNoPayloadServerHandlerInitCode},
		}},

		// streaming payload

		{"streaming-payload", testdata.StreamingPayloadDSL, []*sectionExpectation{
			{"server-handler-init", &testdata.StreamingPayloadServerHandlerInitCode},
			{"server-stream-send", &testdata.StreamingPayloadServerStreamSendCode},
			{"server-stream-recv", &testdata.StreamingPayloadServerStreamRecvCode},
			{"server-stream-close", nil},
			{"server-stream-set-view", nil},
		}},
		{"streaming-payload-no-payload", testdata.StreamingPayloadNoPayloadDSL, []*sectionExpectation{
			{"server-handler-init", &testdata.StreamingPayloadNoPayloadServerHandlerInitCode},
			{"server-stream-close", nil},
		}},
		{"streaming-payload-no-result", testdata.StreamingPayloadNoResultDSL, []*sectionExpectation{
			{"server-stream-send", nil},
			{"server-stream-recv", &testdata.StreamingPayloadNoResultServerStreamRecvCode},
			{"server-stream-close", &testdata.StreamingPayloadNoResultServerStreamCloseCode},
			{"server-stream-set-view", nil},
		}},
		{"streaming-payload-result-with-views", testdata.StreamingPayloadResultWithViewsDSL, []*sectionExpectation{
			{"server-stream-send", &testdata.StreamingPayloadResultWithViewsServerStreamSendCode},
			{"server-stream-recv", &testdata.StreamingPayloadResultWithViewsServerStreamRecvCode},
			{"server-stream-close", nil},
			{"server-stream-set-view", &testdata.StreamingPayloadResultWithViewsServerStreamSetViewCode},
		}},
		{"streaming-payload-result-with-explicit-view", testdata.StreamingPayloadResultWithExplicitViewDSL, []*sectionExpectation{
			{"server-stream-send", &testdata.StreamingPayloadResultWithExplicitViewServerStreamSendCode},
			{"server-stream-recv", &testdata.StreamingPayloadResultWithExplicitViewServerStreamRecvCode},
			{"server-stream-set-view", nil},
		}},
		{"streaming-payload-result-collection-with-views", testdata.StreamingPayloadResultCollectionWithViewsDSL, []*sectionExpectation{
			{"server-stream-send", &testdata.StreamingPayloadResultCollectionWithViewsServerStreamSendCode},
			{"server-stream-recv", &testdata.StreamingPayloadResultCollectionWithViewsServerStreamRecvCode},
			{"server-stream-set-view", &testdata.StreamingPayloadResultCollectionWithViewsServerStreamSetViewCode},
		}},
		{"streaming-payload-result-collection-with-explicit-view", testdata.StreamingPayloadResultCollectionWithExplicitViewDSL, []*sectionExpectation{
			{"server-stream-send", &testdata.StreamingPayloadResultCollectionWithExplicitViewServerStreamSendCode},
			{"server-stream-recv", &testdata.StreamingPayloadResultCollectionWithExplicitViewServerStreamRecvCode},
			{"server-stream-set-view", nil},
		}},
		{"streaming-payload-primitive", testdata.StreamingPayloadPrimitiveDSL, []*sectionExpectation{
			{"server-stream-send", &testdata.StreamingPayloadPrimitiveServerStreamSendCode},
			{"server-stream-recv", &testdata.StreamingPayloadPrimitiveServerStreamRecvCode},
			{"server-stream-set-view", nil},
		}},
		{"streaming-payload-primitive-array", testdata.StreamingPayloadPrimitiveArrayDSL, []*sectionExpectation{
			{"server-stream-send", &testdata.StreamingPayloadPrimitiveArrayServerStreamSendCode},
			{"server-stream-recv", &testdata.StreamingPayloadPrimitiveArrayServerStreamRecvCode},
		}},
		{"streaming-payload-primitive-map", testdata.StreamingPayloadPrimitiveMapDSL, []*sectionExpectation{
			{"server-stream-send", &testdata.StreamingPayloadPrimitiveMapServerStreamSendCode},
			{"server-stream-recv", &testdata.StreamingPayloadPrimitiveMapServerStreamRecvCode},
		}},
		{"streaming-payload-user-type-array", testdata.StreamingPayloadUserTypeArrayDSL, []*sectionExpectation{
			{"server-stream-send", &testdata.StreamingPayloadUserTypeArrayServerStreamSendCode},
			{"server-stream-recv", &testdata.StreamingPayloadUserTypeArrayServerStreamRecvCode},
		}},
		{"streaming-payload-user-type-map", testdata.StreamingPayloadUserTypeMapDSL, []*sectionExpectation{
			{"server-stream-send", &testdata.StreamingPayloadUserTypeMapServerStreamSendCode},
			{"server-stream-recv", &testdata.StreamingPayloadUserTypeMapServerStreamRecvCode},
		}},

		// bidirectional streaming

		{"bidirectional-streaming", testdata.BidirectionalStreamingDSL, []*sectionExpectation{
			{"server-handler-init", &testdata.BidirectionalStreamingServerHandlerInitCode},
			{"server-stream-send", &testdata.BidirectionalStreamingServerStreamSendCode},
			{"server-stream-recv", &testdata.BidirectionalStreamingServerStreamRecvCode},
			{"server-stream-close", &testdata.BidirectionalStreamingServerStreamCloseCode},
			{"server-stream-set-view", nil},
		}},
		{"bidirectional-streaming-no-payload", testdata.BidirectionalStreamingNoPayloadDSL, []*sectionExpectation{
			{"server-handler-init", &testdata.BidirectionalStreamingNoPayloadServerHandlerInitCode},
			{"server-stream-close", &testdata.BidirectionalStreamingNoPayloadServerStreamCloseCode},
		}},
		{"bidirectional-streaming-result-with-views", testdata.BidirectionalStreamingResultWithViewsDSL, []*sectionExpectation{
			{"server-stream-send", &testdata.BidirectionalStreamingResultWithViewsServerStreamSendCode},
			{"server-stream-recv", &testdata.BidirectionalStreamingResultWithViewsServerStreamRecvCode},
			{"server-stream-close", &testdata.BidirectionalStreamingResultWithViewsServerStreamCloseCode},
			{"server-stream-set-view", &testdata.BidirectionalStreamingResultWithViewsServerStreamSetViewCode},
		}},
		{"bidirectional-streaming-result-with-explicit-view", testdata.BidirectionalStreamingResultWithExplicitViewDSL, []*sectionExpectation{
			{"server-stream-send", &testdata.BidirectionalStreamingResultWithExplicitViewServerStreamSendCode},
			{"server-stream-recv", &testdata.BidirectionalStreamingResultWithExplicitViewServerStreamRecvCode},
			{"server-stream-set-view", nil},
		}},
		{"bidirectional-streaming-result-collection-with-views", testdata.BidirectionalStreamingResultCollectionWithViewsDSL, []*sectionExpectation{
			{"server-stream-send", &testdata.BidirectionalStreamingResultCollectionWithViewsServerStreamSendCode},
			{"server-stream-recv", &testdata.BidirectionalStreamingResultCollectionWithViewsServerStreamRecvCode},
			{"server-stream-set-view", &testdata.BidirectionalStreamingResultCollectionWithViewsServerStreamSetViewCode},
		}},
		{"bidirectional-streaming-result-collection-with-explicit-view", testdata.BidirectionalStreamingResultCollectionWithExplicitViewDSL, []*sectionExpectation{
			{"server-stream-send", &testdata.BidirectionalStreamingResultCollectionWithExplicitViewServerStreamSendCode},
			{"server-stream-recv", &testdata.BidirectionalStreamingResultCollectionWithExplicitViewServerStreamRecvCode},
			{"server-stream-set-view", nil},
		}},
		{"bidirectional-streaming-primitive", testdata.BidirectionalStreamingPrimitiveDSL, []*sectionExpectation{
			{"server-stream-send", &testdata.BidirectionalStreamingPrimitiveServerStreamSendCode},
			{"server-stream-recv", &testdata.BidirectionalStreamingPrimitiveServerStreamRecvCode},
			{"server-stream-set-view", nil},
		}},
		{"bidirectional-streaming-primitive-array", testdata.BidirectionalStreamingPrimitiveArrayDSL, []*sectionExpectation{
			{"server-stream-send", &testdata.BidirectionalStreamingPrimitiveArrayServerStreamSendCode},
			{"server-stream-recv", &testdata.BidirectionalStreamingPrimitiveArrayServerStreamRecvCode},
		}},
		{"bidirectional-streaming-primitive-map", testdata.BidirectionalStreamingPrimitiveMapDSL, []*sectionExpectation{
			{"server-stream-send", &testdata.BidirectionalStreamingPrimitiveMapServerStreamSendCode},
			{"server-stream-recv", &testdata.BidirectionalStreamingPrimitiveMapServerStreamRecvCode},
		}},
		{"bidirectional-streaming-user-type-array", testdata.BidirectionalStreamingUserTypeArrayDSL, []*sectionExpectation{
			{"server-stream-send", &testdata.BidirectionalStreamingUserTypeArrayServerStreamSendCode},
			{"server-stream-recv", &testdata.BidirectionalStreamingUserTypeArrayServerStreamRecvCode},
		}},
		{"bidirectional-streaming-user-type-map", testdata.BidirectionalStreamingUserTypeMapDSL, []*sectionExpectation{
			{"server-stream-send", &testdata.BidirectionalStreamingUserTypeMapServerStreamSendCode},
			{"server-stream-recv", &testdata.BidirectionalStreamingUserTypeMapServerStreamRecvCode},
		}},
	}

	filesFn := func() []*codegen.File { return ServerFiles("", expr.Root) }
	runTests(t, cases, filesFn)
}

func TestClientStreaming(t *testing.T) {
	cases := []*testCase{
		{"mixed-endpoints", testdata.StreamingResultDSL, []*sectionExpectation{
			{"client-stream-conn-configurer-struct", &testdata.MixedEndpointsConnConfigurerStructCode},
			{"client-stream-conn-configurer-struct-init", &testdata.MixedEndpointsConnConfigurerInitCode},
		}},

		// streaming result
		{"streaming-result", testdata.StreamingResultDSL, []*sectionExpectation{
			{"client-endpoint-init", &testdata.StreamingResultClientEndpointCode},
			{"client-stream-recv", &testdata.StreamingResultClientStreamRecvCode},
			{"client-stream-close", nil},
			{"client-stream-set-view", nil},
		}},
		{"streaming-result-with-views", testdata.StreamingResultWithViewsDSL, []*sectionExpectation{
			{"client-endpoint-init", &testdata.StreamingResultWithViewsClientEndpointCode},
			{"client-stream-recv", &testdata.StreamingResultWithViewsClientStreamRecvCode},
			{"client-stream-close", nil},
			{"client-stream-set-view", &testdata.StreamingResultWithViewsClientStreamSetViewCode},
		}},
		{"streaming-result-with-explicit-view", testdata.StreamingResultWithExplicitViewDSL, []*sectionExpectation{
			{"client-endpoint-init", &testdata.StreamingResultWithExplicitViewClientEndpointCode},
			{"client-stream-recv", &testdata.StreamingResultWithExplicitViewClientStreamRecvCode},
			{"client-stream-set-view", nil},
		}},
		{"streaming-result-collection-with-views", testdata.StreamingResultCollectionWithViewsDSL, []*sectionExpectation{
			{"client-stream-recv", &testdata.StreamingResultCollectionWithViewsClientStreamRecvCode},
			{"client-stream-set-view", &testdata.StreamingResultCollectionWithViewsClientStreamSetViewCode},
		}},
		{"streaming-result-collection-with-explicit-view", testdata.StreamingResultCollectionWithExplicitViewDSL, []*sectionExpectation{
			{"client-endpoint-init", &testdata.StreamingResultCollectionWithExplicitViewClientEndpointCode},
			{"client-stream-recv", &testdata.StreamingResultCollectionWithExplicitViewClientStreamRecvCode},
			{"client-stream-set-view", nil},
		}},
		{"streaming-result-primitive", testdata.StreamingResultPrimitiveDSL, []*sectionExpectation{
			{"client-stream-recv", &testdata.StreamingResultPrimitiveClientStreamRecvCode},
			{"client-stream-set-view", nil},
		}},
		{"streaming-result-primitive-array", testdata.StreamingResultPrimitiveArrayDSL, []*sectionExpectation{
			{"client-stream-recv", &testdata.StreamingResultPrimitiveArrayClientStreamRecvCode},
		}},
		{"streaming-result-primitive-map", testdata.StreamingResultPrimitiveMapDSL, []*sectionExpectation{
			{"client-stream-recv", &testdata.StreamingResultPrimitiveMapClientStreamRecvCode},
		}},
		{"streaming-result-user-type-array", testdata.StreamingResultUserTypeArrayDSL, []*sectionExpectation{
			{"client-stream-recv", &testdata.StreamingResultUserTypeArrayClientStreamRecvCode},
		}},
		{"streaming-result-user-type-map", testdata.StreamingResultUserTypeMapDSL, []*sectionExpectation{
			{"client-stream-recv", &testdata.StreamingResultUserTypeMapClientStreamRecvCode},
		}},
		{"streaming-result-no-payload", testdata.StreamingResultNoPayloadDSL, []*sectionExpectation{
			{"client-endpoint-init", &testdata.StreamingResultNoPayloadClientEndpointCode},
		}},

		// streaming payload

		{"streaming-payload", testdata.StreamingPayloadDSL, []*sectionExpectation{
			{"client-endpoint-init", &testdata.StreamingPayloadClientEndpointCode},
			{"client-stream-send", &testdata.StreamingPayloadClientStreamSendCode},
			{"client-stream-recv", &testdata.StreamingPayloadClientStreamRecvCode},
			{"client-stream-close", nil},
			{"client-stream-set-view", nil},
		}},
		{"streaming-payload-no-payload", testdata.StreamingPayloadNoPayloadDSL, []*sectionExpectation{
			{"client-endpoint-init", &testdata.StreamingPayloadNoPayloadClientEndpointCode},
			{"client-stream-send", &testdata.StreamingPayloadNoPayloadClientStreamSendCode},
			{"client-stream-recv", &testdata.StreamingPayloadNoPayloadClientStreamRecvCode},
			{"client-stream-close", nil},
		}},
		{"streaming-payload-no-result", testdata.StreamingPayloadNoResultDSL, []*sectionExpectation{
			{"client-stream-send", &testdata.StreamingPayloadNoResultClientStreamSendCode},
			{"client-stream-recv", nil},
			{"client-stream-close", &testdata.StreamingPayloadNoResultClientStreamCloseCode},
			{"client-stream-set-view", nil},
		}},
		{"streaming-payload-result-with-views", testdata.StreamingPayloadResultWithViewsDSL, []*sectionExpectation{
			{"client-stream-send", &testdata.StreamingPayloadResultWithViewsClientStreamSendCode},
			{"client-stream-recv", &testdata.StreamingPayloadResultWithViewsClientStreamRecvCode},
			{"client-stream-close", nil},
			{"client-stream-set-view", &testdata.StreamingPayloadResultWithViewsClientStreamSetViewCode},
		}},
		{"streaming-payload-result-with-explicit-view", testdata.StreamingPayloadResultWithExplicitViewDSL, []*sectionExpectation{
			{"client-stream-send", &testdata.StreamingPayloadResultWithExplicitViewClientStreamSendCode},
			{"client-stream-recv", &testdata.StreamingPayloadResultWithExplicitViewClientStreamRecvCode},
			{"client-stream-set-view", nil},
		}},
		{"streaming-payload-result-collection-with-views", testdata.StreamingPayloadResultCollectionWithViewsDSL, []*sectionExpectation{
			{"client-stream-send", &testdata.StreamingPayloadResultCollectionWithViewsClientStreamSendCode},
			{"client-stream-recv", &testdata.StreamingPayloadResultCollectionWithViewsClientStreamRecvCode},
			{"client-stream-set-view", &testdata.StreamingPayloadResultCollectionWithViewsClientStreamSetViewCode},
		}},
		{"streaming-payload-result-collection-with-explicit-view", testdata.StreamingPayloadResultCollectionWithExplicitViewDSL, []*sectionExpectation{
			{"client-stream-send", &testdata.StreamingPayloadResultCollectionWithExplicitViewClientStreamSendCode},
			{"client-stream-recv", &testdata.StreamingPayloadResultCollectionWithExplicitViewClientStreamRecvCode},
			{"client-stream-set-view", nil},
		}},
		{"streaming-payload-primitive", testdata.StreamingPayloadPrimitiveDSL, []*sectionExpectation{
			{"client-stream-send", &testdata.StreamingPayloadPrimitiveClientStreamSendCode},
			{"client-stream-recv", &testdata.StreamingPayloadPrimitiveClientStreamRecvCode},
			{"client-stream-set-view", nil},
		}},
		{"streaming-payload-primitive-array", testdata.StreamingPayloadPrimitiveArrayDSL, []*sectionExpectation{
			{"client-stream-send", &testdata.StreamingPayloadPrimitiveArrayClientStreamSendCode},
			{"client-stream-recv", &testdata.StreamingPayloadPrimitiveArrayClientStreamRecvCode},
		}},
		{"streaming-payload-primitive-map", testdata.StreamingPayloadPrimitiveMapDSL, []*sectionExpectation{
			{"client-stream-send", &testdata.StreamingPayloadPrimitiveMapClientStreamSendCode},
			{"client-stream-recv", &testdata.StreamingPayloadPrimitiveMapClientStreamRecvCode},
		}},
		{"streaming-payload-user-type-array", testdata.StreamingPayloadUserTypeArrayDSL, []*sectionExpectation{
			{"client-stream-send", &testdata.StreamingPayloadUserTypeArrayClientStreamSendCode},
			{"client-stream-recv", &testdata.StreamingPayloadUserTypeArrayClientStreamRecvCode},
		}},
		{"streaming-payload-user-type-map", testdata.StreamingPayloadUserTypeMapDSL, []*sectionExpectation{
			{"client-stream-send", &testdata.StreamingPayloadUserTypeMapClientStreamSendCode},
			{"client-stream-recv", &testdata.StreamingPayloadUserTypeMapClientStreamRecvCode},
		}},

		// bidirectional streaming

		{"bidirectional-streaming", testdata.BidirectionalStreamingDSL, []*sectionExpectation{
			{"client-endpoint-init", &testdata.BidirectionalStreamingClientEndpointCode},
			{"client-stream-send", &testdata.BidirectionalStreamingClientStreamSendCode},
			{"client-stream-recv", &testdata.BidirectionalStreamingClientStreamRecvCode},
			{"client-stream-close", &testdata.BidirectionalStreamingClientStreamCloseCode},
			{"client-stream-set-view", nil},
		}},
		{"bidirectional-streaming-no-payload", testdata.BidirectionalStreamingNoPayloadDSL, []*sectionExpectation{
			{"client-endpoint-init", &testdata.BidirectionalStreamingNoPayloadClientEndpointCode},
			{"client-stream-send", &testdata.BidirectionalStreamingNoPayloadClientStreamSendCode},
			{"client-stream-recv", &testdata.BidirectionalStreamingNoPayloadClientStreamRecvCode},
			{"client-stream-close", &testdata.BidirectionalStreamingNoPayloadClientStreamCloseCode},
		}},
		{"bidirectional-streaming-result-with-views", testdata.BidirectionalStreamingResultWithViewsDSL, []*sectionExpectation{
			{"client-stream-send", &testdata.BidirectionalStreamingResultWithViewsClientStreamSendCode},
			{"client-stream-recv", &testdata.BidirectionalStreamingResultWithViewsClientStreamRecvCode},
			{"client-stream-close", &testdata.BidirectionalStreamingResultWithViewsClientStreamCloseCode},
			{"client-stream-set-view", &testdata.BidirectionalStreamingResultWithViewsClientStreamSetViewCode},
		}},
		{"bidirectional-streaming-result-with-explicit-view", testdata.BidirectionalStreamingResultWithExplicitViewDSL, []*sectionExpectation{
			{"client-stream-send", &testdata.BidirectionalStreamingResultWithExplicitViewClientStreamSendCode},
			{"client-stream-recv", &testdata.BidirectionalStreamingResultWithExplicitViewClientStreamRecvCode},
			{"client-stream-set-view", nil},
		}},
		{"bidirectional-streaming-result-collection-with-views", testdata.BidirectionalStreamingResultCollectionWithViewsDSL, []*sectionExpectation{
			{"client-stream-send", &testdata.BidirectionalStreamingResultCollectionWithViewsClientStreamSendCode},
			{"client-stream-recv", &testdata.BidirectionalStreamingResultCollectionWithViewsClientStreamRecvCode},
			{"client-stream-set-view", &testdata.BidirectionalStreamingResultCollectionWithViewsClientStreamSetViewCode},
		}},
		{"bidirectional-streaming-result-collection-with-explicit-view", testdata.BidirectionalStreamingResultCollectionWithExplicitViewDSL, []*sectionExpectation{
			{"client-stream-send", &testdata.BidirectionalStreamingResultCollectionWithExplicitViewClientStreamSendCode},
			{"client-stream-recv", &testdata.BidirectionalStreamingResultCollectionWithExplicitViewClientStreamRecvCode},
			{"client-stream-set-view", nil},
		}},
		{"bidirectional-streaming-primitive", testdata.BidirectionalStreamingPrimitiveDSL, []*sectionExpectation{
			{"client-stream-send", &testdata.BidirectionalStreamingPrimitiveClientStreamSendCode},
			{"client-stream-recv", &testdata.BidirectionalStreamingPrimitiveClientStreamRecvCode},
			{"client-stream-set-view", nil},
		}},
		{"bidirectional-streaming-primitive-array", testdata.BidirectionalStreamingPrimitiveArrayDSL, []*sectionExpectation{
			{"client-stream-send", &testdata.BidirectionalStreamingPrimitiveArrayClientStreamSendCode},
			{"client-stream-recv", &testdata.BidirectionalStreamingPrimitiveArrayClientStreamRecvCode},
		}},
		{"bidirectional-streaming-primitive-map", testdata.BidirectionalStreamingPrimitiveMapDSL, []*sectionExpectation{
			{"client-stream-send", &testdata.BidirectionalStreamingPrimitiveMapClientStreamSendCode},
			{"client-stream-recv", &testdata.BidirectionalStreamingPrimitiveMapClientStreamRecvCode},
		}},
		{"bidirectional-streaming-user-type-array", testdata.BidirectionalStreamingUserTypeArrayDSL, []*sectionExpectation{
			{"client-stream-send", &testdata.BidirectionalStreamingUserTypeArrayClientStreamSendCode},
			{"client-stream-recv", &testdata.BidirectionalStreamingUserTypeArrayClientStreamRecvCode},
		}},
		{"bidirectional-streaming-user-type-map", testdata.BidirectionalStreamingUserTypeMapDSL, []*sectionExpectation{
			{"client-stream-send", &testdata.BidirectionalStreamingUserTypeMapClientStreamSendCode},
			{"client-stream-recv", &testdata.BidirectionalStreamingUserTypeMapClientStreamRecvCode},
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
				f := fs[0]
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
