// This file verifies that generator and plugin factories create isolated run
// objects and that every phase receives one retained Plan in stable order.
package generator

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
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

	err := executeGeneration("generated.local/gen", nil, "test", registry)
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

// TestPluginPlannedHTTPDataIsAccepted checks that a factory plugin may declare
// a constructor during Plan and use it as direct HTTP data during Generate.
func TestPluginPlannedHTTPDataIsAccepted(t *testing.T) {
	root := codegen.RunDSL(t, httpdata.ServerSimpleRoutingDSL)
	registry := newDefaultRegistry()
	registry.registerPlugin("planned-http-data", "gen", pluginNormal, func() Plugin {
		var declaration *codegen.NameDeclaration
		return Plugin{
			Plan: func(plan *Plan) error {
				pkg, err := plan.Generation().ClaimPackage("generated.local/gen/http/plugin")
				if err != nil {
					return err
				}
				declaration = codegen.NewExactName(codegen.NameFunction, "BuildPluginBody")
				return pkg.DeclareName(declaration)
			},
			Generate: func(_ *Plan, files []*codegen.File) ([]*codegen.File, error) {
				return append(files, &codegen.File{
					Path: "gen/http/plugin/plugin.go",
					SectionTemplates: []*codegen.SectionTemplate{{
						Name: "plugin-init",
						Data: &httpcodegen.InitData{
							Declaration: declaration,
							Name:        declaration.Name(),
						},
					}},
				}), nil
			},
		}
	})

	run, err := newGenerationRun("gen", registry)
	require.NoError(t, err)
	_, err = run.execute("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
}

// TestPluginOwnedHTTPDeclarationReplacementIsAccepted checks that a later
// plugin may replace a declaration with another name planned by the same run.
func TestPluginOwnedHTTPDeclarationReplacementIsAccepted(t *testing.T) {
	root := codegen.RunDSL(t, httpdata.ServerSimpleRoutingDSL)
	registry := newDefaultRegistry()
	var (
		init        *httpcodegen.InitData
		declaration *codegen.NameDeclaration
		replacement *codegen.NameDeclaration
		laterRan    bool
	)
	registry.registerPlugin("a-add-init", "gen", pluginNormal, func() Plugin {
		return Plugin{
			Plan: func(plan *Plan) error {
				pkg, err := plan.Generation().ClaimPackage("generated.local/gen/http/plugin")
				if err != nil {
					return err
				}
				declaration = codegen.NewExactName(codegen.NameFunction, "BuildPluginBody")
				replacement = codegen.NewExactName(codegen.NameFunction, "BuildOtherBody")
				if err := pkg.DeclareName(declaration); err != nil {
					return err
				}
				return pkg.DeclareName(replacement)
			},
			Generate: func(_ *Plan, files []*codegen.File) ([]*codegen.File, error) {
				init = &httpcodegen.InitData{Declaration: declaration, Name: declaration.Name()}
				return append(files, &codegen.File{
					Path:             "gen/http/plugin/plugin.go",
					SectionTemplates: []*codegen.SectionTemplate{{Name: "plugin-init", Data: init}},
				}), nil
			},
		}
	})
	registry.registerPlugin("b-replace-init", "gen", pluginNormal, func() Plugin {
		return Plugin{Generate: func(_ *Plan, files []*codegen.File) ([]*codegen.File, error) {
			init.Declaration = replacement
			init.Name = replacement.Name()
			return files, nil
		}}
	})
	registry.registerPlugin("c-later", "gen", pluginNormal, func() Plugin {
		return Plugin{Generate: func(_ *Plan, files []*codegen.File) ([]*codegen.File, error) {
			laterRan = true
			return files, nil
		}}
	})

	run, err := newGenerationRun("gen", registry)
	require.NoError(t, err)
	_, err = run.execute("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	require.True(t, laterRan)
}

// TestPluginCallbackErrorIsPreserved checks that an ordinary callback failure
// is returned unchanged and stops later plugins.
func TestPluginCallbackErrorIsPreserved(t *testing.T) {
	root := &expr.RootExpr{API: &expr.APIExpr{
		Name:              "callback-error",
		RandomizerFactory: expr.NewDeterministicRandomizerFactory(),
	}}
	registry := newRegistry()
	registry.addCommand("test")
	laterRan := false
	registry.registerPlugin("fail", "test", pluginNormal, func() Plugin {
		return Plugin{Generate: func(_ *Plan, files []*codegen.File) ([]*codegen.File, error) {
			return files, errors.New("callback failed")
		}}
	})
	registry.registerPlugin("z-later", "test", pluginNormal, func() Plugin {
		return Plugin{Generate: func(_ *Plan, files []*codegen.File) ([]*codegen.File, error) {
			laterRan = true
			return files, nil
		}}
	})

	run, err := newGenerationRun("test", registry)
	require.NoError(t, err)
	_, err = run.execute("generated.local/gen", []eval.Root{root})
	require.EqualError(t, err, "callback failed")
	require.False(t, laterRan)
}

// TestPluginDesignMutationErrorTakesPrecedence checks that Goa reports a
// forbidden design change even when the callback also returns its own error.
// The changed root would otherwise remain visible to later generation runs.
func TestPluginDesignMutationErrorTakesPrecedence(t *testing.T) {
	root := &expr.RootExpr{API: &expr.APIExpr{
		Name:              "before",
		RandomizerFactory: expr.NewDeterministicRandomizerFactory(),
	}}
	registry := newRegistry()
	registry.addCommand("test")
	registry.registerPlugin("mutate-and-fail", "test", pluginNormal, func() Plugin {
		return Plugin{Generate: func(_ *Plan, files []*codegen.File) ([]*codegen.File, error) {
			root.API.Name = "after"
			return files, errors.New("callback failed")
		}}
	})

	run, err := newGenerationRun("test", registry)
	require.NoError(t, err)
	_, err = run.execute("generated.local/gen", []eval.Root{root})
	require.ErrorContains(t, err, `plugin "mutate-and-fail" generate mutated prepared design`)
	require.NotEqual(t, "callback failed", err.Error())
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
		err := executeGeneration(
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
			err := executeGeneration(
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

	err := executeGeneration("generated.local/gen", []eval.Root{root}, "test", registry)
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

			err := executeGeneration("generated.local/gen", []eval.Root{root}, "test", registry)
			require.ErrorContains(t, err, test.phase+" mutated prepared design")
			require.False(t, followingRan)
		})
	}
}

// TestPluginRegistrySealsOnFirstSnapshot verifies that a run cannot observe
// factories registered after the registry's immutable snapshot is established.
func TestPluginRegistrySealsOnFirstSnapshot(t *testing.T) {
	registry := newRegistry()
	registry.addCommand("test", func() coreGenerator {
		return coreGenerator{}
	})
	err := executeGeneration("generated.local/gen", nil, "test", registry)
	require.NoError(t, err)
	require.Panics(t, func() {
		registry.registerPlugin("late", "test", pluginNormal, func() Plugin {
			return Plugin{}
		})
	})
}

// TestReleasedAndFactoryPluginsShareOneRun verifies that plugins registered
// through either API run in one order and receive the same prepared design and
// current file list.
func TestReleasedAndFactoryPluginsShareOneRun(t *testing.T) {
	root := &expr.RootExpr{API: &expr.APIExpr{
		Name:              "prepared",
		RandomizerFactory: expr.NewDeterministicRandomizerFactory(),
	}}
	registry := newRegistry()
	registry.addCommand("test", func() coreGenerator {
		return coreGenerator{Generate: func(_ *Plan) ([]*codegen.File, error) {
			return []*codegen.File{{Path: "core"}}, nil
		}}
	})
	var events []string
	registry.registeredPlugins = func() []registeredPluginDescriptor {
		return []registeredPluginDescriptor{
			releasedPluginForTest("z-first", "test", pluginFirst, root, &events),
			releasedPluginForTest("a-normal", "test", pluginNormal, root, &events),
			releasedPluginForTest("a-last", "test", pluginLast, root, &events),
		}
	}
	registerFactoryPluginForTest(registry, "a-first", pluginFirst, &events)
	registerFactoryPluginForTest(registry, "z-normal", pluginNormal, &events)
	registerFactoryPluginForTest(registry, "z-last", pluginLast, &events)

	run, err := newGenerationRun("test", registry)
	require.NoError(t, err)
	result, err := run.execute("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	require.Equal(t, []string{
		"prepare:a-first", "prepare:z-first", "prepare:a-normal", "prepare:z-normal", "prepare:a-last", "prepare:z-last",
		"generate:a-first:core", "generate:z-first:factory:a-first", "generate:a-normal:released:z-first",
		"generate:z-normal:released:a-normal", "generate:a-last:factory:z-normal", "generate:z-last:released:a-last",
	}, events)
	require.Equal(t, "factory:z-last", result.files[len(result.files)-1].Path)
}

// TestReleasedDuplicatePluginsKeepRegistrationOrder verifies that callbacks
// with the same released command and name still run in registration order.
func TestReleasedDuplicatePluginsKeepRegistrationOrder(t *testing.T) {
	registry := newRegistry()
	registry.addCommand("test", func() coreGenerator {
		return coreGenerator{Generate: func(_ *Plan) ([]*codegen.File, error) {
			return []*codegen.File{{Path: "core"}}, nil
		}}
	})
	var events []string
	registry.registeredPlugins = func() []registeredPluginDescriptor {
		duplicate := func(event string) registeredPluginDescriptor {
			return registeredPluginDescriptor{
				name:     "same",
				command:  "test",
				position: pluginNormal,
				generate: func(_ string, _ []eval.Root, files []*codegen.File) ([]*codegen.File, error) {
					events = append(events, event+":"+files[len(files)-1].Path)
					return append(files, &codegen.File{Path: event}), nil
				},
			}
		}
		return []registeredPluginDescriptor{duplicate("first"), duplicate("second")}
	}

	run, err := newGenerationRun("test", registry)
	require.NoError(t, err)
	_, err = run.execute("generated.local/gen", nil)
	require.NoError(t, err)
	require.Equal(t, []string{"first:core", "second:first"}, events)
}

// TestReleasedPluginCallbackReceivesEachRun checks that the same registered
// function can run twice. Each call receives only that run's generated package,
// design roots, and files.
func TestReleasedPluginCallbackReceivesEachRun(t *testing.T) {
	registry := newRegistry()
	registry.addCommand("test", func() coreGenerator {
		return coreGenerator{Generate: func(plan *Plan) ([]*codegen.File, error) {
			root := plan.Generation().Roots()[0].(*expr.RootExpr)
			return []*codegen.File{{Path: "core-" + root.API.Name}}, nil
		}}
	})

	var (
		preparedPackages  []string
		preparedRoots     [][]eval.Root
		generatedPackages []string
		generatedRoots    [][]eval.Root
		generatedFiles    [][]*codegen.File
	)
	//nolint:unparam // The released callback signature includes an error result.
	prepare := func(genpkg string, roots []eval.Root) error {
		preparedPackages = append(preparedPackages, genpkg)
		preparedRoots = append(preparedRoots, append([]eval.Root(nil), roots...))
		return nil
	}
	//nolint:unparam // The released callback signature includes an error result.
	generate := func(genpkg string, roots []eval.Root, files []*codegen.File) ([]*codegen.File, error) {
		generatedPackages = append(generatedPackages, genpkg)
		generatedRoots = append(generatedRoots, append([]eval.Root(nil), roots...))
		generatedFiles = append(generatedFiles, append([]*codegen.File(nil), files...))
		return append(files, &codegen.File{Path: "released-" + roots[0].(*expr.RootExpr).API.Name}), nil
	}
	registry.registeredPlugins = func() []registeredPluginDescriptor {
		return []registeredPluginDescriptor{{
			name:     "released",
			command:  "test",
			position: pluginNormal,
			prepare:  prepare,
			generate: generate,
		}}
	}

	packages := []string{"generated.local/first", "generated.local/second"}
	roots := []*expr.RootExpr{
		{API: &expr.APIExpr{Name: "first", RandomizerFactory: expr.NewDeterministicRandomizerFactory()}},
		{API: &expr.APIExpr{Name: "second", RandomizerFactory: expr.NewDeterministicRandomizerFactory()}},
	}
	for index, root := range roots {
		run, err := newGenerationRun("test", registry)
		require.NoError(t, err)
		result, err := run.execute(packages[index], []eval.Root{root})
		require.NoError(t, err)
		require.Equal(t, "released-"+root.API.Name, result.files[1].Path)
	}

	require.Equal(t, packages, preparedPackages)
	require.Equal(t, packages, generatedPackages)
	for index, root := range roots {
		require.Len(t, preparedRoots[index], 1)
		require.Same(t, root, preparedRoots[index][0])
		require.Len(t, generatedRoots[index], 1)
		require.Same(t, root, generatedRoots[index][0])
		require.Len(t, generatedFiles[index], 1)
		require.Equal(t, "core-"+root.API.Name, generatedFiles[index][0].Path)
	}
}

// TestReleasedPluginNilFileRemainsVisibleUntilMerge checks that one plugin may
// return a one-item list containing nil. The next plugin receives that list
// unchanged, and Goa omits nil before writing files. Released Goa accidentally
// panicked when nil was the only file; generation now handles every list size
// consistently.
func TestReleasedPluginNilFileRemainsVisibleUntilMerge(t *testing.T) {
	codegen.RunDSL(t, func() {
	})
	registry := newRegistry()
	registry.addCommand("test")
	observedNil := false
	registry.registeredPlugins = func() []registeredPluginDescriptor {
		return []registeredPluginDescriptor{
			{
				name:     "a-return-nil",
				command:  "test",
				position: pluginNormal,
				generate: func(_ string, _ []eval.Root, _ []*codegen.File) ([]*codegen.File, error) {
					return []*codegen.File{nil}, nil
				},
			},
			{
				name:     "b-observe-nil",
				command:  "test",
				position: pluginNormal,
				generate: func(_ string, _ []eval.Root, files []*codegen.File) ([]*codegen.File, error) {
					observedNil = len(files) == 1 && files[0] == nil
					return files, nil
				},
			},
		}
	}
	directory := t.TempDir()
	writeGeneratedModule(t, filepath.Join(directory, codegen.Gendir), "generated.local/gen")

	outputs, err := generate(directory, "test", false, registry)
	require.NoError(t, err)
	require.True(t, observedNil)
	require.Empty(t, outputs)
}

// TestReleasedAndFactoryPluginDuplicatesStopBeforeCallbacks verifies that a
// command/name pair cannot be registered once through each API.
func TestReleasedAndFactoryPluginDuplicatesStopBeforeCallbacks(t *testing.T) {
	registry := newRegistry()
	registry.addCommand("test")
	called := false
	registry.registeredPlugins = func() []registeredPluginDescriptor {
		return []registeredPluginDescriptor{{
			name:     "duplicate",
			command:  "test",
			position: pluginFirst,
			generate: func(_ string, _ []eval.Root, files []*codegen.File) ([]*codegen.File, error) {
				called = true
				return files, nil
			},
		}}
	}
	registry.registerPlugin("duplicate", "test", pluginNormal, func() Plugin {
		called = true
		return Plugin{}
	})

	_, err := newGenerationRun("test", registry)
	require.ErrorContains(t, err, `plugin "duplicate" is already registered for command "test"`)
	require.False(t, called)
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

// releasedPluginForTest creates an old-style callback that checks the exact
// package, root, and file list passed from the shared generation run.
func releasedPluginForTest(name, command string, position pluginPosition, root eval.Root, events *[]string) registeredPluginDescriptor {
	return registeredPluginDescriptor{
		name:     name,
		command:  command,
		position: position,
		prepare: func(genpkg string, roots []eval.Root) error {
			if genpkg != "generated.local/gen" {
				return fmt.Errorf("prepare received package %q", genpkg)
			}
			if len(roots) != 1 || roots[0] != root {
				return fmt.Errorf("prepare received another run's roots")
			}
			*events = append(*events, "prepare:"+name)
			return nil
		},
		generate: func(genpkg string, roots []eval.Root, files []*codegen.File) ([]*codegen.File, error) {
			if genpkg != "generated.local/gen" {
				return nil, fmt.Errorf("generate received package %q", genpkg)
			}
			if len(roots) != 1 || roots[0] != root {
				return nil, fmt.Errorf("generate received another run's roots")
			}
			last := files[len(files)-1].Path
			*events = append(*events, "generate:"+name+":"+last)
			return append(files, &codegen.File{Path: "released:" + name}), nil
		},
	}
}

// registerFactoryPluginForTest adds a factory plugin that records its current
// input file and appends one file for the following plugin.
func registerFactoryPluginForTest(registry *registry, name string, position pluginPosition, events *[]string) {
	registry.registerPlugin(name, "test", position, func() Plugin {
		return Plugin{
			Prepare: func(_ string, _ []eval.Root) error {
				*events = append(*events, "prepare:"+name)
				return nil
			},
			Generate: func(_ *Plan, files []*codegen.File) ([]*codegen.File, error) {
				last := files[len(files)-1].Path
				*events = append(*events, "generate:"+name+":"+last)
				return append(files, &codegen.File{Path: "factory:" + name}), nil
			},
		}
	})
}
