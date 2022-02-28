package codegen

import (
	"testing"

	"goa.design/goa/v3/codegen"
	ctestdata "goa.design/goa/v3/codegen/testdata"
	"goa.design/goa/v3/expr"
)

func TestProtoBufTransform(t *testing.T) {
	root := codegen.RunDSL(t, ctestdata.TestTypesDSL)
	var (
		sd = &ServiceData{Name: "Service", Scope: codegen.NewNameScope()}

		// types to test
		primitive = expr.Int

		simple     = root.UserType("Simple")
		required   = root.UserType("Required")
		defaultT   = root.UserType("Default")
		customtype = root.UserType("CustomTypes")

		simpleMap  = root.UserType("SimpleMap")
		nestedMap  = root.UserType("NestedMap")
		arrayMap   = root.UserType("ArrayMap")
		defaultMap = root.UserType("DefaultMap")

		simpleArray  = root.UserType("SimpleArray")
		nestedArray  = root.UserType("NestedArray")
		mapArray     = root.UserType("MapArray")
		typeArray    = root.UserType("TypeArray")
		defaultArray = root.UserType("DefaultArray")

		recursive   = root.UserType("Recursive")
		composite   = root.UserType("Composite")
		customField = root.UserType("CompositeWithCustomField")
		optional    = root.UserType("Optional")
		defaults    = root.UserType("WithDefaults")

		resultType = root.UserType("ResultType")
		rtCol      = root.UserType("ResultTypeCollection")

		// attribute contexts used in test cases
		svcCtx = serviceTypeContext("", sd.Scope)
		ptrCtx = pointerContext("", sd.Scope)
		pbCtx  = protoBufTypeContext("", sd.Scope)
	)

	tc := map[string][]struct {
		Name    string
		Source  expr.DataType
		Target  expr.DataType
		ToProto bool
		Ctx     *codegen.AttributeContext
		Code    string
	}{
		// test cases to transform service type to protocol buffer type
		"to-protobuf-type": {
			{"primitive-to-primitive", primitive, primitive, true, svcCtx, primitiveSvcToPrimitiveProtoCode},
			{"simple-to-simple", simple, simple, true, svcCtx, simpleSvcToSimpleProtoCode},
			{"simple-to-required", simple, required, true, svcCtx, simpleSvcToRequiredProtoCode},
			{"required-to-simple", required, simple, true, svcCtx, requiredSvcToSimpleProtoCode},
			{"simple-to-default", simple, defaultT, true, svcCtx, simpleSvcToDefaultProtoCode},
			{"default-to-simple", defaultT, simple, true, svcCtx, defaultSvcToSimpleProtoCode},
			{"required-ptr-to-simple", required, simple, true, ptrCtx, requiredPtrSvcToSimpleProtoCode},
			{"simple-to-customtype", customtype, simple, true, svcCtx, customSvcToSimpleProtoCode},
			{"customtype-to-customtype", customtype, customtype, true, svcCtx, customSvcToCustomProtoCode},

			// maps
			{"map-to-map", simpleMap, simpleMap, true, svcCtx, simpleMapSvcToSimpleMapProtoCode},
			{"nested-map-to-nested-map", nestedMap, nestedMap, true, svcCtx, nestedMapSvcToNestedMapProtoCode},
			{"array-map-to-array-map", arrayMap, arrayMap, true, svcCtx, arrayMapSvcToArrayMapProtoCode},
			{"default-map-to-default-map", defaultMap, defaultMap, true, svcCtx, defaultMapSvcToDefaultMapProtoCode},

			// arrays
			{"array-to-array", simpleArray, simpleArray, true, svcCtx, simpleArraySvcToSimpleArrayProtoCode},
			{"nested-array-to-nested-array", nestedArray, nestedArray, true, svcCtx, nestedArraySvcToNestedArrayProtoCode},
			{"type-array-to-type-array", typeArray, typeArray, true, svcCtx, typeArraySvcToTypeArrayProtoCode},
			{"map-array-to-map-array", mapArray, mapArray, true, svcCtx, mapArraySvcToMapArrayProtoCode},
			{"default-array-to-default-array", defaultArray, defaultArray, true, svcCtx, defaultArraySvcToDefaultArrayProtoCode},

			{"recursive-to-recursive", recursive, recursive, true, svcCtx, recursiveSvcToRecursiveProtoCode},
			{"composite-to-custom-field", composite, customField, true, svcCtx, compositeSvcToCustomFieldProtoCode},
			{"custom-field-to-composite", customField, composite, true, svcCtx, customFieldSvcToCompositeProtoCode},
			{"result-type-to-result-type", resultType, resultType, true, svcCtx, resultTypeSvcToResultTypeProtoCode},
			{"result-type-collection-to-result-type-collection", rtCol, rtCol, true, svcCtx, rtColSvcToRTColProtoCode},
			{"optional-to-optional", optional, optional, true, svcCtx, optionalSvcToOptionalProtoCode},
			{"defaults-to-defaults", defaults, defaults, true, svcCtx, defaultsSvcToDefaultsProtoCode},
		},

		// test cases to transform protocol buffer type to service type
		"to-service-type": {
			{"primitive-to-primitive", primitive, primitive, false, svcCtx, primitiveProtoToPrimitiveSvcCode},
			{"simple-to-simple", simple, simple, false, svcCtx, simpleProtoToSimpleSvcCode},
			{"simple-to-required", simple, required, false, svcCtx, simpleProtoToRequiredSvcCode},
			{"required-to-simple", required, simple, false, svcCtx, requiredProtoToSimpleSvcCode},
			{"simple-to-default", simple, defaultT, false, svcCtx, simpleProtoToDefaultSvcCode},
			{"default-to-simple", defaultT, simple, false, svcCtx, defaultProtoToSimpleSvcCode},
			{"simple-to-required-ptr", simple, required, false, ptrCtx, simpleProtoToRequiredPtrSvcCode},
			{"simple-to-customtype", simple, customtype, false, svcCtx, simpleProtoToCustomSvcCode},
			{"customtype-to-customtype", customtype, customtype, false, svcCtx, customProtoToCustomSvcCode},

			// maps
			{"map-to-map", simpleMap, simpleMap, false, svcCtx, simpleMapProtoToSimpleMapSvcCode},
			{"nested-map-to-nested-map", nestedMap, nestedMap, false, svcCtx, nestedMapProtoToNestedMapSvcCode},
			{"array-map-to-array-map", arrayMap, arrayMap, false, svcCtx, arrayMapProtoToArrayMapSvcCode},
			{"default-map-to-default-map", defaultMap, defaultMap, false, svcCtx, defaultMapProtoToDefaultMapSvcCode},

			// arrays
			{"array-to-array", simpleArray, simpleArray, false, svcCtx, simpleArrayProtoToSimpleArraySvcCode},
			{"nested-array-to-nested-array", nestedArray, nestedArray, false, svcCtx, nestedArrayProtoToNestedArraySvcCode},
			{"type-array-to-type-array", typeArray, typeArray, false, svcCtx, typeArrayProtoToTypeArraySvcCode},
			{"map-array-to-map-array", mapArray, mapArray, false, svcCtx, mapArrayProtoToMapArraySvcCode},
			{"default-array-to-default-array", defaultArray, defaultArray, false, svcCtx, defaultArrayProtoToDefaultArraySvcCode},

			{"recursive-to-recursive", recursive, recursive, false, svcCtx, recursiveProtoToRecursiveSvcCode},
			{"composite-to-custom-field", composite, customField, false, svcCtx, compositeProtoToCustomFieldSvcCode},
			{"custom-field-to-composite", customField, composite, false, svcCtx, customFieldProtoToCompositeSvcCode},
			{"result-type-to-result-type", resultType, resultType, false, svcCtx, resultTypeProtoToResultTypeSvcCode},
			{"result-type-collection-to-result-type-collection", rtCol, rtCol, false, svcCtx, rtColProtoToRTColSvcCode},
			{"optional-to-optional", optional, optional, false, svcCtx, optionalProtoToOptionalSvcCode},
			{"defaults-to-defaults", defaults, defaults, false, svcCtx, defaultsProtoToDefaultsSvcCode},
		},
	}
	for name, cases := range tc {
		t.Run(name, func(t *testing.T) {
			for _, c := range cases {
				t.Run(c.Name, func(t *testing.T) {
					source := &expr.AttributeExpr{Type: c.Source}
					target := &expr.AttributeExpr{Type: c.Target}
					srcCtx := c.Ctx
					tgtCtx := c.Ctx
					if c.ToProto {
						target = makeProtoBufMessage(expr.DupAtt(target), target.Type.Name(), sd)
						tgtCtx = pbCtx
					} else {
						source = makeProtoBufMessage(expr.DupAtt(source), source.Type.Name(), sd)
						srcCtx = pbCtx
					}
					code, _, err := protoBufTransform(source, target, "source", "target", srcCtx, tgtCtx, c.ToProto, true)
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

func pointerContext(pkg string, scope *codegen.NameScope) *codegen.AttributeContext {
	return codegen.NewAttributeContext(true, false, true, pkg, scope)
}

const (
	primitiveSvcToPrimitiveProtoCode = `func transform() {
	target := &Int{}
	target.Field = int32(source)
}
`

	simpleSvcToSimpleProtoCode = `func transform() {
	target := &Simple{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
	}
	if source.Integer != nil {
		target.Integer = int32(*source.Integer)
	}
}
`

	simpleSvcToRequiredProtoCode = `func transform() {
	target := &Required{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
	}
	if source.Integer != nil {
		target.Integer = int32(*source.Integer)
	}
}
`
	requiredSvcToSimpleProtoCode = `func transform() {
	target := &Simple{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
		Integer:        int32(source.Integer),
	}
}
`

	simpleSvcToDefaultProtoCode = `func transform() {
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

	defaultSvcToSimpleProtoCode = `func transform() {
	target := &Simple{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
		Integer:        int32(source.Integer),
	}
}
`

	requiredPtrSvcToSimpleProtoCode = `func transform() {
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

	customSvcToSimpleProtoCode = `func transform() {
	target := &Simple{
		RequiredString: string(source.RequiredString),
		DefaultBool:    bool(source.DefaultBool),
	}
	if source.Integer != nil {
		target.Integer = int32(*source.Integer)
	}
}
`

	simpleProtoToCustomSvcCode = `func transform() {
	target := &CustomTypes{
		RequiredString: tdtypes.CustomString(source.RequiredString),
		DefaultBool:    tdtypes.CustomBool(source.DefaultBool),
	}
	if source.Integer != 0 {
		integerptr := tdtypes.CustomInt(source.Integer)
		target.Integer = &integerptr
	}
}
`

	customSvcToCustomProtoCode = `func transform() {
	target := &CustomTypes{
		RequiredString: tdtypes.CustomString(source.RequiredString),
		DefaultBool:    tdtypes.CustomBool(source.DefaultBool),
	}
	if source.Integer != nil {
		target.Integer = tdtypes.CustomInt(*source.Integer)
	}
}
`

	customProtoToCustomSvcCode = simpleProtoToCustomSvcCode

	simpleMapSvcToSimpleMapProtoCode = `func transform() {
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

	nestedMapSvcToNestedMapProtoCode = `func transform() {
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

	arrayMapSvcToArrayMapProtoCode = `func transform() {
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

	defaultMapSvcToDefaultMapProtoCode = `func transform() {
	target := &DefaultMap{}
	if source.Simple != nil {
		target.Simple = make(map[string]int32, len(source.Simple))
		for key, val := range source.Simple {
			tk := key
			tv := int32(val)
			target.Simple[tk] = tv
		}
	}
	if len(source.Simple) == 0 {
		target.Simple = map[string]int{"foo": 1}
	}
}
`

	simpleArraySvcToSimpleArrayProtoCode = `func transform() {
	target := &SimpleArray{}
	if source.StringArray != nil {
		target.StringArray = make([]string, len(source.StringArray))
		for i, val := range source.StringArray {
			target.StringArray[i] = val
		}
	}
}
`

	nestedArraySvcToNestedArrayProtoCode = `func transform() {
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

	typeArraySvcToTypeArrayProtoCode = `func transform() {
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

	mapArraySvcToMapArrayProtoCode = `func transform() {
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

	defaultArraySvcToDefaultArrayProtoCode = `func transform() {
	target := &DefaultArray{}
	if source.StringArray != nil {
		target.StringArray = make([]string, len(source.StringArray))
		for i, val := range source.StringArray {
			target.StringArray[i] = val
		}
	}
	if len(source.StringArray) == 0 {
		target.StringArray = []string{"foo", "bar"}
	}
}
`

	recursiveSvcToRecursiveProtoCode = `func transform() {
	target := &Recursive{
		RequiredString: source.RequiredString,
	}
	if source.Recursive != nil {
		target.Recursive = svcRecursiveToRecursive(source.Recursive)
	}
}
`

	compositeSvcToCustomFieldProtoCode = `func transform() {
	target := &CompositeWithCustomField{}
	if source.RequiredString != nil {
		target.RequiredString = *source.RequiredString
	}
	if source.DefaultInt != nil {
		target.DefaultInt = int32(*source.DefaultInt)
	}
	if source.DefaultInt == nil {
		target.DefaultInt = 100
	}
	if source.Type != nil {
		target.Type = svcSimpleToSimple(source.Type)
	}
	if source.Map != nil {
		target.Map_ = make(map[int32]string, len(source.Map))
		for key, val := range source.Map {
			tk := int32(key)
			tv := val
			target.Map_[tk] = tv
		}
	}
	if source.Array != nil {
		target.Array = make([]string, len(source.Array))
		for i, val := range source.Array {
			target.Array[i] = val
		}
	}
}
`

	customFieldSvcToCompositeProtoCode = `func transform() {
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

	resultTypeSvcToResultTypeProtoCode = `func transform() {
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

	rtColSvcToRTColProtoCode = `func transform() {
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

	optionalSvcToOptionalProtoCode = `func transform() {
	target := &Optional{
		Bytes_: source.Bytes,
		Any:    source.Any,
	}
	if source.Int != nil {
		target.Int = int32(*source.Int)
	}
	if source.Uint != nil {
		target.Uint = uint32(*source.Uint)
	}
	if source.Float != nil {
		target.Float_ = *source.Float
	}
	if source.String != nil {
		target.String_ = *source.String
	}
	if source.Array != nil {
		target.Array = make([]string, len(source.Array))
		for i, val := range source.Array {
			target.Array[i] = val
		}
	}
	if source.Map != nil {
		target.Map_ = make(map[int32]string, len(source.Map))
		for key, val := range source.Map {
			tk := int32(key)
			tv := val
			target.Map_[tk] = tv
		}
	}
	if source.UserType != nil {
		target.UserType = svcOptionalToOptional(source.UserType)
	}
}
`

	defaultsSvcToDefaultsProtoCode = `func transform() {
	target := &WithDefaults{
		Int:            int32(source.Int),
		RawJson:        json.RawMessage(source.RawJSON),
		RequiredInt:    int32(source.RequiredInt),
		String_:        source.String,
		RequiredString: source.RequiredString,
		Bytes_:         source.Bytes,
		RequiredBytes:  source.RequiredBytes,
		Any:            source.Any,
		RequiredAny:    source.RequiredAny,
	}
	if source.Array != nil {
		target.Array = make([]string, len(source.Array))
		for i, val := range source.Array {
			target.Array[i] = val
		}
	}
	if len(source.Array) == 0 {
		target.Array = []string{"foo", "bar"}
	}
	if source.RequiredArray != nil {
		target.RequiredArray = make([]string, len(source.RequiredArray))
		for i, val := range source.RequiredArray {
			target.RequiredArray[i] = val
		}
	}
	if source.Map != nil {
		target.Map_ = make(map[int32]string, len(source.Map))
		for key, val := range source.Map {
			tk := int32(key)
			tv := val
			target.Map_[tk] = tv
		}
	}
	if len(source.Map) == 0 {
		target.Map_ = map[int]string{1: "foo"}
	}
	if source.RequiredMap != nil {
		target.RequiredMap = make(map[int32]string, len(source.RequiredMap))
		for key, val := range source.RequiredMap {
			tk := int32(key)
			tv := val
			target.RequiredMap[tk] = tv
		}
	}
}
`

	primitiveProtoToPrimitiveSvcCode = `func transform() {
	target := int(source.Field)
}
`

	simpleProtoToSimpleSvcCode = `func transform() {
	target := &Simple{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
	}
	if source.Integer != 0 {
		integerptr := int(source.Integer)
		target.Integer = &integerptr
	}
}
`

	simpleProtoToRequiredSvcCode = `func transform() {
	target := &Required{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
		Integer:        int(source.Integer),
	}
}
`

	requiredProtoToSimpleSvcCode = `func transform() {
	target := &Simple{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
	}
	integerptr := int(source.Integer)
	target.Integer = &integerptr
}
`

	simpleProtoToDefaultSvcCode = `func transform() {
	target := &Default{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
		Integer:        int(source.Integer),
	}
	if source.Integer == 0 {
		target.Integer = 1
	}
}
`

	defaultProtoToSimpleSvcCode = `func transform() {
	target := &Simple{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
	}
	integerptr := int(source.Integer)
	target.Integer = &integerptr
}
`

	simpleProtoToRequiredPtrSvcCode = `func transform() {
	target := &Required{
		RequiredString: &source.RequiredString,
		DefaultBool:    &source.DefaultBool,
	}
	if source.Integer != 0 {
		integerptr := int(source.Integer)
		target.Integer = &integerptr
	}
}
`

	simpleMapProtoToSimpleMapSvcCode = `func transform() {
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

	nestedMapProtoToNestedMapSvcCode = `func transform() {
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

	arrayMapProtoToArrayMapSvcCode = `func transform() {
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

	defaultMapProtoToDefaultMapSvcCode = `func transform() {
	target := &DefaultMap{}
	if source.Simple != nil {
		target.Simple = make(map[string]int, len(source.Simple))
		for key, val := range source.Simple {
			tk := key
			tv := int(val)
			target.Simple[tk] = tv
		}
	}
	if len(source.Simple) == 0 {
		target.Simple = map[string]int{"foo": 1}
	}
}
`

	simpleArrayProtoToSimpleArraySvcCode = `func transform() {
	target := &SimpleArray{}
	if source.StringArray != nil {
		target.StringArray = make([]string, len(source.StringArray))
		for i, val := range source.StringArray {
			target.StringArray[i] = val
		}
	}
}
`

	nestedArrayProtoToNestedArraySvcCode = `func transform() {
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

	typeArrayProtoToTypeArraySvcCode = `func transform() {
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

	mapArrayProtoToMapArraySvcCode = `func transform() {
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

	defaultArrayProtoToDefaultArraySvcCode = `func transform() {
	target := &DefaultArray{}
	if source.StringArray != nil {
		target.StringArray = make([]string, len(source.StringArray))
		for i, val := range source.StringArray {
			target.StringArray[i] = val
		}
	}
	if len(source.StringArray) == 0 {
		target.StringArray = []string{"foo", "bar"}
	}
}
`

	recursiveProtoToRecursiveSvcCode = `func transform() {
	target := &Recursive{
		RequiredString: source.RequiredString,
	}
	if source.Recursive != nil {
		target.Recursive = protobufRecursiveToRecursive(source.Recursive)
	}
}
`

	compositeProtoToCustomFieldSvcCode = `func transform() {
	target := &CompositeWithCustomField{
		MyString: source.RequiredString,
		MyInt:    int(source.DefaultInt),
	}
	if source.DefaultInt == 0 {
		target.MyInt = 100
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

	customFieldProtoToCompositeSvcCode = `func transform() {
	target := &Composite{
		RequiredString: &source.RequiredString,
	}
	if source.DefaultInt != 0 {
		defaultIntptr := int(source.DefaultInt)
		target.DefaultInt = &defaultIntptr
	}
	if source.Type != nil {
		target.Type = protobufSimpleToSimple(source.Type)
	}
	if source.Map_ != nil {
		target.Map = make(map[int]string, len(source.Map_))
		for key, val := range source.Map_ {
			tk := int(key)
			tv := val
			target.Map[tk] = tv
		}
	}
	if source.Array != nil {
		target.Array = make([]string, len(source.Array))
		for i, val := range source.Array {
			target.Array[i] = val
		}
	}
}
`

	resultTypeProtoToResultTypeSvcCode = `func transform() {
	target := &ResultType{}
	if source.Int != 0 {
		int_ptr := int(source.Int)
		target.Int = &int_ptr
	}
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

	rtColProtoToRTColSvcCode = `func transform() {
	target := &ResultTypeCollection{}
	if source.Collection != nil {
		target.Collection = make([]*ResultType, len(source.Collection.Field))
		for i, val := range source.Collection.Field {
			target.Collection[i] = &ResultType{}
			if val.Int != 0 {
				int_ptr := int(val.Int)
				target.Collection[i].Int = &int_ptr
			}
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

	optionalProtoToOptionalSvcCode = `func transform() {
	target := &Optional{
		Bytes: source.Bytes_,
		Any:   source.Any,
	}
	if source.Int != 0 {
		int_ptr := int(source.Int)
		target.Int = &int_ptr
	}
	if source.Uint != 0 {
		uint_ptr := uint(source.Uint)
		target.Uint = &uint_ptr
	}
	if source.Float_ != 0 {
		target.Float = &source.Float_
	}
	if source.String_ != "" {
		target.String = &source.String_
	}
	if source.Array != nil {
		target.Array = make([]string, len(source.Array))
		for i, val := range source.Array {
			target.Array[i] = val
		}
	}
	if source.Map_ != nil {
		target.Map = make(map[int]string, len(source.Map_))
		for key, val := range source.Map_ {
			tk := int(key)
			tv := val
			target.Map[tk] = tv
		}
	}
	if source.UserType != nil {
		target.UserType = protobufOptionalToOptional(source.UserType)
	}
}
`

	defaultsProtoToDefaultsSvcCode = `func transform() {
	target := &WithDefaults{
		Int:            int(source.Int),
		RawJSON:        json.RawMessage(source.RawJson),
		RequiredInt:    int(source.RequiredInt),
		String:         source.String_,
		RequiredString: source.RequiredString,
		Bytes:          source.Bytes_,
		RequiredBytes:  source.RequiredBytes,
		Any:            source.Any,
		RequiredAny:    source.RequiredAny,
	}
	if source.Int == 0 {
		target.Int = 100
	}
	var zero json.RawMessage
	if source.RawJson == zero {
		target.RawJSON = json.RawMessage{0x66, 0x6f, 0x6f}
	}
	if source.String_ == "" {
		target.String = "foo"
	}
	if len(source.Bytes_) == 0 {
		target.Bytes = []byte{0x66, 0x6f, 0x6f, 0x62, 0x61, 0x72}
	}
	if source.Any == nil {
		target.Any = "something"
	}
	if source.Array != nil {
		target.Array = make([]string, len(source.Array))
		for i, val := range source.Array {
			target.Array[i] = val
		}
	}
	if len(source.Array) == 0 {
		target.Array = []string{"foo", "bar"}
	}
	if source.RequiredArray != nil {
		target.RequiredArray = make([]string, len(source.RequiredArray))
		for i, val := range source.RequiredArray {
			target.RequiredArray[i] = val
		}
	}
	if source.Map_ != nil {
		target.Map = make(map[int]string, len(source.Map_))
		for key, val := range source.Map_ {
			tk := int(key)
			tv := val
			target.Map[tk] = tv
		}
	}
	if len(source.Map_) == 0 {
		target.Map = map[int]string{1: "foo"}
	}
	if source.RequiredMap != nil {
		target.RequiredMap = make(map[int]string, len(source.RequiredMap))
		for key, val := range source.RequiredMap {
			tk := int(key)
			tv := val
			target.RequiredMap[tk] = tv
		}
	}
}
`
)
