if err2 := Validate{{ .name }}({{ .target }}); err2 != nil {
        err = goa.MergeErrors(err, err2)
}