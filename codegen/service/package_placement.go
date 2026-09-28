// This file chooses shared packages from every design in one generation.
// Service plans retain the same answer without changing imported declarations.
package service

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

type (
	// packageClaim records the shared root and field path requiring a package.
	packageClaim struct {
		location string
		via      string
	}

	// placementVisit keeps a copy reachable under two different requirements
	// visible until ambiguity is checked, while terminating recursive graphs.
	placementVisit struct {
		userType expr.UserType
		location string
	}

	// originalTypeLayout keeps the exact root attribute used to plan one
	// original in one generated package.
	originalTypeLayout struct {
		origin    expr.UserType
		attribute *expr.AttributeExpr
		layout    *codegen.GoTypePlan
	}
)

// UserTypeLayout returns the retained root attribute and complete Go layout for
// an exact original/declaration pair from this generation's Generation.UserTypes.
// The attribute contains the canonical original and is the layout's exact root
// occurrence. Callers must use it with the layout and treat both as read-only.
// Every Plan from NewPlans can query every original in the generation, including
// originals emitted by another root. Each service-local copy has its own pair.
// The query is available before Freeze and never changes expressions or chooses
// an owner from a method, name, root order, or another generation.
func (p *Plan) UserTypeLayout(original expr.UserType, declaration *codegen.TypeDeclaration) (*expr.AttributeExpr, *codegen.GoTypePlan, error) {
	retained, exists := p.facts.rootTypes.originalLayouts[declaration]
	if original == nil || !exists || retained.origin != original.Origin() {
		return nil, nil, fmt.Errorf("original type and declaration are not a registered pair in this generation")
	}
	return retained.attribute, retained.layout, nil
}

// resolveLocations checks the complete shared-root graph before any service
// submits declarations. Repeated equal claims share one location; incompatible
// claims report both paths, independent of root or service order. Distinct
// explicitly located originals keep their own package when reached from a parent.
func (s *rootTypeSet) resolveLocations(roots []*expr.RootExpr) error {
	type sharedRoot struct {
		userType expr.UserType
		location string
		via      string
	}
	var shared []sharedRoot
	// Inspect each actual copy once so different annotations on copies of one
	// origin cannot disappear when that origin is recorded.
	seen := make(map[expr.UserType]struct{})
	collect := func(attribute *expr.AttributeExpr, via string) {
		walkPlacementTypes(attribute, seen, via, func(userType expr.UserType, via string) {
			if location := codegen.UserTypeLocation(userType); location != nil {
				shared = append(shared, sharedRoot{userType, location.RelImportPath, via})
			}
		})
	}
	orderedRoots := append([]*expr.RootExpr(nil), roots...)
	slices.SortFunc(orderedRoots, func(a, b *expr.RootExpr) int {
		return strings.Compare(a.API.Name, b.API.Name)
	})
	for _, root := range orderedRoots {
		for _, userType := range root.Types {
			collect(&expr.AttributeExpr{Type: userType}, root.API.Name+"."+userType.Name())
		}
		for _, result := range root.ResultTypes {
			collect(&expr.AttributeExpr{Type: result}, root.API.Name+"."+result.Name())
		}
		for _, rootError := range root.Errors {
			collect(rootError.AttributeExpr, root.API.Name+"."+rootError.Name)
		}
		services := append([]*expr.ServiceExpr(nil), root.Services...)
		slices.SortFunc(services, func(a, b *expr.ServiceExpr) int {
			return strings.Compare(a.Name, b.Name)
		})
		for _, service := range services {
			for _, serviceError := range service.Errors {
				collect(serviceError.AttributeExpr, root.API.Name+"."+service.Name+"."+serviceError.Name)
			}
			methods := append([]*expr.MethodExpr(nil), service.Methods...)
			slices.SortFunc(methods, func(a, b *expr.MethodExpr) int {
				return strings.Compare(a.Name, b.Name)
			})
			for _, method := range methods {
				prefix := root.API.Name + "." + service.Name + "." + method.Name
				collect(method.Payload, prefix+".payload")
				collect(method.StreamingPayload, prefix+".streaming_payload")
				collect(method.Result, prefix+".result")
				collect(method.StreamingResult, prefix+".streaming_result")
				for _, methodError := range method.Errors {
					collect(methodError.AttributeExpr, prefix+"."+methodError.Name)
				}
			}
		}
	}
	slices.SortFunc(shared, func(a, b sharedRoot) int {
		if order := strings.Compare(a.location, b.location); order != 0 {
			return order
		}
		return strings.Compare(a.via, b.via)
	})
	// Resolve every explicit copy first. Crossing a distinct located original
	// then uses its own package, including for its unlocated descendants.
	explicitClaims := make(map[expr.UserType]map[string]packageClaim)
	for _, root := range shared {
		origin := root.userType.Origin()
		if explicitClaims[origin] == nil {
			explicitClaims[origin] = make(map[string]packageClaim)
		}
		previous, exists := explicitClaims[origin][root.location]
		if !exists || root.via < previous.via {
			explicitClaims[origin][root.location] = packageClaim{root.location, root.via}
		}
	}
	explicit := make(map[expr.UserType]packageClaim, len(explicitClaims))
	var explicitConflicts []string
	for origin, selected := range explicitClaims {
		if len(selected) > 1 {
			explicitConflicts = append(explicitConflicts, placementConflict(origin, selected))
			continue
		}
		for _, claim := range selected {
			explicit[origin] = claim
		}
	}
	if len(explicitConflicts) > 0 {
		slices.Sort(explicitConflicts)
		return fmt.Errorf("explicit type package placement failed:\n%s", strings.Join(explicitConflicts, "\n"))
	}
	claims := make(map[expr.UserType]map[string]packageClaim)
	for _, root := range shared {
		s.inheritLocation(&expr.AttributeExpr{Type: root.userType}, root.location, root.via,
			explicit, claims, make(map[placementVisit]struct{}))
	}
	var conflicts []string
	for origin, selected := range claims {
		paths := make([]string, 0, len(selected))
		for selectedPath := range selected {
			paths = append(paths, selectedPath)
		}
		slices.Sort(paths)
		if len(paths) > 1 {
			conflicts = append(conflicts, placementConflict(origin, selected))
			continue
		}
		s.locations[origin] = &codegen.Location{
			RelImportPath: paths[0],
			FilePath:      filepath.Join(filepath.FromSlash(paths[0]), codegen.SnakeCase(s.canonical(origin).Name())+".go"),
		}
	}
	if len(conflicts) > 0 {
		slices.Sort(conflicts)
		return fmt.Errorf("shared type package placement failed:\n%s", strings.Join(conflicts, "\n"))
	}
	return nil
}

// location returns the selected package with the declaration's own filename.
// Unlocated ordinary values and generated union wrappers inherit their context.
func (s *rootTypeSet) location(dataType expr.DataType) *codegen.Location {
	userType, ok := dataType.(expr.UserType)
	if !ok {
		return nil
	}
	return s.locations[userType.Origin()]
}

// walkPlacementTypes visits named values through every schema edge. Tracking
// actual copies preserves origin-conflict evidence while terminating recursion.
func walkPlacementTypes(attribute *expr.AttributeExpr, seen map[expr.UserType]struct{}, via string, visit func(expr.UserType, string)) {
	if attribute == nil || attribute.Type == expr.Empty {
		return
	}
	recurse := func(attribute *expr.AttributeExpr, via string) {
		walkPlacementTypes(attribute, seen, via, visit)
	}
	switch actual := attribute.Type.(type) {
	case expr.UserType:
		if expr.IsErrorResult(actual) {
			return
		}
		if _, exists := seen[actual]; exists {
			return
		}
		seen[actual] = struct{}{}
		visit(actual, via)
		recurse(actual.Attribute(), via)
	case *expr.Object:
		for _, field := range sortedNamedAttributes(*actual) {
			recurse(field.Attribute, via+"."+field.Name)
		}
	case *expr.Array:
		recurse(actual.ElemType, via+"[]")
	case *expr.Map:
		recurse(actual.KeyType, via+"[key]")
		recurse(actual.ElemType, via+"[value]")
	case *expr.Union:
		for _, branch := range sortedNamedAttributes(actual.Values) {
			recurse(branch.Attribute, via+"."+branch.Name)
		}
	}
}

// retainOriginalLayouts saves each original's root attribute and complete Go
// layout after every root has submitted declarations. All service plans share
// these exact attribute/layout pairs.
func (s *rootTypeSet) retainOriginalLayouts(generation *codegen.Generation) error {
	for original, declaration := range generation.UserTypes() {
		attribute := &expr.AttributeExpr{Type: original}
		layout, err := codegen.PlanGoType(attribute, codegen.GoTypePlanOptions{
			Owner:            declaration.PackagePath(),
			Policy:           codegen.GoLayoutPolicy{UseDefault: true, SumType: true},
			RetainNamedValue: true,
			Bind:             serviceGoTypeBinder(s, generation),
		})
		if err != nil {
			return fmt.Errorf("plan original type %q in %q: %w", original.Name(), declaration.PackagePath(), err)
		}
		s.originalLayouts[declaration] = originalTypeLayout{
			origin:    original.Origin(),
			attribute: attribute,
			layout:    layout,
		}
	}
	return nil
}

// inheritLocation follows each schema edge with the selected enclosing package.
// An explicit original changes that package; it does not stop descendant checks.
func (s *rootTypeSet) inheritLocation(attribute *expr.AttributeExpr, location, via string, explicit map[expr.UserType]packageClaim, claims map[expr.UserType]map[string]packageClaim, seen map[placementVisit]struct{}) {
	if attribute == nil || attribute.Type == expr.Empty {
		return
	}
	recurse := func(attribute *expr.AttributeExpr, next string) {
		s.inheritLocation(attribute, location, next, explicit, claims, seen)
	}
	switch actual := attribute.Type.(type) {
	case expr.UserType:
		if expr.IsErrorResult(actual) {
			return
		}
		origin := actual.Origin()
		selected, located := explicit[origin]
		if located {
			location = selected.location
		}
		key := placementVisit{actual, location}
		if _, visited := seen[key]; visited {
			return
		}
		seen[key] = struct{}{}
		if s.contains(actual) || located {
			if claims[origin] == nil {
				claims[origin] = make(map[string]packageClaim)
			}
			previous, exists := claims[origin][location]
			if !exists || via < previous.via {
				claims[origin][location] = packageClaim{location, via}
			}
		}
		s.inheritLocation(actual.Attribute(), location, via, explicit, claims, seen)
	case *expr.Object:
		for _, field := range sortedNamedAttributes(*actual) {
			recurse(field.Attribute, via+"."+field.Name)
		}
	case *expr.Array:
		recurse(actual.ElemType, via+"[]")
	case *expr.Map:
		recurse(actual.KeyType, via+"[key]")
		recurse(actual.ElemType, via+"[value]")
	case *expr.Union:
		for _, branch := range sortedNamedAttributes(actual.Values) {
			recurse(branch.Attribute, via+"."+branch.Name)
		}
	}
}

// placementConflict reports every incompatible requirement in stable order.
func placementConflict(origin expr.UserType, selected map[string]packageClaim) string {
	paths := make([]string, 0, len(selected))
	for selectedPath := range selected {
		paths = append(paths, selectedPath)
	}
	slices.Sort(paths)
	descriptions := make([]string, 0, len(paths))
	for _, selectedPath := range paths {
		claim := selected[selectedPath]
		descriptions = append(descriptions, fmt.Sprintf("%q via %s", claim.location, claim.via))
	}
	return fmt.Sprintf("type %q requires incompatible packages: %s", origin.Name(), strings.Join(descriptions, "; "))
}
