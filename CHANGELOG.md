### Goa v3.22.1

#### Highlights
- JSON‑RPC 2.0 transport joins HTTP and gRPC as a first‑class option. Generate
  servers, clients, CLI, streaming (WebSocket, SSE), batch requests, and
  notifications from the same design. This makes Goa a great fit for RPC‑style
  APIs across protocols without duplicating effort ([PR #3734](https://github.com/goadesign/goa/pull/3734)).
- Toolchain updated to Go 1.24 (toolchain `go1.24.4`) with refreshed
  dependencies for better performance, diagnostics, and editor support
  ([PR #3765](https://github.com/goadesign/goa/pull/3765)).

#### Compatibility
- Requires Go 1.24+. Please update your Go toolchain before upgrading
  ([PR #3765](https://github.com/goadesign/goa/pull/3765)).
- Generated example CLI: `UsageCommands()` now returns `[]string` (previously a
  single string). Help text and behavior are unchanged; only the return type
  matters if you consume it programmatically
  ([PR #3734](https://github.com/goadesign/goa/pull/3734)).

#### New Features
- JSON‑RPC 2.0 (now first‑class alongside HTTP and gRPC)
  - Full code generation for servers and clients; batch requests and
    notifications; streaming via WebSocket and SSE; CLI integration
    (`-jsonrpc`/`-j` when both HTTP and JSON‑RPC are available).
  - New `jsonrpc` package: request/response types, error codes, helpers
    (`MakeSuccessResponse`, `MakeErrorResponse`, `MakeNotification`,
    `IDToString`) and a `StreamConfig` with `With*` options to tune WebSocket
    behavior (timeouts, buffers, retries, compression, ping).
  - Comprehensive docs and integration tests to help you hit the ground
    running ([PR #3734](https://github.com/goadesign/goa/pull/3734)).

#### Improvements
- Conversion functions are now generated per target package. If your user types
  include `Meta("struct:pkg:path", "...")`, their conversions land in the
  correct `convert.go` next to those types. This fixes cross‑package builds and
  smooths Windows paths ([PR #3748](https://github.com/goadesign/goa/pull/3748),
  fixes [Issue #3745](https://github.com/goadesign/goa/issues/3745)).
- Validation
  - Fix `OneOf` validation in view types when a union arm uses a format (e.g.
    `FormatDateTime`). Prevents incorrect pointer assumptions and the resulting
    compile errors ([PR #3749](https://github.com/goadesign/goa/pull/3749)).
- HTTP decoding
  - Harden query parsing: malformed nested brackets no longer panic and now
    surface a clear error: “invalid query string: malformed brackets”
    ([PR #3746](https://github.com/goadesign/goa/pull/3746)).
- Codegen robustness and clarity
  - Default array handling: only cast defaults when the element type is an
    alias; supports `[]any`, `[]string`, numeric/bool slices, and
    `expr.ArrayVal`; avoids reflection for simpler, faster code
    ([PR #3760](https://github.com/goadesign/goa/pull/3760)).
  - Correct package qualification for inline structs referencing user types.
    This stabilizes builds in mixed package layouts and the JSON‑RPC IT
    harness ([PR #3762](https://github.com/goadesign/goa/pull/3762)).
  - `SnakeCase` also replaces “/” with “_” to produce safer identifiers
    ([PR #3760](https://github.com/goadesign/goa/pull/3760)).
- Testing utilities
  - Golden diffs now use `testify` for cleaner output. Use `-update` (or `-u`)
    to refresh goldens. Legacy diff flags are removed; `CompareOrUpdateGolden`
    remains for easy migration. Auto‑formatting of `.go` and `.json` goldens
    is built‑in ([PR #3766](https://github.com/goadesign/goa/pull/3766)).
- Modernization sweep
  - Broad internal cleanups (`maps.Copy`, `slices.Contains`, `http.NoBody`,
    fewer builtin shadowing issues, consistent file perms) improve code health
    across the board ([PR #3750](https://github.com/goadesign/goa/pull/3750)).

#### Documentation and Misc
- Expanded JSON‑RPC documentation with practical examples so you can adopt it
  confidently ([PR #3734](https://github.com/goadesign/goa/pull/3734)).
- Minor typo fixes and sponsor card update
  ([PR #3756](https://github.com/goadesign/goa/pull/3756),
  [PR #3752](https://github.com/goadesign/goa/pull/3752)).

#### Dependency Updates
- Go 1.24 toolchain; gRPC 1.74.2; x/tools 0.36.0; x/text up to 0.28.0; newer
  protobuf, x/net, x/sys, and related modules
  ([PR #3765](https://github.com/goadesign/goa/pull/3765),
  [PR #3751](https://github.com/goadesign/goa/pull/3751),
  [PR #3739](https://github.com/goadesign/goa/pull/3739),
  [PR #3740](https://github.com/goadesign/goa/pull/3740)).

#### Thanks
Huge thanks to everyone who contributed to this release:
- @raphael
- @rorydbain
- @RedMarcher
- @tchssk
- @chailandau
- @dependabot[bot]

Full changelog: https://github.com/goadesign/goa/compare/v3.21.5...v3.22.1

