// These tests exercise generated clients and servers with completed and
// input-required results. Service JSON uses its own tags, while HTTP and
// JSON-RPC use the native transport names. Protobuf keeps its native oneof.
package unions_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	gengrpcclient "generated.local/gen/grpc/unions/client"
	genpb "generated.local/gen/grpc/unions/pb"
	gengrpcserver "generated.local/gen/grpc/unions/server"
	genclient "generated.local/gen/http/unions/client"
	genserver "generated.local/gen/http/unions/server"
	genrpcclient "generated.local/gen/jsonrpc/rpcunions/client"
	genrpcserver "generated.local/gen/jsonrpc/rpcunions/server"
	genrpcunions "generated.local/gen/rpcunions"
	genunions "generated.local/gen/unions"
	goahttp "goa.design/goa/v3/http"
)

func TestObjectUnionHTTPAndServiceJSON(t *testing.T) {
	spec, err := openapi3.NewLoader().LoadFromFile("../http/openapi3.json")
	require.NoError(t, err)
	for _, test := range []struct {
		name   string
		value  *genunions.Outcome
		wire   string
		stored string
	}{
		{"complete", &genunions.Outcome{Outcome: genunions.NewResultChoiceComplete(&genunions.Complete{Reference: "done"})}, `{"resultType":"complete","reference":"done"}`, `{"Outcome":{"resultType":"complete","stored_reference":"done"}}`},
		{"input", &genunions.Outcome{Outcome: genunions.NewResultChoiceInputRequired(&genunions.Pending{Message: "Provide a name", State: "operation-1"})}, `{"resultType":"input_required","message":"Provide a name","state":"operation-1"}`, `{"Outcome":{"resultType":"input_required","Message":"Provide a name","State":"operation-1"}}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			stored, err := json.Marshal(test.value)
			require.NoError(t, err)
			require.JSONEq(t, test.stored, string(stored))
			var decoded genunions.Outcome
			require.NoError(t, json.Unmarshal(stored, &decoded))
			require.Equal(t, test.value, &decoded)
			fromCLI, err := genclient.BuildSelectedPayload(&test.wire)
			require.NoError(t, err)
			require.Equal(t, test.value, fromCLI)
			for _, transport := range []struct {
				path     string
				encode   func(*http.Request, any) error
				decode   func(*http.Request) (*genunions.Outcome, error)
				response func(context.Context, http.ResponseWriter, any) error
				result   func(*http.Response) (any, error)
				wire     string
			}{
				{"/echo", genclient.EncodeEchoRequest(goahttp.RequestEncoder), genserver.DecodeEchoRequest(goahttp.NewMuxer(), goahttp.RequestDecoder), genserver.EncodeEchoResponse(goahttp.ResponseEncoder), genclient.DecodeEchoResponse(goahttp.ResponseDecoder, false), `{"outcome":` + test.wire + `}`},
				{"/selected", genclient.EncodeSelectedRequest(goahttp.RequestEncoder), genserver.DecodeSelectedRequest(goahttp.NewMuxer(), goahttp.RequestDecoder), genserver.EncodeSelectedResponse(goahttp.ResponseEncoder), genclient.DecodeSelectedResponse(goahttp.ResponseDecoder, false), test.wire},
			} {
				request := httptest.NewRequest(http.MethodPost, transport.path, nil)
				require.NoError(t, transport.encode(request, test.value))
				wire, err := io.ReadAll(request.Body)
				require.NoError(t, err)
				require.NoError(t, request.Body.Close())
				require.JSONEq(t, transport.wire, string(wire))
				var object any
				require.NoError(t, json.Unmarshal(wire, &object))
				operation := spec.Paths.Value(transport.path).Post
				require.NoError(t, operation.RequestBody.Value.Content["application/json"].Schema.Value.VisitJSON(object))
				incoming := httptest.NewRequest(http.MethodPost, transport.path, bytes.NewReader(wire))
				payload, err := transport.decode(incoming)
				require.NoError(t, err)
				require.NoError(t, incoming.Body.Close())
				require.Equal(t, test.value, payload)
				recorder := httptest.NewRecorder()
				require.NoError(t, transport.response(t.Context(), recorder, payload))
				require.JSONEq(t, transport.wire, recorder.Body.String())
				require.NoError(t, operation.Responses.Status(200).Value.Content["application/json"].Schema.Value.VisitJSON(object))
				response := recorder.Result()
				response.Request = request
				result, err := transport.result(response)
				require.NoError(t, err)
				require.NoError(t, response.Body.Close())
				require.Equal(t, test.value, result)
			}
		})
	}
}

func TestObjectUnionInvalidWireStopsBeforeService(t *testing.T) {
	for _, wire := range []string{
		`null`, `[]`, `{"reference":"done"}`, `{"resultType":null}`, `{"resultType":1}`, `{"resultType":"unknown"}`,
		`{"resultType":"complete"}`, `{"resultType":"complete","reference":""}`,
		`{"resultType":"input_required","message":"Provide a name"}`,
		`{"resultType":"complete","value":{"reference":"done"}}`,
	} {
		incoming := httptest.NewRequest(http.MethodPost, "/selected", bytes.NewBufferString(wire))
		_, err := genserver.DecodeSelectedRequest(goahttp.NewMuxer(), goahttp.RequestDecoder)(incoming)
		require.Error(t, err, wire)
		require.NoError(t, incoming.Body.Close())
	}
	for _, choice := range []genunions.ResultChoice{{}, genunions.NewResultChoiceComplete(nil)} {
		_, err := json.Marshal(choice)
		require.Error(t, err)
	}
}

func TestObjectUnionJSONRPC(t *testing.T) {
	value := &genrpcunions.Outcome{Outcome: genrpcunions.NewResultChoiceComplete(&genrpcunions.Complete{Reference: "done"})}
	const want = `{"resultType":"complete","reference":"done"}`
	request := httptest.NewRequest(http.MethodPost, "/rpc", nil)
	id, err := genrpcclient.EncodeEchoRequest(goahttp.RequestEncoder)(request, value)
	require.NoError(t, err)
	wire, err := io.ReadAll(request.Body)
	require.NoError(t, err)
	require.NoError(t, request.Body.Close())
	var message struct{ Params json.RawMessage }
	require.NoError(t, json.Unmarshal(wire, &message))
	require.JSONEq(t, want, string(message.Params))
	server := genrpcserver.New(&genrpcunions.Endpoints{Echo: func(_ context.Context, payload any) (any, error) {
		require.Equal(t, value, payload)
		return payload, nil
	}}, goahttp.NewMuxer(), goahttp.RequestDecoder, goahttp.ResponseEncoder, func(_ context.Context, _ http.ResponseWriter, err error) { t.Errorf("RPC failed: %v", err) })
	incoming := httptest.NewRequest(http.MethodPost, "/rpc", bytes.NewReader(wire))
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, incoming)
	require.NoError(t, incoming.Body.Close())
	var reply struct{ Result json.RawMessage }
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &reply))
	require.JSONEq(t, want, string(reply.Result))
	response := recorder.Result()
	response.Request = request
	result, err := genrpcclient.DecodeEchoResponse(goahttp.ResponseDecoder, false)(response, id)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
	require.Equal(t, value, result)
}

func TestObjectUnionGRPCRoundTrip(t *testing.T) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	echo := func(_ context.Context, value any) (any, error) { return value, nil }
	genpb.RegisterUnionsServer(server, gengrpcserver.New(&genunions.Endpoints{Echo: echo, Selected: echo, Nested: echo, Tagged: echo, TaggedSelected: echo}, nil))
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	t.Cleanup(func() {
		server.Stop()
		require.NoError(t, listener.Close())
		require.NoError(t, <-done)
	})
	connection, err := grpc.NewClient("passthrough:///unions", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, connection.Close()) })
	client := gengrpcclient.NewClient(connection)
	value := &genunions.Outcome{Outcome: genunions.NewResultChoiceInputRequired(&genunions.Pending{Message: "Provide name", State: "operation-1"})}
	result, err := client.Echo()(t.Context(), value)
	require.NoError(t, err)
	require.Equal(t, value, result)
	nested := &genunions.Collection{Items: []*genunions.Outcome{value}, Named: map[string]*genunions.Outcome{"first": value}}
	result, err = client.Nested()(t.Context(), nested)
	require.NoError(t, err)
	require.Equal(t, nested, result)
	tagged := &genunions.Tagged{Choice: genunions.NewTaggedChoiceText("unchanged")}
	result, err = client.Tagged()(t.Context(), tagged)
	require.NoError(t, err)
	require.Equal(t, tagged, result)
}

func TestObjectUnionCollectionsAndTaggedJSON(t *testing.T) {
	value := &genunions.Outcome{Outcome: genunions.NewResultChoiceComplete(&genunions.Complete{Reference: "done"})}
	nested := &genunions.Collection{Items: []*genunions.Outcome{value}, Named: map[string]*genunions.Outcome{"first": value}}
	request := httptest.NewRequest(http.MethodPost, "/nested", nil)
	require.NoError(t, genclient.EncodeNestedRequest(goahttp.RequestEncoder)(request, nested))
	wire, err := io.ReadAll(request.Body)
	require.NoError(t, err)
	require.NoError(t, request.Body.Close())
	const want = `{"items":[{"outcome":{"resultType":"complete","reference":"done"}}],"named":{"first":{"outcome":{"resultType":"complete","reference":"done"}}}}`
	require.JSONEq(t, want, string(wire))
	incoming := httptest.NewRequest(http.MethodPost, "/nested", bytes.NewReader(wire))
	result, err := genserver.DecodeNestedRequest(goahttp.NewMuxer(), goahttp.RequestDecoder)(incoming)
	require.NoError(t, err)
	require.NoError(t, incoming.Body.Close())
	require.Equal(t, nested, result)
	for _, test := range []struct {
		value *genunions.Tagged
		want  string
	}{
		{&genunions.Tagged{Choice: genunions.NewTaggedChoiceText("unchanged")}, `{"Choice":{"kind":"text","data":"unchanged"}}`},
		{&genunions.Tagged{Choice: genunions.NewTaggedChoiceComplete(&genunions.Complete{Reference: "done"})}, `{"Choice":{"kind":"complete","data":{"stored_reference":"done"}}}`},
	} {
		data, err := json.Marshal(test.value)
		require.NoError(t, err)
		require.JSONEq(t, test.want, string(data))
		var decoded genunions.Tagged
		require.NoError(t, json.Unmarshal(data, &decoded))
		require.Equal(t, test.value, &decoded)
		request := httptest.NewRequest(http.MethodPost, "/tagged_selected", nil)
		require.NoError(t, genclient.EncodeTaggedSelectedRequest(goahttp.RequestEncoder)(request, test.value))
		wire, err := io.ReadAll(request.Body)
		require.NoError(t, err)
		require.NoError(t, request.Body.Close())
		var expected struct{ Choice json.RawMessage }
		require.NoError(t, json.Unmarshal(data, &expected))
		// The service stores internal JSON names, while the HTTP object branch
		// uses reference. The discriminator and value envelope stay unchanged.
		want := string(expected.Choice)
		if object, ok := test.value.Choice.AsComplete(); ok {
			require.Equal(t, "done", object.Reference)
			want = `{"kind":"complete","data":{"reference":"done"}}`
		}
		require.JSONEq(t, want, string(wire))
		incoming := httptest.NewRequest(http.MethodPost, "/tagged_selected", bytes.NewReader(wire))
		result, err := genserver.DecodeTaggedSelectedRequest(goahttp.NewMuxer(), goahttp.RequestDecoder)(incoming)
		require.NoError(t, err)
		require.NoError(t, incoming.Body.Close())
		require.Equal(t, test.value, result)
		fromCLI, err := genclient.BuildTaggedSelectedPayload(&want)
		require.NoError(t, err)
		require.Equal(t, test.value, fromCLI)
	}
}
