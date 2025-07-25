# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## About This Project

Goa is a design-first API framework for Go that generates production-ready code
from a declarative DSL. The framework allows developers to define their API
design using a Domain Specific Language (DSL) and then generates server
interfaces, client code, transport adapters, and documentation automatically.

## Common Commands

### Development Commands
- `make all` - Run lint and test (default target)
- `make lint` - Run golangci-lint on all Go files
- `make test` - Run all tests with coverage output to `cover.out`

### Code Generation Commands
- `goa gen <package>` - Generate service interfaces, endpoints, transport code and OpenAPI spec from design package
- `goa example <package>` - Generate example server and client tool from design package
- `goa version` - Print version information

### Testing Specific Components
- Run tests for specific packages: `go test ./dsl/... ./expr/... ./codegen/...`
- Test with verbose output: `go test -v ./...`
- Run single test file: `go test ./path/to/package -run TestName`

## Code Layout

The Goa project favors short-ish files (ideally no more than 2,000 lines of code
per file).  Each file should encapsulate one main construct and potentially
smaller satellite constructs used by the main construct. A construct may consist
of a strut, an interface and related factory functions and methods.

Each file should be organized as follows:

1. Type declarations in a single type ( ) block, public type first then private
   types. The organization should list top level constructs first then
   dependencies.
2. Constant declarations, public then private
3. Variable declarations, public then private
4. Public functions
5. Public methods
6. Private functions
7. Private methods

Do not put markers between these sections, this is solely for organization.

## Architecture Overview

### Core Packages
- **`dsl/`** - Domain Specific Language for defining APIs
  - Contains all DSL functions like `API`, `Service`, `Method`, `Attribute`, etc.
  - Uses "dot imports" pattern for better design readability
  - Entry point for all API designs

- **`expr/`** - Expression data types and processing
  - Transforms DSL into internal expression representation
  - Implements Preparer, Validator, and Finalizer interfaces
  - Contains primitive types (bool, string, integers), arrays, maps, objects, and user types

- **`codegen/`** - Code generation engine
  - File generation framework with sections and templates
  - Contains the `GoTransform` functionality for type transformations
  - Supports multiple output formats and generators

- **`cmd/goa/`** - Main CLI tool
  - Supports `gen`, `example`, and `version` commands
  - Handles command-line argument parsing and generator orchestration

### Transport Packages
- **`http/`** - HTTP transport implementation
  - Client and server code generation
  - OpenAPI v2/v3 specification generation
  - Middleware, encoding, error handling
  - WebSocket and Server-Sent Events (SSE) support

- **`grpc/`** - gRPC transport implementation
  - Protocol buffer generation
  - Client and server stub generation
  - Streaming support

### Supporting Packages
- **`pkg/`** - Core runtime utilities (endpoint, error handling, validation)
- **`middleware/`** - Common middleware implementations
- **`security/`** - Security scheme definitions
- **`eval/`** - Expression evaluation engine

### Code Generation Flow
1. **Design Definition** - Write API design using DSL in a design package
2. **DSL Processing** - DSL functions create expression trees in memory
3. **Expression Building** - Raw expressions are prepared, validated, and finalized
4. **Code Generation** - Templates are executed against expressions to generate code
5. **Output** - Service interfaces, transport code, client libraries, and documentation

### Key Design Patterns
- **Design-First Approach** - All code is generated from the design specification
- **Transport Agnostic** - Same design generates HTTP and gRPC implementations
- **Template-Based Generation** - Extensive use of Go templates for code generation
- **Clean Separation** - Business logic separate from transport concerns
- **Type Safety** - Comprehensive type checking throughout the pipeline

## Testing Strategy

The codebase uses extensive testing:
- **Golden File Testing** - Many tests compare generated output against golden files in `testdata/` directories
- **DSL Testing** - Test files often contain example DSL definitions in `testdata/*_dsls.go`
- **Cross-Package Testing** - Tests verify integration between DSL, expressions, and code generation
- **Protocol Buffer Testing** - Special protoc-based tests for gRPC functionality

## Development Workflow

1. Make changes to DSL, expression, or codegen packages
2. Run `make lint` to check code style
3. Run `make test` to verify all tests pass
4. For template changes, verify golden files are updated correctly
5. Test with real examples using `goa gen` and `goa example` commands

## Branch Naming Conventions

When creating branches for development work, use the following naming conventions:

- **`fix/`** - For bug fixes (e.g., `fix/conversion-files-per-package-issue-3745`)
- **`feature/`** - For new features (e.g., `feature/add-websocket-support`)
- **`refactor/`** - For refactoring work (e.g., `refactor/simplify-codegen-pipeline`) 
- **`docs/`** - For documentation changes (e.g., `docs/update-getting-started`)
- **`test/`** - For test improvements (e.g., `test/add-integration-tests`)
- **`chore/`** - For maintenance tasks (e.g., `chore/update-dependencies`)

Include the GitHub issue number when applicable, and use descriptive names that summarize the work being done. Use kebab-case (lowercase with hyphens) for branch names.

## How to Reproduce an Issue

Here is the issue repro protocol that should be followed when needing to generate the
code from a specific design to reproduce an issue.

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

NOTE: there is no need to recompile goa after changing its code because the code
generation process includes compiling a temporary tool with the goa code which
will include whatever change was made to its source code.

NOTE2: The `goa gen` command will wipe out the `gen` folder but the
`goa example` command will NOT override pre-existing files (the `cmd` folder and
the top level service files).
