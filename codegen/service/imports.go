// This file plans one immutable import alias per complete package path, then
// computes the exact subset of those imports used by each generated service
// file. Qualified references and import declarations share these bindings.
package service

import (
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
		generation *codegen.Generation
	}

	// importCollector accumulates the imports referenced by one generated Go
	// file while traversing recursive service type definitions.
	importCollector struct {
		aliases       *importAliases
		genpkg        string
		outputPackage string
		paths         map[string]struct{}
	}
)

// AttributeImports returns the exact generated-type and metadata imports
// referenced by attributes using the frozen aliases shared with service type
// references.
func (d *ServicesData) AttributeImports(outputPackage string, attributes ...*expr.AttributeExpr) []*codegen.ImportSpec {
	collector := newImportCollector(d.aliases, d.generation.GenPkg(), outputPackage)
	seen := make(map[expr.UserType]struct{})
	for _, attribute := range attributes {
		collector.collectReferences(attribute, seen)
	}
	return collector.imports()
}

// fileImports returns one canonical import per complete path used by a single
// generated file. Explicit paths and attribute-derived paths are deduplicated
// before their frozen aliases are materialized.
func (d *ServicesData) fileImports(outputPackage string, paths []string, attributes ...*expr.AttributeExpr) []*codegen.ImportSpec {
	collector := newImportCollector(d.aliases, d.generation.GenPkg(), outputPackage)
	for _, importPath := range paths {
		collector.addPath(importPath)
	}
	for _, attribute := range attributes {
		collector.collectDefinition(attribute)
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

// newImportAliases returns the generation-owned frozen alias binding used by
// service analysis and rendering.
func newImportAliases(root *expr.RootExpr, generation *codegen.Generation) (*importAliases, error) {
	if !generation.HasRoot(root) {
		return nil, rootMembershipError(root)
	}
	return &importAliases{generation: generation}, nil
}

// planImports registers every fixed package and every external or generated
// package referenced by declarations selected for emission from root.
func planImports(root *expr.RootExpr, inputs []plannedAttribute, generation *codegen.Generation) error {
	fixed := []*codegen.ImportSpec{
		codegen.SimpleImport("bytes"),
		codegen.SimpleImport("context"),
		codegen.SimpleImport("encoding/json"),
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("io"),
		codegen.SimpleImport("strings"),
		codegen.SimpleImport("unicode/utf8"),
		codegen.SimpleImport("goa.design/clue/log"),
		codegen.GoaImport(""),
		codegen.GoaImport("security"),
	}
	for _, spec := range fixed {
		if err := generation.RequireImport(spec); err != nil {
			return err
		}
	}
	for _, service := range root.Services {
		servicePath := servicePackagePath(generation.GenPkg(), service)
		serviceName := strings.ToLower(codegen.Goify(service.Name, false))
		if err := generation.ReserveGeneratedImport(codegen.NewImport(serviceName, servicePath)); err != nil {
			return err
		}
		if err := generation.ReserveGeneratedImport(codegen.NewImport(serviceName+"views", servicePath+"/views")); err != nil {
			return err
		}
	}
	seen := make(map[expr.UserType]struct{})
	for _, input := range inputs {
		if err := planAttributeImports(input.attribute, generation, seen); err != nil {
			return err
		}
	}
	for _, typeMap := range append(append([]*expr.TypeMap(nil), root.Conversions...), root.Creations...) {
		importPath, alias, err := getExternalTypeInfo(typeMap.External)
		if err != nil {
			return err
		}
		if err := generation.DeclareImport(codegen.NewImport(alias, importPath)); err != nil {
			return err
		}
	}
	return nil
}

// planAttributeImports recursively records every explicit generated location and
// struct:field:type import reachable from attribute.
func planAttributeImports(attribute *expr.AttributeExpr, generation *codegen.Generation, seen map[expr.UserType]struct{}) error {
	if attribute == nil || attribute.Type == expr.Empty {
		return nil
	}
	if _, spec := codegen.GetMetaType(attribute); spec != nil {
		if err := generation.DeclareImport(spec); err != nil {
			return err
		}
	}
	switch actual := attribute.Type.(type) {
	case expr.UserType:
		if location := codegen.UserTypeLocation(actual); location != nil {
			owner := generation.Package(path.Join(generation.GenPkg(), location.RelImportPath))
			if err := generation.DeclareImport(codegen.NewImport(
				strings.ToLower(codegen.Goify(path.Base(owner.ImportPath()), false)),
				owner.ImportPath(),
			)); err != nil {
				return err
			}
		}
		origin := actual.Origin()
		if _, ok := seen[origin]; ok {
			return nil
		}
		seen[origin] = struct{}{}
		return planAttributeImports(actual.Attribute(), generation, seen)
	case *expr.Object:
		for _, named := range *actual {
			if err := planAttributeImports(named.Attribute, generation, seen); err != nil {
				return err
			}
		}
	case *expr.Array:
		return planAttributeImports(actual.ElemType, generation, seen)
	case *expr.Map:
		if err := planAttributeImports(actual.KeyType, generation, seen); err != nil {
			return err
		}
		return planAttributeImports(actual.ElemType, generation, seen)
	case *expr.Union:
		for _, named := range actual.Values {
			if err := planAttributeImports(named.Attribute, generation, seen); err != nil {
				return err
			}
		}
	}
	return nil
}

// name returns the frozen qualifier for importPath and panics when rendering
// asks for a package that was absent from alias planning.
func (a *importAliases) name(importPath string) string {
	return a.generation.ImportName(importPath)
}

// spec returns the frozen import declaration for importPath.
func (a *importAliases) spec(importPath string) *codegen.ImportSpec {
	return a.generation.Import(importPath)
}

// newImportCollector creates a file-scoped collector that omits imports of the
// package containing the generated file.
func newImportCollector(aliases *importAliases, genpkg, outputPackage string) *importCollector {
	return &importCollector{
		aliases:       aliases,
		genpkg:        genpkg,
		outputPackage: outputPackage,
		paths:         make(map[string]struct{}),
	}
}

// addPath records an explicitly referenced package unless it is the package
// currently being emitted.
func (c *importCollector) addPath(importPath string) {
	if importPath != "" && importPath != c.outputPackage {
		c.paths[importPath] = struct{}{}
	}
}

// collectDefinition records imports used to render an attribute definition.
// Named references stop traversal because their fields are emitted elsewhere.
func (c *importCollector) collectDefinition(attribute *expr.AttributeExpr) {
	c.collectAttribute(attribute, false, nil)
}

// collectReferences records imports used by recursive conversion and
// validation code, including types and metadata nested in named declarations.
func (c *importCollector) collectReferences(attribute *expr.AttributeExpr, seen map[expr.UserType]struct{}) {
	c.collectAttribute(attribute, true, seen)
}

// collectAttribute implements definition and recursive-reference traversal.
func (c *importCollector) collectAttribute(attribute *expr.AttributeExpr, expandNamed bool, seen map[expr.UserType]struct{}) {
	if attribute == nil || attribute.Type == expr.Empty {
		return
	}
	c.addMetaImport(attribute)
	switch actual := attribute.Type.(type) {
	case expr.UserType:
		c.addLocation(codegen.UserTypeLocation(actual))
		if !expandNamed {
			return
		}
		origin := actual.Origin()
		if _, ok := seen[origin]; ok {
			return
		}
		seen[origin] = struct{}{}
		c.collectAttribute(actual.Attribute(), true, seen)
	case *expr.Object:
		for _, named := range *actual {
			c.collectAttribute(named.Attribute, expandNamed, seen)
		}
	case *expr.Array:
		c.collectAttribute(actual.ElemType, expandNamed, seen)
	case *expr.Map:
		c.collectAttribute(actual.KeyType, expandNamed, seen)
		c.collectAttribute(actual.ElemType, expandNamed, seen)
	case *expr.Union:
		for _, named := range actual.Values {
			c.collectAttribute(named.Attribute, expandNamed, seen)
		}
	}
}

// addLocation records the generated package selected by location unless it is
// the package currently being emitted.
func (c *importCollector) addLocation(location *codegen.Location) {
	if location == nil {
		return
	}
	importPath := c.aliases.generation.Package(path.Join(c.genpkg, location.RelImportPath)).ImportPath()
	if importPath != c.outputPackage {
		c.paths[importPath] = struct{}{}
	}
}

// addMetaImport records the package named by struct:field:type metadata unless
// the metadata refers to the package currently being emitted.
func (c *importCollector) addMetaImport(attribute *expr.AttributeExpr) {
	_, spec := codegen.GetMetaType(attribute)
	if spec != nil && spec.Path != c.outputPackage {
		c.paths[spec.Path] = struct{}{}
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
		imports[i] = c.aliases.spec(importPath)
	}
	return imports
}
