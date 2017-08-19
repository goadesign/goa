package design

import (
	"fmt"
	"mime"
	"strings"

	"goa.design/goa.v2/eval"
)

const (
	// DefaultView is the name of the default result type view.
	DefaultView = "default"
)

type (
	// ResultTypeExpr describes the rendering of a service using field and
	// link definitions. A field corresponds to a single member of the result
	// type, it has a name and a type as well as optional validation rules.
	// A link has a name and a URL that points to a related service. Result
	// types also define views which describe which fields and links to
	// render when building the response body for the corresponding view.
	ResultTypeExpr struct {
		// A result type is a type
		*UserTypeExpr
		// Identifier is the RFC 6838 result type media type identifier.
		Identifier string
		// ContentType identifies the value written to the response
		// "Content-Type" header. Defaults to Identifier.
		ContentType string
		// Views list the supported views indexed by name.
		Views []*ViewExpr
	}

	// ViewExpr defines which fields and links to render when building a
	// response. The view is an object whose field names must match the
	// names of the parent result type field names. The field definitions are
	// inherited from the parent result type but may be overridden.
	ViewExpr struct {
		// Set of properties included in view
		*AttributeExpr
		// Name of view
		Name string
		// Parent result Type
		Parent *ResultTypeExpr
	}
)

var (
	// ErrorResultIdentifier is the result type identifier used for error
	// responses.
	ErrorResultIdentifier = "application/vnd.goa.error"

	// ErrorResult is the built-in result type for error responses.
	ErrorResult = &ResultTypeExpr{
		UserTypeExpr: &UserTypeExpr{
			AttributeExpr: &AttributeExpr{
				Type:        errorResultType,
				Description: "Error response result type",
				UserExamples: []*ExampleExpr{{
					Summary: "BadRequest",
					Value: Val{
						"id":     "3F1FKVRR",
						"status": "400",
						"code":   "invalid_value",
						"detail": "Value of ID must be an integer",
						"meta":   map[string]interface{}{"timestamp": 1458609066},
					},
				}},
			},
			TypeName: "error",
		},
		Identifier: ErrorResultIdentifier,
		Views:      []*ViewExpr{errorResultView},
	}

	errorResultType = &Object{
		{"id", &AttributeExpr{
			Type:        String,
			Description: "a unique identifier for this particular occurrence of the problem.",
		}},
		{"status", &AttributeExpr{
			Type:        String,
			Description: "the HTTP status code applicable to this problem, expressed as a string value.",
		}},
		{"code", &AttributeExpr{
			Type:        String,
			Description: "an application-specific error code, expressed as a string value.",
		}},
		{"detail", &AttributeExpr{
			Type:        String,
			Description: "a human-readable explanation specific to this occurrence of the problem.",
		}},
		{"meta", &AttributeExpr{
			Type: &Map{
				KeyType:  &AttributeExpr{Type: String},
				ElemType: &AttributeExpr{Type: Any},
			},
			Description: "a meta object containing non-standard meta-information about the error.",
		}},
	}

	errorResultView = &ViewExpr{
		AttributeExpr: &AttributeExpr{Type: errorResultType},
		Name:          "default",
	}
)

// NewResultTypeExpr creates a result type definition but does not
// execute the DSL.
func NewResultTypeExpr(name, identifier string, fn func()) *ResultTypeExpr {
	return &ResultTypeExpr{
		UserTypeExpr: &UserTypeExpr{
			AttributeExpr: &AttributeExpr{Type: &Object{}, DSLFunc: fn},
			TypeName:      name,
		},
		Identifier: identifier,
	}
}

// CanonicalIdentifier returns the result type identifier sans suffix
// which is what the DSL uses to store and lookup result types.
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
func (m *ResultTypeExpr) Kind() Kind { return ResultTypeKind }

// Dup creates a deep copy of the result type given a deep copy of its attribute.
func (m *ResultTypeExpr) Dup(att *AttributeExpr) UserType {
	return &ResultTypeExpr{
		UserTypeExpr: m.UserTypeExpr.Dup(att).(*UserTypeExpr),
		Identifier:   m.Identifier,
		Views:        m.Views,
	}
}

// Name returns the result type name.
func (m *ResultTypeExpr) Name() string { return m.TypeName }

// View returns the view with the given name.
func (m *ResultTypeExpr) View(name string) *ViewExpr {
	for _, v := range m.Views {
		if v.Name == name {
			return v
		}
	}
	return nil
}

// IsError returns true if the result type is implemented via a goa struct.
func (m *ResultTypeExpr) IsError() bool {
	base, params, err := mime.ParseMediaType(m.Identifier)
	if err != nil {
		panic("invalid result type identifier " + m.Identifier) // bug
	}
	delete(params, "view")
	return mime.FormatMediaType(base, params) == ErrorResult.Identifier
}

// ComputeViews returns the result type views recursing as necessary if the result
// type is a collection.
func (m *ResultTypeExpr) ComputeViews() []*ViewExpr {
	if m.Views != nil {
		return m.Views
	}
	if a, ok := m.Type.(*Array); ok {
		if mt, ok := a.ElemType.Type.(*ResultTypeExpr); ok {
			return mt.ComputeViews()
		}
	}
	return nil
}

// Finalize builds the default view if not explicitly defined.
func (m *ResultTypeExpr) Finalize() {
	if m.View("default") == nil {
		v := &ViewExpr{
			AttributeExpr: DupAtt(m.AttributeExpr),
			Name:          "default",
			Parent:        m,
		}
		m.Views = append(m.Views, v)
	}
}

// Project creates a ResultTypeExpr containing the fields defined in the view
// expression of m named after the view argument. Project also returns a links
// object created after the link expression of m if there is one.
//
// The resulting result type defines a default view. The result type identifier is
// computed by adding a parameter called "view" to the original identifier. The
// value of the "view" parameter is the name of the view.
func Project(m *ResultTypeExpr, view string) (*ResultTypeExpr, error) {
	_, params, _ := mime.ParseMediaType(m.Identifier)
	if params["view"] == view {
		// nothing to do
		return m, nil
	}
	if _, ok := m.Type.(*Array); ok {
		return projectCollection(m, view)
	}
	return projectSingle(m, view)
}

func projectSingle(m *ResultTypeExpr, view string) (*ResultTypeExpr, error) {
	v := m.View(view)
	if v == nil {
		return nil, fmt.Errorf("unknown view %#v", view)
	}
	viewObj := v.Type.(*Object)

	// Compute validations - view may not have all fields
	var val *ValidationExpr
	if m.Validation != nil {
		var required []string
		for _, n := range m.Validation.Required {
			if att := viewObj.Attribute(n); att != nil {
				required = append(required, n)
			}
		}
		val = m.Validation.Dup()
		val.Required = required
	}

	// Compute description
	desc := m.Description
	if desc == "" {
		desc = m.TypeName + " result type"
	}
	desc += " (" + view + " view)"

	// Compute type name
	typeName := m.TypeName
	if view != "default" {
		typeName += strings.Title(view)
	}

	params := map[string]string{"view": view}
	id := mime.FormatMediaType(m.Identifier, params)
	projected := &ResultTypeExpr{
		Identifier: id,
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

	projectedObj := projected.Type.(*Object)
	mtObj := m.Type.(*Object)
	for _, nat := range *viewObj {
		if at := mtObj.Attribute(nat.Name); at != nil {
			pat, err := projectRecursive(at, nat, view)
			if err != nil {
				return nil, err
			}
			projectedObj.Set(nat.Name, pat)
		}
	}
	return projected, nil
}

func projectCollection(m *ResultTypeExpr, view string) (*ResultTypeExpr, error) {
	// Project the collection element result type
	e := m.Type.(*Array).ElemType.Type.(*ResultTypeExpr) // validation checked this cast would work
	pe, err2 := Project(e, view)
	if err2 != nil {
		return nil, fmt.Errorf("collection element: %s", err2)
	}

	// Build the projected collection with the results
	params := map[string]string{"view": view}
	id := mime.FormatMediaType(m.Identifier, params)
	desc := m.TypeName + " is the result type for an array of " + e.TypeName + " (" + view + " view)"
	proj := &ResultTypeExpr{
		Identifier: id,
		UserTypeExpr: &UserTypeExpr{
			AttributeExpr: &AttributeExpr{
				Description:  desc,
				Type:         &Array{ElemType: &AttributeExpr{Type: pe}},
				UserExamples: m.UserExamples,
			},
			TypeName: pe.TypeName + "Collection",
		},
	}
	proj.Views = []*ViewExpr{&ViewExpr{
		AttributeExpr: DupAtt(pe.View("default").AttributeExpr),
		Name:          "default",
		Parent:        pe,
	}}

	// Run the DSL that was created by the CollectionOf function
	if !eval.Execute(proj.DSL(), proj) {
		return nil, eval.Context.Errors
	}

	return proj, nil
}

func projectRecursive(at *AttributeExpr, vat *NamedAttributeExpr, view string) (*AttributeExpr, error) {
	at = DupAtt(at)
	if rt, ok := at.Type.(*ResultTypeExpr); ok {
		vatt := vat.Attribute
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
		pr, err := Project(rt, view)
		if err != nil {
			return nil, fmt.Errorf("view %#v on field %#v cannot be computed: %s", view, vat.Name, err)
		}
		at.Type = pr
		return at, nil
	}
	if obj := AsObject(at.Type); obj != nil {
		vobj := AsObject(vat.Attribute.Type)
		if vobj == nil {
			return at, nil
		}
		for _, cnat := range *obj {
			var cvnat *NamedAttributeExpr
			for _, nnat := range *vobj {
				if nnat.Name == cnat.Name {
					cvnat = nnat
					break
				}
			}
			if cvnat == nil {
				continue
			}
			pat, err := projectRecursive(cnat.Attribute, cvnat, view)
			if err != nil {
				return nil, err
			}
			cnat.Attribute = pat
		}
		return at, nil
	}
	if ar := AsArray(at.Type); ar != nil {
		pat, err := projectRecursive(ar.ElemType, vat, view)
		if err != nil {
			return nil, err
		}
		ar.ElemType = pat
	}
	return at, nil
}

// projectIdentifier computes the projected result type identifier by adding the
// "view" param. We need the projected result type identifier to be different so
// that looking up projected result types from ProjectedResultTypes works
// correctly. It's also good for clients.
func (m *ResultTypeExpr) projectIdentifier(view string) string {
	base, params, err := mime.ParseMediaType(m.Identifier)
	if err != nil {
		base = m.Identifier
	}
	params["view"] = view
	return mime.FormatMediaType(base, params)
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
