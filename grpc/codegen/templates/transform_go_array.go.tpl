{{ .TargetVar }} {{ if .NewVar }}:={{ else }}={{ end }} make([]{{ .ElemTypeRef }}, len({{ .SourceVar }})) 
for {{ .LoopVar }}{{ if .ValVar }}, {{ .ValVar }}{{ end }} := range {{ .SourceVar }} {
  {{ transformAttribute .SourceElem .TargetElem "val" (printf "%s[%s]" .TargetVar .LoopVar) false .TransformAttrs -}}
}
