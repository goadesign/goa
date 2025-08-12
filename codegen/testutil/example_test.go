package testutil_test

import (
	"fmt"
	"testing"

	"goa.design/goa/v3/codegen/testutil"
)

// Example: Basic usage with AssertString
func TestBasicUsage(t *testing.T) {
	// Generate some code
	code := `package main

func main() {
	fmt.Println("Hello, World!")
}
`

	// Compare with golden file
	testutil.AssertString(t, "testdata/golden/hello_world.go.golden", code)
}

// Example: Using the fluent API
func TestFluentAPI(t *testing.T) {
	gf := testutil.NewGoldenFile(t, "testdata/golden")

	// Generate code
	code := generateServiceCode()

	// Use fluent API for comparison
	gf.StringContent(code).
		Path("service.go.golden").
		CompareContent()
}

// Example: Testing multiple files with batch operations
func TestBatchOperations(t *testing.T) {
	batch := testutil.NewBatch(t)

	// Generate multiple files
	serverCode := generateServerCode()
	clientCode := generateClientCode()
	typesCode := generateTypesCode()

	// Add all files to batch and compare
	batch.
		AddString("server.go.golden", serverCode).
		AddString("client.go.golden", clientCode).
		AddString("types.go.golden", typesCode).
		Compare()
}

// Example: Custom options for specific needs
func TestCustomOptions(t *testing.T) {
	opts := testutil.Options{
		BasePath:            "testdata/custom",
		ContentType:         testutil.ContentTypeGo,
		FormatCode:          true,
		NormalizeWhitespace: true,
        CreateMissing:       true, // Create golden files if they don't exist
	}

	gf := testutil.WithOptions(t, opts)
	
	code := generateComplexCode()
	
	gf.StringContent(code).
		Path("complex.go.golden").
		CompareContent()
}

// Example: Format-aware comparisons
func TestFormatAwareComparisons(t *testing.T) {
	// Test Go code - automatically formatted
	goCode := `package main
import "fmt"
func main(){fmt.Println("unformatted")}`
	
	testutil.AssertGo(t, "testdata/golden/formatted.go.golden", goCode)

	// Test JSON - automatically pretty-printed
	jsonData := []byte(`{"name":"test","value":42,"items":["a","b","c"]}`)
	
	testutil.AssertJSON(t, "testdata/golden/config.json.golden", jsonData)
}


// Example: Legacy migration
func TestLegacyMigration(t *testing.T) {
	code := generateLegacyCode()
	
	// This is a drop-in replacement for the old compareOrUpdateGolden function
	testutil.CompareOrUpdateGolden(t, code, "testdata/golden/legacy.golden")
}

// Example: Conditional golden file creation
func TestConditionalCreation(t *testing.T) {
	code := generateOptionalFeature()
	
	// Only create/compare if feature is enabled
	if code != "" {
		gf := testutil.NewGoldenFile(t, "testdata/golden")
		gf.CompareOrCreate(code, "optional_feature.golden")
	}
}

// Example: Testing with subtests
func TestWithSubtests(t *testing.T) {
	testCases := []struct {
		name     string
		generate func() string
		golden   string
	}{
		{
			name:     "simple",
			generate: generateSimpleCode,
			golden:   "simple.go.golden",
		},
		{
			name:     "complex",
			generate: generateComplexCode,
			golden:   "complex.go.golden",
		},
		{
			name:     "with_errors",
			generate: generateCodeWithErrors,
			golden:   "errors.go.golden",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			code := tc.generate()
			// Create new GoldenFile for subtest
			subGF := testutil.NewGoldenFile(t, "testdata/golden")
			subGF.StringContent(code).Path(tc.golden).CompareContent()
		})
	}
}

// Example: Parallel golden file updates
func TestParallelUpdates(t *testing.T) {
	// When running with -update, files are updated in parallel by default
	files := map[string]string{
		"parallel1.golden": generateParallel1(),
		"parallel2.golden": generateParallel2(),
		"parallel3.golden": generateParallel3(),
		"parallel4.golden": generateParallel4(),
	}

	gf := testutil.NewGoldenFile(t, "testdata/golden")
	gf.CompareMultiple(files)
}

// Helper functions for examples
func generateServiceCode() string {
	return `package service

import (
	"context"
	"log"
)

type Service interface {
	DoSomething(ctx context.Context) error
}

type serviceImpl struct {
	logger *log.Logger
}

func (s *serviceImpl) DoSomething(ctx context.Context) error {
	s.logger.Println("Doing something")
	return nil
}
`
}

func generateServerCode() string {
	return `package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, Server!"))
}
`
}

func generateClientCode() string {
	return `package client

import (
	"net/http"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{},
	}
}
`
}

func generateTypesCode() string {
	return `package types

type User struct {
	ID    string
	Name  string
	Email string
}

type Request struct {
	UserID string
}

type Response struct {
	User  *User
	Error error
}
`
}

func generateComplexCode() string {
	return generateServiceCode() + "\n" + generateTypesCode()
}


func generateLegacyCode() string {
	return "// Legacy code example\n" + generateSimpleCode()
}

func generateOptionalFeature() string {
	// Simulate optional feature generation
	if testing.Short() {
		return ""
	}
	return "// Optional feature code\n"
}

func generateSimpleCode() string {
	return `package simple

func Hello() string {
	return "Hello, World!"
}
`
}

func generateCodeWithErrors() string {
	return `package errors

import "errors"

var ErrNotFound = errors.New("not found")

func Find(id string) error {
	return ErrNotFound
}
`
}

func generateParallel1() string { return fmt.Sprintf("// Parallel 1\n%s", generateSimpleCode()) }
func generateParallel2() string { return fmt.Sprintf("// Parallel 2\n%s", generateSimpleCode()) }
func generateParallel3() string { return fmt.Sprintf("// Parallel 3\n%s", generateSimpleCode()) }
func generateParallel4() string { return fmt.Sprintf("// Parallel 4\n%s", generateSimpleCode()) }