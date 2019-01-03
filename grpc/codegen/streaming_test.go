package codegen

import (
	"strings"
	"testing"

	"goa.design/goa/codegen"
	"goa.design/goa/expr"
	"goa.design/goa/grpc/codegen/testdata"
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
	//var empty string
	cases := []*testCase{
		// streaming result
		{"server-streaming", testdata.ServerStreamingUserTypeDSL, []*sectionExpectation{
			{"server-stream-struct-type", &testdata.ServerStreamingServerStructCode},
			{"server-stream-send", &testdata.ServerStreamingServerSendCode},
			{"server-stream-close", &testdata.ServerStreamingServerCloseCode},
			{"server-stream-set-view", nil},
			{"client-stream-struct-type", &testdata.ServerStreamingClientStructCode},
			{"client-stream-recv", &testdata.ServerStreamingClientRecvCode},
			{"client-stream-close", &testdata.ServerStreamingClientCloseCode},
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
			{"client-stream-recv", nil},
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
			RunGRPCDSL(t, c.DSL)
			serverfs := ServerFiles("", expr.Root)
			if len(serverfs) < 2 {
				t.Fatalf("got %d server files, expected 2", len(serverfs))
			}
			clientfs := ClientFiles("", expr.Root)
			if len(clientfs) < 2 {
				t.Fatalf("got %d client files, expected 2", len(clientfs))
			}
			for _, s := range c.Sections {
				var (
					code     string
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
				if seclen > 0 {
					code = codegen.SectionCode(t, sections[0])
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
					if code != *s.Code {
						t.Errorf("invalid code for %s %s section, got:\n%s\ngot vs. expected:\n%s", path, s.Name, code, codegen.Diff(t, code, *s.Code))
					}
				}
			}
		})
	}
}
