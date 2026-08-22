// This file binds service types selected by one design root to the generated
// packages that declare them. Planning records relocated user types and unions
// before names freeze; rendering stores one canonical section per declaration
// record so each package emits that declaration once.
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
	// generatedTypeEmissionKind identifies the template family that defines one
	// retained relocated type declaration.
	generatedTypeEmissionKind uint8

	// generatedTypeEmissionFacts selects one canonical relocated declaration
	// before names freeze and retains the linked data owner used by rendering.
	generatedTypeEmissionFacts struct {
		kind        generatedTypeEmissionKind
		declaration *codegen.TypeDeclaration
		location    *codegen.Location
		root        *rootFacts
		service     *serviceFacts
		method      *methodFacts
		attribute   *methodAttributeFacts
		userType    *userTypeFacts
		error       bool
	}

	// generatedUnionEmissionFacts selects one canonical union definition before
	// names freeze and retains its linked render data.
	generatedUnionEmissionFacts struct {
		root    *rootFacts
		service *serviceFacts
		union   *unionFacts
	}

	// plannedAttribute identifies one service attribute and the package inherited
	// by nested types that do not select their own struct:pkg:path location.
	plannedAttribute struct {
		attribute *expr.AttributeExpr
		service   *expr.ServiceExpr
		location  *codegen.Location
	}

	// plannedUserType identifies one user type emitted in one generated package.
	// The same expression may be copied into two packages through Extend.
	plannedUserType struct {
		userType expr.UserType
		owner    *codegen.GeneratedPackage
	}

	// unionBranch identifies a generated user type that exists only to name one
	// branch of its owning union.
	unionBranch struct {
		union *expr.Union
		name  string
	}

	// rootTypeSet maps compiler-created copies back to the exact DSL declaration
	// in the same design root. Generated union aliases have different typed
	// origins and are not included.
	rootTypeSet struct {
		byOrigin map[expr.UserType]expr.UserType
	}

	// generatedPackageData owns the render data emitted into one Go package.
	generatedPackageData struct {
		types        map[*codegen.TypeDeclaration]*generatedTypeData
		unions       map[*codegen.UnionDeclaration]*UnionTypeData
		unionImports []*codegen.ImportSpec
	}

	// generatedTypeData owns one relocated user-type declaration and optional
	// error behavior at its metadata-selected file.
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

// collectServiceDeclarations declares every relocated user type and union reachable from root.
// User types are declared across the complete root before any union so exact
// user-authored names always take precedence over generated union names.
func collectServiceDeclarations(facts *rootFacts, generation *codegen.Generation) error {
	if !generation.HasRoot(facts.root) {
		return rootMembershipError(facts.root)
	}
	inputs := planningInputs(facts)
	rootTypes := facts.rootTypes
	for _, serviceFacts := range facts.services {
		service := serviceFacts.service
		// The service package record makes NewServicesData a render-only contract:
		// its scope is unavailable until the generation freezes.
		if _, err := generation.ClaimPackage(servicePackagePath(generation.GenPkg(), service)); err != nil {
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

// collectGeneratedPackageEmissions selects one owner for every relocated type
// and union declaration across all roots before the generation freezes.
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

// selectGeneratedTypeEmission coalesces identical declaration candidates and
// rejects candidates that would give one declaration two emitted contracts.
func selectGeneratedTypeEmission(selected map[*codegen.TypeDeclaration]*generatedTypeEmissionFacts, candidate *generatedTypeEmissionFacts) error {
	existing := selected[candidate.declaration]
	if existing == nil {
		selected[candidate.declaration] = candidate
		return nil
	}
	if err := validateGeneratedTypeEmission(existing, candidate); err != nil {
		return err
	}
	if generatedTypeEmissionLess(candidate, existing) {
		selected[candidate.declaration] = candidate
	}
	return nil
}

// validateGeneratedTypeEmission enforces the one-definition contract for one
// canonical generated type declaration.
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

// generatedTypeEmissionLayout returns the exact retained definition written by
// one emission candidate. References belong to consuming service files and do
// not participate in ownership of this declaration's definition.
func generatedTypeEmissionLayout(emission *generatedTypeEmissionFacts) *codegen.GoTypePlan {
	if emission.userType != nil {
		return emission.userType.layout
	}
	return emission.attribute.definition
}

// sameGeneratedTypeEmissionContent compares the retained comments and error
// behavior that can change the bytes emitted for one declaration.
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

// sameGeneratedTypeEmissionSource compares exact authored sources while
// recognizing generated union branch aliases already proven compatible by the
// package's canonical branch declaration.
func sameGeneratedTypeEmissionSource(left, right *generatedTypeEmissionFacts) bool {
	if generatedTypeEmissionOrigin(left) == generatedTypeEmissionOrigin(right) {
		return true
	}
	return left.userType != nil && right.userType != nil &&
		!left.root.rootTypes.contains(left.userType.userType) &&
		!right.root.rootTypes.contains(right.userType.userType)
}

// generatedTypeEmissionName describes the exact authored or normalized source
// in a planning conflict diagnostic.
func generatedTypeEmissionName(emission *generatedTypeEmissionFacts) string {
	origin := generatedTypeEmissionOrigin(emission)
	if origin == nil {
		return ""
	}
	return origin.Name()
}

// validateGeneratedUnionEmission enforces one location and shape for a
// canonical generated union declaration.
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

// sameGeneratedUnionBranches compares every fact that changes one canonical
// union declaration's type, constructors, validation, or JSON helpers.
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
// name is not part of union ownership.
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

// generatedTypeEmissionLess orders equivalent candidates by stable service
// and method facts, never by root traversal position.
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

// generatedUnionEmissionLess orders equivalent union candidates by their
// stable package and service ownership facts.
func generatedUnionEmissionLess(left, right *generatedUnionEmissionFacts) bool {
	if left.union.declaration.PackagePath() != right.union.declaration.PackagePath() {
		return left.union.declaration.PackagePath() < right.union.declaration.PackagePath()
	}
	return left.service.packagePath < right.service.packagePath
}

// rootMembershipError reports an attempt to plan or analyze a design root
// that the generation does not own.
func rootMembershipError(root *expr.RootExpr) error {
	return fmt.Errorf("service root %p does not belong to the generation", root)
}

// planMethodTypes declares the semantic wrappers created when NewGeneration
// takes ownership of raw method objects. Exact user types in the same package
// are planned separately and therefore keep their authored names.
func planMethodTypes(facts *rootFacts, generation *codegen.Generation) (map[expr.UserType]codegen.DerivedTypeID, error) {
	planned := make(map[expr.UserType]codegen.DerivedTypeID)
	for _, serviceFacts := range facts.services {
		service := serviceFacts.service
		generatedPackage := generation.Package(servicePackagePath(generation.GenPkg(), service))
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

// planningInputs returns the service attributes that can cause service types
// to be emitted. Unused root types are deliberately excluded.
func planningInputs(facts *rootFacts) []plannedAttribute {
	var inputs []plannedAttribute
	for _, serviceFacts := range facts.services {
		service := serviceFacts.service
		for _, serviceError := range serviceFacts.errors {
			inputs = append(inputs, plannedAttribute{attribute: serviceError.AttributeExpr, service: service})
		}
		for _, method := range serviceFacts.methods {
			inputs = append(inputs,
				plannedAttribute{attribute: method.Payload, service: service},
				plannedAttribute{attribute: method.StreamingPayload, service: service},
				plannedAttribute{attribute: method.Result, service: service},
			)
			if method.HasMixedResults() {
				inputs = append(inputs, plannedAttribute{attribute: method.StreamingResult, service: service})
			}
			for _, methodError := range method.Errors {
				inputs = append(inputs, plannedAttribute{attribute: methodError.AttributeExpr, service: service})
			}
		}
		for _, userType := range facts.types {
			services, ok := userType.Attribute().Meta["type:generate:force"]
			if !ok || len(services) > 0 && !slices.Contains(services, service.Name) {
				continue
			}
			inputs = append(inputs, plannedAttribute{
				attribute: &expr.AttributeExpr{Type: userType},
				service:   service,
			})
		}
	}
	return inputs
}

// planUserTypes traverses attribute and declares each relocated user type in
// the package selected by its own or its enclosing type's metadata.
func planUserTypes(attribute *expr.AttributeExpr, service *expr.ServiceExpr, location *codegen.Location, generation *codegen.Generation, rootTypes *rootTypeSet, methodTypes map[expr.UserType]codegen.DerivedTypeID, seen map[plannedUserType]struct{}) error {
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
		owner, err := claimGeneratedPackage(generation, service, typeLocation)
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

// planUnions traverses attribute after all user types have been declared and
// records each relocated union in its owning package.
func planUnions(attribute *expr.AttributeExpr, service *expr.ServiceExpr, location *codegen.Location, generation *codegen.Generation, rootTypes *rootTypeSet, seen map[plannedUserType]struct{}) error {
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
		owner, err := claimGeneratedPackage(generation, service, typeLocation)
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
		generatedPackage, err := claimGeneratedPackage(generation, service, location)
		if err != nil {
			return err
		}
		if _, err := generatedPackage.DeclareUnion(actual); err != nil {
			return err
		}
		for _, named := range actual.Values {
			if userType, ok := generatedUnionBranch(named, rootTypes); ok {
				if _, err := generatedPackage.DeclareUnionBranchType(actual, named.Name, userType); err != nil {
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

// planViews rebuilds the same projected expression graph used by rendering,
// declares every derived view type, and then declares view-local union
// families after the derived type names have been recorded.
func planViews(facts *rootFacts, generation *codegen.Generation) error {
	for _, serviceFacts := range facts.services {
		service := serviceFacts.service
		viewsPath := servicePackagePath(generation.GenPkg(), service) + "/views"
		views, err := generation.ClaimPackage(viewsPath)
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
		retainedUnions := make(map[codegen.UnionTypeID]struct{})
		for _, projected := range projectedRoots {
			if err := planViewUnions(projected, views, derived, seenUnions, retainedUnions, &serviceFacts.viewUnions); err != nil {
				return err
			}
		}
	}
	return nil
}

// collectProjectedTypeFacts selects validators and view-narrowed conversions
// from one projected graph before package names freeze.
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

// collectValidationFacts retains the exact attributes and child validators
// selected for each projected type view.
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
			if nested, ok := attribute.Type.(*expr.ResultTypeExpr); ok {
				selectedView := ""
				if explicit, ok := viewAttribute.Meta.Last(expr.ViewMetaKey); ok && explicit != expr.DefaultView {
					selectedView = explicit
				}
				validation.fields = append(validation.fields, &validationFieldFacts{
					name:      name,
					attribute: attribute,
					view:      selectedView,
					required:  nested.Attribute().IsRequired(name),
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

// collectViewConversionFacts narrows a projected result to each declared view
// and retains the transform operation used in the selected direction.
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
			plan, err := codegen.NewTransformPlan(source, conversion.transformTarget)
			if err != nil {
				return nil, err
			}
			conversion.plan = plan
		}
		result = append(result, conversion)
	}
	return result, nil
}

// planViewUnions declares every union family reachable from one projected
// graph. Projected user types already own their derived declarations; only a
// branch without one is a generated alias owned by its union family.
func planViewUnions(attribute *expr.AttributeExpr, generatedPackage *codegen.GeneratedPackage, derived map[expr.UserType]codegen.DerivedTypeID, seen map[expr.UserType]struct{}, retained map[codegen.UnionTypeID]struct{}, unions *[]*unionFacts) error {
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
		if _, err := generatedPackage.DeclareUnion(actual); err != nil {
			return err
		}
		identity := codegen.NewUnionTypeID(actual)
		if _, exists := retained[identity]; !exists {
			retained[identity] = struct{}{}
			declaration, err := generatedPackage.Union(actual)
			if err != nil {
				return err
			}
			*unions = append(*unions, &unionFacts{
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
					if _, err := generatedPackage.DeclareUnionBranchType(actual, branch.Name, userType); err != nil {
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

// newRootTypeSet records the exact declarations whose compiler-created copies
// share package records. Generated union aliases have independent origins
// and never enter this set.
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

// generatedUnionBranch identifies the user type synthesized by OneOf around a
// branch that was not itself an exact DSL user-type declaration.
func generatedUnionBranch(branch *expr.NamedAttributeExpr, rootTypes *rootTypeSet) (expr.UserType, bool) {
	userType, ok := branch.Attribute.Type.(expr.UserType)
	if !ok {
		return nil, false
	}
	return userType, !rootTypes.contains(userType)
}

// add records one exact root declaration under its typed origin.
func (s *rootTypeSet) add(userType expr.UserType) {
	s.byOrigin[userType.Origin()] = userType
}

// canonical maps only a compiler copy whose typed origin belongs to this root
// back to its exact declaration.
func (s *rootTypeSet) canonical(userType expr.UserType) expr.UserType {
	if canonical, ok := s.byOrigin[userType.Origin()]; ok {
		return canonical
	}
	return userType
}

// contains reports whether userType is an exact root declaration or one of
// its compiler copies.
func (s *rootTypeSet) contains(userType expr.UserType) bool {
	_, ok := s.byOrigin[userType.Origin()]
	return ok
}

// claimGeneratedPackage preserves the relative path spelling supplied by
// design metadata so Generation can reject two claims that resolve to one
// output package. An absolute path violates the metadata contract instead of
// selecting a package beneath the generated module by string concatenation.
func claimGeneratedPackage(generation *codegen.Generation, service *expr.ServiceExpr, location *codegen.Location) (*codegen.GeneratedPackage, error) {
	if location == nil {
		return generation.ClaimPackage(servicePackagePath(generation.GenPkg(), service))
	}
	if path.IsAbs(location.RelImportPath) {
		return nil, fmt.Errorf("generated package location %q must be relative", location.RelImportPath)
	}
	claim := strings.TrimSuffix(generation.GenPkg(), "/") + "/" + location.RelImportPath
	return generation.ClaimPackage(claim)
}

// generatedPackagePath returns the canonical import path selected by location,
// or the service package when location is nil.
func generatedPackagePath(genpkg string, service *expr.ServiceExpr, location *codegen.Location) string {
	if location != nil {
		return path.Join(genpkg, location.RelImportPath)
	}
	return servicePackagePath(genpkg, service)
}

// servicePackagePath returns the actual import path of service's generated Go
// package.
func servicePackagePath(genpkg string, service *expr.ServiceExpr) string {
	return path.Join(genpkg, codegen.SnakeCase(codegen.Goify(service.Name, false)))
}
