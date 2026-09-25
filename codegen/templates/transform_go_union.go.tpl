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
		{{ $.TempVarName }} = {{ transformHelperName .SourceAttr .TargetAttr .TransformAttrs }}(actual)
	{{- else }}
		{{ transformAttribute .SourceAttr .TargetAttr "actual" $.TempVarName false true true .TransformAttrs -}}
	{{- end }}
	}
	{{- else }}
	{{- if .UseHelper }}
	{{ $.TempVarName }} := {{ transformHelperName .SourceAttr .TargetAttr .TransformAttrs }}(actual)
	{{- else }}
	{{ transformAttribute .SourceAttr .TargetAttr "actual" $.TempVarName true true true .TransformAttrs -}}
	{{- end }}
	{{- end }}
	{{- if $.TargetPointer }}
	var u {{ $.ValueTypeRef }}
	{{ $.TargetReceiver }}.Set{{ .TargetFieldName }}(({{ .TargetCastType }})({{ $.TempVarName }}))
	{{ $.TargetVar }} = &u
	{{- else }}
	u := {{ $.TargetVar }}
	{{ $.TargetReceiver }}.Set{{ .TargetFieldName }}(({{ .TargetCastType }})({{ $.TempVarName }}))
	{{ $.TargetVar }} = u
	{{- end }}
{{- end }}
}
