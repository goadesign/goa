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

// TestServerErrorResponses verifies the transport writes request and service
// failures without exposing JSON-RPC error methods through the service stream.
func TestServerErrorResponses(t *testing.T) {
	root := expr.RunDSL(t, testdata.JSONRPCKitchenSinkDSL)
	plan := CreateJSONRPCPlan(root)

	feedServer := renderPlannedFile(t, plan.ServerFiles(), "feed", "server.go")
	require.Contains(t, feedServer, "return strm.sendError(ctx, req.ID, jsonrpc.InvalidParams, err.Error(), nil)")
	require.Contains(t, feedServer, "response := jsonrpc.MakeSuccessResponse(req.ID, nil)")
	require.Contains(t, feedServer, "return strm.sendSSEEvent(ctx, response, nil, nil, nil)")
	require.Contains(t, feedServer, "jsonrpc.MakeSuccessResponse(req.ID, res)")
	require.Equal(t, 1, strings.Count(feedServer, `mux.Handle("POST", "/feed", h.ServeHTTP)`))
	require.NotContains(t, feedServer, "SendError")

	feedStream := renderPlannedFile(t, plan.ServerFiles(), "feed", "sse.go")
	require.Contains(t, feedStream, "func (s *WatchServerStream) Send(event *feed.WatchResult) error")
	require.Contains(t, feedStream, "func (s *WatchServerStream) SendWithContext(ctx context.Context, event *feed.WatchResult) error")
	require.Contains(t, feedStream, "func (s *WatchServerStream) Close() error")
	require.NotContains(t, feedStream, "SendAndClose")
	require.NotContains(t, feedStream, "SendError")

	calcServer := renderPlannedFile(t, plan.ServerFiles(), "calc", "server.go")
	require.Contains(t, calcServer, "if err != nil {")
	require.Contains(t, calcServer, "encodeJSONRPCError(ctx, w, req,")
}

// TestNamedSSEPayloadReceivesLastEventID verifies a named payload receives the
// event ID in its designed pointer field before the endpoint runs.
func TestNamedSSEPayloadReceivesLastEventID(t *testing.T) {
	root := expr.RunDSL(t, testdata.JSONRPCKitchenSinkDSL)
	plan := CreateJSONRPCPlan(root)
	decoder := paramsGoldenSection(t, plan.ServerFiles(), "encode_decode.go", "jsonrpc-request-decoder", "func DecodeWatchRequest")
	feedServer := codegen.SectionCode(t, decoder)

	require.Contains(t, feedServer, `lastEventIDValues, lastEventIDPresent := r.Header["Last-Event-Id"]`)
	require.Contains(t, feedServer, "payload = NewWatchPayload(lastEventID, requestID)")
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
