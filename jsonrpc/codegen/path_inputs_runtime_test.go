// These tests compile JSON-RPC clients and servers for named route inputs. URL
// fields stay outside params, retain each method's types and validations, and
// reach configured endpoints through mounted HTTP middleware.
package codegen_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestGeneratedJSONRPCPathInputs generates and runs real clients and servers
// so URL decoding, validation and endpoint delivery are checked together.
func TestGeneratedJSONRPCPathInputs(t *testing.T) {
	for _, mapped := range []bool{false, true} {
		name := "same names"
		design := pathInputsDesign
		if mapped {
			name = "mapped names"
			design = strings.ReplaceAll(design, "{organization_id}", "{organization}")
			design = strings.ReplaceAll(design, " JSONRPC(func() {", " JSONRPC(func() { Param(\"organization_id:organization\");")
			design = strings.Replace(design, "JSONRPC(func() { Param(\"organization_id:organization\");\n  Path", "JSONRPC(func() {\n  Path", 1)
			design = strings.ReplaceAll(design, "HTTP(func() { GET(", "HTTP(func() { Param(\"organization_id:organization\"); GET(")
		}
		t.Run(name, func(t *testing.T) {
			runGeneratedPathInputs(t, design)
		})
	}
}

// runGeneratedPathInputs generates each authored transport and runs its clients
// against mounted servers, so a successful render alone cannot pass this test.
func runGeneratedPathInputs(t *testing.T, design string) {
	t.Helper()
	directory := t.TempDir()
	repository, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	module := fmt.Sprintf("module path-inputs.local\n\ngo 1.26.0\n\nrequire goa.design/goa/v3 v3.0.0\n\nreplace goa.design/goa/v3 => %s\n", filepath.ToSlash(repository))
	require.NoError(t, os.Mkdir(filepath.Join(directory, "design"), 0o700))
	for name, source := range map[string]string{
		"go.mod":              module,
		"design/design.go":    design,
		"path_inputs_test.go": pathInputsRuntimeTest,
	} {
		require.NoError(t, os.WriteFile(filepath.Join(directory, name), []byte(source), 0o600))
	}
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Minute)
	defer cancel()
	for _, arguments := range [][]string{
		{"run", "-mod=mod", "goa.design/goa/v3/cmd/goa", "gen", "path-inputs.local/design"},
		{"test", "-mod=mod", "-race", "-p=1", "./..."},
	} {
		command := exec.CommandContext(ctx, "go", arguments...)
		command.Dir = directory
		command.Env = append(os.Environ(), "GOWORK=off", "GOFLAGS=-mod=mod -p=1")
		output, err := command.CombinedOutput()
		require.NoError(t, err, string(output))
	}
}

const pathInputsDesign = `package design

import . "goa.design/goa/v3/dsl"

var _ = API("path_inputs", func() {
 Description("Verify URL inputs with synthetic JSON-RPC methods")
})

var organization = Type("Organization", String, func() {
 Pattern("^[a-z]+$")
})

var _ = Service("route", func() {
 Description("Retain each method's inputs on one shared JSON-RPC route")
 JSONRPC(func() {
  Path("/prefix")
  POST("/organizations/{organization_id}/rpc")
 })
 Method("text", func() {
  Payload(func() {
   Field(1, "organization_id", organization, "Organization selected by the address", func() {
    Meta("struct:field:name", "Organization")
   })
   Required("organization_id")
  })
  Result(String)
  JSONRPC(func() {})
 })
 Method("number", func() {
  Payload(func() {
   Field(1, "organization_id", Int64, "Positive synthetic organization number", func() {
    Minimum(1)
   })
   Required("organization_id")
  })
  Result(String)
  JSONRPC(func() {})
 })
 Method("body", func() {
  Payload(func() {
   Field(1, "organization_id", organization, "Organization selected by the address")
   Field(2, "value", String, "Value carried in protocol params")
   Required("organization_id", "value")
  })
  Result(String)
  JSONRPC(func() {})
 })
 Method("collision", func() {
  Payload(func() {
   Field(1, "organization_id", organization, "Organization selected by the address")
   Field(2, "organization", String, "Independent domain value with the URL name")
   Field(3, "query_value", String, "Value carried by the query")
   Field(4, "header_value", String, "Value carried by a header")
   Field(5, "cookie_value", String, "Value carried by a cookie")
   Required("organization_id", "organization", "query_value", "header_value", "cookie_value")
  })
  Result(String)
  JSONRPC(func() {
   Param("query_value:query")
   Header("header_value:X-Value")
   Cookie("cookie_value:value")
  })
 })
 Method("identified", func() {
  Payload(func() {
   Field(1, "organization_id", organization, "Organization selected by the address")
   ID("request_id", String, "Protocol request identity")
   Required("organization_id", "request_id")
  })
  Result(String)
  JSONRPC(func() {})
 })
 Method("optional", func() {
  Payload(func() {
   Field(1, "organization_id", organization, "Organization selected by the address")
  })
  Result(String)
  JSONRPC(func() {})
 })
 Method("list", func() {
  Payload(func() {
   Field(1, "organization_id", ArrayOf(Int), "Synthetic organization numbers", func() {
    MinLength(1)
   })
   Required("organization_id")
  })
  Result(String)
  JSONRPC(func() {})
 })
 Method("notice", func() {
  Payload(func() {
   Field(1, "organization_id", organization, "Organization selected by the address")
   Required("organization_id")
  })
  JSONRPC(func() { Notification() })
 })
 Method("tick", func() {
  Payload(func() {
   Field(1, "organization_id", organization, "Organization selected by the address")
   Required("organization_id")
  })
  StreamingResult(String)
  JSONRPC(func() { ServerSentEvents() })
 })
})

var _ = Service("native_http", func() {
 Description("Verify the shared HTTP request builder with a renamed path field")
 HTTP(func() { Path("/http/prefix") })
 Method("echo", func() {
  Payload(func() {
   Field(1, "organization_id", organization, "Organization selected by the address", func() {
    Meta("struct:field:name", "Organization")
   })
   Required("organization_id")
  })
  Result(String)
  HTTP(func() { GET("/organizations/{organization_id}") })
 })
})
`

const pathInputsRuntimeTest = `package pathinputs_test

import (
 "bytes"
 "context"
 "errors"
 "fmt"
 "io"
 "net/http"
 "net/http/httptest"
 "net/url"
 "strings"
 "sync/atomic"
 "testing"

 "github.com/stretchr/testify/require"
 goahttp "goa.design/goa/v3/http"
 goa "goa.design/goa/v3/pkg"
 genhttpclient "path-inputs.local/gen/http/native_http/client"
 genhttpserver "path-inputs.local/gen/http/native_http/server"
 gennativehttp "path-inputs.local/gen/native_http"
 genclient "path-inputs.local/gen/jsonrpc/route/client"
 genserver "path-inputs.local/gen/jsonrpc/route/server"
 genroute "path-inputs.local/gen/route"
)

type (
 routeContext struct{}
 recordingDoer struct {
  client *http.Client
  path, body string
 }
)

// Do records the actual URL and envelope, then sends the same request unchanged.
func (d *recordingDoer) Do(request *http.Request) (*http.Response, error) {
 encoded, err := io.ReadAll(request.Body)
 if err != nil {
  return nil, err
 }
 if err := request.Body.Close(); err != nil {
  return nil, err
 }
 d.path, d.body = request.URL.EscapedPath(), string(encoded)
 request.Body = io.NopCloser(bytes.NewReader(encoded))
 return d.client.Do(request)
}

// TestNamedRouteInputsReachConfiguredEndpoints sends generated client calls
// and invalid wire requests through the mounted server to check their outcomes.
func TestNamedRouteInputsReachConfiguredEndpoints(t *testing.T) {
 var calls atomic.Int64
 var nativeCalls atomic.Int64
 invoked := func(ctx context.Context) error {
  if ctx.Value(routeContext{}) != "installed" {
   return errors.New("mounted middleware context missing")
  }
  calls.Add(1)
  return nil
 }
 endpoints := &genroute.Endpoints{
  Text: func(ctx context.Context, raw any) (any, error) {
   if err := invoked(ctx); err != nil {
    return nil, err
   }
   return string(raw.(*genroute.TextPayload).Organization), nil
  },
  Number: func(ctx context.Context, raw any) (any, error) {
   if err := invoked(ctx); err != nil {
    return nil, err
   }
   return fmt.Sprint(raw.(*genroute.NumberPayload).OrganizationID), nil
  },
  Body: func(ctx context.Context, raw any) (any, error) {
   if err := invoked(ctx); err != nil {
    return nil, err
   }
   payload := raw.(*genroute.BodyPayload)
   return string(payload.OrganizationID) + "/" + payload.Value, nil
  },
  Collision: func(ctx context.Context, raw any) (any, error) {
   if err := invoked(ctx); err != nil {
    return nil, err
   }
   payload := raw.(*genroute.CollisionPayload)
   return string(payload.OrganizationID) + "/" + payload.Organization + "/" + payload.QueryValue + "/" + payload.HeaderValue + "/" + payload.CookieValue, nil
  },
  Identified: func(ctx context.Context, raw any) (any, error) {
   if err := invoked(ctx); err != nil {
    return nil, err
   }
   payload := raw.(*genroute.IdentifiedPayload)
   return string(payload.OrganizationID) + "/" + payload.RequestID, nil
  },
  Optional: func(ctx context.Context, raw any) (any, error) {
   if err := invoked(ctx); err != nil {
    return nil, err
   }
   return string(*raw.(*genroute.OptionalPayload).OrganizationID), nil
  },
  List: func(ctx context.Context, raw any) (any, error) {
   if err := invoked(ctx); err != nil {
    return nil, err
   }
   return fmt.Sprint(raw.(*genroute.ListPayload).OrganizationID), nil
  },
  Notice: func(ctx context.Context, raw any) (any, error) {
   if err := invoked(ctx); err != nil {
    return nil, err
   }
   if raw.(*genroute.NoticePayload).OrganizationID != "blue" {
    return nil, errors.New("notification route value changed")
   }
   return nil, nil
  },
  Tick: func(ctx context.Context, raw any) (any, error) {
   if err := invoked(ctx); err != nil {
    return nil, err
   }
   input := raw.(*genroute.TickEndpointInput)
   return nil, input.Stream.Send(string(input.Payload.OrganizationID))
  },
 }
 // Mount before installing middleware to check the route uses the current handler.
 mux := goahttp.NewMuxer()
 server := genserver.New(endpoints, mux, goahttp.RequestDecoder, goahttp.ResponseEncoder, nil)
 server.Mount(mux)
 native := genhttpserver.New(&gennativehttp.Endpoints{
  Echo: func(_ context.Context, raw any) (any, error) {
   nativeCalls.Add(1)
   return string(raw.(*gennativehttp.EchoPayload).Organization), nil
  },
 }, mux, goahttp.RequestDecoder, goahttp.ResponseEncoder, nil, nil)
 native.Mount(mux)
 server.Use(func(next http.Handler) http.Handler {
  return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
   ctx := context.WithValue(request.Context(), routeContext{}, "installed")
   next.ServeHTTP(writer, request.WithContext(ctx))
  })
 })
 peer := httptest.NewServer(mux)
 defer peer.Close()
 location, err := url.Parse(peer.URL)
 require.NoError(t, err)
 doer := &recordingDoer{client: peer.Client()}
 client := genclient.NewClient(location.Scheme, location.Host, doer, goahttp.RequestEncoder, goahttp.ResponseDecoder, false)
 organization := genroute.Organization("blue")
 nativeClient := genhttpclient.NewClient(location.Scheme, location.Host, peer.Client(), goahttp.RequestEncoder, goahttp.ResponseDecoder, false)
 nativeResult, err := nativeClient.Echo()(t.Context(), &gennativehttp.EchoPayload{Organization: "green"})
 require.NoError(t, err)
 require.Equal(t, "green", nativeResult)
 require.Equal(t, int64(1), nativeCalls.Load())
 invalidNative, err := peer.Client().Get(peer.URL + "/http/prefix/organizations/BAD")
 require.NoError(t, err)
 require.NoError(t, invalidNative.Body.Close())
 require.Equal(t, http.StatusBadRequest, invalidNative.StatusCode)
 require.Equal(t, int64(1), nativeCalls.Load())
 // Generated clients must send route values only in the URL.
 tests := []struct {
  name string
  endpoint goa.Endpoint
  payload any
  result, path, params string
 }{
  {"named string", client.Text(), &genroute.TextPayload{Organization: "blue"}, "blue", "blue", ""},
  {"numeric lower bound", client.Number(), &genroute.NumberPayload{OrganizationID: 1}, "1", "1", ""},
  {"numeric other value", client.Number(), &genroute.NumberPayload{OrganizationID: 42}, "42", "42", ""},
  {"body", client.Body(), &genroute.BodyPayload{OrganizationID: "blue", Value: "domain"}, "blue/domain", "blue", ` + "`" + `"params":{"value":"domain"}` + "`" + `},
  {"separate domain and transport names", client.Collision(), &genroute.CollisionPayload{OrganizationID: "blue", Organization: "domain", QueryValue: "query", HeaderValue: "header", CookieValue: "cookie"}, "blue/domain/query/header/cookie", "blue", ` + "`" + `"params":{"organization":"domain"}` + "`" + `},
  {"mapped identity", client.Identified(), &genroute.IdentifiedPayload{OrganizationID: "blue", RequestID: "chosen-id"}, "blue/chosen-id", "blue", ""},
  {"optional service field", client.Optional(), &genroute.OptionalPayload{OrganizationID: &organization}, "blue", "blue", ""},
  {"numeric array", client.List(), &genroute.ListPayload{OrganizationID: []int{1, 2}}, "[1 2]", "1,2", ""},
  {"notification", client.Notice(), &genroute.NoticePayload{OrganizationID: "blue"}, "", "blue", ""},
 }
 for _, tc := range tests {
  t.Run(tc.name, func(t *testing.T) {
   before := calls.Load()
   result, err := tc.endpoint(t.Context(), tc.payload)
   require.NoError(t, err)
   if tc.name == "notification" {
    require.Nil(t, result)
   } else {
    require.Equal(t, tc.result, result)
   }
   require.Equal(t, before + 1, calls.Load())
   require.Equal(t, "/prefix/organizations/" + tc.path + "/rpc", doer.path)
   require.NotContains(t, doer.body, "organization_id")
   if tc.params == "" {
    require.NotContains(t, doer.body, ` + "`" + `"params"` + "`" + `)
   } else {
    require.Contains(t, doer.body, tc.params)
   }
  })
 }
 // Invalid URL values must fail decoding before any configured endpoint runs.
 for _, tc := range []struct {name, value, method, accept string}{
  {"string pattern", "BAD", "text", "application/json"},
  {"numeric parse", "wrong", "number", "application/json"},
  {"numeric below bound", "0", "number", "application/json"},
  {"array parse", "1,wrong", "list", "application/json"},
  {"stream pattern", "BAD", "tick", "text/event-stream"},
 } {
  t.Run(tc.name, func(t *testing.T) {
   before := calls.Load()
   envelope := "{\"jsonrpc\":\"2.0\",\"id\":\"invalid\",\"method\":\"" + tc.method + "\"}"
   request, err := http.NewRequestWithContext(t.Context(), http.MethodPost, peer.URL + "/prefix/organizations/" + tc.value + "/rpc", strings.NewReader(envelope))
   require.NoError(t, err)
   request.Header.Set("Content-Type", "application/json")
   request.Header.Set("Accept", tc.accept)
   response, err := peer.Client().Do(request)
   require.NoError(t, err)
   encoded, readErr := io.ReadAll(response.Body)
   closeErr := response.Body.Close()
   require.NoError(t, readErr)
   require.NoError(t, closeErr)
   require.Contains(t, string(encoded), ` + "`" + `"code":-32602` + "`" + `)
   require.Equal(t, before, calls.Load())
  })
 }
 request, err := http.NewRequestWithContext(t.Context(), http.MethodPost, peer.URL + "/prefix/organizations/blue/rpc", strings.NewReader(` + "`" + `{"jsonrpc":"2.0","id":"stream","method":"tick"}` + "`" + `))
 require.NoError(t, err)
 request.Header.Set("Content-Type", "application/json")
 request.Header.Set("Accept", "text/event-stream")
 before := calls.Load()
 response, err := peer.Client().Do(request)
 require.NoError(t, err)
 encoded, readErr := io.ReadAll(response.Body)
 closeErr := response.Body.Close()
 require.NoError(t, readErr)
 require.NoError(t, closeErr)
 require.Equal(t, before + 1, calls.Load())
 require.Contains(t, string(encoded), ` + "`" + `"params":["blue"]` + "`" + `)
 require.Contains(t, string(encoded), ` + "`" + `"id":"stream"` + "`" + `)
 // The generated streaming client uses the same URL and receives the typed event.
 before = calls.Load()
 streamed, err := client.Tick()(t.Context(), &genroute.TickPayload{OrganizationID: "blue"})
 require.NoError(t, err)
 stream, ok := streamed.(interface {
  Recv() (string, error)
  Close() error
 })
 require.True(t, ok)
 value, err := stream.Recv()
 require.NoError(t, err)
 require.Equal(t, "blue", value)
 _, err = stream.Recv()
 require.ErrorIs(t, err, io.EOF)
 require.NoError(t, stream.Close())
 require.Equal(t, before + 1, calls.Load())
 require.Equal(t, "/prefix/organizations/blue/rpc", doer.path)
 require.NotContains(t, doer.body, "organization_id")
 require.NotContains(t, doer.body, ` + "`" + `"params"` + "`" + `)
 // The path-only encoder still rejects values outside the authored payload type.
 _, err = genclient.EncodeTextRequest(goahttp.RequestEncoder)(httptest.NewRequest(http.MethodPost, "/rpc", nil), "wrong type")
 require.Error(t, err)
}
`
