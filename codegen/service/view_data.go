// This file formats retained projected types, views, constructors, validators, and their exact declaration references.
package service

import (
	"bytes"
	"fmt"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// projectTypePairs rewrites a copied result graph into pointer-backed view
// types and returns each generated declaration with its exact source. The
// source Origin makes independently rebuilt plan and render graphs select the
// same package record.
func projectTypePairs(projected, source *expr.AttributeExpr, seen map[expr.UserType]expr.UserType) []*projectedTypePair {
	collect := func(projected, source *expr.AttributeExpr) []*projectedTypePair {
		return projectTypePairs(projected, source, seen)
	}
	switch projectedType := projected.Type.(type) {
	case expr.UserType:
		sourceType := source.Type.(expr.UserType)
		origin := sourceType.Origin()
		if existing, ok := seen[origin]; ok {
			if existing != nil {
				projected.Type = existing
			}
			return nil
		}
		seen[origin] = nil
		projectedType.Rename(projectedType.Name() + "View")
		nested := collect(projectedType.Attribute(), sourceType.Attribute())
		seen[origin] = projectedType
		return append([]*projectedTypePair{{
			source:             sourceType,
			projected:          projectedType,
			sourceAttribute:    source,
			projectedAttribute: projected,
		}}, nested...)
	case *expr.Array:
		return collect(projectedType.ElemType, source.Type.(*expr.Array).ElemType)
	case *expr.Map:
		sourceMap := source.Type.(*expr.Map)
		pairs := collect(projectedType.KeyType, sourceMap.KeyType)
		return append(pairs, collect(projectedType.ElemType, sourceMap.ElemType)...)
	case *expr.Object:
		sourceObject := source.Type.(*expr.Object)
		var pairs []*projectedTypePair
		for _, field := range *projectedType {
			pairs = append(pairs, collect(field.Attribute, sourceObject.Attribute(field.Name))...)
		}
		return pairs
	case *expr.Union:
		sourceUnion := source.Type.(*expr.Union)
		var pairs []*projectedTypePair
		for index, branch := range projectedType.Values {
			pairs = append(pairs, collect(branch.Attribute, sourceUnion.Values[index].Attribute)...)
		}
		return pairs
	default:
		return nil
	}
}

// projectedResultRoot returns the root attribute used to collect projected
// view types for m.Result. Compiler-created method wrappers retain their exact
// provenance in generation, so authored types with matching text stay intact.
func projectedResultRoot(generation *codegen.Generation, m *expr.MethodExpr) (*expr.AttributeExpr, *expr.AttributeExpr) {
	if ut, ok := m.Result.Type.(*expr.UserTypeExpr); ok {
		if _, normalized := generation.NormalizedMethodType(ut); !normalized {
			return expr.DupAtt(m.Result), m.Result
		}
		return expr.DupAtt(ut.Attribute()), ut.Attribute()
	}
	return expr.DupAtt(m.Result), m.Result
}

// hasResultType returns true if the given attribute has a result type recursively.
func hasResultType(att *expr.AttributeExpr, seens ...map[expr.UserType]struct{}) bool {
	if _, ok := att.Type.(*expr.ResultTypeExpr); ok {
		return true
	}
	var seen map[expr.UserType]struct{}
	if len(seens) > 0 {
		seen = seens[0]
	} else {
		seen = make(map[expr.UserType]struct{})
	}
	switch a := att.Type.(type) {
	case expr.UserType:
		origin := a.Origin()
		if _, ok := seen[origin]; ok {
			return false
		}
		seen[origin] = struct{}{}
		return hasResultType(a.Attribute(), seen)
	case *expr.Array:
		return hasResultType(a.ElemType, seen)
	case *expr.Map:
		return hasResultType(a.KeyType, seen) || hasResultType(a.ElemType, seen)
	case *expr.Object:
		for _, nat := range *a {
			if hasResultType(nat.Attribute, seen) {
				return true
			}
		}
	case *expr.Union:
		for _, nat := range a.Values {
			if hasResultType(nat.Attribute, seen) {
				return true
			}
		}
	}
	return false
}

// buildProjectedType returns the render data for one pointer-backed view
// declaration and its conversions to the exact source service type.
func buildProjectedType(facts *projectedTypeFacts, serviceResolver, viewResolver *declarationResolver, declaration *codegen.TypeDeclaration) *ProjectedTypeData {
	var (
		projections []*InitData
		typeInits   []*InitData
		views       []*ViewData

		varname = declaration.Name()
		pt      = facts.projectedType
	)
	if facts.resultType {
		typeInits = buildViewConversions(facts, serviceResolver, viewResolver, true)
		projections = buildViewConversions(facts, serviceResolver, viewResolver, false)
		serviceName := facts.source.Link(
			serviceResolver.outputPath,
			retainedTypeQualifier(serviceResolver.aliases),
		).Name()
		views = buildViews(facts.views, serviceName, facts.mapDeclaration, facts.conversions)
	}
	validations := buildValidations(facts, viewResolver)
	linked := facts.projected.Link(viewResolver.outputPath, retainedTypeQualifier(viewResolver.aliases))
	definition := facts.definition.Link(viewResolver.outputPath, retainedTypeQualifier(viewResolver.aliases))
	return &ProjectedTypeData{
		UserTypeData: &UserTypeData{
			Declaration: declaration,
			Name:        varname,
			Description: fmt.Sprintf("%s is a type that runs validations on a projected type.", varname),
			VarName:     varname,
			Def:         definition.Def(),
			Ref:         linked.Ref(),
			Type:        pt,
		},
		Projections: projections,
		TypeInits:   typeInits,
		Validations: validations,
		Views:       views,
	}
}

// buildViews builds the view data for all the views in the given result type.
func buildViews(facts []*viewRenderFacts, typeName string, mapDeclaration *codegen.NameDeclaration, conversions []*viewConversionFacts) []*ViewData {
	toProjected := make(map[string]*codegen.NameDeclaration)
	toResult := make(map[string]*codegen.NameDeclaration)
	for _, conversion := range conversions {
		calls := toProjected
		if conversion.toResult {
			calls = toResult
		}
		calls[canonicalValidatorView(conversion.viewName)] = conversion.constructor
	}
	views := make([]*ViewData, len(facts))
	for i, view := range facts {
		views[i] = &ViewData{
			Name:           view.name,
			Description:    view.description,
			Attributes:     append([]string(nil), view.attributes...),
			TypeVarName:    typeName,
			MapDeclaration: mapDeclaration,
			ToProjected:    toProjected[canonicalValidatorView(view.name)],
			ToResult:       toResult[canonicalValidatorView(view.name)],
		}
	}
	return views
}

// buildViewedResultType formats the retained viewed result wrapper and its
// constructors without consulting the mutable design expression.
func buildViewedResultType(facts *viewedResultFacts, viewspkg string, serviceResolver, viewResolver *declarationResolver, declaration *codegen.TypeDeclaration) *ViewedResultTypeData {
	isarr := facts.isCollection
	viewName := facts.viewName
	views := buildViews(facts.views, declaration.Name(), facts.mapDeclaration, facts.conversions)

	// build validation data
	qualifier := retainedTypeQualifier(serviceResolver.aliases)
	serviceType := facts.source.layout.Link(serviceResolver.outputPath, qualifier)
	resvar, serviceRef := declaration.Name(), serviceType.Ref()
	projT := facts.wrapped
	wrapperViewType := facts.wrappedLayout.Link(viewResolver.outputPath, qualifier)
	resref := wrapperViewType.Name()
	if !isarr {
		resref = "*" + resref
	}
	validationCalls := make([]*ValidationCallData, len(facts.views))
	for index, view := range facts.views {
		validationCalls[index] = newRetainedValidationCall(facts.validationCalls[index], view.name)
	}
	data := map[string]any{
		"ArgVar":          "result",
		"Source":          "result",
		"ValidationCalls": validationCalls,
		"IsViewed":        true,
	}
	buf := &bytes.Buffer{}
	if err := validateTypeCodeTmpl.Execute(buf, data); err != nil {
		panic(err) // bug
	}
	validatorDeclaration := facts.validator
	name := validatorDeclaration.Name()
	validate := &ValidateData{
		Declaration: validatorDeclaration,
		Description: fmt.Sprintf("%s runs the validations defined on the viewed result type %s.", name, resvar),
		Ref:         resref,
		Validate:    buf.String(),
		Calls:       validationCalls,
	}

	// build constructor to initialize viewed result type from result type
	wrapperServiceType := facts.wrappedLayout.Link(serviceResolver.outputPath, qualifier)
	vresref := wrapperServiceType.Name()
	if !isarr {
		vresref = "*" + vresref
	}
	data = map[string]any{
		"ToViewed":      true,
		"ArgVar":        "res",
		"ReturnVar":     "vres",
		"Views":         views,
		"ReturnTypeRef": vresref,
		"IsCollection":  isarr,
		"TargetType":    wrapperServiceType.Name(),
	}
	buf = &bytes.Buffer{}
	if err := initTypeCodeTmpl.Execute(buf, data); err != nil {
		panic(err) // bug
	}
	name = facts.toViewed.Name()
	init := &InitData{
		Declaration: facts.toViewed,
		Description: fmt.Sprintf("%s initializes viewed result type %s from result type %s using the given view.", name, resvar, resvar),
		Args: []*InitArgData{
			{Name: "res", Ref: serviceRef},
			{Name: "view", Ref: "string"},
		},
		ReturnTypeRef: vresref,
		Code:          buf.String(),
	}

	// build constructor to initialize result type from viewed result type
	resref = serviceRef
	data = map[string]any{
		"ToResult":      true,
		"ArgVar":        "vres",
		"ReturnVar":     "res",
		"Views":         views,
		"ReturnTypeRef": resref,
	}
	buf = &bytes.Buffer{}
	if err := initTypeCodeTmpl.Execute(buf, data); err != nil {
		panic(err) // bug
	}
	name = facts.toResult.Name()
	resinit := &InitData{
		Declaration:   facts.toResult,
		Description:   fmt.Sprintf("%s initializes result type %s from viewed result type %s.", name, resvar, resvar),
		Args:          []*InitArgData{{Name: "vres", Ref: vresref}},
		ReturnTypeRef: resref,
		Code:          buf.String(),
	}

	return &ViewedResultTypeData{
		UserTypeData: &UserTypeData{
			Declaration: declaration,
			Name:        resvar,
			Description: fmt.Sprintf("%s is the viewed result type that is projected based on a view.", resvar),
			VarName:     resvar,
			Def:         facts.wrappedDef.Link(viewResolver.outputPath, qualifier).Def(),
			Ref:         resref,
			Type:        projT,
		},
		FullName:     wrapperServiceType.Name(),
		FullRef:      vresref,
		ResultInit:   resinit,
		Init:         init,
		Views:        views,
		Validate:     validate,
		IsCollection: isarr,
		ViewName:     viewName,
		ViewsPkg:     viewspkg,
	}
}

// wrapProjected builds a viewed result type by wrapping the given projected
// in a result type with "projected" and "view" attributes.
func wrapProjected(projected expr.UserType) expr.UserType {
	rt := projected.(*expr.ResultTypeExpr)
	pratt := &expr.NamedAttributeExpr{
		Name:      "projected",
		Attribute: &expr.AttributeExpr{Type: rt, Description: "Type to project"},
	}
	prview := &expr.NamedAttributeExpr{
		Name:      "view",
		Attribute: &expr.AttributeExpr{Type: expr.String, Description: "View to render"},
	}
	return &expr.ResultTypeExpr{
		UserTypeExpr: &expr.UserTypeExpr{
			AttributeExpr: &expr.AttributeExpr{
				Type:       &expr.Object{pratt, prview},
				Validation: &expr.ValidationExpr{Required: []string{"projected", "view"}},
			},
			TypeName: rt.TypeName,
		},
		Identifier: rt.Identifier,
		Views:      rt.Views,
	}
}

// buildViewConversions builds the data to generate the constructor code that
// converts between a result type and its projected type, one constructor per
// view. When toResult is true the constructors initialize the result type from
// the projected type, otherwise they project the result type to the projected
// type based on the view.
func buildViewConversions(facts *projectedTypeFacts, serviceResolver, viewResolver *declarationResolver, toResult bool) []*InitData {
	init := make([]*InitData, 0, len(facts.conversions)/2)
	serviceType := facts.source.Link(serviceResolver.outputPath, retainedTypeQualifier(serviceResolver.aliases))
	serviceName, serviceRef := serviceType.Name(), serviceType.Ref()
	projectedDeclaration := facts.declaration
	projectedRef := facts.projected.Link(serviceResolver.outputPath, retainedTypeQualifier(serviceResolver.aliases)).Ref()
	serviceViewResolver := viewResolver.withOutputPackage(serviceResolver.outputPath)
	for _, conversion := range facts.conversions {
		if conversion.toResult != toResult {
			continue
		}
		viewedResolver := serviceViewResolver.bindDerived(conversion.contextType, conversion.contextIdentity)
		if conversion.elementType != nil {
			viewedResolver = viewedResolver.bindDerived(conversion.elementType, conversion.elementIdentity)
		}
		targetType := conversion.targetLayout.Link(
			serviceResolver.outputPath,
			retainedTypeQualifier(serviceResolver.aliases),
		).Name()
		if toResult {
			srcCtx := declarationContext(viewedResolver, true)
			tgtCtx := declarationContext(serviceResolver, false)
			resvar := serviceName
			name := conversion.constructor.Name()
			code, helpers := buildConstructorCode(
				conversion,
				"vres",
				"res",
				srcCtx,
				tgtCtx,
				targetType,
			)
			init = append(init, &InitData{
				Declaration:   conversion.constructor,
				Description:   fmt.Sprintf("%s converts projected type %s to service type %s.", name, resvar, resvar),
				Args:          []*InitArgData{{Name: "vres", Ref: projectedRef}},
				ReturnTypeRef: serviceRef,
				Code:          code,
				Helpers:       helpers,
			})
		} else {
			srcCtx := declarationContext(serviceResolver, false)
			tgtCtx := declarationContext(viewedResolver, true)
			tname := projectedDeclaration.Name()
			name := conversion.constructor.Name()
			code, helpers := buildConstructorCode(
				conversion,
				"res",
				"vres",
				srcCtx,
				tgtCtx,
				targetType,
			)
			init = append(init, &InitData{
				Declaration:   conversion.constructor,
				Description:   fmt.Sprintf("%s projects result type %s to projected type %s using the %q view.", name, serviceName, tname, conversion.viewName),
				Args:          []*InitArgData{{Name: "res", Ref: serviceRef}},
				ReturnTypeRef: projectedRef,
				Code:          code,
				Helpers:       helpers,
			})
		}
	}
	return init
}

// buildValidations builds the data required to generate validations for the
// projected types.
func buildValidations(projected *projectedTypeFacts, resolver *declarationResolver) []*ValidateData {
	linkedType := projected.projected.Link(resolver.outputPath, retainedTypeQualifier(resolver.aliases))
	tname := linkedType.Name()
	var validations []*ValidateData
	if projected.resultType {
		// for result types we create a validation function containing view
		// specific validation logic for each view
		for _, facts := range projected.validations {
			viewName := facts.viewName
			data := map[string]any{
				"Projected":    tname,
				"ArgVar":       "result",
				"Source":       "result",
				"IsCollection": facts.collectionElem != nil,
			}
			declaration := facts.declaration
			name := declaration.Name()
			var calls []*ValidationCallData

			if facts.collectionElem != nil {
				// dealing with an array type
				data["Source"] = "item"
				call := newRetainedValidationCall(facts.collectionCall, viewName)
				data["ValidateCall"] = call
				calls = append(calls, call)
			} else {
				fields := make([]*validationFieldData, 0, len(facts.fields))
				for _, field := range facts.fields {
					call := newRetainedValidationCall(field.call, field.view)
					fields = append(fields, &validationFieldData{
						Name:       field.name,
						Call:       call,
						IsRequired: field.required,
					})
					calls = append(calls, call)
				}
				data["Validate"] = renderRetainedValidation(facts, resolver)
				data["Fields"] = fields
			}

			buf := &bytes.Buffer{}
			if err := validateTypeCodeTmpl.Execute(buf, data); err != nil {
				panic(err) // bug
			}

			validations = append(validations, &ValidateData{
				Declaration: declaration,
				Description: fmt.Sprintf("%s runs the validations defined on %s using the %q view.", name, tname, viewName),
				Ref:         linkedType.Ref(),
				Validate:    buf.String(),
				Calls:       calls,
			})
		}
	} else {
		// for a user type or a result type with single view, we generate only one validation
		// function containing the validation logic
		facts := projected.validations[0]
		declaration := facts.declaration
		name := declaration.Name()
		validations = append(validations, &ValidateData{
			Declaration: declaration,
			Description: fmt.Sprintf("%s runs the validations defined on %s.", name, tname),
			Ref:         linkedType.Ref(),
			Validate:    renderRetainedValidation(facts, resolver),
		})
	}
	return validations
}

// renderRetainedValidation formats a symbolic validation plan against the
// frozen view-package declarations and aliases.
func renderRetainedValidation(facts *validationFacts, resolver *declarationResolver) string {
	linkedLayout := facts.layout.Link(resolver.outputPath, retainedTypeQualifier(resolver.aliases))
	linked, err := facts.plan.Link(linkedLayout)
	if err != nil {
		panic(err) // bug
	}
	return linked.Render("result", "result")
}

// newRetainedValidationCall formats one nested call from the exact declaration
// bound during planning.
func newRetainedValidationCall(declaration *codegen.NameDeclaration, view string) *ValidationCallData {
	return &ValidationCallData{
		Declaration: declaration,
		View:        view,
		Default:     canonicalValidatorView(view) == "",
	}
}

// newValidationCall binds a nested call spelling to the exact validator
// declaration retained for attribute and view.
// buildConstructorCode builds the transformation code to create a projected
// type from a service type and vice versa.
//
// source and target contains the projected/service contextual attributes
//
// sourceVar and targetVar contains the variable name that holds the source and
// target data structures in the transformation code.
//
// view is used to generate the constructor function name.
func buildConstructorCode(facts *viewConversionFacts, sourceVar, targetVar string, sourceCtx, targetCtx *codegen.AttributeContext, targetType string) (string, []*codegen.TransformFunctionData) {
	var (
		helpers []*codegen.TransformFunctionData
		buf     bytes.Buffer
	)
	data := map[string]any{
		"ArgVar":       sourceVar,
		"ReturnVar":    targetVar,
		"IsCollection": facts.collection,
		"TargetType":   targetType,
	}

	if facts.collection {
		// result type collection
		data["Init"] = facts.elementCall
		if err := initTypeCodeTmpl.Execute(&buf, data); err != nil {
			panic(err) // bug
		}
		return buf.String(), helpers
	}

	data["Source"] = sourceVar
	data["Target"] = targetVar

	if err := facts.plan.BindContexts(sourceCtx, targetCtx); err != nil {
		panic(err) // bug
	}
	code, helpers, err := facts.plan.Render(sourceVar, targetVar, true)
	if err != nil {
		panic(err) // bug
	}
	data["Code"] = code

	fields := make([]*constructorFieldData, 0, len(facts.fields))
	for _, field := range facts.fields {
		fields = append(fields, &constructorFieldData{
			VarName:     codegen.Goify(field.name, true),
			Declaration: field.call,
		})
	}
	data["Fields"] = fields

	if err := initTypeCodeTmpl.Execute(&buf, data); err != nil {
		panic(err) // bug
	}
	return buf.String(), helpers
}

// walkViewAttrs iterates through the attributes in att that are found in the
// given view and executes the walker function.
func walkViewAttrs(obj *expr.Object, view *expr.ViewExpr, walker func(name string, attr, vatt *expr.AttributeExpr)) {
	for _, nat := range *expr.AsObject(view.Type) {
		if attr := obj.Attribute(nat.Name); attr != nil {
			walker(nat.Name, attr, nat.Attribute)
		}
	}
}

// removeMeta removes the meta attributes from the given attribute. This is
// needed to make sure that any field name overriding is removed when
// generating protobuf types (as protogen itself won't honor these overrides).
func removeMeta(att *expr.AttributeExpr) {
	if err := codegen.Walk(att, func(a *expr.AttributeExpr) error {
		delete(a.Meta, "struct:pkg:path")
		return nil
	}); err != nil {
		panic(err) // bug
	}
}
