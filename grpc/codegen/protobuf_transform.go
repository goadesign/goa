package codegen

import (
	"fmt"
	"strings"

	"goa.design/goa/codegen"
	"goa.design/goa/expr"
)

// protoBufTransform produces Go code to initialize a data structure defined
// by target from an instance of data structure defined by source. The source
// or target is a protocol buffer type.
//
// source, target are the source and target attributes used in transformation
//
// `proto` param if true indicates that the target is a protocol buffer type
//
func protoBufTransform(source, target *codegen.ContextualAttribute, sourceVar, targetVar string, proto bool) (string, []*codegen.TransformFunctionData, error) {
	var prefix string
	{
		prefix = "protobuf"
		if proto {
			prefix = "svc"
		}
	}
	p := &protoBufTransformer{
		helperPrefix: prefix,
		proto:        proto,
	}

	code, err := p.Transform(source, target, &codegen.TransformAttrs{SourceVar: sourceVar, TargetVar: targetVar, NewVar: true})
	if err != nil {
		return "", nil, err
	}

	funcs, err := codegen.GoTransformHelpers(source, target, p, prefix)
	if err != nil {
		return "", nil, err
	}

	return strings.TrimRight(code, "\n"), funcs, nil
}

// protoBufTransformer implements the codegen.Transformer interface
// to transform Go types to protocol buffer generated Go types.
type protoBufTransformer struct {
	// helperPrefix is the prefix for the helper function names.
	helperPrefix string
	// proto if true indicates target type is a protocol buffer type.
	proto bool
	// targetInit is the initialization code for the target type for nested
	// map and array types.
	targetInit string
}

// Transform returns the code to initialize a target data structure from an
// instance of source data structure. It returns an error if source and target
// are not compatible for transformation (different types, fields of
// different type).
func (p *protoBufTransformer) Transform(source, target *codegen.ContextualAttribute, ta *codegen.TransformAttrs) (string, error) {
	var (
		initCode string
		err      error

		srcAtt = source.Attribute.Expr()
		tgtAtt = target.Attribute.Expr()
	)
	if err := codegen.IsCompatible(srcAtt.Type, tgtAtt.Type, ta.SourceVar, ta.TargetVar); err != nil {
		if p.proto {
			initCode += fmt.Sprintf("%s := &%s{}\n", ta.TargetVar, target.Attribute.Name())
			ta.TargetVar += ".Field"
			ta.NewVar = false
			tgtAtt = unwrapAttr(expr.DupAtt(tgtAtt))
		} else {
			srcAtt = unwrapAttr(expr.DupAtt(srcAtt))
			ta.SourceVar += ".Field"
		}
		if err = codegen.IsCompatible(srcAtt.Type, tgtAtt.Type, ta.SourceVar, ta.TargetVar); err != nil {
			return "", err
		}
		source = source.Dup(srcAtt, true)
		target = target.Dup(tgtAtt, true)
	}

	var (
		code string

		sourceType = source.Attribute.Expr().Type
	)
	{
		switch {
		case expr.IsArray(sourceType):
			code, err = p.TransformArray(source, target, ta)
		case expr.IsMap(sourceType):
			code, err = p.TransformMap(source, target, ta)
		case expr.IsObject(sourceType):
			code, err = p.TransformObject(source, target, ta)
		default:
			assign := "="
			if ta.NewVar {
				assign = ":="
			}
			srcField := p.ConvertType(source.Attribute, target.Attribute, ta.SourceVar)
			code = fmt.Sprintf("%s %s %s\n", ta.TargetVar, assign, srcField)
		}
	}
	if err != nil {
		return "", err
	}
	return initCode + code, nil
}

// MakeCompatible checks whether source and target attributes are
// compatible for transformation and returns an error if not. If no error
// is returned, it returns the source and target attributes that are
// compatible.
func (p *protoBufTransformer) MakeCompatible(source, target *codegen.ContextualAttribute, ta *codegen.TransformAttrs, suffix string) (src, tgt *codegen.ContextualAttribute, newTA *codegen.TransformAttrs, err error) {
	src = source
	tgt = target
	newTA = &codegen.TransformAttrs{
		SourceVar: ta.SourceVar,
		TargetVar: ta.TargetVar,
		NewVar:    ta.NewVar,
	}
	if err = codegen.IsCompatible(
		src.Attribute.Expr().Type,
		tgt.Attribute.Expr().Type,
		ta.SourceVar+suffix, ta.TargetVar+suffix); err != nil {
		if p.proto {
			p.targetInit = target.Attribute.Name()
			tgtAtt := unwrapAttr(expr.DupAtt(target.Attribute.Expr()))
			tgt = target.Dup(tgtAtt, true)
		} else {
			srcAtt := unwrapAttr(expr.DupAtt(source.Attribute.Expr()))
			src = source.Dup(srcAtt, true)
			newTA.SourceVar += ".Field"
		}
		if err = codegen.IsCompatible(
			src.Attribute.Expr().Type,
			tgt.Attribute.Expr().Type,
			newTA.SourceVar, newTA.TargetVar); err != nil {
			return src, tgt, newTA, err
		}
	}
	return src, tgt, newTA, nil
}

// ConvertType produces code to initialize a target type from a source type
// held by sourceVar.
// NOTE: For Int and UInt kinds, protocol buffer Go compiler generates
// int32 and uint32 respectively whereas goa v2 generates int and uint.
func (p *protoBufTransformer) ConvertType(source, target codegen.Attributor, sourceVar string) string {
	typ := source.Expr().Type
	if _, ok := typ.(expr.UserType); ok {
		// return a function name for the conversion
		return fmt.Sprintf("%s(%s)", codegen.HelperName(source, target, p.helperPrefix), sourceVar)
	}

	if typ.Kind() != expr.IntKind && typ.Kind() != expr.UIntKind {
		return sourceVar
	}
	if p.proto {
		return fmt.Sprintf("%s(%s)", protoBufNativeGoTypeName(typ), sourceVar)
	}
	return fmt.Sprintf("%s(%s)", codegen.GoNativeTypeName(typ), sourceVar)
}

// transformObject returns the code to transform source attribute of object
// type to target attribute of object type. It returns an error if source
// and target are not compatible for transformation.
func (p *protoBufTransformer) TransformObject(source, target *codegen.ContextualAttribute, ta *codegen.TransformAttrs) (string, error) {
	return codegen.GoObjectTransform(source, target, ta, p)
}

// transformArray returns the code to transform source attribute of array
// type to target attribute of array type. It returns an error if source
// and target are not compatible for transformation.
func (p *protoBufTransformer) TransformArray(source, target *codegen.ContextualAttribute, ta *codegen.TransformAttrs) (string, error) {
	sourceArr := expr.AsArray(source.Attribute.Expr().Type)
	if sourceArr == nil {
		return "", fmt.Errorf("source is not an array type: received %T", source.Attribute.Expr().Type)
	}
	targetArr := expr.AsArray(target.Attribute.Expr().Type)
	if targetArr == nil {
		return "", fmt.Errorf("target is not an array type: received %T", target.Attribute.Expr().Type)
	}

	source = source.Dup(sourceArr.ElemType, true)
	target = target.Dup(targetArr.ElemType, true)
	targetRef := target.Attribute.Ref()

	var code string

	// If targetInit is set, the target array element is in a nested state.
	// See grpc/docs/FAQ.md.
	if p.targetInit != "" {
		assign := "="
		if ta.NewVar {
			assign = ":="
		}
		code = fmt.Sprintf("%s %s &%s{}\n", ta.TargetVar, assign, p.targetInit)
		ta = &codegen.TransformAttrs{
			SourceVar: ta.SourceVar,
			TargetVar: ta.TargetVar + ".Field",
			NewVar:    false,
		}
		p.targetInit = ""
	}
	if err := codegen.IsCompatible(source.Attribute.Expr().Type, target.Attribute.Expr().Type, ta.SourceVar+"[0]", ta.TargetVar+"[0]"); err != nil {
		if p.proto {
			p.targetInit = target.Attribute.Name()
			tAtt := unwrapAttr(expr.DupAtt(targetArr.ElemType))
			target = target.Dup(tAtt, true)
		} else {
			sAtt := unwrapAttr(expr.DupAtt(sourceArr.ElemType))
			source = source.Dup(sAtt, true)
			ta = &codegen.TransformAttrs{
				SourceVar: ta.SourceVar + ".Field",
				TargetVar: ta.TargetVar,
				NewVar:    ta.NewVar,
			}
		}
		if err := codegen.IsCompatible(source.Attribute.Expr().Type, target.Attribute.Expr().Type, ta.SourceVar+"[0]", ta.TargetVar+"[0]"); err != nil {
			return "", err
		}
	}

	data := map[string]interface{}{
		"Transformer": p,
		"ElemTypeRef": targetRef,
		"SourceElem":  source,
		"TargetElem":  target,
		"SourceVar":   ta.SourceVar,
		"TargetVar":   ta.TargetVar,
		"NewVar":      ta.NewVar,
	}
	c, err := codegen.RunGoArrayTemplate(data)
	if err != nil {
		return "", err
	}
	return code + c, nil
}

// transformMap returns the code to transform source attribute of map
// type to target attribute of map type. It returns an error if source
// and target are not compatible for transformation.
func (p *protoBufTransformer) TransformMap(source, target *codegen.ContextualAttribute, ta *codegen.TransformAttrs) (string, error) {
	sourceType := source.Attribute.Expr().Type
	targetType := target.Attribute.Expr().Type
	sourceMap := expr.AsMap(sourceType)
	if sourceMap == nil {
		return "", fmt.Errorf("source is not a map type: received %T", sourceType)
	}
	targetMap := expr.AsMap(targetType)
	if targetMap == nil {
		return "", fmt.Errorf("target is not a map type: received %T", targetType)
	}

	// Target map key cannot be nested in protocol buffers. So no need to worry
	// about unwrapping.
	sourceKey := source.Dup(sourceMap.KeyType, true)
	targetKey := target.Dup(targetMap.KeyType, true)
	if err := codegen.IsCompatible(sourceKey.Attribute.Expr().Type, targetKey.Attribute.Expr().Type, ta.SourceVar+"[key]", ta.TargetVar+"[key]"); err != nil {
		return "", err
	}
	sourceElem := source.Dup(sourceMap.ElemType, true)
	targetElem := target.Dup(targetMap.ElemType, true)
	targetElemRef := targetElem.Attribute.Ref()

	var code string

	// If targetInit is set, the target map element is in a nested state.
	// See grpc/docs/FAQ.md.
	if p.targetInit != "" {
		assign := "="
		if ta.NewVar {
			assign = ":="
		}
		code = fmt.Sprintf("%s %s &%s{}\n", ta.TargetVar, assign, p.targetInit)
		ta = &codegen.TransformAttrs{
			SourceVar: ta.SourceVar,
			TargetVar: ta.TargetVar + ".Field",
			NewVar:    false,
		}
		p.targetInit = ""
	}
	if err := codegen.IsCompatible(sourceMap.ElemType.Type, targetMap.ElemType.Type, ta.SourceVar+"[*]", ta.TargetVar+"[*]"); err != nil {
		if p.proto {
			p.targetInit = targetElem.Attribute.Name()
			tAtt := unwrapAttr(expr.DupAtt(targetMap.ElemType))
			targetElem = target.Dup(tAtt, true)
		} else {
			sAtt := unwrapAttr(expr.DupAtt(sourceMap.ElemType))
			sourceElem = source.Dup(sAtt, true)
			ta = &codegen.TransformAttrs{
				SourceVar: ta.SourceVar + ".Field",
				TargetVar: ta.TargetVar,
				NewVar:    ta.NewVar,
			}
		}
		if err := codegen.IsCompatible(sourceElem.Attribute.Expr().Type, targetElem.Attribute.Expr().Type, ta.SourceVar+"[*]", ta.TargetVar+"[*]"); err != nil {
			return "", err
		}
	}
	data := map[string]interface{}{
		"Transformer": p,
		"KeyTypeRef":  targetKey.Attribute.Ref(),
		"ElemTypeRef": targetElemRef,
		"SourceKey":   sourceKey,
		"TargetKey":   targetKey,
		"SourceElem":  sourceElem,
		"TargetElem":  targetElem,
		"SourceVar":   ta.SourceVar,
		"TargetVar":   ta.TargetVar,
		"NewVar":      ta.NewVar,
		"TargetMap":   targetMap,
	}
	c, err := codegen.RunGoMapTemplate(data)
	if err != nil {
		return "", err
	}
	return code + c, nil
}
