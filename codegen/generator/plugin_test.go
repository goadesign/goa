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
	httpdata "goa.design/goa/v3/http/codegen/testdata"
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
		root := &expr.RootExpr{API: &expr.APIExpr{
			Name:              fmt.Sprintf("run-%d", i),
			RandomizerFactory: expr.NewDeterministicRandomizerFactory(),
		}}
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
			root := &expr.RootExpr{API: &expr.APIExpr{
				Name:              fmt.Sprintf("run-%d", index),
				RandomizerFactory: expr.NewDeterministicRandomizerFactory(),
			}}
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
	root := &expr.RootExpr{API: &expr.APIExpr{
		Name:              "before",
		RandomizerFactory: expr.NewDeterministicRandomizerFactory(),
	}}
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

// TestPreparedRootsRejectNonAttributeMutations verifies that service, method,
// transport, and pointer-topology changes stop before the next callback.
func TestPreparedRootsRejectNonAttributeMutations(t *testing.T) {
	tests := []struct {
		name      string
		phase     string
		configure func(*registry, func())
	}{
		{
			name:  "service identity during plugin plan",
			phase: `plugin "a-mutator" plan`,
			configure: func(registry *registry, following func()) {
				registry.addCommand("test")
				registry.registerPlugin("a-mutator", "test", pluginNormal, func() Plugin {
					return Plugin{Plan: func(plan *Plan) error {
						plan.Generation().Roots()[0].(*expr.RootExpr).Services[0].Name = "changed"
						return nil
					}}
				})
				registry.registerPlugin("z-following", "test", pluginNormal, func() Plugin {
					return Plugin{Plan: func(_ *Plan) error {
						following()
						return nil
					}}
				})
			},
		},
		{
			name:  "method identity during core generate",
			phase: `core "method-mutator" generate`,
			configure: func(registry *registry, following func()) {
				registry.addCommand(
					"test",
					func() coreGenerator {
						return coreGenerator{name: "method-mutator", Generate: func(plan *Plan) ([]*codegen.File, error) {
							plan.Generation().Roots()[0].(*expr.RootExpr).Services[0].Methods[0].Name = "changed"
							return nil, nil
						}}
					},
					func() coreGenerator {
						return coreGenerator{name: "following", Generate: func(_ *Plan) ([]*codegen.File, error) {
							following()
							return nil, nil
						}}
					},
				)
			},
		},
		{
			name:  "HTTP route during plugin generate",
			phase: `plugin "a-mutator" generate`,
			configure: func(registry *registry, following func()) {
				registry.addCommand("test")
				registry.registerPlugin("a-mutator", "test", pluginNormal, func() Plugin {
					return Plugin{Generate: func(plan *Plan, files []*codegen.File) ([]*codegen.File, error) {
						plan.Generation().Roots()[0].(*expr.RootExpr).API.HTTP.Services[0].HTTPEndpoints[0].Routes[0].Path = "/changed"
						return files, nil
					}}
				})
				registry.registerPlugin("z-following", "test", pluginNormal, func() Plugin {
					return Plugin{Generate: func(_ *Plan, files []*codegen.File) ([]*codegen.File, error) {
						following()
						return files, nil
					}}
				})
			},
		},
		{
			name:  "equal service replacement during core plan",
			phase: `core "topology-mutator" plan`,
			configure: func(registry *registry, following func()) {
				registry.addCommand(
					"test",
					func() coreGenerator {
						return coreGenerator{name: "topology-mutator", Plan: func(plan *Plan) error {
							root := plan.Generation().Roots()[0].(*expr.RootExpr)
							copy := *root.Services[0]
							root.Services[0] = &copy
							return nil
						}}
					},
					func() coreGenerator {
						return coreGenerator{name: "following", Plan: func(_ *Plan) error {
							following()
							return nil
						}}
					},
				)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := expr.RunDSL(t, httpdata.AliasTypeDSL)
			registry := newRegistry()
			followingRan := false
			test.configure(registry, func() {
				followingRan = true
			})

			_, err := executeGeneration("generated.local/gen", []eval.Root{root}, "test", registry)
			require.ErrorContains(t, err, test.phase+" mutated prepared design")
			require.False(t, followingRan)
		})
	}
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
		phase := 0
		return coreGenerator{
			name: "state",
			Plan: func(_ *Plan) error {
				if phase != 0 {
					return fmt.Errorf("core plan started at phase %d", phase)
				}
				phase++
				return nil
			},
			Generate: func(plan *Plan) ([]*codegen.File, error) {
				if phase != 1 {
					return nil, fmt.Errorf("core generate started at phase %d", phase)
				}
				phase++
				return []*codegen.File{{Path: plan.Generation().GenPkg()}}, nil
			},
		}
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
