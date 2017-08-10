// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// sommelier HTTP server types
//
// Command:
// $ goa gen goa.design/goa.v2/examples/cellar/design

package server

import (
	"unicode/utf8"

	goa "goa.design/goa.v2"
	"goa.design/goa.v2/examples/cellar/gen/sommelier"
)

// PickRequestBody is the type of the sommelier pick HTTP endpoint request body.
type PickRequestBody struct {
	// Name of bottle to pick
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// Varietals in preference order
	Varietal []string `form:"varietal,omitempty" json:"varietal,omitempty" xml:"varietal,omitempty"`
	// Winery of bottle to pick
	Winery *string `form:"winery,omitempty" json:"winery,omitempty" xml:"winery,omitempty"`
}

// PickResponseBody is the type of the sommelier pick HTTP endpoint response
// body.
type PickResponseBody []*StoredBottleResponseBody

// PickNoCriteriaResponseBody is the type of the sommelier "pick" HTTP endpoint
// no_criteria error response body.
type PickNoCriteriaResponseBody struct {
	// Missing criteria
	Value string `form:"value" json:"value" xml:"value"`
}

// PickNoMatchResponseBody is the type of the sommelier "pick" HTTP endpoint
// no_match error response body.
type PickNoMatchResponseBody struct {
	// No bottle matched given criteria
	Value string `form:"value" json:"value" xml:"value"`
}

// StoredBottleResponseBody is used to define fields on response body types.
type StoredBottleResponseBody struct {
	// ID is the unique id of the bottle.
	ID string `form:"id" json:"id" xml:"id"`
	// Name of bottle
	Name string `form:"name" json:"name" xml:"name"`
	// Winery that produces wine
	Winery *WineryResponseBody `form:"winery" json:"winery" xml:"winery"`
	// Vintage of bottle
	Vintage uint32 `form:"vintage" json:"vintage" xml:"vintage"`
	// Composition is the list of grape varietals and associated percentage.
	Composition []*ComponentResponseBody `form:"composition,omitempty" json:"composition,omitempty" xml:"composition,omitempty"`
	// Description of bottle
	Description *string `form:"description,omitempty" json:"description,omitempty" xml:"description,omitempty"`
	// Rating of bottle from 1 (worst) to 5 (best)
	Rating *uint32 `form:"rating,omitempty" json:"rating,omitempty" xml:"rating,omitempty"`
}

// WineryResponseBody is used to define fields on response body types.
type WineryResponseBody struct {
	// Name of winery
	Name string `form:"name" json:"name" xml:"name"`
	// Region of winery
	Region string `form:"region" json:"region" xml:"region"`
	// Country of winery
	Country string `form:"country" json:"country" xml:"country"`
	// Winery website URL
	URL *string `form:"url,omitempty" json:"url,omitempty" xml:"url,omitempty"`
}

// ComponentResponseBody is used to define fields on response body types.
type ComponentResponseBody struct {
	// Grape varietal
	Varietal string `form:"varietal" json:"varietal" xml:"varietal"`
	// Percentage of varietal in wine
	Percentage *uint32 `form:"percentage,omitempty" json:"percentage,omitempty" xml:"percentage,omitempty"`
}

// NewPickResponseBody builds the sommelier service pick endpoint response body
// from a result.
func NewPickResponseBody(res sommelier.StoredBottleCollection) PickResponseBody {
	body := make([]*StoredBottleResponseBody, len(res))
	for i, val := range res {
		body[i] = &StoredBottleResponseBody{
			ID:          val.ID,
			Name:        val.Name,
			Vintage:     val.Vintage,
			Description: val.Description,
			Rating:      val.Rating,
		}
		body[i].Winery = wineryToWineryResponseBodyNoDefault(val.Winery)
		if val.Composition != nil {
			body[i].Composition = make([]*ComponentResponseBody, len(val.Composition))
			for i, val := range val.Composition {
				body[i].Composition[i] = &ComponentResponseBody{
					Varietal:   val.Varietal,
					Percentage: val.Percentage,
				}
			}
		}
	}

	return body
}

// NewPickNoCriteriaResponseBody builds the sommelier service pick endpoint
// response body from a result.
func NewPickNoCriteriaResponseBody(res *sommelier.NoCriteria) *PickNoCriteriaResponseBody {
	body := &PickNoCriteriaResponseBody{
		Value: res.Value,
	}

	return body
}

// NewPickNoMatchResponseBody builds the sommelier service pick endpoint
// response body from a result.
func NewPickNoMatchResponseBody(res *sommelier.NoMatch) *PickNoMatchResponseBody {
	body := &PickNoMatchResponseBody{
		Value: res.Value,
	}

	return body
}

// NewPickCriteria builds a sommelier service pick endpoint payload.
func NewPickCriteria(body *PickRequestBody) *sommelier.Criteria {
	v := &sommelier.Criteria{
		Name:   body.Name,
		Winery: body.Winery,
	}
	if body.Varietal != nil {
		v.Varietal = make([]string, len(body.Varietal))
		for i, val := range body.Varietal {
			v.Varietal[i] = val
		}
	}

	return v
}

// Validate runs the validations defined on StoredBottleResponseBody
func (body *StoredBottleResponseBody) Validate() (err error) {
	if body.Winery == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("winery", "body"))
	}
	if utf8.RuneCountInString(body.Name) > 100 {
		err = goa.MergeErrors(err, goa.InvalidLengthError("body.name", body.Name, utf8.RuneCountInString(body.Name), 100, false))
	}
	if body.Winery != nil {
		if err2 := body.Winery.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if body.Vintage < 1900 {
		err = goa.MergeErrors(err, goa.InvalidRangeError("body.vintage", body.Vintage, 1900, true))
	}
	if body.Vintage > 2020 {
		err = goa.MergeErrors(err, goa.InvalidRangeError("body.vintage", body.Vintage, 2020, false))
	}
	for _, e := range body.Composition {
		if e != nil {
			if err2 := e.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
	}
	if body.Description != nil {
		if utf8.RuneCountInString(*body.Description) > 2000 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("body.description", *body.Description, utf8.RuneCountInString(*body.Description), 2000, false))
		}
	}
	if body.Rating != nil {
		if *body.Rating < 1 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("body.rating", *body.Rating, 1, true))
		}
	}
	if body.Rating != nil {
		if *body.Rating > 5 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("body.rating", *body.Rating, 5, false))
		}
	}
	return
}

// Validate runs the validations defined on WineryResponseBody
func (body *WineryResponseBody) Validate() (err error) {
	err = goa.MergeErrors(err, goa.ValidatePattern("body.region", body.Region, "(?i)[a-z '\\.]+"))
	err = goa.MergeErrors(err, goa.ValidatePattern("body.country", body.Country, "(?i)[a-z '\\.]+"))
	if body.URL != nil {
		err = goa.MergeErrors(err, goa.ValidatePattern("body.url", *body.URL, "(?i)^(https?|ftp)://[^\\s/$.?#].[^\\s]*$"))
	}
	return
}

// Validate runs the validations defined on ComponentResponseBody
func (body *ComponentResponseBody) Validate() (err error) {
	err = goa.MergeErrors(err, goa.ValidatePattern("body.varietal", body.Varietal, "[A-Za-z' ]+"))
	if utf8.RuneCountInString(body.Varietal) > 100 {
		err = goa.MergeErrors(err, goa.InvalidLengthError("body.varietal", body.Varietal, utf8.RuneCountInString(body.Varietal), 100, false))
	}
	if body.Percentage != nil {
		if *body.Percentage < 1 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("body.percentage", *body.Percentage, 1, true))
		}
	}
	if body.Percentage != nil {
		if *body.Percentage > 100 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("body.percentage", *body.Percentage, 100, false))
		}
	}
	return
}
