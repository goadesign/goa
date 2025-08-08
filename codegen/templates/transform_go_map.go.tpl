{{ .TargetVar }} {{ if .NewVar }}:={{ else }}={{ end }} make({{ if .TypeAliasName }}{{ .TypeAliasName }}{{ else }}map[{{ .KeyTypeRef }}]{{ .ElemTypeRef }}{{ end }}, len({{ .SourceVar }}))
for key, val := range {{ .SourceVar }} {
{{ if .IsKeyStruct -}}
	tk := {{ transformHelperName .SourceKey .TargetKey .TransformAttrs -}}(val)
{{ else -}}
	{{ transformAttribute .SourceKey .TargetKey "key" "tk" true .TransformAttrs }}{{ end -}}
{{ if .IsElemStruct -}}
	if val == nil {
		{{ .TargetVar }}[tk] = nil
		continue
	}
	{{ .TargetVar }}[tk] = {{ transformHelperName .SourceElem .TargetElem .TransformAttrs -}}(val)
{{ else -}}
	{{ transformAttribute .SourceElem .TargetElem "val" (printf "tv%s" .LoopVar) true .TransformAttrs -}}
	{{ .TargetVar }}[tk] = {{ printf "tv%s" .LoopVar -}}
{{ end -}}
}
