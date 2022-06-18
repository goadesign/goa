package codegen

import (
	"reflect"
	"sort"
	"testing"

	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestGetMetaTypeImports(t *testing.T) {
	testdata := []struct {
		name string
		dsl  func()
		want []string
	}{
		{
			name: "payload-primitive",
			dsl: func() {
				dsl.Method("m", func() {
					dsl.Payload(func() {
						dsl.Attribute("a", dsl.String, func() {
							dsl.Meta("struct:field:type", "CustomTypeString", "package/string")
						})
						dsl.Attribute("b", dsl.Int, func() {
							dsl.Meta("struct:field:type", "CustomTypeInt", "package/int")
						})
					})
				})
			},
			want: []string{
				"package/string",
				"package/int",
			},
		},
		{
			name: "payload-map",
			dsl: func() {
				dsl.Method("m", func() {
					dsl.Payload(func() {
						dsl.Attribute("a", dsl.MapOf(dsl.String, dsl.String, func() {
							dsl.Key(func() {
								dsl.Meta("struct:field:type", "CustomTypeMapKey", "package/map-key")
							})
							dsl.Elem(func() {
								dsl.Meta("struct:field:type", "CustomTypeMapElem", "package/map-elem")
							})
						}))
					})
				})
			},
			want: []string{
				"package/map-elem",
				"package/map-key",
			},
		},
		{
			name: "payload-map-map",
			dsl: func() {
				dsl.Method("m", func() {
					dsl.Payload(func() {
						dsl.Attribute("a", dsl.MapOf(dsl.String, dsl.MapOf(dsl.String, dsl.String, func() {
							dsl.Key(func() {
								dsl.Meta("struct:field:type", "CustomTypeMapKey", "package/map-map-key")
							})
							dsl.Elem(func() {
								dsl.Meta("struct:field:type", "CustomTypeMapElem", "package/map-map-elem")
							})
						}), func() {
							dsl.Key(func() {
								dsl.Meta("struct:field:type", "CustomTypeMapKey", "package/map-key")
							})
							dsl.Elem(func() {
								dsl.Meta("struct:field:type", "CustomTypeMapElem", "package/map-elem")
							})
						}))
					})
				})
			},
			want: []string{
				"package/map-key",
				"package/map-map-elem",
				"package/map-map-key",
				"package/map-elem",
			},
		},
		{
			name: "payload-array",
			dsl: func() {
				dsl.Method("m", func() {
					dsl.Payload(func() {
						dsl.Attribute("a", dsl.ArrayOf(dsl.String, func() {
							dsl.Meta("struct:field:type", "SomeCustomTypeArrayElem", "package/array-elem")
						}), func() {
							dsl.Meta("struct:field:type", "SomeCustomTypeArray", "package/array")
						})
					})
				})
			},
			want: []string{
				"package/array-elem",
				"package/array",
			},
		},
		{
			name: "result",
			dsl: func() {
				dsl.Method("m", func() {
					dsl.Result(func() {
						dsl.Attribute("a", dsl.String, func() {
							dsl.Meta("struct:field:type", "CustomTypeString", "package/result-string")
						})
						dsl.Attribute("b", dsl.ArrayOf(dsl.String, func() {
							dsl.Meta("struct:field:type", "SomeCustomTypeArrayElem", "package/result-array-elem")
						}), func() {
							dsl.Meta("struct:field:type", "SomeCustomTypeArray", "package/result-array")
						})
						dsl.Attribute("a", dsl.MapOf(dsl.String, dsl.String, func() {
							dsl.Key(func() {
								dsl.Meta("struct:field:type", "CustomTypeMapKey", "package/result-map-key")
							})
							dsl.Elem(func() {
								dsl.Meta("struct:field:type", "CustomTypeMapElem", "package/result-map-elem")
							})
						}))
					})
				})
			},
			want: []string{
				"package/result-string",
				"package/result-array",
				"package/result-array-elem",
				"package/result-map-key",
				"package/result-map-elem",
			},
		},
	}
	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			eval.Context = &eval.DSLContext{}
			serviceExpr := &expr.ServiceExpr{}
			eval.Execute(tt.dsl, serviceExpr)
			if eval.Context.Errors != nil {
				t.Fatalf("%s: Service DSL failed unexpectedly with %s", tt.name, eval.Context.Errors)
			}
			for _, methodExpr := range serviceExpr.Methods {
				eval.Execute(methodExpr.DSLFunc, methodExpr)
				if eval.Context.Errors != nil {
					t.Fatalf("%s: Method DSL failed unexpectedly with %s", tt.name, eval.Context.Errors)
				}
			}
			for _, methodExpr := range serviceExpr.Methods {
				var got []string
				for _, v := range GetMetaTypeImports(methodExpr.Payload) {
					got = append(got, v.Path)
				}
				for _, v := range GetMetaTypeImports(methodExpr.Result) {
					got = append(got, v.Path)
				}
				sort.Strings(got)
				sort.Strings(tt.want)
				if !reflect.DeepEqual(tt.want, got) {
					t.Errorf("want %+v, got %+v", tt.want, got)
				}
			}
		})
	}
}
