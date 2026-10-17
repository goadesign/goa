// These tests exchange complete values through generated HTTP codecs. The
// service representation and an unmapped service must keep their original names.
package mapped_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"

	genclient "generated.local/gen/http/mapped/client"
	genserver "generated.local/gen/http/mapped/server"
	genplainclient "generated.local/gen/http/plain/client"
	genplainserver "generated.local/gen/http/plain/server"
	genmapped "generated.local/gen/mapped"
	goahttp "goa.design/goa/v3/http"
)

func TestMappedBodyRoundTrip(t *testing.T) {
	item := `{"displayName":"","isEnabled":false,"itemCount":0,"opaqueText":"{\"display_name\":\"unchanged\"}","noteText":"note","label":"retained"}`
	derived := `{"derivedName":"derived","is_enabled":false,"item_count":0,"opaque_text":"same"}`
	for _, branch := range []struct{ name, value string }{{"first", item}, {"second", derived}} {
		t.Run(branch.name, func(t *testing.T) {
			wire := `{"items":[` + item + `,` + item + `],"byKey":{"display_name":` + item + `},"derived":` + derived + `,"tree":{"nodeName":"root","children":[{"nodeName":"child","children":[]}]},"selection":{"type":"` + branch.name + `","value":` + branch.value + `}}`
			missing := strings.Replace(wire, `"isEnabled":false,`, "", 1)
			invalid := httptest.NewRequest(http.MethodPost, "/mapped", bytes.NewBufferString(missing))
			_, invalidErr := genserver.DecodeExchangeRequest(goahttp.NewMuxer(), goahttp.RequestDecoder)(invalid)
			require.Error(t, invalidErr)
			require.NoError(t, invalid.Body.Close())
			incoming := httptest.NewRequest(http.MethodPost, "/mapped", bytes.NewBufferString(wire))
			decoded, err := genserver.DecodeExchangeRequest(goahttp.NewMuxer(), goahttp.RequestDecoder)(incoming)
			require.NoError(t, err)
			require.NoError(t, incoming.Body.Close())
			value := decoded
			require.Len(t, value.Items, 2)
			require.Contains(t, value.ByKey, "display_name")
			require.Equal(t, `{"display_name":"unchanged"}`, value.Items[0].OpaqueText)
			request := httptest.NewRequest(http.MethodPost, "/mapped", nil)
			require.NoError(t, genclient.EncodeExchangeRequest(goahttp.RequestEncoder)(request, decoded))
			encoded, err := io.ReadAll(request.Body)
			require.NoError(t, err)
			require.NoError(t, request.Body.Close())
			require.JSONEq(t, wire, string(encoded))
			recorder := httptest.NewRecorder()
			require.NoError(t, genserver.EncodeExchangeResponse(goahttp.ResponseEncoder)(context.Background(), recorder, decoded))
			require.JSONEq(t, wire, recorder.Body.String())
			response := recorder.Result()
			response.Request = request
			result, err := genclient.DecodeExchangeResponse(goahttp.ResponseDecoder, false)(response)
			require.NoError(t, err)
			require.NoError(t, response.Body.Close())
			require.Equal(t, decoded, result)
			serviceJSON, err := json.Marshal(decoded)
			require.NoError(t, err)
			require.NotContains(t, string(serviceJSON), `"displayName":`)
			require.NotContains(t, string(serviceJSON), `"derivedName":`)
			require.Contains(t, string(serviceJSON), `"LegacyNote":"note"`)
			require.Contains(t, string(serviceJSON), `"Label":"retained"`)
			spec, err := openapi3.NewLoader().LoadFromFile("../http/openapi3.json")
			require.NoError(t, err)
			var document any
			require.NoError(t, json.Unmarshal(encoded, &document))
			operation := spec.Paths.Value("/mapped").Post
			require.NoError(t, operation.RequestBody.Value.Content["application/json"].Schema.Value.VisitJSON(document))
			require.NoError(t, operation.Responses.Status(200).Value.Content["application/json"].Schema.Value.VisitJSON(document))
		})
	}
}

func TestUnmappedBodyRoundTrip(t *testing.T) {
	item := `{"display_name":"plain","is_enabled":false,"item_count":0,"opaque_text":"unchanged"}`
	wire := `{"items":[` + item + `],"by_key":{},"derived":` + item + `,"tree":{"node_name":"root","children":[]},"selection":{"type":"first","value":` + item + `}}`
	incoming := httptest.NewRequest(http.MethodPost, "/plain", bytes.NewBufferString(wire))
	decoded, err := genplainserver.DecodeExchangeRequest(goahttp.NewMuxer(), goahttp.RequestDecoder)(incoming)
	require.NoError(t, err)
	require.NoError(t, incoming.Body.Close())
	request := httptest.NewRequest(http.MethodPost, "/plain", nil)
	require.NoError(t, genplainclient.EncodeExchangeRequest(goahttp.RequestEncoder)(request, decoded))
	encoded, err := io.ReadAll(request.Body)
	require.NoError(t, err)
	require.NoError(t, request.Body.Close())
	require.JSONEq(t, wire, string(encoded))
	recorder := httptest.NewRecorder()
	require.NoError(t, genplainserver.EncodeExchangeResponse(goahttp.ResponseEncoder)(context.Background(), recorder, decoded))
	require.JSONEq(t, wire, recorder.Body.String())
	response := recorder.Result()
	response.Request = request
	result, err := genplainclient.DecodeExchangeResponse(goahttp.ResponseDecoder, false)(response)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
	require.Equal(t, decoded, result)
}

func TestMappedErrorRoundTrip(t *testing.T) {
	problem := &genmapped.Problem{ErrorCode: "rejected"}
	recorder := httptest.NewRecorder()
	require.NoError(t, genserver.EncodeExchangeError(goahttp.ResponseEncoder, nil)(context.Background(), recorder, problem))
	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.JSONEq(t, `{"errorCode":"rejected"}`, recorder.Body.String())
	response := recorder.Result()
	response.Request = httptest.NewRequest(http.MethodPost, "/mapped", nil)
	_, err := genclient.DecodeExchangeResponse(goahttp.ResponseDecoder, false)(response)
	require.NoError(t, response.Body.Close())
	var decoded *genmapped.Problem
	require.ErrorAs(t, err, &decoded)
	require.Equal(t, problem, decoded)
}
