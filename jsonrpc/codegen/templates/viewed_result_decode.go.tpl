{{ printf "%s decodes the JSON body selected by the result view for the %s service %s method." .Decode.Name .ServiceName .MethodName | comment }}
func {{ .Decode.Name }}(decoder func(*http.Response) goahttp.Decoder, resp *http.Response, data json.RawMessage{{ if .SSE }}{{ if .SSE.IDField }}, eventID string, hasEventID bool{{ end }}{{ if .SSE.EventField }}, eventType string, hasEventType bool{{ end }}{{ if .SSE.RetryField }}, eventRetry string, hasEventRetry bool{{ end }}{{ end }}) ({{ .ResultRef }}, error) {
	{{- if .Variable }}
	var representation struct {
		View *string          `json:"view"`
		Body *json.RawMessage `json:"body"`
	}
	if err := {{ .BodyDecoder.Name }}(decoder, data, &representation); err != nil {
		return nil, goahttp.ErrDecodingError("{{ .ServiceName }}", "{{ .MethodName }}", err)
	}
	if representation.View == nil {
		return nil, goahttp.ErrValidationError("{{ .ServiceName }}", "{{ .MethodName }}", goa.MissingFieldError("view", "result"))
	}
	view := *representation.View
	switch view {
	{{- range .Branches }}
	case {{ printf "%q" .View }}:
		{{- if .ClientBody }}
		if representation.Body == nil {
			return nil, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.MethodName }}", goa.MissingFieldError("body", "result"))
		}
		resp.Body = io.NopCloser(bytes.NewBuffer(*representation.Body))
		{{- end }}
		{{- template "partial_single_response" (viewedResponseData . $.ServiceName $.MethodName $.SSE) }}
		projected := {{ .ResultInit.Declaration.Name }}({{ range .ResultInit.ClientArgs }}{{ .Ref }},{{ end }})
		{{ $.ViewedValue }} := {{ if not $.IsCollection }}&{{ end }}{{ $.ViewedPkg }}.{{ $.ViewedVarName }}{
			Projected: projected,
			View:      view,
		}
		if err := {{ $.ViewedPkg }}.{{ $.ViewedValidator }}({{ $.ViewedValue }}); err != nil {
			return nil, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.MethodName }}", err)
		}
		result := {{ $.ServicePkg }}.{{ $.ServiceResultConstructor }}({{ $.ViewedValue }})
		{{- if $.SSE }}
		{{- template "viewed_sse_result_fields" $ }}
		{{- end }}
		return result, nil
	{{- end }}
	default:
		return nil, goahttp.ErrValidationError("{{ .ServiceName }}", "{{ .MethodName }}", goa.InvalidEnumValueError("view", view, []any{
			{{- range .Branches }}{{ printf "%q" .View }},{{ end }}
		}))
	}
	{{- else }}
	{{- with index .Branches 0 }}
	{{- if .ClientBody }}
	resp.Body = io.NopCloser(bytes.NewBuffer(data))
	{{- end }}
	{{- template "partial_single_response" (viewedResponseData . $.ServiceName $.MethodName $.SSE) }}
	projected := {{ .ResultInit.Declaration.Name }}({{ range .ResultInit.ClientArgs }}{{ .Ref }},{{ end }})
	{{ $.ViewedValue }} := {{ if not $.IsCollection }}&{{ end }}{{ $.ViewedPkg }}.{{ $.ViewedVarName }}{
		Projected: projected,
		View:      {{ printf "%q" $.FixedView }},
	}
	if err := {{ $.ViewedPkg }}.{{ $.ViewedValidator }}({{ $.ViewedValue }}); err != nil {
		return nil, goahttp.ErrValidationError("{{ $.ServiceName }}", "{{ $.MethodName }}", err)
	}
	result := {{ $.ServicePkg }}.{{ $.ServiceResultConstructor }}({{ $.ViewedValue }})
	{{- if $.SSE }}
	{{- template "viewed_sse_result_fields" $ }}
	{{- end }}
	return result, nil
	{{- end }}
	{{- end }}
}

{{- define "viewed_sse_outer_fields" }}
	{{- if .SSE.IDField }}
	if !hasEventID {
		{{- if .SSE.ID.HasDefault }}
		value := {{ .SSE.ID.ClientTypeRef }}({{ printf "%q" .SSE.ID.DefaultValue }})
		body.{{ .SSE.IDField }} = {{ if .SSE.ClientIDPointer }}&{{ end }}value
		{{- else if not .SSE.ID.Pointer }}
		return nil, goahttp.ErrValidationError("{{ .ServiceName }}", "{{ .MethodName }}", fmt.Errorf("server-sent event has no id for result field {{ .SSE.IDField }}"))
		{{- end }}
	} else {
		value := {{ .SSE.ID.ClientTypeRef }}(eventID)
		body.{{ .SSE.IDField }} = {{ if .SSE.ClientIDPointer }}&{{ end }}value
	}
	{{- end }}
	{{- if .SSE.EventField }}
	if !hasEventType {
		{{- if .SSE.Event.HasDefault }}
		value := {{ .SSE.Event.ClientTypeRef }}({{ printf "%q" .SSE.Event.DefaultValue }})
		body.{{ .SSE.EventField }} = {{ if .SSE.ClientEventPointer }}&{{ end }}value
		{{- else if not .SSE.Event.Pointer }}
		return nil, goahttp.ErrValidationError("{{ .ServiceName }}", "{{ .MethodName }}", fmt.Errorf("server-sent event has no type for result field {{ .SSE.EventField }}"))
		{{- end }}
	} else {
		value := {{ .SSE.Event.ClientTypeRef }}(eventType)
		body.{{ .SSE.EventField }} = {{ if .SSE.ClientEventPointer }}&{{ end }}value
	}
	{{- end }}
	{{- if .SSE.RetryField }}
	if !hasEventRetry {
		{{- if .SSE.Retry.HasDefault }}
		value := {{ .SSE.Retry.ClientTypeRef }}({{ printf "%v" .SSE.Retry.DefaultValue }})
		body.{{ .SSE.RetryField }} = {{ if .SSE.Retry.ClientPointer }}&{{ end }}value
		{{- else if not .SSE.Retry.Pointer }}
		return nil, goahttp.ErrValidationError("{{ .ServiceName }}", "{{ .MethodName }}", fmt.Errorf("server-sent event has no retry for result field {{ .SSE.RetryField }}"))
		{{- end }}
	} else {
		{{- if sseRetrySigned .SSE.Retry }}
		parsed, parseErr := strconv.ParseInt(eventRetry, 10, {{ sseRetryBits .SSE.Retry }})
		if parseErr != nil || parsed < 0 {
			return nil, goahttp.ErrDecodingError("{{ .ServiceName }}", "{{ .MethodName }}", fmt.Errorf("invalid server-sent event retry %q", eventRetry))
		}
		{{- else }}
		parsed, parseErr := strconv.ParseUint(eventRetry, 10, {{ sseRetryBits .SSE.Retry }})
		if parseErr != nil {
			return nil, goahttp.ErrDecodingError("{{ .ServiceName }}", "{{ .MethodName }}", fmt.Errorf("invalid server-sent event retry %q", eventRetry))
		}
		{{- end }}
		value := {{ .SSE.Retry.ClientTypeRef }}(parsed)
		body.{{ .SSE.RetryField }} = {{ if .SSE.Retry.ClientPointer }}&{{ end }}value
	}
	{{- end }}
{{- end }}

{{- define "viewed_sse_result_fields" }}
	{{- if .SSE.IDField }}
	{{- if .SSE.ID.Pointer }}
	if body.{{ .SSE.IDField }} != nil {
		value := {{ .SSE.ID.TypeRef }}(*body.{{ .SSE.IDField }})
		result.{{ .SSE.IDField }} = &value
	}
	{{- else }}
	result.{{ .SSE.IDField }} = {{ .SSE.ID.TypeRef }}(*body.{{ .SSE.IDField }})
	{{- end }}
	{{- end }}
	{{- if .SSE.EventField }}
	{{- if .SSE.Event.Pointer }}
	if body.{{ .SSE.EventField }} != nil {
		value := {{ .SSE.Event.TypeRef }}(*body.{{ .SSE.EventField }})
		result.{{ .SSE.EventField }} = &value
	}
	{{- else }}
	result.{{ .SSE.EventField }} = {{ .SSE.Event.TypeRef }}(*body.{{ .SSE.EventField }})
	{{- end }}
	{{- end }}
	{{- if .SSE.RetryField }}
	{{- if .SSE.Retry.Pointer }}
	if body.{{ .SSE.RetryField }} != nil {
		value := {{ .SSE.Retry.TypeRef }}(*body.{{ .SSE.RetryField }})
		result.{{ .SSE.RetryField }} = &value
	}
	{{- else }}
	result.{{ .SSE.RetryField }} = {{ .SSE.Retry.TypeRef }}(*body.{{ .SSE.RetryField }})
	{{- end }}
	{{- end }}
{{- end }}
