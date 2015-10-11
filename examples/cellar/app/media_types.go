//************************************************************************//
// cellar: Application Media Types
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

// A tenant account
// Identifier: application/vnd.goa.example.account
type ExampleAccountMedia struct {
	CreatedAt string `json:"created_at,omitempty"`
	CreatedBy string `json:"created_by,omitempty"`
	Href      string `json:"href,omitempty"`
	ID        int    `json:"id,omitempty"`
	Name      string `json:"name"`
}

// ExampleAccountMedia views
type ExampleAccountMediaViewEnum string

const (
	// ExampleAccountMedia default view
	ExampleAccountMediaDefaultView ExampleAccountMediaViewEnum = "default"
	// ExampleAccountMedia full view
	ExampleAccountMediaFullView ExampleAccountMediaViewEnum = "full"
	// ExampleAccountMedia link view
	ExampleAccountMediaLinkView ExampleAccountMediaViewEnum = "link"
)

// A bottle of wine
// Identifier: application/vnd.goa.example.bottle
type ExampleBottleMedia struct {
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

// ExampleBottleMedia views
type ExampleBottleMediaViewEnum string

const (
	// ExampleBottleMedia default view
	ExampleBottleMediaDefaultView ExampleBottleMediaViewEnum = "default"
	// ExampleBottleMedia full view
	ExampleBottleMediaFullView ExampleBottleMediaViewEnum = "full"
)

// ExampleBottleMediaCollection media type
// Identifier: application/vnd.goa.example.bottle; type=collection
type ExampleBottleMediaCollection []*struct {
	Account         *ExampleAccountMedia `json:"account,omitempty"`
	Characteristics []string             `json:"characteristics,omitempty"`
	Color           string               `json:"color,omitempty"`
	Country         string               `json:"country,omitempty"`
	CreatedAt       string               `json:"created_at,omitempty"`
	Href            string               `json:"href,omitempty"`
	ID              int                  `json:"id,omitempty"`
	Name            string               `json:"name,omitempty"`
	Region          string               `json:"region,omitempty"`
	Review          string               `json:"review,omitempty"`
	Sweet           bool                 `json:"sweet,omitempty"`
	UpdatedAt       string               `json:"updated_at,omitempty"`
	Varietal        string               `json:"varietal,omitempty"`
	Vineyard        string               `json:"vineyard,omitempty"`
	Vintage         int                  `json:"vintage,omitempty"`
}

// ExampleBottleMediaCollection views
type ExampleBottleMediaCollectionViewEnum string

const (
	// ExampleBottleMediaCollection default view
	ExampleBottleMediaCollectionDefaultView ExampleBottleMediaCollectionViewEnum = "default"
	// ExampleBottleMediaCollection full view
	ExampleBottleMediaCollectionFullView ExampleBottleMediaCollectionViewEnum = "full"
)
