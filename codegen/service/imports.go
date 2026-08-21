// This file computes imports from the service expressions rendered into one
// generated Go file. Callers identify the file's package explicitly so a file
// never imports itself and imports from unrelated services cannot leak into it.
package service

import (
	"path"
	"sort"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

type (
	// importCollector accumulates the imports referenced by one generated Go
	// file while traversing recursive service type definitions.
	importCollector struct {
		genpkg        string
		outputPackage string
		importsByPath map[string]*codegen.ImportSpec
	}
)

// AttributeImports returns the generated-type and struct:field:type imports
// referenced by attributes. Pass a named user type attribute when the file
// references that declaration. Pass the user type's underlying attribute when
// the file emits its definition. outputPackage is the full Go import path of
// the file receiving the imports.
func AttributeImports(genpkg, outputPackage string, attributes ...*expr.AttributeExpr) []*codegen.ImportSpec {
	collector := newImportCollector(genpkg, outputPackage)
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

// newImportCollector creates a file-scoped collector that omits imports of the
// package containing the generated file.
func newImportCollector(genpkg, outputPackage string) *importCollector {
	return &importCollector{
		genpkg:        genpkg,
		outputPackage: outputPackage,
		importsByPath: make(map[string]*codegen.ImportSpec),
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
	if importPath == c.outputPackage {
		return
	}
	c.importsByPath[importPath] = &codegen.ImportSpec{
		Name: location.PackageName(),
		Path: importPath,
	}
}

// addMetaImport records the package named by struct:field:type metadata unless
// the metadata refers to the package currently being emitted.
func (c *importCollector) addMetaImport(attribute *expr.AttributeExpr) {
	_, spec := codegen.GetMetaType(attribute)
	if spec == nil || spec.Path == c.outputPackage {
		return
	}
	c.importsByPath[spec.Path] = spec
}

// imports returns a deterministic snapshot of the packages collected for one
// generated file.
func (c *importCollector) imports() []*codegen.ImportSpec {
	paths := make([]string, 0, len(c.importsByPath))
	for importPath := range c.importsByPath {
		paths = append(paths, importPath)
	}
	sort.Strings(paths)
	imports := make([]*codegen.ImportSpec, len(paths))
	for i, importPath := range paths {
		imports[i] = c.importsByPath[importPath]
	}
	return imports
}
