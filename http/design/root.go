package design

// Root holds the root expression built on process initialization.
var Root = &RootExp{}

// RootExpr is the data structure built by the http design DSL.
type RootExp struct {
	// MediaTypes contains the set of media types created by the DSL.
	MediaTypes []*MediaTypeExpr
	// Resources contains the resources created by the DSL.
	Resources []*ResourceExpr
}

// MediaType returns the media type expression with the given id and true if found, nil and false
// otherwise.
func (r *RootExpr) MediaType(id string) (*MediaTypeExpr, bool) {
	for _, mt := range r.MediaTypes {
		if mt.ID == id {
			return mt, true
		}
	}
	return nil, false
}
