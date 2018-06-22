package codegen

import (
	"testing"

	"goa.design/goa/codegen"
	"goa.design/goa/http/codegen/testdata"
	httpdesign "goa.design/goa/http/design"
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
		{"streaming-result", testdata.StreamingResultDSL, []*sectionExpectation{
			{"server-handler-init", &testdata.StreamingResultServerHandlerInitCode},
			{"server-stream-send", &testdata.StreamingResultServerStreamSendCode},
			{"server-stream-close", &testdata.StreamingResultServerStreamCloseCode},
			{"server-stream-set-view", nil},
		}},
		{"streaming-result-with-views", testdata.StreamingResultWithViewsDSL, []*sectionExpectation{
			{"server-handler-init", &testdata.StreamingResultWithViewsServerHandlerInitCode},
			{"server-stream-send", &testdata.StreamingResultWithViewsServerStreamSendCode},
			{"server-stream-close", &testdata.StreamingResultWithViewsServerStreamCloseCode},
			{"server-stream-set-view", &testdata.StreamingResultWithViewsServerStreamSetViewCode},
		}},
		{"streaming-result-no-payload", testdata.StreamingResultNoPayloadDSL, []*sectionExpectation{
			{"server-handler-init", &testdata.StreamingResultNoPayloadServerHandlerInitCode},
		}},
	}
	filesFn := func() []*codegen.File { return ServerFiles("", httpdesign.Root) }
	runTests(t, cases, filesFn)
}

func TestClientStreaming(t *testing.T) {
	cases := []*testCase{
		{"streaming-result", testdata.StreamingResultDSL, []*sectionExpectation{
			{"client-endpoint-init", &testdata.StreamingResultClientEndpointCode},
			{"client-stream-recv", &testdata.StreamingResultClientStreamRecvCode},
			{"client-stream-set-view", nil},
		}},
		{"streaming-result-with-views", testdata.StreamingResultWithViewsDSL, []*sectionExpectation{
			{"client-endpoint-init", &testdata.StreamingResultWithViewsClientEndpointCode},
			{"client-stream-recv", &testdata.StreamingResultWithViewsClientStreamRecvCode},
			{"client-stream-set-view", &testdata.StreamingResultWithViewsClientStreamSetViewCode},
		}},
	}
	filesFn := func() []*codegen.File { return ClientFiles("", httpdesign.Root) }
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
