//************************************************************************//
// API "cellar": Application Contexts
//
// Generated with goagen v0.0.1, command line:
// $ goagen
// --out=$(GOPATH)/src/github.com/raphael/goa/examples/cellar
// --design=github.com/raphael/goa/examples/cellar/design
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
func NewCreateAccountPayload(raw interface{}) (p *CreateAccountPayload, err error) {
	p, err = UnmarshalCreateAccountPayload(raw, err)
	return
}

// UnmarshalCreateAccountPayload unmarshals and validates a raw interface{} into an instance of CreateAccountPayload
func UnmarshalCreateAccountPayload(source interface{}, inErr error) (target *CreateAccountPayload, err error) {
	err = inErr
	if val, ok := source.(map[string]interface{}); ok {
		target = new(CreateAccountPayload)
		if v, ok := val["name"]; ok {
			var tmp1 string
			if val, ok := v.(string); ok {
				tmp1 = val
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Name`, v, "string", err)
			}
			target.Name = tmp1
		} else {
			err = goa.MissingAttributeError(`payload`, "name", err)
		}
	} else {
		err = goa.InvalidAttributeTypeError(`payload`, source, "dictionary", err)
	}
	return
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
	rawAccountID := c.Get("accountID")
	if accountID, err2 := strconv.Atoi(rawAccountID); err2 == nil {
		ctx.AccountID = int(accountID)
	} else {
		err = goa.InvalidParamTypeError("accountID", rawAccountID, "integer", err)
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
	rawAccountID := c.Get("accountID")
	if accountID, err2 := strconv.Atoi(rawAccountID); err2 == nil {
		ctx.AccountID = int(accountID)
	} else {
		err = goa.InvalidParamTypeError("accountID", rawAccountID, "integer", err)
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
	ctx.Header().Set("Content-Type", "application/vnd.account+json; charset=utf-8")
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
	rawAccountID := c.Get("accountID")
	if accountID, err2 := strconv.Atoi(rawAccountID); err2 == nil {
		ctx.AccountID = int(accountID)
	} else {
		err = goa.InvalidParamTypeError("accountID", rawAccountID, "integer", err)
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
func NewUpdateAccountPayload(raw interface{}) (p *UpdateAccountPayload, err error) {
	p, err = UnmarshalUpdateAccountPayload(raw, err)
	return
}

// UnmarshalUpdateAccountPayload unmarshals and validates a raw interface{} into an instance of UpdateAccountPayload
func UnmarshalUpdateAccountPayload(source interface{}, inErr error) (target *UpdateAccountPayload, err error) {
	err = inErr
	if val, ok := source.(map[string]interface{}); ok {
		target = new(UpdateAccountPayload)
		if v, ok := val["name"]; ok {
			var tmp2 string
			if val, ok := v.(string); ok {
				tmp2 = val
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Name`, v, "string", err)
			}
			target.Name = tmp2
		} else {
			err = goa.MissingAttributeError(`payload`, "name", err)
		}
	} else {
		err = goa.InvalidAttributeTypeError(`payload`, source, "dictionary", err)
	}
	return
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
	rawAccountID := c.Get("accountID")
	if accountID, err2 := strconv.Atoi(rawAccountID); err2 == nil {
		ctx.AccountID = int(accountID)
	} else {
		err = goa.InvalidParamTypeError("accountID", rawAccountID, "integer", err)
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
	Color     string
	Country   string
	Name      string
	Region    string
	Review    string
	Sweetness int
	Varietal  string
	Vineyard  string
	Vintage   int
}

// NewCreateBottlePayload instantiates a CreateBottlePayload from a raw request body.
// It validates each field and returns an error if any validation fails.
func NewCreateBottlePayload(raw interface{}) (p *CreateBottlePayload, err error) {
	p, err = UnmarshalCreateBottlePayload(raw, err)
	return
}

// UnmarshalCreateBottlePayload unmarshals and validates a raw interface{} into an instance of CreateBottlePayload
func UnmarshalCreateBottlePayload(source interface{}, inErr error) (target *CreateBottlePayload, err error) {
	err = inErr
	if val, ok := source.(map[string]interface{}); ok {
		target = new(CreateBottlePayload)
		if v, ok := val["color"]; ok {
			var tmp3 string
			if val, ok := v.(string); ok {
				tmp3 = val
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Color`, v, "string", err)
			}
			if err == nil {
				if tmp3 != "" {
					if !(tmp3 == "red" || tmp3 == "white" || tmp3 == "rose" || tmp3 == "yellow" || tmp3 == "sparkling") {
						err = goa.InvalidEnumValueError(`payload.Color`, tmp3, []interface{}{"red", "white", "rose", "yellow", "sparkling"}, err)
					}
				}
			}
			target.Color = tmp3
		} else {
			err = goa.MissingAttributeError(`payload`, "color", err)
		}
		if v, ok := val["country"]; ok {
			var tmp4 string
			if val, ok := v.(string); ok {
				tmp4 = val
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Country`, v, "string", err)
			}
			if err == nil {
				if len(tmp4) < 2 {
					err = goa.InvalidLengthError(`payload.Country`, tmp4, len(tmp4), 2, true, err)
				}
			}
			target.Country = tmp4
		}
		if v, ok := val["name"]; ok {
			var tmp5 string
			if val, ok := v.(string); ok {
				tmp5 = val
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Name`, v, "string", err)
			}
			if err == nil {
				if len(tmp5) < 2 {
					err = goa.InvalidLengthError(`payload.Name`, tmp5, len(tmp5), 2, true, err)
				}
			}
			target.Name = tmp5
		} else {
			err = goa.MissingAttributeError(`payload`, "name", err)
		}
		if v, ok := val["region"]; ok {
			var tmp6 string
			if val, ok := v.(string); ok {
				tmp6 = val
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Region`, v, "string", err)
			}
			target.Region = tmp6
		}
		if v, ok := val["review"]; ok {
			var tmp7 string
			if val, ok := v.(string); ok {
				tmp7 = val
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Review`, v, "string", err)
			}
			if err == nil {
				if len(tmp7) < 10 {
					err = goa.InvalidLengthError(`payload.Review`, tmp7, len(tmp7), 10, true, err)
				}
				if len(tmp7) > 300 {
					err = goa.InvalidLengthError(`payload.Review`, tmp7, len(tmp7), 300, false, err)
				}
			}
			target.Review = tmp7
		}
		if v, ok := val["sweetness"]; ok {
			var tmp8 int
			if f, ok := v.(float64); ok {
				tmp8 = int(f)
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Sweetness`, v, "int", err)
			}
			if err == nil {
				if tmp8 < 1 {
					err = goa.InvalidRangeError(`payload.Sweetness`, tmp8, 1, true, err)
				}
				if tmp8 > 5 {
					err = goa.InvalidRangeError(`payload.Sweetness`, tmp8, 5, false, err)
				}
			}
			target.Sweetness = tmp8
		}
		if v, ok := val["varietal"]; ok {
			var tmp9 string
			if val, ok := v.(string); ok {
				tmp9 = val
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Varietal`, v, "string", err)
			}
			if err == nil {
				if len(tmp9) < 4 {
					err = goa.InvalidLengthError(`payload.Varietal`, tmp9, len(tmp9), 4, true, err)
				}
			}
			target.Varietal = tmp9
		} else {
			err = goa.MissingAttributeError(`payload`, "varietal", err)
		}
		if v, ok := val["vineyard"]; ok {
			var tmp10 string
			if val, ok := v.(string); ok {
				tmp10 = val
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Vineyard`, v, "string", err)
			}
			if err == nil {
				if len(tmp10) < 2 {
					err = goa.InvalidLengthError(`payload.Vineyard`, tmp10, len(tmp10), 2, true, err)
				}
			}
			target.Vineyard = tmp10
		} else {
			err = goa.MissingAttributeError(`payload`, "vineyard", err)
		}
		if v, ok := val["vintage"]; ok {
			var tmp11 int
			if f, ok := v.(float64); ok {
				tmp11 = int(f)
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Vintage`, v, "int", err)
			}
			if err == nil {
				if tmp11 < 1900 {
					err = goa.InvalidRangeError(`payload.Vintage`, tmp11, 1900, true, err)
				}
				if tmp11 > 2020 {
					err = goa.InvalidRangeError(`payload.Vintage`, tmp11, 2020, false, err)
				}
			}
			target.Vintage = tmp11
		} else {
			err = goa.MissingAttributeError(`payload`, "vintage", err)
		}
	} else {
		err = goa.InvalidAttributeTypeError(`payload`, source, "dictionary", err)
	}
	return
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
	rawAccountID := c.Get("accountID")
	if accountID, err2 := strconv.Atoi(rawAccountID); err2 == nil {
		ctx.AccountID = int(accountID)
	} else {
		err = goa.InvalidParamTypeError("accountID", rawAccountID, "integer", err)
	}
	rawBottleID := c.Get("bottleID")
	if bottleID, err2 := strconv.Atoi(rawBottleID); err2 == nil {
		ctx.BottleID = int(bottleID)
	} else {
		err = goa.InvalidParamTypeError("bottleID", rawBottleID, "integer", err)
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
	rawAccountID := c.Get("accountID")
	if accountID, err2 := strconv.Atoi(rawAccountID); err2 == nil {
		ctx.AccountID = int(accountID)
	} else {
		err = goa.InvalidParamTypeError("accountID", rawAccountID, "integer", err)
	}
	rawYears := c.Get("years")
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
	return &ctx, err
}

// NotFound sends a HTTP response with status code 404.
func (ctx *ListBottleContext) NotFound() error {
	return ctx.Respond(404, nil)
}

// OK sends a HTTP response with status code 200.
func (ctx *ListBottleContext) OK(resp BottleCollection, view BottleCollectionViewEnum) error {
	r, err := resp.Dump(view)
	if err != nil {
		return fmt.Errorf("invalid response: %s", err)
	}
	ctx.Header().Set("Content-Type", "application/vnd.bottle+json; type=collection; charset=utf-8")
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
	rawAccountID := c.Get("accountID")
	if accountID, err2 := strconv.Atoi(rawAccountID); err2 == nil {
		ctx.AccountID = int(accountID)
	} else {
		err = goa.InvalidParamTypeError("accountID", rawAccountID, "integer", err)
	}
	rawBottleID := c.Get("bottleID")
	if bottleID, err2 := strconv.Atoi(rawBottleID); err2 == nil {
		ctx.BottleID = int(bottleID)
	} else {
		err = goa.InvalidParamTypeError("bottleID", rawBottleID, "integer", err)
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
func NewRateBottlePayload(raw interface{}) (p *RateBottlePayload, err error) {
	p, err = UnmarshalRateBottlePayload(raw, err)
	return
}

// UnmarshalRateBottlePayload unmarshals and validates a raw interface{} into an instance of RateBottlePayload
func UnmarshalRateBottlePayload(source interface{}, inErr error) (target *RateBottlePayload, err error) {
	err = inErr
	if val, ok := source.(map[string]interface{}); ok {
		target = new(RateBottlePayload)
		if v, ok := val["rating"]; ok {
			var tmp12 int
			if f, ok := v.(float64); ok {
				tmp12 = int(f)
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Rating`, v, "int", err)
			}
			if err == nil {
				if tmp12 < 1 {
					err = goa.InvalidRangeError(`payload.Rating`, tmp12, 1, true, err)
				}
				if tmp12 > 5 {
					err = goa.InvalidRangeError(`payload.Rating`, tmp12, 5, false, err)
				}
			}
			target.Rating = tmp12
		} else {
			err = goa.MissingAttributeError(`payload`, "rating", err)
		}
	} else {
		err = goa.InvalidAttributeTypeError(`payload`, source, "dictionary", err)
	}
	return
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
	rawAccountID := c.Get("accountID")
	if accountID, err2 := strconv.Atoi(rawAccountID); err2 == nil {
		ctx.AccountID = int(accountID)
	} else {
		err = goa.InvalidParamTypeError("accountID", rawAccountID, "integer", err)
	}
	rawBottleID := c.Get("bottleID")
	if bottleID, err2 := strconv.Atoi(rawBottleID); err2 == nil {
		ctx.BottleID = int(bottleID)
	} else {
		err = goa.InvalidParamTypeError("bottleID", rawBottleID, "integer", err)
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
	ctx.Header().Set("Content-Type", "application/vnd.bottle+json; charset=utf-8")
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
	rawAccountID := c.Get("accountID")
	if accountID, err2 := strconv.Atoi(rawAccountID); err2 == nil {
		ctx.AccountID = int(accountID)
	} else {
		err = goa.InvalidParamTypeError("accountID", rawAccountID, "integer", err)
	}
	rawBottleID := c.Get("bottleID")
	if bottleID, err2 := strconv.Atoi(rawBottleID); err2 == nil {
		ctx.BottleID = int(bottleID)
	} else {
		err = goa.InvalidParamTypeError("bottleID", rawBottleID, "integer", err)
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
	Color     string
	Country   string
	Name      string
	Region    string
	Review    string
	Sweetness int
	Varietal  string
	Vineyard  string
	Vintage   int
}

// NewUpdateBottlePayload instantiates a UpdateBottlePayload from a raw request body.
// It validates each field and returns an error if any validation fails.
func NewUpdateBottlePayload(raw interface{}) (p *UpdateBottlePayload, err error) {
	p, err = UnmarshalUpdateBottlePayload(raw, err)
	return
}

// UnmarshalUpdateBottlePayload unmarshals and validates a raw interface{} into an instance of UpdateBottlePayload
func UnmarshalUpdateBottlePayload(source interface{}, inErr error) (target *UpdateBottlePayload, err error) {
	err = inErr
	if val, ok := source.(map[string]interface{}); ok {
		target = new(UpdateBottlePayload)
		if v, ok := val["color"]; ok {
			var tmp13 string
			if val, ok := v.(string); ok {
				tmp13 = val
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Color`, v, "string", err)
			}
			if err == nil {
				if tmp13 != "" {
					if !(tmp13 == "red" || tmp13 == "white" || tmp13 == "rose" || tmp13 == "yellow" || tmp13 == "sparkling") {
						err = goa.InvalidEnumValueError(`payload.Color`, tmp13, []interface{}{"red", "white", "rose", "yellow", "sparkling"}, err)
					}
				}
			}
			target.Color = tmp13
		}
		if v, ok := val["country"]; ok {
			var tmp14 string
			if val, ok := v.(string); ok {
				tmp14 = val
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Country`, v, "string", err)
			}
			if err == nil {
				if len(tmp14) < 2 {
					err = goa.InvalidLengthError(`payload.Country`, tmp14, len(tmp14), 2, true, err)
				}
			}
			target.Country = tmp14
		}
		if v, ok := val["name"]; ok {
			var tmp15 string
			if val, ok := v.(string); ok {
				tmp15 = val
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Name`, v, "string", err)
			}
			if err == nil {
				if len(tmp15) < 2 {
					err = goa.InvalidLengthError(`payload.Name`, tmp15, len(tmp15), 2, true, err)
				}
			}
			target.Name = tmp15
		}
		if v, ok := val["region"]; ok {
			var tmp16 string
			if val, ok := v.(string); ok {
				tmp16 = val
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Region`, v, "string", err)
			}
			target.Region = tmp16
		}
		if v, ok := val["review"]; ok {
			var tmp17 string
			if val, ok := v.(string); ok {
				tmp17 = val
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Review`, v, "string", err)
			}
			if err == nil {
				if len(tmp17) < 10 {
					err = goa.InvalidLengthError(`payload.Review`, tmp17, len(tmp17), 10, true, err)
				}
				if len(tmp17) > 300 {
					err = goa.InvalidLengthError(`payload.Review`, tmp17, len(tmp17), 300, false, err)
				}
			}
			target.Review = tmp17
		}
		if v, ok := val["sweetness"]; ok {
			var tmp18 int
			if f, ok := v.(float64); ok {
				tmp18 = int(f)
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Sweetness`, v, "int", err)
			}
			if err == nil {
				if tmp18 < 1 {
					err = goa.InvalidRangeError(`payload.Sweetness`, tmp18, 1, true, err)
				}
				if tmp18 > 5 {
					err = goa.InvalidRangeError(`payload.Sweetness`, tmp18, 5, false, err)
				}
			}
			target.Sweetness = tmp18
		}
		if v, ok := val["varietal"]; ok {
			var tmp19 string
			if val, ok := v.(string); ok {
				tmp19 = val
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Varietal`, v, "string", err)
			}
			if err == nil {
				if len(tmp19) < 4 {
					err = goa.InvalidLengthError(`payload.Varietal`, tmp19, len(tmp19), 4, true, err)
				}
			}
			target.Varietal = tmp19
		}
		if v, ok := val["vineyard"]; ok {
			var tmp20 string
			if val, ok := v.(string); ok {
				tmp20 = val
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Vineyard`, v, "string", err)
			}
			if err == nil {
				if len(tmp20) < 2 {
					err = goa.InvalidLengthError(`payload.Vineyard`, tmp20, len(tmp20), 2, true, err)
				}
			}
			target.Vineyard = tmp20
		}
		if v, ok := val["vintage"]; ok {
			var tmp21 int
			if f, ok := v.(float64); ok {
				tmp21 = int(f)
			} else {
				err = goa.InvalidAttributeTypeError(`payload.Vintage`, v, "int", err)
			}
			if err == nil {
				if tmp21 < 1900 {
					err = goa.InvalidRangeError(`payload.Vintage`, tmp21, 1900, true, err)
				}
				if tmp21 > 2020 {
					err = goa.InvalidRangeError(`payload.Vintage`, tmp21, 2020, false, err)
				}
			}
			target.Vintage = tmp21
		}
	} else {
		err = goa.InvalidAttributeTypeError(`payload`, source, "dictionary", err)
	}
	return
}

// NoContent sends a HTTP response with status code 204.
func (ctx *UpdateBottleContext) NoContent() error {
	return ctx.Respond(204, nil)
}

// NotFound sends a HTTP response with status code 404.
func (ctx *UpdateBottleContext) NotFound() error {
	return ctx.Respond(404, nil)
}
