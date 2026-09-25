// This file checks the original-type catalog through evaluated DSL, the core
// service planner, a planning plugin, and the Go declarations those plans emit.
package generator

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

type (
	// catalogCompanionOrder orders the plugin's functions by original type name.
	catalogCompanionOrder string
)

func TestGenerationUserTypesServiceCatalog(t *testing.T) {
	for _, test := range []struct {
		name    string
		reverse bool
	}{
		{name: "forward services"},
		{name: "reverse services", reverse: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			var leaf, reachable, locatedChild, located, forcedChild, forced expr.UserType
			var report *expr.ResultTypeExpr
			root := codegen.RunDSL(t, func() {
				leaf = dsl.Type("Leaf", func() {
					dsl.Attribute("value", dsl.String)
				})
				reachable = dsl.Type("Reachable", func() {
					dsl.Attribute("leaf", leaf)
					dsl.OneOf("choice", func() {
						dsl.TypeName("Choice")
						dsl.Attribute("text", dsl.String)
						dsl.Attribute("leaf", leaf)
					})
				})
				locatedChild = dsl.Type("LocatedChild", func() {
					dsl.Meta("struct:pkg:path", "shared")
					dsl.Attribute("value", dsl.String)
				})
				located = dsl.Type("Located", func() {
					dsl.Meta("struct:pkg:path", "shared")
					dsl.Attribute("child", locatedChild)
				})
				forcedChild = dsl.Type("ForcedChild", func() {
					dsl.Attribute("value", dsl.String)
				})
				forced = dsl.Type("Forced", func() {
					dsl.Meta("type:generate:force", "Alpha")
					dsl.Attribute("child", forcedChild)
				})
				dsl.Type("Unused", func() {
					dsl.Attribute("value", dsl.String)
				})
				report = dsl.ResultType("application/vnd.catalog.report", func() {
					dsl.TypeName("Report")
					dsl.Attributes(func() {
						dsl.Attribute("title", dsl.String)
						dsl.Attribute("detail", dsl.String)
					})
					dsl.View("default", func() {
						dsl.Attribute("title")
						dsl.Attribute("detail")
					})
					dsl.View("tiny", func() {
						dsl.Attribute("title")
					})
				})
				alpha := func() {
					dsl.Service("Alpha", func() {
						dsl.Method("Read", func() {
							dsl.Payload(func() {
								dsl.Attribute("local", reachable)
								dsl.Attribute("shared", located)
							})
							dsl.Result(report)
						})
						dsl.Method("Repeat", func() {
							dsl.Payload(reachable)
							dsl.Result(func() {
								dsl.Attribute("local", reachable)
							})
						})
					})
				}
				beta := func() {
					dsl.Service("Beta", func() {
						dsl.Method("Read", func() {
							dsl.Payload(reachable)
						})
					})
				}
				if test.reverse {
					beta()
					alpha()
				} else {
					alpha()
					beta()
				}
			})
			wantOriginals := []expr.UserType{
				forced, forcedChild, leaf, reachable, report,
				leaf, reachable,
				located, locatedChild,
			}
			wantPackages := []string{
				"generated.local/gen/alpha",
				"generated.local/gen/alpha",
				"generated.local/gen/alpha",
				"generated.local/gen/alpha",
				"generated.local/gen/alpha",
				"generated.local/gen/beta",
				"generated.local/gen/beta",
				"generated.local/gen/shared",
				"generated.local/gen/shared",
			}
			var originals []expr.UserType
			var declarations, excluded []*codegen.TypeDeclaration
			var companions []*codegen.NameDeclaration
			registry := newRegistry()
			registry.addCommand("gen", genGeneratorFactories()...)
			registry.registerPlugin("original-catalog", "gen", pluginNormal, func() Plugin {
				return Plugin{
					Plan: func(plan *Plan) error {
						generation := plan.Generation()
						require.False(t, generation.Frozen())
						require.Panics(t, func() {
							plan.Service(root).Services()
						})
						for original, declaration := range generation.UserTypes() {
							originals = append(originals, original)
							declarations = append(declarations, declaration)
							owner := generation.Package(declaration.PackagePath())
							companion, err := owner.DeclareDependentName(
								codegen.NameFunction, declaration.Declaration(), "Inspect", "",
								catalogCompanionOrder(original.Name()),
							)
							require.NoError(t, err)
							companions = append(companions, companion)
						}

						alpha := generation.Package("generated.local/gen/alpha")
						for _, attribute := range []*expr.AttributeExpr{
							root.Service("Alpha").Method("Read").Payload,
							root.Service("Alpha").Method("Repeat").Result,
						} {
							wrapper, ok := attribute.Type.(expr.UserType)
							require.True(t, ok)
							_, normalized := generation.NormalizedMethodType(wrapper)
							require.True(t, normalized)
							declaration, err := alpha.Type(wrapper)
							require.NoError(t, err)
							excluded = append(excluded, declaration)
						}
						views := generation.Package("generated.local/gen/alpha/views")
						for _, identity := range []codegen.DerivedTypeID{
							codegen.NewProjectedTypeID(report),
							codegen.NewViewedResultTypeID(report),
						} {
							declaration, err := views.DerivedType(identity)
							require.NoError(t, err)
							excluded = append(excluded, declaration)
						}
						branch, err := alpha.UnionBranchType(reachable.Attribute().Find("choice"), "text")
						require.NoError(t, err)
						excluded = append(excluded, branch)
						return nil
					},
				}
			})
			run, err := newGenerationRun("gen", registry)
			require.NoError(t, err)
			result, err := run.execute("generated.local/gen", []eval.Root{root})
			require.NoError(t, err)
			require.Len(t, originals, len(wantOriginals))
			for i := range wantOriginals {
				require.Same(t, wantOriginals[i].Origin(), originals[i])
				require.Equal(t, wantPackages[i], declarations[i].PackagePath())
				require.Equal(t, "Inspect"+declarations[i].Name(), companions[i].Name())
			}

			generation := result.plan.Generation()
			index := 0
			for original, declaration := range generation.UserTypes() {
				require.Less(t, index, len(declarations))
				require.Same(t, originals[index], original)
				require.Same(t, declarations[index], declaration)
				index++
			}
			require.Equal(t, len(declarations), index)
			emitted := catalogEmittedTypes(t, generation, result.files)
			for _, declaration := range declarations {
				require.Contains(t, emitted[declaration.PackagePath()], declaration.Name())
			}
			require.Len(t, excluded, 5)
			for _, declaration := range excluded {
				for _, original := range declarations {
					require.NotSame(t, original, declaration)
				}
				require.Contains(t, emitted[declaration.PackagePath()], declaration.Name())
			}
			for _, names := range emitted {
				require.NotContains(t, names, "Unused")
			}
			require.NotContains(t, emitted["generated.local/gen/beta"], "Forced")
			require.NotContains(t, emitted["generated.local/gen/beta"], "ForcedChild")
		})
	}
}

func TestGenerationUserTypesForceNeedsServiceContext(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		dsl.Type("Forced", func() {
			dsl.Meta("type:generate:force")
			dsl.Meta("struct:pkg:path", "shared")
			dsl.Attribute("value", dsl.String)
		})
	})
	registry := newRegistry()
	registry.addCommand("gen", genGeneratorFactories()...)
	run, err := newGenerationRun("gen", registry)
	require.NoError(t, err)
	result, err := run.execute("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	count := 0
	for range result.plan.Generation().UserTypes() {
		count++
	}
	require.Zero(t, count)
}

// ComparePackageName keeps companion names independent of service visit order.
func (o catalogCompanionOrder) ComparePackageName(other codegen.PackageNameOrder) int {
	return strings.Compare(string(o), string(other.(catalogCompanionOrder)))
}

// catalogEmittedTypes reads the generated Go declarations so catalog membership
// is checked against actual service output, including the excluded wrappers.
func catalogEmittedTypes(t *testing.T, generation *codegen.Generation, files []*codegen.File) map[string][]string {
	t.Helper()
	merged, err := mergeFilesByPath(files)
	require.NoError(t, err)
	emitted := make(map[string][]string)
	for _, file := range merged {
		if filepath.Ext(file.Path) != ".go" {
			continue
		}
		var output bytes.Buffer
		for _, section := range file.SectionTemplates {
			require.NoError(t, section.Write(&output))
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), file.Path, output.Bytes(), 0)
		require.NoError(t, err)
		owner, ok := generation.PackageForFile(file.Path)
		require.True(t, ok, "unowned generated file %s", file.Path)
		for _, declaration := range parsed.Decls {
			group, ok := declaration.(*ast.GenDecl)
			if !ok || group.Tok != token.TYPE {
				continue
			}
			for _, spec := range group.Specs {
				named := spec.(*ast.TypeSpec)
				emitted[owner.ImportPath()] = append(emitted[owner.ImportPath()], named.Name.Name)
			}
		}
	}
	return emitted
}
