//************************************************************************//
// cellar: Application Contexts
//
// Generated with goagen v0.0.1, command line:
// $ goagen
// --out=/home/raphael/go/src/github.com/raphael/goa/examples/cellar
// --design=github.com/raphael/goa/examples/cellar/design
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
	*goa.Context
	Payload *CreateAccountPayload
}

// NewCreateAccountContext parses the incoming request URL and body, performs validations and creates the
// context used by the account controller create action.
func NewCreateAccountContext(c *goa.Context) (*CreateAccountContext, error) {
	var err error
	ctx := CreateAccountContext{Context: c}

	p, err := NewCreateAccountPayload(c.Payload())
	if err != nil {
		return nil, err
	}
	ctx.Payload = p
	return &ctx, err
}

// CreateAccountPayload is the account create action payload.
type CreateAccountPayload struct {
	// Name of account
	Name string
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
				err = goa.InvalidAttributeTypeError(`payload.Name`, v, "string", err)
			}
			p.Name = tmp1
		} else {
			err = goa.MissingAttributeError(`payload`, "name", err)
		}
	} else {
		err = goa.InvalidAttributeTypeError(`payload`, raw, "map[string]interface{}", err)
	}
	return p, err
}

// Created sends a HTTP response with status code 201.
func (ctx *CreateAccountContext) Created() error {
	return ctx.Respond(201, nil)
}

// DeleteAccountContext provides the account delete action context.
type DeleteAccountContext struct {
	*goa.Context
	AccountID int
}

// NewDeleteAccountContext parses the incoming request URL and body, performs validations and creates the
// context used by the account controller delete action.
func NewDeleteAccountContext(c *goa.Context) (*DeleteAccountContext, error) {
	var err error
	ctx := DeleteAccountContext{Context: c}

	rawAccountID, ok := c.Get("accountID")
	if ok {
		if accountID, err2 := strconv.Atoi(rawAccountID); err2 == nil {
			ctx.AccountID = int(accountID)
		} else {
			err = goa.InvalidParamTypeError("accountID", rawAccountID, "integer", err)
		}
	}
	return &ctx, err
}

// NoContent sends a HTTP response with status code 204.
func (ctx *DeleteAccountContext) NoContent() error {
	return ctx.Respond(204, nil)
}

// NotFound sends a HTTP response with status code 404.
func (ctx *DeleteAccountContext) NotFound() error {
	return ctx.Respond(404, nil)
}

// ShowAccountContext provides the account show action context.
type ShowAccountContext struct {
	*goa.Context
	AccountID int
}

// NewShowAccountContext parses the incoming request URL and body, performs validations and creates the
// context used by the account controller show action.
func NewShowAccountContext(c *goa.Context) (*ShowAccountContext, error) {
	var err error
	ctx := ShowAccountContext{Context: c}

	rawAccountID, ok := c.Get("accountID")
	if ok {
		if accountID, err2 := strconv.Atoi(rawAccountID); err2 == nil {
			ctx.AccountID = int(accountID)
		} else {
			err = goa.InvalidParamTypeError("accountID", rawAccountID, "integer", err)
		}
	}
	return &ctx, err
}

// NotFound sends a HTTP response with status code 404.
func (ctx *ShowAccountContext) NotFound() error {
	return ctx.Respond(404, nil)
}

// OK sends a HTTP response with status code 200.
func (ctx *ShowAccountContext) OK(resp *Account, view AccountViewEnum) error {
	r, err := resp.Dump(view)
	if err != nil {
		return fmt.Errorf("invalid response: %s", err)
	}
	ctx.Header().Set("Content-Type", "application/vnd.goa.example.account; charset=utf-8")
	return ctx.JSON(200, r)
}

// UpdateAccountContext provides the account update action context.
type UpdateAccountContext struct {
	*goa.Context
	AccountID int
	Payload   *UpdateAccountPayload
}

// NewUpdateAccountContext parses the incoming request URL and body, performs validations and creates the
// context used by the account controller update action.
func NewUpdateAccountContext(c *goa.Context) (*UpdateAccountContext, error) {
	var err error
	ctx := UpdateAccountContext{Context: c}

	rawAccountID, ok := c.Get("accountID")
	if ok {
		if accountID, err2 := strconv.Atoi(rawAccountID); err2 == nil {
			ctx.AccountID = int(accountID)
		} else {
			err = goa.InvalidParamTypeError("accountID", rawAccountID, "integer", err)
		}
	}
	p, err := NewUpdateAccountPayload(c.Payload())
	if err != nil {
		return nil, err
	}
	ctx.Payload = p
	return &ctx, err
}

// UpdateAccountPayload is the account update action payload.
type UpdateAccountPayload struct {
	// Name of account
	Name string
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
				err = goa.InvalidAttributeTypeError(`payload.Name`, v, "string", err)
			}
			p.Name = tmp2
		} else {
			err = goa.MissingAttributeError(`payload`, "name", err)
		}
	} else {
		err = goa.InvalidAttributeTypeError(`payload`, raw, "map[string]interface{}", err)
	}
	return p, err
}

// NoContent sends a HTTP response with status code 204.
func (ctx *UpdateAccountContext) NoContent() error {
	return ctx.Respond(204, nil)
}

// NotFound sends a HTTP response with status code 404.
func (ctx *UpdateAccountContext) NotFound() error {
	return ctx.Respond(404, nil)
}

// CreateBottleContext provides the bottle create action context.
type CreateBottleContext struct {
	*goa.Context
	AccountID int
	Payload   *CreateBottlePayload
}

// NewCreateBottleContext parses the incoming request URL and body, performs validations and creates the
// context used by the bottle controller create action.
func NewCreateBottleContext(c *goa.Context) (*CreateBottleContext, error) {
	var err error
	ctx := CreateBottleContext{Context: c}

	rawAccountID, ok := c.Get("accountID")
	if ok {
		if accountID, err2 := strconv.Atoi(rawAccountID); err2 == nil {
			ctx.AccountID = int(accountID)
		} else {
			err = goa.InvalidParamTypeError("accountID", rawAccountID, "integer", err)
		}
	}
	p, err := NewCreateBottlePayload(c.Payload())
	if err != nil {
		return nil, err
	}
	ctx.Payload = p
	return &ctx, err
}

// CreateBottlePayload is the bottle create action payload.
type CreateBottlePayload struct {
	Characteristics string
	Color           string
	Country         string
	Name            string
	Region          string
	Review          string
	Sweetness       int
	Varietal        string
	Vineyard        string
	Vintage         int
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
				err = goa.InvalidAttributeTypeError(`payload.Characteristics`, v, "string", err)
			}
			if err == nil {
				if len(tmp3) < 10 {
					err = goa.InvalidLengthError(`payload.Characteristics`, tmp3, 10, true, err)
				}
				if len(tmp3) > 300 {
					err = goa.InvalidLengthError(`payload.Characteristics`, tmp3, 300, false, err)
				}
			}
			p.Characteristics = tmp3
		}
		if v, ok := val["color"]; ok {
			var tmp4 string
			if val, ok := v.(string); ok {
				tmp4 = val
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Color`, v, "string", err)
			}
			if err == nil {
				if tmp4 != "" {
					if !(tmp4 == "red" || tmp4 == "white" || tmp4 == "rose" || tmp4 == "yellow" || tmp4 == "sparkling") {
						err = goa.InvalidEnumValueError(`payload.Color`, tmp4, []interface{}{"red", "white", "rose", "yellow", "sparkling"}, err)
					}
				}
			}
			p.Color = tmp4
		} else {
			err = goa.MissingAttributeError(`payload`, "color", err)
		}
		if v, ok := val["country"]; ok {
			var tmp5 string
			if val, ok := v.(string); ok {
				tmp5 = val
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Country`, v, "string", err)
			}
			if err == nil {
				if len(tmp5) < 2 {
					err = goa.InvalidLengthError(`payload.Country`, tmp5, 2, true, err)
				}
			}
			p.Country = tmp5
		}
		if v, ok := val["name"]; ok {
			var tmp6 string
			if val, ok := v.(string); ok {
				tmp6 = val
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Name`, v, "string", err)
			}
			if err == nil {
				if len(tmp6) < 2 {
					err = goa.InvalidLengthError(`payload.Name`, tmp6, 2, true, err)
				}
			}
			p.Name = tmp6
		} else {
			err = goa.MissingAttributeError(`payload`, "name", err)
		}
		if v, ok := val["region"]; ok {
			var tmp7 string
			if val, ok := v.(string); ok {
				tmp7 = val
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Region`, v, "string", err)
			}
			p.Region = tmp7
		}
		if v, ok := val["review"]; ok {
			var tmp8 string
			if val, ok := v.(string); ok {
				tmp8 = val
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Review`, v, "string", err)
			}
			if err == nil {
				if len(tmp8) < 10 {
					err = goa.InvalidLengthError(`payload.Review`, tmp8, 10, true, err)
				}
				if len(tmp8) > 300 {
					err = goa.InvalidLengthError(`payload.Review`, tmp8, 300, false, err)
				}
			}
			p.Review = tmp8
		}
		if v, ok := val["sweetness"]; ok {
			var tmp9 int
			if f, ok := v.(float64); ok {
				tmp9 = int(f)
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Sweetness`, v, "int", err)
			}
			if err == nil {
				if tmp9 < 1 {
					err = goa.InvalidRangeError(`payload.Sweetness`, tmp9, 1, true, err)
				}
				if tmp9 > 5 {
					err = goa.InvalidRangeError(`payload.Sweetness`, tmp9, 5, false, err)
				}
			}
			p.Sweetness = tmp9
		}
		if v, ok := val["varietal"]; ok {
			var tmp10 string
			if val, ok := v.(string); ok {
				tmp10 = val
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Varietal`, v, "string", err)
			}
			if err == nil {
				if len(tmp10) < 4 {
					err = goa.InvalidLengthError(`payload.Varietal`, tmp10, 4, true, err)
				}
			}
			p.Varietal = tmp10
		} else {
			err = goa.MissingAttributeError(`payload`, "varietal", err)
		}
		if v, ok := val["vineyard"]; ok {
			var tmp11 string
			if val, ok := v.(string); ok {
				tmp11 = val
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Vineyard`, v, "string", err)
			}
			if err == nil {
				if len(tmp11) < 2 {
					err = goa.InvalidLengthError(`payload.Vineyard`, tmp11, 2, true, err)
				}
			}
			p.Vineyard = tmp11
		} else {
			err = goa.MissingAttributeError(`payload`, "vineyard", err)
		}
		if v, ok := val["vintage"]; ok {
			var tmp12 int
			if f, ok := v.(float64); ok {
				tmp12 = int(f)
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Vintage`, v, "int", err)
			}
			if err == nil {
				if tmp12 < 1900 {
					err = goa.InvalidRangeError(`payload.Vintage`, tmp12, 1900, true, err)
				}
				if tmp12 > 2020 {
					err = goa.InvalidRangeError(`payload.Vintage`, tmp12, 2020, false, err)
				}
			}
			p.Vintage = tmp12
		} else {
			err = goa.MissingAttributeError(`payload`, "vintage", err)
		}
	} else {
		err = goa.InvalidAttributeTypeError(`payload`, raw, "map[string]interface{}", err)
	}
	return p, err
}

// Created sends a HTTP response with status code 201.
func (ctx *CreateBottleContext) Created() error {
	return ctx.Respond(201, nil)
}

// DeleteBottleContext provides the bottle delete action context.
type DeleteBottleContext struct {
	*goa.Context
	AccountID int
	BottleID  int
}

// NewDeleteBottleContext parses the incoming request URL and body, performs validations and creates the
// context used by the bottle controller delete action.
func NewDeleteBottleContext(c *goa.Context) (*DeleteBottleContext, error) {
	var err error
	ctx := DeleteBottleContext{Context: c}

	rawAccountID, ok := c.Get("accountID")
	if ok {
		if accountID, err2 := strconv.Atoi(rawAccountID); err2 == nil {
			ctx.AccountID = int(accountID)
		} else {
			err = goa.InvalidParamTypeError("accountID", rawAccountID, "integer", err)
		}
	}
	rawBottleID, ok := c.Get("bottleID")
	if ok {
		if bottleID, err2 := strconv.Atoi(rawBottleID); err2 == nil {
			ctx.BottleID = int(bottleID)
		} else {
			err = goa.InvalidParamTypeError("bottleID", rawBottleID, "integer", err)
		}
	}
	return &ctx, err
}

// NoContent sends a HTTP response with status code 204.
func (ctx *DeleteBottleContext) NoContent() error {
	return ctx.Respond(204, nil)
}

// NotFound sends a HTTP response with status code 404.
func (ctx *DeleteBottleContext) NotFound() error {
	return ctx.Respond(404, nil)
}

// ListBottleContext provides the bottle list action context.
type ListBottleContext struct {
	*goa.Context
	AccountID int
	Years     []int

	HasYears bool
}

// NewListBottleContext parses the incoming request URL and body, performs validations and creates the
// context used by the bottle controller list action.
func NewListBottleContext(c *goa.Context) (*ListBottleContext, error) {
	var err error
	ctx := ListBottleContext{Context: c}

	rawAccountID, ok := c.Get("accountID")
	if ok {
		if accountID, err2 := strconv.Atoi(rawAccountID); err2 == nil {
			ctx.AccountID = int(accountID)
		} else {
			err = goa.InvalidParamTypeError("accountID", rawAccountID, "integer", err)
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
				err = goa.InvalidParamTypeError("elem", rawElem, "integer", err)
			}
		}
		ctx.Years = elemsYears2
		ctx.HasYears = true
	}
	return &ctx, err
}

// OK sends a HTTP response with status code 200.
func (ctx *ListBottleContext) OK(resp BottleCollection, view BottleCollectionViewEnum) error {
	r, err := resp.Dump(view)
	if err != nil {
		return fmt.Errorf("invalid response: %s", err)
	}
	ctx.Header().Set("Content-Type", "application/vnd.goa.example.bottle; type=collection; charset=utf-8")
	return ctx.JSON(200, r)
}

// RateBottleContext provides the bottle rate action context.
type RateBottleContext struct {
	*goa.Context
	AccountID int
	BottleID  int
	Payload   *RateBottlePayload
}

// NewRateBottleContext parses the incoming request URL and body, performs validations and creates the
// context used by the bottle controller rate action.
func NewRateBottleContext(c *goa.Context) (*RateBottleContext, error) {
	var err error
	ctx := RateBottleContext{Context: c}

	rawAccountID, ok := c.Get("accountID")
	if ok {
		if accountID, err2 := strconv.Atoi(rawAccountID); err2 == nil {
			ctx.AccountID = int(accountID)
		} else {
			err = goa.InvalidParamTypeError("accountID", rawAccountID, "integer", err)
		}
	}
	rawBottleID, ok := c.Get("bottleID")
	if ok {
		if bottleID, err2 := strconv.Atoi(rawBottleID); err2 == nil {
			ctx.BottleID = int(bottleID)
		} else {
			err = goa.InvalidParamTypeError("bottleID", rawBottleID, "integer", err)
		}
	}
	p, err := NewRateBottlePayload(c.Payload())
	if err != nil {
		return nil, err
	}
	ctx.Payload = p
	return &ctx, err
}

// RateBottlePayload is the bottle rate action payload.
type RateBottlePayload struct {
	// Rating of bottle between 1 and 5
	Rating int
}

// NewRateBottlePayload instantiates a RateBottlePayload from a raw request body.
// It validates each field and returns an error if any validation fails.
func NewRateBottlePayload(raw interface{}) (*RateBottlePayload, error) {
	var err error
	var p *RateBottlePayload
	if val, ok := raw.(map[string]interface{}); ok {
		p = new(RateBottlePayload)
		if v, ok := val["rating"]; ok {
			var tmp13 int
			if f, ok := v.(float64); ok {
				tmp13 = int(f)
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Rating`, v, "int", err)
			}
			if err == nil {
				if tmp13 < 1 {
					err = goa.InvalidRangeError(`payload.Rating`, tmp13, 1, true, err)
				}
				if tmp13 > 5 {
					err = goa.InvalidRangeError(`payload.Rating`, tmp13, 5, false, err)
				}
			}
			p.Rating = tmp13
		} else {
			err = goa.MissingAttributeError(`payload`, "rating", err)
		}
	} else {
		err = goa.InvalidAttributeTypeError(`payload`, raw, "map[string]interface{}", err)
	}
	return p, err
}

// NoContent sends a HTTP response with status code 204.
func (ctx *RateBottleContext) NoContent() error {
	return ctx.Respond(204, nil)
}

// NotFound sends a HTTP response with status code 404.
func (ctx *RateBottleContext) NotFound() error {
	return ctx.Respond(404, nil)
}

// ShowBottleContext provides the bottle show action context.
type ShowBottleContext struct {
	*goa.Context
	AccountID int
	BottleID  int
}

// NewShowBottleContext parses the incoming request URL and body, performs validations and creates the
// context used by the bottle controller show action.
func NewShowBottleContext(c *goa.Context) (*ShowBottleContext, error) {
	var err error
	ctx := ShowBottleContext{Context: c}

	rawAccountID, ok := c.Get("accountID")
	if ok {
		if accountID, err2 := strconv.Atoi(rawAccountID); err2 == nil {
			ctx.AccountID = int(accountID)
		} else {
			err = goa.InvalidParamTypeError("accountID", rawAccountID, "integer", err)
		}
	}
	rawBottleID, ok := c.Get("bottleID")
	if ok {
		if bottleID, err2 := strconv.Atoi(rawBottleID); err2 == nil {
			ctx.BottleID = int(bottleID)
		} else {
			err = goa.InvalidParamTypeError("bottleID", rawBottleID, "integer", err)
		}
	}
	return &ctx, err
}

// NotFound sends a HTTP response with status code 404.
func (ctx *ShowBottleContext) NotFound() error {
	return ctx.Respond(404, nil)
}

// OK sends a HTTP response with status code 200.
func (ctx *ShowBottleContext) OK(resp *Bottle, view BottleViewEnum) error {
	r, err := resp.Dump(view)
	if err != nil {
		return fmt.Errorf("invalid response: %s", err)
	}
	ctx.Header().Set("Content-Type", "application/vnd.goa.example.bottle; charset=utf-8")
	return ctx.JSON(200, r)
}

// UpdateBottleContext provides the bottle update action context.
type UpdateBottleContext struct {
	*goa.Context
	AccountID int
	BottleID  int
	Payload   *UpdateBottlePayload
}

// NewUpdateBottleContext parses the incoming request URL and body, performs validations and creates the
// context used by the bottle controller update action.
func NewUpdateBottleContext(c *goa.Context) (*UpdateBottleContext, error) {
	var err error
	ctx := UpdateBottleContext{Context: c}

	rawAccountID, ok := c.Get("accountID")
	if ok {
		if accountID, err2 := strconv.Atoi(rawAccountID); err2 == nil {
			ctx.AccountID = int(accountID)
		} else {
			err = goa.InvalidParamTypeError("accountID", rawAccountID, "integer", err)
		}
	}
	rawBottleID, ok := c.Get("bottleID")
	if ok {
		if bottleID, err2 := strconv.Atoi(rawBottleID); err2 == nil {
			ctx.BottleID = int(bottleID)
		} else {
			err = goa.InvalidParamTypeError("bottleID", rawBottleID, "integer", err)
		}
	}
	p, err := NewUpdateBottlePayload(c.Payload())
	if err != nil {
		return nil, err
	}
	ctx.Payload = p
	return &ctx, err
}

// UpdateBottlePayload is the bottle update action payload.
type UpdateBottlePayload struct {
	Characteristics string
	Color           string
	Country         string
	Name            string
	Region          string
	Review          string
	Sweetness       int
	Varietal        string
	Vineyard        string
	Vintage         int
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
				err = goa.InvalidAttributeTypeError(`payload.Characteristics`, v, "string", err)
			}
			if err == nil {
				if len(tmp14) < 10 {
					err = goa.InvalidLengthError(`payload.Characteristics`, tmp14, 10, true, err)
				}
				if len(tmp14) > 300 {
					err = goa.InvalidLengthError(`payload.Characteristics`, tmp14, 300, false, err)
				}
			}
			p.Characteristics = tmp14
		}
		if v, ok := val["color"]; ok {
			var tmp15 string
			if val, ok := v.(string); ok {
				tmp15 = val
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Color`, v, "string", err)
			}
			if err == nil {
				if tmp15 != "" {
					if !(tmp15 == "red" || tmp15 == "white" || tmp15 == "rose" || tmp15 == "yellow" || tmp15 == "sparkling") {
						err = goa.InvalidEnumValueError(`payload.Color`, tmp15, []interface{}{"red", "white", "rose", "yellow", "sparkling"}, err)
					}
				}
			}
			p.Color = tmp15
		}
		if v, ok := val["country"]; ok {
			var tmp16 string
			if val, ok := v.(string); ok {
				tmp16 = val
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Country`, v, "string", err)
			}
			if err == nil {
				if len(tmp16) < 2 {
					err = goa.InvalidLengthError(`payload.Country`, tmp16, 2, true, err)
				}
			}
			p.Country = tmp16
		}
		if v, ok := val["name"]; ok {
			var tmp17 string
			if val, ok := v.(string); ok {
				tmp17 = val
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Name`, v, "string", err)
			}
			if err == nil {
				if len(tmp17) < 2 {
					err = goa.InvalidLengthError(`payload.Name`, tmp17, 2, true, err)
				}
			}
			p.Name = tmp17
		}
		if v, ok := val["region"]; ok {
			var tmp18 string
			if val, ok := v.(string); ok {
				tmp18 = val
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Region`, v, "string", err)
			}
			p.Region = tmp18
		}
		if v, ok := val["review"]; ok {
			var tmp19 string
			if val, ok := v.(string); ok {
				tmp19 = val
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Review`, v, "string", err)
			}
			if err == nil {
				if len(tmp19) < 10 {
					err = goa.InvalidLengthError(`payload.Review`, tmp19, 10, true, err)
				}
				if len(tmp19) > 300 {
					err = goa.InvalidLengthError(`payload.Review`, tmp19, 300, false, err)
				}
			}
			p.Review = tmp19
		}
		if v, ok := val["sweetness"]; ok {
			var tmp20 int
			if f, ok := v.(float64); ok {
				tmp20 = int(f)
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Sweetness`, v, "int", err)
			}
			if err == nil {
				if tmp20 < 1 {
					err = goa.InvalidRangeError(`payload.Sweetness`, tmp20, 1, true, err)
				}
				if tmp20 > 5 {
					err = goa.InvalidRangeError(`payload.Sweetness`, tmp20, 5, false, err)
				}
			}
			p.Sweetness = tmp20
		}
		if v, ok := val["varietal"]; ok {
			var tmp21 string
			if val, ok := v.(string); ok {
				tmp21 = val
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Varietal`, v, "string", err)
			}
			if err == nil {
				if len(tmp21) < 4 {
					err = goa.InvalidLengthError(`payload.Varietal`, tmp21, 4, true, err)
				}
			}
			p.Varietal = tmp21
		}
		if v, ok := val["vineyard"]; ok {
			var tmp22 string
			if val, ok := v.(string); ok {
				tmp22 = val
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Vineyard`, v, "string", err)
			}
			if err == nil {
				if len(tmp22) < 2 {
					err = goa.InvalidLengthError(`payload.Vineyard`, tmp22, 2, true, err)
				}
			}
			p.Vineyard = tmp22
		}
		if v, ok := val["vintage"]; ok {
			var tmp23 int
			if f, ok := v.(float64); ok {
				tmp23 = int(f)
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Vintage`, v, "int", err)
			}
			if err == nil {
				if tmp23 < 1900 {
					err = goa.InvalidRangeError(`payload.Vintage`, tmp23, 1900, true, err)
				}
				if tmp23 > 2020 {
					err = goa.InvalidRangeError(`payload.Vintage`, tmp23, 2020, false, err)
				}
			}
			p.Vintage = tmp23
		}
	} else {
		err = goa.InvalidAttributeTypeError(`payload`, raw, "map[string]interface{}", err)
	}
	return p, err
}

// NoContent sends a HTTP response with status code 204.
func (ctx *UpdateBottleContext) NoContent() error {
	return ctx.Respond(204, nil)
}

// NotFound sends a HTTP response with status code 404.
func (ctx *UpdateBottleContext) NotFound() error {
	return ctx.Respond(404, nil)
}
