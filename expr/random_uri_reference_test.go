// This file verifies FormatURIReference example generation across faker,
// deterministic, and custom Randomizer implementations, and confirms existing
// formats do not evaluate URI-reference examples across retries.
package expr_test

import (
	"net"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
	goa "goa.design/goa/v3/pkg"
)

type (
	// standaloneRandomizer implements only the v3 expr.Randomizer interface
	// without embedding Goa's built-in randomizers or implementing URIReference.
	standaloneRandomizer struct {
		emailCalls int
		urlCalls   int
		uuidCalls  int
		uuids      []string
	}

	// standaloneRandomizerFactory returns a preconfigured standaloneRandomizer so
	// tests can assert method call counts.
	standaloneRandomizerFactory struct {
		instance *standaloneRandomizer
	}

	// customURIReferenceRandomizer opts into dedicated URI-reference example
	// generation via the optional URIReference method.
	customURIReferenceRandomizer struct {
		*standaloneRandomizer
		uriRefCalls int
	}

	// customURIReferenceFactory returns a preconfigured
	// customURIReferenceRandomizer so tests can verify optional URIReference
	// dispatch and call counts.
	customURIReferenceFactory struct {
		instance *customURIReferenceRandomizer
	}
)

// Compile-time check that the 19-method v3 interface remains satisfied without
// implementing URIReference.
var _ expr.Randomizer = (*standaloneRandomizer)(nil)

func TestURIReferenceExampleFaker(t *testing.T) {
	cases := []struct {
		name   string
		prefix string
	}{
		{"full uri", "http://"},
		{"relative reference", "/"},
	}
	att := &expr.AttributeExpr{
		Type:       expr.String,
		Validation: &expr.ValidationExpr{Format: expr.FormatURIReference},
	}
	identity := expr.MethodPayloadExampleIdentity(exampleMethod("links", "create"))
	refs := make([]string, 64)
	for i := range refs {
		r := expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("seed" + strconv.Itoa(i))).At(identity)
		ref, ok := att.Example(r).(string)
		require.True(t, ok, "expected string example")
		require.NoError(t, goa.ValidateFormat("ref", ref, goa.FormatURIReference), ref)
		refs[i] = ref
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.True(t, containsPrefix(refs, tc.prefix), "no %s in %v", tc.name, refs)
		})
	}
}

func TestURIReferenceExampleRandomizers(t *testing.T) {
	att := &expr.AttributeExpr{
		Type:       expr.String,
		Validation: &expr.ValidationExpr{Format: expr.FormatURIReference},
	}
	identity := expr.MethodPayloadExampleIdentity(exampleMethod("links", "resolve"))
	cases := []struct {
		name     string
		setup    func() (expr.RandomizerFactory, func(*testing.T))
		expected string
	}{
		{
			name: "deterministic",
			setup: func() (expr.RandomizerFactory, func(*testing.T)) {
				return expr.NewDeterministicRandomizerFactory(), nil
			},
			expected: "/foo",
		},
		{
			name: "custom without URIReference falls back to URL",
			setup: func() (expr.RandomizerFactory, func(*testing.T)) {
				rand := &standaloneRandomizer{}
				return standaloneRandomizerFactory{instance: rand}, func(t *testing.T) {
					require.Equal(t, 1, rand.urlCalls)
					require.Equal(t, 0, rand.emailCalls)
				}
			},
			expected: "https://custom.example/resource",
		},
		{
			name: "custom implementing URIReference uses it directly",
			setup: func() (expr.RandomizerFactory, func(*testing.T)) {
				rand := &customURIReferenceRandomizer{standaloneRandomizer: &standaloneRandomizer{}}
				return customURIReferenceFactory{instance: rand}, func(t *testing.T) {
					require.Equal(t, 1, rand.uriRefCalls)
					require.Equal(t, 0, rand.urlCalls)
					require.Equal(t, 0, rand.emailCalls)
				}
			},
			expected: "/custom/ref",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			factory, check := tc.setup()
			r := expr.NewExampleGenerator(factory).At(identity)
			ref, ok := att.Example(r).(string)
			require.True(t, ok)
			require.Equal(t, tc.expected, ref)
			require.NoError(t, goa.ValidateFormat("ref", ref, goa.FormatURIReference))
			if check != nil {
				check(t)
			}
		})
	}
}

func TestFormatWithPatternRetrySkipsURIReference(t *testing.T) {
	rand := &customURIReferenceRandomizer{
		standaloneRandomizer: &standaloneRandomizer{
			uuids: []string{
				"11111111-1111-4111-8111-111111111111",
				"00000000-0000-4000-8000-000000000000",
			},
		},
	}
	att := &expr.AttributeExpr{
		Type: expr.String,
		Validation: &expr.ValidationExpr{
			Format:  expr.FormatUUID,
			Pattern: "^0",
		},
	}
	identity := expr.MethodPayloadExampleIdentity(exampleMethod("items", "get"))
	r := expr.NewExampleGenerator(customURIReferenceFactory{instance: rand}).At(identity)

	example, ok := att.Example(r).(string)
	require.True(t, ok)
	require.Equal(t, "00000000-0000-4000-8000-000000000000", example)
	require.Equal(t, 2, rand.uuidCalls)
	require.Equal(t, 0, rand.uriRefCalls)
	require.Equal(t, 2, rand.urlCalls)
}

func (f standaloneRandomizerFactory) NewRandomizer(expr.ExampleIdentity) expr.Randomizer {
	return f.instance
}

func (f customURIReferenceFactory) NewRandomizer(expr.ExampleIdentity) expr.Randomizer {
	return f.instance
}

func (r *customURIReferenceRandomizer) URIReference() string {
	r.uriRefCalls++
	return "/custom/ref"
}

func (*standaloneRandomizer) ArrayLength() int {
	return 1
}

func (*standaloneRandomizer) Int() int {
	return 1
}

func (*standaloneRandomizer) Int32() int32 {
	return 1
}

func (*standaloneRandomizer) Int64() int64 {
	return 1
}

func (*standaloneRandomizer) String() string {
	return "abc123"
}

func (*standaloneRandomizer) Bool() bool {
	return false
}

func (*standaloneRandomizer) Float32() float32 {
	return 1
}

func (*standaloneRandomizer) Float64() float64 {
	return 1
}

func (*standaloneRandomizer) UInt() uint {
	return 1
}

func (*standaloneRandomizer) UInt32() uint32 {
	return 1
}

func (*standaloneRandomizer) UInt64() uint64 {
	return 1
}

func (*standaloneRandomizer) Name() string {
	return "Alice"
}

func (r *standaloneRandomizer) Email() string {
	r.emailCalls++
	return "alice@example.com"
}

func (*standaloneRandomizer) Hostname() string {
	return "example.com"
}

func (*standaloneRandomizer) IPv4Address() net.IP {
	return net.IPv4zero
}

func (*standaloneRandomizer) IPv6Address() net.IP {
	return net.IPv6zero
}

func (r *standaloneRandomizer) URL() string {
	r.urlCalls++
	return "https://custom.example/resource"
}

func (*standaloneRandomizer) Characters(n int) string {
	return strings.Repeat("a", n)
}

func (r *standaloneRandomizer) UUID() string {
	idx := r.uuidCalls
	r.uuidCalls++
	if idx < len(r.uuids) {
		return r.uuids[idx]
	}
	return "550e8400-e29b-41d4-a716-446655440000"
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
