// This file renders JSON-RPC clients and servers into a temporary Go module.
// The generated tests call each client with an application-supplied decoder
// and send an invalid request to an SSE server, then inspect the response data
// and errors that the generated code gives the application.
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
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
	jsonrpccodegen "goa.design/goa/v3/jsonrpc/codegen"
)

// TestGeneratedViewedClientDecodersReceiveOKStatus renders a unary call and an
// SSE stream, then runs each generated client.
func TestGeneratedViewedClientDecodersReceiveOKStatus(t *testing.T) {
	dir := renderViewedResultRuntimeModule(t)
	writeViewedResultRuntimeTest(t, dir, "unary_status", unaryStatusRuntimeTest)
	writeViewedResultRuntimeTest(t, dir, "sse_status", sseStatusRuntimeTest)
	runViewedResultRuntimeTests(t, dir,
		"./jsonrpc/unary_status/client",
		"./jsonrpc/sse_status/client",
	)
}

// TestGeneratedSSELifecycle renders a result stream and checks every JSON-RPC
// message written when the service sends values and returns.
func TestGeneratedSSELifecycle(t *testing.T) {
	dir := renderViewedResultRuntimeModule(t)
	writeViewedResultServerRuntimeTest(t, dir, "sse_decode", sseLifecycleRuntimeTest)
	runViewedResultRuntimeTests(t, dir, "./jsonrpc/sse_decode/server")
}

// TestGeneratedMappedObjectBodyValidatesRequiredFields renders an explicit
// object response body, decodes both selected views, and checks the required
// field when the generated client receives the selected body.
func TestGeneratedMappedObjectBodyValidatesRequiredFields(t *testing.T) {
	dir := renderViewedResultRuntimeModule(t)
	writeViewedResultRuntimeTest(t, dir, "mapped_body", mappedBodyRuntimeTest)
	runViewedResultRuntimeTests(t, dir, "./jsonrpc/mapped_body/client")
}

// TestGeneratedServerReturnsRequestBodyFailures makes reading and closing one
// request body fail and checks that the generated server reports both errors.
func TestGeneratedServerReturnsRequestBodyFailures(t *testing.T) {
	dir := renderViewedResultRuntimeModule(t)
	writeViewedResultServerRuntimeTest(t, dir, "unary_status", requestBodyFailureRuntimeTest)
	runViewedResultRuntimeTests(t, dir, "./jsonrpc/unary_status/server")
}

// TestGeneratedSSEDecodeErrorReturnsWriteFailure sends a request that omits a
// required parameter and makes writing the JSON-RPC error message fail. The
// generated server must report that failure once without starting a new
// response.
func TestGeneratedSSEDecodeErrorReturnsWriteFailure(t *testing.T) {
	dir := renderViewedResultRuntimeModule(t)
	writeViewedResultServerRuntimeTest(t, dir, "sse_decode", sseDecodeErrorRuntimeTest)
	runViewedResultRuntimeTests(t, dir, "./jsonrpc/sse_decode/server")
}

// TestGeneratedSSERecvWithContextStopsBlockedRead renders an SSE client and
// checks that canceling a receive closes its response body and ends the stream.
func TestGeneratedSSERecvWithContextStopsBlockedRead(t *testing.T) {
	dir := renderViewedResultRuntimeModule(t)
	writeViewedResultRuntimeTest(t, dir, "sse_status", sseCancellationRuntimeTest)
	runViewedResultRuntimeTests(t, dir, "./jsonrpc/sse_status/client")
}

// renderViewedResultRuntimeModule writes the generated service and JSON-RPC
// client files used by these tests. It uses this Goa checkout and leaves the
// repository's generated files unchanged.
func renderViewedResultRuntimeModule(t *testing.T) string {
	t.Helper()
	root := expr.RunDSL(t, viewedResultRuntimeDSL)
	generation, err := goacodegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	httpPlans, err := httpcodegen.NewJSONRPCPlans(generation, httpcodegen.PlanInput{
		Root:    root,
		Service: servicePlan,
	})
	require.NoError(t, err)
	jsonPlans, err := jsonrpccodegen.NewPlans(generation, jsonrpccodegen.PlanInput{
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
	goMod := fmt.Sprintf("module generated.local/gen\n\ngo 1.25\n\nrequire goa.design/goa/v3 v3.0.0\n\nreplace goa.design/goa/v3 => %s\n", repository)
	require.NoError(t, os.WriteFile(filepath.Join(moduleDir, "go.mod"), []byte(goMod), 0o600))
	return moduleDir
}

// writeViewedResultRuntimeTest adds a client test to the temporary module.
func writeViewedResultRuntimeTest(t *testing.T, moduleDir, serviceName, source string) {
	t.Helper()
	dir := filepath.Join(moduleDir, "jsonrpc", serviceName, "client")
	require.NoError(t, os.MkdirAll(dir, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "viewed_result_runtime_test.go"), []byte(source), 0o600))
}

// writeViewedResultServerRuntimeTest adds a server test to the temporary
// module without writing into this repository's generated directories.
func writeViewedResultServerRuntimeTest(t *testing.T, moduleDir, serviceName, source string) {
	t.Helper()
	dir := filepath.Join(moduleDir, "jsonrpc", serviceName, "server")
	require.NoError(t, os.MkdirAll(dir, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sse_error_runtime_test.go"), []byte(source), 0o600))
}

// runViewedResultRuntimeTests runs only the generated packages named by
// patterns so each failure identifies the client call or server request that
// supplied unexpected data.
func runViewedResultRuntimeTests(t *testing.T, moduleDir string, patterns ...string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	args := append([]string{"test", "-mod=mod"}, patterns...)
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = moduleDir
	cmd.Env = append(os.Environ(), "GOWORK=off")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
}

// viewedResultRuntimeDSL defines the JSON-RPC methods rendered into the
// temporary module used by these tests.
func viewedResultRuntimeDSL() {
	result := viewedStatusResult()
	watchFailure := dsl.Type("WatchFailure", func() {
		dsl.Attribute("reason", dsl.String)
		dsl.Required("reason")
	})
	dsl.Service("Unary Status", func() {
		dsl.JSONRPC(func() {
			dsl.POST("/unary")
		})
		dsl.Method("fetch", func() {
			dsl.Result(result)
			dsl.JSONRPC(func() {})
		})
	})
	dsl.Service("SSE Status", func() {
		dsl.JSONRPC(func() {
			dsl.POST("/sse")
		})
		dsl.Method("watch", func() {
			dsl.StreamingResult(result)
			dsl.Error("watch_failed", watchFailure)
			dsl.JSONRPC(func() {
				dsl.Response("watch_failed", func() {
					dsl.Code(-32010)
				})
				dsl.ServerSentEvents(func() {
					dsl.SSEEventID("event_id")
					dsl.SSEEventType("event_type")
					dsl.SSEEventRetry("retry")
				})
			})
		})
	})
	dsl.Service("SSE Decode", func() {
		dsl.JSONRPC(func() {
			dsl.POST("/decode")
		})
		dsl.Method("watch", func() {
			dsl.Payload(func() {
				dsl.Attribute("topic", dsl.String)
				dsl.Required("topic")
			})
			dsl.StreamingResult(func() {
				dsl.Attribute("message", dsl.String)
				dsl.Required("message")
			})
			dsl.JSONRPC(func() {
				dsl.ServerSentEvents(func() {})
			})
		})
	})
	dsl.Service("Protocol", func() {
		dsl.JSONRPC(func() {
			dsl.POST("/protocol")
		})
		dsl.Method("ping", func() {
			dsl.JSONRPC(func() {})
		})
		dsl.Method("watch", func() {
			dsl.StreamingResult(dsl.String)
			dsl.JSONRPC(func() {
				dsl.ServerSentEvents(func() {})
			})
		})
	})
	mapped := dsl.ResultType("application/vnd.mapped-body", func() {
		dsl.TypeName("MappedBody")
		dsl.Attribute("id", func() {
			dsl.Attribute("value", dsl.String)
			dsl.Required("value")
		})
		dsl.Attribute("detail", dsl.String)
		dsl.Required("id")
		dsl.View("summary", func() {
			dsl.Attribute("id")
		})
		dsl.View("detailed", func() {
			dsl.Attribute("id")
			dsl.Attribute("detail")
		})
	})
	dsl.Service("Mapped Body", func() {
		dsl.JSONRPC(func() {
			dsl.POST("/mapped")
		})
		dsl.Method("fetch", func() {
			dsl.Result(mapped)
			dsl.JSONRPC(func() {
				dsl.Response(func() {
					dsl.Body("id")
				})
			})
		})
	})
}

// viewedStatusResult defines two views so each client must decode both the
// selected view name and its corresponding JSON body.
func viewedStatusResult() *expr.ResultTypeExpr {
	return dsl.ResultType("application/vnd.decoder-status", func() {
		dsl.TypeName("DecoderStatus")
		dsl.Attribute("label", dsl.String)
		dsl.Attribute("detail", dsl.String)
		dsl.Attribute("event_id", dsl.String, func() {
			dsl.Pattern(`^(|item-[0-9]+)$`)
		})
		dsl.Attribute("event_type", dsl.String, func() {
			dsl.Enum("", "update")
		})
		dsl.Attribute("retry", dsl.Int, func() {
			dsl.Maximum(10)
		})
		dsl.Required("label")
		dsl.View("summary", func() {
			dsl.Attribute("label")
		})
		dsl.View("detailed", func() {
			dsl.Attribute("label")
			dsl.Attribute("detail")
		})
	})
}

const unaryStatusRuntimeTest = `package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	goahttp "goa.design/goa/v3/http"
)

type doerFunc func(*http.Request) (*http.Response, error)

func (f doerFunc) Do(request *http.Request) (*http.Response, error) {
	return f(request)
}

type trackedResponseBody struct {
	reader   io.Reader
	closeErr error
}

type requestEnvelope struct {
	ID string ` + "`json:\"id\"`" + `
}

func (body *trackedResponseBody) Read(buffer []byte) (int, error) {
	return body.reader.Read(buffer)
}

func (body *trackedResponseBody) Close() error {
	return body.closeErr
}

func requestID(t *testing.T, request *http.Request) string {
	t.Helper()
	var envelope requestEnvelope
	require.NoError(t, json.NewDecoder(request.Body).Decode(&envelope))
	return envelope.ID
}

func responseBody(id, result string) io.ReadCloser {
	body := ` + "`" + `{"jsonrpc":"2.0","id":` + "`" + ` + strconv.Quote(id) + ` + "`" + `,"result":` + "`" + ` + result + "}"
	return io.NopCloser(strings.NewReader(body))
}

func TestUnaryViewedDecoderReceivesHTTPStatusOK(t *testing.T) {
	statuses := make([]int, 0, 3)
	decoder := func(response *http.Response) goahttp.Decoder {
		statuses = append(statuses, response.StatusCode)
		return goahttp.ResponseDecoder(response)
	}
	doer := doerFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body: responseBody(
				requestID(t, request),
				` + "`" + `{"view":"summary","body":{"label":"ready"}}` + "`" + `,
			),
		}, nil
	})
	client := NewClient("http", "example.test", doer, goahttp.RequestEncoder, decoder, false)
	_, err := client.Fetch()(context.Background(), nil)
	require.NoError(t, err)
	require.NotEmpty(t, statuses)
	for _, status := range statuses {
		require.Equal(t, http.StatusOK, status)
	}
}

func TestUnaryDecoderReturnsResponseCloseFailure(t *testing.T) {
	closeErr := errors.New("close failed")
	doer := doerFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body: &trackedResponseBody{
				reader: responseBody(
					requestID(t, request),
					` + "`" + `{"view":"summary","body":{"label":"ready"}}` + "`" + `,
				),
				closeErr: closeErr,
			},
		}, nil
	})
	client := NewClient("http", "example.test", doer, goahttp.RequestEncoder, goahttp.ResponseDecoder, false)
	_, err := client.Fetch()(context.Background(), nil)
	assertDecodingError(t, err)
	require.ErrorIs(t, err, closeErr)
}

func TestUnaryDecoderRejectsMismatchedResponseID(t *testing.T) {
	response := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body: responseBody(
			"response-id",
			` + "`" + `{"view":"summary","body":{"label":"ready"}}` + "`" + `,
		),
	}

	_, err := DecodeFetchResponse(goahttp.ResponseDecoder, false)(response, "request-id")
	var clientErr *goahttp.ClientError
	require.ErrorAs(t, err, &clientErr)
	require.Equal(t, "invalid_response", clientErr.Name)
	require.ErrorContains(t, err, ` + "`" + `response id "response-id" does not match request id "request-id"` + "`" + `)
}

func TestUnaryDecoderReturnsDecodeAndCloseFailures(t *testing.T) {
	decodeErr := errors.New("decode failed")
	closeErr := errors.New("close failed")
	doer := doerFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body: &trackedResponseBody{
				reader:   strings.NewReader("ignored"),
				closeErr: closeErr,
			},
		}, nil
	})
	decoder := func(*http.Response) goahttp.Decoder {
		return goahttp.EncodingFunc(func(any) error {
			return decodeErr
		})
	}
	client := NewClient("http", "example.test", doer, goahttp.RequestEncoder, decoder, false)

	_, err := client.Fetch()(context.Background(), nil)

	assertDecodingError(t, err)
	require.ErrorIs(t, err, decodeErr)
	require.ErrorIs(t, err, closeErr)
}

func assertDecodingError(t *testing.T, err error) {
	t.Helper()
	var clientErr *goahttp.ClientError
	require.ErrorAs(t, err, &clientErr)
	require.Equal(t, "decoding_error", clientErr.Name)
}
`

const sseStatusRuntimeTest = `package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/require"

	service "generated.local/gen/sse_status"
	goahttp "goa.design/goa/v3/http"
	"goa.design/goa/v3/jsonrpc"
)

type doerFunc func(*http.Request) (*http.Response, error)

func (f doerFunc) Do(request *http.Request) (*http.Response, error) {
	return f(request)
}

type failingResponseBody struct {
	reader   io.Reader
	readErr  error
	closeErr error
	closes   int
}

type requestEnvelope struct {
	ID string ` + "`json:\"id\"`" + `
}

func (body *failingResponseBody) Read(buffer []byte) (int, error) {
	if body.reader != nil {
		return body.reader.Read(buffer)
	}
	return 0, body.readErr
}

func (body *failingResponseBody) Close() error {
	body.closes++
	return body.closeErr
}

func requestID(t *testing.T, request *http.Request) string {
	t.Helper()
	var envelope requestEnvelope
	require.NoError(t, json.NewDecoder(request.Body).Decode(&envelope))
	return envelope.ID
}

func terminalResponse(id string) string {
	return "data: " +
		` + "`" + `{"jsonrpc":"2.0","id":` + "`" + ` + strconv.Quote(id) + ` + "`" + `,"result":null}` + "`" + ` + "\n\n"
}

func TestSSEViewedDecoderReceivesHTTPStatusOK(t *testing.T) {
	statuses := make([]int, 0, 2)
	decoder := func(response *http.Response) goahttp.Decoder {
		statuses = append(statuses, response.StatusCode)
		return goahttp.ResponseDecoder(response)
	}
	doer := doerFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body: io.NopCloser(strings.NewReader(
				"id: \nevent: \nretry: 0\ndata: " + ` + "`" + `{"jsonrpc":"2.0","method":"watch","params":{"view":"summary","body":{"label":"ready"}}}` + "`" + ` + "\n\n" +
					terminalResponse(requestID(t, request)),
			)),
		}, nil
	})
	client := NewClient("http", "example.test", doer, goahttp.RequestEncoder, decoder, false)
	raw, err := client.Watch()(context.Background(), nil)
	require.NoError(t, err)
	stream := raw.(*WatchStreamImpl)
	var serviceStream service.WatchClientStream = stream
	result, err := serviceStream.Recv()
	require.NoError(t, err)
	require.NotNil(t, result.EventID)
	require.Empty(t, *result.EventID)
	require.NotNil(t, result.EventType)
	require.Empty(t, *result.EventType)
	require.NotNil(t, result.Retry)
	require.Zero(t, *result.Retry)
	_, err = serviceStream.Recv()
	require.ErrorIs(t, err, io.EOF)
	require.NotEmpty(t, statuses)
	for _, status := range statuses {
		require.Equal(t, http.StatusOK, status)
	}
}

func TestSSEViewedDecoderValidatesMappedOuterFields(t *testing.T) {
	response := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body: io.NopCloser(strings.NewReader(
			"id: item-1\nevent: update\nretry: 11\ndata: " + ` + "`" + `{"jsonrpc":"2.0","method":"watch","params":{"view":"summary","body":{"label":"ready"}}}` + "`" + ` + "\n\n",
		)),
	}

	_, err := NewWatchStream(response, goahttp.ResponseDecoder, "request-1").Recv()

	require.ErrorContains(t, err, "lesser or equal than 10")
}

func TestSSETerminalErrorIsReturned(t *testing.T) {
	stream := NewWatchStream(&http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body: io.NopCloser(strings.NewReader(
			"data: " + ` + "`" + `{"jsonrpc":"2.0","id":"1","error":{"code":-32603,"message":"watch failed","data":{"reason":"overload"}}}` + "`" + ` + "\n\n",
		)),
	}, goahttp.ResponseDecoder, "1")

	_, err := stream.Recv()
	var response *jsonrpc.RawErrorResponse
	require.ErrorAs(t, err, &response)
	require.Equal(t, -32603, response.Code)
	require.Equal(t, "watch failed", response.Message)
	require.JSONEq(t, ` + "`" + `{"reason":"overload"}` + "`" + `, string(response.Data))
}

func TestSSETerminalDesignedErrorIsReturned(t *testing.T) {
	stream := NewWatchStream(&http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body: io.NopCloser(strings.NewReader(
			"data: " + ` + "`" + `{"jsonrpc":"2.0","id":"1","error":{"code":-32010,"message":"watch failed","data":{"name":"watch_failed","body":{"reason":"overload"}}}}` + "`" + ` + "\n\n",
		)),
	}, goahttp.ResponseDecoder, "1")

	_, err := stream.Recv()

	var failure *service.WatchFailure
	require.ErrorAs(t, err, &failure)
	require.Equal(t, "overload", failure.Reason)
}

func TestSSETerminalErrorPreservesCloseFailure(t *testing.T) {
	closeErr := errors.New("close failed")
	body := &failingResponseBody{
		reader: strings.NewReader(
			"data: " + ` + "`" + `{"jsonrpc":"2.0","id":"1","error":{"code":-32603,"message":"watch failed"}}` + "`" + ` + "\n\n",
		),
		closeErr: closeErr,
	}
	stream := NewWatchStream(&http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       body,
	}, goahttp.ResponseDecoder, "1")

	_, err := stream.Recv()
	var response *jsonrpc.RawErrorResponse
	require.ErrorAs(t, err, &response)
	require.Equal(t, -32603, response.Code)
	require.Equal(t, "watch failed", response.Message)
	require.ErrorIs(t, err, closeErr)
	require.Equal(t, 1, body.closes)
	_, err = stream.Recv()
	require.Equal(t, io.EOF, err)
}

func TestSSEInvalidEventClosesBody(t *testing.T) {
	tests := []struct {
		name  string
		event string
		error string
	}{
		{"malformed message", "data: {\n\n", "failed to parse JSON-RPC message"},
		{"notification with ID", "data: " + ` + "`" + `{"jsonrpc":"2.0","method":"watch","id":"unexpected","params":{}}` + "`" + ` + "\n\n", "invalid JSON-RPC notification"},
		{"wrong method", "data: " + ` + "`" + `{"jsonrpc":"2.0","method":"other","params":{}}` + "`" + ` + "\n\n", "received notification for JSON-RPC method"},
		{"response with params", "data: " + ` + "`" + `{"jsonrpc":"2.0","id":"1","params":{}}` + "`" + ` + "\n\n", "JSON-RPC response contains params"},
		{"message without method or response", "event: other\ndata: {}\n\n", ` + "`" + `response jsonrpc must be "2.0"` + "`" + `},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := &failingResponseBody{reader: strings.NewReader(test.event)}
			stream := NewWatchStream(&http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
				Body:       body,
			}, goahttp.ResponseDecoder, "1")

			_, err := stream.Recv()
			require.ErrorContains(t, err, test.error)
			require.Equal(t, 1, body.closes)
			_, err = stream.Recv()
			require.Equal(t, io.EOF, err)
		})
	}
}

func TestSSEReadFailureClosesBody(t *testing.T) {
	readErr := errors.New("read failed")
	body := &failingResponseBody{readErr: readErr}
	stream := NewWatchStream(&http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       body,
	}, goahttp.ResponseDecoder, "1")

	_, err := stream.Recv()
	require.ErrorIs(t, err, readErr)
	require.Equal(t, 1, body.closes)
	_, err = stream.Recv()
	require.Equal(t, io.EOF, err)
}

func TestSSEValidNotificationKeepsBodyOpen(t *testing.T) {
	body := &failingResponseBody{reader: strings.NewReader(
		"id: item-1\nevent: update\nretry: 0\ndata: " + ` + "`" + `{"jsonrpc":"2.0","method":"watch","params":{"view":"summary","body":{"label":"ready"}}}` + "`" + ` + "\n\n" +
			"data: " + ` + "`" + `{"jsonrpc":"2.0","id":"1","result":null}` + "`" + ` + "\n\n",
	)}
	stream := NewWatchStream(&http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       body,
	}, goahttp.ResponseDecoder, "1")

	_, err := stream.Recv()
	require.NoError(t, err)
	require.Zero(t, body.closes)
	_, err = stream.Recv()
	require.Equal(t, io.EOF, err)
	require.Equal(t, 1, body.closes)
}

func TestSSEAcceptsCRLFAndCRLineEndings(t *testing.T) {
	message := ` + "`" + `{"jsonrpc":"2.0","method":"watch","params":{"view":"summary","body":{"label":"ready"}}}` + "`" + `
	for _, ending := range []string{"\r\n", "\r"} {
		reader := io.Reader(strings.NewReader("data: " + message + ending + ending))
		if ending == "\r\n" {
			reader = iotest.OneByteReader(reader)
		}
		stream := NewWatchStream(&http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(reader),
		}, goahttp.ResponseDecoder, "1")

		result, err := stream.Recv()

		require.NoError(t, err)
		require.Equal(t, "ready", result.Label)
		require.NoError(t, stream.Close())
	}
}

func TestSSEPersistsOnlyValidEventIDsAndStripsInitialBOM(t *testing.T) {
	message := ` + "`" + `{"jsonrpc":"2.0","method":"watch","params":{"view":"summary","body":{"label":"ready"}}}` + "`" + `
	body := strings.Join([]string{
		"\uFEFFid: item-1\ndata: " + message + "\n\n",
		"data: " + message + "\n\n",
		"\uFEFFid: later\ndata: " + message + "\n\n",
		"id: ignored\x00id\ndata: " + message + "\n\n",
		"id:\ndata: " + message + "\n\n",
		"data: " + message + "\n\n",
	}, "")
	stream := NewWatchStream(&http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}, goahttp.ResponseDecoder, "1")

	for _, expected := range []string{"item-1", "item-1", "item-1", "item-1", "", ""} {
		result, err := stream.Recv()
		require.NoError(t, err)
		require.NotNil(t, result.EventID)
		require.Equal(t, expected, *result.EventID)
	}
	require.NoError(t, stream.Close())
}

func TestSSEEmptyEventKeepsOnlyItsID(t *testing.T) {
	message := ` + "`" + `{"jsonrpc":"2.0","method":"watch","params":{"view":"summary","body":{"label":"ready"}}}` + "`" + `
	body := "id: item-1\nevent: stale\nretry: 7\n\n" + "data: " + message + "\n\n"
	stream := NewWatchStream(&http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}, goahttp.ResponseDecoder, "1")

	result, err := stream.Recv()

	require.NoError(t, err)
	require.NotNil(t, result.EventID)
	require.Equal(t, "item-1", *result.EventID)
	require.Nil(t, result.EventType)
	require.Nil(t, result.Retry)
	require.Equal(t, "ready", result.Label)
}

func TestSSEViewedFailuresUseClientBoundaryErrors(t *testing.T) {
	tests := []struct {
		name      string
		event     string
		errorName string
	}{
		{
			name:      "representation decode",
			event:     "data: " + ` + "`" + `{"jsonrpc":"2.0","method":"watch","params":{"view":1,"body":{}}}` + "`" + ` + "\n\n",
			errorName: "decoding_error",
		},
		{
			name:      "missing view",
			event:     "data: " + ` + "`" + `{"jsonrpc":"2.0","method":"watch","params":{"body":{"label":"ready"}}}` + "`" + ` + "\n\n",
			errorName: "validation_error",
		},
		{
			name:      "outer retry decode",
			event:     "retry: 999999999999999999999999999999999999\ndata: " + ` + "`" + `{"jsonrpc":"2.0","method":"watch","params":{"view":"summary","body":{"label":"ready"}}}` + "`" + ` + "\n\n",
			errorName: "decoding_error",
		},
		{
			name:      "body validation",
			event:     "data: " + ` + "`" + `{"jsonrpc":"2.0","method":"watch","params":{"view":"summary","body":{}}}` + "`" + ` + "\n\n",
			errorName: "validation_error",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stream := NewWatchStream(&http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
				Body:       io.NopCloser(strings.NewReader(test.event)),
			}, goahttp.ResponseDecoder, "1")

			_, err := stream.Recv()

			var clientError *goahttp.ClientError
			require.ErrorAs(t, err, &clientError)
			require.Equal(t, test.errorName, clientError.Name)
			require.NotContains(t, err.Error(), "failed to decode result")
		})
	}
}

func TestSSEColonlessDataIsAnEmptyDataLine(t *testing.T) {
	stream := NewWatchStream(&http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader("data\n\n")),
	}, goahttp.ResponseDecoder, "1")

	_, err := stream.Recv()

	require.ErrorContains(t, err, "failed to parse JSON-RPC message")
}

func TestSSEDiscardsUnterminatedEventAtEOF(t *testing.T) {
	message := ` + "`" + `{"jsonrpc":"2.0","method":"watch","params":{"view":"summary","body":{"label":"ready"}}}` + "`" + `
	stream := NewWatchStream(&http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader("data: " + message)),
	}, goahttp.ResponseDecoder, "1")

	_, err := stream.Recv()

	require.ErrorIs(t, err, io.EOF)
}

func TestSSERetryUsesOnlyValidDigitFields(t *testing.T) {
	message := ` + "`" + `{"jsonrpc":"2.0","method":"watch","params":{"view":"summary","body":{"label":"ready"}}}` + "`" + `
	for _, retry := range []string{"", "+1", "-1", "١"} {
		stream := NewWatchStream(&http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader("retry: " + retry + "\ndata: " + message + "\n\n")),
		}, goahttp.ResponseDecoder, "1")

		result, err := stream.Recv()
		require.NoError(t, err)
		require.Nil(t, result.Retry)
		require.Equal(t, "ready", result.Label)
	}

	for _, fields := range []string{
		"retry: later\nretry: 7\n",
		"retry: 7\nretry: later\n",
	} {
		stream := NewWatchStream(&http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(fields + "data: " + message + "\n\n")),
		}, goahttp.ResponseDecoder, "1")

		result, err := stream.Recv()
		require.NoError(t, err)
		require.NotNil(t, result.Retry)
		require.Equal(t, 7, *result.Retry)
		require.Equal(t, "ready", result.Label)
	}
}

func TestSSETerminalResponseRejectsMismatchedID(t *testing.T) {
	body := &failingResponseBody{reader: strings.NewReader(terminalResponse("response-id"))}
	stream := NewWatchStream(&http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       body,
	}, goahttp.ResponseDecoder, "request-id")

	_, err := stream.Recv()
	require.ErrorContains(t, err, ` + "`" + `response id "response-id" does not match request id "request-id"` + "`" + `)
	require.Equal(t, 1, body.closes)
}

func TestSSEEndpointReturnsResponseBodyFailures(t *testing.T) {
	readErr := errors.New("read failed")
	closeErr := errors.New("close failed")
	doer := doerFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadGateway,
			Header:     make(http.Header),
			Body:       &failingResponseBody{readErr: readErr, closeErr: closeErr},
		}, nil
	})
	client := NewClient("http", "example.test", doer, goahttp.RequestEncoder, goahttp.ResponseDecoder, false)
	_, err := client.Watch()(context.Background(), nil)
	assertDecodingError(t, err)
	require.ErrorIs(t, err, readErr)
	require.ErrorIs(t, err, closeErr)
}

func TestSSEEndpointReturnsContentTypeAndCloseFailures(t *testing.T) {
	closeErr := errors.New("close failed")
	doer := doerFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       &failingResponseBody{closeErr: closeErr},
		}, nil
	})
	client := NewClient("http", "example.test", doer, goahttp.RequestEncoder, goahttp.ResponseDecoder, false)

	_, err := client.Watch()(context.Background(), nil)

	require.ErrorContains(t, err, "unexpected content type")
	require.ErrorIs(t, err, closeErr)
	assertDecodingError(t, err)
}

func TestSSEEndpointContentTypeRemainsPlainWhenCloseSucceeds(t *testing.T) {
	doer := doerFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       &failingResponseBody{},
		}, nil
	})
	client := NewClient("http", "example.test", doer, goahttp.RequestEncoder, goahttp.ResponseDecoder, false)

	_, err := client.Watch()(context.Background(), nil)

	require.EqualError(t, err, "unexpected content type: application/json (expected text/event-stream)")
	var clientErr *goahttp.ClientError
	require.NotErrorAs(t, err, &clientErr)
}

func TestSSEEndpointRequiresEventStreamContentType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
	}{
		{name: "missing"},
		{name: "malformed", contentType: "text/event-stream; charset"},
		{name: "prefix lookalike", contentType: "text/event-stream-extra"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := &failingResponseBody{}
			doer := doerFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{test.contentType}},
					Body:       body,
				}, nil
			})
			client := NewClient("http", "example.test", doer, goahttp.RequestEncoder, goahttp.ResponseDecoder, false)

			_, err := client.Watch()(context.Background(), nil)

			require.ErrorContains(t, err, "unexpected content type")
			require.Equal(t, 1, body.closes)
		})
	}
}

func TestSSEEndpointAllowsEventStreamParameters(t *testing.T) {
	body := &failingResponseBody{}
	doer := doerFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/event-stream; charset=utf-8"}},
			Body:       body,
		}, nil
	})
	client := NewClient("http", "example.test", doer, goahttp.RequestEncoder, goahttp.ResponseDecoder, false)

	raw, err := client.Watch()(context.Background(), nil)

	require.NoError(t, err)
	require.NoError(t, raw.(*WatchStreamImpl).Close())
	require.Equal(t, 1, body.closes)
}

func assertDecodingError(t *testing.T, err error) {
	t.Helper()
	var clientErr *goahttp.ClientError
	require.ErrorAs(t, err, &clientErr)
	require.Equal(t, "decoding_error", clientErr.Name)
}
`

const sseCancellationRuntimeTest = `package client

import (
	"context"
	"errors"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	goahttp "goa.design/goa/v3/http"
)

type blockingResponseBody struct {
	readStarted chan struct{}
	closed      chan struct{}
	readOnce    sync.Once
	closeOnce   sync.Once
	closeErr    error
	closes      int
}

func newBlockingResponseBody(closeErr error) *blockingResponseBody {
	return &blockingResponseBody{
		readStarted: make(chan struct{}),
		closed:      make(chan struct{}),
		closeErr:    closeErr,
	}
}

func (body *blockingResponseBody) Read([]byte) (int, error) {
	body.readOnce.Do(func() {
		close(body.readStarted)
	})
	<-body.closed
	return 0, io.ErrClosedPipe
}

func (body *blockingResponseBody) Close() error {
	body.closeOnce.Do(func() {
		body.closes++
		close(body.closed)
	})
	return body.closeErr
}

func TestRecvWithContextReturnsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	assertBlockedReceiveEndsWithContext(t, ctx, cancel, context.Canceled)
}

func TestRecvWithContextReturnsDeadlineExceeded(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	assertBlockedReceiveEndsWithContext(t, ctx, func() {}, context.DeadlineExceeded)
}

func assertBlockedReceiveEndsWithContext(t *testing.T, ctx context.Context, endContext func(), want error) {
	t.Helper()
	closeErr := errors.New("close failed")
	body := newBlockingResponseBody(closeErr)
	stream := NewWatchStream(&http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       body,
	}, goahttp.ResponseDecoder, "1")

	received := make(chan error, 1)
	go func() {
		_, err := stream.RecvWithContext(ctx)
		received <- err
	}()
	<-body.readStarted
	endContext()

	select {
	case err := <-received:
		require.ErrorIs(t, err, want)
		require.ErrorIs(t, err, closeErr)
	case <-time.After(time.Second):
		require.NoError(t, body.Close())
		err := <-received
		t.Fatalf("receive remained blocked after context ended; returned %v after closing body", err)
	}
	select {
	case <-body.closed:
	default:
		t.Fatal("receive returned without closing response body")
	}
	_, err := stream.Recv()
	require.ErrorIs(t, err, io.EOF)
	require.Equal(t, 1, body.closes)
}
`

const mappedBodyRuntimeTest = `package client

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	goahttp "goa.design/goa/v3/http"
	goa "goa.design/goa/v3/pkg"
)

func TestMappedObjectBodyAcceptsBothViews(t *testing.T) {
	for _, view := range []string{"summary", "detailed"} {
		response := mappedResponse(` + "`" + `{"view":"` + "`" + ` + view + ` + "`" + `","body":{"value":"record-1"}}` + "`" + `)
		_, err := DecodeFetchResponse(goahttp.ResponseDecoder, false)(response, "1")
		require.NoError(t, err)
	}
}

func TestMappedObjectBodyRejectsMissingRequiredField(t *testing.T) {
	response := mappedResponse(` + "`" + `{"view":"summary","body":{}}` + "`" + `)
	_, err := DecodeFetchResponse(goahttp.ResponseDecoder, false)(response, "1")
	var serviceError *goa.ServiceError
	require.ErrorAs(t, err, &serviceError)
	require.Equal(t, goa.MissingField, serviceError.Name)
	require.NotNil(t, serviceError.Field)
	require.Equal(t, "value", *serviceError.Field)
}

func mappedResponse(result string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body: io.NopCloser(strings.NewReader(
			` + "`" + `{"jsonrpc":"2.0","id":"1","result":` + "`" + ` + result + "}",
		)),
	}
}
`

const sseDecodeErrorRuntimeTest = `package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	service "generated.local/gen/sse_decode"
	goahttp "goa.design/goa/v3/http"
)

var errWriteSSEError = errors.New("write SSE error")
var errEncodeSSE = errors.New("encode SSE event")
var errFlushSSE = errors.New("flush SSE event")

type unusedService struct{}

func (*unusedService) Watch(context.Context, *service.WatchPayload, service.WatchServerStream) error {
	return nil
}

type failingResponseWriter struct {
	header      http.Header
	headerCalls int
}

type stepResponseWriter struct {
	header      http.Header
	headerCalls int
	writes      int
	failWrite   int
	flushError  error
	body        strings.Builder
}

func (writer *stepResponseWriter) Header() http.Header {
	return writer.header
}

func (writer *stepResponseWriter) WriteHeader(int) {
	writer.headerCalls++
}

func (writer *stepResponseWriter) Write(data []byte) (int, error) {
	writer.writes++
	if writer.writes == writer.failWrite {
		return 0, errWriteSSEError
	}
	return writer.body.Write(data)
}

func (writer *stepResponseWriter) FlushError() error {
	return writer.flushError
}

func (writer *failingResponseWriter) Header() http.Header {
	return writer.header
}

func (writer *failingResponseWriter) WriteHeader(int) {
	writer.headerCalls++
}

func (*failingResponseWriter) Write([]byte) (int, error) {
	return 0, errWriteSSEError
}

func TestSSERequestDecodeErrorReturnsWriteFailureOnce(t *testing.T) {
	reported := make([]error, 0, 1)
	server := New(
		service.NewEndpoints(&unusedService{}),
		goahttp.NewMuxer(),
		goahttp.RequestDecoder,
		goahttp.ResponseEncoder,
		func(_ context.Context, _ http.ResponseWriter, err error) {
			reported = append(reported, err)
		},
	)
	writer := &failingResponseWriter{header: make(http.Header)}
	request := httptest.NewRequest(http.MethodPost, "/decode", strings.NewReader(
		` + "`" + `{"jsonrpc":"2.0","id":"request-1","method":"watch","params":{}}` + "`" + `,
	))
	server.ServeHTTP(writer, request)

	require.Len(t, reported, 1)
	require.ErrorIs(t, reported[0], errWriteSSEError)
	require.Equal(t, 1, writer.headerCalls)
}

func TestSSEEventEncodesBeforeStartingResponse(t *testing.T) {
	writer := &stepResponseWriter{header: make(http.Header)}
	stream := &sseServerStream{
		w: writer,
		encoder: func(context.Context, http.ResponseWriter) goahttp.Encoder {
			return goahttp.EncodingFunc(func(any) error { return errEncodeSSE })
		},
	}

	err := stream.sendSSEEvent(context.Background(), map[string]any{"value": "one"}, nil, nil, nil)
	require.ErrorIs(t, err, errEncodeSSE)
	require.Zero(t, writer.headerCalls)
	require.Zero(t, writer.writes)
}

func TestSSEEventReturnsEveryWriteAndFlushError(t *testing.T) {
	tests := []struct {
		name       string
		failWrite  int
		flushError error
	}{
		{name: "data label", failWrite: 1},
		{name: "encoded value", failWrite: 2},
		{name: "data line ending", failWrite: 3},
		{name: "event ending", failWrite: 4},
		{name: "flush", flushError: errFlushSSE},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			writer := &stepResponseWriter{
				header:     make(http.Header),
				failWrite:  test.failWrite,
				flushError: test.flushError,
			}
			stream := &sseServerStream{w: writer, encoder: goahttp.ResponseEncoder}

			err := stream.sendSSEEvent(context.Background(), map[string]any{"value": "one"}, nil, nil, nil)
			if test.flushError != nil {
				require.ErrorIs(t, err, test.flushError)
			} else {
				require.ErrorIs(t, err, errWriteSSEError)
			}
			require.Equal(t, 1, writer.headerCalls)
		})
	}
}

func TestSSEEventPrefixesEveryEncodedDataLine(t *testing.T) {
	for _, encoded := range []string{
		"{\n  \"jsonrpc\": \"2.0\"\n}\n",
		"{\r  \"jsonrpc\": \"2.0\"\r}\r",
		"{\r\n  \"jsonrpc\": \"2.0\"\r\n}\r\n",
	} {
		writer := &stepResponseWriter{header: make(http.Header)}
		stream := &sseServerStream{
			w: writer,
			encoder: func(_ context.Context, target http.ResponseWriter) goahttp.Encoder {
				return goahttp.EncodingFunc(func(any) error {
					_, err := target.Write([]byte(encoded))
					return err
				})
			},
		}

		err := stream.sendSSEEvent(context.Background(), struct{}{}, nil, nil, nil)

		require.NoError(t, err)
		require.Equal(t, "data: {\ndata:   \"jsonrpc\": \"2.0\"\ndata: }\n\n", writer.body.String())
	}
}

func TestSSEEventRejectsInvalidOuterFieldsBeforeStartingResponse(t *testing.T) {
	tests := []struct {
		name      string
		id        *string
		eventType *string
		retry     *string
	}{
		{name: "id nul", id: stringPointer("one\x00two")},
		{name: "id line feed", id: stringPointer("one\ntwo")},
		{name: "event carriage return", eventType: stringPointer("one\rtwo")},
		{name: "retry plus", retry: stringPointer("+1")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			writer := &stepResponseWriter{header: make(http.Header)}
			stream := &sseServerStream{w: writer, encoder: goahttp.ResponseEncoder}

			err := stream.sendSSEEvent(context.Background(), struct{}{}, test.id, test.eventType, test.retry)

			require.Error(t, err)
			require.Zero(t, writer.headerCalls)
			require.Zero(t, writer.writes)
		})
	}
}

func stringPointer(value string) *string {
	return &value
}
`

const requestBodyFailureRuntimeTest = `package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	goahttp "goa.design/goa/v3/http"
)

type failingRequestBody struct {
	readErr  error
	closeErr error
}

type failingResponseWriter struct {
	*httptest.ResponseRecorder
	failAt   int
	writes   int
	writeErr error
}

func (writer *failingResponseWriter) Write(data []byte) (int, error) {
	writer.writes++
	if writer.writes == writer.failAt {
		return 0, writer.writeErr
	}
	return writer.ResponseRecorder.Write(data)
}

func (body *failingRequestBody) Read([]byte) (int, error) {
	return 0, body.readErr
}

func (body *failingRequestBody) Close() error {
	return body.closeErr
}

func TestServerReportsRequestReadAndCloseFailures(t *testing.T) {
	readErr := errors.New("read failed")
	closeErr := errors.New("close failed")
	var reported error
	server := &Server{
		errhandler: func(_ context.Context, _ http.ResponseWriter, err error) {
			reported = err
		},
	}
	request := httptest.NewRequest(http.MethodPost, "/unary", nil)
	request.Body = &failingRequestBody{readErr: readErr, closeErr: closeErr}

	server.handleHTTP(httptest.NewRecorder(), request)

	require.ErrorIs(t, reported, readErr)
	require.ErrorIs(t, reported, closeErr)
}

func TestBatchWriterReturnsOpeningDelimiterFailure(t *testing.T) {
	writeErr := errors.New("write failed")
	response := &failingResponseWriter{
		ResponseRecorder: httptest.NewRecorder(),
		failAt:           1,
		writeErr:         writeErr,
	}
	writer := &batchWriter{Writer: response}

	_, err := writer.Write([]byte("{\"jsonrpc\":\"2.0\",\"result\":null}"))

	require.ErrorIs(t, err, writeErr)
}

func TestServerReportsBatchClosingDelimiterFailure(t *testing.T) {
	writeErr := errors.New("write failed")
	var reported []error
	server := &Server{
		decoder: goahttp.RequestDecoder,
		encoder: goahttp.ResponseEncoder,
		errhandler: func(_ context.Context, _ http.ResponseWriter, err error) {
			reported = append(reported, err)
		},
	}
	request := httptest.NewRequest(
		http.MethodPost,
		"/unary",
		strings.NewReader("[{\"jsonrpc\":\"2.0\",\"id\":\"1\",\"method\":\"missing\"}]"),
	)
	response := &failingResponseWriter{
		ResponseRecorder: httptest.NewRecorder(),
		failAt:           3,
		writeErr:         writeErr,
	}

	server.handleHTTP(response, request)

	require.ErrorIs(t, errors.Join(reported...), writeErr)
}
`

const sseLifecycleRuntimeTest = `package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	service "generated.local/gen/sse_decode"
	goahttp "goa.design/goa/v3/http"
	goa "goa.design/goa/v3/pkg"
)

var errWatch = goa.NewServiceError(errors.New("watch failed"), "watch_failed", false, false, false)

type lifecycleService struct {
	fail bool
	calls int
}

func (s *lifecycleService) Watch(_ context.Context, _ *service.WatchPayload, stream service.WatchServerStream) error {
	s.calls++
	if err := stream.Send(&service.WatchResult{Message: "ready"}); err != nil {
		return err
	}
	if s.fail {
		return errWatch
	}
	return nil
}

func TestSSEStreamWritesNotificationThenNullCompletion(t *testing.T) {
	body, reported := serveLifecycle(&lifecycleService{}, ` + "`" + `{"jsonrpc":"2.0","id":"request-1","method":"watch","params":{"topic":"alerts"}}` + "`" + `)
	require.Empty(t, reported)
	notification := strings.Index(body, ` + "`" + `"method":"watch"` + "`" + `)
	response := strings.Index(body, ` + "`" + `"result":null` + "`" + `)
	require.NotEqual(t, -1, notification)
	require.Greater(t, response, notification)
	require.NotContains(t, body, "event:")
	require.Contains(t, body, ` + "`" + `"method":"watch"` + "`" + `)
	require.Contains(t, body, ` + "`" + `"params":{"message":"ready"}` + "`" + `)
	require.Contains(t, body, ` + "`" + `"id":"request-1"` + "`" + `)
	require.Contains(t, body, ` + "`" + `"result":null` + "`" + `)
}

func TestSSEStreamWritesReturnedErrorAsTerminalResponse(t *testing.T) {
	body, reported := serveLifecycle(&lifecycleService{fail: true}, ` + "`" + `{"jsonrpc":"2.0","id":"request-1","method":"watch","params":{"topic":"alerts"}}` + "`" + `)
	require.Empty(t, reported)
	notification := strings.Index(body, ` + "`" + `"method":"watch"` + "`" + `)
	response := strings.Index(body, ` + "`" + `"error":` + "`" + `)
	require.NotEqual(t, -1, notification)
	require.Greater(t, response, notification)
	require.NotContains(t, body, "event:")
	require.Contains(t, body, ` + "`" + `"code":-32603` + "`" + `)
	require.Contains(t, body, ` + "`" + `"message":"watch failed"` + "`" + `)
}

func TestSSEMethodRejectsMissingAndNullIDsBeforeDispatch(t *testing.T) {
	for _, request := range []string{
		` + "`" + `{"jsonrpc":"2.0","method":"watch","params":{"topic":"alerts"}}` + "`" + `,
		` + "`" + `{"jsonrpc":"2.0","id":null,"method":"watch","params":{"topic":"alerts"}}` + "`" + `,
	} {
		svc := &lifecycleService{}
		body, reported := serveLifecycle(svc, request)
		require.Empty(t, reported)
		require.Zero(t, svc.calls)
		require.Contains(t, body, "data: ")
		require.Contains(t, body, ` + "`" + `"id":null` + "`" + `)
		require.Contains(t, body, ` + "`" + `"code":-32600` + "`" + `)
	}
}

func TestSSEUnknownNotificationReceivesNoResponse(t *testing.T) {
	body, reported := serveLifecycle(&lifecycleService{}, ` + "`" + `{"jsonrpc":"2.0","method":"missing"}` + "`" + `)
	require.Empty(t, reported)
	require.Empty(t, body)
}

func TestSSEInvalidRequestsReceiveError(t *testing.T) {
	for _, request := range []string{
		` + "`" + `{"jsonrpc":"2.0"}` + "`" + `,
		` + "`" + `{"jsonrpc":"1.0","method":"watch","params":{"topic":"alerts"}}` + "`" + `,
	} {
		body, reported := serveLifecycle(&lifecycleService{}, request)
		require.Empty(t, reported)
		require.NotContains(t, body, "event:")
		require.Contains(t, body, "data: ")
		require.Contains(t, body, ` + "`" + `"id":null` + "`" + `)
		require.Contains(t, body, ` + "`" + `"code":-32600` + "`" + `)
	}
}

func serveLifecycle(svc *lifecycleService, body string) (string, []error) {
	reported := make([]error, 0, 1)
	server := New(
		service.NewEndpoints(svc),
		goahttp.NewMuxer(),
		goahttp.RequestDecoder,
		goahttp.ResponseEncoder,
		func(_ context.Context, _ http.ResponseWriter, err error) {
			reported = append(reported, err)
		},
	)
	writer := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/decode", strings.NewReader(body))
	server.ServeHTTP(writer, request)
	return writer.Body.String(), reported
}
`
