{{ .TargetVar }} {{ if .NewVar }}:={{ else }}={{ end }} make({{ if .TypeAliasName }}{{ .TypeAliasName }}{{ else }}[]{{ if .TargetElemPointer }}*{{ end }}{{ .ElemTypeRef }}{{ end }}, len({{ .SourceVar }}))
for {{ .LoopVar }}, val := range {{ .SourceVar }} {
{{ if .SourceIsObject -}}
	if val == nil {
		{{ .TargetVar }}[{{ .LoopVar }}] = nil
		continue
	}
{{ end -}}
{{ if .TargetElemPointer -}}
	var transformed {{ .ElemTypeRef }}
{{ if .UseHelper -}}
	transformed = {{ transformHelperName .SourceElem .TargetElem .TransformAttrs }}({{ .SourceElement }})
{{ else -}}
	{{ transformAttribute .SourceElem .TargetElem .SourceElement "transformed" false .TransformAttrs -}}
{{ end -}}
	{{ .TargetVar }}[{{ .LoopVar }}] = &transformed
{{ else -}}
{{ if .UseHelper -}}
	{{ .TargetVar }}[{{ .LoopVar }}] = {{ transformHelperName .SourceElem .TargetElem .TransformAttrs }}({{ .SourceElement }})
{{ else -}}
	{{ transformAttribute .SourceElem .TargetElem .SourceElement (printf "%s[%s]" .TargetVar .LoopVar) false .TransformAttrs -}}
{{ end -}}
{{ end -}}
}
