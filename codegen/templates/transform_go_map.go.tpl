{{ .TargetVar }} {{ if .NewVar }}:={{ else }}={{ end }} make({{ if .TypeAliasName }}{{ .TypeAliasName }}{{ else }}map[{{ .KeyTypeRef }}]{{ .ElemTypeRef }}{{ end }}, len({{ .SourceVar }}))
for key, val := range {{ .SourceVar }} {
{{ if .UseKeyHelper -}}
	tk := {{ transformHelperName .SourceKey .TargetKey .TransformAttrs -}}(key)
{{ else -}}
	{{ transformAttribute .SourceKey .TargetKey "key" "tk" true .TransformAttrs }}{{ end -}}
{{ if .ElemIsObject -}}
	if val == nil {
		{{ .TargetVar }}[tk] = nil
		continue
	}
{{ end -}}
{{ if .UseElemHelper -}}
	{{ .TargetVar }}[tk] = {{ transformHelperName .SourceElem .TargetElem .TransformAttrs -}}(val)
{{ else -}}
	{{ transformAttribute .SourceElem .TargetElem "val" (printf "tv%s" .LoopVar) true .TransformAttrs -}}
	{{ .TargetVar }}[tk] = {{ printf "tv%s" .LoopVar -}}
{{ end -}}
}
