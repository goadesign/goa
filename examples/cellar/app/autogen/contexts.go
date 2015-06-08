package autogen

import (
	"encoding/json"
	"fmt"

	"github.com/raphael/goa"
)

// ListBottleContext provides the bottles list action context
type ListBottleContext struct {
	*goa.Context
}

// AccountID returns the AccountID request path parameter
func (c *ListBottleContext) AccountID() int {
	return c.Context.IntParam("account")
}

// Years returns the year request query parameter
// HasYears() must return tru otherwise this method panics
func (c *ListBottleContext) Years() []int {
	return c.Context.IntSliceParam("years")
}

// HasYears() returns true if the "year" query string is defined
func (c *ListBottleContext) HasYears() bool {
	return c.Context.HasParam("years")
}

// OK builds a HTTP response with status code 200.
func (c *ListBottleContext) OK(bottles []*BottleResource) error {
	js, err := json.Marshal(bottles)
	if err != nil {
		return fmt.Errorf("failed to serialize response body: %s", err)
	}
	return c.Context.Respond(200, js)
}

// ShowBottleContext provides the bottles show action context
type ShowBottleContext struct {
	*goa.Context
}

// AccountID returns the AccountID request path parameter
func (c *ShowBottleContext) AccountID() int {
	return c.Context.IntParam("account")
}

// id returns the id request path parameter
func (c *ShowBottleContext) ID() int {
	return c.Context.IntParam("id")
}

// OK builds a HTTP response with status code 200.
func (c *ShowBottleContext) OK(bottle *BottleResource) error {
	if err := bottle.Validate(); err != nil {
		return err
	}
	js, err := json.Marshal(bottle)
	if err != nil {
		return fmt.Errorf("failed to serialize response body: %s", err)
	}
	return c.Context.Respond(200, js)
}

// NotFound builds a HTTP response with status code 404.
func (c *ShowBottleContext) NotFound() error {
	return c.Context.Respond(404, nil)
}

// CreateBottleContext provides the bottles create action context
type CreateBottleContext struct {
	*goa.Context
}

// AccountID returns the AccountID request path parameter
func (c *CreateBottleContext) AccountID() int {
	return c.Context.IntParam("account")
}

// Payload returns the request payload
func (c *CreateBottleContext) Payload() (*CreateBottlePayload, error) {
	var p CreateBottlePayload
	if err := c.Context.Bind(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

// CreateBottlePayload provides the bottles create action payload
type CreateBottlePayload struct {
	Name     string  `json:"name"`
	Vintage  string  `json:"vintage"`
	Vineyard string  `json:"vineyard"`
	Varietal *string `json:"vintage,omitempty"`
	Color    *string `json:"color,omitempty"`
	Sweet    *bool   `json:"sweet,omitempty"`
	Country  *string `json:"country,omitempty"`
	Region   *string `json:"region,omitempty"`
	Review   *string `json:"review,omitempty"`
}

// Validate applies the payload validation rules and returns an error in case of failure.
func (p *CreateBottlePayload) Validate() error {
	return nil
}

// UpdateBottleContext provides the bottles update action context
type UpdateBottleContext struct {
	*goa.Context
}

// AccountID returns the AccountID request path parameter
func (c *UpdateBottleContext) AccountID() int {
	return c.Context.IntParam("account")
}

// id returns the id request path parameter
func (c *UpdateBottleContext) ID() int {
	return c.Context.IntParam("id")
}

// Payload returns the request payload
func (c *UpdateBottleContext) Payload() (*UpdateBottlePayload, error) {
	var p UpdateBottlePayload
	if err := c.Context.Bind(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

// UpdateBottlePayload provides the bottles update action payload
type UpdateBottlePayload struct {
	Name     string  `json:"name"`
	Vintage  string  `json:"vintage"`
	Vineyard string  `json:"vineyard"`
	Varietal *string `json:"vintage"`
	Color    *string `json:"color"`
	Sweet    *bool   `json:"sweet"`
	Country  *string `json:"country"`
	Region   *string `json:"region"`
	Review   *string `json:"review"`
}

// Validate implements the validation rules specified by the payload design definition.
func (p *UpdateBottlePayload) Validate() error {
	return nil
}

// DeleteBottleContext provides the bottles delete action context
type DeleteBottleContext struct {
	*goa.Context
}

// AccountID returns the AccountID request path parameter
func (c *DeleteBottleContext) AccountID() int {
	return c.Context.IntParam("account")
}

// id returns the id request path parameter
func (c *DeleteBottleContext) ID() int {
	return c.Context.IntParam("id")
}

// RateBottleContext provides the bottles rate action context
type RateBottleContext struct {
	*goa.Context
}

// AccountID returns the AccountID request path parameter
func (c RateBottleContext) AccountID() int {
	return c.Context.IntParam("account")
}

// id returns the id request path parameter
func (c *RateBottleContext) ID() int {
	return c.Context.IntParam("id")
}
