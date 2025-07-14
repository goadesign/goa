{{ $target := or (and (or (or .array .map) .nonzero) .target) .targetVal -}}
{{ if and .isPointer .string -}}
if {{ .target }} != nil {
{{ end -}}
if {{ if .string }}utf8.RuneCountInString({{ $target }}){{ else }}len({{ $target }}){{ end }} {{ if .isMinLength }}<{{ else }}>{{ end }} {{ if .isMinLength }}{{ .minLength }}{{ else }}{{ .maxLength }}{{ end }} {
        err = goa.MergeErrors(err, goa.InvalidLengthError({{ printf "%q" .context }}, {{ $target }}, {{ if .string }}utf8.RuneCountInString({{ $target }}){{ else }}len({{ $target }}){{ end }}, {{ if .isMinLength }}{{ .minLength }}, true{{ else }}{{ .maxLength }}, false{{ end }}))
}{{- if and .isPointer .string }}
}
{{- end }}