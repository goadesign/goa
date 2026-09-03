// This file defines the JSON-RPC messages shared by generated clients and
// servers. It preserves request IDs exactly so a server can return the value
// it received and a client can reject a response for another request.
package jsonrpc

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type (
	// Request represents a JSON-RPC request.
	Request struct {
		JSONRPC string `json:"jsonrpc"`
		Method  string `json:"method"`
		Params  any    `json:"params,omitempty"`
		ID      any    `json:"id,omitempty"`
	}

	// Response represents a JSON-RPC response.
	Response struct {
		JSONRPC string         `json:"jsonrpc"`
		Result  any            `json:"result,omitempty"`
		Error   *ErrorResponse `json:"error,omitempty"`
		ID      any            `json:"id"`
	}

	// ErrorResponse represents a JSON-RPC error response.
	ErrorResponse struct {
		Code    Code   `json:"code"`
		Message string `json:"message"`
		Data    any    `json:"data,omitempty"`
	}

	// RawRequest represents a JSON-RPC request with a marshalled params.
	RawRequest struct {
		JSONRPC string          `json:"jsonrpc"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params,omitempty"`
		ID      any             `json:"id"`
		// Invalid is true when the JSON value is not shaped like a JSON-RPC
		// request object.
		Invalid bool `json:"-"`
		// HasID is true when the "id" key is present in the incoming JSON, even
		// when its value is null. Generated servers compare it with the method's
		// declared request or notification contract before calling the service.
		HasID bool `json:"-"`
		// HasMethod is true when the "method" key is present, including when its
		// value is an empty string.
		HasMethod bool `json:"-"`
	}

	// RawResponse represents a JSON-RPC response with a marshalled result
	// and error.
	RawResponse struct {
		JSONRPC string            `json:"jsonrpc"`
		Result  json.RawMessage   `json:"result,omitempty"`
		Error   *RawErrorResponse `json:"error,omitempty"`
		ID      any               `json:"id,omitempty"`
		// Invalid is true when the JSON value is not shaped like a JSON-RPC
		// response object.
		Invalid bool `json:"-"`
		// HasResult is true when the response contains the "result" key,
		// including a null result.
		HasResult bool `json:"-"`
		// HasError is true when the response contains the "error" key.
		HasError bool `json:"-"`
		// HasID is true when the response contains the "id" key, including a
		// null ID.
		HasID bool `json:"-"`
	}

	// RawErrorResponse represents a JSON-RPC error response with marshalled
	// data.
	RawErrorResponse struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data,omitempty"`
	}

	// Code is a JSON-RPC error code, see JSON-RPC 2.0 section 5.1
	Code int
)

const (
	ParseError     Code = -32700
	InvalidRequest Code = -32600
	MethodNotFound Code = -32601
	InvalidParams  Code = -32602
	InternalError  Code = -32603
)

// MakeSuccessResponse creates a success response.
func MakeSuccessResponse(id any, result any) *Response {
	return &Response{
		JSONRPC: "2.0",
		Result:  result,
		ID:      id,
	}
}

// MakeErrorResponse creates an error response.
func MakeErrorResponse(id any, code Code, message string, data any) *Response {
	if message == "" {
		switch code {
		case ParseError:
			message = "Parse error"
		case InvalidRequest:
			message = "Invalid request"
		case MethodNotFound:
			message = "Method not found"
		case InvalidParams:
			message = "Invalid params"
		case InternalError:
			message = "Internal error"
		default:
			message = "Unknown error"
		}
	}
	return &Response{
		JSONRPC: "2.0",
		Error:   &ErrorResponse{Code: code, Message: message, Data: data},
		ID:      id,
	}
}

// DecodeServiceErrorData decodes the data object written for a designed Goa
// error. It returns ok false and empty name and body values when data is not
// that object, so callers can preserve the original JSON-RPC error.
func DecodeServiceErrorData(data json.RawMessage) (string, json.RawMessage, bool) {
	var value struct {
		Name *string         `json:"name"`
		Body json.RawMessage `json:"body"`
	}
	if err := json.Unmarshal(data, &value); err != nil || value.Name == nil || len(value.Body) == 0 {
		return "", nil, false
	}
	return *value.Name, value.Body, true
}

// MakeNotification creates a notification.
func MakeNotification(method string, params any) *Request {
	return &Request{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
}

// MarshalJSON writes the result member for every success response, including
// responses whose result is null, and writes only the error for failures.
func (r *Response) MarshalJSON() ([]byte, error) {
	if r.Error != nil {
		return json.Marshal(struct {
			JSONRPC string         `json:"jsonrpc"`
			Error   *ErrorResponse `json:"error"`
			ID      any            `json:"id"`
		}{
			JSONRPC: r.JSONRPC,
			Error:   r.Error,
			ID:      r.ID,
		})
	}
	return json.Marshal(struct {
		JSONRPC string `json:"jsonrpc"`
		Result  any    `json:"result"`
		ID      any    `json:"id"`
	}{
		JSONRPC: r.JSONRPC,
		Result:  r.Result,
		ID:      r.ID,
	})
}

// Error returns a string representation of the error.
func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("jsonrpc: code %d: %s", e.Code, e.Message)
}

// Error returns a string representation of the error.
func (e *RawErrorResponse) Error() string {
	return fmt.Sprintf("jsonrpc: code %d: %s", e.Code, e.Message)
}

// IDToString converts a decoded JSON-RPC string or number ID to its exact text.
func IDToString(id any) (string, error) {
	switch v := id.(type) {
	case string:
		return v, nil
	case json.Number:
		return v.String(), nil
	default:
		return "", fmt.Errorf("JSON-RPC id has unexpected type %T", id)
	}
}

// SinglePositionalParam returns the only value in a positional params array.
// It rejects named params and arrays that do not contain exactly one value.
func SinglePositionalParam(params json.RawMessage) (json.RawMessage, error) {
	var values []json.RawMessage
	if err := json.Unmarshal(params, &values); err != nil || values == nil || len(values) != 1 {
		return nil, fmt.Errorf("params must be an array with exactly one value")
	}
	return values[0], nil
}

// UnmarshalJSON decodes one request and records invalid input and ID presence.
func (r *RawRequest) UnmarshalJSON(data []byte) error {
	*r = RawRequest{}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		if json.Valid(data) {
			r.Invalid = true
			return nil
		}
		return err
	}
	if raw == nil {
		r.Invalid = true
		return nil
	}
	if v, ok := raw["id"]; ok {
		r.HasID = true
		var valid bool
		r.ID, valid = decodeID(v)
		if !valid {
			r.Invalid = true
		}
	}
	if v, ok := raw["jsonrpc"]; ok {
		if json.Unmarshal(v, &r.JSONRPC) != nil {
			r.Invalid = true
		}
	}
	if v, ok := raw["method"]; ok {
		r.HasMethod = true
		var method *string
		if json.Unmarshal(v, &method) != nil || method == nil {
			r.Invalid = true
		} else {
			r.Method = *method
		}
	}
	if v, ok := raw["params"]; ok {
		r.Params = v
		params := bytes.TrimSpace(v)
		if len(params) == 0 || (params[0] != '{' && params[0] != '[') {
			r.Invalid = true
		}
	}
	return nil
}

// UnmarshalJSON decodes one response and records the fields whose presence is
// required to distinguish a valid response from a zero Go value.
func (r *RawResponse) UnmarshalJSON(data []byte) error {
	*r = RawResponse{}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw == nil {
		r.Invalid = true
		return nil
	}
	if value, ok := raw["jsonrpc"]; ok {
		if err := json.Unmarshal(value, &r.JSONRPC); err != nil {
			r.Invalid = true
		}
	}
	if value, ok := raw["result"]; ok {
		r.HasResult = true
		r.Result = append(r.Result[:0], value...)
	}
	if value, ok := raw["error"]; ok {
		r.HasError = true
		var valid bool
		r.Error, valid = decodeRawError(value)
		if !valid {
			r.Invalid = true
		}
	}
	if value, ok := raw["id"]; ok {
		r.HasID = true
		var valid bool
		r.ID, valid = decodeID(value)
		if !valid {
			r.Invalid = true
		}
	}
	return nil
}

// Validate checks that the response is a complete JSON-RPC 2.0 envelope for
// expectedID. Parse and invalid-request errors may use a null ID because the
// server could not recover the request ID from malformed input.
func (r *RawResponse) Validate(expectedID string) error {
	if r.Invalid {
		return fmt.Errorf("response is not a valid JSON-RPC object")
	}
	if r.JSONRPC != "2.0" {
		return fmt.Errorf("response jsonrpc must be \"2.0\"")
	}
	if r.HasResult == r.HasError {
		return fmt.Errorf("response must contain exactly one of result or error")
	}
	if !r.HasID {
		return fmt.Errorf("response has no id")
	}
	if r.ID == nil {
		if r.Error != nil && (r.Error.Code == int(ParseError) || r.Error.Code == int(InvalidRequest)) {
			return nil
		}
		return fmt.Errorf("response id is null")
	}
	actualID, ok := r.ID.(string)
	if !ok {
		switch r.ID.(type) {
		case json.Number, float64:
			return fmt.Errorf("response id is a number")
		default:
			return fmt.Errorf("response is not a valid JSON-RPC object")
		}
	}
	if actualID != expectedID {
		return fmt.Errorf("response id %q does not match request id %q", actualID, expectedID)
	}
	return nil
}

// decodeRawError decodes the required members of one JSON-RPC error object.
// Empty messages and any integer code remain valid values.
func decodeRawError(raw json.RawMessage) (*RawErrorResponse, bool) {
	var values map[string]json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil || values == nil {
		//nolint:nilerr // A malformed error object is not a valid response.
		return nil, false
	}
	code, hasCode := values["code"]
	message, hasMessage := values["message"]
	if !hasCode || !hasMessage {
		return nil, false
	}
	result := new(RawErrorResponse)
	if json.Unmarshal(code, &result.Code) != nil || json.Unmarshal(message, &result.Message) != nil {
		//nolint:nilerr // A malformed error member is not a valid response.
		return nil, false
	}
	if data, ok := values["data"]; ok {
		result.Data = append(result.Data[:0], data...)
	}
	return result, true
}

// decodeID preserves a JSON-RPC string, number, or null ID without converting
// numbers through float64. The boolean is false for every other JSON value.
func decodeID(raw json.RawMessage) (any, bool) {
	if string(raw) == "null" {
		return nil, true
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var id any
	if err := decoder.Decode(&id); err != nil {
		return nil, false
	}
	switch id.(type) {
	case string, json.Number:
		return id, true
	default:
		return nil, false
	}
}
