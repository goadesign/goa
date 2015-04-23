package design

import (
	"fmt"
	"sort"
)

// A goa application exposes one or more resources through a REST interface
type Application struct {
	Name          string
	Description   string
	Resources     map[string]*Resource // Resources exposed by application indexed by name
	ResourceNames []string             // Resource names ordered alphabetically
}

// Create new goa application
func NewApplication(name, desc, version string) *Application {
	app := Application{
		Name:        name,
		Description: desc,
		Resources:   make(map[string]*Resource),
	}
	return &app
}

// NewResource creates a new resource from the given name, base path, description and media type.
func (a *Application) NewResource(name, path, desc string, mtype *MediaType) *Resource {
	if _, ok := a.Resources[name]; ok {
		panic(fmt.Sprintf("Resource %s already defined", name))
	}
	r := Resource{
		Name:        name,
		BasePath:    path,
		Description: desc,
		MediaType:   mtype,
		Actions:     make(map[string]*Action),
	}
	a.ResourceNames = append(a.ResourceNames, name)
	sort.Strings(a.ResourceNames)
	a.Resources[name] = &r
	return &r
}
