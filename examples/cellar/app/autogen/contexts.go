package autogen

import (
	"strconv"

	"github.com/raphael/goa"
)

// ListBottleContext provides the bottles list action context
type ListBottleContext struct {
	*goa.Context
	AccountID int
	Years     []int
	HasYears  bool
}

// NewListBottleContext parses the incoming request URL and body and instantiates the context
// accordingly. It returns an error if a required parameter is missing or if a parameter has an
// invalid value.
func NewListBottleContext(c *goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	rawAccountID, _ := c.Get("accountID")
	if accountID, err := strconv.Atoi(rawAccountID); err == nil {
		ctx.AccountID = int(accountID)
	} else {
		err = goa.InvalidParamValue("accountID", rawAccountID, "number", err)
	}
	if rawYears, ok := c.Query["years"]; ok {
		ctx.HasYears = true
		years := make([]int, len(rawYears))
		for i, rawYear := range rawYears {
			if year, err := strconv.Atoi(rawYear); err == nil {
				years[i] = year
			} else {
				err = goa.InvalidParamValue("years", rawYears, "array of numbers", err)
				break
			}
		}
	}
	return &ctx, err
}

// OK builds a HTTP response with status code 200.
func (c *ListBottleContext) OK(bottles []*Bottle) error {
	return c.JSON(200, bottles)
}

// ShowBottleContext provides the bottles show action context
type ShowBottleContext struct {
	*goa.Context
	AccountID int
	ID        int
}

// NewShowBottleContext parses the incoming request URL and body and instantiates the context
// accordingly. It returns an error if a required parameter is missing or if a parameter has an
// invalid value.
func NewShowBottleContext(c *goa.Context) (*ShowBottleContext, error) {
	var err error
	ctx := ShowBottleContext{Context: c}
	rawAccountID, _ := c.Get("accountID")
	if accountID, err := strconv.Atoi(rawAccountID); err == nil {
		ctx.AccountID = int(accountID)
	} else {
		err = goa.InvalidParamValue("accountID", rawAccountID, "number", err)
	}
	rawID, _ := c.Get("ID")
	if ID, err := strconv.Atoi(rawID); err == nil {
		ctx.ID = int(ID)
	} else {
		err = goa.InvalidParamValue("ID", rawID, "number", err)
	}
	return &ctx, err
}

// OK builds a HTTP response with status code 200.
func (c *ShowBottleContext) OK(bottle *Bottle) error {
	if err := bottle.Validate(); err != nil {
		return err
	}
	return c.JSON(200, bottle)
}

// NotFound builds a HTTP response with status code 404.
func (c *ShowBottleContext) NotFound() error {
	return c.Respond(404, nil)
}

type (
	// CreateBottleContext provides the bottles create action context
	CreateBottleContext struct {
		*goa.Context
		AccountID int
		Payload   *CreateBottlePayload
	}

	// CreateBottlePayload provides the bottles create action payload
	CreateBottlePayload struct {
		Name     string  `json:"name"`
		Vintage  int     `json:"vintage"`
		Vineyard string  `json:"vineyard"`
		Varietal *string `json:"vintage,omitempty"`
		Color    *string `json:"color,omitempty"`
		Sweet    *bool   `json:"sweet,omitempty"`
		Country  *string `json:"country,omitempty"`
		Region   *string `json:"region,omitempty"`
		Review   *string `json:"review,omitempty"`
	}
)

// NewCreateBottleContext parses the incoming request URL and body and instantiates the context
// accordingly. It returns an error if a required parameter is missing or if a parameter has an
// invalid value.
func NewCreateBottleContext(c *goa.Context) (*CreateBottleContext, error) {
	var err error
	ctx := CreateBottleContext{Context: c}
	rawAccountID, _ := c.Get("accountID")
	if accountID, err := strconv.Atoi(rawAccountID); err == nil {
		ctx.AccountID = int(accountID)
	} else {
		err = goa.InvalidParamValue("accountID", rawAccountID, "number", err)
	}
	p, err := NewCreateBottlePayload(c.Payload)
	if err != nil {
		return nil, err
	}
	ctx.Payload = p
	return &ctx, err
}

// NewCreateBottlePayload instantiates a CreateBottlePayload from a raw request body.
// It validates each field and returns an error in case one or more validation fails.
func NewCreateBottlePayload(raw interface{}) (*CreateBottlePayload, error) {
	var err error
	p := CreateBottlePayload{}
	m, ok := raw.(map[string]interface{})
	if !ok {
		err = goa.InvalidPayload("map", err)
		goto end
	}
	if rawName, ok := m["name"]; ok {
		if name, ok := rawName.(string); ok {
			p.Name = name
		} else {
			err = goa.InvalidPayloadField("name", rawName, "string", err)
		}
	} else {
		err = goa.MissingPayloadField("name", err)
	}
	if rawVintage, ok := m["vintage"]; ok {
		if vintage, ok := rawVintage.(int); ok {
			p.Vintage = vintage
		} else {
			err = goa.InvalidPayloadField("vintage", rawVintage, "int", err)
		}
	} else {
		err = goa.MissingPayloadField("vintage", err)
	}
	if rawVineyard, ok := m["vineyard"]; ok {
		if vineyard, ok := rawVineyard.(string); ok {
			p.Vineyard = vineyard
		} else {
			err = goa.InvalidPayloadField("vineyard", rawVineyard, "string", err)
		}
	} else {
		err = goa.MissingPayloadField("vineyard", err)
	}
	if rawVarietal, ok := m["varietal"]; ok {
		if varietal, ok := rawVarietal.(string); ok {
			p.Varietal = &varietal
		} else {
			err = goa.InvalidPayloadField("varietal", rawVarietal, "string", err)
		}
	}
	if rawColor, ok := m["color"]; ok {
		if color, ok := rawColor.(string); ok {
			if color == "red" || color == "white" || color == "rose" || color == "yellow" {
				p.Color = &color
			} else {
				err = goa.InvalidPayloadFieldValue("color", rawColor, []string{"red", "white", "rose", "yellow"}, err)
			}
		} else {
			err = goa.InvalidPayloadField("color", rawColor, "string", err)
		}
	}
	if rawSweet, ok := m["sweet"]; ok {
		if sweet, ok := rawSweet.(bool); ok {
			p.Sweet = &sweet
		} else {
			err = goa.InvalidPayloadField("sweet", rawSweet, "bool", err)
		}
	}
	if rawRegion, ok := m["region"]; ok {
		if region, ok := rawRegion.(string); ok {
			p.Region = &region
		} else {
			err = goa.InvalidPayloadField("region", rawRegion, "string", err)
		}
	}
	if rawCountry, ok := m["country"]; ok {
		if country, ok := rawCountry.(string); ok {
			p.Country = &country
		} else {
			err = goa.InvalidPayloadField("country", rawCountry, "string", err)
		}
	}
	if rawReview, ok := m["review"]; ok {
		if review, ok := rawReview.(string); ok {
			p.Review = &review
		} else {
			err = goa.InvalidPayloadField("review", rawReview, "string", err)
		}
	}
end:
	return &p, err
}

// Created sends a HTTP response with status code 201 and an empty body.
func (c *CreateBottleContext) Created(bottle *Bottle) error {
	return c.JSON(201, bottle)
}

type (
	// UpdateBottleContext provides the bottles update action context
	UpdateBottleContext struct {
		*goa.Context
		AccountID int
		ID        int
		Payload   *UpdateBottlePayload
	}

	// UpdateBottlePayload provides the bottles update action payload
	UpdateBottlePayload struct {
		Name     *string
		Vintage  *int
		Vineyard *string
		Varietal *string
		Color    *string
		Sweet    *bool
		Country  *string
		Region   *string
		Review   *string
	}
)

// NewUpdateBottleContext parses the incoming request URL and body and instantiates the context
// accordingly. It returns an error if a required parameter is missing or if a parameter has an
// invalid value.
func NewUpdateBottleContext(c *goa.Context) (*UpdateBottleContext, error) {
	var err error
	ctx := UpdateBottleContext{Context: c}
	rawAccountID, _ := c.Get("accountID")
	if accountID, err := strconv.Atoi(rawAccountID); err == nil {
		ctx.AccountID = int(accountID)
	} else {
		err = goa.InvalidParamValue("accountID", rawAccountID, "number", err)
	}
	rawID, _ := c.Get("ID")
	if ID, err := strconv.Atoi(rawID); err == nil {
		ctx.ID = int(ID)
	} else {
		err = goa.InvalidParamValue("ID", rawID, "number", err)
	}
	p, err := NewUpdateBottlePayload(c.Payload)
	if err != nil {
		return nil, err
	}
	ctx.Payload = p
	return &ctx, err
}

// NewUpdateBottlePayload instantiates a UpdateBottlePayload from a raw request body.
// It validates each field and returns an error in case one or more validation fails.
func NewUpdateBottlePayload(raw interface{}) (*UpdateBottlePayload, error) {
	var err error
	p := UpdateBottlePayload{}
	m, ok := raw.(map[string]interface{})
	if !ok {
		err = goa.InvalidPayload("map", err)
		goto end
	}
	if rawName, ok := m["name"]; ok {
		if name, ok := rawName.(string); ok {
			p.Name = &name
		} else {
			err = goa.InvalidPayloadField("name", rawName, "string", err)
		}
	}
	if rawVintage, ok := m["vintage"]; ok {
		if vintage, ok := rawVintage.(int); ok {
			p.Vintage = &vintage
		} else {
			err = goa.InvalidPayloadField("vintage", rawVintage, "int", err)
		}
	}
	if rawVineyard, ok := m["vineyard"]; ok {
		if vineyard, ok := rawVineyard.(string); ok {
			p.Vineyard = &vineyard
		} else {
			err = goa.InvalidPayloadField("vineyard", rawVineyard, "string", err)
		}
	}
	if rawVarietal, ok := m["varietal"]; ok {
		if varietal, ok := rawVarietal.(string); ok {
			p.Varietal = &varietal
		} else {
			err = goa.InvalidPayloadField("varietal", rawVarietal, "string", err)
		}
	}
	if rawColor, ok := m["color"]; ok {
		if color, ok := rawColor.(string); ok {
			if color == "red" || color == "white" || color == "rose" || color == "yellow" {
				p.Color = &color
			} else {
				err = goa.InvalidPayloadFieldValue("color", rawColor, []string{"red", "white", "rose", "yellow"}, err)
			}
		} else {
			err = goa.InvalidPayloadField("color", rawColor, "string", err)
		}
	}
	if rawSweet, ok := m["sweet"]; ok {
		if sweet, ok := rawSweet.(bool); ok {
			p.Sweet = &sweet
		} else {
			err = goa.InvalidPayloadField("sweet", rawSweet, "bool", err)
		}
	}
	if rawRegion, ok := m["region"]; ok {
		if region, ok := rawRegion.(string); ok {
			p.Region = &region
		} else {
			err = goa.InvalidPayloadField("region", rawRegion, "string", err)
		}
	}
	if rawCountry, ok := m["country"]; ok {
		if country, ok := rawCountry.(string); ok {
			p.Country = &country
		} else {
			err = goa.InvalidPayloadField("country", rawCountry, "string", err)
		}
	}
	if rawReview, ok := m["review"]; ok {
		if review, ok := rawReview.(string); ok {
			p.Review = &review
		} else {
			err = goa.InvalidPayloadField("review", rawReview, "string", err)
		}
	}
end:
	return &p, err
}

// NotFound sends a HTTP response with status code 404 and an empty body.
func (c *UpdateBottleContext) NotFound() error {
	return c.Respond(404, nil)
}

// NoContent sends a HTTP response with status code 204 and an empty body.
func (c *UpdateBottleContext) NoContent() error {
	return c.Respond(204, nil)
}

// DeleteBottleContext provides the bottles delete action context
type DeleteBottleContext struct {
	*goa.Context
	AccountID int
	ID        int
}

// NewDeleteBottleContext parses the incoming request URL and body and instantiates the context
// accordingly. It returns an error if a required parameter is missing or if a parameter has an
// invalid value.
func NewDeleteBottleContext(c *goa.Context) (*DeleteBottleContext, error) {
	var err error
	ctx := DeleteBottleContext{Context: c}
	rawAccountID, _ := c.Get("accountID")
	if accountID, err := strconv.Atoi(rawAccountID); err == nil {
		ctx.AccountID = int(accountID)
	} else {
		err = goa.InvalidParamValue("accountID", rawAccountID, "number", err)
	}
	rawID, _ := c.Get("ID")
	if ID, err := strconv.Atoi(rawID); err == nil {
		ctx.ID = int(ID)
	} else {
		err = goa.InvalidParamValue("ID", rawID, "number", err)
	}
	return &ctx, err
}

// NotFound sends a HTTP response with status code 404 and an empty body.
func (c *DeleteBottleContext) NotFound() error {
	return c.Respond(404, nil)
}

// NoContent sends a HTTP response with status code 204 and an empty body.
func (c *DeleteBottleContext) NoContent() error {
	return c.Respond(204, nil)
}

// RateBottleContext provides the bottles rate action context
type RateBottleContext struct {
	*goa.Context
	AccountID int
	ID        int
	Payload   *RateBottlePayload
}

// NewRateBottleContext parses the incoming request URL and body and instantiates the context
// accordingly. It returns an error if a required parameter is missing or if a parameter has an
// invalid value.
func NewRateBottleContext(c *goa.Context) (*RateBottleContext, error) {
	var err error
	ctx := RateBottleContext{Context: c}
	rawAccountID, _ := c.Get("accountID")
	if accountID, err := strconv.Atoi(rawAccountID); err == nil {
		ctx.AccountID = int(accountID)
	} else {
		err = goa.InvalidParamValue("accountID", rawAccountID, "number", err)
	}
	rawID, _ := c.Get("ID")
	if ID, err := strconv.Atoi(rawID); err == nil {
		ctx.ID = int(ID)
	} else {
		err = goa.InvalidParamValue("ID", rawID, "number", err)
	}
	var p RateBottlePayload
	if err := c.Bind(&p); err != nil {
		return nil, err
	}
	ctx.Payload = &p
	return &ctx, err
}

// RateBottlePayload provides the bottles create action payload
type RateBottlePayload struct {
	Ratings int `json:"ratings"`
}

// Validate applies the payload validation rules and returns an error in case of failure.
func (p *RateBottlePayload) Validate() error {
	return nil
}

// NotFound sends a HTTP response with status code 404 and an empty body.
func (c *RateBottleContext) NotFound() error {
	return c.Respond(404, nil)
}

// NoContent sends a HTTP response with status code 204 and an empty body.
func (c *RateBottleContext) NoContent() error {
	return c.Respond(204, nil)
}
