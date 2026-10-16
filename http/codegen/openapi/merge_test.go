// This file checks that merging and copying schemas preserves their constraints
// in the documents returned to generator callers.
package openapi

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaMergePreservesMapBounds(t *testing.T) {
	cases := []struct {
		name  string
		first string
		next  string
		want  string
	}{
		{
			name:  "missing",
			first: `{}`,
			next:  `{"minProperties":1,"maxProperties":2}`,
			want:  `{"minProperties":1,"maxProperties":2}`,
		},
		{
			name:  "zero",
			first: `{}`,
			next:  `{"minProperties":0,"maxProperties":0}`,
			want:  `{"minProperties":0,"maxProperties":0}`,
		},
		{
			name:  "unchanged",
			first: `{"minProperties":1,"maxProperties":2}`,
			next:  `{}`,
			want:  `{"minProperties":1,"maxProperties":2}`,
		},
		{
			name:  "wider",
			first: `{"minProperties":1,"maxProperties":2}`,
			next:  `{"minProperties":0,"maxProperties":3}`,
			want:  `{"minProperties":0,"maxProperties":3}`,
		},
		{
			name:  "narrower",
			first: `{"minProperties":0,"maxProperties":3}`,
			next:  `{"minProperties":1,"maxProperties":2}`,
			want:  `{"minProperties":0,"maxProperties":3}`,
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			var first, next Schema
			require.NoError(t, json.Unmarshal([]byte(test.first), &first))
			require.NoError(t, json.Unmarshal([]byte(test.next), &next))
			first.Merge(&next)
			data, err := json.Marshal(first.Dup())
			require.NoError(t, err)
			require.JSONEq(t, test.want, string(data))
		})
	}
}

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
