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
- [ ] Declaration-form vs constructor-form `OneOf(...)` must not misparse based on context.
- [x] Malformed constructor calls must fail closed and must not degrade into another type.
- [x] Constructor-form string variants support already-defined named types.
- [x] Constructor-form string variants support true forward declarations.
- [x] Constructor-form string variants support recursion.
- [x] Constructor-form string variants report precise unresolved-name errors.
- [x] Constructor-form discriminator key overrides propagate correctly from enclosing payload/result attributes.

## Variant Naming and Wire Contract

- [ ] Discriminator values must be deterministic.
- [x] Variant naming is collision-safe for named variants.
- [ ] Ambiguous unnamed complex variants must not create brittle wire contracts.
- [ ] Reordering unrelated declarations must not silently change public discriminator values or OpenAPI component names.
- [ ] Constructor-form discriminator values remain stable under `TypeName`, aliasing, and renamed/generated user types.
- [ ] Distinct discriminator names remain distinct after Go/codegen name normalization.

## Service and Transform Codegen

- [x] Generated union types, constructors, and transforms are nil-safe in covered paths.
- [x] Invalid DSL states do not cascade into misleading secondary codegen behavior in covered paths.
- [ ] Generated code remains symmetric with existing named/declaration-form union behavior where intended.
- [ ] Collections of constructor-form unions (`ArrayOf(OneOf(...))`, `MapOf(..., OneOf(...))`) generate stable transform code without helper collisions or semantic drift.
- [ ] Result-view projection remains correct when constructor-form unions appear in results and branch views differ or are missing.
- [ ] Anonymous constructor unions used repeatedly across methods/packages dedupe intentionally without type/helper collisions or package clutter.

## Validation and Defaults

- [x] Validation code generation validates the active constructor-union branch in covered constructor-union smoke tests.
- [ ] Validation code generation correctly validates the active constructor-union branch, including nested branch validations.
- [ ] Invalid constructor unions in validation paths fail precisely instead of skipping validation or generating uncompilable code.
- [ ] Default values for constructor unions either render correctly in generated Go code or fail with a precise DSL/codegen error.

## HTTP Transport Codegen

- [ ] Top-level and nested unions encode and decode symmetrically.
- [x] Optional `null` bodies do not skip required non-body decoding or validation.
- [ ] Required payload semantics remain intact after nil/null handling changes.
- [ ] Client and server generation remain behaviorally aligned.
- [ ] Nested constructor unions with custom discriminator/value keys are client/server symmetric, not just OpenAPI-correct.
- [x] Constructor unions are rejected or handled intentionally when placed in HTTP params, headers, or cookies.
- [ ] Multipart request generation either supports constructor unions correctly or rejects them with a precise error.
- [ ] WebSocket streaming generation handles constructor-form unions for `StreamingPayload` / `StreamingResult` correctly.
- [ ] Security-scheme extraction through constructor unions either works intentionally or fails with a precise DSL error.

## OpenAPI v3 Schemas

- [x] Wrapper/discriminator generation is internally consistent in covered cases.
- [x] Discriminator mapping targets match the emitted wrapper schemas in covered cases.
- [x] Recursive unions do not recurse forever during schema generation.
- [x] Wrapper component names are collision-safe in covered cases.
- [ ] Schema refs and names remain stable under traversal-order changes.

## OpenAPI v2 Compatibility

- [x] Constructor-form unions do not panic or crash OpenAPI v2 generation in covered smoke tests.
- [ ] Constructor-form unions do not panic or crash OpenAPI v2 generation.
- [ ] OpenAPI v2 degrades gracefully for constructor-form unions with a stable, intentional fallback shape.

## OpenAPI Examples

- [x] Schema examples, payload examples, and media-type examples use the canonical wire shape in covered cases.
- [x] User-provided examples select the intended union branch, including later object branches, in covered cases.
- [ ] Ambiguous raw user examples for overlapping object branches do not silently canonicalize to the wrong branch.
- [ ] Ambiguous user-provided union examples fail explicitly instead of silently falling back to the first branch.
- [x] Generated examples remain canonical through nested unions in covered cases.
- [x] Schema-level property examples agree with enclosing payload and request/response examples in covered cases.
- [ ] Custom discriminator/value keys are reflected in examples as well as schemas.
- [ ] Schema-level example selection remains stable and documented when multiple user examples exist.
- [ ] Nested objects that mix user-provided examples and generated examples preserve the correct canonical union shape at every level.
- [ ] Reordering constructor union branches does not silently change generated examples unless that behavior is accepted as a limitation.

## Regression Surface

- [ ] The change does not alter unrelated code paths.
- [x] Existing repo rules in `AGENTS.md` are currently satisfied in the touched code.
- [x] `make test` passes after the current change set.
- [x] `make lint` passes after the current change set.

## Other Goa Surfaces

- [x] Unary gRPC code generation supports top-level constructor-form union payloads/results in covered smoke tests without invalid `.proto` output or generator panics.
- [x] HTTP and gRPC CLI generation support top-level constructor-form union payloads in covered smoke tests without generation failures.
- [ ] gRPC code generation supports constructor-form unions, including top-level method payloads/results, without invalid `.proto` output or generator panics.
- [ ] CLI generation handles constructor-form unions, including top-level payloads/results, without generation failures or unusable clients.
- [x] `Error(...)` declarations with constructor-form unions fail with a precise, intentional DSL validation error in covered cases.
- [x] gRPC error conversion no longer needs to handle constructor-form union error types because union-typed errors are rejected before transport/codegen.

## Open Items Under Active Audit

- Traversal-order stability for schema refs and wrapper component naming.
- Symmetry between constructor-form unions and named/declaration-form unions.
- Symmetry between HTTP client and server handling for nested and top-level unions.
- Remaining edge cases for custom discriminator/value keys.
- Remaining context ambiguity when declaration-form `OneOf(...)` is supplied as an `Attribute` type argument inside declaration contexts.
- gRPC, OpenAPI v2, CLI, collections, views, and error-declaration behavior for constructor-form unions.
- Broader validation coverage beyond the covered active-branch smoke path.
- Ambiguous overlapping object-branch examples and multi-example schema/example divergence.
- Discriminator drift under `TypeName`, aliasing, and generated renames.
- Ambiguous branch inference for user examples.
- Mixed user-example/generated-example propagation through nested objects and unions.
- Post-normalization identifier collisions in generated Go/OpenAPI helper names.
- Example churn caused by branch-order-dependent generated examples.
- Invalid HTTP placements for constructor unions.
- Validation/default rendering gaps for constructor unions.
- Multipart, WebSocket streaming, and security-analysis behavior for constructor unions.
- Broader gRPC coverage beyond the covered unary top-level payload/result smoke path.
- Anonymous-union deduping and helper/type collisions across repeated use sites.

## Priority Tests To Add

- [ ] A union with two object branches that both match `{}` or another sparse example, asserting OpenAPI generation reports ambiguity instead of emitting the first branch.
- [ ] A nested payload where:
  - outer is generated,
  - `outer.choice` has a user example,
  - `outer.choice.value.inner` is another generated union,
  - and one level uses custom `kind` / `data` keys.
  Assert schema example, payload example, and media-type examples all agree.
- [ ] A constructor union whose branch names differ only by formatting/casing, to catch collisions after identifier normalization in generated service and transport code.
- [ ] A golden pair with the same constructor union branches in opposite order, asserting runtime discriminator values stay stable where promised, and emitted examples either stay stable or the limitation is documented explicitly.
- [ ] A negative test for ambiguous object-branch matching with custom keys, not just default `type` / `value`.
- [ ] A DSL/codegen test asserting constructor unions in HTTP params, headers, and cookies fail precisely rather than panicking or generating broken code.
- [ ] Validation generation coverage for constructor unions with branch-specific required fields and validations, asserting only the active branch is validated.
- [ ] Default-value generation coverage for constructor unions, asserting generated Go is valid or the error is explicit.
- [ ] Multipart and WebSocket streaming coverage for constructor-form unions, asserting intentional support or precise rejection.
- [ ] Security-analysis coverage where an auth token is reachable only inside a constructor-union branch, asserting a precise failure mode.
- [ ] Repeated identical anonymous constructor unions across multiple methods, asserting deduping/naming behavior is intentional and collision-free.
- [x] `Error(...)` coverage where the error type is a constructor-form union, asserting HTTP and gRPC either generate the required transforms and mappings or fail with a precise DSL/codegen error.

## Deferred Limitations

- None currently recorded. If a limitation is accepted instead of fixed, move it here with a rationale.
