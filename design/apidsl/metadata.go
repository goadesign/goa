package apidsl

// Metadata is a set of key/value pairs that can be assigned
// to an object. Each value consists of a slice of strings so
// that multiple invocation of the Metadata function on the
// same target using the same key builds up the slice.
//
// While keys can have any value the following names are
// handled explicitly by goagen:
//
// "struct:tag=xxx": sets the struct field tag xxx on generated structs.
//               Overrides tags that goagen would otherwise set.
//               If the metadata value is a slice then the
//               strings are joined with the space character as
//               separator.
//
// "swagger:tag=xxx": sets the Swagger object field tag xxx. The value
//               must be one to three strings. The first string is
//               the tag description while the second and third strings
//               are the documentation url and description for the tag.
//               Subsequent calls to Metadata on the same attribute
//               with key "swagger:tag" builds up the Swagger tag list.
//
// Usage:
//        Metadata("struct:tag=json", "myName,omitempty")
//        Metadata("struct:tag=xml", "myName,attr")
//        Metadata("swagger:tag=backend")
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
