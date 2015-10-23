//************************************************************************//
// cellar: Application Resource Href Factories
//
// Generated with codegen v0.0.1, command line:
// $ codegen
// --out=/home/raphael/go/src/github.com/raphael/goa/examples/cellar
// --design=github.com/raphael/goa/examples/cellar/design
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
