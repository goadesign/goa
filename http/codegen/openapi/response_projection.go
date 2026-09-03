// This file selects a response view on a private result type copy and keeps
// the component names produced by released Goa versions.
package openapi

import (
	"fmt"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// ResponseProjection contains one detached response result and the types whose
// released names should be preferred when they each describe one schema.
type ResponseProjection struct {
	Result    *expr.ResultTypeExpr
	Preferred []expr.UserType
}

// ProjectResponseResult selects view without changing the design. Collection
// result names keep the Response suffix before the view name.
func ProjectResponseResult(result *expr.ResultTypeExpr, view string) ResponseProjection {
	projected, err := expr.Project(result, view)
	if err != nil {
		panic(fmt.Sprintf("failed to project result type %q to view %q: %s", result.Identifier, view, err))
	}
	copy := expr.DupAtt(&expr.AttributeExpr{Type: projected}).Type.(*expr.ResultTypeExpr)
	originalArray := expr.AsArray(result.Type)
	projectedArray := expr.AsArray(copy.Type)
	if originalArray == nil || projectedArray == nil {
		return ResponseProjection{Result: copy}
	}
	originalElement, originalNamed := originalArray.ElemType.Type.(*expr.ResultTypeExpr)
	projectedElement, projectedNamed := projectedArray.ElemType.Type.(*expr.ResultTypeExpr)
	if !originalNamed || !projectedNamed {
		return ResponseProjection{Result: copy}
	}
	name := codegen.Goify(originalElement.Name(), true) + "Response"
	if view != "" && view != expr.DefaultView {
		name += codegen.Goify(view, true)
	}
	projectedElement.Rename(name)
	copy.Rename(name + "Collection")
	return ResponseProjection{
		Result:    copy,
		Preferred: []expr.UserType{copy, projectedElement},
	}
}
