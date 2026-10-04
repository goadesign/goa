// These tests run generated service endpoints before transport encoding. A service
// result that violates its selected view must be a server fault; a domain error
// returned by the service must retain its declared meaning and original cause.
package codegen_test

import "testing"

// TestGeneratedResultValidationIsServerFault runs a generated service and checks
// its output errors through the HTTP, gRPC and JSON-RPC error encoders.
func TestGeneratedResultValidationIsServerFault(t *testing.T) {
	dir := renderViewedResultRuntimeModule(t)
	writeViewedResultServerRuntimeTest(t, dir, "unary_status", resultValidationFaultRuntimeTest)
	runViewedResultRuntimeTests(t, dir, "./jsonrpc/unary_status/server")
}

const resultValidationFaultRuntimeTest = `package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	genmapped "generated.local/gen/mapped_body"
	genunary "generated.local/gen/unary_status"
	goagrpc "goa.design/goa/v3/grpc"
	goahttp "goa.design/goa/v3/http"
	"goa.design/goa/v3/jsonrpc"
	goa "goa.design/goa/v3/pkg"
)

type (
	resultFaultService struct {
		result *genunary.DecoderStatus
		view string
		failure error
	}
	mappedFaultService struct {
		result *genmapped.MappedBody
		view string
		failure error
	}
)

func (s *resultFaultService) Fetch(context.Context) (*genunary.DecoderStatus, string, error) {
	return s.result, s.view, s.failure
}

func (s *mappedFaultService) Fetch(context.Context) (*genmapped.MappedBody, string, error) {
	return s.result, s.view, s.failure
}

func TestSelectedResultValidationFault(t *testing.T) {
	for _, test := range []struct {
		name string
		view string
		input string
		valid bool
	}{
		{name: "valid summary", view: "summary", input: "{\"id\":{\"value\":\"ready\"}}", valid: true},
		{name: "valid detailed", view: "detailed", input: "{\"id\":{\"value\":\"ready\"},\"detail\":\"more\"}", valid: true},
		{name: "missing required result field", view: "summary", input: "{}"},
		{name: "null required result field", view: "detailed", input: "{\"id\":null}"},
		{name: "undeclared view", view: "other", input: "{\"id\":{\"value\":\"ready\"}}"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var result genmapped.MappedBody
			require.NoError(t, json.Unmarshal([]byte(test.input), &result))
			endpoint := genmapped.NewEndpoints(&mappedFaultService{result: &result, view: test.view}).Fetch
			returned, err := endpoint(t.Context(), nil)
			if test.valid {
				require.NoError(t, err)
				require.NotNil(t, returned)
				return
			}
			require.Nil(t, returned)
			var fault *goa.ServiceError
			require.ErrorAs(t, err, &fault)
			require.True(t, fault.Fault)
			require.Equal(t, "fault", fault.Name)
			cause := errors.Unwrap(fault)
			require.Error(t, cause)
			var validation *goa.ServiceError
			require.ErrorAs(t, cause, &validation)
			require.False(t, validation.Fault)
			require.ErrorIs(t, err, cause)
			require.Equal(t, http.StatusInternalServerError, goahttp.NewErrorResponse(t.Context(), err).StatusCode())
			require.Equal(t, codes.Internal, status.Code(goagrpc.EncodeError(err)))
		})
	}
}

func TestServiceDomainErrorRemainsUnchanged(t *testing.T) {
	failure := goa.MissingFieldError("domain_value", "service operation")
	endpoint := genmapped.NewEndpoints(&mappedFaultService{failure: failure}).Fetch
	result, err := endpoint(t.Context(), nil)
	require.Nil(t, result)
	require.Same(t, failure, err)
	require.Equal(t, http.StatusBadRequest, goahttp.NewErrorResponse(t.Context(), err).StatusCode())
	require.Equal(t, codes.InvalidArgument, status.Code(goagrpc.EncodeError(err)))
}

func TestJSONRPCServerReturnsResultFault(t *testing.T) {
	service := &resultFaultService{result: &genunary.DecoderStatus{Label: "ready"}, view: "other"}
	server := New(genunary.NewEndpoints(service), goahttp.NewMuxer(), goahttp.RequestDecoder, goahttp.ResponseEncoder, nil)
	request := httptest.NewRequest(http.MethodPost, "/unary", strings.NewReader("{\"jsonrpc\":\"2.0\",\"id\":\"result-1\",\"method\":\"fetch\",\"params\":{}}"))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	var envelope struct {
		ID string
		Error *jsonrpc.ErrorResponse
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &envelope))
	require.Equal(t, "result-1", envelope.ID)
	require.NotNil(t, envelope.Error)
	require.Equal(t, jsonrpc.InternalError, envelope.Error.Code)
}
`
