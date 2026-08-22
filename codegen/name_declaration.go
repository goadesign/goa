// This file defines the canonical package-level Go name shared by declaration
// planning and rendering. Generated packages allocate exact names before
// compiler-preferred names and freeze each record before source rendering.
package codegen

import (
	"fmt"
	"reflect"
	"strings"
)

type (
	// PackageNameKind identifies the Go declaration category for diagnostics.
	// Types, functions, constants, and variables still share one package namespace.
	PackageNameKind uint8

	// PackageNameVisibility specifies whether a preferred generated declaration
	// is visible outside its Go package.
	PackageNameVisibility uint8

	// PackageNameOrder supplies a deterministic total order for preferred names
	// in one subsystem-owned declaration family. Implementations must be named,
	// non-pointer value types whose fields recursively contain immutable values.
	// The package catalog compares only values of the same concrete type.
	PackageNameOrder interface {
		// ComparePackageName compares two values of the same concrete type. It
		// must return a negative value when the receiver sorts first, zero only
		// when both values contain identical stable ordering facts, and a positive
		// value when the receiver sorts last. Its sign must be antisymmetric, and
		// its less-than relation must be transitive.
		ComparePackageName(PackageNameOrder) int
	}

	// NameDeclaration records one package-level Go identifier. Its final name is
	// unavailable until the owning generation freezes.
	NameDeclaration struct {
		kind       PackageNameKind
		visibility PackageNameVisibility
		preferred  string
		final      string
		owner      *GeneratedPackage
		exact      bool
		order      PackageNameOrder
		base       *NameDeclaration
		prefix     string
		suffix     string
		hashes     []Hasher
		frozen     bool
	}
)

const (
	// NameType identifies a package-level type declaration.
	NameType PackageNameKind = iota + 1
	// NameFunction identifies a package-level function declaration.
	NameFunction
	// NameConstant identifies a package-level constant declaration.
	NameConstant
	// NameVariable identifies a package-level variable declaration.
	NameVariable
)

const (
	// ExportedName makes the preferred generated identifier package-visible.
	ExportedName PackageNameVisibility = iota + 1
	// UnexportedName keeps the preferred generated identifier package-private.
	UnexportedName
)

// NewExactName creates an authored or external declaration whose exported Go
// identifier must not change. The owning generated package rejects collisions.
func NewExactName(kind PackageNameKind, preferred string) *NameDeclaration {
	return &NameDeclaration{
		kind:      kind,
		preferred: Goify(preferred, true),
		exact:     true,
	}
}

// NewPreferredName creates a compiler-owned declaration whose preferred Go
// identifier may receive a deterministic numeric suffix. order must be a
// named, non-pointer value whose fields recursively contain immutable values;
// the owning package validates that constraint when it accepts the record.
func NewPreferredName(kind PackageNameKind, preferred string, visibility PackageNameVisibility, order PackageNameOrder) *NameDeclaration {
	return &NameDeclaration{
		kind:       kind,
		visibility: visibility,
		preferred:  Goify(preferred, visibility == ExportedName),
		order:      order,
	}
}

// Name returns the frozen Go identifier. It panics before the owning
// generation freezes because no renderer may observe a provisional spelling.
func (d *NameDeclaration) Name() string {
	if !d.frozen {
		panic(fmt.Sprintf("package name %q requested before generation freeze", d.preferredName()))
	}
	return d.final
}

// Kind returns the declaration category used for collision diagnostics.
func (d *NameDeclaration) Kind() PackageNameKind {
	return d.kind
}

// String returns the declaration category used in planning errors.
func (k PackageNameKind) String() string {
	switch k {
	case NameType:
		return "type"
	case NameFunction:
		return "function"
	case NameConstant:
		return "constant"
	case NameVariable:
		return "variable"
	default:
		return "unknown"
	}
}

// newDependentName creates a compiler-owned name whose preferred spelling is
// derived from another canonical declaration after that declaration freezes.
func newDependentName(kind PackageNameKind, base *NameDeclaration, prefix, suffix string, order PackageNameOrder) *NameDeclaration {
	if base == nil {
		panic("dependent package name requires a base declaration")
	}
	return &NameDeclaration{
		kind:   kind,
		order:  order,
		base:   base,
		prefix: prefix,
		suffix: suffix,
	}
}

// comparePackageNames orders independent records without consulting discovery
// order. Equal ordering facts for distinct records are a planning error.
func comparePackageNames(left, right *NameDeclaration) int {
	leftType := reflect.TypeOf(left.order)
	rightType := reflect.TypeOf(right.order)
	if compared := strings.Compare(leftType.PkgPath(), rightType.PkgPath()); compared != 0 {
		return compared
	}
	if compared := strings.Compare(leftType.Name(), rightType.Name()); compared != 0 {
		return compared
	}
	return left.order.ComparePackageName(right.order)
}

// validateNameDeclaration rejects records that cannot identify a package-level
// Go declaration before the owning package changes its declaration catalog.
func validateNameDeclaration(declaration *NameDeclaration) error {
	if !declaration.kind.valid() {
		return fmt.Errorf("invalid package name kind %d", declaration.kind)
	}
	if declaration.base == nil && !declaration.exact && !declaration.visibility.valid() {
		return fmt.Errorf("invalid package name visibility %d", declaration.visibility)
	}
	if declaration.preferredName() == "" {
		return fmt.Errorf("package name must not be empty")
	}
	return nil
}

// valid reports whether visibility is represented by the preferred-name catalog.
func (v PackageNameVisibility) valid() bool {
	return v == ExportedName || v == UnexportedName
}

// validatePackageNameOrder rejects ordering values whose identity or contents
// can change after collection. A named value type gives independent generators
// a stable family identity without coordinating through caller-chosen strings.
func validatePackageNameOrder(order PackageNameOrder) error {
	if order == nil {
		return fmt.Errorf("package name order must be a stable concrete named value type")
	}
	typeOf := reflect.TypeOf(order)
	if typeOf.Name() == "" || typeOf.PkgPath() == "" || !isStablePackageNameOrderType(typeOf) {
		return fmt.Errorf("package name order %T must be a stable concrete named value type", order)
	}
	return nil
}

// isStablePackageNameOrderType reports whether values of typeOf contain only
// immutable value fields suitable for deterministic comparison after freeze.
func isStablePackageNameOrderType(typeOf reflect.Type) bool {
	switch typeOf.Kind() {
	case reflect.Array:
		return isStablePackageNameOrderType(typeOf.Elem())
	case reflect.Struct:
		for i := range typeOf.NumField() {
			if !isStablePackageNameOrderType(typeOf.Field(i).Type) {
				return false
			}
		}
		return true
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128,
		reflect.String:
		return true
	default:
		return false
	}
}

// packagePath returns the generated import path that owns the declaration.
// Access before package collection is an internal planning bug.
func (d *NameDeclaration) packagePath() string {
	if d.owner == nil {
		panic(fmt.Sprintf("package name %q has no generated package owner", d.preferredName()))
	}
	return d.owner.path
}

// preferredName returns the requested name, using the base declaration's
// frozen spelling for linked declaration families such as union constructors.
func (d *NameDeclaration) preferredName() string {
	if d.base == nil {
		return d.preferred
	}
	base := d.base.preferred
	if d.base.frozen {
		base = d.base.final
	}
	return d.prefix + base + d.suffix
}

// valid reports whether the category is represented by this catalog.
func (k PackageNameKind) valid() bool {
	switch k {
	case NameType, NameFunction, NameConstant, NameVariable:
		return true
	default:
		return false
	}
}
