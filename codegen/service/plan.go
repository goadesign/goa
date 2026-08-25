// This file stores the service declarations and file data selected for one
// generation run. File writers use this stored data without rereading the
// design.
package service

import (
	"path"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

type (
	// PlanInput supplies one evaluated Goa design and the generator used to make
	// examples for types in that design.
	PlanInput struct {
		// Root is one evaluated Goa design included in this generation command.
		Root *expr.RootExpr
		// Examples produces the values written as examples for Root.
		Examples *expr.ExampleGenerator
	}

	// Plan stores one design's services, generated Go declarations, and template
	// data from name collection through file rendering.
	Plan struct {
		generation *codegen.Generation
		facts      *rootFacts
		services   *ServicesData
	}

	// rootFacts stores the services and API values selected from one design.
	// Copying its slices prevents later steps from walking the design again and
	// finding a different set of services.
	rootFacts struct {
		root                *expr.RootExpr
		apiName             string
		apiVersion          string
		examplePackageName  string
		exampleImports      []*codegen.ImportSpec
		services            []*serviceFacts
		serviceByID         map[string]*serviceFacts
		types               []expr.UserType
		rootTypes           *rootTypeSet
		examples            *expr.ExampleGenerator
		externalConversions []*externalConversionFileFacts
		generatedTypes      []*generatedTypeEmissionFacts
		generatedUnions     []*generatedUnionEmissionFacts
	}

	// serviceFacts stores the service values needed to name and write its files.
	serviceFacts struct {
		service                  *expr.ServiceExpr
		apiName                  string
		name                     string
		description              string
		packagePath              string
		viewsPath                string
		packageImport            *codegen.ImportSpec
		viewsImport              *codegen.ImportSpec
		methods                  []*expr.MethodExpr
		orderedMethods           []*methodFacts
		methodByExpr             map[*expr.MethodExpr]*methodFacts
		errors                   []*expr.ErrorExpr
		errorFacts               []*errorRenderFacts
		serverInterceptors       []*expr.InterceptorExpr
		clientInterceptors       []*expr.InterceptorExpr
		serverInterceptorFacts   []*interceptorFacts
		clientInterceptorFacts   []*interceptorFacts
		referenceAttributes      []*expr.AttributeExpr
		reachableTypes           map[expr.UserType]struct{}
		projections              map[*expr.MethodExpr]*projectionFacts
		userTypes                []*userTypeFacts
		errorTypes               []*userTypeFacts
		unions                   []*unionFacts
		viewUnions               []*unionFacts
		names                    serviceNames
		validators               map[validatorKey]*codegen.NameDeclaration
		errorConstructors        map[string]*codegen.NameDeclaration
		generatedTypeImports     map[*codegen.TypeDeclaration]*retainedFileImports
		exampleStruct            *codegen.NameDeclaration
		exampleConstructor       *codegen.NameDeclaration
		exampleServerStruct      *codegen.NameDeclaration
		exampleServerConstructor *codegen.NameDeclaration
		exampleClientStruct      *codegen.NameDeclaration
		exampleClientConstructor *codegen.NameDeclaration
		imports                  serviceFileImports
		data                     *Data
	}

	// methodFacts stores the method values used by service and transport files.
	methodFacts struct {
		method                       *expr.MethodExpr
		serviceName                  string
		name                         string
		description                  string
		idempotent                   bool
		payload                      *methodAttributeFacts
		result                       *methodAttributeFacts
		streamingPayload             *methodAttributeFacts
		streamingResult              *methodAttributeFacts
		errors                       []*errorRenderFacts
		requirements                 RequirementsData
		schemes                      SchemesData
		streamKind                   expr.StreamKind
		isStreaming                  bool
		hasMixedResults              bool
		varName                      string
		serverStreamVarName          string
		clientStreamVarName          string
		endpointField                string
		streamEndpointField          string
		viewedResult                 *viewedResultFacts
		projection                   *projectionFacts
		skipRequestBodyEncodeDecode  bool
		skipResponseBodyEncodeDecode bool
	}

	// methodAttributeFacts stores one payload or result description, default, and
	// example. GoTypePlan separately stores the Go fields nested inside it.
	methodAttributeFacts struct {
		attribute    *expr.AttributeExpr
		layout       *codegen.GoTypePlan
		definition   *codegen.GoTypePlan
		normalized   bool
		present      bool
		isObject     bool
		location     *codegen.Location
		description  string
		defaultValue any
		example      any
	}

	// errorRenderFacts stores the error type, description, and marker fields
	// written to service, client, and endpoint files.
	errorRenderFacts struct {
		attribute   *expr.AttributeExpr
		layout      *codegen.GoTypePlan
		name        string
		description string
		location    *codegen.Location
		temporary   bool
		timeout     bool
		fault       bool
		serviceType bool
	}

	// interceptorFacts stores the methods that call one interceptor on either the
	// client or server side.
	interceptorFacts struct {
		name                        string
		description                 string
		readPayload                 *expr.AttributeExpr
		writePayload                *expr.AttributeExpr
		readResult                  *expr.AttributeExpr
		writeResult                 *expr.AttributeExpr
		readStreamingPayload        *expr.AttributeExpr
		writeStreamingPayload       *expr.AttributeExpr
		readStreamingResult         *expr.AttributeExpr
		writeStreamingResult        *expr.AttributeExpr
		readPayloadFields           []*interceptorAccessFacts
		writePayloadFields          []*interceptorAccessFacts
		readResultFields            []*interceptorAccessFacts
		writeResultFields           []*interceptorAccessFacts
		readStreamingPayloadFields  []*interceptorAccessFacts
		writeStreamingPayloadFields []*interceptorAccessFacts
		readStreamingResultFields   []*interceptorAccessFacts
		writeStreamingResultFields  []*interceptorAccessFacts
		methods                     []*methodFacts
	}

	// interceptorAccessFacts stores one field exposed to an interceptor and the
	// Go type written for that field.
	interceptorAccessFacts struct {
		attribute *expr.AttributeExpr
		name      string
		pointer   bool
		layout    *codegen.GoTypePlan
	}

	// projectionFacts stores copies of one method's result types containing only
	// the fields in each selected view. Name collection and template data both
	// read these same copies.
	projectionFacts struct {
		pairs []*projectedTypePair
		types []*projectedTypeFacts
	}

	// projectedTypeFacts stores one result type containing the fields selected by
	// a view, together with the validation and conversion code generated for it.
	projectedTypeFacts struct {
		pair           *projectedTypePair
		projectedType  expr.UserType
		projected      *codegen.GoTypePlan
		definition     *codegen.GoTypePlan
		source         *codegen.GoTypePlan
		resultType     bool
		views          []*viewRenderFacts
		validations    []*validationFacts
		conversions    []*viewConversionFacts
		mapDeclaration *codegen.NameDeclaration
		declaration    *codegen.TypeDeclaration
	}

	// viewRenderFacts stores the description and ordered field names from one
	// declared result view for the service and views templates.
	viewRenderFacts struct {
		name        string
		description string
		attributes  []string
	}

	// validationFacts stores the field checks and child validation calls emitted
	// by one view-specific validation function. Function names are added later.
	validationFacts struct {
		viewName       string
		attribute      *expr.AttributeExpr
		layout         *codegen.GoTypePlan
		plan           *codegen.ValidationPlan
		declaration    *codegen.NameDeclaration
		needed         bool
		alias          bool
		pointer        bool
		collectionElem *expr.AttributeExpr
		collectionCall *codegen.NameDeclaration
		fields         []*validationFieldFacts
	}

	// validationFieldFacts stores one child result field and the validation call
	// emitted for it.
	validationFieldFacts struct {
		name      string
		attribute *expr.AttributeExpr
		view      string
		required  bool
		call      *codegen.NameDeclaration
	}

	// viewValidationKey identifies one result type and view while Goa decides
	// whether its generated validation function can return an error.
	viewValidationKey struct {
		origin expr.UserType
		view   string
	}

	// viewConversionFacts stores one conversion between a service result and a
	// selected result view, including conversions for nested fields.
	viewConversionFacts struct {
		toResult        bool
		viewName        string
		source          *expr.AttributeExpr
		target          *expr.AttributeExpr
		transformTarget *expr.AttributeExpr
		fields          []*viewConversionFieldFacts
		plan            *codegen.TransformPlan
		targetLayout    *codegen.GoTypePlan
		collection      bool
		contextType     expr.UserType
		contextIdentity codegen.DerivedTypeID
		elementType     expr.UserType
		elementIdentity codegen.DerivedTypeID
		constructor     *codegen.NameDeclaration
		elementCall     *codegen.NameDeclaration
	}

	// viewConversionFieldFacts stores one child result constructor call emitted
	// separately from the general type conversion.
	viewConversionFieldFacts struct {
		name      string
		attribute *expr.AttributeExpr
		view      string
		call      *codegen.NameDeclaration
	}

	// viewedResultFacts stores the wrapper type and selected view written for one
	// method result.
	viewedResultFacts struct {
		wrapped         expr.UserType
		wrappedLayout   *codegen.GoTypePlan
		wrappedDef      *codegen.GoTypePlan
		projected       *projectedTypeFacts
		origin          expr.UserType
		source          *methodAttributeFacts
		viewName        string
		views           []*viewRenderFacts
		conversions     []*viewConversionFacts
		toViewed        *codegen.NameDeclaration
		toResult        *codegen.NameDeclaration
		mapDeclaration  *codegen.NameDeclaration
		declaration     *codegen.TypeDeclaration
		validator       *codegen.NameDeclaration
		validationCalls []*codegen.NameDeclaration
		isCollection    bool
	}

	// userTypeFacts records one selected design type, its generated Go type, and
	// the file location inherited from an enclosing type when needed.
	userTypeFacts struct {
		userType     expr.UserType
		name         string
		description  string
		errorName    string
		serviceError bool
		location     *codegen.Location
		declaration  *codegen.TypeDeclaration
		layout       *codegen.GoTypePlan
		imports      retainedFileImports
	}

	// unionFacts records one Goa OneOf type and its generated Go declaration.
	unionFacts struct {
		attribute   *expr.AttributeExpr
		union       *expr.Union
		identity    codegen.UnionDeclarationID
		typeKey     string
		valueKey    string
		branches    []*unionBranchFacts
		location    *codegen.Location
		declaration *codegen.UnionDeclaration
		imports     retainedFileImports
		data        *UnionTypeData
	}

	// unionBranchFacts stores one Goa OneOf branch, its generated names, and the
	// Go type written for its value.
	unionBranchFacts struct {
		name               string
		fieldName          string
		storageName        string
		declaration        *codegen.UnionBranchDeclaration
		layout             *codegen.GoTypePlan
		nilable            bool
		emitPrimitiveAlias bool
		primitiveAliasType string
	}

	// validatorKey selects the validation function for one generated result type
	// and view. Validation code for view-specific result copies uses that exact
	// function.
	validatorKey struct {
		declaration *codegen.TypeDeclaration
		view        string
	}

	// viewConversionCallKey identifies one private constructor by the source
	// result declaration, selected view, and conversion direction.
	viewConversionCallKey struct {
		origin   expr.UserType
		view     string
		toResult bool
	}

	// streamWrapperKey identifies the client or server wrapper for one method's
	// stream.
	streamWrapperKey struct {
		method *expr.MethodExpr
		server bool
	}
)

// collectServiceNames submits every package-level Go declaration written for
// one service and its views before Generation.Freeze chooses the final Go
// names.
func collectServiceNames(facts *serviceFacts, rootTypes *rootTypeSet, generation *codegen.Generation) error {
	service := facts.service
	serviceName := service.Name
	servicePackage := generation.Package(facts.packagePath)
	viewsPackage := generation.Package(facts.viewsPath)
	examplePackage, err := generation.ClaimOutputPackage(path.Dir(generation.GenPkg()), ".")
	if err != nil {
		return err
	}
	exampleInterceptorsPackage, err := generation.ClaimOutputPackage(
		path.Join(path.Dir(generation.GenPkg()), "interceptors"),
		"interceptors",
	)
	if err != nil {
		return err
	}
	facts.names = make(serviceNames)
	facts.validators = make(map[validatorKey]*codegen.NameDeclaration)
	facts.errorConstructors = make(map[string]*codegen.NameDeclaration)
	declare := func(pkg *codegen.GeneratedPackage, role serviceNameRole, preferred string, id serviceSymbolID) error {
		id.role = role
		id.service = serviceName
		_, err := facts.names.declareForAPI(pkg, id, preferred, facts.apiName)
		return err
	}
	static := []struct {
		role      serviceNameRole
		preferred string
	}{
		{serviceInterfaceNameRole, "Service"},
		{serviceAPINameRole, "APIName"},
		{serviceAPIVersionNameRole, "APIVersion"},
		{serviceNameConstantRole, "ServiceName"},
		{serviceMethodNamesRole, "MethodNames"},
		{serviceEndpointsNameRole, "Endpoints"},
		{serviceNewEndpointsNameRole, "NewEndpoints"},
		{serviceClientNameRole, "Client"},
		{serviceNewClientNameRole, "NewClient"},
	}
	if serviceHasSchemes(facts) {
		static = append(static, struct {
			role      serviceNameRole
			preferred string
		}{serviceAutherNameRole, "Auther"}) //nolint:misspell // Keep Goa's existing generated interface name.
	}
	for _, symbol := range static {
		if err := declare(servicePackage, symbol.role, symbol.preferred, serviceSymbolID{}); err != nil {
			return err
		}
	}
	facts.exampleStruct, err = facts.names.declareForAPI(examplePackage, serviceSymbolID{
		role: serviceExampleStructNameRole, service: serviceName,
	}, codegen.Goify(serviceName, false)+"srvc", facts.apiName)
	if err != nil {
		return err
	}
	facts.exampleConstructor, err = facts.names.declareForAPI(examplePackage, serviceSymbolID{
		role: serviceExampleConstructorNameRole, service: serviceName,
	}, "New"+codegen.Goify(serviceName, true), facts.apiName)
	if err != nil {
		return err
	}
	structName := codegen.Goify(serviceName, true)
	if len(facts.serverInterceptors) > 0 {
		facts.exampleServerStruct, err = facts.names.declareForAPI(exampleInterceptorsPackage, serviceSymbolID{
			role: serviceExampleServerInterceptorsStructNameRole, service: serviceName,
		}, structName+"ServerInterceptors", facts.apiName)
		if err != nil {
			return err
		}
		facts.exampleServerConstructor, err = facts.names.declareForAPI(exampleInterceptorsPackage, serviceSymbolID{
			role: serviceExampleServerInterceptorsConstructorNameRole, service: serviceName,
		}, "New"+structName+"ServerInterceptors", facts.apiName)
		if err != nil {
			return err
		}
	}
	if len(facts.clientInterceptors) > 0 {
		facts.exampleClientStruct, err = facts.names.declareForAPI(exampleInterceptorsPackage, serviceSymbolID{
			role: serviceExampleClientInterceptorsStructNameRole, service: serviceName,
		}, structName+"ClientInterceptors", facts.apiName)
		if err != nil {
			return err
		}
		facts.exampleClientConstructor, err = facts.names.declareForAPI(exampleInterceptorsPackage, serviceSymbolID{
			role: serviceExampleClientInterceptorsConstructorNameRole, service: serviceName,
		}, "New"+structName+"ClientInterceptors", facts.apiName)
		if err != nil {
			return err
		}
	}
	for _, method := range facts.methods {
		methodFacts := facts.methodByExpr[method]
		methodID := serviceSymbolID{method: methodFacts.varName}
		if method.IsStreaming() || method.HasMixedResults() {
			if err := declare(servicePackage, serviceServerStreamNameRole, methodFacts.varName+"ServerStream", methodID); err != nil {
				return err
			}
			if err := declare(servicePackage, serviceClientStreamNameRole, methodFacts.varName+"ClientStream", methodID); err != nil {
				return err
			}
			if err := declare(servicePackage, serviceEndpointInputNameRole, methodFacts.varName+"EndpointInput", methodID); err != nil {
				return err
			}
		}
		if err := declare(servicePackage, serviceMethodEndpointNameRole, "New"+methodFacts.varName+"Endpoint", methodID); err != nil {
			return err
		}
		if methodFacts.skipRequestBodyEncodeDecode {
			if err := declare(servicePackage, serviceRequestNameRole, methodFacts.varName+"RequestData", methodID); err != nil {
				return err
			}
		}
		if methodFacts.skipResponseBodyEncodeDecode {
			if err := declare(servicePackage, serviceResponseNameRole, methodFacts.varName+"ResponseData", methodID); err != nil {
				return err
			}
		}
		if len(method.ServerInterceptors) > 0 {
			if err := declare(servicePackage, serviceServerEndpointWrapperNameRole, "Wrap"+methodFacts.varName+"Endpoint", methodID); err != nil {
				return err
			}
		}
		if len(method.ClientInterceptors) > 0 {
			if err := declare(servicePackage, serviceClientEndpointWrapperNameRole, "Wrap"+methodFacts.varName+"ClientEndpoint", methodID); err != nil {
				return err
			}
		}
	}
	if err := collectErrorNames(facts, servicePackage); err != nil {
		return err
	}
	if err := collectInterceptorNames(facts, servicePackage); err != nil {
		return err
	}
	return collectViewNames(facts, servicePackage, viewsPackage, rootTypes, generation)
}

// serviceHasSchemes reports whether any selected method needs generated
// authorization functions.
func serviceHasSchemes(facts *serviceFacts) bool {
	for _, method := range facts.methods {
		if len(method.Requirements) > 0 {
			return true
		}
	}
	return false
}

// collectErrorNames declares the constructors emitted for distinct Goa
// service errors shared by service-level and method-level declarations.
func collectErrorNames(facts *serviceFacts, servicePackage *codegen.GeneratedPackage) error {
	seen := make(map[string]struct{})
	errors := append([]*expr.ErrorExpr(nil), facts.errors...)
	for _, method := range facts.methods {
		errors = append(errors, method.Errors...)
	}
	for _, serviceError := range errors {
		if !expr.IsErrorResult(serviceError.Type) {
			continue
		}
		if _, exists := seen[serviceError.Name]; exists {
			continue
		}
		seen[serviceError.Name] = struct{}{}
		declaration, err := facts.names.declareForAPI(servicePackage, serviceSymbolID{
			role:    serviceErrorConstructorNameRole,
			service: facts.service.Name,
			subject: serviceError.Name,
		}, "Make"+codegen.Goify(serviceError.Name, true), facts.apiName)
		if err != nil {
			return err
		}
		facts.errorConstructors[serviceError.Name] = declaration
	}
	return nil
}

// collectInterceptorNames declares interceptor interfaces, typed accessors,
// wrappers, and stream wrapper structs in the service package.
func collectInterceptorNames(facts *serviceFacts, servicePackage *codegen.GeneratedPackage) error {
	declare := func(role serviceNameRole, preferred string, id serviceSymbolID) error {
		id.role = role
		id.service = facts.service.Name
		_, err := facts.names.declareForAPI(servicePackage, id, preferred, facts.apiName)
		return err
	}
	if len(facts.serverInterceptors) > 0 {
		if err := declare(serviceServerInterceptorsNameRole, "ServerInterceptors", serviceSymbolID{}); err != nil {
			return err
		}
	}
	if len(facts.clientInterceptors) > 0 {
		if err := declare(serviceClientInterceptorsNameRole, "ClientInterceptors", serviceSymbolID{}); err != nil {
			return err
		}
	}
	interceptors := append(append([]*expr.InterceptorExpr(nil), facts.serverInterceptors...), facts.clientInterceptors...)
	seenInterceptors := make(map[string]struct{})
	seenStreams := make(map[streamWrapperKey]struct{})
	for _, interceptor := range interceptors {
		if _, exists := seenInterceptors[interceptor.Name]; !exists {
			seenInterceptors[interceptor.Name] = struct{}{}
			base := codegen.Goify(interceptor.Name, true)
			for _, symbol := range []struct {
				role   serviceNameRole
				suffix string
				emit   bool
			}{
				{serviceInterceptorInfoNameRole, "Info", true},
				{serviceInterceptorPayloadNameRole, "Payload", interceptor.ReadPayload != nil || interceptor.WritePayload != nil},
				{serviceInterceptorResultNameRole, "Result", interceptor.ReadResult != nil || interceptor.WriteResult != nil},
				{serviceInterceptorStreamingPayloadNameRole, "StreamingPayload", interceptor.ReadStreamingPayload != nil || interceptor.WriteStreamingPayload != nil},
				{serviceInterceptorStreamingResultNameRole, "StreamingResult", interceptor.ReadStreamingResult != nil || interceptor.WriteStreamingResult != nil},
			} {
				if !symbol.emit {
					continue
				}
				if err := declare(symbol.role, base+symbol.suffix, serviceSymbolID{subject: interceptor.Name}); err != nil {
					return err
				}
			}
		}
		for _, method := range facts.methods {
			server := interceptorNamed(method.ServerInterceptors, interceptor.Name)
			client := interceptorNamed(method.ClientInterceptors, interceptor.Name)
			if !server && !client {
				continue
			}
			methodName := facts.methodByExpr[method].varName
			base := codegen.Goify(codegen.SnakeCase(interceptor.Name), false) + methodName
			methodID := serviceSymbolID{method: facts.methodByExpr[method].varName, subject: interceptor.Name}
			streamingAccess := interceptorHasStreamingAccess(interceptor) && (method.IsStreaming() || method.HasMixedResults())
			hasPayloadAccess := interceptor.ReadPayload != nil || interceptor.WritePayload != nil
			hasStreamingPayloadAccess := interceptor.ReadStreamingPayload != nil || interceptor.WriteStreamingPayload != nil
			hasStreamingResultAccess := interceptor.ReadStreamingResult != nil || interceptor.WriteStreamingResult != nil
			for _, symbol := range []struct {
				role   serviceNameRole
				suffix string
				emit   bool
			}{
				{serviceInterceptorPayloadAccessNameRole, "Payload", hasPayloadAccess},
				{serviceInterceptorResultAccessNameRole, "Result", interceptor.ReadResult != nil || interceptor.WriteResult != nil},
				{serviceInterceptorStreamingPayloadAccessNameRole, "StreamingPayload", hasStreamingPayloadAccess},
				{serviceInterceptorStreamingResultAccessNameRole, "StreamingResult", hasStreamingResultAccess},
				{serviceInterceptorMethodInfoNameRole, "Info", true},
				{serviceInterceptorServerUnaryInfoNameRole, "ServerUnaryInfo", server && (!streamingAccess || hasPayloadAccess)},
				{serviceInterceptorClientUnaryInfoNameRole, "ClientUnaryInfo", client && (!streamingAccess || hasPayloadAccess)},
				{serviceInterceptorStreamingSendInfoNameRole, "StreamingSendInfo", streamingAccess && (server && hasStreamingResultAccess || client && hasStreamingPayloadAccess)},
				{serviceInterceptorStreamingRecvInfoNameRole, "StreamingRecvInfo", streamingAccess && (server && hasStreamingPayloadAccess || client && hasStreamingResultAccess)},
			} {
				if !symbol.emit {
					continue
				}
				if err := declare(symbol.role, base+symbol.suffix, methodID); err != nil {
					return err
				}
			}
			if server {
				if err := declare(serviceServerInterceptorWrapperNameRole, "wrap"+methodName+codegen.Goify(interceptor.Name, true), methodID); err != nil {
					return err
				}
			}
			if client {
				if err := declare(serviceClientInterceptorWrapperNameRole, "wrapClient"+methodName+codegen.Goify(interceptor.Name, true), methodID); err != nil {
					return err
				}
			}
			if (!method.IsStreaming() && !method.HasMixedResults()) || !interceptorHasStreamingAccess(interceptor) {
				continue
			}
			for _, side := range []struct {
				server bool
				role   serviceNameRole
				name   string
			}{
				{true, serviceServerStreamWrapperNameRole, "wrapped" + methodName + "ServerStream"},
				{false, serviceClientStreamWrapperNameRole, "wrapped" + methodName + "ClientStream"},
			} {
				key := streamWrapperKey{method: method, server: side.server}
				if _, exists := seenStreams[key]; exists || side.server && !server || !side.server && !client {
					continue
				}
				seenStreams[key] = struct{}{}
				if err := declare(side.role, side.name, serviceSymbolID{method: methodName}); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// collectViewNames declares validators and constructor/map companions from
// the exact service and view type declarations allocated during view planning.
func collectViewNames(facts *serviceFacts, servicePackage, viewsPackage *codegen.GeneratedPackage, rootTypes *rootTypeSet, generation *codegen.Generation) error {
	markNeededViewValidators(facts)
	for _, method := range facts.methods {
		projection := facts.projections[method]
		if projection == nil {
			continue
		}
		for _, projectedFacts := range projection.types {
			pair := projectedFacts.pair
			declaration, err := viewsPackage.DerivedType(codegen.NewProjectedTypeID(pair.source))
			if err != nil {
				return err
			}
			projectedFacts.declaration = declaration
			for _, validation := range projectedFacts.validations {
				if !validation.needed {
					continue
				}
				view := canonicalValidatorView(validation.viewName)
				suffix := ""
				if view != "" {
					suffix = codegen.Goify(view, true)
				}
				key := validatorKey{declaration: declaration, view: view}
				if facts.validators[key] != nil {
					continue
				}
				id := serviceSymbolID{
					role:    serviceValidatorNameRole,
					service: facts.service.Name,
					subject: pair.source.ID(),
					source:  pair.source.Name(),
					view:    view,
					side:    "projected",
				}
				validator, err := facts.names.declareDependentForAPI(viewsPackage, id, declaration.Declaration(), "Validate", suffix, facts.apiName)
				if err != nil {
					return err
				}
				facts.validators[key] = validator
				validation.declaration = validator
			}
			if _, ok := pair.projected.(*expr.ResultTypeExpr); ok {
				projectedFacts.mapDeclaration, err = facts.names.declareForAPI(viewsPackage, serviceSymbolID{
					role:    serviceViewMapNameRole,
					service: facts.service.Name,
					subject: pair.source.ID(),
					source:  pair.source.Name(),
				}, codegen.Goify(pair.source.Name(), true)+"Map", facts.apiName)
				if err != nil {
					return err
				}
			}
			for _, conversion := range projectedFacts.conversions {
				side := "to-projected"
				preferredBase := codegen.Goify(pair.projected.Name(), true)
				if conversion.toResult {
					side = "to-result"
					preferredBase = codegen.Goify(pair.source.Name(), true)
				}
				suffix := ""
				if conversion.viewName != expr.DefaultView {
					suffix = codegen.Goify(conversion.viewName, true)
				}
				conversion.constructor, err = facts.names.declareForAPI(servicePackage, serviceSymbolID{
					role:    servicePrivateProjectionConstructorNameRole,
					service: facts.service.Name,
					subject: pair.source.ID(),
					source:  pair.source.Name(),
					view:    canonicalValidatorView(conversion.viewName),
					side:    side,
				}, "new"+preferredBase+suffix, facts.apiName)
				if err != nil {
					return err
				}
				if conversion.plan == nil {
					continue
				}
				for _, definition := range conversion.plan.HelperDefinitions() {
					sourceName, sourceID := transformDataTypeName(definition.Source.Type)
					targetName, targetID := transformDataTypeName(definition.Target.Type)
					sourcePreferred := sourceName
					targetPreferred := targetName
					viewsPackageName := strings.ToLower(codegen.Goify(facts.service.Name, false)) + "views"
					if conversion.toResult {
						sourcePreferred = viewsPackageName + codegen.Goify(sourceName, true)
					} else {
						targetPreferred = viewsPackageName + codegen.Goify(targetName, true)
					}
					declaration, err := facts.names.declareForAPI(servicePackage, serviceSymbolID{
						role:       serviceTransformHelperNameRole,
						service:    facts.service.Name,
						subject:    pair.source.ID(),
						view:       canonicalValidatorView(conversion.viewName),
						source:     sourceID,
						target:     targetID,
						side:       side,
						definition: definition.Location,
					}, "transform"+codegen.Goify(sourcePreferred, true)+"To"+codegen.Goify(targetPreferred, true), facts.apiName)
					if err != nil {
						return err
					}
					if err := conversion.plan.BindHelperDefinition(definition.ID, declaration); err != nil {
						return err
					}
				}
			}
		}
		resultType, hasViews := method.Result.Type.(*expr.ResultTypeExpr)
		if !hasViews {
			continue
		}
		for _, projected := range projection.types {
			if projected.pair.source.Origin() == resultType.Origin() {
				facts.methodByExpr[method].viewedResult.conversions = projected.conversions
				break
			}
		}
		viewedDeclaration, err := viewsPackage.DerivedType(codegen.NewViewedResultTypeID(resultType))
		if err != nil {
			return err
		}
		facts.methodByExpr[method].viewedResult.declaration = viewedDeclaration
		viewedValidatorKey := validatorKey{declaration: viewedDeclaration}
		if facts.validators[viewedValidatorKey] == nil {
			validatorID := serviceSymbolID{
				role:    serviceValidatorNameRole,
				service: facts.service.Name,
				method:  facts.methodByExpr[method].varName,
				subject: resultType.ID(),
				source:  resultType.Name(),
				side:    "viewed",
			}
			validator, err := facts.names.declareDependentForAPI(viewsPackage, validatorID, viewedDeclaration.Declaration(), "Validate", "", facts.apiName)
			if err != nil {
				return err
			}
			facts.validators[viewedValidatorKey] = validator
		}
		facts.methodByExpr[method].viewedResult.validator = facts.validators[viewedValidatorKey]
		for _, symbol := range []struct {
			role   serviceNameRole
			prefix string
			side   string
		}{
			{serviceViewConstructorNameRole, "NewViewed", "to-viewed"},
			{serviceViewConstructorNameRole, "New", "to-result"},
		} {
			constructor, err := facts.names.declareForAPI(servicePackage, serviceSymbolID{
				role:    symbol.role,
				service: facts.service.Name,
				subject: resultType.ID(),
				source:  resultType.Name(),
				side:    symbol.side,
			}, symbol.prefix+codegen.Goify(resultType.Name(), true), facts.apiName)
			if err != nil {
				return err
			}
			if symbol.side == "to-viewed" {
				facts.methodByExpr[method].viewedResult.toViewed = constructor
			} else {
				facts.methodByExpr[method].viewedResult.toResult = constructor
			}
		}
		facts.methodByExpr[method].viewedResult.mapDeclaration, err = facts.names.declareForAPI(viewsPackage, serviceSymbolID{
			role:    serviceViewMapNameRole,
			service: facts.service.Name,
			subject: resultType.ID(),
			source:  resultType.Name(),
		}, codegen.Goify(resultType.Name(), true)+"Map", facts.apiName)
		if err != nil {
			return err
		}
		viewedFacts := facts.methodByExpr[method].viewedResult
		for _, view := range viewedFacts.views {
			declaration := facts.validators[validatorKey{
				declaration: viewedFacts.projected.declaration,
				view:        canonicalValidatorView(view.name),
			}]
			viewedFacts.validationCalls = append(viewedFacts.validationCalls, declaration)
		}
	}
	linkViewConversionCalls(facts)
	return planServiceValidations(facts, rootTypes, generation)
}

// linkViewConversionCalls gives collection and child constructor calls the Go
// function names chosen for their result type and view.
func linkViewConversionCalls(facts *serviceFacts) {
	lookup := make(map[viewConversionCallKey]*codegen.NameDeclaration)
	for _, method := range facts.methods {
		projection := facts.projections[method]
		if projection == nil {
			continue
		}
		for _, projected := range projection.types {
			for _, conversion := range projected.conversions {
				origin := projected.pair.projected.Origin()
				if conversion.toResult {
					origin = projected.pair.source.Origin()
				}
				lookup[viewConversionCallKey{
					origin:   origin,
					view:     canonicalValidatorView(conversion.viewName),
					toResult: conversion.toResult,
				}] = conversion.constructor
			}
		}
	}
	for _, method := range facts.methods {
		projection := facts.projections[method]
		if projection == nil {
			continue
		}
		for _, projected := range projection.types {
			for _, conversion := range projected.conversions {
				collection := projected.pair.projected
				if conversion.toResult {
					collection = projected.pair.source
				}
				if array := expr.AsArray(collection); array != nil {
					userType := array.ElemType.Type.(expr.UserType)
					conversion.elementCall = lookup[viewConversionCallKey{
						origin:   userType.Origin(),
						view:     canonicalValidatorView(conversion.viewName),
						toResult: conversion.toResult,
					}]
				}
				for _, field := range conversion.fields {
					userType := field.attribute.Type.(expr.UserType)
					field.call = lookup[viewConversionCallKey{
						origin:   userType.Origin(),
						view:     canonicalValidatorView(field.view),
						toResult: conversion.toResult,
					}]
				}
			}
		}
	}
}

// interceptorHasStreamingAccess reports whether interceptor causes a wrapped
// stream implementation to be emitted.
func interceptorHasStreamingAccess(interceptor *expr.InterceptorExpr) bool {
	return interceptor.ReadStreamingPayload != nil || interceptor.WriteStreamingPayload != nil ||
		interceptor.ReadStreamingResult != nil || interceptor.WriteStreamingResult != nil
}

// interceptorNamed reports whether interceptors contains name.
func interceptorNamed(interceptors []*expr.InterceptorExpr, name string) bool {
	for _, interceptor := range interceptors {
		if interceptor.Name == name {
			return true
		}
	}
	return false
}
