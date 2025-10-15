# Repository Guidelines

## Project Structure & Module Organization
- `dsl/`: Public design language definitions (uses dot imports intentionally).
- `expr/`: Internal AST and validation for the DSL.
- `codegen/`: Generators for transports, types, and docs.
- `http/`, `grpc/`, `jsonrpc/`: Transport-specific runtime and codegen helpers.
- `middleware/`: Built-in interceptors and middleware.
- `pkg/`: Core runtime (endpoints, errors, version, helpers).
- `cmd/goa/`: Goa CLI source. Install locally to try changes.
- `docs/`, `security/`, `eval/`: Documentation assets, security notes, evaluation code.

## Build, Test, and Development Commands
- `make lint` — Run linters per `.golangci.yml` across `./...`.
- `make test` — Run unit and integration tests; writes coverage to `cover.out`.
- `cd cmd/goa && go install .` — Build/install the Goa CLI from source.

## Coding Style & Naming Conventions
- Go 1.24+. Format with `go fmt ./...`; keep imports grouped.
- Files: `lower_snake_case.go`; packages: lowercase, short, meaningful.
- Exported identifiers require GoDoc comments; avoid stutter.
- Errors: wrap with `%w`; prefer `errors.Is/As` (enforced by `errorlint`).

## Code File Organization
- Declarations order:
  1) Types (public first, then private) in a single `type (...)` block when practical
  2) Constants (public, then private)
  3) Variables (public, then private)
  4) Public functions
  5) Public methods
  6) Private functions
  7) Private methods
- No commented‑out code; remove dead code instead of commenting.
- Keep imports grouped; stdlib separated from external. Let `gofmt` manage ordering.

## Repro Protocol
To reproduce a code generation issue from a specific design, follow this protocol:
1. Create a new directory under ~/src/repros: ~/src/repros/<issue>. Choose a meaningful short name, for example ~/src/repros/customtype.
2. Create a design subdirectory and write the design file in ~/src/repros/<issue>/design/design.go.
3. Run `go mod init <issue>` in the issue directory where <issue> is the same short name, for example `go mod init customtype`.
4. Run `goa gen <issue>/design` in the issue directory, this will create a 'gen' directory.
5. Run `go mod tidy` to download all dependencies.
6. Run `go mod edit -replace goa.design/goa/v3=$HOME/src/goa`
7. Run `goa gen <issue>/design` in the issue directory a second time, this time it will use the development version of goa.
8. [OPTIONAL] if needed also generate the example command line tools with `goa example <issue>/design`
You are now ready to troubleshoot goa by making changes in ~/src/goa and
running the `goa gen` and/or `goa example` commands as per the above.

## Code Generation Behavior
- After modifying goa source code, you do not need to manually rebuild the goa
CLI. The `goa gen` and `goa example` commands automatically compile and use a
temporary binary that includes your latest changes.
- The `goa gen` command deletes and recreates the entire `gen` directory each
time it runs, removing any previously generated files. In contrast, the `goa
example` command generates example service code but does not overwrite existing
files in the `cmd` directory or any top-level service files; it only creates new
files if they do not already exist.

## General Principles
- Fail fast: do not mask invariant violations with defensive nil checks in hot paths; surface precise errors.
- Keep code tidy: no commented‑out code blocks; keep PRs focused.

## Additional Style Details
- Prefer `any` over `interface{}` in new code.
- Use multi‑line `if` blocks; target ~80‑column lines when practical.
- Struct/composite literals: break long/named fields onto one field per line with trailing commas; close brace on its own line.

## Testing Guidelines
- Write table‑driven tests in `*_test.go` using `testing` (optionally `testify`).
- Name tests `TestXxx`; keep unit tests fast and deterministic.
- Run `make test` locally and ensure coverage does not regress.

## Commit & Pull Request Guidelines
- Commit style mirrors history: `scope(subscope): concise summary`.
  Example: `http/codegen: fix SSE Last-Event-ID handling (#1234)`.
- Reference issues using `Fixes #NNNN` in PRs. Include rationale, tests, and docs updates when behavior/design changes.
- Keep PRs focused and small; ensure `make lint` and `make test` pass.
- If changing generators or DSL, run the CLI against examples to validate.

## Agent‑Specific Notes
- Avoid release automation (`make release*`). Keep patches minimal and scoped.
- Do not run `git` commands in this environment; maintainers handle tagging and releases.
