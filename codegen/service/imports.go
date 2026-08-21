// This file plans one immutable import alias per complete package path, then
// computes the exact subset of those imports used by each generated service
// file. Qualified references and import declarations share these bindings.
package service

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

type (
	// importAliases is the frozen render-model binding from complete import paths
	// to their unique Go qualifiers.
	importAliases struct {
		bindings map[string]importBinding
	}

	// importBinding records the allocated qualifier and whether the import must
	// spell it explicitly in a generated header.
	importBinding struct {
		name      string
		preferred string
		explicit  bool
	}

	// importAliasCandidate records one package path before deterministic alias
	// allocation. Fixed generator imports receive priority over design metadata.
	importAliasCandidate struct {
		preferred string
		explicit  bool
		fixed     bool
	}

	// importAliasPlan collects all package paths before any render string is
	// produced.
	importAliasPlan struct {
		candidates map[string]importAliasCandidate
	}

	// importCollector accumulates the imports referenced by one generated Go
	// file while traversing recursive service type definitions.
	importCollector struct {
		aliases       *importAliases
		genpkg        string
		outputPackage string
		paths         map[string]struct{}
		legacy        map[string]*codegen.ImportSpec
	}
)

// AttributeImports returns the generated-type and struct:field:type imports
// referenced by attributes using their preferred, unshared aliases. Transport
// generators retain this contract until their service-side references move to
// the declaration resolver in Task 5.
func AttributeImports(genpkg, outputPackage string, attributes ...*expr.AttributeExpr) []*codegen.ImportSpec {
	collector := newImportCollector(nil, genpkg, outputPackage)
	for _, attribute := range attributes {
		collector.collect(attribute)
	}
	return collector.imports()
}

// AttributeImports returns the exact generated-type and metadata imports
// referenced by attributes using the frozen aliases shared with service type
// references.
func (d *ServicesData) AttributeImports(outputPackage string, attributes ...*expr.AttributeExpr) []*codegen.ImportSpec {
	collector := newImportCollector(d.aliases, d.generation.GenPkg, outputPackage)
	for _, attribute := range attributes {
		collector.collect(attribute)
	}
	return collector.imports()
}

// serviceReferenceAttributes returns the method and error attributes whose
// named declarations are referenced by service, endpoint, and client files.
func serviceReferenceAttributes(service *expr.ServiceExpr) []*expr.AttributeExpr {
	attributes := make([]*expr.AttributeExpr, 0, len(service.Methods)*4+len(service.Errors))
	for _, serviceError := range service.Errors {
		attributes = append(attributes, serviceError.AttributeExpr)
	}
	for _, method := range service.Methods {
		attributes = append(attributes, method.Payload, method.StreamingPayload, method.Result)
		if method.HasMixedResults() {
			attributes = append(attributes, method.StreamingResult)
		}
		for _, methodError := range method.Errors {
			attributes = append(attributes, methodError.AttributeExpr)
		}
	}
	return attributes
}

// newImportAliases scans every participating design root plus the explicit
// root being rendered, then freezes deterministic aliases before service
// analysis creates type-reference strings.
func newImportAliases(root *expr.RootExpr, generation *codegen.Generation) (*importAliases, error) {
	plan := &importAliasPlan{candidates: make(map[string]importAliasCandidate)}
	if err := plan.addFixedImports(); err != nil {
		return nil, err
	}
	seenRoots := make(map[*expr.RootExpr]struct{}, len(generation.Roots)+1)
	for _, evaluated := range generation.Roots {
		design, ok := evaluated.(*expr.RootExpr)
		if !ok {
			continue
		}
		seenRoots[design] = struct{}{}
		if err := plan.addRoot(design, generation.GenPkg); err != nil {
			return nil, err
		}
	}
	if _, ok := seenRoots[root]; !ok {
		if err := plan.addRoot(root, generation.GenPkg); err != nil {
			return nil, err
		}
	}
	return plan.freeze(), nil
}

// addFixedImports reserves qualifiers used directly by service and view
// templates before metadata-selected packages compete for those names.
func (p *importAliasPlan) addFixedImports() error {
	fixed := []*codegen.ImportSpec{
		codegen.SimpleImport("bytes"),
		codegen.SimpleImport("context"),
		codegen.SimpleImport("encoding/json"),
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("io"),
		codegen.SimpleImport("unicode/utf8"),
		codegen.GoaImport(""),
		codegen.GoaImport("security"),
	}
	for _, spec := range fixed {
		if err := p.add(spec.Path, spec.Name, spec.Name != "", true); err != nil {
			return err
		}
	}
	return nil
}

// addRoot collects generated package locations and metadata imports reachable
// from one complete service design.
func (p *importAliasPlan) addRoot(root *expr.RootExpr, genpkg string) error {
	for _, service := range root.Services {
		servicePath := servicePackagePath(genpkg, service)
		serviceName := strings.ToLower(codegen.Goify(service.Name, false))
		if err := p.add(servicePath, serviceName, true, false); err != nil {
			return err
		}
		if err := p.add(servicePath+"/views", serviceName+"views", true, false); err != nil {
			return err
		}
	}
	seen := make(map[expr.UserType]struct{})
	for _, userType := range root.Types {
		if err := p.addAttribute(&expr.AttributeExpr{Type: userType}, genpkg, seen); err != nil {
			return err
		}
	}
	for _, resultType := range root.ResultTypes {
		if err := p.addAttribute(&expr.AttributeExpr{Type: resultType}, genpkg, seen); err != nil {
			return err
		}
	}
	for _, service := range root.Services {
		for _, attribute := range serviceReferenceAttributes(service) {
			if err := p.addAttribute(attribute, genpkg, seen); err != nil {
				return err
			}
		}
	}
	return nil
}

// addAttribute recursively records every explicit generated location and
// struct:field:type import reachable from attribute.
func (p *importAliasPlan) addAttribute(attribute *expr.AttributeExpr, genpkg string, seen map[expr.UserType]struct{}) error {
	if attribute == nil || attribute.Type == expr.Empty {
		return nil
	}
	if _, spec := codegen.GetMetaType(attribute); spec != nil {
		if err := p.add(spec.Path, spec.Name, spec.Name != "", false); err != nil {
			return err
		}
	}
	switch actual := attribute.Type.(type) {
	case expr.UserType:
		if location := codegen.UserTypeLocation(actual); location != nil {
			if err := p.add(
				path.Join(genpkg, location.RelImportPath),
				location.PackageName(),
				true,
				false,
			); err != nil {
				return err
			}
		}
		origin := actual.Origin()
		if _, ok := seen[origin]; ok {
			return nil
		}
		seen[origin] = struct{}{}
		return p.addAttribute(actual.Attribute(), genpkg, seen)
	case *expr.Object:
		for _, named := range *actual {
			if err := p.addAttribute(named.Attribute, genpkg, seen); err != nil {
				return err
			}
		}
	case *expr.Array:
		return p.addAttribute(actual.ElemType, genpkg, seen)
	case *expr.Map:
		if err := p.addAttribute(actual.KeyType, genpkg, seen); err != nil {
			return err
		}
		return p.addAttribute(actual.ElemType, genpkg, seen)
	case *expr.Union:
		for _, named := range actual.Values {
			if err := p.addAttribute(named.Attribute, genpkg, seen); err != nil {
				return err
			}
		}
	}
	return nil
}

// add records one complete import path and rejects contradictory preferred
// package names before aliases are allocated.
func (p *importAliasPlan) add(importPath, preferred string, explicit, fixed bool) error {
	if importPath == "" {
		return nil
	}
	if preferred == "" {
		preferred = path.Base(importPath)
	}
	if existing, ok := p.candidates[importPath]; ok {
		if existing.preferred != preferred {
			return fmt.Errorf(
				"import path %q cannot use both package names %q and %q",
				importPath,
				existing.preferred,
				preferred,
			)
		}
		existing.explicit = existing.explicit || explicit
		existing.fixed = existing.fixed || fixed
		p.candidates[importPath] = existing
		return nil
	}
	p.candidates[importPath] = importAliasCandidate{
		preferred: preferred,
		explicit:  explicit,
		fixed:     fixed,
	}
	return nil
}

// freeze allocates aliases in fixed-priority, full-path order and returns the
// immutable lookup used throughout rendering.
func (p *importAliasPlan) freeze() *importAliases {
	paths := make([]string, 0, len(p.candidates))
	for importPath := range p.candidates {
		paths = append(paths, importPath)
	}
	sort.Slice(paths, func(i, j int) bool {
		left, right := p.candidates[paths[i]], p.candidates[paths[j]]
		if left.fixed != right.fixed {
			return left.fixed
		}
		return paths[i] < paths[j]
	})
	scope := codegen.NewNameScope()
	bindings := make(map[string]importBinding, len(paths))
	for _, importPath := range paths {
		candidate := p.candidates[importPath]
		bindings[importPath] = importBinding{
			name:      scope.Unique(candidate.preferred),
			preferred: candidate.preferred,
			explicit:  candidate.explicit,
		}
	}
	scope.Freeze()
	return &importAliases{bindings: bindings}
}

// name returns the frozen qualifier for importPath and panics when rendering
// asks for a package that was absent from alias planning.
func (a *importAliases) name(importPath string) string {
	binding, ok := a.bindings[importPath]
	if !ok {
		panic(fmt.Sprintf("import path %q has no planned alias", importPath))
	}
	return binding.name
}

// spec returns the frozen import declaration for importPath.
func (a *importAliases) spec(importPath string) *codegen.ImportSpec {
	binding, ok := a.bindings[importPath]
	if !ok {
		panic(fmt.Sprintf("import path %q has no planned alias", importPath))
	}
	spec := &codegen.ImportSpec{Path: importPath}
	if binding.explicit || binding.name != binding.preferred {
		spec.Name = binding.name
	}
	return spec
}

// newImportCollector creates a file-scoped collector that omits imports of the
// package containing the generated file.
func newImportCollector(aliases *importAliases, genpkg, outputPackage string) *importCollector {
	return &importCollector{
		aliases:       aliases,
		genpkg:        genpkg,
		outputPackage: outputPackage,
		paths:         make(map[string]struct{}),
		legacy:        make(map[string]*codegen.ImportSpec),
	}
}

// collect walks inline shapes but stops at named types because a reference to a
// named declaration does not render that declaration's fields in the file.
func (c *importCollector) collect(attribute *expr.AttributeExpr) {
	if attribute == nil || attribute.Type == expr.Empty {
		return
	}
	c.addMetaImport(attribute)
	switch actual := attribute.Type.(type) {
	case expr.UserType:
		c.addLocation(codegen.UserTypeLocation(actual))
	case *expr.Object:
		for _, named := range *actual {
			c.collect(named.Attribute)
		}
	case *expr.Array:
		c.collect(actual.ElemType)
	case *expr.Map:
		c.collect(actual.KeyType)
		c.collect(actual.ElemType)
	case *expr.Union:
		for _, named := range actual.Values {
			c.collect(named.Attribute)
		}
	}
}

// addLocation records the generated package selected by location unless it is
// the package currently being emitted.
func (c *importCollector) addLocation(location *codegen.Location) {
	if location == nil {
		return
	}
	importPath := path.Join(c.genpkg, location.RelImportPath)
	if importPath != c.outputPackage {
		c.paths[importPath] = struct{}{}
		c.legacy[importPath] = &codegen.ImportSpec{
			Name: location.PackageName(),
			Path: importPath,
		}
	}
}

// addMetaImport records the package named by struct:field:type metadata unless
// the metadata refers to the package currently being emitted.
func (c *importCollector) addMetaImport(attribute *expr.AttributeExpr) {
	_, spec := codegen.GetMetaType(attribute)
	if spec != nil && spec.Path != c.outputPackage {
		c.paths[spec.Path] = struct{}{}
		c.legacy[spec.Path] = spec
	}
}

// imports returns a deterministic snapshot of the packages collected for one
// generated file.
func (c *importCollector) imports() []*codegen.ImportSpec {
	paths := make([]string, 0, len(c.paths))
	for importPath := range c.paths {
		paths = append(paths, importPath)
	}
	sort.Strings(paths)
	imports := make([]*codegen.ImportSpec, len(paths))
	for i, importPath := range paths {
		if c.aliases != nil {
			imports[i] = c.aliases.spec(importPath)
			continue
		}
		imports[i] = c.legacy[importPath]
	}
	return imports
}
