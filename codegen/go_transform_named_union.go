// Default union conversions keep each variable's declared Go type while using
// the underlying union's methods and package for branch references.
package codegen

import (
	"bytes"
	"fmt"
	"strconv"

	"goa.design/goa/v3/expr"
)

// transformUnion converts one selected branch. Pointer arguments describe the
// stored source and destination; newVar only selects whether to declare the
// destination. Object field callers supply their existing value temporary.
func transformUnion(source, target *expr.AttributeExpr, sourceVar, targetVar string, newVar, sourcePointer, targetPointer bool, ta *TransformAttrs) (string, error) {
	if !expr.IsUnion(target.Type) {
		return "", fmt.Errorf("cannot transform union %s to non-union %s", source.Type.Name(), target.Type.Name())
	}
	srcUnion, tgtUnion := expr.AsUnion(source.Type), expr.AsUnion(target.Type)
	if len(srcUnion.Values) != len(tgtUnion.Values) {
		return "", fmt.Errorf("cannot transform union: number of union types differ (%s has %d, %s has %d)",
			source.Type.Name(), len(srcUnion.Values), target.Type.Name(), len(tgtUnion.Values))
	}
	for i, st := range srcUnion.Values {
		if err := IsCompatible(st.Attribute.Type, tgtUnion.Values[i].Attribute.Type, sourceVar, targetVar); err != nil {
			return "", fmt.Errorf("cannot transform union %s to %s: type at index %d: %w",
				source.Type.Name(), target.Type.Name(), i, err)
		}
	}

	// A named definition can change packages several times before reaching the
	// union. Its unlocated branches belong to that union's package.
	_, sourceContext := unionDefinition(source, ta.SourceCtx)
	_, targetContext := unionDefinition(target, ta.TargetCtx)
	sourceReceiver := unionMethodReceiver(source, ta.SourceCtx, sourceVar, sourcePointer)
	targetReceiver := "u"
	if _, named := target.Type.(expr.UserType); named {
		targetReceiver = unionMethodReceiver(target, ta.TargetCtx, "&u", true)
	}
	unionPkg := ta.TargetCtx.Pkg(target)
	valueType := ta.TargetCtx.Scope.Name(target, unionPkg, false, ta.TargetCtx.UseDefault)
	typeRef := valueType
	if targetPointer {
		typeRef = ta.TargetCtx.Scope.Ref(target, unionPkg)
	}

	tempVarName := "obj"
	if ta.unionDepth > 0 {
		tempVarName = "tmp" + strconv.Itoa(ta.unionDepth+1)
	}
	childAttrs := *ta
	childAttrs.unionDepth++
	cases := make([]map[string]any, 0, len(srcUnion.Values))
	for i, st := range srcUnion.Values {
		tt := tgtUnion.Values[i]
		branchAttrs := childAttrs
		branchAttrs.SourceCtx = sourceContext.Enter(st.Attribute)
		branchAttrs.TargetCtx = targetContext.Enter(tt.Attribute)
		cases = append(cases, map[string]any{
			"CaseName":        st.Name,
			"SourceFieldName": Goify(st.Name, true),
			"TargetFieldName": Goify(tt.Name, true),
			"SourceAttr":      st.Attribute,
			"TargetAttr":      tt.Attribute,
			"TargetCastType":  branchAttrs.TargetCtx.Scope.Ref(tt.Attribute, branchAttrs.TargetCtx.Pkg(tt.Attribute)),
			"SourceNilable":   IsNilable(st.Attribute.Type) || expr.IsUnion(st.Attribute.Type),
			"UseHelper":       usesTransformHelper(st.Attribute, tt.Attribute, ta.Hooks),
			"TransformAttrs":  &branchAttrs,
		})
	}
	data := map[string]any{
		"SourceVar":      sourceReceiver,
		"TargetVar":      targetVar,
		"TargetReceiver": targetReceiver,
		"TargetPointer":  targetPointer,
		"NewVar":         newVar,
		"TypeRef":        typeRef,
		"ValueTypeRef":   valueType,
		"TempVarName":    tempVarName,
		"Cases":          cases,
	}
	var buf bytes.Buffer
	if err := transformGoUnionT.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// unionDefinition follows only retained named definitions and enters each
// owner before returning the union that declares the methods and branch types.
func unionDefinition(attribute *expr.AttributeExpr, context *AttributeContext) (*expr.AttributeExpr, *AttributeContext) {
	for {
		context = context.Enter(attribute)
		named, ok := attribute.Type.(expr.UserType)
		if !ok {
			return attribute, context
		}
		attribute = named.Attribute()
	}
}

// unionMethodReceiver converts a named value or pointer to the exact type that
// owns its union methods. Direct union values already have those methods.
func unionMethodReceiver(attribute *expr.AttributeExpr, context *AttributeContext, value string, pointer bool) string {
	if _, named := attribute.Type.(expr.UserType); !named {
		return value
	}
	definition, owner := unionDefinition(attribute, context)
	var ref string
	if pointer {
		ref = owner.Scope.Ref(definition, owner.Pkg(definition))
	} else {
		ref = owner.Scope.Name(definition, owner.Pkg(definition), false, owner.UseDefault)
	}
	return fmt.Sprintf("(%s)(%s)", ref, value)
}
