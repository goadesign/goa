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
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	// Global flags for updating golden files
	updateGolden = flag.Bool("update", false, "update golden files")
	u            = flag.Bool("u", false, "update golden files (shorthand)")
	w            = flag.Bool("w", false, "update golden files (legacy compatibility)")

	// Parallel update control
	parallelUpdate = flag.Bool("golden.parallel", true, "update golden files in parallel")

	// Global registry for tracking golden file operations
	goldenRegistry = &registry{
		files: make(map[string]bool),
		mu:    sync.RWMutex{},
	}
)

// registry tracks golden file operations to prevent conflicts
type registry struct {
	files map[string]bool
	mu    sync.RWMutex
}

func (r *registry) register(path string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.files[path] {
		return false
	}
	r.files[path] = true
	return true
}

func (r *registry) unregister(path string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.files, path)
}

// isUpdateMode returns true if any update flag is set
func isUpdateMode() bool {
	return *updateGolden || *u || *w
}

// ContentType specifies the type of content for format-aware operations
type ContentType int

const (
	// ContentTypeAuto detects content type from file extension
	ContentTypeAuto ContentType = iota
	// ContentTypeGo indicates Go source code
	ContentTypeGo
	// ContentTypeJSON indicates JSON data
	ContentTypeJSON
	// ContentTypeText indicates plain text
	ContentTypeText
	// ContentTypeGoTemplate indicates Go template code
	ContentTypeGoTemplate
)

// Options configures golden file operations
type Options struct {
	// BasePath is the base directory for golden files (default: "testdata/golden")
	BasePath string

	// ContentType specifies the content type for formatting
	ContentType ContentType

	// FormatCode formats Go code before comparison (default: true for .go files)
	FormatCode bool

	// NormalizeWhitespace trims trailing whitespace and ensures consistent line endings
	NormalizeWhitespace bool

	// CreateMissing creates golden files if they don't exist
	CreateMissing bool


	// FileMode controls file permissions (default: 0644)
	FileMode os.FileMode

	// UpdateMode allows overriding the global update mode
	UpdateMode *bool
}

// DefaultOptions returns sensible defaults for most use cases
func DefaultOptions() Options {
	return Options{
		BasePath:            filepath.Join("testdata", "golden"),
		ContentType:         ContentTypeAuto,
		FormatCode:          true,
		NormalizeWhitespace: true,
        CreateMissing:       false,
		FileMode:            0644,
	}
}

// GoldenFile manages golden file testing operations with a fluent API
type GoldenFile struct {
	t       testing.TB
	options Options
	content []byte
	path    string
}

// NewGoldenFile creates a new GoldenFile instance with default options
func NewGoldenFile(t testing.TB, basePath string) *GoldenFile {
	t.Helper()
	opts := DefaultOptions()
	if basePath != "" {
		opts.BasePath = basePath
	}
	return &GoldenFile{
		t:       t,
		options: opts,
	}
}

// WithOptions creates a new GoldenFile instance with custom options
func WithOptions(t testing.TB, opts Options) *GoldenFile {
	t.Helper()
	// Fill in defaults for unset options
	if opts.BasePath == "" {
		opts.BasePath = DefaultOptions().BasePath
	}
	if opts.FileMode == 0 {
		opts.FileMode = DefaultOptions().FileMode
	}
	return &GoldenFile{
		t:       t,
		options: opts,
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
	if !filepath.IsAbs(g.path) && g.options.BasePath != "" {
		goldenPath = filepath.Join(g.options.BasePath, g.path)
	}

	// Register the file to prevent concurrent access
	if !goldenRegistry.register(goldenPath) {
		g.t.Fatalf("golden file %q is already being processed by another test", goldenPath)
	}
	defer goldenRegistry.unregister(goldenPath)

	// Prepare content
	content := g.prepareContent()

	// Check update mode
	updateMode := isUpdateMode()
	if g.options.UpdateMode != nil {
		updateMode = *g.options.UpdateMode
	}

	if updateMode {
		g.updateFile(content, goldenPath)
		return
	}

	// Check if file exists
	if _, err := os.Stat(goldenPath); os.IsNotExist(err) {
		if g.options.CreateMissing {
			g.updateFile(content, goldenPath)
			g.t.Logf("Created new golden file: %s", goldenPath)
			return
		}
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

	// Detect content type if auto
	contentType := g.options.ContentType
	if contentType == ContentTypeAuto && g.path != "" {
		switch {
		case strings.HasSuffix(g.path, ".go"):
			contentType = ContentTypeGo
		case strings.HasSuffix(g.path, ".json"):
			contentType = ContentTypeJSON
		case strings.HasSuffix(g.path, ".tmpl") || strings.HasSuffix(g.path, ".gotmpl"):
			contentType = ContentTypeGoTemplate
		default:
			contentType = ContentTypeText
		}
	}

	// Format based on content type
	if g.options.FormatCode {
		switch contentType {
		case ContentTypeGo:
			if formatted, err := format.Source(content); err == nil {
				content = formatted
			}
		case ContentTypeJSON:
			var v any
			if err := json.Unmarshal(content, &v); err == nil {
				if formatted, err := json.MarshalIndent(v, "", "  "); err == nil {
					content = formatted
				}
			}
		}
	}

	// Normalize whitespace
	if g.options.NormalizeWhitespace {
		// Convert Windows line endings to Unix
		content = bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
		// Trim trailing whitespace from each line
		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			lines[i] = strings.TrimRight(line, " \t")
		}
		content = []byte(strings.Join(lines, "\n"))
		// Ensure file ends with newline
		if len(content) > 0 && content[len(content)-1] != '\n' {
			content = append(content, '\n')
		}
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
	if err := os.WriteFile(goldenPath, content, g.options.FileMode); err != nil {
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
	if g.options.NormalizeWhitespace {
		golden = bytes.ReplaceAll(golden, []byte("\r\n"), []byte("\n"))
		lines := strings.Split(string(golden), "\n")
		for i, line := range lines {
			lines[i] = strings.TrimRight(line, " \t")
		}
		golden = []byte(strings.Join(lines, "\n"))
		if len(golden) > 0 && golden[len(golden)-1] != '\n' {
			golden = append(golden, '\n')
		}
	}

	if !bytes.Equal(content, golden) {
		// Use testify for readable equality assertion
		require.Equalf(g.t, string(golden), string(content), "golden file mismatch for %q", goldenPath)
		g.t.Logf("Run with -update to update the golden file")
	}
}

// IsUpdateMode returns true if golden file update mode is enabled
func (g *GoldenFile) IsUpdateMode() bool {
	if g.options.UpdateMode != nil {
		return *g.options.UpdateMode
	}
	return isUpdateMode()
}

// SetUpdateMode allows overriding the update mode for specific tests
func (g *GoldenFile) SetUpdateMode(update bool) {
	g.options.UpdateMode = &update
}

// Exists checks if a golden file exists
func (g *GoldenFile) Exists(golden string) bool {
	goldenPath := golden
	if !filepath.IsAbs(golden) {
		goldenPath = filepath.Join(g.options.BasePath, golden)
	}

	_, err := os.Stat(goldenPath)
	return err == nil
}

// CompareOrCreate compares content with a golden file if it exists,
// or creates it if it doesn't exist (useful for initial test creation)
func (g *GoldenFile) CompareOrCreate(actual string, golden string) {
	g.t.Helper()

	// Temporarily enable CreateMissing
	origCreateMissing := g.options.CreateMissing
	g.options.CreateMissing = true
	defer func() { g.options.CreateMissing = origCreateMissing }()

	g.StringContent(actual).Path(golden).CompareContent()
}

// CompareMultiple compares multiple actual/golden file pairs
// The pairs parameter is a map where keys are golden file names and values are the actual content
func (g *GoldenFile) CompareMultiple(pairs map[string]string) {
	g.t.Helper()

	// Type assert to *testing.T to use Run method
	t, ok := g.t.(*testing.T)
	if !ok {
		// If not a *testing.T, just compare directly without subtests
		for golden, actual := range pairs {
			newG := &GoldenFile{t: g.t, options: g.options}
			newG.StringContent(actual).Path(golden).CompareContent()
		}
		return
	}

	if *parallelUpdate && isUpdateMode() {
		// Update files in parallel
		var wg sync.WaitGroup
		for golden, actual := range pairs {
			wg.Add(1)
			go func(golden, actual string) {
				defer wg.Done()
				newG := &GoldenFile{t: g.t, options: g.options}
				newG.StringContent(actual).Path(golden).CompareContent()
			}(golden, actual)
		}
		wg.Wait()
	} else {
		// Run as subtests
		for golden, actual := range pairs {
			t.Run(filepath.Base(golden), func(t *testing.T) {
				// Create a new GoldenFile instance to use the sub-test's t
				subGolden := WithOptions(t, g.options)
				subGolden.StringContent(actual).Path(golden).CompareContent()
			})
		}
	}
}

// Batch provides batch operations for multiple golden files
type Batch struct {
	t       testing.TB
	options Options
	files   []batchFile
}

type batchFile struct {
	path    string
	content []byte
}

// NewBatch creates a new batch operation
func NewBatch(t testing.TB, opts ...Options) *Batch {
	t.Helper()
	options := DefaultOptions()
	if len(opts) > 0 {
		options = opts[0]
	}
	return &Batch{
		t:       t,
		options: options,
		files:   make([]batchFile, 0),
	}
}

// Add adds a file to the batch
func (b *Batch) Add(path string, content []byte) *Batch {
	b.files = append(b.files, batchFile{path: path, content: content})
	return b
}

// AddString adds a file with string content to the batch
func (b *Batch) AddString(path string, content string) *Batch {
	return b.Add(path, []byte(content))
}

// Compare performs all comparisons in the batch
func (b *Batch) Compare() {
	b.t.Helper()

	if *parallelUpdate && isUpdateMode() {
		// Update files in parallel
		var wg sync.WaitGroup
		for _, file := range b.files {
			wg.Add(1)
			go func(file batchFile) {
				defer wg.Done()
				g := WithOptions(b.t, b.options)
				g.Content(file.content).Path(file.path).CompareContent()
			}(file)
		}
		wg.Wait()
	} else {
		// Compare sequentially
		for _, file := range b.files {
			g := WithOptions(b.t, b.options)
			g.Content(file.content).Path(file.path).CompareContent()
		}
	}
}

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
	gf := &GoldenFile{t: t, options: DefaultOptions()}
	gf.options.BasePath = ""
	gf.Content(got).Path(goldenPath).CompareContent()
}

// AssertString provides a simple assertion API for strings
func AssertString(t testing.TB, goldenPath string, got string) {
	t.Helper()
	gf := &GoldenFile{t: t, options: DefaultOptions()}
	gf.options.BasePath = ""
	gf.StringContent(got).Path(goldenPath).CompareContent()
}

// AssertJSON compares JSON content with proper formatting
func AssertJSON(t testing.TB, goldenPath string, got []byte) {
	t.Helper()
	gf := &GoldenFile{t: t, options: DefaultOptions()}
	gf.options.BasePath = ""
	gf.options.ContentType = ContentTypeJSON
	gf.Content(got).Path(goldenPath).CompareContent()
}

// AssertGo compares Go source code with proper formatting
func AssertGo(t testing.TB, goldenPath string, got string) {
	t.Helper()
	gf := &GoldenFile{t: t, options: DefaultOptions()}
	gf.options.BasePath = ""
	gf.options.ContentType = ContentTypeGo
	gf.StringContent(got).Path(goldenPath).CompareContent()
}
