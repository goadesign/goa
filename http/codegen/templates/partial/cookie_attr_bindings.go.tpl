{{- range .Cookies }}
	{{- if .MaxAgeFrom }}
		{{- if .MaxAgeFrom.FieldPointer }}
		if {{ .MaxAgeFrom.VarName }} != 0 {
			{{ .MaxAgeFrom.VarName }}Tmp := {{ .MaxAgeFrom.VarName }}
			{{ $.Target }}.{{ .MaxAgeFrom.FieldName }} = &{{ .MaxAgeFrom.VarName }}Tmp
		}
		{{- else }}
		{{ $.Target }}.{{ .MaxAgeFrom.FieldName }} = {{ .MaxAgeFrom.VarName }}
		{{- end }}
	{{- end }}
	{{- if .DomainFrom }}
		{{- if .DomainFrom.FieldPointer }}
		if {{ .DomainFrom.VarName }} != "" {
			{{ .DomainFrom.VarName }}Tmp := {{ .DomainFrom.VarName }}
			{{ $.Target }}.{{ .DomainFrom.FieldName }} = &{{ .DomainFrom.VarName }}Tmp
		}
		{{- else }}
		{{ $.Target }}.{{ .DomainFrom.FieldName }} = {{ .DomainFrom.VarName }}
		{{- end }}
	{{- end }}
	{{- if .PathFrom }}
		{{- if .PathFrom.FieldPointer }}
		if {{ .PathFrom.VarName }} != "" {
			{{ .PathFrom.VarName }}Tmp := {{ .PathFrom.VarName }}
			{{ $.Target }}.{{ .PathFrom.FieldName }} = &{{ .PathFrom.VarName }}Tmp
		}
		{{- else }}
		{{ $.Target }}.{{ .PathFrom.FieldName }} = {{ .PathFrom.VarName }}
		{{- end }}
	{{- end }}
	{{- if .SecureFrom }}
		{{- if .SecureFrom.FieldPointer }}
		if {{ .SecureFrom.VarName }} {
			{{ .SecureFrom.VarName }}Tmp := {{ .SecureFrom.VarName }}
			{{ $.Target }}.{{ .SecureFrom.FieldName }} = &{{ .SecureFrom.VarName }}Tmp
		}
		{{- else }}
		{{ $.Target }}.{{ .SecureFrom.FieldName }} = {{ .SecureFrom.VarName }}
		{{- end }}
	{{- end }}
	{{- if .HTTPOnlyFrom }}
		{{- if .HTTPOnlyFrom.FieldPointer }}
		if {{ .HTTPOnlyFrom.VarName }} {
			{{ .HTTPOnlyFrom.VarName }}Tmp := {{ .HTTPOnlyFrom.VarName }}
			{{ $.Target }}.{{ .HTTPOnlyFrom.FieldName }} = &{{ .HTTPOnlyFrom.VarName }}Tmp
		}
		{{- else }}
		{{ $.Target }}.{{ .HTTPOnlyFrom.FieldName }} = {{ .HTTPOnlyFrom.VarName }}
		{{- end }}
	{{- end }}
	{{- if .SameSiteFrom }}
		{{- if .SameSiteFrom.FieldPointer }}
		if {{ .SameSiteFrom.VarName }} != "" && {{ .SameSiteFrom.VarName }} != "Default" {
			{{ .SameSiteFrom.VarName }}Tmp := {{ .SameSiteFrom.VarName }}
			{{ $.Target }}.{{ .SameSiteFrom.FieldName }} = &{{ .SameSiteFrom.VarName }}Tmp
		}
		{{- else }}
		{{ $.Target }}.{{ .SameSiteFrom.FieldName }} = {{ .SameSiteFrom.VarName }}
		{{- end }}
	{{- end }}
{{- end }}
