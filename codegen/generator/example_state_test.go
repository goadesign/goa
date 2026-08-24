// This file verifies that each generator execution owns independent mutable
// example state while sharing only immutable API factory configuration.
package generator

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

type (
	// mutatingRandomizerFactory violates the immutable API factory contract so
	// the lifecycle can prove it attributes the mutation at construction.
	mutatingRandomizerFactory struct {
		calls int
	}
)

// NewRandomizer records a call before returning a fresh deterministic stream.
func (f *mutatingRandomizerFactory) NewRandomizer(identity expr.ExampleIdentity) expr.Randomizer {
	f.calls++
	return expr.NewDeterministicRandomizerFactory().NewRandomizer(identity)
}

func TestGenerationRejectsFactoryMutationWhenStreamIsCreated(t *testing.T) {
	root := expr.RunDSL(t, func() {})
	root.API.RandomizerFactory = &mutatingRandomizerFactory{}
	registry := newRegistry()
	registry.addCommand("test", func() coreGenerator {
		return coreGenerator{name: "examples", Plan: func(plan *Plan) error {
			plan.exampleGenerator(root).At(generatorTestIdentity())
			return nil
		}}
	})

	err := executeGeneration("generated.local/gen", []eval.Root{root}, "test", registry)

	require.ErrorContains(t, err, `core "examples" plan mutated prepared design`)
}

func TestGenerationRunsOwnIndependentExampleGenerators(t *testing.T) {
	root := expr.RunDSL(t, func() {})
	factory := root.API.RandomizerFactory
	registry := newRegistry()
	var (
		mu         sync.Mutex
		generators []*expr.ExampleGenerator
		examples   []string
	)
	registry.addCommand("test", func() coreGenerator {
		return coreGenerator{name: "examples", Plan: func(plan *Plan) error {
			generator := plan.exampleGenerator(root)
			stream := generator.At(generatorTestIdentity())
			mu.Lock()
			defer mu.Unlock()
			generators = append(generators, generator)
			examples = append(examples, stream.String())
			return nil
		}}
	})

	for range 2 {
		err := executeGeneration("generated.local/gen", []eval.Root{root}, "test", registry)
		require.NoError(t, err)
	}

	require.Len(t, generators, 2)
	require.NotSame(t, generators[0], generators[1])
	require.Equal(t, examples[0], examples[1])
	require.Equal(t, factory, root.API.RandomizerFactory)
}

// generatorTestIdentity returns a typed owner for values drawn directly by
// lifecycle tests rather than by a code-generation subsystem.
func generatorTestIdentity() expr.ExampleIdentity {
	method := &expr.MethodExpr{
		Name:    "lifecycle",
		Service: &expr.ServiceExpr{Name: "generator-test"},
	}
	return expr.MethodPayloadExampleIdentity(method)
}

func TestConcurrentGenerationRunsOwnIndependentExampleGenerators(t *testing.T) {
	registry := newRegistry()
	roots := []*expr.RootExpr{expr.RunDSL(t, func() {}), expr.RunDSL(t, func() {})}
	recursive := make(map[*expr.RootExpr]*expr.UserTypeExpr, len(roots))
	for _, root := range roots {
		node := &expr.UserTypeExpr{TypeName: "Node", UID: "test-node"}
		node.AttributeExpr = &expr.AttributeExpr{Type: &expr.Object{
			{Name: "value", Attribute: &expr.AttributeExpr{Type: expr.String}},
			{Name: "children", Attribute: &expr.AttributeExpr{Type: &expr.Array{
				ElemType: &expr.AttributeExpr{Type: node},
			}}},
		}}
		recursive[root] = node
	}
	var (
		mu         sync.Mutex
		generators []*expr.ExampleGenerator
		examples   []any
	)
	registry.addCommand("test", func() coreGenerator {
		return coreGenerator{name: "examples", Plan: func(plan *Plan) error {
			root := plan.preparedRoots[0].(*expr.RootExpr)
			generator := plan.exampleGenerator(root)
			node := recursive[root]
			example := node.Example(generator.At(expr.UserTypeExampleIdentity(node)))
			mu.Lock()
			defer mu.Unlock()
			generators = append(generators, generator)
			examples = append(examples, example)
			return nil
		}}
	})

	var runs sync.WaitGroup
	errs := make(chan error, 2)
	for _, root := range roots {
		runs.Add(1)
		go func(root *expr.RootExpr) {
			defer runs.Done()
			err := executeGeneration("generated.local/gen", []eval.Root{root}, "test", registry)
			errs <- err
		}(root)
	}
	runs.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}

	require.Len(t, generators, 2)
	require.NotSame(t, generators[0], generators[1])
	require.Equal(t, examples[0], examples[1])
}
