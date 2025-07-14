if {{ $.target }}.{{ .attCtx.Scope.Field $.reqAtt .req true }} == nil {
        err = goa.MergeErrors(err, goa.MissingFieldError("{{ .req }}", {{ printf "%q" $.context }}))
}