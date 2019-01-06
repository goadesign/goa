package codegen

import (
	"fmt"
	"strings"

	"goa.design/goa/expr"
)

type (
	// Transformer transforms a source attribute to a target attribute.
	Transformer interface {
		// TransformAttribute returns the code to transform source attribute to
		// target attribute. It returns an error if source and target are not
		// compatible for transformation.
		TransformAttribute(source, target AttributeAnalyzer, ta *TransformAttrs) (code string, err error)
		// TransformPrimitive returns the code to transform source attribute of
		// primitve type to target attribute of primitive type. It returns an error
		// if source and target are not compatible for transformation.
		TransformPrimitive(source, target AttributeAnalyzer, ta *TransformAttrs) (code string, err error)
		// TransformObject returns the code to transform source attribute of object
		// type to target attribute of object type. It returns an error if source
		// and target are not compatible for transformation.
		TransformObject(source, target AttributeAnalyzer, ta *TransformAttrs) (code string, err error)
		// TransformArray returns the code to transform source attribute of array
		// type to target attribute of array type. It returns an error if source
		// and target are not compatible for transformation.
		TransformArray(source, target AttributeAnalyzer, ta *TransformAttrs) (code string, err error)
		// TransformMap returns the code to transform source attribute of map
		// type to target attribute of map type. It returns an error if source
		// and target are not compatible for transformation.
		TransformMap(source, target AttributeAnalyzer, ta *TransformAttrs) (code string, err error)
		// TransformHelpers returns the helper functions that assist in the
		// transformation. It returns an error if source and target are not
		// compatible for transformation.
		TransformHelpers(source, target AttributeAnalyzer, seen ...map[string]*TransformFunctionData) (tfds []*TransformFunctionData, err error)
		// HelperName returns the name for the transform function to transform
		// source to the target attribute.
		HelperName(source, target AttributeAnalyzer) string
		// ConvertType adds type conversion code (if any) against varn based on
		// the attribute type.
		ConvertType(varn string, typ expr.DataType) (string, bool)
	}

	// TransformAttrs are the attributes that help in the transformation.
	TransformAttrs struct {
		// SourceVar and TargetVar are the source and target variable names used
		// in the transformation code.
		SourceVar, TargetVar string
		// NewVar is used to determine the assignment operator to initialize
		// TargetVar.
		NewVar bool
	}

	// AttributeTransformer defines the fields to transform a source attribute
	// to a target attribute.
	AttributeTransformer struct {
		// HelperPrefix is the prefix for the helper functions generated during
		// the transformation. The helper functions are named based on this
		// pattern - <HelperPrefix><SourceTypeName>To<TargetTypeName>. If no prefix
		// specified, "transform" is used as a prefix by default.
		HelperPrefix string
	}

	// TransformFunctionData describes a helper function used to transform
	// user types. These are necessary to prevent potential infinite
	// recursion when a type attribute is defined recursively. For example:
	//
	//     var Recursive = Type("Recursive", func() {
	//         Attribute("r", "Recursive")
	//     }
	//
	// Transforming this type requires generating an intermediary function:
	//
	//     func recursiveToRecursive(r *Recursive) *service.Recursive {
	//         var t service.Recursive
	//         if r.R != nil {
	//             t.R = recursiveToRecursive(r.R)
	//         }
	//    }
	//
	TransformFunctionData struct {
		Name          string
		ParamTypeRef  string
		ResultTypeRef string
		Code          string
	}
)

// Transform transforms source attribute to target attribute with the given
// transformer and returns the transformation code and the helper functions
// used in the transformation. It returns an error if source and target
// attributes are not compatible for transformation.
//
// source, target are the source and target attributes used in transformation
//
// sourceVar and targetVar are the variable names used in the transformation
//
// t is the Transformer
//
func Transform(source, target AttributeAnalyzer, sourceVar, targetVar string, t Transformer) (string, []*TransformFunctionData, error) {
	code, err := t.TransformAttribute(source, target, &TransformAttrs{SourceVar: sourceVar, TargetVar: targetVar, NewVar: true})
	if err != nil {
		return "", nil, err
	}

	funcs, err := t.TransformHelpers(source, target)
	if err != nil {
		return "", nil, err
	}

	return strings.TrimRight(code, "\n"), funcs, nil
}

// IsCompatible returns an error if a and b are not both objects, both arrays,
// both maps or both the same primitive type. actx and bctx are used to build
// the error message if any.
func IsCompatible(a, b expr.DataType, actx, bctx string) error {
	switch {
	case expr.IsObject(a):
		if !expr.IsObject(b) {
			return fmt.Errorf("%s is an object but %s type is %s", actx, bctx, b.Name())
		}
	case expr.IsArray(a):
		if !expr.IsArray(b) {
			return fmt.Errorf("%s is an array but %s type is %s", actx, bctx, b.Name())
		}
	case expr.IsMap(a):
		if !expr.IsMap(b) {
			return fmt.Errorf("%s is a hash but %s type is %s", actx, bctx, b.Name())
		}
	default:
		if a.Kind() != b.Kind() {
			return fmt.Errorf("%s is a %s but %s type is %s", actx, a.Name(), bctx, b.Name())
		}
	}
	return nil
}

// AppendHelpers takes care of only appending helper functions from newH that
// are not already in oldH.
func AppendHelpers(oldH, newH []*TransformFunctionData) []*TransformFunctionData {
	for _, h := range newH {
		found := false
		for _, h2 := range oldH {
			if h.Name == h2.Name {
				found = true
				break
			}
		}
		if !found {
			oldH = append(oldH, h)
		}
	}
	return oldH
}

// HelperName returns the name for the transform function.
func (t *AttributeTransformer) HelperName(source, target AttributeAnalyzer) string {
	var (
		sname  string
		tname  string
		prefix string
	)
	{
		sname = Goify(source.Name(true), true)
		tname = Goify(target.Name(true), true)
		prefix = t.HelperPrefix
		if prefix == "" {
			prefix = "transform"
		}
	}
	return Goify(prefix+sname+"To"+tname, false)
}

// ConvertType converts varn to type typ.
func (t *AttributeTransformer) ConvertType(varn string, typ expr.DataType) (string, bool) {
	return varn, false
}
