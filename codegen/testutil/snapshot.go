// Package testutil provides utilities for testing code generation.
package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"goa.design/goa/v3/codegen"
)

// Snapshot provides advanced snapshot testing for generated code
type Snapshot struct {
	t       testing.TB
	name    string
	options Options
	files   map[string]*codegen.File
}

// NewSnapshot creates a new snapshot test
func NewSnapshot(t testing.TB, name string, opts ...Options) *Snapshot {
	t.Helper()
	options := DefaultOptions()
	if len(opts) > 0 {
		options = opts[0]
	}
	return &Snapshot{
		t:       t,
		name:    name,
		options: options,
		files:   make(map[string]*codegen.File),
	}
}

// AddFile adds a generated file to the snapshot
func (s *Snapshot) AddFile(file *codegen.File) *Snapshot {
	if file == nil {
		return s
	}
	s.files[file.Path] = file
	return s
}

// AddFiles adds multiple generated files to the snapshot
func (s *Snapshot) AddFiles(files []*codegen.File) *Snapshot {
	for _, file := range files {
		s.AddFile(file)
	}
	return s
}

// Compare compares all files in the snapshot against golden files
func (s *Snapshot) Compare() {
	s.t.Helper()
	
	// Create sorted list of paths for deterministic order
	paths := make([]string, 0, len(s.files))
	for path := range s.files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	
	// If we're a *testing.T, use subtests
	if t, ok := s.t.(*testing.T); ok {
		for _, path := range paths {
			file := s.files[path]
			t.Run(sanitizeTestName(path), func(t *testing.T) {
				s.compareFile(t, file)
			})
		}
	} else {
		// Otherwise compare directly
		for _, path := range paths {
			s.compareFile(s.t, s.files[path])
		}
	}
}

// compareFile compares a single generated file
func (s *Snapshot) compareFile(t testing.TB, file *codegen.File) {
	t.Helper()
	
	// Get the file content by executing templates
	var buf strings.Builder
	for _, section := range file.SectionTemplates {
		if section.FuncMap != nil {
			continue // Skip sections with function maps for now
		}
		buf.WriteString(section.Source)
		if !strings.HasSuffix(section.Source, "\n") {
			buf.WriteString("\n")
		}
	}
	content := buf.String()
	
	// Determine golden path
	goldenPath := s.goldenPath(file.Path)
	
	// Use GoldenFile for comparison
	gf := WithOptions(t, s.options)
	gf.StringContent(content).Path(goldenPath).CompareContent()
}

// goldenPath generates the golden file path for a given file
func (s *Snapshot) goldenPath(filePath string) string {
	// Replace path separators with underscores for flat structure
	safeName := strings.ReplaceAll(filePath, string(filepath.Separator), "_")
	// Remove leading underscore if present
	safeName = strings.TrimPrefix(safeName, "_")
	
	// Add snapshot name as prefix if provided
	if s.name != "" {
		safeName = fmt.Sprintf("%s_%s", s.name, safeName)
	}
	
	// Ensure .golden extension
	if !strings.HasSuffix(safeName, ".golden") {
		safeName += ".golden"
	}
	
	return safeName
}

// sanitizeTestName creates a valid test name from a file path
func sanitizeTestName(path string) string {
	// Replace problematic characters
	name := strings.ReplaceAll(path, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.TrimPrefix(name, "_")
	return name
}

// SnapshotFiles provides a simpler API for common snapshot testing
func SnapshotFiles(t testing.TB, name string, files []*codegen.File) {
	t.Helper()
	NewSnapshot(t, name).AddFiles(files).Compare()
}

// SnapshotService captures all files generated for a service
type SnapshotService struct {
	t       testing.TB
	name    string
	options Options
	groups  map[string][]*codegen.File
}

// NewSnapshotService creates a service-oriented snapshot test
func NewSnapshotService(t testing.TB, serviceName string, opts ...Options) *SnapshotService {
	t.Helper()
	options := DefaultOptions()
	if len(opts) > 0 {
		options = opts[0]
	}
	return &SnapshotService{
		t:       t,
		name:    serviceName,
		options: options,
		groups:  make(map[string][]*codegen.File),
	}
}

// AddGroup adds a group of files (e.g., "server", "client", "types")
func (ss *SnapshotService) AddGroup(name string, files []*codegen.File) *SnapshotService {
	ss.groups[name] = files
	return ss
}

// Compare runs comparison for all groups
func (ss *SnapshotService) Compare() {
	ss.t.Helper()
	
	// Sort group names for deterministic order
	groupNames := make([]string, 0, len(ss.groups))
	for name := range ss.groups {
		groupNames = append(groupNames, name)
	}
	sort.Strings(groupNames)
	
	// If we're a *testing.T, use subtests for groups
	if t, ok := ss.t.(*testing.T); ok {
		for _, groupName := range groupNames {
			files := ss.groups[groupName]
			t.Run(groupName, func(t *testing.T) {
				snapshot := NewSnapshot(t, fmt.Sprintf("%s_%s", ss.name, groupName), ss.options)
				snapshot.AddFiles(files).Compare()
			})
		}
	} else {
		// Otherwise compare directly
		for _, groupName := range groupNames {
			snapshot := NewSnapshot(ss.t, fmt.Sprintf("%s_%s", ss.name, groupName), ss.options)
			snapshot.AddFiles(ss.groups[groupName]).Compare()
		}
	}
}

// DirSnapshot compares an entire directory structure
type DirSnapshot struct {
	t         testing.TB
	sourceDir string
	goldenDir string
	options   Options
	ignore    []string
}

// NewDirSnapshot creates a directory snapshot comparison
func NewDirSnapshot(t testing.TB, sourceDir, goldenDir string, opts ...Options) *DirSnapshot {
	t.Helper()
	options := DefaultOptions()
	if len(opts) > 0 {
		options = opts[0]
	}
	
	// Default golden dir if not specified
	if goldenDir == "" {
		goldenDir = filepath.Join(options.BasePath, "snapshots", filepath.Base(sourceDir))
	}
	
	return &DirSnapshot{
		t:         t,
		sourceDir: sourceDir,
		goldenDir: goldenDir,
		options:   options,
		ignore: []string{
			".git",
			"node_modules",
			"vendor",
			"*.test",
			"*.golden",
		},
	}
}

// Ignore adds patterns to ignore during comparison
func (ds *DirSnapshot) Ignore(patterns ...string) *DirSnapshot {
	ds.ignore = append(ds.ignore, patterns...)
	return ds
}

// Compare performs the directory comparison
func (ds *DirSnapshot) Compare() {
	ds.t.Helper()
	
	// Walk the source directory
	err := filepath.Walk(ds.sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories and ignored files
		if info.IsDir() || ds.shouldIgnore(path) {
			if info.IsDir() && ds.shouldIgnore(path) {
				return filepath.SkipDir
			}
			return nil
		}
		
		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %q: %w", path, err)
		}
		
		// Calculate relative path
		relPath, err := filepath.Rel(ds.sourceDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}
		
		// Determine golden path
		goldenPath := filepath.Join(ds.goldenDir, relPath)
		
		// Compare using GoldenFile
		gf := WithOptions(ds.t, ds.options)
		gf.Content(content).Path(goldenPath).CompareContent()
		
		return nil
	})
	
	if err != nil {
		ds.t.Fatalf("directory walk failed: %v", err)
	}
}

// shouldIgnore checks if a path matches any ignore pattern
func (ds *DirSnapshot) shouldIgnore(path string) bool {
	base := filepath.Base(path)
	for _, pattern := range ds.ignore {
		if matched, _ := filepath.Match(pattern, base); matched {
			return true
		}
		// Also check against full path
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}
	}
	return false
}