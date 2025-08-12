package testutil_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	"goa.design/goa/v3/codegen/testutil"
)

func TestGoldenFile(t *testing.T) {
	// Create a temporary directory for test golden files
	tmpDir := t.TempDir()
	
	t.Run("Compare", func(t *testing.T) {
		// Create a test golden file
		goldenPath := filepath.Join(tmpDir, "test.golden")
		expectedContent := "package main\n\nfunc main() {\n\t// Test content\n}\n"
		require.NoError(t, os.WriteFile(goldenPath, []byte(expectedContent), 0644))
		
		gf := testutil.NewGoldenFile(t, tmpDir)
		
		// Test successful comparison
		gf.Compare(expectedContent, "test.golden")
		
		// Test failed comparison would normally fail the test
		// We can't easily test this without a full mock of testing.TB
	})
	
	t.Run("Update", func(t *testing.T) {
		gf := testutil.NewGoldenFile(t, tmpDir)
		gf.SetUpdateMode(true)
		
		newContent := "updated content"
		goldenFile := "update_test.golden"
		
		// Update should create the file
		gf.Compare(newContent, goldenFile)
		
		// Verify file was created with correct content
		goldenPath := filepath.Join(tmpDir, goldenFile)
		actual, err := os.ReadFile(goldenPath)
		require.NoError(t, err)
		assert.Equal(t, newContent+"\n", string(actual))
	})
	
	t.Run("CompareOrCreate", func(t *testing.T) {
		gf := testutil.NewGoldenFile(t, tmpDir)
		// Test creating new file (using update override)
		newFile := "new_file.golden"
		content := "new file content"
		gf.SetUpdateMode(true)
		gf.Compare(content, newFile)
		// Verify file was created
		assert.True(t, gf.Exists(newFile))
		// Now compare without update
		gf.SetUpdateMode(false)
		gf.Compare(content, newFile)
		
		// Test comparing with different content would fail the test
		// We verify the file was created correctly above
	})
	
	t.Run("CompareMultiple", func(t *testing.T) {
		gf := testutil.NewGoldenFile(t, tmpDir)
		gf.SetUpdateMode(true)
		
		pairs := map[string]string{
			"file1.golden": "content 1",
			"file2.golden": "content 2",
			"file3.golden": "content 3",
		}
		
		// Create files (using update override)
		for golden, actual := range pairs {
			gf.Compare(actual, golden)
		}
		
		// Verify all files were created
		for golden, expected := range pairs {
			actual, err := os.ReadFile(filepath.Join(tmpDir, golden))
			require.NoError(t, err)
			assert.Equal(t, expected+"\n", string(actual))
		}
	})
	
	t.Run("AbsolutePath", func(t *testing.T) {
		gf := testutil.NewGoldenFile(t, tmpDir)
		gf.SetUpdateMode(true)
		
		// Test with absolute path
		absPath := filepath.Join(tmpDir, "subdir", "abs.golden")
		content := "absolute path content"
		
		gf.Compare(content, absPath)
		
		// Verify file was created at absolute path
		actual, err := os.ReadFile(absPath)
		require.NoError(t, err)
		assert.Equal(t, content+"\n", string(actual))
	})
	
	t.Run("WindowsLineEndings", func(t *testing.T) {
		// Create a golden file with Windows line endings
		goldenPath := filepath.Join(tmpDir, "windows.golden")
		windowsContent := "line1\r\nline2\r\nline3\r\n"
		require.NoError(t, os.WriteFile(goldenPath, []byte(windowsContent), 0644))
		
		gf := testutil.NewGoldenFile(t, tmpDir)
		
		// Compare with Unix line endings (should pass due to normalization)
		unixContent := "line1\nline2\nline3\n"
		gf.Compare(unixContent, "windows.golden")
	})
}