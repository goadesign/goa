//************************************************************************//
// cellar: Application Resource Href Factories
//
// Generated with codegen v0.0.1, command line:
// $ /home/raphael/go/src/github.com/raphael/goa/examples/cellar/codegen599314475/codegen
// --out=/home/raphael/go/src/github.com/raphael/goa/examples/cellar
// --design=github.com/raphael/goa/examples/cellar/design
// --force
// --pkg=app
//
// The content of this file is auto-generated, DO NOT MODIFY
//************************************************************************//

package app

import "fmt"

// AccountHref returns the resource href.
func AccountHref(accountID interface{}) string {
	return fmt.Sprintf("/cellar/accounts/%v", accountID)
}

// BottleHref returns the resource href.
func BottleHref(accountID, id interface{}) string {
	return fmt.Sprintf("/cellar/accounts/%v/bottles/%v", accountID, id)
}
