package codegen

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestSSE(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"string", testdata.SSEStringDSL},
		{"int", testdata.SSEIntDSL},
		{"bool", testdata.SSEBoolDSL},
		{"object", testdata.SSEObjectDSL},
		{"data-field", testdata.SSEDataFieldDSL},
		{"data-id-field", testdata.SSEDataIDFieldDSL},
		{"request-id", testdata.SSERequestIDDSL},
		{"all-fields", testdata.SSEAllFieldsDSL},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := RunHTTPDSL(t, c.DSL)
			services := CreateHTTPServices(root)
			fs := ServerFiles("", services)
			// Simple types (string, int, bool) and request-id don't generate SSE-specific files
			// because they have no fields to map to SSE attributes
			expectedFiles := 2
			if c.Name == "object" || c.Name == "data-field" || c.Name == "data-id-field" || c.Name == "all-fields" {
				expectedFiles = 3
			}
			require.Len(t, fs, expectedFiles)
			// For cases with SSE files, check the SSE file (index 2)
			// For cases without SSE files, check the encode/decode file (index 1)
			fileIndex := 1
			if expectedFiles == 3 {
				fileIndex = 2
			}
			sections := fs[fileIndex].SectionTemplates
			require.Greater(t, len(sections), 1)
			code := codegen.SectionCode(t, sections[1])
			golden := filepath.Join("testdata", "golden", "sse-"+c.Name+".golden")
			compareOrUpdateGolden(t, code, golden)
		})
	}
}
