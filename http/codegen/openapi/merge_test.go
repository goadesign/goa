package openapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaMergeAndDupPreserveAnyOf(t *testing.T) {
	source := &Schema{
		AnyOf: []*Schema{
			{Type: Type(String)},
			{Type: Type(Integer)},
		},
	}
	merged := NewSchema()
	merged.Merge(source)

	require.Len(t, merged.AnyOf, 2)
	assert.Equal(t, Type(String), merged.AnyOf[0].Type)

	dup := merged.Dup()
	require.Len(t, dup.AnyOf, 2)
	assert.Equal(t, Type(Integer), dup.AnyOf[1].Type)
	assert.NotSame(t, merged.AnyOf[0], dup.AnyOf[0])
}
