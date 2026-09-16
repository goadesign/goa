{{ printf "%s returns the name of the service served." .ServerService | comment }}
func (s *{{ .ServerStructDeclaration.Name }}) {{ .ServerService }}() string { return "{{ .Service.Name }}" }

// reportRejectedNotification reports a notification sent to a request-only
// method without allowing the error handler to write a response.
func (s *{{ .ServerStructDeclaration.Name }}) reportRejectedNotification(ctx context.Context, req *jsonrpc.RawRequest) {
	outputWriter := &{{ .NoOutputWriter.Name }}{header: make(http.Header)}
	s.errhandler(ctx, outputWriter, fmt.Errorf("JSON-RPC notification cannot call request method %q", req.Method))
}
