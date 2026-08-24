// This file records the fields, pointers, tags, declarations, and unions
// needed to write generated service types.
package service

import (
	"fmt"
	"path"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// planServiceTypeLayouts records every field, pointer, struct tag, package, and
// declaration needed to write service types after Generation.Freeze chooses
// every declaration and import name.
func planServiceTypeLayouts(facts *serviceFacts, rootTypes *rootTypeSet, generation *codegen.Generation) error {
	binder := serviceGoTypeBinder(rootTypes, generation)
	plan := func(attribute *expr.AttributeExpr, owner string) (*codegen.GoTypePlan, error) {
		if attribute == nil {
			return nil, nil
		}
		return codegen.PlanGoType(attribute, codegen.GoTypePlanOptions{
			Owner: owner,
			Policy: codegen.GoLayoutPolicy{
				UseDefault: true,
				SumType:    true,
			},
			Bind: binder,
		})
	}
	userTypes := append(append([]*userTypeFacts(nil), facts.userTypes...), facts.errorTypes...)
	for _, userType := range userTypes {
		layout, err := plan(userType.userType.Attribute(), userType.declaration.PackagePath())
		if err != nil {
			return err
		}
		userType.layout = layout
	}
	for _, errorFacts := range facts.errorFacts {
		layout, err := plan(errorFacts.attribute, facts.packagePath)
		if err != nil {
			return err
		}
		errorFacts.layout = layout
	}
	for _, method := range facts.orderedMethods {
		for _, attribute := range []*methodAttributeFacts{
			method.payload,
			method.streamingPayload,
			method.result,
			method.streamingResult,
		} {
			if attribute == nil {
				continue
			}
			layout, err := plan(attribute.attribute, facts.packagePath)
			if err != nil {
				return err
			}
			attribute.layout = layout
			if userType, ok := attribute.attribute.Type.(expr.UserType); ok {
				_, attribute.normalized = generation.NormalizedMethodType(userType)
			}
			if layout.TypeDeclaration() != nil {
				userType := attribute.attribute.Type.(expr.UserType)
				definition, err := plan(userType.Attribute(), layout.Owner())
				if err != nil {
					return err
				}
				attribute.definition = definition
			}
		}
		for _, errorFacts := range method.errors {
			layout, err := plan(errorFacts.attribute, facts.packagePath)
			if err != nil {
				return err
			}
			errorFacts.layout = layout
		}
	}
	for _, interceptor := range append(
		append([]*interceptorFacts(nil), facts.serverInterceptorFacts...),
		facts.clientInterceptorFacts...,
	) {
		if len(interceptor.methods) == 0 {
			continue
		}
		method := interceptor.methods[0]
		accesses := []struct {
			selection *expr.AttributeExpr
			parent    *methodAttributeFacts
			target    *[]*interceptorAccessFacts
		}{
			{interceptor.readPayload, method.payload, &interceptor.readPayloadFields},
			{interceptor.writePayload, method.payload, &interceptor.writePayloadFields},
			{interceptor.readResult, method.result, &interceptor.readResultFields},
			{interceptor.writeResult, method.result, &interceptor.writeResultFields},
			{interceptor.readStreamingPayload, method.streamingPayload, &interceptor.readStreamingPayloadFields},
			{interceptor.writeStreamingPayload, method.streamingPayload, &interceptor.writeStreamingPayloadFields},
			{interceptor.readStreamingResult, method.streamingResult, &interceptor.readStreamingResultFields},
			{interceptor.writeStreamingResult, method.streamingResult, &interceptor.writeStreamingResultFields},
		}
		for _, access := range accesses {
			planned, err := planInterceptorAccess(access.selection, access.parent, facts.packagePath, binder)
			if err != nil {
				return err
			}
			*access.target = planned
		}
	}
	for _, union := range facts.unions {
		if err := planUnionRenderFacts(union, binder, generation.Package(union.declaration.PackagePath())); err != nil {
			return err
		}
	}
	return nil
}

// planUnionRenderFacts records every Goa OneOf branch, nil rule, and Go type
// before Generation.Freeze chooses declaration and import names.
func planUnionRenderFacts(facts *unionFacts, binder codegen.GoTypeBinder, generatedPackage *codegen.GeneratedPackage) error {
	facts.identity = codegen.NewUnionDeclarationID(facts.attribute)
	facts.typeKey = facts.union.GetTypeKey()
	facts.valueKey = facts.union.GetValueKey()
	facts.branches = make([]*unionBranchFacts, len(facts.union.Values))
	storageNames := codegen.NewNameScope()
	// Reserve kind for the selector so a branch named kind uses another field.
	storageNames.Unique("kind")
	for index, branch := range facts.union.Values {
		declaration, err := generatedPackage.UnionBranch(facts.attribute, branch.Name)
		if err != nil {
			return err
		}
		layout, err := codegen.PlanGoType(branch.Attribute, codegen.GoTypePlanOptions{
			Owner: generatedPackage.ImportPath(),
			Policy: codegen.GoLayoutPolicy{
				UseDefault: true,
				SumType:    true,
			},
			Bind: binder,
		})
		if err != nil {
			return err
		}
		primitiveAliasType, hasPrimitiveAlias := primitiveAliasGoType(branch.Attribute.Type)
		_, isUserType := branch.Attribute.Type.(expr.UserType)
		_, hasCustomImport := layout.Import()
		facts.branches[index] = &unionBranchFacts{
			name:               branch.Name,
			fieldName:          codegen.Goify(branch.Name, true),
			storageName:        storageNames.Unique(codegen.Goify(branch.Name, false)),
			declaration:        declaration,
			layout:             layout,
			nilable:            codegen.IsNilable(branch.Attribute.Type),
			emitPrimitiveAlias: hasPrimitiveAlias && !isUserType && !hasCustomImport,
			primitiveAliasType: primitiveAliasType,
		}
	}
	return nil
}

// planInterceptorAccess records the field names, pointer choices, and Go types
// exposed to an interceptor while the design expressions are available.
func planInterceptorAccess(selection *expr.AttributeExpr, parent *methodAttributeFacts, owner string, binder codegen.GoTypeBinder) ([]*interceptorAccessFacts, error) {
	if selection == nil {
		return nil, nil
	}
	object := expr.AsObject(selection.Type)
	if object == nil {
		return nil, fmt.Errorf("plan interceptor access: selection must be an object")
	}
	if len(*object) == 0 {
		return nil, nil
	}
	result := make([]*interceptorAccessFacts, len(*object))
	for index, field := range *object {
		attribute := parent.attribute.Find(field.Name)
		if attribute == nil {
			return nil, fmt.Errorf("plan interceptor access: attribute %q is not present in its method value", field.Name)
		}
		layout, err := codegen.PlanGoType(attribute, codegen.GoTypePlanOptions{
			Owner: owner,
			Policy: codegen.GoLayoutPolicy{
				UseDefault: true,
				SumType:    true,
			},
			Bind: binder,
		})
		if err != nil {
			return nil, err
		}
		result[index] = &interceptorAccessFacts{
			attribute: expr.DupAtt(attribute),
			name:      codegen.Goify(field.Name, true),
			pointer:   parent.attribute.IsPrimitivePointer(field.Name, true),
			layout:    layout,
		}
	}
	return result, nil
}

// serviceGoTypeBinder maps authored service types and compiler-created copies
// to the generated Go declarations selected during collection.
func serviceGoTypeBinder(rootTypes *rootTypeSet, generation *codegen.Generation) codegen.GoTypeBinder {
	return func(request codegen.GoTypeBindingRequest) (codegen.GoTypeBinding, error) {
		owner := request.InheritedOwner
		if location := codegen.UserTypeLocation(request.Attribute.Type); location != nil {
			owner = path.Join(generation.GenPkg(), location.RelImportPath)
		}
		generatedPackage := generation.Package(owner)
		switch request.Kind {
		case codegen.GoNamed:
			userType := request.Attribute.Type.(expr.UserType)
			declaration, err := generatedPackage.Type(rootTypes.canonical(userType))
			if err != nil {
				return codegen.GoTypeBinding{}, err
			}
			return codegen.GoTypeBinding{Owner: owner, Type: declaration}, nil
		case codegen.GoUnion:
			declaration, err := generatedPackage.Union(request.Attribute)
			if err != nil {
				return codegen.GoTypeBinding{}, err
			}
			return codegen.GoTypeBinding{Owner: owner, Union: declaration}, nil
		default:
			return codegen.GoTypeBinding{}, fmt.Errorf("bind unsupported retained Go type kind %s", request.Kind)
		}
	}
}

// collectServiceUnionFacts selects each service Goa OneOf type once in every
// package that writes it and records its generated declaration.
func collectServiceUnionFacts(facts *serviceFacts, rootTypes *rootTypeSet, generation *codegen.Generation) error {
	seenTypes := make(map[plannedUserType]struct{})
	seenUnions := make(map[unionDataKey]struct{})
	collect := func(attribute *expr.AttributeExpr, location *codegen.Location) error {
		return collectUnionFacts(attribute, facts.packagePath, location, rootTypes, generation, seenTypes, seenUnions, &facts.unions)
	}
	for _, userType := range facts.userTypes {
		if err := collect(&expr.AttributeExpr{Type: userType.userType}, userType.location); err != nil {
			return err
		}
	}
	for _, errorType := range facts.errorTypes {
		if err := collect(&expr.AttributeExpr{Type: errorType.userType}, errorType.location); err != nil {
			return err
		}
	}
	for _, method := range facts.methods {
		attributes := []*expr.AttributeExpr{method.Payload, method.StreamingPayload, method.Result}
		if method.HasMixedResults() {
			attributes = append(attributes, method.StreamingResult)
		}
		for _, attribute := range attributes {
			var location *codegen.Location
			if attribute != nil {
				location = codegen.UserTypeLocation(attribute.Type)
			}
			if err := collect(attribute, location); err != nil {
				return err
			}
		}
		for _, methodError := range method.Errors {
			if err := collect(methodError.AttributeExpr, codegen.UserTypeLocation(methodError.Type)); err != nil {
				return err
			}
		}
	}
	return nil
}

// collectUnionFacts recursively records union declarations while keeping
// unlocated nested types in the package inherited from their enclosing type.
func collectUnionFacts(attribute *expr.AttributeExpr, servicePath string, location *codegen.Location, rootTypes *rootTypeSet, generation *codegen.Generation, seenTypes map[plannedUserType]struct{}, seenUnions map[unionDataKey]struct{}, unions *[]*unionFacts) error {
	if attribute == nil || attribute.Type == expr.Empty {
		return nil
	}
	recurse := func(attribute *expr.AttributeExpr, location *codegen.Location) error {
		return collectUnionFacts(attribute, servicePath, location, rootTypes, generation, seenTypes, seenUnions, unions)
	}
	switch actual := attribute.Type.(type) {
	case expr.UserType:
		typeLocation := codegen.UserTypeLocation(actual)
		if typeLocation == nil {
			typeLocation = location
		}
		owner := generation.Package(generatedPackagePath(generation.GenPkg(), servicePath, typeLocation))
		key := plannedUserType{userType: rootTypes.canonical(actual), owner: owner}
		if _, exists := seenTypes[key]; exists {
			return nil
		}
		seenTypes[key] = struct{}{}
		return recurse(actual.Attribute(), typeLocation)
	case *expr.Object:
		for _, field := range sortedNamedAttributes(*actual) {
			if err := recurse(field.Attribute, location); err != nil {
				return err
			}
		}
	case *expr.Array:
		return recurse(actual.ElemType, location)
	case *expr.Map:
		if err := recurse(actual.KeyType, location); err != nil {
			return err
		}
		return recurse(actual.ElemType, location)
	case *expr.Union:
		packagePath := generatedPackagePath(generation.GenPkg(), servicePath, location)
		identity := codegen.NewUnionDeclarationID(attribute)
		key := unionDataKey{packagePath: packagePath, identity: identity}
		if _, exists := seenUnions[key]; !exists {
			declaration, err := generation.Package(packagePath).Union(attribute)
			if err != nil {
				return err
			}
			seenUnions[key] = struct{}{}
			*unions = append(*unions, &unionFacts{
				attribute:   attribute,
				union:       actual,
				identity:    identity,
				typeKey:     actual.GetTypeKey(),
				valueKey:    actual.GetValueKey(),
				location:    location,
				declaration: declaration,
			})
		}
		for _, branch := range actual.Values {
			if err := recurse(branch.Attribute, location); err != nil {
				return err
			}
		}
	}
	return nil
}

// typeMapMatchesFacts reports whether a user-supplied Go type mapping applies
// to a payload, result, error, stream value, or child type selected for this
// service.
func typeMapMatchesFacts(typeMap *expr.TypeMap, facts *serviceFacts) bool {
	_, reachable := facts.reachableTypes[typeMap.User.Origin()]
	return reachable
}

// collectServiceTypeFacts selects the exact named types emitted for one
// service. Linking later formats these records without searching for the types
// again or deciding which generated package contains them.
func collectServiceTypeFacts(facts *serviceFacts, rootTypes []expr.UserType, canonical *rootTypeSet, generation *codegen.Generation) error {
	seen := make(map[userTypeDataKey]struct{})
	for _, serviceError := range facts.errors {
		selected, err := collectUserTypeFacts(serviceError.AttributeExpr, facts.packagePath, nil, canonical, generation, seen)
		if err != nil {
			return err
		}
		facts.errorTypes = append(facts.errorTypes, selected...)
	}
	for _, method := range facts.methods {
		attributes := []*expr.AttributeExpr{method.Payload, method.StreamingPayload, method.Result}
		if method.HasMixedResults() {
			attributes = append(attributes, method.StreamingResult)
		}
		for _, attribute := range attributes {
			if attribute == nil {
				continue
			}
			location := (*codegen.Location)(nil)
			inner := attribute
			if userType, ok := attribute.Type.(expr.UserType); ok {
				location = codegen.UserTypeLocation(userType)
				if _, normalized := generation.NormalizedMethodType(userType); normalized || location == nil {
					inner = userType.Attribute()
				}
			}
			selected, err := collectUserTypeFacts(inner, facts.packagePath, location, canonical, generation, seen)
			if err != nil {
				return err
			}
			facts.userTypes = append(facts.userTypes, selected...)
		}
		for _, methodError := range method.Errors {
			selected, err := collectUserTypeFacts(methodError.AttributeExpr, facts.packagePath, nil, canonical, generation, seen)
			if err != nil {
				return err
			}
			facts.errorTypes = append(facts.errorTypes, selected...)
		}
	}
	for _, method := range facts.methods {
		attributes := []*expr.AttributeExpr{method.Payload, method.StreamingPayload, method.Result}
		if method.HasMixedResults() {
			attributes = append(attributes, method.StreamingResult)
		}
		for _, attribute := range attributes {
			if attribute == nil || attribute.Type == expr.Empty {
				continue
			}
			if _, raw := attribute.Type.(*expr.Object); raw {
				panic(fmt.Sprintf(
					"service %q method %q declares a raw object type: codegen.NewGeneration must own the finalized design before generators read it",
					facts.service.Name, method.Name))
			}
			if userType, ok := attribute.Type.(expr.UserType); ok {
				declaration, err := generation.Package(generatedPackagePath(
					generation.GenPkg(), facts.packagePath, codegen.UserTypeLocation(userType),
				)).Type(userType)
				if err != nil {
					return err
				}
				seen[userTypeDataKey{origin: userType.Origin(), declaration: declaration}] = struct{}{}
			}
		}
	}
	for _, userType := range rootTypes {
		services, forced := userType.Attribute().Meta["type:generate:force"]
		if !forced || len(services) > 0 && !containsString(services, facts.service.Name) {
			continue
		}
		selected, err := collectUserTypeFacts(
			&expr.AttributeExpr{Type: userType}, facts.packagePath, nil, canonical, generation, seen,
		)
		if err != nil {
			return err
		}
		facts.userTypes = append(facts.userTypes, selected...)
	}
	for _, userType := range facts.userTypes {
		facts.reachableTypes[userType.userType.Origin()] = struct{}{}
	}
	return nil
}

// collectUserTypeFacts recursively selects named types while carrying the
// package location inherited from an enclosing generated type.
func collectUserTypeFacts(attribute *expr.AttributeExpr, servicePath string, location *codegen.Location, canonical *rootTypeSet, generation *codegen.Generation, seen map[userTypeDataKey]struct{}) ([]*userTypeFacts, error) {
	if attribute == nil || attribute.Type == expr.Empty {
		return nil, nil
	}
	collect := func(attribute *expr.AttributeExpr, location *codegen.Location) ([]*userTypeFacts, error) {
		return collectUserTypeFacts(attribute, servicePath, location, canonical, generation, seen)
	}
	var result []*userTypeFacts
	switch actual := attribute.Type.(type) {
	case expr.UserType:
		typeLocation := codegen.UserTypeLocation(actual)
		if typeLocation == nil {
			typeLocation = location
		}
		declaration, err := generation.Package(
			generatedPackagePath(generation.GenPkg(), servicePath, typeLocation),
		).Type(canonical.canonical(actual))
		if err != nil {
			return nil, err
		}
		key := userTypeDataKey{origin: actual.Origin(), declaration: declaration}
		if _, exists := seen[key]; exists {
			return nil, nil
		}
		seen[key] = struct{}{}
		result = append(result, &userTypeFacts{
			userType:     actual,
			name:         actual.Name(),
			description:  actual.Attribute().Description,
			errorName:    retainedErrorName(actual),
			serviceError: expr.IsErrorResult(actual),
			location:     typeLocation,
			declaration:  declaration,
		})
		nested, err := collect(actual.Attribute(), typeLocation)
		return append(result, nested...), err
	case *expr.Object:
		for _, field := range *actual {
			selected, err := collect(field.Attribute, location)
			if err != nil {
				return nil, err
			}
			result = append(result, selected...)
		}
	case *expr.Array:
		return collect(actual.ElemType, location)
	case *expr.Map:
		key, err := collect(actual.KeyType, location)
		if err != nil {
			return nil, err
		}
		value, err := collect(actual.ElemType, location)
		return append(key, value...), err
	case *expr.Union:
		for _, branch := range actual.Values {
			if userType, generated := generatedUnionBranch(branch, canonical); generated && location != nil {
				selected, err := collect(&expr.AttributeExpr{Type: userType}, location)
				if err != nil {
					return nil, err
				}
				result = append(result, selected...)
				continue
			}
			selected, err := collect(branch.Attribute, location)
			if err != nil {
				return nil, err
			}
			result = append(result, selected...)
		}
	}
	return result, nil
}

// This helper copies the Go expression returned by GoaErrorName while the
// design error metadata is still available.
func retainedErrorName(userType expr.UserType) string {
	if object := expr.AsObject(userType); object != nil {
		for _, field := range *object {
			if _, ok := field.Attribute.Meta["struct:error:name"]; ok {
				return fmt.Sprintf("e.%s", codegen.GoifyAtt(field.Attribute, field.Name, true))
			}
		}
	}
	if value, ok := userType.Attribute().Meta["struct:error:name"]; ok {
		return fmt.Sprintf("%q", value[0])
	}
	return fmt.Sprintf("%q", userType.Name())
}

// containsString reports whether values contains target without introducing a
// second service-selection representation.
func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
