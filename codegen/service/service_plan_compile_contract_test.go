// This file compiles generated service, views, and starter implementation
// packages for nested validation collisions shaped like AURA tool contracts.
package service

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service/testdata"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

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
	compileGeneratedServiceFiles(t, files)

	rendered := renderedServiceFiles(t, files)
	serviceSource := string(rendered[filepath.Join(codegen.Gendir, "secured", "service.go")])
	require.Equal(t, 1, strings.Count(serviceSource, "type InvalidScopes string"))
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
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := codegen.RunDSL(t, c.DSL)
			plan := retainedServicePlanForPackage(t, root)
			files, err := Files(plan)
			require.NoError(t, err)
			files = append(files, ExampleServiceFiles(plan)...)
			compileGeneratedServiceFiles(t, files)
		})
	}
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
	goaRoot := serviceModuleDirectory(t, "goa.design/goa/v3")
	module := "module generated.local\n\ngo 1.24\n\n" +
		"require goa.design/goa/v3 v3.0.0\n\n" +
		"replace goa.design/goa/v3 => " + filepath.ToSlash(goaRoot) + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(directory, "go.mod"), []byte(module), 0o600))
	for _, file := range files {
		_, err := file.Render(directory)
		require.NoError(t, err)
	}
	for path, source := range additional {
		fullPath := filepath.Join(directory, path)
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0o700))
		require.NoError(t, os.WriteFile(fullPath, []byte(source), 0o600))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "test", "-mod=mod", "./...")
	command.Dir = directory
	command.Env = append(os.Environ(), "GOWORK=off")
	output, err := command.CombinedOutput()
	require.NoError(t, err, "compile generated packages:\n%s", output)
}

// serviceModuleDirectory resolves the local checkout for a module used by a
// temporary generated module.
func serviceModuleDirectory(t *testing.T, module string) string {
	t.Helper()
	command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", module)
	output, err := command.CombinedOutput()
	require.NoError(t, err, "resolve module %s:\n%s", module, output)
	directory := strings.TrimSpace(string(output))
	require.NotEmpty(t, directory)
	return directory
}
