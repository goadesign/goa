package design

import (
	"errors"

	"github.com/goadesign/goa/dslengine"
)

// FindAPIDefinition iterates over the RootDefinitions to find the registered
// APIDefinition.
// It will return an error if there's more than one APIDefinition registered or
// no APIDefinition could be found.
func FindAPIDefinition(roots dslengine.RootDefinitions) (*APIDefinition, error) {
	var api *APIDefinition
	err := roots.IterateRoots(func(root dslengine.Root) error {
		if def, ok := root.(*APIDefinition); ok && api == nil {
			api = def
		} else if ok {
			return errors.New("encountered more than one APIDefinition registered")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if api == nil {
		return nil, errors.New("no APIDefinition registered in the API design")
	}
	return api, nil
}
