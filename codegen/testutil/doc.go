// Package testutil provides testing utilities for the Goa code generation framework.
//
// Golden File Testing
//
// The package provides utilities for golden file testing, a technique where
// expected outputs are stored in files and compared against actual outputs
// during tests. This is particularly useful for testing code generation where
// outputs can be large and complex.
//
// Basic Usage:
//
//	func TestCodeGeneration(t *testing.T) {
//		// Create a golden file manager
//		gf := testutil.NewGoldenFile(t, "testdata/golden")
//		
//		// Generate code
//		code := generateCode()
//		
//		// Compare with golden file
//		gf.Compare(code, "mytest.golden")
//	}
//
// Updating Golden Files:
//
// To update golden files when the expected output changes, run tests with
// the -update flag:
//
//	go test ./... -update
//
// Legacy Compatibility:
//
// For backward compatibility with existing tests, use CompareOrUpdateGolden:
//
//	testutil.CompareOrUpdateGolden(t, actual, "path/to/file.golden")
//
// This function uses the same -update flag but requires the full path to the
// golden file.
//
// Advanced Usage:
//
// The GoldenFile type provides additional methods for more complex scenarios:
//
//	// Compare multiple files at once
//	gf.CompareMultiple(map[string]string{
//		"file1.golden": code1,
//		"file2.golden": code2,
//	})
//	
//	// Create golden file if it doesn't exist
//	gf.CompareOrCreate(code, "new.golden")
//	
//	// Check if a golden file exists
//	if gf.Exists("optional.golden") {
//		gf.Compare(code, "optional.golden")
//	}
//
// Organization:
//
// Golden files are typically organized under a testdata/golden directory
// within each package. This keeps test data close to the tests while
// maintaining a clean structure.
package testutil