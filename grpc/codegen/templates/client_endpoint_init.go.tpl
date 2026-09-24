{{- $retry := and .Method.Idempotent (eq .Method.StreamKind 1) }}
{{ printf "%s calls the %q function in %s.%s interface." .Method.VarName .Method.VarName .ClientProtobufPkgName .ClientInterface | comment }}
func (c *{{ .ClientStructDeclaration.Name }}) {{ .Method.VarName }}() goa.Endpoint {
	return func(ctx context.Context, v any) (any, error) {
		remote := {{ .ClientBuildDeclaration.Name }}(c.grpccli, c.opts...)
		// Convert errors from the RPC call here so local encoding and decoding
		// errors keep their original types and validation details.
		inv := goagrpc.NewInvoker(
			func(ctx context.Context, request any, opts ...grpc.CallOption) (any, error) {
				{{- if $retry }}
				// The request is already encoded. Retry this RPC call, then
				// let the invoker decode its successful response once.
				rpc := func(ctx context.Context, request any) (any, error) {
				{{- end }}
				res, err := remote(ctx, request, opts...)
				if err != nil {
				{{- if .Errors }}
					resp := goagrpc.DecodeError(err)
					switch message := resp.(type) {
					{{- range .Errors }}
						{{- if .Response.ClientConvert }}
							case {{ .Response.ClientConvert.SrcRef }}:
								{{- if .Response.ClientConvert.Validation }}
									if err := {{ .Response.ClientConvert.Validation.Declaration.Name }}(message); err != nil {
										return nil, err
									}
								{{- end }}
								return nil, {{ .Response.ClientConvert.Init.Declaration.Name }}({{ range .Response.ClientConvert.Init.Args }}{{ .Name }}, {{ end }})
						{{- end }}
					{{- end }}
					case *goapb.ErrorResponse:
						return nil, goagrpc.NewServiceError(message)
					default:
						if ctxErr := goagrpc.ContextError(ctx, err); ctxErr != nil {
							return nil, ctxErr
						}
						{{- if $retry }}
						return nil, goagrpc.NewTransportError(err)
						{{- else }}
						return nil, goa.Fault("%s", err.Error())
						{{- end }}
					}
				{{- else }}
					{{- if $retry }}
					// Decode a Goa error detail before returning a matching context error or preserving the transport error.
					{{- else }}
					// Decode a Goa error detail before returning a matching context error or falling back to Fault.
					{{- end }}
					resp := goagrpc.DecodeError(err)
					if eresp, ok := resp.(*goapb.ErrorResponse); ok {
						return nil, goagrpc.NewServiceError(eresp)
					}
					if ctxErr := goagrpc.ContextError(ctx, err); ctxErr != nil {
						return nil, ctxErr
					}
					{{- if $retry }}
					return nil, goagrpc.NewTransportError(err)
					{{- else }}
					return nil, goa.Fault("%s", err.Error())
					{{- end }}
				{{- end }}
				}
				return res, nil
				{{- if $retry }}
				}
				return goa.RetryEndpoint(rpc{{ range .Method.Errors }}{{ if .Temporary }}, {{ printf "%q" .ErrName }}{{ end }}{{ end }})(ctx, request)
				{{- end }}
			},
			{{ if .PayloadRef }}{{ .ClientEncodeDeclaration.Name }}{{ else }}nil{{ end }},
			{{ if or .ResultRef .ClientStream }}{{ .ClientDecodeDeclaration.Name }}{{ else }}nil{{ end }})
		return inv.Invoke(ctx, v)
	}
}
