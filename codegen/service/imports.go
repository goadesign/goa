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
	collector := newImportCollector(d.aliases, d.generation.GenPkg(), outputPackage)
	for _, attribute := range attributes {
		collector.collect(attribute)
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

// newImportAliases returns the generation-owned frozen alias binding used by
// service analysis and rendering.
func newImportAliases(root *expr.RootExpr, generation *codegen.Generation) (*importAliases, error) {
	if !generation.HasRoot(root) {
		return nil, rootMembershipError(root)
	}
	return &importAliases{generation: generation}, nil
}

// planImports registers every fixed and design-selected package path reachable
// from root in the generation-wide alias catalog.
func planImports(root *expr.RootExpr, generation *codegen.Generation) error {
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
	for _, userType := range root.Types {
		if err := planAttributeImports(&expr.AttributeExpr{Type: userType}, generation, seen); err != nil {
			return err
		}
	}
	for _, resultType := range root.ResultTypes {
		if err := planAttributeImports(&expr.AttributeExpr{Type: resultType}, generation, seen); err != nil {
			return err
		}
	}
	for _, service := range root.Services {
		for _, attribute := range serviceReferenceAttributes(service) {
			if err := planAttributeImports(attribute, generation, seen); err != nil {
				return err
			}
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
			if err := generation.DeclareImport(codegen.NewImport(
				location.PackageName(),
				path.Join(generation.GenPkg(), location.RelImportPath),
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
		legacy:        make(map[string]*codegen.ImportSpec),
	}
}

// addPath records an explicitly referenced package unless it is the package
// currently being emitted.
func (c *importCollector) addPath(importPath string) {
	if importPath != "" && importPath != c.outputPackage {
		c.paths[importPath] = struct{}{}
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
