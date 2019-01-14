package codegen

import (
	"testing"

	"goa.design/goa/codegen/testdata"
	"goa.design/goa/expr"
)

func defaultContext(typ expr.DataType, pkg string, scope *NameScope) *ContextualAttribute {
	att := NewGoAttribute(&expr.AttributeExpr{Type: typ}, pkg, scope)
	return NewUseDefaultContext(att)
}

func pointerContext(typ expr.DataType, pkg string, scope *NameScope) *ContextualAttribute {
	att := NewGoAttribute(&expr.AttributeExpr{Type: typ}, pkg, scope)
	return NewPointerContext(att)
}

func TestGoTransform(t *testing.T) {
	root := RunDSL(t, testdata.TestTypesDSL)
	var (
		scope = NewNameScope()

		// types to test
		simple   = root.UserType("Simple")
		super    = root.UserType("Super")
		required = root.UserType("Required")
		defaultT = root.UserType("Default")

		simpleMap   = root.UserType("SimpleMap")
		requiredMap = root.UserType("RequiredMap")
		defaultMap  = root.UserType("DefaultMap")
		nestedMap   = root.UserType("NestedMap")
		typeMap     = root.UserType("TypeMap")
		arrayMap    = root.UserType("ArrayMap")

		simpleArray   = root.UserType("SimpleArray")
		requiredArray = root.UserType("RequiredArray")
		defaultArray  = root.UserType("DefaultArray")
		nestedArray   = root.UserType("NestedArray")
		typeArray     = root.UserType("TypeArray")
		mapArray      = root.UserType("MapArray")

		recursive   = root.UserType("Recursive")
		composite   = root.UserType("Composite")
		customField = root.UserType("CompositeWithCustomField")

		resultType = root.UserType("ResultType")
		rtCol      = root.UserType("ResultTypeCollection")

		// attribute analyzers used in test cases
		simpleUseDefault         = defaultContext(simple, "", scope)
		requiredUseDefault       = defaultContext(required, "", scope)
		superUseDefault          = defaultContext(super, "", scope)
		defaultTUseDefault       = defaultContext(defaultT, "", scope)
		simpleMapUseDefault      = defaultContext(simpleMap, "", scope)
		requiredMapUseDefault    = defaultContext(requiredMap, "", scope)
		defaultMapUseDefault     = defaultContext(defaultMap, "", scope)
		nestedMapUseDefault      = defaultContext(nestedMap, "", scope)
		typeMapUseDefault        = defaultContext(typeMap, "", scope)
		arrayMapUseDefault       = defaultContext(arrayMap, "", scope)
		simpleArrayUseDefault    = defaultContext(simpleArray, "", scope)
		requiredArrayUseDefault  = defaultContext(requiredArray, "", scope)
		defaultArrayUseDefault   = defaultContext(defaultArray, "", scope)
		nestedArrayUseDefault    = defaultContext(nestedArray, "", scope)
		typeArrayUseDefault      = defaultContext(typeArray, "", scope)
		mapArrayUseDefault       = defaultContext(mapArray, "", scope)
		recursiveUseDefault      = defaultContext(recursive, "", scope)
		compositeUseDefault      = defaultContext(composite, "", scope)
		customFieldUseDefault    = defaultContext(customField, "", scope)
		customFieldPkgUseDefault = defaultContext(customField, "mypkg", scope)
		resultTypeUseDefault     = defaultContext(resultType, "", scope)
		rtColUseDefault          = defaultContext(rtCol, "", scope)

		simplePointer        = pointerContext(simple, "", scope)
		requiredPointer      = pointerContext(required, "", scope)
		superPointer         = pointerContext(super, "", scope)
		defaultTPointer      = pointerContext(defaultT, "", scope)
		requiredMapPointer   = pointerContext(requiredMap, "", scope)
		defaultMapPointer    = pointerContext(defaultMap, "", scope)
		requiredArrayPointer = pointerContext(requiredArray, "", scope)
		defaultArrayPointer  = pointerContext(defaultArray, "", scope)
		recursivePointer     = pointerContext(recursive, "", scope)
		customFieldPointer   = pointerContext(customField, "", scope)
	)
	tc := map[string][]struct {
		Name   string
		Source *ContextualAttribute
		Target *ContextualAttribute
		Code   string
	}{
		// source and target type use default
		"source-target-type-use-default": {
			{"simple-to-simple", simpleUseDefault, simpleUseDefault, srcTgtUseDefaultSimpleToSimpleCode},
			{"simple-to-required", simpleUseDefault, requiredUseDefault, srcTgtUseDefaultSimpleToRequiredCode},
			{"required-to-simple", requiredUseDefault, simpleUseDefault, srcTgtUseDefaultRequiredToSimpleCode},
			{"simple-to-super", simpleUseDefault, superUseDefault, srcTgtUseDefaultSimpleToSuperCode},
			{"super-to-simple", superUseDefault, simpleUseDefault, srcTgtUseDefaultSuperToSimpleCode},
			{"simple-to-default", simpleUseDefault, defaultTUseDefault, srcTgtUseDefaultSimpleToDefaultCode},
			{"default-to-simple", defaultTUseDefault, simpleUseDefault, srcTgtUseDefaultDefaultToSimpleCode},

			// maps
			{"map-to-map", simpleMapUseDefault, simpleMapUseDefault, srcTgtUseDefaultMapToMapCode},
			{"map-to-required-map", simpleMapUseDefault, requiredMapUseDefault, srcTgtUseDefaultMapToRequiredMapCode},
			{"required-map-to-map", requiredMapUseDefault, simpleMapUseDefault, srcTgtUseDefaultRequiredMapToMapCode},
			{"map-to-default-map", simpleMapUseDefault, defaultMapUseDefault, srcTgtUseDefaultMapToDefaultMapCode},
			{"default-map-to-map", defaultMapUseDefault, simpleMapUseDefault, srcTgtUseDefaultDefaultMapToMapCode},
			{"required-map-to-default-map", requiredMapUseDefault, defaultMapUseDefault, srcTgtUseDefaultRequiredMapToDefaultMapCode},
			{"default-map-to-required-map", defaultMapUseDefault, requiredMapUseDefault, srcTgtUseDefaultDefaultMapToRequiredMapCode},
			{"nested-map-to-nested-map", nestedMapUseDefault, nestedMapUseDefault, srcTgtUseDefaultNestedMapToNestedMapCode},
			{"type-map-to-type-map", typeMapUseDefault, typeMapUseDefault, srcTgtUseDefaultTypeMapToTypeMapCode},
			{"array-map-to-array-map", arrayMapUseDefault, arrayMapUseDefault, srcTgtUseDefaultArrayMapToArrayMapCode},

			// arrays
			{"array-to-array", simpleArrayUseDefault, simpleArrayUseDefault, srcTgtUseDefaultArrayToArrayCode},
			{"array-to-required-array", simpleArrayUseDefault, requiredArrayUseDefault, srcTgtUseDefaultArrayToRequiredArrayCode},
			{"required-array-to-array", requiredArrayUseDefault, simpleArrayUseDefault, srcTgtUseDefaultRequiredArrayToArrayCode},
			{"array-to-default-array", simpleArrayUseDefault, defaultArrayUseDefault, srcTgtUseDefaultArrayToDefaultArrayCode},
			{"default-array-to-array", defaultArrayUseDefault, simpleArrayUseDefault, srcTgtUseDefaultDefaultArrayToArrayCode},
			{"required-array-to-default-array", requiredArrayUseDefault, defaultArrayUseDefault, srcTgtUseDefaultRequiredArrayToDefaultArrayCode},
			{"default-array-to-required-array", defaultArrayUseDefault, requiredArrayUseDefault, srcTgtUseDefaultDefaultArrayToRequiredArrayCode},
			{"nested-array-to-nested-array", nestedArrayUseDefault, nestedArrayUseDefault, srcTgtUseDefaultNestedArrayToNestedArrayCode},
			{"type-array-to-type-array", typeArrayUseDefault, typeArrayUseDefault, srcTgtUseDefaultTypeArrayToTypeArrayCode},
			{"map-array-to-map-array", mapArrayUseDefault, mapArrayUseDefault, srcTgtUseDefaultMapArrayToMapArrayCode},

			// others
			{"recursive-to-recursive", recursiveUseDefault, recursiveUseDefault, srcTgtUseDefaultRecursiveToRecursiveCode},
			{"composite-to-custom-field", compositeUseDefault, customFieldUseDefault, srcTgtUseDefaultCompositeToCustomFieldCode},
			{"custom-field-to-composite", customFieldUseDefault, compositeUseDefault, srcTgtUseDefaultCustomFieldToCompositeCode},
			{"composite-to-custom-field-pkg", compositeUseDefault, customFieldPkgUseDefault, srcTgtUseDefaultCompositeToCustomFieldPkgCode},
			{"result-type-to-result-type", resultTypeUseDefault, resultTypeUseDefault, srcTgtUseDefaultResultTypeToResultTypeCode},
			{"result-type-collection-to-result-type-collection", rtColUseDefault, rtColUseDefault, srcTgtUseDefaultRTColToRTColCode},
		},

		// source type uses pointers for all fields, target type uses default
		"source-type-all-ptrs-target-type-uses-default": {
			{"simple-to-simple", simplePointer, simpleUseDefault, srcAllPtrsTgtUseDefaultSimpleToSimpleCode},
			{"simple-to-required", simplePointer, requiredUseDefault, srcAllPtrsTgtUseDefaultSimpleToRequiredCode},
			{"required-to-simple", requiredPointer, simpleUseDefault, srcAllPtrsTgtUseDefaultRequiredToSimpleCode},
			{"simple-to-super", simplePointer, superUseDefault, srcAllPtrsTgtUseDefaultSimpleToSuperCode},
			{"super-to-simple", superPointer, simpleUseDefault, srcAllPtrsTgtUseDefaultSuperToSimpleCode},
			{"simple-to-default", simplePointer, defaultTUseDefault, srcAllPtrsTgtUseDefaultSimpleToDefaultCode},
			{"default-to-simple", defaultTPointer, simpleUseDefault, srcAllPtrsTgtUseDefaultDefaultToSimpleCode},

			// maps
			{"required-map-to-map", requiredMapPointer, simpleMapUseDefault, srcAllPtrsTgtUseDefaultRequiredMapToMapCode},
			{"default-map-to-map", defaultMapPointer, simpleMapUseDefault, srcAllPtrsTgtUseDefaultDefaultMapToMapCode},
			{"required-map-to-default-map", requiredMapPointer, defaultMapUseDefault, srcAllPtrsTgtUseDefaultRequiredMapToDefaultMapCode},
			{"default-map-to-required-map", defaultMapPointer, requiredMapUseDefault, srcAllPtrsTgtUseDefaultDefaultMapToRequiredMapCode},

			// arrays
			{"default-array-to-array", defaultArrayPointer, simpleArrayUseDefault, srcAllPtrsTgtUseDefaultDefaultArrayToArrayCode},
			{"required-array-to-default-array", requiredArrayPointer, defaultArrayUseDefault, srcAllPtrsTgtUseDefaultRequiredArrayToDefaultArrayCode},
			{"default-array-to-required-array", defaultArrayPointer, requiredArrayUseDefault, srcAllPtrsTgtUseDefaultDefaultArrayToRequiredArrayCode},

			// others
			{"custom-field-to-composite", customFieldPointer, compositeUseDefault, srcAllPtrsTgtUseDefaultCustomFieldToCompositeCode},
		},

		// source type uses default, target type uses pointers for all fields
		"source-type-uses-default-target-type-all-ptrs": {
			{"simple-to-simple", simpleUseDefault, simplePointer, srcUseDefaultTgtAllPtrsSimpleToSimpleCode},
			{"simple-to-required", simpleUseDefault, requiredPointer, srcUseDefaultTgtAllPtrsSimpleToRequiredCode},
			{"required-to-simple", requiredUseDefault, simplePointer, srcUseDefaultTgtAllPtrsRequiredToSimpleCode},
			{"simple-to-default", simpleUseDefault, defaultTPointer, srcUseDefaultTgtAllPtrsSimpleToDefaultCode},
			{"default-to-simple", defaultTUseDefault, simplePointer, srcUseDefaultTgtAllPtrsDefaultToSimpleCode},

			// maps
			{"map-to-default-map", simpleMapUseDefault, defaultMapPointer, srcUseDefaultTgtAllPtrsMapToDefaultMapCode},

			// arrays
			{"array-to-default-array", simpleArrayUseDefault, defaultArrayPointer, srcUseDefaultTgtAllPtrsArrayToDefaultArrayCode},

			// others
			{"recursive-to-recursive", recursiveUseDefault, recursivePointer, srcUseDefaultTgtAllPtrsRecursiveToRecursiveCode},
			{"composite-to-custom-field", compositeUseDefault, customFieldPointer, srcUseDefaultTgtAllPtrsCompositeToCustomFieldCode},
		},
	}
	for name, cases := range tc {
		t.Run(name, func(t *testing.T) {
			for _, c := range cases {
				t.Run(c.Name, func(t *testing.T) {
					if c.Source == nil {
						t.Fatal("source type not found in testdata")
					}
					if c.Target == nil {
						t.Fatal("target type not found in testdata")
					}
					code, _, err := GoTransform(c.Source, c.Target, "source", "target", "")
					if err != nil {
						t.Fatal(err)
					}
					code = FormatTestCode(t, "package foo\nfunc transform(){\n"+code+"}")
					if code != c.Code {
						t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code, Diff(t, code, c.Code))
					}
				})
			}
		})
	}
}

const (
	srcTgtUseDefaultSimpleToSimpleCode = `func transform() {
	target := &Simple{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
		Integer:        source.Integer,
	}
}
`

	srcTgtUseDefaultSimpleToRequiredCode = `func transform() {
	target := &Required{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
	}
	if source.Integer != nil {
		target.Integer = *source.Integer
	}
}
`

	srcTgtUseDefaultRequiredToSimpleCode = `func transform() {
	target := &Simple{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
		Integer:        &source.Integer,
	}
}
`

	srcTgtUseDefaultSimpleToSuperCode = `func transform() {
	target := &Super{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
		Integer:        source.Integer,
	}
}
`

	srcTgtUseDefaultSuperToSimpleCode = `func transform() {
	target := &Simple{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
		Integer:        source.Integer,
	}
}
`

	srcTgtUseDefaultSimpleToDefaultCode = `func transform() {
	target := &Default{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
	}
	if source.Integer != nil {
		target.Integer = *source.Integer
	}
	if source.Integer == nil {
		target.Integer = 1
	}
}
`

	srcTgtUseDefaultDefaultToSimpleCode = `func transform() {
	target := &Simple{
		RequiredString: source.RequiredString,
		DefaultBool:    source.DefaultBool,
		Integer:        &source.Integer,
	}
}
`

	srcTgtUseDefaultMapToMapCode = `func transform() {
	target := &SimpleMap{}
	if source.Simple != nil {
		target.Simple = make(map[string]int, len(source.Simple))
		for key, val := range source.Simple {
			tk := key
			tv := val
			target.Simple[tk] = tv
		}
	}
}
`

	srcTgtUseDefaultMapToRequiredMapCode = `func transform() {
	target := &RequiredMap{}
	if source.Simple != nil {
		target.Simple = make(map[string]int, len(source.Simple))
		for key, val := range source.Simple {
			tk := key
			tv := val
			target.Simple[tk] = tv
		}
	}
}
`

	srcTgtUseDefaultRequiredMapToMapCode = `func transform() {
	target := &SimpleMap{}
	target.Simple = make(map[string]int, len(source.Simple))
	for key, val := range source.Simple {
		tk := key
		tv := val
		target.Simple[tk] = tv
	}
}
`

	srcTgtUseDefaultMapToDefaultMapCode = `func transform() {
	target := &DefaultMap{}
	if source.Simple != nil {
		target.Simple = make(map[string]int, len(source.Simple))
		for key, val := range source.Simple {
			tk := key
			tv := val
			target.Simple[tk] = tv
		}
	}
	if source.Simple == nil {
		target.Simple = map[string]int{"foo": 1}
	}
}
`

	srcTgtUseDefaultDefaultMapToMapCode = `func transform() {
	target := &SimpleMap{}
	target.Simple = make(map[string]int, len(source.Simple))
	for key, val := range source.Simple {
		tk := key
		tv := val
		target.Simple[tk] = tv
	}
}
`

	srcTgtUseDefaultRequiredMapToDefaultMapCode = `func transform() {
	target := &DefaultMap{}
	target.Simple = make(map[string]int, len(source.Simple))
	for key, val := range source.Simple {
		tk := key
		tv := val
		target.Simple[tk] = tv
	}
}
`

	srcTgtUseDefaultDefaultMapToRequiredMapCode = `func transform() {
	target := &RequiredMap{}
	target.Simple = make(map[string]int, len(source.Simple))
	for key, val := range source.Simple {
		tk := key
		tv := val
		target.Simple[tk] = tv
	}
}
`

	srcTgtUseDefaultNestedMapToNestedMapCode = `func transform() {
	target := &NestedMap{}
	if source.NestedMap != nil {
		target.NestedMap = make(map[float64]map[int]map[float64]uint64, len(source.NestedMap))
		for key, val := range source.NestedMap {
			tk := key
			tvc := make(map[int]map[float64]uint64, len(val))
			for key, val := range val {
				tk := key
				tvb := make(map[float64]uint64, len(val))
				for key, val := range val {
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

	srcTgtUseDefaultTypeMapToTypeMapCode = `func transform() {
	target := &TypeMap{}
	if source.TypeMap != nil {
		target.TypeMap = make(map[string]*SimpleMap, len(source.TypeMap))
		for key, val := range source.TypeMap {
			tk := key
			tvb := &SimpleMap{}
			if val.Simple != nil {
				tvb.Simple = make(map[string]int, len(val.Simple))
				for key, val := range val.Simple {
					tk := key
					tv := val
					tvb.Simple[tk] = tv
				}
			}
			target.TypeMap[tk] = tvb
		}
	}
}
`

	srcTgtUseDefaultArrayMapToArrayMapCode = `func transform() {
	target := &ArrayMap{}
	if source.ArrayMap != nil {
		target.ArrayMap = make(map[uint32][]float32, len(source.ArrayMap))
		for key, val := range source.ArrayMap {
			tk := key
			tv := make([]float32, len(val))
			for i, val := range val {
				tv[i] = val
			}
			target.ArrayMap[tk] = tv
		}
	}
}
`

	srcTgtUseDefaultArrayToArrayCode = `func transform() {
	target := &SimpleArray{}
	if source.StringArray != nil {
		target.StringArray = make([]string, len(source.StringArray))
		for i, val := range source.StringArray {
			target.StringArray[i] = val
		}
	}
}
`

	srcTgtUseDefaultArrayToRequiredArrayCode = `func transform() {
	target := &RequiredArray{}
	if source.StringArray != nil {
		target.StringArray = make([]string, len(source.StringArray))
		for i, val := range source.StringArray {
			target.StringArray[i] = val
		}
	}
}
`

	srcTgtUseDefaultRequiredArrayToArrayCode = `func transform() {
	target := &SimpleArray{}
	target.StringArray = make([]string, len(source.StringArray))
	for i, val := range source.StringArray {
		target.StringArray[i] = val
	}
}
`

	srcTgtUseDefaultArrayToDefaultArrayCode = `func transform() {
	target := &DefaultArray{}
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

	srcTgtUseDefaultDefaultArrayToArrayCode = `func transform() {
	target := &SimpleArray{}
	target.StringArray = make([]string, len(source.StringArray))
	for i, val := range source.StringArray {
		target.StringArray[i] = val
	}
}
`

	srcTgtUseDefaultRequiredArrayToDefaultArrayCode = `func transform() {
	target := &DefaultArray{}
	target.StringArray = make([]string, len(source.StringArray))
	for i, val := range source.StringArray {
		target.StringArray[i] = val
	}
}
`

	srcTgtUseDefaultDefaultArrayToRequiredArrayCode = `func transform() {
	target := &RequiredArray{}
	target.StringArray = make([]string, len(source.StringArray))
	for i, val := range source.StringArray {
		target.StringArray[i] = val
	}
}
`

	srcTgtUseDefaultNestedArrayToNestedArrayCode = `func transform() {
	target := &NestedArray{}
	if source.NestedArray != nil {
		target.NestedArray = make([][][]float64, len(source.NestedArray))
		for i, val := range source.NestedArray {
			target.NestedArray[i] = make([][]float64, len(val))
			for j, val := range val {
				target.NestedArray[i][j] = make([]float64, len(val))
				for k, val := range val {
					target.NestedArray[i][j][k] = val
				}
			}
		}
	}
}
`

	srcTgtUseDefaultTypeArrayToTypeArrayCode = `func transform() {
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

	srcTgtUseDefaultMapArrayToMapArrayCode = `func transform() {
	target := &MapArray{}
	if source.MapArray != nil {
		target.MapArray = make([]map[int]string, len(source.MapArray))
		for i, val := range source.MapArray {
			target.MapArray[i] = make(map[int]string, len(val))
			for key, val := range val {
				tk := key
				tv := val
				target.MapArray[i][tk] = tv
			}
		}
	}
}
`

	srcTgtUseDefaultRecursiveToRecursiveCode = `func transform() {
	target := &Recursive{
		RequiredString: source.RequiredString,
	}
	if source.Recursive != nil {
		target.Recursive = transformRecursiveToRecursive(source.Recursive)
	}
}
`

	srcTgtUseDefaultCompositeToCustomFieldCode = `func transform() {
	target := &CompositeWithCustomField{}
	if source.RequiredString != nil {
		target.MyString = *source.RequiredString
	}
	if source.DefaultInt != nil {
		target.MyInt = *source.DefaultInt
	}
	if source.DefaultInt == nil {
		target.MyInt = 100
	}
	if source.Type != nil {
		target.MyType = transformSimpleToSimple(source.Type)
	}
	if source.Map != nil {
		target.MyMap = make(map[int]string, len(source.Map))
		for key, val := range source.Map {
			tk := key
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

	srcTgtUseDefaultCustomFieldToCompositeCode = `func transform() {
	target := &Composite{
		RequiredString: &source.MyString,
		DefaultInt:     &source.MyInt,
	}
	target.Type = transformSimpleToSimple(source.MyType)
	target.Map = make(map[int]string, len(source.MyMap))
	for key, val := range source.MyMap {
		tk := key
		tv := val
		target.Map[tk] = tv
	}
	target.Array = make([]string, len(source.MyArray))
	for i, val := range source.MyArray {
		target.Array[i] = val
	}
}
`

	srcTgtUseDefaultCompositeToCustomFieldPkgCode = `func transform() {
	target := &mypkg.CompositeWithCustomField{}
	if source.RequiredString != nil {
		target.MyString = *source.RequiredString
	}
	if source.DefaultInt != nil {
		target.MyInt = *source.DefaultInt
	}
	if source.DefaultInt == nil {
		target.MyInt = 100
	}
	if source.Type != nil {
		target.MyType = transformSimpleToMypkgSimple(source.Type)
	}
	if source.Map != nil {
		target.MyMap = make(map[int]string, len(source.Map))
		for key, val := range source.Map {
			tk := key
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

	srcTgtUseDefaultResultTypeToResultTypeCode = `func transform() {
	target := &ResultType{
		Int: source.Int,
	}
	if source.Map != nil {
		target.Map = make(map[int]string, len(source.Map))
		for key, val := range source.Map {
			tk := key
			tv := val
			target.Map[tk] = tv
		}
	}
}
`

	srcTgtUseDefaultRTColToRTColCode = `func transform() {
	target := &ResultTypeCollection{}
	if source.Collection != nil {
		target.Collection = make([]*ResultType, len(source.Collection))
		for i, val := range source.Collection {
			target.Collection[i] = &ResultType{
				Int: val.Int,
			}
			if val.Map != nil {
				target.Collection[i].Map = make(map[int]string, len(val.Map))
				for key, val := range val.Map {
					tk := key
					tv := val
					target.Collection[i].Map[tk] = tv
				}
			}
		}
	}
}
`

	srcAllPtrsTgtUseDefaultSimpleToSimpleCode = `func transform() {
	target := &Simple{
		RequiredString: *source.RequiredString,
		Integer:        source.Integer,
	}
	if source.DefaultBool != nil {
		target.DefaultBool = *source.DefaultBool
	}
	if source.DefaultBool == nil {
		target.DefaultBool = true
	}
}
`

	srcAllPtrsTgtUseDefaultSimpleToRequiredCode = `func transform() {
	target := &Required{
		RequiredString: *source.RequiredString,
	}
	if source.DefaultBool != nil {
		target.DefaultBool = *source.DefaultBool
	}
	if source.Integer != nil {
		target.Integer = *source.Integer
	}
	if source.DefaultBool == nil {
		target.DefaultBool = true
	}
}
`

	srcAllPtrsTgtUseDefaultRequiredToSimpleCode = `func transform() {
	target := &Simple{
		RequiredString: *source.RequiredString,
		DefaultBool:    *source.DefaultBool,
		Integer:        source.Integer,
	}
	if source.DefaultBool == nil {
		target.DefaultBool = true
	}
}
`

	srcAllPtrsTgtUseDefaultSimpleToSuperCode = `func transform() {
	target := &Super{
		RequiredString: *source.RequiredString,
		Integer:        source.Integer,
	}
	if source.DefaultBool != nil {
		target.DefaultBool = *source.DefaultBool
	}
	if source.DefaultBool == nil {
		target.DefaultBool = true
	}
}
`

	srcAllPtrsTgtUseDefaultSuperToSimpleCode = `func transform() {
	target := &Simple{
		RequiredString: *source.RequiredString,
		Integer:        source.Integer,
	}
	if source.DefaultBool != nil {
		target.DefaultBool = *source.DefaultBool
	}
	if source.DefaultBool == nil {
		target.DefaultBool = true
	}
}
`

	srcAllPtrsTgtUseDefaultSimpleToDefaultCode = `func transform() {
	target := &Default{
		RequiredString: *source.RequiredString,
	}
	if source.DefaultBool != nil {
		target.DefaultBool = *source.DefaultBool
	}
	if source.Integer != nil {
		target.Integer = *source.Integer
	}
	if source.RequiredString == nil {
		target.RequiredString = "foo"
	}
	if source.DefaultBool == nil {
		target.DefaultBool = true
	}
	if source.Integer == nil {
		target.Integer = 1
	}
}
`

	srcAllPtrsTgtUseDefaultDefaultToSimpleCode = `func transform() {
	target := &Simple{
		Integer: source.Integer,
	}
	if source.RequiredString != nil {
		target.RequiredString = *source.RequiredString
	}
	if source.DefaultBool != nil {
		target.DefaultBool = *source.DefaultBool
	}
	if source.DefaultBool == nil {
		target.DefaultBool = true
	}
}
`

	srcAllPtrsTgtUseDefaultRequiredMapToMapCode = `func transform() {
	target := &SimpleMap{}
	if source.Simple != nil {
		target.Simple = make(map[string]int, len(source.Simple))
		for key, val := range source.Simple {
			tk := key
			tv := val
			target.Simple[tk] = tv
		}
	}
}
`

	srcAllPtrsTgtUseDefaultDefaultMapToMapCode = `func transform() {
	target := &SimpleMap{}
	if source.Simple != nil {
		target.Simple = make(map[string]int, len(source.Simple))
		for key, val := range source.Simple {
			tk := key
			tv := val
			target.Simple[tk] = tv
		}
	}
}
`

	srcAllPtrsTgtUseDefaultRequiredMapToDefaultMapCode = `func transform() {
	target := &DefaultMap{}
	if source.Simple != nil {
		target.Simple = make(map[string]int, len(source.Simple))
		for key, val := range source.Simple {
			tk := key
			tv := val
			target.Simple[tk] = tv
		}
	}
	if source.Simple == nil {
		target.Simple = map[string]int{"foo": 1}
	}
}
`

	srcAllPtrsTgtUseDefaultDefaultMapToRequiredMapCode = `func transform() {
	target := &RequiredMap{}
	if source.Simple != nil {
		target.Simple = make(map[string]int, len(source.Simple))
		for key, val := range source.Simple {
			tk := key
			tv := val
			target.Simple[tk] = tv
		}
	}
}
`

	srcAllPtrsTgtUseDefaultDefaultArrayToArrayCode = `func transform() {
	target := &SimpleArray{}
	if source.StringArray != nil {
		target.StringArray = make([]string, len(source.StringArray))
		for i, val := range source.StringArray {
			target.StringArray[i] = val
		}
	}
}
`

	srcAllPtrsTgtUseDefaultRequiredArrayToDefaultArrayCode = `func transform() {
	target := &DefaultArray{}
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

	srcAllPtrsTgtUseDefaultDefaultArrayToRequiredArrayCode = `func transform() {
	target := &RequiredArray{}
	if source.StringArray != nil {
		target.StringArray = make([]string, len(source.StringArray))
		for i, val := range source.StringArray {
			target.StringArray[i] = val
		}
	}
}
`

	srcAllPtrsTgtUseDefaultCustomFieldToCompositeCode = `func transform() {
	target := &Composite{
		RequiredString: source.MyString,
		DefaultInt:     source.MyInt,
	}
	if source.MyType != nil {
		target.Type = transformSimpleToSimple(source.MyType)
	}
	if source.MyMap != nil {
		target.Map = make(map[int]string, len(source.MyMap))
		for key, val := range source.MyMap {
			tk := key
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

	srcUseDefaultTgtAllPtrsSimpleToSimpleCode = `func transform() {
	target := &Simple{
		RequiredString: &source.RequiredString,
		DefaultBool:    &source.DefaultBool,
		Integer:        source.Integer,
	}
}
`

	srcUseDefaultTgtAllPtrsSimpleToRequiredCode = `func transform() {
	target := &Required{
		RequiredString: &source.RequiredString,
		DefaultBool:    &source.DefaultBool,
		Integer:        source.Integer,
	}
}
`

	srcUseDefaultTgtAllPtrsRequiredToSimpleCode = `func transform() {
	target := &Simple{
		RequiredString: &source.RequiredString,
		DefaultBool:    &source.DefaultBool,
		Integer:        &source.Integer,
	}
}
`

	srcUseDefaultTgtAllPtrsSimpleToDefaultCode = `func transform() {
	target := &Default{
		RequiredString: &source.RequiredString,
		DefaultBool:    &source.DefaultBool,
		Integer:        source.Integer,
	}
}
`

	srcUseDefaultTgtAllPtrsDefaultToSimpleCode = `func transform() {
	target := &Simple{
		RequiredString: &source.RequiredString,
		DefaultBool:    &source.DefaultBool,
		Integer:        &source.Integer,
	}
}
`

	srcUseDefaultTgtAllPtrsMapToDefaultMapCode = `func transform() {
	target := &DefaultMap{}
	if source.Simple != nil {
		target.Simple = make(map[string]int, len(source.Simple))
		for key, val := range source.Simple {
			tk := key
			tv := val
			target.Simple[tk] = tv
		}
	}
}
`

	srcUseDefaultTgtAllPtrsArrayToDefaultArrayCode = `func transform() {
	target := &DefaultArray{}
	if source.StringArray != nil {
		target.StringArray = make([]string, len(source.StringArray))
		for i, val := range source.StringArray {
			target.StringArray[i] = val
		}
	}
}
`

	srcUseDefaultTgtAllPtrsRecursiveToRecursiveCode = `func transform() {
	target := &Recursive{
		RequiredString: &source.RequiredString,
	}
	if source.Recursive != nil {
		target.Recursive = transformRecursiveToRecursive(source.Recursive)
	}
}
`

	srcUseDefaultTgtAllPtrsCompositeToCustomFieldCode = `func transform() {
	target := &CompositeWithCustomField{
		MyString: source.RequiredString,
		MyInt:    source.DefaultInt,
	}
	if source.Type != nil {
		target.MyType = transformSimpleToSimple(source.Type)
	}
	if source.Map != nil {
		target.MyMap = make(map[int]string, len(source.Map))
		for key, val := range source.Map {
			tk := key
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
)
