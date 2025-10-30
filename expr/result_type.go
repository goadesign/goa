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

	// ViewMetaKey is the key used to store the view name in the attribute meta.
	ViewMetaKey = "view"
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
		// "Content-Type" header. Deprecated.
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
				Validation:  &ValidationExpr{Required: []string{"name", "id", "message", "temporary", "timeout", "fault"}},
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
		Name:          DefaultView,
	}
)

// NewResultTypeExpr creates a result type definition but does not
// execute the DSL.
func NewResultTypeExpr(name, identifier string, fn func()) *ResultTypeExpr {
	return &ResultTypeExpr{
		UserTypeExpr: &UserTypeExpr{
			AttributeExpr: &AttributeExpr{Type: &Object{}, DSLFunc: fn},
			TypeName:      name,
			UID:           identifier,
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
func (*ResultTypeExpr) Kind() Kind { return ResultTypeKind }

// Dup creates a deep copy of the result type given a deep copy of its attribute.
func (rt *ResultTypeExpr) Dup(att *AttributeExpr) UserType {
	return &ResultTypeExpr{
		UserTypeExpr: rt.UserTypeExpr.Dup(att).(*UserTypeExpr),
		Identifier:   rt.Identifier,
		Views:        rt.Views,
	}
}

// ID returns the identifier of the result type.
func (rt *ResultTypeExpr) ID() string {
	return rt.Identifier
}

// Name returns the result type name.
func (rt *ResultTypeExpr) Name() string { return rt.TypeName }

// View returns the view with the given name.
func (rt *ResultTypeExpr) View(name string) *ViewExpr {
	for _, v := range rt.Views {
		if v.Name == name {
			return v
		}
	}
	return nil
}

// HasMultipleViews returns true if the result type has more than one view.
func (rt *ResultTypeExpr) HasMultipleViews() bool {
	return len(rt.Views) > 1
}

// ViewHasAttribute returns true if the result type view has the given
// attribute.
func (rt *ResultTypeExpr) ViewHasAttribute(view, attr string) bool {
	v := rt.View(view)
	if v == nil {
		return false
	}
	return v.Find(attr) != nil
}

// Finalize builds the default view if not explicitly defined and finalizes
// the underlying UserTypeExpr.
func (rt *ResultTypeExpr) Finalize() {
	rt.useExplicitView()
	rt.ensureDefaultView()
	rt.UserTypeExpr.Finalize()
	seen := make(map[string]struct{})
	walkAttribute(rt.AttributeExpr, func(_ string, att *AttributeExpr) error { // nolint: errcheck
		if rt, ok := att.Type.(*ResultTypeExpr); ok {
			if _, ok := seen[rt.Identifier]; !ok {
				seen[rt.Identifier] = struct{}{}
				rt.useExplicitView()
				rt.ensureDefaultView()
			}
		}
		return nil
	})
}

// useExplicitView projects the result type using the view explicitly set on the
// attribute if any.
func (rt *ResultTypeExpr) useExplicitView() {
	if view, ok := rt.Meta.Last(ViewMetaKey); ok {
		p, err := Project(rt, view)
		if err != nil {
			panic(err) // bug - presence of view meta should have been validated before
		}
		*rt = *p
	}
}

// ensureDefaultView builds the default view if not explicitly defined.
func (rt *ResultTypeExpr) ensureDefaultView() {
	if rt.View(DefaultView) == nil {
		att := DupAtt(rt.AttributeExpr)
		if arr := AsArray(att.Type); arr != nil {
			att.Type = AsObject(arr.ElemType.Type)
		}
		v := &ViewExpr{
			AttributeExpr: att,
			Name:          DefaultView,
			Parent:        rt,
		}
		rt.Views = append(rt.Views, v)
	}
}

// Project creates a ResultTypeExpr containing the fields defined in the view
// expression of m named after the view argument.
//
// The resulting result type defines a default view. The result type identifier is
// computed by adding a parameter called "view" to the original identifier. The
// value of the "view" parameter is the name of the view.
//
// Project returns an error if the view does not exist for the given result type
// or any result type that makes up its attributes recursively. Note that
// individual attributes may use a different view. In this case Project uses
// that view and returns an error if it isn't defined on the attribute type.
func Project(rt *ResultTypeExpr, view string) (*ResultTypeExpr, error) {
	return project(rt, view, make(map[string]*AttributeExpr))
}

func project(rt *ResultTypeExpr, view string, seen map[string]*AttributeExpr) (*ResultTypeExpr, error) {
	_, params, _ := mime.ParseMediaType(rt.Identifier)
	if params["view"] == view {
		// nothing to do
		return rt, nil
	}
	if _, ok := rt.Type.(*Array); ok {
		return projectCollection(rt, view, seen)
	}
	return projectSingle(rt, view, seen)
}

func projectSingle(rt *ResultTypeExpr, view string, seen map[string]*AttributeExpr) (*ResultTypeExpr, error) {
	v := rt.View(view)
	if v == nil {
		return nil, fmt.Errorf("unknown view %#v", view)
	}
	viewObj := v.Type.(*Object)

	// Compute validations - view may not have all fields
	var val *ValidationExpr
	if rt.Validation != nil {
		var required []string
		for _, n := range rt.Validation.Required {
			if att := viewObj.Attribute(n); att != nil {
				required = append(required, n)
			}
		}
		val = rt.Validation.Dup()
		val.Required = required
	}

	// Compute description
	desc := rt.Description
	if desc == "" {
		desc = rt.TypeName + " result type"
	}
	desc += " (" + view + " view)"

	// Compute type name
	typeName := rt.TypeName
	if view != DefaultView {
		typeName += Title(view)
	}

	var ut *UserTypeExpr
	if att, ok := seen[hashAttrAndView(rt.Attribute(), view)]; ok {
		if rt, ok2 := att.Type.(*ResultTypeExpr); ok2 {
			ut = &UserTypeExpr{
				AttributeExpr: DupAtt(rt.Attribute()),
				TypeName:      rt.TypeName,
			}
		}
	}
	id := rt.projectIdentifier(view)
	if ut == nil {
		ut = &UserTypeExpr{
			AttributeExpr: &AttributeExpr{
				Description: desc,
				Validation:  val,
			},
		}
	}
	ut.TypeName = typeName
	ut.UID = id
	ut.Type = Dup(v.Type)
	ut.UserExamples = v.UserExamples
	projected := &ResultTypeExpr{
		Identifier:   id,
		UserTypeExpr: ut,
	}
	projected.Views = []*ViewExpr{{
		Name:          DefaultView,
		AttributeExpr: DupAtt(v.AttributeExpr),
		Parent:        projected,
	}}

	projectedObj := projected.Type.(*Object)
	mtObj := AsObject(rt.Type)
	for _, nat := range *viewObj {
		if at := mtObj.Attribute(nat.Name); at != nil {
			pat, err := projectRecursive(at, nat, view, seen)
			if err != nil {
				return nil, err
			}
			projectedObj.Set(nat.Name, pat)
		}
	}
	return projected, nil
}

func projectCollection(rt *ResultTypeExpr, view string, seen map[string]*AttributeExpr) (*ResultTypeExpr, error) {
	// Project the collection element result type
	e := rt.Type.(*Array).ElemType.Type.(*ResultTypeExpr) // validation checked this cast would work
	pe, err2 := project(e, view, seen)
	if err2 != nil {
		return nil, fmt.Errorf("collection element: %w", err2)
	}

	// Build the projected collection with the results
	id := rt.projectIdentifier(view)
	proj := &ResultTypeExpr{
		Identifier: id,
		UserTypeExpr: &UserTypeExpr{
			AttributeExpr: &AttributeExpr{
				Description:  rt.TypeName + " is the result type for an array of " + e.TypeName + " (" + view + " view)",
				Type:         &Array{ElemType: &AttributeExpr{Type: pe}},
				UserExamples: rt.UserExamples,
			},
			TypeName: pe.TypeName + "Collection",
			UID:      id,
		},
		Views: []*ViewExpr{{
			AttributeExpr: DupAtt(pe.View(DefaultView).AttributeExpr),
			Name:          DefaultView,
			Parent:        pe,
		}},
	}

	// Run the DSL that was created by the CollectionOf function
	if !eval.Execute(proj.DSL(), proj) {
		return nil, eval.Context.Errors
	}

	return proj, nil
}

func projectRecursive(at *AttributeExpr, vat *NamedAttributeExpr, view string, seen map[string]*AttributeExpr) (*AttributeExpr, error) {
	if att, ok := seen[hashAttrAndView(at, view)]; ok {
		return att, nil
	}
	at = DupAtt(at)

	if rt, ok := at.Type.(*ResultTypeExpr); ok {
		vatt := vat.Attribute
		view, ok := vatt.Meta.Last(ViewMetaKey)
		if !ok {
			if v, ok := at.Meta.Last(ViewMetaKey); ok {
				view = v
			} else {
				view = DefaultView
			}
		}
		seen[hashAttrAndView(at, view)] = at
		pr, err := project(rt, view, seen)
		if err != nil {
			return nil, fmt.Errorf("view %#v on field %#v cannot be computed: %w", view, vat.Name, err)
		}
		at.Type = pr
		return at, nil
	}

	if _, ok := at.Type.(*UserTypeExpr); ok {
		seen[hashAttrAndView(at, view)] = at
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
			pat, err := projectRecursive(cnat.Attribute, cvnat, view, seen)
			if err != nil {
				return nil, err
			}
			cnat.Attribute = pat
		}
		return at, nil
	}

	if ar := AsArray(at.Type); ar != nil {
		pat, err := projectRecursive(ar.ElemType, vat, view, seen)
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
func (rt *ResultTypeExpr) projectIdentifier(view string) string {
	base, params, err := mime.ParseMediaType(rt.Identifier)
	if err != nil {
		base = rt.Identifier
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

// hashAttrAndView computes a hash for an attribute and a view that returns the
// same value for two attributes and views that produce the same projected type.
func hashAttrAndView(att *AttributeExpr, view string) string {
	return Hash(att.Type, false, false, false) + "::" + view
}
