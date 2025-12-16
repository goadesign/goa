{{- if and (isUnion .reqAtt) (isAttributeScope .attCtx.Scope) }}
if {{ $.target }}.{{ .attCtx.Scope.Field $.reqAtt .req true }}.Kind() == "" {
        err = goa.MergeErrors(err, goa.MissingFieldError("{{ .req }}", {{ printf "%q" $.context }}))
}
{{- else }}
if {{ $.target }}.{{ .attCtx.Scope.Field $.reqAtt .req true }} == nil {
        err = goa.MergeErrors(err, goa.MissingFieldError("{{ .req }}", {{ printf "%q" $.context }}))
}
{{- end }}
