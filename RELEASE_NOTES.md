# Goa v3.23.0

We are thrilled to announce Goa v3.23.0! This release brings massive performance improvements to the code generation process, speeding it up by over 80%. It also includes important updates to the DSL, improved validation logic, and support for the latest JSON Schema draft in OpenAPI.

## Performance

* **Massive Code Generation Speedup:** The code generator has been optimized to be over 80% faster, making your development loop tighter than ever. ([#3832](https://github.com/goadesign/goa/pull/3832), [#3833](https://github.com/goadesign/goa/pull/3833))
* **Optimized File Writes:** Reduced I/O overhead during generation. ([#3852](https://github.com/goadesign/goa/pull/3852))

## New Features

* **Absolute Mount Paths:** The `Files` DSL now supports absolute paths using the `//` prefix, allowing you to mount static files at the server root regardless of API or service prefixes. ([#3837](https://github.com/goadesign/goa/pull/3837))
* **Method DSL Update:** The `Method` function now returns `MethodExpr`, enabling more flexible DSL composition. ([#3850](https://github.com/goadesign/goa/pull/3850))
* **OpenAPI Update:** The generated OpenAPI definitions now use JSON Schema Draft 2020-12. ([#3838](https://github.com/goadesign/goa/pull/3838))

## Bug Fixes

* **Validation Improvements:**
    * Fixed recursion guard propagation and alias type handling in validation code. ([#3836](https://github.com/goadesign/goa/pull/3836))
    * Added support for non-nullable elements in `ArrayOfRequired` validation. ([#3834](https://github.com/goadesign/goa/pull/3834))
    * Preserved underscores in validation function names to avoid naming conflicts. ([#3842](https://github.com/goadesign/goa/pull/3842))
    * Fixed validation helper naming for protobuf messages. ([#3844](https://github.com/goadesign/goa/pull/3844))
    * Added guards for nil elements in array transforms to prevent panics. ([#3830](https://github.com/goadesign/goa/pull/3830))
* **HTTP/JSON-RPC:** Fixed handler argument ordering for HTTP servers hosting JSON-RPC services. ([#3831](https://github.com/goadesign/goa/pull/3831))

## Dependency Updates

* Updated module dependencies to their latest versions. ([#3851](https://github.com/goadesign/goa/pull/3851))
* Bumped actions/checkout to v6. ([#3845](https://github.com/goadesign/goa/pull/3845))
* Bumped github/codeql-action to v4. ([#3818](https://github.com/goadesign/goa/pull/3818))
* Bumped actions/upload-artifact to v5. ([#3829](https://github.com/goadesign/goa/pull/3829))

## Contributors

A huge thank you to our contributors for this release:

* Isaac Seymour (@isaacseymour)
* Lawrence Jones (@lawrencejones)
* Raphael Simon (@raphael)
