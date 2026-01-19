// This file wires Goa's core expressions into the Meta DSL extension points.
//
// The `dsl.Meta` and `dsl.RemoveMeta` helpers operate on the expression returned
// by `eval.Current()`. Most expressions that accept metadata store it in a
// `MetaExpr` field. By implementing the `MetaAdder` / `MetaDeleter` interfaces
// on those expressions, the DSL implementation stays small and does not need to
// enumerate all supported expression types.
package expr

func appendMeta(meta MetaExpr, name string, vals ...string) MetaExpr {
	if meta == nil {
		meta = make(MetaExpr)
	}
	meta[name] = append(meta[name], vals...)
	return meta
}

// AddMeta appends metadata values to the API expression.
func (a *APIExpr) AddMeta(name string, value ...string) {
	a.Meta = appendMeta(a.Meta, name, value...)
}

// DeleteMeta removes a metadata entry from the API expression.
func (a *APIExpr) DeleteMeta(name string) {
	delete(a.Meta, name)
}

// AddMeta appends metadata values to the server expression.
func (s *ServerExpr) AddMeta(name string, value ...string) {
	s.Meta = appendMeta(s.Meta, name, value...)
}

// DeleteMeta removes a metadata entry from the server expression.
func (s *ServerExpr) DeleteMeta(name string) {
	delete(s.Meta, name)
}

// AddMeta appends metadata values to the host expression.
func (h *HostExpr) AddMeta(name string, value ...string) {
	h.Meta = appendMeta(h.Meta, name, value...)
}

// DeleteMeta removes a metadata entry from the host expression.
func (h *HostExpr) DeleteMeta(name string) {
	delete(h.Meta, name)
}

// AddMeta appends metadata values to the service expression.
func (s *ServiceExpr) AddMeta(name string, value ...string) {
	s.Meta = appendMeta(s.Meta, name, value...)
}

// DeleteMeta removes a metadata entry from the service expression.
func (s *ServiceExpr) DeleteMeta(name string) {
	delete(s.Meta, name)
}

// AddMeta appends metadata values to the method expression.
func (m *MethodExpr) AddMeta(name string, value ...string) {
	m.Meta = appendMeta(m.Meta, name, value...)
}

// DeleteMeta removes a metadata entry from the method expression.
func (m *MethodExpr) DeleteMeta(name string) {
	delete(m.Meta, name)
}

// AddMeta appends metadata values to the HTTP service expression.
func (s *HTTPServiceExpr) AddMeta(name string, value ...string) {
	s.Meta = appendMeta(s.Meta, name, value...)
}

// DeleteMeta removes a metadata entry from the HTTP service expression.
func (s *HTTPServiceExpr) DeleteMeta(name string) {
	delete(s.Meta, name)
}

// AddMeta appends metadata values to the HTTP endpoint expression.
func (e *HTTPEndpointExpr) AddMeta(name string, value ...string) {
	e.Meta = appendMeta(e.Meta, name, value...)
}

// DeleteMeta removes a metadata entry from the HTTP endpoint expression.
func (e *HTTPEndpointExpr) DeleteMeta(name string) {
	delete(e.Meta, name)
}

// AddMeta appends metadata values to the route expression.
func (r *RouteExpr) AddMeta(name string, value ...string) {
	r.Meta = appendMeta(r.Meta, name, value...)
}

// DeleteMeta removes a metadata entry from the route expression.
func (r *RouteExpr) DeleteMeta(name string) {
	delete(r.Meta, name)
}

// AddMeta appends metadata values to the HTTP file server expression.
func (f *HTTPFileServerExpr) AddMeta(name string, value ...string) {
	f.Meta = appendMeta(f.Meta, name, value...)
}

// DeleteMeta removes a metadata entry from the HTTP file server expression.
func (f *HTTPFileServerExpr) DeleteMeta(name string) {
	delete(f.Meta, name)
}

// AddMeta appends metadata values to the HTTP response expression.
func (r *HTTPResponseExpr) AddMeta(name string, value ...string) {
	r.Meta = appendMeta(r.Meta, name, value...)
}

// DeleteMeta removes a metadata entry from the HTTP response expression.
func (r *HTTPResponseExpr) DeleteMeta(name string) {
	delete(r.Meta, name)
}
