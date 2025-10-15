{{ .TargetVar }} {{ if .NewVar }}:={{ else }}={{ end }} make({{ if .TypeAliasName }}{{ .TypeAliasName }}{{ else }}[]{{ .ElemTypeRef }}{{ end }}, len({{ .SourceVar }}))
for {{ .LoopVar }}, val := range {{ .SourceVar }} {
{{ if .IsStruct -}}
	{{ .TargetVar }}[{{ .LoopVar }}] = {{ transformHelperName .SourceElem .TargetElem .TransformAttrs }}(val)
{{ else -}}
	{{ transformAttribute .SourceElem .TargetElem "val" (printf "%s[%s]" .TargetVar .LoopVar) false .TransformAttrs -}}
{{ end -}}
}
