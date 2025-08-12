// Package testutil provides world-class utilities for testing code generation with golden files.
// It offers a fluent API, intelligent diffing, batch operations, and format-aware comparisons.
package testutil

import (
	"bytes"
	"encoding/json"
	"flag"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	// Global flags for updating golden files
	update = flag.Bool("update", false, "update golden files")
	u      = flag.Bool("u", false, "update golden files (shorthand)")
)

// isUpdateMode returns true if any update flag is set
func isUpdateMode() bool {
	return *update || *u
}

// DefaultBasePath is the default directory for golden files
const DefaultBasePath = "testdata/golden"

// GoldenFile manages golden file testing operations with a fluent API
type GoldenFile struct {
	t        testing.TB
	basePath string
    update   bool
	content  []byte
	path     string
}

// NewGoldenFile creates a new GoldenFile instance with default options
func NewGoldenFile(t testing.TB, basePath string) *GoldenFile {
	t.Helper()
	return &GoldenFile{
		t: t,
		basePath: func() string {
			if basePath != "" {
				return basePath
			}
			return DefaultBasePath
		}(),
	}
}

// Content sets the content to compare (fluent API)
func (g *GoldenFile) Content(content []byte) *GoldenFile {
	g.content = content
	return g
}

// StringContent sets string content to compare (fluent API)
func (g *GoldenFile) StringContent(content string) *GoldenFile {
	return g.Content([]byte(content))
}

// Path sets the golden file path (fluent API)
func (g *GoldenFile) Path(path string) *GoldenFile {
	g.path = path
	return g
}

// CompareContent performs the golden file comparison
func (g *GoldenFile) CompareContent() {
	g.t.Helper()

	if g.path == "" {
		g.t.Fatal("golden file path not set")
	}
	if g.content == nil {
		g.t.Fatal("content not set")
	}

	// Determine the full path
	goldenPath := g.path
	if !filepath.IsAbs(g.path) && g.basePath != "" {
		goldenPath = filepath.Join(g.basePath, g.path)
	}

	// Prepare content
	content := g.prepareContent()

	// Check update mode
    updateMode := g.update || isUpdateMode()

	if updateMode {
		g.updateFile(content, goldenPath)
		return
	}

	// Check if file exists
	if _, err := os.Stat(goldenPath); os.IsNotExist(err) {
		g.t.Fatalf("golden file %q does not exist (run with -update to create)", goldenPath)
	}

	g.compareContent(content, goldenPath)
}

// Compare compares the actual content with the golden file content (legacy API)
// Deprecated: Use StringContent().Path().CompareContent() for the fluent API
func (g *GoldenFile) Compare(actual string, golden string) {
	g.t.Helper()
	g.StringContent(actual).Path(golden).CompareContent()
}

// CompareBytes is like Compare but works with byte slices (legacy API)
func (g *GoldenFile) CompareBytes(actual []byte, golden string) {
	g.t.Helper()
	g.Content(actual).Path(golden).CompareContent()
}

// prepareContent applies transformations based on content type and options
func (g *GoldenFile) prepareContent() []byte {
	content := g.content

	// Auto-detect format from file extension
	isGo := strings.HasSuffix(g.path, ".go") || strings.HasSuffix(g.path, ".go.golden")
	isJSON := strings.HasSuffix(g.path, ".json") || strings.HasSuffix(g.path, ".json.golden")

	// Always format based on detected type
	switch {
	case isGo:
		if formatted, err := format.Source(content); err == nil {
			content = formatted
		}
	case isJSON:
		var v any
		if err := json.Unmarshal(content, &v); err == nil {
			if formatted, err := json.MarshalIndent(v, "", "  "); err == nil {
				content = formatted
			}
		}
	}

	// Normalize whitespace
	// Always normalize whitespace
	content = bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	content = []byte(strings.Join(lines, "\n"))
	if len(content) > 0 && content[len(content)-1] != '\n' {
		content = append(content, '\n')
	}

	return content
}

// updateFile writes content to the golden file
func (g *GoldenFile) updateFile(content []byte, goldenPath string) {
	g.t.Helper()

	// Create directory if it doesn't exist
	dir := filepath.Dir(goldenPath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		g.t.Fatalf("failed to create golden file directory %q: %v", dir, err)
	}

	// Write the golden file
	if err := os.WriteFile(goldenPath, content, 0644); err != nil {
		g.t.Fatalf("failed to update golden file %q: %v", goldenPath, err)
	}

	g.t.Logf("Updated golden file: %s", goldenPath)
}

// compareContent reads the golden file and compares with content
func (g *GoldenFile) compareContent(content []byte, goldenPath string) {
	g.t.Helper()

	golden, err := os.ReadFile(goldenPath)
	if err != nil {
		g.t.Fatalf("failed to read golden file %q: %v", goldenPath, err)
	}

	// Apply same transformations to golden content
	golden = bytes.ReplaceAll(golden, []byte("\r\n"), []byte("\n"))
	lines := strings.Split(string(golden), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	golden = []byte(strings.Join(lines, "\n"))
	if len(golden) > 0 && golden[len(golden)-1] != '\n' {
		golden = append(golden, '\n')
	}

	if !bytes.Equal(content, golden) {
		// Use testify for readable equality assertion
		require.Equalf(g.t, string(golden), string(content), "golden file mismatch for %q", goldenPath)
		g.t.Logf("Run with -update to update the golden file")
	}
}

// Exists checks if a golden file exists
func (g *GoldenFile) Exists(golden string) bool {
	goldenPath := golden
	if !filepath.IsAbs(golden) {
		goldenPath = filepath.Join(g.basePath, golden)
	}

	_, err := os.Stat(goldenPath)
	return err == nil
}

// SetUpdateMode sets whether this GoldenFile should update golden files on compare.
func (g *GoldenFile) SetUpdateMode(update bool) { g.update = update }

// CompareOrUpdateGolden provides a drop-in replacement for the legacy function
// used throughout the codebase. New code should use GoldenFile instead.
// The golden parameter should be a full path to the golden file.
func CompareOrUpdateGolden(t *testing.T, actual, golden string) {
	t.Helper()
	gf := NewGoldenFile(t, "")
	// Since this is a legacy function, golden is expected to be a full path
	// We use an absolute path to bypass the base path handling
	absGolden := golden
	if !filepath.IsAbs(golden) {
		// If it's already relative, make it absolute from current directory
		absGolden, _ = filepath.Abs(golden)
	}
	gf.StringContent(actual).Path(absGolden).CompareContent()
}

// Assert provides a simple assertion API
func Assert(t testing.TB, goldenPath string, got []byte) {
	t.Helper()
	gf := &GoldenFile{t: t, basePath: ""}
	gf.Content(got).Path(goldenPath).CompareContent()
}

// AssertString provides a simple assertion API for strings
func AssertString(t testing.TB, goldenPath string, got string) {
	t.Helper()
	gf := &GoldenFile{t: t, basePath: ""}
	gf.StringContent(got).Path(goldenPath).CompareContent()
}

// AssertJSON compares JSON content with proper formatting
func AssertJSON(t testing.TB, goldenPath string, got []byte) {
	t.Helper()
	gf := &GoldenFile{t: t, basePath: ""}
	var v any
	if err := json.Unmarshal(got, &v); err == nil {
		if formatted, err := json.MarshalIndent(v, "", "  "); err == nil {
			got = formatted
		}
	}
	gf.Content(got).Path(goldenPath).CompareContent()
}

// AssertGo compares Go source code with proper formatting
func AssertGo(t testing.TB, goldenPath string, got string) {
	t.Helper()
	gf := &GoldenFile{t: t, basePath: ""}
	formatted := []byte(got)
	if src, err := format.Source([]byte(got)); err == nil {
		formatted = src
	}
	gf.Content(formatted).Path(goldenPath).CompareContent()
}
