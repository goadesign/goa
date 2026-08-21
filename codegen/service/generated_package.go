// This file binds service types selected by one design root to the generated
// packages that declare them. Planning records relocated user types and unions
// before names freeze; rendering stores one canonical section per declaration
// record so each package emits that declaration once.
package service

import (
	"fmt"
	"path"
	"slices"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

type (
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
		types  map[*codegen.TypeDeclaration]*generatedTypeData
		unions map[codegen.UnionTypeID]*UnionTypeData
	}

	// generatedTypeData owns one relocated user-type declaration and optional
	// error behavior at its metadata-selected file.
	generatedTypeData struct {
		declaration *codegen.TypeDeclaration
		userType    expr.UserType
		location    *codegen.Location
		section     *codegen.SectionTemplate
		error       *codegen.SectionTemplate
	}
)

// Plan declares every relocated user type and union reachable from root.
// User types are declared across the complete root before any union so exact
// user-authored names always take precedence over generated union names.
func Plan(root *expr.RootExpr, generation *codegen.Generation) error {
	if !generation.HasRoot(root) {
		return rootMembershipError(root)
	}
	inputs := planningInputs(root)
	rootTypes := newRootTypeSet(root)
	for _, service := range root.Services {
		// The service package record makes NewServicesData a render-only contract:
		// its scope is unavailable until the generation freezes.
		if _, err := generation.ClaimPackage(servicePackagePath(generation.GenPkg(), service)); err != nil {
			return err
		}
	}
	methodTypes, err := planMethodTypes(root, generation)
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
	if err := planViews(root, generation, rootTypes); err != nil {
		return err
	}
	return planImports(root, inputs, generation)
}

// rootMembershipError reports an attempt to plan or analyze a design root
// that the generation does not own.
func rootMembershipError(root *expr.RootExpr) error {
	return fmt.Errorf("service root %p does not belong to the generation", root)
}

// planMethodTypes declares the semantic wrappers created when NewGeneration
// takes ownership of raw method objects. Exact user types in the same package
// are planned separately and therefore keep their authored names.
func planMethodTypes(root *expr.RootExpr, generation *codegen.Generation) (map[expr.UserType]codegen.DerivedTypeID, error) {
	planned := make(map[expr.UserType]codegen.DerivedTypeID)
	for _, service := range root.Services {
		generatedPackage := generation.Package(servicePackagePath(generation.GenPkg(), service))
		for _, method := range service.Methods {
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
func planningInputs(root *expr.RootExpr) []plannedAttribute {
	var inputs []plannedAttribute
	for _, service := range root.Services {
		for _, serviceError := range service.Errors {
			inputs = append(inputs, plannedAttribute{attribute: serviceError.AttributeExpr, service: service})
		}
		for _, method := range service.Methods {
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
		for _, userType := range root.Types {
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
func planViews(root *expr.RootExpr, generation *codegen.Generation, rootTypes *rootTypeSet) error {
	for _, service := range root.Services {
		viewsPath := servicePackagePath(generation.GenPkg(), service) + "/views"
		views, err := generation.ClaimPackage(viewsPath)
		if err != nil {
			return err
		}
		seenProjected := make(map[expr.UserType]expr.UserType)
		derived := make(map[expr.UserType]codegen.DerivedTypeID)
		var projectedRoots []*expr.AttributeExpr
		for _, method := range service.Methods {
			if !hasResultType(method.Result) {
				continue
			}
			projected, source := projectedResultRoot(generation, method)
			pairs := projectTypePairs(projected, source, seenProjected)
			for _, pair := range pairs {
				identity := codegen.NewProjectedTypeID(pair.source)
				if _, err := views.DeclareDerivedType(identity, codegen.Goify(pair.projected.Name(), true)); err != nil {
					return err
				}
				derived[pair.projected.Origin()] = identity
			}
			removeMeta(projected)
			projectedRoots = append(projectedRoots, projected)

			if resultType, ok := method.Result.Type.(*expr.ResultTypeExpr); ok {
				serviceTypes := generation.Package(servicePackagePath(generation.GenPkg(), service))
				if _, err := serviceTypes.Type(rootTypes.canonical(resultType)); err != nil {
					return err
				}
				if _, err := views.DeclareDerivedType(
					codegen.NewViewedResultTypeID(resultType),
					codegen.Goify(resultType.Name(), true),
				); err != nil {
					return err
				}
			}
		}
		seenUnions := make(map[expr.UserType]struct{})
		for _, projected := range projectedRoots {
			if err := planViewUnions(projected, views, derived, seenUnions); err != nil {
				return err
			}
		}
	}
	return nil
}

// planViewUnions declares every union family reachable from one projected
// graph. Projected user types already own their derived declarations; only a
// branch without one is a generated alias owned by its union family.
func planViewUnions(attribute *expr.AttributeExpr, generatedPackage *codegen.GeneratedPackage, derived map[expr.UserType]codegen.DerivedTypeID, seen map[expr.UserType]struct{}) error {
	if attribute == nil || attribute.Type == expr.Empty {
		return nil
	}
	recurse := func(attribute *expr.AttributeExpr) error {
		return planViewUnions(attribute, generatedPackage, derived, seen)
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

// generatedPackage returns the root-owned render data for the package selected
// by location, creating that owner on first use.
func (d *ServicesData) generatedPackage(service *expr.ServiceExpr, location *codegen.Location) *generatedPackageData {
	importPath := generatedPackagePath(d.generation.GenPkg(), service, location)
	owner := d.generation.Package(importPath)
	if generatedPackage, ok := d.packages[owner]; ok {
		return generatedPackage
	}
	generatedPackage := &generatedPackageData{
		types:  make(map[*codegen.TypeDeclaration]*generatedTypeData),
		unions: make(map[codegen.UnionTypeID]*UnionTypeData),
	}
	d.packages[owner] = generatedPackage
	return generatedPackage
}

// registerPackageData gives each relocated user type's canonical declaration
// record one render section at its metadata-selected file.
func (d *ServicesData) registerPackageData(service *expr.ServiceExpr, data *Data) error {
	for i, method := range service.Methods {
		methodData := data.Methods[i]
		if err := d.registerMethodType(service, method.Payload, methodData.PayloadLoc, methodData.PayloadDef, &codegen.SectionTemplate{
			Name:   "service-payload",
			Source: serviceTemplates.Read(payloadT),
			Data:   methodData,
		}); err != nil {
			return err
		}
		if method.StreamingPayload != nil {
			if err := d.registerMethodType(service, method.StreamingPayload, codegen.UserTypeLocation(method.StreamingPayload.Type), methodData.StreamingPayloadDef, &codegen.SectionTemplate{
				Name:   "service-streaming-payload",
				Source: serviceTemplates.Read(streamingPayloadT),
				Data:   methodData,
			}); err != nil {
				return err
			}
		}
		if err := d.registerMethodType(service, method.Result, methodData.ResultLoc, methodData.ResultDef, &codegen.SectionTemplate{
			Name:   "service-result",
			Source: serviceTemplates.Read(resultT),
			Data:   methodData,
		}); err != nil {
			return err
		}
		if method.HasMixedResults() && method.StreamingResult != nil {
			if err := d.registerMethodType(service, method.StreamingResult, codegen.UserTypeLocation(method.StreamingResult.Type), methodData.StreamingResultDef, &codegen.SectionTemplate{
				Name:   "service-streaming-result",
				Source: serviceTemplates.Read(resultT),
				Data: map[string]any{
					"Result":     methodData.StreamingResult,
					"ResultDef":  methodData.StreamingResultDef,
					"ResultDesc": methodData.StreamingResultDesc,
				},
			}); err != nil {
				return err
			}
		}
	}
	for _, userType := range data.userTypes {
		if userType.Loc == nil {
			continue
		}
		d.registerType(service, userType.Declaration, userType.Type, userType.Loc, &codegen.SectionTemplate{
			Name:   "service-user-type",
			Source: serviceTemplates.Read(userTypeT),
			Data:   userType,
		})
	}
	for _, errorType := range data.errorTypes {
		if errorType.Loc == nil || errorType.Type == expr.ErrorResult {
			continue
		}
		d.registerType(service, errorType.Declaration, errorType.Type, errorType.Loc, &codegen.SectionTemplate{
			Name:   "error-user-type",
			Source: serviceTemplates.Read(userTypeT),
			Data:   errorType,
		})
		generatedType := d.generatedPackage(service, errorType.Loc).types[errorType.Declaration]
		if generatedType.error == nil {
			generatedType.error = &codegen.SectionTemplate{
				Name:    "service-error",
				Source:  serviceTemplates.Read(errorT),
				FuncMap: map[string]any{"errorName": errorName},
				Data:    errorType,
			}
		}
	}
	return nil
}

// registerMethodType records one relocated method payload or result when it
// has a generated declaration body.
func (d *ServicesData) registerMethodType(service *expr.ServiceExpr, attribute *expr.AttributeExpr, location *codegen.Location, definition string, section *codegen.SectionTemplate) error {
	if location == nil || definition == "" {
		return nil
	}
	userType := attribute.Type.(expr.UserType)
	declaration, err := d.generation.Package(
		generatedPackagePath(d.generation.GenPkg(), service, location),
	).UserType(d.rootTypes.canonical(userType))
	if err != nil {
		return err
	}
	d.registerType(service, declaration, userType, location, section)
	return nil
}

// registerType stores section under declaration. Repeated uses of the same
// canonical record retain the first root-order section and emit once.
func (d *ServicesData) registerType(service *expr.ServiceExpr, declaration *codegen.TypeDeclaration, userType expr.UserType, location *codegen.Location, section *codegen.SectionTemplate) {
	generatedPackage := d.generatedPackage(service, location)
	if _, ok := generatedPackage.types[declaration]; ok {
		return
	}
	generatedPackage.types[declaration] = &generatedTypeData{
		declaration: declaration,
		userType:    userType,
		location:    location,
		section:     section,
	}
}
