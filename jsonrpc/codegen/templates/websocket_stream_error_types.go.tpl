// Stream error types for comprehensive error reporting
type StreamErrorType int

const (
	StreamErrorConnection StreamErrorType = iota // WebSocket connection errors
	StreamErrorProtocol                          // Invalid JSON-RPC protocol
	StreamErrorParsing                           // Failed to parse/decode response
	StreamErrorOrphaned                          // Response with no matching request
	StreamErrorTimeout                           // Request timeout
)

// StreamErrorHandler allows users to handle stream errors
type StreamErrorHandler func(ctx context.Context, errorType StreamErrorType, err error, response *jsonrpc.RawResponse)
