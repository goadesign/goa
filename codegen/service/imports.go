// This file chooses one Go package name for each import path and records which
// imports each generated service file uses.
package service

import (
	"path"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

type (
	// importAliases returns the Go package name chosen for an import path in one
	// output package before files are rendered.
	importAliases struct {
		generation *codegen.Generation
	}

	// serviceFileImports keeps imports separate for files that emit different
	// subsets of one service's types and runtime helpers.
	serviceFileImports struct {
		service                   *codegen.GeneratedImportPlan
		endpoint                  *codegen.GeneratedImportPlan
		client                    *codegen.GeneratedImportPlan
		views                     *codegen.GeneratedImportPlan
		serverInterceptors        *codegen.GeneratedImportPlan
		clientInterceptors        *codegen.GeneratedImportPlan
		interceptorWrappers       *codegen.GeneratedImportPlan
		exampleService            *codegen.GeneratedImportPlan
		exampleServerInterceptors *codegen.GeneratedImportPlan
		exampleClientInterceptors *codegen.GeneratedImportPlan
	}
)

// newImportAliases returns package-name lookups for the supplied generation.
// Service analysis and rendering use these lookups after all import names are
// chosen.
func newImportAliases(root *expr.RootExpr, generation *codegen.Generation) (*importAliases, error) {
	if !generation.HasRoot(root) {
		return nil, rootMembershipError(root)
	}
	return &importAliases{generation: generation}, nil
}

// name returns the chosen Go package name for importPath. It panics when a
// renderer asks for a package that was not recorded during planning.
func (a *importAliases) name(outputPackage, importPath string) string {
	return a.generation.Package(outputPackage).ImportName(importPath)
}

// spec returns the completed import declaration for importPath.
func (a *importAliases) spec(outputPackage, importPath string) *codegen.ImportSpec {
	return a.generation.Package(outputPackage).Import(importPath)
}

// retainFileImports collects the fixed, generated, type-definition, and
// recursive-reference package paths used by one emitted file before
// Generation.Freeze chooses their Go package names.
func retainFileImports(
	generation *codegen.Generation,
	outputPackage string,
	fixed, generated []*codegen.ImportSpec,
	definitions, references []*expr.AttributeExpr,
) (*codegen.GeneratedImportPlan, error) {
	imports := codegen.NewGeneratedImportPlan(generation.Package(outputPackage))
	if err := imports.Require(fixed...); err != nil {
		return nil, err
	}
	if err := imports.AddGenerated(generated...); err != nil {
		return nil, err
	}
	if err := imports.AddTypeExpressions(definitions...); err != nil {
		return nil, err
	}
	if err := imports.AddRecursiveTypeReferences(references...); err != nil {
		return nil, err
	}
	return imports, nil
}

// planServiceFileImports selects the package paths used by each concrete file
// emitted for one service copied into the plan. It requests their preferred Go
// package names before Generation.Freeze chooses the final names.
func planServiceFileImports(facts *serviceFacts, rootTypes *rootTypeSet, generation *codegen.Generation) error {
	servicePath := facts.packagePath
	serviceImport := facts.packageImport
	viewsImport := facts.viewsImport
	facts.generatedTypeImports = make(map[*codegen.TypeDeclaration]*codegen.GeneratedImportPlan)

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
		viewsFixed := []*codegen.ImportSpec{goaImport}
		validationFixed, validationGenerated := viewValidationImports(facts)
		viewsFixed = append(viewsFixed, validationFixed...)
		if len(facts.viewUnions) > 0 {
			viewsFixed = append(viewsFixed,
				codegen.SimpleImport("bytes"),
				codegen.SimpleImport("encoding/json"),
				codegen.SimpleImport("fmt"),
			)
		}
		facts.imports.views, err = retainFileImports(
			generation, servicePath+"/views", viewsFixed, validationGenerated, viewDefinitions, nil,
		)
		if err != nil {
			return err
		}
	}

	interceptorFixed := []*codegen.ImportSpec{contextImport, goaImport}
	if len(facts.serverInterceptors) > 0 {
		serverInterceptorNames := interceptorNames(facts.serverInterceptorFacts)
		serverInterceptorTypes := interceptorTypes(facts.serverInterceptorFacts)
		serverInterceptorTypes = append(
			serverInterceptorTypes,
			interceptorTypesOnly(facts.clientInterceptorFacts, serverInterceptorNames)...,
		)
		facts.imports.serverInterceptors, err = retainFileImports(
			generation, servicePath, interceptorFixed, nil, serverInterceptorTypes, nil,
		)
		if err != nil {
			return err
		}
	}
	if len(facts.clientInterceptors) > 0 {
		clientInterceptorTypes := interceptorTypesWithout(
			facts.clientInterceptorFacts,
			interceptorNames(facts.serverInterceptorFacts),
		)
		facts.imports.clientInterceptors, err = retainFileImports(
			generation, servicePath, interceptorFixed, nil, clientInterceptorTypes, nil,
		)
		if err != nil {
			return err
		}
	}
	if len(facts.serverInterceptors) > 0 || len(facts.clientInterceptors) > 0 {
		streamTypes := interceptorStreamTypes(facts.serverInterceptorFacts)
		streamTypes = append(streamTypes, interceptorStreamTypes(facts.clientInterceptorFacts)...)
		facts.imports.interceptorWrappers, err = retainFileImports(
			generation, servicePath, interceptorFixed, nil, streamTypes, nil,
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
		facts.generatedTypeImports[userType.declaration] = userType.imports
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
				generation.GenPkg(), facts.packagePath, codegen.UserTypeLocation(userType),
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
			facts.generatedTypeImports[declaration] = retained
		}
	}
	unionFixed := []*codegen.ImportSpec{
		codegen.SimpleImport("bytes"),
		codegen.SimpleImport("encoding/json"),
		codegen.SimpleImport("fmt"),
		goaImport,
	}
	for _, union := range facts.unions {
		definitions := make([]*expr.AttributeExpr, 0, len(union.union.Values))
		for _, branch := range union.union.Values {
			definitions = append(definitions, branch.Attribute)
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

// interceptorTypes returns the selected fields whose Go types are written
// in an interceptor interface or accessor method.
func interceptorTypes(interceptors []*interceptorFacts) []*expr.AttributeExpr {
	return interceptorTypesWithout(interceptors, nil)
}

// interceptorTypesWithout returns selected interceptor fields except for
// interceptors whose names are emitted by another file.
func interceptorTypesWithout(interceptors []*interceptorFacts, excluded map[string]struct{}) []*expr.AttributeExpr {
	var attributes []*expr.AttributeExpr
	for _, interceptor := range interceptors {
		if _, skip := excluded[interceptor.name]; skip {
			continue
		}
		attributes = append(attributes, interceptorValueTypes(interceptor)...)
	}
	return attributes
}

// interceptorTypesOnly returns types for interceptor definitions
// written in another file with the same name.
func interceptorTypesOnly(interceptors []*interceptorFacts, included map[string]struct{}) []*expr.AttributeExpr {
	var attributes []*expr.AttributeExpr
	for _, interceptor := range interceptors {
		if _, keep := included[interceptor.name]; !keep {
			continue
		}
		attributes = append(attributes, interceptorValueTypes(interceptor)...)
	}
	return attributes
}

// interceptorValueTypes returns the selected fields and the complete
// method values stored behind their generated accessors.
func interceptorValueTypes(interceptor *interceptorFacts) []*expr.AttributeExpr {
	accesses := [][]*interceptorAccessFacts{
		interceptor.readPayloadFields,
		interceptor.writePayloadFields,
		interceptor.readResultFields,
		interceptor.writeResultFields,
		interceptor.readStreamingPayloadFields,
		interceptor.writeStreamingPayloadFields,
		interceptor.readStreamingResultFields,
		interceptor.writeStreamingResultFields,
	}
	var attributes []*expr.AttributeExpr
	for _, fields := range accesses {
		for _, field := range fields {
			attributes = append(attributes, field.attribute)
		}
	}
	hasPayload := len(interceptor.readPayloadFields) > 0 || len(interceptor.writePayloadFields) > 0
	hasResult := len(interceptor.readResultFields) > 0 || len(interceptor.writeResultFields) > 0
	hasStreamingPayload := len(interceptor.readStreamingPayloadFields) > 0 || len(interceptor.writeStreamingPayloadFields) > 0
	hasStreamingResult := len(interceptor.readStreamingResultFields) > 0 || len(interceptor.writeStreamingResultFields) > 0
	for _, method := range interceptor.methods {
		if hasPayload {
			attributes = append(attributes, method.payload.attribute)
		}
		if hasResult {
			attributes = append(attributes, method.result.attribute)
		}
		if hasStreamingPayload {
			attributes = append(attributes, method.streamingPayload.attribute)
		}
		if hasStreamingResult {
			attributes = append(attributes, method.streamingResult.attribute)
		}
	}
	return attributes
}

// interceptorStreamTypes returns the streaming values written by wrapper
// methods for interceptors that inspect those values.
func interceptorStreamTypes(interceptors []*interceptorFacts) []*expr.AttributeExpr {
	var attributes []*expr.AttributeExpr
	for _, interceptor := range interceptors {
		payload := len(interceptor.readStreamingPayloadFields) > 0 || len(interceptor.writeStreamingPayloadFields) > 0
		result := len(interceptor.readStreamingResultFields) > 0 || len(interceptor.writeStreamingResultFields) > 0
		for _, method := range interceptor.methods {
			if payload {
				attributes = append(attributes, method.streamingPayload.attribute)
			}
			if result {
				attributes = append(attributes, method.streamingResult.attribute)
			}
		}
	}
	return attributes
}

// interceptorNames returns the names of interceptors emitted in a file.
func interceptorNames(interceptors []*interceptorFacts) map[string]struct{} {
	names := make(map[string]struct{}, len(interceptors))
	for _, interceptor := range interceptors {
		names[interceptor.name] = struct{}{}
	}
	return names
}

// viewValidationImports separates packages named directly by templates from
// generated packages whose import names may change to avoid a collision.
func viewValidationImports(facts *serviceFacts) (fixed, generated []*codegen.ImportSpec) {
	for _, method := range facts.methods {
		projection := facts.projections[method]
		if projection == nil {
			continue
		}
		for _, projected := range projection.types {
			for _, validation := range projected.validations {
				if validation.plan == nil {
					continue
				}
				for _, preference := range validation.plan.ImportPreferences() {
					spec := codegen.NewImport(preference.Name, preference.Path)
					switch preference.Path {
					case codegen.GoaImport("").Path, "unicode/utf8":
						fixed = append(fixed, spec)
					default:
						generated = append(generated, spec)
					}
				}
			}
		}
	}
	return
}

// linkServiceFileImports resolves every file's imports after
// Generation.Freeze chooses all package names.
func linkServiceFileImports(facts *serviceFacts) error {
	imports := []*codegen.GeneratedImportPlan{
		facts.imports.service,
		facts.imports.endpoint,
		facts.imports.client,
		facts.imports.views,
		facts.imports.serverInterceptors,
		facts.imports.clientInterceptors,
		facts.imports.interceptorWrappers,
		facts.imports.exampleService,
		facts.imports.exampleServerInterceptors,
		facts.imports.exampleClientInterceptors,
	}
	for _, retained := range imports {
		if retained != nil {
			if err := retained.Link(); err != nil {
				return err
			}
		}
	}
	for _, union := range facts.unions {
		if union.imports != nil {
			if err := union.imports.Link(); err != nil {
				return err
			}
		}
	}
	for _, imports := range facts.generatedTypeImports {
		if err := imports.Link(); err != nil {
			return err
		}
	}
	return nil
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

// viewDefinitionAttributes returns each view-specific definition emitted in
// the service views file exactly once, even when several entries point to the
// same attribute value.
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
		if expr.IsErrorResult(serviceError.Type) {
			return true
		}
	}
	for _, method := range facts.methods {
		for _, methodError := range method.Errors {
			if expr.IsErrorResult(methodError.Type) {
				return true
			}
		}
	}
	return false
}
