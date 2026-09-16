// Code generators use this file to turn type declarations and attributes into
// unique Go names and type references. Generated user types keep the identity
// of the declaration they came from. Other lookups use the caller's exact Hash
// value. After Freeze, callers can read existing names but cannot reserve new
// ones.
package codegen

import (
	"fmt"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"goa.design/goa/v3/expr"
)

type (
	// NameScope defines a naming scope.
	NameScope struct {
		names            map[string]string             // type hash to unique name
		userTypeNames    map[expr.UserType]string      // original user type to exact name
		unionNames       map[UnionDeclarationID]string // authored OneOf to exact name
		importNames      map[string]string             // import path to final package name
		generatedImports map[string]string             // generated relative path to final package name
		counts           map[string]int                // raw type name to occurrence count
		frozen           bool                          // true after this set rejects new names
	}

	// Hasher is the interface implemented by the objects that must be
	// scoped.
	Hasher interface {
		// Hash computes a unique instance hash suitable for indexing
		// in a map.
		Hash() string
	}

	// Scoper provides a scope for generating unique names.
	Scoper interface {
		Scope() *NameScope
	}
)

// NewNameScope creates an empty name scope.
func NewNameScope() *NameScope {
	return &NameScope{
		names:            make(map[string]string),
		userTypeNames:    make(map[expr.UserType]string),
		unionNames:       make(map[UnionDeclarationID]string),
		importNames:      make(map[string]string),
		generatedImports: make(map[string]string),
		counts:           make(map[string]int),
	}
}

// Fork returns a new scope containing every lookup key and name already
// recorded in s. The new scope can add private helper names without changing s
// or colliding with names already chosen there.
func (s *NameScope) Fork() *NameScope {
	fork := NewNameScope()
	for hash, name := range s.names {
		fork.names[hash] = name
	}
	for origin, name := range s.userTypeNames {
		fork.userTypeNames[origin] = name
	}
	for identity, name := range s.unionNames {
		fork.unionNames[identity] = name
	}
	for importPath, name := range s.importNames {
		fork.importNames[importPath] = name
	}
	for importPath, name := range s.generatedImports {
		fork.generatedImports[importPath] = name
	}
	for name, count := range s.counts {
		fork.counts[name] = count
	}
	return fork
}

// bindImport makes type rendering use the package name selected for an import
// path. Generated packages call it before the scope is frozen.
func (s *NameScope) bindImport(importPath, name string) {
	if s.frozen {
		panic("cannot bind an import name in a frozen name scope")
	}
	if existing, ok := s.importNames[importPath]; ok && existing != name {
		panic(fmt.Sprintf("import path %q is already bound to package name %q", importPath, existing))
	}
	s.importNames[importPath] = name
}

// bindGeneratedImport records the final name for a path relative to the
// generated module root.
func (s *NameScope) bindGeneratedImport(importPath, name string) {
	if s.frozen {
		panic("cannot bind a generated import name in a frozen name scope")
	}
	if existing, ok := s.generatedImports[importPath]; ok && existing != name {
		panic(fmt.Sprintf("generated import path %q is already bound to package name %q", importPath, existing))
	}
	s.generatedImports[importPath] = name
}

// HashedUnique builds the unique name for key using name and - if not unique -
// appending suffix and - if still not unique - a counter value. It returns
// the same value when called multiple times for a key returning the same hash.
func (s *NameScope) HashedUnique(key Hasher, name string, suffix ...string) string {
	hash := key.Hash()
	if n, ok := s.names[hash]; ok {
		return n
	}
	if s.frozen {
		panic("cannot reserve a new hashed name in a frozen name scope")
	}
	name = s.Unique(name, suffix...)
	s.names[hash] = name
	return name
}

// Unique returns a unique name for the given name. A suffix is appended to the
// name if given name is not unique. If suffixed name is still not unique, a
// counter value is added to the suffixed name until unique. The returned name
// is reserved in the scope.
func (s *NameScope) Unique(name string, suffix ...string) string {
	if s.frozen {
		panic("cannot reserve a name in a frozen name scope")
	}
	ret := s.PeekUnique(name, suffix...)
	s.counts[ret]++
	return ret
}

// Freeze prevents the scope from reserving new names. Names already associated
// with hashes remain readable through HashedUnique and type-reference methods.
func (s *NameScope) Freeze() {
	s.frozen = true
}

// bind makes key return an already reserved Go name without reserving another
// name. Generated packages call it while Generation.Freeze assigns final
// declaration names.
func (s *NameScope) bind(key Hasher, name string) {
	if s.frozen {
		panic("cannot bind a hashed name in a frozen name scope")
	}
	hash := key.Hash()
	if existing, ok := s.names[hash]; ok && existing != name {
		panic(fmt.Sprintf("hash %q is already bound to package name %q", hash, existing))
	}
	if _, ok := s.counts[name]; !ok {
		panic(fmt.Sprintf("package name %q must be reserved before hash binding", name))
	}
	s.names[hash] = name
}

// bindUserType makes a user type and copies made from it return the same exact
// generated name.
func (s *NameScope) bindUserType(userType expr.UserType, name string) {
	if s.frozen {
		panic("cannot bind a user type name in a frozen name scope")
	}
	origin := userType.Origin()
	if existing, ok := s.userTypeNames[origin]; ok && existing != name {
		panic(fmt.Sprintf("user type %q is already bound to package name %q", origin.Name(), existing))
	}
	if _, ok := s.counts[name]; !ok {
		panic(fmt.Sprintf("package name %q must be reserved before user type binding", name))
	}
	s.userTypeNames[origin] = name
}

// bindUnion makes copies of one authored OneOf return its exact generated name.
func (s *NameScope) bindUnion(identity UnionDeclarationID, name string) {
	if s.frozen {
		panic("cannot bind a union name in a frozen name scope")
	}
	if existing, ok := s.unionNames[identity]; ok && existing != name {
		panic(fmt.Sprintf("union declaration is already bound to package name %q", existing))
	}
	if _, ok := s.counts[name]; !ok {
		panic(fmt.Sprintf("package name %q must be reserved before union binding", name))
	}
	s.unionNames[identity] = name
}

// PeekUnique returns the name that Unique would return for the same inputs,
// without mutating the scope.
//
// This is useful when synthesizing type names or identifiers that are later
// reserved via Hash-based naming (e.g., GoTypeName/HashedUnique) and therefore
// must not increment the scope counters twice.
func (s *NameScope) PeekUnique(name string, suffix ...string) string {
	c, ok := s.counts[name]
	if !ok {
		return name
	}
	if len(suffix) > 0 {
		name += suffix[0]
		c, ok = s.counts[name]
		if !ok {
			return name
		}
	}
	for i := c; ; i++ {
		ret := name + strconv.Itoa(i+1)
		if _, ok := s.counts[ret]; !ok {
			return ret
		}
	}
}

// Name returns the unique name the scope would currently assign to the given
// name: the name itself when unused, otherwise the name with its next counter
// value appended. Unlike Unique it does not reserve the returned name.
func (s *NameScope) Name(name string) string {
	i, ok := s.counts[name]
	if !ok {
		return name
	}
	name += strconv.Itoa(i + 1)
	return name
}

// GoTypeDef returns the Go code that defines a Go type which matches the data
// structure definition (the part that comes after `type foo`).
//
// ptr if true indicates that the attribute must be stored in a pointer
// (except array and map types which are always non-pointers)
//
// useDefault if true indicates that the attribute must not be a pointer
// if it has a default value.
func (s *NameScope) GoTypeDef(att *expr.AttributeExpr, ptr, useDefault bool) string {
	pkg := ""
	if loc := UserTypeLocation(att.Type); loc != nil {
		pkg = s.generatedImportName(loc.RelImportPath, loc.PackageName())
	} else if p, ok := att.Meta.Last("struct:pkg:path"); ok && p != "" {
		pkg = s.generatedImportName(p, Goify(filepath.Base(p), false))
	}
	return s.goTypeDefWithPkgOverride(att, ptr, useDefault, pkg, "")
}

// goTypeDefWithPkgOverride generates the Go type definition string for the attribute.
// When targetPkg is not empty, user types referenced inside inline structs are
// qualified against targetPkg unless they are defined in a different package,
// in which case their own package is used. When targetPkg is empty, the
// package qualification falls back to pkg and the user type location.
func (s *NameScope) goTypeDefWithPkgOverride(att *expr.AttributeExpr, ptr, useDefault bool, pkg, targetPkg string) string {
	switch actual := att.Type.(type) {
	case expr.Primitive:
		if typeName := s.metaTypeName(att); typeName != "" {
			return typeName
		}
		return GoNativeTypeName(actual)
	case *expr.Array:
		d := s.goTypeDefWithPkgOverride(actual.ElemType, ptr, useDefault, pkg, targetPkg)
		if expr.IsObject(actual.ElemType.Type) {
			d = "*" + d
		}
		return "[]" + d
	case *expr.Map:
		keyDef := s.goTypeDefWithPkgOverride(actual.KeyType, ptr, useDefault, pkg, targetPkg)
		if expr.IsObject(actual.KeyType.Type) {
			keyDef = "*" + keyDef
		}
		elemDef := s.goTypeDefWithPkgOverride(actual.ElemType, ptr, useDefault, pkg, targetPkg)
		if expr.IsObject(actual.ElemType.Type) {
			elemDef = "*" + elemDef
		}
		return fmt.Sprintf("map[%s]%s", keyDef, elemDef)
	case *expr.Union:
		// Unions are generated as named sum-type structs. Refer to the named type
		// here; the concrete definition is emitted separately by the service
		// code generator.
		if targetPkg != "" {
			return s.GoFullTypeName(att, targetPkg)
		}
		return s.GoFullTypeName(att, "")
	case *expr.Object:
		ss := []string{"struct {"}
		for _, nat := range *actual {
			var (
				fn   string
				tdef string
				desc string
				tags string

				name = nat.Name
				at   = nat.Attribute
			)
			{
				fn = GoifyAtt(at, name, true)
				tdef = s.goTypeDefWithPkgOverride(at, ptr, useDefault, pkg, targetPkg)
				if goFieldIsPointer(att, name, ptr, useDefault) {
					tdef = "*" + tdef
				}
				if at.Description != "" {
					desc = Comment(at.Description) + "\n\t"
				}
				tags = AttributeTagsWithName(att, name, at)
			}
			ss = append(ss, fmt.Sprintf("\t%s%s %s%s", desc, fn, tdef, tags))
		}
		ss = append(ss, "}")
		return strings.Join(ss, "\n")
	case expr.UserType:
		if actual == expr.Empty {
			return "struct {}"
		}
		var referencedPkg string
		if loc := UserTypeLocation(actual); loc != nil {
			if targetPkg != "" || loc.PackageName() != pkg {
				referencedPkg = s.generatedImportName(loc.RelImportPath, loc.PackageName())
			}
		} else if targetPkg != "" {
			referencedPkg = targetPkg
		}
		return s.scopedTypeName(actual, Goify(actual.Name(), true), referencedPkg)
	default:
		panic(fmt.Sprintf("unknown data type %T", actual)) // bug
	}
}

// GoVar returns the Go code that returns the address of a variable of the Go type
// which matches the given attribute type.
func (*NameScope) GoVar(varName string, dt expr.DataType) string {
	// For a raw struct, no need to indirecting
	if isRawStruct(dt) {
		return varName
	}
	return "&" + varName
}

// GoTypeRef returns the Go code that refers to the Go type which matches the
// given attribute type.
func (s *NameScope) GoTypeRef(att *expr.AttributeExpr) string {
	name := s.GoTypeName(att)
	return goTypeRef(name, att.Type)
}

// GoTypeRefWithDefaults returns the Go code that refers to the Go type which
// matches the given attribute type. The result of this function differs from
// GoTypeRef when the attribute type is an object (note: not a user type) and
// the reference is thus an inline struct definition. In this case accounting
// for default values may cause child attributes to use non-pointer fields.
func (s *NameScope) GoTypeRefWithDefaults(att *expr.AttributeExpr) string {
	name := s.GoTypeNameWithDefaults(att)
	return goTypeRef(name, att.Type)
}

// GoFullTypeRef returns the Go code that refers to the Go type which matches
// the given attribute type defined in the given package if a user type.
func (s *NameScope) GoFullTypeRef(att *expr.AttributeExpr, pkg string) string {
	name := s.GoFullTypeName(att, pkg)
	return goTypeRef(name, att.Type)
}

// GoTypeName returns the Go type name of the given attribute type.
func (s *NameScope) GoTypeName(att *expr.AttributeExpr) string {
	return s.GoFullTypeName(att, "")
}

// GoTypeNameWithDefaults returns the Go type name of the given attribute type.
// The result of this function differs from GoTypeName when the attribute type
// is an object (note: not a user type) and the name is thus an inline struct
// definition. In this case accounting for default values may cause child
// attributes to use non-pointer fields.
func (s *NameScope) GoTypeNameWithDefaults(att *expr.AttributeExpr) string {
	if _, ok := att.Type.(*expr.Object); ok {
		return s.GoTypeDef(att, false, true)
	}
	return s.GoTypeName(att)
}

// GoFullTypeName returns the Go type name of the given data type qualified with
// the given package name if applicable and if not the empty string.
func (s *NameScope) GoFullTypeName(att *expr.AttributeExpr, pkg string) string {
	switch actual := att.Type.(type) {
	case expr.Primitive:
		if typeName := s.metaTypeName(att); typeName != "" {
			return typeName
		}
		return GoNativeTypeName(actual)
	case *expr.Array:
		return "[]" + s.GoFullTypeRef(actual.ElemType, s.pkgWithDefault(actual.ElemType.Type, pkg))
	case *expr.Map:
		return fmt.Sprintf("map[%s]%s",
			s.GoFullTypeRef(actual.KeyType, s.pkgWithDefault(actual.KeyType.Type, pkg)),
			s.GoFullTypeRef(actual.ElemType, s.pkgWithDefault(actual.ElemType.Type, pkg)))
	case *expr.Object:
		return s.GoTypeDef(att, false, false)
	case expr.UserType:
		if expr.IsErrorResult(actual) {
			return "goa.ServiceError"
		}
		return s.scopedTypeName(actual, Goify(actual.Name(), true), pkg)
	case *expr.Union:
		return s.scopedUnionTypeName(NewUnionDeclarationID(att), Goify(actual.Name(), true), pkg)
	case expr.CompositeExpr:
		return s.GoFullTypeName(actual.Attribute(), s.pkgWithDefault(actual.Attribute().Type, pkg))
	default:
		panic(fmt.Sprintf("unknown data type %T", actual)) // bug
	}
}

// scopedUnionTypeName returns the declaration recorded for one authored OneOf
// occurrence. A package qualifier is added when the caller names another Go
// package.
func (s *NameScope) scopedUnionTypeName(identity UnionDeclarationID, base, pkg string) string {
	name := base
	if selected, ok := s.unionNames[identity]; ok {
		name = selected
	} else if s.frozen {
		panic(fmt.Sprintf("union %q was not declared before the package name scope was frozen", base))
	}
	if pkg == "" {
		return name
	}
	return pkg + "." + name
}

// scopedTypeName returns the generated name recorded for userType. A copied
// user type first uses the name of its original declaration. Types without an
// explicit declaration binding keep the existing Hash-based lookup.
func (s *NameScope) scopedTypeName(userType expr.UserType, base, pkg string) string {
	if name, ok := s.userTypeNames[userType.Origin()]; ok {
		if pkg == "" {
			return name
		}
		return pkg + "." + name
	}
	if pkg == "" {
		return s.HashedUnique(userType, base, "")
	}
	if name, ok := s.names[userType.Hash()]; ok {
		return pkg + "." + name
	}
	return pkg + "." + base
}

// pkgWithDefault returns the package defining the given type. If the types is a
// user type with "struct:pkg:path" metadata then it returns the corresponding
// value, otherwise it returns pkg.
func (s *NameScope) pkgWithDefault(dt expr.DataType, pkg string) string {
	if loc := UserTypeLocation(dt); loc != nil {
		return s.generatedImportName(loc.RelImportPath, loc.PackageName())
	}
	return pkg
}

// importName returns the final package name for importPath when the generated
// package planned that import.
func (s *NameScope) importName(importPath, fallback string) string {
	if name := s.importNames[importPath]; name != "" {
		return name
	}
	return fallback
}

// generatedImportName returns the final package name for a path relative to
// the generated module root.
func (s *NameScope) generatedImportName(importPath, fallback string) string {
	if name := s.generatedImports[path.Clean(importPath)]; name != "" {
		return name
	}
	return s.importName(importPath, fallback)
}

// metaTypeName applies the final package name to a type named by
// struct:field:type metadata.
func (s *NameScope) metaTypeName(attribute *expr.AttributeExpr) string {
	typeName, spec := GetMetaType(attribute)
	if typeName == "" || spec == nil {
		return typeName
	}
	preferred := spec.preferredName()
	selected := s.importName(spec.Path, preferred)
	return rebindMetaTypeQualifier(typeName, preferred, selected)
}

func goTypeRef(name string, dt expr.DataType) string {
	// For a raw struct, no need to dereference
	if isRawStruct(dt) {
		return name
	}
	return "*" + name
}

func isRawStruct(dt expr.DataType) bool {
	if _, ok := dt.(*expr.Object); ok {
		return true
	}
	if expr.IsObject(dt) {
		return false
	}
	if expr.IsUnion(dt) {
		return false
	}
	return true
}
