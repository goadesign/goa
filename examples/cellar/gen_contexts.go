package main

import (
	"encoding/json"

	"github.com/raphael/goa"
)

// ListBottleContext provides the bottles list action context
type ListBottleContext struct {
	*goa.Context
}

// accountID returns the accountID request path parameter
func (c *ListBottleContext) accountID() int {
	return c.Context.IntPathParam("accountID")
}

// year returns the year request query parameter
// hasYear() must return true otherwise this method panics
func (c *ListBottleContext) year() int {
	return c.Context.IntQueryParam("year")
}

// hasYear() returns true if the "year" query string is defined
func (c *ListBottleContext) hasYear() bool {
	return c.Context.HasQuery("year")
}

// ShowBottleContext provides the bottles show action context
type ShowBottleContext struct {
	*goa.Context
}

// accountID returns the accountID request path parameter
func (c *ShowBottleContext) accountID() int {
	return c.Context.IntPathParam("accountID")
}

// id returns the id request path parameter
func (c *ShowBottleContext) id() int {
	return c.Context.IntPathParam("id")
}

// CreateBottleContext provides the bottles create action context
type CreateBottleContext struct {
	*goa.Context
}

// accountID returns the accountID request path parameter
func (c *CreateBottleContext) accountID() int {
	return c.Context.IntPathParam("accountID")
}

// Payload returns the request payload
func (c *CreateBottleContext) Payload() *CreateBottlePayload {
	decoder := json.NewDecoder(c.Context.R.Body)
	var p CreateBottlePayload
	err := decoder.Decode(&p)
	if err != nil {
		return nil, err
	}
	if err := validateRequired("create Bottle payload", "Name", p.Name, ""); err != nil {
		return nil, err
	}
	if err := validateRequired("create Bottle payload", "Vintage", p.Vintage, ""); err != nil {
		return nil, err
	}
	if err := validateRequired("create Bottle payload", "Vineyard", p.Vineyard, ""); err != nil {
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

// UpdateBottleContext provides the bottles update action context
type UpdateBottleContext struct {
	*goa.Context
}

// accountID returns the accountID request path parameter
func (c *UpdateBottleContext) accountID() int {
	return c.Context.IntPathParam("accountID")
}

// id returns the id request path parameter
func (c *UpdateBottleContext) id() int {
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

// accountID returns the accountID request path parameter
func (c *DeleteBottleContext) accountID() int {
	return c.Context.IntPathParam("accountID")
}

// id returns the id request path parameter
func (c *DeleteBottleContext) id() int {
	return c.Context.IntPathParam("id")
}
