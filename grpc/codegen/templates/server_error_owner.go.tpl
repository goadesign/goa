// {{ .Owner.Name }} finds the name and value that supply a declared response.
// If an error lists several causes, only a named error outside that list can
// supply the response. Otherwise, the caller receives generic complete details.
func {{ .Owner.Name }}(err error) (goa.GoaErrorNamer, error, bool) {
	var name goa.GoaErrorNamer
	var owner error
	var explicit interface{ GRPCStatus() *status.Status }
	for current := err; current != nil; {
		if name == nil {
			if explicit == nil {
				explicit, _ = current.(interface{ GRPCStatus() *status.Status })
			}
			if candidate, ok := current.(goa.GoaErrorNamer); ok {
				name, owner = candidate, current
			}
		}
		next, independent := {{ .Next.Name }}(current)
		if independent {
			if name == nil || (explicit != nil && explicit.GRPCStatus() != nil) {
				return nil, nil, false
			}
			return name, owner, true
		}
		current = next
	}

	// If every error has at most one cause, keep the first name found through
	// custom As methods. A status supplied before that name takes precedence.
	statusSeen := false
	for current := err; current != nil; {
		if !statusSeen {
			if explicit, ok := current.(interface{ GRPCStatus() *status.Status }); ok {
				statusSeen = true
				if explicit.GRPCStatus() != nil {
					return nil, nil, false
				}
			}
		}
		if name, ok := current.(goa.GoaErrorNamer); ok {
			return name, err, false
		}
		if as, ok := current.(interface{ As(any) bool }); ok {
			var name goa.GoaErrorNamer
			if as.As(&name) {
				if name == nil {
					panic("custom As returned a nil error name")
				}
				return name, err, false
			}
		}
		current, _ = {{ .Next.Name }}(current)
	}
	return nil, nil, false
}

// {{ .Next.Name }} follows one non-nil cause. Several cause entries remain
// independent even when they have equal values, names, or status codes.
func {{ .Next.Name }}(err error) (error, bool) {
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		var next error
		for _, cause := range joined.Unwrap() {
			if cause != nil {
				if next != nil {
					return nil, true
				}
				next = cause
			}
		}
		return next, false
	}
	return errors.Unwrap(err), false
}
