switch string({{ if .SourcePtr }}(*{{ .SourceVar }}){{ else }}{{ .SourceVar }}{{ end }}.Kind()) {
{{- range .Cases }}
case {{ printf "%q" .typeTag }}:
	actual, _ := {{ if $.SourcePtr }}(*{{ $.SourceVar }}){{ else }}{{ $.SourceVar }}{{ end }}.As{{ .sourceFieldName }}()
	{{- $val := (convertType .sourceAttr .targetAttr false false "actual" $.TransformAttrs) }}
	{{ $.TargetVar }} = &{{ .targetWrapperType }}{ {{ .targetFieldName }}: {{ $val }} }
{{- end }}
}
