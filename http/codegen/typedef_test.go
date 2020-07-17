package codegen

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
	"testing"
)

func TestGoTypeDef(t *testing.T) {
	// types to test
	var (
		mixed = &expr.AttributeExpr{
			Type: &expr.Object{
				&expr.NamedAttributeExpr{
					Name:      "required",
					Attribute: &expr.AttributeExpr{Type: expr.String},
				},
				&expr.NamedAttributeExpr{
					Name:      "default",
					Attribute: &expr.AttributeExpr{Type: expr.Int, DefaultValue: 0},
				},
				&expr.NamedAttributeExpr{
					Name:      "optional",
					Attribute: &expr.AttributeExpr{Type: expr.Float32},
				},
				&expr.NamedAttributeExpr{
					Name:      "custom_type",
					Attribute: &expr.AttributeExpr{Type: expr.String, Meta: expr.MetaExpr{"struct:field:type": []string{"pkg.String"}}},
				},
			},
			Validation: &expr.ValidationExpr{
				Required: []string{"required"},
			},
		}
	)

	cases := []struct {
		Name       string
		Attr       *expr.AttributeExpr
		UsePtr     bool
		UseDefault bool
		Def        string
	}{
		{"no-default", mixed, false, false, mixedNoDefault},
		{"use-default", mixed, false, true, mixedUseDefault},
		{"use-pointer", mixed, true, true, mixedUsePointer},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			def := goTypeDef(codegen.NewNameScope(), c.Attr, c.UsePtr, c.UseDefault)
			if def != c.Def {
				t.Errorf("invalid type definition:\ngot:\n%s\n\nexpected:\n%s\n\ndiff:\n%s\n", def, c.Def, codegen.Diff(t, def, c.Def))
			}
		})
	}
}

var (
	mixedNoDefault = `struct {
	Required string ` + "`" + `form:"required" json:"required" xml:"required"` + "`" + `
	Default *int ` + "`" + `form:"default,omitempty" json:"default,omitempty" xml:"default,omitempty"` + "`" + `
	Optional *float32 ` + "`" + `form:"optional,omitempty" json:"optional,omitempty" xml:"optional,omitempty"` + "`" + `
	CustomType *pkg.String ` + "`" + `form:"custom_type,omitempty" json:"custom_type,omitempty" xml:"custom_type,omitempty"` + "`" + `
}`

	mixedUseDefault = `struct {
	Required string ` + "`" + `form:"required" json:"required" xml:"required"` + "`" + `
	Default int ` + "`" + `form:"default" json:"default" xml:"default"` + "`" + `
	Optional *float32 ` + "`" + `form:"optional,omitempty" json:"optional,omitempty" xml:"optional,omitempty"` + "`" + `
	CustomType *pkg.String ` + "`" + `form:"custom_type,omitempty" json:"custom_type,omitempty" xml:"custom_type,omitempty"` + "`" + `
}`

	mixedUsePointer = `struct {
	Required *string ` + "`" + `form:"required,omitempty" json:"required,omitempty" xml:"required,omitempty"` + "`" + `
	Default *int ` + "`" + `form:"default,omitempty" json:"default,omitempty" xml:"default,omitempty"` + "`" + `
	Optional *float32 ` + "`" + `form:"optional,omitempty" json:"optional,omitempty" xml:"optional,omitempty"` + "`" + `
	CustomType *pkg.String ` + "`" + `form:"custom_type,omitempty" json:"custom_type,omitempty" xml:"custom_type,omitempty"` + "`" + `
}`
)
