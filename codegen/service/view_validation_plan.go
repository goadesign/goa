// This file records validation for result types narrowed to their declared
// views. Each field check and child call uses the Go type and function
// declarations submitted for the generated views package.
package service

import (
	"fmt"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// viewValidationPolicy states that generated view fields use pointers, apply
// defaults, and represent Goa OneOf values with generated structs. Both type
// layout and validation use these same choices.
func viewValidationPolicy() codegen.GoLayoutPolicy {
	return codegen.GoLayoutPolicy{
		Pointer:    true,
		UseDefault: true,
		SumType:    true,
	}
}

// planServiceValidations records every field check and child validation call
// written for service result views after all view type names are submitted.
func planServiceValidations(facts *serviceFacts, rootTypes *rootTypeSet, generation *codegen.Generation) error {
	hasProjection := false
	for _, method := range facts.orderedMethods {
		hasProjection = hasProjection || method.projection != nil
	}
	if !hasProjection {
		return nil
	}
	views := generation.Package(facts.viewsPath)
	derived := make(map[expr.UserType]*codegen.TypeDeclaration)
	wrappers := make(map[expr.UserType]*codegen.TypeDeclaration)
	for _, method := range facts.orderedMethods {
		if method.viewedResult != nil {
			wrappers[method.viewedResult.wrapped] = method.viewedResult.declaration
		}
		if method.projection == nil {
			continue
		}
		for _, projected := range method.projection.types {
			declaration, err := views.DerivedType(codegen.NewProjectedTypeID(projected.pair.source))
			if err != nil {
				return err
			}
			derived[projected.pair.projected.Origin()] = declaration
		}
	}
	binder := func(request codegen.GoTypeBindingRequest) (codegen.GoTypeBinding, error) {
		switch request.Kind {
		case codegen.GoNamed:
			userType := request.Attribute.Type.(expr.UserType)
			declaration := wrappers[userType]
			if declaration == nil {
				declaration = derived[userType.Origin()]
			}
			if declaration == nil {
				var err error
				declaration, err = views.Type(userType)
				if err != nil {
					return codegen.GoTypeBinding{}, err
				}
			}
			return codegen.GoTypeBinding{Owner: facts.viewsPath, Type: declaration}, nil
		case codegen.GoUnion:
			declaration, err := views.Union(request.Attribute.Type.(*expr.Union))
			if err != nil {
				return codegen.GoTypeBinding{}, err
			}
			return codegen.GoTypeBinding{Owner: facts.viewsPath, Union: declaration}, nil
		default:
			return codegen.GoTypeBinding{}, fmt.Errorf("bind unsupported view validation type %s", request.Kind)
		}
	}
	planLayout := func(attribute *expr.AttributeExpr, pointer bool) (*codegen.GoTypePlan, error) {
		policy := viewValidationPolicy()
		policy.Pointer = pointer
		return codegen.PlanGoType(attribute, codegen.GoTypePlanOptions{
			Owner:  facts.viewsPath,
			Policy: policy,
			Bind:   binder,
		})
	}
	validator := func(request codegen.ValidatorBindingRequest) (*codegen.NameDeclaration, error) {
		declaration := request.Layout.TypeDeclaration()
		retained := facts.validators[validatorKey{
			declaration: declaration,
			view:        canonicalValidatorView(request.View),
		}]
		if retained == nil {
			return nil, fmt.Errorf(
				"validator for declaration %p and view %q was not retained",
				declaration,
				request.View,
			)
		}
		return retained, nil
	}
	validatorCall := func(attribute *expr.AttributeExpr, view string, required bool) (*codegen.NameDeclaration, error) {
		layout, err := planLayout(attribute, true)
		if err != nil {
			return nil, err
		}
		request := codegen.ValidatorBindingRequest{
			Attribute: attribute,
			Layout:    layout,
			View:      view,
		}
		declaration := facts.validators[validatorKey{
			declaration: layout.TypeDeclaration(),
			view:        canonicalValidatorView(view),
		}]
		if declaration == nil && required {
			return validator(request)
		}
		return declaration, nil
	}
	for _, method := range facts.orderedMethods {
		if method.projection == nil {
			continue
		}
		for _, projected := range method.projection.types {
			projected.resultType = len(projected.views) > 0
			layout, err := planLayout(projected.pair.projectedAttribute, true)
			if err != nil {
				return err
			}
			definition, err := planLayout(projected.pair.projected.Attribute(), true)
			if err != nil {
				return err
			}
			source, err := codegen.PlanGoType(projected.pair.sourceAttribute, codegen.GoTypePlanOptions{
				Owner: facts.packagePath,
				Policy: codegen.GoLayoutPolicy{
					UseDefault: true,
					SumType:    true,
				},
				Bind: serviceGoTypeBinder(rootTypes, generation),
			})
			if err != nil {
				return err
			}
			projected.projected = layout
			projected.definition = definition
			projected.source = source
			for _, conversion := range projected.conversions {
				conversion.collection = expr.AsArray(conversion.target.Type) != nil
				conversionBinder := binder
				conversionOwner := facts.viewsPath
				if conversion.toResult {
					conversionBinder = serviceGoTypeBinder(rootTypes, generation)
					conversionOwner = facts.packagePath
				} else {
					viewBinder := conversionBinder
					targetType := conversion.target.Type.(expr.UserType)
					conversionBinder = func(request codegen.GoTypeBindingRequest) (codegen.GoTypeBinding, error) {
						if request.Kind == codegen.GoNamed && request.Attribute.Type == targetType {
							return codegen.GoTypeBinding{Owner: facts.viewsPath, Type: projected.declaration}, nil
						}
						return viewBinder(request)
					}
				}
				conversion.targetLayout, err = codegen.PlanGoType(conversion.target, codegen.GoTypePlanOptions{
					Owner: conversionOwner,
					Policy: codegen.GoLayoutPolicy{
						UseDefault: true,
						SumType:    true,
					},
					Bind: conversionBinder,
				})
				if err != nil {
					return err
				}
				context := conversion.target
				if conversion.toResult {
					context = conversion.source
				}
				conversion.contextType = context.Type.(expr.UserType)
				conversion.contextIdentity = codegen.NewProjectedTypeID(projected.pair.source)
				if projectedArray := expr.AsArray(projected.pair.projectedAttribute.Type); projectedArray != nil {
					conversion.elementType = expr.AsArray(context.Type).ElemType.Type.(expr.UserType)
					conversion.elementIdentity = codegen.NewProjectedTypeID(
						expr.AsArray(projected.pair.sourceAttribute.Type).ElemType.Type.(expr.UserType),
					)
				}
			}
			for _, validation := range projected.validations {
				if !validation.needed {
					continue
				}
				if validation.collectionElem != nil {
					declaration, err := validatorCall(validation.collectionElem, validation.viewName, true)
					if err != nil {
						return err
					}
					validation.collectionCall = declaration
					continue
				}
				layout, err := planLayout(validation.attribute, validation.pointer)
				if err != nil {
					return err
				}
				plan, err := codegen.NewValidationPlan(
					validation.attribute,
					layout,
					codegen.ValidationPlanOptions{
						Required: true,
						Alias:    validation.alias,
						Bind:     validator,
					},
				)
				if err != nil {
					return err
				}
				validation.layout = layout
				validation.plan = plan
				for _, field := range validation.fields {
					declaration, err := validatorCall(field.attribute, field.view, false)
					if err != nil {
						return err
					}
					field.call = declaration
				}
			}
		}
		viewed := method.viewedResult
		if viewed == nil {
			continue
		}
		wrapped, err := planLayout(&expr.AttributeExpr{Type: viewed.wrapped}, false)
		if err != nil {
			return err
		}
		if wrapped.TypeDeclaration() != viewed.declaration {
			return fmt.Errorf("viewed result %q layout was bound to the wrong declaration", viewed.wrapped.Name())
		}
		wrappedDef, err := planLayout(viewed.wrapped.Attribute(), false)
		if err != nil {
			return err
		}
		viewed.wrappedLayout = wrapped
		viewed.wrappedDef = wrappedDef
	}
	for _, union := range facts.viewUnions {
		if err := planUnionRenderFacts(union, binder, views); err != nil {
			return err
		}
	}
	return nil
}
