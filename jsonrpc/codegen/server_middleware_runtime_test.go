// These tests generate ordinary, streaming and mixed JSON-RPC services. Mounted
// requests and direct HTTP calls pass through Server.Use middleware before
// any endpoint runs, and a middleware rejection must prevent endpoint work.
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
)

// TestGeneratedJSONRPCServerMiddleware verifies direct and mounted HTTP calls
// through the middleware installed after route registration.
func TestGeneratedJSONRPCServerMiddleware(t *testing.T) {
	dir := t.TempDir()
	workingDir, err := os.Getwd()
	require.NoError(t, err)
	repository := filepath.Clean(filepath.Join(workingDir, "..", ".."))
	module := fmt.Sprintf("module middleware-contract.local\n\ngo 1.26.0\n\nrequire goa.design/goa/v3 v3.0.0\n\nreplace goa.design/goa/v3 => %s\n", filepath.ToSlash(repository))
	require.NoError(t, os.Mkdir(filepath.Join(dir, "design"), 0o700))
	for name, source := range map[string]string{
		"go.mod":             module,
		"design/design.go":   middlewareRuntimeDesign,
		"middleware_test.go": middlewareRuntimeTest,
	} {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(source), 0o600))
	}
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Minute)
	defer cancel()
	for _, arguments := range [][]string{
		{"run", "-mod=mod", "goa.design/goa/v3/cmd/goa", "gen", "middleware-contract.local/design"},
		{"test", "-mod=mod", "-race", "-p=1", "./..."},
	} {
		command := exec.CommandContext(ctx, "go", arguments...)
		command.Dir = dir
		command.Env = append(os.Environ(), "GOWORK=off", "GOFLAGS=-mod=mod -p=1")
		output, err := command.CombinedOutput()
		require.NoError(t, err, string(output))
	}
}

const middlewareRuntimeDesign = `package design

import . "goa.design/goa/v3/dsl"

var _ = API("middleware_contract", func() {
 Description("Verify the native JSON-RPC server middleware contract")
})

var _ = Service("ordinary", func() {
 JSONRPC(func() { POST("/ordinary") })
 Method("echo", func() {
  Result(String)
  JSONRPC(func() {})
 })
})

var _ = Service("streaming", func() {
 JSONRPC(func() { POST("/streaming") })
 Method("tick", func() {
  StreamingResult(String)
  JSONRPC(func() { ServerSentEvents() })
 })
})

var _ = Service("mixed", func() {
 JSONRPC(func() { POST("/mixed") })
 Method("echo", func() {
  Result(String)
  JSONRPC(func() {})
 })
 Method("tick", func() {
  StreamingResult(String)
  JSONRPC(func() { ServerSentEvents() })
 })
})
`

const middlewareRuntimeTest = `package middlewarecontract_test

import (
 "context"
 "errors"
 "io"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"

 "github.com/stretchr/testify/require"
 genordinary "middleware-contract.local/gen/ordinary"
 genordinarysrv "middleware-contract.local/gen/jsonrpc/ordinary/server"
 genstreaming "middleware-contract.local/gen/streaming"
 genstreamingsrv "middleware-contract.local/gen/jsonrpc/streaming/server"
 genmixed "middleware-contract.local/gen/mixed"
 genmixedsrv "middleware-contract.local/gen/jsonrpc/mixed/server"
 goahttp "goa.design/goa/v3/http"
)

type (
 contextKey struct{}
 middlewareServer interface {
  goahttp.Server
  http.Handler
 }
)

// TestServerMiddlewareRunsInOrderAndCanReject checks context changes and
// middleware rejections through every native HTTP response type.
func TestServerMiddlewareRunsInOrderAndCanReject(t *testing.T) {
 tests := []struct {
  name, path, method, accept string
  makeServer func(goahttp.Muxer, func(context.Context) error) middlewareServer
 }{
  {"ordinary", "/ordinary", "echo", "application/json", func(mux goahttp.Muxer, invoked func(context.Context) error) middlewareServer {
   endpoints := &genordinary.Endpoints{Echo: func(ctx context.Context, _ any) (any, error) {
    if err := invoked(ctx); err != nil {
     return nil, err
    }
    return "value", nil
   }}
   return genordinarysrv.New(endpoints, mux, goahttp.RequestDecoder, goahttp.ResponseEncoder, nil)
  }},
  {"streaming", "/streaming", "tick", "text/event-stream", func(mux goahttp.Muxer, invoked func(context.Context) error) middlewareServer {
   endpoints := &genstreaming.Endpoints{Tick: func(ctx context.Context, raw any) (any, error) {
    if err := invoked(ctx); err != nil {
     return nil, err
    }
    return nil, raw.(*genstreaming.TickEndpointInput).Stream.Send("value")
   }}
   return genstreamingsrv.New(endpoints, mux, goahttp.RequestDecoder, goahttp.ResponseEncoder, nil)
  }},
  {"mixed_json", "/mixed", "echo", "application/json", func(mux goahttp.Muxer, invoked func(context.Context) error) middlewareServer {
   endpoints := &genmixed.Endpoints{Echo: func(ctx context.Context, _ any) (any, error) {
    if err := invoked(ctx); err != nil {
     return nil, err
    }
    return "value", nil
   }}
   return genmixedsrv.New(endpoints, mux, goahttp.RequestDecoder, goahttp.ResponseEncoder, nil)
  }},
  {"mixed_stream", "/mixed", "tick", "text/event-stream", func(mux goahttp.Muxer, invoked func(context.Context) error) middlewareServer {
   endpoints := &genmixed.Endpoints{Tick: func(ctx context.Context, raw any) (any, error) {
    if err := invoked(ctx); err != nil {
     return nil, err
    }
    return nil, raw.(*genmixed.TickEndpointInput).Stream.Send("value")
   }}
   return genmixedsrv.New(endpoints, mux, goahttp.RequestDecoder, goahttp.ResponseEncoder, nil)
  }},
 }
 for _, tc := range tests {
  for _, mounted := range []bool{false, true} {
   entry := "direct"
   if mounted {
    entry = "mounted"
   }
   t.Run(tc.name+"/"+entry, func(t *testing.T) {
   events := make(chan string, 3)
   invoked := func(ctx context.Context) error {
    events <- "endpoint"
    if ctx.Value(contextKey{}) != "outer/inner" {
     return errors.New("middleware context missing")
    }
    return nil
   }
   mux := goahttp.NewMuxer()
   server := tc.makeServer(mux, invoked)
   var handler http.Handler = server
   if mounted {
    goahttp.Servers{server}.Mount(mux)
    handler = mux
   }
   // Registering middleware after mounting must still wrap the served request.
   server.Use(func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
     events <- "inner"
     if request.Context().Value(contextKey{}) != "outer" {
      http.Error(writer, "outer context missing", http.StatusInternalServerError)
      return
     }
     ctx := context.WithValue(request.Context(), contextKey{}, "outer/inner")
     next.ServeHTTP(writer, request.WithContext(ctx))
    })
   })
   server.Use(func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
     events <- "outer"
     writer.Header().Set("X-Synthetic-Middleware", "installed")
     if request.Header.Get("X-Synthetic-Deny") == "yes" {
      http.Error(writer, "Rejected by middleware", http.StatusForbidden)
      return
     }
     ctx := context.WithValue(request.Context(), contextKey{}, "outer")
     next.ServeHTTP(writer, request.WithContext(ctx))
    })
   })
   peer := httptest.NewServer(handler)
   defer peer.Close()
   for _, denied := range []bool{false, true} {
    body := "{\"jsonrpc\":\"2.0\",\"id\":\"request-1\",\"method\":\"" + tc.method + "\"}"
    request, err := http.NewRequestWithContext(t.Context(), http.MethodPost, peer.URL+tc.path, strings.NewReader(body))
    require.NoError(t, err)
    request.Header.Set("Content-Type", "application/json")
    request.Header.Set("Accept", tc.accept)
    if denied {
     request.Header.Set("X-Synthetic-Deny", "yes")
    }
    response, err := peer.Client().Do(request)
    require.NoError(t, err)
    encoded, readErr := io.ReadAll(response.Body)
    closeErr := response.Body.Close()
    require.NoError(t, readErr)
    require.NoError(t, closeErr)
    require.Equal(t, "installed", response.Header.Get("X-Synthetic-Middleware"))
    expectedEvents := 3
    if denied {
     expectedEvents = 1
    }
    require.Len(t, events, expectedEvents)
    require.Equal(t, "outer", <-events)
    if denied {
     require.Equal(t, http.StatusForbidden, response.StatusCode)
     require.Empty(t, events)
     continue
    }
    require.Equal(t, "inner", <-events)
    require.Equal(t, "endpoint", <-events)
    require.Empty(t, events)
    require.Equal(t, http.StatusOK, response.StatusCode)
    if tc.accept == "application/json" {
     require.JSONEq(t, "{\"jsonrpc\":\"2.0\",\"id\":\"request-1\",\"result\":\"value\"}", string(encoded))
    } else {
     require.Equal(t, "text/event-stream", response.Header.Get("Content-Type"))
     require.Contains(t, string(encoded), "\"params\":[\"value\"]")
     require.Contains(t, string(encoded), "\"id\":\"request-1\"")
     require.Contains(t, string(encoded), "\"result\":null")
    }
   }
   })
  }
 }
}
`
