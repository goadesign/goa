//************************************************************************//
// cellar: Application User Types
//
// Generated with goagen v0.0.1, command line:
// $ goagen
// --out=$(GOPATH)/src/github.com/raphael/goa/examples/cellar
// --design=github.com/raphael/goa/examples/cellar/design
// --pkg=app
//
// The content of this file is auto-generated, DO NOT MODIFY
//************************************************************************//

package app

// BottlePayload type
type BottlePayload struct {
	Color     string
	Country   string
	Name      string
	Region    string
	Review    string
	Sweetness int
	Varietal  string
	Vineyard  string
	Vintage   int
}
