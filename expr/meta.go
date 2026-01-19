package expr

// MetaAdder is implemented by expressions that can receive design-time metadata.
//
// Goa's standard DSL helper `Meta(name, value...)` attaches metadata to the
// current expression. Most Goa core expressions store metadata directly in a
// `MetaExpr` field. Third-party DSLs (and plugins) may add new expression types
// that also want to support `Meta` without Goa needing to know about those
// concrete types.
//
// An expression that implements MetaAdder opts into Goa's `Meta` DSL by
// providing an AddMeta implementation that:
//   - appends values to the existing entry for the given name, and
//   - preserves existing metadata for other names.
//
// Callers should treat metadata keys and values as opaque strings; their
// interpretation is application-specific.
type MetaAdder interface {
	// AddMeta appends the given values to the metadata entry identified by name.
	AddMeta(name string, value ...string)
}

// MetaDeleter is implemented by expressions that can remove design-time metadata.
//
// Goa's standard DSL helper `RemoveMeta(name)` removes the metadata entry with
// the given name from the current expression.
//
// Implementations should be a no-op when no metadata is present.
type MetaDeleter interface {
	// DeleteMeta removes the metadata entry identified by name.
	DeleteMeta(name string)
}
