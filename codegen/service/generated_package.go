// This file assigns service types and unions to the generated packages that
// write them. Each generated package writes each declaration once.
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

type (
	// generatedTypeEmissionKind identifies the template that writes one type
	// outside its service package.
	generatedTypeEmissionKind uint8

	// generatedTypeEmissionFacts records one type declaration written outside its
	// service package and the service data used to render it.
	generatedTypeEmissionFacts struct {
		kind        generatedTypeEmissionKind
		declaration *codegen.TypeDeclaration
		location    *codegen.Location
		root        *rootFacts
		service     *serviceFacts
		method      *methodFacts
		attribute   *methodAttributeFacts
		userType    *userTypeFacts
		uses        []generatedTypeMethodUse
		error       bool
	}

	// generatedTypeMethodUse records how one service method directly uses an
	// authored type written outside the service package.
	generatedTypeMethodUse struct {
		service string
		method  string
		roles   generatedTypeMethodRoles
	}

	// generatedTypeMethodRoles records the method fields whose declared type is
	// the authored type being written.
	generatedTypeMethodRoles uint8

	// generatedUnionEmissionFacts records one Goa OneOf declaration and the data
	// used to write it outside its service package.
	generatedUnionEmissionFacts struct {
		root    *rootFacts
		service *serviceFacts
		union   *unionFacts
	}

	// plannedAttribute records one service field and the Go package used by child
	// types that do not declare their own struct:pkg:path location.
	plannedAttribute struct {
		attribute *expr.AttributeExpr
		service   *serviceFacts
		location  *codegen.Location
	}

	// plannedUserType identifies one user type written to one generated package.
	// Extend may copy the same expression into more than one package.
	plannedUserType struct {
		userType expr.UserType
		owner    *codegen.GeneratedPackage
	}

	// rootTypeSet maps compiler-created copies back to the user type declared in
	// the same design. Generated Goa OneOf branch aliases are not included.
	rootTypeSet struct {
		byOrigin map[expr.UserType]expr.UserType
	}

	// generatedPackageData stores the render data emitted into one Go package.
	generatedPackageData struct {
		types        map[*codegen.TypeDeclaration]*generatedTypeData
		unions       map[*codegen.UnionDeclaration]*UnionTypeData
		unionImports []*codegen.ImportSpec
	}

	// generatedTypeData stores one user-type declaration placed in the file
	// selected by its metadata, plus optional error behavior.
	generatedTypeData struct {
		declaration *codegen.TypeDeclaration
		location    *codegen.Location
		imports     []*codegen.ImportSpec
		section     *codegen.SectionTemplate
		error       *codegen.SectionTemplate
	}
)

const (
	generatedPayloadEmission generatedTypeEmissionKind = iota + 1
	generatedStreamingPayloadEmission
	generatedResultEmission
	generatedStreamingResultEmission
	generatedUserTypeEmission
	generatedErrorTypeEmission
)

const (
	generatedPayloadRole generatedTypeMethodRoles = 1 << iota
	generatedStreamingPayloadRole
	generatedResultRole
	generatedStreamingResultRole
)

// collectServiceDeclarations submits every user type and Goa OneOf declaration
// reachable from root. It submits authored type names first, so generated union
// names receive a number when both request the same Go name.
func collectServiceDeclarations(facts *rootFacts, generation *codegen.Generation) error {
	if !generation.HasRoot(facts.root) {
		return rootMembershipError(facts.root)
	}
	inputs := planningInputs(facts)
	rootTypes := facts.rootTypes
	for _, serviceFacts := range facts.services {
		// Record the service package now. File building reads its Go names after all
		// generators finish submitting declarations.
		if _, err := generation.ClaimPackage(serviceFacts.packagePath); err != nil {
			return err
		}
	}
	methodTypes, err := planMethodTypes(facts, generation)
	if err != nil {
		return err
	}

	seenTypes := make(map[plannedUserType]struct{})
	for _, input := range inputs {
		if err := planUserTypes(input.attribute, input.service, input.location, generation, rootTypes, methodTypes, seenTypes); err != nil {
			return err
		}
	}

	seenTypes = make(map[plannedUserType]struct{})
	for _, input := range inputs {
		if err := planUnions(input.attribute, input.service, input.location, generation, rootTypes, seenTypes); err != nil {
			return err
		}
	}
	if err := planViews(facts, generation); err != nil {
		return err
	}
	return nil
}

// collectGeneratedPackageEmissions selects the one service that supplies the
// definition for each type and Goa OneOf declaration shared across the designs
// in this generation command.
func collectGeneratedPackageEmissions(roots []*rootFacts) error {
	types := make(map[*codegen.TypeDeclaration]*generatedTypeEmissionFacts)
	unions := make(map[*codegen.UnionDeclaration]*generatedUnionEmissionFacts)
	for _, root := range roots {
		root.generatedTypes = nil
		root.generatedUnions = nil
		for _, service := range root.services {
			for _, union := range service.unions {
				emission := &generatedUnionEmissionFacts{root: root, service: service, union: union}
				if existing := unions[union.declaration]; existing != nil {
					if err := validateGeneratedUnionEmission(existing, emission); err != nil {
						return err
					}
					if generatedUnionEmissionLess(emission, existing) {
						unions[union.declaration] = emission
					}
				} else {
					unions[union.declaration] = emission
				}
			}
			for _, method := range service.orderedMethods {
				candidates := []struct {
					kind      generatedTypeEmissionKind
					attribute *methodAttributeFacts
				}{
					{generatedPayloadEmission, method.payload},
					{generatedStreamingPayloadEmission, method.streamingPayload},
					{generatedResultEmission, method.result},
				}
				if method.hasMixedResults {
					candidates = append(candidates, struct {
						kind      generatedTypeEmissionKind
						attribute *methodAttributeFacts
					}{generatedStreamingResultEmission, method.streamingResult})
				}
				for _, candidate := range candidates {
					attribute := candidate.attribute
					if attribute == nil || attribute.location == nil || !attribute.normalized || attribute.definition == nil {
						continue
					}
					emission := &generatedTypeEmissionFacts{
						kind:        candidate.kind,
						declaration: attribute.layout.TypeDeclaration(),
						location:    copyGeneratedLocation(attribute.location),
						root:        root,
						service:     service,
						method:      method,
						attribute:   attribute,
					}
					if err := selectGeneratedTypeEmission(types, emission); err != nil {
						return err
					}
				}
			}
			for _, userType := range service.userTypes {
				if userType.location == nil {
					continue
				}
				emission := &generatedTypeEmissionFacts{
					kind:        generatedUserTypeEmission,
					declaration: userType.declaration,
					location:    copyGeneratedLocation(userType.location),
					root:        root,
					service:     service,
					userType:    userType,
					uses:        generatedTypeMethodUses(service, userType.declaration),
				}
				if err := selectGeneratedTypeEmission(types, emission); err != nil {
					return err
				}
			}
			for _, errorType := range service.errorTypes {
				if errorType.location == nil || errorType.serviceError {
					continue
				}
				emission := &generatedTypeEmissionFacts{
					kind:        generatedErrorTypeEmission,
					declaration: errorType.declaration,
					location:    copyGeneratedLocation(errorType.location),
					root:        root,
					service:     service,
					userType:    errorType,
					error:       true,
				}
				if err := selectGeneratedTypeEmission(types, emission); err != nil {
					return err
				}
			}
		}
	}
	for _, emission := range types {
		emission.root.generatedTypes = append(emission.root.generatedTypes, emission)
	}
	for _, emission := range unions {
		emission.root.generatedUnions = append(emission.root.generatedUnions, emission)
	}
	for _, root := range roots {
		sort.Slice(root.generatedTypes, func(i, j int) bool {
			return generatedTypeEmissionLess(root.generatedTypes[i], root.generatedTypes[j])
		})
		sort.Slice(root.generatedUnions, func(i, j int) bool {
			return generatedUnionEmissionLess(root.generatedUnions[i], root.generatedUnions[j])
		})
	}
	return nil
}

// selectGeneratedTypeEmission keeps one copy when two services would write the
// same type declaration and rejects them when their definitions differ.
func selectGeneratedTypeEmission(selected map[*codegen.TypeDeclaration]*generatedTypeEmissionFacts, candidate *generatedTypeEmissionFacts) error {
	existing := selected[candidate.declaration]
	if existing == nil {
		selected[candidate.declaration] = candidate
		return nil
	}
	if err := validateGeneratedTypeEmission(existing, candidate); err != nil {
		return err
	}
	uses := mergeGeneratedTypeMethodUses(existing.uses, candidate.uses)
	if generatedTypeEmissionLess(candidate, existing) {
		candidate.uses = uses
		selected[candidate.declaration] = candidate
	} else {
		existing.uses = uses
	}
	return nil
}

// generatedTypeMethodUses records each method that declares the authored type
// as its payload, result, or streamed value. Types used only inside another
// type do not match these declarations.
func generatedTypeMethodUses(service *serviceFacts, declaration *codegen.TypeDeclaration) []generatedTypeMethodUse {
	uses := make([]generatedTypeMethodUse, 0, len(service.orderedMethods))
	for _, method := range service.orderedMethods {
		var roles generatedTypeMethodRoles
		roles = addGeneratedTypeMethodRole(roles, method.payload, declaration, generatedPayloadRole)
		roles = addGeneratedTypeMethodRole(roles, method.streamingPayload, declaration, generatedStreamingPayloadRole)
		roles = addGeneratedTypeMethodRole(roles, method.result, declaration, generatedResultRole)
		roles = addGeneratedTypeMethodRole(roles, method.streamingResult, declaration, generatedStreamingResultRole)
		if roles == 0 {
			continue
		}
		uses = append(uses, generatedTypeMethodUse{
			service: service.name,
			method:  method.name,
			roles:   roles,
		})
	}
	return mergeGeneratedTypeMethodUses(nil, uses)
}

// addGeneratedTypeMethodRole records role when the method field directly uses
// declaration.
func addGeneratedTypeMethodRole(roles generatedTypeMethodRoles, attribute *methodAttributeFacts, declaration *codegen.TypeDeclaration, role generatedTypeMethodRoles) generatedTypeMethodRoles {
	if attribute != nil && attribute.layout != nil && attribute.layout.TypeDeclaration() == declaration {
		return roles | role
	}
	return roles
}

// mergeGeneratedTypeMethodUses combines the methods found through different
// services or design roots and returns them in a stable order.
func mergeGeneratedTypeMethodUses(left, right []generatedTypeMethodUse) []generatedTypeMethodUse {
	uses := append(append(make([]generatedTypeMethodUse, 0, len(left)+len(right)), left...), right...)
	sort.Slice(uses, func(i, j int) bool {
		if uses[i].service != uses[j].service {
			return uses[i].service < uses[j].service
		}
		return uses[i].method < uses[j].method
	})
	merged := uses[:0]
	for _, use := range uses {
		last := len(merged) - 1
		if last >= 0 && merged[last].service == use.service && merged[last].method == use.method {
			merged[last].roles |= use.roles
			continue
		}
		merged = append(merged, use)
	}
	return merged
}

// validateGeneratedTypeEmission returns an error when two services would write
// different definitions for the same generated Go type declaration.
func validateGeneratedTypeEmission(left, right *generatedTypeEmissionFacts) error {
	if left.kind != right.kind || !sameGeneratedLocation(left.location, right.location) ||
		!sameGeneratedTypeEmissionSource(left, right) ||
		!sameGeneratedTypeEmissionContent(left, right) ||
		!generatedTypeEmissionLayout(left).Equivalent(generatedTypeEmissionLayout(right)) ||
		!slices.Equal(
			left.service.generatedTypeImports[left.declaration].paths,
			right.service.generatedTypeImports[right.declaration].paths,
		) {
		return fmt.Errorf(
			"conflicting generated type emission in package %q: roles %d and %d, sources %q and %q",
			left.declaration.PackagePath(),
			left.kind,
			right.kind,
			generatedTypeEmissionName(left),
			generatedTypeEmissionName(right),
		)
	}
	return nil
}

// generatedTypeEmissionLayout returns the Go type definition supplied by one
// service. References from other service files do not affect which definition
// is written.
func generatedTypeEmissionLayout(emission *generatedTypeEmissionFacts) *codegen.GoTypePlan {
	if emission.userType != nil {
		return emission.userType.layout
	}
	return emission.attribute.definition
}

// sameGeneratedTypeEmissionContent reports whether two services would write
// the same comment and error behavior for one type declaration.
func sameGeneratedTypeEmissionContent(left, right *generatedTypeEmissionFacts) bool {
	if left.error != right.error {
		return false
	}
	if left.userType != nil || right.userType != nil {
		return left.userType != nil && right.userType != nil &&
			left.userType.name == right.userType.name &&
			left.userType.description == right.userType.description &&
			left.userType.errorName == right.userType.errorName &&
			left.userType.serviceError == right.userType.serviceError
	}
	return left.service.packagePath == right.service.packagePath &&
		left.method.name == right.method.name &&
		left.attribute.description == right.attribute.description
}

// sameGeneratedTypeEmissionSource reports whether two candidates came from the
// same authored type. Generated aliases for the same Goa OneOf branch also
// match because the package already owns one declaration for that branch.
func sameGeneratedTypeEmissionSource(left, right *generatedTypeEmissionFacts) bool {
	if generatedTypeEmissionOrigin(left) == generatedTypeEmissionOrigin(right) {
		return true
	}
	return left.userType != nil && right.userType != nil &&
		!left.root.rootTypes.contains(left.userType.userType) &&
		!right.root.rootTypes.contains(right.userType.userType)
}

// generatedTypeEmissionName describes the design type that caused a conflict
// while selecting one generated definition.
func generatedTypeEmissionName(emission *generatedTypeEmissionFacts) string {
	origin := generatedTypeEmissionOrigin(emission)
	if origin == nil {
		return ""
	}
	return origin.Name()
}

// validateGeneratedUnionEmission returns an error when two services would
// write the same Goa OneOf declaration in different files or with different
// fields.
func validateGeneratedUnionEmission(left, right *generatedUnionEmissionFacts) error {
	if left.union.declaration != right.union.declaration ||
		left.union.identity != right.union.identity ||
		left.union.typeKey != right.union.typeKey ||
		left.union.valueKey != right.union.valueKey ||
		generatedLocationPath(left.union.location) != generatedLocationPath(right.union.location) ||
		!sameGeneratedUnionBranches(left.union.branches, right.union.branches) ||
		!slices.Equal(left.union.imports.paths, right.union.imports.paths) {
		return fmt.Errorf(
			"conflicting generated union emission in package %q: declarations equal=%t, keys %q/%q and %q/%q",
			left.union.declaration.PackagePath(),
			left.union.declaration == right.union.declaration,
			left.union.typeKey,
			left.union.valueKey,
			right.union.typeKey,
			right.union.valueKey,
		)
	}
	return nil
}

// sameGeneratedUnionBranches reports whether two services would write the same
// branch fields, constructors, validation, and JSON functions for one Goa OneOf
// declaration.
func sameGeneratedUnionBranches(left, right []*unionBranchFacts) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		leftBranch, rightBranch := left[index], right[index]
		if leftBranch.name != rightBranch.name ||
			leftBranch.fieldName != rightBranch.fieldName ||
			leftBranch.declaration != rightBranch.declaration ||
			!leftBranch.layout.Equivalent(rightBranch.layout) ||
			leftBranch.nilable != rightBranch.nilable ||
			leftBranch.emitPrimitiveAlias != rightBranch.emitPrimitiveAlias ||
			leftBranch.primitiveAliasType != rightBranch.primitiveAliasType {
			return false
		}
	}
	return true
}

// generatedLocationPath returns the generated package selected by location.
// Union declarations always emit in unions.go, so their enclosing type's file
// name does not affect which generated package contains the union.
func generatedLocationPath(location *codegen.Location) string {
	if location == nil {
		return ""
	}
	return location.RelImportPath
}

// generatedTypeEmissionOrigin returns the exact authored or normalized source
// whose layout defines an emitted declaration.
func generatedTypeEmissionOrigin(emission *generatedTypeEmissionFacts) expr.UserType {
	if emission.userType != nil {
		return emission.userType.userType.Origin()
	}
	if userType, ok := emission.attribute.attribute.Type.(expr.UserType); ok {
		return userType.Origin()
	}
	return nil
}

// generatedTypeEmissionLess orders equal definitions by service and method
// names, so walking designs in another order does not change the selected copy.
func generatedTypeEmissionLess(left, right *generatedTypeEmissionFacts) bool {
	if left.declaration.PackagePath() != right.declaration.PackagePath() {
		return left.declaration.PackagePath() < right.declaration.PackagePath()
	}
	if left.location.FilePath != right.location.FilePath {
		return left.location.FilePath < right.location.FilePath
	}
	if left.service.packagePath != right.service.packagePath {
		return left.service.packagePath < right.service.packagePath
	}
	leftMethod, rightMethod := "", ""
	if left.method != nil {
		leftMethod = left.method.name
	}
	if right.method != nil {
		rightMethod = right.method.name
	}
	return leftMethod < rightMethod
}

// generatedUnionEmissionLess orders equal Goa OneOf definitions by package and
// service names.
func generatedUnionEmissionLess(left, right *generatedUnionEmissionFacts) bool {
	if left.union.declaration.PackagePath() != right.union.declaration.PackagePath() {
		return left.union.declaration.PackagePath() < right.union.declaration.PackagePath()
	}
	return left.service.packagePath < right.service.packagePath
}

// rootMembershipError reports an attempt to generate files from a design that
// was not supplied to this generation command.
func rootMembershipError(root *expr.RootExpr) error {
	return fmt.Errorf("service root %p does not belong to the generation", root)
}

// planMethodTypes submits the payload and result wrapper types created for raw
// object definitions. Authored user types are submitted separately and keep
// their requested names when no declaration conflicts.
func planMethodTypes(facts *rootFacts, generation *codegen.Generation) (map[expr.UserType]codegen.DerivedTypeID, error) {
	planned := make(map[expr.UserType]codegen.DerivedTypeID)
	for _, serviceFacts := range facts.services {
		generatedPackage := generation.Package(serviceFacts.packagePath)
		for _, method := range serviceFacts.methods {
			attributes := []*expr.AttributeExpr{
				method.Payload,
				method.StreamingPayload,
				method.Result,
			}
			if method.HasMixedResults() {
				attributes = append(attributes, method.StreamingResult)
			}
			for _, attribute := range attributes {
				userType, ok := attribute.Type.(expr.UserType)
				if !ok {
					continue
				}
				identity, ok := generation.NormalizedMethodType(userType)
				if !ok {
					continue
				}
				_, derived, err := generatedPackage.DeclareMethodType(identity, userType)
				if err != nil {
					return nil, err
				}
				planned[userType.Origin()] = derived
			}
		}
	}
	return planned, nil
}

// planningInputs returns the payloads, results, errors, and stream values that
// can write service types. It excludes design types that no service uses.
func planningInputs(facts *rootFacts) []plannedAttribute {
	var inputs []plannedAttribute
	for _, serviceFacts := range facts.services {
		for _, serviceError := range serviceFacts.errors {
			inputs = append(inputs, plannedAttribute{attribute: serviceError.AttributeExpr, service: serviceFacts})
		}
		for _, method := range serviceFacts.methods {
			inputs = append(inputs,
				plannedAttribute{attribute: method.Payload, service: serviceFacts},
				plannedAttribute{attribute: method.StreamingPayload, service: serviceFacts},
				plannedAttribute{attribute: method.Result, service: serviceFacts},
			)
			if method.HasMixedResults() {
				inputs = append(inputs, plannedAttribute{attribute: method.StreamingResult, service: serviceFacts})
			}
			for _, methodError := range method.Errors {
				inputs = append(inputs, plannedAttribute{attribute: methodError.AttributeExpr, service: serviceFacts})
			}
		}
		for _, userType := range facts.types {
			services, ok := userType.Attribute().Meta["type:generate:force"]
			if !ok || len(services) > 0 && !slices.Contains(services, serviceFacts.name) {
				continue
			}
			inputs = append(inputs, plannedAttribute{
				attribute: &expr.AttributeExpr{Type: userType},
				service:   serviceFacts,
			})
		}
	}
	return inputs
}

// planUserTypes walks attribute and submits each user type to the Go package
// selected by its own metadata or by its enclosing type.
func planUserTypes(attribute *expr.AttributeExpr, service *serviceFacts, location *codegen.Location, generation *codegen.Generation, rootTypes *rootTypeSet, methodTypes map[expr.UserType]codegen.DerivedTypeID, seen map[plannedUserType]struct{}) error {
	if attribute == nil || attribute.Type == expr.Empty {
		return nil
	}
	recurse := func(attribute *expr.AttributeExpr, location *codegen.Location) error {
		return planUserTypes(attribute, service, location, generation, rootTypes, methodTypes, seen)
	}
	switch actual := attribute.Type.(type) {
	case expr.UserType:
		if _, normalized := methodTypes[actual.Origin()]; normalized {
			return recurse(actual.Attribute(), location)
		}
		declaredType := rootTypes.canonical(actual)
		typeLocation := codegen.UserTypeLocation(actual)
		if typeLocation == nil {
			typeLocation = location
		}
		owner, err := claimGeneratedPackage(generation, service.packagePath, typeLocation)
		if err != nil {
			return err
		}
		key := plannedUserType{userType: declaredType, owner: owner}
		if _, ok := seen[key]; ok {
			return nil
		}
		seen[key] = struct{}{}
		if _, err := owner.DeclareUserType(declaredType); err != nil {
			return err
		}
		return recurse(actual.Attribute(), typeLocation)
	case *expr.Object:
		for _, named := range *actual {
			if err := recurse(named.Attribute, location); err != nil {
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
		for _, named := range actual.Values {
			if userType, ok := generatedUnionBranch(named, rootTypes); ok {
				if err := recurse(userType.Attribute(), location); err != nil {
					return err
				}
				continue
			}
			if err := recurse(named.Attribute, location); err != nil {
				return err
			}
		}
	}
	return nil
}

// planUnions walks attribute after authored user type names have been submitted
// and records each Goa OneOf declaration in the package that writes it.
func planUnions(attribute *expr.AttributeExpr, service *serviceFacts, location *codegen.Location, generation *codegen.Generation, rootTypes *rootTypeSet, seen map[plannedUserType]struct{}) error {
	if attribute == nil || attribute.Type == expr.Empty {
		return nil
	}
	recurse := func(attribute *expr.AttributeExpr, location *codegen.Location) error {
		return planUnions(attribute, service, location, generation, rootTypes, seen)
	}
	switch actual := attribute.Type.(type) {
	case expr.UserType:
		declaredType := rootTypes.canonical(actual)
		typeLocation := codegen.UserTypeLocation(actual)
		if typeLocation == nil {
			typeLocation = location
		}
		owner, err := claimGeneratedPackage(generation, service.packagePath, typeLocation)
		if err != nil {
			return err
		}
		key := plannedUserType{userType: declaredType, owner: owner}
		if _, ok := seen[key]; ok {
			return nil
		}
		seen[key] = struct{}{}
		return recurse(actual.Attribute(), typeLocation)
	case *expr.Object:
		for _, named := range sortedNamedAttributes(*actual) {
			if err := recurse(named.Attribute, location); err != nil {
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
		generatedPackage, err := claimGeneratedPackage(generation, service.packagePath, location)
		if err != nil {
			return err
		}
		if _, err := generatedPackage.DeclareUnion(attribute); err != nil {
			return err
		}
		for _, named := range actual.Values {
			if userType, ok := generatedUnionBranch(named, rootTypes); ok {
				if _, err := generatedPackage.DeclareUnionBranchType(attribute, named.Name, userType); err != nil {
					return err
				}
				if err := recurse(userType.Attribute(), location); err != nil {
					return err
				}
				continue
			}
			if err := recurse(named.Attribute, location); err != nil {
				return err
			}
		}
	}
	return nil
}

// planViews builds the result types for each declared view, submits their Go
// type names, and then submits Goa OneOf types written in the views package.
func planViews(facts *rootFacts, generation *codegen.Generation) error {
	for _, serviceFacts := range facts.services {
		views, err := generation.ClaimPackage(serviceFacts.viewsPath)
		if err != nil {
			return err
		}
		seenProjected := make(map[expr.UserType]expr.UserType)
		projectedFactsByOrigin := make(map[expr.UserType]*projectedTypeFacts)
		derived := make(map[expr.UserType]codegen.DerivedTypeID)
		var projectedRoots []*expr.AttributeExpr
		for _, method := range serviceFacts.methods {
			if !hasResultType(method.Result) {
				continue
			}
			projected, source := projectedResultRoot(generation, method)
			pairs := projectTypePairs(projected, source, seenProjected)
			projection := &projectionFacts{pairs: pairs}
			serviceFacts.projections[method] = projection
			serviceFacts.methodByExpr[method].projection = projection
			for _, pair := range pairs {
				identity := codegen.NewProjectedTypeID(pair.source)
				if _, err := views.DeclareDerivedType(identity, codegen.Goify(pair.projected.Name(), true)); err != nil {
					return err
				}
				derived[pair.projected.Origin()] = identity
				projectedFacts, err := collectProjectedTypeFacts(pair)
				if err != nil {
					return err
				}
				projection.types = append(projection.types, projectedFacts)
				projectedFactsByOrigin[pair.source.Origin()] = projectedFacts
			}
			if resultType, ok := method.Result.Type.(*expr.ResultTypeExpr); ok {
				projectedType := seenProjected[resultType.Origin()]
				if len(pairs) > 0 {
					projectedType = pairs[0].projected
				}
				if projectedType != nil {
					viewName := ""
					if !resultType.HasMultipleViews() {
						viewName = expr.DefaultView
					}
					if selected, ok := method.Result.Meta.Last(expr.ViewMetaKey); ok {
						viewName = selected
					}
					serviceFacts.methodByExpr[method].viewedResult = &viewedResultFacts{
						wrapped:      wrapProjected(projectedType),
						origin:       resultType.Origin(),
						source:       serviceFacts.methodByExpr[method].result,
						viewName:     viewName,
						views:        projectedFactsByOrigin[resultType.Origin()].views,
						conversions:  projectedFactsByOrigin[resultType.Origin()].conversions,
						projected:    projectedFactsByOrigin[resultType.Origin()],
						isCollection: expr.IsArray(method.Result.Type),
					}
				}
			}
			removeMeta(projected)
			projectedRoots = append(projectedRoots, projected)

			if resultType, ok := method.Result.Type.(*expr.ResultTypeExpr); ok {
				if _, err := views.DeclareDerivedType(
					codegen.NewViewedResultTypeID(resultType),
					codegen.Goify(resultType.Name(), true),
				); err != nil {
					return err
				}
			}
		}
		seenUnions := make(map[expr.UserType]struct{})
		retainedUnions := make(map[codegen.UnionDeclarationID]struct{})
		for _, projected := range projectedRoots {
			if err := planViewUnions(projected, views, derived, seenUnions, retainedUnions, &serviceFacts.viewUnions); err != nil {
				return err
			}
		}
	}
	return nil
}

// collectProjectedTypeFacts selects validation and conversion code for one set
// of result types containing only the fields in their declared views.
func collectProjectedTypeFacts(pair *projectedTypePair) (*projectedTypeFacts, error) {
	facts := &projectedTypeFacts{
		pair:          pair,
		projectedType: pair.projected,
		validations:   collectValidationFacts(pair.projectedAttribute),
	}
	resultType, viewed := pair.projected.(*expr.ResultTypeExpr)
	if !viewed {
		return facts, nil
	}
	for _, view := range resultType.Views {
		object := expr.AsObject(view.Type)
		attributes := make([]string, len(*object))
		for index, field := range *object {
			attributes[index] = field.Name
		}
		facts.views = append(facts.views, &viewRenderFacts{
			name:        view.Name,
			description: view.Description,
			attributes:  attributes,
		})
	}
	for _, toResult := range []bool{true, false} {
		conversions, err := collectViewConversionFacts(pair.projectedAttribute, pair.sourceAttribute, toResult)
		if err != nil {
			return nil, err
		}
		facts.conversions = append(facts.conversions, conversions...)
	}
	return facts, nil
}

// collectValidationFacts stores the field checks and child validation calls
// selected for each result type and view.
func collectValidationFacts(projected *expr.AttributeExpr) []*validationFacts {
	userType := projected.Type.(expr.UserType)
	resultType, viewed := userType.(*expr.ResultTypeExpr)
	if !viewed {
		return []*validationFacts{{
			attribute: userType.Attribute(),
			alias:     expr.IsAlias(userType),
			pointer:   !expr.IsPrimitive(projected.Type),
		}}
	}
	facts := make([]*validationFacts, 0, len(resultType.Views))
	array := expr.AsArray(projected.Type)
	for _, view := range resultType.Views {
		validation := &validationFacts{viewName: view.Name, pointer: true}
		if array != nil {
			validation.collectionElem = array.ElemType
			facts = append(facts, validation)
			continue
		}
		object := &expr.Object{}
		walkViewAttrs(expr.AsObject(projected.Type), view, func(name string, attribute, viewAttribute *expr.AttributeExpr) {
			if _, ok := attribute.Type.(*expr.ResultTypeExpr); ok {
				selectedView := ""
				if explicit, ok := viewAttribute.Meta.Last(expr.ViewMetaKey); ok && explicit != expr.DefaultView {
					selectedView = explicit
				}
				validation.fields = append(validation.fields, &validationFieldFacts{
					name:      name,
					attribute: attribute,
					view:      selectedView,
					required:  resultType.Attribute().IsRequired(name),
				})
				return
			}
			object.Set(name, attribute)
		})
		validation.attribute = &expr.AttributeExpr{Type: object, Validation: resultType.Validation}
		facts = append(facts, validation)
	}
	return facts
}

// markNeededViewValidators keeps only functions that can return an error. A
// parent is needed when it checks one of its own fields or calls a needed child.
func markNeededViewValidators(facts *serviceFacts) {
	validations := make(map[viewValidationKey]*validationFacts)
	for _, projection := range facts.projections {
		for _, projected := range projection.types {
			for _, validation := range projected.validations {
				key := viewValidationKey{
					origin: projected.pair.projected.Origin(),
					view:   canonicalValidatorView(validation.viewName),
				}
				validations[key] = validation
				if validation.collectionElem == nil && codegen.NeedsValidation(validation.attribute, viewValidationPolicy()) {
					validation.needed = true
				}
				for _, field := range validation.fields {
					validation.needed = validation.needed || field.required
				}
			}
		}
	}

	for changed := true; changed; {
		changed = false
		for _, validation := range validations {
			if validation.needed {
				continue
			}
			if validation.collectionElem != nil && neededValidation(validations, validation.collectionElem, validation.viewName) {
				validation.needed = true
				changed = true
				continue
			}
			for _, field := range validation.fields {
				if neededValidation(validations, field.attribute, field.view) {
					validation.needed = true
					changed = true
					break
				}
			}
		}
	}
}

// neededValidation reports whether the selected view-specific result type can
// return a validation error.
func neededValidation(validations map[viewValidationKey]*validationFacts, attribute *expr.AttributeExpr, view string) bool {
	userType, ok := attribute.Type.(expr.UserType)
	if !ok {
		return false
	}
	validation := validations[viewValidationKey{
		origin: userType.Origin(),
		view:   canonicalValidatorView(view),
	}]
	return validation != nil && validation.needed
}

// collectViewConversionFacts selects the fields in each declared result view
// and stores the conversion used in each direction.
func collectViewConversionFacts(projected, service *expr.AttributeExpr, toResult bool) ([]*viewConversionFacts, error) {
	views := service.Type.(*expr.ResultTypeExpr).Views
	projectedObject := expr.AsObject(projected.Type)
	projectedArray := expr.AsArray(projected.Type)
	if projectedArray != nil {
		projectedObject = expr.AsObject(projectedArray.ElemType.Type)
	}
	result := make([]*viewConversionFacts, 0, len(views))
	for _, view := range views {
		object := &expr.Object{}
		walkViewAttrs(projectedObject, view, func(name string, attribute, _ *expr.AttributeExpr) {
			object.Set(name, attribute)
		})
		var narrowedType expr.DataType = object
		if projectedArray != nil {
			narrowedType = &expr.Array{ElemType: &expr.AttributeExpr{Type: &expr.ResultTypeExpr{
				UserTypeExpr: &expr.UserTypeExpr{
					AttributeExpr: &expr.AttributeExpr{Type: object},
					TypeName:      projectedArray.ElemType.Type.Name(),
				},
			}}}
		}
		narrowed := &expr.AttributeExpr{Type: &expr.ResultTypeExpr{
			UserTypeExpr: &expr.UserTypeExpr{
				AttributeExpr: &expr.AttributeExpr{Type: narrowedType},
				TypeName:      projected.Type.Name(),
			},
			Views:      views,
			Identifier: service.Type.(*expr.ResultTypeExpr).Identifier,
		}}
		source, target := service, narrowed
		if toResult {
			source, target = narrowed, service
		}
		conversion := &viewConversionFacts{
			toResult: toResult,
			viewName: view.Name,
			source:   source,
			target:   target,
		}
		if projectedArray == nil {
			conversion.transformTarget = expr.DupAtt(target)
			targetObject := expr.AsObject(conversion.transformTarget.Type)
			for _, field := range *targetObject {
				if _, nested := field.Attribute.Type.(*expr.ResultTypeExpr); !nested {
					continue
				}
				nestedView := ""
				if selected := source.Type.(*expr.ResultTypeExpr).View(view.Name).Find(field.Name); selected != nil {
					if explicit, ok := selected.Meta.Last(expr.ViewMetaKey); ok && explicit != expr.DefaultView {
						nestedView = explicit
					}
				}
				conversion.fields = append(conversion.fields, &viewConversionFieldFacts{
					name:      field.Name,
					attribute: field.Attribute,
					view:      nestedView,
				})
				targetObject.Delete(field.Name)
			}
			plan, err := codegen.NewTransformPlan(source, conversion.transformTarget, "", nil)
			if err != nil {
				return nil, err
			}
			conversion.plan = plan
		}
		result = append(result, conversion)
	}
	return result, nil
}

// planViewUnions submits every Goa OneOf type reachable from view-specific
// result types. Existing view types keep their declarations; a branch without
// one receives a generated alias declaration.
func planViewUnions(attribute *expr.AttributeExpr, generatedPackage *codegen.GeneratedPackage, derived map[expr.UserType]codegen.DerivedTypeID, seen map[expr.UserType]struct{}, retained map[codegen.UnionDeclarationID]struct{}, unions *[]*unionFacts) error {
	if attribute == nil || attribute.Type == expr.Empty {
		return nil
	}
	recurse := func(attribute *expr.AttributeExpr) error {
		return planViewUnions(attribute, generatedPackage, derived, seen, retained, unions)
	}
	switch actual := attribute.Type.(type) {
	case expr.UserType:
		origin := actual.Origin()
		if _, ok := seen[origin]; ok {
			return nil
		}
		seen[origin] = struct{}{}
		return recurse(actual.Attribute())
	case *expr.Object:
		for _, field := range *actual {
			if err := recurse(field.Attribute); err != nil {
				return err
			}
		}
	case *expr.Array:
		return recurse(actual.ElemType)
	case *expr.Map:
		if err := recurse(actual.KeyType); err != nil {
			return err
		}
		return recurse(actual.ElemType)
	case *expr.Union:
		if _, err := generatedPackage.DeclareUnion(attribute); err != nil {
			return err
		}
		identity := codegen.NewUnionDeclarationID(attribute)
		if _, exists := retained[identity]; !exists {
			retained[identity] = struct{}{}
			declaration, err := generatedPackage.Union(attribute)
			if err != nil {
				return err
			}
			*unions = append(*unions, &unionFacts{
				attribute:   attribute,
				union:       actual,
				identity:    identity,
				typeKey:     actual.GetTypeKey(),
				valueKey:    actual.GetValueKey(),
				location:    &codegen.Location{RelImportPath: "views"},
				declaration: declaration,
			})
		}
		for _, branch := range actual.Values {
			if userType, ok := branch.Attribute.Type.(expr.UserType); ok {
				if _, projected := derived[userType.Origin()]; !projected {
					if _, err := generatedPackage.DeclareUnionBranchType(attribute, branch.Name, userType); err != nil {
						return err
					}
				}
			}
			if err := recurse(branch.Attribute); err != nil {
				return err
			}
		}
	}
	return nil
}

// newRootTypeSet records the authored types in one design so compiler-created
// copies can use the same generated Go declarations. Generated Goa OneOf branch
// aliases are not added.
func newRootTypeSet(root *expr.RootExpr) *rootTypeSet {
	userTypes := &rootTypeSet{
		byOrigin: make(map[expr.UserType]expr.UserType, len(root.Types)+len(root.ResultTypes)+1),
	}
	for _, userType := range root.Types {
		userTypes.add(userType)
	}
	for _, resultType := range root.ResultTypes {
		userTypes.add(resultType)
	}
	userTypes.add(expr.ErrorResult)
	return userTypes
}

// generatedUnionBranch reports whether OneOf created a user type around a
// branch that was not declared as a user type in the design.
func generatedUnionBranch(branch *expr.NamedAttributeExpr, rootTypes *rootTypeSet) (expr.UserType, bool) {
	userType, ok := branch.Attribute.Type.(expr.UserType)
	if !ok {
		return nil, false
	}
	return userType, !rootTypes.contains(userType)
}

// add records one user type declared in this design under its original type.
func (s *rootTypeSet) add(userType expr.UserType) {
	s.byOrigin[userType.Origin()] = userType
}

// canonical returns the original authored declaration for a compiler-created
// copy from this design. Types originating in another design are returned
// unchanged.
func (s *rootTypeSet) canonical(userType expr.UserType) expr.UserType {
	if canonical, ok := s.byOrigin[userType.Origin()]; ok {
		return canonical
	}
	return userType
}

// contains reports whether userType was declared in this design or copied from
// one of its declarations.
func (s *rootTypeSet) contains(userType expr.UserType) bool {
	_, ok := s.byOrigin[userType.Origin()]
	return ok
}

// claimGeneratedPackage passes the relative path from design metadata to
// Generation unchanged. Generation can then reject two different path strings
// that resolve to the same output package. An absolute path is invalid metadata
// and does not select a package under the generated module.
func claimGeneratedPackage(generation *codegen.Generation, servicePath string, location *codegen.Location) (*codegen.GeneratedPackage, error) {
	if location == nil {
		return generation.ClaimPackage(servicePath)
	}
	if path.IsAbs(location.RelImportPath) {
		return nil, fmt.Errorf("generated package location %q must be relative", location.RelImportPath)
	}
	claim := strings.TrimSuffix(generation.GenPkg(), "/") + "/" + location.RelImportPath
	return generation.ClaimPackage(claim)
}

// generatedPackagePath returns the cleaned import path selected by location,
// or the service package when location is nil.
func generatedPackagePath(genpkg, servicePath string, location *codegen.Location) string {
	if location != nil {
		return path.Join(genpkg, location.RelImportPath)
	}
	return servicePath
}
