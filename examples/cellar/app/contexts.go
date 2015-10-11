//************************************************************************//
// cellar: Application Contexts
//
// Generated with codegen v0.0.1, command line:
// $ /home/raphael/go/src/github.com/raphael/goa/examples/cellar/codegen485234072/codegen
// --out=/home/raphael/go/src/github.com/raphael/goa/examples/cellar
// --design=github.com/raphael/goa/examples/cellar/design
// --force
// --pkg=app
//
// The content of this file is auto-generated, DO NOT MODIFY
//************************************************************************//

package app

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/raphael/goa"
)

// CreateAccountContext provides the account create action context.
type CreateAccountContext struct {
	goa.Context
	AccountID int

	Payload *CreateAccountPayload
}

// NewCreateAccountContext parses the incoming request URL and body, performs validations and creates the
// context used by the account controller create action.
func NewCreateAccountContext(c goa.Context) (*CreateAccountContext, error) {
	var err error
	ctx := CreateAccountContext{Context: c}
	if c.Header().Get("Auth-Token") == "" {
		err = goa.MissingHeaderError("Auth-Token", err)
	}
	rawAccountID, ok := c.Get("accountID")
	if ok {
		if accountID, err2 := strconv.Atoi(rawAccountID); err2 == nil {
			ctx.AccountID = int(accountID)
		} else {
			err = goa.InvalidParamTypeError("accountID", rawAccountID, "integer", err2)
		}
	}
	if payload := c.Payload(); payload != nil {
		p, err := NewCreateAccountPayload(payload)
		if err != nil {
			return nil, err
		}
		ctx.Payload = p
	}
	return &ctx, err
}

// CreateAccountPayload is the account create action payload.
type CreateAccountPayload struct {
	Name string `json:"name"`
}

// NewCreateAccountPayload instantiates a CreateAccountPayload from a raw request body.
// It validates each field and returns an error if any validation fails.
func NewCreateAccountPayload(raw interface{}) (*CreateAccountPayload, error) {
	var err error
	var p *CreateAccountPayload
	if val, ok := raw.(map[string]interface{}); ok {
		p = new(CreateAccountPayload)
		if v, ok := val["name"]; ok {
			var tmp1 string
			if val, ok := v.(string); ok {
				tmp1 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Name`, v, "string", err)
			}
			p.Name = tmp1
		}
	} else {
		err = goa.InvalidAttributeTypeError(``, raw, "map[string]interface{}", err)
	}

	return p, err
}

// Created sends a HTTP response with status code 201.
func (c *CreateAccountContext) Created() error {
	return c.Respond(201, nil)
}

// DeleteAccountContext provides the account delete action context.
type DeleteAccountContext struct {
	goa.Context
	AccountID int
	ID        int
}

// NewDeleteAccountContext parses the incoming request URL and body, performs validations and creates the
// context used by the account controller delete action.
func NewDeleteAccountContext(c goa.Context) (*DeleteAccountContext, error) {
	var err error
	ctx := DeleteAccountContext{Context: c}
	if c.Header().Get("Auth-Token") == "" {
		err = goa.MissingHeaderError("Auth-Token", err)
	}
	rawAccountID, ok := c.Get("accountID")
	if ok {
		if accountID, err2 := strconv.Atoi(rawAccountID); err2 == nil {
			ctx.AccountID = int(accountID)
		} else {
			err = goa.InvalidParamTypeError("accountID", rawAccountID, "integer", err2)
		}
	}
	rawID, ok := c.Get("id")
	if ok {
		if id, err2 := strconv.Atoi(rawID); err2 == nil {
			ctx.ID = int(id)
		} else {
			err = goa.InvalidParamTypeError("id", rawID, "integer", err2)
		}
	}
	return &ctx, err
}

// NoContent sends a HTTP response with status code 204.
func (c *DeleteAccountContext) NoContent() error {
	return c.Respond(204, nil)
}

// NotFound sends a HTTP response with status code 404.
func (c *DeleteAccountContext) NotFound() error {
	return c.Respond(404, nil)
}

// ShowAccountContext provides the account show action context.
type ShowAccountContext struct {
	goa.Context
	AccountID int
	ID        int
}

// NewShowAccountContext parses the incoming request URL and body, performs validations and creates the
// context used by the account controller show action.
func NewShowAccountContext(c goa.Context) (*ShowAccountContext, error) {
	var err error
	ctx := ShowAccountContext{Context: c}
	if c.Header().Get("Auth-Token") == "" {
		err = goa.MissingHeaderError("Auth-Token", err)
	}
	rawAccountID, ok := c.Get("accountID")
	if ok {
		if accountID, err2 := strconv.Atoi(rawAccountID); err2 == nil {
			ctx.AccountID = int(accountID)
		} else {
			err = goa.InvalidParamTypeError("accountID", rawAccountID, "integer", err2)
		}
	}
	rawID, ok := c.Get("id")
	if ok {
		if id, err2 := strconv.Atoi(rawID); err2 == nil {
			ctx.ID = int(id)
		} else {
			err = goa.InvalidParamTypeError("id", rawID, "integer", err2)
		}
	}
	return &ctx, err
}

// NotFound sends a HTTP response with status code 404.
func (c *ShowAccountContext) NotFound() error {
	return c.Respond(404, nil)
}

// OK sends a HTTP response with status code 200.
func (c *ShowAccountContext) OK(resp *ExampleAccountMedia, view ExampleAccountMediaViewEnum) error {
	var r interface{}
	if view == ExampleAccountMediaDefaultView {
		if resp.Name == "" {
			return fmt.Errorf("missing required attribute \"name\"")
		}
		r = map[string]interface{}{
			"href": resp.Href,
			"id":   resp.ID,
			"name": resp.Name,
		}

	}
	if view == ExampleAccountMediaFullView {
		if resp.Name == "" {
			return fmt.Errorf("missing required attribute \"name\"")
		}
		r = map[string]interface{}{
			"created_at": resp.CreatedAt,
			"created_by": resp.CreatedBy,
			"href":       resp.Href,
			"id":         resp.ID,
			"name":       resp.Name,
		}

	}
	if view == ExampleAccountMediaLinkView {
		if resp.Name == "" {
			return fmt.Errorf("missing required attribute \"name\"")
		}
		r = map[string]interface{}{
			"href": resp.Href,
			"name": resp.Name,
		}

	}
	return c.JSON(200, r)
}

// UpdateAccountContext provides the account update action context.
type UpdateAccountContext struct {
	goa.Context
	AccountID int
	ID        int

	Payload *UpdateAccountPayload
}

// NewUpdateAccountContext parses the incoming request URL and body, performs validations and creates the
// context used by the account controller update action.
func NewUpdateAccountContext(c goa.Context) (*UpdateAccountContext, error) {
	var err error
	ctx := UpdateAccountContext{Context: c}
	if c.Header().Get("Auth-Token") == "" {
		err = goa.MissingHeaderError("Auth-Token", err)
	}
	rawAccountID, ok := c.Get("accountID")
	if ok {
		if accountID, err2 := strconv.Atoi(rawAccountID); err2 == nil {
			ctx.AccountID = int(accountID)
		} else {
			err = goa.InvalidParamTypeError("accountID", rawAccountID, "integer", err2)
		}
	}
	rawID, ok := c.Get("id")
	if ok {
		if id, err2 := strconv.Atoi(rawID); err2 == nil {
			ctx.ID = int(id)
		} else {
			err = goa.InvalidParamTypeError("id", rawID, "integer", err2)
		}
	}
	if payload := c.Payload(); payload != nil {
		p, err := NewUpdateAccountPayload(payload)
		if err != nil {
			return nil, err
		}
		ctx.Payload = p
	}
	return &ctx, err
}

// UpdateAccountPayload is the account update action payload.
type UpdateAccountPayload struct {
	Name string `json:"name"`
}

// NewUpdateAccountPayload instantiates a UpdateAccountPayload from a raw request body.
// It validates each field and returns an error if any validation fails.
func NewUpdateAccountPayload(raw interface{}) (*UpdateAccountPayload, error) {
	var err error
	var p *UpdateAccountPayload
	if val, ok := raw.(map[string]interface{}); ok {
		p = new(UpdateAccountPayload)
		if v, ok := val["name"]; ok {
			var tmp2 string
			if val, ok := v.(string); ok {
				tmp2 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Name`, v, "string", err)
			}
			p.Name = tmp2
		}
	} else {
		err = goa.InvalidAttributeTypeError(``, raw, "map[string]interface{}", err)
	}

	return p, err
}

// NoContent sends a HTTP response with status code 204.
func (c *UpdateAccountContext) NoContent() error {
	return c.Respond(204, nil)
}

// NotFound sends a HTTP response with status code 404.
func (c *UpdateAccountContext) NotFound() error {
	return c.Respond(404, nil)
}

// CreateBottleContext provides the bottle create action context.
type CreateBottleContext struct {
	goa.Context
	AccountID int

	Payload *CreateBottlePayload
}

// NewCreateBottleContext parses the incoming request URL and body, performs validations and creates the
// context used by the bottle controller create action.
func NewCreateBottleContext(c goa.Context) (*CreateBottleContext, error) {
	var err error
	ctx := CreateBottleContext{Context: c}
	if c.Header().Get("Auth-Token") == "" {
		err = goa.MissingHeaderError("Auth-Token", err)
	}
	rawAccountID, ok := c.Get("accountID")
	if ok {
		if accountID, err2 := strconv.Atoi(rawAccountID); err2 == nil {
			ctx.AccountID = int(accountID)
		} else {
			err = goa.InvalidParamTypeError("accountID", rawAccountID, "integer", err2)
		}
	}
	if payload := c.Payload(); payload != nil {
		p, err := NewCreateBottlePayload(payload)
		if err != nil {
			return nil, err
		}
		ctx.Payload = p
	}
	return &ctx, err
}

// CreateBottlePayload is the bottle create action payload.
type CreateBottlePayload struct {
	Characteristics string `json:"characteristics,omitempty"`
	Color           string `json:"color,omitempty"`
	Country         string `json:"country,omitempty"`
	Name            string `json:"name"`
	Region          string `json:"region,omitempty"`
	Review          string `json:"review,omitempty"`
	Sweet           string `json:"sweet,omitempty"`
	Varietal        string `json:"varietal,omitempty"`
	Vineyard        string `json:"vineyard"`
	Vintage         string `json:"vintage"`
}

// NewCreateBottlePayload instantiates a CreateBottlePayload from a raw request body.
// It validates each field and returns an error if any validation fails.
func NewCreateBottlePayload(raw interface{}) (*CreateBottlePayload, error) {
	var err error
	var p *CreateBottlePayload
	if val, ok := raw.(map[string]interface{}); ok {
		p = new(CreateBottlePayload)
		if v, ok := val["characteristics"]; ok {
			var tmp3 string
			if val, ok := v.(string); ok {
				tmp3 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Characteristics`, v, "string", err)
			}
			p.Characteristics = tmp3
		}
		if v, ok := val["color"]; ok {
			var tmp4 string
			if val, ok := v.(string); ok {
				tmp4 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Color`, v, "string", err)
			}
			p.Color = tmp4
		}
		if v, ok := val["country"]; ok {
			var tmp5 string
			if val, ok := v.(string); ok {
				tmp5 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Country`, v, "string", err)
			}
			p.Country = tmp5
		}
		if v, ok := val["name"]; ok {
			var tmp6 string
			if val, ok := v.(string); ok {
				tmp6 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Name`, v, "string", err)
			}
			p.Name = tmp6
		}
		if v, ok := val["region"]; ok {
			var tmp7 string
			if val, ok := v.(string); ok {
				tmp7 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Region`, v, "string", err)
			}
			p.Region = tmp7
		}
		if v, ok := val["review"]; ok {
			var tmp8 string
			if val, ok := v.(string); ok {
				tmp8 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Review`, v, "string", err)
			}
			p.Review = tmp8
		}
		if v, ok := val["sweet"]; ok {
			var tmp9 string
			if val, ok := v.(string); ok {
				tmp9 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Sweet`, v, "string", err)
			}
			p.Sweet = tmp9
		}
		if v, ok := val["varietal"]; ok {
			var tmp10 string
			if val, ok := v.(string); ok {
				tmp10 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Varietal`, v, "string", err)
			}
			p.Varietal = tmp10
		}
		if v, ok := val["vineyard"]; ok {
			var tmp11 string
			if val, ok := v.(string); ok {
				tmp11 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Vineyard`, v, "string", err)
			}
			p.Vineyard = tmp11
		}
		if v, ok := val["vintage"]; ok {
			var tmp12 string
			if val, ok := v.(string); ok {
				tmp12 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Vintage`, v, "string", err)
			}
			p.Vintage = tmp12
		}
	} else {
		err = goa.InvalidAttributeTypeError(``, raw, "map[string]interface{}", err)
	}

	return p, err
}

// Created sends a HTTP response with status code 201.
func (c *CreateBottleContext) Created() error {
	return c.Respond(201, nil)
}

// DeleteBottleContext provides the bottle delete action context.
type DeleteBottleContext struct {
	goa.Context
	AccountID int
	ID        int
}

// NewDeleteBottleContext parses the incoming request URL and body, performs validations and creates the
// context used by the bottle controller delete action.
func NewDeleteBottleContext(c goa.Context) (*DeleteBottleContext, error) {
	var err error
	ctx := DeleteBottleContext{Context: c}
	if c.Header().Get("Auth-Token") == "" {
		err = goa.MissingHeaderError("Auth-Token", err)
	}
	rawAccountID, ok := c.Get("accountID")
	if ok {
		if accountID, err2 := strconv.Atoi(rawAccountID); err2 == nil {
			ctx.AccountID = int(accountID)
		} else {
			err = goa.InvalidParamTypeError("accountID", rawAccountID, "integer", err2)
		}
	}
	rawID, ok := c.Get("id")
	if ok {
		if id, err2 := strconv.Atoi(rawID); err2 == nil {
			ctx.ID = int(id)
		} else {
			err = goa.InvalidParamTypeError("id", rawID, "integer", err2)
		}
	}
	return &ctx, err
}

// NoContent sends a HTTP response with status code 204.
func (c *DeleteBottleContext) NoContent() error {
	return c.Respond(204, nil)
}

// NotFound sends a HTTP response with status code 404.
func (c *DeleteBottleContext) NotFound() error {
	return c.Respond(404, nil)
}

// ListBottleContext provides the bottle list action context.
type ListBottleContext struct {
	goa.Context
	AccountID int
	Years     []int

	HasYears bool
}

// NewListBottleContext parses the incoming request URL and body, performs validations and creates the
// context used by the bottle controller list action.
func NewListBottleContext(c goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}
	if c.Header().Get("Auth-Token") == "" {
		err = goa.MissingHeaderError("Auth-Token", err)
	}
	rawAccountID, ok := c.Get("accountID")
	if ok {
		if accountID, err2 := strconv.Atoi(rawAccountID); err2 == nil {
			ctx.AccountID = int(accountID)
		} else {
			err = goa.InvalidParamTypeError("accountID", rawAccountID, "integer", err2)
		}
	}
	rawYears, ok := c.Get("years")
	if ok {
		elemsYears := strings.Split(rawYears, ",")
		elemsYears2 := make([]int, len(elemsYears))
		for i, rawElem := range elemsYears {
			if elem, err2 := strconv.Atoi(rawElem); err2 == nil {
				elemsYears2[i] = int(elem)
			} else {
				err = goa.InvalidParamTypeError("elem", rawElem, "integer", err2)
			}
		}
		ctx.Years = elemsYears2
		ctx.HasYears = true
	}
	return &ctx, err
}

// OK sends a HTTP response with status code 200.
func (c *ListBottleContext) OK(resp *ExampleBottleMediaCollection, view ExampleBottleMediaCollectionViewEnum) error {
	var r interface{}
	if view == ExampleBottleMediaCollectionDefaultView {
		r = map[string]interface{}{
			"href":     resp.Href,
			"id":       resp.ID,
			"links":    resp.Links,
			"name":     resp.Name,
			"varietal": resp.Varietal,
			"vineyard": resp.Vineyard,
			"vintage":  resp.Vintage,
		}
		links := make(map[string]interface{})
		if resp.Account.Name == "" {
			return fmt.Errorf("missing required attribute \"name\"")
		}
		links["Account"] = map[string]interface{}{
			"href": resp.Account.Href,
			"name": resp.Account.Name,
		}
		r["links"] = links

	}
	if view == ExampleBottleMediaCollectionFullView {
		r = map[string]interface{}{
			"account":         resp.Account,
			"characteristics": resp.Characteristics,
			"color":           resp.Color,
			"country":         resp.Country,
			"created_at":      resp.CreatedAt,
			"href":            resp.Href,
			"id":              resp.ID,
			"name":            resp.Name,
			"region":          resp.Region,
			"review":          resp.Review,
			"sweet":           resp.Sweet,
			"updated_at":      resp.UpdatedAt,
			"varietal":        resp.Varietal,
			"vineyard":        resp.Vineyard,
			"vintage":         resp.Vintage,
		}
		links := make(map[string]interface{})
		if resp.Account.Name == "" {
			return fmt.Errorf("missing required attribute \"name\"")
		}
		links["Account"] = map[string]interface{}{
			"href": resp.Account.Href,
			"name": resp.Account.Name,
		}
		r["links"] = links

	}
	return c.JSON(200, r)
}

// RateBottleContext provides the bottle rate action context.
type RateBottleContext struct {
	goa.Context
	AccountID int
	ID        int

	Payload *RateBottlePayload
}

// NewRateBottleContext parses the incoming request URL and body, performs validations and creates the
// context used by the bottle controller rate action.
func NewRateBottleContext(c goa.Context) (*RateBottleContext, error) {
	var err error
	ctx := RateBottleContext{Context: c}
	if c.Header().Get("Auth-Token") == "" {
		err = goa.MissingHeaderError("Auth-Token", err)
	}
	rawAccountID, ok := c.Get("accountID")
	if ok {
		if accountID, err2 := strconv.Atoi(rawAccountID); err2 == nil {
			ctx.AccountID = int(accountID)
		} else {
			err = goa.InvalidParamTypeError("accountID", rawAccountID, "integer", err2)
		}
	}
	rawID, ok := c.Get("id")
	if ok {
		if id, err2 := strconv.Atoi(rawID); err2 == nil {
			ctx.ID = int(id)
		} else {
			err = goa.InvalidParamTypeError("id", rawID, "integer", err2)
		}
	}
	if payload := c.Payload(); payload != nil {
		p, err := NewRateBottlePayload(payload)
		if err != nil {
			return nil, err
		}
		ctx.Payload = p
	}
	return &ctx, err
}

// RateBottlePayload is the bottle rate action payload.
type RateBottlePayload struct {
	Rating string `json:"rating,omitempty"`
}

// NewRateBottlePayload instantiates a RateBottlePayload from a raw request body.
// It validates each field and returns an error if any validation fails.
func NewRateBottlePayload(raw interface{}) (*RateBottlePayload, error) {
	var err error
	var p *RateBottlePayload
	if val, ok := raw.(map[string]interface{}); ok {
		p = new(RateBottlePayload)
		if v, ok := val["rating"]; ok {
			var tmp13 string
			if val, ok := v.(string); ok {
				tmp13 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Rating`, v, "string", err)
			}
			p.Rating = tmp13
		}
	} else {
		err = goa.InvalidAttributeTypeError(``, raw, "map[string]interface{}", err)
	}

	return p, err
}

// NoContent sends a HTTP response with status code 204.
func (c *RateBottleContext) NoContent() error {
	return c.Respond(204, nil)
}

// NotFound sends a HTTP response with status code 404.
func (c *RateBottleContext) NotFound() error {
	return c.Respond(404, nil)
}

// ShowBottleContext provides the bottle show action context.
type ShowBottleContext struct {
	goa.Context
	AccountID int
	ID        int
}

// NewShowBottleContext parses the incoming request URL and body, performs validations and creates the
// context used by the bottle controller show action.
func NewShowBottleContext(c goa.Context) (*ShowBottleContext, error) {
	var err error
	ctx := ShowBottleContext{Context: c}
	if c.Header().Get("Auth-Token") == "" {
		err = goa.MissingHeaderError("Auth-Token", err)
	}
	rawAccountID, ok := c.Get("accountID")
	if ok {
		if accountID, err2 := strconv.Atoi(rawAccountID); err2 == nil {
			ctx.AccountID = int(accountID)
		} else {
			err = goa.InvalidParamTypeError("accountID", rawAccountID, "integer", err2)
		}
	}
	rawID, ok := c.Get("id")
	if ok {
		if id, err2 := strconv.Atoi(rawID); err2 == nil {
			ctx.ID = int(id)
		} else {
			err = goa.InvalidParamTypeError("id", rawID, "integer", err2)
		}
	}
	return &ctx, err
}

// NotFound sends a HTTP response with status code 404.
func (c *ShowBottleContext) NotFound() error {
	return c.Respond(404, nil)
}

// OK sends a HTTP response with status code 200.
func (c *ShowBottleContext) OK(resp *ExampleBottleMedia, view ExampleBottleMediaViewEnum) error {
	var r interface{}
	if view == ExampleBottleMediaDefaultView {
		if resp.Name == "" {
			return fmt.Errorf("missing required attribute \"name\"")
		}
		r = map[string]interface{}{
			"href":     resp.Href,
			"id":       resp.ID,
			"links":    resp.Links,
			"name":     resp.Name,
			"varietal": resp.Varietal,
			"vineyard": resp.Vineyard,
			"vintage":  resp.Vintage,
		}
		links := make(map[string]interface{})
		if resp.Account.Name == "" {
			return fmt.Errorf("missing required attribute \"name\"")
		}
		links["Account"] = map[string]interface{}{
			"href": resp.Account.Href,
			"name": resp.Account.Name,
		}
		r["links"] = links

	}
	if view == ExampleBottleMediaFullView {
		if resp.Account == "" {
			return fmt.Errorf("missing required attribute \"account\"")
		}
		r = map[string]interface{}{
			"account":         resp.Account,
			"characteristics": resp.Characteristics,
			"color":           resp.Color,
			"country":         resp.Country,
			"created_at":      resp.CreatedAt,
			"href":            resp.Href,
			"id":              resp.ID,
			"name":            resp.Name,
			"region":          resp.Region,
			"review":          resp.Review,
			"sweet":           resp.Sweet,
			"updated_at":      resp.UpdatedAt,
			"varietal":        resp.Varietal,
			"vineyard":        resp.Vineyard,
			"vintage":         resp.Vintage,
		}
		links := make(map[string]interface{})
		if resp.Account.Name == "" {
			return fmt.Errorf("missing required attribute \"name\"")
		}
		links["Account"] = map[string]interface{}{
			"href": resp.Account.Href,
			"name": resp.Account.Name,
		}
		r["links"] = links

	}
	return c.JSON(200, r)
}

// UpdateBottleContext provides the bottle update action context.
type UpdateBottleContext struct {
	goa.Context
	AccountID int
	ID        int

	Payload *UpdateBottlePayload
}

// NewUpdateBottleContext parses the incoming request URL and body, performs validations and creates the
// context used by the bottle controller update action.
func NewUpdateBottleContext(c goa.Context) (*UpdateBottleContext, error) {
	var err error
	ctx := UpdateBottleContext{Context: c}
	if c.Header().Get("Auth-Token") == "" {
		err = goa.MissingHeaderError("Auth-Token", err)
	}
	rawAccountID, ok := c.Get("accountID")
	if ok {
		if accountID, err2 := strconv.Atoi(rawAccountID); err2 == nil {
			ctx.AccountID = int(accountID)
		} else {
			err = goa.InvalidParamTypeError("accountID", rawAccountID, "integer", err2)
		}
	}
	rawID, ok := c.Get("id")
	if ok {
		if id, err2 := strconv.Atoi(rawID); err2 == nil {
			ctx.ID = int(id)
		} else {
			err = goa.InvalidParamTypeError("id", rawID, "integer", err2)
		}
	}
	if payload := c.Payload(); payload != nil {
		p, err := NewUpdateBottlePayload(payload)
		if err != nil {
			return nil, err
		}
		ctx.Payload = p
	}
	return &ctx, err
}

// UpdateBottlePayload is the bottle update action payload.
type UpdateBottlePayload struct {
	Characteristics string `json:"characteristics,omitempty"`
	Color           string `json:"color,omitempty"`
	Country         string `json:"country,omitempty"`
	Name            string `json:"name,omitempty"`
	Region          string `json:"region,omitempty"`
	Review          string `json:"review,omitempty"`
	Sweet           string `json:"sweet,omitempty"`
	Varietal        string `json:"varietal,omitempty"`
	Vineyard        string `json:"vineyard,omitempty"`
	Vintage         string `json:"vintage,omitempty"`
}

// NewUpdateBottlePayload instantiates a UpdateBottlePayload from a raw request body.
// It validates each field and returns an error if any validation fails.
func NewUpdateBottlePayload(raw interface{}) (*UpdateBottlePayload, error) {
	var err error
	var p *UpdateBottlePayload
	if val, ok := raw.(map[string]interface{}); ok {
		p = new(UpdateBottlePayload)
		if v, ok := val["characteristics"]; ok {
			var tmp14 string
			if val, ok := v.(string); ok {
				tmp14 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Characteristics`, v, "string", err)
			}
			p.Characteristics = tmp14
		}
		if v, ok := val["color"]; ok {
			var tmp15 string
			if val, ok := v.(string); ok {
				tmp15 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Color`, v, "string", err)
			}
			p.Color = tmp15
		}
		if v, ok := val["country"]; ok {
			var tmp16 string
			if val, ok := v.(string); ok {
				tmp16 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Country`, v, "string", err)
			}
			p.Country = tmp16
		}
		if v, ok := val["name"]; ok {
			var tmp17 string
			if val, ok := v.(string); ok {
				tmp17 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Name`, v, "string", err)
			}
			p.Name = tmp17
		}
		if v, ok := val["region"]; ok {
			var tmp18 string
			if val, ok := v.(string); ok {
				tmp18 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Region`, v, "string", err)
			}
			p.Region = tmp18
		}
		if v, ok := val["review"]; ok {
			var tmp19 string
			if val, ok := v.(string); ok {
				tmp19 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Review`, v, "string", err)
			}
			p.Review = tmp19
		}
		if v, ok := val["sweet"]; ok {
			var tmp20 string
			if val, ok := v.(string); ok {
				tmp20 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Sweet`, v, "string", err)
			}
			p.Sweet = tmp20
		}
		if v, ok := val["varietal"]; ok {
			var tmp21 string
			if val, ok := v.(string); ok {
				tmp21 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Varietal`, v, "string", err)
			}
			p.Varietal = tmp21
		}
		if v, ok := val["vineyard"]; ok {
			var tmp22 string
			if val, ok := v.(string); ok {
				tmp22 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Vineyard`, v, "string", err)
			}
			p.Vineyard = tmp22
		}
		if v, ok := val["vintage"]; ok {
			var tmp23 string
			if val, ok := v.(string); ok {
				tmp23 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Vintage`, v, "string", err)
			}
			p.Vintage = tmp23
		}
	} else {
		err = goa.InvalidAttributeTypeError(``, raw, "map[string]interface{}", err)
	}

	return p, err
}

// NoContent sends a HTTP response with status code 204.
func (c *UpdateBottleContext) NoContent() error {
	return c.Respond(204, nil)
}

// NotFound sends a HTTP response with status code 404.
func (c *UpdateBottleContext) NotFound() error {
	return c.Respond(404, nil)
}
