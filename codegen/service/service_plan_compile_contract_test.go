// This file compiles generated service, views, and starter implementation
// packages for nested validation and generated-name collisions.
package service

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service/testdata"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

type (
	serviceModuleTemplate struct {
		goMod []byte
		goSum []byte
	}

	compiledServiceCase struct {
		name       string
		files      []*codegen.File
		additional map[string]string
	}
)

const serviceModuleSeed = `package seed

import (
	_ "goa.design/clue/debug"
	_ "goa.design/clue/health"
	_ "goa.design/clue/interceptors"
	_ "goa.design/clue/log"
	_ "goa.design/goa/v3/codegen/service/testdata/external"
	_ "goa.design/goa/v3/codegen/service/testdata/external-accessibility"
	_ "goa.design/goa/v3/codegen/service/testdata/external-union"
	_ "goa.design/goa/v3/codegen/service/testdata/nested-alpha"
	_ "goa.design/goa/v3/codegen/service/testdata/nested-beta"
	_ "goa.design/goa/v3/codegen/service/testdata/nested-outer"
	_ "goa.design/goa/v3/pkg"
)
`

var (
	parallelServiceTests sync.Map
	cachedServiceModule  = sync.OnceValues(func() (serviceModuleTemplate, error) {
		goaRoot, err := filepath.Abs(filepath.Join("..", ".."))
		if err != nil {
			return serviceModuleTemplate{}, err
		}
		tmpDir, err := os.MkdirTemp("", "goa-service-mod-template-*")
		if err != nil {
			return serviceModuleTemplate{}, err
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		module := "module generated.local\n\ngo 1.25.0\n\n" +
			"require (\n\tgoa.design/goa/v3 v3.0.0\n\tgoa.design/clue v1.2.6\n)\n\n" +
			"replace goa.design/goa/v3 => " + filepath.ToSlash(goaRoot) + "\n"
		if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(module), 0o600); err != nil {
			return serviceModuleTemplate{}, err
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "seed.go"), []byte(serviceModuleSeed), 0o600); err != nil {
			return serviceModuleTemplate{}, err
		}
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir = tmpDir
		cmd.Env = append(os.Environ(), "GOWORK=off")
		if out, err := cmd.CombinedOutput(); err != nil {
			return serviceModuleTemplate{}, &os.PathError{Op: "go mod tidy", Path: string(out), Err: err}
		}
		goMod, err := os.ReadFile(filepath.Join(tmpDir, "go.mod"))
		if err != nil {
			return serviceModuleTemplate{}, err
		}
		goSum, err := os.ReadFile(filepath.Join(tmpDir, "go.sum"))
		if err != nil {
			return serviceModuleTemplate{}, err
		}
		return serviceModuleTemplate{goMod: goMod, goSum: goSum}, nil
	})
)

func parallelServiceTest(t *testing.T) {
	t.Helper()
	if _, loaded := parallelServiceTests.LoadOrStore(t, struct{}{}); !loaded {
		t.Parallel()
	}
}

func compileGeneratedServiceCases(t *testing.T, cases []compiledServiceCase) {
	t.Helper()
	parallelServiceTest(t)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			compileGeneratedServiceFilesWith(t, tc.files, tc.additional)
		})
	}
}

// TestServicePackageNameUsesClaimedImportPath verifies service package names
// remain lowercase after the final generated import path is claimed.
func TestServicePackageNameUsesClaimedImportPath(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		dsl.Service("api_key_service", func() {
			dsl.Method("Read", func() {
				dsl.Payload(func() {
					dsl.OneOf("credential", func() {
						dsl.Attribute("api_key", dsl.String)
						dsl.Attribute("token", dsl.String)
					})
					dsl.Required("credential")
				})
			})
		})
	})
	plan := retainedServicePlanForPackage(t, root)
	require.Equal(t, "apikeyservice", plan.Services().Get("api_key_service").PkgName)

	files, err := Files(plan)
	require.NoError(t, err)
	directory := t.TempDir()
	for _, file := range files {
		_, err := file.Render(directory)
		require.NoError(t, err)
	}
	source, err := os.ReadFile(filepath.Join(directory, codegen.Gendir, "api_key_service", "service.go"))
	require.NoError(t, err)
	require.Contains(t, string(source), "package apikeyservice")
	unionSource, err := os.ReadFile(filepath.Join(directory, codegen.Gendir, "api_key_service", "unions.go"))
	require.NoError(t, err)
	require.Contains(t, string(unionSource), "package apikeyservice")
}

// TestRepeatedInlineMethodErrorsCompile verifies equivalent method errors use
// one generated public error declaration.
func TestRepeatedInlineMethodErrorsCompile(t *testing.T) {
	root := codegen.RunDSL(t, testdata.RepeatedInlineErrorsDSL)
	plan := retainedServicePlanForPackage(t, root)
	files, err := Files(plan)
	require.NoError(t, err)
	rendered := renderedServiceFiles(t, files)
	serviceSource := string(rendered[filepath.Join(codegen.Gendir, "secured", "service.go")])
	require.Equal(t, 1, strings.Count(serviceSource, "type InvalidScopes string"))
	compileGeneratedServiceFiles(t, files)
}

// TestExtendedLocatedUnionBranchAliasesCompile verifies that Goa writes each
// compiler-created branch type once when several relocated payloads extend the
// same OneOf owner.
func TestExtendedLocatedUnionBranchAliasesCompile(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		owner := dsl.Type("RunOwner", func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.Meta("type:generate:force")
			dsl.OneOf("owner", func() {
				dsl.Attribute("session_id", dsl.String, func() {
					dsl.MinLength(1)
				})
				dsl.Attribute("task_id", dsl.String, func() {
					dsl.MinLength(1)
				})
			})
			dsl.Required("owner")
		})
		first := dsl.Type("FirstPayload", func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.Extend(owner)
			dsl.Attribute("first", dsl.String)
		})
		second := dsl.Type("SecondPayload", func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.Extend(owner)
			dsl.Attribute("second", dsl.String)
		})
		dsl.Service("Runs", func() {
			dsl.Method("First", func() {
				dsl.Payload(first)
			})
			dsl.Method("Second", func() {
				dsl.Payload(second)
			})
		})
	})

	plan := retainedServicePlanForPackage(t, root)
	files, err := Files(plan)
	require.NoError(t, err)
	rendered := renderedServiceFiles(t, files)
	for _, name := range []string{"owner_session_id.go", "owner_task_id.go", "unions.go"} {
		path := filepath.Join(codegen.Gendir, "types", name)
		source, ok := rendered[path]
		require.True(t, ok, "generated file %s", path)
		testutil.AssertGo(
			t,
			filepath.Join("testdata", "golden", "service_extended_located_union_"+name+".golden"),
			string(source),
		)
	}
	compileGeneratedServiceFiles(t, files)
}

// TestNestedViewValidatorCollisionCompiles catches a parent validator that
// reconstructs its child's preferred name after the child function was
// suffixed by another projected declaration in the views package.
func TestNestedViewValidatorCollisionCompiles(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		child := dsl.ResultType("application/vnd.child", func() {
			dsl.TypeName("Child")
			dsl.Attribute("name", dsl.String, func() {
				dsl.MinLength(1)
			})
			dsl.Required("name")
			dsl.View("default", func() {
				dsl.Attribute("name")
			})
		})
		collision := dsl.Type("ValidateChild", func() {
			dsl.Attribute("value", dsl.String)
			dsl.Required("value")
		})
		parent := dsl.ResultType("application/vnd.parent", func() {
			dsl.TypeName("Parent")
			dsl.Attribute("child", child)
			dsl.Attribute("children", dsl.CollectionOf(child))
			dsl.Attribute("validator_name_collision", collision)
			dsl.Required("child", "children", "validator_name_collision")
			dsl.View("default", func() {
				dsl.Attribute("child")
				dsl.Attribute("children")
				dsl.Attribute("validator_name_collision")
			})
		})
		dsl.Service("Values", func() {
			dsl.Method("Read", func() {
				dsl.Result(parent)
			})
		})
	})

	plan := retainedServicePlanForPackage(t, root)
	data := plan.Services().Get("Values")
	var childValidation, parentValidation *ValidateData
	for _, projected := range data.projectedTypes {
		for _, validation := range projected.Validations {
			switch projected.Name {
			case "ChildView":
				childValidation = validation
			case "ParentView":
				parentValidation = validation
			}
		}
	}
	require.NotNil(t, childValidation)
	require.NotNil(t, parentValidation)
	require.NotEmpty(t, parentValidation.Calls)
	require.Same(t, childValidation.Declaration, parentValidation.Calls[0].Declaration)
	require.Equal(t, "ValidateChildView2", childValidation.Declaration.Name())

	files, err := Files(plan)
	require.NoError(t, err)
	files = append(files, ExampleServiceFiles(plan)...)
	compileGeneratedServiceFiles(t, files)
}

// TestMixedResultStarterCompiles checks that a fresh starter implements the
// service method that returns one normal result and may also send stream values.
func TestMixedResultStarterCompiles(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"result and stream", testdata.MixedResultsEndpointDSL},
		{"result view and stream", testdata.MixedResultsWithViewsEndpointDSL},
	}
	compiled := make([]compiledServiceCase, 0, len(cases))
	for _, c := range cases {
		root := codegen.RunDSL(t, c.DSL)
		plan := retainedServicePlanForPackage(t, root)
		files, err := Files(plan)
		require.NoError(t, err)
		files = append(files, ExampleServiceFiles(plan)...)
		compiled = append(compiled, compiledServiceCase{name: c.Name, files: files})
	}
	compileGeneratedServiceCases(t, compiled)
}

// retainedServicePlanForPackage builds and links service generation data using
// the import path shared by these compilation tests.
func retainedServicePlanForPackage(t *testing.T, root *expr.RootExpr) *Plan {
	t.Helper()
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	plan, err := NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, plan.Link())
	return plan
}

// compileGeneratedServiceFiles renders files into a temporary module and runs
// the Go compiler against every generated package.
func compileGeneratedServiceFiles(t *testing.T, files []*codegen.File) {
	compileGeneratedServiceFilesWith(t, files, nil)
}

// compileGeneratedServiceFilesWith renders files and additional test source
// into a temporary module, then runs every generated package test.
func compileGeneratedServiceFilesWith(t *testing.T, files []*codegen.File, additional map[string]string) {
	t.Helper()
	directory := t.TempDir()
	tmpl, err := cachedServiceModule()
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(directory, "go.mod"), tmpl.goMod, 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(directory, "go.sum"), tmpl.goSum, 0o600))
	for _, file := range files {
		_, err := file.Render(directory)
		require.NoError(t, err)
	}
	for path, source := range additional {
		fullPath := filepath.Join(directory, path)
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0o700))
		require.NoError(t, os.WriteFile(fullPath, []byte(source), 0o600))
	}

	parallelServiceTest(t)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	args := []string{"build", "-mod=readonly", "./..."}
	if len(additional) > 0 {
		args = []string{"test", "-mod=readonly", "./..."}
	}
	command := exec.CommandContext(ctx, "go", args...)
	command.Dir = directory
	command.Env = append(os.Environ(), "GOWORK=off")
	output, err := command.CombinedOutput()
	require.NoError(t, err, "compile generated packages:\n%s", output)
}
