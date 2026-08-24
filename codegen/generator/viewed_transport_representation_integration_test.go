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

// TestGeneratedJSONRPCViewedServiceNameCompiles checks that the selected-view
// value does not hide the generated service package with the same Go name.
func TestGeneratedJSONRPCViewedServiceNameCompiles(t *testing.T) {
	dir := generateViewedTransportModule(t, viewedJSONRPCQualifierCollisionDSL)
	runGeneratedPackageTests(t, dir, "./jsonrpc/viewed/client")
}

// TestGeneratedJSONRPCSSEViewedRepresentation checks that JSON-RPC SSE pairs
// every view name with its matching body and rebuilds service results from
// notifications before the terminal response ends the stream.
func TestGeneratedJSONRPCSSEViewedRepresentation(t *testing.T) {
	dir := generateViewedTransportModule(t, viewedJSONRPCSSEDSL)
	writeGeneratedContractTest(t, dir, filepath.Join("jsonrpc", "jsonrpcsse", "client"), jsonRPCViewedSSEClientTest)
	runGeneratedPackageTests(t, dir, "./jsonrpc/jsonrpcsse/client")
}

// TestGeneratedJSONRPCSSEServerEmitsViewedRepresentation checks that SSE
// notifications contain the same view name and body that clients read.
// Methods with one fixed view contain only the body.
func TestGeneratedJSONRPCSSEServerEmitsViewedRepresentation(t *testing.T) {
	dir := generateViewedTransportModule(t, viewedJSONRPCSSEDSL)
	writeGeneratedContractTest(t, dir, filepath.Join("jsonrpc", "jsonrpcsse", "server"), jsonRPCViewedSSEServerTest)
	runGeneratedPackageTests(t, dir, "./jsonrpc/jsonrpcsse/server")
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

// viewedJSONRPCQualifierCollisionDSL uses the service name that previously
// matched the local selected-view value in the generated decoder.
func viewedJSONRPCQualifierCollisionDSL() {
	event := viewedResultType()
	dsl.Service("viewed", func() {
		dsl.JSONRPC(func() {
			dsl.POST("/rpc")
		})
		dsl.Method("fetch", func() {
			dsl.Result(event)
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
