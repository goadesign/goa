{{ .TargetVar }} {{ if .NewVar }}:={{ else }}={{ end }} make({{ if .TypeAliasName }}{{ .TypeAliasName }}{{ else }}map[{{ .KeyTypeRef }}]{{ .ElemTypeRef }}{{ end }}, len({{ .SourceVar }}))
for key, val := range {{ .SourceVar }} {
{{ if .UseKeyHelper -}}
	tk := {{ transformHelperName .SourceKey .TargetKey .TransformAttrs -}}(key)
{{ else -}}
	{{ transformAttribute .SourceKey .TargetKey "key" "tk" true false false .TransformAttrs }}{{ end -}}
{{ if .ElemIsNilable -}}
	if val == nil {
		{{ .TargetVar }}[tk] = nil
		continue
	}
{{ end -}}
{{ if .UseElemHelper -}}
	{{ .TargetVar }}[tk] = {{ transformHelperName .SourceElem .TargetElem .TransformAttrs -}}(val)
{{ else -}}
{{ if .ElemIsUnion -}}
	var {{ printf "tv%s" .LoopVar }} {{ .ElemTypeRef }}
{{ end -}}
	{{ transformAttribute .SourceElem .TargetElem "val" (printf "tv%s" .LoopVar) (not .ElemIsUnion) false false .TransformAttrs -}}
	{{ .TargetVar }}[tk] = {{ printf "tv%s" .LoopVar -}}
{{ end -}}
}
