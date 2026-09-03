// This file checks that each command builds files from its own Plan and that
// simultaneous commands cannot change each other's output.
package generator

import (
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

type (
	// generatorCall records the Plan passed to one command's two functions.
	generatorCall struct {
		planCalls     int
		generateCalls int
		planned       *Plan
		generated     *Plan
	}

	// commandResult contains every file written by one command, indexed by its
	// path beneath the output directory.
	commandResult struct {
		files map[string][]byte
		err   error
	}
)

// TestCommandsUseEachSelectedGeneratorOnce checks that each command calls only
// its listed generators and passes the same Plan to both functions.
func TestCommandsUseEachSelectedGeneratorOnce(t *testing.T) {
	root := codegen.RunDSL(t, commandIsolationDSL("first"))
	cases := []struct {
		name      string
		factories []generatorFactory
		selected  []string
	}{
		{"gen", genGeneratorFactories(), []string{"service", "transport", "openapi"}},
		{"example", exampleGeneratorFactories(), []string{"example"}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			calls := make(map[string]*generatorCall, len(test.selected))
			registry := newRegistry()
			registry.addCommand(test.name, observedGenerators(test.factories, calls)...)

			run, err := newGenerationRun(test.name, registry)
			require.NoError(t, err)
			result, err := run.execute("generated.local/gen", []eval.Root{root})
			require.NoError(t, err)

			require.ElementsMatch(t, test.selected, mapKeys(calls))
			for _, name := range test.selected {
				call := calls[name]
				require.Equal(t, 1, call.planCalls, "%s plan calls", name)
				require.Equal(t, 1, call.generateCalls, "%s file calls", name)
				require.Same(t, result.plan, call.planned, "%s planned Plan", name)
				require.Same(t, call.planned, call.generated, "%s generated Plan", name)
			}
		})
	}
}

// TestFocusedCommandDoesNotBuildExamplesOrOpenAPI checks that a command with no
// example or OpenAPI generator creates neither result.
func TestFocusedCommandDoesNotBuildExamplesOrOpenAPI(t *testing.T) {
	root := codegen.RunDSL(t, commandIsolationDSL("focused"))
	registry := testRegistry("focused", testGenerator(nil, nil))
	run, err := newGenerationRun("focused", registry)
	require.NoError(t, err)
	result, err := run.execute("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	require.Nil(t, result.plan.example)
	require.Nil(t, result.plan.openapi)
}

// TestCommandsProduceTheSameFilesWhenRunTogether checks that gen and example
// produce the same bytes alone, repeatedly, and beside another command run.
func TestCommandsProduceTheSameFilesWhenRunTogether(t *testing.T) {
	for _, command := range []string{"gen", "example"} {
		t.Run(command, func(t *testing.T) {
			first := codegen.RunDSL(t, commandIsolationDSL("first"))
			second := codegen.RunDSL(t, commandIsolationDSL("second"))
			firstExpected, err := renderCommand(command, first, t.TempDir())
			require.NoError(t, err)
			firstAgain, err := renderCommand(command, first, t.TempDir())
			require.NoError(t, err)
			require.Equal(t, firstExpected, firstAgain)
			secondExpected, err := renderCommand(command, second, t.TempDir())
			require.NoError(t, err)

			start := make(chan struct{})
			results := make(chan commandResult, 2)
			var ready sync.WaitGroup
			ready.Add(2)
			for _, input := range []struct {
				root *expr.RootExpr
				dir  string
			}{{first, t.TempDir()}, {second, t.TempDir()}} {
				go runCommandTogether(input.root, input.dir, command, start, &ready, results)
			}
			ready.Wait()
			close(start)
			firstResult := <-results
			secondResult := <-results
			require.NoError(t, firstResult.err)
			require.NoError(t, secondResult.err)
			if !reflect.DeepEqual(firstResult.files, firstExpected) {
				firstResult, secondResult = secondResult, firstResult
			}
			require.Equal(t, firstExpected, firstResult.files)
			require.Equal(t, secondExpected, secondResult.files)
		})
	}
}

// observedGenerators wraps each selected generator and records the Plan passed
// to the function that chooses names and the function that builds files.
func observedGenerators(factories []generatorFactory, calls map[string]*generatorCall) []generatorFactory {
	observed := make([]generatorFactory, len(factories))
	for index, factory := range factories {
		generator := factory()
		call := &generatorCall{}
		calls[generator.name] = call
		observed[index] = observedGenerator(generator, call)
	}
	return observed
}

// observedGenerator returns a factory that records calls before running the
// selected generator's original functions.
func observedGenerator(generator coreGenerator, call *generatorCall) generatorFactory {
	return func() coreGenerator {
		return coreGenerator{
			name: generator.name,
			Plan: func(plan *Plan) error {
				call.planCalls++
				call.planned = plan
				return generator.Plan(plan)
			},
			Generate: func(plan *Plan) ([]*codegen.File, error) {
				call.generateCalls++
				call.generated = plan
				return generator.Generate(plan)
			},
		}
	}
}

// mapKeys returns the generator names recorded by one command.
func mapKeys(calls map[string]*generatorCall) []string {
	keys := make([]string, 0, len(calls))
	for key := range calls {
		keys = append(keys, key)
	}
	return keys
}

// renderCommand builds and writes every file for one command.
func renderCommand(command string, root *expr.RootExpr, dir string) (map[string][]byte, error) {
	run, err := newGenerationRun(command, newDefaultRegistry())
	if err != nil {
		return nil, err
	}
	result, err := run.execute("generated.local/gen", []eval.Root{root})
	if err != nil {
		return nil, err
	}
	files, err := mergeFilesByPath(result.files)
	if err != nil {
		return nil, err
	}
	written := make(map[string][]byte, len(files))
	for _, file := range files {
		filename, err := file.Render(dir)
		if err != nil {
			return nil, err
		}
		content, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		relative, err := filepath.Rel(dir, filename)
		if err != nil {
			return nil, err
		}
		written[filepath.ToSlash(relative)] = content
	}
	return written, nil
}

// runCommandTogether waits until both command runs are ready, then builds and
// writes one command's files.
func runCommandTogether(root *expr.RootExpr, dir, command string, start <-chan struct{}, ready *sync.WaitGroup, results chan<- commandResult) {
	ready.Done()
	<-start
	files, err := renderCommand(command, root, dir)
	results <- commandResult{files: files, err: err}
}

// commandIsolationDSL defines one HTTP service whose names identify its output.
func commandIsolationDSL(name string) func() {
	return func() {
		serviceName := name + " service"
		dsl.API(name, func() {
			dsl.Server(name, func() {
				dsl.Services(serviceName)
				dsl.Host(name, func() {
					dsl.URI("http://localhost")
				})
			})
		})
		dsl.Service(serviceName, func() {
			dsl.Method("show", func() {
				dsl.Payload(func() {
					dsl.Attribute("message", dsl.String)
				})
				dsl.Result(dsl.String)
				dsl.HTTP(func() {
					dsl.POST("/show")
				})
			})
		})
	}
}
