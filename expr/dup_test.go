package expr

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDupPreservesNonNullableArrayElements(t *testing.T) {
	t.Parallel()

	original := &Array{
		ElemType:         &AttributeExpr{Type: String},
		NonNullableElems: true,
	}

	duplicate, ok := Dup(original).(*Array)
	require.True(t, ok)
	assert.NotSame(t, original, duplicate)
	assert.NotSame(t, original.ElemType, duplicate.ElemType)
	assert.True(t, duplicate.NonNullableElems)
}
