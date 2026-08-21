// This file verifies that generator and plugin factories create isolated run
// objects and that every phase receives one retained Plan in stable order.
package generator

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// TestPluginFactoryOrderAndPlan verifies First, normal, and Last ordering and
// proves that plugin planning and rendering receive the exact same Plan pointer.
func TestPluginFactoryOrderAndPlan(t *testing.T) {
	registry := newRegistry()
	registry.addCommand("test", func() coreGenerator {
		return coreGenerator{}
	})
	var (
		events  []string
		plans   []*Plan
		planMux sync.Mutex
	)
	register := func(position pluginPosition, name string) {
		registry.registerPlugin(name, "test", position, func() Plugin {
			return Plugin{
				Prepare: func(_ string, _ []eval.Root) error {
					events = append(events, "prepare:"+name)
					return nil
				},
				Plan: func(plan *Plan) error {
					events = append(events, "plan:"+name)
					planMux.Lock()
					plans = append(plans, plan)
					planMux.Unlock()
					return nil
				},
				Generate: func(plan *Plan, files []*codegen.File) ([]*codegen.File, error) {
					events = append(events, "generate:"+name)
					planMux.Lock()
					plans = append(plans, plan)
					planMux.Unlock()
					return files, nil
				},
			}
		})
	}
	register(pluginLast, "z-last")
	register(pluginNormal, "z-normal")
	register(pluginFirst, "b-first")
	register(pluginFirst, "a-first")
	register(pluginNormal, "a-normal")
	register(pluginLast, "a-last")

	_, err := executeGeneration("generated.local/gen", nil, "test", registry)
	require.NoError(t, err)
	require.Equal(t, []string{
		"prepare:a-first", "prepare:b-first", "prepare:a-normal", "prepare:z-normal", "prepare:a-last", "prepare:z-last",
		"plan:a-first", "plan:b-first", "plan:a-normal", "plan:z-normal", "plan:a-last", "plan:z-last",
		"generate:a-first", "generate:b-first", "generate:a-normal", "generate:z-normal", "generate:a-last", "generate:z-last",
	}, events)
	require.Len(t, plans, 12)
	for _, plan := range plans[1:] {
		require.Same(t, plans[0], plan)
	}
	require.NotNil(t, plans[0].Generation())
}

// TestPluginFactorySequentialIsolation verifies that every run invokes the
// factory again and no mutable callback state survives from an earlier run.
func TestPluginFactorySequentialIsolation(t *testing.T) {
	registry := isolatedPluginRegistry(t)

	for i := range 2 {
		root := &expr.RootExpr{API: &expr.APIExpr{Name: fmt.Sprintf("run-%d", i)}}
		_, err := executeGeneration(
			fmt.Sprintf("generated.local/gen%d", i),
			[]eval.Root{root},
			"test",
			registry,
		)
		require.NoError(t, err)
	}
}

// TestPluginFactoryConcurrentIsolation verifies that registry snapshots are
// race-safe and concurrent runs own independent callback state.
func TestPluginFactoryConcurrentIsolation(t *testing.T) {
	registry := isolatedPluginRegistry(t)
	var wait sync.WaitGroup
	errs := make(chan error, 2)
	for i := range 2 {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			root := &expr.RootExpr{API: &expr.APIExpr{Name: fmt.Sprintf("run-%d", index)}}
			_, err := executeGeneration(
				fmt.Sprintf("generated.local/gen%d", index),
				[]eval.Root{root},
				"test",
				registry,
			)
			errs <- err
		}(i)
	}
	wait.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
}

// TestPreparedRootsBecomeExactGenerationSnapshot verifies that plugin
// preparation completes before Generation copies root membership and values.
func TestPreparedRootsBecomeExactGenerationSnapshot(t *testing.T) {
	root := &expr.RootExpr{API: &expr.APIExpr{Name: "before"}}
	registry := newRegistry()
	registry.addCommand("test", func() coreGenerator {
		return coreGenerator{
			Plan: func(plan *Plan) error {
				if !plan.Generation().HasRoot(root) {
					return fmt.Errorf("prepared root is absent from generation")
				}
				if root.API.Name != "after" {
					return fmt.Errorf("generation observed API name %q", root.API.Name)
				}
				return nil
			},
		}
	})
	registry.registerPlugin("prepare", "test", pluginNormal, func() Plugin {
		return Plugin{Prepare: func(_ string, roots []eval.Root) error {
			roots[0].(*expr.RootExpr).API.Name = "after"
			return nil
		}}
	})

	_, err := executeGeneration("generated.local/gen", []eval.Root{root}, "test", registry)
	require.NoError(t, err)
}

// TestPluginRegistrySealsOnFirstSnapshot verifies that a run cannot observe
// factories registered after the registry's immutable snapshot is established.
func TestPluginRegistrySealsOnFirstSnapshot(t *testing.T) {
	registry := newRegistry()
	registry.addCommand("test", func() coreGenerator { return coreGenerator{} })
	_, err := executeGeneration("generated.local/gen", nil, "test", registry)
	require.NoError(t, err)
	require.Panics(t, func() {
		registry.registerPlugin("late", "test", pluginNormal, func() Plugin { return Plugin{} })
	})
}

// isolatedPluginRegistry builds a factory whose private phase counter must
// always start at zero and advance exactly once through the three phases.
func isolatedPluginRegistry(t *testing.T) *registry {
	t.Helper()
	registry := newRegistry()
	registry.addCommand("test", func() coreGenerator {
		return coreGenerator{Generate: func(plan *Plan) ([]*codegen.File, error) {
			return []*codegen.File{{Path: plan.Generation().GenPkg()}}, nil
		}}
	})
	registry.registerPlugin("state", "test", pluginNormal, func() Plugin {
		var (
			phase        int
			preparedRoot eval.Root
			planned      *Plan
		)
		return Plugin{
			Prepare: func(_ string, roots []eval.Root) error {
				if phase != 0 {
					return fmt.Errorf("prepare started at phase %d", phase)
				}
				if len(roots) != 1 {
					return fmt.Errorf("prepare received %d roots", len(roots))
				}
				preparedRoot = roots[0]
				phase++
				return nil
			},
			Plan: func(plan *Plan) error {
				if phase != 1 {
					return fmt.Errorf("plan started at phase %d", phase)
				}
				roots := plan.Generation().Roots()
				if len(roots) != 1 || roots[0] != preparedRoot {
					return fmt.Errorf("plan received another run's roots")
				}
				planned = plan
				phase++
				return nil
			},
			Generate: func(plan *Plan, files []*codegen.File) ([]*codegen.File, error) {
				if phase != 2 {
					return nil, fmt.Errorf("generate started at phase %d", phase)
				}
				if plan != planned {
					return nil, fmt.Errorf("generate received another run's plan")
				}
				if len(files) != 1 || files[0].Path != plan.Generation().GenPkg() {
					return nil, fmt.Errorf("generate received another run's files")
				}
				phase++
				return files, nil
			},
		}
	})
	return registry
}
