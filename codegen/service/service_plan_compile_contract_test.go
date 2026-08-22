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
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// TestServicePackageNameUsesClaimedImportPath verifies mixed-case service
// names keep the canonical Go casing derived from their claimed package path.
func TestServicePackageNameUsesClaimedImportPath(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		dsl.Service("UnionValidation", func() {
			dsl.Method("Read", func() {})
		})
	})
	plan := retainedServicePlanForPackage(t, root, "generated.local/gen")
	require.Equal(t, "unionValidation", plan.Services().Get("UnionValidation").PkgName)
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

	plan := retainedServicePlanForPackage(t, root, "generated.local/gen")
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
	compileGeneratedServiceFiles(t, "generated.local", files)
}

// retainedServicePlanForPackage runs the service lifecycle with the generated
// import root used by the temporary compilation module.
func retainedServicePlanForPackage(t *testing.T, root *expr.RootExpr, generatedPackage string) *Plan {
	t.Helper()
	generation, err := codegen.NewGeneration(generatedPackage, []eval.Root{root})
	require.NoError(t, err)
	plan, err := NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, plan.Link())
	return plan
}

// compileGeneratedServiceFiles renders files into a temporary module and runs
// the Go compiler against every generated package.
func compileGeneratedServiceFiles(t *testing.T, modulePath string, files []*codegen.File) {
	t.Helper()
	directory := t.TempDir()
	goaRoot := serviceModuleDirectory(t, "goa.design/goa/v3")
	module := "module " + modulePath + "\n\ngo 1.24\n\n" +
		"require goa.design/goa/v3 v3.0.0\n\n" +
		"replace goa.design/goa/v3 => " + filepath.ToSlash(goaRoot) + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(directory, "go.mod"), []byte(module), 0o600))
	for _, file := range files {
		_, err := file.Render(directory)
		require.NoError(t, err)
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
