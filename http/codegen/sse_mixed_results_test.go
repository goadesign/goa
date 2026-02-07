package codegen

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestSSE_MixedResults(t *testing.T) {
	root := RunHTTPDSL(t, testdata.MixedResultsDSL)
	services := CreateHTTPServices(root)

	t.Run("server", func(t *testing.T) {
		files := ServerFiles("", services)
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
		files := ClientFiles("", services)
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

