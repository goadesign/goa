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
		// HasID is true when the "id" key is present in the incoming JSON (even if null).
		// It is consumed by generated templates (WebSocket/SSE/HTTP) to decide whether
		// to send a response for this request. Do not remove even if unused by this package.
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

// UnmarshalJSON decodes RawRequest and records whether the id field was present.
func (r *RawRequest) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if v, ok := raw["jsonrpc"]; ok {
		if err := json.Unmarshal(v, &r.JSONRPC); err != nil {
			return err
		}
	}
	if v, ok := raw["method"]; ok {
		if err := json.Unmarshal(v, &r.Method); err != nil {
			return err
		}
	}
	if v, ok := raw["params"]; ok {
		r.Params = v
	}
	if v, ok := raw["id"]; ok {
		r.HasID = true
		// Preserve null vs non-null values
		if string(v) == "null" {
			r.ID = nil
		} else {
			var id any
			if err := json.Unmarshal(v, &id); err != nil {
				return err
			}
			r.ID = id
		}
	} else {
		r.HasID = false
		r.ID = nil
	}
	return nil
}
