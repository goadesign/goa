// This file checks the validation generated for protobuf OneOf values.
// It rejects a missing selected branch and a selected branch whose value is nil.
package codegen

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	d "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

type (
	// protobufUnionTestScope returns the Go names used by the generated struct for
	// each selected branch. This lets the test run without generating a service.
	protobufUnionTestScope struct {
		scope *NameScope
	}
)

func TestProtobufUnionValidationRequiresCompleteSelectedBranch(t *testing.T) {
	root := RunDSL(t, protobufUnionValidationDSL)
	message := root.UserType("Message")
	ctx := NewAttributeContext(false, true, false, "pb", NewNameScope())
	ctx.Scope = &protobufUnionTestScope{scope: NewNameScope()}

	generated := AttributeValidationCode(message.Attribute(), message, ctx, true, false, "message", "message")

	require.Contains(t, generated, `goa.MissingFieldError("choice", "message")`)
	require.Contains(t, generated, `goa.MissingFieldError("detail", "message.choice")`)
	require.Contains(t, generated, `goa.MissingFieldError("inactive", "message.choice")`)
	require.Contains(t, generated, `goa.MissingFieldError("blob", "message.choice")`)
	require.Contains(t, generated, `goa.MissingFieldError("metadata", "message.choice")`)
	require.Contains(t, generated, "if v == nil {")
	require.Contains(t, generated, "if v.Detail == nil {")
	require.Contains(t, generated, "if v.Inactive == nil {")
	require.Contains(t, generated, "if v.Metadata == nil {")
	require.Contains(t, generated, "if v.Blob == nil {")
	require.NotContains(t, generated, "if v.Token == nil {")
}

func (s *protobufUnionTestScope) Name(att *expr.AttributeExpr, pkg string, _, _ bool) string {
	name := Goify(att.Type.Name(), true)
	if pkg != "" {
		return pkg + "." + name
	}
	return name
}

func (s *protobufUnionTestScope) Ref(att *expr.AttributeExpr, pkg string) string {
	return "*" + s.Name(att, pkg, false, false)
}

func (*protobufUnionTestScope) Field(_ *expr.AttributeExpr, name string, firstUpper bool) string {
	return Goify(name, firstUpper)
}

func (*protobufUnionTestScope) Package(*expr.AttributeExpr) string {
	return "pb"
}

func (s *protobufUnionTestScope) Enter(*expr.AttributeExpr) Attributor {
	return s
}

func (*protobufUnionTestScope) IsSumType() bool {
	return false
}

func (s *protobufUnionTestScope) ValidatorCall(att *expr.AttributeExpr, view, target, _ string) string {
	name := "Validate" + s.Name(att, "", false, false) + Goify(view, true)
	return fmt.Sprintf("%s(%s)", name, target)
}

func (s *protobufUnionTestScope) Scope() *NameScope {
	return s.scope
}

// protobufUnionValidationDSL creates OneOf branches stored as pointers,
// scalars, and byte slices.
func protobufUnionValidationDSL() {
	token := d.Type("Token", d.String)
	detail := d.Type("Detail", func() {
		d.Attribute("label", d.String)
		d.Required("label")
	})
	inactive := d.Type("Inactive", func() {})
	d.Type("Message", func() {
		d.OneOf("choice", func() {
			d.Attribute("number", d.Int, func() { d.Minimum(1) })
			d.Attribute("detail", detail)
			d.Attribute("inactive", inactive)
			d.Attribute("blob", d.Bytes)
			d.Attribute("token", token)
			d.Attribute("metadata", d.Any)
		})
		d.Required("choice")
	})
}
