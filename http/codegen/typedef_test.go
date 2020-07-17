package codegen

import (
	"testing"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
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
					Name:      "bytes",
					Attribute: &expr.AttributeExpr{Type: expr.Bytes},
				},
				&expr.NamedAttributeExpr{
					Name:      "any",
					Attribute: &expr.AttributeExpr{Type: expr.Any},
				},
				&expr.NamedAttributeExpr{
					Name:      "required_bytes",
					Attribute: &expr.AttributeExpr{Type: expr.Bytes},
				},
				&expr.NamedAttributeExpr{
					Name:      "required_any",
					Attribute: &expr.AttributeExpr{Type: expr.Any},
				},
				&expr.NamedAttributeExpr{
					Name:      "default_bytes",
					Attribute: &expr.AttributeExpr{Type: expr.Bytes, DefaultValue: []byte("foo")},
				},
				&expr.NamedAttributeExpr{
					Name:      "default_any",
					Attribute: &expr.AttributeExpr{Type: expr.Any, DefaultValue: "foo"},
				},
				&expr.NamedAttributeExpr{
					Name:      "custom_type",
					Attribute: &expr.AttributeExpr{Type: expr.String, Meta: expr.MetaExpr{"struct:field:type": []string{"pkg.String"}}},
				},
				&expr.NamedAttributeExpr{
					Name:      "custom_tag",
					Attribute: &expr.AttributeExpr{Type: expr.String, Meta: expr.MetaExpr{"struct:tag:foo": []string{"bar"}}},
				},
			},
			Validation: &expr.ValidationExpr{
				Required: []string{"required", "required_bytes", "required_any"},
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
	Bytes []byte ` + "`" + `form:"bytes,omitempty" json:"bytes,omitempty" xml:"bytes,omitempty"` + "`" + `
	Any interface{} ` + "`" + `form:"any,omitempty" json:"any,omitempty" xml:"any,omitempty"` + "`" + `
	RequiredBytes []byte ` + "`" + `form:"required_bytes" json:"required_bytes" xml:"required_bytes"` + "`" + `
	RequiredAny interface{} ` + "`" + `form:"required_any" json:"required_any" xml:"required_any"` + "`" + `
	DefaultBytes []byte ` + "`" + `form:"default_bytes,omitempty" json:"default_bytes,omitempty" xml:"default_bytes,omitempty"` + "`" + `
	DefaultAny interface{} ` + "`" + `form:"default_any,omitempty" json:"default_any,omitempty" xml:"default_any,omitempty"` + "`" + `
	CustomType *pkg.String ` + "`" + `form:"custom_type,omitempty" json:"custom_type,omitempty" xml:"custom_type,omitempty"` + "`" + `
	CustomTag *string ` + "`" + `foo:"bar"` + "`" + `
}`

	mixedUseDefault = `struct {
	Required string ` + "`" + `form:"required" json:"required" xml:"required"` + "`" + `
	Default int ` + "`" + `form:"default" json:"default" xml:"default"` + "`" + `
	Optional *float32 ` + "`" + `form:"optional,omitempty" json:"optional,omitempty" xml:"optional,omitempty"` + "`" + `
	Bytes []byte ` + "`" + `form:"bytes,omitempty" json:"bytes,omitempty" xml:"bytes,omitempty"` + "`" + `
	Any interface{} ` + "`" + `form:"any,omitempty" json:"any,omitempty" xml:"any,omitempty"` + "`" + `
	RequiredBytes []byte ` + "`" + `form:"required_bytes" json:"required_bytes" xml:"required_bytes"` + "`" + `
	RequiredAny interface{} ` + "`" + `form:"required_any" json:"required_any" xml:"required_any"` + "`" + `
	DefaultBytes []byte ` + "`" + `form:"default_bytes" json:"default_bytes" xml:"default_bytes"` + "`" + `
	DefaultAny interface{} ` + "`" + `form:"default_any" json:"default_any" xml:"default_any"` + "`" + `
	CustomType *pkg.String ` + "`" + `form:"custom_type,omitempty" json:"custom_type,omitempty" xml:"custom_type,omitempty"` + "`" + `
	CustomTag *string ` + "`" + `foo:"bar"` + "`" + `
}`

	mixedUsePointer = `struct {
	Required *string ` + "`" + `form:"required,omitempty" json:"required,omitempty" xml:"required,omitempty"` + "`" + `
	Default *int ` + "`" + `form:"default,omitempty" json:"default,omitempty" xml:"default,omitempty"` + "`" + `
	Optional *float32 ` + "`" + `form:"optional,omitempty" json:"optional,omitempty" xml:"optional,omitempty"` + "`" + `
	Bytes []byte ` + "`" + `form:"bytes,omitempty" json:"bytes,omitempty" xml:"bytes,omitempty"` + "`" + `
	Any interface{} ` + "`" + `form:"any,omitempty" json:"any,omitempty" xml:"any,omitempty"` + "`" + `
	RequiredBytes []byte ` + "`" + `form:"required_bytes,omitempty" json:"required_bytes,omitempty" xml:"required_bytes,omitempty"` + "`" + `
	RequiredAny interface{} ` + "`" + `form:"required_any,omitempty" json:"required_any,omitempty" xml:"required_any,omitempty"` + "`" + `
	DefaultBytes []byte ` + "`" + `form:"default_bytes,omitempty" json:"default_bytes,omitempty" xml:"default_bytes,omitempty"` + "`" + `
	DefaultAny interface{} ` + "`" + `form:"default_any,omitempty" json:"default_any,omitempty" xml:"default_any,omitempty"` + "`" + `
	CustomType *pkg.String ` + "`" + `form:"custom_type,omitempty" json:"custom_type,omitempty" xml:"custom_type,omitempty"` + "`" + `
	CustomTag *string ` + "`" + `foo:"bar"` + "`" + `
}`
)
