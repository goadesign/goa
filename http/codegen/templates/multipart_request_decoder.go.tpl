{{ printf "%s returns a decoder to decode the multipart request for the %q service %q endpoint." .InitName .ServiceName .MethodName | comment }}
func {{ .InitName }}(mux goahttp.Muxer, {{ .VarName }} {{ .FuncName }}) func(r *http.Request) goahttp.Decoder {
	return func(r *http.Request) goahttp.Decoder {
		return goahttp.EncodingFunc(func(v any) error {
			mr, merr := r.MultipartReader()
			if merr != nil {
				return merr
			}
			p := v.(*{{ .Payload.Ref }})
			if err := {{ .VarName }}(mr, p); err != nil {
				return err
			}
			{{- template "partial_request_elements" .Payload.Request }}
			{{- if .Payload.Request.MustValidate }}
			if err != nil {
				return err
			}
			{{- end }}
			{{- if .Payload.Request.PayloadInit }}
				{{- range .Payload.Request.PayloadInit.ServerArgs }}
					{{- if .FieldName }}
			(*p).{{ .FieldName }} = {{ if and (not .Pointer) .FieldPointer }}&{{ end }}{{ .VarName }}
					{{- end }}
				{{- end }}
			{{- end }}
			return nil
		})
	}
}
