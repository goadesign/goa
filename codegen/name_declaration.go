// This file defines the canonical package-level Go name shared by declaration
// planning and rendering. Generated packages allocate exact names before
// compiler-preferred names and freeze each record before source rendering.
package codegen

import (
	"fmt"
	"strings"
)

type (
	// PackageNameKind identifies the Go declaration category for diagnostics.
	// Types, functions, constants, and variables still share one package namespace.
	PackageNameKind uint8

	// PackageNameOrder supplies deterministic ordering for preferred names in
	// one subsystem-owned declaration family. Implementations compare only
	// values with the same PackageNameFamily.
	PackageNameOrder interface {
		PackageNameFamily() string
		ComparePackageName(PackageNameOrder) int
	}

	// NameDeclaration records one package-level Go identifier. Its final name is
	// unavailable until the owning generation freezes.
	NameDeclaration struct {
		kind        PackageNameKind
		preferred   string
		final       string
		packagePath string
		exact       bool
		order       PackageNameOrder
		base        *NameDeclaration
		prefix      string
		suffix      string
		hashes      []Hasher
		frozen      bool
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
// identifier may receive a deterministic numeric suffix.
func NewPreferredName(kind PackageNameKind, preferred string, order PackageNameOrder) *NameDeclaration {
	if order == nil {
		panic("preferred package name requires stable ordering")
	}
	return &NameDeclaration{
		kind:      kind,
		preferred: Goify(preferred, true),
		order:     order,
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

// PreferredName returns the unsuffixed Go identifier requested during planning.
func (d *NameDeclaration) PreferredName() string {
	return d.preferredName()
}

// Kind returns the declaration category used for collision diagnostics.
func (d *NameDeclaration) Kind() PackageNameKind {
	return d.kind
}

// PackagePath returns the generated import path that owns the declaration. It
// is empty until a generated package accepts the record.
func (d *NameDeclaration) PackagePath() string {
	return d.packagePath
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
	if order == nil {
		panic("dependent package name requires stable ordering")
	}
	return &NameDeclaration{
		kind:   kind,
		order:  order,
		base:   base,
		prefix: prefix,
		suffix: suffix,
	}
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

// comparePackageNames orders independent records without consulting discovery
// order. Equal ordering facts for distinct records are a planning error.
func comparePackageNames(left, right *NameDeclaration) int {
	if compared := strings.Compare(left.order.PackageNameFamily(), right.order.PackageNameFamily()); compared != 0 {
		return compared
	}
	return left.order.ComparePackageName(right.order)
}
