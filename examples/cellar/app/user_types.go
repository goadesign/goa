//************************************************************************//
// cellar: Application User Types
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

// MarshalBottlePayload validates and renders an instance of BottlePayload into a interface{}
func MarshalBottlePayload(source *BottlePayload, inErr error) (target map[string]interface{}, err error) {
	err = inErr
	if source.Color != "" {
		if !(source.Color == "red" || source.Color == "white" || source.Color == "rose" || source.Color == "yellow" || source.Color == "sparkling") {
			err = goa.InvalidEnumValueError(`.color`, source.Color, []interface{}{"red", "white", "rose", "yellow", "sparkling"}, err)
		}
	}
	if len(source.Country) < 2 {
		err = goa.InvalidLengthError(`.country`, source.Country, 2, true, err)
	}
	if len(source.Name) < 2 {
		err = goa.InvalidLengthError(`.name`, source.Name, 2, true, err)
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
	tmp51 := map[string]interface{}{
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
	target = tmp51
	return
}

// UnmarshalBottlePayload unmarshals and validates a raw interface{} into an instance of BottlePayload
func UnmarshalBottlePayload(source interface{}, inErr error) (target *BottlePayload, err error) {
	err = inErr
	if val, ok := source.(map[string]interface{}); ok {
		target = new(BottlePayload)
		if v, ok := val["color"]; ok {
			var tmp52 string
			if val, ok := v.(string); ok {
				tmp52 = val
			} else {
				err = goa.InvalidAttributeTypeError(`load.Color`, v, "string", err)
			}
			if err == nil {
				if tmp52 != "" {
					if !(tmp52 == "red" || tmp52 == "white" || tmp52 == "rose" || tmp52 == "yellow" || tmp52 == "sparkling") {
						err = goa.InvalidEnumValueError(`load.Color`, tmp52, []interface{}{"red", "white", "rose", "yellow", "sparkling"}, err)
					}
				}
			}
			target.Color = tmp52
		}
		if v, ok := val["country"]; ok {
			var tmp53 string
			if val, ok := v.(string); ok {
				tmp53 = val
			} else {
				err = goa.InvalidAttributeTypeError(`load.Country`, v, "string", err)
			}
			if err == nil {
				if len(tmp53) < 2 {
					err = goa.InvalidLengthError(`load.Country`, tmp53, 2, true, err)
				}
			}
			target.Country = tmp53
		}
		if v, ok := val["name"]; ok {
			var tmp54 string
			if val, ok := v.(string); ok {
				tmp54 = val
			} else {
				err = goa.InvalidAttributeTypeError(`load.Name`, v, "string", err)
			}
			if err == nil {
				if len(tmp54) < 2 {
					err = goa.InvalidLengthError(`load.Name`, tmp54, 2, true, err)
				}
			}
			target.Name = tmp54
		}
		if v, ok := val["region"]; ok {
			var tmp55 string
			if val, ok := v.(string); ok {
				tmp55 = val
			} else {
				err = goa.InvalidAttributeTypeError(`load.Region`, v, "string", err)
			}
			target.Region = tmp55
		}
		if v, ok := val["review"]; ok {
			var tmp56 string
			if val, ok := v.(string); ok {
				tmp56 = val
			} else {
				err = goa.InvalidAttributeTypeError(`load.Review`, v, "string", err)
			}
			if err == nil {
				if len(tmp56) < 10 {
					err = goa.InvalidLengthError(`load.Review`, tmp56, 10, true, err)
				}
				if len(tmp56) > 300 {
					err = goa.InvalidLengthError(`load.Review`, tmp56, 300, false, err)
				}
			}
			target.Review = tmp56
		}
		if v, ok := val["sweetness"]; ok {
			var tmp57 int
			if f, ok := v.(float64); ok {
				tmp57 = int(f)
			} else {
				err = goa.InvalidAttributeTypeError(`load.Sweetness`, v, "int", err)
			}
			if err == nil {
				if tmp57 < 1 {
					err = goa.InvalidRangeError(`load.Sweetness`, tmp57, 1, true, err)
				}
				if tmp57 > 5 {
					err = goa.InvalidRangeError(`load.Sweetness`, tmp57, 5, false, err)
				}
			}
			target.Sweetness = tmp57
		}
		if v, ok := val["varietal"]; ok {
			var tmp58 string
			if val, ok := v.(string); ok {
				tmp58 = val
			} else {
				err = goa.InvalidAttributeTypeError(`load.Varietal`, v, "string", err)
			}
			if err == nil {
				if len(tmp58) < 4 {
					err = goa.InvalidLengthError(`load.Varietal`, tmp58, 4, true, err)
				}
			}
			target.Varietal = tmp58
		}
		if v, ok := val["vineyard"]; ok {
			var tmp59 string
			if val, ok := v.(string); ok {
				tmp59 = val
			} else {
				err = goa.InvalidAttributeTypeError(`load.Vineyard`, v, "string", err)
			}
			if err == nil {
				if len(tmp59) < 2 {
					err = goa.InvalidLengthError(`load.Vineyard`, tmp59, 2, true, err)
				}
			}
			target.Vineyard = tmp59
		}
		if v, ok := val["vintage"]; ok {
			var tmp60 int
			if f, ok := v.(float64); ok {
				tmp60 = int(f)
			} else {
				err = goa.InvalidAttributeTypeError(`load.Vintage`, v, "int", err)
			}
			if err == nil {
				if tmp60 < 1900 {
					err = goa.InvalidRangeError(`load.Vintage`, tmp60, 1900, true, err)
				}
				if tmp60 > 2020 {
					err = goa.InvalidRangeError(`load.Vintage`, tmp60, 2020, false, err)
				}
			}
			target.Vintage = tmp60
		}
	} else {
		err = goa.InvalidAttributeTypeError(`load`, source, "dictionary", err)
	}
	return
}
