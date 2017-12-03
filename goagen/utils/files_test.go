package utils

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

var sep = string(os.PathSeparator)

// Test RemoveFiles. Creates new directories and files in temp directory.
// Expects that RemoveFiles delete only files, and leaving directory structure intact.
// Files from .svn directory should not be deleted.
func TestRemoveFiles(t *testing.T) {

	//Creating temp directory. When error occures, temp directory would remain, for visual check.
	testDir := filepath.Join(".", time.Now().Format("20060102150405"))
	currentDir := testDir
	createDir(currentDir, t)

	currentDir = testDir + sep + "app"
	createDir(currentDir, t)
	createFile(currentDir, "app_file.go", t)

	currentDir += sep + "test"
	createDir(currentDir, t)
	createFile(currentDir, "bookings_testing.go", t)

	currentDir = testDir + sep + "client"
	createDir(currentDir, t)
	createFile(currentDir, "bookings.go", t)

	// Add .svn ignore directory, which is not allowed to be deleted.
	currentDir = testDir + sep + "client" + sep + ".svn"
	createDir(currentDir, t)
	createFile(currentDir, "all-wcprops", t)

	// Call function RemoveFiles to be tested.
	if err := RemoveFiles(testDir); err != nil {
		t.Fatalf("Error when calling RemoveFiles %s", err)
	}

	// Check if directories were deleted.
	if !exists(testDir+sep+"app") || !exists(testDir+sep+"app"+sep+"test") || !exists(testDir+sep+"client") {
		t.Fatalf("Directories where deleted, expected to be not deleted.")
	}

	// Check if files were deleted.
	if exists(testDir+sep+"app"+sep+"app_file.go") ||
		exists(testDir+sep+"app"+sep+"test"+sep+"bookings_testing.go") ||
		exists(testDir+sep+"app"+sep+"test"+sep+"bookings") {
		t.Fatalf("Files where not deleted, expected to be deleted.")
	}

	// Check if .svn directory was deleted.
	if !exists(testDir + sep + "client" + sep + ".svn" + sep + "all-wcprops") {
		t.Fatalf("File in .svn folder was deleted, expected to be not deleted.")
	}

	// Delete temp directory only when no error occured.
	os.RemoveAll(testDir)
}

func createDir(d string, t *testing.T) {
	if err := os.MkdirAll(d, os.ModePerm); err != nil {
		t.Fatalf("Can't create temp directory " + d)
	}
}

func createFile(d, f string, t *testing.T) {
	file, err := os.OpenFile(d+sep+f, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		t.Fatalf("Can't create temp file " + d)
	}
	file.Close()
}

func exists(f string) bool {
	if _, err := os.Stat(f); err == nil {
		return true
	}
	return false
}
