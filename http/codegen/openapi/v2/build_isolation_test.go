// This file checks that Swagger builds do not share or change each other's schemas.
package openapiv2_test

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
	openapiv2 "goa.design/goa/v3/http/codegen/openapi/v2"
)

func TestBuildsKeepDefinitionsSeparate(t *testing.T) {
	firstRoot := expr.RunDSL(t, schemaBuildDSL("first"))
	secondRoot := expr.RunDSL(t, schemaBuildDSL("second"))

	first, err := openapiv2.NewV2(
		firstRoot,
		firstRoot.API.Servers[0].Hosts[0],
	)
	require.NoError(t, err)
	firstJSON, err := json.Marshal(first)
	require.NoError(t, err)

	second, err := openapiv2.NewV2(
		secondRoot,
		secondRoot.API.Servers[0].Hosts[0],
	)
	require.NoError(t, err)

	firstDefinition := definitionWithProperty(first.Definitions, "first")
	require.NotNil(t, firstDefinition, "definitions: %#v", first.Definitions)
	require.Contains(t, firstDefinition.Properties, "first")
	require.NotContains(t, firstDefinition.Properties, "second")
	secondDefinition := definitionWithProperty(second.Definitions, "second")
	require.NotNil(t, secondDefinition, "definitions: %#v", second.Definitions)
	require.Contains(t, secondDefinition.Properties, "second")
	require.NotContains(t, secondDefinition.Properties, "first")

	firstJSONAfterSecondBuild, err := json.Marshal(first)
	require.NoError(t, err)
	require.Equal(t, firstJSON, firstJSONAfterSecondBuild)
}

func TestBuildsAreSafeToRunTogether(t *testing.T) {
	firstRoot := expr.RunDSL(t, schemaBuildDSL("first"))
	secondRoot := expr.RunDSL(t, schemaBuildDSL("second"))

	type result struct {
		spec *openapiv2.V2
		err  error
	}
	start := make(chan struct{})
	results := make(chan result, 2)
	var ready sync.WaitGroup
	ready.Add(2)
	build := func(root *expr.RootExpr) {
		ready.Done()
		<-start
		spec, err := openapiv2.NewV2(
			root,
			root.API.Servers[0].Hosts[0],
		)
		results <- result{spec: spec, err: err}
	}
	go build(firstRoot)
	go build(secondRoot)
	ready.Wait()
	close(start)

	properties := make(map[string]int)
	for range 2 {
		built := <-results
		require.NoError(t, built.err)
		var found []string
		for _, definition := range built.spec.Definitions {
			for property := range definition.Properties {
				if property == "first" || property == "second" {
					found = append(found, property)
				}
			}
		}
		require.Len(t, found, 1)
		properties[found[0]]++
	}
	require.Equal(t, map[string]int{"first": 1, "second": 1}, properties)
}

// definitionWithProperty finds the returned schema that contains property.
func definitionWithProperty(definitions map[string]*openapi.Schema, property string) *openapi.Schema {
	for _, definition := range definitions {
		if _, ok := definition.Properties[property]; ok {
			return definition
		}
	}
	return nil
}

// schemaBuildDSL returns an API whose Shared result contains only field.
func schemaBuildDSL(field string) func() {
	return func() {
		shared := dsl.Type("Shared", func() {
			dsl.Attribute(field, dsl.String)
		})
		dsl.API("test", func() {
			dsl.Server("test", func() {
				dsl.Host("localhost", func() {
					dsl.URI("https://goa.design")
				})
			})
		})
		dsl.Service("testService", func() {
			dsl.Method("show", func() {
				dsl.Result(shared)
				dsl.HTTP(func() {
					dsl.GET("/")
				})
			})
		})
	}
}
