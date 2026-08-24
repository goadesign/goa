// This file defines the evaluated design root and validates relationships
// between its API, services, generated types, and explicitly relocated user
// types before code generation begins.
package expr

import (
	"fmt"
	"maps"
	"reflect"
	"slices"
	"sort"

	"goa.design/goa/v3/eval"
	goa "goa.design/goa/v3/pkg"
)

// Root is the root object built by the DSL.
var Root = new(RootExpr)

// DefaultProtoc is the default command to be invoked for generating code from protobuf schemas.
const DefaultProtoc = "protoc"

type (
	// RootExpr is the struct built by the DSL on process start.
	RootExpr struct {
		// API contains the API expression built by the DSL.
		API *APIExpr
		// Services contains the list of services exposed by the API.
		Services []*ServiceExpr
		// Interceptors contains the list of interceptors.
		Interceptors []*InterceptorExpr
		// Errors contains the list of errors returned by all the API
		// methods.
		Errors []*ErrorExpr
		// Types contains the user types described in the DSL.
		Types []UserType
		// ResultTypes contains the result types generated during DSL
		// execution.
		ResultTypes []*ResultTypeExpr
		// Conversions list the user type to external type mappings.
		Conversions []*TypeMap
		// Creations list the external type to user type mappings.
		Creations []*TypeMap
		// Schemes list the registered security schemes.
		Schemes []*SchemeExpr
	}

	// MetaExpr is a set of key/value pairs
	MetaExpr map[string][]string

	// TypeMap defines a user to external type mapping.
	TypeMap struct {
		// User is the user type being converted or created.
		User UserType

		// External is an instance of the type being converted from or to.
		External any
	}
)

// WalkSets returns the expressions in order of evaluation.
func (r *RootExpr) WalkSets(walk eval.SetWalker) {
	if r.API == nil {
		name := "API"
		if len(r.Services) > 0 {
			name = r.Services[0].Name
		}
		r.API = NewAPIExpr(name, func() {})
	}

	// Top level API DSL
	walk(eval.ExpressionSet{r.API})

	// Servers
	walk(eval.ToExpressionSet(r.API.Servers))

	// User types
	types := make(eval.ExpressionSet, len(r.Types))
	for i, t := range r.Types {
		types[i] = t.Attribute()
	}
	walk(types)

	// Result types, note: initialized when the generated result types are
	// walked so empty the first time around (in the first pass DSL
	// execution). See ResultTypesRoot.WalkSets for details.
	rtypes := make(eval.ExpressionSet, len(r.ResultTypes))
	for i, rt := range r.ResultTypes {
		rtypes[i] = rt
	}
	walk(rtypes)

	// Services
	for _, service := range r.Services {
		service.design = r
	}
	walk(eval.ToExpressionSet(r.Services))

	// Methods (must be done after services)
	var methods eval.ExpressionSet
	for _, s := range r.Services {
		for _, m := range s.Methods {
			methods = append(methods, m)
		}
	}
	walk(methods)

	// HTTP services and endpoints
	r.walkHTTPServices(r.API.HTTP.Services, walk)

	// JSON-RPC services and endpoints
	r.walkHTTPServices(r.API.JSONRPC.Services, walk)

	// GRPC services and endpoints
	grpcsvcs := make(eval.ExpressionSet, len(r.API.GRPC.Services))
	sort.SliceStable(r.API.GRPC.Services, func(i, j int) bool {
		return r.API.GRPC.Services[j].ParentName == r.API.GRPC.Services[i].Name()
	})
	var grpcepts eval.ExpressionSet
	for i, svc := range r.API.GRPC.Services {
		grpcsvcs[i] = svc
		for _, e := range svc.GRPCEndpoints {
			grpcepts = append(grpcepts, e)
		}
	}
	walk(eval.ExpressionSet{r.API.GRPC})
	walk(grpcsvcs)
	walk(grpcepts)
}

// DependsOn returns nil, the core DSL has no dependency.
func (*RootExpr) DependsOn() []eval.Root { return nil }

// Packages returns the Go import path to this and the dsl packages.
func (*RootExpr) Packages() []string {
	return []string{
		"goa.design/goa/v3/expr",
		"goa.design/goa/v3/dsl",
		fmt.Sprintf("goa.design/goa/v3@%s/expr", goa.Version()),
		fmt.Sprintf("goa.design/goa/v3@%s/dsl", goa.Version()),
	}
}

// UserType returns the user type expression with the given name if found, nil otherwise.
func (r *RootExpr) UserType(name string) UserType {
	for _, t := range r.Types {
		if t.Name() == name {
			return t
		}
	}
	for _, t := range r.ResultTypes {
		if t.Name() == name {
			return t
		}
	}
	return nil
}

// Service returns the service with the given name.
func (r *RootExpr) Service(name string) *ServiceExpr {
	for _, s := range r.Services {
		if s.Name == name {
			return s
		}
	}
	return nil
}

// Error returns the error with the given name.
func (r *RootExpr) Error(name string) *ErrorExpr {
	for _, e := range r.Errors {
		if e.Name == name {
			return e
		}
	}
	return nil
}

// EvalName is the name of the DSL.
func (*RootExpr) EvalName() string {
	return "design"
}

// Validate makes sure the root expression is valid for code generation.
func (r *RootExpr) Validate() error {
	var verr eval.ValidationErrors
	if r.API == nil {
		verr.Add(r, "Missing API declaration")
	}
	// Ensure user type Go type names are unique across the design. Duplicate
	// user type names (e.g., via TypeName) can lead to conflicting generated
	// code. The Type DSL checks for duplicate declared names; this check
	// covers collisions introduced by renaming.
	useen := make(map[string]struct{})
	for _, ut := range r.Types {
		name := ut.Name()
		if _, ok := useen[name]; ok {
			verr.Add(r, "type %#v defined twice", name)
		} else {
			useen[name] = struct{}{}
		}
	}
	// Ensure result type Go type names are unique across declared result types
	// (exclude generated collection/result types). Generated types (e.g.,
	// CollectionOf) do not set the openapi:typename meta, whereas declared
	// result types do (set when calling dsl.ResultType). This prevents
	// collisions like two declared ResultType blocks both using TypeName("A"),
	// while allowing declared types to coexist with generated collection types
	// that may share a name like "XCollection".
	seen := make(map[string]struct{})
	for _, rt := range r.ResultTypes {
		if _, declared := rt.Meta["openapi:typename"]; !declared {
			continue // skip generated result types
		}
		name := rt.Name()
		if _, ok := seen[name]; ok {
			verr.Add(r, "result type %#v defined twice", name)
		} else {
			seen[name] = struct{}{}
		}
	}

	verr.Merge(r.validateRelocatedUserTypes())
	verr.Merge(validateTypeMappings("conversion", r.Conversions))
	verr.Merge(validateTypeMappings("creation", r.Creations))

	return &verr
}

// validateTypeMappings rejects repeated declarations that would generate the
// same method on one user type. A reflected type includes its package path, so
// equally named external types from different packages remain distinct.
func validateTypeMappings(direction string, mappings []*TypeMap) *eval.ValidationErrors {
	type mappingIdentity struct {
		user     UserType
		external reflect.Type
	}
	var verr eval.ValidationErrors
	seen := make(map[mappingIdentity]struct{}, len(mappings))
	for _, mapping := range mappings {
		identity := mappingIdentity{
			user:     mapping.User.Origin(),
			external: reflect.TypeOf(mapping.External),
		}
		if _, ok := seen[identity]; ok {
			if direction == "conversion" {
				verr.Add(
					mapping.User,
					"conversion from user type %q to external type %q defined twice",
					mapping.User.Name(), identity.external,
				)
			} else {
				verr.Add(
					mapping.User,
					"creation from external type %q to user type %q defined twice",
					identity.external, mapping.User.Name(),
				)
			}
			continue
		}
		seen[identity] = struct{}{}
	}
	return &verr
}

// validateRelocatedUserTypes enforces that relocated user types (those with
// `struct:pkg:path`) only depend on other declared user types with an explicit
// generation location.
//
// Without this constraint, generated code would need to reference service-local
// user types across packages, which is not safe (it either fails to compile or
// forces import cycles). Generated/derived user types (e.g. union branch
// wrappers) are exempt because they are materialized alongside their owning
// types.
func (r *RootExpr) validateRelocatedUserTypes() *eval.ValidationErrors {
	var verr eval.ValidationErrors
	declared := make(map[UserType]struct{}, len(r.Types))
	for _, ut := range r.Types {
		declared[ut.Origin()] = struct{}{}
	}
	for _, ut := range r.Types {
		pkgPath, ok := ut.Attribute().Meta.Last("struct:pkg:path")
		if !ok || pkgPath == "" {
			continue
		}
		seen := make(map[UserType]struct{})
		r.walkUserTypeDependencies(ut, seen, "", func(dep UserType, path string) {
			if dep.Origin() == ut.Origin() {
				return
			}
			if _, ok := declared[dep.Origin()]; !ok {
				// Generated/derived user types (e.g. union branch wrappers) are
				// materialized alongside their owning types and do not require an
				// explicit struct:pkg:path.
				return
			}
			if _, ok := dep.Attribute().Meta.Last("struct:pkg:path"); ok {
				return
			}
			msg := fmt.Sprintf(
				`type %q is generated in package %q (struct:pkg:path) but depends on %q with no explicit struct:pkg:path`,
				ut.Name(),
				pkgPath,
				dep.Name(),
			)
			if path != "" {
				msg = fmt.Sprintf("%s (referenced via %s)", msg, path)
			}
			msg = fmt.Sprintf(
				`%s; fix: add Meta("struct:pkg:path", %q) to %q (or move it to an explicitly located shared type)`,
				msg,
				pkgPath,
				dep.Name(),
			)
			verr.Add(dep, "%s", msg)
		})
	}
	return &verr
}

// walkUserTypeDependencies traverses the attribute graph reachable from root and
// invokes visit for each encountered user type.
func (r *RootExpr) walkUserTypeDependencies(root UserType, seen map[UserType]struct{}, path string, visit func(UserType, string)) {
	if root == nil || root.Attribute() == nil {
		return
	}
	r.walkAttributeUserTypes(root.Attribute(), seen, path, visit)
}

// walkAttributeUserTypes traverses att and invokes visit for each encountered
// user type.
//
// The path argument records the traversal path through objects, arrays, maps,
// and unions and is intended for diagnostics.
func (r *RootExpr) walkAttributeUserTypes(att *AttributeExpr, seen map[UserType]struct{}, path string, visit func(UserType, string)) {
	if att == nil || att.Type == Empty {
		return
	}
	switch t := att.Type.(type) {
	case UserType:
		origin := t.Origin()
		if _, ok := seen[origin]; ok {
			return
		}
		seen[origin] = struct{}{}
		visit(t, path)
		r.walkAttributeUserTypes(t.Attribute(), seen, path, visit)
	case *Object:
		for _, nat := range *t {
			next := nat.Name
			if path != "" {
				next = path + "." + nat.Name
			}
			r.walkAttributeUserTypes(nat.Attribute, seen, next, visit)
		}
	case *Array:
		next := path + "[]"
		r.walkAttributeUserTypes(t.ElemType, seen, next, visit)
	case *Map:
		r.walkAttributeUserTypes(t.KeyType, seen, path+"[key]", visit)
		r.walkAttributeUserTypes(t.ElemType, seen, path+"[value]", visit)
	case *Union:
		for _, nat := range t.Values {
			next := nat.Name
			if path != "" {
				next = path + "." + nat.Name
			}
			r.walkAttributeUserTypes(nat.Attribute, seen, next, visit)
		}
	}
}

// Finalize finalizes the server expressions.
func (r *RootExpr) Finalize() {
	if r.API == nil {
		r.API = &APIExpr{}
	}
	if len(r.API.Servers) == 0 {
		r.API.Servers = []*ServerExpr{r.API.DefaultServer()}
	}
	for _, e := range r.Errors {
		e.Finalize()
	}
	for _, s := range r.API.Servers {
		s.Finalize()
	}
}

// walkHTTPServices walks the HTTP services and endpoints.
func (r *RootExpr) walkHTTPServices(svcs []*HTTPServiceExpr, walk eval.SetWalker) {
	sort.SliceStable(svcs, func(i, j int) bool {
		return svcs[j].ParentName == svcs[i].Name()
	})
	var httpepts eval.ExpressionSet
	var httpsvrs eval.ExpressionSet
	httpsvcs := make(eval.ExpressionSet, len(svcs))
	for i, svc := range svcs {
		httpsvcs[i] = svc
		for _, e := range svc.HTTPEndpoints {
			httpepts = append(httpepts, e)
		}
		for _, s := range svc.FileServers {
			httpsvrs = append(httpsvrs, s)
		}
	}
	walk(eval.ExpressionSet{r.API.HTTP})
	walk(httpsvcs)
	walk(httpepts)
	walk(httpsvrs)
}

// Dup creates a new map from the given expression.
func (m MetaExpr) Dup() MetaExpr {
	d := make(MetaExpr, len(m))
	maps.Copy(d, m)
	return d
}

// Merge merges src meta expression with m. If meta has intersecting set of
// keys on both m and src, then the values for those keys in src is appended
// to the values of the keys in m if not already existing.
func (m MetaExpr) Merge(src MetaExpr) {
	for k, vals := range src {
		if mvals, ok := m[k]; ok {
			var found bool
			for _, v := range vals {
				found = slices.Contains(mvals, v)
				if !found {
					mvals = append(mvals, v)
				}
			}
			m[k] = mvals
		} else {
			m[k] = vals
		}
	}
}

// Last returns the last value for a specific key, if the key exists and has
// values; otherwise returns an empty string, with the "ok" flag set to false.
func (m MetaExpr) Last(key string) (string, bool) {
	v, ok := m[key]
	if !ok {
		return "", false
	}

	l := len(v)
	if l < 1 {
		return "", false
	}

	return v[l-1], true
}
