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

		simpleOneOf    = root.UserType("SimpleOneOf")
		embeddedOneOf  = root.UserType("EmbeddedOneOf")
		recursiveOneOf = root.UserType("RecursiveOneOf")

		pkgOverride = root.UserType("CompositePkgOverride")

		// attribute contexts used in test cases
		svcCtx = serviceTypeContext("proto", sd.Scope)
		ptrCtx = pointerContext("proto", sd.Scope)
		pbCtx  = protoBufTypeContext("proto", sd.Scope, true)
	)

	// gRPC does not support any
	obj := expr.AsObject(defaults)
	for _, nat := range *obj {
		if nat.Name == "any" {
			nat.Attribute.Type = expr.String
		}
		if nat.Name == "required_any" {
			nat.Attribute.Type = expr.String
		}
	}

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

			// oneofs
			{"oneof-to-oneof", simpleOneOf, simpleOneOf, true, svcCtx, oneOfSvcToOneOfProtoCode},
			{"embedded-oneof-to-embedded-oneof", embeddedOneOf, embeddedOneOf, true, svcCtx, embeddedOneOfSvcToEmbeddedOneOfProtoCode},
			{"recursive-oneof-to-recursive-oneof", recursiveOneOf, recursiveOneOf, true, svcCtx, recursiveOneOfSvcToRecursiveOneOfProtoCode},

			// package override
			{"pkg-override-to-pkg-override", pkgOverride, pkgOverride, true, svcCtx, pkgOverrideSvcToPkgOverrideProtoCode},
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

			// oneofs
			{"oneof-to-oneof", simpleOneOf, simpleOneOf, false, svcCtx, oneOfProtoToOneOfSvcCode},
			{"embedded-oneof-to-embedded-oneof", embeddedOneOf, embeddedOneOf, false, svcCtx, embeddedOneOfProtoToEmbeddedOneOfSvcCode},
			{"recursive-oneof-to-recursive-oneof", recursiveOneOf, recursiveOneOf, false, svcCtx, recursiveOneOfProtoToRecursiveOneOfSvcCode},

			// package override
			{"pkg-override-to-pkg-override", pkgOverride, pkgOverride, false, svcCtx, pkgOverrideProtoToPkgOverrideSvcCode},
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
	target := &proto.Int{}
	target.Field = int32(source)
}
`

	simpleSvcToSimpleProtoCode = `func transform() {
	target := &proto.Simple{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
	}
	if source.Integer != nil {
		integer := int32(*source.Integer)
		target.Integer = &integer
	}
	{
		var zero bool
		if target.DefaultBool == zero {
			target.DefaultBool = true
		}
	}
}
`

	simpleSvcToRequiredProtoCode = `func transform() {
	target := &proto.Required{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
	}
	if source.Integer != nil {
		target.Integer = int32(*source.Integer)
	}
	{
		var zero bool
		if target.DefaultBool == zero {
			target.DefaultBool = true
		}
	}
}
`
	requiredSvcToSimpleProtoCode = `func transform() {
	target := &proto.Simple{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
	}
	integer := int32(source.Integer)
	target.Integer = &integer
}
`

	simpleSvcToDefaultProtoCode = `func transform() {
	target := &proto.Default{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
	}
	if source.Integer != nil {
		target.Integer = int32(*source.Integer)
	}
	{
		var zero bool
		if target.DefaultBool == zero {
			target.DefaultBool = true
		}
	}
	if source.Integer == nil {
		target.Integer = 1
	}
}
`

	defaultSvcToSimpleProtoCode = `func transform() {
	target := &proto.Simple{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
	}
	integer := int32(source.Integer)
	target.Integer = &integer
	{
		var zero bool
		if target.DefaultBool == zero {
			target.DefaultBool = true
		}
	}
}
`

	requiredPtrSvcToSimpleProtoCode = `func transform() {
	target := &proto.Simple{
		RequiredString: *source.RequiredString,
		DefaultBool:    *source.DefaultBool,
	}
	integer := int32(*source.Integer)
	target.Integer = &integer
}
`

	customSvcToSimpleProtoCode = `func transform() {
	target := &proto.Simple{
		RequiredString: string(source.RequiredString),
		DefaultBool:    bool(source.DefaultBool),
	}
	if source.Integer != nil {
		integer := int32(*source.Integer)
		target.Integer = &integer
	}
	{
		var zero bool
		if target.DefaultBool == zero {
			target.DefaultBool = true
		}
	}
}
`

	simpleProtoToCustomSvcCode = `func transform() {
	target := &proto.CustomTypes{
		RequiredString: tdtypes.CustomString(source.RequiredString),
		DefaultBool:    tdtypes.CustomBool(source.DefaultBool),
	}
	if source.Integer != nil {
		integer := tdtypes.CustomInt(*source.Integer)
		target.Integer = &integer
	}
	{
		var zero tdtypes.CustomBool
		if target.DefaultBool == zero {
			target.DefaultBool = true
		}
	}
}
`

	customSvcToCustomProtoCode = `func transform() {
	target := &proto.CustomTypes{
		RequiredString: string(source.RequiredString),
		DefaultBool:    bool(source.DefaultBool),
	}
	if source.Integer != nil {
		integer := int32(*source.Integer)
		target.Integer = &integer
	}
	{
		var zero bool
		if target.DefaultBool == zero {
			target.DefaultBool = true
		}
	}
}
`

	customProtoToCustomSvcCode = `func transform() {
	target := &proto.CustomTypes{
		RequiredString: tdtypes.CustomString(source.RequiredString),
		DefaultBool:    tdtypes.CustomBool(source.DefaultBool),
	}
	if source.Integer != nil {
		integer := tdtypes.CustomInt(*source.Integer)
		target.Integer = &integer
	}
	{
		var zero tdtypes.CustomBool
		if target.DefaultBool == zero {
			target.DefaultBool = true
		}
	}
}
`

	simpleMapSvcToSimpleMapProtoCode = `func transform() {
	target := &proto.SimpleMap{}
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
	target := &proto.NestedMap{}
	if source.NestedMap != nil {
		target.NestedMap = make(map[float64]*proto.MapOfSint32MapOfDoubleUint64, len(source.NestedMap))
		for key, val := range source.NestedMap {
			tk := key
			tvc := &proto.MapOfSint32MapOfDoubleUint64{}
			tvc.Field = make(map[int32]*proto.MapOfDoubleUint64, len(val))
			for key, val := range val {
				tk := int32(key)
				tvb := &proto.MapOfDoubleUint64{}
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
	target := &proto.ArrayMap{}
	if source.ArrayMap != nil {
		target.ArrayMap = make(map[uint32]*proto.ArrayOfFloat, len(source.ArrayMap))
		for key, val := range source.ArrayMap {
			tk := key
			tv := &proto.ArrayOfFloat{}
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
	target := &proto.DefaultMap{}
	if source.Simple != nil {
		target.Simple = make(map[string]int32, len(source.Simple))
		for key, val := range source.Simple {
			tk := key
			tv := int32(val)
			target.Simple[tk] = tv
		}
	}
	if source.Simple == nil {
		target.Simple = map[string]int{"foo": 1}
	}
}
`

	simpleArraySvcToSimpleArrayProtoCode = `func transform() {
	target := &proto.SimpleArray{}
	if source.StringArray != nil {
		target.StringArray = make([]string, len(source.StringArray))
		for i, val := range source.StringArray {
			target.StringArray[i] = val
		}
	}
}
`

	nestedArraySvcToNestedArrayProtoCode = `func transform() {
	target := &proto.NestedArray{}
	if source.NestedArray != nil {
		target.NestedArray = make([]*proto.ArrayOfArrayOfDouble, len(source.NestedArray))
		for i, val := range source.NestedArray {
			target.NestedArray[i] = &proto.ArrayOfArrayOfDouble{}
			target.NestedArray[i].Field = make([]*proto.ArrayOfDouble, len(val))
			for j, val := range val {
				target.NestedArray[i].Field[j] = &proto.ArrayOfDouble{}
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
	target := &proto.TypeArray{}
	if source.TypeArray != nil {
		target.TypeArray = make([]*proto.SimpleArray, len(source.TypeArray))
		for i, val := range source.TypeArray {
			target.TypeArray[i] = &proto.SimpleArray{}
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
	target := &proto.MapArray{}
	if source.MapArray != nil {
		target.MapArray = make([]*proto.MapOfSint32String, len(source.MapArray))
		for i, val := range source.MapArray {
			target.MapArray[i] = &proto.MapOfSint32String{}
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
	target := &proto.DefaultArray{}
	if source.StringArray != nil {
		target.StringArray = make([]string, len(source.StringArray))
		for i, val := range source.StringArray {
			target.StringArray[i] = val
		}
	}
	if source.StringArray == nil {
		target.StringArray = []string{"foo", "bar"}
	}
}
`

	recursiveSvcToRecursiveProtoCode = `func transform() {
	target := &proto.Recursive{
		RequiredString: source.RequiredString,
	}
	if source.Recursive != nil {
		target.Recursive = svcProtoRecursiveToProtoRecursive(source.Recursive)
	}
}
`

	compositeSvcToCustomFieldProtoCode = `func transform() {
	target := &proto.CompositeWithCustomField{}
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
		target.Type = svcProtoSimpleToProtoSimple(source.Type)
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
	target := &proto.Composite{
		RequiredString: &source.MyString,
	}
	defaultInt := int32(source.MyInt)
	target.DefaultInt = &defaultInt
	if source.MyType != nil {
		target.Type = svcProtoSimpleToProtoSimple(source.MyType)
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
	target := &proto.ResultType{}
	if source.Int != nil {
		int_ := int32(*source.Int)
		target.Int = &int_
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
	target := &proto.ResultTypeCollection{}
	if source.Collection != nil {
		target.Collection = &proto.ResultTypeCollection{}
		target.Collection.Field = make([]*proto.ResultType, len(source.Collection))
		for i, val := range source.Collection {
			target.Collection.Field[i] = &proto.ResultType{}
			if val.Int != nil {
				int_ := int32(*val.Int)
				target.Collection.Field[i].Int = &int_
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
	target := &proto.Optional{
		Float_:  source.Float,
		String_: source.String,
		Bytes_:  source.Bytes,
		Any:     source.Any,
	}
	if source.Int != nil {
		int_ := int32(*source.Int)
		target.Int = &int_
	}
	if source.Uint != nil {
		uint_ := uint32(*source.Uint)
		target.Uint = &uint_
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
		target.UserType = svcProtoOptionalToProtoOptional(source.UserType)
	}
}
`

	defaultsSvcToDefaultsProtoCode = `func transform() {
	target := &proto.WithDefaults{
		Int:            int32(source.Int),
		RawJson:        string(source.RawJSON),
		RequiredInt:    int32(source.RequiredInt),
		String_:        source.String,
		RequiredString: source.RequiredString,
		Bytes_:         source.Bytes,
		RequiredBytes:  source.RequiredBytes,
		Any:            source.Any,
		RequiredAny:    source.RequiredAny,
	}
	{
		var zero int32
		if target.Int == zero {
			target.Int = 100
		}
	}
	{
		var zero string
		if target.RawJson == zero {
			target.RawJson = json.RawMessage{0x66, 0x6f, 0x6f}
		}
	}
	{
		var zero string
		if target.String_ == zero {
			target.String_ = "foo"
		}
	}
	{
		var zero []byte
		if target.Bytes_ == zero {
			target.Bytes_ = []byte{0x66, 0x6f, 0x6f, 0x62, 0x61, 0x72}
		}
	}
	{
		var zero string
		if target.Any == zero {
			target.Any = "something"
		}
	}
	if source.Array != nil {
		target.Array = make([]string, len(source.Array))
		for i, val := range source.Array {
			target.Array[i] = val
		}
	}
	if source.Array == nil {
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
	if source.Map == nil {
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

	oneOfSvcToOneOfProtoCode = `func transform() {
	target := &proto.SimpleOneOf{}
	if source.SimpleOneOf != nil {
		switch src := source.SimpleOneOf.(type) {
		case proto.SimpleOneOfString:
			target.SimpleOneOf = &proto.SimpleOneOf_String_{String_: string(src)}
		case proto.SimpleOneOfInteger:
			target.SimpleOneOf = &proto.SimpleOneOf_Integer{Integer: int32(src)}
		}
	}
}
`

	embeddedOneOfSvcToEmbeddedOneOfProtoCode = `func transform() {
	target := &proto.EmbeddedOneOf{
		String_: source.String,
	}
	if source.EmbeddedOneOf != nil {
		switch src := source.EmbeddedOneOf.(type) {
		case proto.EmbeddedOneOfString:
			target.EmbeddedOneOf = &proto.EmbeddedOneOf_String_{String_: string(src)}
		case proto.EmbeddedOneOfInteger:
			target.EmbeddedOneOf = &proto.EmbeddedOneOf_Integer{Integer: int32(src)}
		case proto.EmbeddedOneOfBoolean:
			target.EmbeddedOneOf = &proto.EmbeddedOneOf_Boolean{Boolean: bool(src)}
		case proto.EmbeddedOneOfNumber:
			target.EmbeddedOneOf = &proto.EmbeddedOneOf_Number{Number: int32(src)}
		case proto.EmbeddedOneOfArray:
			target.EmbeddedOneOf = &proto.EmbeddedOneOf_Array{Array: svcProtoEmbeddedOneOfArrayToProtoEmbeddedOneOfArray(src)}
		case proto.EmbeddedOneOfMap:
			target.EmbeddedOneOf = &proto.EmbeddedOneOf_Map_{Map_: svcProtoEmbeddedOneOfMapToProtoEmbeddedOneOfMap(src)}
		case *proto.SimpleOneOf:
			target.EmbeddedOneOf = &proto.EmbeddedOneOf_UserType{UserType: svcProtoSimpleOneOfToProtoSimpleOneOf(src)}
		}
	}
}
`

	recursiveOneOfSvcToRecursiveOneOfProtoCode = `func transform() {
	target := &proto.RecursiveOneOf{
		String_: source.String,
	}
	if source.RecursiveOneOf != nil {
		switch src := source.RecursiveOneOf.(type) {
		case proto.RecursiveOneOfInteger:
			target.RecursiveOneOf = &proto.RecursiveOneOf_Integer{Integer: int32(src)}
		case *proto.RecursiveOneOf:
			target.RecursiveOneOf = &proto.RecursiveOneOf_Recurse{Recurse: svcProtoRecursiveOneOfToProtoRecursiveOneOf(src)}
		}
	}
}
`

	pkgOverrideSvcToPkgOverrideProtoCode = `func transform() {
	target := &proto.CompositePkgOverride{}
	if source.WithOverride != nil {
		target.WithOverride = svcTypesWithOverrideToProtoWithOverride(source.WithOverride)
	}
}
`

	primitiveProtoToPrimitiveSvcCode = `func transform() {
	target := int(source.Field)
}
`

	simpleProtoToSimpleSvcCode = `func transform() {
	target := &proto.Simple{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
	}
	if source.Integer != nil {
		integer := int(*source.Integer)
		target.Integer = &integer
	}
	{
		var zero bool
		if target.DefaultBool == zero {
			target.DefaultBool = true
		}
	}
}
`

	simpleProtoToRequiredSvcCode = `func transform() {
	target := &proto.Required{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
	}
	if source.Integer != nil {
		target.Integer = int(*source.Integer)
	}
	{
		var zero bool
		if target.DefaultBool == zero {
			target.DefaultBool = true
		}
	}
}
`

	requiredProtoToSimpleSvcCode = `func transform() {
	target := &proto.Simple{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
	}
	integer := int(source.Integer)
	target.Integer = &integer
}
`

	simpleProtoToDefaultSvcCode = `func transform() {
	target := &proto.Default{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
	}
	if source.Integer != nil {
		target.Integer = int(*source.Integer)
	}
	{
		var zero bool
		if target.DefaultBool == zero {
			target.DefaultBool = true
		}
	}
	if source.Integer == nil {
		target.Integer = 1
	}
}
`

	defaultProtoToSimpleSvcCode = `func transform() {
	target := &proto.Simple{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
	}
	integer := int(source.Integer)
	target.Integer = &integer
	{
		var zero bool
		if target.DefaultBool == zero {
			target.DefaultBool = true
		}
	}
}
`

	simpleProtoToRequiredPtrSvcCode = `func transform() {
	target := &proto.Required{
		RequiredString: &source.RequiredString,
		DefaultBool:    &source.DefaultBool,
	}
	if source.Integer != nil {
		integer := int(*source.Integer)
		target.Integer = &integer
	}
}
`

	simpleMapProtoToSimpleMapSvcCode = `func transform() {
	target := &proto.SimpleMap{}
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
	target := &proto.NestedMap{}
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
	target := &proto.ArrayMap{}
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
	target := &proto.DefaultMap{}
	if source.Simple != nil {
		target.Simple = make(map[string]int, len(source.Simple))
		for key, val := range source.Simple {
			tk := key
			tv := int(val)
			target.Simple[tk] = tv
		}
	}
	if source.Simple == nil {
		target.Simple = map[string]int{"foo": 1}
	}
}
`

	simpleArrayProtoToSimpleArraySvcCode = `func transform() {
	target := &proto.SimpleArray{}
	if source.StringArray != nil {
		target.StringArray = make([]string, len(source.StringArray))
		for i, val := range source.StringArray {
			target.StringArray[i] = val
		}
	}
}
`

	nestedArrayProtoToNestedArraySvcCode = `func transform() {
	target := &proto.NestedArray{}
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
	target := &proto.TypeArray{}
	if source.TypeArray != nil {
		target.TypeArray = make([]*proto.SimpleArray, len(source.TypeArray))
		for i, val := range source.TypeArray {
			target.TypeArray[i] = &proto.SimpleArray{}
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
	target := &proto.MapArray{}
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
	target := &proto.DefaultArray{}
	if source.StringArray != nil {
		target.StringArray = make([]string, len(source.StringArray))
		for i, val := range source.StringArray {
			target.StringArray[i] = val
		}
	}
	if source.StringArray == nil {
		target.StringArray = []string{"foo", "bar"}
	}
}
`

	recursiveProtoToRecursiveSvcCode = `func transform() {
	target := &proto.Recursive{
		RequiredString: source.RequiredString,
	}
	if source.Recursive != nil {
		target.Recursive = protobufProtoRecursiveToProtoRecursive(source.Recursive)
	}
}
`

	compositeProtoToCustomFieldSvcCode = `func transform() {
	target := &proto.CompositeWithCustomField{}
	if source.RequiredString != nil {
		target.MyString = *source.RequiredString
	}
	if source.DefaultInt != nil {
		target.MyInt = int(*source.DefaultInt)
	}
	if source.DefaultInt == nil {
		target.MyInt = 100
	}
	if source.Type != nil {
		target.MyType = protobufProtoSimpleToProtoSimple(source.Type)
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
	target := &proto.Composite{
		RequiredString: &source.RequiredString,
	}
	defaultInt := int(source.DefaultInt)
	target.DefaultInt = &defaultInt
	if source.Type != nil {
		target.Type = protobufProtoSimpleToProtoSimple(source.Type)
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
	target := &proto.ResultType{}
	if source.Int != nil {
		int_ := int(*source.Int)
		target.Int = &int_
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
	target := &proto.ResultTypeCollection{}
	if source.Collection != nil {
		target.Collection = make([]*proto.ResultType, len(source.Collection.Field))
		for i, val := range source.Collection.Field {
			target.Collection[i] = &proto.ResultType{}
			if val.Int != nil {
				int_ := int(*val.Int)
				target.Collection[i].Int = &int_
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
	target := &proto.Optional{
		Float:  source.Float_,
		String: source.String_,
		Bytes:  source.Bytes_,
		Any:    source.Any,
	}
	if source.Int != nil {
		int_ := int(*source.Int)
		target.Int = &int_
	}
	if source.Uint != nil {
		uint_ := uint(*source.Uint)
		target.Uint = &uint_
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
		target.UserType = protobufProtoOptionalToProtoOptional(source.UserType)
	}
}
`

	defaultsProtoToDefaultsSvcCode = `func transform() {
	target := &proto.WithDefaults{
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
	{
		var zero int
		if target.Int == zero {
			target.Int = 100
		}
	}
	{
		var zero json.RawMessage
		if target.RawJSON == zero {
			target.RawJSON = json.RawMessage{0x66, 0x6f, 0x6f}
		}
	}
	{
		var zero string
		if target.String == zero {
			target.String = "foo"
		}
	}
	{
		var zero []byte
		if target.Bytes == zero {
			target.Bytes = []byte{0x66, 0x6f, 0x6f, 0x62, 0x61, 0x72}
		}
	}
	{
		var zero string
		if target.Any == zero {
			target.Any = "something"
		}
	}
	if source.Array != nil {
		target.Array = make([]string, len(source.Array))
		for i, val := range source.Array {
			target.Array[i] = val
		}
	}
	if source.Array == nil {
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
	if source.Map_ == nil {
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

	oneOfProtoToOneOfSvcCode = `func transform() {
	target := &proto.SimpleOneOf{}
	if source.SimpleOneOf != nil {
		switch val := source.SimpleOneOf.(type) {
		case *proto.SimpleOneOf_String_:
			target.SimpleOneOf = proto.SimpleOneOfString(val.String_)
		case *proto.SimpleOneOf_Integer:
			target.SimpleOneOf = proto.SimpleOneOfInteger(val.Integer)
		}
	}
}
`

	embeddedOneOfProtoToEmbeddedOneOfSvcCode = `func transform() {
	target := &proto.EmbeddedOneOf{
		String: source.String_,
	}
	if source.EmbeddedOneOf != nil {
		switch val := source.EmbeddedOneOf.(type) {
		case *proto.EmbeddedOneOf_String_:
			target.EmbeddedOneOf = proto.EmbeddedOneOfString(val.String_)
		case *proto.EmbeddedOneOf_Integer:
			target.EmbeddedOneOf = proto.EmbeddedOneOfInteger(val.Integer)
		case *proto.EmbeddedOneOf_Boolean:
			target.EmbeddedOneOf = proto.EmbeddedOneOfBoolean(val.Boolean)
		case *proto.EmbeddedOneOf_Number:
			target.EmbeddedOneOf = proto.EmbeddedOneOfNumber(val.Number)
		case *proto.EmbeddedOneOf_Array:
			target.EmbeddedOneOf = protobufProtoEmbeddedOneOfArrayToProtoEmbeddedOneOfArray(val.Array)
		case *proto.EmbeddedOneOf_Map_:
			target.EmbeddedOneOf = protobufProtoEmbeddedOneOfMapToProtoEmbeddedOneOfMap(val.Map_)
		case *proto.EmbeddedOneOf_UserType:
			target.EmbeddedOneOf = protobufProtoSimpleOneOfToProtoSimpleOneOf(val.UserType)
		}
	}
}
`

	recursiveOneOfProtoToRecursiveOneOfSvcCode = `func transform() {
	target := &proto.RecursiveOneOf{
		String: source.String_,
	}
	if source.RecursiveOneOf != nil {
		switch val := source.RecursiveOneOf.(type) {
		case *proto.RecursiveOneOf_Integer:
			target.RecursiveOneOf = proto.RecursiveOneOfInteger(val.Integer)
		case *proto.RecursiveOneOf_Recurse:
			target.RecursiveOneOf = protobufProtoRecursiveOneOfToProtoRecursiveOneOf(val.Recurse)
		}
	}
}
`

	pkgOverrideProtoToPkgOverrideSvcCode = `func transform() {
	target := &types.CompositePkgOverride{}
	if source.WithOverride != nil {
		target.WithOverride = protobufProtoWithOverrideToTypesWithOverride(source.WithOverride)
	}
}
`
)
