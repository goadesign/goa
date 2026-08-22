// This file checks that generated clients and servers send each result view
// with the JSON fields selected by that view.
package generator

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

const jsonRPCViewedWebSocketServerTest = `package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"

	service "generated.local/gen/jsonrpc_web_socket"
	goahttp "goa.design/goa/v3/http"
	goa "goa.design/goa/v3/pkg"
)

type viewedService struct {
	releaseWatch chan struct{}
	sendErrors   chan error
}

func (*viewedService) HandleStream(context.Context, service.Stream) error {
	return nil
}

func (s *viewedService) Watch(ctx context.Context, stream service.WatchServerStream) error {
	stream.SetView("summary")
	go func() {
		<-s.releaseWatch
		s.sendErrors <- stream.SendResponse(ctx, viewedEvent("watch-event"))
	}()
	return nil
}

func (s *viewedService) Inspect(ctx context.Context, stream service.InspectServerStream) error {
	stream.SetView("detailed")
	s.sendErrors <- stream.SendResponse(ctx, viewedEvent("inspect-event"))
	close(s.releaseWatch)
	return nil
}

func (*viewedService) Fixed(context.Context, service.FixedServerStream) error {
	return nil
}

type wireMessage struct {
	ID     any             ` + "`" + `json:"id"` + "`" + `
	Method string          ` + "`" + `json:"method"` + "`" + `
	Params json.RawMessage ` + "`" + `json:"params"` + "`" + `
	Result json.RawMessage ` + "`" + `json:"result"` + "`" + `
}

func TestViewedWebSocketServerUsesPerMessageView(t *testing.T) {
	svc := &viewedService{
		releaseWatch: make(chan struct{}),
		sendErrors:   make(chan error, 2),
	}
	directError := make(chan error, 1)
	acknowledged := make(chan struct{})
	var acknowledge sync.Once
	releaseServer := func() {
		acknowledge.Do(func() { close(acknowledged) })
	}
	handler := func(ctx context.Context, stream service.Stream) error {
		if err := stream.SendWatchNotification(ctx, viewedEvent("direct-watch"), "summary"); err != nil {
			return err
		}
		if err := stream.SendInspectNotification(ctx, viewedEvent("direct-inspect"), "detailed"); err != nil {
			return err
		}
		if err := stream.SendFixedNotification(ctx, viewedEvent("direct-fixed")); err != nil {
			return err
		}
		directError <- stream.SendWatchNotification(ctx, viewedEvent("invalid"), "unknown")
		for range 2 {
			if err := stream.Recv(ctx); err != nil {
				return err
			}
		}
		<-acknowledged
		return nil
	}
	server := New(
		handler,
		service.NewEndpoints(svc),
		goahttp.NewMuxer(),
		goahttp.RequestDecoder,
		goahttp.ResponseEncoder,
		func(_ context.Context, _ http.ResponseWriter, err error) { t.Errorf("serve WebSocket: %v", err) },
		&websocket.Upgrader{},
		nil,
	)
	httpServer := httptest.NewServer(server)
	defer httpServer.Close()
	conn, _, err := websocket.DefaultDialer.Dial(
		"ws"+strings.TrimPrefix(httpServer.URL, "http")+"/stream",
		nil,
	)
	require.NoError(t, err)
	defer func() { require.NoError(t, conn.Close()) }()
	defer releaseServer()
	require.NoError(t, conn.SetReadDeadline(time.Now().Add(5*time.Second)))

	directWatch := readWireMessage(t, conn)
	require.Equal(t, "watch", directWatch.Method)
	require.JSONEq(t,
		` + "`" + `{"view":"summary","body":{"event_id":"direct-watch"}}` + "`" + `,
		string(directWatch.Params),
	)
	directInspect := readWireMessage(t, conn)
	require.Equal(t, "inspect", directInspect.Method)
	require.JSONEq(t,
		` + "`" + `{"view":"detailed","body":{"event_id":"direct-inspect","profile":{"display_name":"Ada"}}}` + "`" + `,
		string(directInspect.Params),
	)
	directFixed := readWireMessage(t, conn)
	require.Equal(t, "fixed", directFixed.Method)
	require.JSONEq(t,
		` + "`" + `{"event_id":"direct-fixed","profile":{"display_name":"Ada"}}` + "`" + `,
		string(directFixed.Params),
	)
	requireBoundaryError(t, <-directError, goa.InvalidEnumValue, "view")

	require.NoError(t, conn.WriteJSON(map[string]any{
		"jsonrpc": "2.0", "id": "watch-id", "method": "watch", "params": map[string]any{"key": "watch"},
	}))
	require.NoError(t, conn.WriteJSON(map[string]any{
		"jsonrpc": "2.0", "id": "inspect-id", "method": "inspect", "params": map[string]any{"key": "inspect"},
	}))
	inspect := readWireMessage(t, conn)
	require.Equal(t, "inspect-id", inspect.ID)
	require.JSONEq(t,
		` + "`" + `{"view":"detailed","body":{"event_id":"inspect-event","profile":{"display_name":"Ada"}}}` + "`" + `,
		string(inspect.Result),
	)
	watch := readWireMessage(t, conn)
	require.Equal(t, "watch-id", watch.ID)
	require.JSONEq(t,
		` + "`" + `{"view":"summary","body":{"event_id":"watch-event"}}` + "`" + `,
		string(watch.Result),
	)
	for range 2 {
		require.NoError(t, <-svc.sendErrors)
	}
	releaseServer()
}

func readWireMessage(t *testing.T, conn *websocket.Conn) wireMessage {
	t.Helper()
	var message wireMessage
	require.NoError(t, conn.ReadJSON(&message))
	return message
}

func viewedEvent(id string) *service.Event {
	return &service.Event{EventID: id, Profile: &service.Profile{DisplayName: "Ada"}}
}

func requireBoundaryError(t *testing.T, err error, name, field string) {
	t.Helper()
	var serviceError *goa.ServiceError
	require.ErrorAs(t, err, &serviceError)
	require.Equal(t, name, serviceError.Name)
	require.NotNil(t, serviceError.Field)
	require.Equal(t, field, *serviceError.Field)
}
`

// TestGeneratedHTTPViewedSSEServerUsesRequestView checks that an SSE request
// uses the view selected by the service call. A method with one fixed view does
// not choose a view while it runs.
func TestGeneratedHTTPViewedSSEServerUsesRequestView(t *testing.T) {
	dir := generateViewedTransportModule(t, viewedHTTPSSEDSL)
	writeGeneratedContractTest(t, dir, filepath.Join("http", "http_view_stream", "server"), httpViewedSSEServerTest)
	runGeneratedPackageTests(t, dir, "./http/http_view_stream/server")
}

// TestGeneratedHTTPViewedSSEClientRebuildsResult checks that an SSE client
// reads the selected HTTP body before rebuilding the service result, including
// JSON field names and nested fields.
func TestGeneratedHTTPViewedSSEClientRebuildsResult(t *testing.T) {
	dir := generateViewedTransportModule(t, viewedHTTPSSEDSL)
	writeGeneratedContractTest(t, dir, filepath.Join("http", "http_view_stream", "client"), httpViewedSSEClientTest)
	runGeneratedPackageTests(t, dir, "./http/http_view_stream/client")
}

// TestGeneratedJSONRPCUnaryViewedRepresentation checks that a one-result call
// sends both the selected view name and its JSON body. The view name must not
// come from an HTTP header.
func TestGeneratedJSONRPCUnaryViewedRepresentation(t *testing.T) {
	dir := generateViewedTransportModule(t, viewedJSONRPCUnaryDSL)
	writeGeneratedContractTest(t, dir, filepath.Join("jsonrpc", "jsonrpc_unary", "client"), jsonRPCViewedUnaryClientTest)
	runGeneratedPackageTests(t, dir, "./jsonrpc/jsonrpc_unary/client")
}

// TestGeneratedJSONRPCUnaryServerEmitsViewedRepresentation checks that the
// server writes the selected view and matching body in the JSON-RPC result. A
// method with one fixed view writes only the body.
func TestGeneratedJSONRPCUnaryServerEmitsViewedRepresentation(t *testing.T) {
	dir := generateViewedTransportModule(t, viewedJSONRPCUnaryDSL)
	writeGeneratedContractTest(t, dir, filepath.Join("jsonrpc", "jsonrpc_unary", "server"), jsonRPCViewedUnaryServerTest)
	runGeneratedPackageTests(t, dir, "./jsonrpc/jsonrpc_unary/server")
}

// TestGeneratedJSONRPCSSEViewedRepresentation checks that JSON-RPC SSE pairs
// every view name with its matching body and rebuilds service results for both
// notifications and final responses.
func TestGeneratedJSONRPCSSEViewedRepresentation(t *testing.T) {
	dir := generateViewedTransportModule(t, viewedJSONRPCSSEDSL)
	writeGeneratedContractTest(t, dir, filepath.Join("jsonrpc", "jsonrpcsse", "client"), jsonRPCViewedSSEClientTest)
	runGeneratedPackageTests(t, dir, "./jsonrpc/jsonrpcsse/client")
}

// TestGeneratedJSONRPCSSEServerEmitsViewedRepresentation checks that SSE
// notifications and final responses contain the same view name and body that
// clients read. Methods with one fixed view contain only the body.
func TestGeneratedJSONRPCSSEServerEmitsViewedRepresentation(t *testing.T) {
	dir := generateViewedTransportModule(t, viewedJSONRPCSSEDSL)
	writeGeneratedContractTest(t, dir, filepath.Join("jsonrpc", "jsonrpcsse", "server"), jsonRPCViewedSSEServerTest)
	runGeneratedPackageTests(t, dir, "./jsonrpc/jsonrpcsse/server")
}

// TestGeneratedJSONRPCWebSocketDirectSendsRequireView checks that each direct
// send chooses a view when several are legal. A method with one fixed view does
// not accept a view argument, and only methods that need a choice have SetView.
func TestGeneratedJSONRPCWebSocketDirectSendsRequireView(t *testing.T) {
	dir := generateViewedTransportModule(t, viewedJSONRPCWebSocketDSL)
	writeGeneratedContractTest(t, dir, "jsonrpc_web_socket", jsonRPCViewedWebSocketInterfaceTest)
	runGeneratedPackageTests(t, dir, "./jsonrpc_web_socket")
}

// TestGeneratedJSONRPCWebSocketRoutesResponsesByRequestID checks that two
// methods can share one connection, write at the same time, and receive
// responses in reverse order without either method receiving the wrong result.
func TestGeneratedJSONRPCWebSocketRoutesResponsesByRequestID(t *testing.T) {
	dir := generateViewedTransportModule(t, viewedJSONRPCWebSocketDSL)
	writeGeneratedContractTest(t, dir, filepath.Join("jsonrpc", "jsonrpc_web_socket", "client"), jsonRPCViewedWebSocketRuntimeTest)
	runGeneratedPackageTests(t, dir, "./jsonrpc/jsonrpc_web_socket/client")
}

// TestGeneratedJSONRPCWebSocketServerUsesPerCallViews checks that each request
// and direct send writes the body for its own view, even when two methods share
// one connection and finish out of order.
func TestGeneratedJSONRPCWebSocketServerUsesPerCallViews(t *testing.T) {
	dir := generateViewedTransportModule(t, viewedJSONRPCWebSocketDSL)
	writeGeneratedContractTest(t, dir, filepath.Join("jsonrpc", "jsonrpc_web_socket", "server"), jsonRPCViewedWebSocketServerTest)
	runGeneratedPackageTests(t, dir, "./jsonrpc/jsonrpc_web_socket/server")
}

// generateViewedTransportModule generates a temporary Go module for one test.
func generateViewedTransportModule(t *testing.T, design func()) string {
	t.Helper()
	registry := testRegistryFromGenfuncs([]testGenfunc{
		{Plan: planServiceData, Generate: testServiceFiles},
		{Plan: planTransportData, Generate: testTransportFiles},
	})
	codegen.RunDSL(t, design)
	dir := filepath.Join(t.TempDir(), codegen.Gendir)
	writeGeneratedModule(t, dir, "generated.local/gen")
	_, err := generate(filepath.Dir(dir), "gen", false, registry)
	require.NoError(t, err)
	return dir
}

// writeGeneratedContractTest adds a test that calls one generated package.
// The generated module is temporary; the source tree remains untouched.
func writeGeneratedContractTest(t *testing.T, moduleDir, packageDir, source string) {
	t.Helper()
	dir := filepath.Join(moduleDir, packageDir)
	require.NoError(t, os.MkdirAll(dir, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "viewed_contract_test.go"), []byte(source), 0o600))
}

// runGeneratedPackageTests compiles and runs one generated package. Limiting
// the command to that package makes a failure point to the code under test.
func runGeneratedPackageTests(t *testing.T, dir, packagePattern string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "test", "-mod=mod", packagePattern)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("test generated package %s: %v\n%s", packagePattern, err, output)
	}
}

// viewedResultType defines a result whose selected view changes both the JSON
// body fields and their JSON names.
func viewedResultType() *expr.ResultTypeExpr {
	profile := dsl.Type("Profile", func() {
		dsl.Attribute("display_name", dsl.String)
		dsl.Required("display_name")
	})
	return dsl.ResultType("application/vnd.viewed-event", func() {
		dsl.TypeName("Event")
		dsl.Attribute("event_id", dsl.String)
		dsl.Attribute("profile", profile)
		dsl.Required("event_id", "profile")
		dsl.View("summary", func() {
			dsl.Attribute("event_id")
		})
		dsl.View("detailed", func() {
			dsl.Attribute("event_id")
			dsl.Attribute("profile")
		})
	})
}

// viewedHTTPSSEDSL creates HTTP SSE methods with selectable and fixed views.
func viewedHTTPSSEDSL() {
	event := viewedResultType()
	immediate := dsl.Type("Immediate", func() {
		dsl.Attribute("message", dsl.String)
	})
	dsl.Service("HTTP View Stream", func() {
		dsl.Method("watch", func() {
			dsl.StreamingResult(event)
			dsl.HTTP(func() {
				dsl.GET("/watch")
				dsl.ServerSentEvents()
			})
		})
		dsl.Method("fixed", func() {
			dsl.StreamingResult(event, func() {
				dsl.View("detailed")
			})
			dsl.HTTP(func() {
				dsl.GET("/fixed")
				dsl.ServerSentEvents()
			})
		})
		dsl.Method("mixed", func() {
			dsl.Result(immediate)
			dsl.StreamingResult(event)
			dsl.HTTP(func() {
				dsl.GET("/mixed")
				dsl.ServerSentEvents()
			})
		})
	})
}

// viewedJSONRPCUnaryDSL creates one-result JSON-RPC methods with selectable
// and fixed views.
func viewedJSONRPCUnaryDSL() {
	event := viewedResultType()
	dsl.Service("JSON RPC Unary", func() {
		dsl.JSONRPC(func() {
			dsl.POST("/rpc")
		})
		dsl.Method("fetch", func() {
			dsl.Result(event)
			dsl.JSONRPC(func() {})
		})
		dsl.Method("fixed", func() {
			dsl.Result(event, func() {
				dsl.View("detailed")
			})
			dsl.JSONRPC(func() {})
		})
	})
}

// viewedJSONRPCSSEDSL creates JSON-RPC SSE methods with selectable and fixed
// views.
func viewedJSONRPCSSEDSL() {
	event := viewedResultType()
	dsl.Service("JSON RPC SSE", func() {
		dsl.JSONRPC(func() {
			dsl.POST("/events")
		})
		dsl.Method("watch", func() {
			dsl.StreamingResult(event)
			dsl.JSONRPC(func() {
				dsl.ServerSentEvents(func() {})
			})
		})
		dsl.Method("fixed", func() {
			dsl.StreamingResult(event, func() {
				dsl.View("detailed")
			})
			dsl.JSONRPC(func() {
				dsl.ServerSentEvents(func() {})
			})
		})
	})
}

// viewedJSONRPCWebSocketDSL creates JSON-RPC WebSocket methods with selectable
// and fixed views on one service.
func viewedJSONRPCWebSocketDSL() {
	event := viewedResultType()
	dsl.Service("JSON RPC WebSocket", func() {
		dsl.JSONRPC(func() {
			dsl.Path("/stream")
		})
		dsl.Method("watch", func() {
			dsl.StreamingPayload(func() {
				dsl.Attribute("key", dsl.String)
				dsl.Required("key")
			})
			dsl.StreamingResult(event)
			dsl.JSONRPC(func() {})
		})
		dsl.Method("inspect", func() {
			dsl.StreamingPayload(func() {
				dsl.Attribute("key", dsl.String)
				dsl.Required("key")
			})
			dsl.StreamingResult(event)
			dsl.JSONRPC(func() {})
		})
		dsl.Method("fixed", func() {
			dsl.StreamingPayload(func() {
				dsl.Attribute("key", dsl.String)
				dsl.Required("key")
			})
			dsl.StreamingResult(event, func() {
				dsl.View("detailed")
			})
			dsl.JSONRPC(func() {})
		})
	})
}
