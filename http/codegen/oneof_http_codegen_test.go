// These tests regress HTTP code generation around OneOf request bodies and
// single-view ResultType responses. Client validation, union collection, and
// decode/init must all derive from the same effective transport body shape.
package codegen

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	. "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestClientCLIInlinesOneOfRequestValidation(t *testing.T) {
	code := renderClientCLISectionCode(t, testdata.PayloadBodyUnionUserValidateDSL, 1, 1)

	require.Contains(t, code, "BuildMethodBodyUnionUserValidatePayload")
	require.Contains(t, code, "if body.A == nil")
	require.Contains(t, code, "marshalUnionUserValidateTo")
	require.NotContains(t, code, "ValidateMethodBodyUnionUserValidateRequestBody")
}

func TestClientResponseCodeProjectsSingleViewOneOfResults(t *testing.T) {
	cases := []struct {
		name string
		dsl  func()
	}{
		{
			name: "single-result",
			dsl:  oneOfResultSingleViewDSL,
		},
		{
			name: "collection-result",
			dsl:  oneOfResultCollectionSingleViewDSL,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			typeCode := renderClientTypesCode(t, c.dsl)
			decodeCode := renderClientDecodeCode(t, c.dsl)

			require.Contains(t, typeCode, "AnimalView")
			require.NotContains(t, typeCode, "CatDetailsResponseBody")
			require.NotContains(t, typeCode, "DogDetailsResponseBody")
			require.NotContains(t, typeCode, "BirdDetailsResponseBody")
			require.NotContains(t, typeCode, "FishDetailsResponseBody")
			require.NotContains(t, typeCode, "json:\"details,omitempty\"")

			require.NotContains(t, decodeCode, "body.Details")
			require.NotContains(t, decodeCode, "CatDetailsResponseBody")
			require.NotContains(t, decodeCode, "DogDetailsResponseBody")
			require.NotContains(t, decodeCode, "BirdDetailsResponseBody")
			require.NotContains(t, decodeCode, "FishDetailsResponseBody")
		})
	}
}

// renderClientCLISectionCode renders the requested client CLI section for a
// DSL and returns the generated code.
func renderClientCLISectionCode(t *testing.T, dsl func(), fileIndex, sectionIndex int) string {
	t.Helper()

	root := expr.RunDSL(t, dsl)
	services := CreateHTTPServices(root)
	fs := ClientCLIFiles(services)

	return codegen.SectionCode(t, fs[fileIndex].SectionTemplates[sectionIndex])
}

// renderClientTypesCode renders the client type file for a single-service DSL.
func renderClientTypesCode(t *testing.T, dsl func()) string {
	t.Helper()

	const genpkg = "gen"

	root := expr.RunDSL(t, dsl)
	services := CreateHTTPServices(root)
	fs := typesFile(root.API.HTTP.Services[0], false, services)

	var buf bytes.Buffer
	for _, s := range fs.SectionTemplates[1:] {
		require.NoError(t, s.Write(&buf))
	}

	return codegen.FormatTestCode(t, "package foo\n"+buf.String())
}

// renderClientDecodeCode renders the client decode section for a single-service
// DSL and returns the generated code.
func renderClientDecodeCode(t *testing.T, dsl func()) string {
	t.Helper()

	root := expr.RunDSL(t, dsl)
	services := CreateHTTPServices(root)
	fs := ClientFiles(services)
	require.Len(t, fs, 2)

	sections := fs[1].SectionTemplates
	require.Greater(t, len(sections), 2)

	return codegen.SectionCode(t, sections[2])
}

// oneOfResultSingleViewDSL defines a ResultType whose only view drops the OneOf
// field. Client response code must therefore treat the transport body as the
// projected view, not the raw ResultType.
func oneOfResultSingleViewDSL() {
	animal := oneOfAnimalResultType("application/vnd.oneof-http-single-view.animal")

	Service("ServiceOneOfSingleView", func() {
		Method("MethodShowAnimal", func() {
			Payload(func() {
				Attribute("id", String)
				Required("id")
			})
			Result(animal)
			HTTP(func() {
				GET("/animals/{id}")
				Response(StatusOK)
			})
		})
	})
}

// oneOfResultCollectionSingleViewDSL defines a collection result whose
// transport body is fixed to a single view. The generated client collection
// code must project before collecting unions and body types.
func oneOfResultCollectionSingleViewDSL() {
	animal := oneOfAnimalResultType("application/vnd.oneof-http-collection-view.animal")

	Service("ServiceOneOfCollectionSingleView", func() {
		Method("MethodListAnimals", func() {
			Result(CollectionOf(animal), func() {
				View("default")
			})
			HTTP(func() {
				GET("/animals")
				Response(StatusOK)
			})
		})
	})
}

// oneOfAnimalResultType returns a ResultType whose default view excludes the
// OneOf details attribute. The omitted union is the regression target.
func oneOfAnimalResultType(mediaType string) *expr.ResultTypeExpr {
	var catDetails = Type("CatDetails", func() {
		Attribute("favorite_spot", String)
		Attribute("lives_left", Int)
		Required("favorite_spot", "lives_left")
	})
	var dogDetails = Type("DogDetails", func() {
		Attribute("favorite_park", String)
		Attribute("plays_fetch", Boolean)
		Required("favorite_park", "plays_fetch")
	})
	var birdDetails = Type("BirdDetails", func() {
		Attribute("can_fly", Boolean)
		Attribute("vocabulary_size", Int)
		Required("can_fly", "vocabulary_size")
	})
	var fishDetails = Type("FishDetails", func() {
		Attribute("water_type", String)
		Required("water_type")
	})

	return ResultType(mediaType, func() {
		TypeName("Animal")
		Attributes(func() {
			Attribute("name", String)
			OneOf("details", func() {
				Attribute("cat", catDetails)
				Attribute("dog", dogDetails)
				Attribute("bird", birdDetails)
				Attribute("fish", fishDetails)
			})
			Attribute("id", String)
		})
		View("default", func() {
			Attribute("name")
			Attribute("id")
		})
		Required("name", "id")
	})
}
