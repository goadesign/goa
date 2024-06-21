{{ printf "%s calls the %q function in %s.%s interface." .Method.VarName .Method.VarName .PkgName .ClientInterface | comment }}
func (c *{{ .ClientStruct }}) {{ .Method.VarName }}() goa.Endpoint {
	return func(ctx context.Context, v any) (any, error) {
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
				return nil, goa.Fault(err.Error())
			}
		{{- else }}
			return nil, goa.Fault(err.Error())
		{{- end }}
		}
		return res, nil
	}
}