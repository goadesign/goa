// Code generated by goa v2.0.0-wip, DO NOT EDIT.
//
// storage service
//
// Command:
// $ goa gen goa.design/goa/examples/cellar/design -o
// $(GOPATH)/src/goa.design/goa/examples/cellar

package storage

import (
	"context"

	storageviews "goa.design/goa/examples/cellar/gen/storage/views"
)

// The storage service makes it possible to view, add or remove wine bottles.
type Service interface {
	// List all stored bottles
	List(context.Context) (res StoredBottleCollection, err error)
	// Show bottle by ID
	// The "view" return value must have one of the following views
	// * "default"
	// * "tiny"
	Show(context.Context, *ShowPayload) (res *StoredBottle, view string, err error)
	// Add new bottle and return its ID.
	Add(context.Context, *Bottle) (res string, err error)
	// Remove bottle from storage
	Remove(context.Context, *RemovePayload) (err error)
	// Rate bottles by IDs
	Rate(context.Context, map[uint32][]string) (err error)
	// Add n number of bottles and return their IDs. This is a multipart request
	// and each part has field name 'bottle' and contains the encoded bottle info
	// to be added.
	MultiAdd(context.Context, []*Bottle) (res []string, err error)
	// Update bottles with the given IDs. This is a multipart request and each part
	// has field name 'bottle' and contains the encoded bottle info to be updated.
	// The IDs in the query parameter is mapped to each part in the request.
	MultiUpdate(context.Context, *MultiUpdatePayload) (err error)
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

// ErrorName returns "NotFound".
func (e *NotFound) ErrorName() string {
	return e.Message
}

// NewStoredBottleCollection initializes result type StoredBottleCollection
// from viewed result type StoredBottleCollection.
func NewStoredBottleCollection(vres storageviews.StoredBottleCollection) StoredBottleCollection {
	var res StoredBottleCollection
	switch vres.View {
	case "default", "":
		res = newStoredBottleCollection(vres.Projected)
	case "tiny":
		res = newStoredBottleCollectionTiny(vres.Projected)
	}
	return res
}

// NewViewedStoredBottleCollection initializes viewed result type
// StoredBottleCollection from result type StoredBottleCollection using the
// given view.
func NewViewedStoredBottleCollection(res StoredBottleCollection, view string) storageviews.StoredBottleCollection {
	var vres storageviews.StoredBottleCollection
	switch view {
	case "default", "":
		p := newStoredBottleCollectionView(res)
		vres = storageviews.StoredBottleCollection{p, "default"}
	case "tiny":
		p := newStoredBottleCollectionViewTiny(res)
		vres = storageviews.StoredBottleCollection{p, "tiny"}
	}
	return vres
}

// NewStoredBottle initializes result type StoredBottle from viewed result type
// StoredBottle.
func NewStoredBottle(vres *storageviews.StoredBottle) *StoredBottle {
	var res *StoredBottle
	switch vres.View {
	case "default", "":
		res = newStoredBottle(vres.Projected)
	case "tiny":
		res = newStoredBottleTiny(vres.Projected)
	}
	return res
}

// NewViewedStoredBottle initializes viewed result type StoredBottle from
// result type StoredBottle using the given view.
func NewViewedStoredBottle(res *StoredBottle, view string) *storageviews.StoredBottle {
	var vres *storageviews.StoredBottle
	switch view {
	case "default", "":
		p := newStoredBottleView(res)
		vres = &storageviews.StoredBottle{p, "default"}
	case "tiny":
		p := newStoredBottleViewTiny(res)
		vres = &storageviews.StoredBottle{p, "tiny"}
	}
	return vres
}

// newStoredBottleCollection converts projected type StoredBottleCollection to
// service type StoredBottleCollection.
func newStoredBottleCollection(vres storageviews.StoredBottleCollectionView) StoredBottleCollection {
	res := make(StoredBottleCollection, len(vres))
	for i, n := range vres {
		res[i] = newStoredBottle(n)
	}
	return res
}

// newStoredBottleCollectionTiny converts projected type StoredBottleCollection
// to service type StoredBottleCollection.
func newStoredBottleCollectionTiny(vres storageviews.StoredBottleCollectionView) StoredBottleCollection {
	res := make(StoredBottleCollection, len(vres))
	for i, n := range vres {
		res[i] = newStoredBottleTiny(n)
	}
	return res
}

// newStoredBottleCollectionView projects result type StoredBottleCollection
// into projected type StoredBottleCollectionView using the "default" view.
func newStoredBottleCollectionView(res StoredBottleCollection) storageviews.StoredBottleCollectionView {
	vres := make(storageviews.StoredBottleCollectionView, len(res))
	for i, n := range res {
		vres[i] = newStoredBottleView(n)
	}
	return vres
}

// newStoredBottleCollectionViewTiny projects result type
// StoredBottleCollection into projected type StoredBottleCollectionView using
// the "tiny" view.
func newStoredBottleCollectionViewTiny(res StoredBottleCollection) storageviews.StoredBottleCollectionView {
	vres := make(storageviews.StoredBottleCollectionView, len(res))
	for i, n := range res {
		vres[i] = newStoredBottleViewTiny(n)
	}
	return vres
}

// newStoredBottle converts projected type StoredBottle to service type
// StoredBottle.
func newStoredBottle(vres *storageviews.StoredBottleView) *StoredBottle {
	res := &StoredBottle{
		Description: vres.Description,
		Rating:      vres.Rating,
	}
	if vres.ID != nil {
		res.ID = *vres.ID
	}
	if vres.Name != nil {
		res.Name = *vres.Name
	}
	if vres.Vintage != nil {
		res.Vintage = *vres.Vintage
	}
	if vres.Composition != nil {
		res.Composition = make([]*Component, len(vres.Composition))
		for i, val := range vres.Composition {
			res.Composition[i] = &Component{
				Varietal:   *val.Varietal,
				Percentage: val.Percentage,
			}
		}
	}
	if vres.Winery != nil {
		res.Winery = newWineryTiny(vres.Winery)
	}
	return res
}

// newStoredBottleTiny converts projected type StoredBottle to service type
// StoredBottle.
func newStoredBottleTiny(vres *storageviews.StoredBottleView) *StoredBottle {
	res := &StoredBottle{}
	if vres.ID != nil {
		res.ID = *vres.ID
	}
	if vres.Name != nil {
		res.Name = *vres.Name
	}
	if vres.Winery != nil {
		res.Winery = newWineryTiny(vres.Winery)
	}
	return res
}

// newStoredBottleView projects result type StoredBottle into projected type
// StoredBottleView using the "default" view.
func newStoredBottleView(res *StoredBottle) *storageviews.StoredBottleView {
	vres := &storageviews.StoredBottleView{
		ID:          &res.ID,
		Name:        &res.Name,
		Vintage:     &res.Vintage,
		Description: res.Description,
		Rating:      res.Rating,
	}
	if res.Composition != nil {
		vres.Composition = make([]*storageviews.ComponentView, len(res.Composition))
		for i, val := range res.Composition {
			vres.Composition[i] = &storageviews.ComponentView{
				Varietal:   &val.Varietal,
				Percentage: val.Percentage,
			}
		}
	}
	if res.Winery != nil {
		vres.Winery = newWineryViewTiny(res.Winery)
	}
	return vres
}

// newStoredBottleViewTiny projects result type StoredBottle into projected
// type StoredBottleView using the "tiny" view.
func newStoredBottleViewTiny(res *StoredBottle) *storageviews.StoredBottleView {
	vres := &storageviews.StoredBottleView{
		ID:   &res.ID,
		Name: &res.Name,
	}
	if res.Winery != nil {
		vres.Winery = newWineryViewTiny(res.Winery)
	}
	return vres
}

// newWinery converts projected type Winery to service type Winery.
func newWinery(vres *storageviews.WineryView) *Winery {
	res := &Winery{
		URL: vres.URL,
	}
	if vres.Name != nil {
		res.Name = *vres.Name
	}
	if vres.Region != nil {
		res.Region = *vres.Region
	}
	if vres.Country != nil {
		res.Country = *vres.Country
	}
	return res
}

// newWineryTiny converts projected type Winery to service type Winery.
func newWineryTiny(vres *storageviews.WineryView) *Winery {
	res := &Winery{}
	if vres.Name != nil {
		res.Name = *vres.Name
	}
	return res
}

// newWineryView projects result type Winery into projected type WineryView
// using the "default" view.
func newWineryView(res *Winery) *storageviews.WineryView {
	vres := &storageviews.WineryView{
		Name:    &res.Name,
		Region:  &res.Region,
		Country: &res.Country,
		URL:     res.URL,
	}
	return vres
}

// newWineryViewTiny projects result type Winery into projected type WineryView
// using the "tiny" view.
func newWineryViewTiny(res *Winery) *storageviews.WineryView {
	vres := &storageviews.WineryView{
		Name: &res.Name,
	}
	return vres
}
