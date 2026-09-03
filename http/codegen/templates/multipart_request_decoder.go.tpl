{{ printf "%s returns a decoder to decode the multipart request for the %q service %q endpoint." .InitDeclaration.Name .ServiceName .MethodName | comment }}
func {{ .InitDeclaration.Name }}(_ goahttp.Muxer, {{ .VarName }} {{ .FuncDeclaration.Name }}) func(r *http.Request) goahttp.Decoder {
	return func(r *http.Request) goahttp.Decoder {
		return goahttp.EncodingFunc(func(v any) error {
			mr, merr := r.MultipartReader()
			if merr != nil {
				return merr
			}
			body := v.(*{{ if .Payload.Request.ServerBody.Declaration }}{{ .Payload.Request.ServerBody.Declaration.Name }}{{ else }}{{ .Payload.Request.ServerBody.VarName }}{{ end }})
			if err := {{ .VarName }}(mr, body); err != nil {
				return err
			}
			return nil
		})
	}
}
