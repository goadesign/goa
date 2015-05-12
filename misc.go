package goa

import (
	"fmt"
	"os"
)

// fatalf prints the given error message and exits process with status 1.
func fatalf(format string, val ...interface{}) {
	fmt.Printf("goa: "+format, val)
	os.Exit(1)
}
