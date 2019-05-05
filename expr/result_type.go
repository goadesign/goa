package expr

import (
	"fmt"
	"mime"
	"strings"

	"goa.design/goa/v3/eval"
)

const (
	// DefaultView is the name of the default result type view.
	DefaultView = "default"
)

type (
	// ResultTypeExpr is a user type which describes views used to
	// render responses.
	ResultTypeExpr struct {
		// A result type is a user type
		*UserTypeExpr
		// Identifier is the RFC 6838 result type media type identifier.
		Identifier string
		// ContentType identifies the value written to the response
		// "Content-Type" header.
		ContentType string
		// Views list the supported views indexed by name.
		Views []*ViewExpr
	}

	// ViewExpr defines which fields to render when building a response. The view
	// is an object whose field names must match the names of the parent result
	// type field names. The field definitions are inherited from the parent
	// result type but may be overridden.
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
						"name":    "bad_request",
						"id":      "3F1FKVRR",
						"message": "Value of ID must be an integer",
					},
				}},
				Validation: &ValidationExpr{Required: []string{"name", "id", "message", "temporary", "timeout", "fault"}},
			},
			TypeName: "error",
		},
		Identifier: ErrorResultIdentifier,
		Views:      []*ViewExpr{errorResultView},
	}

	errorResultType = &Object{
		{"name", &AttributeExpr{
			Type:         String,
			Description:  "Name is the name of this class of errors.",
			Meta:         MetaExpr{"struct:error:name": nil},
			UserExamples: []*ExampleExpr{{Value: "bad_request"}},
		}},
		{"id", &AttributeExpr{
			Type:         String,
			Description:  "ID is a unique identifier for this particular occurrence of the problem.",
			UserExamples: []*ExampleExpr{{Value: "123abc"}},
		}},
		{"message", &AttributeExpr{
			Type:         String,
			Description:  "Message is a human-readable explanation specific to this occurrence of the problem.",
			UserExamples: []*ExampleExpr{{Value: "parameter 'p' must be an integer"}},
		}},
		{"temporary", &AttributeExpr{
			Type:        Boolean,
			Description: "Is the error temporary?",
		}},
		{"timeout", &AttributeExpr{
			Type:        Boolean,
			Description: "Is the error a timeout?",
		}},
		{"fault", &AttributeExpr{
			Type:        Boolean,
			Description: "Is the error a server-side fault?",
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

// ID returns the identifier of the result type.
func (m *ResultTypeExpr) ID() string {
	return m.Identifier
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

// HasMultipleViews returns true if the result type has more than one view.
func (m *ResultTypeExpr) HasMultipleViews() bool {
	return len(m.Views) > 1
}

// ViewHasAttribute returns true if the result type view has the given
// attribute.
func (m *ResultTypeExpr) ViewHasAttribute(view, attr string) bool {
	v := m.View(view)
	if v == nil {
		return false
	}
	return v.AttributeExpr.Find(attr) != nil
}

// Finalize builds the default view if not explicitly defined and finalizes
// the underlying UserTypeExpr.
func (m *ResultTypeExpr) Finalize() {
	if m.View("default") == nil {
		att := DupAtt(m.AttributeExpr)
		if arr := AsArray(att.Type); arr != nil {
			att.Type = AsObject(arr.ElemType.Type)
		}
		v := &ViewExpr{
			AttributeExpr: att,
			Name:          "default",
			Parent:        m,
		}
		m.Views = append(m.Views, v)
	}
	m.UserTypeExpr.Finalize()
}

// Project creates a ResultTypeExpr containing the fields defined in the view
// expression of m named after the view argument.
//
// The resulting result type defines a default view. The result type identifier is
// computed by adding a parameter called "view" to the original identifier. The
// value of the "view" parameter is the name of the view.
func Project(m *ResultTypeExpr, view string, seen ...map[string]*AttributeExpr) (*ResultTypeExpr, error) {
	_, params, _ := mime.ParseMediaType(m.Identifier)
	if params["view"] == view {
		// nothing to do
		return m, nil
	}
	if _, ok := m.Type.(*Array); ok {
		return projectCollection(m, view, seen...)
	}
	return projectSingle(m, view, seen...)
}

func projectSingle(m *ResultTypeExpr, view string, seen ...map[string]*AttributeExpr) (*ResultTypeExpr, error) {
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

	var ut *UserTypeExpr
	if len(seen) > 0 {
		s := seen[0]
		if att, ok := s[m.Identifier]; ok {
			if rt, ok2 := att.Type.(*ResultTypeExpr); ok2 {
				ut = &UserTypeExpr{
					AttributeExpr: DupAtt(rt.Attribute()),
					TypeName:      rt.TypeName,
				}
			}
		}
	} else {
		seen = append(seen, make(map[string]*AttributeExpr))
	}
	if ut == nil {
		ut = &UserTypeExpr{
			AttributeExpr: &AttributeExpr{
				Description: desc,
				Validation:  val,
			},
		}
	}
	ut.TypeName = typeName
	ut.AttributeExpr.Type = Dup(v.Type)
	projected := &ResultTypeExpr{
		Identifier:   m.projectIdentifier(view),
		UserTypeExpr: ut,
	}
	projected.Views = []*ViewExpr{{
		Name:          "default",
		AttributeExpr: DupAtt(v.AttributeExpr),
		Parent:        projected,
	}}

	projectedObj := projected.Type.(*Object)
	mtObj := m.Type.(*Object)
	for _, nat := range *viewObj {
		if at := mtObj.Attribute(nat.Name); at != nil {
			pat, err := projectRecursive(at, nat, view, seen...)
			if err != nil {
				return nil, err
			}
			projectedObj.Set(nat.Name, pat)
		}
	}
	return projected, nil
}

func projectCollection(m *ResultTypeExpr, view string, seen ...map[string]*AttributeExpr) (*ResultTypeExpr, error) {
	// Project the collection element result type
	e := m.Type.(*Array).ElemType.Type.(*ResultTypeExpr) // validation checked this cast would work
	pe, err2 := Project(e, view, seen...)
	if err2 != nil {
		return nil, fmt.Errorf("collection element: %s", err2)
	}

	// Build the projected collection with the results
	proj := &ResultTypeExpr{
		Identifier: m.projectIdentifier(view),
		UserTypeExpr: &UserTypeExpr{
			AttributeExpr: &AttributeExpr{
				Description:  m.TypeName + " is the result type for an array of " + e.TypeName + " (" + view + " view)",
				Type:         &Array{ElemType: &AttributeExpr{Type: pe}},
				UserExamples: m.UserExamples,
			},
			TypeName: pe.TypeName + "Collection",
		},
		Views: []*ViewExpr{{
			AttributeExpr: DupAtt(pe.View("default").AttributeExpr),
			Name:          "default",
			Parent:        pe,
		}},
	}

	// Run the DSL that was created by the CollectionOf function
	if !eval.Execute(proj.DSL(), proj) {
		return nil, eval.Context.Errors
	}

	return proj, nil
}

func projectRecursive(at *AttributeExpr, vat *NamedAttributeExpr, view string, seen ...map[string]*AttributeExpr) (*AttributeExpr, error) {
	s := seen[0]
	ut, isUT := at.Type.(UserType)
	if isUT {
		if att, ok := s[ut.ID()]; ok {
			return att, nil
		}
	}
	at = DupAtt(at)
	if isUT {
		s[ut.ID()] = at
	}
	if rt, ok := at.Type.(*ResultTypeExpr); ok {
		vatt := vat.Attribute
		var view string
		if len(vatt.Meta["view"]) > 0 {
			view = vatt.Meta["view"][0]
		}
		if view == "" && len(at.Meta["view"]) > 0 {
			view = at.Meta["view"][0]
		}
		if view == "" {
			view = DefaultView
		}
		pr, err := Project(rt, view, seen...)
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
			pat, err := projectRecursive(cnat.Attribute, cvnat, view, seen...)
			if err != nil {
				return nil, err
			}
			cnat.Attribute = pat
		}
		return at, nil
	}
	if ar := AsArray(at.Type); ar != nil {
		pat, err := projectRecursive(ar.ElemType, vat, view, seen...)
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
	if params == nil {
		params = make(map[string]string)
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
