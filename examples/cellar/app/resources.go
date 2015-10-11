//************************************************************************//
// cellar: Application Resources
//
// Generated with codegen v0.0.1, command line:
// $ /home/raphael/go/src/github.com/raphael/goa/examples/cellar/codegen485234072/codegen
// --out=/home/raphael/go/src/github.com/raphael/goa/examples/cellar
// --design=github.com/raphael/goa/examples/cellar/design
// --force
// --pkg=app
//
// The content of this file is auto-generated, DO NOT MODIFY
//************************************************************************//

package app

import "fmt"

// Account resource
type Account struct {
	CreatedAt string `json:"created_at,omitempty"`
	CreatedBy string `json:"created_by,omitempty"`
	Href      string `json:"href,omitempty"`
	ID        int    `json:"id,omitempty"`
	Name      string `json:"name"`
}

// AccountHref returns the resource href.
func AccountHref(accountID, id string) string {
	return fmt.Sprintf("%s", accountID, id)
}

// Bottle resource
type Bottle struct {
	Account         *ExampleAccountMedia `json:"account"`
	Characteristics []string             `json:"characteristics,omitempty"`
	Color           string               `json:"color,omitempty"`
	Country         string               `json:"country,omitempty"`
	CreatedAt       string               `json:"created_at,omitempty"`
	Href            string               `json:"href,omitempty"`
	ID              int                  `json:"id,omitempty"`
	Name            string               `json:"name"`
	Region          string               `json:"region,omitempty"`
	Review          string               `json:"review,omitempty"`
	Sweet           bool                 `json:"sweet,omitempty"`
	UpdatedAt       string               `json:"updated_at,omitempty"`
	Varietal        string               `json:"varietal,omitempty"`
	Vineyard        string               `json:"vineyard"`
	Vintage         int                  `json:"vintage,omitempty"`
}

// BottleHref returns the resource href.
func BottleHref(accountID, id string) string {
	return fmt.Sprintf("%s%s", accountID, id)
}
