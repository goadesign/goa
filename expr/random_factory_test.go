// This file verifies that immutable example configuration creates independent
// mutable value streams for each code generation run.
package expr_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

type (
	// customRandomizerFactory exercises the public factory contract without
	// depending on Goa's built-in factory implementations.
	customRandomizerFactory struct {
		seed string
	}

	recordingRandomizerFactory struct {
		identities *[]expr.ExampleIdentity
	}
)

// NewRandomizer creates an independent seeded stream for identity.
func (f customRandomizerFactory) NewRandomizer(identity expr.ExampleIdentity) expr.Randomizer {
	if identity.Seed() == "" {
		panic("custom randomizer received an empty identity")
	}
	return expr.NewFakerRandomizerFactory(f.seed).NewRandomizer(identity)
}

// NewRandomizer records the exact owner selected by example traversal and
// delegates value generation to Goa's deterministic factory.
func (f recordingRandomizerFactory) NewRandomizer(identity expr.ExampleIdentity) expr.Randomizer {
	*f.identities = append(*f.identities, identity)
	return expr.NewDeterministicRandomizerFactory().NewRandomizer(identity)
}

func TestRandomizerFactoriesCreateIndependentStreams(t *testing.T) {
	cases := []struct {
		Name    string
		Factory expr.RandomizerFactory
	}{
		{"faker", expr.NewFakerRandomizerFactory("seed")},
		{"deterministic", expr.NewDeterministicRandomizerFactory()},
		{"custom", customRandomizerFactory{seed: "seed"}},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			identity := expr.MethodPayloadExampleIdentity(exampleMethod("service", "method"))
			first := expr.NewExampleGenerator(c.Factory).At(identity)
			second := expr.NewExampleGenerator(c.Factory).At(identity)

			require.NotSame(t, first, second)
			require.Equal(t, first.String(), second.String())
			require.Equal(t, first.Int(), second.Int())
		})
	}
}

// TestReleasedStandaloneRandomizers checks the released concrete randomizer
// types, seed, and repeatable values.
func TestReleasedStandaloneRandomizers(t *testing.T) {
	faker := expr.NewFakerRandomizer("seed")
	concrete, ok := faker.(*expr.FakerRandomizer)
	require.True(t, ok)
	require.Equal(t, "seed", concrete.Seed)
	require.Equal(t, expr.NewFakerRandomizer("seed").String(), faker.String())

	deterministic := expr.NewDeterministicRandomizer()
	_, ok = deterministic.(*expr.DeterministicRandomizer)
	require.True(t, ok)
	require.Equal(t, "abc123", deterministic.String())
}

func TestRandomizerFactoriesPreserveDerivedExampleStability(t *testing.T) {
	factory := expr.NewFakerRandomizerFactory("seed")
	identity := expr.MethodPayloadExampleIdentity(exampleMethod("service", "method"))
	first := expr.NewExampleGenerator(factory).At(identity)
	second := expr.NewExampleGenerator(factory).At(identity)

	require.Equal(t, first.Member("payload").String(), second.Member("payload").String())
	require.Equal(t, first.ArrayElement(0).Int(), second.ArrayElement(0).Int())
}

func TestExampleIdentitiesFrameComponents(t *testing.T) {
	for _, delimiter := range []string{".", "/", ":"} {
		t.Run(delimiter, func(t *testing.T) {
			left := expr.MethodPayloadExampleIdentity(exampleMethod("a"+delimiter+"b", "c"))
			right := expr.MethodPayloadExampleIdentity(exampleMethod("a", "b"+delimiter+"c"))

			require.NotEqual(t, left.Seed(), right.Seed())
		})
	}
}

func TestExampleIdentitiesDistinguishSemanticAndStructuralKinds(t *testing.T) {
	method := exampleMethod("service", "method")
	payload := expr.MethodPayloadExampleIdentity(method)
	result := expr.MethodResultExampleIdentity(method)

	require.NotEqual(t, payload.Seed(), result.Seed())
	require.NotEqual(t, payload.Member("0").Seed(), payload.ArrayElement(0).Seed())
	require.NotEqual(t, payload.Member("value").Seed(), payload.UnionMember("value").Seed())
	require.NotEqual(t, payload.MapKey(0).Seed(), payload.MapValue(0).Seed())
	errorIdentity := expr.MethodErrorExampleIdentity(method, &expr.ErrorExpr{Name: "failure"})
	require.NotEqual(t, result.Member("failure").Seed(), errorIdentity.Seed())
}

func TestHTTPResponseIdentitiesIgnoreTraversalOrderAndDistinguishErrors(t *testing.T) {
	method := exampleMethod("service", "method")
	endpoint := &expr.HTTPEndpointExpr{MethodExpr: method}
	ok := &expr.HTTPResponseExpr{StatusCode: expr.StatusOK}
	created := &expr.HTTPResponseExpr{StatusCode: expr.StatusCreated}
	responses := []*expr.HTTPResponseExpr{ok, created}
	before := map[int]string{
		ok.StatusCode:      expr.ResponseBodyExampleIdentity(endpoint, responses[0]).Seed(),
		created.StatusCode: expr.ResponseBodyExampleIdentity(endpoint, responses[1]).Seed(),
	}

	responses[0], responses[1] = responses[1], responses[0]
	require.Equal(t, before[created.StatusCode], expr.ResponseBodyExampleIdentity(endpoint, responses[0]).Seed())
	require.Equal(t, before[ok.StatusCode], expr.ResponseBodyExampleIdentity(endpoint, responses[1]).Seed())

	firstError := &expr.HTTPErrorExpr{Name: "missing", Response: &expr.HTTPResponseExpr{StatusCode: expr.StatusNotFound}}
	secondError := &expr.HTTPErrorExpr{Name: "gone", Response: &expr.HTTPResponseExpr{StatusCode: expr.StatusNotFound}}
	require.NotEqual(t,
		expr.ErrorResponseBodyExampleIdentity(endpoint, firstError).Seed(),
		expr.ErrorResponseBodyExampleIdentity(endpoint, secondError).Seed(),
	)
	require.NotEqual(t,
		expr.ResponseBodyExampleIdentity(endpoint, &expr.HTTPResponseExpr{StatusCode: expr.StatusNotFound}).Seed(),
		expr.ErrorResponseBodyExampleIdentity(endpoint, firstError).Seed(),
	)
}

func TestHTTPBodyIdentitiesDistinguishHTTPAndJSONRPCMappings(t *testing.T) {
	method := exampleMethod("service", "method")
	httpEndpoint := &expr.HTTPEndpointExpr{MethodExpr: method}
	jsonRPCEndpoint := &expr.HTTPEndpointExpr{
		MethodExpr: method,
		Meta:       expr.MetaExpr{"jsonrpc": {}},
	}

	require.NotEqual(t,
		expr.RequestBodyExampleIdentity(httpEndpoint).Seed(),
		expr.RequestBodyExampleIdentity(jsonRPCEndpoint).Seed(),
	)
}

func TestGRPCMessageIdentitiesDistinguishExactMethodsAndRoles(t *testing.T) {
	dashed := exampleMethod("service", "foo-bar")
	underscore := exampleMethod("service", "foo_bar")
	errorExpr := &expr.ErrorExpr{Name: "failure"}

	require.NotEqual(t,
		expr.GRPCRequestMessageExampleIdentity(dashed).Seed(),
		expr.GRPCRequestMessageExampleIdentity(underscore).Seed(),
	)
	require.NotEqual(t,
		expr.GRPCRequestMessageExampleIdentity(dashed).Seed(),
		expr.GRPCResponseMessageExampleIdentity(dashed).Seed(),
	)
	require.NotEqual(t,
		expr.GRPCStreamingRequestMessageExampleIdentity(dashed).Seed(),
		expr.GRPCStreamingResponseMessageExampleIdentity(dashed).Seed(),
	)
	require.NotEqual(t,
		expr.GRPCResponseMessageExampleIdentity(dashed).Seed(),
		expr.GRPCErrorMessageExampleIdentity(dashed, errorExpr).Seed(),
	)
}

func TestInlineMethodErrorsRetainMethodErrorIdentity(t *testing.T) {
	root := expr.RunDSL(t, func() {
		var authored = dsl.Type("AuthoredError", func() {
			dsl.Attribute("message", dsl.String)
		})
		dsl.Service("Values", func() {
			dsl.Method("Read", func() {
				dsl.Error("inline", dsl.String)
				dsl.Error("authored", authored)
			})
		})
	})
	method := root.Service("Values").Method("Read")
	cases := []struct {
		name     string
		error    *expr.ErrorExpr
		expected expr.ExampleIdentity
	}{
		{
			name:     "inline",
			error:    method.Error("inline"),
			expected: expr.MethodErrorExampleIdentity(method, method.Error("inline")),
		},
		{
			name:     "authored",
			error:    method.Error("authored"),
			expected: expr.UserTypeExampleIdentity(root.UserType("AuthoredError")),
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			var identities []expr.ExampleIdentity
			generator := expr.NewExampleGenerator(recordingRandomizerFactory{identities: &identities})
			test.error.AttributeExpr.Example(generator.At(expr.MethodPayloadExampleIdentity(method)))

			require.NotEmpty(t, identities)
			require.Contains(t, identities, test.expected)
		})
	}
}

func TestConfiguredExampleGeneratorRequiresIdentity(t *testing.T) {
	attribute := &expr.AttributeExpr{Type: expr.String}

	require.Panics(t, func() {
		attribute.Example(expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("seed")))
	})
}

func TestZeroExampleIdentityCannotCreateStructuralIdentity(t *testing.T) {
	var identity expr.ExampleIdentity

	require.Panics(t, func() {
		identity.Member("field")
	})
}

func TestDisabledExampleGeneratorSuppressesAuthoredExamples(t *testing.T) {
	attribute := &expr.AttributeExpr{
		Type:         expr.String,
		UserExamples: []*expr.ExampleExpr{{Value: "authored"}},
	}

	require.Nil(t, attribute.Example(&expr.ExampleGenerator{}))
}

func exampleMethod(service, method string) *expr.MethodExpr {
	svc := &expr.ServiceExpr{Name: service}
	return &expr.MethodExpr{Name: method, Service: svc}
}
