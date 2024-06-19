{{ printf "Build%sFunc builds the remote method to invoke for %q service %q endpoint." .Method.VarName .ServiceName .Method.Name | comment }}
func Build{{ .Method.VarName }}Func(grpccli {{ .PkgName }}.{{ .ClientInterface }}, cliopts ...grpc.CallOption) goagrpc.RemoteFunc {
	return func(ctx context.Context, reqpb any, opts ...grpc.CallOption) (any, error) {
		for _, opt := range cliopts {
			opts = append(opts, opt)
		}
		if reqpb != nil {
			return grpccli.{{ .ClientMethodName }}(ctx{{ if not .Method.StreamingPayload }}, reqpb.({{ .Request.ClientConvert.TgtRef }}){{ end }}, opts...)
		}
		return grpccli.{{ .ClientMethodName }}(ctx{{ if not .Method.StreamingPayload }}, &{{ .Request.ClientConvert.TgtName }}{}{{ end }}, opts...)
	}
}