{{ printf "%s returns a decoder for responses returned by the %s %s endpoint. restoreBody controls whether the response body should be restored after having been read." .ResponseDecoderDeclaration.Name .ServiceName .Method.Name | comment }}
{{- if .Errors }}
{{ printf "%s may return the following errors:" .ResponseDecoderDeclaration.Name | comment }}
	{{- range $gerr := .Errors }}
	{{- range $errors := .Errors }}
//	- {{ printf "%q" .Name }} (type {{ .Ref }}): {{ .Response.StatusCode }}{{ if .Response.Description }}, {{ .Response.Description }}{{ end }}
	{{- end }}
	{{- end }}
//	- error: internal error
{{- end }}
func {{ .ResponseDecoderDeclaration.Name }}(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (any, error) {
	return func(resp *http.Response) {{ if .Method.SkipResponseBodyEncodeDecode }}(any, error){{ else }}(result any, decodeErr error){{ end }} {
		{{- if not .Method.SkipResponseBodyEncodeDecode }}
		responseBody := resp.Body
		if restoreBody {
			b, readErr := io.ReadAll(responseBody)
			closeErr := responseBody.Close()
			if err := errors.Join(readErr, closeErr); err != nil {
				return nil, goahttp.ErrDecodingError("{{ .ServiceName }}", "{{ .Method.Name }}", err)
			}
			resp.Body = io.NopCloser(bytes.NewBuffer(b))
			defer func() {
				resp.Body = io.NopCloser(bytes.NewBuffer(b))
			}()
		} else {
			defer func() {
				if err := responseBody.Close(); err != nil {
					decodeErr = errors.Join(decodeErr, goahttp.ErrDecodingError("{{ .ServiceName }}", "{{ .Method.Name }}", err))
				}
			}()
		}
		{{- end }}
		switch resp.StatusCode {
	{{- range .Result.Responses }}
		case {{ .StatusCode }}:
		{{- if .SelectClientBodyByView }}
			view := resp.Header.Get("goa-view")
			if view == "" {
				view = "default"
			}
			switch view {
			{{- $response := . }}
			{{- range .ViewedRepresentations }}
			case {{ printf "%q" .View }}:
				{{- template "partial_single_response" (buildViewedResponseData $response . $.ServiceName $.Method) }}
				p := {{ .ResultInit.Declaration.Name }}({{ range .ResultInit.ClientArgs }}{{ .Ref }},{{ end }})
				{{- if $response.TagName }}
				tmp := {{ printf "%q" $response.TagValue }}
				p.{{ $response.TagName }} = &tmp
				{{- end }}
				vres := {{ if not $.Method.ViewedResult.IsCollection }}&{{ end }}{{ $.Method.ViewedResult.ViewsPkg}}.{{ $.Method.ViewedResult.VarName }}{Projected: p, View: view}
				{{- if .ClientBody }}
				if err = {{ $.Method.ViewedResult.ViewsPkg}}.{{ $.Method.ViewedResult.Validate.Declaration.Name }}(vres); err != nil {
					return nil, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", err)
				}
				{{- end }}
				res := {{ $.ServicePkgName }}.{{ $.Method.ViewedResult.ResultInit.Declaration.Name }}(vres)
				return res, nil
			{{- end }}
			default:
				return nil, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", goa.InvalidEnumValueError("view", view, []any{ {{ range $.Method.ViewedResult.Views }}{{ printf "%q" .Name }}, {{ end }} }))
			}
		{{- else }}
			{{- template "partial_single_response" (buildResponseData . $.ServiceName $.Method) }}
		{{- if .ResultInit }}
			{{- if .ViewedResult }}
			p := {{ .ResultInit.Declaration.Name }}({{ range .ResultInit.ClientArgs }}{{ .Ref }},{{ end }})
				{{- if .TagName }}
				tmp := {{ printf "%q" .TagValue }}
				p.{{ .TagName }} = &tmp
				{{- end }}
				{{- if $.Method.ViewedResult.ViewName }}
			view := {{ printf "%q" $.Method.ViewedResult.ViewName }}
				{{- else }}
			view := resp.Header.Get("goa-view")
			if view == "" {
				view = "default"
			}
				{{- end }}
			vres := {{ if not $.Method.ViewedResult.IsCollection }}&{{ end }}{{ $.Method.ViewedResult.ViewsPkg}}.{{ $.Method.ViewedResult.VarName }}{Projected: p, View: view}
				{{- if .ClientBody }}
				if err = {{ $.Method.ViewedResult.ViewsPkg}}.{{ $.Method.ViewedResult.Validate.Declaration.Name }}(vres); err != nil {
					return nil, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.Method.Name }}", err)
				}
				{{- end }}
			res := {{ $.ServicePkgName }}.{{ $.Method.ViewedResult.ResultInit.Declaration.Name }}(vres)
			{{- else }}
			res := {{ .ResultInit.Declaration.Name }}({{ range .ResultInit.ClientArgs }}{{ .Ref }},{{ end }})
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
			return nil, {{ .ResultInit.Declaration.Name }}({{ range .ResultInit.ClientArgs }}{{ .Ref }},{{ end }})
					{{- else if .ClientBody }}
			return nil, body
					{{- else }}
			return nil, nil
					{{- end }}
				{{- end }}
			{{- end }}
		default:
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("{{ $.ServiceName }}", "{{ $.Method.Name }}", err)
			}
			return nil, goahttp.ErrInvalidResponse({{ printf "%q" $.ServiceName }}, {{ printf "%q" $.Method.Name }}, resp.StatusCode, string(body))
		}
		{{- else }}
			{{- with (index .Errors 0).Response }}
				{{- template "partial_single_response" (buildResponseData . $.ServiceName $.Method) }}
				{{- if .ResultInit }}
			return nil, {{ .ResultInit.Declaration.Name }}({{ range .ResultInit.ClientArgs }}{{ .Ref }},{{ end }})
				{{- else if .ClientBody }}
			return nil, body
				{{- else }}
			return nil, nil
				{{- end }}
			{{- end }}
		{{- end }}
	{{- end }}
		default:
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("{{ .ServiceName }}", "{{ .Method.Name }}", err)
			}
			return nil, goahttp.ErrInvalidResponse({{ printf "%q" .ServiceName }}, {{ printf "%q" .Method.Name }}, resp.StatusCode, string(body))
		}
	}
}
