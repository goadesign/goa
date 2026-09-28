{{ printf "%s builds the remote method to invoke for %q service %q endpoint." .ClientBuildDeclaration.Name .ServiceName .Method.Name | comment }}
func {{ .ClientBuildDeclaration.Name }}(grpccli {{ .ClientProtobufPkgName }}.{{ .ClientInterface }}, cliopts ...grpc.CallOption) goagrpc.RemoteFunc {
	return func(ctx context.Context, reqpb any, opts ...grpc.CallOption) (any, error) {
		for _, opt := range cliopts {
			opts = append(opts, opt)
		}
		{{- if .Request.StreamEnvelope }}
			stream, err := grpccli.{{ .GRPCMethodName }}(ctx, opts...)
			if err != nil {
				return nil, err
			}
			if reqpb != nil {
				// An initial-send EOF leaves the server status for Recv or
				// CloseAndRecv. Keep the stream so its decoder can read it.
				if err := stream.Send(reqpb.({{ .Request.ClientMessageRef }})); err != nil && !errors.Is(err, io.EOF) {
					return nil, err
				}
			}
			return stream, nil
		{{- else }}
			if reqpb != nil {
				return grpccli.{{ .GRPCMethodName }}(ctx{{ if not .Method.StreamingPayload }}, reqpb.({{ .Request.ClientConvert.TgtRef }}){{ end }}, opts...)
			}
			return grpccli.{{ .GRPCMethodName }}(ctx{{ if not .Method.StreamingPayload }}, &{{ .Request.ClientConvert.TgtName }}{}{{ end }}, opts...)
		{{- end }}
	}
}
