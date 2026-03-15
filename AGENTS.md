# Repository Guidelines

## Common Rules

### Agent Behavior

- **Plan before acting**: For ≤2 files, state a brief plan then implement. For ≥3 files, write a step-by-step plan first.
- **Read before editing**: Always read files before modifying. Search over guessing.
- **Fix root causes**: Do not produce local workarounds—fix the real issue.
- **Be concise**: Give short status updates during multi-step work. Present a short summary when done.

### Go Code Style

- **Go 1.24+**. Format with `go fmt ./...`.
- **Imports**: Group stdlib separate from external. Let gofmt manage ordering.
- **Files**: Use `lower_snake_case.go`. Keep ≤1000 lines; split proactively.
- **Naming**: Packages are lowercase and short. Exported identifiers need GoDoc. Avoid stutter.
- **Types**: Use `any` over `interface{}`. Prefer concrete types over `interface{}`.
- **Errors**: Wrap with `%w`. Use `errors.Is/As`. **Never ignore errors or use `_ = call()`**.
- **Signatures**: Keep on one line when ≤100 columns. Only wrap genuinely long signatures.
- **Slice/map nil**: Do not check nil before `len`. `len(nil)` returns 0. Use `len(x) == 0` directly.

### Code Blocks and Literals

- Always place a newline after `{` and before `}` for `if`, `for`, `switch`, `func`, `type`.
- No single-line blocks: `if cond { do() }` → use multiple lines.
- Short struct literals are fine inline: `&T{A: 1}`. Break long literals to one field per line with trailing commas.

### File Organization

Order declarations as:
1. Types (public, then private) in a single `type (...)` block when practical
2. Constants (public, then private)
3. Variables (public, then private)
4. Public functions
5. Public methods
6. Private functions
7. Private methods

No commented-out code—delete dead code.

### Error Handling & Contracts

- **Always check errors**. Never discard with `_`.
- **Strong contracts**: Goa validates payloads at boundaries. Do not re-validate inside service code.
- **No defensive programming**: Do not add nil/empty guards for values guaranteed by construction, Goa, or prior validation.
- **Validate only at boundaries**: HTTP/gRPC handlers, event consumers, DB results, third-party APIs, `ctx.Value()`, type assertions, required map lookups.
- **Fail fast**: Unexpected states are bugs. Return precise errors or panic—do not silently recover or skip.

### Goa DSL Rules

- **Never edit `gen/`**: Always regenerate.
- **DSL validation**: Put validations (lengths, enums, formats) in the design. Do not re-validate in code.
- **Avoid `Any`**: Use concrete types to enable gRPC generation.

### Codegen Implementation

- **Use NameScope helpers** for type references: `GoTypeRef`, `GoFullTypeRef`, `GoTypeName`. Never concatenate strings for types.
- Let Goa decide pointer/value semantics. Do not force `pointer=true` except in transport validation.
- **Keep helper visibility minimal**: If logic is shared only inside one codegen area, keep it package-private or move it under an `internal` package. Do not export helpers from a parent package just to share them across sibling generators.
- **Avoid pass-through wrappers**: When two helper functions differ only by forwarding arguments or hard-coding `nil`, collapse them into a single implementation instead of adding an extra layer.

### Documentation

- Every exported type, function, method, and field must have a GoDoc comment explaining its contract—like Go stdlib documentation.

### Safety & Forbidden Operations

| Action | Policy |
|--------|--------|
| `git clean/stash/reset/checkout` | **FORBIDDEN** |
| `go clean -cache` | **FORBIDDEN** during normal work |
| Edit `gen/` directly | **FORBIDDEN** |
| Changes ≥3 files | Describe plan first |
| New dependencies | Explain why first |

### Testing

- Write table-driven tests in `*_test.go`.
- Name tests `TestXxx`. Keep fast and deterministic.
- Use `testify/require` for assertions.
- Prefer `t.Errorf` over `t.Fatalf` so tests report multiple failures.

---

## Goa-Specific Rules

### Project Structure

- `dsl/`: Public DSL definitions (dot imports allowed per `.golangci.yml`)
- `expr/`: Internal AST and validation
- `codegen/`: Generators for transports, types, docs
- `http/`, `grpc/`, `jsonrpc/`: Transport-specific codegen
- `middleware/`: Built-in interceptors
- `pkg/`: Core runtime
- `cmd/goa/`: CLI source

### Build & Test

```bash
make lint          # Run linters
make test          # Run tests
cd cmd/goa && go install .  # Install CLI locally
```

### Code Generation Behavior

- After modifying goa source, `goa gen` and `goa example` automatically compile and use your changes—no manual rebuild needed.
- `goa gen` deletes and recreates the entire `gen/` directory.
- `goa example` only creates new files; it does not overwrite existing `cmd/` files.

### Repro Protocol

To reproduce a codegen issue:
1. Create `~/src/repros/<issue>/design/design.go`
2. `go mod init <issue>` in the issue directory
3. `goa gen <issue>/design`
4. `go mod tidy`
5. `go mod edit -replace goa.design/goa/v3=$HOME/src/goa`
6. `goa gen <issue>/design` again with local goa
7. Optional: `goa example <issue>/design`

### Slices/Maps and Required Fields

Do not rely on nil vs empty to encode presence. Goa uses `omitempty`—both nil and empty serialize as "missing". If empty is valid, do not mark the field as required.
