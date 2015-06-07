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
func (c *ListBottleContext) OK(bottles *BottleResourceCollection) error {
	if err := bottles.Validate(); err != nil {
		return err
	}
	js, err := json.Marshal(bottles)
	if err != nil {
		return fmt.Errorf("failed to serialize response body: %s", err)
	}
	c.Context.Respond(200, string(js))
	return nil
}

// ShowBottleContext provides the bottles show action context
type ShowBottleContext struct {
	*goa.Context
}

// AccountID returns the AccountID request path parameter
func (c *ShowBottleContext) AccountID() int {
	return c.Context.IntPathParam("account")
}

// id returns the id request path parameter
func (c *ShowBottleContext) ID() int {
	return c.Context.IntPathParam("id")
}

// OK builds a HTTP response with status code 200.
func (c *ListBottleContext) OK(bottle *BottleResource) error {
	if err := bottle.Validate(); err != nil {
		return err
	}
	js, err := json.Marshal(bottle)
	if err != nil {
		return fmt.Errorf("failed to serialize response body: %s", err)
	}
	c.Context.Respond(200, string(js))
	return nil
}

// CreateBottleContext provides the bottles create action context
type CreateBottleContext struct {
	*goa.Context
}

// AccountID returns the AccountID request path parameter
func (c *CreateBottleContext) AccountID() int {
	return c.Context.IntPathParam("account")
}

// Payload returns the request payload
func (c *CreateBottleContext) Payload() (*CreateBottlePayload, error) {
	decoder := json.NewDecoder(c.Context.R.Body)
	var p CreateBottlePayload
	err := decoder.Decode(&p)
	if err != nil {
		return nil, err
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return &p, nil
}

// CreateBottlePayload provides the bottles create action payload
type CreateBottlePayload struct {
	Name            string    `json:"name"`
	Vintage         string    `json:"vintage"`
	Vineyard        string    `json:"vineyard"`
	Varietal        *string   `json:"vintage,omitempty"`
	Color           *string   `json:"color,omitempty"`
	Sweet           *bool     `json:"sweet,omitempty"`
	Country         *string   `json:"country,omitempty"`
	Region          *string   `json:"region,omitempty"`
	Review          *string   `json:"review,omitempty"`
	Characteristics *[]string `json:"characteristics,omitempty"`
}

// Validate applies the payload validation rules and returns an error in case of failure.
func (p *CreateBottlePayload) Validate() error {
}

// UpdateBottleContext provides the bottles update action context
type UpdateBottleContext struct {
	*goa.Context
}

// AccountID returns the AccountID request path parameter
func (c *UpdateBottleContext) AccountID() int {
	return c.Context.IntPathParam("account")
}

// id returns the id request path parameter
func (c *UpdateBottleContext) ID() int {
	return c.Context.IntPathParam("id")
}

// Payload returns the request payload
func (c *UpdateBottleContext) Payload() *UpdateBottleContext {
	var p UpdateBottlePayload
	c.Context.Bind(&p)
	return &p
}

// UpdateBottlePayload provides the bottles update action payload
type UpdateBottlePayload struct {
	Name            *string   `json:"name,omitempty"`
	Vintage         *string   `json:"vintage,omitempty"`
	Vineyard        *string   `json:"vineyard,omitempty"`
	Varietal        *string   `json:"vintage,omitempty"`
	Color           *string   `json:"color,omitempty"`
	Sweet           *bool     `json:"sweet,omitempty"`
	Country         *string   `json:"country,omitempty"`
	Region          *string   `json:"region,omitempty"`
	Review          *string   `json:"review,omitempty"`
	Characteristics *[]string `json:"characteristics,omitempty"`
}

// DeleteBottleContext provides the bottles delete action context
type DeleteBottleContext struct {
	*goa.Context
}

// AccountID returns the AccountID request path parameter
func (c *DeleteBottleContext) AccountID() int {
	return c.Context.IntPathParam("account")
}

// id returns the id request path parameter
func (c *DeleteBottleContext) ID() int {
	return c.Context.IntPathParam("id")
}
