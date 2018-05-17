// Code generated by goa v2.0.0-wip, DO NOT EDIT.
//
// sommelier views
//
// Command:
// $ goa gen goa.design/goa/examples/cellar/design -o
// $(GOPATH)/src/goa.design/goa/examples/cellar

package views

import (
	"unicode/utf8"

	goa "goa.design/goa"
)

// StoredBottleCollection is the viewed result type that is projected based on
// a view.
type StoredBottleCollection []*StoredBottle

// StoredBottleView is a type that runs validations on a projected type.
type StoredBottleView struct {
	// ID is the unique id of the bottle.
	ID *string
	// Name of bottle
	Name *string
	// Winery that produces wine
	Winery *Winery
	// Vintage of bottle
	Vintage *uint32
	// Composition is the list of grape varietals and associated percentage.
	Composition []*Component
	// Description of bottle
	Description *string
	// Rating of bottle from 1 (worst) to 5 (best)
	Rating *uint32
}

// StoredBottle is the viewed result type that is projected based on a view.
type StoredBottle struct {
	// Type to project
	Projected *StoredBottleView
	// View to render
	View string
}

// WineryView is a type that runs validations on a projected type.
type WineryView struct {
	// Name of winery
	Name *string
	// Region of winery
	Region *string
	// Country of winery
	Country *string
	// Winery website URL
	URL *string
}

// Winery is the viewed result type that is projected based on a view.
type Winery struct {
	// Type to project
	Projected *WineryView
	// View to render
	View string
}

// Component is a type that runs validations on a projected type.
type Component struct {
	// Grape varietal
	Varietal *string
	// Percentage of varietal in wine
	Percentage *uint32
}

// Validate runs the validations defined on StoredBottleCollection.
func (result StoredBottleCollection) Validate() (err error) {
	for _, projected := range result {
		if err2 := projected.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// Validate runs the validations defined on StoredBottle.
func (result *StoredBottle) Validate() (err error) {
	projected := result.Projected
	switch result.View {
	case "tiny":
		if projected.ID == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("id", "projected"))
		}
		if projected.Name == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("name", "projected"))
		}
		if projected.Winery == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("winery", "projected"))
		}
		if projected.Name != nil {
			if utf8.RuneCountInString(*projected.Name) > 100 {
				err = goa.MergeErrors(err, goa.InvalidLengthError("projected.name", *projected.Name, utf8.RuneCountInString(*projected.Name), 100, false))
			}
		}
		if projected.Winery != nil {
			if err2 := projected.Winery.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
	default:
		if projected.ID == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("id", "projected"))
		}
		if projected.Name == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("name", "projected"))
		}
		if projected.Winery == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("winery", "projected"))
		}
		if projected.Vintage == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("vintage", "projected"))
		}
		if projected.Name != nil {
			if utf8.RuneCountInString(*projected.Name) > 100 {
				err = goa.MergeErrors(err, goa.InvalidLengthError("projected.name", *projected.Name, utf8.RuneCountInString(*projected.Name), 100, false))
			}
		}
		if projected.Winery != nil {
			if err2 := projected.Winery.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
		if projected.Vintage != nil {
			if *projected.Vintage < 1900 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("projected.vintage", *projected.Vintage, 1900, true))
			}
		}
		if projected.Vintage != nil {
			if *projected.Vintage > 2020 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("projected.vintage", *projected.Vintage, 2020, false))
			}
		}
		for _, e := range projected.Composition {
			if e != nil {
				if err2 := e.Validate(); err2 != nil {
					err = goa.MergeErrors(err, err2)
				}
			}
		}
		if projected.Description != nil {
			if utf8.RuneCountInString(*projected.Description) > 2000 {
				err = goa.MergeErrors(err, goa.InvalidLengthError("projected.description", *projected.Description, utf8.RuneCountInString(*projected.Description), 2000, false))
			}
		}
		if projected.Rating != nil {
			if *projected.Rating < 1 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("projected.rating", *projected.Rating, 1, true))
			}
		}
		if projected.Rating != nil {
			if *projected.Rating > 5 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("projected.rating", *projected.Rating, 5, false))
			}
		}
	}
	return
}

// Validate runs the validations defined on Winery.
func (result *Winery) Validate() (err error) {
	projected := result.Projected
	switch result.View {
	case "tiny":
		if projected.Name == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("name", "projected"))
		}
	default:
		if projected.Name == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("name", "projected"))
		}
		if projected.Region == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("region", "projected"))
		}
		if projected.Country == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("country", "projected"))
		}
		if projected.Region != nil {
			err = goa.MergeErrors(err, goa.ValidatePattern("projected.region", *projected.Region, "(?i)[a-z '\\.]+"))
		}
		if projected.Country != nil {
			err = goa.MergeErrors(err, goa.ValidatePattern("projected.country", *projected.Country, "(?i)[a-z '\\.]+"))
		}
		if projected.URL != nil {
			err = goa.MergeErrors(err, goa.ValidatePattern("projected.url", *projected.URL, "(?i)^(https?|ftp)://[^\\s/$.?#].[^\\s]*$"))
		}
	}
	return
}

// Validate runs the validations defined on Component.
func (result *Component) Validate() (err error) {
	if result.Varietal != nil {
		err = goa.MergeErrors(err, goa.ValidatePattern("result.varietal", *result.Varietal, "[A-Za-z' ]+"))
	}
	if result.Varietal != nil {
		if utf8.RuneCountInString(*result.Varietal) > 100 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("result.varietal", *result.Varietal, utf8.RuneCountInString(*result.Varietal), 100, false))
		}
	}
	if result.Percentage != nil {
		if *result.Percentage < 1 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("result.percentage", *result.Percentage, 1, true))
		}
	}
	if result.Percentage != nil {
		if *result.Percentage > 100 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("result.percentage", *result.Percentage, 100, false))
		}
	}
	return
}
