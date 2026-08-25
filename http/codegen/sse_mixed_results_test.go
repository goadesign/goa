package codegen

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestSSE_MixedResults(t *testing.T) {
	root := expr.RunDSL(t, testdata.MixedResultsDSL)
	plan := linkedHTTPPlanForRoot(t, root)

	t.Run("server", func(t *testing.T) {
		files := plan.ServerFiles()
		var sseFile *codegen.File
		for _, f := range files {
			if strings.HasSuffix(f.Path, filepath.Join("server", "sse.go")) {
				sseFile = f
				break
			}
		}
		require.NotNil(t, sseFile)

		sections := sseFile.Section("server-sse")
		require.NotEmpty(t, sections)
		code := codegen.SectionCode(t, sections[0])

		require.Contains(t, code, "body := NewEvent(res)")
		require.Contains(t, code, "json.Marshal(body)")
		require.NotContains(t, code, "var payload any")
		require.NotContains(t, code, "json.Marshal(res)")
	})

	t.Run("client", func(t *testing.T) {
		files := plan.ClientFiles()
		var sseFile *codegen.File
		for _, f := range files {
			if strings.HasSuffix(f.Path, filepath.Join("client", "sse.go")) {
				sseFile = f
				break
			}
		}
		require.NotNil(t, sseFile)

		sections := sseFile.Section("client-sse")
		require.NotEmpty(t, sections)
		code := codegen.SectionCode(t, sections[0])

		require.Contains(t, code, "var body Event")
		require.Contains(t, code, "err = ValidateEvent(&body)")
		require.Contains(t, code, "result := &mixedresultsservice.Event{")
		require.Contains(t, code, "return result, true, nil")
	})
}

// TestSSE_MixedResultConversionSelection verifies that mixed SSE clients use a
// direct assignment only for wire values that already have the service type.
// Anonymous objects need a planned conversion, while an empty streamed result
// returns the method's zero event without declaring an HTTP body.
func TestSSE_MixedResultConversionSelection(t *testing.T) {
	tests := []struct {
		name       string
		streaming  any
		contains   string
		notContain string
	}{
		{"primitive", dsl.Int, "result := body", "result := &"},
		{"primitive collection", dsl.ArrayOf(dsl.Int), "result := body", "result := &"},
		{"inline object", func() { dsl.Attribute("value", dsl.Int) }, "result := &", "result := body"},
		{"empty body", func() {}, "return event, true, nil", "var body"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := expr.RunDSL(t, mixedSSEResultShapeDSL(test.streaming))
			plan := linkedHTTPPlanForRoot(t, root)
			code := mixedSSEClientCode(t, plan)
			require.Contains(t, code, test.contains)
			require.NotContains(t, code, test.notContain)
		})
	}
}

// mixedSSEResultShapeDSL defines one ordinary string result and a separately
// streamed result so each test exercises the mixed SSE client path.
func mixedSSEResultShapeDSL(streaming any) func() {
	return func() {
		dsl.Service("Mixed Shape", func() {
			dsl.Method("watch", func() {
				dsl.Result(dsl.String)
				dsl.StreamingResult(streaming)
				dsl.HTTP(func() {
					dsl.GET("/watch")
					dsl.ServerSentEvents()
				})
			})
		})
	}
}

// mixedSSEClientCode renders the mixed endpoint's client stream implementation.
func mixedSSEClientCode(t *testing.T, plan *Plan) string {
	t.Helper()
	for _, file := range plan.ClientFiles() {
		if !strings.HasSuffix(file.Path, filepath.Join("client", "sse.go")) {
			continue
		}
		sections := file.Section("client-sse")
		require.NotEmpty(t, sections)
		return codegen.SectionCode(t, sections[0])
	}
	t.Fatal("mixed SSE client file was not generated")
	return ""
}
