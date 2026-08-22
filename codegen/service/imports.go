// This file plans one immutable import alias per complete package path, then
// computes the exact subset of those imports used by each generated service
// file. Qualified references and import declarations share these bindings.
package service

import (
	"path"
	"slices"
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
		planning      bool
		err           error
	}

	// retainedFileImports stores the complete package paths selected for one
	// emitted file and their frozen import declarations after linking.
	retainedFileImports struct {
		paths []string
		specs []*codegen.ImportSpec
	}

	// serviceFileImports keeps imports separate for files that emit different
	// subsets of one service's types and runtime helpers.
	serviceFileImports struct {
		service                   retainedFileImports
		endpoint                  retainedFileImports
		client                    retainedFileImports
		views                     retainedFileImports
		serverInterceptors        retainedFileImports
		clientInterceptors        retainedFileImports
		interceptorWrappers       retainedFileImports
		exampleService            retainedFileImports
		exampleServerInterceptors retainedFileImports
		exampleClientInterceptors retainedFileImports
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

// newImportAliases returns the generation-owned frozen alias binding used by
// service analysis and rendering.
func newImportAliases(root *expr.RootExpr, generation *codegen.Generation) (*importAliases, error) {
	if !generation.HasRoot(root) {
		return nil, rootMembershipError(root)
	}
	return &importAliases{generation: generation}, nil
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

// newPlanningImportCollector creates the same path walker used by rendering
// and additionally declares each discovered metadata or generated-package
// preference in the generation alias catalog.
func newPlanningImportCollector(generation *codegen.Generation, outputPackage string) *importCollector {
	return &importCollector{
		aliases:       &importAliases{generation: generation},
		genpkg:        generation.GenPkg(),
		outputPackage: outputPackage,
		paths:         make(map[string]struct{}),
		planning:      true,
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
		if c.planning && c.err == nil {
			c.err = c.aliases.generation.ReserveGeneratedImport(codegen.NewImport(
				strings.ToLower(codegen.Goify(path.Base(importPath), false)),
				importPath,
			))
		}
	}
}

// addMetaImport records the package named by struct:field:type metadata unless
// the metadata refers to the package currently being emitted.
func (c *importCollector) addMetaImport(attribute *expr.AttributeExpr) {
	_, spec := codegen.GetMetaType(attribute)
	if spec != nil && spec.Path != c.outputPackage {
		c.paths[spec.Path] = struct{}{}
		if c.planning && c.err == nil {
			c.err = c.aliases.generation.DeclareImport(spec)
		}
	}
}

// retainFileImports collects one emitted file's exact fixed, generated, type
// definition, and recursive-reference package paths before names freeze.
func retainFileImports(
	generation *codegen.Generation,
	outputPackage string,
	fixed, generated []*codegen.ImportSpec,
	definitions, references []*expr.AttributeExpr,
) (retainedFileImports, error) {
	collector := newPlanningImportCollector(generation, outputPackage)
	for _, spec := range fixed {
		collector.addPath(spec.Path)
		if err := generation.RequireImport(spec); err != nil {
			return retainedFileImports{}, err
		}
	}
	for _, spec := range generated {
		collector.addPath(spec.Path)
		if err := generation.ReserveGeneratedImport(spec); err != nil {
			return retainedFileImports{}, err
		}
	}
	for _, attribute := range definitions {
		collector.collectDefinition(attribute)
	}
	seen := make(map[expr.UserType]struct{})
	for _, attribute := range references {
		collector.collectReferences(attribute, seen)
	}
	if collector.err != nil {
		return retainedFileImports{}, collector.err
	}
	paths := make([]string, 0, len(collector.paths))
	for importPath := range collector.paths {
		paths = append(paths, importPath)
	}
	sort.Strings(paths)
	return retainedFileImports{paths: paths}, nil
}

// linkFileImports resolves one retained path list through the frozen
// Generation alias catalog without traversing service attributes.
func linkFileImports(imports *retainedFileImports, generation *codegen.Generation) {
	imports.specs = make([]*codegen.ImportSpec, len(imports.paths))
	for index, importPath := range imports.paths {
		imports.specs[index] = generation.Import(importPath)
	}
}

// addRetainedImportPath adds one explicitly declared package to a file's
// retained path set while preserving deterministic order.
func addRetainedImportPath(imports *retainedFileImports, importPath string) {
	index, found := slices.BinarySearch(imports.paths, importPath)
	if found {
		return
	}
	imports.paths = slices.Insert(imports.paths, index, importPath)
}

// planServiceFileImports selects the package paths used by each concrete file
// emitted for one retained service and declares their alias preferences before
// generation freezes.
func planServiceFileImports(facts *serviceFacts, rootTypes *rootTypeSet, generation *codegen.Generation) error {
	service := facts.service
	servicePath := servicePackagePath(generation.GenPkg(), service)
	serviceImport := codegen.NewImport(strings.ToLower(codegen.Goify(service.Name, false)), servicePath)
	viewsImport := codegen.NewImport(serviceImport.Name+"views", servicePath+"/views")
	facts.generatedTypeImports = make(map[*codegen.TypeDeclaration]*retainedFileImports)

	definitions := serviceDefinitionAttributes(facts)
	serviceDefinitions := append([]*expr.AttributeExpr(nil), facts.referenceAttributes...)
	serviceDefinitions = append(serviceDefinitions, definitions...)
	viewDefinitions := viewDefinitionAttributes(facts)

	contextImport := codegen.SimpleImport("context")
	ioImport := codegen.SimpleImport("io")
	goaImport := codegen.GoaImport("")
	securityImport := codegen.GoaImport("security")
	logImport := codegen.SimpleImport("goa.design/clue/log")

	serviceFixed := []*codegen.ImportSpec{contextImport}
	if serviceUsesIO(facts) {
		serviceFixed = append(serviceFixed, ioImport)
	}
	if serviceUsesGoaErrors(facts) {
		serviceFixed = append(serviceFixed, goaImport)
	}
	if serviceHasSchemes(facts) {
		serviceFixed = append(serviceFixed, securityImport)
	}
	var serviceGenerated []*codegen.ImportSpec
	if len(facts.projections) > 0 {
		serviceGenerated = append(serviceGenerated, viewsImport)
	}
	var err error
	facts.imports.service, err = retainFileImports(
		generation, servicePath, serviceFixed, serviceGenerated, serviceDefinitions, nil,
	)
	if err != nil {
		return err
	}

	endpointFixed := []*codegen.ImportSpec{contextImport, goaImport}
	if serviceUsesIO(facts) {
		endpointFixed = append(endpointFixed, ioImport)
	}
	if serviceHasSchemes(facts) {
		endpointFixed = append(endpointFixed, securityImport)
	}
	var endpointGenerated []*codegen.ImportSpec
	for _, method := range facts.methods {
		if facts.methodByExpr[method].viewedResult != nil {
			endpointGenerated = append(endpointGenerated, viewsImport)
			break
		}
	}
	facts.imports.endpoint, err = retainFileImports(
		generation, servicePath, endpointFixed, endpointGenerated, facts.referenceAttributes, nil,
	)
	if err != nil {
		return err
	}

	clientFixed := []*codegen.ImportSpec{contextImport, goaImport}
	if serviceUsesIO(facts) {
		clientFixed = append(clientFixed, ioImport)
	}
	facts.imports.client, err = retainFileImports(
		generation, servicePath, clientFixed, nil, facts.referenceAttributes, nil,
	)
	if err != nil {
		return err
	}

	if len(facts.projections) > 0 {
		viewsFixed := []*codegen.ImportSpec{goaImport, codegen.SimpleImport("unicode/utf8")}
		if len(facts.viewUnions) > 0 {
			viewsFixed = append(viewsFixed,
				codegen.SimpleImport("bytes"),
				codegen.SimpleImport("encoding/json"),
				codegen.SimpleImport("fmt"),
			)
		}
		facts.imports.views, err = retainFileImports(
			generation, servicePath+"/views", viewsFixed, nil, viewDefinitions, nil,
		)
		if err != nil {
			return err
		}
	}

	interceptorFixed := []*codegen.ImportSpec{contextImport, goaImport}
	if len(facts.serverInterceptors) > 0 {
		facts.imports.serverInterceptors, err = retainFileImports(
			generation, servicePath, interceptorFixed, nil, nil, nil,
		)
		if err != nil {
			return err
		}
	}
	if len(facts.clientInterceptors) > 0 {
		facts.imports.clientInterceptors, err = retainFileImports(
			generation, servicePath, interceptorFixed, nil, nil, nil,
		)
		if err != nil {
			return err
		}
	}
	if len(facts.serverInterceptors) > 0 || len(facts.clientInterceptors) > 0 {
		facts.imports.interceptorWrappers, err = retainFileImports(
			generation, servicePath, interceptorFixed, nil, nil, nil,
		)
		if err != nil {
			return err
		}
	}

	exampleFixed := []*codegen.ImportSpec{contextImport, logImport}
	if serviceUsesIO(facts) {
		exampleFixed = append(exampleFixed, ioImport)
	}
	if serviceUsesResponseBody(facts) {
		exampleFixed = append(exampleFixed, codegen.SimpleImport("strings"))
	}
	if serviceHasSchemes(facts) {
		exampleFixed = append(exampleFixed, codegen.SimpleImport("fmt"), securityImport)
	}
	facts.imports.exampleService, err = retainFileImports(
		generation, path.Dir(generation.GenPkg()), exampleFixed,
		[]*codegen.ImportSpec{serviceImport}, facts.referenceAttributes, nil,
	)
	if err != nil {
		return err
	}

	exampleInterceptorFixed := []*codegen.ImportSpec{contextImport, logImport, goaImport}
	if len(facts.serverInterceptors) > 0 {
		facts.imports.exampleServerInterceptors, err = retainFileImports(
			generation, path.Join(path.Dir(generation.GenPkg()), "interceptors"),
			exampleInterceptorFixed, []*codegen.ImportSpec{serviceImport}, nil, nil,
		)
		if err != nil {
			return err
		}
	}
	if len(facts.clientInterceptors) > 0 {
		facts.imports.exampleClientInterceptors, err = retainFileImports(
			generation, path.Join(path.Dir(generation.GenPkg()), "interceptors"),
			exampleInterceptorFixed, []*codegen.ImportSpec{serviceImport}, nil, nil,
		)
		if err != nil {
			return err
		}
	}
	for _, userType := range append(append([]*userTypeFacts(nil), facts.userTypes...), facts.errorTypes...) {
		if userType.location == nil {
			continue
		}
		userType.imports, err = retainFileImports(
			generation,
			userType.declaration.PackagePath(),
			nil,
			nil,
			[]*expr.AttributeExpr{userType.userType.Attribute()},
			nil,
		)
		if err != nil {
			return err
		}
		facts.generatedTypeImports[userType.declaration] = &userType.imports
	}
	for _, method := range facts.methods {
		attributes := []*expr.AttributeExpr{method.Payload, method.StreamingPayload, method.Result}
		if method.HasMixedResults() {
			attributes = append(attributes, method.StreamingResult)
		}
		for _, attribute := range attributes {
			if attribute == nil || codegen.UserTypeLocation(attribute.Type) == nil {
				continue
			}
			userType, ok := attribute.Type.(expr.UserType)
			if !ok {
				continue
			}
			if _, normalized := generation.NormalizedMethodType(userType); !normalized {
				continue
			}
			owner := generation.Package(generatedPackagePath(
				generation.GenPkg(), facts.service, codegen.UserTypeLocation(userType),
			))
			declaration, err := owner.UserType(rootTypes.canonical(userType))
			if err != nil {
				return err
			}
			if _, exists := facts.generatedTypeImports[declaration]; exists {
				continue
			}
			retained, err := retainFileImports(
				generation,
				declaration.PackagePath(),
				nil,
				nil,
				[]*expr.AttributeExpr{userType.Attribute()},
				nil,
			)
			if err != nil {
				return err
			}
			facts.generatedTypeImports[declaration] = &retained
		}
	}
	unionFixed := []*codegen.ImportSpec{
		codegen.SimpleImport("bytes"),
		codegen.SimpleImport("encoding/json"),
		codegen.SimpleImport("fmt"),
		goaImport,
	}
	for _, union := range facts.unions {
		definitions := make([]*expr.AttributeExpr, 0, len(union.union.Values)*2)
		for _, branch := range union.union.Values {
			definitions = append(definitions, branch.Attribute)
			if userType, ok := branch.Attribute.Type.(expr.UserType); ok {
				definitions = append(definitions, userType.Attribute())
			}
		}
		union.imports, err = retainFileImports(
			generation,
			union.declaration.PackagePath(),
			unionFixed,
			nil,
			definitions,
			nil,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// linkServiceFileImports resolves every concrete file contribution after the
// generation alias catalog freezes.
func linkServiceFileImports(facts *serviceFacts, generation *codegen.Generation) {
	imports := []*retainedFileImports{
		&facts.imports.service,
		&facts.imports.endpoint,
		&facts.imports.client,
		&facts.imports.views,
		&facts.imports.serverInterceptors,
		&facts.imports.clientInterceptors,
		&facts.imports.interceptorWrappers,
		&facts.imports.exampleService,
		&facts.imports.exampleServerInterceptors,
		&facts.imports.exampleClientInterceptors,
	}
	for _, retained := range imports {
		linkFileImports(retained, generation)
	}
	for _, userType := range append(append([]*userTypeFacts(nil), facts.userTypes...), facts.errorTypes...) {
		linkFileImports(&userType.imports, generation)
	}
	for _, union := range facts.unions {
		linkFileImports(&union.imports, generation)
	}
	for _, imports := range facts.generatedTypeImports {
		linkFileImports(imports, generation)
	}
}

// serviceDefinitionAttributes returns the exact named definitions written to
// service.go in addition to method references.
func serviceDefinitionAttributes(facts *serviceFacts) []*expr.AttributeExpr {
	var definitions []*expr.AttributeExpr
	for _, method := range facts.methods {
		attributes := []*expr.AttributeExpr{method.Payload, method.StreamingPayload, method.Result}
		if method.HasMixedResults() {
			attributes = append(attributes, method.StreamingResult)
		}
		for _, attribute := range attributes {
			if attribute == nil || codegen.UserTypeLocation(attribute.Type) != nil {
				continue
			}
			if userType, ok := attribute.Type.(expr.UserType); ok {
				definitions = append(definitions, userType.Attribute())
			}
		}
	}
	for _, userType := range append(append([]*userTypeFacts(nil), facts.userTypes...), facts.errorTypes...) {
		if userType.location == nil {
			definitions = append(definitions, userType.userType.Attribute())
		}
	}
	return definitions
}

// viewDefinitionAttributes returns each projected definition emitted in the
// service views file exactly once by expression identity.
func viewDefinitionAttributes(facts *serviceFacts) []*expr.AttributeExpr {
	seen := make(map[*expr.AttributeExpr]struct{})
	var definitions []*expr.AttributeExpr
	for _, projection := range facts.projections {
		for _, projected := range projection.types {
			attribute := projected.pair.projectedAttribute
			if userType, ok := attribute.Type.(expr.UserType); ok {
				attribute = userType.Attribute()
			}
			if _, ok := seen[attribute]; ok {
				continue
			}
			seen[attribute] = struct{}{}
			definitions = append(definitions, attribute)
		}
	}
	return definitions
}

// serviceUsesIO reports whether generated method signatures expose a raw
// request or response body stream.
func serviceUsesIO(facts *serviceFacts) bool {
	for _, method := range facts.methodByExpr {
		if method.skipRequestBodyEncodeDecode || method.skipResponseBodyEncodeDecode {
			return true
		}
	}
	return false
}

// serviceUsesResponseBody reports whether the starter implementation creates
// a raw response body from a string reader.
func serviceUsesResponseBody(facts *serviceFacts) bool {
	for _, method := range facts.methodByExpr {
		if method.skipResponseBodyEncodeDecode {
			return true
		}
	}
	return false
}

// serviceUsesGoaErrors reports whether service.go emits a constructor that
// calls the Goa service-error runtime.
func serviceUsesGoaErrors(facts *serviceFacts) bool {
	for _, serviceError := range facts.errors {
		if serviceError.Type == expr.ErrorResult {
			return true
		}
	}
	for _, method := range facts.methods {
		for _, methodError := range method.Errors {
			if methodError.Type == expr.ErrorResult {
				return true
			}
		}
	}
	return false
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
