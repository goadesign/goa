package apidsl

// Metadata is a set of key/value pairs that can be assigned
// to an object. Each value consists of a slice of strings so
// that multiple invocation of the Metadata function on the
// same target using the same key builds up the slice.
//
// While keys can have any value the following names are
// handled explicitly by goagen:
//
// "struct:tag:xxx": sets the struct field tag xxx on generated structs.
//               Overrides tags that goagen would otherwise set.
//               If the metadata value is a slice then the
//               strings are joined with the space character as
//               separator.
//    Example:
//        Metadata("struct:tag:json", "myName,omitempty")
//        Metadata("struct:tag:xml", "myName,attr")
//
// "swagger:tag:xxx": sets the Swagger object field tag xxx.
//
//    Example:
//        Metadata("swagger:tag:Backend")
//        Metadata("swagger:tag:Backend:desc", "Quick description of what 'Backend' is")
//        Metadata("swagger:tag:Backend:url", "http://example.com")
//        Metadata("swagger:tag:Backend:url:desc", "See more docs here")
//
func Metadata(name string, value ...string) {
	if at, ok := attributeDefinition(false); ok {
		if at.Metadata == nil {
			at.Metadata = make(map[string][]string)
		}
		at.Metadata[name] = append(at.Metadata[name], value...)
		return
	}
	if mt, ok := mediaTypeDefinition(false); ok {
		if mt.Metadata == nil {
			mt.Metadata = make(map[string][]string)
		}
		mt.Metadata[name] = append(mt.Metadata[name], value...)
		return
	}
	if act, ok := actionDefinition(false); ok {
		if act.Metadata == nil {
			act.Metadata = make(map[string][]string)
		}
		act.Metadata[name] = append(act.Metadata[name], value...)
		return
	}
	if res, ok := resourceDefinition(false); ok {
		if res.Metadata == nil {
			res.Metadata = make(map[string][]string)
		}
		res.Metadata[name] = append(res.Metadata[name], value...)
		return
	}
	if rd, ok := responseDefinition(false); ok {
		if rd.Metadata == nil {
			rd.Metadata = make(map[string][]string)
		}
		rd.Metadata[name] = append(rd.Metadata[name], value...)
		return
	}
	if api, ok := apiDefinition(true); ok {
		if api.Metadata == nil {
			api.Metadata = make(map[string][]string)
		}
		api.Metadata[name] = append(api.Metadata[name], value...)
		return
	}
}
