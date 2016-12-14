package design

import (
	"fmt"
	"go/build"
	"mime"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/goadesign/goa/dslengine"
)

type routeInfo struct {
	Key       string
	Resource  *ResourceDefinition
	Action    *ActionDefinition
	Route     *RouteDefinition
	Wildcards []*wildCardInfo
}

type wildCardInfo struct {
	Name string
	Orig dslengine.Definition
}

func newRouteInfo(resource *ResourceDefinition, action *ActionDefinition, route *RouteDefinition) *routeInfo {
	vars := route.Params()
	wi := make([]*wildCardInfo, len(vars))
	for i, v := range vars {
		var orig dslengine.Definition
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
func (a *APIDefinition) Validate() error {

	// This is a little bit hacky but we need the generated media types DSLs to run first so
	// that their views are defined otherwise we risk running into validation errors where an
	// attribute defined on a non generated media type uses a generated mediatype (i.e.
	// CollectionOf(Foo)) with a specific view that hasn't been set yet.
	// TBD: Maybe GeneratedMediaTypes should not be a separate DSL root.
	for _, mt := range GeneratedMediaTypes {
		dslengine.Execute(mt.DSLFunc, mt)
		mt.DSLFunc = nil // So that it doesn't run again when the generated media types DSL root is executed
	}

	verr := new(dslengine.ValidationErrors)
	if a.Params != nil {
		verr.Merge(a.Params.Validate("base parameters", a))
	}

	a.validateContact(verr)
	a.validateLicense(verr)
	a.validateDocs(verr)
	a.validateOrigins(verr)

	var allRoutes []*routeInfo
	a.IterateResources(func(r *ResourceDefinition) error {
		verr.Merge(r.Validate())
		r.IterateActions(func(ac *ActionDefinition) error {
			if ac.Docs != nil && ac.Docs.URL != "" {
				if _, err := url.ParseRequestURI(ac.Docs.URL); err != nil {
					verr.Add(ac, "invalid action docs URL value: %s", err)
				}
			}
			for _, ro := range ac.Routes {
				if ro.IsAbsolute() {
					continue
				}
				info := newRouteInfo(r, ac, ro)
				allRoutes = append(allRoutes, info)
				rwcs := ExtractWildcards(ac.Parent.FullPath())
				wcs := ExtractWildcards(ro.Path)
				for _, rwc := range rwcs {
					for _, wc := range wcs {
						if rwc == wc {
							verr.Add(ac, `duplicate wildcard "%s" in resource base path "%s" and action route "%s"`,
								wc, ac.Parent.FullPath(), ro.Path)
						}
					}
				}
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
		verr.Merge(mt.Validate())
		return nil
	})
	a.IterateUserTypes(func(t *UserTypeDefinition) error {
		verr.Merge(t.Validate("", a))
		return nil
	})
	a.IterateResponses(func(r *ResponseDefinition) error {
		verr.Merge(r.Validate())
		return nil
	})
	for _, dec := range a.Consumes {
		verr.Merge(dec.Validate())
	}
	for _, enc := range a.Produces {
		verr.Merge(enc.Validate())
	}

	err := verr.AsError()
	if err == nil {
		// *ValidationErrors(nil) != error(nil)
		return nil
	}
	return err
}

func (a *APIDefinition) validateContact(verr *dslengine.ValidationErrors) {
	if a.Contact != nil && a.Contact.URL != "" {
		if _, err := url.ParseRequestURI(a.Contact.URL); err != nil {
			verr.Add(a, "invalid contact URL value: %s", err)
		}
	}
}

func (a *APIDefinition) validateLicense(verr *dslengine.ValidationErrors) {
	if a.License != nil && a.License.URL != "" {
		if _, err := url.ParseRequestURI(a.License.URL); err != nil {
			verr.Add(a, "invalid license URL value: %s", err)
		}
	}
}

func (a *APIDefinition) validateDocs(verr *dslengine.ValidationErrors) {
	if a.Docs != nil && a.Docs.URL != "" {
		if _, err := url.ParseRequestURI(a.Docs.URL); err != nil {
			verr.Add(a, "invalid docs URL value: %s", err)
		}
	}
}

func (a *APIDefinition) validateOrigins(verr *dslengine.ValidationErrors) {
	for _, origin := range a.Origins {
		verr.Merge(origin.Validate())
	}
}

// Validate tests whether the resource definition is consistent: action names are valid and each action is
// valid.
func (r *ResourceDefinition) Validate() *dslengine.ValidationErrors {
	verr := new(dslengine.ValidationErrors)
	if r.Name == "" {
		verr.Add(r, "Resource name cannot be empty")
	}
	r.validateActions(verr)
	if r.ParentName != "" {
		r.validateParent(verr)
	}
	for _, resp := range r.Responses {
		verr.Merge(resp.Validate())
	}
	if r.Params != nil {
		verr.Merge(r.Params.Validate("resource parameters", r))
	}
	for _, origin := range r.Origins {
		verr.Merge(origin.Validate())
	}
	return verr.AsError()
}

func (r *ResourceDefinition) validateActions(verr *dslengine.ValidationErrors) {
	found := false
	for _, a := range r.Actions {
		if a.Name == r.CanonicalActionName {
			found = true
		}
		verr.Merge(a.Validate())
	}
	for _, f := range r.FileServers {
		verr.Merge(f.Validate())
	}
	if r.CanonicalActionName != "" && !found {
		verr.Add(r, `unknown canonical action "%s"`, r.CanonicalActionName)
	}
}

func (r *ResourceDefinition) validateParent(verr *dslengine.ValidationErrors) {
	p, ok := Design.Resources[r.ParentName]
	if !ok {
		verr.Add(r, "Parent resource named %#v not found", r.ParentName)
	} else {
		if p.CanonicalAction() == nil {
			verr.Add(r, "Parent resource %#v has no canonical action", r.ParentName)
		}
	}
}

// Validate makes sure the CORS definition origin is valid.
func (cors *CORSDefinition) Validate() *dslengine.ValidationErrors {
	verr := new(dslengine.ValidationErrors)
	if !cors.Regexp && strings.Count(cors.Origin, "*") > 1 {
		verr.Add(cors, "invalid origin, can only contain one wildcard character")
	}
	if cors.Regexp {
		_, err := regexp.Compile(cors.Origin)
		if err != nil {
			verr.Add(cors, "invalid origin, should be a valid regular expression")
		}
	}
	return verr
}

// Validate validates the encoding MIME type and Go package path if set.
func (enc *EncodingDefinition) Validate() *dslengine.ValidationErrors {
	verr := new(dslengine.ValidationErrors)
	if len(enc.MIMETypes) == 0 {
		verr.Add(enc, "missing MIME type")
		return verr
	}
	for _, m := range enc.MIMETypes {
		_, _, err := mime.ParseMediaType(m)
		if err != nil {
			verr.Add(enc, "invalid MIME type %#v: %s", m, err)
		}
	}
	if len(enc.PackagePath) > 0 {
		rel := filepath.FromSlash(enc.PackagePath)
		dir, err := os.Getwd()
		if err != nil {
			verr.Add(enc, "couldn't retrieve working directory %s", err)
			return verr
		}
		_, err = build.Default.Import(rel, dir, build.FindOnly)
		if err != nil {
			verr.Add(enc, "invalid Go package path %#v: %s", enc.PackagePath, err)
			return verr
		}
	} else {
		for _, m := range enc.MIMETypes {
			if _, ok := KnownEncoders[m]; !ok {
				knownMIMETypes := make([]string, len(KnownEncoders))
				i := 0
				for k := range KnownEncoders {
					knownMIMETypes[i] = k
					i++
				}
				sort.Strings(knownMIMETypes)
				verr.Add(enc, "Encoders not known for all MIME types, use Package to specify encoder Go package. MIME types with known encoders are %s",
					strings.Join(knownMIMETypes, ", "))
			}
		}
	}
	if enc.Function != "" && enc.PackagePath == "" {
		verr.Add(enc, "Must specify encoder package page with PackagePath")
	}
	return verr
}

// Validate tests whether the action definition is consistent: parameters have unique names and it has at least
// one response.
func (a *ActionDefinition) Validate() *dslengine.ValidationErrors {
	verr := new(dslengine.ValidationErrors)
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
		verr.Merge(r.Validate())
	}
	verr.Merge(a.ValidateParams())
	if a.Payload != nil {
		verr.Merge(a.Payload.Validate("action payload", a))
	}
	if a.Parent == nil {
		verr.Add(a, "missing parent resource")
	}
	if a.Params != nil {
		for n, p := range a.Params.Type.ToObject() {
			if p.Type.IsPrimitive() {
				continue
			}
			if p.Type.IsArray() {
				if p.Type.ToArray().ElemType.Type.IsPrimitive() {
					continue
				}
			}
			verr.Add(a, "Param %s has an invalid type, action params must be primitives or arrays of primitives", n)
		}
	}

	return verr.AsError()
}

// Validate checks the file server is properly initialized.
func (f *FileServerDefinition) Validate() *dslengine.ValidationErrors {
	verr := new(dslengine.ValidationErrors)
	if f.FilePath == "" {
		verr.Add(f, "File server must have a non empty file path")
	}
	if f.RequestPath == "" {
		verr.Add(f, "File server must have a non empty route path")
	}
	if f.Parent == nil {
		verr.Add(f, "missing parent resource")
	}
	matches := WildcardRegex.FindAllString(f.RequestPath, -1)
	if len(matches) == 1 {
		if !strings.HasSuffix(f.RequestPath, matches[0]) {
			verr.Add(f, "invalid request path %s, must end with a wildcard starting with *", f.RequestPath)
		}
	}
	if len(matches) > 2 {
		verr.Add(f, "invalid request path, may only contain one wildcard")
	}

	return verr.AsError()
}

// ValidateParams checks the action parameters (make sure they have names, members and types).
func (a *ActionDefinition) ValidateParams() *dslengine.ValidationErrors {
	verr := new(dslengine.ValidationErrors)
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
		} else if p.Type.Kind() == HashKind {
			verr.Add(a, `parameter %s cannot be a hash, only action payloads may be of type hash`, n)
		}
		ctx := fmt.Sprintf("parameter %s", n)
		verr.Merge(p.Validate(ctx, a))
	}
	for _, resp := range a.Responses {
		verr.Merge(resp.Validate())
	}
	return verr.AsError()
}

// validated keeps track of validated attributes to handle cyclical definitions.
var validated = make(map[*AttributeDefinition]bool)

// Validate tests whether the attribute definition is consistent: required fields exist.
// Since attributes are unaware of their context, additional context information can be provided
// to be used in error messages.
// The parent definition context is automatically added to error messages.
func (a *AttributeDefinition) Validate(ctx string, parent dslengine.Definition) *dslengine.ValidationErrors {
	if validated[a] {
		return nil
	}
	validated[a] = true
	verr := new(dslengine.ValidationErrors)
	if a.Type == nil {
		verr.Add(parent, "attribute type is nil")
		return verr
	}
	if ctx != "" {
		ctx += " - "
	}
	// If both Default and Enum are given, make sure the Default value is one of Enum values.
	// TODO: We only do the default value and enum check just for primitive types.
	// Issue 388 (https://github.com/goadesign/goa/issues/388) will address this for other types.
	if a.Type.IsPrimitive() && a.DefaultValue != nil && a.Validation != nil && a.Validation.Values != nil {
		var found bool
		for _, e := range a.Validation.Values {
			if e == a.DefaultValue {
				found = true
				break
			}
		}
		if !found {
			verr.Add(parent, "%sdefault value %#v is not one of the accepted values: %#v", ctx, a.DefaultValue, a.Validation.Values)
		}
	}
	o := a.Type.ToObject()
	if o != nil {
		for _, n := range a.AllRequired() {
			found := false
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
		for n, att := range o {
			ctx = fmt.Sprintf("field %s", n)
			verr.Merge(att.Validate(ctx, parent))
		}
	} else {
		if a.Type.IsArray() {
			elemType := a.Type.ToArray().ElemType
			verr.Merge(elemType.Validate(ctx, a))
		}
	}

	return verr.AsError()
}

// Validate checks that the response definition is consistent: its status is set and the media
// type definition if any is valid.
func (r *ResponseDefinition) Validate() *dslengine.ValidationErrors {
	verr := new(dslengine.ValidationErrors)
	if r.Headers != nil {
		verr.Merge(r.Headers.Validate("response headers", r))
	}
	if r.Status == 0 {
		verr.Add(r, "response status not defined")
	}
	return verr.AsError()
}

// Validate checks that the route definition is consistent: it has a parent.
func (r *RouteDefinition) Validate() *dslengine.ValidationErrors {
	verr := new(dslengine.ValidationErrors)
	if r.Parent == nil {
		verr.Add(r, "missing route parent action")
	}
	return verr.AsError()
}

// Validate checks that the user type definition is consistent: it has a name and the attribute
// backing the type is valid.
func (u *UserTypeDefinition) Validate(ctx string, parent dslengine.Definition) *dslengine.ValidationErrors {
	verr := new(dslengine.ValidationErrors)
	if u.TypeName == "" {
		verr.Add(parent, "%s - %s", ctx, "User type must have a name")
	}
	verr.Merge(u.AttributeDefinition.Validate(ctx, u))
	return verr.AsError()
}

// Validate checks that the media type definition is consistent: its identifier is a valid media
// type identifier.
func (m *MediaTypeDefinition) Validate() *dslengine.ValidationErrors {
	verr := new(dslengine.ValidationErrors)
	verr.Merge(m.UserTypeDefinition.Validate("", m))
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
			verr.Merge(att.Validate("attribute "+n, m))
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
	hasDefaultView := false
	for n, v := range m.Views {
		if n == "default" {
			hasDefaultView = true
		}
		verr.Merge(v.Validate())
	}
	if !hasDefaultView {
		verr.Add(m, `media type does not define the default view, use View("default", ...) to define it.`)
	}

	for _, l := range m.Links {
		verr.Merge(l.Validate())
	}
	return verr.AsError()
}

// Validate checks that the link definition is consistent: it has a media type or the name of an
// attribute part of the parent media type.
func (l *LinkDefinition) Validate() *dslengine.ValidationErrors {
	verr := new(dslengine.ValidationErrors)
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
		} else {
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
	}
	return verr.AsError()
}

// Validate checks that the view definition is consistent: it has a  parent media type and the
// underlying definition type is consistent.
func (v *ViewDefinition) Validate() *dslengine.ValidationErrors {
	verr := new(dslengine.ValidationErrors)
	if v.Parent == nil {
		verr.Add(v, "View must have a parent media type")
	}
	verr.Merge(v.AttributeDefinition.Validate("", v))
	return verr.AsError()
}
