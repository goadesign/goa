package codegen

import (
	"testing"

	"goa.design/goa/codegen"
	ctestdata "goa.design/goa/codegen/testdata"
	"goa.design/goa/expr"
)

func TestProtoBufTransform(t *testing.T) {
	root := codegen.RunDSL(t, ctestdata.TestTypesDSL)
	var (
		scope = codegen.NewNameScope()

		// types to test
		primitive = expr.Int

		simple   = root.UserType("Simple")
		required = root.UserType("Required")
		defaultT = root.UserType("Default")

		simpleMap = root.UserType("SimpleMap")
		nestedMap = root.UserType("NestedMap")
		arrayMap  = root.UserType("ArrayMap")

		simpleArray = root.UserType("SimpleArray")
		nestedArray = root.UserType("NestedArray")
		mapArray    = root.UserType("MapArray")
		typeArray   = root.UserType("TypeArray")

		recursive   = root.UserType("Recursive")
		composite   = root.UserType("Composite")
		customField = root.UserType("CompositeWithCustomField")

		resultType = root.UserType("ResultType")
		rtCol      = root.UserType("ResultTypeCollection")

		// attribute analyzers used in test cases
		primitiveGoa   = newGoaContextAttr(primitive, "", scope)
		simpleGoa      = newGoaContextAttr(simple, "", scope)
		requiredGoa    = newGoaContextAttr(required, "", scope)
		defaultTGoa    = newGoaContextAttr(defaultT, "", scope)
		simpleMapGoa   = newGoaContextAttr(simpleMap, "", scope)
		nestedMapGoa   = newGoaContextAttr(nestedMap, "", scope)
		arrayMapGoa    = newGoaContextAttr(arrayMap, "", scope)
		simpleArrayGoa = newGoaContextAttr(simpleArray, "", scope)
		nestedArrayGoa = newGoaContextAttr(nestedArray, "", scope)
		mapArrayGoa    = newGoaContextAttr(mapArray, "", scope)
		typeArrayGoa   = newGoaContextAttr(typeArray, "", scope)
		recursiveGoa   = newGoaContextAttr(recursive, "", scope)
		compositeGoa   = newGoaContextAttr(composite, "", scope)
		customFieldGoa = newGoaContextAttr(customField, "", scope)
		resultTypeGoa  = newGoaContextAttr(resultType, "", scope)
		rtColGoa       = newGoaContextAttr(rtCol, "", scope)
		requiredPtrGoa = pointerContext(required, "", scope)

		primitiveProto   = newProtoContextAttr(primitive, "", scope)
		simpleProto      = newProtoContextAttr(simple, "", scope)
		requiredProto    = newProtoContextAttr(required, "", scope)
		defaultTProto    = newProtoContextAttr(defaultT, "", scope)
		simpleMapProto   = newProtoContextAttr(simpleMap, "", scope)
		nestedMapProto   = newProtoContextAttr(nestedMap, "", scope)
		arrayMapProto    = newProtoContextAttr(arrayMap, "", scope)
		simpleArrayProto = newProtoContextAttr(simpleArray, "", scope)
		nestedArrayProto = newProtoContextAttr(nestedArray, "", scope)
		mapArrayProto    = newProtoContextAttr(mapArray, "", scope)
		typeArrayProto   = newProtoContextAttr(typeArray, "", scope)
		recursiveProto   = newProtoContextAttr(recursive, "", scope)
		compositeProto   = newProtoContextAttr(composite, "", scope)
		customFieldProto = newProtoContextAttr(customField, "", scope)
		resultTypeProto  = newProtoContextAttr(resultType, "", scope)
		rtColProto       = newProtoContextAttr(rtCol, "", scope)
	)

	tc := map[string][]struct {
		Name    string
		Source  *codegen.ContextualAttribute
		Target  *codegen.ContextualAttribute
		ToProto bool
		Code    string
	}{
		// test cases to transform goa type to protocol buffer type
		"to-protobuf-type": {
			{"primitive-to-primitive", primitiveGoa, primitiveProto, true, primitiveGoaToPrimitiveProtoCode},
			{"simple-to-simple", simpleGoa, simpleProto, true, simpleGoaToSimpleProtoCode},
			{"simple-to-required", simpleGoa, requiredProto, true, simpleGoaToRequiredProtoCode},
			{"required-to-simple", requiredGoa, simpleProto, true, requiredGoaToSimpleProtoCode},
			{"simple-to-default", simpleGoa, defaultTProto, true, simpleGoaToDefaultProtoCode},
			{"default-to-simple", defaultTGoa, simpleProto, true, defaultGoaToSimpleProtoCode},
			{"required-ptr-to-simple", requiredPtrGoa, simpleProto, true, requiredPtrGoaToSimpleProtoCode},

			// maps
			{"map-to-map", simpleMapGoa, simpleMapProto, true, simpleMapGoaToSimpleMapProtoCode},
			{"nested-map-to-nested-map", nestedMapGoa, nestedMapProto, true, nestedMapGoaToNestedMapProtoCode},
			{"array-map-to-array-map", arrayMapGoa, arrayMapProto, true, arrayMapGoaToArrayMapProtoCode},

			// arrays
			{"array-to-array", simpleArrayGoa, simpleArrayProto, true, simpleArrayGoaToSimpleArrayProtoCode},
			{"nested-array-to-nested-array", nestedArrayGoa, nestedArrayProto, true, nestedArrayGoaToNestedArrayProtoCode},
			{"type-array-to-type-array", typeArrayGoa, typeArrayProto, true, typeArrayGoaToTypeArrayProtoCode},
			{"map-array-to-map-array", mapArrayGoa, mapArrayProto, true, mapArrayGoaToMapArrayProtoCode},

			{"recursive-to-recursive", recursiveGoa, recursiveProto, true, recursiveGoaToRecursiveProtoCode},
			{"composite-to-custom-field", compositeGoa, customFieldProto, true, compositeGoaToCustomFieldProtoCode},
			{"custom-field-to-composite", customFieldGoa, compositeProto, true, customFieldGoaToCompositeProtoCode},
			{"result-type-to-result-type", resultTypeGoa, resultTypeProto, true, resultTypeGoaToResultTypeProtoCode},
			{"result-type-collection-to-result-type-collection", rtColGoa, rtColProto, true, rtColGoaToRTColProtoCode},
		},

		// test cases to transform protocol buffer type to goa type
		"to-goa-type": {
			{"primitive-to-primitive", primitiveProto, primitiveGoa, false, primitiveProtoToPrimitiveGoaCode},
			{"simple-to-simple", simpleProto, simpleGoa, false, simpleProtoToSimpleGoaCode},
			{"simple-to-required", simpleProto, requiredGoa, false, simpleProtoToRequiredGoaCode},
			{"required-to-simple", requiredProto, simpleGoa, false, requiredProtoToSimpleGoaCode},
			{"simple-to-default", simpleProto, defaultTGoa, false, simpleProtoToDefaultGoaCode},
			{"default-to-simple", defaultTProto, simpleGoa, false, defaultProtoToSimpleGoaCode},
			{"simple-to-required-ptr", simpleProto, requiredPtrGoa, false, simpleProtoToRequiredPtrGoaCode},

			// maps
			{"map-to-map", simpleMapProto, simpleMapGoa, false, simpleMapProtoToSimpleMapGoaCode},
			{"nested-map-to-nested-map", nestedMapProto, nestedMapGoa, false, nestedMapProtoToNestedMapGoaCode},
			{"array-map-to-array-map", arrayMapProto, arrayMapGoa, false, arrayMapProtoToArrayMapGoaCode},

			// arrays
			{"array-to-array", simpleArrayProto, simpleArrayGoa, false, simpleArrayProtoToSimpleArrayGoaCode},
			{"nested-array-to-nested-array", nestedArrayProto, nestedArrayGoa, false, nestedArrayProtoToNestedArrayGoaCode},
			{"type-array-to-type-array", typeArrayProto, typeArrayGoa, false, typeArrayProtoToTypeArrayGoaCode},
			{"map-array-to-map-array", mapArrayProto, mapArrayGoa, false, mapArrayProtoToMapArrayGoaCode},

			{"recursive-to-recursive", recursiveProto, recursiveGoa, false, recursiveProtoToRecursiveGoaCode},
			{"composite-to-custom-field", compositeProto, customFieldGoa, false, compositeProtoToCustomFieldGoaCode},
			{"custom-field-to-composite", customFieldProto, compositeGoa, false, customFieldProtoToCompositeGoaCode},
			{"result-type-to-result-type", resultTypeProto, resultTypeGoa, false, resultTypeProtoToResultTypeGoaCode},
			{"result-type-collection-to-result-type-collection", rtColProto, rtColGoa, false, rtColProtoToRTColGoaCode},
		},
	}
	for name, cases := range tc {
		t.Run(name, func(t *testing.T) {
			for _, c := range cases {
				t.Run(c.Name, func(t *testing.T) {
					code, _, err := protoBufTransform(c.Source, c.Target, "source", "target", c.ToProto)
					if err != nil {
						t.Fatal(err)
					}
					code = codegen.FormatTestCode(t, "package foo\nfunc transform(){\n"+code+"}")
					if code != c.Code {
						t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, codegen.Diff(t, code, c.Code))
					}
				})
			}
		})
	}
}

func newGoaContextAttr(dt expr.DataType, pkg string, scope *codegen.NameScope) *codegen.ContextualAttribute {
	att := codegen.NewGoAttribute(&expr.AttributeExpr{Type: dt}, pkg, scope)
	return codegen.NewUseDefaultContext(att)
}

func pointerContext(typ expr.DataType, pkg string, scope *codegen.NameScope) *codegen.ContextualAttribute {
	att := codegen.NewGoAttribute(&expr.AttributeExpr{Type: typ}, pkg, scope)
	ca := codegen.NewPointerContext(att)
	ca.OverrideRequired = true
	return ca
}

func newProtoContextAttr(dt expr.DataType, pkg string, scope *codegen.NameScope) *codegen.ContextualAttribute {
	att := &expr.AttributeExpr{Type: expr.Dup(dt)}
	att = makeProtoBufMessage(att, dt.Name(), scope)
	return protoBufContext(att, pkg, scope)
}

const (
	primitiveGoaToPrimitiveProtoCode = `func transform() {
	target := &Int{}
	target.Field = int32(source)
}
`

	simpleGoaToSimpleProtoCode = `func transform() {
	target := &Simple{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
	}
	if source.Integer != nil {
		target.Integer = int32(*source.Integer)
	}
}
`

	simpleGoaToRequiredProtoCode = `func transform() {
	target := &Required{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
	}
	if source.Integer != nil {
		target.Integer = int32(*source.Integer)
	}
}
`
	requiredGoaToSimpleProtoCode = `func transform() {
	target := &Simple{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
		Integer:        int32(source.Integer),
	}
}
`

	simpleGoaToDefaultProtoCode = `func transform() {
	target := &Default{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
	}
	if source.Integer != nil {
		target.Integer = int32(*source.Integer)
	}
	if source.Integer == nil {
		target.Integer = 1
	}
}
`

	defaultGoaToSimpleProtoCode = `func transform() {
	target := &Simple{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
		Integer:        int32(source.Integer),
	}
}
`

	requiredPtrGoaToSimpleProtoCode = `func transform() {
	target := &Simple{}
	if source.RequiredString != nil {
		target.RequiredString = *source.RequiredString
	}
	if source.DefaultBool != nil {
		target.DefaultBool = *source.DefaultBool
	}
	if source.Integer != nil {
		target.Integer = int32(*source.Integer)
	}
	if source.DefaultBool == nil {
		target.DefaultBool = true
	}
}
`

	simpleMapGoaToSimpleMapProtoCode = `func transform() {
	target := &SimpleMap{}
	if source.Simple != nil {
		target.Simple = make(map[string]int32, len(source.Simple))
		for key, val := range source.Simple {
			tk := key
			tv := int32(val)
			target.Simple[tk] = tv
		}
	}
}
`

	nestedMapGoaToNestedMapProtoCode = `func transform() {
	target := &NestedMap{}
	if source.NestedMap != nil {
		target.NestedMap = make(map[float64]*MapOfSint32MapOfDoubleUint64, len(source.NestedMap))
		for key, val := range source.NestedMap {
			tk := key
			tvc := &MapOfSint32MapOfDoubleUint64{}
			tvc.Field = make(map[int32]*MapOfDoubleUint64, len(val))
			for key, val := range val {
				tk := int32(key)
				tvb := &MapOfDoubleUint64{}
				tvb.Field = make(map[float64]uint64, len(val))
				for key, val := range val {
					tk := key
					tv := val
					tvb.Field[tk] = tv
				}
				tvc.Field[tk] = tvb
			}
			target.NestedMap[tk] = tvc
		}
	}
}
`

	arrayMapGoaToArrayMapProtoCode = `func transform() {
	target := &ArrayMap{}
	if source.ArrayMap != nil {
		target.ArrayMap = make(map[uint32]*ArrayOfFloat, len(source.ArrayMap))
		for key, val := range source.ArrayMap {
			tk := key
			tv := &ArrayOfFloat{}
			tv.Field = make([]float32, len(val))
			for i, val := range val {
				tv.Field[i] = val
			}
			target.ArrayMap[tk] = tv
		}
	}
}
`

	simpleArrayGoaToSimpleArrayProtoCode = `func transform() {
	target := &SimpleArray{}
	if source.StringArray != nil {
		target.StringArray = make([]string, len(source.StringArray))
		for i, val := range source.StringArray {
			target.StringArray[i] = val
		}
	}
}
`

	nestedArrayGoaToNestedArrayProtoCode = `func transform() {
	target := &NestedArray{}
	if source.NestedArray != nil {
		target.NestedArray = make([]*ArrayOfArrayOfDouble, len(source.NestedArray))
		for i, val := range source.NestedArray {
			target.NestedArray[i] = &ArrayOfArrayOfDouble{}
			target.NestedArray[i].Field = make([]*ArrayOfDouble, len(val))
			for j, val := range val {
				target.NestedArray[i].Field[j] = &ArrayOfDouble{}
				target.NestedArray[i].Field[j].Field = make([]float64, len(val))
				for k, val := range val {
					target.NestedArray[i].Field[j].Field[k] = val
				}
			}
		}
	}
}
`

	typeArrayGoaToTypeArrayProtoCode = `func transform() {
	target := &TypeArray{}
	if source.TypeArray != nil {
		target.TypeArray = make([]*SimpleArray, len(source.TypeArray))
		for i, val := range source.TypeArray {
			target.TypeArray[i] = &SimpleArray{}
			if val.StringArray != nil {
				target.TypeArray[i].StringArray = make([]string, len(val.StringArray))
				for j, val := range val.StringArray {
					target.TypeArray[i].StringArray[j] = val
				}
			}
		}
	}
}
`

	mapArrayGoaToMapArrayProtoCode = `func transform() {
	target := &MapArray{}
	if source.MapArray != nil {
		target.MapArray = make([]*MapOfSint32String, len(source.MapArray))
		for i, val := range source.MapArray {
			target.MapArray[i] = &MapOfSint32String{}
			target.MapArray[i].Field = make(map[int32]string, len(val))
			for key, val := range val {
				tk := int32(key)
				tv := val
				target.MapArray[i].Field[tk] = tv
			}
		}
	}
}
`

	recursiveGoaToRecursiveProtoCode = `func transform() {
	target := &Recursive{
		RequiredString: source.RequiredString,
	}
	if source.Recursive != nil {
		target.Recursive = svcRecursiveToRecursive(source.Recursive)
	}
}
`

	compositeGoaToCustomFieldProtoCode = `func transform() {
	target := &CompositeWithCustomField{}
	if source.RequiredString != nil {
		target.MyString = *source.RequiredString
	}
	if source.DefaultInt != nil {
		target.MyInt = int32(*source.DefaultInt)
	}
	if source.DefaultInt == nil {
		target.MyInt = 100
	}
	if source.Type != nil {
		target.MyType = svcSimpleToSimple(source.Type)
	}
	if source.Map != nil {
		target.MyMap = make(map[int32]string, len(source.Map))
		for key, val := range source.Map {
			tk := int32(key)
			tv := val
			target.MyMap[tk] = tv
		}
	}
	if source.Array != nil {
		target.MyArray = make([]string, len(source.Array))
		for i, val := range source.Array {
			target.MyArray[i] = val
		}
	}
}
`

	customFieldGoaToCompositeProtoCode = `func transform() {
	target := &Composite{
		RequiredString: source.MyString,
		DefaultInt:     int32(source.MyInt),
	}
	if source.MyType != nil {
		target.Type = svcSimpleToSimple(source.MyType)
	}
	if source.MyMap != nil {
		target.Map_ = make(map[int32]string, len(source.MyMap))
		for key, val := range source.MyMap {
			tk := int32(key)
			tv := val
			target.Map_[tk] = tv
		}
	}
	if source.MyArray != nil {
		target.Array = make([]string, len(source.MyArray))
		for i, val := range source.MyArray {
			target.Array[i] = val
		}
	}
}
`

	resultTypeGoaToResultTypeProtoCode = `func transform() {
	target := &ResultType{}
	if source.Int != nil {
		target.Int = int32(*source.Int)
	}
	if source.Map != nil {
		target.Map_ = make(map[int32]string, len(source.Map))
		for key, val := range source.Map {
			tk := int32(key)
			tv := val
			target.Map_[tk] = tv
		}
	}
}
`

	rtColGoaToRTColProtoCode = `func transform() {
	target := &ResultTypeCollection{}
	if source.Collection != nil {
		target.Collection = &ResultTypeCollection{}
		target.Collection.Field = make([]*ResultType, len(source.Collection))
		for i, val := range source.Collection {
			target.Collection.Field[i] = &ResultType{}
			if val.Int != nil {
				target.Collection.Field[i].Int = int32(*val.Int)
			}
			if val.Map != nil {
				target.Collection.Field[i].Map_ = make(map[int32]string, len(val.Map))
				for key, val := range val.Map {
					tk := int32(key)
					tv := val
					target.Collection.Field[i].Map_[tk] = tv
				}
			}
		}
	}
}
`

	primitiveProtoToPrimitiveGoaCode = `func transform() {
	target := int(source.Field)
}
`

	simpleProtoToSimpleGoaCode = `func transform() {
	target := &Simple{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
	}
	integerptr := int(source.Integer)
	target.Integer = &integerptr
}
`

	simpleProtoToRequiredGoaCode = `func transform() {
	target := &Required{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
		Integer:        int(source.Integer),
	}
}
`

	requiredProtoToSimpleGoaCode = `func transform() {
	target := &Simple{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
	}
	integerptr := int(source.Integer)
	target.Integer = &integerptr
}
`

	simpleProtoToDefaultGoaCode = `func transform() {
	target := &Default{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
		Integer:        int(source.Integer),
	}
}
`

	defaultProtoToSimpleGoaCode = `func transform() {
	target := &Simple{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
	}
	integerptr := int(source.Integer)
	target.Integer = &integerptr
}
`

	simpleProtoToRequiredPtrGoaCode = `func transform() {
	target := &Required{
		RequiredString: &source.RequiredString,
		DefaultBool:    &source.DefaultBool,
	}
	integerptr := int(source.Integer)
	target.Integer = &integerptr
}
`

	simpleMapProtoToSimpleMapGoaCode = `func transform() {
	target := &SimpleMap{}
	if source.Simple != nil {
		target.Simple = make(map[string]int, len(source.Simple))
		for key, val := range source.Simple {
			tk := key
			tv := int(val)
			target.Simple[tk] = tv
		}
	}
}
`

	nestedMapProtoToNestedMapGoaCode = `func transform() {
	target := &NestedMap{}
	if source.NestedMap != nil {
		target.NestedMap = make(map[float64]map[int]map[float64]uint64, len(source.NestedMap))
		for key, val := range source.NestedMap {
			tk := key
			tvc := make(map[int]map[float64]uint64, len(val.Field))
			for key, val := range val.Field {
				tk := int(key)
				tvb := make(map[float64]uint64, len(val.Field))
				for key, val := range val.Field {
					tk := key
					tv := val
					tvb[tk] = tv
				}
				tvc[tk] = tvb
			}
			target.NestedMap[tk] = tvc
		}
	}
}
`

	arrayMapProtoToArrayMapGoaCode = `func transform() {
	target := &ArrayMap{}
	if source.ArrayMap != nil {
		target.ArrayMap = make(map[uint32][]float32, len(source.ArrayMap))
		for key, val := range source.ArrayMap {
			tk := key
			tv := make([]float32, len(val.Field))
			for i, val := range val.Field {
				tv[i] = val
			}
			target.ArrayMap[tk] = tv
		}
	}
}
`

	simpleArrayProtoToSimpleArrayGoaCode = `func transform() {
	target := &SimpleArray{}
	if source.StringArray != nil {
		target.StringArray = make([]string, len(source.StringArray))
		for i, val := range source.StringArray {
			target.StringArray[i] = val
		}
	}
}
`

	nestedArrayProtoToNestedArrayGoaCode = `func transform() {
	target := &NestedArray{}
	if source.NestedArray != nil {
		target.NestedArray = make([][][]float64, len(source.NestedArray))
		for i, val := range source.NestedArray {
			target.NestedArray[i] = make([][]float64, len(val.Field))
			for j, val := range val.Field {
				target.NestedArray[i][j] = make([]float64, len(val.Field))
				for k, val := range val.Field {
					target.NestedArray[i][j][k] = val
				}
			}
		}
	}
}
`

	typeArrayProtoToTypeArrayGoaCode = `func transform() {
	target := &TypeArray{}
	if source.TypeArray != nil {
		target.TypeArray = make([]*SimpleArray, len(source.TypeArray))
		for i, val := range source.TypeArray {
			target.TypeArray[i] = &SimpleArray{}
			if val.StringArray != nil {
				target.TypeArray[i].StringArray = make([]string, len(val.StringArray))
				for j, val := range val.StringArray {
					target.TypeArray[i].StringArray[j] = val
				}
			}
		}
	}
}
`

	mapArrayProtoToMapArrayGoaCode = `func transform() {
	target := &MapArray{}
	if source.MapArray != nil {
		target.MapArray = make([]map[int]string, len(source.MapArray))
		for i, val := range source.MapArray {
			target.MapArray[i] = make(map[int]string, len(val.Field))
			for key, val := range val.Field {
				tk := int(key)
				tv := val
				target.MapArray[i][tk] = tv
			}
		}
	}
}
`

	recursiveProtoToRecursiveGoaCode = `func transform() {
	target := &Recursive{
		RequiredString: source.RequiredString,
	}
	if source.Recursive != nil {
		target.Recursive = protobufRecursiveToRecursive(source.Recursive)
	}
}
`

	compositeProtoToCustomFieldGoaCode = `func transform() {
	target := &CompositeWithCustomField{
		MyString: source.RequiredString,
		MyInt:    int(source.DefaultInt),
	}
	if source.Type != nil {
		target.MyType = protobufSimpleToSimple(source.Type)
	}
	if source.Map_ != nil {
		target.MyMap = make(map[int]string, len(source.Map_))
		for key, val := range source.Map_ {
			tk := int(key)
			tv := val
			target.MyMap[tk] = tv
		}
	}
	if source.Array != nil {
		target.MyArray = make([]string, len(source.Array))
		for i, val := range source.Array {
			target.MyArray[i] = val
		}
	}
}
`

	customFieldProtoToCompositeGoaCode = `func transform() {
	target := &Composite{
		RequiredString: &source.MyString,
	}
	defaultIntptr := int(source.MyInt)
	target.DefaultInt = &defaultIntptr
	if source.MyType != nil {
		target.Type = protobufSimpleToSimple(source.MyType)
	}
	if source.MyMap != nil {
		target.Map = make(map[int]string, len(source.MyMap))
		for key, val := range source.MyMap {
			tk := int(key)
			tv := val
			target.Map[tk] = tv
		}
	}
	if source.MyArray != nil {
		target.Array = make([]string, len(source.MyArray))
		for i, val := range source.MyArray {
			target.Array[i] = val
		}
	}
}
`

	resultTypeProtoToResultTypeGoaCode = `func transform() {
	target := &ResultType{}
	int_ptr := int(source.Int)
	target.Int = &int_ptr
	if source.Map_ != nil {
		target.Map = make(map[int]string, len(source.Map_))
		for key, val := range source.Map_ {
			tk := int(key)
			tv := val
			target.Map[tk] = tv
		}
	}
}
`

	rtColProtoToRTColGoaCode = `func transform() {
	target := &ResultTypeCollection{}
	if source.Collection != nil {
		target.Collection = make([]*ResultType, len(source.Collection.Field))
		for i, val := range source.Collection.Field {
			target.Collection[i] = &ResultType{}
			int_ptr := int(val.Int)
			target.Collection[i].Int = &int_ptr
			if val.Map_ != nil {
				target.Collection[i].Map = make(map[int]string, len(val.Map_))
				for key, val := range val.Map_ {
					tk := int(key)
					tv := val
					target.Collection[i].Map[tk] = tv
				}
			}
		}
	}
}
`
)
