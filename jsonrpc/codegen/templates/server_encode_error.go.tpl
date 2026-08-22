{{ printf "encodeJSONRPCError writes one JSON-RPC error response and preserves a missing request ID." | comment }}
func (s *{{ .ServerStructDeclaration.Name }}) encodeJSONRPCError(ctx context.Context, w http.ResponseWriter, req *jsonrpc.RawRequest, code jsonrpc.Code, message string, data any) {
	{{ .EncodeError.Name }}(ctx, w, req, code, message, data, s.encoder, s.errhandler)
}

{{ printf "%s writes one JSON-RPC error response and preserves a missing request ID." .EncodeError.Name | comment }}
func {{ .EncodeError.Name }}(
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
		response := jsonrpc.MakeErrorResponse(req.ID, code, message, data)
		if err := encoder(ctx, w).Encode(response); err != nil {
			errhandler(ctx, w, fmt.Errorf("failed to encode JSON-RPC response: %w", err))
		}
	}
}
