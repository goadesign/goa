{{ if .isPointer }}if {{ .target }} != nil {
{{ end -}}
if {{ .targetVal }} {{ if .isExclMin }}<={{ else }}>={{ end }} {{ if .isExclMin }}{{ .exclMin }}{{ else }}{{ .exclMax }}{{ end }} {
        err = goa.MergeErrors(err, goa.InvalidRangeError({{ printf "%q" .context }}, {{ .targetVal }}, {{ if .isExclMin }}{{ .exclMin }}, true{{ else }}{{ .exclMax }}, false{{ end }}))
}
{{- if .isPointer }}
}
{{- end }}