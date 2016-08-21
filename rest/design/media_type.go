package design

import (
	"fmt"
	"mime"
	"sort"
	"strings"

	goa "github.com/goadesign/goa/design"
)

const (
	// MediaTypeKind represents a media type.
	MediaTypeKind = goa.AnyKind + 1
)

type (
	// MediaTypeExpr describes the rendering of a resource using field and link
	// definitions. A field corresponds to a single member of the media type, it has a name and
	// a type as well as optional validation rules. A link has a name and a URL that points to a
	// related resource.  Media types also define views which describe which fields and links to
	// render when building the response body for the corresponding view.
	MediaTypeExpr struct {
		// A media type is a type
		*goa.UserTypeExpr
		// Identifier is the RFC 6838 media type identifier.
		Identifier string
		// ContentType identifies the value written to the response "Content-Type" header.
		// Defaults to Identifier.
		ContentType string
		// Links list the rendered links indexed by name.
		Links map[string]*LinkExpr
		// Views list the supported views indexed by name.
		Views map[string]*ViewExpr
	}

	// LinkExpr defines a media type link, it specifies a URL to a related resource.
	LinkExpr struct {
		// Link name
		Name string
		// View used to render link if not "link"
		View string
		// URITemplate is the RFC6570 URI template of the link Href.
		URITemplate string

		// Parent media Type
		Parent *MediaTypeExpr
	}

	// ViewExpr defines which fields and links to render when building a response.  The
	// view is an object whose field names must match the names of the parent media type field
	// names.  The field definitions are inherited from the parent media type but may be
	// overridden.
	ViewExpr struct {
		// Set of properties included in view
		*FieldExpr
		// Name of view
		Name string
		// Parent media Type
		Parent *MediaTypeExpr
	}

	// Projector derives media types using a canonical media type expression and a view
	// expression.
	Projector struct {
		// Projected is a cache of projected media types indexed by identifier.
		Projected map[string]*ProjectedMTExpr
	}

	// ProjectedMTExpr represents a media type that was derived from a media type
	// expression defined in a DSL by applying a view. The result of applying a view is a media
	// type expression that contains the subset of the original media type fields listed in the
	// view recursively projected if they are media types themselves. The result also include a
	// links object that corresponds to the media type links defined in the design.
	ProjectedMTExpr struct {
		// View used to create projected media type
		View string
		// MediaType is the projected media type.
		MediaType *MediaTypeExpr
		// Links lists the media type links if any.
		Links Object
	}
)

var (
	// ErrorMediaIdentifier is the media type identifier used for error responses.
	ErrorMediaIdentifier = "application/vnd.goa.error"

	// ErrorMedia is the built-in media type for error responses.
	ErrorMedia = &MediaTypeExpr{
		UserTypeExpr: &UserTypeExpr{
			FieldExpr: &FieldExpr{
				Type:        errorMediaType,
				Description: "Error response media type",
				Example: map[string]interface{}{
					"id":     "3F1FKVRR",
					"status": "400",
					"code":   "invalid_value",
					"detail": "Value of ID must be an integer",
					"meta":   map[string]interface{}{"timestamp": 1458609066},
				},
			},
			TypeName: "error",
		},
		Identifier: ErrorMediaIdentifier,
		Views:      map[string]*ViewExpr{"default": errorMediaView},
	}

	errorMediaType = Object{
		"id": &FieldExpr{
			Type:        String,
			Description: "a unique identifier for this particular occurrence of the problem.",
			Example:     "3F1FKVRR",
		},
		"status": &FieldExpr{
			Type:        String,
			Description: "the HTTP status code applicable to this problem, expressed as a string value.",
			Example:     "400",
		},
		"code": &FieldExpr{
			Type:        String,
			Description: "an application-specific error code, expressed as a string value.",
			Example:     "invalid_value",
		},
		"detail": &FieldExpr{
			Type:        String,
			Description: "a human-readable explanation specific to this occurrence of the problem.",
			Example:     "Value of ID must be an integer",
		},
		"meta": &FieldExpr{
			Type: &Map{
				KeyType:  &FieldExpr{Type: String},
				ElemType: &FieldExpr{Type: Any},
			},
			Description: "a meta object containing non-standard meta-information about the error.",
			Example:     map[string]interface{}{"timestamp": 1458609066},
		},
	}

	errorMediaView = &ViewExpr{
		FieldExpr: &FieldExpr{Type: errorMediaType},
		Name:      "default",
	}
)

// NewMediaTypeExpr creates a media type definition but does not
// execute the DSL.
func NewMediaTypeExpr(name, identifier string, dsl func()) *MediaTypeExpr {
	return &MediaTypeExpr{
		UserTypeExpr: &UserTypeExpr{
			FieldExpr: &FieldExpr{Type: Object{}, DSLFunc: dsl},
			TypeName:  name,
		},
		Identifier: identifier,
	}
}

// Kind implements DataKind.
func (m *MediaTypeExpr) Kind() Kind { return MediaTypeKind }

// IsError returns true if the media type is implemented via a goa struct.
func (m *MediaTypeExpr) IsError() bool {
	base, params, err := mime.ParseMediaType(m.Identifier)
	if err != nil {
		panic("invalid media type identifier " + m.Identifier) // bug
	}
	delete(params, "view")
	return mime.FormatMediaType(base, params) == ErrorMedia.Identifier
}

// ComputeViews returns the media type views recursing as necessary if the media type is a
// collection.
func (m *MediaTypeExpr) ComputeViews() map[string]*ViewExpr {
	if m.Views != nil {
		return m.Views
	}
	if m.IsArray() {
		if mt, ok := m.ToArray().ElemType.Type.(*MediaTypeExpr); ok {
			return mt.ComputeViews()
		}
	}
	return nil
}

// ViewIterator is the type of the function given to IterateViews.
type ViewIterator func(*ViewExpr) error

// IterateViews calls the given iterator passing in each field sorted in alphabetical order.
// Iteration stops if an iterator returns an error and in this case IterateViews returns that
// error.
func (m *MediaTypeExpr) IterateViews(it ViewIterator) error {
	o := m.Views
	// gather names and sort them
	names := make([]string, len(o))
	i := 0
	for n := range o {
		names[i] = n
		i++
	}
	sort.Strings(names)
	// iterate
	for _, n := range names {
		if err := it(o[n]); err != nil {
			return err
		}
	}
	return nil
}

// Project creates a MediaTypeExpr containing the fields defined in the view expression of m named
// after the view argument. Project also returns a links object created after the link expression of
// m if there is one.
//
// The resulting media type defines a default view. The media type identifier is computed by adding
// a parameter called "view" to the original identifier. The value of the "view" parameter is the
// name of the view.
func (p *Projector) Project(m *MediaTypeExpr, view string) (*MediaTypeExpr, Object, error) {
	var viewID string
	cano := CanonicalIdentifier(m.Identifier)
	base, params, _ := mime.ParseMediaType(cano)
	if params["view"] != "" {
		viewID = cano // Already projected
	} else {
		params["view"] = view
		viewID = mime.FormatMediaType(base, params)
	}

	var ok bool
	if proj, ok = p.Projected[viewID]; ok {
		return proj.MediaType, proj.Links, nil
	}
	if m.IsArray() {
		proj, err := p.projectCollection(m, view, viewID)
		return proj, nil, err
	}
	return p.projectSingle(m, view, viewID)
}

func (m *MediaTypeExpr) projectSingle(view, viewID string) (p *MediaTypeExpr, links Object, err error) {
	v, ok := m.Views[view]
	if !ok {
		return nil, nil, fmt.Errorf("unknown view %#v", view)
	}
	viewObj := v.Type.ToObject()

	// Compute validations - view may not have all fields
	var val *ValidationExpr
	if m.Validation != nil {
		var required []string
		for _, n := range m.Validation.Required {
			if _, ok := viewObj[n]; ok {
				required = append(required, n)
			}
		}
		val = m.Validation.Dup()
		val.Required = required
	}

	// Compute description
	desc := m.Description
	if desc == "" {
		desc = m.TypeName + " media type"
	}
	desc += " (" + view + " view)"

	// Compute type name
	typeName := m.TypeName
	if view != "default" {
		typeName += strings.Title(view)
	}

	p = &MediaTypeExpr{
		Identifier: viewID,
		UserTypeExpr: &UserTypeExpr{
			TypeName: typeName,
			FieldExpr: &FieldExpr{
				Description: desc,
				Type:        Dup(v.Type),
				Validation:  val,
			},
		},
	}
	p.Views = map[string]*ViewExpr{"default": {
		Name:      "default",
		FieldExpr: DupAtt(v.FieldExpr),
		Parent:    p,
	}}

	proj := ProjectedMTExpr{View: view, MediaType: p}
	ProjectedMediaTypes[viewID] = proj
	projectedObj := p.Type.ToObject()
	mtObj := m.Type.ToObject()
	for n := range viewObj {
		if n == "links" {
			linkObj := make(Object)
			for n, link := range m.Links {
				linkView := link.View
				if linkView == "" {
					linkView = "link"
				}
				mtAtt, ok := mtObj[n]
				if !ok {
					return nil, nil, fmt.Errorf("unknown field %#v used in links", n)
				}
				mtt := mtAtt.Type.(*MediaTypeExpr)
				vl, _, err := projector.Project(mtt, linkView)
				if err != nil {
					return nil, nil, err
				}
				linkObj[n] = &FieldExpr{Type: vl, Validation: mtt.Validation, Metadata: mtAtt.Metadata}
			}
			proj.Links = linkObj
		} else {
			if at := mtObj[n]; at != nil {
				at = DupAtt(at)
				if mt, ok := at.Type.(*MediaTypeExpr); ok {
					vatt := viewObj[n]
					view := vatt.View
					if view == "" {
						view = at.View
					}
					if view == "" {
						view = DefaultView
					}
					pr, _, err := projector.Project(mt, view)
					if err != nil {
						return nil, nil, fmt.Errorf("view %#v on field %#v cannot be computed: %s", view, n, err)
					}
					at.Type = pr
				}
				projectedObj[n] = at
			}
		}
	}
	return
}

func (p *Projector) projectCollection(m *MediaTypeExpr, view, viewID string) (*MediaTypeExpr, error) {
	// Project the collection element media type
	e := m.ToArray().ElemType.Type.(*MediaTypeExpr) // validation checked this cast would work
	pe, le, err2 := p.Project(e, view)
	if err2 != nil {
		return nil, nil, fmt.Errorf("collection element: %s", err2)
	}

	// Build the projected collection with the results
	desc := m.TypeName + " is the media type for an array of " + e.TypeName + " (" + view + " view)"
	proj := &MediaTypeExpr{
		Identifier: viewID,
		UserTypeExpr: &UserTypeExpr{
			FieldExpr: &FieldExpr{
				Description: desc,
				Type:        &Array{ElemType: &FieldExpr{Type: pe}},
				Example:     m.Example,
			},
			TypeName: pe.TypeName + "Collection",
		},
	}
	proj.Views = map[string]*ViewExpr{"default": &ViewExpr{
		FieldExpr: DupAtt(pe.Views["default"].FieldExpr),
		Name:      "default",
		Parent:    p,
	}}

	// Run the DSL that was created by the CollectionOf function
	if !eval.Execute(proj.DSL(), proj) {
		return nil, nil, eval.Errors
	}

	return proj, nil
}

// projectIdentifier computes the projected media type identifier by adding the "view" param.  We
// need the projected media type identifier to be different so that looking up projected media types
// from ProjectedMediaTypes works correctly. It's also good for clients.
func (m *MediaTypeExpr) projectIdentifier(view string) string {
	base, params, err := mime.ParseMediaType(m.Identifier)
	if err != nil {
		base = m.Identifier
	}
	params["view"] = view
	return mime.FormatMediaType(base, params)
}

// Context returns the generic definition name used in error messages.
func (l *LinkExpr) Context() string {
	var prefix, suffix string
	if l.Name != "" {
		prefix = fmt.Sprintf("link %#v", l.Name)
	} else {
		prefix = "unnamed link"
	}
	if l.Parent != nil {
		suffix = fmt.Sprintf(" of %s", l.Parent.Context())
	}
	return prefix + suffix
}

// Attribute returns the linked attribute.
func (l *LinkExpr) Attribute() *FieldExpr {
	p := l.Parent.ToObject()
	if p == nil {
		return nil
	}
	att, _ := p[l.Name]

	return att
}

// MediaType returns the media type of the linked attribute.
func (l *LinkExpr) MediaType() *MediaTypeExpr {
	att := l.Attribute()
	mt, _ := att.Type.(*MediaTypeExpr)
	return mt
}

// Context returns the generic definition name used in error messages.
func (v *ViewExpr) Context() string {
	var prefix, suffix string
	if v.Name != "" {
		prefix = fmt.Sprintf("view %#v", v.Name)
	} else {
		prefix = "unnamed view"
	}
	if v.Parent != nil {
		suffix = fmt.Sprintf(" of %s", v.Parent.Context())
	}
	return prefix + suffix
}
