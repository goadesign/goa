{{- range .Cookies }}
	{{- if .MaxAgeFrom }}
		{{- if .MaxAgeFrom.FieldPointer }}
		{{ $.Target }}.{{ .MaxAgeFrom.FieldName }} = goahttp.CookieIntAttr({{ .MaxAgeFrom.VarName }})
		{{- else }}
		{{ $.Target }}.{{ .MaxAgeFrom.FieldName }} = {{ .MaxAgeFrom.VarName }}
		{{- end }}
	{{- end }}
	{{- if .DomainFrom }}
		{{- if .DomainFrom.FieldPointer }}
		{{ $.Target }}.{{ .DomainFrom.FieldName }} = goahttp.CookieStringAttr({{ .DomainFrom.VarName }})
		{{- else }}
		{{ $.Target }}.{{ .DomainFrom.FieldName }} = {{ .DomainFrom.VarName }}
		{{- end }}
	{{- end }}
	{{- if .PathFrom }}
		{{- if .PathFrom.FieldPointer }}
		{{ $.Target }}.{{ .PathFrom.FieldName }} = goahttp.CookieStringAttr({{ .PathFrom.VarName }})
		{{- else }}
		{{ $.Target }}.{{ .PathFrom.FieldName }} = {{ .PathFrom.VarName }}
		{{- end }}
	{{- end }}
	{{- if .SecureFrom }}
		{{- if .SecureFrom.FieldPointer }}
		{{ $.Target }}.{{ .SecureFrom.FieldName }} = goahttp.CookieBoolAttr({{ .SecureFrom.VarName }})
		{{- else }}
		{{ $.Target }}.{{ .SecureFrom.FieldName }} = {{ .SecureFrom.VarName }}
		{{- end }}
	{{- end }}
	{{- if .HTTPOnlyFrom }}
		{{- if .HTTPOnlyFrom.FieldPointer }}
		{{ $.Target }}.{{ .HTTPOnlyFrom.FieldName }} = goahttp.CookieBoolAttr({{ .HTTPOnlyFrom.VarName }})
		{{- else }}
		{{ $.Target }}.{{ .HTTPOnlyFrom.FieldName }} = {{ .HTTPOnlyFrom.VarName }}
		{{- end }}
	{{- end }}
	{{- if .SameSiteFrom }}
		{{- if .SameSiteFrom.FieldPointer }}
		{{ $.Target }}.{{ .SameSiteFrom.FieldName }} = goahttp.CookieSameSiteAttr({{ .SameSiteFrom.VarName }})
		{{- else }}
		{{ $.Target }}.{{ .SameSiteFrom.FieldName }} = {{ .SameSiteFrom.VarName }}
		{{- end }}
	{{- end }}
{{- end }}
