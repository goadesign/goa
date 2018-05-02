// Code generated by goa v2.0.0-wip, DO NOT EDIT.
//
// storage views
//
// Command:
// $ goa gen goa.design/goa/examples/cellar/design -o
// $(GOPATH)/src/goa.design/goa/examples/cellar

package views

import (
	"unicode/utf8"

	goa "goa.design/goa"
)

// StoredBottleView is a type which is projected based on a view.
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

// StoredBottle is the viewed result type that projects StoredBottleView based
// on a view.
type StoredBottle struct {
	*StoredBottleView
	// View to render
	View string
}

// WineryView is a type which is projected based on a view.
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

// Winery is the viewed result type that projects WineryView based on a view.
type Winery struct {
	*WineryView
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

// AsDefault projects viewed result type StoredBottle using the default view.
func (result *StoredBottle) AsDefault() *StoredBottle {
	t := &StoredBottleView{
		ID:          result.ID,
		Name:        result.Name,
		Vintage:     result.Vintage,
		Description: result.Description,
		Rating:      result.Rating,
	}
	if result.Composition != nil {
		t.Composition = make([]*Component, len(result.Composition))
		for j, val := range result.Composition {
			t.Composition[j] = &Component{
				Varietal:   val.Varietal,
				Percentage: val.Percentage,
			}
		}
	}
	if result.Winery != nil {
		t.Winery = result.Winery.AsTiny()
	}

	return &StoredBottle{
		StoredBottleView: t,
		View:             "default",
	}
}

// AsTiny projects viewed result type StoredBottle using the tiny view.
func (result *StoredBottle) AsTiny() *StoredBottle {
	t := &StoredBottleView{
		ID:   result.ID,
		Name: result.Name,
	}
	if result.Winery != nil {
		t.Winery = result.Winery.AsTiny()
	}

	return &StoredBottle{
		StoredBottleView: t,
		View:             "tiny",
	}
}

// AsDefault projects viewed result type Winery using the default view.
func (result *Winery) AsDefault() *Winery {
	t := &WineryView{
		Name:    result.Name,
		Region:  result.Region,
		Country: result.Country,
		URL:     result.URL,
	}
	return &Winery{
		WineryView: t,
		View:       "default",
	}
}

// AsTiny projects viewed result type Winery using the tiny view.
func (result *Winery) AsTiny() *Winery {
	t := &WineryView{
		Name: result.Name,
	}
	return &Winery{
		WineryView: t,
		View:       "tiny",
	}
}

// Validate runs the validations defined on StoredBottle.
func (result *StoredBottle) Validate() (err error) {
	switch result.View {
	case "default":
		if result.ID == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("id", "result"))
		}
		if result.Name == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("name", "result"))
		}
		if result.Winery == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("winery", "result"))
		}
		if result.Vintage == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("vintage", "result"))
		}
		if result.Name != nil {
			if utf8.RuneCountInString(*result.Name) > 100 {
				err = goa.MergeErrors(err, goa.InvalidLengthError("result.name", *result.Name, utf8.RuneCountInString(*result.Name), 100, false))
			}
		}
		if result.Winery != nil {
			if err2 := result.Winery.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
		if result.Vintage != nil {
			if *result.Vintage < 1900 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("result.vintage", *result.Vintage, 1900, true))
			}
		}
		if result.Vintage != nil {
			if *result.Vintage > 2020 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("result.vintage", *result.Vintage, 2020, false))
			}
		}
		for _, e := range result.Composition {
			if e != nil {
				if err2 := e.Validate(); err2 != nil {
					err = goa.MergeErrors(err, err2)
				}
			}
		}
		if result.Description != nil {
			if utf8.RuneCountInString(*result.Description) > 2000 {
				err = goa.MergeErrors(err, goa.InvalidLengthError("result.description", *result.Description, utf8.RuneCountInString(*result.Description), 2000, false))
			}
		}
		if result.Rating != nil {
			if *result.Rating < 1 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("result.rating", *result.Rating, 1, true))
			}
		}
		if result.Rating != nil {
			if *result.Rating > 5 {
				err = goa.MergeErrors(err, goa.InvalidRangeError("result.rating", *result.Rating, 5, false))
			}
		}
	case "tiny":
		if result.ID == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("id", "result"))
		}
		if result.Name == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("name", "result"))
		}
		if result.Winery == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("winery", "result"))
		}
		if result.Name != nil {
			if utf8.RuneCountInString(*result.Name) > 100 {
				err = goa.MergeErrors(err, goa.InvalidLengthError("result.name", *result.Name, utf8.RuneCountInString(*result.Name), 100, false))
			}
		}
		if result.Winery != nil {
			if err2 := result.Winery.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
	}
	return
}

// Validate runs the validations defined on Winery.
func (result *Winery) Validate() (err error) {
	switch result.View {
	case "default":
		if result.Name == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("name", "result"))
		}
		if result.Region == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("region", "result"))
		}
		if result.Country == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("country", "result"))
		}
		if result.Region != nil {
			err = goa.MergeErrors(err, goa.ValidatePattern("result.region", *result.Region, "(?i)[a-z '\\.]+"))
		}
		if result.Country != nil {
			err = goa.MergeErrors(err, goa.ValidatePattern("result.country", *result.Country, "(?i)[a-z '\\.]+"))
		}
		if result.URL != nil {
			err = goa.MergeErrors(err, goa.ValidatePattern("result.url", *result.URL, "(?i)^(https?|ftp)://[^\\s/$.?#].[^\\s]*$"))
		}
	case "tiny":
		if result.Name == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("name", "result"))
		}
	}
	return
}

// Validate runs the validations defined on Component.
func (result *Component) Validate() (err error) {
	if result.Varietal == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("varietal", "result"))
	}
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
