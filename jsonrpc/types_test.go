// This file checks the JSON-RPC envelope values shared by generated clients
// and servers.
package jsonrpc_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/jsonrpc"
)

// TestRequestKeepsEmptyStringID checks that an explicitly selected empty
// string remains an ID instead of turning the request into a notification.
func TestRequestKeepsEmptyStringID(t *testing.T) {
	encoded, err := json.Marshal(&jsonrpc.Request{
		JSONRPC: "2.0",
		Method:  "lookup",
		ID:      "",
	})
	require.NoError(t, err)
	require.JSONEq(t, `{"jsonrpc":"2.0","id":"","method":"lookup"}`, string(encoded))
}

// TestRawRequestPreservesNumericID checks that a server can repeat a numeric
// request ID without converting it through floating-point arithmetic.
func TestRawRequestPreservesNumericID(t *testing.T) {
	const request = `{"jsonrpc":"2.0","id":9007199254740993123456789,"method":"lookup"}`
	var decoded jsonrpc.RawRequest
	require.NoError(t, json.Unmarshal([]byte(request), &decoded))
	require.True(t, decoded.HasID)
	require.Equal(t, json.Number("9007199254740993123456789"), decoded.ID)

	encoded, err := json.Marshal(jsonrpc.MakeSuccessResponse(decoded.ID, nil))
	require.NoError(t, err)
	require.JSONEq(t, `{"jsonrpc":"2.0","id":9007199254740993123456789,"result":null}`, string(encoded))
}

// TestIDToString checks the exact conversion used when a JSON-RPC ID is
// assigned to a string field in a service payload.
func TestIDToString(t *testing.T) {
	tests := []struct {
		name    string
		id      any
		want    string
		wantErr string
	}{
		{name: "string", id: "call-1", want: "call-1"},
		{name: "empty string", id: ""},
		{name: "number", id: json.Number("9007199254740993123456789"), want: "9007199254740993123456789"},
		{name: "floating point number", id: float64(1), wantErr: "JSON-RPC id has unexpected type float64"},
		{name: "nil", wantErr: "JSON-RPC id has unexpected type <nil>"},
		{name: "object", id: map[string]any{}, wantErr: "JSON-RPC id has unexpected type map[string]interface {}"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, err := jsonrpc.IDToString(test.id)
			if test.wantErr != "" {
				require.EqualError(t, err, test.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.want, value)
		})
	}
}

// TestRawRequestRequiresStructuredParams checks the JSON-RPC rule that params,
// when present, must be an object or array.
func TestRawRequestRequiresStructuredParams(t *testing.T) {
	tests := []struct {
		name    string
		params  string
		invalid bool
	}{
		{name: "object", params: `{}`},
		{name: "array", params: `[]`},
		{name: "string", params: `"value"`, invalid: true},
		{name: "number", params: `1`, invalid: true},
		{name: "boolean", params: `true`, invalid: true},
		{name: "null", params: `null`, invalid: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := `{"jsonrpc":"2.0","method":"lookup","params":` + test.params + `}`
			var decoded jsonrpc.RawRequest
			require.NoError(t, json.Unmarshal([]byte(request), &decoded))
			require.Equal(t, test.invalid, decoded.Invalid)
		})
	}
}

// TestRawRequestRecordsMethodPresence checks that a missing method can be
// distinguished from an explicitly empty method name.
func TestRawRequestRecordsMethodPresence(t *testing.T) {
	tests := []struct {
		name      string
		request   string
		hasMethod bool
		method    string
		invalid   bool
	}{
		{name: "missing", request: `{"jsonrpc":"2.0"}`},
		{name: "empty", request: `{"jsonrpc":"2.0","method":""}`, hasMethod: true},
		{name: "named", request: `{"jsonrpc":"2.0","method":"lookup"}`, hasMethod: true, method: "lookup"},
		{name: "null", request: `{"jsonrpc":"2.0","method":null}`, hasMethod: true, invalid: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var request jsonrpc.RawRequest
			require.NoError(t, json.Unmarshal([]byte(test.request), &request))
			require.Equal(t, test.hasMethod, request.HasMethod)
			require.Equal(t, test.method, request.Method)
			require.Equal(t, test.invalid, request.Invalid)
		})
	}
}

// TestSinglePositionalParam checks the JSON-RPC array used to carry one Goa
// value without changing the value's service type.
func TestSinglePositionalParam(t *testing.T) {
	tests := []struct {
		name    string
		params  string
		want    string
		wantErr string
	}{
		{name: "string", params: `["value"]`, want: `"value"`},
		{name: "object value", params: `[{"name":"value"}]`, want: `{"name":"value"}`},
		{name: "null value", params: `[null]`, want: `null`},
		{name: "object container", params: `{"name":"value"}`, wantErr: "params must be an array with exactly one value"},
		{name: "empty array", params: `[]`, wantErr: "params must be an array with exactly one value"},
		{name: "two values", params: `["one","two"]`, wantErr: "params must be an array with exactly one value"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, err := jsonrpc.SinglePositionalParam(json.RawMessage(test.params))
			if test.wantErr != "" {
				require.EqualError(t, err, test.wantErr)
				return
			}
			require.NoError(t, err)
			require.JSONEq(t, test.want, string(value))
		})
	}
}

// TestDecodeServiceErrorData checks the typed data object that distinguishes a
// designed Goa error from a JSON-RPC protocol failure using the same code.
func TestDecodeServiceErrorData(t *testing.T) {
	for _, test := range []struct {
		name     string
		data     string
		wantName string
		wantBody string
		wantOK   bool
	}{
		{name: "structured body", data: `{"name":"busy","body":{"retryAfter":3}}`, wantName: "busy", wantBody: `{"retryAfter":3}`, wantOK: true},
		{name: "empty body", data: `{"name":"quiet","body":null}`, wantName: "quiet", wantBody: `null`, wantOK: true},
		{name: "unknown field", data: `{"name":"busy","body":{},"future":true}`, wantName: "busy", wantBody: `{}`, wantOK: true},
		{name: "missing name", data: `{"body":{}}`},
		{name: "null name", data: `{"name":null,"body":{}}`},
		{name: "wrong name type", data: `{"name":1,"body":{}}`},
		{name: "missing body", data: `{"name":"busy"}`},
		{name: "array", data: `[]`},
		{name: "invalid JSON", data: `{`},
	} {
		t.Run(test.name, func(t *testing.T) {
			name, body, ok := jsonrpc.DecodeServiceErrorData(json.RawMessage(test.data))

			require.Equal(t, test.wantOK, ok)
			if !test.wantOK {
				require.Empty(t, name)
				require.Nil(t, body)
				return
			}
			require.Equal(t, test.wantName, name)
			require.JSONEq(t, test.wantBody, string(body))
		})
	}
}

// TestRawResponseValidation checks the full response envelope before generated
// clients decode its result or service error.
func TestRawResponseValidation(t *testing.T) {
	tests := []struct {
		name       string
		response   string
		expectedID string
		wantErr    string
	}{
		{name: "matching result", response: `{"jsonrpc":"2.0","id":"call-1","result":null}`, expectedID: "call-1"},
		{name: "matching empty ID", response: `{"jsonrpc":"2.0","id":"","result":null}`},
		{name: "matching error", response: `{"jsonrpc":"2.0","id":"call-1","error":{"code":-32602,"message":"bad params"}}`, expectedID: "call-1"},
		{name: "parse error with null ID", response: `{"jsonrpc":"2.0","id":null,"error":{"code":-32700,"message":"parse error"}}`, expectedID: "call-1"},
		{name: "invalid request with null ID", response: `{"jsonrpc":"2.0","id":null,"error":{"code":-32600,"message":"invalid request"}}`, expectedID: "call-1"},
		{name: "method not found with null ID", response: `{"jsonrpc":"2.0","id":null,"error":{"code":-32601,"message":"method not found"}}`, expectedID: "call-1", wantErr: "response id is null"},
		{name: "invalid params with null ID", response: `{"jsonrpc":"2.0","id":null,"error":{"code":-32602,"message":"invalid params"}}`, expectedID: "call-1", wantErr: "response id is null"},
		{name: "internal error with null ID", response: `{"jsonrpc":"2.0","id":null,"error":{"code":-32603,"message":"internal error"}}`, expectedID: "call-1", wantErr: "response id is null"},
		{name: "server error with null ID", response: `{"jsonrpc":"2.0","id":null,"error":{"code":-32000,"message":"server error"}}`, expectedID: "call-1", wantErr: "response id is null"},
		{name: "application error with null ID", response: `{"jsonrpc":"2.0","id":null,"error":{"code":7001,"message":"application error"}}`, expectedID: "call-1", wantErr: "response id is null"},
		{name: "missing ID", response: `{"jsonrpc":"2.0","result":null}`, expectedID: "call-1", wantErr: "response has no id"},
		{name: "null result ID", response: `{"jsonrpc":"2.0","id":null,"result":null}`, expectedID: "call-1", wantErr: "response id is null"},
		{name: "numeric ID", response: `{"jsonrpc":"2.0","id":1,"result":null}`, expectedID: "1", wantErr: "response id is a number"},
		{name: "different ID", response: `{"jsonrpc":"2.0","id":"call-2","result":null}`, expectedID: "call-1", wantErr: `response id "call-2" does not match request id "call-1"`},
		{name: "missing version", response: `{"id":"call-1","result":null}`, expectedID: "call-1", wantErr: "response jsonrpc must be \"2.0\""},
		{name: "wrong version", response: `{"jsonrpc":"1.0","id":"call-1","result":null}`, expectedID: "call-1", wantErr: "response jsonrpc must be \"2.0\""},
		{name: "result and error", response: `{"jsonrpc":"2.0","id":"call-1","result":null,"error":{"code":-32603,"message":"failed"}}`, expectedID: "call-1", wantErr: "response must contain exactly one of result or error"},
		{name: "no result or error", response: `{"jsonrpc":"2.0","id":"call-1"}`, expectedID: "call-1", wantErr: "response must contain exactly one of result or error"},
		{name: "error missing code", response: `{"jsonrpc":"2.0","id":"call-1","error":{"message":"failed"}}`, expectedID: "call-1", wantErr: "response is not a valid JSON-RPC object"},
		{name: "error code is not an integer", response: `{"jsonrpc":"2.0","id":"call-1","error":{"code":"bad","message":"failed"}}`, expectedID: "call-1", wantErr: "response is not a valid JSON-RPC object"},
		{name: "error missing message", response: `{"jsonrpc":"2.0","id":"call-1","error":{"code":7}}`, expectedID: "call-1", wantErr: "response is not a valid JSON-RPC object"},
		{name: "error message is not a string", response: `{"jsonrpc":"2.0","id":"call-1","error":{"code":7,"message":false}}`, expectedID: "call-1", wantErr: "response is not a valid JSON-RPC object"},
		{name: "invalid ID type", response: `{"jsonrpc":"2.0","id":{},"result":null}`, expectedID: "call-1", wantErr: "response is not a valid JSON-RPC object"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var response jsonrpc.RawResponse
			require.NoError(t, json.Unmarshal([]byte(test.response), &response))
			err := response.Validate(test.expectedID)
			if test.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.EqualError(t, err, test.wantErr)
		})
	}
}
