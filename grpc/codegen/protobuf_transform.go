package codegen

import (
	"bytes"
	"fmt"

	"goa.design/goa/codegen"
	"goa.design/goa/expr"
)

// protoBufTransformer implements the codegen.Transformer interface
// to transform Go types to protocol buffer generated Go types.
type protoBufTransformer struct {
	*codegen.GoTransformer
	// proto if true indicates target type is a protocol buffer type.
	proto bool
	// targetInit is the initialization code for the target type for nested
	// map and array types.
	targetInit string
}

// protoBufTransform transforms Go type to protocol buffer Go type and
// vice versa.
//
// source, target are the source and target attributes used in transformation
//
// `proto` param if true indicates that the target is a protocol buffer type
//
func protoBufTransform(source, target codegen.AttributeAnalyzer, sourceVar, targetVar string, proto bool) (string, []*codegen.TransformFunctionData, error) {
	var prefix string
	{
		prefix = "protobuf"
		if proto {
			prefix = "svc"
		}
	}
	p := &protoBufTransformer{
		GoTransformer: codegen.NewGoTransformer(prefix).(*codegen.GoTransformer),
		proto:         proto,
	}
	return codegen.Transform(source, target, sourceVar, targetVar, p)
}

// TransformAttribute returns the code to transform source attribute to
// target attribute. It returns an error if source and target are not
// compatible for transformation.
func (p *protoBufTransformer) TransformAttribute(source, target codegen.AttributeAnalyzer, ta *codegen.TransformAttrs) (string, error) {
	var (
		code string
		err  error

		sourceType = source.Attribute().Type
		targetType = target.Attribute().Type
	)
	{
		switch {
		case expr.IsArray(sourceType):
			code, err = p.TransformArray(source, target, ta)
		case expr.IsMap(sourceType):
			code, err = p.TransformMap(source, target, ta)
		case expr.IsObject(sourceType):
			if expr.IsPrimitive(targetType) {
				code, err = p.TransformPrimitive(source, target, ta)
			} else {
				code, err = p.TransformObject(source, target, ta)
			}
		default:
			code, err = p.TransformPrimitive(source, target, ta)
		}
	}
	if err != nil {
		return "", err
	}
	return code, nil
}

// TransformPrimitive returns the code to transform source attribute of
// primitive type to target attribute of primitive type. It returns an error
// if source and target are not compatible for transformation.
func (p *protoBufTransformer) TransformPrimitive(source, target codegen.AttributeAnalyzer, ta *codegen.TransformAttrs) (string, error) {
	var code string
	srcAtt := source.Attribute()
	tgtAtt := target.Attribute()
	if err := codegen.IsCompatible(srcAtt.Type, tgtAtt.Type, ta.SourceVar, ta.TargetVar); err != nil {
		if p.proto {
			code += fmt.Sprintf("%s := &%s{}\n", ta.TargetVar, target.Name(true))
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
	}
	assign := "="
	if ta.NewVar {
		assign = ":="
	}
	srcField, _ := p.ConvertType(ta.SourceVar, srcAtt.Type)
	code += fmt.Sprintf("%s %s %s\n", ta.TargetVar, assign, srcField)
	return code, nil
}

// TransformObject returns the code to transform source attribute of object
// type to target attribute of object type. It returns an error if source
// and target are not compatible for transformation.
func (p *protoBufTransformer) TransformObject(source, target codegen.AttributeAnalyzer, ta *codegen.TransformAttrs) (string, error) {
	return codegen.GoObjectTransform(source, target, ta, p)
}

// TransformArray returns the code to transform source attribute of array
// type to target attribute of array type. It returns an error if source
// and target are not compatible for transformation.
func (p *protoBufTransformer) TransformArray(source, target codegen.AttributeAnalyzer, ta *codegen.TransformAttrs) (string, error) {
	sourceArr := expr.AsArray(source.Attribute().Type)
	if sourceArr == nil {
		return "", fmt.Errorf("source is not an array type: received %T", source.Attribute().Type)
	}
	targetArr := expr.AsArray(target.Attribute().Type)
	if targetArr == nil {
		return "", fmt.Errorf("target is not an array type: received %T", target.Attribute().Type)
	}

	source = source.Dup(sourceArr.ElemType, true)
	target = target.Dup(targetArr.ElemType, true)
	targetRef := target.Ref(true)

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
	}
	if err := codegen.IsCompatible(source.Attribute().Type, target.Attribute().Type, ta.SourceVar+"[0]", ta.TargetVar+"[0]"); err != nil {
		if p.proto {
			p.targetInit = target.Name(true)
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
		if err := codegen.IsCompatible(source.Attribute().Type, target.Attribute().Type, ta.SourceVar+"[0]", ta.TargetVar+"[0]"); err != nil {
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
	var buf bytes.Buffer
	if err := codegen.TransformGoArrayT.Execute(&buf, data); err != nil {
		return "", err
	}
	code += buf.String()
	return code, nil
}

// TransformMap returns the code to transform source attribute of map
// type to target attribute of map type. It returns an error if source
// and target are not compatible for transformation.
func (p *protoBufTransformer) TransformMap(source, target codegen.AttributeAnalyzer, ta *codegen.TransformAttrs) (string, error) {
	sourceType := source.Attribute().Type
	targetType := target.Attribute().Type
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
	if err := codegen.IsCompatible(sourceKey.Attribute().Type, targetKey.Attribute().Type, ta.SourceVar+"[key]", ta.TargetVar+"[key]"); err != nil {
		return "", err
	}
	sourceElem := source.Dup(sourceMap.ElemType, true)
	targetElem := target.Dup(targetMap.ElemType, true)
	targetElemRef := targetElem.Ref(true)

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
	}
	if err := codegen.IsCompatible(sourceMap.ElemType.Type, targetMap.ElemType.Type, ta.SourceVar+"[*]", ta.TargetVar+"[*]"); err != nil {
		if p.proto {
			p.targetInit = targetElem.Name(true)
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
		if err := codegen.IsCompatible(sourceElem.Attribute().Type, targetElem.Attribute().Type, ta.SourceVar+"[*]", ta.TargetVar+"[*]"); err != nil {
			return "", err
		}
	}
	data := map[string]interface{}{
		"Transformer": p,
		"KeyTypeRef":  targetKey.Ref(true),
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
	var buf bytes.Buffer
	if err := codegen.TransformGoMapT.Execute(&buf, data); err != nil {
		return "", err
	}
	code += buf.String()
	return code, nil
}

// MakeCompatible checks whether source and target attributes are
// compatible for transformation and returns an error if not. If no error
// is returned, it returns the source and target attributes that are
// compatible.
func (p *protoBufTransformer) MakeCompatible(source, target codegen.AttributeAnalyzer, ta *codegen.TransformAttrs, suffix string) (src, tgt codegen.AttributeAnalyzer, newTA *codegen.TransformAttrs, err error) {
	src = source
	tgt = target
	if err = codegen.IsCompatible(src.Attribute().Type, tgt.Attribute().Type, ta.SourceVar+suffix, ta.TargetVar+suffix); err != nil {
		if p.proto {
			tgtAtt := unwrapAttr(expr.DupAtt(target.Attribute()))
			tgt = target.Dup(tgtAtt, true)
		} else {
			srcAtt := unwrapAttr(expr.DupAtt(source.Attribute()))
			src = source.Dup(srcAtt, true)
		}
		if err = codegen.IsCompatible(src.Attribute().Type, tgt.Attribute().Type, ta.SourceVar, ta.TargetVar); err != nil {
			return src, tgt, ta, err
		}
	}
	return src, tgt, ta, nil
}

// TransformHelpers returns the transform functions required to transform
// source attribute to target attribute. It returns an error if source and
// target are incompatible.
func (p *protoBufTransformer) TransformHelpers(source, target codegen.AttributeAnalyzer, seen ...map[string]*codegen.TransformFunctionData) ([]*codegen.TransformFunctionData, error) {
	return codegen.GoTransformHelpers(source, target, p, seen...)
}

// ConvertType converts varn to type typ.
// NOTE: For Int and UInt kinds, protocol buffer Go compiler generates
// int32 and uint32 respectively whereas goa v2 generates int and uint.
func (p *protoBufTransformer) ConvertType(varn string, typ expr.DataType) (string, bool) {
	if typ.Kind() != expr.IntKind && typ.Kind() != expr.UIntKind {
		return varn, false
	}

	if p.proto {
		return fmt.Sprintf("%s(%s)", protoBufNativeGoTypeName(typ), varn), true
	}
	return fmt.Sprintf("%s(%s)", codegen.GoNativeTypeName(typ), varn), true
}
