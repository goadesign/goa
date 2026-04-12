{{ printf "Build%sFunc builds the remote method to invoke for %q service %q endpoint." .Method.VarName .ServiceName .Method.Name | comment }}
func Build{{ .Method.VarName }}Func(grpccli {{ .PkgName }}.{{ .ClientInterface }}, cliopts ...grpc.CallOption) goagrpc.RemoteFunc {
	return func(ctx context.Context, reqpb any, opts ...grpc.CallOption) (any, error) {
		for _, opt := range cliopts {
			opts = append(opts, opt)
		}
		{{- if .Request.StreamEnvelope }}
			stream, err := grpccli.{{ .ClientMethodName }}(ctx, opts...)
			if err != nil {
				return nil, err
			}
			if reqpb != nil {
				if err := stream.Send(reqpb.({{ .Request.Message.Ref }})); err != nil {
					return nil, err
				}
			}
			return stream, nil
		{{- else }}
			if reqpb != nil {
				return grpccli.{{ .ClientMethodName }}(ctx{{ if not .Method.StreamingPayload }}, reqpb.({{ .Request.ClientConvert.TgtRef }}){{ end }}, opts...)
			}
			return grpccli.{{ .ClientMethodName }}(ctx{{ if not .Method.StreamingPayload }}, &{{ .Request.ClientConvert.TgtName }}{}{{ end }}, opts...)
		{{- end }}
	}
}
