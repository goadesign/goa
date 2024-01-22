// Error returns an error description.
func (e {{ .Ref }}) Error() string {
	return {{ printf "%q" .Description }}
}

// ErrorName returns {{ printf "%q" .Name }}.
//
// Deprecated: Use GoaErrorName - https://github.com/goadesign/goa/issues/3105
func (e {{ .Ref }}) ErrorName() string {
	return e.GoaErrorName()
}

// GoaErrorName returns {{ printf "%q" .Name }}.
func (e {{ .Ref }}) GoaErrorName() string {
	return {{ errorName . }}
}
