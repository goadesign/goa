// These tests evaluate flattened OneOf declarations through the ordinary DSL.
// Invalid object mappings must stop generation; ordinary unions keep their
// existing scalar branches and value-field metadata.
package dsl_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	. "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestOneOfFlattenMapping(t *testing.T) {
	for _, test := range []struct {
		name          string
		branch        func()
		discriminator string
		meta          func()
		failure       string
	}{
		{name: "object", branch: func() { Attribute("reference", String) }},
		{name: "mapped collision", branch: func() { Attribute("kind:resultType", String) }, failure: "discriminator JSON name"},
		{name: "complete JSON tag collision", branch: func() { Attribute("kind", String, func() { Meta("struct:tag:json", "resultType,omitempty") }) }, failure: "discriminator JSON name"},
		{name: "service JSON tag collision", branch: func() { Attribute("kind", String, func() { Meta("struct:tag:json:name", "resultType") }) }, failure: "discriminator JSON name"},
		{name: "service field collision", discriminator: "Reference", branch: func() { Attribute("reference", String) }, failure: "discriminator JSON name"},
		{name: "renamed service field collision", discriminator: "Choice", branch: func() { Attribute("reference", String, func() { Meta("struct:field:name", "Choice") }) }, failure: "discriminator JSON name"},
		{name: "complete tag selects independent name", branch: func() { Attribute("resultType", String, func() { Meta("struct:tag:json", "reference") }) }},
		{name: "ignored field", branch: func() { Attribute("resultType", String, func() { Meta("struct:tag:json", "-") }) }},
		{name: "value field conflict", branch: func() { Attribute("reference", String) }, meta: func() { Meta("oneof:value:field", "data") }, failure: "cannot be combined"},
		{name: "flag with value", branch: func() { Attribute("reference", String) }, meta: func() { Meta("oneof:json:flatten", "true") }, failure: "takes no values"},
	} {
		t.Run(test.name, func(t *testing.T) {
			eval.Context = &eval.DSLContext{}
			declaration := &expr.UserTypeExpr{TypeName: "Outcome", AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{}}}
			eval.Execute(func() {
				OneOf("outcome", func() {
					Meta("oneof:json:flatten")
					discriminator := test.discriminator
					if discriminator == "" {
						discriminator = "resultType"
					}
					Meta("oneof:type:field", discriminator)
					if test.meta != nil {
						test.meta()
					}
					Attribute("complete", test.branch)
				})
			}, declaration)
			if eval.Context.Errors != nil {
				require.NotEmpty(t, test.failure)
				require.Contains(t, eval.Context.Errors.Error(), test.failure)
				return
			}
			failure := declaration.Attribute().Validate("", declaration)
			if test.failure != "" {
				require.NotNil(t, failure)
				require.Contains(t, failure.Error(), test.failure)
				return
			}
			require.Empty(t, failure.Errors)
			union := expr.AsUnion(declaration.Attribute().Find("outcome").Type)
			require.True(t, union.Flatten)
			require.True(t, expr.AsUnion(expr.DupAtt(declaration.Attribute().Find("outcome")).Type).Flatten)
			copier := expr.NewAttributeGraphCopier()
			require.True(t, expr.AsUnion(copier.Copy(declaration.Attribute().Find("outcome")).Type).Flatten)
		})
	}
}

func TestOneOfFlattenRejectsNonObjectBranches(t *testing.T) {
	for _, branch := range []any{String, ArrayOf(String), MapOf(String, String)} {
		eval.Context = &eval.DSLContext{}
		declaration := &expr.UserTypeExpr{TypeName: "Outcome", AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{}}}
		eval.Execute(func() {
			OneOf("outcome", func() {
				Meta("oneof:json:flatten")
				Attribute("complete", branch)
			})
		}, declaration)
		failure := declaration.Attribute().Validate("", declaration)
		require.NotNil(t, failure)
		require.Contains(t, failure.Error(), "must be an object")
	}
}
