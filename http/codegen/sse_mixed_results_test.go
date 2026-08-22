package codegen

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
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

		require.Contains(t, code, "payload = res")
		require.NotContains(t, code, "NewCreateResponseBody")
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

		require.Contains(t, code, "event = new(")
	})
}
