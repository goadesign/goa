//************************************************************************//
// cellar: Application Media Types
//
// Generated with goagen v0.0.1, command line:
// $ goagen
// --out=.
// --design=github.com/raphael/goa/examples/cellar/design
// --pkg=app
//
// The content of this file is auto-generated, DO NOT MODIFY
//************************************************************************//

package app

import "github.com/raphael/goa"

// A tenant account
// Identifier: application/vnd.goa.example.account+json
type Account struct {
	// Date of creation
	CreatedAt string
	// Email of account owner
	CreatedBy string
	// API href of account
	Href string
	// ID of account
	ID int
	// Name of account
	Name string
}

// object views
type AccountViewEnum string

const (
	// Account default view
	AccountDefaultView AccountViewEnum = "default"
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
			var tmp22 string
			if val, ok := v.(string); ok {
				tmp22 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.CreatedAt`, v, "string", err)
			}
			if err == nil {
				if tmp22 != "" {
					if err2 := goa.ValidateFormat(goa.FormatDateTime, tmp22); err2 != nil {
						err = goa.InvalidFormatError(`.CreatedAt`, tmp22, goa.FormatDateTime, err2, err)
					}
				}
			}
			res.CreatedAt = tmp22
		}
		if v, ok := val["created_by"]; ok {
			var tmp23 string
			if val, ok := v.(string); ok {
				tmp23 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.CreatedBy`, v, "string", err)
			}
			if err == nil {
				if tmp23 != "" {
					if err2 := goa.ValidateFormat(goa.FormatEmail, tmp23); err2 != nil {
						err = goa.InvalidFormatError(`.CreatedBy`, tmp23, goa.FormatEmail, err2, err)
					}
				}
			}
			res.CreatedBy = tmp23
		}
		if v, ok := val["href"]; ok {
			var tmp24 string
			if val, ok := v.(string); ok {
				tmp24 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Href`, v, "string", err)
			}
			res.Href = tmp24
		}
		if v, ok := val["id"]; ok {
			var tmp25 int
			if f, ok := v.(float64); ok {
				tmp25 = int(f)
			} else {
				err = goa.InvalidAttributeTypeError(`.ID`, v, "int", err)
			}
			res.ID = tmp25
		}
		if v, ok := val["name"]; ok {
			var tmp26 string
			if val, ok := v.(string); ok {
				tmp26 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Name`, v, "string", err)
			}
			res.Name = tmp26
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
		if mt.CreatedAt != "" {
			if err2 := goa.ValidateFormat(goa.FormatDateTime, mt.CreatedAt); err2 != nil {
				err = goa.InvalidFormatError(`default view.created_at`, mt.CreatedAt, goa.FormatDateTime, err2, err)
			}
		}
		if mt.CreatedBy != "" {
			if err2 := goa.ValidateFormat(goa.FormatEmail, mt.CreatedBy); err2 != nil {
				err = goa.InvalidFormatError(`default view.created_by`, mt.CreatedBy, goa.FormatEmail, err2, err)
			}
		}
		tmp27 := map[string]interface{}{
			"created_at": mt.CreatedAt,
			"created_by": mt.CreatedBy,
			"href":       mt.Href,
			"id":         mt.ID,
			"name":       mt.Name,
		}
		res = tmp27
	}
	if view == AccountLinkView {
		tmp28 := map[string]interface{}{
			"href": mt.Href,
			"id":   mt.ID,
			"name": mt.Name,
		}
		res = tmp28
	}
	return res, err
}

// Validate validates the media type instance.
func (mt *Account) Validate() (err error) {
	if mt.CreatedAt != "" {
		if err2 := goa.ValidateFormat(goa.FormatDateTime, mt.CreatedAt); err2 != nil {
			err = goa.InvalidFormatError(`response.created_at`, mt.CreatedAt, goa.FormatDateTime, err2, err)
		}
	}
	if mt.CreatedBy != "" {
		if err2 := goa.ValidateFormat(goa.FormatEmail, mt.CreatedBy); err2 != nil {
			err = goa.InvalidFormatError(`response.created_by`, mt.CreatedBy, goa.FormatEmail, err2, err)
		}
	}
	return
}

// A bottle of wine
// Identifier: application/vnd.goa.example.bottle+json
type Bottle struct {
	// Account that owns bottle
	Account *Account
	Color   string
	Country string
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

// object views
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
			var tmp29 *Account
			if val, ok := v.(map[string]interface{}); ok {
				tmp29 = new(Account)
				if v, ok := val["created_at"]; ok {
					var tmp30 string
					if val, ok := v.(string); ok {
						tmp30 = val
					} else {
						err = goa.InvalidAttributeTypeError(`.Account.CreatedAt`, v, "string", err)
					}
					if err == nil {
						if tmp30 != "" {
							if err2 := goa.ValidateFormat(goa.FormatDateTime, tmp30); err2 != nil {
								err = goa.InvalidFormatError(`.Account.CreatedAt`, tmp30, goa.FormatDateTime, err2, err)
							}
						}
					}
					tmp29.CreatedAt = tmp30
				}
				if v, ok := val["created_by"]; ok {
					var tmp31 string
					if val, ok := v.(string); ok {
						tmp31 = val
					} else {
						err = goa.InvalidAttributeTypeError(`.Account.CreatedBy`, v, "string", err)
					}
					if err == nil {
						if tmp31 != "" {
							if err2 := goa.ValidateFormat(goa.FormatEmail, tmp31); err2 != nil {
								err = goa.InvalidFormatError(`.Account.CreatedBy`, tmp31, goa.FormatEmail, err2, err)
							}
						}
					}
					tmp29.CreatedBy = tmp31
				}
				if v, ok := val["href"]; ok {
					var tmp32 string
					if val, ok := v.(string); ok {
						tmp32 = val
					} else {
						err = goa.InvalidAttributeTypeError(`.Account.Href`, v, "string", err)
					}
					tmp29.Href = tmp32
				}
				if v, ok := val["id"]; ok {
					var tmp33 int
					if f, ok := v.(float64); ok {
						tmp33 = int(f)
					} else {
						err = goa.InvalidAttributeTypeError(`.Account.ID`, v, "int", err)
					}
					tmp29.ID = tmp33
				}
				if v, ok := val["name"]; ok {
					var tmp34 string
					if val, ok := v.(string); ok {
						tmp34 = val
					} else {
						err = goa.InvalidAttributeTypeError(`.Account.Name`, v, "string", err)
					}
					tmp29.Name = tmp34
				}
			} else {
				err = goa.InvalidAttributeTypeError(`.Account`, v, "map[string]interface{}", err)
			}
			res.Account = tmp29
		}
		if v, ok := val["color"]; ok {
			var tmp35 string
			if val, ok := v.(string); ok {
				tmp35 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Color`, v, "string", err)
			}
			if err == nil {
				if tmp35 != "" {
					if !(tmp35 == "red" || tmp35 == "white" || tmp35 == "rose" || tmp35 == "yellow" || tmp35 == "sparkling") {
						err = goa.InvalidEnumValueError(`.Color`, tmp35, []interface{}{"red", "white", "rose", "yellow", "sparkling"}, err)
					}
				}
			}
			res.Color = tmp35
		}
		if v, ok := val["country"]; ok {
			var tmp36 string
			if val, ok := v.(string); ok {
				tmp36 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Country`, v, "string", err)
			}
			if err == nil {
				if len(tmp36) < 2 {
					err = goa.InvalidLengthError(`.Country`, tmp36, 2, true, err)
				}
			}
			res.Country = tmp36
		}
		if v, ok := val["created_at"]; ok {
			var tmp37 string
			if val, ok := v.(string); ok {
				tmp37 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.CreatedAt`, v, "string", err)
			}
			if err == nil {
				if tmp37 != "" {
					if err2 := goa.ValidateFormat(goa.FormatDateTime, tmp37); err2 != nil {
						err = goa.InvalidFormatError(`.CreatedAt`, tmp37, goa.FormatDateTime, err2, err)
					}
				}
			}
			res.CreatedAt = tmp37
		}
		if v, ok := val["href"]; ok {
			var tmp38 string
			if val, ok := v.(string); ok {
				tmp38 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Href`, v, "string", err)
			}
			res.Href = tmp38
		}
		if v, ok := val["id"]; ok {
			var tmp39 int
			if f, ok := v.(float64); ok {
				tmp39 = int(f)
			} else {
				err = goa.InvalidAttributeTypeError(`.ID`, v, "int", err)
			}
			res.ID = tmp39
		}
		if v, ok := val["name"]; ok {
			var tmp40 string
			if val, ok := v.(string); ok {
				tmp40 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Name`, v, "string", err)
			}
			if err == nil {
				if len(tmp40) < 2 {
					err = goa.InvalidLengthError(`.Name`, tmp40, 2, true, err)
				}
			}
			res.Name = tmp40
		}
		if v, ok := val["rating"]; ok {
			var tmp41 int
			if f, ok := v.(float64); ok {
				tmp41 = int(f)
			} else {
				err = goa.InvalidAttributeTypeError(`.Rating`, v, "int", err)
			}
			if err == nil {
				if tmp41 < 1 {
					err = goa.InvalidRangeError(`.Rating`, tmp41, 1, true, err)
				}
				if tmp41 > 5 {
					err = goa.InvalidRangeError(`.Rating`, tmp41, 5, false, err)
				}
			}
			res.Rating = tmp41
		}
		if v, ok := val["region"]; ok {
			var tmp42 string
			if val, ok := v.(string); ok {
				tmp42 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Region`, v, "string", err)
			}
			res.Region = tmp42
		}
		if v, ok := val["review"]; ok {
			var tmp43 string
			if val, ok := v.(string); ok {
				tmp43 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Review`, v, "string", err)
			}
			if err == nil {
				if len(tmp43) < 10 {
					err = goa.InvalidLengthError(`.Review`, tmp43, 10, true, err)
				}
				if len(tmp43) > 300 {
					err = goa.InvalidLengthError(`.Review`, tmp43, 300, false, err)
				}
			}
			res.Review = tmp43
		}
		if v, ok := val["sweetness"]; ok {
			var tmp44 int
			if f, ok := v.(float64); ok {
				tmp44 = int(f)
			} else {
				err = goa.InvalidAttributeTypeError(`.Sweetness`, v, "int", err)
			}
			if err == nil {
				if tmp44 < 1 {
					err = goa.InvalidRangeError(`.Sweetness`, tmp44, 1, true, err)
				}
				if tmp44 > 5 {
					err = goa.InvalidRangeError(`.Sweetness`, tmp44, 5, false, err)
				}
			}
			res.Sweetness = tmp44
		}
		if v, ok := val["updated_at"]; ok {
			var tmp45 string
			if val, ok := v.(string); ok {
				tmp45 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.UpdatedAt`, v, "string", err)
			}
			if err == nil {
				if tmp45 != "" {
					if err2 := goa.ValidateFormat(goa.FormatDateTime, tmp45); err2 != nil {
						err = goa.InvalidFormatError(`.UpdatedAt`, tmp45, goa.FormatDateTime, err2, err)
					}
				}
			}
			res.UpdatedAt = tmp45
		}
		if v, ok := val["varietal"]; ok {
			var tmp46 string
			if val, ok := v.(string); ok {
				tmp46 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Varietal`, v, "string", err)
			}
			if err == nil {
				if len(tmp46) < 4 {
					err = goa.InvalidLengthError(`.Varietal`, tmp46, 4, true, err)
				}
			}
			res.Varietal = tmp46
		}
		if v, ok := val["vineyard"]; ok {
			var tmp47 string
			if val, ok := v.(string); ok {
				tmp47 = val
			} else {
				err = goa.InvalidAttributeTypeError(`.Vineyard`, v, "string", err)
			}
			if err == nil {
				if len(tmp47) < 2 {
					err = goa.InvalidLengthError(`.Vineyard`, tmp47, 2, true, err)
				}
			}
			res.Vineyard = tmp47
		}
		if v, ok := val["vintage"]; ok {
			var tmp48 int
			if f, ok := v.(float64); ok {
				tmp48 = int(f)
			} else {
				err = goa.InvalidAttributeTypeError(`.Vintage`, v, "int", err)
			}
			if err == nil {
				if tmp48 < 1900 {
					err = goa.InvalidRangeError(`.Vintage`, tmp48, 1900, true, err)
				}
				if tmp48 > 2020 {
					err = goa.InvalidRangeError(`.Vintage`, tmp48, 2020, false, err)
				}
			}
			res.Vintage = tmp48
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
		if len(mt.Name) < 2 {
			err = goa.InvalidLengthError(`default view.name`, mt.Name, 2, true, err)
		}
		if mt.Rating < 1 {
			err = goa.InvalidRangeError(`default view.rating`, mt.Rating, 1, true, err)
		}
		if mt.Rating > 5 {
			err = goa.InvalidRangeError(`default view.rating`, mt.Rating, 5, false, err)
		}
		if len(mt.Varietal) < 4 {
			err = goa.InvalidLengthError(`default view.varietal`, mt.Varietal, 4, true, err)
		}
		if len(mt.Vineyard) < 2 {
			err = goa.InvalidLengthError(`default view.vineyard`, mt.Vineyard, 2, true, err)
		}
		if mt.Vintage < 1900 {
			err = goa.InvalidRangeError(`default view.vintage`, mt.Vintage, 1900, true, err)
		}
		if mt.Vintage > 2020 {
			err = goa.InvalidRangeError(`default view.vintage`, mt.Vintage, 2020, false, err)
		}
		tmp50 := map[string]interface{}{
			"href":     mt.Href,
			"id":       mt.ID,
			"name":     mt.Name,
			"rating":   mt.Rating,
			"varietal": mt.Varietal,
			"vineyard": mt.Vineyard,
			"vintage":  mt.Vintage,
		}
		res = tmp50
		if err == nil {
			links := make(map[string]interface{})
			tmp49 := map[string]interface{}{
				"href": mt.Account.Href,
				"id":   mt.Account.ID,
				"name": mt.Account.Name,
			}
			links["account"] = tmp49
			res["links"] = links
		}
	}
	if view == BottleFullView {
		if mt.Color != "" {
			if !(mt.Color == "red" || mt.Color == "white" || mt.Color == "rose" || mt.Color == "yellow" || mt.Color == "sparkling") {
				err = goa.InvalidEnumValueError(`full view.color`, mt.Color, []interface{}{"red", "white", "rose", "yellow", "sparkling"}, err)
			}
		}
		if len(mt.Country) < 2 {
			err = goa.InvalidLengthError(`full view.country`, mt.Country, 2, true, err)
		}
		if mt.CreatedAt != "" {
			if err2 := goa.ValidateFormat(goa.FormatDateTime, mt.CreatedAt); err2 != nil {
				err = goa.InvalidFormatError(`full view.created_at`, mt.CreatedAt, goa.FormatDateTime, err2, err)
			}
		}
		if len(mt.Name) < 2 {
			err = goa.InvalidLengthError(`full view.name`, mt.Name, 2, true, err)
		}
		if mt.Rating < 1 {
			err = goa.InvalidRangeError(`full view.rating`, mt.Rating, 1, true, err)
		}
		if mt.Rating > 5 {
			err = goa.InvalidRangeError(`full view.rating`, mt.Rating, 5, false, err)
		}
		if len(mt.Review) < 10 {
			err = goa.InvalidLengthError(`full view.review`, mt.Review, 10, true, err)
		}
		if len(mt.Review) > 300 {
			err = goa.InvalidLengthError(`full view.review`, mt.Review, 300, false, err)
		}
		if mt.Sweetness < 1 {
			err = goa.InvalidRangeError(`full view.sweetness`, mt.Sweetness, 1, true, err)
		}
		if mt.Sweetness > 5 {
			err = goa.InvalidRangeError(`full view.sweetness`, mt.Sweetness, 5, false, err)
		}
		if mt.UpdatedAt != "" {
			if err2 := goa.ValidateFormat(goa.FormatDateTime, mt.UpdatedAt); err2 != nil {
				err = goa.InvalidFormatError(`full view.updated_at`, mt.UpdatedAt, goa.FormatDateTime, err2, err)
			}
		}
		if len(mt.Varietal) < 4 {
			err = goa.InvalidLengthError(`full view.varietal`, mt.Varietal, 4, true, err)
		}
		if len(mt.Vineyard) < 2 {
			err = goa.InvalidLengthError(`full view.vineyard`, mt.Vineyard, 2, true, err)
		}
		if mt.Vintage < 1900 {
			err = goa.InvalidRangeError(`full view.vintage`, mt.Vintage, 1900, true, err)
		}
		if mt.Vintage > 2020 {
			err = goa.InvalidRangeError(`full view.vintage`, mt.Vintage, 2020, false, err)
		}
		tmp52 := map[string]interface{}{
			"color":      mt.Color,
			"country":    mt.Country,
			"created_at": mt.CreatedAt,
			"href":       mt.Href,
			"id":         mt.ID,
			"name":       mt.Name,
			"rating":     mt.Rating,
			"region":     mt.Region,
			"review":     mt.Review,
			"sweetness":  mt.Sweetness,
			"updated_at": mt.UpdatedAt,
			"varietal":   mt.Varietal,
			"vineyard":   mt.Vineyard,
			"vintage":    mt.Vintage,
		}
		if mt.Account != nil {
			if mt.Account.CreatedAt != "" {
				if err2 := goa.ValidateFormat(goa.FormatDateTime, mt.Account.CreatedAt); err2 != nil {
					err = goa.InvalidFormatError(`full view.Account.created_at`, mt.Account.CreatedAt, goa.FormatDateTime, err2, err)
				}
			}
			if mt.Account.CreatedBy != "" {
				if err2 := goa.ValidateFormat(goa.FormatEmail, mt.Account.CreatedBy); err2 != nil {
					err = goa.InvalidFormatError(`full view.Account.created_by`, mt.Account.CreatedBy, goa.FormatEmail, err2, err)
				}
			}
			tmp53 := map[string]interface{}{
				"created_at": mt.Account.CreatedAt,
				"created_by": mt.Account.CreatedBy,
				"href":       mt.Account.Href,
				"id":         mt.Account.ID,
				"name":       mt.Account.Name,
			}
			tmp52["account"] = tmp53
		}
		res = tmp52
		if err == nil {
			links := make(map[string]interface{})
			tmp51 := map[string]interface{}{
				"href": mt.Account.Href,
				"id":   mt.Account.ID,
				"name": mt.Account.Name,
			}
			links["account"] = tmp51
			res["links"] = links
		}
	}
	if view == BottleTinyView {
		if len(mt.Name) < 2 {
			err = goa.InvalidLengthError(`tiny view.name`, mt.Name, 2, true, err)
		}
		if mt.Rating < 1 {
			err = goa.InvalidRangeError(`tiny view.rating`, mt.Rating, 1, true, err)
		}
		if mt.Rating > 5 {
			err = goa.InvalidRangeError(`tiny view.rating`, mt.Rating, 5, false, err)
		}
		tmp55 := map[string]interface{}{
			"href":   mt.Href,
			"id":     mt.ID,
			"name":   mt.Name,
			"rating": mt.Rating,
		}
		res = tmp55
		if err == nil {
			links := make(map[string]interface{})
			tmp54 := map[string]interface{}{
				"href": mt.Account.Href,
				"id":   mt.Account.ID,
				"name": mt.Account.Name,
			}
			links["account"] = tmp54
			res["links"] = links
		}
	}
	return res, err
}

// Validate validates the media type instance.
func (mt *Bottle) Validate() (err error) {
	if mt.Account.CreatedAt != "" {
		if err2 := goa.ValidateFormat(goa.FormatDateTime, mt.Account.CreatedAt); err2 != nil {
			err = goa.InvalidFormatError(`response.account.created_at`, mt.Account.CreatedAt, goa.FormatDateTime, err2, err)
		}
	}
	if mt.Account.CreatedBy != "" {
		if err2 := goa.ValidateFormat(goa.FormatEmail, mt.Account.CreatedBy); err2 != nil {
			err = goa.InvalidFormatError(`response.account.created_by`, mt.Account.CreatedBy, goa.FormatEmail, err2, err)
		}
	}
	if mt.Color != "" {
		if !(mt.Color == "red" || mt.Color == "white" || mt.Color == "rose" || mt.Color == "yellow" || mt.Color == "sparkling") {
			err = goa.InvalidEnumValueError(`response.color`, mt.Color, []interface{}{"red", "white", "rose", "yellow", "sparkling"}, err)
		}
	}
	if len(mt.Country) < 2 {
		err = goa.InvalidLengthError(`response.country`, mt.Country, 2, true, err)
	}
	if mt.CreatedAt != "" {
		if err2 := goa.ValidateFormat(goa.FormatDateTime, mt.CreatedAt); err2 != nil {
			err = goa.InvalidFormatError(`response.created_at`, mt.CreatedAt, goa.FormatDateTime, err2, err)
		}
	}
	if len(mt.Name) < 2 {
		err = goa.InvalidLengthError(`response.name`, mt.Name, 2, true, err)
	}
	if mt.Rating < 1 {
		err = goa.InvalidRangeError(`response.rating`, mt.Rating, 1, true, err)
	}
	if mt.Rating > 5 {
		err = goa.InvalidRangeError(`response.rating`, mt.Rating, 5, false, err)
	}
	if len(mt.Review) < 10 {
		err = goa.InvalidLengthError(`response.review`, mt.Review, 10, true, err)
	}
	if len(mt.Review) > 300 {
		err = goa.InvalidLengthError(`response.review`, mt.Review, 300, false, err)
	}
	if mt.Sweetness < 1 {
		err = goa.InvalidRangeError(`response.sweetness`, mt.Sweetness, 1, true, err)
	}
	if mt.Sweetness > 5 {
		err = goa.InvalidRangeError(`response.sweetness`, mt.Sweetness, 5, false, err)
	}
	if mt.UpdatedAt != "" {
		if err2 := goa.ValidateFormat(goa.FormatDateTime, mt.UpdatedAt); err2 != nil {
			err = goa.InvalidFormatError(`response.updated_at`, mt.UpdatedAt, goa.FormatDateTime, err2, err)
		}
	}
	if len(mt.Varietal) < 4 {
		err = goa.InvalidLengthError(`response.varietal`, mt.Varietal, 4, true, err)
	}
	if len(mt.Vineyard) < 2 {
		err = goa.InvalidLengthError(`response.vineyard`, mt.Vineyard, 2, true, err)
	}
	if mt.Vintage < 1900 {
		err = goa.InvalidRangeError(`response.vintage`, mt.Vintage, 1900, true, err)
	}
	if mt.Vintage > 2020 {
		err = goa.InvalidRangeError(`response.vintage`, mt.Vintage, 2020, false, err)
	}
	return
}

// BottleCollection media type
// Identifier: application/vnd.goa.example.bottle+json; type=collection
type BottleCollection []*Bottle

// array views
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
			var tmp56 *Bottle
			if val, ok := v.(map[string]interface{}); ok {
				tmp56 = new(Bottle)
				if v, ok := val["account"]; ok {
					var tmp57 *Account
					if val, ok := v.(map[string]interface{}); ok {
						tmp57 = new(Account)
						if v, ok := val["created_at"]; ok {
							var tmp58 string
							if val, ok := v.(string); ok {
								tmp58 = val
							} else {
								err = goa.InvalidAttributeTypeError(`[*].Account.CreatedAt`, v, "string", err)
							}
							if err == nil {
								if tmp58 != "" {
									if err2 := goa.ValidateFormat(goa.FormatDateTime, tmp58); err2 != nil {
										err = goa.InvalidFormatError(`[*].Account.CreatedAt`, tmp58, goa.FormatDateTime, err2, err)
									}
								}
							}
							tmp57.CreatedAt = tmp58
						}
						if v, ok := val["created_by"]; ok {
							var tmp59 string
							if val, ok := v.(string); ok {
								tmp59 = val
							} else {
								err = goa.InvalidAttributeTypeError(`[*].Account.CreatedBy`, v, "string", err)
							}
							if err == nil {
								if tmp59 != "" {
									if err2 := goa.ValidateFormat(goa.FormatEmail, tmp59); err2 != nil {
										err = goa.InvalidFormatError(`[*].Account.CreatedBy`, tmp59, goa.FormatEmail, err2, err)
									}
								}
							}
							tmp57.CreatedBy = tmp59
						}
						if v, ok := val["href"]; ok {
							var tmp60 string
							if val, ok := v.(string); ok {
								tmp60 = val
							} else {
								err = goa.InvalidAttributeTypeError(`[*].Account.Href`, v, "string", err)
							}
							tmp57.Href = tmp60
						}
						if v, ok := val["id"]; ok {
							var tmp61 int
							if f, ok := v.(float64); ok {
								tmp61 = int(f)
							} else {
								err = goa.InvalidAttributeTypeError(`[*].Account.ID`, v, "int", err)
							}
							tmp57.ID = tmp61
						}
						if v, ok := val["name"]; ok {
							var tmp62 string
							if val, ok := v.(string); ok {
								tmp62 = val
							} else {
								err = goa.InvalidAttributeTypeError(`[*].Account.Name`, v, "string", err)
							}
							tmp57.Name = tmp62
						}
					} else {
						err = goa.InvalidAttributeTypeError(`[*].Account`, v, "map[string]interface{}", err)
					}
					tmp56.Account = tmp57
				}
				if v, ok := val["color"]; ok {
					var tmp63 string
					if val, ok := v.(string); ok {
						tmp63 = val
					} else {
						err = goa.InvalidAttributeTypeError(`[*].Color`, v, "string", err)
					}
					if err == nil {
						if tmp63 != "" {
							if !(tmp63 == "red" || tmp63 == "white" || tmp63 == "rose" || tmp63 == "yellow" || tmp63 == "sparkling") {
								err = goa.InvalidEnumValueError(`[*].Color`, tmp63, []interface{}{"red", "white", "rose", "yellow", "sparkling"}, err)
							}
						}
					}
					tmp56.Color = tmp63
				}
				if v, ok := val["country"]; ok {
					var tmp64 string
					if val, ok := v.(string); ok {
						tmp64 = val
					} else {
						err = goa.InvalidAttributeTypeError(`[*].Country`, v, "string", err)
					}
					if err == nil {
						if len(tmp64) < 2 {
							err = goa.InvalidLengthError(`[*].Country`, tmp64, 2, true, err)
						}
					}
					tmp56.Country = tmp64
				}
				if v, ok := val["created_at"]; ok {
					var tmp65 string
					if val, ok := v.(string); ok {
						tmp65 = val
					} else {
						err = goa.InvalidAttributeTypeError(`[*].CreatedAt`, v, "string", err)
					}
					if err == nil {
						if tmp65 != "" {
							if err2 := goa.ValidateFormat(goa.FormatDateTime, tmp65); err2 != nil {
								err = goa.InvalidFormatError(`[*].CreatedAt`, tmp65, goa.FormatDateTime, err2, err)
							}
						}
					}
					tmp56.CreatedAt = tmp65
				}
				if v, ok := val["href"]; ok {
					var tmp66 string
					if val, ok := v.(string); ok {
						tmp66 = val
					} else {
						err = goa.InvalidAttributeTypeError(`[*].Href`, v, "string", err)
					}
					tmp56.Href = tmp66
				}
				if v, ok := val["id"]; ok {
					var tmp67 int
					if f, ok := v.(float64); ok {
						tmp67 = int(f)
					} else {
						err = goa.InvalidAttributeTypeError(`[*].ID`, v, "int", err)
					}
					tmp56.ID = tmp67
				}
				if v, ok := val["name"]; ok {
					var tmp68 string
					if val, ok := v.(string); ok {
						tmp68 = val
					} else {
						err = goa.InvalidAttributeTypeError(`[*].Name`, v, "string", err)
					}
					if err == nil {
						if len(tmp68) < 2 {
							err = goa.InvalidLengthError(`[*].Name`, tmp68, 2, true, err)
						}
					}
					tmp56.Name = tmp68
				}
				if v, ok := val["rating"]; ok {
					var tmp69 int
					if f, ok := v.(float64); ok {
						tmp69 = int(f)
					} else {
						err = goa.InvalidAttributeTypeError(`[*].Rating`, v, "int", err)
					}
					if err == nil {
						if tmp69 < 1 {
							err = goa.InvalidRangeError(`[*].Rating`, tmp69, 1, true, err)
						}
						if tmp69 > 5 {
							err = goa.InvalidRangeError(`[*].Rating`, tmp69, 5, false, err)
						}
					}
					tmp56.Rating = tmp69
				}
				if v, ok := val["region"]; ok {
					var tmp70 string
					if val, ok := v.(string); ok {
						tmp70 = val
					} else {
						err = goa.InvalidAttributeTypeError(`[*].Region`, v, "string", err)
					}
					tmp56.Region = tmp70
				}
				if v, ok := val["review"]; ok {
					var tmp71 string
					if val, ok := v.(string); ok {
						tmp71 = val
					} else {
						err = goa.InvalidAttributeTypeError(`[*].Review`, v, "string", err)
					}
					if err == nil {
						if len(tmp71) < 10 {
							err = goa.InvalidLengthError(`[*].Review`, tmp71, 10, true, err)
						}
						if len(tmp71) > 300 {
							err = goa.InvalidLengthError(`[*].Review`, tmp71, 300, false, err)
						}
					}
					tmp56.Review = tmp71
				}
				if v, ok := val["sweetness"]; ok {
					var tmp72 int
					if f, ok := v.(float64); ok {
						tmp72 = int(f)
					} else {
						err = goa.InvalidAttributeTypeError(`[*].Sweetness`, v, "int", err)
					}
					if err == nil {
						if tmp72 < 1 {
							err = goa.InvalidRangeError(`[*].Sweetness`, tmp72, 1, true, err)
						}
						if tmp72 > 5 {
							err = goa.InvalidRangeError(`[*].Sweetness`, tmp72, 5, false, err)
						}
					}
					tmp56.Sweetness = tmp72
				}
				if v, ok := val["updated_at"]; ok {
					var tmp73 string
					if val, ok := v.(string); ok {
						tmp73 = val
					} else {
						err = goa.InvalidAttributeTypeError(`[*].UpdatedAt`, v, "string", err)
					}
					if err == nil {
						if tmp73 != "" {
							if err2 := goa.ValidateFormat(goa.FormatDateTime, tmp73); err2 != nil {
								err = goa.InvalidFormatError(`[*].UpdatedAt`, tmp73, goa.FormatDateTime, err2, err)
							}
						}
					}
					tmp56.UpdatedAt = tmp73
				}
				if v, ok := val["varietal"]; ok {
					var tmp74 string
					if val, ok := v.(string); ok {
						tmp74 = val
					} else {
						err = goa.InvalidAttributeTypeError(`[*].Varietal`, v, "string", err)
					}
					if err == nil {
						if len(tmp74) < 4 {
							err = goa.InvalidLengthError(`[*].Varietal`, tmp74, 4, true, err)
						}
					}
					tmp56.Varietal = tmp74
				}
				if v, ok := val["vineyard"]; ok {
					var tmp75 string
					if val, ok := v.(string); ok {
						tmp75 = val
					} else {
						err = goa.InvalidAttributeTypeError(`[*].Vineyard`, v, "string", err)
					}
					if err == nil {
						if len(tmp75) < 2 {
							err = goa.InvalidLengthError(`[*].Vineyard`, tmp75, 2, true, err)
						}
					}
					tmp56.Vineyard = tmp75
				}
				if v, ok := val["vintage"]; ok {
					var tmp76 int
					if f, ok := v.(float64); ok {
						tmp76 = int(f)
					} else {
						err = goa.InvalidAttributeTypeError(`[*].Vintage`, v, "int", err)
					}
					if err == nil {
						if tmp76 < 1900 {
							err = goa.InvalidRangeError(`[*].Vintage`, tmp76, 1900, true, err)
						}
						if tmp76 > 2020 {
							err = goa.InvalidRangeError(`[*].Vintage`, tmp76, 2020, false, err)
						}
					}
					tmp56.Vintage = tmp76
				}
			} else {
				err = goa.InvalidAttributeTypeError(`[*]`, v, "map[string]interface{}", err)
			}
			res[i] = tmp56
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
		res = make([]map[string]interface{}, len(mt))
		for i, tmp77 := range mt {
			if len(tmp77.Name) < 2 {
				err = goa.InvalidLengthError(`default view[*].name`, tmp77.Name, 2, true, err)
			}
			if tmp77.Rating < 1 {
				err = goa.InvalidRangeError(`default view[*].rating`, tmp77.Rating, 1, true, err)
			}
			if tmp77.Rating > 5 {
				err = goa.InvalidRangeError(`default view[*].rating`, tmp77.Rating, 5, false, err)
			}
			if len(tmp77.Varietal) < 4 {
				err = goa.InvalidLengthError(`default view[*].varietal`, tmp77.Varietal, 4, true, err)
			}
			if len(tmp77.Vineyard) < 2 {
				err = goa.InvalidLengthError(`default view[*].vineyard`, tmp77.Vineyard, 2, true, err)
			}
			if tmp77.Vintage < 1900 {
				err = goa.InvalidRangeError(`default view[*].vintage`, tmp77.Vintage, 1900, true, err)
			}
			if tmp77.Vintage > 2020 {
				err = goa.InvalidRangeError(`default view[*].vintage`, tmp77.Vintage, 2020, false, err)
			}
			tmp79 := map[string]interface{}{
				"href":     tmp77.Href,
				"id":       tmp77.ID,
				"name":     tmp77.Name,
				"rating":   tmp77.Rating,
				"varietal": tmp77.Varietal,
				"vineyard": tmp77.Vineyard,
				"vintage":  tmp77.Vintage,
			}
			res[i] = tmp79
			if err == nil {
				links := make(map[string]interface{})
				tmp78 := map[string]interface{}{
					"href": tmp77.Account.Href,
					"id":   tmp77.Account.ID,
					"name": tmp77.Account.Name,
				}
				links["account"] = tmp78
				res[i]["links"] = links
			}
		}
	}
	if view == BottleCollectionTinyView {
		res = make([]map[string]interface{}, len(mt))
		for i, tmp80 := range mt {
			if len(tmp80.Name) < 2 {
				err = goa.InvalidLengthError(`tiny view[*].name`, tmp80.Name, 2, true, err)
			}
			if tmp80.Rating < 1 {
				err = goa.InvalidRangeError(`tiny view[*].rating`, tmp80.Rating, 1, true, err)
			}
			if tmp80.Rating > 5 {
				err = goa.InvalidRangeError(`tiny view[*].rating`, tmp80.Rating, 5, false, err)
			}
			tmp82 := map[string]interface{}{
				"href":   tmp80.Href,
				"id":     tmp80.ID,
				"name":   tmp80.Name,
				"rating": tmp80.Rating,
			}
			res[i] = tmp82
			if err == nil {
				links := make(map[string]interface{})
				tmp81 := map[string]interface{}{
					"href": tmp80.Account.Href,
					"id":   tmp80.Account.ID,
					"name": tmp80.Account.Name,
				}
				links["account"] = tmp81
				res[i]["links"] = links
			}
		}
	}
	return res, err
}

// Validate validates the media type instance.
func (mt BottleCollection) Validate() (err error) {
	for _, e := range mt {
		if e.Account.CreatedAt != "" {
			if err2 := goa.ValidateFormat(goa.FormatDateTime, e.Account.CreatedAt); err2 != nil {
				err = goa.InvalidFormatError(`response[*].account.created_at`, e.Account.CreatedAt, goa.FormatDateTime, err2, err)
			}
		}
		if e.Account.CreatedBy != "" {
			if err2 := goa.ValidateFormat(goa.FormatEmail, e.Account.CreatedBy); err2 != nil {
				err = goa.InvalidFormatError(`response[*].account.created_by`, e.Account.CreatedBy, goa.FormatEmail, err2, err)
			}
		}
		if e.Color != "" {
			if !(e.Color == "red" || e.Color == "white" || e.Color == "rose" || e.Color == "yellow" || e.Color == "sparkling") {
				err = goa.InvalidEnumValueError(`response[*].color`, e.Color, []interface{}{"red", "white", "rose", "yellow", "sparkling"}, err)
			}
		}
		if len(e.Country) < 2 {
			err = goa.InvalidLengthError(`response[*].country`, e.Country, 2, true, err)
		}
		if e.CreatedAt != "" {
			if err2 := goa.ValidateFormat(goa.FormatDateTime, e.CreatedAt); err2 != nil {
				err = goa.InvalidFormatError(`response[*].created_at`, e.CreatedAt, goa.FormatDateTime, err2, err)
			}
		}
		if len(e.Name) < 2 {
			err = goa.InvalidLengthError(`response[*].name`, e.Name, 2, true, err)
		}
		if e.Rating < 1 {
			err = goa.InvalidRangeError(`response[*].rating`, e.Rating, 1, true, err)
		}
		if e.Rating > 5 {
			err = goa.InvalidRangeError(`response[*].rating`, e.Rating, 5, false, err)
		}
		if len(e.Review) < 10 {
			err = goa.InvalidLengthError(`response[*].review`, e.Review, 10, true, err)
		}
		if len(e.Review) > 300 {
			err = goa.InvalidLengthError(`response[*].review`, e.Review, 300, false, err)
		}
		if e.Sweetness < 1 {
			err = goa.InvalidRangeError(`response[*].sweetness`, e.Sweetness, 1, true, err)
		}
		if e.Sweetness > 5 {
			err = goa.InvalidRangeError(`response[*].sweetness`, e.Sweetness, 5, false, err)
		}
		if e.UpdatedAt != "" {
			if err2 := goa.ValidateFormat(goa.FormatDateTime, e.UpdatedAt); err2 != nil {
				err = goa.InvalidFormatError(`response[*].updated_at`, e.UpdatedAt, goa.FormatDateTime, err2, err)
			}
		}
		if len(e.Varietal) < 4 {
			err = goa.InvalidLengthError(`response[*].varietal`, e.Varietal, 4, true, err)
		}
		if len(e.Vineyard) < 2 {
			err = goa.InvalidLengthError(`response[*].vineyard`, e.Vineyard, 2, true, err)
		}
		if e.Vintage < 1900 {
			err = goa.InvalidRangeError(`response[*].vintage`, e.Vintage, 1900, true, err)
		}
		if e.Vintage > 2020 {
			err = goa.InvalidRangeError(`response[*].vintage`, e.Vintage, 2020, false, err)
		}
	}
	return
}
