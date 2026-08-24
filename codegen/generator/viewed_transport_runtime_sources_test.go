// This file contains source code that calls generated HTTP and JSON-RPC code
// with result views. Each source string runs in a temporary Go module.
package generator

const httpViewedSSEServerTest = `package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	service "generated.local/gen/http_view_stream"
	goahttp "goa.design/goa/v3/http"
	goa "goa.design/goa/v3/pkg"
)

type viewedService struct {
	serviceDefaults
	watchView string
}

type unknownViewService struct {
	serviceDefaults
	sendError chan error
}

type changingViewService struct {
	serviceDefaults
	sendError chan error
}

type sendResultService struct {
	serviceDefaults
	sendError chan error
}

type fixedErrorAfterEventService struct{ serviceDefaults }

type mixedErrorAfterEventService struct{ serviceDefaults }

type serviceDefaults struct{}

type statusRecorder struct {
	*httptest.ResponseRecorder
	statuses []int
}

type streamResponseWriter struct {
	header     http.Header
	status     int
	body       strings.Builder
	writeError error
}

var errEventWrite = errors.New("event write failed")
var errAfterEvent = errors.New("service failed after event")

func (serviceDefaults) Watch(_ context.Context, _ service.WatchServerStream) error {
	return nil
}

func (serviceDefaults) Fixed(_ context.Context, _ service.FixedServerStream) error {
	return nil
}

func (serviceDefaults) Mixed(_ context.Context, _ service.MixedServerStream) (*service.Immediate, error) {
	return nil, nil
}

func (s *viewedService) Watch(_ context.Context, stream service.WatchServerStream) error {
	stream.SetView(s.watchView)
	return stream.Send(viewedEvent())
}

func (s *viewedService) Fixed(_ context.Context, stream service.FixedServerStream) error {
	return stream.Send(viewedEvent())
}

func (s *unknownViewService) Watch(_ context.Context, stream service.WatchServerStream) error {
	stream.SetView("unknown")
	err := stream.Send(viewedEvent())
	s.sendError <- err
	return err
}

func (*unknownViewService) Fixed(_ context.Context, stream service.FixedServerStream) error {
	return stream.Send(viewedEvent())
}

func (s *changingViewService) Watch(_ context.Context, stream service.WatchServerStream) error {
	stream.SetView("summary")
	if err := stream.Send(viewedEvent()); err != nil {
		return err
	}
	stream.SetView("detailed")
	err := stream.Send(viewedEvent())
	s.sendError <- err
	return nil
}

func (*changingViewService) Fixed(_ context.Context, stream service.FixedServerStream) error {
	return stream.Send(viewedEvent())
}

func (s *sendResultService) Watch(_ context.Context, stream service.WatchServerStream) error {
	err := stream.Send(viewedEvent())
	s.sendError <- err
	return err
}

func (*sendResultService) Fixed(_ context.Context, stream service.FixedServerStream) error {
	return stream.Send(viewedEvent())
}

func (*fixedErrorAfterEventService) Watch(_ context.Context, stream service.WatchServerStream) error {
	return stream.Send(viewedEvent())
}

func (*fixedErrorAfterEventService) Fixed(_ context.Context, stream service.FixedServerStream) error {
	if err := stream.Send(viewedEvent()); err != nil {
		return err
	}
	return errAfterEvent
}

func (*mixedErrorAfterEventService) Mixed(_ context.Context, stream service.MixedServerStream) (*service.Immediate, error) {
	if err := stream.Send(viewedEvent()); err != nil {
		return nil, err
	}
	return nil, errAfterEvent
}

func (w *statusRecorder) WriteHeader(status int) {
	w.statuses = append(w.statuses, status)
	w.ResponseRecorder.WriteHeader(status)
}

func (w *streamResponseWriter) Header() http.Header {
	return w.header
}

func (w *streamResponseWriter) WriteHeader(status int) {
	w.status = status
}

func (w *streamResponseWriter) Write(data []byte) (int, error) {
	if w.writeError != nil {
		return 0, w.writeError
	}
	return w.body.Write(data)
}

func TestViewedSSEServerUsesRequestView(t *testing.T) {
	svc := &viewedService{watchView: "detailed"}
	handler := NewWatchHandler(
		service.NewWatchEndpoint(svc),
		goahttp.NewMuxer(),
		goahttp.RequestDecoder,
		goahttp.ResponseEncoder,
		func(context.Context, http.ResponseWriter, error) {},
		nil,
	)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/watch", nil))
	require.Equal(t, "detailed", recorder.Header().Get("goa-view"))
	require.JSONEq(t,
		` + "`" + `{"event_id":"event-1","profile":{"display_name":"Ada"}}` + "`" + `,
		sseData(t, recorder.Body.String()),
	)
}

func TestFixedViewedSSEServerIsSpecialized(t *testing.T) {
	_, exposesSetView := reflect.TypeOf((*service.FixedServerStream)(nil)).Elem().MethodByName("SetView")
	require.False(t, exposesSetView)

	svc := &viewedService{}
	handler := NewFixedHandler(
		service.NewFixedEndpoint(svc),
		goahttp.NewMuxer(),
		goahttp.RequestDecoder,
		goahttp.ResponseEncoder,
		func(context.Context, http.ResponseWriter, error) {},
		nil,
	)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/fixed", nil))
	require.JSONEq(t,
		` + "`" + `{"event_id":"event-1","profile":{"display_name":"Ada"}}` + "`" + `,
		sseData(t, recorder.Body.String()),
	)
}

func TestFixedViewedSSEServerDoesNotEncodeServiceErrorAfterEvent(t *testing.T) {
	var handled error
	handler := NewFixedHandler(
		service.NewFixedEndpoint(&fixedErrorAfterEventService{}),
		goahttp.NewMuxer(),
		goahttp.RequestDecoder,
		goahttp.ResponseEncoder,
		func(_ context.Context, _ http.ResponseWriter, err error) { handled = err },
		nil,
	)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/fixed", nil))
	require.ErrorIs(t, handled, errAfterEvent)
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, 1, strings.Count(recorder.Body.String(), "data:"))
	require.NotContains(t, recorder.Body.String(), ` + "`" + `"name":` + "`" + `)
}

func TestMixedResultSSEServerDoesNotEncodeServiceErrorAfterEvent(t *testing.T) {
	var handled error
	handler := NewMixedHandler(
		service.NewMixedEndpoint(&mixedErrorAfterEventService{}),
		goahttp.NewMuxer(),
		goahttp.RequestDecoder,
		goahttp.ResponseEncoder,
		func(_ context.Context, _ http.ResponseWriter, err error) { handled = err },
		nil,
	)
	request := httptest.NewRequest("GET", "/mixed", nil)
	request.Header.Set("Accept", "text/event-stream")
	recorder := &statusRecorder{ResponseRecorder: httptest.NewRecorder()}
	handler.ServeHTTP(recorder, request)
	require.ErrorIs(t, handled, errAfterEvent)
	require.Equal(t, []int{http.StatusOK}, recorder.statuses)
	require.Equal(t, 1, strings.Count(recorder.Body.String(), "data:"))
	require.JSONEq(t,
		` + "`" + `{"event_id":"event-1","profile":{"display_name":"Ada"}}` + "`" + `,
		sseData(t, recorder.Body.String()),
	)
	require.NotContains(t, recorder.Body.String(), ` + "`" + `"name":` + "`" + `)
}

func TestUnknownViewedSSEServerSelectionIsRejectedBeforeWriting(t *testing.T) {
	sendError := make(chan error, 1)
	handler := NewWatchHandler(
		service.NewWatchEndpoint(&unknownViewService{sendError: sendError}),
		goahttp.NewMuxer(),
		goahttp.RequestDecoder,
		goahttp.ResponseEncoder,
		func(context.Context, http.ResponseWriter, error) {},
		nil,
	)
	server := httptest.NewServer(handler)
	defer server.Close()
	response, err := server.Client().Get(server.URL + "/watch")
	require.NoError(t, err)
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	requireBoundaryError(t, <-sendError, goa.InvalidEnumValue, "view")
	require.Equal(t, http.StatusBadRequest, response.StatusCode)
	var serviceError map[string]any
	require.NoError(t, json.Unmarshal(body, &serviceError))
	require.Equal(t, goa.InvalidEnumValue, serviceError["name"])
	require.Contains(t, serviceError["message"], "value of view")
}

func TestEmptyViewedSSEServerSelectionUsesDefaultView(t *testing.T) {
	handler := NewWatchHandler(
		service.NewWatchEndpoint(&viewedService{}),
		goahttp.NewMuxer(),
		goahttp.RequestDecoder,
		goahttp.ResponseEncoder,
		func(context.Context, http.ResponseWriter, error) {},
		nil,
	)
	server := httptest.NewServer(handler)
	defer server.Close()
	response, err := server.Client().Get(server.URL + "/watch")
	require.NoError(t, err)
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Equal(t, "default", response.Header.Get("goa-view"))
	require.JSONEq(t,
		` + "`" + `{"event_id":"event-1","profile":{"display_name":"Ada"}}` + "`" + `,
		sseData(t, string(body)),
	)
}

func TestViewedSSEServerRejectsViewChangesAfterFirstEvent(t *testing.T) {
	sendError := make(chan error, 1)
	handler := NewWatchHandler(
		service.NewWatchEndpoint(&changingViewService{sendError: sendError}),
		goahttp.NewMuxer(),
		goahttp.RequestDecoder,
		goahttp.ResponseEncoder,
		func(context.Context, http.ResponseWriter, error) {},
		nil,
	)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/watch", nil))
	requireBoundaryError(t, <-sendError, goa.InvalidEnumValue, "view")
	require.Equal(t, "summary", recorder.Header().Get("goa-view"))
	require.Equal(t, 1, strings.Count(recorder.Body.String(), "data:"))
}

func TestViewedSSEServerReturnsEventWriteError(t *testing.T) {
	sendError := make(chan error, 1)
	handler := NewWatchHandler(
		service.NewWatchEndpoint(&sendResultService{sendError: sendError}),
		goahttp.NewMuxer(),
		goahttp.RequestDecoder,
		goahttp.ResponseEncoder,
		func(context.Context, http.ResponseWriter, error) {},
		nil,
	)
	writer := &streamResponseWriter{header: make(http.Header), writeError: errEventWrite}
	handler.ServeHTTP(writer, httptest.NewRequest("GET", "/watch", nil))
	require.ErrorIs(t, <-sendError, errEventWrite)
	require.Equal(t, http.StatusOK, writer.status)
}

func TestViewedSSEServerReturnsFlushError(t *testing.T) {
	sendError := make(chan error, 1)
	handler := NewWatchHandler(
		service.NewWatchEndpoint(&sendResultService{sendError: sendError}),
		goahttp.NewMuxer(),
		goahttp.RequestDecoder,
		goahttp.ResponseEncoder,
		func(context.Context, http.ResponseWriter, error) {},
		nil,
	)
	writer := &streamResponseWriter{header: make(http.Header)}
	handler.ServeHTTP(writer, httptest.NewRequest("GET", "/watch", nil))
	require.ErrorIs(t, <-sendError, http.ErrNotSupported)
	require.Equal(t, http.StatusOK, writer.status)
	require.NotEmpty(t, writer.body.String())
}

func viewedEvent() *service.Event {
	return &service.Event{
		EventID: "event-1",
		Profile: &service.Profile{DisplayName: "Ada"},
	}
}

func sseData(t *testing.T, event string) string {
	t.Helper()
	for _, line := range strings.Split(event, "\n") {
		if strings.HasPrefix(line, "data:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		}
	}
	t.Errorf("SSE event has no data field: %q", event)
	return ""
}

func requireBoundaryError(t *testing.T, err error, name, field string) {
	t.Helper()
	var serviceError *goa.ServiceError
	require.ErrorAs(t, err, &serviceError)
	require.Equal(t, name, serviceError.Name)
	require.NotNil(t, serviceError.Field)
	require.Equal(t, field, *serviceError.Field)
}
`

const httpViewedSSEClientTest = `package client

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	service "generated.local/gen/http_view_stream"
	goahttp "goa.design/goa/v3/http"
)

type doerFunc func(*http.Request) (*http.Response, error)

func (f doerFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestViewedSSEClientReconstructsSelectedBody(t *testing.T) {
	cases := []struct {
		name        string
		endpoint    func(*Client) func(context.Context, any) (any, error)
		recv        func(any) (*service.Event, error)
		view        string
		body        string
		wantProfile bool
	}{
		{
			name:     "summary",
			endpoint: func(c *Client) func(context.Context, any) (any, error) { return c.Watch() },
			recv: func(raw any) (*service.Event, error) {
				var stream service.WatchClientStream = raw.(WatchClientStream)
				return stream.Recv()
			},
			view:     "summary",
			body:     ` + "`" + `{"event_id":"summary-event"}` + "`" + `,
		},
		{
			name:        "detailed fixed",
			endpoint:    func(c *Client) func(context.Context, any) (any, error) { return c.Fixed() },
			recv: func(raw any) (*service.Event, error) {
				var stream service.FixedClientStream = raw.(FixedClientStream)
				return stream.Recv()
			},
			body:        ` + "`" + `{"event_id":"detailed-event","profile":{"display_name":"Ada"}}` + "`" + `,
			wantProfile: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			doer := doerFunc(func(*http.Request) (*http.Response, error) {
				header := http.Header{"Content-Type": []string{"text/event-stream"}}
				if tc.view != "" {
					header.Set("goa-view", tc.view)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     header,
					Body: io.NopCloser(strings.NewReader("data: " + tc.body + "\n\n")),
				}, nil
			})
			client := NewClient(
				"http", "example.test", doer,
				goahttp.RequestEncoder, goahttp.ResponseDecoder, false,
			)
			rawStream, err := tc.endpoint(client)(context.Background(), nil)
			require.NoError(t, err)
			event, err := tc.recv(rawStream)
			require.NoError(t, err)
			require.Equal(t, strings.TrimSuffix(tc.name, " fixed")+"-event", event.EventID)
			if tc.wantProfile {
				require.Equal(t, "Ada", event.Profile.DisplayName)
			} else {
				require.Nil(t, event.Profile)
			}
		})
	}
}
`

const jsonRPCViewedUnaryClientTest = `package client

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	service "generated.local/gen/jsonrpc_unary"
	goahttp "goa.design/goa/v3/http"
	goa "goa.design/goa/v3/pkg"
)

func TestVariableViewedUnaryResponseUsesRepresentationBody(t *testing.T) {
	response := jsonRPCResponse(
		` + "`" + `{"jsonrpc":"2.0","id":"1","result":{"view":"detailed","body":{"event_id":"event-1","profile":{"display_name":"Ada"}}}}` + "`" + `,
	)
	response.Header.Set("goa-view", "summary")
	var result any
	var err error
	require.NotPanics(t, func() {
		result, err = DecodeFetchResponse(goahttp.ResponseDecoder, false)(response)
	})
	require.NoError(t, err)
	event := result.(*service.Event)
	require.Equal(t, "event-1", event.EventID)
	require.Equal(t, "Ada", event.Profile.DisplayName)
}

func TestVariableViewedUnaryResponseRejectsInvalidRepresentation(t *testing.T) {
	cases := []struct {
		name      string
		result    string
		errorName string
		field     string
	}{
		{"missing view", ` + "`" + `{"body":{"event_id":"event-1"}}` + "`" + `, goa.MissingField, "view"},
		{"null view", ` + "`" + `{"view":null,"body":{"event_id":"event-1"}}` + "`" + `, goa.MissingField, "view"},
		{"missing body", ` + "`" + `{"view":"summary"}` + "`" + `, goa.MissingField, "body"},
		{"null body", ` + "`" + `{"view":"summary","body":null}` + "`" + `, goa.MissingField, "body"},
		{"unknown view", ` + "`" + `{"view":"unknown","body":{"event_id":"event-1"}}` + "`" + `, goa.InvalidEnumValue, "view"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			response := jsonRPCResponse(` + "`" + `{"jsonrpc":"2.0","id":"1","result":` + "`" + ` + tc.result + "}")
			var err error
			require.NotPanics(t, func() {
				_, err = DecodeFetchResponse(goahttp.ResponseDecoder, false)(response)
			})
			requireBoundaryError(t, err, tc.errorName, tc.field)
		})
	}
}

func TestFixedViewedUnaryResponseUsesBodyOnly(t *testing.T) {
	response := jsonRPCResponse(
		` + "`" + `{"jsonrpc":"2.0","id":"1","result":{"event_id":"event-1","profile":{"display_name":"Ada"}}}` + "`" + `,
	)
	result, err := DecodeFixedResponse(goahttp.ResponseDecoder, false)(response)
	require.NoError(t, err)
	event := result.(*service.Event)
	require.Equal(t, "event-1", event.EventID)
	require.Equal(t, "Ada", event.Profile.DisplayName)
}

func jsonRPCResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func requireBoundaryError(t *testing.T, err error, name, field string) {
	t.Helper()
	var serviceError *goa.ServiceError
	require.ErrorAs(t, err, &serviceError)
	require.Equal(t, name, serviceError.Name)
	require.NotNil(t, serviceError.Field)
	require.Equal(t, field, *serviceError.Field)
}
`

const jsonRPCViewedUnaryServerTest = `package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	service "generated.local/gen/jsonrpc_unary"
	goahttp "goa.design/goa/v3/http"
	goa "goa.design/goa/v3/pkg"
)

type viewedService struct {
	fetchView string
}

func (s *viewedService) Fetch(context.Context) (*service.Event, string, error) {
	return viewedEvent(), s.fetchView, nil
}

func (*viewedService) Fixed(context.Context) (*service.Event, error) {
	return viewedEvent(), nil
}

func TestVariableViewedUnaryServerEmitsRepresentation(t *testing.T) {
	recorder := serveJSONRPC(t, "fetch", "detailed")
	require.Empty(t, recorder.Header().Get("goa-view"))
	require.JSONEq(t,
		` + "`" + `{"view":"detailed","body":{"event_id":"event-1","profile":{"display_name":"Ada"}}}` + "`" + `,
		jsonRPCResult(t, recorder),
	)
}

func TestFixedViewedUnaryServerEmitsBodyOnly(t *testing.T) {
	recorder := serveJSONRPC(t, "fixed", "")
	require.JSONEq(t,
		` + "`" + `{"event_id":"event-1","profile":{"display_name":"Ada"}}` + "`" + `,
		jsonRPCResult(t, recorder),
	)
}

func TestUnknownViewedUnaryServerSelectionIsRejected(t *testing.T) {
	result, err := service.NewFetchEndpoint(&viewedService{fetchView: "unknown"})(context.Background(), nil)
	require.Nil(t, result)
	requireBoundaryError(t, err, goa.InvalidEnumValue, "view")
}

func serveJSONRPC(t *testing.T, method, view string) *httptest.ResponseRecorder {
	t.Helper()
	server := New(
		service.NewEndpoints(&viewedService{fetchView: view}),
		goahttp.NewMuxer(),
		goahttp.RequestDecoder,
		goahttp.ResponseEncoder,
		func(context.Context, http.ResponseWriter, error) {},
	)
	body := []byte(` + "`" + `{"jsonrpc":"2.0","id":"1","method":"` + "`" + ` + method + ` + "`" + `"}` + "`" + `)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest("POST", "/rpc", bytes.NewReader(body)))
	require.Equal(t, http.StatusOK, recorder.Code)
	return recorder
}

func jsonRPCResult(t *testing.T, recorder *httptest.ResponseRecorder) string {
	t.Helper()
	var response struct {
		Result json.RawMessage ` + "`" + `json:"result"` + "`" + `
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	return string(response.Result)
}

func viewedEvent() *service.Event {
	return &service.Event{
		EventID: "event-1",
		Profile: &service.Profile{DisplayName: "Ada"},
	}
}

func requireBoundaryError(t *testing.T, err error, name, field string) {
	t.Helper()
	var serviceError *goa.ServiceError
	require.ErrorAs(t, err, &serviceError)
	require.Equal(t, name, serviceError.Name)
	require.NotNil(t, serviceError.Field)
	require.Equal(t, field, *serviceError.Field)
}
`

const jsonRPCViewedSSEClientTest = `package client

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	service "generated.local/gen/jsonrpcsse"
	goahttp "goa.design/goa/v3/http"
	goa "goa.design/goa/v3/pkg"
)

type doerFunc func(*http.Request) (*http.Response, error)

func (f doerFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestViewedSSENotificationReconstructsTransportBody(t *testing.T) {
	data := ` + "`" + `{"jsonrpc":"2.0","method":"watch","params":{"view":"detailed","body":{"event_id":"event-1","profile":{"display_name":"Ada"}}}}` + "`" + `
	event, err := recvWatch("notification", data)
	require.NoError(t, err)
	require.Equal(t, "event-1", event.EventID)
	require.Equal(t, "Ada", event.Profile.DisplayName)
}

func TestViewedSSEFinalResponseEndsStream(t *testing.T) {
	data := ` + "`" + `{"jsonrpc":"2.0","id":"1","result":null}` + "`" + `
	_, err := recvWatch("response", data)
	require.ErrorIs(t, err, io.EOF)
}

func TestViewedSSERejectsInvalidRepresentation(t *testing.T) {
	cases := []struct {
		name      string
		params    string
		errorName string
		field     string
	}{
		{"missing view", ` + "`" + `{"body":{"event_id":"event-1"}}` + "`" + `, goa.MissingField, "view"},
		{"null view", ` + "`" + `{"view":null,"body":{"event_id":"event-1"}}` + "`" + `, goa.MissingField, "view"},
		{"missing body", ` + "`" + `{"view":"summary"}` + "`" + `, goa.MissingField, "body"},
		{"null body", ` + "`" + `{"view":"summary","body":null}` + "`" + `, goa.MissingField, "body"},
		{"unknown view", ` + "`" + `{"view":"unknown","body":{"event_id":"event-1"}}` + "`" + `, goa.InvalidEnumValue, "view"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := ` + "`" + `{"jsonrpc":"2.0","method":"watch","params":` + "`" + ` + tc.params + "}"
			_, err := recvWatch("notification", data)
			requireBoundaryError(t, err, tc.errorName, tc.field)
		})
	}
}

func TestFixedViewedSSEUsesBodyOnly(t *testing.T) {
	data := ` + "`" + `{"jsonrpc":"2.0","method":"fixed","params":{"event_id":"event-1","profile":{"display_name":"Ada"}}}` + "`" + `
	event, err := recvFixed("notification", data)
	require.NoError(t, err)
	require.Equal(t, "event-1", event.EventID)
	require.Equal(t, "Ada", event.Profile.DisplayName)
}

func recvWatch(eventType, data string) (*service.Event, error) {
	client := sseClient(eventType, data)
	raw, err := client.Watch()(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	transport := raw.(*WatchStreamImpl)
	var stream service.WatchClientStream = transport
	return stream.Recv()
}

func recvFixed(eventType, data string) (*service.Event, error) {
	client := sseClient(eventType, data)
	raw, err := client.Fixed()(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	transport := raw.(*FixedStreamImpl)
	var stream service.FixedClientStream = transport
	return stream.Recv()
}

func sseClient(eventType, data string) *Client {
	doer := doerFunc(func(*http.Request) (*http.Response, error) {
		body := "event: " + eventType + "\ndata: " + data + "\n\n"
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})
	return NewClient(
		"http", "example.test", doer,
		goahttp.RequestEncoder, goahttp.ResponseDecoder, false,
	)
}

func requireBoundaryError(t *testing.T, err error, name, field string) {
	t.Helper()
	var serviceError *goa.ServiceError
	require.ErrorAs(t, err, &serviceError)
	require.Equal(t, name, serviceError.Name)
	require.NotNil(t, serviceError.Field)
	require.Equal(t, field, *serviceError.Field)
}
`

const jsonRPCViewedSSEServerTest = `package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	service "generated.local/gen/jsonrpcsse"
	goahttp "goa.design/goa/v3/http"
	goa "goa.design/goa/v3/pkg"
)

type viewedService struct{}

type unknownViewService struct {
	sendError chan error
}

type changedViewService struct {
	sendError chan error
}

func (*viewedService) Watch(_ context.Context, stream service.WatchServerStream) error {
	stream.SetView("summary")
	if err := stream.Send(viewedEvent()); err != nil {
		return err
	}
	return stream.Send(viewedEvent())
}

func (*viewedService) Fixed(_ context.Context, stream service.FixedServerStream) error {
	if err := stream.Send(viewedEvent()); err != nil {
		return err
	}
	return stream.Send(viewedEvent())
}

func (s *unknownViewService) Watch(_ context.Context, stream service.WatchServerStream) error {
	stream.SetView("unknown")
	err := stream.Send(viewedEvent())
	s.sendError <- err
	return err
}

func (*unknownViewService) Fixed(_ context.Context, stream service.FixedServerStream) error {
	return stream.Send(viewedEvent())
}

func (s *changedViewService) Watch(_ context.Context, stream service.WatchServerStream) error {
	stream.SetView("summary")
	if err := stream.Send(viewedEvent()); err != nil {
		return err
	}
	stream.SetView("detailed")
	err := stream.Send(viewedEvent())
	s.sendError <- err
	return err
}

func (*changedViewService) Fixed(_ context.Context, stream service.FixedServerStream) error {
	return stream.Send(viewedEvent())
}

func TestVariableViewedSSEServerEmitsRepresentation(t *testing.T) {
	recorder := serveSSE(t, "watch")
	records := jsonRPCSSERecords(t, recorder.Body.String())
	require.Len(t, records, 3)
	require.JSONEq(t,
		` + "`" + `{"view":"summary","body":{"event_id":"event-1"}}` + "`" + `,
		string(records[0].Params),
	)
	require.JSONEq(t,
		` + "`" + `{"view":"summary","body":{"event_id":"event-1"}}` + "`" + `,
		string(records[1].Params),
	)
	require.JSONEq(t, ` + "`" + `null` + "`" + `, string(records[2].Result))
}

func TestChangedViewedSSEServerSelectionIsRejectedBeforeWriting(t *testing.T) {
	sendError := make(chan error, 1)
	recorder := serveSSEService(t, "watch", &changedViewService{sendError: sendError})
	requireBoundaryError(t, <-sendError, goa.InvalidEnumValue, "view")
	records := jsonRPCSSERecords(t, recorder.Body.String())
	require.Len(t, records, 2)
	require.JSONEq(t,
		` + "`" + `{"view":"summary","body":{"event_id":"event-1"}}` + "`" + `,
		string(records[0].Params),
	)
	require.NotEmpty(t, records[1].Error)
}

func TestFixedViewedSSEServerEmitsBodyOnly(t *testing.T) {
	recorder := serveSSE(t, "fixed")
	records := jsonRPCSSERecords(t, recorder.Body.String())
	require.Len(t, records, 3)
	require.JSONEq(t,
		` + "`" + `{"event_id":"event-1","profile":{"display_name":"Ada"}}` + "`" + `,
		string(records[0].Params),
	)
	require.JSONEq(t,
		` + "`" + `{"event_id":"event-1","profile":{"display_name":"Ada"}}` + "`" + `,
		string(records[1].Params),
	)
	require.JSONEq(t, ` + "`" + `null` + "`" + `, string(records[2].Result))
}

func TestUnknownViewedSSEServerSelectionIsRejectedBeforeWriting(t *testing.T) {
	sendError := make(chan error, 1)
	recorder := serveSSEService(t, "watch", &unknownViewService{sendError: sendError})
	requireBoundaryError(t, <-sendError, goa.InvalidEnumValue, "view")
	records := jsonRPCSSERecords(t, recorder.Body.String())
	require.Len(t, records, 1)
	require.Empty(t, records[0].Params)
	require.NotEmpty(t, records[0].Error)
}

func serveSSE(t *testing.T, method string) *httptest.ResponseRecorder {
	t.Helper()
	return serveSSEService(t, method, &viewedService{})
}

func serveSSEService(t *testing.T, method string, svc service.Service) *httptest.ResponseRecorder {
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
	server.ServeHTTP(recorder, httptest.NewRequest("POST", "/events", bytes.NewReader(body)))
	require.Equal(t, http.StatusOK, recorder.Code)
	return recorder
}

type sseRecord struct {
	Params json.RawMessage ` + "`" + `json:"params"` + "`" + `
	Result json.RawMessage ` + "`" + `json:"result"` + "`" + `
	Error  json.RawMessage ` + "`" + `json:"error"` + "`" + `
}

func jsonRPCSSERecords(t *testing.T, event string) []sseRecord {
	t.Helper()
	var records []sseRecord
	for _, line := range strings.Split(event, "\n") {
		if strings.HasPrefix(line, "data:") {
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			var record sseRecord
			require.NoError(t, json.Unmarshal([]byte(data), &record))
			records = append(records, record)
		}
	}
	return records
}

func viewedEvent() *service.Event {
	return &service.Event{
		EventID: "event-1",
		Profile: &service.Profile{DisplayName: "Ada"},
	}
}

func requireBoundaryError(t *testing.T, err error, name, field string) {
	t.Helper()
	var serviceError *goa.ServiceError
	require.ErrorAs(t, err, &serviceError)
	require.Equal(t, name, serviceError.Name)
	require.NotNil(t, serviceError.Field)
	require.Equal(t, field, *serviceError.Field)
}
`
