package expr

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGeneratedResultTypeMatchesCanonicalIdentifier verifies that a lookup by
// canonical identifier finds a generated result type recorded with its authored
// media type suffix. CollectionOf looks its cache up with a canonical
// identifier, so a suffixed identifier must not create a second type.
func TestGeneratedResultTypeMatchesCanonicalIdentifier(t *testing.T) {
	ResetDSL(t)

	suffixed := "application/vnd.item+json; type=collection"
	collection := NewResultTypeExpr("ItemCollection", suffixed, func() {})
	GeneratedResultTypes.Append(collection)

	canonical := CanonicalIdentifier(suffixed)
	require.NotEqual(t, suffixed, canonical, "the identifier must lose its suffix to exercise the lookup")

	assert.Same(t, collection, GeneratedResultType(canonical))
	assert.Same(t, collection, GeneratedResultType(suffixed))
	assert.Nil(t, GeneratedResultType("application/vnd.other; type=collection"))
}
