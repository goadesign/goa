package codegen

import (
	"testing"

	"goa.design/goa/v3/codegen/testutil"

	"github.com/stretchr/testify/require"

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
	}{
		// test cases to transform service type to protocol buffer type
		"to-protobuf-type": {
			{"primitive-to-primitive", primitive, primitive, true, svcCtx},
			{"simple-to-simple", simple, simple, true, svcCtx},
			{"simple-to-required", simple, required, true, svcCtx},
			{"required-to-simple", required, simple, true, svcCtx},
			{"simple-to-default", simple, defaultT, true, svcCtx},
			{"default-to-simple", defaultT, simple, true, svcCtx},
			{"required-ptr-to-simple", required, simple, true, ptrCtx},
			{"simple-to-customtype", customtype, simple, true, svcCtx},
			{"customtype-to-customtype", customtype, customtype, true, svcCtx},

			// maps
			{"map-to-map", simpleMap, simpleMap, true, svcCtx},
			{"nested-map-to-nested-map", nestedMap, nestedMap, true, svcCtx},
			{"array-map-to-array-map", arrayMap, arrayMap, true, svcCtx},
			{"default-map-to-default-map", defaultMap, defaultMap, true, svcCtx},

			// arrays
			{"array-to-array", simpleArray, simpleArray, true, svcCtx},
			{"nested-array-to-nested-array", nestedArray, nestedArray, true, svcCtx},
			{"type-array-to-type-array", typeArray, typeArray, true, svcCtx},
			{"map-array-to-map-array", mapArray, mapArray, true, svcCtx},
			{"default-array-to-default-array", defaultArray, defaultArray, true, svcCtx},

			{"recursive-to-recursive", recursive, recursive, true, svcCtx},
			{"composite-to-custom-field", composite, customField, true, svcCtx},
			{"custom-field-to-composite", customField, composite, true, svcCtx},
			{"result-type-to-result-type", resultType, resultType, true, svcCtx},
			{"result-type-collection-to-result-type-collection", rtCol, rtCol, true, svcCtx},
			{"optional-to-optional", optional, optional, true, svcCtx},
			{"defaults-to-defaults", defaults, defaults, true, svcCtx},

			// oneofs
			{"oneof-to-oneof", simpleOneOf, simpleOneOf, true, svcCtx},
			{"embedded-oneof-to-embedded-oneof", embeddedOneOf, embeddedOneOf, true, svcCtx},
			{"recursive-oneof-to-recursive-oneof", recursiveOneOf, recursiveOneOf, true, svcCtx},

			// package override
			{"pkg-override-to-pkg-override", pkgOverride, pkgOverride, true, svcCtx},
		},

		// test cases to transform protocol buffer type to service type
		"to-service-type": {
			{"primitive-to-primitive", primitive, primitive, false, svcCtx},
			{"simple-to-simple", simple, simple, false, svcCtx},
			{"simple-to-required", simple, required, false, svcCtx},
			{"required-to-simple", required, simple, false, svcCtx},
			{"simple-to-default", simple, defaultT, false, svcCtx},
			{"default-to-simple", defaultT, simple, false, svcCtx},
			{"simple-to-required-ptr", simple, required, false, ptrCtx},
			{"simple-to-customtype", simple, customtype, false, svcCtx},
			{"customtype-to-customtype", customtype, customtype, false, svcCtx},

			// maps
			{"map-to-map", simpleMap, simpleMap, false, svcCtx},
			{"nested-map-to-nested-map", nestedMap, nestedMap, false, svcCtx},
			{"array-map-to-array-map", arrayMap, arrayMap, false, svcCtx},
			{"default-map-to-default-map", defaultMap, defaultMap, false, svcCtx},

			// arrays
			{"array-to-array", simpleArray, simpleArray, false, svcCtx},
			{"nested-array-to-nested-array", nestedArray, nestedArray, false, svcCtx},
			{"type-array-to-type-array", typeArray, typeArray, false, svcCtx},
			{"map-array-to-map-array", mapArray, mapArray, false, svcCtx},
			{"default-array-to-default-array", defaultArray, defaultArray, false, svcCtx},

			{"recursive-to-recursive", recursive, recursive, false, svcCtx},
			{"composite-to-custom-field", composite, customField, false, svcCtx},
			{"custom-field-to-composite", customField, composite, false, svcCtx},
			{"result-type-to-result-type", resultType, resultType, false, svcCtx},
			{"result-type-collection-to-result-type-collection", rtCol, rtCol, false, svcCtx},
			{"optional-to-optional", optional, optional, false, svcCtx},
			{"defaults-to-defaults", defaults, defaults, false, svcCtx},

			// oneofs
			{"oneof-to-oneof", simpleOneOf, simpleOneOf, false, svcCtx},
			{"embedded-oneof-to-embedded-oneof", embeddedOneOf, embeddedOneOf, false, svcCtx},
			{"recursive-oneof-to-recursive-oneof", recursiveOneOf, recursiveOneOf, false, svcCtx},

			// package override
			{"pkg-override-to-pkg-override", pkgOverride, pkgOverride, false, svcCtx},
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
					require.NoError(t, err)
					code = codegen.FormatTestCode(t, "package foo\nfunc transform(){\n"+code+"}")
					testutil.AssertGo(t, "testdata/golden/protobuf_type_encode_"+name+"_"+c.Name+".go.golden", code)
				})
			}
		})
	}
}

func TestProtoBufTransformAnyType(t *testing.T) {
	var (
		sd     = &ServiceData{Name: "Service", Scope: codegen.NewNameScope()}
		svcCtx = codegen.NewAttributeContext(false, false, true, "", sd.Scope)
		pbCtx  = protoBufTypeContext("", sd.Scope, false)
	)

	cases := []struct {
		Name    string
		ToProto bool
		NewVar  bool
		Ctx     *codegen.AttributeContext
	}{
		{"any-to-proto-new-var", true, true, svcCtx},
		{"proto-to-any-new-var", false, true, svcCtx},
		{"any-to-proto-assign-var", true, false, svcCtx},
		{"proto-to-any-assign-var", false, false, svcCtx},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			source := &expr.AttributeExpr{Type: expr.Any}
			target := &expr.AttributeExpr{Type: expr.Any}
			srcCtx := c.Ctx
			tgtCtx := c.Ctx
			if c.ToProto {
				tgtCtx = pbCtx
			} else {
				srcCtx = pbCtx
			}
			code, _, err := protoBufTransform(source, target, "source", "target", srcCtx, tgtCtx, c.ToProto, c.NewVar)
			require.NoError(t, err)
			t.Logf("Generated code: %s", code)

			// Check if transformation contains Any type conversion logic
			if c.ToProto {
				require.Contains(t, code, "func() *structpb.Value", "To proto conversion should generate Value type conversion function")
			} else {
				require.Contains(t, code, "func() any", "From proto conversion should generate any type conversion function")
			}

			// Check the assignment operator used based on newVar parameter
			if c.NewVar {
				require.Contains(t, code, "target :=", "New variable should use := operator")
			} else {
				require.Contains(t, code, "target =", "Assignment should use = operator")
				require.NotContains(t, code, "target :=", "Assignment should not use := operator")
			}
		})
	}
}

func pointerContext(pkg string, scope *codegen.NameScope) *codegen.AttributeContext {
	return codegen.NewAttributeContext(true, false, true, pkg, scope)
}
