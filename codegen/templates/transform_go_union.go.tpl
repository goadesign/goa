{{- if .NewVar }}var {{ .TargetVar }} {{ .TypeRef }}
{{ end -}}
switch string({{ .SourceVar }}.Kind()) {
{{- range .Cases }}
case {{ printf "%q" .CaseName }}:
	actual, _ := {{ $.SourceVar }}.As{{ .SourceFieldName }}()
	{{ transformAttribute .SourceAttr .TargetAttr "actual" $.TempVarName true $.TransformAttrs -}}
	{{- if $.NewVar }}
	var u {{ $.ValueTypeRef }}
	u.Set{{ .TargetFieldName }}(({{ .TargetCastType }})({{ $.TempVarName }}))
	{{- if $.TargetIsPointer }}
	{{ $.TargetVar }} = &u
	{{- else }}
	{{ $.TargetVar }} = u
	{{- end }}
	{{- else }}
	u := {{ $.TargetVar }}
	u.Set{{ .TargetFieldName }}(({{ .TargetCastType }})({{ $.TempVarName }}))
	{{ $.TargetVar }} = u
	{{- end }}
{{- end }}
}
