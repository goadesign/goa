// This file checks that each example generation keeps its copied server data
// separate from every other generation.
package example

import (
	"bytes"
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example/testdata"
	"goa.design/goa/v3/codegen/service"
	dsl "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestPlansKeepSameNamedServersSeparate(t *testing.T) {
	httpRoot := codegen.RunDSL(t, testdata.ServiceForOnlyHTTPDSL)
	grpcRoot := codegen.RunDSL(t, testdata.ServiceForOnlyGRPCDSL)

	httpGeneration, err := codegen.NewGeneration("example.local/http/gen", []eval.Root{httpRoot})
	require.NoError(t, err)
	grpcGeneration, err := codegen.NewGeneration("example.local/grpc/gen", []eval.Root{grpcRoot})
	require.NoError(t, err)

	httpService, err := service.NewPlan(httpRoot, httpGeneration, expr.NewExampleGenerator(httpRoot.API.RandomizerFactory))
	require.NoError(t, err)
	httpPlan, err := NewPlan(httpGeneration, httpService)
	require.NoError(t, err)
	httpData, ok := httpPlan.Root(httpService)
	require.True(t, ok)
	httpServer := httpData.Servers[0]
	require.True(t, httpServer.HasHTTP)
	require.False(t, httpServer.HasTransport(TransportGRPC))

	grpcService, err := service.NewPlan(grpcRoot, grpcGeneration, expr.NewExampleGenerator(grpcRoot.API.RandomizerFactory))
	require.NoError(t, err)
	grpcPlan, err := NewPlan(grpcGeneration, grpcService)
	require.NoError(t, err)
	grpcData, ok := grpcPlan.Root(grpcService)
	require.True(t, ok)
	grpcServer := grpcData.Servers[0]
	require.False(t, grpcServer.HasHTTP)
	require.True(t, grpcServer.HasTransport(TransportGRPC))

	require.True(t, httpServer.HasHTTP)
	require.False(t, httpServer.HasTransport(TransportGRPC))
}

func TestPlanKeepsHostVariableDefaultAndAllowedValues(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		dsl.API("host variables", func() {
			dsl.Server("public", func() {
				dsl.Services("status")
				dsl.Host("production", func() {
					dsl.URI("https://{region}.example.com")
					dsl.Variable("region", dsl.String, func() {
						dsl.Default("west")
						dsl.Enum("west", "east")
					})
				})
			})
		})
		dsl.Service("status", func() {
			dsl.Method("read", func() {
				dsl.HTTP(func() {
					dsl.GET("/status")
				})
			})
		})
	})
	generation, err := codegen.NewGeneration("example.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	plan, err := NewPlan(generation, servicePlan)
	require.NoError(t, err)
	plannedRoot, ok := plan.Root(servicePlan)
	require.True(t, ok)

	variable := plannedRoot.Servers[0].Hosts[0].Variables[0]
	require.Equal(t, "west", variable.DefaultValue)
	require.Equal(t, []string{"west", "east"}, variable.Values)
}

// TestPlanUsesURLRoleWhenAHostVariableMatchesABuiltInFlag checks that a
// generated flag says what it configures instead of receiving a number.
func TestPlanUsesURLRoleWhenAHostVariableMatchesABuiltInFlag(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		dsl.API("host variable collision", func() {
			dsl.Server("public", func() {
				dsl.Services("status")
				dsl.Host("production", func() {
					dsl.URI("https://{host}.example.com")
					dsl.Variable("host", dsl.String, func() {
						dsl.Default("west")
						dsl.Enum("west", "east")
					})
				})
			})
		})
		dsl.Service("status", func() {
			dsl.Method("read", func() {
				dsl.HTTP(func() {
					dsl.GET("/status")
				})
			})
		})
	})
	generation, err := codegen.NewGeneration("example.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	plan, err := NewPlan(generation, servicePlan)
	require.NoError(t, err)
	plannedRoot, ok := plan.Root(servicePlan)
	require.True(t, ok)

	plannedClient := planClientMainServer(plannedRoot.Servers[0])
	variable := plannedClient.Variables[0]
	require.Equal(t, "url-host", variable.FlagName)
	require.Equal(t, "urlHostF", variable.VarName)
	require.Same(t, variable, plannedClient.Hosts[0].Variables[0])

	plannedServer := planMainVariables(plannedRoot.Servers[0].Variables, []string{"host"})
	require.Equal(t, "url-host", plannedServer.all[0].FlagName)
	require.Equal(t, "urlHostF", plannedServer.all[0].VarName)
}

// TestPlanFindsRootForExactServicePlan checks that copied server data belongs
// only to the service plan from which it was built.
func TestPlanFindsRootForExactServicePlan(t *testing.T) {
	firstRoot := codegen.RunDSL(t, testdata.ServiceForOnlyHTTPDSL)
	secondRoot := codegen.RunDSL(t, testdata.ServiceForOnlyHTTPDSL)
	firstGeneration, err := codegen.NewGeneration("example.local/first/gen", []eval.Root{firstRoot})
	require.NoError(t, err)
	firstService, err := service.NewPlan(firstRoot, firstGeneration, expr.NewExampleGenerator(firstRoot.API.RandomizerFactory))
	require.NoError(t, err)
	secondGeneration, err := codegen.NewGeneration("example.local/second/gen", []eval.Root{secondRoot})
	require.NoError(t, err)
	secondService, err := service.NewPlan(secondRoot, secondGeneration, expr.NewExampleGenerator(secondRoot.API.RandomizerFactory))
	require.NoError(t, err)
	plan, err := NewPlan(firstGeneration, firstService)
	require.NoError(t, err)

	root, ok := plan.Root(firstService)
	require.True(t, ok)
	require.Equal(t, firstRoot.API.Name, root.APIName)
	_, ok = plan.Root(secondService)
	require.False(t, ok)
}

func TestPlansBuildConcurrently(t *testing.T) {
	httpRoot := codegen.RunDSL(t, testdata.ServiceForOnlyHTTPDSL)
	grpcRoot := codegen.RunDSL(t, testdata.ServiceForOnlyGRPCDSL)
	httpGeneration, err := codegen.NewGeneration("example.local/http/gen", []eval.Root{httpRoot})
	require.NoError(t, err)
	grpcGeneration, err := codegen.NewGeneration("example.local/grpc/gen", []eval.Root{grpcRoot})
	require.NoError(t, err)
	httpService, err := service.NewPlan(httpRoot, httpGeneration, expr.NewExampleGenerator(httpRoot.API.RandomizerFactory))
	require.NoError(t, err)
	grpcService, err := service.NewPlan(grpcRoot, grpcGeneration, expr.NewExampleGenerator(grpcRoot.API.RandomizerFactory))
	require.NoError(t, err)

	start := make(chan struct{})
	var (
		plans [2]*Plan
		errs  [2]error
		ready sync.WaitGroup
		wait  sync.WaitGroup
	)
	build := func(index int, generation *codegen.Generation, servicePlan *service.Plan) {
		defer wait.Done()
		ready.Done()
		<-start
		plans[index], errs[index] = NewPlan(generation, servicePlan)
	}
	ready.Add(2)
	wait.Add(2)
	go build(0, httpGeneration, httpService)
	go build(1, grpcGeneration, grpcService)
	ready.Wait()
	close(start)
	wait.Wait()

	require.NoError(t, errs[0])
	require.NoError(t, errs[1])
	httpData, ok := plans[0].Root(httpService)
	require.True(t, ok)
	grpcData, ok := plans[1].Root(grpcService)
	require.True(t, ok)
	require.True(t, httpData.Servers[0].HasHTTP)
	require.False(t, grpcData.Servers[0].HasHTTP)
}

// TestPlanCopiesEveryServerValue checks that example output does not keep a
// path back to the design values it copied.
func TestPlanCopiesEveryServerValue(t *testing.T) {
	root := codegen.RunDSL(t, testdata.ServiceForOnlyHTTPDSL)
	generation, err := codegen.NewGeneration("example.local/http/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	plan, err := NewPlan(generation, servicePlan)
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())

	copied, ok := plan.Root(servicePlan)
	require.True(t, ok)
	server := root.API.Servers[0]
	require.False(t, pointsToDesignValue(reflect.ValueOf(copied), map[uintptr]struct{}{
		reflect.ValueOf(root).Pointer():   {},
		reflect.ValueOf(server).Pointer(): {},
	}, make(map[uintptr]struct{})))

	files := CLIFiles(copied)
	require.NotEmpty(t, files)
	before := renderExampleSections(t, files[0])
	root.API.Name = "changed api"
	server.Name = "changed server"
	server.Description = "changed description"
	server.Services = nil
	server.Hosts = nil
	require.Equal(t, before, renderExampleSections(t, files[0]))
}

// renderExampleSections writes the complete file without touching the file
// system so a test can compare the exact generated text.
func renderExampleSections(t *testing.T, file *codegen.File) string {
	t.Helper()
	var output bytes.Buffer
	for _, section := range file.SectionTemplates {
		require.NoError(t, section.Write(&output))
	}
	return output.String()
}

// pointsToDesignValue reports whether value contains one of the original
// design pointers. visited prevents loops in linked values.
func pointsToDesignValue(value reflect.Value, targets, visited map[uintptr]struct{}) bool {
	if !value.IsValid() {
		return false
	}
	switch value.Kind() {
	case reflect.Interface:
		return pointsToDesignValue(value.Elem(), targets, visited)
	case reflect.Pointer:
		pointer := value.Pointer()
		if _, ok := targets[pointer]; ok {
			return true
		}
		if _, ok := visited[pointer]; ok {
			return false
		}
		visited[pointer] = struct{}{}
		return pointsToDesignValue(value.Elem(), targets, visited)
	case reflect.Map:
		for iterator := value.MapRange(); iterator.Next(); {
			if pointsToDesignValue(iterator.Key(), targets, visited) ||
				pointsToDesignValue(iterator.Value(), targets, visited) {
				return true
			}
		}
	case reflect.Slice, reflect.Array:
		for index := 0; index < value.Len(); index++ {
			if pointsToDesignValue(value.Index(index), targets, visited) {
				return true
			}
		}
	case reflect.Struct:
		for index := 0; index < value.NumField(); index++ {
			if pointsToDesignValue(value.Field(index), targets, visited) {
				return true
			}
		}
	}
	return false
}
