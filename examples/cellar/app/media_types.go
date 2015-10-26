//************************************************************************//
// cellar: Application Media Types
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

	"github.com/raphael/goa"
)

// A tenant account
// Identifier: application/vnd.goa.example.account
type Account struct {
	// Date of creation
	CreatedAt string
	// Email of account ownder
	CreatedBy string
	// API href of account
	Href string
	// ID of account
	ID int
	// Name of account
	Name string
}

// Account views
type AccountViewEnum string

const (
	// Account default view
	AccountDefaultView AccountViewEnum = "default"
	// Account full view
	AccountFullView AccountViewEnum = "full"
	// Account link view
	AccountLinkView AccountViewEnum = "link"
)

// LoadAccount loads raw data into an instance of Account running all the
// validations. Raw data is defined by data that the JSON unmarshaler would create when unmarshaling
// into a variable of type interface{}. See https://golang.org/pkg/encoding/json/#Unmarshal for the
// complete list of supported data types.
func LoadAccount(raw interface{}) (*Account, error) {
	var err error
	var res *Account
	if val, ok := raw.(map[string]interface{}); ok {
		res = new(Account)
		if v, ok := val["created_at"]; ok {
			var tmp24 string
			if val, ok := v.(string); ok {
				tmp24 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.CreatedAt`, v, "string", err)
			}
			if err2 := goa.ValidateFormat(goa.FormatDateTime, tmp24); err2 != nil {
				err = goa.InvalidFormatError(`.CreatedAt`, tmp24, goa.FormatDateTime, err2, err)
			}
			res.CreatedAt = tmp24
		}
		if v, ok := val["created_by"]; ok {
			var tmp25 string
			if val, ok := v.(string); ok {
				tmp25 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.CreatedBy`, v, "string", err)
			}
			if err2 := goa.ValidateFormat(goa.FormatEmail, tmp25); err2 != nil {
				err = goa.InvalidFormatError(`.CreatedBy`, tmp25, goa.FormatEmail, err2, err)
			}
			res.CreatedBy = tmp25
		}
		if v, ok := val["href"]; ok {
			var tmp26 string
			if val, ok := v.(string); ok {
				tmp26 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Href`, v, "string", err)
			}
			res.Href = tmp26
		}
		if v, ok := val["id"]; ok {
			var tmp27 int
			if val, ok := v.(int); ok {
				tmp27 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.ID`, v, "int", err)
			}
			res.ID = tmp27
		}
		if v, ok := val["name"]; ok {
			var tmp28 string
			if val, ok := v.(string); ok {
				tmp28 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Name`, v, "string", err)
			}
			res.Name = tmp28
		} else {
			err = goa.MissingAttributeError(``, "name", err)
		}
	} else {
		err = goa.InvalidAttributeTypeError(``, raw, "map[string]interface{}", err)
	}
	return res, err
}

// Dump produces raw data from an instance of Account running all the
// validations. See LoadAccount for the definition of raw data.
func (mt *Account) Dump(view AccountViewEnum) (map[string]interface{}, error) {
	var err error
	var res map[string]interface{}
	if view == AccountDefaultView {
		if mt.Name == "" {
			err = fmt.Errorf("missing required attribute \"name\"")
		}
		if err == nil {
			tmp29 := map[string]interface{}{
				"href": mt.Href,
				"id":   mt.ID,
				"name": mt.Name,
			}
			res = tmp29
		}
	}
	if view == AccountFullView {
		if mt.Name == "" {
			err = fmt.Errorf("missing required attribute \"name\"")
		}
		if err2 := goa.ValidateFormat(goa.FormatDateTime, mt.CreatedAt); err2 != nil {
			err = goa.InvalidFormatError(`.CreatedAt`, mt.CreatedAt, goa.FormatDateTime, err2, err)
		}
		if err2 := goa.ValidateFormat(goa.FormatEmail, mt.CreatedBy); err2 != nil {
			err = goa.InvalidFormatError(`.CreatedBy`, mt.CreatedBy, goa.FormatEmail, err2, err)
		}
		if err == nil {
			tmp30 := map[string]interface{}{
				"created_at": mt.CreatedAt,
				"created_by": mt.CreatedBy,
				"href":       mt.Href,
				"id":         mt.ID,
				"name":       mt.Name,
			}
			res = tmp30
		}
	}
	if view == AccountLinkView {
		if mt.Name == "" {
			err = fmt.Errorf("missing required attribute \"name\"")
		}
		if err == nil {
			tmp31 := map[string]interface{}{
				"href": mt.Href,
				"name": mt.Name,
			}
			res = tmp31
		}
	}
	return res, err
}

// Validate validates the media type instance.
func (mt *Account) Validate() (err error) {
	if err2 := goa.ValidateFormat(goa.FormatDateTime, mt.CreatedAt); err2 != nil {
		err = goa.InvalidFormatError(`.CreatedAt`, mt.CreatedAt, goa.FormatDateTime, err2, err)
	}
	if err2 := goa.ValidateFormat(goa.FormatEmail, mt.CreatedBy); err2 != nil {
		err = goa.InvalidFormatError(`.CreatedBy`, mt.CreatedBy, goa.FormatEmail, err2, err)
	}
	return
}

// A bottle of wine
// Identifier: application/vnd.goa.example.bottle
type Bottle struct {
	// Account that owns bottle
	Account         *Account
	Characteristics string
	Color           string
	Country         string
	// Date of creation
	CreatedAt string
	// API href of bottle
	Href string
	// ID of bottle
	ID   int
	Name string
	// Rating of bottle between 1 and 5
	Rating    int
	Region    string
	Review    string
	Sweetness int
	// Date of last update
	UpdatedAt string
	Varietal  string
	Vineyard  string
	Vintage   int
}

// Bottle views
type BottleViewEnum string

const (
	// Bottle default view
	BottleDefaultView BottleViewEnum = "default"
	// Bottle full view
	BottleFullView BottleViewEnum = "full"
	// Bottle tiny view
	BottleTinyView BottleViewEnum = "tiny"
)

// LoadBottle loads raw data into an instance of Bottle running all the
// validations. Raw data is defined by data that the JSON unmarshaler would create when unmarshaling
// into a variable of type interface{}. See https://golang.org/pkg/encoding/json/#Unmarshal for the
// complete list of supported data types.
func LoadBottle(raw interface{}) (*Bottle, error) {
	var err error
	var res *Bottle
	if val, ok := raw.(map[string]interface{}); ok {
		res = new(Bottle)
		if v, ok := val["account"]; ok {
			var tmp32 *Account
			if val, ok := v.(map[string]interface{}); ok {
				tmp32 = new(Account)
				if v, ok := val["created_at"]; ok {
					var tmp33 string
					if val, ok := v.(string); ok {
						tmp33 = val
					} else {
						err = goa.InvalidAttributeTypeError(`.Account.CreatedAt`, v, "string", err)
					}
					if err2 := goa.ValidateFormat(goa.FormatDateTime, tmp33); err2 != nil {
						err = goa.InvalidFormatError(`.Account.CreatedAt`, tmp33, goa.FormatDateTime, err2, err)
					}
					tmp32.CreatedAt = tmp33
				}
				if v, ok := val["created_by"]; ok {
					var tmp34 string
					if val, ok := v.(string); ok {
						tmp34 = val
					} else {
						err = goa.InvalidAttributeTypeError(`.Account.CreatedBy`, v, "string", err)
					}
					if err2 := goa.ValidateFormat(goa.FormatEmail, tmp34); err2 != nil {
						err = goa.InvalidFormatError(`.Account.CreatedBy`, tmp34, goa.FormatEmail, err2, err)
					}
					tmp32.CreatedBy = tmp34
				}
				if v, ok := val["href"]; ok {
					var tmp35 string
					if val, ok := v.(string); ok {
						tmp35 = val
					} else {
						err = goa.InvalidAttributeTypeError(`.Account.Href`, v, "string", err)
					}
					tmp32.Href = tmp35
				}
				if v, ok := val["id"]; ok {
					var tmp36 int
					if val, ok := v.(int); ok {
						tmp36 = val
					} else {
						err = goa.InvalidAttributeTypeError(`.Account.ID`, v, "int", err)
					}
					tmp32.ID = tmp36
				}
				if v, ok := val["name"]; ok {
					var tmp37 string
					if val, ok := v.(string); ok {
						tmp37 = val
					} else {
						err = goa.InvalidAttributeTypeError(`.Account.Name`, v, "string", err)
					}
					tmp32.Name = tmp37
				} else {
					err = goa.MissingAttributeError(`.Account`, "name", err)
				}
			} else {
				err = goa.InvalidAttributeTypeError(`.Account`, v, "map[string]interface{}", err)
			}
			if err2 := goa.ValidateFormat(goa.FormatDateTime, tmp32.CreatedAt); err2 != nil {
				err = goa.InvalidFormatError(`.Account.CreatedAt`, tmp32.CreatedAt, goa.FormatDateTime, err2, err)
			}
			if err2 := goa.ValidateFormat(goa.FormatEmail, tmp32.CreatedBy); err2 != nil {
				err = goa.InvalidFormatError(`.Account.CreatedBy`, tmp32.CreatedBy, goa.FormatEmail, err2, err)
			}
			res.Account = tmp32
		} else {
			err = goa.MissingAttributeError(``, "account", err)
		}
		if v, ok := val["characteristics"]; ok {
			var tmp38 string
			if val, ok := v.(string); ok {
				tmp38 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Characteristics`, v, "string", err)
			}
			if len(tmp38) < 10 {
				err = goa.InvalidLengthError(`.Characteristics`, tmp38, 10, true, err)
			}
			if len(tmp38) > 300 {
				err = goa.InvalidLengthError(`.Characteristics`, tmp38, 300, false, err)
			}
			res.Characteristics = tmp38
		}
		if v, ok := val["color"]; ok {
			var tmp39 string
			if val, ok := v.(string); ok {
				tmp39 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Color`, v, "string", err)
			}
			if !(tmp39 == "red" || tmp39 == "white" || tmp39 == "rose" || tmp39 == "yellow" || tmp39 == "sparkling") {
				err = goa.InvalidEnumValueError(`.Color`, tmp39, []interface{}{"red", "white", "rose", "yellow", "sparkling"}, err)
			}
			res.Color = tmp39
		}
		if v, ok := val["country"]; ok {
			var tmp40 string
			if val, ok := v.(string); ok {
				tmp40 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Country`, v, "string", err)
			}
			if len(tmp40) < 2 {
				err = goa.InvalidLengthError(`.Country`, tmp40, 2, true, err)
			}
			res.Country = tmp40
		}
		if v, ok := val["created_at"]; ok {
			var tmp41 string
			if val, ok := v.(string); ok {
				tmp41 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.CreatedAt`, v, "string", err)
			}
			if err2 := goa.ValidateFormat(goa.FormatDateTime, tmp41); err2 != nil {
				err = goa.InvalidFormatError(`.CreatedAt`, tmp41, goa.FormatDateTime, err2, err)
			}
			res.CreatedAt = tmp41
		}
		if v, ok := val["href"]; ok {
			var tmp42 string
			if val, ok := v.(string); ok {
				tmp42 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Href`, v, "string", err)
			}
			res.Href = tmp42
		}
		if v, ok := val["id"]; ok {
			var tmp43 int
			if val, ok := v.(int); ok {
				tmp43 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.ID`, v, "int", err)
			}
			res.ID = tmp43
		}
		if v, ok := val["name"]; ok {
			var tmp44 string
			if val, ok := v.(string); ok {
				tmp44 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Name`, v, "string", err)
			}
			if len(tmp44) < 2 {
				err = goa.InvalidLengthError(`.Name`, tmp44, 2, true, err)
			}
			res.Name = tmp44
		} else {
			err = goa.MissingAttributeError(``, "name", err)
		}
		if v, ok := val["rating"]; ok {
			var tmp45 int
			if val, ok := v.(int); ok {
				tmp45 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Rating`, v, "int", err)
			}
			if tmp45 < 1 {
				err = goa.InvalidRangeError(`.Rating`, tmp45, 1, true, err)
			}
			if tmp45 > 5 {
				err = goa.InvalidRangeError(`.Rating`, tmp45, 5, false, err)
			}
			res.Rating = tmp45
		}
		if v, ok := val["region"]; ok {
			var tmp46 string
			if val, ok := v.(string); ok {
				tmp46 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Region`, v, "string", err)
			}
			res.Region = tmp46
		}
		if v, ok := val["review"]; ok {
			var tmp47 string
			if val, ok := v.(string); ok {
				tmp47 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Review`, v, "string", err)
			}
			if len(tmp47) < 10 {
				err = goa.InvalidLengthError(`.Review`, tmp47, 10, true, err)
			}
			if len(tmp47) > 300 {
				err = goa.InvalidLengthError(`.Review`, tmp47, 300, false, err)
			}
			res.Review = tmp47
		}
		if v, ok := val["sweetness"]; ok {
			var tmp48 int
			if val, ok := v.(int); ok {
				tmp48 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Sweetness`, v, "int", err)
			}
			if tmp48 < 1 {
				err = goa.InvalidRangeError(`.Sweetness`, tmp48, 1, true, err)
			}
			if tmp48 > 5 {
				err = goa.InvalidRangeError(`.Sweetness`, tmp48, 5, false, err)
			}
			res.Sweetness = tmp48
		}
		if v, ok := val["updated_at"]; ok {
			var tmp49 string
			if val, ok := v.(string); ok {
				tmp49 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.UpdatedAt`, v, "string", err)
			}
			if err2 := goa.ValidateFormat(goa.FormatDateTime, tmp49); err2 != nil {
				err = goa.InvalidFormatError(`.UpdatedAt`, tmp49, goa.FormatDateTime, err2, err)
			}
			res.UpdatedAt = tmp49
		}
		if v, ok := val["varietal"]; ok {
			var tmp50 string
			if val, ok := v.(string); ok {
				tmp50 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Varietal`, v, "string", err)
			}
			if len(tmp50) < 4 {
				err = goa.InvalidLengthError(`.Varietal`, tmp50, 4, true, err)
			}
			res.Varietal = tmp50
		}
		if v, ok := val["vineyard"]; ok {
			var tmp51 string
			if val, ok := v.(string); ok {
				tmp51 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Vineyard`, v, "string", err)
			}
			if len(tmp51) < 2 {
				err = goa.InvalidLengthError(`.Vineyard`, tmp51, 2, true, err)
			}
			res.Vineyard = tmp51
		} else {
			err = goa.MissingAttributeError(``, "vineyard", err)
		}
		if v, ok := val["vintage"]; ok {
			var tmp52 int
			if val, ok := v.(int); ok {
				tmp52 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Vintage`, v, "int", err)
			}
			if tmp52 < 1900 {
				err = goa.InvalidRangeError(`.Vintage`, tmp52, 1900, true, err)
			}
			if tmp52 > 2020 {
				err = goa.InvalidRangeError(`.Vintage`, tmp52, 2020, false, err)
			}
			res.Vintage = tmp52
		}
	} else {
		err = goa.InvalidAttributeTypeError(``, raw, "map[string]interface{}", err)
	}
	return res, err
}

// Dump produces raw data from an instance of Bottle running all the
// validations. See LoadBottle for the definition of raw data.
func (mt *Bottle) Dump(view BottleViewEnum) (map[string]interface{}, error) {
	var err error
	var res map[string]interface{}
	if view == BottleDefaultView {
		if mt.Name == "" {
			err = fmt.Errorf("missing required attribute \"name\"")
		}
		if len(mt.Name) < 2 {
			err = goa.InvalidLengthError(`.Name`, mt.Name, 2, true, err)
		}
		if len(mt.Varietal) < 4 {
			err = goa.InvalidLengthError(`.Varietal`, mt.Varietal, 4, true, err)
		}
		if len(mt.Vineyard) < 2 {
			err = goa.InvalidLengthError(`.Vineyard`, mt.Vineyard, 2, true, err)
		}
		if mt.Vintage < 1900 {
			err = goa.InvalidRangeError(`.Vintage`, mt.Vintage, 1900, true, err)
		}
		if mt.Vintage > 2020 {
			err = goa.InvalidRangeError(`.Vintage`, mt.Vintage, 2020, false, err)
		}
		if err == nil {
			tmp54 := map[string]interface{}{
				"href":     mt.Href,
				"id":       mt.ID,
				"name":     mt.Name,
				"varietal": mt.Varietal,
				"vineyard": mt.Vineyard,
				"vintage":  mt.Vintage,
			}
			res = tmp54
		}
		links := make(map[string]interface{})
		if mt.Account.Name == "" {
			err = fmt.Errorf("missing required attribute \"name\"")
		}
		if err == nil {
			tmp53 := map[string]interface{}{
				"href": mt.Account.Href,
				"name": mt.Account.Name,
			}
			links["Account"] = tmp53
		}
		res["links"] = links

	}
	if view == BottleFullView {
		if mt.Account == nil {
			err = fmt.Errorf("missing required attribute \"account\"")
		}
		if err2 := goa.ValidateFormat(goa.FormatDateTime, mt.Account.CreatedAt); err2 != nil {
			err = goa.InvalidFormatError(`.Account.CreatedAt`, mt.Account.CreatedAt, goa.FormatDateTime, err2, err)
		}
		if err2 := goa.ValidateFormat(goa.FormatEmail, mt.Account.CreatedBy); err2 != nil {
			err = goa.InvalidFormatError(`.Account.CreatedBy`, mt.Account.CreatedBy, goa.FormatEmail, err2, err)
		}
		if len(mt.Characteristics) < 10 {
			err = goa.InvalidLengthError(`.Characteristics`, mt.Characteristics, 10, true, err)
		}
		if len(mt.Characteristics) > 300 {
			err = goa.InvalidLengthError(`.Characteristics`, mt.Characteristics, 300, false, err)
		}
		if !(mt.Color == "red" || mt.Color == "white" || mt.Color == "rose" || mt.Color == "yellow" || mt.Color == "sparkling") {
			err = goa.InvalidEnumValueError(`.Color`, mt.Color, []interface{}{"red", "white", "rose", "yellow", "sparkling"}, err)
		}
		if len(mt.Country) < 2 {
			err = goa.InvalidLengthError(`.Country`, mt.Country, 2, true, err)
		}
		if err2 := goa.ValidateFormat(goa.FormatDateTime, mt.CreatedAt); err2 != nil {
			err = goa.InvalidFormatError(`.CreatedAt`, mt.CreatedAt, goa.FormatDateTime, err2, err)
		}
		if len(mt.Name) < 2 {
			err = goa.InvalidLengthError(`.Name`, mt.Name, 2, true, err)
		}
		if len(mt.Review) < 10 {
			err = goa.InvalidLengthError(`.Review`, mt.Review, 10, true, err)
		}
		if len(mt.Review) > 300 {
			err = goa.InvalidLengthError(`.Review`, mt.Review, 300, false, err)
		}
		if mt.Sweetness < 1 {
			err = goa.InvalidRangeError(`.Sweetness`, mt.Sweetness, 1, true, err)
		}
		if mt.Sweetness > 5 {
			err = goa.InvalidRangeError(`.Sweetness`, mt.Sweetness, 5, false, err)
		}
		if err2 := goa.ValidateFormat(goa.FormatDateTime, mt.UpdatedAt); err2 != nil {
			err = goa.InvalidFormatError(`.UpdatedAt`, mt.UpdatedAt, goa.FormatDateTime, err2, err)
		}
		if len(mt.Varietal) < 4 {
			err = goa.InvalidLengthError(`.Varietal`, mt.Varietal, 4, true, err)
		}
		if len(mt.Vineyard) < 2 {
			err = goa.InvalidLengthError(`.Vineyard`, mt.Vineyard, 2, true, err)
		}
		if mt.Vintage < 1900 {
			err = goa.InvalidRangeError(`.Vintage`, mt.Vintage, 1900, true, err)
		}
		if mt.Vintage > 2020 {
			err = goa.InvalidRangeError(`.Vintage`, mt.Vintage, 2020, false, err)
		}
		if err == nil {
			tmp55 := map[string]interface{}{
				"characteristics": mt.Characteristics,
				"color":           mt.Color,
				"country":         mt.Country,
				"created_at":      mt.CreatedAt,
				"href":            mt.Href,
				"id":              mt.ID,
				"name":            mt.Name,
				"region":          mt.Region,
				"review":          mt.Review,
				"sweetness":       mt.Sweetness,
				"updated_at":      mt.UpdatedAt,
				"varietal":        mt.Varietal,
				"vineyard":        mt.Vineyard,
				"vintage":         mt.Vintage,
			}
			if mt.Account != nil {
				if mt.Account.Name == "" {
					err = fmt.Errorf("missing required attribute \"name\"")
				}
				if err == nil {
					tmp56 := map[string]interface{}{
						"href": mt.Account.Href,
						"id":   mt.Account.ID,
						"name": mt.Account.Name,
					}
					tmp55["account"] = tmp56
				}
			}
			res = tmp55
		}
	}
	if view == BottleTinyView {
		if mt.Name == "" {
			err = fmt.Errorf("missing required attribute \"name\"")
		}
		if len(mt.Name) < 2 {
			err = goa.InvalidLengthError(`.Name`, mt.Name, 2, true, err)
		}
		if err == nil {
			tmp58 := map[string]interface{}{
				"href": mt.Href,
				"id":   mt.ID,
				"name": mt.Name,
			}
			res = tmp58
		}
		links := make(map[string]interface{})
		if mt.Account.Name == "" {
			err = fmt.Errorf("missing required attribute \"name\"")
		}
		if err == nil {
			tmp57 := map[string]interface{}{
				"href": mt.Account.Href,
				"name": mt.Account.Name,
			}
			links["Account"] = tmp57
		}
		res["links"] = links

	}
	return res, err
}

// Validate validates the media type instance.
func (mt *Bottle) Validate() (err error) {
	if err2 := goa.ValidateFormat(goa.FormatDateTime, mt.Account.CreatedAt); err2 != nil {
		err = goa.InvalidFormatError(`.Account.CreatedAt`, mt.Account.CreatedAt, goa.FormatDateTime, err2, err)
	}
	if err2 := goa.ValidateFormat(goa.FormatEmail, mt.Account.CreatedBy); err2 != nil {
		err = goa.InvalidFormatError(`.Account.CreatedBy`, mt.Account.CreatedBy, goa.FormatEmail, err2, err)
	}
	if len(mt.Characteristics) < 10 {
		err = goa.InvalidLengthError(`.Characteristics`, mt.Characteristics, 10, true, err)
	}
	if len(mt.Characteristics) > 300 {
		err = goa.InvalidLengthError(`.Characteristics`, mt.Characteristics, 300, false, err)
	}
	if !(mt.Color == "red" || mt.Color == "white" || mt.Color == "rose" || mt.Color == "yellow" || mt.Color == "sparkling") {
		err = goa.InvalidEnumValueError(`.Color`, mt.Color, []interface{}{"red", "white", "rose", "yellow", "sparkling"}, err)
	}
	if len(mt.Country) < 2 {
		err = goa.InvalidLengthError(`.Country`, mt.Country, 2, true, err)
	}
	if err2 := goa.ValidateFormat(goa.FormatDateTime, mt.CreatedAt); err2 != nil {
		err = goa.InvalidFormatError(`.CreatedAt`, mt.CreatedAt, goa.FormatDateTime, err2, err)
	}
	if len(mt.Name) < 2 {
		err = goa.InvalidLengthError(`.Name`, mt.Name, 2, true, err)
	}
	if mt.Rating < 1 {
		err = goa.InvalidRangeError(`.Rating`, mt.Rating, 1, true, err)
	}
	if mt.Rating > 5 {
		err = goa.InvalidRangeError(`.Rating`, mt.Rating, 5, false, err)
	}
	if len(mt.Review) < 10 {
		err = goa.InvalidLengthError(`.Review`, mt.Review, 10, true, err)
	}
	if len(mt.Review) > 300 {
		err = goa.InvalidLengthError(`.Review`, mt.Review, 300, false, err)
	}
	if mt.Sweetness < 1 {
		err = goa.InvalidRangeError(`.Sweetness`, mt.Sweetness, 1, true, err)
	}
	if mt.Sweetness > 5 {
		err = goa.InvalidRangeError(`.Sweetness`, mt.Sweetness, 5, false, err)
	}
	if err2 := goa.ValidateFormat(goa.FormatDateTime, mt.UpdatedAt); err2 != nil {
		err = goa.InvalidFormatError(`.UpdatedAt`, mt.UpdatedAt, goa.FormatDateTime, err2, err)
	}
	if len(mt.Varietal) < 4 {
		err = goa.InvalidLengthError(`.Varietal`, mt.Varietal, 4, true, err)
	}
	if len(mt.Vineyard) < 2 {
		err = goa.InvalidLengthError(`.Vineyard`, mt.Vineyard, 2, true, err)
	}
	if mt.Vintage < 1900 {
		err = goa.InvalidRangeError(`.Vintage`, mt.Vintage, 1900, true, err)
	}
	if mt.Vintage > 2020 {
		err = goa.InvalidRangeError(`.Vintage`, mt.Vintage, 2020, false, err)
	}
	return
}

// BottleCollection media type
// Identifier: application/vnd.goa.example.bottle; type=collection
type BottleCollection []*Bottle

// BottleCollection views
type BottleCollectionViewEnum string

const (
	// BottleCollection default view
	BottleCollectionDefaultView BottleCollectionViewEnum = "default"
	// BottleCollection tiny view
	BottleCollectionTinyView BottleCollectionViewEnum = "tiny"
)

// LoadBottleCollection loads raw data into an instance of BottleCollection running all the
// validations. Raw data is defined by data that the JSON unmarshaler would create when unmarshaling
// into a variable of type interface{}. See https://golang.org/pkg/encoding/json/#Unmarshal for the
// complete list of supported data types.
func LoadBottleCollection(raw interface{}) (BottleCollection, error) {
	var err error
	var res BottleCollection
	if val, ok := raw.([]interface{}); ok {
		res = make([]*Bottle, len(val))
		for i, v := range val {
			var tmp59 *Bottle
			if val, ok := v.(map[string]interface{}); ok {
				tmp59 = new(Bottle)
				if v, ok := val["account"]; ok {
					var tmp60 *Account
					if val, ok := v.(map[string]interface{}); ok {
						tmp60 = new(Account)
						if v, ok := val["created_at"]; ok {
							var tmp61 string
							if val, ok := v.(string); ok {
								tmp61 = val
							} else {
								err = goa.InvalidAttributeTypeError(`[*].Account.CreatedAt`, v, "string", err)
							}
							if err2 := goa.ValidateFormat(goa.FormatDateTime, tmp61); err2 != nil {
								err = goa.InvalidFormatError(`[*].Account.CreatedAt`, tmp61, goa.FormatDateTime, err2, err)
							}
							tmp60.CreatedAt = tmp61
						}
						if v, ok := val["created_by"]; ok {
							var tmp62 string
							if val, ok := v.(string); ok {
								tmp62 = val
							} else {
								err = goa.InvalidAttributeTypeError(`[*].Account.CreatedBy`, v, "string", err)
							}
							if err2 := goa.ValidateFormat(goa.FormatEmail, tmp62); err2 != nil {
								err = goa.InvalidFormatError(`[*].Account.CreatedBy`, tmp62, goa.FormatEmail, err2, err)
							}
							tmp60.CreatedBy = tmp62
						}
						if v, ok := val["href"]; ok {
							var tmp63 string
							if val, ok := v.(string); ok {
								tmp63 = val
							} else {
								err = goa.InvalidAttributeTypeError(`[*].Account.Href`, v, "string", err)
							}
							tmp60.Href = tmp63
						}
						if v, ok := val["id"]; ok {
							var tmp64 int
							if val, ok := v.(int); ok {
								tmp64 = val
							} else {
								err = goa.InvalidAttributeTypeError(`[*].Account.ID`, v, "int", err)
							}
							tmp60.ID = tmp64
						}
						if v, ok := val["name"]; ok {
							var tmp65 string
							if val, ok := v.(string); ok {
								tmp65 = val
							} else {
								err = goa.InvalidAttributeTypeError(`[*].Account.Name`, v, "string", err)
							}
							tmp60.Name = tmp65
						} else {
							err = goa.MissingAttributeError(`[*].Account`, "name", err)
						}
					} else {
						err = goa.InvalidAttributeTypeError(`[*].Account`, v, "map[string]interface{}", err)
					}
					if err2 := goa.ValidateFormat(goa.FormatDateTime, tmp60.CreatedAt); err2 != nil {
						err = goa.InvalidFormatError(`[*].Account.CreatedAt`, tmp60.CreatedAt, goa.FormatDateTime, err2, err)
					}
					if err2 := goa.ValidateFormat(goa.FormatEmail, tmp60.CreatedBy); err2 != nil {
						err = goa.InvalidFormatError(`[*].Account.CreatedBy`, tmp60.CreatedBy, goa.FormatEmail, err2, err)
					}
					tmp59.Account = tmp60
				} else {
					err = goa.MissingAttributeError(`[*]`, "account", err)
				}
				if v, ok := val["characteristics"]; ok {
					var tmp66 string
					if val, ok := v.(string); ok {
						tmp66 = val
					} else {
						err = goa.InvalidAttributeTypeError(`[*].Characteristics`, v, "string", err)
					}
					if len(tmp66) < 10 {
						err = goa.InvalidLengthError(`[*].Characteristics`, tmp66, 10, true, err)
					}
					if len(tmp66) > 300 {
						err = goa.InvalidLengthError(`[*].Characteristics`, tmp66, 300, false, err)
					}
					tmp59.Characteristics = tmp66
				}
				if v, ok := val["color"]; ok {
					var tmp67 string
					if val, ok := v.(string); ok {
						tmp67 = val
					} else {
						err = goa.InvalidAttributeTypeError(`[*].Color`, v, "string", err)
					}
					if !(tmp67 == "red" || tmp67 == "white" || tmp67 == "rose" || tmp67 == "yellow" || tmp67 == "sparkling") {
						err = goa.InvalidEnumValueError(`[*].Color`, tmp67, []interface{}{"red", "white", "rose", "yellow", "sparkling"}, err)
					}
					tmp59.Color = tmp67
				}
				if v, ok := val["country"]; ok {
					var tmp68 string
					if val, ok := v.(string); ok {
						tmp68 = val
					} else {
						err = goa.InvalidAttributeTypeError(`[*].Country`, v, "string", err)
					}
					if len(tmp68) < 2 {
						err = goa.InvalidLengthError(`[*].Country`, tmp68, 2, true, err)
					}
					tmp59.Country = tmp68
				}
				if v, ok := val["created_at"]; ok {
					var tmp69 string
					if val, ok := v.(string); ok {
						tmp69 = val
					} else {
						err = goa.InvalidAttributeTypeError(`[*].CreatedAt`, v, "string", err)
					}
					if err2 := goa.ValidateFormat(goa.FormatDateTime, tmp69); err2 != nil {
						err = goa.InvalidFormatError(`[*].CreatedAt`, tmp69, goa.FormatDateTime, err2, err)
					}
					tmp59.CreatedAt = tmp69
				}
				if v, ok := val["href"]; ok {
					var tmp70 string
					if val, ok := v.(string); ok {
						tmp70 = val
					} else {
						err = goa.InvalidAttributeTypeError(`[*].Href`, v, "string", err)
					}
					tmp59.Href = tmp70
				}
				if v, ok := val["id"]; ok {
					var tmp71 int
					if val, ok := v.(int); ok {
						tmp71 = val
					} else {
						err = goa.InvalidAttributeTypeError(`[*].ID`, v, "int", err)
					}
					tmp59.ID = tmp71
				}
				if v, ok := val["name"]; ok {
					var tmp72 string
					if val, ok := v.(string); ok {
						tmp72 = val
					} else {
						err = goa.InvalidAttributeTypeError(`[*].Name`, v, "string", err)
					}
					if len(tmp72) < 2 {
						err = goa.InvalidLengthError(`[*].Name`, tmp72, 2, true, err)
					}
					tmp59.Name = tmp72
				} else {
					err = goa.MissingAttributeError(`[*]`, "name", err)
				}
				if v, ok := val["rating"]; ok {
					var tmp73 int
					if val, ok := v.(int); ok {
						tmp73 = val
					} else {
						err = goa.InvalidAttributeTypeError(`[*].Rating`, v, "int", err)
					}
					if tmp73 < 1 {
						err = goa.InvalidRangeError(`[*].Rating`, tmp73, 1, true, err)
					}
					if tmp73 > 5 {
						err = goa.InvalidRangeError(`[*].Rating`, tmp73, 5, false, err)
					}
					tmp59.Rating = tmp73
				}
				if v, ok := val["region"]; ok {
					var tmp74 string
					if val, ok := v.(string); ok {
						tmp74 = val
					} else {
						err = goa.InvalidAttributeTypeError(`[*].Region`, v, "string", err)
					}
					tmp59.Region = tmp74
				}
				if v, ok := val["review"]; ok {
					var tmp75 string
					if val, ok := v.(string); ok {
						tmp75 = val
					} else {
						err = goa.InvalidAttributeTypeError(`[*].Review`, v, "string", err)
					}
					if len(tmp75) < 10 {
						err = goa.InvalidLengthError(`[*].Review`, tmp75, 10, true, err)
					}
					if len(tmp75) > 300 {
						err = goa.InvalidLengthError(`[*].Review`, tmp75, 300, false, err)
					}
					tmp59.Review = tmp75
				}
				if v, ok := val["sweetness"]; ok {
					var tmp76 int
					if val, ok := v.(int); ok {
						tmp76 = val
					} else {
						err = goa.InvalidAttributeTypeError(`[*].Sweetness`, v, "int", err)
					}
					if tmp76 < 1 {
						err = goa.InvalidRangeError(`[*].Sweetness`, tmp76, 1, true, err)
					}
					if tmp76 > 5 {
						err = goa.InvalidRangeError(`[*].Sweetness`, tmp76, 5, false, err)
					}
					tmp59.Sweetness = tmp76
				}
				if v, ok := val["updated_at"]; ok {
					var tmp77 string
					if val, ok := v.(string); ok {
						tmp77 = val
					} else {
						err = goa.InvalidAttributeTypeError(`[*].UpdatedAt`, v, "string", err)
					}
					if err2 := goa.ValidateFormat(goa.FormatDateTime, tmp77); err2 != nil {
						err = goa.InvalidFormatError(`[*].UpdatedAt`, tmp77, goa.FormatDateTime, err2, err)
					}
					tmp59.UpdatedAt = tmp77
				}
				if v, ok := val["varietal"]; ok {
					var tmp78 string
					if val, ok := v.(string); ok {
						tmp78 = val
					} else {
						err = goa.InvalidAttributeTypeError(`[*].Varietal`, v, "string", err)
					}
					if len(tmp78) < 4 {
						err = goa.InvalidLengthError(`[*].Varietal`, tmp78, 4, true, err)
					}
					tmp59.Varietal = tmp78
				}
				if v, ok := val["vineyard"]; ok {
					var tmp79 string
					if val, ok := v.(string); ok {
						tmp79 = val
					} else {
						err = goa.InvalidAttributeTypeError(`[*].Vineyard`, v, "string", err)
					}
					if len(tmp79) < 2 {
						err = goa.InvalidLengthError(`[*].Vineyard`, tmp79, 2, true, err)
					}
					tmp59.Vineyard = tmp79
				} else {
					err = goa.MissingAttributeError(`[*]`, "vineyard", err)
				}
				if v, ok := val["vintage"]; ok {
					var tmp80 int
					if val, ok := v.(int); ok {
						tmp80 = val
					} else {
						err = goa.InvalidAttributeTypeError(`[*].Vintage`, v, "int", err)
					}
					if tmp80 < 1900 {
						err = goa.InvalidRangeError(`[*].Vintage`, tmp80, 1900, true, err)
					}
					if tmp80 > 2020 {
						err = goa.InvalidRangeError(`[*].Vintage`, tmp80, 2020, false, err)
					}
					tmp59.Vintage = tmp80
				}
			} else {
				err = goa.InvalidAttributeTypeError(`[*]`, v, "map[string]interface{}", err)
			}
			if err2 := goa.ValidateFormat(goa.FormatDateTime, tmp59.Account.CreatedAt); err2 != nil {
				err = goa.InvalidFormatError(`[*].Account.CreatedAt`, tmp59.Account.CreatedAt, goa.FormatDateTime, err2, err)
			}
			if err2 := goa.ValidateFormat(goa.FormatEmail, tmp59.Account.CreatedBy); err2 != nil {
				err = goa.InvalidFormatError(`[*].Account.CreatedBy`, tmp59.Account.CreatedBy, goa.FormatEmail, err2, err)
			}
			if len(tmp59.Characteristics) < 10 {
				err = goa.InvalidLengthError(`[*].Characteristics`, tmp59.Characteristics, 10, true, err)
			}
			if len(tmp59.Characteristics) > 300 {
				err = goa.InvalidLengthError(`[*].Characteristics`, tmp59.Characteristics, 300, false, err)
			}
			if !(tmp59.Color == "red" || tmp59.Color == "white" || tmp59.Color == "rose" || tmp59.Color == "yellow" || tmp59.Color == "sparkling") {
				err = goa.InvalidEnumValueError(`[*].Color`, tmp59.Color, []interface{}{"red", "white", "rose", "yellow", "sparkling"}, err)
			}
			if len(tmp59.Country) < 2 {
				err = goa.InvalidLengthError(`[*].Country`, tmp59.Country, 2, true, err)
			}
			if err2 := goa.ValidateFormat(goa.FormatDateTime, tmp59.CreatedAt); err2 != nil {
				err = goa.InvalidFormatError(`[*].CreatedAt`, tmp59.CreatedAt, goa.FormatDateTime, err2, err)
			}
			if len(tmp59.Name) < 2 {
				err = goa.InvalidLengthError(`[*].Name`, tmp59.Name, 2, true, err)
			}
			if tmp59.Rating < 1 {
				err = goa.InvalidRangeError(`[*].Rating`, tmp59.Rating, 1, true, err)
			}
			if tmp59.Rating > 5 {
				err = goa.InvalidRangeError(`[*].Rating`, tmp59.Rating, 5, false, err)
			}
			if len(tmp59.Review) < 10 {
				err = goa.InvalidLengthError(`[*].Review`, tmp59.Review, 10, true, err)
			}
			if len(tmp59.Review) > 300 {
				err = goa.InvalidLengthError(`[*].Review`, tmp59.Review, 300, false, err)
			}
			if tmp59.Sweetness < 1 {
				err = goa.InvalidRangeError(`[*].Sweetness`, tmp59.Sweetness, 1, true, err)
			}
			if tmp59.Sweetness > 5 {
				err = goa.InvalidRangeError(`[*].Sweetness`, tmp59.Sweetness, 5, false, err)
			}
			if err2 := goa.ValidateFormat(goa.FormatDateTime, tmp59.UpdatedAt); err2 != nil {
				err = goa.InvalidFormatError(`[*].UpdatedAt`, tmp59.UpdatedAt, goa.FormatDateTime, err2, err)
			}
			if len(tmp59.Varietal) < 4 {
				err = goa.InvalidLengthError(`[*].Varietal`, tmp59.Varietal, 4, true, err)
			}
			if len(tmp59.Vineyard) < 2 {
				err = goa.InvalidLengthError(`[*].Vineyard`, tmp59.Vineyard, 2, true, err)
			}
			if tmp59.Vintage < 1900 {
				err = goa.InvalidRangeError(`[*].Vintage`, tmp59.Vintage, 1900, true, err)
			}
			if tmp59.Vintage > 2020 {
				err = goa.InvalidRangeError(`[*].Vintage`, tmp59.Vintage, 2020, false, err)
			}
			res[i] = tmp59
		}
	} else {
		err = goa.InvalidAttributeTypeError(``, raw, "[]interface{}", err)
	}
	return res, err
}

// Dump produces raw data from an instance of BottleCollection running all the
// validations. See LoadBottleCollection for the definition of raw data.
func (mt BottleCollection) Dump(view BottleCollectionViewEnum) ([]map[string]interface{}, error) {
	var err error
	var res []map[string]interface{}
	if view == BottleCollectionDefaultView {
		tmp81 := make([]map[string]interface{}, len(mt))
		for i, res := range mt {
			if res.Name == "" {
				err = fmt.Errorf("missing required attribute \"name\"")
			}
			if len(res.Name) < 2 {
				err = goa.InvalidLengthError(`[*].Name`, res.Name, 2, true, err)
			}
			if len(res.Varietal) < 4 {
				err = goa.InvalidLengthError(`[*].Varietal`, res.Varietal, 4, true, err)
			}
			if len(res.Vineyard) < 2 {
				err = goa.InvalidLengthError(`[*].Vineyard`, res.Vineyard, 2, true, err)
			}
			if res.Vintage < 1900 {
				err = goa.InvalidRangeError(`[*].Vintage`, res.Vintage, 1900, true, err)
			}
			if res.Vintage > 2020 {
				err = goa.InvalidRangeError(`[*].Vintage`, res.Vintage, 2020, false, err)
			}
			if err == nil {
				tmp83 := map[string]interface{}{
					"href":     res.Href,
					"id":       res.ID,
					"name":     res.Name,
					"varietal": res.Varietal,
					"vineyard": res.Vineyard,
					"vintage":  res.Vintage,
				}
				tmp81[i] = tmp83
			}
			links := make(map[string]interface{})
			if res.Account.Name == "" {
				err = fmt.Errorf("missing required attribute \"name\"")
			}
			if err == nil {
				tmp82 := map[string]interface{}{
					"href": res.Account.Href,
					"name": res.Account.Name,
				}
				links["Account"] = tmp82
			}
			tmp81[i]["links"] = links

		}
		res = tmp81
	}
	if view == BottleCollectionTinyView {
		tmp84 := make([]map[string]interface{}, len(mt))
		for i, res := range mt {
			if res.Name == "" {
				err = fmt.Errorf("missing required attribute \"name\"")
			}
			if len(res.Name) < 2 {
				err = goa.InvalidLengthError(`[*].Name`, res.Name, 2, true, err)
			}
			if err == nil {
				tmp86 := map[string]interface{}{
					"href": res.Href,
					"id":   res.ID,
					"name": res.Name,
				}
				tmp84[i] = tmp86
			}
			links := make(map[string]interface{})
			if res.Account.Name == "" {
				err = fmt.Errorf("missing required attribute \"name\"")
			}
			if err == nil {
				tmp85 := map[string]interface{}{
					"href": res.Account.Href,
					"name": res.Account.Name,
				}
				links["Account"] = tmp85
			}
			tmp84[i]["links"] = links

		}
		res = tmp84
	}
	return res, err
}

// Validate validates the media type instance.
func (mt BottleCollection) Validate() (err error) {
	return
}
