// This file checks the generated JSON-RPC code that carries a selected result
// view through unary responses and server-sent events.
package codegen

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/dsl"
)

func TestVariableViewedResultGeneratedSource(t *testing.T) {
	_, plan := linkedJSONRPCPlan(t, variableViewedResultGoldenDSL)
	require.Len(t, plan.services, 1)
	service := plan.services[0]

	clientConversions := clientViewedResultSections(service)
	serverConversions := serverViewedResultSections(service)
	require.Len(t, clientConversions, 2)
	require.Len(t, serverConversions, 2)
	clientData := clientConversions[0].Data.(*viewedResultTemplateData)
	require.Equal(t, "fetch", clientData.MethodName)
	require.Equal(t, "viewed2", clientData.ViewedValue)
	require.Equal(t, "fetch", serverConversions[0].Data.(*viewedResultTemplateData).MethodName)
	testutil.AssertGo(
		t,
		"testdata/golden/viewed_result_variable_decoder.go.golden",
		codegen.SectionCode(t, clientConversions[0]),
	)
	testutil.AssertGo(
		t,
		"testdata/golden/viewed_result_variable_encoder.go.golden",
		codegen.SectionCode(t, serverConversions[0]),
	)

	tests := []struct {
		name         string
		files        []*codegen.File
		packageName  string
		fileName     string
		sectionName  string
		sectionCount int
		golden       string
	}{
		{
			name:         "unary client",
			files:        plan.ClientFiles(),
			packageName:  "client",
			fileName:     "encode_decode.go",
			sectionName:  "jsonrpc-response-decoder",
			sectionCount: 1,
			golden:       "testdata/golden/viewed_result_variable_unary_client.go.golden",
		},
		{
			name:         "unary server",
			files:        plan.ServerFiles(),
			packageName:  "server",
			fileName:     "server.go",
			sectionName:  "jsonrpc-server-handler-init",
			sectionCount: 2,
			golden:       "testdata/golden/viewed_result_variable_unary_server.go.golden",
		},
		{
			name:         "SSE client",
			files:        plan.ClientFiles(),
			packageName:  "client",
			fileName:     "stream.go",
			sectionName:  "jsonrpc-sse-client-stream",
			sectionCount: 1,
			golden:       "testdata/golden/viewed_result_variable_sse_client.go.golden",
		},
		{
			name:         "SSE server",
			files:        plan.ServerFiles(),
			packageName:  "server",
			fileName:     "sse.go",
			sectionName:  "jsonrpc-sse-server-stream",
			sectionCount: 1,
			golden:       "testdata/golden/viewed_result_variable_sse_server.go.golden",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			file := viewedResultGoldenFile(t, test.files, test.packageName, test.fileName)
			sections := file.Section(test.sectionName)
			require.Len(t, sections, test.sectionCount)
			testutil.AssertGo(t, test.golden, codegen.SectionCode(t, sections[0]))
		})
	}
}

// viewedResultGoldenFile returns one generated client or server file for the
// viewed service used by the snapshots.
func viewedResultGoldenFile(t *testing.T, files []*codegen.File, packageName, fileName string) *codegen.File {
	t.Helper()
	for _, file := range files {
		if filepath.Base(file.Path) != fileName {
			continue
		}
		if filepath.Base(filepath.Dir(file.Path)) != packageName {
			continue
		}
		if filepath.Base(filepath.Dir(filepath.Dir(file.Path))) == "viewed" {
			return file
		}
	}
	t.Fatalf("generated viewed/%s/%s file not found", packageName, fileName)
	return nil
}

// variableViewedResultGoldenDSL defines one unary method and one server stream
// whose callers choose between the two named views or the generated default.
func variableViewedResultGoldenDSL() {
	result := dsl.ResultType("application/vnd.viewed-golden", func() {
		dsl.TypeName("ViewedGolden")
		dsl.Attribute("id", dsl.String)
		dsl.Attribute("detail", dsl.String)
		dsl.Required("id", "detail")
		dsl.View("summary", func() {
			dsl.Attribute("id")
		})
		dsl.View("detailed", func() {
			dsl.Attribute("id")
			dsl.Attribute("detail")
		})
	})
	dsl.Service("viewed", func() {
		dsl.JSONRPC(func() {
			dsl.POST("/rpc")
		})
		dsl.Method("fetch", func() {
			dsl.Result(result)
			dsl.JSONRPC(func() {})
		})
		dsl.Method("watch", func() {
			dsl.StreamingResult(result)
			dsl.JSONRPC(func() {
				dsl.ServerSentEvents(func() {})
			})
		})
	})
}
