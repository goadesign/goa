// This file records the packages named by one generated source contribution.
// A generator records complete paths before names are chosen, then reads the
// final import declarations after the generation is frozen. Contributions
// that share a file merge their completed import declarations.
package codegen

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"goa.design/goa/v3/expr"
)

type (
	// GeneratedImportPlan stores the packages named by one source contribution.
	// Every package path must be added before Generation.Freeze.
	GeneratedImportPlan struct {
		output  *GeneratedPackage
		paths   map[string]struct{}
		imports []*ImportSpec
		linked  bool
	}
)

// NewGeneratedImportPlan creates an empty import plan for source written in
// output.
func NewGeneratedImportPlan(output *GeneratedPackage) *GeneratedImportPlan {
	if output == nil {
		panic("generated import plan requires an output package")
	}
	return &GeneratedImportPlan{
		output: output,
		paths:  make(map[string]struct{}),
	}
}

// Require adds packages whose names are written directly by generated source.
func (p *GeneratedImportPlan) Require(specs ...*ImportSpec) error {
	return p.add(specs, p.output.RequireImport)
}

// AddGenerated adds generated packages referenced by this contribution.
func (p *GeneratedImportPlan) AddGenerated(specs ...*ImportSpec) error {
	return p.add(specs, p.output.ReserveGeneratedImport)
}

// AddDesign adds packages named by design metadata referenced by this
// contribution.
func (p *GeneratedImportPlan) AddDesign(specs ...*ImportSpec) error {
	return p.add(specs, p.output.DeclareImport)
}

// AddTypeExpressions adds custom and relocated packages visible in written Go
// type expressions. A named type adds its package but not packages used only
// by its separately written fields.
func (p *GeneratedImportPlan) AddTypeExpressions(attributes ...*expr.AttributeExpr) error {
	if err := p.ensurePlanning(); err != nil {
		return err
	}
	for _, attribute := range attributes {
		if err := walkAttributeImports(p.output, attribute, false, nil, p.addAttributeImport); err != nil {
			return err
		}
	}
	return nil
}

// AddRecursiveTypeReferences adds packages named while generated code reads,
// converts, or validates nested fields. It descends into named types because
// those fields are referenced by the emitted code.
func (p *GeneratedImportPlan) AddRecursiveTypeReferences(attributes ...*expr.AttributeExpr) error {
	if err := p.ensurePlanning(); err != nil {
		return err
	}
	seen := make(map[expr.UserType]struct{})
	for _, attribute := range attributes {
		if err := walkAttributeImports(p.output, attribute, true, seen, p.addAttributeImport); err != nil {
			return err
		}
	}
	return nil
}

// Paths returns independent copies of the complete package paths used by this
// contribution. It can compare plans before their package names are fixed.
func (p *GeneratedImportPlan) Paths() []string {
	paths := make([]string, 0, len(p.paths))
	for importPath := range p.paths {
		paths = append(paths, importPath)
	}
	sort.Strings(paths)
	return paths
}

// Link resolves every recorded path to the package name selected by
// Generation.Freeze.
func (p *GeneratedImportPlan) Link() error {
	if p.linked {
		return fmt.Errorf("generated import plan for %q linked more than once", p.output.path)
	}
	if !p.output.frozen {
		return fmt.Errorf("generated import plan for %q linked before freeze", p.output.path)
	}
	paths := p.Paths()
	p.imports = make([]*ImportSpec, len(paths))
	for index, importPath := range paths {
		p.imports[index] = p.output.Import(importPath)
	}
	p.linked = true
	return nil
}

// Imports returns independent copies of the final import declarations. It
// panics when Link has not completed.
func (p *GeneratedImportPlan) Imports() []*ImportSpec {
	if !p.linked {
		panic("generated imports requested before link")
	}
	imports := make([]*ImportSpec, len(p.imports))
	for index, spec := range p.imports {
		copy := *spec
		imports[index] = &copy
	}
	return imports
}

// add registers each import preference on the output package and saves its
// complete path for this contribution.
func (p *GeneratedImportPlan) add(specs []*ImportSpec, register func(*ImportSpec) error) error {
	if err := p.ensurePlanning(); err != nil {
		return err
	}
	for _, spec := range specs {
		if spec == nil || spec.Path == "" || spec.Path == p.output.path {
			continue
		}
		if err := register(spec); err != nil {
			return err
		}
		p.paths[spec.Path] = struct{}{}
	}
	return nil
}

// ensurePlanning rejects changes after package and import names are fixed.
func (p *GeneratedImportPlan) ensurePlanning() error {
	if p.output.frozen {
		return fmt.Errorf("generated import plan for %q changed after freeze", p.output.path)
	}
	return nil
}

// addAttributeImport registers one package found in a generated type shape.
func (p *GeneratedImportPlan) addAttributeImport(spec *ImportSpec, generated bool) error {
	if generated {
		return p.AddGenerated(spec)
	}
	return p.AddDesign(spec)
}

// walkAttributeImports visits packages named by one generated type shape.
func walkAttributeImports(
	output *GeneratedPackage,
	attribute *expr.AttributeExpr,
	expandNamed bool,
	seen map[expr.UserType]struct{},
	visit func(*ImportSpec, bool) error,
) error {
	if attribute == nil || attribute.Type == nil || attribute.Type == expr.Empty {
		return nil
	}
	if _, spec := GetMetaType(attribute); spec != nil && spec.Path != output.path {
		if err := visit(spec, false); err != nil {
			return err
		}
	}
	switch actual := attribute.Type.(type) {
	case expr.UserType:
		if expr.IsErrorResult(actual) {
			return visit(GoaImport(""), false)
		}
		if location := UserTypeLocation(actual); location != nil {
			importPath := path.Join(output.genpkg, location.RelImportPath)
			if importPath != output.path {
				if err := visit(NewImport(
					strings.ToLower(Goify(path.Base(importPath), false)),
					importPath,
				), true); err != nil {
					return err
				}
			}
		}
		if !expandNamed {
			return nil
		}
		origin := actual.Origin()
		if _, exists := seen[origin]; exists {
			return nil
		}
		seen[origin] = struct{}{}
		defer delete(seen, origin)
		return walkAttributeImports(output, actual.Attribute(), true, seen, visit)
	case *expr.Object:
		for _, field := range *actual {
			if err := walkAttributeImports(output, field.Attribute, expandNamed, seen, visit); err != nil {
				return err
			}
		}
	case *expr.Array:
		return walkAttributeImports(output, actual.ElemType, expandNamed, seen, visit)
	case *expr.Map:
		if err := walkAttributeImports(output, actual.KeyType, expandNamed, seen, visit); err != nil {
			return err
		}
		return walkAttributeImports(output, actual.ElemType, expandNamed, seen, visit)
	case *expr.Union:
		for _, branch := range actual.Values {
			if err := walkAttributeImports(output, branch.Attribute, expandNamed, seen, visit); err != nil {
				return err
			}
		}
	}
	return nil
}
