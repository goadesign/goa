package goa

// Application defines a goa application.
type Application struct {
	Name        string
	Controllers map[string]*Controller
}

// New instantiates a new goa application with the given name.
func New(name string) *Application {
	return &Application{
		Name:        name,
		Controllers: make(map[string]*Controller),
	}
}

// NewController adds a controller for the resource with given name to the application.
func (a *Application) NewController(name string) *Controller {
	if _, ok := a.Controllers[name]; ok {
		fatalf("multiple controllers for %s.", name)
	}
	c := Controller{
		ResourceName: name,
		actions:      make(map[string]interface{}),
	}
	a.Controllers[name] = &c
	return &c
}
