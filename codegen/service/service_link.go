// This file turns stored service information into the data used by templates
// after all generated names and imported package names are final.
package service

import (
	"fmt"
	"path"
	"slices"
	"sort"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// Link reads the final Go declaration and import names and builds the data used
// by service templates. Generation.Freeze must run first so each definition
// and every reference to it use the same name.
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

// Services returns the service data passed to templates. It panics before Link
// because that data does not exist until the final Go names are available.
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
	pkgName := strings.ToLower(codegen.Goify(path.Base(servicePackage.ImportPath()), false))
	var viewsPkg string
	seenErrors := make(map[string]struct{})
	type viewedResultKey struct {
		origin expr.UserType
		view   string
	}
	seenViewed := make(map[viewedResultKey]*ViewedResultTypeData)
	seenViewedDeclarations := make(map[*codegen.TypeDeclaration]struct{})
	viewDerived := make(map[expr.UserType]codegen.DerivedTypeID)
	serviceResolver := newServiceResolver(
		d.generation,
		d.aliases,
		facts.name,
		facts.packagePath,
		facts.packagePath,
	).withValidators(facts.validators)
	types = formatUserTypeFacts(facts.userTypes, d.aliases)
	errTypes = formatUserTypeFacts(facts.errorTypes, d.aliases)

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
		// Build template data for each result type containing only a view's fields.
		if projection := method.projection; projection != nil {
			viewsPkg = d.aliases.spec(facts.packagePath, facts.viewsPath).Name
			views := d.generation.Package(facts.viewsPath)
			for _, projectedFacts := range projection.types {
				pair := projectedFacts.pair
				identity := codegen.NewProjectedTypeID(pair.source)
				viewDerived[pair.projected.Origin()] = identity
			}
			viewResolver := newViewResolver(
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
					viewsPkg,
				)
				projTypes = append(projTypes, projectedType)
			}
		}
		for _, errorFacts := range method.errors {
			recordError(errorFacts)
		}
	}
	viewUnions := d.formatViewUnions(facts)

	var (
		methods []*MethodData
		schemes SchemesData
	)
	methods = make([]*MethodData, len(facts.orderedMethods))
	methodDataByFacts := make(map[*methodFacts]*MethodData, len(facts.orderedMethods))
	for i, method := range facts.orderedMethods {
		m := buildMethodData(method, serviceResolver, facts)
		methods[i] = m
		methodDataByFacts[method] = m
		for _, s := range m.Schemes {
			schemes = schemes.Append(s)
		}
		viewedFacts := method.viewedResult
		if viewedFacts == nil {
			continue
		}
		viewsPkg = d.aliases.spec(facts.packagePath, facts.viewsPath).Name
		key := viewedResultKey{origin: viewedFacts.origin, view: viewedFacts.viewName}
		if vrt, ok := seenViewed[key]; ok {
			m.ViewedResult = vrt
			continue
		}
		vrt := buildViewedResultType(
			viewedFacts,
			viewsPkg,
			serviceResolver,
			newViewResolver(
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

	unions := d.formatServiceUnions(facts)

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
		ServerInterceptorsDeclaration: facts.names[serviceSymbolID{
			role: serviceServerInterceptorsNameRole, service: facts.name,
		}].declaration,
		ClientInterceptorsDeclaration: facts.names[serviceSymbolID{
			role: serviceClientInterceptorsNameRole, service: facts.name,
		}].declaration,
		ExampleStructDeclaration:                        facts.exampleStruct,
		ExampleConstructorDeclaration:                   facts.exampleConstructor,
		ExampleServerInterceptorsConstructorDeclaration: facts.exampleServerConstructor,
		Name:               facts.name,
		Description:        desc,
		APIName:            d.facts.apiName,
		APIVersion:         d.facts.apiVersion,
		VarName:            varName,
		PathName:           path.Base(facts.packagePath),
		StructName:         codegen.Goify(facts.name, true),
		PkgName:            pkgName,
		ViewsPkg:           viewsPkg,
		Methods:            methods,
		Schemes:            schemes,
		ServerInterceptors: d.collectInterceptors(facts, facts.serverInterceptorFacts, methodDataByFacts, serviceResolver, true),
		ClientInterceptors: d.collectInterceptors(facts, facts.clientInterceptorFacts, methodDataByFacts, serviceResolver, false),
		Scope:              scope,
		ViewScope:          viewScope,
		errorTypes:         errTypes,
		errorInits:         errorInits,
		userTypes:          types,
		projectedTypes:     projTypes,
		viewedResultTypes:  viewedRTs,
		unions:             unions,
		viewUnions:         viewUnions,
		viewDerived:        viewDerived,
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

// formatUserTypeFacts resolves the final names and definitions of types that
// collection already selected and assigned to generated packages.
func formatUserTypeFacts(facts []*userTypeFacts, aliases *importAliases) []*UserTypeData {
	data := make([]*UserTypeData, len(facts))
	for index, facts := range facts {
		linked := facts.layout.Link(
			facts.declaration.PackagePath(),
			retainedTypeQualifier(aliases, facts.declaration.PackagePath()),
		)
		data[index] = &UserTypeData{
			Declaration:    facts.declaration,
			Name:           facts.name,
			VarName:        facts.declaration.Name(),
			Description:    facts.description,
			ErrorName:      facts.errorName,
			IsServiceError: facts.serviceError,
			Def:            linked.Def(),
			Ref:            facts.declaration.Ref(facts.userType),
			Loc:            facts.location,
			Type:           facts.userType,
		}
	}
	return data
}

// formatServiceUnions builds template data for the Goa OneOf declarations
// recorded during collection and adds one entry for each generated package.
func (d *ServicesData) formatServiceUnions(facts *serviceFacts) []*UnionTypeData {
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
	return unions
}

// formatViewUnions builds template data for the Goa OneOf declarations found
// while collecting result views. It does not walk the result types again.
func (d *ServicesData) formatViewUnions(facts *serviceFacts) []*UnionTypeData {
	unions := make([]*UnionTypeData, len(facts.viewUnions))
	for index, union := range facts.viewUnions {
		unions[index] = buildRetainedUnionTypeData(union, d.aliases)
	}
	sort.Slice(unions, func(i, j int) bool {
		return unions[i].Name < unions[j].Name
	})
	return unions
}

// This helper builds template data for one Goa OneOf type from the branch names
// and Go types selected during planning.
func buildRetainedUnionTypeData(facts *unionFacts, aliases *importAliases) *UnionTypeData {
	fields := make([]*UnionFieldData, len(facts.branches))
	for index, branch := range facts.branches {
		fields[index] = &UnionFieldData{
			Name:                   branch.name,
			KindConst:              branch.declaration.KindConst(),
			Constructor:            branch.declaration.Constructor(),
			KindDeclaration:        branch.declaration.KindDeclaration(),
			ConstructorDeclaration: branch.declaration.ConstructorDeclaration(),
			FieldName:              branch.fieldName,
			StorageName:            branch.storageName,
			FieldType:              branch.layout.Link(facts.declaration.PackagePath(), retainedTypeQualifier(aliases, facts.declaration.PackagePath())).Ref(),
			Nilable:                branch.nilable,
			EmitPrimitiveAlias:     branch.emitPrimitiveAlias,
			PrimitiveAliasType:     branch.primitiveAliasType,
			TypeTag:                branch.name,
		}
	}
	return &UnionTypeData{
		TypeDeclaration: facts.declaration.Declaration(),
		KindDeclaration: facts.declaration.KindDeclaration(),
		Name:            facts.declaration.Name(),
		KindName:        facts.declaration.KindName(),
		Fields:          fields,
		Loc:             facts.location,
		TypeKey:         facts.typeKey,
		ValueKey:        facts.valueKey,
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

// This helper builds constructor data for an error copied during collection
// without reading the design expression again.
func buildRetainedErrorInitData(facts *errorRenderFacts, resolver *declarationResolver, declaration *codegen.NameDeclaration) *ErrorInitData {
	if facts.layout == nil {
		panic(fmt.Sprintf("retained error %q has no Go type layout", facts.name))
	}
	linked := facts.layout.Link(resolver.outputPath, retainedTypeQualifier(resolver.aliases, resolver.outputPath))
	name := ""
	if facts.serviceType {
		name = declaration.Name()
	}
	return &ErrorInitData{
		Declaration: declaration,
		Name:        name,
		Description: facts.description,
		ErrName:     facts.name,
		TypeName:    linked.Name(),
		TypeRef:     linked.Ref(),
		Temporary:   facts.temporary,
		Timeout:     facts.timeout,
		Fault:       facts.fault,
	}
}

// This helper returns the Go import name chosen for a type recorded during
// planning.
func retainedTypeQualifier(aliases *importAliases, outputPackage string) codegen.GoTypeQualifier {
	return func(importPath string) string {
		return aliases.name(outputPackage, importPath)
	}
}
