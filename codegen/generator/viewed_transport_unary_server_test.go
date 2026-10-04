// This source runs a generated JSON-RPC server and checks selected result views.
// Invalid service selections return a server fault with the original validation
// cause; successful fixed and service-selected views retain their wire fields.
package generator

const jsonRPCViewedUnaryServerTest = `package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	genunary "generated.local/gen/jsonrpc_unary"
	goahttp "goa.design/goa/v3/http"
	"goa.design/goa/v3/jsonrpc"
	goa "goa.design/goa/v3/pkg"
)

type (
	viewedService struct {
		fetchView string
	}
)

func (s *viewedService) Fetch(context.Context) (*genunary.Event, string, error) {
	return viewedEvent(), s.fetchView, nil
}

func (*viewedService) Fixed(context.Context) (*genunary.Event, error) {
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
	result, err := genunary.NewFetchEndpoint(&viewedService{fetchView: "unknown"})(context.Background(), nil)
	require.Nil(t, result)
	var fault *goa.ServiceError
	require.ErrorAs(t, err, &fault)
	require.True(t, fault.Fault)
	require.Equal(t, "fault", fault.Name)
	requireBoundaryError(t, errors.Unwrap(fault), goa.InvalidEnumValue, "view")
	var response jsonrpc.Response
	require.NoError(t, json.Unmarshal(serveJSONRPC(t, "fetch", "unknown").Body.Bytes(), &response))
	require.Equal(t, "1", response.ID)
	require.NotNil(t, response.Error)
	require.Equal(t, jsonrpc.InternalError, response.Error.Code)
}

// serveJSONRPC sends one request to the generated server and returns its response.
func serveJSONRPC(t *testing.T, method, view string) *httptest.ResponseRecorder {
	t.Helper()
	server := New(
		genunary.NewEndpoints(&viewedService{fetchView: view}),
		goahttp.NewMuxer(),
		goahttp.RequestDecoder,
		goahttp.ResponseEncoder,
		func(_ context.Context, _ http.ResponseWriter, err error) {
			t.Error(err)
		},
	)
	body := []byte(` + "`" + `{"jsonrpc":"2.0","id":"1","method":"` + "`" + ` + method + ` + "`" + `"}` + "`" + `)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest("POST", "/rpc", bytes.NewReader(body)))
	require.Equal(t, http.StatusOK, recorder.Code)
	return recorder
}

// jsonRPCResult reads the structured result from a successful server response.
func jsonRPCResult(t *testing.T, recorder *httptest.ResponseRecorder) string {
	t.Helper()
	var response struct {
		Result json.RawMessage ` + "`" + `json:"result"` + "`" + `
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	return string(response.Result)
}

func viewedEvent() *genunary.Event {
	return &genunary.Event{
		EventID: "event-1",
		Profile: &genunary.Profile{DisplayName: "Ada"},
	}
}

// requireBoundaryError checks the original validation name and field retained as the fault cause.
func requireBoundaryError(t *testing.T, err error, name, field string) {
	t.Helper()
	var serviceError *goa.ServiceError
	require.ErrorAs(t, err, &serviceError)
	require.Equal(t, name, serviceError.Name)
	require.NotNil(t, serviceError.Field)
	require.Equal(t, field, *serviceError.Field)
}
`
