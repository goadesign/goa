// This file renders a JSON-RPC WebSocket client and runs its request, response,
// timeout, close, and concurrent-send tests with Go's race detector.
package codegen_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	goacodegen "goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
	jsonrpccodegen "goa.design/goa/v3/jsonrpc/codegen"
)

// TestGeneratedWebSocketClientAndServer renders the Chat WebSocket packages
// and runs their request, response, failure, and close tests with Go's race
// detector.
func TestGeneratedWebSocketClientAndServer(t *testing.T) {
	dir := renderWebSocketRuntimeModule(t)
	clientDir := filepath.Join(dir, "jsonrpc", "chat", "client")
	serverDir := filepath.Join(dir, "jsonrpc", "chat", "server")
	require.NoError(t, os.WriteFile(filepath.Join(clientDir, "websocket_client_test.go"), []byte(websocketClientTest), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(serverDir, "websocket_server_test.go"), []byte(websocketServerTest), 0o600))

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "test", "-race", "-mod=mod", "./jsonrpc/chat/client", "./jsonrpc/chat/server")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
}

// renderWebSocketRuntimeModule writes the service, client, and server files
// needed by the generated tests. The temporary module uses this Goa checkout.
func renderWebSocketRuntimeModule(t *testing.T) string {
	t.Helper()
	root := expr.RunDSL(t, websocketConnectionDSL)
	generation, err := goacodegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	plan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	httpPlans, err := httpcodegen.NewJSONRPCPlans(generation, httpcodegen.PlanInput{
		Root:    root,
		Service: plan,
	})
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

	files, err := service.Files(plan)
	require.NoError(t, err)
	files = append(files, jsonPlans[0].ClientFiles()...)
	files = append(files, jsonPlans[0].ServerFiles()...)
	files = append(files, jsonPlans[0].ClientTypeFiles()...)
	files = append(files, jsonPlans[0].ServerTypeFiles()...)
	files = append(files, jsonPlans[0].PathFiles()...)

	base := t.TempDir()
	for _, file := range files {
		_, err := file.Render(base)
		require.NoError(t, err)
	}
	moduleDir := filepath.Join(base, goacodegen.Gendir)
	workingDir, err := os.Getwd()
	require.NoError(t, err)
	repository := filepath.Clean(filepath.Join(workingDir, "..", ".."))
	goMod := fmt.Sprintf("module generated.local/gen\n\ngo 1.25\n\nrequire goa.design/goa/v3 v3.0.0\n\nreplace goa.design/goa/v3 => %s\n", repository)
	require.NoError(t, os.WriteFile(filepath.Join(moduleDir, "go.mod"), []byte(goMod), 0o600))
	return moduleDir
}

const websocketClientTest = `package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	chat "generated.local/gen/chat"
	goahttp "goa.design/goa/v3/http"
	"goa.design/goa/v3/jsonrpc"
)

type contextKey string

type callbackResult struct {
	contextValue any
	err          error
}

type errorEvent struct {
	type_        jsonrpc.StreamErrorType
	contextValue any
	err          error
	response     *jsonrpc.RawResponse
}

type routedResponse struct {
	stream   string
	response *jsonrpc.RawResponse
}

func TestMethodCloseWhileRecvWaitsForResponse(t *testing.T) {
	client, conn, stop := newClientConnection(t, nil, serverConfig{})
	defer stop()
	defer closeClient(t, client)
	for range 100 {
		streamCtx, cancel := context.WithCancel(context.Background())
		stream := &EchoClientStream{
			conn:         conn,
			owner:        &websocketRequestOwner{},
			ctx:          streamCtx,
			cancel:       cancel,
			pendingReady: make(chan struct{}, 1),
		}
		pending := &echoClientStreamPendingRequest{
			resultChan: make(chan echoClientStreamStreamResult, 1),
		}
		id, err := conn.sendRequest(context.Background(), request("close-pending"), stream.owner, func(ctx context.Context, response *jsonrpc.RawResponse, err error) {
			stream.completeResponse(ctx, pending, response, err)
		})
		if err != nil {
			t.Fatal(err)
		}
		pending.id = id
		stream.enqueuePending(pending)
		received := make(chan error, 1)
		go func() {
			_, err := stream.Recv()
			received <- err
		}()
		if err := stream.Close(); err != nil {
			t.Fatal(err)
		}
		if err := receive(t, received); err != errWebsocketMethodStreamClosed {
			t.Fatalf("Recv after Close returned %v", err)
		}
	}
}

func TestMethodCloseWhileRecvWaitsForFirstSend(t *testing.T) {
	client, conn, stop := newClientConnection(t, nil, serverConfig{})
	defer stop()
	defer closeClient(t, client)
	for range 100 {
		streamCtx, cancel := context.WithCancel(context.Background())
		stream := &EchoClientStream{
			conn:         conn,
			owner:        &websocketRequestOwner{},
			ctx:          streamCtx,
			cancel:       cancel,
			pendingReady: make(chan struct{}, 1),
		}
		started := make(chan struct{})
		received := make(chan error, 1)
		go func() {
			close(started)
			_, err := stream.Recv()
			received <- err
		}()
		<-started
		if err := stream.Close(); err != nil {
			t.Fatal(err)
		}
		if err := receive(t, received); err != errWebsocketMethodStreamClosed {
			t.Fatalf("Recv before Send returned %v after Close", err)
		}
	}
}

func TestConcurrentMethodStreamsReceiveTheirResponses(t *testing.T) {
	client, conn, stop := newClientConnection(t, nil, serverConfig{reverseResponses: true})
	defer stop()
	defer closeClient(t, client)

	type sentRequest struct {
		stream string
		id     string
		err    error
	}
	routed := make(chan routedResponse, 2)
	sent := make(chan sentRequest, 2)
	start := make(chan struct{})
	for _, stream := range []string{"alpha", "beta"} {
		stream := stream
		go func() {
			<-start
			id, err := conn.sendRequest(context.Background(), request(stream), &websocketRequestOwner{}, func(_ context.Context, response *jsonrpc.RawResponse, err error) {
				if err != nil {
					t.Errorf("complete %s: %v", stream, err)
					return
				}
				routed <- routedResponse{stream: stream, response: response}
			})
			sent <- sentRequest{stream: stream, id: id, err: err}
		}()
	}
	close(start)

	ids := make(map[string]string, 2)
	for range 2 {
		request := receive(t, sent)
		if request.err != nil {
			t.Fatalf("send %s: %v", request.stream, request.err)
		}
		ids[request.stream] = request.id
	}
	if ids["alpha"] == ids["beta"] {
		t.Fatalf("connection reused request ID %q", ids["alpha"])
	}
	for range 2 {
		response := receive(t, routed)
		if got := jsonrpc.IDToString(response.response.ID); got != ids[response.stream] {
			t.Errorf("%s response routed with ID %q, want %q", response.stream, got, ids[response.stream])
		}
		if !strings.Contains(string(response.response.Result), response.stream) {
			t.Errorf("%s callback received another method's result: %s", response.stream, response.response.Result)
		}
	}
}

func TestClosedMethodStreamRejectsRequestsAndNotifications(t *testing.T) {
	client, conn, stop := newClientConnection(t, nil, serverConfig{})
	defer stop()
	defer closeClient(t, client)
	owner := &websocketRequestOwner{}
	ctx := context.WithValue(context.Background(), contextKey("request"), "closed-owner")
	completed := make(chan callbackResult, 1)
	_, err := conn.sendRequest(ctx, request("first"), owner, func(ctx context.Context, _ *jsonrpc.RawResponse, err error) {
		if terminalErr := conn.terminalError(); terminalErr != nil {
			t.Errorf("owner close made connection terminal: %v", terminalErr)
		}
		completed <- callbackResult{contextValue: ctx.Value(contextKey("request")), err: err}
	})
	if err != nil {
		t.Fatal(err)
	}
	conn.closeOwner(owner)
	result := receive(t, completed)
	if result.contextValue != "closed-owner" || result.err == nil {
		t.Fatalf("unexpected completion: %#v", result)
	}
	if _, err := conn.sendRequest(ctx, request("after-close"), owner, func(context.Context, *jsonrpc.RawResponse, error) {}); err == nil || !strings.Contains(err.Error(), "stream is closed") {
		t.Fatalf("request after close error = %v", err)
	}
	if err := conn.sendNotification(ctx, request("after-close"), owner); err == nil || !strings.Contains(err.Error(), "stream is closed") {
		t.Fatalf("notification after close error = %v", err)
	}
	conn.stateMu.Lock()
	pending := len(conn.pending)
	conn.stateMu.Unlock()
	if pending != 0 {
		t.Fatalf("closed owner retained %d pending requests", pending)
	}
}

func TestRequestTimeoutReturnsWithoutRecv(t *testing.T) {
	events := make(chan errorEvent, 4)
	var conn *websocketClientConn
	client, gotConn, stop := newClientConnection(t, func(ctx context.Context, type_ jsonrpc.StreamErrorType, err error, response *jsonrpc.RawResponse) {
		if terminalErr := connTerminalError(conn); terminalErr != nil {
			t.Errorf("timeout made connection terminal: %v", terminalErr)
		}
		events <- errorEvent{type_: type_, contextValue: ctx.Value(contextKey("request")), err: err, response: response}
	}, serverConfig{}, jsonrpc.WithRequestTimeout(20*time.Millisecond))
	conn = gotConn
	defer stop()
	defer closeClient(t, client)
	owner := &websocketRequestOwner{}
	ctx := context.WithValue(context.Background(), contextKey("request"), "timeout-request")
	completed := make(chan callbackResult, 1)
	_, err := conn.sendRequest(ctx, request("timeout"), owner, func(ctx context.Context, _ *jsonrpc.RawResponse, err error) {
		if terminalErr := conn.terminalError(); terminalErr != nil {
			t.Errorf("timeout completion made connection terminal: %v", terminalErr)
		}
		completed <- callbackResult{contextValue: ctx.Value(contextKey("request")), err: err}
	})
	if err != nil {
		t.Fatal(err)
	}
	result := receive(t, completed)
	if result.contextValue != "timeout-request" || result.err == nil || !strings.Contains(result.err.Error(), "timed out") {
		t.Fatalf("unexpected timeout completion: %#v", result)
	}
	event := receive(t, events)
	if event.type_ != jsonrpc.StreamErrorTimeout || event.contextValue != "timeout-request" {
		t.Fatalf("unexpected timeout event: %#v", event)
	}
	conn.stateMu.Lock()
	pending := len(conn.pending)
	conn.stateMu.Unlock()
	if pending != 0 {
		t.Fatalf("timeout retained %d pending requests", pending)
	}
}

func TestConnectionFailureUsesConnectionContext(t *testing.T) {
	events := make(chan errorEvent, 8)
	closeAfterRequest := make(chan struct{})
	var conn *websocketClientConn
	client, gotConn, stop := newClientConnection(t, func(ctx context.Context, type_ jsonrpc.StreamErrorType, err error, response *jsonrpc.RawResponse) {
		if conn != nil {
			if terminalErr := conn.terminalError(); terminalErr == nil {
				t.Error("connection error callback ran before terminal state")
			}
		}
		events <- errorEvent{type_: type_, contextValue: ctx.Value(contextKey("request")), err: err, response: response}
	}, serverConfig{requestHook: closeAfterRequest})
	conn = gotConn
	defer stop()
	defer closeClient(t, client)
	owner := &websocketRequestOwner{}
	ctx := context.WithValue(context.Background(), contextKey("request"), "failed-request")
	completed := make(chan callbackResult, 1)
	_, err := conn.sendRequest(ctx, request("fail"), owner, func(ctx context.Context, _ *jsonrpc.RawResponse, err error) {
		if terminalErr := conn.terminalError(); terminalErr == nil {
			t.Error("failure completion ran before terminal state")
		}
		completed <- callbackResult{contextValue: ctx.Value(contextKey("request")), err: err}
	})
	if err != nil {
		t.Fatal(err)
	}
	close(closeAfterRequest)
	result := receive(t, completed)
	if result.contextValue != "failed-request" || result.err == nil {
		t.Fatalf("unexpected failure completion: %#v", result)
	}
	for {
		event := receive(t, events)
		if event.type_ == jsonrpc.StreamErrorConnection {
			if event.contextValue != nil {
				t.Fatalf("connection event inherited request context: %#v", event.contextValue)
			}
			break
		}
	}
	if err := conn.sendNotification(context.Background(), request("failed"), &websocketRequestOwner{}); err == nil {
		t.Fatal("terminal connection accepted a notification")
	}
}

func TestEchoRecvReturnsConnectionFailure(t *testing.T) {
	closeAfterRequest := make(chan struct{})
	client, conn, stop := newClientConnection(t, nil, serverConfig{requestHook: closeAfterRequest})
	defer stop()
	defer closeClient(t, client)
	streamCtx, cancel := context.WithCancel(context.Background())
	stream := &EchoClientStream{
		conn:         conn,
		owner:        &websocketRequestOwner{},
		ctx:          streamCtx,
		cancel:       cancel,
		pendingReady: make(chan struct{}, 1),
	}
	if err := stream.Send(&chat.EchoPayload{}); err != nil {
		t.Fatal(err)
	}
	received := make(chan error, 1)
	go func() {
		_, err := stream.Recv()
		received <- err
	}()
	close(closeAfterRequest)
	err := receive(t, received)
	if err == errWebsocketMethodStreamClosed {
		t.Fatalf("Recv returned the method Close error after a socket failure: %v", err)
	}
	connectionErr := conn.terminalError()
	if connectionErr == nil {
		t.Fatal("connection has no error after the server closed the socket")
	}
	if err != connectionErr {
		t.Fatalf("Recv error = %v, want the connection error %v", err, connectionErr)
	}
}

func TestServerNotificationsAndNullIDsReportCorrectErrors(t *testing.T) {
	events := make(chan errorEvent, 8)
	serverMessages := make(chan []any, 1)
	serverMessages <- []any{
		map[string]any{"jsonrpc": "2.0", "method": "tick", "params": map[string]any{"value": 1}},
		map[string]any{"jsonrpc": "2.0", "id": nil, "error": map[string]any{"code": -32603, "message": "failed"}},
	}
	client, _, stop := newClientConnection(t, func(ctx context.Context, type_ jsonrpc.StreamErrorType, err error, response *jsonrpc.RawResponse) {
		events <- errorEvent{type_: type_, contextValue: ctx.Value(contextKey("request")), err: err, response: response}
	}, serverConfig{messages: serverMessages})
	defer stop()
	defer closeClient(t, client)
	notification := receive(t, events)
	if notification.type_ != jsonrpc.StreamErrorNotification || !strings.Contains(notification.err.Error(), "tick") {
		t.Fatalf("method notification misclassified: %#v", notification)
	}
	nullID := receive(t, events)
	if nullID.type_ != jsonrpc.StreamErrorProtocol || nullID.response == nil || nullID.response.Error == nil {
		t.Fatalf("null-ID error misclassified: %#v", nullID)
	}
}

func TestMethodCloseRejectsConcurrentRequests(t *testing.T) {
	client, conn, stop := newClientConnection(t, nil, serverConfig{})
	defer stop()
	defer closeClient(t, client)
	owner := &websocketRequestOwner{}
	var sends sync.WaitGroup
	for i := 0; i < 32; i++ {
		sends.Add(1)
		go func(index int) {
			defer sends.Done()
			_, err := conn.sendRequest(context.Background(), request(fmt.Sprintf("race-%d", index)), owner, func(context.Context, *jsonrpc.RawResponse, error) {})
			if err != nil && !strings.Contains(err.Error(), "stream is closed") {
				t.Errorf("concurrent send: %v", err)
			}
		}(i)
	}
	conn.closeOwner(owner)
	sends.Wait()
	if _, err := conn.sendRequest(context.Background(), request("after-race"), owner, func(context.Context, *jsonrpc.RawResponse, error) {}); err == nil {
		t.Fatal("closed owner accepted request after concurrent sends")
	}
	conn.stateMu.Lock()
	pending := len(conn.pending)
	conn.stateMu.Unlock()
	if pending != 0 {
		t.Fatalf("owner close race retained %d requests", pending)
	}
}

type serverConfig struct {
	requestHook     <-chan struct{}
	messages        <-chan []any
	reverseResponses bool
}

func newClientConnection(t *testing.T, handler jsonrpc.StreamErrorHandler, config serverConfig, options ...jsonrpc.StreamConfigOption) (*Client, *websocketClientConn, func()) {
	t.Helper()
	streamOptions := append([]jsonrpc.StreamConfigOption(nil), options...)
	if handler != nil {
		streamOptions = append(streamOptions, jsonrpc.WithErrorHandler(handler))
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer func() {
			if err := ws.Close(); err != nil {
				t.Errorf("close server WebSocket: %v", err)
			}
		}()
		if config.messages != nil {
			for _, message := range <-config.messages {
				if err := ws.WriteJSON(message); err != nil {
					return
				}
			}
		}
		if config.reverseResponses {
			requests := make([]jsonrpc.RawRequest, 2)
			for i := range requests {
				if err := ws.ReadJSON(&requests[i]); err != nil {
					return
				}
			}
			for i := len(requests) - 1; i >= 0; i-- {
				result := map[string]any{"method": requests[i].Method}
				if err := ws.WriteJSON(jsonrpc.MakeSuccessResponse(requests[i].ID, result)); err != nil {
					return
				}
			}
		}
		for {
			var message any
			if err := ws.ReadJSON(&message); err != nil {
				return
			}
			if config.requestHook != nil {
				<-config.requestHook
				return
			}
		}
	}))
	host := strings.TrimPrefix(server.URL, "http://")
	client := NewClient("http", host, http.DefaultClient, goahttp.RequestEncoder, goahttp.ResponseDecoder, false, websocket.DefaultDialer, nil, streamOptions...)
	conn, err := client.getConn(context.Background())
	if err != nil {
		server.Close()
		t.Fatal(err)
	}
	return client, conn, server.Close
}

func request(method string) *jsonrpc.Request {
	return &jsonrpc.Request{JSONRPC: "2.0", Method: method}
}

func receive[T any](t *testing.T, channel <-chan T) T {
	t.Helper()
	select {
	case value := <-channel:
		return value
	case <-time.After(3 * time.Second):
		var zero T
		t.Fatal("timed out waiting for generated WebSocket lifecycle event")
		return zero
	}
}

func connTerminalError(conn *websocketClientConn) error {
	if conn == nil {
		return nil
	}
	return conn.terminalError()
}

func closeClient(t *testing.T, client *Client) {
	t.Helper()
	if err := client.Close(); err != nil {
		t.Errorf("close client: %v", err)
	}
}
`

const websocketServerTest = `package server

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

type trackedConn struct {
	net.Conn
	closed chan struct{}
	once   sync.Once
}

type trackingWriter struct {
	http.ResponseWriter
	connections chan<- *trackedConn
}

func TestCloseSendsNormalCodeWithoutWaitingForDataWrite(t *testing.T) {
	stream, peer, _, stop := newServerConnection(t)
	defer stop()
	peerResult := make(chan error, 1)
	go func() {
		_, _, err := peer.ReadMessage()
		peerResult <- err
	}()
	stream.writeMu.Lock()
	defer stream.writeMu.Unlock()
	closeResult := make(chan error, 1)
	go func() {
		closeResult <- stream.Close()
	}()
	if err := receiveServer(t, closeResult); err != nil {
		t.Fatal(err)
	}
	err := receiveServer(t, peerResult)
	var closeErr *websocket.CloseError
	if !errors.As(err, &closeErr) || closeErr.Code != websocket.CloseNormalClosure {
		t.Fatalf("peer close error = %v, want code %d", err, websocket.CloseNormalClosure)
	}
}

func TestCloseClosesSocketWhenNormalMessageFails(t *testing.T) {
	stream, peer, connection, stop := newServerConnection(t)
	defer stop()
	tcp, ok := peer.UnderlyingConn().(*net.TCPConn)
	if !ok {
		t.Fatalf("peer connection type = %T, want *net.TCPConn", peer.UnderlyingConn())
	}
	if err := tcp.SetLinger(0); err != nil {
		t.Fatal(err)
	}
	if err := peer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := stream.conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, _, err := stream.conn.ReadMessage(); err == nil {
		t.Fatal("server read succeeded after the peer reset the connection")
	}
	err := stream.Close()
	if err == nil || !strings.Contains(err.Error(), "write normal WebSocket close message") {
		t.Fatalf("Close error = %v, want normal-close write failure", err)
	}
	select {
	case <-connection.closed:
	case <-time.After(3 * time.Second):
		t.Fatal("Close did not close the server socket after the control write failed")
	}
}

func (c *trackedConn) Close() error {
	c.once.Do(func() {
		close(c.closed)
	})
	return c.Conn.Close()
}

func (w *trackingWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("HTTP response writer does not support connection takeover")
	}
	connection, buffer, err := hijacker.Hijack()
	if err != nil {
		return nil, nil, err
	}
	tracked := &trackedConn{Conn: connection, closed: make(chan struct{})}
	w.connections <- tracked
	return tracked, buffer, nil
}

func newServerConnection(t *testing.T) (*chatStream, *websocket.Conn, *trackedConn, func()) {
	t.Helper()
	serverConnections := make(chan *websocket.Conn, 1)
	trackedConnections := make(chan *trackedConn, 1)
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := &trackingWriter{ResponseWriter: w, connections: trackedConnections}
		connection, err := (&websocket.Upgrader{}).Upgrade(writer, r, nil)
		if err != nil {
			t.Errorf("upgrade server connection: %v", err)
			return
		}
		serverConnections <- connection
		<-release
	}))
	peer, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), nil)
	if err != nil {
		server.Close()
		t.Fatal(err)
	}
	connection := receiveServer(t, serverConnections)
	tracked := receiveServer(t, trackedConnections)
	var once sync.Once
	stop := func() {
		once.Do(func() {
			close(release)
			if err := peer.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
				t.Errorf("close peer WebSocket: %v", err)
			}
			server.Close()
		})
	}
	return &chatStream{conn: connection}, peer, tracked, stop
}

func receiveServer[T any](t *testing.T, values <-chan T) T {
	t.Helper()
	select {
	case value := <-values:
		return value
	case <-time.After(3 * time.Second):
		var zero T
		t.Fatal("timed out waiting for generated WebSocket server test")
		return zero
	}
}
`
