// This file renders a JSON-RPC service into a temporary module and runs the
// generated unary and streaming error paths. It verifies that error names,
// codes, and designed response bodies survive the complete server/client flow.
package codegen

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

	goacodegen "goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

// TestGeneratedDesignedErrorEnvelope renders and runs both JSON-RPC transports
// with errors that share codes but have different names and response bodies.
func TestGeneratedDesignedErrorEnvelope(t *testing.T) {
	dir, plan := renderDesignedErrorRuntimeModule(t)
	testutil.AssertGo(
		t,
		"testdata/golden/designed_error_unary_client.go.golden",
		codegenSection(t, plan.ClientFiles(), "encode_decode.go", "jsonrpc-response-decoder", "func DecodeRunResponse"),
	)
	streamClient := codegenSection(t, plan.ClientFiles(), "stream.go", "jsonrpc-sse-client-stream", "decodeError(")
	require.Contains(t, streamClient, "decodeError(response.Error, response.ID == nil)")
	require.Contains(t, streamClient, "decodeError(response *jsonrpc.RawErrorResponse, nullID bool)")
	testutil.AssertGo(
		t,
		"testdata/golden/designed_error_unary_server.go.golden",
		codegenSection(t, plan.ServerFiles(), "server.go", "jsonrpc-server-handler-init", "func NewRunHandler"),
	)
	clientDir := filepath.Join(dir, "jsonrpc", "error_service", "client")
	serverDir := filepath.Join(dir, "jsonrpc", "error_service", "server")
	require.NoError(t, os.WriteFile(filepath.Join(clientDir, "designed_error_runtime_test.go"), []byte(designedErrorClientRuntimeTest), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(serverDir, "designed_error_runtime_test.go"), []byte(designedErrorServerRuntimeTest), 0o600))

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "test", "-mod=mod", "./jsonrpc/error_service/client", "./jsonrpc/error_service/server")
	command.Dir = dir
	command.Env = append(os.Environ(), "GOWORK=off")
	output, err := command.CombinedOutput()
	require.NoError(t, err, string(output))
}

// renderDesignedErrorRuntimeModule renders one service whose errors exercise
// shared codes, protocol-code collisions, renamed fields, and empty bodies.
func renderDesignedErrorRuntimeModule(t *testing.T) (string, *Plan) {
	t.Helper()
	root := expr.RunDSL(t, designedErrorRuntimeDSL)
	generation, err := goacodegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	httpPlans, err := httpcodegen.NewJSONRPCPlans(generation, httpcodegen.PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	jsonPlans, err := NewPlans(generation, PlanInput{
		Root:    root,
		Service: servicePlan,
		HTTP:    httpPlans[0],
	})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, httpPlans[0].Link())
	require.NoError(t, jsonPlans[0].Link())

	files, err := service.Files(servicePlan)
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
	module := fmt.Sprintf(
		"module generated.local/gen\n\ngo 1.25\n\nrequire goa.design/goa/v3 v3.0.0\n\nreplace goa.design/goa/v3 => %s\n",
		repository,
	)
	require.NoError(t, os.WriteFile(filepath.Join(moduleDir, "go.mod"), []byte(module), 0o600))
	return moduleDir, jsonPlans[0]
}

// codegenSection returns the generated method section identified by its
// declaration instead of depending on design order.
func codegenSection(t *testing.T, files []*goacodegen.File, fileName, sectionName, declaration string) string {
	t.Helper()
	for _, file := range files {
		if filepath.Base(file.Path) != fileName {
			continue
		}
		for _, section := range file.Section(sectionName) {
			source := goacodegen.SectionCode(t, section)
			if strings.Contains(source, declaration) {
				return source
			}
		}
	}
	require.Fail(t, "generated section not found", "%s in %s", declaration, fileName)
	return ""
}

// designedErrorRuntimeDSL declares errors whose code alone is deliberately
// insufficient to identify the designed service error.
func designedErrorRuntimeDSL() {
	firstProblem := dsl.Type("FirstProblem", func() {
		dsl.Attribute("detail", dsl.String, func() {
			dsl.MinLength(3)
		})
		dsl.Attribute("private", dsl.String)
		dsl.Required("detail")
	})
	secondProblem := dsl.Type("SecondProblem", func() {
		dsl.Attribute("detail", dsl.String)
		dsl.Required("detail")
	})
	internalProblem := dsl.Type("InternalProblem", func() {
		dsl.Attribute("detail", dsl.String)
		dsl.Required("detail")
	})
	selectedProblem := dsl.Type("SelectedProblem", func() {
		dsl.Attribute("detail", dsl.String)
		dsl.Required("detail")
	})
	parseProblem := dsl.Type("ParseProblem", func() {
		dsl.Attribute("detail", dsl.String)
		dsl.Required("detail")
	})
	invalidProblem := dsl.Type("InvalidProblem", func() {
		dsl.Attribute("detail", dsl.String)
		dsl.Required("detail")
	})
	emptyProblem := dsl.Type("EmptyProblem", func() {
		dsl.Attribute("detail", dsl.String)
	})

	dsl.Service("Error Service", func() {
		dsl.JSONRPC(func() {
			dsl.POST("/rpc")
		})
		dsl.Method("run", func() {
			dsl.Error("first", firstProblem)
			dsl.Error("second", secondProblem)
			dsl.Error("internal", internalProblem)
			dsl.Error("selected", selectedProblem)
			dsl.Error("parse", parseProblem)
			dsl.Error("invalid", invalidProblem)
			dsl.Error("empty", emptyProblem)
			dsl.JSONRPC(func() {
				dsl.Response("first", 7001, func() {
					dsl.Body(func() {
						dsl.Attribute("detail:message")
					})
				})
				dsl.Response("second", 7001)
				dsl.Response("internal", func() {})
				dsl.Response("selected", 7003, func() {
					dsl.Body("detail")
				})
				dsl.Response("parse", -32700)
				dsl.Response("invalid", -32600)
				dsl.Response("empty", 7002, func() {
					dsl.Body(dsl.Empty)
				})
			})
		})
		dsl.Method("watch", func() {
			dsl.StreamingResult(dsl.String)
			dsl.Error("first", firstProblem)
			dsl.Error("second", secondProblem)
			dsl.Error("selected", selectedProblem)
			dsl.Error("parse", parseProblem)
			dsl.Error("invalid", invalidProblem)
			dsl.JSONRPC(func() {
				dsl.Response("first", 7001, func() {
					dsl.Body(func() {
						dsl.Attribute("detail:message")
					})
				})
				dsl.Response("second", 7001)
				dsl.Response("selected", 7003, func() {
					dsl.Body("detail")
				})
				dsl.Response("parse", -32700)
				dsl.Response("invalid", -32600)
				dsl.ServerSentEvents()
			})
		})
	})
}

const designedErrorClientRuntimeTest = `package client

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	service "generated.local/gen/error_service"
	goahttp "goa.design/goa/v3/http"
	"goa.design/goa/v3/jsonrpc"
)

func TestUnaryDesignedErrorsUseCodeAndName(t *testing.T) {
	tests := []struct {
		name string
		response string
		check func(*testing.T, error)
	}{
		{
			name: "first shared code",
			response: ` + "`" + `{"jsonrpc":"2.0","id":"request-1","error":{"code":7001,"message":"first","data":{"name":"first","body":{"message":"visible"}}}}` + "`" + `,
			check: func(t *testing.T, err error) {
				var problem *service.FirstProblem
				require.ErrorAs(t, err, &problem)
				require.Equal(t, "visible", problem.Detail)
				require.Empty(t, problem.Private)
			},
		},
		{
			name: "second shared code",
			response: ` + "`" + `{"jsonrpc":"2.0","id":"request-1","error":{"code":7001,"message":"second","data":{"name":"second","body":{"detail":"other"}}}}` + "`" + `,
			check: func(t *testing.T, err error) {
				var problem *service.SecondProblem
				require.ErrorAs(t, err, &problem)
				require.Equal(t, "other", problem.Detail)
			},
		},
		{
			name: "designed internal code",
			response: ` + "`" + `{"jsonrpc":"2.0","id":"request-1","error":{"code":-32603,"message":"designed","data":{"name":"internal","body":{"detail":"known"}}}}` + "`" + `,
			check: func(t *testing.T, err error) {
				var problem *service.InternalProblem
				require.ErrorAs(t, err, &problem)
				require.Equal(t, "known", problem.Detail)
			},
		},
		{
			name: "empty body",
			response: ` + "`" + `{"jsonrpc":"2.0","id":"request-1","error":{"code":7002,"message":"empty","data":{"name":"empty","body":null}}}` + "`" + `,
			check: func(t *testing.T, err error) {
				var problem *service.EmptyProblem
				require.ErrorAs(t, err, &problem)
			},
		},
		{
			name: "selected primitive body",
			response: ` + "`" + `{"jsonrpc":"2.0","id":"request-1","error":{"code":7003,"message":"selected","data":{"name":"selected","body":"field"}}}` + "`" + `,
			check: func(t *testing.T, err error) {
				var problem *service.SelectedProblem
				require.ErrorAs(t, err, &problem)
				require.Equal(t, "field", problem.Detail)
			},
		},
		{
			name: "designed parse code with exact id",
			response: ` + "`" + `{"jsonrpc":"2.0","id":"request-1","error":{"code":-32700,"message":"parse","data":{"name":"parse","body":{"detail":"known"}}}}` + "`" + `,
			check: func(t *testing.T, err error) {
				var problem *service.ParseProblem
				require.ErrorAs(t, err, &problem)
				require.Equal(t, "known", problem.Detail)
			},
		},
		{
			name: "designed invalid request code with exact id",
			response: ` + "`" + `{"jsonrpc":"2.0","id":"request-1","error":{"code":-32600,"message":"invalid","data":{"name":"invalid","body":{"detail":"known"}}}}` + "`" + `,
			check: func(t *testing.T, err error) {
				var problem *service.InvalidProblem
				require.ErrorAs(t, err, &problem)
				require.Equal(t, "known", problem.Detail)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(test.response))}
			result, err := DecodeRunResponse(goahttp.ResponseDecoder, false)(response, "request-1")
			require.Nil(t, result)
			test.check(t, err)
		})
	}
}

func TestUnaryProtocolAndUnknownErrorsStayRaw(t *testing.T) {
	for _, test := range []struct {
		name string
		response string
		message string
	}{
		{name: "protocol internal error", response: ` + "`" + `{"jsonrpc":"2.0","id":"request-1","error":{"code":-32603,"message":"Internal error"}}` + "`" + `, message: "Internal error"},
		{name: "protocol parse error with null id", response: ` + "`" + `{"jsonrpc":"2.0","id":null,"error":{"code":-32700,"message":"Parse error"}}` + "`" + `, message: "Parse error"},
		{name: "unknown parse envelope with null id", response: ` + "`" + `{"jsonrpc":"2.0","id":null,"error":{"code":-32700,"message":"unknown","data":{"name":"other","body":{"detail":"known"}}}}` + "`" + `, message: "unknown"},
		{name: "protocol invalid request with null id", response: ` + "`" + `{"jsonrpc":"2.0","id":null,"error":{"code":-32600,"message":"Invalid request"}}` + "`" + `, message: "Invalid request"},
		{name: "malformed invalid envelope with null id", response: ` + "`" + `{"jsonrpc":"2.0","id":null,"error":{"code":-32600,"message":"malformed","data":[]}}` + "`" + `, message: "malformed"},
		{name: "unknown name", response: ` + "`" + `{"jsonrpc":"2.0","id":"request-1","error":{"code":7001,"message":"unknown","data":{"name":"other","body":{}}}}` + "`" + `, message: "unknown"},
		{name: "name not allowed for code", response: ` + "`" + `{"jsonrpc":"2.0","id":"request-1","error":{"code":7002,"message":"wrong code","data":{"name":"first","body":{"message":"visible"}}}}` + "`" + `, message: "wrong code"},
		{name: "missing body", response: ` + "`" + `{"jsonrpc":"2.0","id":"request-1","error":{"code":7001,"message":"missing body","data":{"name":"first"}}}` + "`" + `, message: "missing body"},
		{name: "malformed data", response: ` + "`" + `{"jsonrpc":"2.0","id":"request-1","error":{"code":7001,"message":"malformed","data":[]}}` + "`" + `, message: "malformed"},
		{name: "non-null empty body", response: ` + "`" + `{"jsonrpc":"2.0","id":"request-1","error":{"code":7002,"message":"non-null empty","data":{"name":"empty","body":{}}}}` + "`" + `, message: "non-null empty"},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(test.response))}
			result, err := DecodeRunResponse(goahttp.ResponseDecoder, false)(response, "request-1")
			require.Nil(t, result)
			var rpcError *jsonrpc.RawErrorResponse
			require.ErrorAs(t, err, &rpcError)
			require.Equal(t, test.message, rpcError.Message)
		})
	}
}

func TestUnaryDesignedErrorsRejectNullIDs(t *testing.T) {
	for _, responseBody := range []string{
		` + "`" + `{"jsonrpc":"2.0","id":null,"error":{"code":-32700,"message":"parse","data":{"name":"parse","body":{"detail":"known"}}}}` + "`" + `,
		` + "`" + `{"jsonrpc":"2.0","id":null,"error":{"code":-32600,"message":"invalid","data":{"name":"invalid","body":{"detail":"known"}}}}` + "`" + `,
		` + "`" + `{"jsonrpc":"2.0","id":null,"error":{"code":-32603,"message":"internal","data":{"name":"internal","body":{"detail":"known"}}}}` + "`" + `,
		` + "`" + `{"jsonrpc":"2.0","id":null,"error":{"code":7001,"message":"first","data":{"name":"first","body":{"message":"known"}}}}` + "`" + `,
	} {
		response := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(responseBody))}
		result, err := DecodeRunResponse(goahttp.ResponseDecoder, false)(response, "request-1")
		require.Nil(t, result)
		require.ErrorContains(t, err, "response id is null")
		var rpcError *jsonrpc.RawErrorResponse
		require.NotErrorAs(t, err, &rpcError)
	}
}

func TestUnaryDesignedErrorValidatesItsResponseBody(t *testing.T) {
	response := &http.Response{
		StatusCode: http.StatusOK,
		Body: io.NopCloser(strings.NewReader(` + "`" + `{"jsonrpc":"2.0","id":"request-1","error":{"code":7001,"message":"second","data":{"name":"second","body":{}}}}` + "`" + `)),
	}
	result, err := DecodeRunResponse(goahttp.ResponseDecoder, false)(response, "request-1")
	require.Nil(t, result)
	require.ErrorContains(t, err, "\"detail\" is missing from body")
	var rpcError *jsonrpc.RawErrorResponse
	require.NotErrorAs(t, err, &rpcError)
}

func TestSSEDesignedErrorUsesCodeAndName(t *testing.T) {
	response := &http.Response{
		StatusCode: http.StatusOK,
		Body: io.NopCloser(strings.NewReader("data: " + ` + "`" + `{"jsonrpc":"2.0","id":"request-1","error":{"code":7003,"message":"selected","data":{"name":"selected","body":"stream"}}}` + "`" + ` + "\n\n")),
	}
	_, err := NewWatchStream(response, goahttp.ResponseDecoder, "request-1").Recv()
	var problem *service.SelectedProblem
	require.ErrorAs(t, err, &problem)
	require.Equal(t, "stream", problem.Detail)
}

func TestSSEStandardCodesRequireTheExactRequestIDForDesignedErrors(t *testing.T) {
	for _, test := range []struct {
		name string
		response string
		check func(*testing.T, error)
	}{
		{
			name: "parse exact id",
			response: ` + "`" + `{"jsonrpc":"2.0","id":"request-1","error":{"code":-32700,"message":"parse","data":{"name":"parse","body":{"detail":"known"}}}}` + "`" + `,
			check: func(t *testing.T, err error) {
				var problem *service.ParseProblem
				require.ErrorAs(t, err, &problem)
			},
		},
		{
			name: "invalid request exact id",
			response: ` + "`" + `{"jsonrpc":"2.0","id":"request-1","error":{"code":-32600,"message":"invalid","data":{"name":"invalid","body":{"detail":"known"}}}}` + "`" + `,
			check: func(t *testing.T, err error) {
				var problem *service.InvalidProblem
				require.ErrorAs(t, err, &problem)
			},
		},
		{
			name: "parse protocol null id",
			response: ` + "`" + `{"jsonrpc":"2.0","id":null,"error":{"code":-32700,"message":"Parse error"}}` + "`" + `,
			check: requireRawError,
		},
		{
			name: "parse designed envelope null id",
			response: ` + "`" + `{"jsonrpc":"2.0","id":null,"error":{"code":-32700,"message":"parse","data":{"name":"parse","body":{"detail":"known"}}}}` + "`" + `,
			check: requireInvalidResponseID,
		},
		{
			name: "invalid request protocol null id",
			response: ` + "`" + `{"jsonrpc":"2.0","id":null,"error":{"code":-32600,"message":"Invalid request"}}` + "`" + `,
			check: requireRawError,
		},
		{
			name: "invalid request designed envelope null id",
			response: ` + "`" + `{"jsonrpc":"2.0","id":null,"error":{"code":-32600,"message":"invalid","data":{"name":"invalid","body":{"detail":"known"}}}}` + "`" + `,
			check: requireInvalidResponseID,
		},
		{
			name: "unknown parse envelope null id",
			response: ` + "`" + `{"jsonrpc":"2.0","id":null,"error":{"code":-32700,"message":"unknown","data":{"name":"other","body":{"detail":"known"}}}}` + "`" + `,
			check: requireRawError,
		},
		{
			name: "malformed invalid envelope null id",
			response: ` + "`" + `{"jsonrpc":"2.0","id":null,"error":{"code":-32600,"message":"malformed","data":[]}}` + "`" + `,
			check: requireRawError,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(strings.NewReader("data: " + test.response + "\n\n")),
			}
			_, err := NewWatchStream(response, goahttp.ResponseDecoder, "request-1").Recv()
			test.check(t, err)
		})
	}
}

// requireRawError checks that a protocol failure was not rebuilt as a designed error.
func requireRawError(t *testing.T, err error) {
	t.Helper()
	var response *jsonrpc.RawErrorResponse
	require.ErrorAs(t, err, &response)
}

// requireInvalidResponseID checks that a designed error did not accept a null ID.
func requireInvalidResponseID(t *testing.T, err error) {
	t.Helper()
	require.ErrorContains(t, err, "response id is null")
	var response *jsonrpc.RawErrorResponse
	require.NotErrorAs(t, err, &response)
}
`

const designedErrorServerRuntimeTest = `package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	service "generated.local/gen/error_service"
	goa "goa.design/goa/v3/pkg"
	goahttp "goa.design/goa/v3/http"
	"goa.design/goa/v3/jsonrpc"
)

func TestUnaryServerEncodesWrappedDesignedBody(t *testing.T) {
	private := "secret"
	endpoint := goa.Endpoint(func(context.Context, any) (any, error) {
		return nil, fmt.Errorf("wrapped: %w", &service.FirstProblem{Detail: "visible", Private: &private})
	})
	handler := NewRunHandler(endpoint, nil, goahttp.RequestDecoder, goahttp.ResponseEncoder, func(context.Context, http.ResponseWriter, error) {})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/rpc", nil)
	raw := &jsonrpc.RawRequest{JSONRPC: "2.0", Method: "run", ID: "request-1", HasID: true, HasMethod: true}

	require.NoError(t, handler(context.Background(), request, raw, recorder))
	var response jsonrpc.RawResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.NotNil(t, response.Error)
	require.Equal(t, 7001, response.Error.Code)
	require.JSONEq(t, ` + "`" + `{"name":"first","body":{"message":"visible"}}` + "`" + `, string(response.Error.Data))
}

func TestUnaryServerEncodesEmptyDesignedBody(t *testing.T) {
	detail := "not sent"
	endpoint := goa.Endpoint(func(context.Context, any) (any, error) {
		return nil, &service.EmptyProblem{Detail: &detail}
	})
	handler := NewRunHandler(endpoint, nil, goahttp.RequestDecoder, goahttp.ResponseEncoder, func(context.Context, http.ResponseWriter, error) {})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/rpc", nil)
	raw := &jsonrpc.RawRequest{JSONRPC: "2.0", Method: "run", ID: "request-1", HasID: true, HasMethod: true}

	require.NoError(t, handler(context.Background(), request, raw, recorder))
	var response jsonrpc.RawResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.NotNil(t, response.Error)
	require.JSONEq(t, ` + "`" + `{"name":"empty","body":null}` + "`" + `, string(response.Error.Data))
}

func TestUnaryServerEncodesSelectedDesignedBody(t *testing.T) {
	endpoint := goa.Endpoint(func(context.Context, any) (any, error) {
		return nil, fmt.Errorf("wrapped: %w", &service.SelectedProblem{Detail: "chosen"})
	})
	handler := NewRunHandler(endpoint, nil, goahttp.RequestDecoder, goahttp.ResponseEncoder, func(context.Context, http.ResponseWriter, error) {})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/rpc", nil)
	raw := &jsonrpc.RawRequest{JSONRPC: "2.0", Method: "run", ID: "request-1", HasID: true, HasMethod: true}

	require.NoError(t, handler(context.Background(), request, raw, recorder))
	var response jsonrpc.RawResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.NotNil(t, response.Error)
	require.JSONEq(t, ` + "`" + `{"name":"selected","body":"chosen"}` + "`" + `, string(response.Error.Data))
}

func TestSSEServerEncodesWrappedDesignedBody(t *testing.T) {
	endpoint := goa.Endpoint(func(context.Context, any) (any, error) {
		return nil, fmt.Errorf("wrapped: %w", &service.SelectedProblem{Detail: "stream"})
	})
	handler := NewWatchHandler(endpoint, nil, goahttp.RequestDecoder, goahttp.ResponseEncoder, func(context.Context, http.ResponseWriter, error) {})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/rpc", nil)
	raw := &jsonrpc.RawRequest{JSONRPC: "2.0", Method: "watch", ID: "request-1", HasID: true, HasMethod: true}

	require.NoError(t, handler(context.Background(), request, raw, recorder))
	data := strings.TrimSuffix(strings.TrimPrefix(recorder.Body.String(), "data: "), "\n\n")
	var response jsonrpc.RawResponse
	require.NoError(t, json.Unmarshal([]byte(data), &response))
	require.NotNil(t, response.Error)
	require.Equal(t, 7003, response.Error.Code)
	require.JSONEq(t, ` + "`" + `{"name":"selected","body":"stream"}` + "`" + `, string(response.Error.Data))
}
`
