{{- $retry := and .Method.Idempotent (eq .Method.StreamKind 1) }}
{{ printf "%s calls the %q function in %s.%s interface." .Method.VarName .Method.VarName .PkgName .ClientInterface | comment }}
func (c *{{ .ClientStruct }}) {{ .Method.VarName }}() goa.Endpoint {
	{{- if $retry }}
	endpoint := func(ctx context.Context, v any) (any, error) {
	{{- else }}
	return func(ctx context.Context, v any) (any, error) {
	{{- end }}
		inv := goagrpc.NewInvoker(
			Build{{ .Method.VarName }}Func(c.grpccli, c.opts...),
			{{ if .PayloadRef }}Encode{{ .Method.VarName }}Request{{ else }}nil{{ end }},
			{{ if or .ResultRef .ClientStream }}Decode{{ .Method.VarName }}Response{{ else }}nil{{ end }})
		res, err := inv.Invoke(ctx, v)
		if err != nil {
		{{- if .Errors }}
			resp := goagrpc.DecodeError(err)
			switch message := resp.(type) {
			{{- range .Errors }}
				{{- if .Response.ClientConvert }}
					case {{ .Response.ClientConvert.SrcRef }}:
						{{- if .Response.ClientConvert.Validation }}
							if err := {{ .Response.ClientConvert.Validation.Name }}(message); err != nil {
								return nil, err
							}
						{{- end }}
						return nil, {{ .Response.ClientConvert.Init.Name }}({{ range .Response.ClientConvert.Init.Args }}{{ .Name }}, {{ end }})
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
	}
	{{- if $retry }}
	return goa.RetryEndpoint(endpoint{{ range .Method.Errors }}{{ if .Temporary }}, {{ printf "%q" .ErrName }}{{ end }}{{ end }})
	{{- end }}
}
