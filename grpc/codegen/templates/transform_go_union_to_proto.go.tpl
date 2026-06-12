switch string({{ if .SourcePtr }}(*{{ .SourceVar }}){{ else }}{{ .SourceVar }}{{ end }}.Kind()) {
{{- range .Cases }}
case {{ printf "%q" .TypeTag }}:
	actual, _ := {{ if $.SourcePtr }}(*{{ $.SourceVar }}){{ else }}{{ $.SourceVar }}{{ end }}.As{{ .SourceFieldName }}()
	{{ $.TargetVar }} = &{{ .TargetWrapperType }}{ {{ .TargetFieldName }}: {{ .ConvertedValue }} }
{{- end }}
}
