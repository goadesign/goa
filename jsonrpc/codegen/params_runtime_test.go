// This file renders JSON-RPC methods into a temporary Go module. The generated
// tests check positional primitive values, direct structured values, and the
// difference between an absent optional value and an authored empty value.
package codegen_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	goacodegen "goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
	jsonrpccodegen "goa.design/goa/v3/jsonrpc/codegen"
)

// TestGeneratedJSONRPCParams renders and runs the client, server, and stream
// code for single values and for objects, arrays, and maps.
func TestGeneratedJSONRPCParams(t *testing.T) {
	dir := renderParamsRuntimeModule(t)
	clientDir := filepath.Join(dir, "jsonrpc", "param_shapes", "client")
	serverDir := filepath.Join(dir, "jsonrpc", "param_shapes", "server")
	require.NoError(t, os.WriteFile(filepath.Join(clientDir, "params_runtime_test.go"), []byte(paramsClientRuntimeTest), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(serverDir, "params_runtime_test.go"), []byte(paramsServerRuntimeTest), 0o600))
	streamSource, err := os.ReadFile(filepath.Join(clientDir, "stream.go"))
	require.NoError(t, err)
	require.Contains(t, string(streamSource), "decodeError(response.Error)")
	require.NotContains(t, string(streamSource), "decodeError(response.Error, response.ID == nil)")
	require.NotContains(t, string(streamSource), "decodeError(response *jsonrpc.RawErrorResponse, ")
	runParamsRuntimeTests(t, dir)
}

// renderParamsRuntimeModule writes generated service, client, and server
// packages without changing any generated directory in this repository.
func renderParamsRuntimeModule(t *testing.T) string {
	t.Helper()
	root := expr.RunDSL(t, paramsRuntimeDSL)
	generation, err := goacodegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	httpPlans, err := httpcodegen.NewJSONRPCPlans(generation, httpcodegen.PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	jsonPlans, err := jsonrpccodegen.NewPlans(generation, jsonrpccodegen.PlanInput{
		Root:    root,
		Service: servicePlan,
		HTTP:    httpPlans[0],
	})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, httpPlans[0].Link())
	require.NoError(t, jsonPlans[0].Link())

	files, err := service.Files(servicePlan)
	require.NoError(t, err)
	files = append(files, jsonPlans[0].ClientFiles()...)
	files = append(files, jsonPlans[0].ServerFiles()...)
	files = append(files, jsonPlans[0].ClientTypeFiles()...)
	files = append(files, jsonPlans[0].ServerTypeFiles()...)
	files = append(files, jsonPlans[0].PathFiles()...)
	base := t.TempDir()
	for _, file := range files {
		_, err := file.Render(base)
		require.NoError(t, err)
	}

	moduleDir := filepath.Join(base, goacodegen.Gendir)
	workingDir, err := os.Getwd()
	require.NoError(t, err)
	repository := filepath.Clean(filepath.Join(workingDir, "..", ".."))
	module := fmt.Sprintf(
		"module generated.local/gen\n\ngo 1.25\n\nrequire goa.design/goa/v3 v3.0.0\n\nreplace goa.design/goa/v3 => %s\n",
		repository,
	)
	require.NoError(t, os.WriteFile(filepath.Join(moduleDir, "go.mod"), []byte(module), 0o600))
	return moduleDir
}

const paramsClientRuntimeTest = `package client

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	service "generated.local/gen/param_shapes"
	goahttp "goa.design/goa/v3/http"
	"goa.design/goa/v3/jsonrpc"
)

func TestRequestParamsKeepTheirJSONRPCShape(t *testing.T) {
	tests := []struct {
		name   string
		value  any
		encode func(func(*http.Request) goahttp.Encoder) func(*http.Request, any) (string, error)
		want   string
	}{
		{name: "string", value: "value", encode: EncodeTextRequest, want: ` + "`" + `{"params":["value"]}` + "`" + `},
		{name: "primitive alias", value: service.Alias("value"), encode: EncodeAliasRequest, want: ` + "`" + `{"params":["value"]}` + "`" + `},
		{name: "any", value: map[string]any{"name": "value"}, encode: EncodeAnythingRequest, want: ` + "`" + `{"params":[{"name":"value"}]}` + "`" + `},
		{name: "bytes", value: []byte("value"), encode: EncodeBytesRequest, want: ` + "`" + `{"params":["dmFsdWU="]}` + "`" + `},
		{name: "object", value: &service.Object{Name: "value"}, encode: EncodeObjectRequest, want: ` + "`" + `{"params":{"name":"value"}}` + "`" + `},
		{name: "array", value: []string{"one", "two"}, encode: EncodeArrayRequest, want: ` + "`" + `{"params":["one","two"]}` + "`" + `},
		{name: "map", value: map[string]int{"one": 1}, encode: EncodeMapRequest, want: ` + "`" + `{"params":{"one":1}}` + "`" + `},
		{name: "optional primitive body", value: &service.OptionalTextPayload{Value: stringPointer("value")}, encode: EncodeOptionalTextRequest, want: ` + "`" + `{"params":["value"]}` + "`" + `},
		{name: "optional primitive null", value: &service.OptionalTextPayload{}, encode: EncodeOptionalTextRequest, want: ` + "`" + `{"params":[null]}` + "`" + `},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/rpc", nil)
			_, err := test.encode(goahttp.RequestEncoder)(request, test.value)
			require.NoError(t, err)
			body, err := io.ReadAll(request.Body)
			require.NoError(t, err)
			require.JSONEq(t, test.want, paramsOnly(body))
		})
	}

}

func TestOptionalDirectRequestParamsOmitOnlyNilValues(t *testing.T) {
	tests := []struct {
		name   string
		nilValue any
		value  any
		encode func(func(*http.Request) goahttp.Encoder) func(*http.Request, any) (string, error)
		want   string
	}{
		{
			name: "named object",
			nilValue: &service.OptionalObjectPayload{},
			value: &service.OptionalObjectPayload{Value: &service.Object{Name: "value"}},
			encode: EncodeOptionalObjectRequest,
			want: ` + "`" + `{"params":{"name":"value"}}` + "`" + `,
		},
		{
			name: "named array",
			nilValue: &service.OptionalArrayPayload{},
			value: &service.OptionalArrayPayload{Value: service.Objects{&service.Object{Name: "value"}}},
			encode: EncodeOptionalArrayRequest,
			want: ` + "`" + `{"params":[{"name":"value"}]}` + "`" + `,
		},
		{
			name: "named empty array",
			nilValue: &service.OptionalArrayPayload{},
			value: &service.OptionalArrayPayload{Value: service.Objects{}},
			encode: EncodeOptionalArrayRequest,
			want: ` + "`" + `{"params":[]}` + "`" + `,
		},
		{
			name: "named map",
			nilValue: &service.OptionalMapPayload{},
			value: &service.OptionalMapPayload{Value: service.ObjectMap{"one": &service.Object{Name: "value"}}},
			encode: EncodeOptionalMapRequest,
			want: ` + "`" + `{"params":{"one":{"name":"value"}}}` + "`" + `,
		},
		{
			name: "named empty map",
			nilValue: &service.OptionalMapPayload{},
			value: &service.OptionalMapPayload{Value: service.ObjectMap{}},
			encode: EncodeOptionalMapRequest,
			want: ` + "`" + `{"params":{}}` + "`" + `,
		},
		{
			name: "named union",
			nilValue: &service.OptionalUnionPayload{},
			value: optionalUnionPayload("value"),
			encode: EncodeOptionalUnionRequest,
			want: ` + "`" + `{"params":{"type":"name","value":"value"}}` + "`" + `,
		},
		{
			name: "named union empty message",
			nilValue: &service.OptionalUnionPayload{},
			value: optionalUnionInactivePayload(),
			encode: EncodeOptionalUnionRequest,
			want: ` + "`" + `{"params":{"type":"inactive","value":{}}}` + "`" + `,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/rpc", nil)
			_, err := test.encode(goahttp.RequestEncoder)(request, test.nilValue)
			require.NoError(t, err)
			body, err := io.ReadAll(request.Body)
			require.NoError(t, err)
			require.NotContains(t, string(body), ` + "`" + `"params"` + "`" + `)

			request = httptest.NewRequest(http.MethodPost, "/rpc", nil)
			_, err = test.encode(goahttp.RequestEncoder)(request, test.value)
			require.NoError(t, err)
			body, err = io.ReadAll(request.Body)
			require.NoError(t, err)
			require.JSONEq(t, test.want, paramsOnly(body))
		})
	}

}

func TestDefaultedSelectedRequestClientsPreserveExplicitValues(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/rpc", nil)
	_, err := EncodeDefaultTextRequest(goahttp.RequestEncoder)(request, &service.DefaultTextPayload{
		Value: service.DefaultStreamData(""),
	})
	require.NoError(t, err)
	body, err := io.ReadAll(request.Body)
	require.NoError(t, err)
	require.JSONEq(t, ` + "`" + `{"params":[""]}` + "`" + `, paramsOnly(body))

	request = httptest.NewRequest(http.MethodPost, "/rpc", nil)
	_, err = EncodeDefaultArrayRequest(goahttp.RequestEncoder)(request, &service.DefaultArrayPayload{})
	require.NoError(t, err)
	body, err = io.ReadAll(request.Body)
	require.NoError(t, err)
	require.NotContains(t, string(body), ` + "`" + `"params"` + "`" + `)

	request = httptest.NewRequest(http.MethodPost, "/rpc", nil)
	_, err = EncodeDefaultArrayRequest(goahttp.RequestEncoder)(request, &service.DefaultArrayPayload{
		Value: service.DefaultValues{},
	})
	require.NoError(t, err)
	body, err = io.ReadAll(request.Body)
	require.NoError(t, err)
	require.JSONEq(t, ` + "`" + `{"params":[]}` + "`" + `, paramsOnly(body))
}

func stringPointer(value string) *string {
	return &value
}

func optionalUnionPayload(value string) *service.OptionalUnionPayload {
	payload := &service.OptionalUnionPayload{}
	payload.Value.SetName(service.OptionalRequestValueBranchName(value))
	return payload
}

func optionalUnionInactivePayload() *service.OptionalUnionPayload {
	payload := &service.OptionalUnionPayload{}
	payload.Value.SetInactive(&service.Empty{})
	return payload
}

func TestStreamParamsKeepTheirJSONRPCShape(t *testing.T) {
	tests := []struct {
		name   string
		event  string
		receive func(*http.Response) (any, error)
		want   any
	}{
		{
			name: "string",
			event: "data: {\"jsonrpc\":\"2.0\",\"method\":\"stream_text\",\"params\":[\"value\"]}\n\n",
			receive: func(response *http.Response) (any, error) {
				return NewStreamTextStream(response, goahttp.ResponseDecoder, "request-1").RecvWithContext(context.Background())
			},
			want: "value",
		},
		{
			name: "any",
			event: "data: {\"jsonrpc\":\"2.0\",\"method\":\"stream_any\",\"params\":[{\"name\":\"value\"}]}\n\n",
			receive: func(response *http.Response) (any, error) {
				return NewStreamAnyStream(response, goahttp.ResponseDecoder, "request-1").RecvWithContext(context.Background())
			},
			want: map[string]any{"name": "value"},
		},
		{
			name: "structured array",
			event: "data: {\"jsonrpc\":\"2.0\",\"method\":\"stream_array\",\"params\":[\"one\",\"two\"]}\n\n",
			receive: func(response *http.Response) (any, error) {
				return NewStreamArrayStream(response, goahttp.ResponseDecoder, "request-1").RecvWithContext(context.Background())
			},
			want: []string{"one", "two"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(test.event))}
			value, err := test.receive(response)
			require.NoError(t, err)
			require.Equal(t, test.want, value)
		})
	}
}

func TestMappedStreamFieldsRebuildServiceResults(t *testing.T) {
	aliasResponse := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(
		"id: \nevent: \nretry: 0\ndata: {\"jsonrpc\":\"2.0\",\"method\":\"stream_alias_fields\",\"params\":[\"value\"]}\n\n",
	))}
	alias, err := NewStreamAliasFieldsStream(aliasResponse, goahttp.ResponseDecoder, "request-1").Recv()
	require.NoError(t, err)
	require.Equal(t, service.Alias("value"), alias.Data)
	require.NotNil(t, alias.EventID)
	require.Empty(t, *alias.EventID)
	require.NotNil(t, alias.EventType)
	require.Empty(t, *alias.EventType)
	require.NotNil(t, alias.Retry)
	require.Zero(t, *alias.Retry)

	objectResponse := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(
		"id: item-1\nevent: update\nretry: 0\ndata: {\"jsonrpc\":\"2.0\",\"method\":\"stream_object_fields\",\"params\":{\"name\":\"value\"}}\n\n",
	))}
	object, err := NewStreamObjectFieldsStream(objectResponse, goahttp.ResponseDecoder, "request-1").Recv()
	require.NoError(t, err)
	require.Equal(t, &service.Details{Name: "value"}, object.Data)
	require.Equal(t, "item-1", object.EventID)
	require.Equal(t, "update", object.EventType)
	require.Zero(t, object.Retry)
}

func TestMappedStreamFieldsApplyAuthoredDefaultsWhenLinesAreAbsent(t *testing.T) {
	response := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(
		"data: {\"jsonrpc\":\"2.0\",\"method\":\"stream_default_fields\",\"params\":[\"value\"]}\n\n",
	))}

	result, err := NewStreamDefaultFieldsStream(response, goahttp.ResponseDecoder, "request-1").Recv()

	require.NoError(t, err)
	require.Equal(t, service.DefaultEventID("fallback"), result.EventID)
	require.Equal(t, service.DefaultEventType("update"), result.EventType)
	require.Equal(t, service.DefaultRetry(0), result.Retry)
}

func TestOptionalPrimitiveStreamParamsPreservePresenceAndDefaults(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		params  string
		receive func(*http.Response) (any, error)
		check   func(*testing.T, any)
		wantErr string
	}{
		{
			name:   "optional omitted",
			method: "stream_optional_text",
			receive: func(response *http.Response) (any, error) {
				return NewStreamOptionalTextStream(response, goahttp.ResponseDecoder, "request-1").Recv()
			},
			check: func(t *testing.T, value any) {
				require.Nil(t, value.(*service.OptionalTextEvent).Data)
			},
		},
		{
			name:   "optional null",
			method: "stream_optional_text",
			params: ` + "`" + `[null]` + "`" + `,
			receive: func(response *http.Response) (any, error) {
				return NewStreamOptionalTextStream(response, goahttp.ResponseDecoder, "request-1").Recv()
			},
			check: func(t *testing.T, value any) {
				require.Nil(t, value.(*service.OptionalTextEvent).Data)
			},
		},
		{
			name:   "optional empty",
			method: "stream_optional_text",
			params: ` + "`" + `[""]` + "`" + `,
			receive: func(response *http.Response) (any, error) {
				return NewStreamOptionalTextStream(response, goahttp.ResponseDecoder, "request-1").Recv()
			},
			check: func(t *testing.T, value any) {
				result := value.(*service.OptionalTextEvent)
				require.NotNil(t, result.Data)
				require.Empty(t, *result.Data)
			},
		},
		{
			name:    "optional invalid",
			method:  "stream_optional_text",
			params:  ` + "`" + `["1"]` + "`" + `,
			receive: func(response *http.Response) (any, error) {
				return NewStreamOptionalTextStream(response, goahttp.ResponseDecoder, "request-1").Recv()
			},
			wantErr: "must match",
		},
		{
			name:   "default omitted",
			method: "stream_default_text",
			receive: func(response *http.Response) (any, error) {
				return NewStreamDefaultTextStream(response, goahttp.ResponseDecoder, "request-1").Recv()
			},
			check: func(t *testing.T, value any) {
				require.Equal(t, service.DefaultStreamData("fallback"), value.(*service.DefaultTextEvent).Data)
			},
		},
		{
			name:   "default null",
			method: "stream_default_text",
			params: ` + "`" + `[null]` + "`" + `,
			receive: func(response *http.Response) (any, error) {
				return NewStreamDefaultTextStream(response, goahttp.ResponseDecoder, "request-1").Recv()
			},
			check: func(t *testing.T, value any) {
				require.Equal(t, service.DefaultStreamData("fallback"), value.(*service.DefaultTextEvent).Data)
			},
		},
		{
			name:   "default empty",
			method: "stream_default_text",
			params: ` + "`" + `[""]` + "`" + `,
			receive: func(response *http.Response) (any, error) {
				return NewStreamDefaultTextStream(response, goahttp.ResponseDecoder, "request-1").Recv()
			},
			check: func(t *testing.T, value any) {
				require.Empty(t, value.(*service.DefaultTextEvent).Data)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			message := ` + "`" + `{"jsonrpc":"2.0","method":"` + "`" + ` + test.method + ` + "`" + `"` + "`" + `
			if test.params != "" {
				message += ` + "`" + `,"params":` + "`" + ` + test.params
			}
			message += "}\n\n"
			response := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("data: " + message))}
			value, err := test.receive(response)
			if test.wantErr != "" {
				require.ErrorContains(t, err, test.wantErr)
				return
			}
			require.NoError(t, err)
			test.check(t, value)
		})
	}
}

func TestOptionalDirectStreamParamsRoundTripAbsentAndPresent(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		params  string
		receive func(*http.Response) (any, error)
		check   func(*testing.T, any, bool)
	}{
		{
			name: "named object",
			method: "stream_optional_object",
			params: ` + "`" + `{"name":"value"}` + "`" + `,
			receive: func(response *http.Response) (any, error) {
				return NewStreamOptionalObjectStream(response, goahttp.ResponseDecoder, "request-1").Recv()
			},
			check: func(t *testing.T, value any, present bool) {
				result := value.(*service.OptionalObjectEvent)
				if present {
					require.Equal(t, &service.Object{Name: "value"}, result.Data)
				} else {
					require.Nil(t, result.Data)
				}
			},
		},
		{
			name: "named union empty message",
			method: "stream_optional_union",
			params: ` + "`" + `{"type":"inactive","value":{}}` + "`" + `,
			receive: func(response *http.Response) (any, error) {
				return NewStreamOptionalUnionStream(response, goahttp.ResponseDecoder, "request-1").Recv()
			},
			check: func(t *testing.T, value any, present bool) {
				result := value.(*service.OptionalUnionEvent)
				if present {
					inactive, ok := result.Data.AsInactive()
					require.True(t, ok)
					require.Equal(t, &service.Empty{}, inactive)
				} else {
					require.Empty(t, result.Data.Kind())
				}
			},
		},
		{
			name: "named array",
			method: "stream_optional_array",
			params: ` + "`" + `[{"name":"value"}]` + "`" + `,
			receive: func(response *http.Response) (any, error) {
				return NewStreamOptionalArrayStream(response, goahttp.ResponseDecoder, "request-1").Recv()
			},
			check: func(t *testing.T, value any, present bool) {
				result := value.(*service.OptionalArrayEvent)
				if present {
					require.Equal(t, service.Objects{&service.Object{Name: "value"}}, result.Data)
				} else {
					require.Nil(t, result.Data)
				}
			},
		},
		{
			name: "named map",
			method: "stream_optional_map",
			params: ` + "`" + `{"one":{"name":"value"}}` + "`" + `,
			receive: func(response *http.Response) (any, error) {
				return NewStreamOptionalMapStream(response, goahttp.ResponseDecoder, "request-1").Recv()
			},
			check: func(t *testing.T, value any, present bool) {
				result := value.(*service.OptionalMapEvent)
				if present {
					require.Equal(t, service.ObjectMap{"one": &service.Object{Name: "value"}}, result.Data)
				} else {
					require.Nil(t, result.Data)
				}
			},
		},
		{
			name: "named union",
			method: "stream_optional_union",
			params: ` + "`" + `{"type":"name","value":"value"}` + "`" + `,
			receive: func(response *http.Response) (any, error) {
				return NewStreamOptionalUnionStream(response, goahttp.ResponseDecoder, "request-1").Recv()
			},
			check: func(t *testing.T, value any, present bool) {
				result := value.(*service.OptionalUnionEvent)
				if present {
					name, ok := result.Data.AsName()
					require.True(t, ok)
					require.Equal(t, "value", string(name))
				} else {
					require.Empty(t, result.Data.Kind())
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, present := range []bool{false, true} {
				params := ""
				if present {
					params = ` + "`" + `,"params":` + "`" + ` + test.params
				}
				event := ` + "`" + `data: {"jsonrpc":"2.0","method":"` + "`" + ` + test.method + ` + "`" + `"` + "`" + ` + params + "}\n\n"
				response := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(event))}
				value, err := test.receive(response)
				require.NoError(t, err)
				test.check(t, value, present)
			}
		})
	}

	response := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(
		"data: {\"jsonrpc\":\"2.0\",\"method\":\"stream_optional_object\",\"params\":null}\n\n",
	))}
	_, err := NewStreamOptionalObjectStream(response, goahttp.ResponseDecoder, "request-1").Recv()
	require.ErrorContains(t, err, "invalid JSON-RPC notification")
}

func TestMappedStreamRejectsMalformedMessages(t *testing.T) {
	tests := []struct {
		name  string
		event string
		error string
	}{
		{name: "missing required outer field", event: "event: update\nretry: 0\ndata: {\"jsonrpc\":\"2.0\",\"method\":\"stream_object_fields\",\"params\":{\"name\":\"value\"}}\n\n", error: "has no id"},
		{name: "invalid required retry", event: "id: item-1\nevent: update\nretry: later\ndata: {\"jsonrpc\":\"2.0\",\"method\":\"stream_object_fields\",\"params\":{\"name\":\"value\"}}\n\n", error: "has no retry"},
		{name: "invalid data", event: "id: item-1\nevent: update\nretry: 0\ndata: {\"jsonrpc\":\"2.0\",\"method\":\"stream_alias_fields\",\"params\":[]}\n\n", error: "exactly one value"},
		{name: "missing required data", event: "id: item\nevent: update\nretry: 0\ndata: {\"jsonrpc\":\"2.0\",\"method\":\"stream_alias_fields\",\"params\":[null]}\n\n", error: ` + "`" + `"data" is missing` + "`" + `},
		{name: "data pattern", event: "id: item\nevent: update\nretry: 0\ndata: {\"jsonrpc\":\"2.0\",\"method\":\"stream_alias_fields\",\"params\":[\"BAD\"]}\n\n", error: "must match"},
		{name: "event ID pattern", event: "id: 123\nevent: update\nretry: 0\ndata: {\"jsonrpc\":\"2.0\",\"method\":\"stream_alias_fields\",\"params\":[\"value\"]}\n\n", error: "must match"},
		{name: "event type enum", event: "id: item\nevent: delete\nretry: 0\ndata: {\"jsonrpc\":\"2.0\",\"method\":\"stream_alias_fields\",\"params\":[\"value\"]}\n\n", error: "one of"},
		{name: "retry maximum", event: "id: item\nevent: update\nretry: 11\ndata: {\"jsonrpc\":\"2.0\",\"method\":\"stream_alias_fields\",\"params\":[\"value\"]}\n\n", error: "lesser or equal"},
		{name: "missing method", event: "data: {\"jsonrpc\":\"2.0\",\"params\":[\"value\"]}\n\n", error: "JSON-RPC response contains params"},
		{name: "null method", event: "data: {\"jsonrpc\":\"2.0\",\"method\":null,\"params\":[\"value\"]}\n\n", error: "invalid JSON-RPC notification"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(test.event))}
			var err error
			if strings.Contains(test.event, "stream_object_fields") {
				_, err = NewStreamObjectFieldsStream(response, goahttp.ResponseDecoder, "request-1").Recv()
			} else {
				_, err = NewStreamAliasFieldsStream(response, goahttp.ResponseDecoder, "request-1").Recv()
			}
			require.ErrorContains(t, err, test.error)
		})
	}
}

func TestStreamTerminalMessagesDoNotNeedPrivateEventNames(t *testing.T) {
	for _, message := range []string{
		` + "`" + `{"jsonrpc":"2.0","id":"request-1","result":null}` + "`" + `,
		` + "`" + `{"jsonrpc":"2.0","id":"request-1","error":{"code":-32603,"message":"failed"}}` + "`" + `,
	} {
		response := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("data: " + message + "\n\n"))}
		_, err := NewStreamTextStream(response, goahttp.ResponseDecoder, "request-1").Recv()
		if strings.Contains(message, ` + "`" + `"result"` + "`" + `) {
			require.ErrorIs(t, err, io.EOF)
		} else {
			var response *jsonrpc.RawErrorResponse
			require.ErrorAs(t, err, &response)
			require.Equal(t, -32603, response.Code)
			require.Equal(t, "failed", response.Message)
		}
	}
}

func TestStreamTerminalSuccessRequiresNullResult(t *testing.T) {
	response := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(
		"data: " + ` + "`" + `{"jsonrpc":"2.0","id":"request-1","result":"unexpected"}` + "`" + ` + "\n\n",
	))}

	_, err := NewStreamTextStream(response, goahttp.ResponseDecoder, "request-1").Recv()

	require.EqualError(t, err, "JSON-RPC stream completion result must be null")
}

func TestUnknownUnaryErrorPreservesJSONRPCDetails(t *testing.T) {
	response := &http.Response{
		StatusCode: http.StatusOK,
		Body: io.NopCloser(strings.NewReader(
			` + "`" + `{"jsonrpc":"2.0","id":"request-1","error":{"code":-32099,"message":"upstream failed","data":{"reason":"overload"}}}` + "`" + `,
		)),
	}

	result, err := DecodeTextResponse(goahttp.ResponseDecoder, false)(response, "request-1")

	require.Nil(t, result)
	var rpcError *jsonrpc.RawErrorResponse
	require.ErrorAs(t, err, &rpcError)
	require.Equal(t, -32099, rpcError.Code)
	require.Equal(t, "upstream failed", rpcError.Message)
	require.JSONEq(t, ` + "`" + `{"reason":"overload"}` + "`" + `, string(rpcError.Data))
}

func TestResumeRequestWritesLastEventIDOutsideParams(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		encode  func(func(*http.Request) goahttp.Encoder) func(*http.Request, any) (string, error)
		present bool
	}{
		{name: "required empty", value: &service.RequiredResumePayload{}, encode: EncodeRequiredResumeRequest, present: true},
		{name: "optional absent", value: &service.OptionalResumePayload{}, encode: EncodeOptionalResumeRequest},
		{name: "optional empty", value: &service.OptionalResumePayload{LastEventID: stringPointer("")}, encode: EncodeOptionalResumeRequest, present: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/rpc", nil)
			_, err := test.encode(goahttp.RequestEncoder)(request, test.value)
			require.NoError(t, err)
			values, present := request.Header["Last-Event-Id"]
			require.Equal(t, test.present, present)
			if present {
				require.Equal(t, []string{""}, values)
			}
			body, err := io.ReadAll(request.Body)
			require.NoError(t, err)
			require.NotContains(t, string(body), "last_event_id")
		})
	}
}

func TestResumeRequestRejectsInvalidLastEventID(t *testing.T) {
	for _, value := range []string{"one\x00two", "one\rtwo", "one\ntwo"} {
		request := httptest.NewRequest(http.MethodPost, "/rpc", nil)

		_, err := EncodeOptionalResumeRequest(goahttp.RequestEncoder)(request, &service.OptionalResumePayload{LastEventID: stringPointer(value)})

		require.ErrorContains(t, err, "last_event_id")
		require.Empty(t, request.Header.Values("Last-Event-ID"))
	}
}

func TestPrimitiveStreamRejectsWrongPositionalCount(t *testing.T) {
	for _, params := range []string{"[]", "[\"one\",\"two\"]"} {
		event := "data: {\"jsonrpc\":\"2.0\",\"method\":\"stream_text\",\"params\":" + params + "}\n\n"
		response := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(event))}
		_, err := NewStreamTextStream(response, goahttp.ResponseDecoder, "request-1").Recv()
		require.EqualError(t, err, "params must be an array with exactly one value")
	}
}

func paramsOnly(body []byte) string {
	text := string(body)
	start := strings.Index(text, ` + "`" + `"params"` + "`" + `)
	end := strings.LastIndex(text, ` + "`" + `,"id"` + "`" + `)
	return "{" + text[start:end] + "}"
}
`
