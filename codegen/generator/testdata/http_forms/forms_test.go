// These tests exchange generated form requests over HTTP and check the typed
// values received by service endpoints. Invalid bodies must never reach them.
package formtests_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	genforms "generated.local/gen/forms"
	genclient "generated.local/gen/http/forms/client"
	genserver "generated.local/gen/http/forms/server"
	goahttp "goa.design/goa/v3/http"
)

func TestFormClientServer(t *testing.T) {
	empty, token, cookie := "", "credential", "session-value"
	flag, small, large := true, int32(-12), int64(-9223372036854775807)
	unsigned, unsigned32, unsigned64 := uint(9), uint32(4294967295), uint64(18446744073709551615)
	fraction32, fraction64 := float32(1.25), 2.5
	input := &genforms.ExchangePayload{
		Name: "a + b&世界", Empty: empty, Flag: &flag, Count: 7,
		Small: &small, Large: &large, Unsigned: &unsigned,
		Unsigned32: &unsigned32, Unsigned64: &unsigned64,
		Fraction32: &fraction32, Fraction64: &fraction64,
		Data: []byte{0, 255, 128}, Numbers: []int{0, -1, 42},
		Labels: []string{"", "two + three"}, Blobs: [][]byte{{0, 255}, {}},
		Site: "west", Page: 2, Token: &token, Cookie: &cookie,
	}
	seen := make(chan *genforms.ExchangePayload, 1)
	mux := goahttp.NewMuxer()
	endpoints := &genforms.Endpoints{
		Exchange: func(_ context.Context, value any) (any, error) {
			seen <- value.(*genforms.ExchangePayload)
			return "accepted", nil
		},
	}
	server := genserver.New(endpoints, mux, goahttp.RequestDecoder, goahttp.ResponseEncoder, nil, nil)
	genserver.Mount(mux, server)
	httpServer := httptest.NewServer(mux)
	t.Cleanup(httpServer.Close)
	address, err := url.Parse(httpServer.URL)
	require.NoError(t, err)
	client := genclient.NewClient(address.Scheme, address.Host, httpServer.Client(), goahttp.RequestEncoder, goahttp.ResponseDecoder, false)
	result, err := client.Exchange()(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, "accepted", result)
	require.Equal(t, input, <-seen)

	invalid, err := httpServer.Client().Post(httpServer.URL+"/forms/west?page_number=2", "application/x-www-form-urlencoded", strings.NewReader("empty=&numbers=1"))
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, invalid.StatusCode)
	require.NoError(t, invalid.Body.Close())
	select {
	case <-seen:
		t.Fatal("invalid body reached the service endpoint")
	default:
	}

	request := httptest.NewRequest(http.MethodPost, "/forms/west", nil)
	require.NoError(t, genclient.EncodeExchangeRequest(nil)(request, input))
	wire, err := io.ReadAll(request.Body)
	require.NoError(t, err)
	require.NoError(t, request.Body.Close())
	require.Equal(t, "application/x-www-form-urlencoded", request.Header.Get("Content-Type"))
	require.Equal(t, int64(len(wire)), request.ContentLength)
	values, err := url.ParseQuery(string(wire))
	require.NoError(t, err)
	require.Equal(t, []string{"0", "-1", "42"}, values["numbers"])
	require.Equal(t, []string{"", "two + three"}, values["labels"])
	require.Equal(t, "AP+A", values.Get("data"))
	require.NotContains(t, values, "token")
	require.NotContains(t, values, "cookie")
	require.NotContains(t, values, "site")
	require.NotContains(t, values, "page")
	require.Equal(t, "2", request.URL.Query().Get("page_number"))
	replay, err := request.GetBody()
	require.NoError(t, err)
	replayBytes, err := io.ReadAll(replay)
	require.NoError(t, err)
	require.NoError(t, replay.Close())
	require.Equal(t, wire, replayBytes)
}

func TestFormValidationAndDefaults(t *testing.T) {
	for _, test := range []struct {
		name, body, mediaType, query string
		valid                        bool
		count                        int
	}{
		{"defaults", "name=ready&empty=&numbers=1", "application/x-www-form-urlencoded", "", true, 3},
		{"explicit zero", "name=ready&empty=&numbers=1&count=0", "application/x-www-form-urlencoded; charset=UTF-8", "", true, 0},
		{"unknown extension", "name=ready&empty=&numbers=1&extra=value", "application/x-www-form-urlencoded", "", true, 3},
		{"missing name", "empty=&numbers=1", "application/x-www-form-urlencoded", "", false, 0},
		{"query cannot provide body", "empty=&numbers=1", "application/x-www-form-urlencoded", "&name=ready", false, 0},
		{"empty constrained name", "name=&empty=&numbers=1", "application/x-www-form-urlencoded", "", false, 0},
		{"missing empty field", "name=ready&numbers=1", "application/x-www-form-urlencoded", "", false, 0},
		{"missing array", "name=ready&empty=", "application/x-www-form-urlencoded", "", false, 0},
		{"duplicate scalar", "name=ready&name=again&empty=&numbers=1", "application/x-www-form-urlencoded", "", false, 0},
		{"bad integer", "name=ready&empty=&numbers=x", "application/x-www-form-urlencoded", "", false, 0},
		{"overflow", "name=ready&empty=&numbers=1&small=2147483648", "application/x-www-form-urlencoded", "", false, 0},
		{"negative unsigned", "name=ready&empty=&numbers=1&unsigned=-1", "application/x-www-form-urlencoded", "", false, 0},
		{"bad boolean", "name=ready&empty=&numbers=1&flag=perhaps", "application/x-www-form-urlencoded", "", false, 0},
		{"bad bytes", "name=ready&empty=&numbers=1&data=%%%", "application/x-www-form-urlencoded", "", false, 0},
		{"bad base64", "name=ready&empty=&numbers=1&data=invalid", "application/x-www-form-urlencoded", "", false, 0},
		{"invalid UTF8", "name=%FF&empty=&numbers=1", "application/x-www-form-urlencoded", "", false, 0},
		{"invalid escape", "name=%QQ&empty=&numbers=1", "application/x-www-form-urlencoded", "", false, 0},
		{"wrong media type", "name=ready&empty=&numbers=1", "application/json", "", false, 0},
		{"empty body", "", "application/x-www-form-urlencoded", "", false, 0},
		{"native constraint", "name=ready&empty=&numbers=1&count=-1", "application/x-www-form-urlencoded", "", false, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/forms/west?page_number=2"+test.query, strings.NewReader(test.body))
			request.Header.Set("Content-Type", test.mediaType)
			mux := goahttp.NewMuxer()
			mux.Handle(http.MethodPost, "/forms/{site}", func(_ http.ResponseWriter, incoming *http.Request) {
				payload, err := genserver.DecodeExchangeRequest(mux, nil)(incoming)
				if !test.valid {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
				require.Equal(t, "ready", string(payload.Name))
				require.Equal(t, "", payload.Empty)
				require.Equal(t, test.count, payload.Count)
				require.Equal(t, "west", payload.Site)
				require.Equal(t, 2, payload.Page)
			})
			mux.ServeHTTP(httptest.NewRecorder(), request)
		})
	}
}

func TestOptionalBodyMappedNamesAndJSON(t *testing.T) {
	for _, test := range []struct {
		name string
		body *genforms.Details
	}{
		{"absent", nil}, {"present empty message", &genforms.Details{Attempts: 2}},
	} {
		t.Run(test.name, func(t *testing.T) {
			input := &genforms.OptionalPayload{Body: test.body}
			request := httptest.NewRequest(http.MethodPost, "/optional", nil)
			require.NoError(t, genclient.EncodeOptionalRequest(nil)(request, input))
			payload, err := genserver.DecodeOptionalRequest(goahttp.NewMuxer(), nil)(request)
			require.NoError(t, err)
			require.Equal(t, input, payload)
		})
	}
	input := &genforms.MappedPayload{ClientID: "registered"}
	request := httptest.NewRequest(http.MethodPost, "/mapped", nil)
	require.NoError(t, genclient.EncodeMappedRequest(nil)(request, input))
	wire, err := io.ReadAll(request.Body)
	require.NoError(t, err)
	require.NoError(t, request.Body.Close())
	require.Equal(t, "client_id=registered", string(wire))
	incoming := httptest.NewRequest(http.MethodPost, "/mapped", strings.NewReader(string(wire)))
	incoming.Header.Set("Content-Type", request.Header.Get("Content-Type"))
	payload, err := genserver.DecodeMappedRequest(goahttp.NewMuxer(), nil)(incoming)
	require.NoError(t, err)
	require.Equal(t, input, payload)

	jsonInput := &genforms.Details{Message: "hello", Attempts: 2}
	jsonRequest := httptest.NewRequest(http.MethodPost, "/json", nil)
	require.NoError(t, genclient.EncodeJSONRequest(goahttp.RequestEncoder)(jsonRequest, jsonInput))
	jsonPayload, err := genserver.DecodeJSONRequest(goahttp.NewMuxer(), goahttp.RequestDecoder)(jsonRequest)
	require.NoError(t, err)
	require.Equal(t, jsonInput, jsonPayload)
}

func TestFormOpenAPI(t *testing.T) {
	for _, name := range []string{"openapi3.json", "openapi3.2.json"} {
		source, err := os.ReadFile("../http/" + name)
		require.NoError(t, err)
		var spec struct {
			Paths map[string]struct {
				Post struct {
					RequestBody struct {
						Required bool
						Content  map[string]json.RawMessage
					}
				}
			}
		}
		require.NoError(t, json.Unmarshal(source, &spec))
		require.Contains(t, spec.Paths["/mapped"].Post.RequestBody.Content, "application/x-www-form-urlencoded")
		require.Contains(t, spec.Paths["/json"].Post.RequestBody.Content, "application/json")
		require.True(t, spec.Paths["/mapped"].Post.RequestBody.Required)
		require.False(t, spec.Paths["/optional"].Post.RequestBody.Required)
		require.False(t, spec.Paths["/optional-json"].Post.RequestBody.Required)
	}
	source, err := os.ReadFile("../http/openapi.json")
	require.NoError(t, err)
	var spec struct {
		Paths map[string]struct {
			Post struct {
				Consumes   []string
				Parameters []struct {
					Name, In, Type, CollectionFormat string
					Required                         bool
				}
			}
		}
	}
	require.NoError(t, json.Unmarshal(source, &spec))
	require.Equal(t, []string{"application/x-www-form-urlencoded"}, spec.Paths["/mapped"].Post.Consumes)
	for _, parameter := range spec.Paths["/mapped"].Post.Parameters {
		require.Equal(t, "client_id", parameter.Name)
		require.Equal(t, "formData", parameter.In)
		require.True(t, parameter.Required)
	}
	for _, route := range []string{"/optional", "/optional-json"} {
		for _, parameter := range spec.Paths[route].Post.Parameters {
			require.False(t, parameter.Required)
		}
	}
	for _, parameter := range spec.Paths["/forms/{site}"].Post.Parameters {
		if parameter.Name == "numbers" {
			require.Equal(t, "multi", parameter.CollectionFormat)
		}
	}
}

func TestNativeClientBodyDefaultsMatchJSON(t *testing.T) {
	input := &genforms.Details{Message: "hello"}
	jsonRequest := httptest.NewRequest(http.MethodPost, "/json", nil)
	require.NoError(t, genclient.EncodeJSONRequest(goahttp.RequestEncoder)(jsonRequest, input))
	jsonPayload, err := genserver.DecodeJSONRequest(goahttp.NewMuxer(), goahttp.RequestDecoder)(jsonRequest)
	require.NoError(t, err)
	require.Equal(t, 3, jsonPayload.Attempts)

	formRequest := httptest.NewRequest(http.MethodPost, "/optional", nil)
	require.NoError(t, genclient.EncodeOptionalRequest(nil)(formRequest, &genforms.OptionalPayload{Body: input}))
	formPayload, err := genserver.DecodeOptionalRequest(goahttp.NewMuxer(), nil)(formRequest)
	require.NoError(t, err)
	require.Equal(t, jsonPayload, formPayload.Body)
}
