{{ printf "encodeJSONRPCError creates and sends a JSON-RPC error response (handles nil ID gracefully)" | comment }}
func (s *Server) encodeJSONRPCError(ctx context.Context, w http.ResponseWriter, req *jsonrpc.RawRequest, code jsonrpc.Code, message string, data any) {
	encodeJSONRPCError(ctx, w, req, code, message, data, s.encoder, s.errhandler)
}

{{ printf "encodeJSONRPCError creates and sends a JSON-RPC error response (handles nil ID gracefully)" | comment }}
func encodeJSONRPCError(
	ctx context.Context,
	w http.ResponseWriter,
	req *jsonrpc.RawRequest,
	code jsonrpc.Code,
	message string,
	data any,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
) {
	if req.ID != nil {
		response := jsonrpc.MakeErrorResponse(req.ID, code, "", message)
		if data != nil {
			response.Error.Data = data
		}
		if err := encoder(ctx, w).Encode(response); err != nil {
			errhandler(ctx, w, fmt.Errorf("failed to encode JSON-RPC response: %w", err))
		}
	}
}
