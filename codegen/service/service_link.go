// This file links retained service facts into immutable render data after names and import aliases freeze.
package service

import (
	"fmt"
	"path"
	"slices"
	"sort"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// Link resolves the plan's collected facts through frozen declarations into
// the immutable template data consumed by renderers.
func (p *Plan) Link() error {
	if !p.generation.Frozen() {
		return fmt.Errorf("service plan cannot link before generation freeze")
	}
	if p.services != nil {
		return fmt.Errorf("service plan is already linked")
	}
	aliases, err := newImportAliases(p.facts.root, p.generation)
	if err != nil {
		return err
	}
	for _, facts := range p.facts.services {
		linkServiceFileImports(facts, p.generation)
	}
	if err := linkExternalConversions(p.facts, p.generation, aliases); err != nil {
		return err
	}
	services, err := linkServicesData(p.facts, p.generation, aliases)
	if err != nil {
		return err
	}
	p.services = services
	return nil
}

// Services returns the linked service render model. It panics before Link
// because no renderer or transport may observe provisional generated names.
func (p *Plan) Services() *ServicesData {
	if p.services == nil {
		panic("service render model requested before plan linking")
	}
	return p.services
}

// analyze creates the data necessary to render the code of the given service.
// It records the user types needed by the service definition in userTypes.
func (d *ServicesData) analyze(facts *serviceFacts) (*Data, error) {
	var (
		types      []*UserTypeData
		errTypes   []*UserTypeData
		errorInits []*ErrorInitData
		projTypes  []*ProjectedTypeData
		viewedRTs  []*ViewedResultTypeData
	)
	servicePackage := d.generation.Package(facts.packagePath)
	scope := servicePackage.Scope()
	viewScope := d.generation.Package(
		facts.viewsPath,
	).Scope()
	pkgName := codegen.Goify(path.Base(servicePackage.ImportPath()), false)
	seenErrors := make(map[string]struct{})
	type viewedResultKey struct {
		origin expr.UserType
		view   string
	}
	seenViewed := make(map[viewedResultKey]*ViewedResultTypeData)
	seenViewedDeclarations := make(map[*codegen.TypeDeclaration]struct{})
	viewDerived := make(map[expr.UserType]codegen.DerivedTypeID)
	serviceResolver := newRetainedServiceResolver(
		d.generation,
		d.aliases,
		facts.name,
		facts.packagePath,
		facts.packagePath,
	).withValidators(facts.validators)
	types = formatUserTypeFacts(facts.userTypes, facts.packagePath, d.aliases)
	errTypes = formatUserTypeFacts(facts.errorTypes, facts.packagePath, d.aliases)

	// recordError formats each selected ErrorResult constructor once.
	recordError := func(errorFacts *errorRenderFacts) {
		if errorFacts.serviceType {
			if _, ok := seenErrors[errorFacts.name]; ok {
				return
			}
			seenErrors[errorFacts.name] = struct{}{}
			errorInits = append(errorInits, buildRetainedErrorInitData(
				errorFacts,
				serviceResolver,
				facts.errorConstructors[errorFacts.name],
			))
		}
	}
	for _, errorFacts := range facts.errorFacts {
		recordError(errorFacts)
	}

	for _, method := range facts.orderedMethods {
		// Collect projected types
		if projection := method.projection; projection != nil {
			views := d.generation.Package(facts.viewsPath)
			for _, projectedFacts := range projection.types {
				pair := projectedFacts.pair
				identity := codegen.NewProjectedTypeID(pair.source)
				viewDerived[pair.projected.Origin()] = identity
			}
			viewResolver := newRetainedViewResolver(
				d.generation,
				d.aliases,
				facts.name,
				facts.viewsPath,
				viewDerived,
			).
				withValidators(facts.validators)
			for _, projectedFacts := range projection.types {
				pair := projectedFacts.pair
				identity := codegen.NewProjectedTypeID(pair.source)
				declaration, err := views.DerivedType(identity)
				if err != nil {
					return nil, err
				}
				projectedType := buildProjectedType(
					projectedFacts,
					serviceResolver,
					viewResolver,
					declaration,
				)
				projTypes = append(projTypes, projectedType)
			}
		}
		for _, errorFacts := range method.errors {
			recordError(errorFacts)
		}
	}
	viewUnions, err := d.formatViewUnions(facts)
	if err != nil {
		return nil, err
	}

	var (
		methods []*MethodData
		schemes SchemesData
	)
	methods = make([]*MethodData, len(facts.orderedMethods))
	methodDataByFacts := make(map[*methodFacts]*MethodData, len(facts.orderedMethods))
	for i, method := range facts.orderedMethods {
		m, err := d.buildMethodData(method, serviceResolver, facts)
		if err != nil {
			return nil, err
		}
		methods[i] = m
		methodDataByFacts[method] = m
		for _, s := range m.Schemes {
			schemes = schemes.Append(s)
		}
		viewedFacts := method.viewedResult
		if viewedFacts == nil {
			continue
		}
		key := viewedResultKey{origin: viewedFacts.origin, view: viewedFacts.viewName}
		if vrt, ok := seenViewed[key]; ok {
			m.ViewedResult = vrt
			continue
		}
		vrt := buildViewedResultType(
			viewedFacts,
			d.aliases.spec(facts.viewsPath).Name,
			serviceResolver,
			newRetainedViewResolver(
				d.generation,
				d.aliases,
				facts.name,
				facts.viewsPath,
				viewDerived,
			).
				withValidators(facts.validators),
			viewedFacts.declaration,
		)
		if _, found := seenViewedDeclarations[viewedFacts.declaration]; !found {
			viewedRTs = append(viewedRTs, vrt)
			seenViewedDeclarations[viewedFacts.declaration] = struct{}{}
		}
		m.ViewedResult = vrt
		seenViewed[key] = vrt
	}

	unions, err := d.formatServiceUnions(facts)
	if err != nil {
		return nil, err
	}

	desc := facts.description
	if desc == "" {
		desc = fmt.Sprintf("Service is the %s service interface.", facts.name)
	}

	varName := codegen.Goify(facts.name, false)
	data := &Data{
		ServiceDeclaration:      facts.names.declaration(serviceSymbolID{role: serviceInterfaceNameRole, service: facts.name}),
		AutherDeclaration:       facts.names[serviceSymbolID{role: serviceAutherNameRole, service: facts.name}].declaration,
		APINameDeclaration:      facts.names.declaration(serviceSymbolID{role: serviceAPINameRole, service: facts.name}),
		APIVersionDeclaration:   facts.names.declaration(serviceSymbolID{role: serviceAPIVersionNameRole, service: facts.name}),
		ServiceNameDeclaration:  facts.names.declaration(serviceSymbolID{role: serviceNameConstantRole, service: facts.name}),
		MethodNamesDeclaration:  facts.names.declaration(serviceSymbolID{role: serviceMethodNamesRole, service: facts.name}),
		EndpointsDeclaration:    facts.names.declaration(serviceSymbolID{role: serviceEndpointsNameRole, service: facts.name}),
		NewEndpointsDeclaration: facts.names.declaration(serviceSymbolID{role: serviceNewEndpointsNameRole, service: facts.name}),
		ClientDeclaration:       facts.names.declaration(serviceSymbolID{role: serviceClientNameRole, service: facts.name}),
		NewClientDeclaration:    facts.names.declaration(serviceSymbolID{role: serviceNewClientNameRole, service: facts.name}),
		StreamDeclaration:       facts.names[serviceSymbolID{role: serviceStreamNameRole, service: facts.name}].declaration,
		EventDeclaration:        facts.names[serviceSymbolID{role: serviceEventNameRole, service: facts.name}].declaration,
		ServerInterceptorsDeclaration: facts.names[serviceSymbolID{
			role: serviceServerInterceptorsNameRole, service: facts.name,
		}].declaration,
		ClientInterceptorsDeclaration: facts.names[serviceSymbolID{
			role: serviceClientInterceptorsNameRole, service: facts.name,
		}].declaration,
		ExampleStructDeclaration:      facts.exampleStruct,
		ExampleConstructorDeclaration: facts.exampleConstructor,
		Name:                          facts.name,
		Description:                   desc,
		APIName:                       d.facts.apiName,
		APIVersion:                    d.facts.apiVersion,
		VarName:                       varName,
		PathName:                      codegen.SnakeCase(varName),
		StructName:                    codegen.Goify(facts.name, true),
		PkgName:                       pkgName,
		Methods:                       methods,
		Schemes:                       schemes,
		ServerInterceptors:            d.collectInterceptors(facts, facts.serverInterceptorFacts, methodDataByFacts, serviceResolver, true),
		ClientInterceptors:            d.collectInterceptors(facts, facts.clientInterceptorFacts, methodDataByFacts, serviceResolver, false),
		Scope:                         scope,
		ViewScope:                     viewScope,
		errorTypes:                    errTypes,
		errorInits:                    errorInits,
		userTypes:                     types,
		projectedTypes:                projTypes,
		viewedResultTypes:             viewedRTs,
		unions:                        unions,
		viewUnions:                    viewUnions,
		viewDerived:                   viewDerived,
	}
	return data, nil
}

// collectInterceptors returns the set of interceptors defined on the given
// service including any interceptor defined on specific service methods or API.
func (d *ServicesData) collectInterceptors(service *serviceFacts, facts []*interceptorFacts, methods map[*methodFacts]*MethodData, resolver *declarationResolver, server bool) []*InterceptorData {
	res := make([]*InterceptorData, 0, len(facts))
	for _, interceptor := range facts {
		res = append(res, buildInterceptorData(service, interceptor, methods, resolver, server))
	}
	return res
}

// declarationContext configures transformations and validations to resolve
// every named service or view type through its planned package declaration.
func declarationContext(resolver codegen.Attributor, pointer bool) *codegen.AttributeContext {
	return &codegen.AttributeContext{
		Pointer:    pointer,
		UseDefault: true,
		Scope:      resolver,
	}
}

// formatUserTypeFacts resolves the final names and definitions of types whose
// reachability and declaration ownership were fixed during collection.
func formatUserTypeFacts(facts []*userTypeFacts, outputPath string, aliases *importAliases) []*UserTypeData {
	data := make([]*UserTypeData, len(facts))
	for index, facts := range facts {
		description := facts.description
		if description == "" && facts.location != nil {
			description = fmt.Sprintf("%s is a generated service type.", facts.declaration.Name())
		}
		definition := facts.layout.Link(
			facts.declaration.PackagePath(),
			retainedTypeQualifier(aliases),
		)
		reference := facts.reference.Link(
			outputPath,
			retainedTypeQualifier(aliases),
		)
		data[index] = &UserTypeData{
			Declaration:    facts.declaration,
			Name:           facts.name,
			VarName:        facts.declaration.Name(),
			Description:    description,
			ErrorName:      facts.errorName,
			IsServiceError: facts.serviceError,
			Def:            definition.Def(),
			Ref:            reference.Ref(),
			Loc:            facts.location,
			Type:           facts.userType,
		}
	}
	return data
}

// formatServiceUnions resolves the exact service union declarations retained
// during collection and registers one render record per generated package.
func (d *ServicesData) formatServiceUnions(facts *serviceFacts) ([]*UnionTypeData, error) {
	unions := make([]*UnionTypeData, 0, len(facts.unions))
	for _, facts := range facts.unions {
		union := buildRetainedUnionTypeData(facts, d.aliases)
		facts.data = union
		unions = append(unions, union)
	}
	sort.Slice(unions, func(i, j int) bool {
		if unions[i].Name != unions[j].Name {
			return unions[i].Name < unions[j].Name
		}
		var left, right string
		if unions[i].Loc != nil {
			left = unions[i].Loc.RelImportPath
		}
		if unions[j].Loc != nil {
			right = unions[j].Loc.RelImportPath
		}
		return left < right
	})
	return unions, nil
}

// formatViewUnions resolves the exact view union expressions retained while
// their declarations were collected. It does not traverse projected types.
func (d *ServicesData) formatViewUnions(facts *serviceFacts) ([]*UnionTypeData, error) {
	unions := make([]*UnionTypeData, len(facts.viewUnions))
	for index, union := range facts.viewUnions {
		unions[index] = buildRetainedUnionTypeData(union, d.aliases)
	}
	sort.Slice(unions, func(i, j int) bool {
		return unions[i].Name < unions[j].Name
	})
	return unions, nil
}

// buildRetainedUnionTypeData formats one union from the branch declarations
// and Go layouts selected before the generation froze.
func buildRetainedUnionTypeData(facts *unionFacts, aliases *importAliases) *UnionTypeData {
	fields := make([]*UnionFieldData, len(facts.branches))
	for index, branch := range facts.branches {
		fields[index] = &UnionFieldData{
			Name:               branch.name,
			KindConst:          branch.declaration.KindConst(),
			Constructor:        branch.declaration.Constructor(),
			FieldName:          branch.fieldName,
			FieldType:          branch.layout.Link(facts.declaration.PackagePath(), retainedTypeQualifier(aliases)).Ref(),
			Nilable:            branch.nilable,
			EmitPrimitiveAlias: branch.emitPrimitiveAlias,
			PrimitiveAliasType: branch.primitiveAliasType,
			TypeTag:            branch.name,
		}
	}
	return &UnionTypeData{
		Declaration: facts.declaration,
		Name:        facts.declaration.Name(),
		KindName:    facts.declaration.KindName(),
		Fields:      fields,
		Loc:         facts.location,
		TypeKey:     facts.typeKey,
		ValueKey:    facts.valueKey,
	}
}

// sortedNamedAttributes returns object fields sorted by attribute name.
// Union naming uses NameScope uniqueness, so callers that discover unions while
// traversing objects must use a deterministic field order to avoid oscillating
// generated identifiers across runs.
func sortedNamedAttributes(attrs []*expr.NamedAttributeExpr) []*expr.NamedAttributeExpr {
	if len(attrs) < 2 {
		return attrs
	}
	sorted := slices.Clone(attrs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})
	return sorted
}

// primitiveAliasGoType resolves the native Go type for a primitive alias branch.
// It uses expr.IsPrimitive to enforce the type contract and then unwraps aliases.
func primitiveAliasGoType(dt expr.DataType) (string, bool) {
	if !expr.IsPrimitive(dt) {
		return "", false
	}
	for {
		ut, ok := dt.(expr.UserType)
		if !ok {
			return codegen.GoNativeTypeName(dt), true
		}
		dt = ut.Attribute().Type
	}
}

// buildRetainedErrorInitData formats an error selected during collection
// without consulting the mutable design expression.
func buildRetainedErrorInitData(facts *errorRenderFacts, resolver *declarationResolver, declaration *codegen.NameDeclaration) *ErrorInitData {
	if facts.layout == nil {
		panic(fmt.Sprintf("retained error %q has no Go type layout", facts.name))
	}
	linked := facts.layout.Link(resolver.outputPath, retainedTypeQualifier(resolver.aliases))
	return &ErrorInitData{
		Declaration: declaration,
		Description: facts.description,
		ErrName:     facts.name,
		TypeName:    linked.Name(),
		TypeRef:     linked.Ref(),
		Temporary:   facts.temporary,
		Timeout:     facts.timeout,
		Fault:       facts.fault,
	}
}

// retainedTypeQualifier returns the frozen qualifier assigned to one retained
// Go type import.
func retainedTypeQualifier(aliases *importAliases) codegen.GoTypeQualifier {
	return func(importPath string) string {
		return aliases.name(importPath)
	}
}
