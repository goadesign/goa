package jsonrpc

import (
	"encoding/json"
	"fmt"
	"strconv"
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
		// when its value is null. Generated servers use it to decide whether to
		// send a response for this request.
		HasID bool `json:"-"`
	}

	// RawResponse represents a JSON-RPC response with a marshalled result
	// and error.
	RawResponse struct {
		JSONRPC string            `json:"jsonrpc"`
		Result  json.RawMessage   `json:"result,omitempty"`
		Error   *RawErrorResponse `json:"error,omitempty"`
		ID      any               `json:"id,omitempty"`
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

// IDToString converts a JSON-RPC ID to a string.
// JSON unmarshaling produces string or float64 for numeric values.
func IDToString(id any) string {
	switch v := id.(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return ""
	}
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
		if string(v) != "null" {
			if err := json.Unmarshal(v, &r.ID); err != nil {
				r.Invalid = true
				return nil
			}
			switch r.ID.(type) {
			case string, float64:
			default:
				r.ID = nil
				r.Invalid = true
			}
		}
	}
	if v, ok := raw["jsonrpc"]; ok {
		if err := json.Unmarshal(v, &r.JSONRPC); err != nil {
			r.Invalid = true
			return nil
		}
	}
	if v, ok := raw["method"]; ok {
		if err := json.Unmarshal(v, &r.Method); err != nil {
			r.Invalid = true
			return nil
		}
	}
	if v, ok := raw["params"]; ok {
		r.Params = v
	}
	return nil
}
