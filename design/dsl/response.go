package dsl

// Status sets the ResponseTemplate status
func Status(status int) error {
	if r, ok := responseDefinition(); ok {
		r.Status = status
	}
	return nil
}
