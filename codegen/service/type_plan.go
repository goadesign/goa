// This file collects retained service user-type and union emission facts before generated package names freeze.
package service

import (
	"fmt"
	"path"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// planServiceTypeLayouts retains every field, pointer, tag, owner, and exact
// declaration used to spell core service types after names freeze.
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
		reference, err := plan(&expr.AttributeExpr{Type: userType.userType}, facts.packagePath)
		if err != nil {
			return err
		}
		userType.reference = reference
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
			{interceptor.readStreamingResult, method.result, &interceptor.readStreamingResultFields},
			{interceptor.writeStreamingResult, method.result, &interceptor.writeStreamingResultFields},
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

// planUnionRenderFacts retains every semantic branch decision and exact type
// layout before generated names and import aliases freeze.
func planUnionRenderFacts(facts *unionFacts, binder codegen.GoTypeBinder, generatedPackage *codegen.GeneratedPackage) error {
	facts.identity = codegen.NewUnionTypeID(facts.union)
	facts.typeKey = facts.union.GetTypeKey()
	facts.valueKey = facts.union.GetValueKey()
	facts.branches = make([]*unionBranchFacts, len(facts.union.Values))
	for index, branch := range facts.union.Values {
		declaration, err := generatedPackage.UnionBranch(facts.union, branch.Name)
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
			declaration:        declaration,
			layout:             layout,
			nilable:            codegen.IsNilable(branch.Attribute.Type),
			emitPrimitiveAlias: hasPrimitiveAlias && !isUserType && !hasCustomImport,
			primitiveAliasType: primitiveAliasType,
		}
	}
	return nil
}

// planInterceptorAccess retains the selected generated field names, pointer
// behavior, and exact type layouts while the design expressions are inputs.
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
			name:    codegen.Goify(field.Name, true),
			pointer: parent.attribute.IsPrimitivePointer(field.Name, true),
			layout:  layout,
		}
	}
	return result, nil
}

// serviceGoTypeBinder binds authored and normalized service occurrences to
// the package declarations selected during collection.
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
			union := request.Attribute.Type.(*expr.Union)
			declaration, err := generatedPackage.Union(union)
			if err != nil {
				return codegen.GoTypeBinding{}, err
			}
			return codegen.GoTypeBinding{Owner: owner, Union: declaration}, nil
		default:
			return codegen.GoTypeBinding{}, fmt.Errorf("bind unsupported retained Go type kind %s", request.Kind)
		}
	}
}

// collectServiceUnionFacts selects every service sum type once per generated
// package and retains the declaration allocated during planning.
func collectServiceUnionFacts(facts *serviceFacts, rootTypes *rootTypeSet, generation *codegen.Generation) error {
	seenTypes := make(map[plannedUserType]struct{})
	seenUnions := make(map[unionDataKey]struct{})
	collect := func(attribute *expr.AttributeExpr, location *codegen.Location) error {
		return collectUnionFacts(attribute, facts.service, location, rootTypes, generation, seenTypes, seenUnions, &facts.unions)
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
func collectUnionFacts(attribute *expr.AttributeExpr, service *expr.ServiceExpr, location *codegen.Location, rootTypes *rootTypeSet, generation *codegen.Generation, seenTypes map[plannedUserType]struct{}, seenUnions map[unionDataKey]struct{}, unions *[]*unionFacts) error {
	if attribute == nil || attribute.Type == expr.Empty {
		return nil
	}
	recurse := func(attribute *expr.AttributeExpr, location *codegen.Location) error {
		return collectUnionFacts(attribute, service, location, rootTypes, generation, seenTypes, seenUnions, unions)
	}
	switch actual := attribute.Type.(type) {
	case expr.UserType:
		typeLocation := codegen.UserTypeLocation(actual)
		if typeLocation == nil {
			typeLocation = location
		}
		owner := generation.Package(generatedPackagePath(generation.GenPkg(), service, typeLocation))
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
		packagePath := generatedPackagePath(generation.GenPkg(), service, location)
		key := unionDataKey{packagePath: packagePath, identity: codegen.NewUnionTypeID(actual)}
		if _, exists := seenUnions[key]; !exists {
			declaration, err := generation.Package(packagePath).Union(actual)
			if err != nil {
				return err
			}
			seenUnions[key] = struct{}{}
			*unions = append(*unions, &unionFacts{
				union:       actual,
				identity:    codegen.NewUnionTypeID(actual),
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

// typeMapMatchesFacts reports whether a mapping's user type belongs to the
// retained method or nested service types selected during collection.
func typeMapMatchesFacts(typeMap *expr.TypeMap, facts *serviceFacts) bool {
	_, reachable := facts.reachableTypes[typeMap.User.Origin()]
	return reachable
}

// collectServiceTypeFacts selects the exact named types emitted for one
// service. Linking later formats these records without repeating reachability
// or package-ownership decisions.
func collectServiceTypeFacts(facts *serviceFacts, rootTypes []expr.UserType, canonical *rootTypeSet, generation *codegen.Generation) error {
	seen := make(map[userTypeDataKey]struct{})
	for _, serviceError := range facts.errors {
		selected, err := collectUserTypeFacts(serviceError.AttributeExpr, facts.service, nil, canonical, generation, seen)
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
			selected, err := collectUserTypeFacts(inner, facts.service, location, canonical, generation, seen)
			if err != nil {
				return err
			}
			facts.userTypes = append(facts.userTypes, selected...)
		}
		for _, methodError := range method.Errors {
			selected, err := collectUserTypeFacts(methodError.AttributeExpr, facts.service, nil, canonical, generation, seen)
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
					generation.GenPkg(), facts.service, codegen.UserTypeLocation(userType),
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
			&expr.AttributeExpr{Type: userType}, facts.service, nil, canonical, generation, seen,
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
func collectUserTypeFacts(attribute *expr.AttributeExpr, service *expr.ServiceExpr, location *codegen.Location, canonical *rootTypeSet, generation *codegen.Generation, seen map[userTypeDataKey]struct{}) ([]*userTypeFacts, error) {
	if attribute == nil || attribute.Type == expr.Empty {
		return nil, nil
	}
	collect := func(attribute *expr.AttributeExpr, location *codegen.Location) ([]*userTypeFacts, error) {
		return collectUserTypeFacts(attribute, service, location, canonical, generation, seen)
	}
	var result []*userTypeFacts
	switch actual := attribute.Type.(type) {
	case expr.UserType:
		typeLocation := codegen.UserTypeLocation(actual)
		if typeLocation == nil {
			typeLocation = location
		}
		declaration, err := generation.Package(
			generatedPackagePath(generation.GenPkg(), service, typeLocation),
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
			serviceError: actual == expr.ErrorResult,
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

// retainedErrorName copies the exact Go expression returned by GoaErrorName
// before error metadata can be changed by a later generator phase.
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
