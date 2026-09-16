{{- if .TargetPtr -}}
{{- $m := printf "m%s" .LoopVar -}}
{{ $m }} := make(map[{{ .KeyTypeRef }}]{{ .ElemTypeRef }}, len({{ .SourceVar }}))
for key, val := range {{ .SourceVar }} {
    {{ transformAttribute .SourceKey .TargetKey "key" "tk" true .TransformAttrs -}}
    {{ .ElemTransform -}}
    {{ $m }}[tk] = {{ printf "tv%s" .LoopVar }}
}
{{ .TargetVar }} = &{{ $m }}
{{- else -}}
{{ .TargetVar }} {{ if .NewVar }}:={{ else }}={{ end }} make(map[{{ .KeyTypeRef }}]{{ .ElemTypeRef }}, len({{ .SourceVar }}))
for key, val := range {{ .SourceVar }} {
    {{ transformAttribute .SourceKey .TargetKey "key" "tk" true .TransformAttrs -}}
    {{ .ElemTransform -}}
    {{ .TargetVar }}[tk] = {{ printf "tv%s" .LoopVar }}
}
{{- end -}}
