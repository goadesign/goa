// This file verifies that an HTTP server-sent event encodes and decodes the
// body selected for its result view.
package codegen

import (
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// TestViewedSSEServerLocksFirstView verifies that a request-scoped stream
// rejects a later representation before encoding a body under the first view.
func TestViewedSSEServerLocksFirstView(t *testing.T) {
	root := expr.RunDSL(t, viewedSSEDSL)
	plan := linkedHTTPPlanForRoot(t, root)
	code := renderedFile(t, plan.ServerFiles())

	require.Contains(t, code, `if s.sentView != "" && view != s.sentView`)
	require.Contains(t, code, `goa.InvalidEnumValueError("view", view, []any{s.sentView})`)
	start := strings.LastIndex(code, `s.start(view)`)
	require.NotEqual(t, -1, start)
	require.Less(t,
		strings.Index(code, `if s.sentView != "" && view != s.sentView`),
		start,
	)
	require.Less(t, strings.Index(code, "res := "), start)
	require.Less(t, strings.Index(code, "body := "), start)
	require.Less(t, strings.Index(code, `s.sentView = view`), start)
}

// TestViewedSSEClientReconstructsCollections verifies collection events decode
// the selected body, run its constructor and validator, and return the service
// method's result type.
func TestViewedSSEClientReconstructsCollections(t *testing.T) {
	root := expr.RunDSL(t, viewedSSECollectionDSL)
	plan := linkedHTTPPlanForRoot(t, root)
	code := renderedFile(t, plan.ClientFiles())

	require.Contains(t, code, "switch view {")
	require.Contains(t, code, "Decode(&body)")
	require.Contains(t, code, "projected := New")
	require.Contains(t, code, "views.Validate")
	require.Contains(t, code, "result := viewedssecollection.New")
	require.NotContains(t, code, `partial_sse_parse`)
}

// TestViewedSSEUsesConfiguredDataField verifies the server encodes and the
// client decodes only the result field selected as the event data.
func TestViewedSSEUsesConfiguredDataField(t *testing.T) {
	root := expr.RunDSL(t, viewedSSEDataFieldDSL)
	plan := linkedHTTPPlanForRoot(t, root)
	client := renderedFile(t, plan.ClientFiles())
	server := renderedFile(t, plan.ServerFiles())

	require.Contains(t, client, "value := dataContent")
	require.Contains(t, client, "body.Data = &value")
	require.Contains(t, client, "projected := New")
	require.Contains(t, client, "views.Validate")
	require.Contains(t, server, "data = string(body.Data)")
	require.NotContains(t, server, "var payload any")
}

// TestViewedSSERebuildsRequiredResponseFields checks that the client reads the
// event id, event type, and data before it calls the generated result
// constructor and validator.
func TestViewedSSERebuildsRequiredResponseFields(t *testing.T) {
	root := expr.RunDSL(t, viewedSSERequiredFieldsDSL)
	plan := linkedHTTPPlanForRoot(t, root)
	client := renderedFile(t, plan.ClientFiles())
	service := plan.services.Get("Viewed SSE Required")
	require.Len(t, service.Endpoints, 1)
	require.Len(t, service.Endpoints[0].Result.Responses, 1)
	representations := service.Endpoints[0].Result.Responses[0].ViewedRepresentations
	require.Len(t, representations, 3)
	for _, representation := range representations {
		require.NotNil(t, representation.ClientBody)
		require.NotNil(t, representation.ClientBody.ValidatorDeclaration)
		require.NotEmpty(t, representation.ClientBody.ValidationTarget)
	}

	for _, assignment := range []string{
		"idContent = s.lastEventID",
		"eventContent = string(value)",
		"body.ID = &idContent",
		"body.Kind = &eventContent",
		"value := dataContent",
		"body.Data = &value",
	} {
		require.Contains(t, client, assignment)
		require.Less(t, strings.Index(client, assignment), strings.Index(client, "projected := New"))
	}
	require.NotContains(t, client, "result.ID =")
	require.NotContains(t, client, "result.Kind =")
	require.Less(t, strings.Index(client, "projected := New"), strings.Index(client, "views.Validate"))
}

// TestViewedClientsUseAssignedValidator checks that each HTTP client calls the
// validator name chosen for the service views package when another declaration
// requests the validator's preferred spelling.
func TestViewedClientsUseAssignedValidator(t *testing.T) {
	root := expr.RunDSL(t, viewedClientValidatorCollisionDSL)
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	viewsPackage, err := generation.ClaimPackage(path.Join(generation.GenPkg(), "viewed_validator", "views"))
	require.NoError(t, err)
	preferred := "ValidateViewedClientCollision"
	require.NoError(t, viewsPackage.DeclareName(codegen.NewExactName(codegen.NameFunction, preferred)))
	plans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, plans[0].Link())

	var declaration *codegen.NameDeclaration
	for _, endpoint := range plans[0].services.Get("Viewed Validator").Endpoints {
		if endpoint.Method.ViewedResult != nil {
			declaration = endpoint.Method.ViewedResult.Validate.Declaration
			break
		}
	}
	require.NotNil(t, declaration)
	require.NotEqual(t, preferred, declaration.Name())
	client := renderedFiles(t, plans[0].ClientFiles())
	require.GreaterOrEqual(t, strings.Count(client, "."+declaration.Name()+"("), 3)
	require.NotContains(t, client, "."+preferred+"(")
}

// TestViewedSSESoleViewIsFixed checks that the generated service supplies its
// only legal view, so the HTTP response needs no view selector.
func TestViewedSSESoleViewIsFixed(t *testing.T) {
	plan := linkedHTTPPlan(t, viewedSSESoleViewDSL)
	viewed, ok := plan.ViewedResult("Viewed SSE Sole View", "Watch")
	require.True(t, ok)

	require.False(t, viewed.Variable)
	require.Equal(t, expr.DefaultView, viewed.FixedView)
	require.Len(t, viewed.Representations, 1)
}

// TestViewedResultConstructorsUsePackageDeclarations verifies Go-equivalent
// view names receive distinct stable functions and every definition and call
// uses the same package-owned declaration.
func TestViewedResultConstructorsUsePackageDeclarations(t *testing.T) {
	root := expr.RunDSL(t, viewedSSEConstructorCollisionDSL)
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	plans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, plans[0].Link())

	endpoint := root.API.HTTP.Services[0].HTTPEndpoints[0]
	response := endpoint.Responses[0]
	retained, ok := plans[0].ViewedResult("Viewed SSE Collision", "Watch")
	require.True(t, ok)
	require.GreaterOrEqual(t, len(retained.Representations), 2)
	names := make(map[string]struct{}, len(retained.Representations))
	collidingNames := make(map[string]string, 2)
	definitions := renderedFiles(t, plans[0].ClientTypeFiles())
	calls := renderedFile(t, plans[0].ClientFiles())
	for _, representation := range retained.Representations {
		declaration := plans[0].constructors[viewedConstructorKey{
			endpoint: endpoint,
			response: response,
			view:     representation.View,
		}]
		require.Same(t, declaration, representation.ResultInit.Declaration)
		name := declaration.Name()
		names[name] = struct{}{}
		if representation.View == "foo-bar" || representation.View == "foo bar" {
			collidingNames[representation.View] = name
		}
		require.Contains(t, definitions, "func "+name+"(")
		require.Contains(t, calls, name+"(")
	}
	require.Len(t, names, len(retained.Representations))
	require.NotEqual(t, collidingNames["foo-bar"], collidingNames["foo bar"])
}

// renderedFile renders the generated sse.go file.
func renderedFile(t *testing.T, files []*codegen.File) string {
	t.Helper()
	for _, file := range files {
		if strings.HasSuffix(file.Path, "sse.go") {
			return renderedFiles(t, []*codegen.File{file})
		}
	}
	t.Error("generated sse.go file was not planned")
	return ""
}

// renderedFiles renders every planned section so tests compare generated Go
// definitions and calls rather than template text.
func renderedFiles(t *testing.T, files []*codegen.File) string {
	t.Helper()
	var rendered strings.Builder
	for _, file := range files {
		for _, section := range file.SectionTemplates[1:] {
			rendered.WriteString(codegen.SectionCode(t, section))
		}
	}
	return rendered.String()
}

// viewedSSEType defines two legal branches with different body shapes.
func viewedSSEType(name string) *expr.ResultTypeExpr {
	return dsl.ResultType("application/vnd."+name, func() {
		dsl.TypeName(name)
		dsl.Attribute("id", dsl.String)
		dsl.Attribute("detail", dsl.String)
		dsl.Required("id")
		dsl.View("summary", func() { dsl.Attribute("id") })
		dsl.View("detailed", func() {
			dsl.Attribute("id")
			dsl.Attribute("detail")
		})
	})
}

// viewedSSEDSL defines a variable-view object stream.
func viewedSSEDSL() {
	event := viewedSSEType("ViewedSSEEvent")
	dsl.Service("Viewed SSE", func() {
		dsl.Method("Watch", func() {
			dsl.StreamingResult(event)
			dsl.HTTP(func() {
				dsl.GET("/watch")
				dsl.ServerSentEvents()
			})
		})
	})
}

// viewedSSECollectionDSL defines a variable-view collection stream.
func viewedSSECollectionDSL() {
	event := viewedSSEType("ViewedSSECollectionEvent")
	dsl.Service("Viewed SSE Collection", func() {
		dsl.Method("Watch", func() {
			dsl.StreamingResult(dsl.CollectionOf(event))
			dsl.HTTP(func() {
				dsl.GET("/watch")
				dsl.ServerSentEvents()
			})
		})
	})
}

// viewedSSEDataFieldDSL defines a variable-view stream whose data line carries
// one configured result field.
func viewedSSEDataFieldDSL() {
	event := dsl.ResultType("application/vnd.viewed-sse-data", func() {
		dsl.TypeName("ViewedSSEData")
		dsl.Attribute("data", dsl.String)
		dsl.Attribute("detail", dsl.String)
		dsl.Required("data")
		dsl.View("summary", func() { dsl.Attribute("data") })
		dsl.View("detailed", func() {
			dsl.Attribute("data")
			dsl.Attribute("detail")
		})
	})
	dsl.Service("Viewed SSE Data", func() {
		dsl.Method("Watch", func() {
			dsl.StreamingResult(event)
			dsl.HTTP(func() {
				dsl.GET("/watch")
				dsl.ServerSentEvents("data")
			})
		})
	})
}

// viewedSSEPrimitiveAliasDataFieldDSL defines a viewed stream whose data line
// carries a required field declared with a named string type.
func viewedSSEPrimitiveAliasDataFieldDSL() {
	text := dsl.Type("ViewedEventText", dsl.String)
	event := dsl.ResultType("application/vnd.viewed-sse-alias-data", func() {
		dsl.TypeName("ViewedSSEAliasData")
		dsl.Attribute("data", text)
		dsl.Attribute("detail", dsl.String)
		dsl.Required("data")
		dsl.View("summary", func() { dsl.Attribute("data") })
		dsl.View("detailed", func() {
			dsl.Attribute("data")
			dsl.Attribute("detail")
		})
	})
	dsl.Service("Viewed SSE Alias Data", func() {
		dsl.Method("Watch", func() {
			dsl.StreamingResult(event)
			dsl.HTTP(func() {
				dsl.GET("/watch")
				dsl.ServerSentEvents("data")
			})
		})
	})
}

// viewedSSERequiredFieldsDSL maps required result fields across every input a
// streamed HTTP response can carry.
func viewedSSERequiredFieldsDSL() {
	event := dsl.ResultType("application/vnd.viewed-sse-required", func() {
		dsl.TypeName("ViewedSSERequired")
		dsl.Attribute("id", dsl.String)
		dsl.Attribute("kind", dsl.String)
		dsl.Attribute("data", dsl.String)
		dsl.Attribute("hidden", dsl.String)
		dsl.Required("id", "kind", "data", "hidden")
		for _, name := range []string{"summary", "detailed"} {
			dsl.View(name, func() {
				dsl.Attribute("id")
				dsl.Attribute("kind")
				dsl.Attribute("data")
			})
		}
	})
	dsl.Service("Viewed SSE Required", func() {
		dsl.Method("Watch", func() {
			dsl.StreamingResult(event)
			dsl.HTTP(func() {
				dsl.GET("/watch")
				dsl.ServerSentEvents("data", func() {
					dsl.SSEEventID("id")
					dsl.SSEEventType("kind")
				})
			})
		})
	})
}

// viewedClientValidatorCollisionDSL exposes one viewed type through unary,
// server-sent event, and WebSocket responses.
func viewedClientValidatorCollisionDSL() {
	result := dsl.ResultType("application/vnd.viewed-client-collision", func() {
		dsl.TypeName("ViewedClientCollision")
		dsl.Attribute("id", dsl.String)
		dsl.Required("id")
		dsl.View("summary", func() { dsl.Attribute("id") })
		dsl.View("detailed", func() { dsl.Attribute("id") })
	})
	dsl.Service("Viewed Validator", func() {
		dsl.Method("Read", func() {
			dsl.Result(result)
			dsl.HTTP(func() { dsl.GET("/read") })
		})
		dsl.Method("Watch", func() {
			dsl.StreamingResult(result)
			dsl.HTTP(func() {
				dsl.GET("/watch")
				dsl.ServerSentEvents()
			})
		})
		dsl.Method("Socket", func() {
			dsl.StreamingResult(result)
			dsl.HTTP(func() { dsl.GET("/socket") })
		})
	})
}

// viewedSSESoleViewDSL defines only Goa's default view.
func viewedSSESoleViewDSL() {
	event := dsl.ResultType("application/vnd.viewed-sse-sole-view", func() {
		dsl.TypeName("ViewedSSESoleView")
		dsl.Attribute("id", dsl.String)
		dsl.Required("id")
		dsl.View(expr.DefaultView, func() { dsl.Attribute("id") })
	})
	dsl.Service("Viewed SSE Sole View", func() {
		dsl.Method("Watch", func() {
			dsl.StreamingResult(event)
			dsl.HTTP(func() {
				dsl.GET("/watch")
				dsl.ServerSentEvents()
			})
		})
	})
}

// linkedHTTPPlan runs the same planning steps as the generator and returns the
// HTTP plan after all service and package names are available.
func linkedHTTPPlan(t *testing.T, design func()) *Plan {
	t.Helper()
	root := expr.RunDSL(t, design)
	return linkedHTTPPlanForRoot(t, root)
}

// viewedSSEConstructorCollisionDSL defines two view names that Goify maps to
// the same preferred constructor spelling.
func viewedSSEConstructorCollisionDSL() {
	event := dsl.ResultType("application/vnd.viewed-sse-collision", func() {
		dsl.TypeName("ViewedSSECollisionEvent")
		dsl.Attribute("id", dsl.String)
		dsl.View("foo-bar", func() { dsl.Attribute("id") })
		dsl.View("foo bar", func() { dsl.Attribute("id") })
	})
	dsl.Service("Viewed SSE Collision", func() {
		dsl.Method("Watch", func() {
			dsl.StreamingResult(event)
			dsl.HTTP(func() {
				dsl.GET("/watch")
				dsl.ServerSentEvents()
			})
		})
	})
}
