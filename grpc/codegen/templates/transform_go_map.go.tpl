{{ .TargetVar }} {{ if .NewVar }}:={{ else }}={{ end }} make(map[{{ .KeyTypeRef }}]{{ .ElemTypeRef }}, len({{ .SourceVar }}))
for key, val := range {{ .SourceVar }} {
	{{ transformAttribute .SourceKey .TargetKey "key" "tk" true .TransformAttrs -}}
	{{ transformAttribute .SourceElem .TargetElem "val" (printf "tv%s" .LoopVar) true .TransformAttrs -}}
	{{ .TargetVar }}[tk] = {{ printf "tv%s" .LoopVar }}
}
