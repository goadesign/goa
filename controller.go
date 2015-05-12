package goa

// Controllers implement a resource actions.
type Controller struct {
	ResourceName string
	actions      map[string]interface{}
}

// Action defines an action handler.
func (c *Controller) Action(name string, handler interface{}) {
	if _, ok := c.actions[name]; ok {
		fatalf("multiple handlers for %s of %s.", name, c.ResourceName)
	}
	c.actions[name] = handler
}
