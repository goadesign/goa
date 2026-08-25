// This file contains the generated server tests written into the temporary
// module used to verify JSON-RPC parameter encoding and decoding.
package codegen_test

const paramsServerRuntimeTest = `package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	service "generated.local/gen/param_shapes"
	goahttp "goa.design/goa/v3/http"
	"goa.design/goa/v3/jsonrpc"
)

func TestRequestParamsDecodeToServiceValues(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/rpc", nil)

	text, err := DecodeTextRequest(nil, goahttp.RequestDecoder)(request, rawParams("[\"value\"]"))
	require.NoError(t, err)
	require.Equal(t, "value", text)

	alias, err := DecodeAliasRequest(nil, goahttp.RequestDecoder)(request, rawParams("[\"value\"]"))
	require.NoError(t, err)
	require.Equal(t, service.Alias("value"), alias)

	anything, err := DecodeAnythingRequest(nil, goahttp.RequestDecoder)(request, rawParams("[{\"name\":\"value\"}]"))
	require.NoError(t, err)
	require.Equal(t, map[string]any{"name": "value"}, anything)

	bytes, err := DecodeBytesRequest(nil, goahttp.RequestDecoder)(request, rawParams("[\"dmFsdWU=\"]"))
	require.NoError(t, err)
	require.Equal(t, []byte("value"), bytes)

	object, err := DecodeObjectRequest(nil, goahttp.RequestDecoder)(request, rawParams("{\"name\":\"value\"}"))
	require.NoError(t, err)
	require.Equal(t, &service.Object{Name: "value"}, object)

	array, err := DecodeArrayRequest(nil, goahttp.RequestDecoder)(request, rawParams("[\"one\",\"two\"]"))
	require.NoError(t, err)
	require.Equal(t, []string{"one", "two"}, array)

	values, err := DecodeMapRequest(nil, goahttp.RequestDecoder)(request, rawParams("{\"one\":1}"))
	require.NoError(t, err)
	require.Equal(t, map[string]int{"one": 1}, values)

	optional, err := DecodeOptionalTextRequest(nil, goahttp.RequestDecoder)(request, rawParams(""))
	require.NoError(t, err)
	require.Nil(t, optional.Value)

	optional, err = DecodeOptionalTextRequest(nil, goahttp.RequestDecoder)(request, rawParams("[null]"))
	require.NoError(t, err)
	require.Nil(t, optional.Value)

	optional, err = DecodeOptionalTextRequest(nil, goahttp.RequestDecoder)(request, rawParams("[\"\"]"))
	require.NoError(t, err)
	require.NotNil(t, optional.Value)
	require.Empty(t, *optional.Value)
}

func TestOptionalDirectRequestParamsDecodeAbsentAndPresent(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/rpc", nil)

	object, err := DecodeOptionalObjectRequest(nil, goahttp.RequestDecoder)(request, rawParams(""))
	require.NoError(t, err)
	require.Nil(t, object.Value)
	object, err = DecodeOptionalObjectRequest(nil, goahttp.RequestDecoder)(request, rawParams(` + "`" + `{"name":"value"}` + "`" + `))
	require.NoError(t, err)
	require.Equal(t, &service.Object{Name: "value"}, object.Value)

	array, err := DecodeOptionalArrayRequest(nil, goahttp.RequestDecoder)(request, rawParams(""))
	require.NoError(t, err)
	require.Nil(t, array.Value)
	array, err = DecodeOptionalArrayRequest(nil, goahttp.RequestDecoder)(request, rawParams(` + "`" + `[{"name":"value"}]` + "`" + `))
	require.NoError(t, err)
	require.Equal(t, service.Objects{&service.Object{Name: "value"}}, array.Value)
	array, err = DecodeOptionalArrayRequest(nil, goahttp.RequestDecoder)(request, rawParams("[]"))
	require.NoError(t, err)
	require.NotNil(t, array.Value)
	require.Empty(t, array.Value)

	values, err := DecodeOptionalMapRequest(nil, goahttp.RequestDecoder)(request, rawParams(""))
	require.NoError(t, err)
	require.Nil(t, values.Value)
	values, err = DecodeOptionalMapRequest(nil, goahttp.RequestDecoder)(request, rawParams(` + "`" + `{"one":{"name":"value"}}` + "`" + `))
	require.NoError(t, err)
	require.Equal(t, service.ObjectMap{"one": &service.Object{Name: "value"}}, values.Value)
	values, err = DecodeOptionalMapRequest(nil, goahttp.RequestDecoder)(request, rawParams("{}"))
	require.NoError(t, err)
	require.NotNil(t, values.Value)
	require.Empty(t, values.Value)

	choice, err := DecodeOptionalUnionRequest(nil, goahttp.RequestDecoder)(request, rawParams(""))
	require.NoError(t, err)
	require.Empty(t, choice.Value.Kind())
	choice, err = DecodeOptionalUnionRequest(nil, goahttp.RequestDecoder)(request, rawParams(` + "`" + `{"type":"name","value":"value"}` + "`" + `))
	require.NoError(t, err)
	name, ok := choice.Value.AsName()
	require.True(t, ok)
	require.Equal(t, "value", string(name))
	choice, err = DecodeOptionalUnionRequest(nil, goahttp.RequestDecoder)(request, rawParams(` + "`" + `{"type":"inactive","value":{}}` + "`" + `))
	require.NoError(t, err)
	inactive, ok := choice.Value.AsInactive()
	require.True(t, ok)
	require.Equal(t, &service.Empty{}, inactive)
}

func TestDefaultedSelectedRequestParamsPreserveAbsenceAndEmptyValues(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/rpc", nil)

	text, err := DecodeDefaultTextRequest(nil, goahttp.RequestDecoder)(request, rawParams(""))
	require.NoError(t, err)
	require.Equal(t, service.DefaultStreamData("fallback"), text.Value)

	text, err = DecodeDefaultTextRequest(nil, goahttp.RequestDecoder)(request, rawParams("[null]"))
	require.NoError(t, err)
	require.Equal(t, service.DefaultStreamData("fallback"), text.Value)

	text, err = DecodeDefaultTextRequest(nil, goahttp.RequestDecoder)(request, rawParams("[\"\"]"))
	require.NoError(t, err)
	require.Empty(t, text.Value)

	values, err := DecodeDefaultArrayRequest(nil, goahttp.RequestDecoder)(request, rawParams(""))
	require.NoError(t, err)
	require.Equal(t, service.DefaultValues{"fallback"}, values.Value)

	values, err = DecodeDefaultArrayRequest(nil, goahttp.RequestDecoder)(request, rawParams("[]"))
	require.NoError(t, err)
	require.NotNil(t, values.Value)
	require.Empty(t, values.Value)
}

func TestRequiredPrimitiveRejectsInvalidPositionalParams(t *testing.T) {
	tests := []struct {
		name   string
		params string
		error  string
	}{
		{name: "object", params: ` + "`" + `{}` + "`" + `, error: "params must be an array with exactly one value"},
		{name: "empty", params: ` + "`" + `[]` + "`" + `, error: "params must be an array with exactly one value"},
		{name: "multiple", params: ` + "`" + `["one","two"]` + "`" + `, error: "params must be an array with exactly one value"},
		{name: "null", params: ` + "`" + `[null]` + "`" + `, error: "missing required payload"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/rpc", nil)
			_, err := DecodeTextRequest(nil, goahttp.RequestDecoder)(request, rawParams(test.params))
			require.ErrorContains(t, err, test.error)
		})
	}
}

func TestOptionalPrimitiveRejectsInvalidPresentParams(t *testing.T) {
	for _, params := range []string{"{}", "[]", "[\"one\",\"two\"]"} {
		request := httptest.NewRequest(http.MethodPost, "/rpc", nil)
		_, err := DecodeOptionalTextRequest(nil, goahttp.RequestDecoder)(request, rawParams(params))
		require.ErrorContains(t, err, "params must be an array with exactly one value")
	}
	request := httptest.NewRequest(http.MethodPost, "/rpc", nil)
	_, err := DecodeOptionalTextRequest(nil, goahttp.RequestDecoder)(request, rawParams("[\"BAD\"]"))
	require.ErrorContains(t, err, "must match")
}

func TestStreamWritersKeepTheirJSONRPCShape(t *testing.T) {
	tests := []struct {
		name string
		send func(*httptest.ResponseRecorder) error
		want string
	}{
		{
			name: "string",
			send: func(recorder *httptest.ResponseRecorder) error {
				return (&StreamTextServerStream{sseServerStream: sseServerStream{w: recorder, encoder: goahttp.ResponseEncoder}}).Send("value")
			},
			want: ` + "`" + `{"jsonrpc":"2.0","method":"stream_text","params":["value"]}` + "`" + `,
		},
		{
			name: "any",
			send: func(recorder *httptest.ResponseRecorder) error {
				return (&StreamAnyServerStream{sseServerStream: sseServerStream{w: recorder, encoder: goahttp.ResponseEncoder}}).Send(map[string]any{"name": "value"})
			},
			want: ` + "`" + `{"jsonrpc":"2.0","method":"stream_any","params":[{"name":"value"}]}` + "`" + `,
		},
		{
			name: "structured array",
			send: func(recorder *httptest.ResponseRecorder) error {
				return (&StreamArrayServerStream{sseServerStream: sseServerStream{w: recorder, encoder: goahttp.ResponseEncoder}}).Send([]string{"one", "two"})
			},
			want: ` + "`" + `{"jsonrpc":"2.0","method":"stream_array","params":["one","two"]}` + "`" + `,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			require.NoError(t, test.send(recorder))
			body := recorder.Body.String()
			require.NotContains(t, body, "event:")
			message := strings.TrimSuffix(strings.TrimPrefix(body, "data: "), "\n\n")
			require.JSONEq(t, test.want, message)
		})
	}
}

func TestMappedStreamWritersUseCompleteJSONRPCData(t *testing.T) {
	emptyID := service.EventID("")
	emptyType := service.EventType("")
	zero := service.Retry(0)
	recorder := httptest.NewRecorder()
	stream := &StreamAliasFieldsServerStream{sseServerStream: sseServerStream{w: recorder, encoder: goahttp.ResponseEncoder}}
	err := stream.Send(&service.AliasEvent{
		Data:      service.Alias("value"),
		EventID:   &emptyID,
		EventType: &emptyType,
		Retry:     &zero,
	})
	require.NoError(t, err)
	require.Equal(t,
		"id: \nevent: \nretry: 0\ndata: {\"jsonrpc\":\"2.0\",\"method\":\"stream_alias_fields\",\"params\":[\"value\"]}\n\n",
		recorder.Body.String(),
	)

	objectRecorder := httptest.NewRecorder()
	objectStream := &StreamObjectFieldsServerStream{sseServerStream: sseServerStream{w: objectRecorder, encoder: goahttp.ResponseEncoder}}
	err = objectStream.Send(&service.ObjectEvent{
		Data:      &service.Details{Name: "value"},
		EventID:   "item-1",
		EventType: "update",
		Retry:     0,
	})
	require.NoError(t, err)
	require.Equal(t,
		"id: item-1\nevent: update\nretry: 0\ndata: {\"jsonrpc\":\"2.0\",\"method\":\"stream_object_fields\",\"params\":{\"name\":\"value\"}}\n\n",
		objectRecorder.Body.String(),
	)
}

func TestOptionalPrimitiveStreamWriterPreservesNilAndEmpty(t *testing.T) {
	nilRecorder := httptest.NewRecorder()
	nilStream := &StreamOptionalTextServerStream{sseServerStream: sseServerStream{w: nilRecorder, encoder: goahttp.ResponseEncoder}}
	require.NoError(t, nilStream.Send(&service.OptionalTextEvent{}))
	require.Equal(t,
		"data: {\"jsonrpc\":\"2.0\",\"method\":\"stream_optional_text\",\"params\":[null]}\n\n",
		nilRecorder.Body.String(),
	)

	empty := ""
	emptyRecorder := httptest.NewRecorder()
	emptyStream := &StreamOptionalTextServerStream{sseServerStream: sseServerStream{w: emptyRecorder, encoder: goahttp.ResponseEncoder}}
	require.NoError(t, emptyStream.Send(&service.OptionalTextEvent{Data: &empty}))
	require.Equal(t,
		"data: {\"jsonrpc\":\"2.0\",\"method\":\"stream_optional_text\",\"params\":[\"\"]}\n\n",
		emptyRecorder.Body.String(),
	)
}

func TestOptionalDirectStreamWritersOmitOnlyNilParams(t *testing.T) {
	tests := []struct {
		name string
		sendNil func(*httptest.ResponseRecorder) error
		sendValue func(*httptest.ResponseRecorder) error
		want string
	}{
		{
			name: "named object",
			sendNil: func(recorder *httptest.ResponseRecorder) error {
				return (&StreamOptionalObjectServerStream{sseServerStream: sseServerStream{w: recorder, encoder: goahttp.ResponseEncoder}}).Send(&service.OptionalObjectEvent{})
			},
			sendValue: func(recorder *httptest.ResponseRecorder) error {
				return (&StreamOptionalObjectServerStream{sseServerStream: sseServerStream{w: recorder, encoder: goahttp.ResponseEncoder}}).Send(&service.OptionalObjectEvent{Data: &service.Object{Name: "value"}})
			},
			want: ` + "`" + `{"jsonrpc":"2.0","method":"stream_optional_object","params":{"name":"value"}}` + "`" + `,
		},
		{
			name: "named array",
			sendNil: func(recorder *httptest.ResponseRecorder) error {
				return (&StreamOptionalArrayServerStream{sseServerStream: sseServerStream{w: recorder, encoder: goahttp.ResponseEncoder}}).Send(&service.OptionalArrayEvent{})
			},
			sendValue: func(recorder *httptest.ResponseRecorder) error {
				return (&StreamOptionalArrayServerStream{sseServerStream: sseServerStream{w: recorder, encoder: goahttp.ResponseEncoder}}).Send(&service.OptionalArrayEvent{Data: service.Objects{&service.Object{Name: "value"}}})
			},
			want: ` + "`" + `{"jsonrpc":"2.0","method":"stream_optional_array","params":[{"name":"value"}]}` + "`" + `,
		},
		{
			name: "named empty array",
			sendNil: func(recorder *httptest.ResponseRecorder) error {
				return (&StreamOptionalArrayServerStream{sseServerStream: sseServerStream{w: recorder, encoder: goahttp.ResponseEncoder}}).Send(&service.OptionalArrayEvent{})
			},
			sendValue: func(recorder *httptest.ResponseRecorder) error {
				return (&StreamOptionalArrayServerStream{sseServerStream: sseServerStream{w: recorder, encoder: goahttp.ResponseEncoder}}).Send(&service.OptionalArrayEvent{Data: service.Objects{}})
			},
			want: ` + "`" + `{"jsonrpc":"2.0","method":"stream_optional_array","params":[]}` + "`" + `,
		},
		{
			name: "named map",
			sendNil: func(recorder *httptest.ResponseRecorder) error {
				return (&StreamOptionalMapServerStream{sseServerStream: sseServerStream{w: recorder, encoder: goahttp.ResponseEncoder}}).Send(&service.OptionalMapEvent{})
			},
			sendValue: func(recorder *httptest.ResponseRecorder) error {
				return (&StreamOptionalMapServerStream{sseServerStream: sseServerStream{w: recorder, encoder: goahttp.ResponseEncoder}}).Send(&service.OptionalMapEvent{Data: service.ObjectMap{"one": &service.Object{Name: "value"}}})
			},
			want: ` + "`" + `{"jsonrpc":"2.0","method":"stream_optional_map","params":{"one":{"name":"value"}}}` + "`" + `,
		},
		{
			name: "named empty map",
			sendNil: func(recorder *httptest.ResponseRecorder) error {
				return (&StreamOptionalMapServerStream{sseServerStream: sseServerStream{w: recorder, encoder: goahttp.ResponseEncoder}}).Send(&service.OptionalMapEvent{})
			},
			sendValue: func(recorder *httptest.ResponseRecorder) error {
				return (&StreamOptionalMapServerStream{sseServerStream: sseServerStream{w: recorder, encoder: goahttp.ResponseEncoder}}).Send(&service.OptionalMapEvent{Data: service.ObjectMap{}})
			},
			want: ` + "`" + `{"jsonrpc":"2.0","method":"stream_optional_map","params":{}}` + "`" + `,
		},
		{
			name: "named union",
			sendNil: func(recorder *httptest.ResponseRecorder) error {
				return (&StreamOptionalUnionServerStream{sseServerStream: sseServerStream{w: recorder, encoder: goahttp.ResponseEncoder}}).Send(&service.OptionalUnionEvent{})
			},
			sendValue: func(recorder *httptest.ResponseRecorder) error {
				event := &service.OptionalUnionEvent{}
				event.Data.SetName("value")
				return (&StreamOptionalUnionServerStream{sseServerStream: sseServerStream{w: recorder, encoder: goahttp.ResponseEncoder}}).Send(event)
			},
			want: ` + "`" + `{"jsonrpc":"2.0","method":"stream_optional_union","params":{"type":"name","value":"value"}}` + "`" + `,
		},
		{
			name: "named union empty message",
			sendNil: func(recorder *httptest.ResponseRecorder) error {
				return (&StreamOptionalUnionServerStream{sseServerStream: sseServerStream{w: recorder, encoder: goahttp.ResponseEncoder}}).Send(&service.OptionalUnionEvent{})
			},
			sendValue: func(recorder *httptest.ResponseRecorder) error {
				event := &service.OptionalUnionEvent{}
				event.Data.SetInactive(&service.Empty{})
				return (&StreamOptionalUnionServerStream{sseServerStream: sseServerStream{w: recorder, encoder: goahttp.ResponseEncoder}}).Send(event)
			},
			want: ` + "`" + `{"jsonrpc":"2.0","method":"stream_optional_union","params":{"type":"inactive","value":{}}}` + "`" + `,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			nilRecorder := httptest.NewRecorder()
			require.NoError(t, test.sendNil(nilRecorder))
			require.NotContains(t, nilRecorder.Body.String(), ` + "`" + `"params"` + "`" + `)

			valueRecorder := httptest.NewRecorder()
			require.NoError(t, test.sendValue(valueRecorder))
			message := strings.TrimSuffix(strings.TrimPrefix(valueRecorder.Body.String(), "data: "), "\n\n")
			require.JSONEq(t, test.want, message)
		})
	}
}

func TestResumeRequestReadsLastEventIDBeforeValidation(t *testing.T) {
	tests := []struct {
		name      string
		decode    func(*http.Request, *jsonrpc.RawRequest) (any, error)
		header    []string
		wantValue *string
		wantError string
	}{
		{
			name: "required empty",
			decode: func(request *http.Request, raw *jsonrpc.RawRequest) (any, error) {
				return DecodeRequiredResumeRequest(nil, goahttp.RequestDecoder)(request, raw)
			},
			header:    []string{""},
			wantValue: stringPointer(""),
		},
		{
			name: "required absent",
			decode: func(request *http.Request, raw *jsonrpc.RawRequest) (any, error) {
				return DecodeRequiredResumeRequest(nil, goahttp.RequestDecoder)(request, raw)
			},
			wantError: "is missing from header",
		},
		{
			name: "optional absent",
			decode: func(request *http.Request, raw *jsonrpc.RawRequest) (any, error) {
				return DecodeOptionalResumeRequest(nil, goahttp.RequestDecoder)(request, raw)
			},
		},
		{
			name: "default absent",
			decode: func(request *http.Request, raw *jsonrpc.RawRequest) (any, error) {
				return DecodeDefaultResumeRequest(nil, goahttp.RequestDecoder)(request, raw)
			},
			wantValue: stringPointer("cursor"),
		},
		{
			name: "default empty",
			decode: func(request *http.Request, raw *jsonrpc.RawRequest) (any, error) {
				return DecodeDefaultResumeRequest(nil, goahttp.RequestDecoder)(request, raw)
			},
			header:    []string{""},
			wantValue: stringPointer(""),
		},
		{
			name: "optional empty",
			decode: func(request *http.Request, raw *jsonrpc.RawRequest) (any, error) {
				return DecodeOptionalResumeRequest(nil, goahttp.RequestDecoder)(request, raw)
			},
			header:    []string{""},
			wantValue: stringPointer(""),
		},
		{
			name: "optional nul",
			decode: func(request *http.Request, raw *jsonrpc.RawRequest) (any, error) {
				return DecodeOptionalResumeRequest(nil, goahttp.RequestDecoder)(request, raw)
			},
			header:    []string{"one\x00two"},
			wantError: "last_event_id",
		},
		{
			name: "optional line feed",
			decode: func(request *http.Request, raw *jsonrpc.RawRequest) (any, error) {
				return DecodeOptionalResumeRequest(nil, goahttp.RequestDecoder)(request, raw)
			},
			header:    []string{"one\ntwo"},
			wantError: "last_event_id",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/rpc", nil)
			if test.header != nil {
				request.Header["Last-Event-Id"] = test.header
			}
			value, err := test.decode(request, rawParams(""))
			if test.wantError != "" {
				require.ErrorContains(t, err, test.wantError)
				return
			}
			require.NoError(t, err)
			switch payload := value.(type) {
			case *service.RequiredResumePayload:
				require.Equal(t, *test.wantValue, payload.LastEventID)
			case *service.OptionalResumePayload:
				require.Equal(t, test.wantValue, payload.LastEventID)
			case *service.DefaultResumePayload:
				require.Equal(t, service.ResumeID(*test.wantValue), payload.LastEventID)
			default:
				t.Fatalf("unexpected payload type %T", value)
			}
		})
	}
}

func stringPointer(value string) *string {
	return &value
}

func rawParams(params string) *jsonrpc.RawRequest {
	return &jsonrpc.RawRequest{Params: json.RawMessage(params)}
}
`
