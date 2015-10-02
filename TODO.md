# Things that are "on the list"

## Versioning

* gopkg.in "integration"

## Generation Targets

* Angular target that generates angular services for each resource.
* Client target that generates an API client package and command line tool.
* Docs target that generates swagger and / or praxis JSON docs.
* Test target that generates API integration tests using gomega's ghttp package.
* Generic target that takes the path to a Go package and the name of the "Generate" method
  and calls it passing in the metadata.

## Integrations

* NewRelic
* Errbit
* Plugin system a la Praxis

## Enhancements

* Documentation: DSL reference, middleware support more examples etc
* Add examples to DSL including auto-generated examples
* Praxis JSON to goa metadata generator
* Implement response inline media type (with resource media type inheritance)
