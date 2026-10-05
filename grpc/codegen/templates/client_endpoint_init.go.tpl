{{- $retry := and .Method.Idempotent (eq .Method.StreamKind 1) }}
{{ printf "%s calls the %q function in %s.%s interface." .Method.VarName .Method.VarName .ClientProtobufPkgName .ClientInterface | comment }}
func (c *{{ .ClientStructDeclaration.Name }}) {{ .Method.VarName }}() goa.Endpoint {
	return func(ctx context.Context, v any) (any, error) {
		remote := {{ .ClientBuildDeclaration.Name }}(c.grpccli, c.opts...)
		// Convert errors from the RPC call here so local encoding and decoding
		// errors keep their original types and validation details.
		inv := goagrpc.NewInvoker(
			func(ctx context.Context, request any, {{ if .ClientStream }}_{{ else }}opts{{ end }} ...grpc.CallOption) (any, error) {
				{{- if $retry }}
				// The request is already encoded. Retry this RPC call, then
				// let the invoker decode its successful response once.
				rpc := func(ctx context.Context, request any) (any, error) {
				{{- end }}
				{{- if .ClientStream }}
				// Opening a stream does not wait for completion, so omit
				// the invoker's unary header/trailer capture options. The
				// remote builder still applies the client's own options.
				res, err := remote(ctx, request)
				{{- else }}
				res, err := remote(ctx, request, opts...)
				{{- end }}
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
						return nil, goagrpc.NewServiceErrorWithCause(err, message)
					}
				{{- else }}
					// Decode service fields before considering a native context stop.
					if message, ok := goagrpc.DecodeError(err).(*goapb.ErrorResponse); ok {
						return nil, goagrpc.NewServiceErrorWithCause(err, message)
					}
				{{- end }}
					if ctxErr := goagrpc.ContextError(ctx, err); ctxErr != nil {
						return nil, ctxErr
					}
					// Inspect one cause chain so an independent failure cannot select
					// a child's cancellation code for the complete returned error.
					singleCause := true
					for cause := err; cause != nil && singleCause; {
						if joined, ok := cause.(interface{ Unwrap() []error }); ok {
							cause = nil
							for _, child := range joined.Unwrap() {
								if child != nil {
									if cause != nil {
										singleCause = false
										break
									}
									cause = child
								}
							}
						} else {
							cause = errors.Unwrap(cause)
						}
					}
					if singleCause {
						// A remote stop keeps its status while the caller is still active.
						switch status.Code(err) {
						case codes.Canceled, codes.DeadlineExceeded:
							return nil, err
						}
					}
					{{- if $retry }}
					return nil, goagrpc.NewTransportError(err)
					{{- else }}
					return nil, goa.Fault("%s", err.Error())
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
