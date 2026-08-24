// This file checks that example generation uses only values copied before Go
// names are finalized.
package generator

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example/testdata"
	"goa.design/goa/v3/eval"
)

// TestExampleFilesDoNotReadChangedServerDesign checks that changing the API or
// server after planning cannot change any example file.
func TestExampleFilesDoNotReadChangedServerDesign(t *testing.T) {
	root := codegen.RunDSL(t, testdata.ServiceForOnlyHTTPDSL)
	plan := mustTestPlan(t, "generated.local/gen", []eval.Root{root}, planExampleData)

	beforeFiles, err := exampleFiles(plan)
	require.NoError(t, err)
	before := renderExampleFiles(t, beforeFiles)

	server := root.API.Servers[0]
	root.API.Name = "changed api"
	server.Name = "changed server"
	server.Description = "changed description"
	server.Services = nil
	server.Hosts = nil
	root.API.Servers = nil

	afterFiles, err := exampleFiles(plan)
	require.NoError(t, err)
	require.Equal(t, before, renderExampleFiles(t, afterFiles))
}

// renderExampleFiles writes each section and indexes the complete text by file
// path.
func renderExampleFiles(t *testing.T, files []*codegen.File) map[string]string {
	t.Helper()
	rendered := make(map[string]string, len(files))
	for _, file := range files {
		var output bytes.Buffer
		for _, section := range file.SectionTemplates {
			require.NoError(t, section.Write(&output))
		}
		rendered[file.Path] = output.String()
	}
	return rendered
}
