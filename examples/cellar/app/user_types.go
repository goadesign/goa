//************************************************************************//
// API "cellar": Application User Types
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
	"github.com/raphael/goa"
)

// BottlePayload type
type BottlePayload struct {
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

// Validate validates the type instance.
func (ut *BottlePayload) Validate() (err error) {
	if ut.Color != "" {
		if !(ut.Color == "red" || ut.Color == "white" || ut.Color == "rose" || ut.Color == "yellow" || ut.Color == "sparkling") {
			err = goa.InvalidEnumValueError(`response.color`, ut.Color, []interface{}{"red", "white", "rose", "yellow", "sparkling"}, err)
		}
	}
	if len(ut.Country) < 2 {
		err = goa.InvalidLengthError(`response.country`, ut.Country, len(ut.Country), 2, true, err)
	}
	if len(ut.Name) < 2 {
		err = goa.InvalidLengthError(`response.name`, ut.Name, len(ut.Name), 2, true, err)
	}
	if len(ut.Review) < 10 {
		err = goa.InvalidLengthError(`response.review`, ut.Review, len(ut.Review), 10, true, err)
	}
	if len(ut.Review) > 300 {
		err = goa.InvalidLengthError(`response.review`, ut.Review, len(ut.Review), 300, false, err)
	}
	if ut.Sweetness < 1 {
		err = goa.InvalidRangeError(`response.sweetness`, ut.Sweetness, 1, true, err)
	}
	if ut.Sweetness > 5 {
		err = goa.InvalidRangeError(`response.sweetness`, ut.Sweetness, 5, false, err)
	}
	if len(ut.Varietal) < 4 {
		err = goa.InvalidLengthError(`response.varietal`, ut.Varietal, len(ut.Varietal), 4, true, err)
	}
	if len(ut.Vineyard) < 2 {
		err = goa.InvalidLengthError(`response.vineyard`, ut.Vineyard, len(ut.Vineyard), 2, true, err)
	}
	if ut.Vintage < 1900 {
		err = goa.InvalidRangeError(`response.vintage`, ut.Vintage, 1900, true, err)
	}
	if ut.Vintage > 2020 {
		err = goa.InvalidRangeError(`response.vintage`, ut.Vintage, 2020, false, err)
	}
	return
}

// MarshalBottlePayload validates and renders an instance of BottlePayload into a interface{}
func MarshalBottlePayload(source *BottlePayload, inErr error) (target map[string]interface{}, err error) {
	err = inErr
	if source.Color != "" {
		if !(source.Color == "red" || source.Color == "white" || source.Color == "rose" || source.Color == "yellow" || source.Color == "sparkling") {
			err = goa.InvalidEnumValueError(`.color`, source.Color, []interface{}{"red", "white", "rose", "yellow", "sparkling"}, err)
		}
	}
	if len(source.Country) < 2 {
		err = goa.InvalidLengthError(`.country`, source.Country, len(source.Country), 2, true, err)
	}
	if len(source.Name) < 2 {
		err = goa.InvalidLengthError(`.name`, source.Name, len(source.Name), 2, true, err)
	}
	if len(source.Review) < 10 {
		err = goa.InvalidLengthError(`.review`, source.Review, len(source.Review), 10, true, err)
	}
	if len(source.Review) > 300 {
		err = goa.InvalidLengthError(`.review`, source.Review, len(source.Review), 300, false, err)
	}
	if source.Sweetness < 1 {
		err = goa.InvalidRangeError(`.sweetness`, source.Sweetness, 1, true, err)
	}
	if source.Sweetness > 5 {
		err = goa.InvalidRangeError(`.sweetness`, source.Sweetness, 5, false, err)
	}
	if len(source.Varietal) < 4 {
		err = goa.InvalidLengthError(`.varietal`, source.Varietal, len(source.Varietal), 4, true, err)
	}
	if len(source.Vineyard) < 2 {
		err = goa.InvalidLengthError(`.vineyard`, source.Vineyard, len(source.Vineyard), 2, true, err)
	}
	if source.Vintage < 1900 {
		err = goa.InvalidRangeError(`.vintage`, source.Vintage, 1900, true, err)
	}
	if source.Vintage > 2020 {
		err = goa.InvalidRangeError(`.vintage`, source.Vintage, 2020, false, err)
	}
	tmp52 := map[string]interface{}{
		"color":     source.Color,
		"country":   source.Country,
		"name":      source.Name,
		"region":    source.Region,
		"review":    source.Review,
		"sweetness": source.Sweetness,
		"varietal":  source.Varietal,
		"vineyard":  source.Vineyard,
		"vintage":   source.Vintage,
	}
	target = tmp52
	return
}

// UnmarshalBottlePayload unmarshals and validates a raw interface{} into an instance of BottlePayload
func UnmarshalBottlePayload(source interface{}, inErr error) (target *BottlePayload, err error) {
	err = inErr
	if val, ok := source.(map[string]interface{}); ok {
		target = new(BottlePayload)
		if v, ok := val["color"]; ok {
			var tmp53 string
			if val, ok := v.(string); ok {
				tmp53 = val
			} else {
				err = goa.InvalidAttributeTypeError(`load.Color`, v, "string", err)
			}
			if err == nil {
				if tmp53 != "" {
					if !(tmp53 == "red" || tmp53 == "white" || tmp53 == "rose" || tmp53 == "yellow" || tmp53 == "sparkling") {
						err = goa.InvalidEnumValueError(`load.Color`, tmp53, []interface{}{"red", "white", "rose", "yellow", "sparkling"}, err)
					}
				}
			}
			target.Color = tmp53
		}
		if v, ok := val["country"]; ok {
			var tmp54 string
			if val, ok := v.(string); ok {
				tmp54 = val
			} else {
				err = goa.InvalidAttributeTypeError(`load.Country`, v, "string", err)
			}
			if err == nil {
				if len(tmp54) < 2 {
					err = goa.InvalidLengthError(`load.Country`, tmp54, len(tmp54), 2, true, err)
				}
			}
			target.Country = tmp54
		}
		if v, ok := val["name"]; ok {
			var tmp55 string
			if val, ok := v.(string); ok {
				tmp55 = val
			} else {
				err = goa.InvalidAttributeTypeError(`load.Name`, v, "string", err)
			}
			if err == nil {
				if len(tmp55) < 2 {
					err = goa.InvalidLengthError(`load.Name`, tmp55, len(tmp55), 2, true, err)
				}
			}
			target.Name = tmp55
		}
		if v, ok := val["region"]; ok {
			var tmp56 string
			if val, ok := v.(string); ok {
				tmp56 = val
			} else {
				err = goa.InvalidAttributeTypeError(`load.Region`, v, "string", err)
			}
			target.Region = tmp56
		}
		if v, ok := val["review"]; ok {
			var tmp57 string
			if val, ok := v.(string); ok {
				tmp57 = val
			} else {
				err = goa.InvalidAttributeTypeError(`load.Review`, v, "string", err)
			}
			if err == nil {
				if len(tmp57) < 10 {
					err = goa.InvalidLengthError(`load.Review`, tmp57, len(tmp57), 10, true, err)
				}
				if len(tmp57) > 300 {
					err = goa.InvalidLengthError(`load.Review`, tmp57, len(tmp57), 300, false, err)
				}
			}
			target.Review = tmp57
		}
		if v, ok := val["sweetness"]; ok {
			var tmp58 int
			if f, ok := v.(float64); ok {
				tmp58 = int(f)
			} else {
				err = goa.InvalidAttributeTypeError(`load.Sweetness`, v, "int", err)
			}
			if err == nil {
				if tmp58 < 1 {
					err = goa.InvalidRangeError(`load.Sweetness`, tmp58, 1, true, err)
				}
				if tmp58 > 5 {
					err = goa.InvalidRangeError(`load.Sweetness`, tmp58, 5, false, err)
				}
			}
			target.Sweetness = tmp58
		}
		if v, ok := val["varietal"]; ok {
			var tmp59 string
			if val, ok := v.(string); ok {
				tmp59 = val
			} else {
				err = goa.InvalidAttributeTypeError(`load.Varietal`, v, "string", err)
			}
			if err == nil {
				if len(tmp59) < 4 {
					err = goa.InvalidLengthError(`load.Varietal`, tmp59, len(tmp59), 4, true, err)
				}
			}
			target.Varietal = tmp59
		}
		if v, ok := val["vineyard"]; ok {
			var tmp60 string
			if val, ok := v.(string); ok {
				tmp60 = val
			} else {
				err = goa.InvalidAttributeTypeError(`load.Vineyard`, v, "string", err)
			}
			if err == nil {
				if len(tmp60) < 2 {
					err = goa.InvalidLengthError(`load.Vineyard`, tmp60, len(tmp60), 2, true, err)
				}
			}
			target.Vineyard = tmp60
		}
		if v, ok := val["vintage"]; ok {
			var tmp61 int
			if f, ok := v.(float64); ok {
				tmp61 = int(f)
			} else {
				err = goa.InvalidAttributeTypeError(`load.Vintage`, v, "int", err)
			}
			if err == nil {
				if tmp61 < 1900 {
					err = goa.InvalidRangeError(`load.Vintage`, tmp61, 1900, true, err)
				}
				if tmp61 > 2020 {
					err = goa.InvalidRangeError(`load.Vintage`, tmp61, 2020, false, err)
				}
			}
			target.Vintage = tmp61
		}
	} else {
		err = goa.InvalidAttributeTypeError(`load`, source, "dictionary", err)
	}
	return
}
