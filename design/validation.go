package design

import (
	"fmt"
	"strings"
)

// ValidationErrors records the errors encountered when running Validate.
type ValidationErrors struct {
	Errors      []error
	Definitions []DSLDefinition
}

// Error implements the error interface.
func (verr *ValidationErrors) Error() string {
	msg := make([]string, len(verr.Errors))
	for i, err := range verr.Errors {
		msg[i] = fmt.Sprintf("%s: %s", verr.Definitions[i].Context(), err)
	}
	return strings.Join(msg, "\n")
}

// Merge merges validation errors into the target.
func (verr *ValidationErrors) Merge(err *ValidationErrors) {
	verr.Errors = append(verr.Errors, err.Errors...)
	verr.Definitions = append(verr.Definitions, err.Definitions...)
}

// Add adds a validation error to the target.
// Add "flattens" validation errors so that the recorded errors are never ValidationErrors
// themselves.
func (verr *ValidationErrors) Add(def DSLDefinition, format string, vals ...interface{}) {
	err := fmt.Errorf(format, vals...)
	verr.Errors = append(verr.Errors, err)
	verr.Definitions = append(verr.Definitions, def)
}

// AsError returns an error if there are validation errors, nil otherwise.
func (verr *ValidationErrors) AsError() *ValidationErrors {
	if len(verr.Errors) > 0 {
		return verr
	}
	return nil
}

type routeInfo struct {
	Key       string
	Resource  *ResourceDefinition
	Action    *ActionDefinition
	Route     *RouteDefinition
	Wildcards []*wildCardInfo
}

type wildCardInfo struct {
	Name string
	Orig DSLDefinition
}

func newRouteInfo(resource *ResourceDefinition, action *ActionDefinition, route *RouteDefinition) *routeInfo {
	vars := route.Params()
	wi := make([]*wildCardInfo, len(vars))
	for i, v := range vars {
		var orig DSLDefinition
		if strings.Contains(route.Path, v) {
			orig = route
		} else if strings.Contains(resource.BasePath, v) {
			orig = resource
		} else {
			orig = Design
		}
		wi[i] = &wildCardInfo{Name: v, Orig: orig}
	}
	key := WildcardRegex.ReplaceAllLiteralString(route.FullPath(), "*")
	return &routeInfo{
		Key:       key,
		Resource:  resource,
		Action:    action,
		Route:     route,
		Wildcards: wi,
	}
}

// DifferentWildcards returns the list of wildcards in other that have a different name from the
// wildcard in target at the same position.
func (r *routeInfo) DifferentWildcards(other *routeInfo) (res [][2]*wildCardInfo) {
	for i, wc := range other.Wildcards {
		if r.Wildcards[i].Name != wc.Name {
			res = append(res, [2]*wildCardInfo{r.Wildcards[i], wc})
		}
	}
	return
}

// Validate tests whether the API definition is consistent: all resource parent names resolve to
// an actual resource.
func (a *APIDefinition) Validate() *ValidationErrors {
	verr := new(ValidationErrors)
	var allRoutes []*routeInfo
	if a.BaseParams != nil {
		if err := a.BaseParams.Validate("base parameters", a); err != nil {
			verr.Merge(err)
		}
	}
	a.IterateResources(func(r *ResourceDefinition) error {
		if err := r.Validate(); err != nil {
			verr.Merge(err)
		}
		r.IterateActions(func(ac *ActionDefinition) error {
			for _, ro := range ac.Routes {
				info := newRouteInfo(r, ac, ro)
				allRoutes = append(allRoutes, info)
			}
			return nil
		})
		return nil
	})
	for _, route := range allRoutes {
		for _, other := range allRoutes {
			if route == other {
				continue
			}
			if strings.HasPrefix(route.Key, other.Key) {
				diffs := route.DifferentWildcards(other)
				if len(diffs) > 0 {
					var msg string
					conflicts := make([]string, len(diffs))
					for i, d := range diffs {
						conflicts[i] = fmt.Sprintf(`"%s" from %s and "%s" from %s`, d[0].Name, d[0].Orig.Context(), d[1].Name, d[1].Orig.Context())
					}
					msg = fmt.Sprintf("%s", strings.Join(conflicts, ", "))
					verr.Add(route.Action,
						`route "%s" conflicts with route "%s" of %s action %s. Make sure wildcards at the same positions have the same name. Conflicting wildcards are %s.`,
						route.Route.FullPath(),
						other.Route.FullPath(),
						other.Resource.Name,
						other.Action.Name,
						msg,
					)
				}
			}
		}
	}
	a.IterateMediaTypes(func(mt *MediaTypeDefinition) error {
		if err := mt.Validate(); err != nil {
			verr.Merge(err)
		}
		return nil
	})
	a.IterateUserTypes(func(t *UserTypeDefinition) error {
		if err := t.Validate("", a); err != nil {
			verr.Merge(err)
		}
		return nil
	})
	a.IterateResponses(func(r *ResponseDefinition) error {
		if err := r.Validate(); err != nil {
			verr.Merge(err)
		}
		return nil
	})

	return verr.AsError()
}

// Validate tests whether the resource definition is consistent: action names are valid and each action is
// valid.
func (r *ResourceDefinition) Validate() *ValidationErrors {
	verr := new(ValidationErrors)
	if r.Name == "" {
		verr.Add(r, "Resource name cannot be empty")
	}
	found := false
	for _, a := range r.Actions {
		if a.Name == r.CanonicalActionName {
			found = true
		}
		if err := a.Validate(); err != nil {
			verr.Merge(err)
		}
	}
	if r.CanonicalActionName != "" && !found {
		verr.Add(r, `unknown canonical action "%s"`, r.CanonicalActionName)
	}
	if r.BaseParams != nil {
		baseParams, ok := r.BaseParams.Type.(Object)
		if !ok {
			verr.Add(r, "invalid type for BaseParams, must be an Object", r)
		} else {
			vars := ExtractWildcards(r.BasePath)
			if len(vars) > 1 {
				if len(vars) != len(baseParams) {
					verr.Add(r, "BasePath defines parameters %s but BaseParams has %d elements",
						strings.Join([]string{
							strings.Join(vars[:len(vars)-1], ", "),
							vars[len(vars)-1],
						}, " and "),
						len(baseParams),
					)
				}
				for _, v := range vars {
					found := false
					for n := range baseParams {
						if v == n {
							found = true
							break
						}
					}
					if !found {
						verr.Add(r, "Variable %s from base path %s does not match any parameter from BaseParams",
							v, r.BasePath)
					}
				}
			} else {
				if len(baseParams) > 0 {
					verr.Add(r, "BasePath does not use variables defines in BaseParams")
				}
			}
		}
	}
	if r.ParentName != "" {
		p, ok := Design.Resources[r.ParentName]
		if !ok {
			verr.Add(r, "Parent resource named %#v not found", r.ParentName)
		} else {
			if p.CanonicalAction() == nil {
				verr.Add(r, "Parent resource %#v has no canonical action", r.ParentName)
			}
		}
	}
	for _, resp := range r.Responses {
		if err := resp.Validate(); err != nil {
			verr.Merge(err)
		}
	}
	if r.Params != nil {
		if err := r.Params.Validate("resource parameters", r); err != nil {
			verr.Merge(err)
		}
	}
	return verr.AsError()
}

// Validate tests whether the action definition is consistent: parameters have unique names and it has at least
// one response.
func (a *ActionDefinition) Validate() *ValidationErrors {
	verr := new(ValidationErrors)
	if a.Name == "" {
		verr.Add(a, "Action name cannot be empty")
	}
	if len(a.Routes) == 0 {
		verr.Add(a, "No route defined for action")
	}
	for i, r := range a.Responses {
		for j, r2 := range a.Responses {
			if i != j && r.Status == r2.Status {
				verr.Add(r, "Multiple response definitions with status code %d", r.Status)
			}
		}
		if err := r.Validate(); err != nil {
			verr.Merge(err)
		}
	}
	if err := a.ValidateParams(); err != nil {
		verr.Merge(err)
	}
	if a.Payload != nil {
		if err := a.Payload.Validate("action payload", a); err != nil {
			verr.Merge(err)
		}
	}
	if a.Parent == nil {
		verr.Add(a, "missing parent resource")
	}
	return verr.AsError()
}

// ValidateParams checks the action parameters (make sure they have names, members and types).
func (a *ActionDefinition) ValidateParams() *ValidationErrors {
	verr := new(ValidationErrors)
	if a.Params == nil {
		return nil
	}
	params, ok := a.Params.Type.(Object)
	if !ok {
		verr.Add(a, `"Params" field of action is not an object`)
	}
	var wcs []string
	for _, r := range a.Routes {
		rwcs := ExtractWildcards(r.FullPath())
		for _, rwc := range rwcs {
			found := false
			for _, wc := range wcs {
				if rwc == wc {
					found = true
					break
				}
			}
			if !found {
				wcs = append(wcs, rwc)
			}
		}
	}
	for n, p := range params {
		if n == "" {
			verr.Add(a, "action has parameter with no name")
		} else if p == nil {
			verr.Add(a, "definition of parameter %s cannot be nil", n)
		} else if p.Type == nil {
			verr.Add(a, "type of parameter %s cannot be nil", n)
		}
		if p.Type.Kind() == ObjectKind {
			verr.Add(a, `parameter %s cannot be an object, only action payloads may be of type object`, n)
		}
		ctx := fmt.Sprintf("parameter %s", n)
		if err := p.Validate(ctx, a); err != nil {
			verr.Merge(err)
		}
	}
	for _, resp := range a.Responses {
		if err := resp.Validate(); err != nil {
			verr.Merge(err)
		}
	}
	return verr.AsError()
}

// Validate tests whether the attribute definition is consistent: required fields exist.
// Since attributes are unaware of their context, additional context information can be provided
// to be used in error messages.
// The parent definition context is automatically added to error messages.
func (a *AttributeDefinition) Validate(ctx string, parent DSLDefinition) *ValidationErrors {
	verr := new(ValidationErrors)
	if a.Type == nil {
		verr.Add(parent, "attribute type is nil")
		return verr
	}
	if ctx != "" {
		ctx += " - "
	}
	o, isObject := a.Type.(Object)
	for _, v := range a.Validations {
		if r, ok := v.(*RequiredValidationDefinition); ok {
			if !isObject {
				verr.Add(parent, `%sonly objects may define a "Required" validation`, ctx)
			}
			for _, n := range r.Names {
				var found bool
				for an := range o {
					if n == an {
						found = true
						break
					}
				}
				if !found {
					verr.Add(parent, `%srequired field "%s" does not exist`, ctx, n)
				}
			}
		}
	}
	if isObject {
		for _, i := range a.IdentityProperties {
			found := false
			for n := range o {
				if i == n {
					found = true
					break
				}
			}
			if !found {
				verr.Add(a, `identity attribute name %#v does not match an existing attribute name`, i)
			}
		}
		for n, att := range o {
			ctx = fmt.Sprintf("field %s", n)
			if err := att.Validate(ctx, a); err != nil {
				verr.Merge(err)
			}
		}
	} else {
		if len(a.IdentityProperties) > 0 {
			verr.Add(a, "non-object attribute cannot define identity properties")
		}
		if a.Type.IsArray() {
			elemType := a.Type.ToArray().ElemType
			if err := elemType.Validate(ctx, a); err != nil {
				verr.Merge(err)
			}
		}
	}

	return verr.AsError()
}

// Validate checks that the response definition is consistent: its status is set and the media
// type definition if any is valid.
func (r *ResponseDefinition) Validate() *ValidationErrors {
	verr := new(ValidationErrors)
	if r.Headers != nil {
		if err := r.Headers.Validate("response headers", r); err != nil {
			verr.Merge(err)
		}
	}
	if r.Status == 0 {
		verr.Add(r, "response status not defined")
	}
	return verr.AsError()
}

// Validate checks that the route definition is consistent: it has a parent.
func (r *RouteDefinition) Validate() *ValidationErrors {
	verr := new(ValidationErrors)
	if r.Parent == nil {
		verr.Add(r, "missing route parent action")
	}
	return verr.AsError()
}

// Validate checks that the user type definition is consistent: it has a name.
func (u *UserTypeDefinition) Validate(ctx string, parent DSLDefinition) *ValidationErrors {
	verr := new(ValidationErrors)
	if u.TypeName == "" {
		verr.Add(parent, "%s - %s", ctx, "User type must have a name")
	}
	if err := u.AttributeDefinition.Validate(ctx, parent); err != nil {
		verr.Merge(err)
	}
	return verr.AsError()
}

// Validate checks that the media type definition is consistent: its identifier is a valid media
// type identifier.
func (m *MediaTypeDefinition) Validate() *ValidationErrors {
	verr := new(ValidationErrors)
	if err := m.UserTypeDefinition.Validate("", m); err != nil {
		verr.Merge(err)
	}
	if m.Type == nil { // TBD move this to somewhere else than validation code
		m.Type = String
	}
	var obj Object
	if a := m.Type.ToArray(); a != nil {
		if a.ElemType == nil {
			verr.Add(m, "array element type is nil")
		} else {
			if err := a.ElemType.Validate("array element", m); err != nil {
				verr.Merge(err)
			} else {
				if _, ok := a.ElemType.Type.(*MediaTypeDefinition); !ok {
					verr.Add(m, "collection media type array element type must be a media type, got %s", a.ElemType.Type.Name())
				} else {
					obj = a.ElemType.Type.ToObject()
				}
			}
		}
	} else {
		obj = m.Type.ToObject()
	}
	if obj != nil {
		for n, att := range obj {
			if err := att.Validate("attribute "+n, m); err != nil {
				verr.Merge(err)
			}
			if att.View != "" {
				cmt, ok := att.Type.(*MediaTypeDefinition)
				if !ok {
					verr.Add(m, "attribute %s of media type defines a view for rendering but its type is not MediaTypeDefinition", n)
				}
				if _, ok := cmt.Views[att.View]; !ok {
					verr.Add(m, "attribute %s of media type uses unknown view %#v", n, att.View)
				}
			}
		}
	}
	if !m.Type.IsArray() {
		hasDefaultView := false
		for n, v := range m.Views {
			if n == "default" {
				hasDefaultView = true
			}
			if err := v.Validate(); err != nil {
				verr.Merge(err)
			}
		}
		if !hasDefaultView {
			verr.Add(m, `media type does not define the default view, use View("default", ...) to define it.`)
		}
	}
	for _, l := range m.Links {
		if err := l.Validate(); err != nil {
			verr.Merge(err)
		}
	}
	return verr.AsError()
}

// Validate checks that the link definition is consistent: it has a media type or the name of an
// attribute part of the parent media type.
func (l *LinkDefinition) Validate() *ValidationErrors {
	verr := new(ValidationErrors)
	if l.Name == "" {
		verr.Add(l, "Links must have a name")
	}
	if l.Parent == nil {
		verr.Add(l, "Link must have a parent media type")
	}
	if l.Parent.ToObject() == nil {
		verr.Add(l, "Link parent media type must be an Object")
	}
	att, ok := l.Parent.ToObject()[l.Name]
	if !ok {
		verr.Add(l, "Link name must match one of the parent media type attribute names")
	} else {
		mediaType, ok := att.Type.(*MediaTypeDefinition)
		if !ok {
			verr.Add(l, "attribute type must be a media type")
		}
		viewFound := false
		view := l.View
		for v := range mediaType.Views {
			if v == view {
				viewFound = true
				break
			}
		}
		if !viewFound {
			verr.Add(l, "view %#v does not exist on target media type %#v", view, mediaType.Identifier)
		}
	}
	return verr.AsError()
}

// Validate checks that the view definition is consistent: it has a  parent media type and the
// underlying definition type is consistent.
func (v *ViewDefinition) Validate() *ValidationErrors {
	verr := new(ValidationErrors)
	if v.Parent == nil {
		verr.Add(v, "View must have a parent media type")
	}
	if err := v.AttributeDefinition.Validate("", v); err != nil {
		verr.Merge(err)
	}
	return verr.AsError()
}
