{{ printf "%s decodes the JSON body selected by the result view for the %s service %s method." .Decode.Name .ServiceName .MethodName | comment }}
func {{ .Decode.Name }}(decoder func(*http.Response) goahttp.Decoder, resp *http.Response, data json.RawMessage) ({{ .ResultRef }}, error) {
	{{- if .Variable }}
	var representation struct {
		View *string          `json:"view"`
		Body *json.RawMessage `json:"body"`
	}
	if err := {{ .BodyDecoder.Name }}(decoder, data, &representation); err != nil {
		return nil, err
	}
	if representation.View == nil {
		return nil, goa.MissingFieldError("view", "result")
	}
	view := *representation.View
	switch view {
	{{- range .Branches }}
	case {{ printf "%q" .View }}:
		{{- if .ClientBody }}
		if representation.Body == nil {
			return nil, goa.MissingFieldError("body", "result")
		}
		resp.Body = io.NopCloser(bytes.NewBuffer(*representation.Body))
		{{- end }}
		{{- template "partial_single_response" (viewedResponseData . $.ServiceName $.MethodName) }}
		projected := {{ .ResultInit.Name }}({{ range .ResultInit.ClientArgs }}{{ .Ref }},{{ end }})
		viewed := {{ if not $.IsCollection }}&{{ end }}{{ $.ViewedPkg }}.{{ $.ViewedVarName }}{
			Projected: projected,
			View:      view,
		}
		if err := {{ $.ViewedPkg }}.{{ $.ViewedValidator }}(viewed); err != nil {
			return nil, err
		}
		return {{ $.ServicePkg }}.{{ $.ServiceResultConstructor }}(viewed), nil
	{{- end }}
	default:
		return nil, goa.InvalidEnumValueError("view", view, []any{
			{{- range .Branches }}{{ printf "%q" .View }},{{ end }}
		})
	}
	{{- else }}
	{{- with index .Branches 0 }}
	{{- if .ClientBody }}
	resp.Body = io.NopCloser(bytes.NewBuffer(data))
	{{- end }}
	{{- template "partial_single_response" (viewedResponseData . $.ServiceName $.MethodName) }}
	projected := {{ .ResultInit.Name }}({{ range .ResultInit.ClientArgs }}{{ .Ref }},{{ end }})
	viewed := {{ if not $.IsCollection }}&{{ end }}{{ $.ViewedPkg }}.{{ $.ViewedVarName }}{
		Projected: projected,
		View:      {{ printf "%q" $.FixedView }},
	}
	if err := {{ $.ViewedPkg }}.{{ $.ViewedValidator }}(viewed); err != nil {
		return nil, err
	}
	return {{ $.ServicePkg }}.{{ $.ServiceResultConstructor }}(viewed), nil
	{{- end }}
	{{- end }}
}
