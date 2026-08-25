// This file renders a JSON-RPC SSE server and proves that notifications run
// service streaming code without writing an HTTP response.
package codegen_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestGeneratedSSENotificationProducesNoHTTPOutput checks that a no-ID call
// still runs Send through the configured encoder while an ID call keeps the
// ordinary event stream response.
func TestGeneratedSSENotificationProducesNoHTTPOutput(t *testing.T) {
	dir := renderParamsRuntimeModule(t)
	serverDir := filepath.Join(dir, "jsonrpc", "param_shapes", "server")
	require.NoError(t, os.WriteFile(
		filepath.Join(serverDir, "sse_notification_output_runtime_test.go"),
		[]byte(sseNotificationOutputRuntimeTest),
		0o600,
	))
	runParamsRuntimeTests(t, dir)
}

const sseNotificationOutputRuntimeTest = `package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	service "generated.local/gen/param_shapes"
	goahttp "goa.design/goa/v3/http"
	goa "goa.design/goa/v3/pkg"
)

func TestSSENotificationRunsServiceWithoutHTTPOutput(t *testing.T) {
	response, calls, sends, encodes := serveStreamText(
		` + "`" + `{"jsonrpc":"2.0","method":"stream_text"}` + "`" + `,
	)
	require.Equal(t, 1, calls)
	require.NoError(t, sends)
	require.Equal(t, 1, encodes)
	require.Empty(t, response.Header())
	require.Empty(t, response.Body.String())
	require.False(t, response.Flushed)
}

func TestSSERequestStillStreams(t *testing.T) {
	response, calls, sends, encodes := serveStreamText(
		` + "`" + `{"jsonrpc":"2.0","id":"request-1","method":"stream_text"}` + "`" + `,
	)
	require.Equal(t, 1, calls)
	require.NoError(t, sends)
	require.Equal(t, 2, encodes)
	require.Equal(t, "text/event-stream", response.Header().Get("Content-Type"))
	events := strings.Split(strings.TrimSpace(response.Body.String()), "\n\n")
	require.Len(t, events, 2)
	require.JSONEq(t,
		` + "`" + `{"jsonrpc":"2.0","method":"stream_text","params":["value"]}` + "`" + `,
		strings.TrimPrefix(events[0], "data: "),
	)
	require.JSONEq(t,
		` + "`" + `{"jsonrpc":"2.0","id":"request-1","result":null}` + "`" + `,
		strings.TrimPrefix(events[1], "data: "),
	)
	require.True(t, response.Flushed)
}

func TestSSENotificationReportsFailuresWithoutHTTPOutput(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		serviceError error
		wantCalls    int
	}{
		{
			name:      "decode failure",
			body:      ` + "`" + `{"jsonrpc":"2.0","method":"required_resume"}` + "`" + `,
		},
		{
			name:         "service failure",
			body:         ` + "`" + `{"jsonrpc":"2.0","method":"stream_text"}` + "`" + `,
			serviceError: errors.New("service failed"),
			wantCalls:    1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response, calls, reports, writeErrors := serveSSEFailure(test.body, test.serviceError)
			require.Equal(t, test.wantCalls, calls)
			require.Equal(t, 1, reports)
			require.Empty(t, writeErrors)
			require.Empty(t, response.Header())
			require.Empty(t, response.Body.String())
			require.False(t, response.Flushed)
		})
	}
}

func TestUnaryNotificationReportsDecodeFailureWithoutHTTPOutput(t *testing.T) {
	response, calls, reports, writeErrors := serveTextNotification(
		` + "`" + `{"jsonrpc":"2.0","method":"text","params":{}}` + "`" + `,
		nil,
	)
	require.Zero(t, calls)
	require.Equal(t, 1, reports)
	require.Empty(t, writeErrors)
	require.Empty(t, response.Header())
	require.Empty(t, response.Body.String())
}

func TestUnaryNotificationReportsServiceFailureWithoutHTTPOutput(t *testing.T) {
	response, calls, reports, writeErrors := serveTextNotification(
		` + "`" + `{"jsonrpc":"2.0","method":"text","params":["value"]}` + "`" + `,
		errors.New("service failed"),
	)
	require.Equal(t, 1, calls)
	require.Equal(t, 1, reports)
	require.Empty(t, writeErrors)
	require.Empty(t, response.Header())
	require.Empty(t, response.Body.String())
}

func TestSuccessfulUnaryNotificationDoesNotReportAnError(t *testing.T) {
	response, calls, reports, writeErrors := serveTextNotification(
		` + "`" + `{"jsonrpc":"2.0","method":"text","params":["value"]}` + "`" + `,
		nil,
	)
	require.Equal(t, 1, calls)
	require.Zero(t, reports)
	require.Empty(t, writeErrors)
	require.Empty(t, response.Header())
	require.Empty(t, response.Body.String())
}

func serveStreamText(body string) (*httptest.ResponseRecorder, int, error, int) {
	var calls int
	var sendErr error
	var encodes int
	encoder := func(ctx context.Context, writer http.ResponseWriter) goahttp.Encoder {
		encodes++
		return goahttp.ResponseEncoder(ctx, writer)
	}
	errhandler := func(context.Context, http.ResponseWriter, error) {
	}
	server := &Server{
		StreamText: NewStreamTextHandler(
			goa.Endpoint(func(_ context.Context, value any) (any, error) {
				calls++
				input := value.(*service.StreamTextEndpointInput)
				sendErr = input.Stream.Send("value")
				return nil, nil
			}),
			goahttp.NewMuxer(),
			goahttp.RequestDecoder,
			encoder,
			errhandler,
		),
		decoder:    goahttp.RequestDecoder,
		encoder:    encoder,
		errhandler: errhandler,
	}
	request := httptest.NewRequest(http.MethodPost, "/rpc", strings.NewReader(body))
	request.Header.Set("Accept", "text/event-stream")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	return response, calls, sendErr, encodes
}

func serveTextNotification(body string, endpointErr error) (*httptest.ResponseRecorder, int, int, []error) {
	var calls int
	var reports int
	var writeErrors []error
	errhandler := func(_ context.Context, writer http.ResponseWriter, _ error) {
		reports++
		writer.Header().Set("X-Error", "reported")
		writer.WriteHeader(http.StatusInternalServerError)
		if _, err := writer.Write([]byte("reported")); err != nil {
			writeErrors = append(writeErrors, err)
		}
	}
	server := &Server{
		Text: NewTextHandler(
			goa.Endpoint(func(context.Context, any) (any, error) {
				calls++
				return nil, endpointErr
			}),
			goahttp.NewMuxer(),
			goahttp.RequestDecoder,
			goahttp.ResponseEncoder,
			errhandler,
		),
		decoder:    goahttp.RequestDecoder,
		encoder:    goahttp.ResponseEncoder,
		errhandler: errhandler,
	}
	request := httptest.NewRequest(http.MethodPost, "/rpc", strings.NewReader(body))
	request.Header.Set("Accept", "application/json")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	return response, calls, reports, writeErrors
}

func serveSSEFailure(body string, endpointErr error) (*httptest.ResponseRecorder, int, int, []error) {
	var calls int
	var reports int
	var writeErrors []error
	errhandler := func(_ context.Context, writer http.ResponseWriter, _ error) {
		reports++
		writer.Header().Set("X-Error", "reported")
		writer.WriteHeader(http.StatusInternalServerError)
		if _, err := writer.Write([]byte("reported")); err != nil {
			writeErrors = append(writeErrors, err)
		}
	}
	endpoint := goa.Endpoint(func(context.Context, any) (any, error) {
		calls++
		return nil, endpointErr
	})
	server := &Server{
		decoder:    goahttp.RequestDecoder,
		encoder:    goahttp.ResponseEncoder,
		errhandler: errhandler,
	}
	if endpointErr == nil {
		server.RequiredResume = NewRequiredResumeHandler(
			endpoint,
			goahttp.NewMuxer(),
			goahttp.RequestDecoder,
			goahttp.ResponseEncoder,
			errhandler,
		)
	} else {
		server.StreamText = NewStreamTextHandler(
			endpoint,
			goahttp.NewMuxer(),
			goahttp.RequestDecoder,
			goahttp.ResponseEncoder,
			errhandler,
		)
	}
	request := httptest.NewRequest(http.MethodPost, "/rpc", strings.NewReader(body))
	request.Header.Set("Accept", "text/event-stream")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	return response, calls, reports, writeErrors
}
`
