// This file verifies that viewed HTTP response constructors keep their
// released names while sharing the package name plan with generated callers.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestViewedResultConstructorsKeepReleasedNames(t *testing.T) {
	code := renderClientTypesCode(t, viewedResultConstructorCompatibilityDSL)

	testutil.AssertGo(t, "testdata/golden/viewed_result_constructor_names.go.golden", code)
	require.Contains(t, code, "func NewArchiveMediaViewOK(")
	require.Contains(t, code, "func NewReadArchiveMediaOK(")
	require.NotContains(t, code, "NewArchiveArchiveMediaOK")
}

func TestViewedResultConstructorCallsUsePlannedName(t *testing.T) {
	root := expr.RunDSL(t, viewedResultConstructorCompatibilityDSL)
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	viewsPackage, err := generation.ClaimPackage("generated.local/gen/archiver/views")
	require.NoError(t, err)
	require.NoError(t, viewsPackage.DeclareName(codegen.NewExactName(codegen.NameType, "ArchiveMediaView")))

	plans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, plans[0].Link())

	archive := plans[0].services.Get("archiver").Endpoint("archive")
	constructor := archive.Result.Responses[0].ResultInit.Declaration
	require.Equal(t, "NewArchiveMediaView2OK", constructor.Name())
	require.Equal(t, constructor.Name(), archive.Result.Responses[0].ResultInit.Name)
	require.Contains(t, renderedFiles(t, plans[0].ClientTypeFiles()), "func NewArchiveMediaView2OK(")
	require.Contains(t, renderedFiles(t, plans[0].ClientFiles()), "NewArchiveMediaView2OK(")
}

// viewedResultConstructorCompatibilityDSL defines one result used by a method
// whose name overlaps the result name and one whose name does not.
func viewedResultConstructorCompatibilityDSL() {
	media := dsl.ResultType("application/vnd.archiver.media", func() {
		dsl.TypeName("ArchiveMedia")
		dsl.Attributes(func() {
			dsl.Attribute("status", dsl.Int)
			dsl.Required("status")
		})
		dsl.View("default", func() {
			dsl.Attribute("status")
		})
	})

	dsl.Service("archiver", func() {
		dsl.Method("archive", func() {
			dsl.Result(media)
			dsl.HTTP(func() {
				dsl.POST("/archive")
			})
		})
		dsl.Method("read", func() {
			dsl.Result(media)
			dsl.HTTP(func() {
				dsl.GET("/archive")
			})
		})
	})
}
