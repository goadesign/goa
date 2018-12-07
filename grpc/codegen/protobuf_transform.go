package codegen

import (
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

// Transform returns the code to transform source attribute to
// target attribute. It returns an error if source and target are not
// compatible for transformation.
func (p *protoBufTransformer) Transform(source, target codegen.AttributeAnalyzer, ta *codegen.TransformAttrs) (string, error) {
	return codegen.GoAttributeTransform(source, target, ta, p)
}

// TransformHelpers returns the transform functions required to transform
// source attribute to target attribute. It returns an error if source and
// target are incompatible.
func (p *protoBufTransformer) TransformHelpers(source, target codegen.AttributeAnalyzer) ([]*codegen.TransformFunctionData, error) {
	return codegen.GoTransformHelpers(source, target, p)
}

func (p *protoBufTransformer) MakeCompatible(source, target codegen.AttributeAnalyzer, ta *codegen.TransformAttrs, suffix string) (src, tgt codegen.AttributeAnalyzer, newTA *codegen.TransformAttrs, err error) {
	if src, tgt, newTA, err = p.GoTransformer.MakeCompatible(source, target, ta, suffix); err != nil {
		if p.proto {
			tAtt := unwrapAttr(expr.DupAtt(tgt.Attribute()))
			tgt = target.Dup(tAtt, true)
			newTA = &codegen.TransformAttrs{
				SourceVar: newTA.SourceVar,
				TargetVar: newTA.TargetVar + ".Field",
				NewVar:    false,
			}
		} else {
			sAtt := unwrapAttr(expr.DupAtt(src.Attribute()))
			src = src.Dup(sAtt, true)
			newTA = &codegen.TransformAttrs{
				SourceVar: newTA.SourceVar + ".Field",
				TargetVar: newTA.TargetVar,
				NewVar:    newTA.NewVar,
			}
		}
		if src, tgt, newTA, err = p.GoTransformer.MakeCompatible(src, tgt, newTA, ""); err != nil {
			return src, tgt, newTA, err
		}
	}
	return src, tgt, newTA, nil
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
