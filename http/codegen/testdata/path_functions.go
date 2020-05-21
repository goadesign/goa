package testdata

var PathNoParamCode = `// MethodPathNoParamServicePathNoParamPath returns the URL path to the ServicePathNoParam service MethodPathNoParam HTTP endpoint.
func MethodPathNoParamServicePathNoParamPath() string {
	return "/one/two"
}
`

var BasePathNoTrailing_SlashWithBasePathNoTrailingCode = `// SlashWithBasePathNoTrailingBasePathNoTrailingPath returns the URL path to the BasePathNoTrailing service SlashWithBasePathNoTrailing HTTP endpoint.
func SlashWithBasePathNoTrailingBasePathNoTrailingPath() string {
	return "/foo"
}
`

var BasePathNoTrailing_TrailingWithBasePathNoTrailingCode = `// TrailingWithBasePathNoTrailingBasePathNoTrailingPath returns the URL path to the BasePathNoTrailing service TrailingWithBasePathNoTrailing HTTP endpoint.
func TrailingWithBasePathNoTrailingBasePathNoTrailingPath() string {
	return "/foo/bar/"
}
`

var BasePathWithTrailingSlash_WithBasePathWithTrailingCode = `// SlashWithBasePathWithTrailingBasePathWithTrailingPath returns the URL path to the BasePathWithTrailing service SlashWithBasePathWithTrailing HTTP endpoint.
func SlashWithBasePathWithTrailingBasePathWithTrailingPath() string {
	return "/foo/"
}
`

var NoBasePath_SlashNoBasePathCode = `// SlashNoBasePathNoBasePathPath returns the URL path to the NoBasePath service SlashNoBasePath HTTP endpoint.
func SlashNoBasePathNoBasePathPath() string {
	return "/"
}
`

var NoBasePath_TrailingNoBasePathCode = `// TrailingNoBasePathNoBasePathPath returns the URL path to the NoBasePath service TrailingNoBasePath HTTP endpoint.
func TrailingNoBasePathNoBasePathPath() string {
	return "/foo/"
}
`

var BasePath_SpecialTrailingSlashCode = `// SpecialTrailingSlashBasePathPath returns the URL path to the BasePath service SpecialTrailingSlash HTTP endpoint.
func SpecialTrailingSlashBasePathPath() string {
	return "/foo/"
}
`

var PathOneParamCode = `// MethodPathOneParamServicePathOneParamPath returns the URL path to the ServicePathOneParam service MethodPathOneParam HTTP endpoint.
func MethodPathOneParamServicePathOneParamPath(a string) string {
	return fmt.Sprintf("/one/%v/two", a)
}
`

var PathMultipleParamsCode = `// MethodPathMultipleParamServicePathMultipleParamPath returns the URL path to the ServicePathMultipleParam service MethodPathMultipleParam HTTP endpoint.
func MethodPathMultipleParamServicePathMultipleParamPath(a string, b string) string {
	return fmt.Sprintf("/one/%v/two/%v/three", a, b)
}
`

var PathAlternativesCode = `// MethodPathAlternativesServicePathAlternativesPath returns the URL path to the ServicePathAlternatives service MethodPathAlternatives HTTP endpoint.
func MethodPathAlternativesServicePathAlternativesPath(a string, b string) string {
	return fmt.Sprintf("/one/%v/two/%v/three", a, b)
}

// MethodPathAlternativesServicePathAlternativesPath2 returns the URL path to the ServicePathAlternatives service MethodPathAlternatives HTTP endpoint.
func MethodPathAlternativesServicePathAlternativesPath2(b string, a string) string {
	return fmt.Sprintf("/one/two/%v/three/%v", b, a)
}
`

var PathStringSliceParamCode = `// MethodPathStringSliceParamServicePathStringSliceParamPath returns the URL path to the ServicePathStringSliceParam service MethodPathStringSliceParam HTTP endpoint.
func MethodPathStringSliceParamServicePathStringSliceParamPath(a []string) string {
	aSlice := make([]string, len(a))
	for i, v := range a {
		aSlice[i] = url.QueryEscape(v)
	}
	return fmt.Sprintf("/one/%v/two", strings.Join(aSlice, ","))
}
`

var PathIntSliceParamCode = `// MethodPathIntSliceParamServicePathIntSliceParamPath returns the URL path to the ServicePathIntSliceParam service MethodPathIntSliceParam HTTP endpoint.
func MethodPathIntSliceParamServicePathIntSliceParamPath(a []int) string {
	aSlice := make([]string, len(a))
	for i, v := range a {
		aSlice[i] = strconv.FormatInt(int64(v), 10)
	}
	return fmt.Sprintf("/one/%v/two", strings.Join(aSlice, ","))
}
`

var PathInt32SliceParamCode = `// MethodPathInt32SliceParamServicePathInt32SliceParamPath returns the URL path to the ServicePathInt32SliceParam service MethodPathInt32SliceParam HTTP endpoint.
func MethodPathInt32SliceParamServicePathInt32SliceParamPath(a []int32) string {
	aSlice := make([]string, len(a))
	for i, v := range a {
		aSlice[i] = strconv.FormatInt(int64(v), 10)
	}
	return fmt.Sprintf("/one/%v/two", strings.Join(aSlice, ","))
}
`

var PathInt64SliceParamCode = `// MethodPathInt64SliceParamServicePathInt64SliceParamPath returns the URL path to the ServicePathInt64SliceParam service MethodPathInt64SliceParam HTTP endpoint.
func MethodPathInt64SliceParamServicePathInt64SliceParamPath(a []int64) string {
	aSlice := make([]string, len(a))
	for i, v := range a {
		aSlice[i] = strconv.FormatInt(v, 10)
	}
	return fmt.Sprintf("/one/%v/two", strings.Join(aSlice, ","))
}
`

var PathUintSliceParamCode = `// MethodPathUintSliceParamServicePathUintSliceParamPath returns the URL path to the ServicePathUintSliceParam service MethodPathUintSliceParam HTTP endpoint.
func MethodPathUintSliceParamServicePathUintSliceParamPath(a []uint) string {
	aSlice := make([]string, len(a))
	for i, v := range a {
		aSlice[i] = strconv.FormatUint(uint64(v), 10)
	}
	return fmt.Sprintf("/one/%v/two", strings.Join(aSlice, ","))
}
`

var PathUint32SliceParamCode = `// MethodPathUint32SliceParamServicePathUint32SliceParamPath returns the URL path to the ServicePathUint32SliceParam service MethodPathUint32SliceParam HTTP endpoint.
func MethodPathUint32SliceParamServicePathUint32SliceParamPath(a []uint32) string {
	aSlice := make([]string, len(a))
	for i, v := range a {
		aSlice[i] = strconv.FormatUint(uint64(v), 10)
	}
	return fmt.Sprintf("/one/%v/two", strings.Join(aSlice, ","))
}
`

var PathUint64SliceParamCode = `// MethodPathUint64SliceParamServicePathUint64SliceParamPath returns the URL path to the ServicePathUint64SliceParam service MethodPathUint64SliceParam HTTP endpoint.
func MethodPathUint64SliceParamServicePathUint64SliceParamPath(a []uint64) string {
	aSlice := make([]string, len(a))
	for i, v := range a {
		aSlice[i] = strconv.FormatUint(v, 10)
	}
	return fmt.Sprintf("/one/%v/two", strings.Join(aSlice, ","))
}
`

var PathFloat32SliceParamCode = `// MethodPathFloat32SliceParamServicePathFloat32SliceParamPath returns the URL path to the ServicePathFloat32SliceParam service MethodPathFloat32SliceParam HTTP endpoint.
func MethodPathFloat32SliceParamServicePathFloat32SliceParamPath(a []float32) string {
	aSlice := make([]string, len(a))
	for i, v := range a {
		aSlice[i] = strconv.FormatFloat(float64(v), 'f', -1, 32)
	}
	return fmt.Sprintf("/one/%v/two", strings.Join(aSlice, ","))
}
`

var PathFloat64SliceParamCode = `// MethodPathFloat64SliceParamServicePathFloat64SliceParamPath returns the URL path to the ServicePathFloat64SliceParam service MethodPathFloat64SliceParam HTTP endpoint.
func MethodPathFloat64SliceParamServicePathFloat64SliceParamPath(a []float64) string {
	aSlice := make([]string, len(a))
	for i, v := range a {
		aSlice[i] = strconv.FormatFloat(v, 'f', -1, 64)
	}
	return fmt.Sprintf("/one/%v/two", strings.Join(aSlice, ","))
}
`

var PathBoolSliceParamCode = `// MethodPathBoolSliceParamServicePathBoolSliceParamPath returns the URL path to the ServicePathBoolSliceParam service MethodPathBoolSliceParam HTTP endpoint.
func MethodPathBoolSliceParamServicePathBoolSliceParamPath(a []bool) string {
	aSlice := make([]string, len(a))
	for i, v := range a {
		aSlice[i] = strconv.FormatBool(v)
	}
	return fmt.Sprintf("/one/%v/two", strings.Join(aSlice, ","))
}
`

var PathInterfaceSliceParamCode = `// MethodPathInterfaceSliceParamServicePathInterfaceSliceParamPath returns the URL path to the ServicePathInterfaceSliceParam service MethodPathInterfaceSliceParam HTTP endpoint.
func MethodPathInterfaceSliceParamServicePathInterfaceSliceParamPath(a []interface{}) string {
	aSlice := make([]string, len(a))
	for i, v := range a {
		aSlice[i] = url.QueryEscape(fmt.Sprintf("%v", v))
	}
	return fmt.Sprintf("/one/%v/two", strings.Join(aSlice, ","))
}
`
