// These tests use generated HTTP codecs for both directions and validate the
// emitted bodies against generated OpenAPI. Direct service serialization must
// keep its internal names, including fields inside nested array elements.
package names_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"

	genclient "generated.local/gen/http/names/client"
	genserver "generated.local/gen/http/names/server"
	genrpcclient "generated.local/gen/jsonrpc/rpcnames/client"
	genrpcserver "generated.local/gen/jsonrpc/rpcnames/server"
	gennames "generated.local/gen/names"
	genrpcnames "generated.local/gen/rpcnames"
	goahttp "goa.design/goa/v3/http"
)

func TestHTTPNamesAndInternalSerialization(t *testing.T) {
	spec, err := openapi3.NewLoader().LoadFromFile("../http/openapi3.json")
	require.NoError(t, err)
	empty := ""
	for _, test := range []struct {
		name     string
		value    *gennames.Document
		wire     string
		internal string
	}{
		{
			name: "nested array and present empty optional field",
			value: &gennames.Document{
				DisplayName: "sample",
				NoteText:    &empty,
				Entries: []*gennames.Entry{{
					LabelText: "first",
					Detail:    &gennames.Detail{UnitName: "units"},
				}},
			},
			wire:     `{"displayName":"sample","noteText":"","entries":[{"labelText":"first","detail":{"unitName":"units"}}]}`,
			internal: `{"display_name":"sample","note_text":"","stored_entries":[{"label_text":"first","stored_detail":{"unit_name":"units"}}]}`,
		},
		{
			name:     "required empty values and absent optional field",
			value:    &gennames.Document{Entries: []*gennames.Entry{}},
			wire:     `{"displayName":"","entries":[]}`,
			internal: `{"display_name":"","stored_entries":[]}`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			internal, err := json.Marshal(test.value)
			require.NoError(t, err)
			require.JSONEq(t, test.internal, string(internal))

			request := httptest.NewRequest(http.MethodPost, "/documents", nil)
			require.NoError(t, genclient.EncodeExchangeRequest(goahttp.RequestEncoder)(request, test.value))
			wire, err := io.ReadAll(request.Body)
			require.NoError(t, err)
			require.NoError(t, request.Body.Close())
			require.JSONEq(t, test.wire, string(wire))
			operation := spec.Paths.Value("/documents").Post
			assertSchemaAccepts(t, operation.RequestBody.Value.Content["application/json"].Schema.Value, wire)

			incoming := httptest.NewRequest(http.MethodPost, "/documents", bytes.NewReader(wire))
			decoded, err := genserver.DecodeExchangeRequest(goahttp.NewMuxer(), goahttp.RequestDecoder)(incoming)
			require.NoError(t, err)
			require.NoError(t, incoming.Body.Close())
			require.Equal(t, test.value, decoded)

			recorder := httptest.NewRecorder()
			require.NoError(t, genserver.EncodeExchangeResponse(goahttp.ResponseEncoder)(context.Background(), recorder, decoded))
			require.Equal(t, http.StatusOK, recorder.Code)
			require.JSONEq(t, test.wire, recorder.Body.String())
			assertSchemaAccepts(t, operation.Responses.Status(200).Value.Content["application/json"].Schema.Value, recorder.Body.Bytes())
			response := recorder.Result()
			response.Request = request
			result, err := genclient.DecodeExchangeResponse(goahttp.ResponseDecoder, false)(response)
			require.NoError(t, err)
			require.NoError(t, response.Body.Close())
			require.Equal(t, test.value, result)
		})
	}
}

func TestExplicitPublicNamePreservesGoFieldName(t *testing.T) {
	value := &gennames.ExternalName{DisplayName: "sample"}
	const want = `{"public_name":"sample"}`
	internal, err := json.Marshal(value)
	require.NoError(t, err)
	require.JSONEq(t, `{"stored_name":"sample"}`, string(internal))
	request := httptest.NewRequest(http.MethodPost, "/mapped", nil)
	require.NoError(t, genclient.EncodeMappedRequest(goahttp.RequestEncoder)(request, value))
	wire, err := io.ReadAll(request.Body)
	require.NoError(t, err)
	require.NoError(t, request.Body.Close())
	require.JSONEq(t, want, string(wire))
	incoming := httptest.NewRequest(http.MethodPost, "/mapped", bytes.NewReader(wire))
	decoded, err := genserver.DecodeMappedRequest(goahttp.NewMuxer(), goahttp.RequestDecoder)(incoming)
	require.NoError(t, err)
	require.NoError(t, incoming.Body.Close())
	require.Equal(t, value, decoded)

	recorder := httptest.NewRecorder()
	require.NoError(t, genserver.EncodeMappedResponse(goahttp.ResponseEncoder)(context.Background(), recorder, decoded))
	require.JSONEq(t, want, recorder.Body.String())
	response := recorder.Result()
	response.Request = request
	result, err := genclient.DecodeMappedResponse(goahttp.ResponseDecoder, false)(response)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
	require.Equal(t, value, result)
}

func TestJSONRPCExchangeUsesTransportNames(t *testing.T) {
	value := &genrpcnames.Document{DisplayName: "sample", Entries: []*genrpcnames.Entry{}}
	const want = `{"displayName":"sample","entries":[]}`
	request := httptest.NewRequest(http.MethodPost, "/rpc", nil)
	requestID, err := genrpcclient.EncodeExchangeRequest(goahttp.RequestEncoder)(request, value)
	require.NoError(t, err)
	wire, err := io.ReadAll(request.Body)
	require.NoError(t, err)
	require.NoError(t, request.Body.Close())
	var message struct {
		Params json.RawMessage
	}
	require.NoError(t, json.Unmarshal(wire, &message))
	require.JSONEq(t, want, string(message.Params))

	endpoints := &genrpcnames.Endpoints{
		Exchange: func(_ context.Context, payload any) (any, error) {
			require.Equal(t, value, payload)
			return payload, nil
		},
	}
	server := genrpcserver.New(
		endpoints, goahttp.NewMuxer(), goahttp.RequestDecoder, goahttp.ResponseEncoder,
		func(_ context.Context, _ http.ResponseWriter, err error) {
			t.Errorf("JSON-RPC exchange failed: %v", err)
		},
	)
	incoming := httptest.NewRequest(http.MethodPost, "/rpc", bytes.NewReader(wire))
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, incoming)
	require.NoError(t, incoming.Body.Close())
	require.Equal(t, http.StatusOK, recorder.Code)
	var reply struct {
		Result json.RawMessage
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &reply))
	require.JSONEq(t, want, string(reply.Result))
	response := recorder.Result()
	response.Request = request
	result, err := genrpcclient.DecodeExchangeResponse(goahttp.ResponseDecoder, false)(response, requestID)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
	require.Equal(t, value, result)
}

// assertSchemaAccepts decodes a generated HTTP body and validates it against
// the corresponding generated OpenAPI request or response schema.
func assertSchemaAccepts(t *testing.T, schema *openapi3.Schema, body []byte) {
	t.Helper()
	var value any
	require.NoError(t, json.Unmarshal(body, &value))
	require.NoError(t, schema.VisitJSON(value))
}
