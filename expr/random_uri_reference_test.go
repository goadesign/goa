// This file verifies that faker URI-reference examples cover both full URIs
// and relative references, and that every example is a valid URI-reference.
package expr_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
	goa "goa.design/goa/v3/pkg"
)

func TestFakerRandomizerURIReference(t *testing.T) {
	cases := []struct {
		name   string
		prefix string
	}{
		{"full uri", "http://"},
		{"relative reference", "/"},
	}
	var refs []string
	for i := range 64 {
		ref := expr.NewFakerRandomizer("seed" + strconv.Itoa(i)).URIReference()
		require.NoError(t, goa.ValidateFormat("ref", ref, goa.FormatURIReference), ref)
		refs = append(refs, ref)
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.True(t, containsPrefix(refs, tc.prefix), "no %s in %v", tc.name, refs)
		})
	}
}

// containsPrefix reports whether any value in refs starts with prefix.
func containsPrefix(refs []string, prefix string) bool {
	for _, ref := range refs {
		if strings.HasPrefix(ref, prefix) {
			return true
		}
	}
	return false
}
