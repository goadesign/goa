// This file renders a JSON-RPC server and runs requests whose wire form decides
// whether the server returns one response, a batch, no body, or an event stream.
package codegen_test

import "testing"

// TestGeneratedServerFollowsJSONRPCRequestRules checks the request forms that
// determine whether the server sends one response, a batch, or an event stream.
func TestGeneratedServerFollowsJSONRPCRequestRules(t *testing.T) {
	dir := renderViewedResultRuntimeModule(t)
	writeViewedResultServerRuntimeTest(t, dir, "protocol", protocolRuntimeTest)
	runViewedResultRuntimeTests(t, dir, "./jsonrpc/protocol/server")
}

// TestGeneratedPureSSEServerClosesRequestBodies checks that an event-only
// server closes the body supplied by the HTTP server after the stream ends.
func TestGeneratedPureSSEServerClosesRequestBodies(t *testing.T) {
	dir := renderViewedResultRuntimeModule(t)
	writeViewedResultServerRuntimeTest(t, dir, "sse_decode", pureSSEBodyRuntimeTest)
	runViewedResultRuntimeTests(t, dir, "./jsonrpc/sse_decode/server")
}

const protocolRuntimeTest = `package server

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	goahttp "goa.design/goa/v3/http"
	goa "goa.design/goa/v3/pkg"
)

func TestRequestIDPresenceControlsResponses(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{name: "missing ID", body: ` + "`" + `{"jsonrpc":"2.0","method":"ping"}` + "`" + `},
		{name: "empty string ID", body: ` + "`" + `{"jsonrpc":"2.0","id":"","method":"ping"}` + "`" + `, want: ` + "`" + `{"jsonrpc":"2.0","id":"","result":null}` + "`" + `},
		{name: "null ID", body: ` + "`" + `{"jsonrpc":"2.0","id":null,"method":"ping"}` + "`" + `, want: ` + "`" + `{"jsonrpc":"2.0","id":null,"result":null}` + "`" + `},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := serveProtocol(test.body, "")
			if test.want == "" {
				require.Empty(t, response.Body.String())
				return
			}
			require.JSONEq(t, test.want, response.Body.String())
		})
	}
}

func TestRequestIDPresenceControlsErrors(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{name: "missing ID", body: ` + "`" + `{"jsonrpc":"2.0","method":"missing"}` + "`" + `},
		{name: "empty string ID", body: ` + "`" + `{"jsonrpc":"2.0","id":"","method":"missing"}` + "`" + `, want: ` + "`" + `{"jsonrpc":"2.0","id":"","error":{"code":-32601,"message":"Method not found"}}` + "`" + `},
		{name: "null ID", body: ` + "`" + `{"jsonrpc":"2.0","id":null,"method":"missing"}` + "`" + `, want: ` + "`" + `{"jsonrpc":"2.0","id":null,"error":{"code":-32601,"message":"Method not found"}}` + "`" + `},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := serveProtocol(test.body, "")
			if test.want == "" {
				require.Empty(t, response.Body.String())
				return
			}
			require.JSONEq(t, test.want, response.Body.String())
		})
	}
}

func TestBatchFormFollowsJSONWhitespaceAndEmptyArrayRules(t *testing.T) {
	response := serveProtocol(" \n\t["+` + "`" + `{"jsonrpc":"2.0","id":"one","method":"ping"}` + "`" + `+"]", "")
	require.JSONEq(t, ` + "`" + `[{"jsonrpc":"2.0","id":"one","result":null}]` + "`" + `, response.Body.String())

	response = serveProtocol("[]", "")
	require.JSONEq(t, ` + "`" + `{"jsonrpc":"2.0","id":null,"error":{"code":-32600,"message":"Invalid request"}}` + "`" + `, response.Body.String())

	response = serveProtocol(` + "`" + `[{"jsonrpc":"2.0","method":"ping"}]` + "`" + `, "")
	require.Empty(t, response.Body.String())
}

func TestInvalidRequestsReturnErrors(t *testing.T) {
	for _, body := range []string{
		` + "`" + `{}` + "`" + `,
		` + "`" + `{"jsonrpc":"1.0","method":"ping"}` + "`" + `,
	} {
		response := serveProtocol(body, "")
		require.JSONEq(t, ` + "`" + `{"jsonrpc":"2.0","id":null,"error":{"code":-32600,"message":"Invalid request"}}` + "`" + `, response.Body.String())
	}
}

func TestBatchProcessesInvalidMembersIndependently(t *testing.T) {
	response := serveProtocol("[1]", "")
	require.JSONEq(t, ` + "`" + `[{"jsonrpc":"2.0","id":null,"error":{"code":-32600,"message":"Invalid request"}}]` + "`" + `, response.Body.String())

	response = serveProtocol(` + "`" + `[{"jsonrpc":"2.0","id":"one","method":"ping"},1,{"jsonrpc":"2.0","method":"ping"}]` + "`" + `, "")
	require.JSONEq(t, ` + "`" + `[
		{"jsonrpc":"2.0","id":"one","result":null},
		{"jsonrpc":"2.0","id":null,"error":{"code":-32600,"message":"Invalid request"}}
	]` + "`" + `, response.Body.String())
}

func TestAcceptQualityControlsEventStreamSelection(t *testing.T) {
	response := serveProtocol(` + "`" + `{"jsonrpc":"2.0","id":"one","method":"ping"}` + "`" + `, "text/event-stream;q=0, application/json")
	require.JSONEq(t, ` + "`" + `{"jsonrpc":"2.0","id":"one","result":null}` + "`" + `, response.Body.String())

	response = serveProtocol(` + "`" + `{"jsonrpc":"2.0","id":"one","method":"ping"}` + "`" + `, "Text/Event-Stream;Q=0.5")
	require.Equal(t, http.StatusNotAcceptable, response.Code)
	require.Empty(t, response.Body.String())

	response = serveProtocol(` + "`" + `{"jsonrpc":"2.0","id":"one","method":"ping"}` + "`" + `, "application/json, text/event-stream;q=0.5")
	require.JSONEq(t, ` + "`" + `{"jsonrpc":"2.0","id":"one","result":null}` + "`" + `, response.Body.String())
}

func TestMixedServerSelectsTheRequestedMethodsResponseType(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		accept    string
		wantCode  int
		wantEvent bool
		wantPing  int
		wantWatch int
	}{
		{name: "unary with both", method: "ping", accept: "application/json, text/event-stream", wantCode: http.StatusOK, wantPing: 1},
		{name: "stream with both", method: "watch", accept: "application/json, text/event-stream", wantCode: http.StatusOK, wantEvent: true, wantWatch: 1},
		{name: "unary with events only", method: "ping", accept: "text/event-stream", wantCode: http.StatusNotAcceptable},
		{name: "stream with JSON only", method: "watch", accept: "application/json", wantCode: http.StatusNotAcceptable},
		{name: "unary without accept", method: "ping", wantCode: http.StatusOK, wantPing: 1},
		{name: "stream without accept", method: "watch", wantCode: http.StatusOK, wantEvent: true, wantWatch: 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := ` + "`" + `{"jsonrpc":"2.0","id":"one","method":"` + "`" + ` + test.method + ` + "`" + `"}` + "`" + `
			response, calls := serveProtocolWithCalls(body, test.accept)
			require.Equal(t, test.wantCode, response.Code)
			require.Equal(t, test.wantPing, calls.ping)
			require.Equal(t, test.wantWatch, calls.watch)
			if test.wantEvent {
				require.Contains(t, response.Body.String(), ` + "`" + `"jsonrpc":"2.0"` + "`" + `)
				require.Contains(t, response.Body.String(), ` + "`" + `"result":null` + "`" + `)
				require.NotContains(t, response.Body.String(), "event:")
			}
			if test.wantCode == http.StatusNotAcceptable {
				require.Empty(t, response.Body.String())
			}
		})
	}
}

func TestMixedServerChoosesAFormatForRequestErrors(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		accept    string
		wantCode  int
		wantEvent bool
		wantBody  bool
	}{
		{name: "malformed prefers JSON", body: "{", accept: "application/json, text/event-stream", wantCode: http.StatusOK, wantBody: true},
		{name: "malformed uses events", body: "{", accept: "text/event-stream", wantCode: http.StatusOK, wantEvent: true, wantBody: true},
		{name: "malformed unsupported", body: "{", accept: "application/xml", wantCode: http.StatusNotAcceptable},
		{name: "invalid prefers JSON", body: ` + "`" + `{}` + "`" + `, accept: "application/json, text/event-stream", wantCode: http.StatusOK, wantBody: true},
		{name: "unknown prefers JSON", body: ` + "`" + `{"jsonrpc":"2.0","id":"one","method":"missing"}` + "`" + `, accept: "application/json, text/event-stream", wantCode: http.StatusOK, wantBody: true},
		{name: "unknown uses events", body: ` + "`" + `{"jsonrpc":"2.0","id":"one","method":"missing"}` + "`" + `, accept: "text/event-stream", wantCode: http.StatusOK, wantEvent: true, wantBody: true},
		{name: "unknown notification has no response", body: ` + "`" + `{"jsonrpc":"2.0","method":"missing"}` + "`" + `, accept: "text/event-stream", wantCode: http.StatusOK},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response, calls := serveProtocolWithCalls(test.body, test.accept)
			require.Equal(t, test.wantCode, response.Code)
			require.Zero(t, calls.ping)
			require.Zero(t, calls.watch)
			require.Equal(t, test.wantBody, response.Body.Len() > 0)
			if test.wantEvent {
				require.Contains(t, response.Body.String(), ` + "`" + `"jsonrpc":"2.0"` + "`" + `)
				require.Contains(t, response.Body.String(), ` + "`" + `"error":` + "`" + `)
				require.NotContains(t, response.Body.String(), "event:")
			}
		})
	}
}

func TestMixedServerKeepsBatchesOnJSON(t *testing.T) {
	body := ` + "`" + `[
		{"jsonrpc":"2.0","id":"one","method":"ping"},
		{"jsonrpc":"2.0","id":"two","method":"watch"},
		{"jsonrpc":"2.0","method":"watch"},
		{"jsonrpc":"2.0","id":"three","method":"ping"}
	]` + "`" + `
	response, calls := serveProtocolWithCalls(body, "application/json, text/event-stream")
	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, 2, calls.ping)
	require.Zero(t, calls.watch)
	require.JSONEq(t, ` + "`" + `[
		{"jsonrpc":"2.0","id":"one","result":null},
		{"jsonrpc":"2.0","id":"two","error":{"code":-32601,"message":"Method is not available in a batch request"}},
		{"jsonrpc":"2.0","id":"three","result":null}
	]` + "`" + `, response.Body.String())

	response, calls = serveProtocolWithCalls(body, "text/event-stream")
	require.Equal(t, http.StatusNotAcceptable, response.Code)
	require.Zero(t, calls.ping)
	require.Zero(t, calls.watch)
	require.Empty(t, response.Body.String())
}

func TestMixedServerClosesTheOriginalBodyOnce(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		accept string
	}{
		{name: "single", body: ` + "`" + `{"jsonrpc":"2.0","id":"one","method":"ping"}` + "`" + `},
		{name: "batch", body: ` + "`" + `[{"jsonrpc":"2.0","id":"one","method":"ping"}]` + "`" + `},
		{name: "parse error", body: "{"},
		{name: "not acceptable", body: ` + "`" + `{"jsonrpc":"2.0","id":"one","method":"ping"}` + "`" + `, accept: "text/event-stream"},
		{name: "stream completion", body: ` + "`" + `{"jsonrpc":"2.0","id":"one","method":"watch"}` + "`" + `, accept: "text/event-stream"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := &trackedRequestBody{Reader: strings.NewReader(test.body)}
			serveProtocolBody(body, test.accept)
			require.Equal(t, 1, body.closes)
		})
	}
}

func TestMixedServerReportsReadAndCloseErrors(t *testing.T) {
	readErr := errors.New("read failed")
	closeErr := errors.New("close failed")
	body := &trackedRequestBody{Reader: errorReader{err: readErr}, closeErr: closeErr}
	_, _, reported := serveProtocolBody(body, "application/json")
	require.ErrorIs(t, errors.Join(reported...), readErr)
	require.ErrorIs(t, errors.Join(reported...), closeErr)
	require.Equal(t, 1, body.closes)
}

func TestUndeclaredServiceErrorIsInternal(t *testing.T) {
	err := goa.NewServiceError(errors.New("failed"), "invalid_params", false, false, false)
	response, _, reported := serveProtocolBodyWithDecoder(
		io.NopCloser(strings.NewReader(` + "`" + `{"jsonrpc":"2.0","id":"one","method":"ping"}` + "`" + `)),
		"application/json, text/event-stream",
		goahttp.RequestDecoder,
		err,
	)
	require.Empty(t, reported)
	require.JSONEq(t, ` + "`" + `{"jsonrpc":"2.0","id":"one","error":{"code":-32603,"message":"failed"}}` + "`" + `, response.Body.String())
}

func TestMixedServerClosesTheBodySuppliedByTheHTTPServer(t *testing.T) {
	original := &trackedRequestBody{Reader: strings.NewReader(` + "`" + `{"jsonrpc":"2.0","id":"one","method":"ping"}` + "`" + `)}
	replacement := &trackedRequestBody{Reader: strings.NewReader("")}
	decoder := func(r *http.Request) goahttp.Decoder {
		result := goahttp.RequestDecoder(r)
		r.Body = replacement
		return result
	}

	serveProtocolBodyWithDecoder(original, "application/json", decoder, nil)

	require.Equal(t, 1, original.closes)
	require.Zero(t, replacement.closes)
}

func serveProtocol(body, accept string) *httptest.ResponseRecorder {
	response, _ := serveProtocolWithCalls(body, accept)
	return response
}

type protocolCalls struct {
	ping  int
	watch int
}

type trackedRequestBody struct {
	io.Reader
	closeErr error
	closes   int
}

type errorReader struct {
	err error
}

func (reader errorReader) Read([]byte) (int, error) {
	return 0, reader.err
}

func (body *trackedRequestBody) Close() error {
	body.closes++
	return body.closeErr
}

func serveProtocolWithCalls(body, accept string) (*httptest.ResponseRecorder, *protocolCalls) {
	response, calls, _ := serveProtocolBody(io.NopCloser(strings.NewReader(body)), accept)
	return response, calls
}

func serveProtocolBody(body io.ReadCloser, accept string) (*httptest.ResponseRecorder, *protocolCalls, []error) {
	return serveProtocolBodyWithDecoder(body, accept, goahttp.RequestDecoder, nil)
}

func serveProtocolBodyWithDecoder(body io.ReadCloser, accept string, decoder func(*http.Request) goahttp.Decoder, pingError error) (*httptest.ResponseRecorder, *protocolCalls, []error) {
	encoder := goahttp.ResponseEncoder
	var reported []error
	errhandler := func(_ context.Context, _ http.ResponseWriter, err error) {
		reported = append(reported, err)
	}
	calls := &protocolCalls{}
	server := &Server{
		Ping: NewPingHandler(
			goa.Endpoint(func(context.Context, any) (any, error) {
				calls.ping++
				return nil, pingError
			}),
			goahttp.NewMuxer(),
			decoder,
			encoder,
			errhandler,
		),
		Watch: NewWatchHandler(
			goa.Endpoint(func(context.Context, any) (any, error) {
				calls.watch++
				return nil, nil
			}),
			goahttp.NewMuxer(),
			decoder,
			encoder,
			errhandler,
		),
		decoder:    decoder,
		encoder:    encoder,
		errhandler: errhandler,
	}
	request := httptest.NewRequest(http.MethodPost, "/protocol", nil)
	request.Body = body
	request.Header.Set("Accept", accept)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	return response, calls, reported
}
`

const pureSSEBodyRuntimeTest = `package server

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	goahttp "goa.design/goa/v3/http"
	goa "goa.design/goa/v3/pkg"
)

type trackedRequestBody struct {
	io.Reader
	closeErr error
	closes   int
}

type errorReader struct {
	err error
}

func (reader errorReader) Read([]byte) (int, error) {
	return 0, reader.err
}

func (body *trackedRequestBody) Close() error {
	body.closes++
	return body.closeErr
}

func TestPureSSEServerClosesTheOriginalBodyOnce(t *testing.T) {
	body := &trackedRequestBody{Reader: strings.NewReader(` + "`" + `{"jsonrpc":"2.0","id":"one","method":"watch","params":{"topic":"alerts"}}` + "`" + `)}
	_, reported := servePureSSE(body)
	require.Empty(t, reported)
	require.Equal(t, 1, body.closes)
}

func TestPureSSEServerReportsReadAndCloseErrors(t *testing.T) {
	readErr := errors.New("read failed")
	closeErr := errors.New("close failed")
	body := &trackedRequestBody{Reader: errorReader{err: readErr}, closeErr: closeErr}
	_, reported := servePureSSE(body)
	require.ErrorIs(t, errors.Join(reported...), readErr)
	require.ErrorIs(t, errors.Join(reported...), closeErr)
	require.Equal(t, 1, body.closes)
}

func servePureSSE(body io.ReadCloser) (*httptest.ResponseRecorder, []error) {
	encoder := goahttp.ResponseEncoder
	var reported []error
	errhandler := func(_ context.Context, _ http.ResponseWriter, err error) {
		reported = append(reported, err)
	}
	server := &Server{
		Watch: NewWatchHandler(
			goa.Endpoint(func(context.Context, any) (any, error) { return nil, nil }),
			goahttp.NewMuxer(),
			goahttp.RequestDecoder,
			encoder,
			errhandler,
		),
		decoder:    goahttp.RequestDecoder,
		encoder:    encoder,
		errhandler: errhandler,
	}
	request := httptest.NewRequest(http.MethodPost, "/decode", nil)
	request.Body = body
	response := httptest.NewRecorder()
	server.handleSSE(response, request)
	return response, reported
}
`
