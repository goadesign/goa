package jsonrpc

import "encoding/json"

type (
	// Request represents a JSON-RPC request.
	Request struct {
		JSONRPC string           `json:"jsonrpc"`
		Method  string           `json:"method"`
		Params  json.RawMessage  `json:"params,omitempty"`
		ID      *json.RawMessage `json:"id,omitempty"`
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
	return &Response{
		JSONRPC: "2.0",
		Error:   &ErrorResponse{Code: code, Message: message, Data: data},
		ID:      id,
	}
}
