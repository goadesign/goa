package design

import (
	"fmt"
	"mime"
	"strings"

	"goa.design/goa.v2/eval"
)

const (
	// DefaultView is the name of the default media type view.
	DefaultView = "default"
)

type (
	// MediaTypeExpr describes the rendering of a resource using field and link
	// definitions. A field corresponds to a single member of the media type, it has a name and
	// a type as well as optional validation rules. A link has a name and a URL that points to a
	// related resource.  Media types also define views which describe which fields and links to
	// render when building the response body for the corresponding view.
	MediaTypeExpr struct {
		// A media type is a type
		*UserTypeExpr
		// Identifier is the RFC 6838 media type identifier.
		Identifier string
		// ContentType identifies the value written to the response "Content-Type" header.
		// Defaults to Identifier.
		ContentType string
		// Views list the supported views indexed by name.
		Views []*ViewExpr
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
		*AttributeExpr
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
		Links *UserTypeExpr
	}
)

var (
	// ErrorMediaIdentifier is the media type identifier used for error responses.
	ErrorMediaIdentifier = "application/vnd.goa.error"

	// ErrorMedia is the built-in media type for error responses.
	ErrorMedia = &MediaTypeExpr{
		UserTypeExpr: &UserTypeExpr{
			AttributeExpr: &AttributeExpr{
				Type:        errorMediaType,
				Description: "Error response media type",
				UserExample: map[string]interface{}{
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
		Views:      []*ViewExpr{errorMediaView},
	}

	errorMediaType = Object{
		"id": &AttributeExpr{
			Type:        String,
			Description: "a unique identifier for this particular occurrence of the problem.",
			UserExample: "3F1FKVRR",
		},
		"status": &AttributeExpr{
			Type:        String,
			Description: "the HTTP status code applicable to this problem, expressed as a string value.",
			UserExample: "400",
		},
		"code": &AttributeExpr{
			Type:        String,
			Description: "an application-specific error code, expressed as a string value.",
			UserExample: "invalid_value",
		},
		"detail": &AttributeExpr{
			Type:        String,
			Description: "a human-readable explanation specific to this occurrence of the problem.",
			UserExample: "Value of ID must be an integer",
		},
		"meta": &AttributeExpr{
			Type: &Map{
				KeyType:  &AttributeExpr{Type: String},
				ElemType: &AttributeExpr{Type: Any},
			},
			Description: "a meta object containing non-standard meta-information about the error.",
			UserExample: map[string]interface{}{"timestamp": 1458609066},
		},
	}

	errorMediaView = &ViewExpr{
		AttributeExpr: &AttributeExpr{Type: errorMediaType},
		Name:          "default",
	}
)

// NewMediaTypeExpr creates a media type definition but does not
// execute the DSL.
func NewMediaTypeExpr(name, identifier string, fn func()) *MediaTypeExpr {
	return &MediaTypeExpr{
		UserTypeExpr: &UserTypeExpr{
			AttributeExpr: &AttributeExpr{Type: Object{}, DSLFunc: fn},
			TypeName:      name,
		},
		Identifier: identifier,
	}
}

// CanonicalIdentifier returns the media type identifier sans suffix
// which is what the DSL uses to store and lookup media types.
func CanonicalIdentifier(identifier string) string {
	base, params, err := mime.ParseMediaType(identifier)
	if err != nil {
		return identifier
	}
	id := base
	if i := strings.Index(id, "+"); i != -1 {
		id = id[:i]
	}
	return mime.FormatMediaType(id, params)
}

// Kind implements DataKind.
func (m *MediaTypeExpr) Kind() Kind { return MediaTypeKind }

// Dup creates a deep copy of the media type given a deep copy of its attribute.
func (m *MediaTypeExpr) Dup(att *AttributeExpr) UserType {
	return &MediaTypeExpr{
		UserTypeExpr: m.UserTypeExpr.Dup(att).(*UserTypeExpr),
		Identifier:   m.Identifier,
		Views:        m.Views,
	}
}

// Name returns the media type canonical identifier.
func (m *MediaTypeExpr) Name() string { return CanonicalIdentifier(m.Identifier) }

// View returns the view with the given name.
func (m *MediaTypeExpr) View(name string) *ViewExpr {
	for _, v := range m.Views {
		if v.Name == name {
			return v
		}
	}
	return nil
}

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
func (m *MediaTypeExpr) ComputeViews() []*ViewExpr {
	if m.Views != nil {
		return m.Views
	}
	if a, ok := m.Type.(*Array); ok {
		if mt, ok := a.ElemType.Type.(*MediaTypeExpr); ok {
			return mt.ComputeViews()
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
func (p *Projector) Project(m *MediaTypeExpr, view string) (*ProjectedMTExpr, error) {
	var viewID string
	cano := CanonicalIdentifier(m.Identifier)
	base, params, _ := mime.ParseMediaType(cano)
	if params["view"] != "" {
		viewID = cano // Already projected
	} else {
		params["view"] = view
		viewID = mime.FormatMediaType(base, params)
	}
	if proj, ok := p.Projected[viewID]; ok {
		return proj, nil
	}
	if _, ok := m.Type.(*Array); ok {
		return p.projectCollection(m, view, viewID)
	}
	return p.projectSingle(m, view, viewID)
}

func (p *Projector) projectSingle(m *MediaTypeExpr, view, viewID string) (*ProjectedMTExpr, error) {
	v := m.View(view)
	if v == nil {
		return nil, fmt.Errorf("unknown view %#v", view)
	}
	viewObj := v.Type.(Object)

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

	projected := &MediaTypeExpr{
		Identifier: viewID,
		UserTypeExpr: &UserTypeExpr{
			TypeName: typeName,
			AttributeExpr: &AttributeExpr{
				Description: desc,
				Type:        Dup(v.Type),
				Validation:  val,
			},
		},
	}
	projected.Views = []*ViewExpr{&ViewExpr{
		Name:          "default",
		AttributeExpr: DupAtt(v.AttributeExpr),
		Parent:        projected,
	}}

	proj := ProjectedMTExpr{View: view, MediaType: projected}
	p.Projected[viewID] = &proj
	projectedObj := projected.Type.(Object)
	mtObj := m.Type.(Object)
	for n := range viewObj {
		if at := mtObj[n]; at != nil {
			at = DupAtt(at)
			if mt, ok := at.Type.(*MediaTypeExpr); ok {
				vatt := viewObj[n]
				var view string
				if len(vatt.Metadata["view"]) > 0 {
					view = vatt.Metadata["view"][0]
				}
				if view == "" && len(at.Metadata["view"]) > 0 {
					view = at.Metadata["view"][0]
				}
				if view == "" {
					view = DefaultView
				}
				pr, err := p.Project(mt, view)
				if err != nil {
					return nil, fmt.Errorf("view %#v on field %#v cannot be computed: %s", view, n, err)
				}
				at.Type = pr.MediaType
			}
			projectedObj[n] = at
		}
	}
	return &proj, nil
}

func (p *Projector) projectCollection(m *MediaTypeExpr, view, viewID string) (*ProjectedMTExpr, error) {
	// Project the collection element media type
	e := m.Type.(*Array).ElemType.Type.(*MediaTypeExpr) // validation checked this cast would work
	pe, err2 := p.Project(e, view)
	if err2 != nil {
		return nil, fmt.Errorf("collection element: %s", err2)
	}

	// Build the projected collection with the results
	desc := m.TypeName + " is the media type for an array of " + e.TypeName + " (" + view + " view)"
	proj := &MediaTypeExpr{
		Identifier: viewID,
		UserTypeExpr: &UserTypeExpr{
			AttributeExpr: &AttributeExpr{
				Description: desc,
				Type:        &Array{ElemType: &AttributeExpr{Type: pe.MediaType}},
				UserExample: m.UserExample,
			},
			TypeName: pe.MediaType.TypeName + "Collection",
		},
	}
	proj.Views = []*ViewExpr{&ViewExpr{
		AttributeExpr: DupAtt(pe.MediaType.View("default").AttributeExpr),
		Name:          "default",
		Parent:        pe.MediaType,
	}}

	// Run the DSL that was created by the CollectionOf function
	if !eval.Execute(proj.DSL(), proj) {
		return nil, eval.Context.Errors
	}

	// Build the links user type
	var links *UserTypeExpr
	if pe.Links != nil {
		lTypeName := pe.Links.TypeName + "Array"
		links = &UserTypeExpr{
			AttributeExpr: &AttributeExpr{
				Type:        &Array{ElemType: &AttributeExpr{Type: pe.Links}},
				Description: fmt.Sprintf("%s contains links to related resources of %s.", lTypeName, m.TypeName),
			},
			TypeName: lTypeName,
		}
	}

	return &ProjectedMTExpr{view, proj, links}, nil
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

// EvalName returns the generic definition name used in error messages.
func (l *LinkExpr) EvalName() string {
	var prefix, suffix string
	if l.Name != "" {
		prefix = fmt.Sprintf("link %#v", l.Name)
	} else {
		prefix = "unnamed link"
	}
	if l.Parent != nil {
		suffix = fmt.Sprintf(" of %s", l.Parent.EvalName())
	}
	return prefix + suffix
}

// Attribute returns the linked attribute.
func (l *LinkExpr) Attribute() *AttributeExpr {
	p := l.Parent.Type.(Object)
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

// EvalName returns the generic definition name used in error messages.
func (v *ViewExpr) EvalName() string {
	var prefix, suffix string
	if v.Name != "" {
		prefix = fmt.Sprintf("view %#v", v.Name)
	} else {
		prefix = "unnamed view"
	}
	if v.Parent != nil {
		suffix = fmt.Sprintf(" of %s", v.Parent.EvalName())
	}
	return prefix + suffix
}
