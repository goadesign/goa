package harness

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DSLLoader loads DSL code from files
type DSLLoader struct {
	baseDir string
}

// NewDSLLoader creates a new DSL loader with the given base directory
func NewDSLLoader(baseDir string) *DSLLoader {
	return &DSLLoader{
		baseDir: baseDir,
	}
}

// Load loads DSL code from a file
func (l *DSLLoader) Load(name string) (string, error) {
	// Try with .go extension if not provided
	if !strings.HasSuffix(name, ".go") {
		name = name + ".go"
	}

	path := filepath.Join(l.baseDir, name)
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to load DSL file %s: %w", name, err)
	}

	// Extract the DSL code from the file
	// We expect files to have a specific format with the DSL inside an init() function
	dslCode := extractDSLCode(string(content))
	if dslCode == "" {
		return "", fmt.Errorf("no DSL code found in file %s", name)
	}

	return dslCode, nil
}

// LoadTemplate loads a DSL template and replaces placeholders
func (l *DSLLoader) LoadTemplate(name string, replacements map[string]string) (string, error) {
	dslCode, err := l.Load(name)
	if err != nil {
		return "", err
	}

	// Replace placeholders
	for placeholder, value := range replacements {
		dslCode = strings.ReplaceAll(dslCode, placeholder, value)
	}

	return dslCode, nil
}

// extractDSLCode extracts the DSL code from a Go file
// It looks for code between specific markers or within init() function
func extractDSLCode(content string) string {
	// Look for DSL markers
	const startMarker = "// DSL-START"
	const endMarker = "// DSL-END"
	
	startIdx := strings.Index(content, startMarker)
	if startIdx != -1 {
		startIdx += len(startMarker)
		endIdx := strings.Index(content[startIdx:], endMarker)
		if endIdx != -1 {
			return strings.TrimSpace(content[startIdx : startIdx+endIdx])
		}
	}

	// Fallback: extract content of init() function
	initStart := strings.Index(content, "func init() {")
	if initStart != -1 {
		// Find the content between the braces
		braceCount := 0
		startIdx := initStart + len("func init() {")
		
		for i := startIdx; i < len(content); i++ {
			if content[i] == '{' {
				braceCount++
			} else if content[i] == '}' {
				if braceCount == 0 {
					// Found the closing brace of init()
					return strings.TrimSpace(content[startIdx:i])
				}
				braceCount--
			}
		}
	}

	return ""
}

// ListDSLs returns a list of available DSL files
func (l *DSLLoader) ListDSLs() ([]string, error) {
	entries, err := os.ReadDir(l.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read DSL directory: %w", err)
	}

	var dsls []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			name := strings.TrimSuffix(entry.Name(), ".go")
			dsls = append(dsls, name)
		}
	}

	return dsls, nil
}