// This file selects nested result fields before Goa plans service-to-view and
// view-to-service conversions. The same conversion engine handles objects,
// collections, and unions; fields outside the selected view are never copied.
package service

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

type (
	// viewTransformType identifies one retained result type and selected view.
	viewTransformType struct {
		// origin is the root result declaration shared by all copies of the type.
		origin expr.UserType
		// view is the result view selected for this conversion step.
		view string
	}
)

// viewTransformHooks narrows the view side of each conversion while keeping
// its original Go declaration. Reusing each type and view also keeps recursive
// result references finite when the transform planner follows their fields.
func viewTransformHooks(toResult bool) *codegen.TransformHooks {
	selected := make(map[viewTransformType]*expr.ResultTypeExpr)
	return &codegen.TransformHooks{
		UnwrapPair: func(source, target *expr.AttributeExpr) (*expr.AttributeExpr, *expr.AttributeExpr, *codegen.WrapDirective) {
			if toResult {
				source = selectTransformView(source, selected)
			} else {
				target = selectTransformView(target, selected)
			}
			return source, target, nil
		},
	}
}

// selectTransformView copies only a result's selected fields and required
// checks. Its declaration identity stays unchanged so generated helpers use
// the already-planned service or view type rather than another public type.
func selectTransformView(attribute *expr.AttributeExpr, selected map[viewTransformType]*expr.ResultTypeExpr) *expr.AttributeExpr {
	result, ok := attribute.Type.(*expr.ResultTypeExpr)
	if !ok {
		return attribute
	}
	view := expr.DefaultView
	if explicit, ok := attribute.Meta.Last(expr.ViewMetaKey); ok {
		view = explicit
	}
	key := viewTransformType{origin: result.Origin(), view: view}
	if existing := selected[key]; existing != nil {
		copied := expr.DupAtt(attribute)
		copied.Type = existing
		return copied
	}
	copied := expr.NewAttributeGraphCopier().Copy(attribute)
	narrowed := copied.Type.(*expr.ResultTypeExpr)
	selected[key] = narrowed
	if array := expr.AsArray(narrowed); array != nil {
		applyCascadedView(array.ElemType, view)
		return copied
	}
	object := &expr.Object{}
	full := expr.AsObject(narrowed)
	for _, field := range *full {
		viewField := expr.AsObject(result.View(view).Type).Attribute(field.Name)
		if viewField == nil {
			continue
		}
		value := field.Attribute
		if explicit, ok := viewField.Meta.Last(expr.ViewMetaKey); ok {
			value = propagateSelectedView(value, explicit)
		}
		object.Set(field.Name, value)
	}
	narrowed.Type = object
	if narrowed.Validation != nil {
		required := make([]string, 0, len(narrowed.Validation.Required))
		for _, name := range narrowed.Validation.Required {
			if object.Attribute(name) != nil {
				required = append(required, name)
			}
		}
		narrowed.Validation.Required = required
	}
	return copied
}

// propagateSelectedView returns a detached copy of attribute where view is
// recorded on every result type reached directly or through anonymous array
// and map containers.
func propagateSelectedView(attribute *expr.AttributeExpr, view string) *expr.AttributeExpr {
	copied := expr.NewAttributeGraphCopier().Copy(attribute)
	applyCascadedView(copied, view)
	return copied
}

// applyCascadedView attaches view to result-type occurrences (including
// CollectionOf result collections) reached directly or through anonymous array
// and map containers. Named non-result user types, objects, and unions own
// their own field definitions and stop the cascade, matching view validation.
func applyCascadedView(attribute *expr.AttributeExpr, view string) {
	if attribute == nil || attribute.Type == expr.Empty {
		return
	}
	switch actual := attribute.Type.(type) {
	case *expr.ResultTypeExpr:
		attribute.AddMeta(expr.ViewMetaKey, view)
	case *expr.Array:
		applyCascadedView(actual.ElemType, view)
	case *expr.Map:
		applyCascadedView(actual.KeyType, view)
		applyCascadedView(actual.ElemType, view)
	}
}
