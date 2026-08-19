package openapi

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"goa.design/goa/v3/expr"
)

func TestExtensionsFromMethod(t *testing.T) {
	method := &expr.MethodExpr{
		Idempotent: true,
		Meta: expr.MetaExpr{
			"openapi:extension:x-example": {"value"},
		},
	}

	extensions := ExtensionsFromMethod(method)

	assert.Equal(t, true, extensions["x-goa-idempotent"])
	assert.Equal(t, "value", extensions["x-example"])
}
