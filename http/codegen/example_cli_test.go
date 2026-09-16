// This file verifies generated HTTP command-line client examples.
package codegen

import (
	"bytes"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	ctestdata "goa.design/goa/v3/codegen/example/testdata"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestExampleCLIFiles(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"no-server", ctestdata.NoServerDSL},
		{"server-hosting-service-subset", ctestdata.ServerHostingServiceSubsetDSL},
		{"server-hosting-multiple-services", ctestdata.ServerHostingMultipleServicesDSL},
		{"streaming", testdata.StreamingResultDSL},
		{"streaming-multiple-services", testdata.StreamingMultipleServicesDSL},
		{"streaming-input-only", testdata.StreamingPayloadDSL},
		{"mixed-results", testdata.MixedResultsDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := codegen.RunDSL(t, c.DSL)
			examples := linkedHTTPExamplePlanForRoot(t, root)
			fs := examples.CLIFiles()
			require.Len(t, fs, 1)
			require.Greater(t, len(fs[0].SectionTemplates), 0)
			var buf bytes.Buffer
			for _, s := range fs[0].SectionTemplates[1:] {
				require.NoError(t, s.Write(&buf))
			}
			code := codegen.FormatTestCode(t, "package foo\n"+buf.String())
			golden := filepath.Join("testdata", "golden", "client-"+c.Name+".golden")
			testutil.CompareOrUpdateGolden(t, code, golden)
		})
	}
}

func TestExampleCLIFilesOmitFilesOnlyServer(t *testing.T) {
	root := codegen.RunDSL(t, ctestdata.ServerHostingServiceWithFileServerDSL)
	examples := linkedHTTPExamplePlanForRoot(t, root)

	require.NotEmpty(t, examples.ServerFiles())
	require.Empty(t, examples.CLIFiles())
}

func TestExampleCLIUsesServicePathsForCommands(t *testing.T) {
	root := codegen.RunDSL(t, collidingServiceNamesDSL)
	examples := linkedHTTPExamplePlanForRoot(t, root)
	files := examples.CLIFiles()
	require.Len(t, files, 1)

	var output bytes.Buffer
	for _, section := range files[0].SectionTemplates {
		require.NoError(t, section.Write(&output))
	}
	first := examples.transport.services.Get("read_value").Service
	second := examples.transport.services.Get("read-value").Service
	firstCommand := codegen.KebabCase(first.PathName)
	secondCommand := codegen.KebabCase(second.PathName)
	require.NotEqual(t, firstCommand, secondCommand)
	require.Contains(t, output.String(), `case "`+firstCommand+`":`)
	require.Contains(t, output.String(), `case "`+secondCommand+`":`)
}

// TestExampleCLIImportsOnlyConfiguredServices verifies that an example client
// does not reserve packages for services its server does not expose or for
// streaming code it does not write.
func TestExampleCLIImportsOnlyConfiguredServices(t *testing.T) {
	root := codegen.RunDSL(t, ctestdata.ServerHostingServiceSubsetDSL)
	examples := linkedHTTPExamplePlanForRoot(t, root)
	files := examples.CLIFiles()
	require.Len(t, files, 1)

	imports := importPaths(files[0].SectionTemplates[0].Data.(map[string]any)["Imports"].([]*codegen.ImportSpec))
	ignored := examples.transport.services.Get("IgnoredService").Service
	require.NotContains(t, imports, path.Join(examples.transport.services.GenPkg(), ignored.PathName))
	require.NotContains(t, imports, "strings")
	require.NotContains(t, imports, "github.com/gorilla/websocket")
}
