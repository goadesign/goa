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

// TestGeneratedViewedUnaryResponseMetadata sends viewed results through a
// generated server and client. It checks a response with a JSON body and a
// response whose result is carried only by an HTTP header and cookie.
func TestGeneratedViewedUnaryResponseMetadata(t *testing.T) {
	dir := renderViewedResultRuntimeModule(t)
	writeViewedResultServerRuntimeTest(t, dir, "unary_metadata", unaryMetadataRuntimeTest)
	runViewedResultRuntimeTests(t, dir, "./jsonrpc/unary_metadata/server")
}

// TestGeneratedServerReturnsRequestBodyFailures makes reading and closing one
// request body fail and checks that the generated server reports both errors.
func TestGeneratedServerReturnsRequestBodyFailures(t *testing.T) {
	dir := renderViewedResultRuntimeModule(t)
	writeViewedResultServerRuntimeTest(t, dir, "unary_status", requestBodyFailureRuntimeTest)
	runViewedResultRuntimeTests(t, dir, "./jsonrpc/unary_status/server")
}

// TestGeneratedSSEDecodeErrorReturnsWriteFailure sends a request that omits a
// required parameter and makes writing the JSON-RPC error event fail. The
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
			dsl.JSONRPC(func() {
				dsl.ServerSentEvents(func() {})
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

	metadata := dsl.ResultType("application/vnd.unary-metadata", func() {
		dsl.TypeName("UnaryMetadata")
		dsl.Attribute("value", dsl.String)
		dsl.Attribute("etag", dsl.String)
		dsl.Attribute("session", dsl.String)
		dsl.Required("etag", "session")
		dsl.View("summary", func() {
			dsl.Attribute("value")
			dsl.Attribute("etag")
			dsl.Attribute("session")
		})
		dsl.View("detailed", func() {
			dsl.Attribute("value")
			dsl.Attribute("etag")
			dsl.Attribute("session")
		})
	})
	metadataOnly := dsl.ResultType("application/vnd.unary-metadata-only", func() {
		dsl.TypeName("UnaryMetadataOnly")
		dsl.Attribute("etag", dsl.String)
		dsl.Attribute("session", dsl.String)
		dsl.Required("etag", "session")
		dsl.View("default", func() {
			dsl.Attribute("etag")
			dsl.Attribute("session")
		})
	})
	dsl.Service("Unary Metadata", func() {
		dsl.JSONRPC(func() {
			dsl.POST("/metadata")
		})
		dsl.Method("fetch", func() {
			dsl.Result(metadata)
			dsl.JSONRPC(func() {
				dsl.Response(func() {
					dsl.Body("value")
					dsl.Header("etag:X-ETag")
					dsl.Cookie("session:SID")
				})
			})
		})
		dsl.Method("only", func() {
			dsl.Result(metadataOnly)
			dsl.JSONRPC(func() {
				dsl.Response(func() {
					dsl.Body(dsl.Empty)
					dsl.Header("etag:X-ETag")
					dsl.Cookie("session:SID")
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
	"errors"
	"io"
	"net/http"
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

func (body *trackedResponseBody) Read(buffer []byte) (int, error) {
	return body.reader.Read(buffer)
}

func (body *trackedResponseBody) Close() error {
	return body.closeErr
}

func TestUnaryViewedDecoderReceivesHTTPStatusOK(t *testing.T) {
	statuses := make([]int, 0, 3)
	decoder := func(response *http.Response) goahttp.Decoder {
		statuses = append(statuses, response.StatusCode)
		return goahttp.ResponseDecoder(response)
	}
	doer := doerFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body: io.NopCloser(strings.NewReader(
				` + "`" + `{"jsonrpc":"2.0","id":"1","result":{"view":"summary","body":{"label":"ready"}}}` + "`" + `,
			)),
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
	doer := doerFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body: &trackedResponseBody{
				reader: strings.NewReader(
					` + "`" + `{"jsonrpc":"2.0","id":"1","result":{"view":"summary","body":{"label":"ready"}}}` + "`" + `,
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
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

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

func TestSSEViewedDecoderReceivesHTTPStatusOK(t *testing.T) {
	statuses := make([]int, 0, 2)
	decoder := func(response *http.Response) goahttp.Decoder {
		statuses = append(statuses, response.StatusCode)
		return goahttp.ResponseDecoder(response)
	}
	doer := doerFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body: io.NopCloser(strings.NewReader(
				"event: notification\ndata: " + ` + "`" + `{"jsonrpc":"2.0","method":"watch","params":{"view":"summary","body":{"label":"ready"}}}` + "`" + ` + "\n\n" +
					"event: response\ndata: " + ` + "`" + `{"jsonrpc":"2.0","id":"1","result":null}` + "`" + ` + "\n\n",
			)),
		}, nil
	})
	client := NewClient("http", "example.test", doer, goahttp.RequestEncoder, decoder, false)
	raw, err := client.Watch()(context.Background(), nil)
	require.NoError(t, err)
	stream := raw.(*WatchStreamImpl)
	var serviceStream service.WatchClientStream = stream
	_, err = serviceStream.Recv()
	require.NoError(t, err)
	_, err = serviceStream.Recv()
	require.ErrorIs(t, err, io.EOF)
	require.NotEmpty(t, statuses)
	for _, status := range statuses {
		require.Equal(t, http.StatusOK, status)
	}
}

func TestSSETerminalErrorIsReturned(t *testing.T) {
	stream := NewWatchStream(&http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body: io.NopCloser(strings.NewReader(
			"event: error\ndata: " + ` + "`" + `{"jsonrpc":"2.0","id":"1","error":{"code":-32603,"message":"watch failed"}}` + "`" + ` + "\n\n",
		)),
	}, goahttp.ResponseDecoder)

	_, err := stream.Recv()
	var responseError *jsonrpc.ErrorResponse
	require.ErrorAs(t, err, &responseError)
	require.Equal(t, jsonrpc.InternalError, responseError.Code)
	require.Equal(t, "watch failed", responseError.Message)
}

func TestSSETerminalErrorPreservesCloseFailure(t *testing.T) {
	closeErr := errors.New("close failed")
	body := &failingResponseBody{
		reader: strings.NewReader(
			"event: error\ndata: " + ` + "`" + `{"jsonrpc":"2.0","id":"1","error":{"code":-32603,"message":"watch failed"}}` + "`" + ` + "\n\n",
		),
		closeErr: closeErr,
	}
	stream := NewWatchStream(&http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       body,
	}, goahttp.ResponseDecoder)

	_, err := stream.Recv()
	var responseError *jsonrpc.ErrorResponse
	require.ErrorAs(t, err, &responseError)
	require.Equal(t, "watch failed", responseError.Message)
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
		{"malformed notification", "event: notification\ndata: {\n\n", "failed to parse notification"},
		{"wrong method", "event: notification\ndata: " + ` + "`" + `{"jsonrpc":"2.0","method":"other","params":{}}` + "`" + ` + "\n\n", "received notification for JSON-RPC method"},
		{"unsupported event", "event: other\ndata: {}\n\n", "unsupported server-sent event type"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := &failingResponseBody{reader: strings.NewReader(test.event)}
			stream := NewWatchStream(&http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
				Body:       body,
			}, goahttp.ResponseDecoder)

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
	}, goahttp.ResponseDecoder)

	_, err := stream.Recv()
	require.ErrorIs(t, err, readErr)
	require.Equal(t, 1, body.closes)
	_, err = stream.Recv()
	require.Equal(t, io.EOF, err)
}

func TestSSEValidNotificationKeepsBodyOpen(t *testing.T) {
	body := &failingResponseBody{reader: strings.NewReader(
		"event: notification\ndata: " + ` + "`" + `{"jsonrpc":"2.0","method":"watch","params":{"view":"summary","body":{"label":"ready"}}}` + "`" + ` + "\n\n" +
			"event: response\ndata: " + ` + "`" + `{"jsonrpc":"2.0","id":"1","result":null}` + "`" + ` + "\n\n",
	)}
	stream := NewWatchStream(&http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       body,
	}, goahttp.ResponseDecoder)

	_, err := stream.Recv()
	require.NoError(t, err)
	require.Zero(t, body.closes)
	_, err = stream.Recv()
	require.Equal(t, io.EOF, err)
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
	}, goahttp.ResponseDecoder)

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
		_, err := DecodeFetchResponse(goahttp.ResponseDecoder, false)(response)
		require.NoError(t, err)
	}
}

func TestMappedObjectBodyRejectsMissingRequiredField(t *testing.T) {
	response := mappedResponse(` + "`" + `{"view":"summary","body":{}}` + "`" + `)
	_, err := DecodeFetchResponse(goahttp.ResponseDecoder, false)(response)
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
	return len(data), nil
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

	err := stream.sendSSEEvent(context.Background(), "notification", map[string]any{"value": "one"})
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
		{name: "event name", failWrite: 1},
		{name: "data label", failWrite: 2},
		{name: "encoded value", failWrite: 3},
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

			err := stream.sendSSEEvent(context.Background(), "notification", map[string]any{"value": "one"})
			if test.flushError != nil {
				require.ErrorIs(t, err, test.flushError)
			} else {
				require.ErrorIs(t, err, errWriteSSEError)
			}
			require.Equal(t, 1, writer.headerCalls)
		})
	}
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
}

func (s *lifecycleService) Watch(_ context.Context, _ *service.WatchPayload, stream service.WatchServerStream) error {
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
	notification := strings.Index(body, "event: notification")
	response := strings.Index(body, "event: response")
	require.NotEqual(t, -1, notification)
	require.Greater(t, response, notification)
	require.Contains(t, body, ` + "`" + `"method":"watch"` + "`" + `)
	require.Contains(t, body, ` + "`" + `"params":{"message":"ready"}` + "`" + `)
	require.Contains(t, body, ` + "`" + `"id":"request-1"` + "`" + `)
	require.Contains(t, body, ` + "`" + `"result":null` + "`" + `)
}

func TestSSEStreamWritesReturnedErrorAsTerminalResponse(t *testing.T) {
	body, reported := serveLifecycle(&lifecycleService{fail: true}, ` + "`" + `{"jsonrpc":"2.0","id":"request-1","method":"watch","params":{"topic":"alerts"}}` + "`" + `)
	require.Empty(t, reported)
	notification := strings.Index(body, "event: notification")
	response := strings.Index(body, "event: error")
	require.NotEqual(t, -1, notification)
	require.Greater(t, response, notification)
	require.Contains(t, body, ` + "`" + `"code":-32603` + "`" + `)
	require.Contains(t, body, ` + "`" + `"message":"watch failed"` + "`" + `)
}

func TestSSENotificationRequestHasNoTerminalResponse(t *testing.T) {
	body, reported := serveLifecycle(&lifecycleService{}, ` + "`" + `{"jsonrpc":"2.0","method":"watch","params":{"topic":"alerts"}}` + "`" + `)
	require.Empty(t, reported)
	require.Contains(t, body, "event: notification")
	require.NotContains(t, body, "event: response")
	require.NotContains(t, body, "event: error")
}

func TestSSENotificationServiceErrorHasNoTerminalResponse(t *testing.T) {
	body, reported := serveLifecycle(&lifecycleService{fail: true}, ` + "`" + `{"jsonrpc":"2.0","method":"watch","params":{"topic":"alerts"}}` + "`" + `)
	require.Empty(t, reported)
	require.Contains(t, body, "event: notification")
	require.NotContains(t, body, "event: response")
	require.NotContains(t, body, "event: error")
}

func TestSSENotificationDecodeErrorHasNoResponse(t *testing.T) {
	body, reported := serveLifecycle(&lifecycleService{}, ` + "`" + `{"jsonrpc":"2.0","method":"watch","params":{}}` + "`" + `)
	require.Empty(t, reported)
	require.Empty(t, body)
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
		require.Contains(t, body, "event: error")
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

const unaryMetadataRuntimeTest = `package server

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	service "generated.local/gen/unary_metadata"
	genclient "generated.local/gen/jsonrpc/unary_metadata/client"
	goahttp "goa.design/goa/v3/http"
)

type metadataService struct {
	fetchView string
}

func (s *metadataService) Fetch(context.Context) (*service.UnaryMetadata, string, error) {
	value := "record-1"
	return &service.UnaryMetadata{Value: &value, Etag: "etag-1", Session: "session-1"}, s.fetchView, nil
}

func (*metadataService) Only(context.Context) (*service.UnaryMetadataOnly, error) {
	return &service.UnaryMetadataOnly{Etag: "etag-2", Session: "session-2"}, nil
}

func TestViewedUnaryResponseCarriesBodyHeaderAndCookie(t *testing.T) {
	response := serve(t, &metadataService{fetchView: "summary"}, "fetch")
	require.Equal(t, "etag-1", response.Header.Get("X-ETag"))
	cookies := response.Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, "SID", cookies[0].Name)
	require.Equal(t, "session-1", cookies[0].Value)

	decoded, err := genclient.DecodeFetchResponse(goahttp.ResponseDecoder, false)(response)
	require.NoError(t, err)
	result := decoded.(*service.UnaryMetadata)
	require.Equal(t, "record-1", *result.Value)
	require.Equal(t, "etag-1", result.Etag)
	require.Equal(t, "session-1", result.Session)
}

func TestViewedUnaryResponseCarriesOnlyHeaderAndCookie(t *testing.T) {
	response := serve(t, &metadataService{}, "only")
	require.Equal(t, "etag-2", response.Header.Get("X-ETag"))
	cookies := response.Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, "SID", cookies[0].Name)
	require.Equal(t, "session-2", cookies[0].Value)

	decoded, err := genclient.DecodeOnlyResponse(goahttp.ResponseDecoder, false)(response)
	require.NoError(t, err)
	result := decoded.(*service.UnaryMetadataOnly)
	require.Equal(t, "etag-2", result.Etag)
	require.Equal(t, "session-2", result.Session)
}

func TestUnknownViewWritesNoSuccessMetadata(t *testing.T) {
	response := serve(t, &metadataService{fetchView: "unknown"}, "fetch")
	require.Empty(t, response.Header.Get("X-ETag"))
	require.Empty(t, response.Cookies())

	_, err := genclient.DecodeFetchResponse(goahttp.ResponseDecoder, false)(response)
	require.Error(t, err)
}

func serve(t *testing.T, svc *metadataService, method string) *http.Response {
	t.Helper()
	server := New(
		service.NewEndpoints(svc),
		goahttp.NewMuxer(),
		goahttp.RequestDecoder,
		goahttp.ResponseEncoder,
		func(context.Context, http.ResponseWriter, error) {},
	)
	body := []byte(` + "`" + `{"jsonrpc":"2.0","id":"1","method":"` + "`" + ` + method + ` + "`" + `"}` + "`" + `)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/metadata", bytes.NewReader(body)))
	require.Equal(t, http.StatusOK, recorder.Code)
	return recorder.Result()
}
`
