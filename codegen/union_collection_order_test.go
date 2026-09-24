package codegen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

// TestUnionCollectionHelperOrder compiles and executes both branch orders.
// Distinct element types detect a consumed-but-swapped helper call even when
// the render cursor has consumed every planned occurrence.
func TestUnionCollectionHelperOrder(t *testing.T) {
	for _, collection := range []string{"map", "slice"} {
		for _, nesting := range []string{"flat", "array", "union"} {
			for _, directFirst := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/direct-first=%t", collection, nesting, directFirst), func(t *testing.T) {
					shape := func(prefix string) (*expr.AttributeExpr, string) {
						var declarations strings.Builder
						named := func(name string, element expr.DataType) *expr.AttributeExpr {
							var underlying expr.DataType = &expr.Array{ElemType: &expr.AttributeExpr{Type: element}}
							if collection == "map" {
								underlying = &expr.Map{
									KeyType:  &expr.AttributeExpr{Type: expr.String},
									ElemType: &expr.AttributeExpr{Type: element},
								}
							}
							named := &expr.UserTypeExpr{
								TypeName:      prefix + name,
								AttributeExpr: &expr.AttributeExpr{Type: underlying},
							}
							fmt.Fprintf(&declarations, "type %s %s\n", named.Name(), NewNameScope().GoTypeDef(named.Attribute(), false, true))
							return &expr.AttributeExpr{Type: named}
						}
						union := func(name string, inline, direct *expr.AttributeExpr) *expr.AttributeExpr {
							branches := []*expr.NamedAttributeExpr{
								{Name: "inline", Attribute: inline},
								{Name: "direct", Attribute: direct},
							}
							if directFirst {
								branches[0], branches[1] = branches[1], branches[0]
							}
							scope := NewNameScope()
							fmt.Fprintf(&declarations, `type %[1]s struct { kind string; inline %[2]s; direct %[3]s }
func (v *%[1]s) Kind() string { return v.kind }
func (v *%[1]s) AsInline() (%[2]s, bool) { return v.inline, v.kind == "inline" }
func (v *%[1]s) AsDirect() (%[3]s, bool) { return v.direct, v.kind == "direct" }
func (v *%[1]s) SetInline(value %[2]s) { v.kind = "inline"; v.inline = value }
func (v *%[1]s) SetDirect(value %[3]s) { v.kind = "direct"; v.direct = value }
`, prefix+name, scope.GoTypeRef(inline), scope.GoTypeRef(direct))
							return &expr.AttributeExpr{Type: &expr.Union{TypeName: prefix + name, Values: branches}}
						}
						inline := &expr.AttributeExpr{Type: &expr.Array{ElemType: named("First", expr.String)}}
						direct := named("Second", expr.Int)
						switch nesting {
						case "array":
							inline = &expr.AttributeExpr{Type: &expr.Array{ElemType: inline}}
						case "union":
							inline = union("Inner", inline, direct)
							direct = named("Third", expr.Boolean)
						}
						result := union("Choice", inline, direct)
						return result, declarations.String()
					}
					source, sourceTypes := shape("Source")
					target, targetTypes := shape("Target")
					context := NewAttributeContext(false, false, true, "", NewNameScope())
					code, helpers, err := GoTransform(source, target, "source", "target", context, context, "", true)
					require.NoError(t, err) // Includes complete planned-call consumption.
					expectedHelpers := 2
					if nesting == "union" {
						expectedHelpers = 3
					}
					require.Len(t, helpers, expectedHelpers)
					repeated, repeatedHelpers, err := GoTransform(source, target, "source", "target", context, context, "", true)
					require.NoError(t, err)
					require.Equal(t, code, repeated)
					require.Equal(t, helpers, repeatedHelpers)

					var generated strings.Builder
					generated.WriteString("package fixture\n")
					generated.WriteString(sourceTypes)
					generated.WriteString(targetTypes)
					fmt.Fprintf(&generated, "func convert(source *SourceChoice) *TargetChoice {\n%s\nreturn target\n}\n", code)
					for _, helper := range helpers {
						fmt.Fprintf(&generated, "func %s(v %s) %s {\n%s\nreturn res\n}\n",
							helper.Name, helper.ParamTypeRef, helper.ResultTypeRef, helper.Code)
					}

					first, second, third := `First{"text"}`, `Second{7}`, `Third{true}`
					if collection == "map" {
						first, second, third = `First{"key":"text"}`, `Second{"key":7}`, `Third{"key":true}`
					}
					// Replace the explicit synthetic prefix in each typed value,
					// retaining nil and empty values in both source and target.
					inline := "[]PREFIXFirst{PREFIXFirst{}, PREFIX" + first + "}"
					emptyInline := "[]PREFIXFirst{}"
					nilInline := "nil"
					direct, emptyDirect := "PREFIX"+second, "PREFIXSecond{}"
					var extra string
					switch nesting {
					case "array":
						inline = "[][]PREFIXFirst{" + inline + ", {}}"
						emptyInline = "[][]PREFIXFirst{}"
					case "union":
						extra = `&PREFIXChoice{kind:"inline", inline:&PREFIXInner{kind:"direct", direct:PREFIX` + second + `}}`
						inline = `&PREFIXInner{kind:"inline", inline:` + inline + "}"
						emptyInline = `&PREFIXInner{kind:"inline", inline:` + emptyInline + "}"
						nilInline = `&PREFIXInner{kind:"inline"}`
						direct, emptyDirect = "PREFIX"+third, "PREFIXThird{}"
					}
					values := []string{
						`&PREFIXChoice{kind:"inline", inline:` + inline + "}",
						`&PREFIXChoice{kind:"inline", inline:` + emptyInline + "}",
						`&PREFIXChoice{kind:"inline", inline:` + nilInline + "}",
						`&PREFIXChoice{kind:"direct", direct:` + direct + "}",
						`&PREFIXChoice{kind:"direct", direct:` + emptyDirect + "}",
						`&PREFIXChoice{kind:"direct"}`,
					}
					if extra != "" {
						values = append(values, extra)
					}
					var checks strings.Builder
					checks.WriteString("package fixture\nimport(\"reflect\";\"testing\")\nfunc TestConvert(t *testing.T) {\n")
					checks.WriteString("tests := []struct { source *SourceChoice; target *TargetChoice }{\n")
					for _, value := range values {
						fmt.Fprintf(&checks, "{%s, %s},\n", strings.ReplaceAll(value, "PREFIX", "Source"), strings.ReplaceAll(value, "PREFIX", "Target"))
					}
					checks.WriteString("}\nfor i, test := range tests {\nif got := convert(test.source); !reflect.DeepEqual(got, test.target) { t.Errorf(\"case %d: got %#v, want %#v\", i, got, test.target) }\n}\n}\n")
					runCollectionFixture(t, map[string]string{
						"convert.go":      generated.String(),
						"convert_test.go": checks.String(),
					})
				})
			}
		}
	}
}
