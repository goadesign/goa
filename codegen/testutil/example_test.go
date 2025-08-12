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

// Example: Format-aware comparisons
func TestFormatAwareComparisons(t *testing.T) {
	// Test Go code - automatically formatted (auto-detected by .go.golden)
	goCode := `package main
import "fmt"
func main(){fmt.Println("unformatted")}`

	testutil.AssertGo(t, "testdata/golden/formatted.go.golden", goCode)

	// Test JSON - automatically pretty-printed (auto-detected by .json.golden)
	jsonData := []byte(`{"name":"test","value":42,"items":["a","b","c"]}`)

	testutil.AssertJSON(t, "testdata/golden/config.json.golden", jsonData)
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
	files := map[string]string{}
	for i := 1; i <= 4; i++ {
		files[fmt.Sprintf("parallel%d.golden", i)] = generateParallel(i)
	}

	for golden, actual := range files {
		gf := testutil.NewGoldenFile(t, "testdata/golden")
		gf.StringContent(actual).Path(golden).CompareContent()
	}
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

// removed verbose server/client helpers to keep file concise

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

// removed legacy code generator

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

func generateParallel(i int) string {
	return fmt.Sprintf("// Parallel %d\n%s", i, generateSimpleCode())
}
