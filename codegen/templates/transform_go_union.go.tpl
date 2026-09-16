{{- if .NewVar }}var {{ .TargetVar }} {{ .TypeRef }}
{{ end -}}
switch string({{ .SourceVar }}.Kind()) {
{{- range .Cases }}
case {{ printf "%q" .CaseName }}:
	actual, _ := {{ $.SourceVar }}.As{{ .SourceFieldName }}()
	{{- if .SourceNilable }}
	var {{ $.TempVarName }} {{ .TargetCastType }}
	if actual != nil {
	{{- if .UseHelper }}
		{{ $.TempVarName }} = {{ .HelperName }}(actual)
	{{- else }}
		{{ transformAttribute .SourceAttr .TargetAttr "actual" $.TempVarName false $.TransformAttrs -}}
	{{- end }}
	}
	{{- else }}
	{{- if .UseHelper }}
	{{ $.TempVarName }} := {{ .HelperName }}(actual)
	{{- else }}
	{{ transformAttribute .SourceAttr .TargetAttr "actual" $.TempVarName true $.TransformAttrs -}}
	{{- end }}
	{{- end }}
	{{- if $.NewVar }}
	var u {{ $.ValueTypeRef }}
	u.Set{{ .TargetFieldName }}(({{ .TargetCastType }})({{ $.TempVarName }}))
	{{ $.TargetVar }} = &u
	{{- else }}
	u := {{ $.TargetVar }}
	u.Set{{ .TargetFieldName }}(({{ .TargetCastType }})({{ $.TempVarName }}))
	{{ $.TargetVar }} = u
	{{- end }}
{{- end }}
}
