// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// storage HTTP client CLI support package
//
// Command:
// $ goa gen goa.design/goa.v2/examples/cellar/design

package client

import (
	"encoding/json"
	"fmt"
	"unicode/utf8"

	goa "goa.design/goa.v2"
	"goa.design/goa.v2/examples/cellar/gen/storage"
)

// BuildShowPayload builds the payload for the storage show endpoint from CLI
// flags.
func BuildShowPayload(storageShowID string) (*storage.ShowPayload, error) {
	var id string
	{
		id = storageShowID
	}
	payload := &storage.ShowPayload{
		ID: id,
	}
	return payload, nil
}

// BuildBottle builds the payload for the storage add endpoint from CLI flags.
func BuildBottle(storageAddBody string) (*storage.Bottle, error) {
	var body AddRequestBody
	{
		err := json.Unmarshal([]byte(storageAddBody), &body)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON for body, example of valid JSON:\n%s", "{\"composition\":[{\"percentage\":67,\"varietal\":\"Syrah\"},{\"percentage\":67,\"varietal\":\"Syrah\"},{\"percentage\":67,\"varietal\":\"Syrah\"}],\"description\":\"Red wine blend with an emphasis on the Cabernet Franc grape and including other Bordeaux grape varietals and some Syrah\",\"name\":\"Blue's Cuvee\",\"rating\":3,\"vintage\":1905,\"winery\":{\"country\":\"USA\",\"name\":\"Longoria\",\"region\":\"Central Coast, California\",\"url\":\"http://www.longoriawine.com/\"}}")
		}
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
		if err != nil {
			return nil, err
		}
	}
	v := &storage.Bottle{
		Name:        body.Name,
		Vintage:     body.Vintage,
		Description: body.Description,
		Rating:      body.Rating,
	}
	v.Winery = wineryRequestBodyToWinery(body.Winery)
	if body.Composition != nil {
		v.Composition = make([]*storage.Component, len(body.Composition))
		for i, val := range body.Composition {
			v.Composition[i] = &storage.Component{
				Varietal:   val.Varietal,
				Percentage: val.Percentage,
			}
		}
	}

	return v, nil
}

// BuildRemovePayload builds the payload for the storage remove endpoint from
// CLI flags.
func BuildRemovePayload(storageRemoveID string) (*storage.RemovePayload, error) {
	var id string
	{
		id = storageRemoveID
	}
	payload := &storage.RemovePayload{
		ID: id,
	}
	return payload, nil
}
