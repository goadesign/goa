//************************************************************************//
// cellar: Application Media Types
//
// Generated with goagen v0.0.1, command line:
// $ goagen
// --out=$(GOPATH)/src/github.com/raphael/goa/examples/cellar
// --design=github.com/raphael/goa/examples/cellar/design
// --pkg=app
//
// The content of this file is auto-generated, DO NOT MODIFY
//************************************************************************//

package app

import "github.com/raphael/goa"

// A tenant account
// Identifier: application/vnd.account+json
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

// Account views
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
func LoadAccount(raw interface{}) (res *Account, err error) {
	res, err = UnmarshalAccount(raw, err)
	return
}

// Dump produces raw data from an instance of Account running all the
// validations. See LoadAccount for the definition of raw data.
func (mt *Account) Dump(view AccountViewEnum) (res map[string]interface{}, err error) {
	if view == AccountDefaultView {
		res, err = MarshalAccount(mt, err)
	}
	if view == AccountLinkView {
		res, err = MarshalAccountLink(mt, err)
	}
	return
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

// MarshalAccount validates and renders an instance of Account into a interface{}
// using view "default".
func MarshalAccount(source *Account, inErr error) (target map[string]interface{}, err error) {
	err = inErr
	if source.CreatedAt != "" {
		if err2 := goa.ValidateFormat(goa.FormatDateTime, source.CreatedAt); err2 != nil {
			err = goa.InvalidFormatError(`.created_at`, source.CreatedAt, goa.FormatDateTime, err2, err)
		}
	}
	if source.CreatedBy != "" {
		if err2 := goa.ValidateFormat(goa.FormatEmail, source.CreatedBy); err2 != nil {
			err = goa.InvalidFormatError(`.created_by`, source.CreatedBy, goa.FormatEmail, err2, err)
		}
	}
	tmp22 := map[string]interface{}{
		"created_at": source.CreatedAt,
		"created_by": source.CreatedBy,
		"href":       source.Href,
		"id":         source.ID,
		"name":       source.Name,
	}
	target = tmp22
	return
}

// MarshalAccountLink validates and renders an instance of Account into a interface{}
// using view "link".
func MarshalAccountLink(source *Account, inErr error) (target map[string]interface{}, err error) {
	err = inErr
	tmp23 := map[string]interface{}{
		"href": source.Href,
		"id":   source.ID,
		"name": source.Name,
	}
	target = tmp23
	return
}

// UnmarshalAccount unmarshals and validates a raw interface{} into an instance of Account
func UnmarshalAccount(source interface{}, inErr error) (target *Account, err error) {
	err = inErr
	if val, ok := source.(map[string]interface{}); ok {
		target = new(Account)
		if v, ok := val["created_at"]; ok {
			var tmp24 string
			if val, ok := v.(string); ok {
				tmp24 = val
			} else {
				err = goa.InvalidAttributeTypeError(`load.CreatedAt`, v, "string", err)
			}
			if err == nil {
				if tmp24 != "" {
					if err2 := goa.ValidateFormat(goa.FormatDateTime, tmp24); err2 != nil {
						err = goa.InvalidFormatError(`load.CreatedAt`, tmp24, goa.FormatDateTime, err2, err)
					}
				}
			}
			target.CreatedAt = tmp24
		}
		if v, ok := val["created_by"]; ok {
			var tmp25 string
			if val, ok := v.(string); ok {
				tmp25 = val
			} else {
				err = goa.InvalidAttributeTypeError(`load.CreatedBy`, v, "string", err)
			}
			if err == nil {
				if tmp25 != "" {
					if err2 := goa.ValidateFormat(goa.FormatEmail, tmp25); err2 != nil {
						err = goa.InvalidFormatError(`load.CreatedBy`, tmp25, goa.FormatEmail, err2, err)
					}
				}
			}
			target.CreatedBy = tmp25
		}
		if v, ok := val["href"]; ok {
			var tmp26 string
			if val, ok := v.(string); ok {
				tmp26 = val
			} else {
				err = goa.InvalidAttributeTypeError(`load.Href`, v, "string", err)
			}
			target.Href = tmp26
		}
		if v, ok := val["id"]; ok {
			var tmp27 int
			if f, ok := v.(float64); ok {
				tmp27 = int(f)
			} else {
				err = goa.InvalidAttributeTypeError(`load.ID`, v, "int", err)
			}
			target.ID = tmp27
		}
		if v, ok := val["name"]; ok {
			var tmp28 string
			if val, ok := v.(string); ok {
				tmp28 = val
			} else {
				err = goa.InvalidAttributeTypeError(`load.Name`, v, "string", err)
			}
			target.Name = tmp28
		}
	} else {
		err = goa.InvalidAttributeTypeError(`load`, source, "dictionary", err)
	}
	return
}

// A bottle of wine
// Identifier: application/vnd.bottle+json
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
func LoadBottle(raw interface{}) (res *Bottle, err error) {
	res, err = UnmarshalBottle(raw, err)
	return
}

// Dump produces raw data from an instance of Bottle running all the
// validations. See LoadBottle for the definition of raw data.
func (mt *Bottle) Dump(view BottleViewEnum) (res map[string]interface{}, err error) {
	if view == BottleDefaultView {
		res, err = MarshalBottle(mt, err)
	}
	if view == BottleFullView {
		res, err = MarshalBottleFull(mt, err)
	}
	if view == BottleTinyView {
		res, err = MarshalBottleTiny(mt, err)
	}
	return
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

// MarshalBottle validates and renders an instance of Bottle into a interface{}
// using view "default".
func MarshalBottle(source *Bottle, inErr error) (target map[string]interface{}, err error) {
	err = inErr
	if len(source.Name) < 2 {
		err = goa.InvalidLengthError(`.name`, source.Name, 2, true, err)
	}
	if source.Rating < 1 {
		err = goa.InvalidRangeError(`.rating`, source.Rating, 1, true, err)
	}
	if source.Rating > 5 {
		err = goa.InvalidRangeError(`.rating`, source.Rating, 5, false, err)
	}
	if len(source.Varietal) < 4 {
		err = goa.InvalidLengthError(`.varietal`, source.Varietal, 4, true, err)
	}
	if len(source.Vineyard) < 2 {
		err = goa.InvalidLengthError(`.vineyard`, source.Vineyard, 2, true, err)
	}
	if source.Vintage < 1900 {
		err = goa.InvalidRangeError(`.vintage`, source.Vintage, 1900, true, err)
	}
	if source.Vintage > 2020 {
		err = goa.InvalidRangeError(`.vintage`, source.Vintage, 2020, false, err)
	}
	tmp29 := map[string]interface{}{
		"href":     source.Href,
		"id":       source.ID,
		"name":     source.Name,
		"rating":   source.Rating,
		"varietal": source.Varietal,
		"vineyard": source.Vineyard,
		"vintage":  source.Vintage,
	}
	target = tmp29
	if err == nil {
		links := make(map[string]interface{})
		links["account"], err = MarshalAccountLink(source.Account, err)
		target["links"] = links
	}
	return
}

// MarshalBottleFull validates and renders an instance of Bottle into a interface{}
// using view "full".
func MarshalBottleFull(source *Bottle, inErr error) (target map[string]interface{}, err error) {
	err = inErr
	if source.Color != "" {
		if !(source.Color == "red" || source.Color == "white" || source.Color == "rose" || source.Color == "yellow" || source.Color == "sparkling") {
			err = goa.InvalidEnumValueError(`.color`, source.Color, []interface{}{"red", "white", "rose", "yellow", "sparkling"}, err)
		}
	}
	if len(source.Country) < 2 {
		err = goa.InvalidLengthError(`.country`, source.Country, 2, true, err)
	}
	if source.CreatedAt != "" {
		if err2 := goa.ValidateFormat(goa.FormatDateTime, source.CreatedAt); err2 != nil {
			err = goa.InvalidFormatError(`.created_at`, source.CreatedAt, goa.FormatDateTime, err2, err)
		}
	}
	if len(source.Name) < 2 {
		err = goa.InvalidLengthError(`.name`, source.Name, 2, true, err)
	}
	if source.Rating < 1 {
		err = goa.InvalidRangeError(`.rating`, source.Rating, 1, true, err)
	}
	if source.Rating > 5 {
		err = goa.InvalidRangeError(`.rating`, source.Rating, 5, false, err)
	}
	if len(source.Review) < 10 {
		err = goa.InvalidLengthError(`.review`, source.Review, 10, true, err)
	}
	if len(source.Review) > 300 {
		err = goa.InvalidLengthError(`.review`, source.Review, 300, false, err)
	}
	if source.Sweetness < 1 {
		err = goa.InvalidRangeError(`.sweetness`, source.Sweetness, 1, true, err)
	}
	if source.Sweetness > 5 {
		err = goa.InvalidRangeError(`.sweetness`, source.Sweetness, 5, false, err)
	}
	if source.UpdatedAt != "" {
		if err2 := goa.ValidateFormat(goa.FormatDateTime, source.UpdatedAt); err2 != nil {
			err = goa.InvalidFormatError(`.updated_at`, source.UpdatedAt, goa.FormatDateTime, err2, err)
		}
	}
	if len(source.Varietal) < 4 {
		err = goa.InvalidLengthError(`.varietal`, source.Varietal, 4, true, err)
	}
	if len(source.Vineyard) < 2 {
		err = goa.InvalidLengthError(`.vineyard`, source.Vineyard, 2, true, err)
	}
	if source.Vintage < 1900 {
		err = goa.InvalidRangeError(`.vintage`, source.Vintage, 1900, true, err)
	}
	if source.Vintage > 2020 {
		err = goa.InvalidRangeError(`.vintage`, source.Vintage, 2020, false, err)
	}
	tmp30 := map[string]interface{}{
		"color":      source.Color,
		"country":    source.Country,
		"created_at": source.CreatedAt,
		"href":       source.Href,
		"id":         source.ID,
		"name":       source.Name,
		"rating":     source.Rating,
		"region":     source.Region,
		"review":     source.Review,
		"sweetness":  source.Sweetness,
		"updated_at": source.UpdatedAt,
		"varietal":   source.Varietal,
		"vineyard":   source.Vineyard,
		"vintage":    source.Vintage,
	}
	if source.Account != nil {
		tmp30["account"], err = MarshalAccount(source.Account, err)
	}
	target = tmp30
	if err == nil {
		links := make(map[string]interface{})
		links["account"], err = MarshalAccountLink(source.Account, err)
		target["links"] = links
	}
	return
}

// MarshalBottleTiny validates and renders an instance of Bottle into a interface{}
// using view "tiny".
func MarshalBottleTiny(source *Bottle, inErr error) (target map[string]interface{}, err error) {
	err = inErr
	if len(source.Name) < 2 {
		err = goa.InvalidLengthError(`.name`, source.Name, 2, true, err)
	}
	if source.Rating < 1 {
		err = goa.InvalidRangeError(`.rating`, source.Rating, 1, true, err)
	}
	if source.Rating > 5 {
		err = goa.InvalidRangeError(`.rating`, source.Rating, 5, false, err)
	}
	tmp31 := map[string]interface{}{
		"href":   source.Href,
		"id":     source.ID,
		"name":   source.Name,
		"rating": source.Rating,
	}
	target = tmp31
	if err == nil {
		links := make(map[string]interface{})
		links["account"], err = MarshalAccountLink(source.Account, err)
		target["links"] = links
	}
	return
}

// UnmarshalBottle unmarshals and validates a raw interface{} into an instance of Bottle
func UnmarshalBottle(source interface{}, inErr error) (target *Bottle, err error) {
	err = inErr
	if val, ok := source.(map[string]interface{}); ok {
		target = new(Bottle)
		if v, ok := val["account"]; ok {
			var tmp32 *Account
			tmp32, err = UnmarshalAccount(v, err)
			target.Account = tmp32
		}
		if v, ok := val["color"]; ok {
			var tmp33 string
			if val, ok := v.(string); ok {
				tmp33 = val
			} else {
				err = goa.InvalidAttributeTypeError(`load.Color`, v, "string", err)
			}
			if err == nil {
				if tmp33 != "" {
					if !(tmp33 == "red" || tmp33 == "white" || tmp33 == "rose" || tmp33 == "yellow" || tmp33 == "sparkling") {
						err = goa.InvalidEnumValueError(`load.Color`, tmp33, []interface{}{"red", "white", "rose", "yellow", "sparkling"}, err)
					}
				}
			}
			target.Color = tmp33
		}
		if v, ok := val["country"]; ok {
			var tmp34 string
			if val, ok := v.(string); ok {
				tmp34 = val
			} else {
				err = goa.InvalidAttributeTypeError(`load.Country`, v, "string", err)
			}
			if err == nil {
				if len(tmp34) < 2 {
					err = goa.InvalidLengthError(`load.Country`, tmp34, 2, true, err)
				}
			}
			target.Country = tmp34
		}
		if v, ok := val["created_at"]; ok {
			var tmp35 string
			if val, ok := v.(string); ok {
				tmp35 = val
			} else {
				err = goa.InvalidAttributeTypeError(`load.CreatedAt`, v, "string", err)
			}
			if err == nil {
				if tmp35 != "" {
					if err2 := goa.ValidateFormat(goa.FormatDateTime, tmp35); err2 != nil {
						err = goa.InvalidFormatError(`load.CreatedAt`, tmp35, goa.FormatDateTime, err2, err)
					}
				}
			}
			target.CreatedAt = tmp35
		}
		if v, ok := val["href"]; ok {
			var tmp36 string
			if val, ok := v.(string); ok {
				tmp36 = val
			} else {
				err = goa.InvalidAttributeTypeError(`load.Href`, v, "string", err)
			}
			target.Href = tmp36
		}
		if v, ok := val["id"]; ok {
			var tmp37 int
			if f, ok := v.(float64); ok {
				tmp37 = int(f)
			} else {
				err = goa.InvalidAttributeTypeError(`load.ID`, v, "int", err)
			}
			target.ID = tmp37
		}
		if v, ok := val["name"]; ok {
			var tmp38 string
			if val, ok := v.(string); ok {
				tmp38 = val
			} else {
				err = goa.InvalidAttributeTypeError(`load.Name`, v, "string", err)
			}
			if err == nil {
				if len(tmp38) < 2 {
					err = goa.InvalidLengthError(`load.Name`, tmp38, 2, true, err)
				}
			}
			target.Name = tmp38
		}
		if v, ok := val["rating"]; ok {
			var tmp39 int
			if f, ok := v.(float64); ok {
				tmp39 = int(f)
			} else {
				err = goa.InvalidAttributeTypeError(`load.Rating`, v, "int", err)
			}
			if err == nil {
				if tmp39 < 1 {
					err = goa.InvalidRangeError(`load.Rating`, tmp39, 1, true, err)
				}
				if tmp39 > 5 {
					err = goa.InvalidRangeError(`load.Rating`, tmp39, 5, false, err)
				}
			}
			target.Rating = tmp39
		}
		if v, ok := val["region"]; ok {
			var tmp40 string
			if val, ok := v.(string); ok {
				tmp40 = val
			} else {
				err = goa.InvalidAttributeTypeError(`load.Region`, v, "string", err)
			}
			target.Region = tmp40
		}
		if v, ok := val["review"]; ok {
			var tmp41 string
			if val, ok := v.(string); ok {
				tmp41 = val
			} else {
				err = goa.InvalidAttributeTypeError(`load.Review`, v, "string", err)
			}
			if err == nil {
				if len(tmp41) < 10 {
					err = goa.InvalidLengthError(`load.Review`, tmp41, 10, true, err)
				}
				if len(tmp41) > 300 {
					err = goa.InvalidLengthError(`load.Review`, tmp41, 300, false, err)
				}
			}
			target.Review = tmp41
		}
		if v, ok := val["sweetness"]; ok {
			var tmp42 int
			if f, ok := v.(float64); ok {
				tmp42 = int(f)
			} else {
				err = goa.InvalidAttributeTypeError(`load.Sweetness`, v, "int", err)
			}
			if err == nil {
				if tmp42 < 1 {
					err = goa.InvalidRangeError(`load.Sweetness`, tmp42, 1, true, err)
				}
				if tmp42 > 5 {
					err = goa.InvalidRangeError(`load.Sweetness`, tmp42, 5, false, err)
				}
			}
			target.Sweetness = tmp42
		}
		if v, ok := val["updated_at"]; ok {
			var tmp43 string
			if val, ok := v.(string); ok {
				tmp43 = val
			} else {
				err = goa.InvalidAttributeTypeError(`load.UpdatedAt`, v, "string", err)
			}
			if err == nil {
				if tmp43 != "" {
					if err2 := goa.ValidateFormat(goa.FormatDateTime, tmp43); err2 != nil {
						err = goa.InvalidFormatError(`load.UpdatedAt`, tmp43, goa.FormatDateTime, err2, err)
					}
				}
			}
			target.UpdatedAt = tmp43
		}
		if v, ok := val["varietal"]; ok {
			var tmp44 string
			if val, ok := v.(string); ok {
				tmp44 = val
			} else {
				err = goa.InvalidAttributeTypeError(`load.Varietal`, v, "string", err)
			}
			if err == nil {
				if len(tmp44) < 4 {
					err = goa.InvalidLengthError(`load.Varietal`, tmp44, 4, true, err)
				}
			}
			target.Varietal = tmp44
		}
		if v, ok := val["vineyard"]; ok {
			var tmp45 string
			if val, ok := v.(string); ok {
				tmp45 = val
			} else {
				err = goa.InvalidAttributeTypeError(`load.Vineyard`, v, "string", err)
			}
			if err == nil {
				if len(tmp45) < 2 {
					err = goa.InvalidLengthError(`load.Vineyard`, tmp45, 2, true, err)
				}
			}
			target.Vineyard = tmp45
		}
		if v, ok := val["vintage"]; ok {
			var tmp46 int
			if f, ok := v.(float64); ok {
				tmp46 = int(f)
			} else {
				err = goa.InvalidAttributeTypeError(`load.Vintage`, v, "int", err)
			}
			if err == nil {
				if tmp46 < 1900 {
					err = goa.InvalidRangeError(`load.Vintage`, tmp46, 1900, true, err)
				}
				if tmp46 > 2020 {
					err = goa.InvalidRangeError(`load.Vintage`, tmp46, 2020, false, err)
				}
			}
			target.Vintage = tmp46
		}
	} else {
		err = goa.InvalidAttributeTypeError(`load`, source, "dictionary", err)
	}
	return
}

// BottleCollection media type
// Identifier: application/vnd.bottle+json; type=collection
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
func LoadBottleCollection(raw interface{}) (res BottleCollection, err error) {
	res, err = UnmarshalBottleCollection(raw, err)
	return
}

// Dump produces raw data from an instance of BottleCollection running all the
// validations. See LoadBottleCollection for the definition of raw data.
func (mt BottleCollection) Dump(view BottleCollectionViewEnum) (res []map[string]interface{}, err error) {
	if view == BottleCollectionDefaultView {
		res = make([]map[string]interface{}, len(mt))
		for i, tmp47 := range mt {
			var tmp48 map[string]interface{}
			tmp48, err = MarshalBottle(tmp47, err)
			res[i] = tmp48
		}
	}
	if view == BottleCollectionTinyView {
		res = make([]map[string]interface{}, len(mt))
		for i, tmp49 := range mt {
			var tmp50 map[string]interface{}
			tmp50, err = MarshalBottleTiny(tmp49, err)
			res[i] = tmp50
		}
	}
	return
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

// MarshalBottleCollection validates and renders an instance of BottleCollection into a interface{}
// using view "default".
func MarshalBottleCollection(source BottleCollection, inErr error) (target []map[string]interface{}, err error) {
	err = inErr
	target = make([]map[string]interface{}, len(source))
	for i, res := range source {
		target[i], err = MarshalBottle(res, err)
	}
	return
}

// MarshalBottleCollectionTiny validates and renders an instance of BottleCollection into a interface{}
// using view "tiny".
func MarshalBottleCollectionTiny(source BottleCollection, inErr error) (target []map[string]interface{}, err error) {
	err = inErr
	target = make([]map[string]interface{}, len(source))
	for i, res := range source {
		target[i], err = MarshalBottleTiny(res, err)
	}
	return
}

// UnmarshalBottleCollection unmarshals and validates a raw interface{} into an instance of BottleCollection
func UnmarshalBottleCollection(source interface{}, inErr error) (target BottleCollection, err error) {
	err = inErr
	if val, ok := source.([]interface{}); ok {
		target = make([]*Bottle, len(val))
		for i, v := range val {
			target[i], err = UnmarshalBottle(v, err)
		}
	} else {
		err = goa.InvalidAttributeTypeError(`load`, source, "array", err)
	}
	return
}
