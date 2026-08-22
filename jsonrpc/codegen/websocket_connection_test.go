// This file verifies that every generated JSON-RPC method uses the same
// WebSocket, while one reader matches each response to the request with that ID.
package codegen_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	goacodegen "goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/codegen/service"
	. "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
	jsonrpccodegen "goa.design/goa/v3/jsonrpc/codegen"
)

// websocketConnectionDSL defines one bidirectional WebSocket method so these
// tests do not depend on unrelated HTTP, SSE, or unary JSON-RPC generation.
var websocketConnectionDSL = func() {
	API("websocket-connection", func() {
		JSONRPC(func() {})
	})
	Service("Chat", func() {
		JSONRPC(func() {
			Path("/ws")
		})
		Method("echo", func() {
			StreamingPayload(func() {
				ID("id", String, "Request ID")
				Attribute("msg", String)
			})
			StreamingResult(func() {
				ID("id", String, "Request ID")
				Attribute("echo", String)
			})
			JSONRPC(func() {})
		})
	})
}

// TestGeneratedWebSocketSharesOneConnection proves that one generated client
// reads and writes its shared WebSocket in one place. Each method stream keeps
// the requests that its Recv calls will return.
func TestGeneratedWebSocketSharesOneConnection(t *testing.T) {
	root := expr.RunDSL(t, websocketConnectionDSL)
	generation, err := goacodegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	plan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	httpPlans, err := httpcodegen.NewJSONRPCPlans(generation, httpcodegen.PlanInput{Root: root, Service: plan})
	require.NoError(t, err)
	jsonPlans, err := jsonrpccodegen.NewPlans(generation, jsonrpccodegen.PlanInput{
		Root:    root,
		Service: plan,
		HTTP:    httpPlans[0],
	})
	require.NoError(t, err)
	require.NoError(t, example.Plan(generation))
	require.NoError(t, generation.Freeze())
	require.NoError(t, plan.Link())
	require.NoError(t, httpPlans[0].Link())
	require.NoError(t, jsonPlans[0].Link())

	dir := t.TempDir()
	for _, file := range append(jsonPlans[0].ClientFiles(), jsonPlans[0].ServerFiles()...) {
		_, err := file.Render(dir)
		require.NoError(t, err)
	}

	client := readGeneratedFile(t, dir, "gen/jsonrpc/chat/client/client.go")
	clientStream := readGeneratedFile(t, dir, "gen/jsonrpc/chat/client/websocket.go")
	serverStream := readGeneratedFile(t, dir, "gen/jsonrpc/chat/server/websocket.go")

	require.Contains(t, client, "websocketClientConn struct")
	require.Contains(t, client, "atomic.Uint64")
	require.Contains(t, client, "pending map[string]*websocketPendingRequest")
	require.Contains(t, client, "go conn.readResponses()")
	require.Contains(t, client, "case owner.closed.Load():")
	require.Contains(t, client, "time.AfterFunc(c.config.RequestTimeout")
	require.Contains(t, client, "request.complete(request.ctx, nil, err)")
	require.Contains(t, client, "func (c *websocketClientConn) closeSocket() error")
	require.Contains(t, client, "var errWebsocketMethodStreamClosed = errors.New(")
	require.NotContains(t, client, "websocket.PingMessage")
	require.Equal(t, 1, strings.Count(client, "ReadJSON("))
	require.NotContains(t, clientStream, "ReadJSON(")
	require.NotContains(t, clientStream, "idGenerator")
	require.NotContains(t, clientStream, "writeMu")
	require.Contains(t, clientStream, "*websocketClientConn")
	require.Contains(t, clientStream, "s.conn.closeOwner(s.owner")
	require.Contains(t, clientStream, "return errWebsocketMethodStreamClosed")
	require.NotContains(t, clientStream, "EchoClientStreamPendingRequest")
	require.NotContains(t, clientStream, "EchoClientStreamStreamResult")
	require.Contains(t, clientStream, "echoClientStreamPendingRequest")
	require.Contains(t, clientStream, "echoClientStreamStreamResult")

	require.Contains(t, serverStream, "writeMu sync.Mutex")
	require.Contains(t, serverStream, "func (s *chatStream) writeJSON(")
	require.Equal(t, 1, strings.Count(serverStream, "s.conn.WriteJSON("))
	require.Contains(t, serverStream, "websocket.FormatCloseMessage(websocket.CloseNormalClosure")
	require.Contains(t, serverStream, "time.Now().Add(time.Second)")
	require.Contains(t, serverStream, "closeErr := s.conn.Close()")
	require.Contains(t, serverStream, "return errors.Join(controlErr, closeErr)")
	require.NotContains(t, serverStream, "func (s *chatStream) Close() error {\n\ts.writeMu.Lock()")
}

// readGeneratedFile returns generated source inspected by the checks above.
func readGeneratedFile(t *testing.T, dir, path string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(dir, path))
	require.NoError(t, err)
	return string(content)
}
