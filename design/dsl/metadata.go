package dsl

// Metadata is a key/value pair that can be assigned
// to an object.  The value is expected be a JSON string, but is
// not currently validated as such.
// Metadata is not used in standard generation but may be
// used by user-defined generators.
// Usage:
//	 Metadata("creator", `{"name":"goagen"}`)
func Metadata(name string, value string) {
	if at, ok := attributeDefinition(false); ok {
		if at.Metadata == nil {
			at.Metadata = make(map[string]string)
		}
		at.Metadata[name] = value
		return
	}
	if mt, ok := mediaTypeDefinition(false); ok {
		if mt.Metadata == nil {
			mt.Metadata = make(map[string]string)
		}
		mt.Metadata[name] = value
		return
	}
	if act, ok := actionDefinition(false); ok {
		if act.Metadata == nil {
			act.Metadata = make(map[string]string)
		}
		act.Metadata[name] = value
		return
	}
	if res, ok := resourceDefinition(false); ok {
		if res.Metadata == nil {
			res.Metadata = make(map[string]string)
		}
		res.Metadata[name] = value
		return
	}
	if rd, ok := responseDefinition(false); ok {
		if rd.Metadata == nil {
			rd.Metadata = make(map[string]string)
		}
		rd.Metadata[name] = value
		return
	}
	if api, ok := apiDefinition(true); ok {
		if api.Metadata == nil {
			api.Metadata = make(map[string]string)
		}
		api.Metadata[name] = value
		return
	}
}
