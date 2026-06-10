package openapi

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func TestTagsFromExpr(t *testing.T) {
	meta := expr.MetaExpr{
		"openapi:tag:Users":         {"Users"},
		"openapi:tag:Users:desc":    {"User management"},
		"openapi:tag:Users:summary": {"Users"},
		"openapi:tag:Users:parent":  {"Accounts"},
		"openapi:tag:Users:kind":    {"nav"},
		"openapi:tag:Users:url":     {"https://example.com/users"},
	}

	t.Run("3.0 ignores 3.2 tag fields", func(t *testing.T) {
		tags := TagsFromExpr(meta, Version30)
		require.Len(t, tags, 1)
		tag := tags[0]
		require.Equal(t, "Users", tag.Name)
		require.Equal(t, "User management", tag.Description)
		require.NotNil(t, tag.ExternalDocs)
		require.Equal(t, "https://example.com/users", tag.ExternalDocs.URL)
		require.Empty(t, tag.Summary)
		require.Empty(t, tag.Parent)
		require.Empty(t, tag.Kind)
	})

	t.Run("3.2 extracts 3.2 tag fields", func(t *testing.T) {
		tags := TagsFromExpr(meta, Version32)
		require.Len(t, tags, 1)
		tag := tags[0]
		require.Equal(t, "Users", tag.Name)
		require.Equal(t, "User management", tag.Description)
		require.Equal(t, "Users", tag.Summary)
		require.Equal(t, "Accounts", tag.Parent)
		require.Equal(t, "nav", tag.Kind)
	})
}
