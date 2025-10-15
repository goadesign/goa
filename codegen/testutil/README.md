# Package testutil

Package testutil provides utilities for testing code generation using golden files.

## Overview

Golden file testing compares generated output against expected output stored in
files. When code generation changes, you can review the differences and update
the golden files if the changes are correct.

This package provides:
- Simple assertion functions for common cases
- A fluent API for advanced scenarios
- Automatic formatting for Go code and JSON
- Cross-platform line ending normalization
- Batch operations for testing multiple files

## Basic Usage

The simplest way to test generated code:

```go
func TestCodeGen(t *testing.T) {
    // Generate your code
    code := generateSomeCode()
    
    // Compare with golden file
    testutil.AssertString(t, "testdata/golden/expected.golden", code)
}
```

Run tests normally:
```bash
go test
```

Update golden files when output changes:
```bash
go test -update
```

## Common Patterns

### Testing Multiple Files

When generating multiple related files:

```go
func TestMultipleFiles(t *testing.T) {
    batch := testutil.NewBatch(t)
    
    batch.AddString("server.go.golden", generateServer()).
          AddString("client.go.golden", generateClient()).
          AddString("types.go.golden", generateTypes()).
          Compare()
}
```

### Format-Specific Testing

The package automatically formats content based on file type:

```go
func TestFormattedOutput(t *testing.T) {
    // Go code is automatically formatted
    goCode := generateGoCode() // even if unformatted
    testutil.AssertGo(t, "output.go.golden", goCode)
    
    // JSON is pretty-printed
    jsonData := generateJSON() // even if minified
    testutil.AssertJSON(t, "config.json.golden", jsonData)
}
```

## Testing Multiple Generated Files

When testing code generators that produce multiple files, you can use batch
operations to test them all at once. This ensures all generated files remain
consistent with each other.

## Advanced Usage

### Fluent API

For more control over the comparison process:

```go
func TestWithFluentAPI(t *testing.T) {
    gf := testutil.NewGoldenFile(t, "testdata/golden")
    
    code := generateCode()
    
    gf.StringContent(code).
       Path("service.go.golden").
       CompareContent()
}
```

### Directory Comparison

Compare entire directory structures:

```go
func TestGeneratedDirectory(t *testing.T) {
    // Generate files to a directory
    generateToDirectory("./generated")
    
    // Compare against golden directory
    snapshot := testutil.NewDirSnapshot(t, "./generated", "testdata/golden/expected")
    snapshot.Ignore("*.tmp", "*.log").Compare()
}
```

## Command Line Flags

```bash
# Update golden files
go test -update    # or -u or -w

# Sequential updates (for debugging)
go test -update -golden.parallel=false
```

## File Organization

Golden files are typically stored in `testdata/golden/` directories. This is the default when using `NewGoldenFile` with an empty base path:

```go
// Uses "testdata/golden" as base path
gf := testutil.NewGoldenFile(t, "")
gf.StringContent(code).Path("output.golden").CompareContent()
// Creates: testdata/golden/output.golden

// Uses custom base path
gf := testutil.NewGoldenFile(t, "testdata/custom")
gf.StringContent(code).Path("output.golden").CompareContent()
// Creates: testdata/custom/output.golden

// Assert functions use paths exactly as provided
testutil.AssertString(t, "testdata/golden/output.golden", code)
// Creates: testdata/golden/output.golden
```

Typical directory structure:
```
mypackage/
├── generator.go
├── generator_test.go
└── testdata/
    └── golden/
        ├── server.go.golden
        ├── client.go.golden
        └── types.go.golden
```

## Content Type Detection

The package automatically detects and formats content based on file extensions:

- `.go` or `.go.golden`: Formatted with `go/format`
- `.json` or `.json.golden`: Pretty-printed with proper indentation
- Other extensions: Treated as plain text
- Other extensions: Treated as plain text

## Notes

- Golden files should be committed to version control
- Always review diffs carefully before updating golden files
- Line endings are automatically normalized for cross-platform compatibility
- The package ensures thread-safe access to golden files

## API Reference

See the [package documentation](https://pkg.go.dev/goa.design/goa/v3/codegen/testutil) for complete API details.