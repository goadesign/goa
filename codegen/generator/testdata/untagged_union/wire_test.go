// These checks run against freshly generated contracts. The same external
// branch values must reach HTTP and JSON-RPC services and survive response views.
package checks_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	gengrpcclient "generated.local/gen/grpc/records/client"
	genpb "generated.local/gen/grpc/records/pb"
	gengrpcserver "generated.local/gen/grpc/records/server"
	genhttpclient "generated.local/gen/http/records/client"
	genhttpserver "generated.local/gen/http/records/server"
	genrpcserver "generated.local/gen/jsonrpc/rpc/server"
	genrecords "generated.local/gen/records"
	genrpc "generated.local/gen/rpc"
	goahttp "goa.design/goa/v3/http"
)

type (
	recordsService struct{ calls int }
	rpcService     struct{ calls int }
)

func (s *recordsService) Echo(_ context.Context, p *genrecords.Entry) (*genrecords.Entry, error) {
	s.calls++
	return p, nil
}

func (s *recordsService) Viewed(_ context.Context, p *genrecords.Entry) (*genrecords.ViewedEntry, string, error) {
	s.calls++
	return &genrecords.ViewedEntry{Resources: p.Resources}, "selected", nil
}

func (s *recordsService) Select(_ context.Context, p *genrecords.Selection) (*genrecords.Selection, error) {
	s.calls++
	return p, nil
}

func (s *recordsService) Configure(_ context.Context, p *genrecords.Settings) (*genrecords.Settings, error) {
	s.calls++
	return p, nil
}

func (s *rpcService) Echo(_ context.Context, p *genrpc.Entry) (*genrpc.Entry, error) {
	s.calls++
	return p, nil
}

func TestUntaggedWire(t *testing.T) {
	service := &recordsService{}
	rpc := &rpcService{}
	mux := goahttp.NewMuxer()
	genhttpserver.Mount(mux, genhttpserver.New(genrecords.NewEndpoints(service), mux, goahttp.RequestDecoder, goahttp.ResponseEncoder, nil, nil))
	genrpcserver.Mount(mux, genrpcserver.New(genrpc.NewEndpoints(rpc), mux, goahttp.RequestDecoder, goahttp.ResponseEncoder, nil))
	server := httptest.NewServer(mux)
	defer server.Close()
	for _, value := range []string{`"dynamic"`, `[]`, `[{"uri":"skill://review/SKILL.md"}]`} {
		t.Run(value, func(t *testing.T) {
			body := []byte(`{"resources":` + value + "}")
			for _, route := range []string{"/echo", "/viewed"} {
				response, err := http.Post(server.URL+route, "application/json", bytes.NewReader(body))
				require.NoError(t, err)
				received, err := io.ReadAll(response.Body)
				require.NoError(t, response.Body.Close())
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, response.StatusCode, string(received))
				assert.JSONEq(t, string(body), string(received))
				var clientBody genhttpclient.EchoResponseBody
				require.NoError(t, json.Unmarshal(received, &clientBody))
				encoded, err := json.Marshal(clientBody)
				require.NoError(t, err)
				assert.JSONEq(t, string(body), string(encoded))
			}
			request := []byte(`{"jsonrpc":"2.0","id":1,"method":"echo","params":` + string(body) + "}")
			response, err := http.Post(server.URL+"/rpc", "application/json", bytes.NewReader(request))
			require.NoError(t, err)
			received, err := io.ReadAll(response.Body)
			require.NoError(t, response.Body.Close())
			require.NoError(t, err)
			assert.JSONEq(t, `{"jsonrpc":"2.0","id":1,"result":`+string(body)+"}", string(received))
		})
	}
	calls := service.calls
	for _, value := range []string{"null", "true", "1", `"other"`, `[{"uri":""}]`, `[{"uri":null}]`, `[null]`, `{"type":"dynamic","value":"dynamic"}`} {
		response, err := http.Post(server.URL+"/echo", "application/json", bytes.NewReader([]byte(`{"resources":`+value+"}")))
		require.NoError(t, err)
		require.NoError(t, response.Body.Close())
		assert.Equal(t, http.StatusBadRequest, response.StatusCode, value)
	}
	assert.Equal(t, calls, service.calls, "invalid values reached the service")
	for _, value := range []string{`""`, "9007199254740993", "-9007199254740993", "true", "false", `{"uri":"file://record"}`, "[]"} {
		body := `{"value":` + value + "}"
		response, err := http.Post(server.URL+"/select", "application/json", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		received, err := io.ReadAll(response.Body)
		require.NoError(t, response.Body.Close())
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, response.StatusCode, string(received))
		assert.JSONEq(t, body, string(received))
	}
	var selected genrecords.Value
	require.NoError(t, json.Unmarshal([]byte("9007199254740993"), &selected))
	number, ok := selected.AsNumber()
	assert.True(t, ok)
	assert.EqualValues(t, int64(9007199254740993), number)
	for _, invalid := range []string{"", "null", "tru", "1.5", "9223372036854775808", "\v1\v", "\u00a01\u00a0"} {
		assert.Error(t, selected.UnmarshalJSON([]byte(invalid)), invalid)
		number, ok = selected.AsNumber()
		assert.True(t, ok)
		assert.EqualValues(t, int64(9007199254740993), number)
	}
}

func TestUntaggedProtobufRoundTrip(t *testing.T) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	genpb.RegisterRecordsServer(server, gengrpcserver.New(genrecords.NewEndpoints(&recordsService{}), nil))
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	t.Cleanup(func() {
		server.Stop()
		require.NoError(t, listener.Close())
		require.NoError(t, <-done)
	})
	connection, err := grpc.NewClient("passthrough:///records", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
		return listener.DialContext(ctx)
	}))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, connection.Close()) })
	client := gengrpcclient.NewClient(connection)
	for _, value := range []genrecords.Setting{
		genrecords.NewSettingText("text"),
		genrecords.NewSettingEnabled(true),
		genrecords.NewSettingEnabled(false),
	} {
		input := &genrecords.Settings{Setting: value}
		result, err := client.Configure()(t.Context(), input)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	}
}
