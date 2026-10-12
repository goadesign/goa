package sample_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	genclient "generated.local/gen/grpc/sample/client"
	genpb "generated.local/gen/grpc/sample/pb"
	genserver "generated.local/gen/grpc/sample/server"
	genhttpclient "generated.local/gen/http/sample/client"
	genhttpserver "generated.local/gen/http/sample/server"
	gensample "generated.local/gen/sample"
	"github.com/stretchr/testify/require"
	goahttp "goa.design/goa/v3/http"
	"google.golang.org/protobuf/proto"
)

func TestRawJSONByteRoundTrip(t *testing.T) {
	for _, raw := range []string{"null", "true", "9007199254740993123456789", "1e9999999999", `[]`, `{}`, `{"nested":[null,9007199254740993123456789]}`, `"` + strings.Repeat("x", 800000) + `"`} {
		original := &gensample.Operation{Value: gensample.NewOperationValueJSON(gensample.OperationValueBranchJSON(raw))}
		encoded, err := json.Marshal(original)
		require.NoError(t, err)
		require.Contains(t, string(encoded), `"value":`+raw)
		var restored gensample.Operation
		require.NoError(t, json.Unmarshal(encoded, &restored))
		require.Equal(t, original, &restored)
		wire, err := proto.Marshal(genclient.NewProtoEchoRequest(original))
		require.NoError(t, err)
		message := &genpb.EchoRequest{}
		require.NoError(t, proto.Unmarshal(wire, message))
		require.Equal(t, original, genserver.NewEchoPayload(message))
	}
}

func TestDistinctRawJSONBranchesAndNesting(t *testing.T) {
	raw := []byte(`{"number":9007199254740993123456789}`)
	for _, choice := range []gensample.TwoValue{
		gensample.NewTwoValueFirst(gensample.TwoValueBranchFirst(raw)),
		gensample.NewTwoValueSecond(gensample.TwoValueBranchSecond(raw)),
		gensample.NewTwoValueNamed(gensample.RawValue(raw)),
		gensample.NewTwoValuePlain(gensample.TwoValueBranchPlain(raw)),
	} {
		value := &gensample.Two{Value: choice}
		original := &gensample.Nested{One: &gensample.Operation{Value: gensample.NewOperationValueJSON(gensample.OperationValueBranchJSON("null"))}, Many: []*gensample.Two{value}, Named: map[string]*gensample.Two{"item": value}}
		encoded, err := json.Marshal(original)
		require.NoError(t, err)
		if choice.Kind() == gensample.TwoValueKindPlain {
			require.NotContains(t, string(encoded), string(raw))
		} else {
			require.Contains(t, string(encoded), string(raw))
		}
		var restored gensample.Nested
		require.NoError(t, json.Unmarshal(encoded, &restored))
		require.Equal(t, original, &restored)
		wire, err := proto.Marshal(genclient.NewProtoNestedRequest(original))
		require.NoError(t, err)
		message := &genpb.NestedRequest{}
		require.NoError(t, proto.Unmarshal(wire, message))
		require.Equal(t, original, genserver.NewNestedPayload(message))
	}
	for _, branch := range []string{"first", "second", "named"} {
		var value gensample.Two
		require.NoError(t, json.Unmarshal([]byte(`{"Value":{"type":"`+branch+`","value":null}}`), &value))
		require.Equal(t, branch, string(value.Value.Kind()))
	}
	for _, wire := range []string{`{"Value":{"type":"first"}}`, `{"Value":{"type":"plain","value":null}}`, `{"Value":{"type":"plain","value":{}}}`} {
		var value gensample.Two
		require.Error(t, json.Unmarshal([]byte(wire), &value), wire)
	}
}

func TestRawJSONHTTP(t *testing.T) {
	for _, raw := range []string{"null", `{"number":9007199254740993123456789}`} {
		for _, choice := range []gensample.TwoValue{
			gensample.NewTwoValueFirst(gensample.TwoValueBranchFirst(raw)),
			gensample.NewTwoValueSecond(gensample.TwoValueBranchSecond(raw)),
			gensample.NewTwoValueNamed(gensample.RawValue(raw)),
		} {
			value := &gensample.Two{Value: choice}
			request := httptest.NewRequest(http.MethodPost, "/two", nil)
			require.NoError(t, genhttpclient.EncodeTwoRequest(goahttp.RequestEncoder)(request, value))
			wire, err := io.ReadAll(request.Body)
			require.NoError(t, err)
			require.NoError(t, request.Body.Close())
			require.Contains(t, string(wire), `"value":`+raw)
			incoming := httptest.NewRequest(http.MethodPost, "/two", bytes.NewReader(wire))
			decoded, err := genhttpserver.DecodeTwoRequest(goahttp.NewMuxer(), goahttp.RequestDecoder)(incoming)
			require.NoError(t, err)
			require.NoError(t, incoming.Body.Close())
			require.Equal(t, value, decoded)
			recorder := httptest.NewRecorder()
			require.NoError(t, genhttpserver.EncodeTwoResponse(goahttp.ResponseEncoder)(context.Background(), recorder, value))
			response := recorder.Result()
			result, err := genhttpclient.DecodeTwoResponse(goahttp.ResponseDecoder, false)(response)
			require.NoError(t, err)
			require.NoError(t, response.Body.Close())
			require.Equal(t, value, result)
		}
	}
}
