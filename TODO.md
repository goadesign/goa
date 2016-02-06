# Things that are "on the list"

## Versioning

* gopkg.in "integration"

## Generation Targets

* Angular target that generates angular services for each resource.
* [DONE] Client target that generates an API client package and command line tool.
* [DONE] Docs target that generates swagger and / or praxis JSON docs.
* Test target that generates API integration tests using gomega's ghttp package.
* [DONE] Generic target that takes the path to a Go package and the name of the "Generate" method
  and calls it passing in the metadata.

## Integrations

* NewRelic
* Errbit
* Plugin system a la Praxis

## Enhancements

* [DONE] Make sure "type overload" works. I.e. Param("foo", Type, func() { Attribute(...) })
* Documentation: [DONE] DSL reference, middleware support more examples etc
* [DONE] Add examples to DSL including auto-generated examples
* Praxis JSON to goa metadata generator
* Implement response inline media type (with resource media type inheritance)

## Praxis Mismatches

### Notes

* Load then validate
* Render produces native types so that it works with JSON, YAML etc.
* Reference should walk each attribute and look for a parent definition (consider embedded structs)

### TODO

* Configuration
* [DONE] Handle the case where an action handler did not write a response
* Before / After filter? (is middleware enough?)
* Handle attribute default value
* [DONE] Examples (same behavior as Load / Dump)
* [DONE] Default view is required
* [WILLNOTDO] Rendering caching
* [DONE] Versioning
* [DONE] Encoding handlers (produces, consumes)
* [DONE] Rename "MediaType" to "DefaultMediaType" in Resource DSL
* [WILLNOTDO] Remove support for multiple routes?
* Default base path for resources built after resource name
* [DONE] // for absolute routes
* Generate action route builder helpers (other than canonical href)
* Equivalent to parse_href from praxis ResourceDefinition ?
* Only use default medai type if response template takes media type as arg (instead of hardcoded to 200)
* Parameterize traits
* Add swagger-like CollectionFormat
* Add swagger-like support for security definitions
* Add swagger-like support for deprecated, [DONE] schemes
