// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// storage service
//
// Command:
// $ goa gen goa.design/goa/examples/cellar/design -o
// $(GOPATH)/src/goa.design/goa/examples/cellar

package storage

import (
	"context"
)

// The storage service makes it possible to view, add or remove wine bottles.
type Service interface {
	// List all stored bottles
	List(context.Context) (StoredBottleCollection, error)
	// Show bottle by ID
	Show(context.Context, *ShowPayload) (*StoredBottle, error)
	// Add new bottle and return its ID.
	Add(context.Context, *Bottle) (string, error)
	// Remove bottle from storage
	Remove(context.Context, *RemovePayload) error
	// Rate bottles by IDs
	Rate(context.Context, map[uint32][]string) error
	// Add n number of bottles and return their IDs. This is a multipart request
	// and each part has field name 'bottle' and contains the encoded bottle info
	// to be added.
	MultiAdd(context.Context, []*Bottle) ([]string, error)
	// Update bottles with the given IDs. This is a multipart request and each part
	// has field name 'bottle' and contains the encoded bottle info to be updated.
	// The IDs in the query parameter is mapped to each part in the request.
	MultiUpdate(context.Context, *MultiUpdatePayload) error
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "storage"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [7]string{"list", "show", "add", "remove", "rate", "multi_add", "multi_update"}

// StoredBottleCollection is the result type of the storage service list method.
type StoredBottleCollection []*StoredBottle

// ShowPayload is the payload type of the storage service show method.
type ShowPayload struct {
	// ID of bottle to show
	ID string
	// View to render
	View *string
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

// RemovePayload is the payload type of the storage service remove method.
type RemovePayload struct {
	// ID of bottle to remove
	ID string
}

// MultiUpdatePayload is the payload type of the storage service multi_update
// method.
type MultiUpdatePayload struct {
	// IDs of the bottles to be updated
	Ids []string
	// Array of bottle info that matches the ids attribute
	Bottles []*Bottle
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

// Error returns an error description.
func (e *NotFound) Error() string {
	return "NotFound is the type returned when attempting to show or delete a bottle that does not exist."
}

// ErrorName returns q%!(EXTRA string=NotFound).
func (e *NotFound) ErrorName() string {
	return "NotFound"
}
