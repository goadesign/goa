// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// storage service
//
// Command:
// $ goa gen goa.design/goa.v2/examples/cellar/design

package storage

import "context"

// The storage service makes it possible to add new or to remove existing wine
// bottles.
type Service interface {
	// Add new bottle and return its ID.
	Add(context.Context, *Bottle) (string, error)
	// List all stored bottles
	List(context.Context) (StoredBottleCollection, error)
	// Show bottle by ID
	Show(context.Context, *ShowPayload) (*StoredBottle, error)
	// Remove bottle from storage
	Remove(context.Context, *RemovePayload) error
}

// Bottle is the payload type of the storage service add method.
type Bottle struct {
	// Name of bottle
	Name string
	// Winery that produces wine
	Winery *Winery
	// Vintage of bottle
	Vintage uint32
	// Composition is the list of grape varietals and associated percentage.
	Composition []*Component
	// Description of bottle
	Description *string
	// Rating of bottle from 1 (worst) to 5 (best)
	Rating *uint32
}

// StoredBottleCollection is the result type of the storage service list method.
type StoredBottleCollection []*StoredBottle

// ShowPayload is the payload type of the storage service show method.
type ShowPayload struct {
	// ID of bottle to show
	ID string
}

// StoredBottle is the result type of the storage service show method.
type StoredBottle struct {
	// ID is the unique id of the bottle.
	ID string
	// Name of bottle
	Name string
	// Winery that produces wine
	Winery *Winery
	// Vintage of bottle
	Vintage uint32
	// Composition is the list of grape varietals and associated percentage.
	Composition []*Component
	// Description of bottle
	Description *string
	// Rating of bottle from 1 (worst) to 5 (best)
	Rating *uint32
}

// RemovePayload is the payload type of the storage service remove method.
type RemovePayload struct {
	// ID of bottle to remove
	ID string
}

type Winery struct {
	// Name of winery
	Name string
	// Region of winery
	Region string
	// Country of winery
	Country string
	// Winery website URL
	URL *string
}

type Component struct {
	// Grape varietal
	Varietal string
	// Percentage of varietal in wine
	Percentage *uint32
}

// NotFound is the type returned when attempting to show or delete a bottle
// that does not exist.
type NotFound struct {
	// Message of error
	Message string
	// ID of missing bottle
	ID string
}

// Error returns "NotFound".
func (e *NotFound) Error() string {
	return "NotFound"
}
