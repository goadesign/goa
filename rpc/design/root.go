package design

var (
	// Root holds the root expression built on process initialization.
	Root = &RootExpr{}
)

// RootExpr is the data structure built by the http design DSL.
type RootExpr struct {
	// Services contains the services created by the DSL.
	Services []*ServiceExpr
}

// Service returns the service expression with the given name, nil if there isn't one.
func (r *RootExpr) Service(name string) *ServiceExpr {
	for _, s := range r.Services {
		if s.Name == name {
			return s
		}
	}
	return nil
}
