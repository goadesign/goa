// This file renders a JSON-RPC SSE server and proves that streaming methods
// require request IDs before any service code runs.
package codegen_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestGeneratedSSERequiresRequestID checks both rejected missing IDs and a
// valid stream with a response ID.
func TestGeneratedSSERequiresRequestID(t *testing.T) {
	dir := renderParamsRuntimeModule(t)
	serverDir := filepath.Join(dir, "jsonrpc", "param_shapes", "server")
	require.NoError(t, os.WriteFile(
		filepath.Join(serverDir, "sse_notification_output_runtime_test.go"),
		[]byte(sseRequestIDRuntimeTest),
		0o600,
	))
	runParamsRuntimeTests(t, dir)
}

const sseRequestIDRuntimeTest = `package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	service "generated.local/gen/param_shapes"
	goahttp "goa.design/goa/v3/http"
	goa "goa.design/goa/v3/pkg"
)

func TestSSEMethodRejectsNotificationWithoutAResponse(t *testing.T) {
	response, calls, sends, encodes, reported := serveStreamText(
		` + "`" + `{"jsonrpc":"2.0","method":"stream_text"}` + "`" + `,
	)
	require.Zero(t, calls)
	require.NoError(t, sends)
	require.Zero(t, encodes)
	require.Len(t, reported, 1)
	require.Empty(t, response.Body.String())
}

func TestSSEMethodRejectsNullIDWithAResponse(t *testing.T) {
	response, calls, sends, encodes, reported := serveStreamText(
		` + "`" + `{"jsonrpc":"2.0","id":null,"method":"stream_text"}` + "`" + `,
	)
	require.Zero(t, calls)
	require.NoError(t, sends)
	require.Equal(t, 1, encodes)
	require.Empty(t, reported)
	require.Equal(t, "text/event-stream", response.Header().Get("Content-Type"))
	require.Contains(t, response.Body.String(), ` + "`" + `"id":null` + "`" + `)
	require.Contains(t, response.Body.String(), ` + "`" + `"code":-32600` + "`" + `)
	require.True(t, response.Flushed)
}

func TestSSERequestStillStreams(t *testing.T) {
	response, calls, sends, encodes, reported := serveStreamText(
		` + "`" + `{"jsonrpc":"2.0","id":"request-1","method":"stream_text"}` + "`" + `,
	)
	require.Equal(t, 1, calls)
	require.NoError(t, sends)
	require.Equal(t, 2, encodes)
	require.Empty(t, reported)
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

func serveStreamText(body string) (*httptest.ResponseRecorder, int, error, int, []error) {
	var calls int
	var sendErr error
	var encodes int
	encoder := func(ctx context.Context, writer http.ResponseWriter) goahttp.Encoder {
		encodes++
		return goahttp.ResponseEncoder(ctx, writer)
	}
	reported := make([]error, 0, 1)
	errhandler := func(_ context.Context, _ http.ResponseWriter, err error) {
		reported = append(reported, err)
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
	return response, calls, sendErr, encodes, reported
}
`
