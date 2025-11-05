{{- /*
   Generate array transform supporting pointer-to-slice targets.
   When TargetPtr is true, allocate a local slice variable and assign its
   address to the target field after populating. Otherwise assign directly
   to the target slice field.
*/ -}}
{{- if .TargetPtr -}}
{{- $arr := printf "arr%s" .LoopVar -}}
{{ $arr }} := make([]{{ .ElemTypeRef }}, len({{ if .SourcePtr }}*{{ end }}{{ .SourceVar }}))
for {{ .LoopVar }}{{ if .ValVar }}, {{ .ValVar }}{{ end }} := range {{ if .SourcePtr }}*{{ end }}{{ .SourceVar }} {
  {{ transformAttribute .SourceElem .TargetElem "val" (printf "%s[%s]" $arr .LoopVar) false .TransformAttrs -}}
}
{{ .TargetVar }} = &{{ $arr }}
{{- else -}}
{{ .TargetVar }} {{ if .NewVar }}:={{ else }}={{ end }} make([]{{ .ElemTypeRef }}, len({{ if .SourcePtr }}*{{ end }}{{ .SourceVar }}))
for {{ .LoopVar }}{{ if .ValVar }}, {{ .ValVar }}{{ end }} := range {{ if .SourcePtr }}*{{ end }}{{ .SourceVar }} {
  {{ transformAttribute .SourceElem .TargetElem "val" (printf "%s[%s]" .TargetVar .LoopVar) false .TransformAttrs -}}
}
{{- end -}}
