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
		ID      any            `json:"id,omitempty"`
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
		ID      any             `json:"id,omitempty"`
	}

	// RawResponse represents a JSON-RPC response with a marshalled result
	// and error.
	RawResponse struct {
		JSONRPC string            `json:"jsonrpc"`
		Result  json.RawMessage   `json:"result,omitempty"`
		Error   *RawErrorResponse `json:"error,omitempty"`
		ID      string            `json:"id,omitempty"`
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
