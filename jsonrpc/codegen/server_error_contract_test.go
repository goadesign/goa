// This file verifies how generated JSON-RPC servers report request, service,
// and stream errors to callers.
package codegen

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/jsonrpc/codegen/testdata"
)

// TestServerErrorResponses verifies request decoding writes JSON-RPC errors,
// service code can explicitly write a server-sent event error, service method
// failures return to the server, and unary failures still become responses.
func TestServerErrorResponses(t *testing.T) {
	root := expr.RunDSL(t, testdata.JSONRPCKitchenSinkDSL)
	plan := CreateJSONRPCPlan(root)

	feedServer := renderPlannedFile(t, plan.ServerFiles(), "feed", "server.go")
	require.Contains(t, feedServer, "if err := strm.SendError(ctx, jsonrpc.IDToString(req.ID), err); err != nil {\n\t\t\t\t\treturn err\n\t\t\t\t}")
	require.Contains(t, feedServer, "if _, err := endpoint(ctx, v); err != nil {\n\t\t\treturn err")
	require.NotContains(t, feedServer, "return strm.SendError")

	feedStream := renderPlannedFile(t, plan.ServerFiles(), "feed", "sse.go")
	require.Contains(t, feedStream, "func (s *WatchServerStream) SendError(")

	calcServer := renderPlannedFile(t, plan.ServerFiles(), "calc", "server.go")
	require.Contains(t, calcServer, "if err != nil {")
	require.Contains(t, calcServer, "encodeJSONRPCError(ctx, w, req,")
}

// renderPlannedFile renders one file stored by the plan into memory without
// writing generated output to the repository.
func renderPlannedFile(t *testing.T, files []*codegen.File, service, name string) string {
	t.Helper()
	for _, file := range files {
		if filepath.Base(file.Path) != name || filepath.Base(filepath.Dir(filepath.Dir(file.Path))) != service {
			continue
		}
		var source strings.Builder
		for _, section := range file.SectionTemplates {
			require.NoError(t, section.Write(&source))
		}
		return source.String()
	}
	t.Errorf("generated %s/%s file not found", service, name)
	return ""
}
