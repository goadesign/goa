# Constructor Union Failure Inventory

This checklist is the merge bar for constructor-form `OneOf(...)` support.

Do not call the feature done until each item is:
- checked off with a direct test or golden,
- checked off after inspection with a concrete reason recorded in this file,
- or moved into a deferred limitations section.

## Status

- Branch: `feat/oneof-payload-result-unions`
- Audit mode: active
- Rule: do not claim the feature is done while any unchecked item below remains.
- Rule: update this file as part of every fix batch.

## DSL Evaluation

- [x] Declaration-signature `OneOf(...)` outside declaration contexts fails with a precise DSL-context error in covered cases.
- [x] Declaration-form vs constructor-form `OneOf(...)` ambiguity is documented as a deferred limitation where the DSL evaluation context cannot distinguish declaration-form use from type-expression use.
- [x] Malformed constructor calls must fail closed and must not degrade into another type.
- [x] Constructor-form string variants support already-defined named types.
- [x] Constructor-form string variants support true forward declarations.
- [x] Constructor-form string variants support recursion.
- [x] Constructor-form string variants report precise unresolved-name errors.
- [x] Constructor-form discriminator key overrides propagate correctly from enclosing payload/result attributes.
- [x] Removed hacky signature checks in `dsl/attribute.go` in favor of cleaner type assertions.

## Variant Naming and Wire Contract

- [x] Discriminator values must be deterministic in covered constructor-form union cases.
- [x] Variant naming is collision-safe for named variants.
- [x] Anonymous constructor-union type names remain stable when branch order is reversed in covered cases.
- [x] Ambiguous unnamed complex variants must not create brittle wire contracts in covered constructor-form cases.
- [x] Reordering unrelated declarations does not change public discriminator values or rendered OpenAPI output in the covered end-to-end fixture.
- [x] Constructor-form discriminator values remain stable under covered `TypeName`, aliasing, and generated rename paths.
- [x] Distinct discriminator names remain distinct after Go/codegen name normalization in covered constructor-form duplicate-name cases.
- [x] Shared `UniqueUnionFieldNames` logic moved to `codegen/types.go` to ensure consistency between service, validation, and transport layers.

## Service and Transform Codegen

- [x] Generated union types, constructors, and transforms are nil-safe in covered paths.
- [x] Invalid DSL states do not cascade into misleading secondary codegen behavior in covered paths.
- [x] Generated code remains symmetric with existing named/declaration-form union behavior in covered service/transform cases.
- [x] Collections of constructor-form unions (`ArrayOf(OneOf(...))`, `MapOf(..., OneOf(...))`) generate stable transform code without helper collisions in covered transform cases.
- [x] Constructor-form result unions with method-level views fail with a precise validation error instead of attempting inconsistent projection.
- [x] Repeated identical anonymous constructor unions used across methods in the same service reuse shared generated union types without duplicate top-level identifiers.
- [x] Generated Go union kind constants remain collision-safe after Go identifier normalization, not just generated field names.
- [x] Union field-name suffixing preserves already-reserved normalized names so cases like `Foo`, `Foo!`, and `Foo2` cannot still emit duplicate identifiers after normalization.
- [x] Union transforms (`transformUnion`) are now name-aware instead of index-based, preventing semantic swaps when branch order differs.

## Validation and Defaults

- [x] Validation code generation validates the active constructor-union branch in covered constructor-union smoke tests.
- [x] Validation code generation correctly validates the active constructor-union branch, including nested branch validations, in covered constructor-union validation paths.
- [x] Invalid constructor unions in validation paths fail precisely at the DSL boundary instead of skipping validation or generating uncompilable code.
- [x] Default values for constructor unions fail with a precise DSL error instead of leaking into codegen.
- [x] Default-value rejection is intentionally correct for all union forms, not an accidental broadening of behavior from constructor unions to declaration-form unions.
- [x] Constructor-union default rejection does not regress long-standing declaration-form union behavior unless that contract change is explicit and intentionally accepted.
- [x] Accessor de-duplication is correctly propagated into validation code via `UniqueUnionFieldNames`.

## HTTP Transport Codegen

- [x] Top-level and nested unions encode and decode symmetrically in covered HTTP transport paths.
- [x] Optional `null` bodies do not skip required non-body decoding or validation.
- [x] Required payload semantics remain intact after nil/null handling changes in covered HTTP decoder paths.
- [x] Client and server generation remain behaviorally aligned in covered top-level and nested HTTP transport paths.
- [x] Nested constructor unions with custom discriminator/value keys are client/server symmetric in covered HTTP transport paths.
- [x] Constructor unions are rejected or handled intentionally when placed in HTTP params, headers, or cookies.
- [x] Multipart request generation rejects constructor unions with a precise endpoint-validation error.
- [x] WebSocket streaming generation handles constructor-form unions for `StreamingPayload` / `StreamingResult` in covered server/client WebSocket smoke tests.
- [x] Security-scheme extraction through constructor unions fails with a precise DSL validation error in covered method-validation paths.
- [x] `isPointerTypeRef` string-hack in server templates replaced with robust `expr.DataType` inspection via `isNilable`.

## OpenAPI v3 Schemas

- [x] Wrapper/discriminator generation is internally consistent in covered cases.
- [x] Discriminator mapping targets match the emitted wrapper schemas in covered cases.
- [x] Recursive unions do not recurse forever during schema generation.
- [x] Wrapper component names are collision-safe in covered cases.
- [x] Reversing constructor-union branch order does not change rendered OpenAPI output in the covered end-to-end fixture.
- [x] Schema refs and names remain stable under covered service traversal-order changes.

## OpenAPI v2 Compatibility

- [x] Constructor-form unions do not panic or crash OpenAPI v2 generation in covered smoke tests.
- [x] Constructor-form unions do not panic or crash OpenAPI v2 generation in covered top-level, custom-key, nested, and recursive cases.
- [x] OpenAPI v2 degrades gracefully for constructor-form unions with a covered object-schema fallback.
- [x] Cookie-backed API key security schemes fallback to `in: header` for Swagger 2.0 documents to ensure spec compliance.
- [x] Swagger 2.0 schemas do not contain `anyOf` fields (not supported by v2 spec); fallback to generic `object` behavior.

## OpenAPI Examples

- [x] Schema examples, payload examples, and media-type examples use the canonical wire shape in covered cases.
- [x] User-provided examples select the intended union branch, including later object branches, in covered cases.
- [x] Ambiguous raw user examples for overlapping object branches do not silently canonicalize to the wrong branch in covered schema/payload example paths.
- [x] Ambiguous user-provided union examples fail explicitly by omission instead of silently falling back to the first branch in covered OpenAPI example paths.
- [x] Ambiguous user-provided union examples are an intentional upstream behavior with documented rationale, not just an implementation detail that drops examples silently. Inspected `http/codegen/openapi/v3/example.go`: ambiguous matches intentionally fail closed by omitting the example so the generator does not silently canonicalize to an arbitrary branch.
- [x] Generated examples remain canonical through nested unions in covered cases.
- [x] Schema-level property examples agree with enclosing payload and request/response examples in covered cases.
- [x] Custom discriminator/value keys are reflected in examples as well as schemas in covered explicit and generated example paths.
- [x] Schema-level example selection remains stable and documented in covered multiple-user-example cases.
- [x] Nested objects that mix user-provided examples and generated examples preserve the correct canonical union shape in covered nested object paths.
- [x] Reordering constructor union branches does not silently change generated examples in covered generated-example paths.

## Regression Surface

- [x] No unrelated regressions were detected in the covered packages after `make test` and `make lint`.
- [x] Existing repo rules in `AGENTS.md` are currently satisfied in the touched code.
- [x] `make test` passes after the current change set.
- [x] `make lint` passes after the current change set.

## Other Goa Surfaces

- [x] Unary gRPC code generation supports top-level constructor-form union payloads/results in covered smoke tests without invalid `.proto` output or generator panics.
- [x] HTTP and gRPC CLI generation support top-level constructor-form union payloads in covered smoke tests without generation failures.
- [x] gRPC code generation supports constructor-form unions in covered unary and bidirectional streaming smoke tests without invalid `.proto` output or generator panics.
- [x] gRPC validation rejects explicit `rpc:tag` collisions between constructor-union branches and sibling protobuf fields before `.proto` generation.
- [x] gRPC validation rejects explicit duplicate `rpc:tag` values across constructor-union branches before `.proto` generation.
- [x] gRPC/protobuf generation rejects or dedupes constructor-union branch names that collide after protobuf field-name normalization.
- [x] CLI generation handles constructor-form union payloads in covered HTTP and gRPC payload-builder paths without generation failures or unusable clients.
- [x] `Error(...)` declarations with constructor-form unions fail with a precise, intentional DSL validation error in covered cases.
- [x] gRPC error conversion no longer needs to handle constructor-form union error types because union-typed errors are rejected before transport/codegen.
- [x] gRPC `.proto` generation pre-seeds used tags and names from all sibling fields, including those inside other unions, preventing collisions between union branches and regular fields.

## Open Items Under Active Audit

- Symmetry between constructor-form unions and named/declaration-form unions beyond the covered service/transform cases.
- Remaining edge cases for custom discriminator/value keys beyond the covered HTTP/OpenAPI example paths.
- Post-normalization identifier collisions in generated Go/OpenAPI helper names beyond the covered discriminator-name cases.
- Multipart and broader streaming/security-analysis behavior for constructor unions beyond the covered WebSocket and method-validation paths.
- Broader gRPC coverage beyond the covered unary and bidirectional-streaming smoke paths.
- Anonymous-union deduping and helper/type collisions across repeated use sites beyond the covered same-service repeated-use and collection-transform paths.
- Whether silent omission is the right upstream contract for ambiguous union examples, or whether the generator should preserve the raw example or raise a design-time error instead.
- Whether rejecting defaults on all unions is an intentional upstream contract change or constructor-union-specific behavior that widened unintentionally.
- Duplication of constructor-union naming/stability logic across DSL, expr finalization, and transport codegen, which increases drift risk.

## Priority Tests To Add

- [x] A union with two object branches that both match `{}` or another sparse example, asserting OpenAPI generation omits the ambiguous example instead of emitting the first branch.
- [x] A gRPC fixture where a constructor-union branch uses the same explicit `rpc:tag` as a sibling message field, asserting validation fails before `.proto` generation.
- [x] A gRPC fixture where two constructor-union branches use the same explicit `rpc:tag`, asserting validation fails before `.proto` generation.
- [x] A constructor union whose branch names stay distinct as discriminator values but collide after Go identifier normalization, asserting generated kind constants and helpers remain collision-free.
- [x] A constructor union whose normalized field names are `Foo`, `Foo!`, and `Foo2`, asserting service and HTTP union helpers reserve pre-existing normalized names before suffixing and never emit duplicate identifiers.
- [x] A constructor union whose branch names collide after protobuf field-name normalization, asserting `.proto` generation fails precisely or emits collision-safe names.
- [x] A nested payload where:
  - outer is generated,
  - `outer.choice` has a user example,
  - `outer.choice.value.inner` is another generated union,
  - and one level uses custom `kind` / `data` keys.
  Assert schema example, payload example, and media-type examples all agree.
- [x] A constructor union whose branch names differ only by formatting/casing, to catch collisions after identifier normalization in the constructor discriminator names.
- [x] A golden pair with the same constructor union branches in opposite order, asserting runtime discriminator values stay stable where promised and emitted examples stay stable end-to-end.
- [x] A negative test for ambiguous object-branch matching with custom keys, not just default `type` / `value`.
- [x] A DSL/codegen test asserting constructor unions in HTTP params, headers, and cookies fail precisely rather than panicking or generating broken code.
- [x] Validation generation coverage for constructor unions with branch-specific required fields and validations, asserting only the active branch is validated.
- [x] Default-value generation coverage for constructor unions, asserting generated Go is valid or the error is explicit.
- [x] A declaration-form union attribute with a default value, asserting constructor-union default rejection does not widen into a backwards-incompatible declaration-form regression unless explicitly intended.
- [x] Multipart and broader streaming coverage for constructor-form unions, asserting intentional support or precise rejection.
- [x] Security-analysis coverage where an auth token is reachable only inside a constructor-union branch, asserting a precise failure mode.
- [x] Repeated identical anonymous constructor unions across multiple methods, asserting deduping/naming behavior is intentional and collision-free.
- [x] `Error(...)` coverage where the error type is a constructor-form union, asserting HTTP and gRPC either generate the required transforms and mappings or fail with a precise DSL/codegen error.

## Deferred Limitations

- Declaration-form `OneOf("X", func(){...})` used as a type expression inside another declaration context cannot be distinguished reliably from a real declaration-form `OneOf(...)` call at DSL-evaluation time. Covered behavior: the DSL fails instead of succeeding silently, but the error is still the generic type-path failure for `Attribute("choice", OneOf("Inner", func(){...}))` rather than a precise `invalid use of OneOf`.
