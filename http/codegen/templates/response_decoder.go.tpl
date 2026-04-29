{{ printf "%s returns a decoder for responses returned by the %s %s endpoint. restoreBody controls whether the response body should be restored after having been read." .ResponseDecoder .ServiceName .Method.Name | comment }}
{{- if .Errors }}
{{ printf "%s may return the following errors:" .ResponseDecoder | comment }}
	{{- range $gerr := .Errors }}
	{{- range $errors := .Errors }}
//	- {{ printf "%q" .Name }} (type {{ .Ref }}): {{ .Response.StatusCode }}{{ if .Response.Description }}, {{ .Response.Description }}{{ end }}
	{{- end }}
	{{- end }}
//	- error: internal error
{{- end }}
func {{ .ResponseDecoder }}(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (any, error) {
	return func(resp *http.Response) (any, error) {
		if restoreBody {
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			resp.Body = io.NopCloser(bytes.NewBuffer(b))
			defer func() {
				resp.Body = io.NopCloser(bytes.NewBuffer(b))
			}()
		}
		{{- if not .Method.SkipResponseBodyEncodeDecode }} else {
			defer resp.Body.Close()
		}
		{{- end }}
		switch resp.StatusCode {
	{{- range .Result.Responses }}
		case {{ .StatusCode }}:
			{{- template "partial_single_response" (buildResponseData . $.ServiceName $.Method) }}
		{{- if .ResultInit }}
			{{- if .ViewedResult }}
			p := {{ .ResultInit.Name }}({{ range .ResultInit.ClientArgs }}{{ .Ref }},{{ end }})
				{{- range .Cookies }}
					{{- if .MaxAgeFrom }}{{ if .MaxAgeFrom.FieldPointer }}
				{{ .MaxAgeFrom.VarName }}Tmp := {{ .MaxAgeFrom.VarName }}
				p.{{ .MaxAgeFrom.FieldName }} = &{{ .MaxAgeFrom.VarName }}Tmp
					{{- else }}
				p.{{ .MaxAgeFrom.FieldName }} = {{ .MaxAgeFrom.VarName }}
					{{- end }}{{- end }}
					{{- if .DomainFrom }}{{ if .DomainFrom.FieldPointer }}
				{{ .DomainFrom.VarName }}Tmp := {{ .DomainFrom.VarName }}
				p.{{ .DomainFrom.FieldName }} = &{{ .DomainFrom.VarName }}Tmp
					{{- else }}
				p.{{ .DomainFrom.FieldName }} = {{ .DomainFrom.VarName }}
					{{- end }}{{- end }}
					{{- if .PathFrom }}{{ if .PathFrom.FieldPointer }}
				{{ .PathFrom.VarName }}Tmp := {{ .PathFrom.VarName }}
				p.{{ .PathFrom.FieldName }} = &{{ .PathFrom.VarName }}Tmp
					{{- else }}
				p.{{ .PathFrom.FieldName }} = {{ .PathFrom.VarName }}
					{{- end }}{{- end }}
					{{- if .SecureFrom }}{{ if .SecureFrom.FieldPointer }}
				{{ .SecureFrom.VarName }}Tmp := {{ .SecureFrom.VarName }}
				p.{{ .SecureFrom.FieldName }} = &{{ .SecureFrom.VarName }}Tmp
					{{- else }}
				p.{{ .SecureFrom.FieldName }} = {{ .SecureFrom.VarName }}
					{{- end }}{{- end }}
					{{- if .HTTPOnlyFrom }}{{ if .HTTPOnlyFrom.FieldPointer }}
				{{ .HTTPOnlyFrom.VarName }}Tmp := {{ .HTTPOnlyFrom.VarName }}
				p.{{ .HTTPOnlyFrom.FieldName }} = &{{ .HTTPOnlyFrom.VarName }}Tmp
					{{- else }}
				p.{{ .HTTPOnlyFrom.FieldName }} = {{ .HTTPOnlyFrom.VarName }}
					{{- end }}{{- end }}
					{{- if .SameSiteFrom }}{{ if .SameSiteFrom.FieldPointer }}
				{{ .SameSiteFrom.VarName }}Tmp := {{ .SameSiteFrom.VarName }}
				p.{{ .SameSiteFrom.FieldName }} = &{{ .SameSiteFrom.VarName }}Tmp
					{{- else }}
				p.{{ .SameSiteFrom.FieldName }} = {{ .SameSiteFrom.VarName }}
					{{- end }}{{- end }}
				{{- end }}
				{{- if .TagName }}
				tmp := {{ printf "%q" .TagValue }}
				p.{{ .TagName }} = &tmp
				{{- end }}
				{{- if $.Method.ViewedResult.ViewName }}
			view := {{ printf "%q" $.Method.ViewedResult.ViewName }}
				{{- else }}
			view := resp.Header.Get("goa-view")
				{{- end }}
			vres := {{ if not $.Method.ViewedResult.IsCollection }}&{{ end }}{{ $.Method.ViewedResult.ViewsPkg}}.{{ $.Method.ViewedResult.VarName }}{Projected: p, View: view}
				{{- if .ClientBody }}
				if err = {{ $.Method.ViewedResult.ViewsPkg}}.Validate{{ $.Method.Result }}(vres); err != nil {
					return nil, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", err)
				}
				{{- end }}
			res := {{ $.ServicePkgName }}.{{ $.Method.ViewedResult.ResultInit.Name }}(vres)
			{{- else }}
			res := {{ .ResultInit.Name }}({{ range .ResultInit.ClientArgs }}{{ .Ref }},{{ end }})
				{{- range .Cookies }}
					{{- if .MaxAgeFrom }}{{ if .MaxAgeFrom.FieldPointer }}
				{{ .MaxAgeFrom.VarName }}Tmp := {{ .MaxAgeFrom.VarName }}
				res.{{ .MaxAgeFrom.FieldName }} = &{{ .MaxAgeFrom.VarName }}Tmp
					{{- else }}
				res.{{ .MaxAgeFrom.FieldName }} = {{ .MaxAgeFrom.VarName }}
					{{- end }}{{- end }}
					{{- if .DomainFrom }}{{ if .DomainFrom.FieldPointer }}
				{{ .DomainFrom.VarName }}Tmp := {{ .DomainFrom.VarName }}
				res.{{ .DomainFrom.FieldName }} = &{{ .DomainFrom.VarName }}Tmp
					{{- else }}
				res.{{ .DomainFrom.FieldName }} = {{ .DomainFrom.VarName }}
					{{- end }}{{- end }}
					{{- if .PathFrom }}{{ if .PathFrom.FieldPointer }}
				{{ .PathFrom.VarName }}Tmp := {{ .PathFrom.VarName }}
				res.{{ .PathFrom.FieldName }} = &{{ .PathFrom.VarName }}Tmp
					{{- else }}
				res.{{ .PathFrom.FieldName }} = {{ .PathFrom.VarName }}
					{{- end }}{{- end }}
					{{- if .SecureFrom }}{{ if .SecureFrom.FieldPointer }}
				{{ .SecureFrom.VarName }}Tmp := {{ .SecureFrom.VarName }}
				res.{{ .SecureFrom.FieldName }} = &{{ .SecureFrom.VarName }}Tmp
					{{- else }}
				res.{{ .SecureFrom.FieldName }} = {{ .SecureFrom.VarName }}
					{{- end }}{{- end }}
					{{- if .HTTPOnlyFrom }}{{ if .HTTPOnlyFrom.FieldPointer }}
				{{ .HTTPOnlyFrom.VarName }}Tmp := {{ .HTTPOnlyFrom.VarName }}
				res.{{ .HTTPOnlyFrom.FieldName }} = &{{ .HTTPOnlyFrom.VarName }}Tmp
					{{- else }}
				res.{{ .HTTPOnlyFrom.FieldName }} = {{ .HTTPOnlyFrom.VarName }}
					{{- end }}{{- end }}
					{{- if .SameSiteFrom }}{{ if .SameSiteFrom.FieldPointer }}
				{{ .SameSiteFrom.VarName }}Tmp := {{ .SameSiteFrom.VarName }}
				res.{{ .SameSiteFrom.FieldName }} = &{{ .SameSiteFrom.VarName }}Tmp
					{{- else }}
				res.{{ .SameSiteFrom.FieldName }} = {{ .SameSiteFrom.VarName }}
					{{- end }}{{- end }}
				{{- end }}
			{{- end }}
			{{- if and .TagName (not .ViewedResult) }}
				{{- if .TagPointer }}
					tmp := {{ printf "%q" .TagValue }}
					res.{{ .TagName }} = &tmp
				{{- else }}
					res.{{ .TagName }} = {{ printf "%q" .TagValue }}
				{{- end }}
			{{- end }}
			return res, nil
		{{- else if .ClientBody }}
			return body, nil
		{{- else if .Headers }}
			return {{ (index .Headers 0).VarName }}, nil
		{{- else if .Cookies }}
			return {{ (index .Cookies 0).VarName }}, nil
		{{- else }}
			return nil, nil
		{{- end }}
	{{- end }}
	{{- range .Errors }}
		case {{ .StatusCode }}:
		{{- if gt (len .Errors) 1 }}
		en := resp.Header.Get("goa-error")
		switch en {
			{{- range .Errors }}
		case {{ printf "%q" .Name }}:
				{{- with .Response }}
					{{- template "partial_single_response" (buildResponseData . $.ServiceName $.Method) }}
					{{- if .ResultInit }}
			return nil, {{ .ResultInit.Name }}({{ range .ResultInit.ClientArgs }}{{ .Ref }},{{ end }})
					{{- else if .ClientBody }}
			return nil, body
					{{- else }}
			return nil, nil
					{{- end }}
				{{- end }}
			{{- end }}
		default:
			body, _ := io.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse({{ printf "%q" $.ServiceName }}, {{ printf "%q" $.Method.Name }}, resp.StatusCode, string(body))
		}
		{{- else }}
			{{- with (index .Errors 0).Response }}
				{{- template "partial_single_response" (buildResponseData . $.ServiceName $.Method) }}
				{{- if .ResultInit }}
			return nil, {{ .ResultInit.Name }}({{ range .ResultInit.ClientArgs }}{{ .Ref }},{{ end }})
				{{- else if .ClientBody }}
			return nil, body
				{{- else }}
			return nil, nil
				{{- end }}
			{{- end }}
		{{- end }}
	{{- end }}
		default:
			body, _ := io.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse({{ printf "%q" .ServiceName }}, {{ printf "%q" .Method.Name }}, resp.StatusCode, string(body))
		}
	}
}
