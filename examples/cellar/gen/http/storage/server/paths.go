// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// HTTP request path constructors for the storage service.
//
// Command:
// $ goa gen goa.design/goa.v2/examples/cellar/design

package server

import (
	"fmt"
)

// ListStoragePath returns the URL path to the storage service list HTTP endpoint.
func ListStoragePath() string {
	return "/storage"
}

// ShowStoragePath returns the URL path to the storage service show HTTP endpoint.
func ShowStoragePath(id string) string {
	return fmt.Sprintf("/storage/%v", id)
}

// AddStoragePath returns the URL path to the storage service add HTTP endpoint.
func AddStoragePath() string {
	return "/storage"
}

// RemoveStoragePath returns the URL path to the storage service remove HTTP endpoint.
func RemoveStoragePath(id string) string {
	return fmt.Sprintf("/storage/%v", id)
}
