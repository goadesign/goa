// {{ . }} keeps an omitted command-line flag distinct from an explicitly empty flag.
type {{ . }} struct {
	value *string
}

// String returns the flag text shown by the standard flag package.
func (f *{{ . }}) String() string {
	if f.value == nil {
		return ""
	}
	return *f.value
}

// Set records that the user supplied the flag, even when value is empty.
func (f *{{ . }}) Set(value string) error {
	f.value = &value
	return nil
}
