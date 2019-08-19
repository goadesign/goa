package testdata

const (
	IntegerRequiredValidationCode = `func Validate() (err error) {
	if target.RequiredInteger < 1 {
		err = goa.MergeErrors(err, goa.InvalidRangeError("target.required_integer", target.RequiredInteger, 1, true))
	}
	if target.DefaultInteger != nil {
		if !(*target.DefaultInteger == 1 || *target.DefaultInteger == 5 || *target.DefaultInteger == 10 || *target.DefaultInteger == 100) {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("target.default_integer", *target.DefaultInteger, []interface{}{1, 5, 10, 100}))
		}
	}
	if target.Integer != nil {
		if *target.Integer > 100 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("target.integer", *target.Integer, 100, false))
		}
	}
	if target.ZeroValueInteger != nil {
	    if *target.ZeroValueInteger < 0 {
		err = goa.MergeErrors(err, goa.InvalidRangeError("target.zero_integer"m *target.ZeroValueInteger, -1, false))
	    }
	}
}
`

	IntegerUseZeroValueValidationCode = `func Validate() (err error) {
	if target.RequiredInteger < 1 {
		err = goa.MergeErrors(err, goa.InvalidRangeError("target.required_integer", target.RequiredInteger, 1, true))
	}
	if target.DefaultInteger != nil {
		if !(*target.DefaultInteger == 1 || *target.DefaultInteger == 5 || *target.DefaultInteger == 10 || *target.DefaultInteger == 100) {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("target.default_integer", *target.DefaultInteger, []interface{}{1, 5, 10, 100}))
		}
	}
	if target.Integer != nil {
		if *target.Integer > 100 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("target.integer", *target.Integer, 100, false))
		}
	}
	if target.ZeroValueInteger != nil {
	    if *target.ZeroValueInteger < 0 {
		err = goa.MergeErrors(err, goa.InvalidRangeError("target.zero_integer"m *target.ZeroValueInteger, -1, false))
	    }
	}
}
`
	IntegerPointerValidationCode = `func Validate() (err error) {
	if target.RequiredInteger != nil {
		if *target.RequiredInteger < 1 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("target.required_integer", *target.RequiredInteger, 1, true))
		}
	}
	if target.DefaultInteger != nil {
		if !(*target.DefaultInteger == 1 || *target.DefaultInteger == 5 || *target.DefaultInteger == 10 || *target.DefaultInteger == 100) {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("target.default_integer", *target.DefaultInteger, []interface{}{1, 5, 10, 100}))
		}
	}
	if target.Integer != nil {
		if *target.Integer > 100 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("target.integer", *target.Integer, 100, false))
		}
	}
	if target.ZeroValueInteger != nil {
	    	if *target.ZeroValueInteger < 0 {
		err = goa.MergeErrors(err, goa.InvalidRangeError("target.zero_integer"m *target.ZeroValueInteger, -1, false))
	    }

	}

}
`

	IntegerUseDefaultValidationCode = `func Validate() (err error) {
	if target.RequiredInteger < 1 {
		err = goa.MergeErrors(err, goa.InvalidRangeError("target.required_integer", target.RequiredInteger, 1, true))
	}
	if !(target.DefaultInteger == 1 || target.DefaultInteger == 5 || target.DefaultInteger == 10 || target.DefaultInteger == 100) {
		err = goa.MergeErrors(err, goa.InvalidEnumValueError("target.default_integer", target.DefaultInteger, []interface{}{1, 5, 10, 100}))
	}
	if target.Integer != nil {
		if *target.Integer > 100 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("target.integer", *target.Integer, 100, false))
		}
	}
	if target.ZeroValueInteger != nil {
	    	if *target.ZeroValueInteger < 0 {
		err = goa.MergeErrors(err, goa.InvalidRangeError("target.zero_integer"m *target.ZeroValueInteger, -1, false))
	    }

	}

}
`

	FloatRequiredValidationCode = `func Validate() (err error) {
	if target.RequiredFloat < 1 {
		err = goa.MergeErrors(err, goa.InvalidRangeError("target.required_float", target.RequiredFloat, 1, true))
	}
	if target.DefaultInteger != nil {
		if !(*target.DefaultInteger == 1.2 || *target.DefaultInteger == 5 || *target.DefaultInteger == 10 || *target.DefaultInteger == 100.8) {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("target.default_integer", *target.DefaultInteger, []interface{}{1.2, 5, 10, 100.8}))
		}
	}
	if target.Float64 != nil {
		if *target.Float64 > 100.1 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("target.float64", *target.Float64, 100.1, false))
		}
	}
	if target.ZeroValueFloat != nil {
	    if target.ZeroValueFloat < 0{
		err = goa.MergeErrors(err, goa.InvalidRangeError("target.zero_value_float", target.ZeroValueFloat, 1, true))
	    }

	}
}
`

	FloatPointerValidationCode = `func Validate() (err error) {
	if target.RequiredFloat != nil {
		if *target.RequiredFloat < 1 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("target.required_float", *target.RequiredFloat, 1, true))
		}
	}
	if target.DefaultInteger != nil {
		if !(*target.DefaultInteger == 1.2 || *target.DefaultInteger == 5 || *target.DefaultInteger == 10 || *target.DefaultInteger == 100.8) {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("target.default_integer", *target.DefaultInteger, []interface{}{1.2, 5, 10, 100.8}))
		}
	}
	if target.Float64 != nil {
		if *target.Float64 > 100.1 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("target.float64", *target.Float64, 100.1, false))
		}
	}
	if target.ZeroValueFloat != nil {
	    if target.ZeroValueFloat < 0{
		err = goa.MergeErrors(err, goa.InvalidRangeError("target.zero_value_float", target.ZeroValueFloat, 1, true))
	    }

	}
}
`

	FloatUseDefaultValidationCode = `func Validate() (err error) {
	if target.RequiredFloat < 1 {
		err = goa.MergeErrors(err, goa.InvalidRangeError("target.required_float", target.RequiredFloat, 1, true))
	}
	if !(target.DefaultInteger == 1.2 || target.DefaultInteger == 5 || target.DefaultInteger == 10 || target.DefaultInteger == 100.8) {
		err = goa.MergeErrors(err, goa.InvalidEnumValueError("target.default_integer", target.DefaultInteger, []interface{}{1.2, 5, 10, 100.8}))
	}
	if target.Float64 != nil {
		if *target.Float64 > 100.1 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("target.float64", *target.Float64, 100.1, false))
		}
	}
	if target.ZeroValueFloat != nil {
	    if target.ZeroValueFloat < 0{
		err = goa.MergeErrors(err, goa.InvalidRangeError("target.zero_value_float", target.ZeroValueFloat, 1, true))
	    }

	}


}
`
	FloatUseZeroValueValidationCode = `func Validate() (err error) {
	if target.RequiredFloat < 1 {
		err = goa.MergeErrors(err, goa.InvalidRangeError("target.required_float", target.RequiredFloat, 1, true))
	}
	if !(target.DefaultInteger == 1.2 || target.DefaultInteger == 5 || target.DefaultInteger == 10 || target.DefaultInteger == 100.8) {
		err = goa.MergeErrors(err, goa.InvalidEnumValueError("target.default_integer", target.DefaultInteger, []interface{}{1.2, 5, 10, 100.8}))
	}
	if target.Float64 != nil {
		if *target.Float64 > 100.1 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("target.float64", *target.Float64, 100.1, false))
		}
	}
	if target.ZeroValueFloat != nil {
	    if target.ZeroValueFloat < 0{
		err = goa.MergeErrors(err, goa.InvalidRangeError("target.zero_value_float", target.ZeroValueFloat, 1, true))
	    }

	}
}
`

	StringRequiredValidationCode = `func Validate() (err error) {
	err = goa.MergeErrors(err, goa.ValidatePattern("target.required_string", target.RequiredString, "^[A-z].*[a-z]$"))
	if utf8.RuneCountInString(target.RequiredString) < 1 {
		err = goa.MergeErrors(err, goa.InvalidLengthError("target.required_string", target.RequiredString, utf8.RuneCountInString(target.RequiredString), 1, true))
	}
	if utf8.RuneCountInString(target.RequiredString) > 10 {
		err = goa.MergeErrors(err, goa.InvalidLengthError("target.required_string", target.RequiredString, utf8.RuneCountInString(target.RequiredString), 10, false))
	}
	if target.DefaultString != nil {
		if !(*target.DefaultString == "foo" || *target.DefaultString == "bar") {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("target.default_string", *target.DefaultString, []interface{}{"foo", "bar"}))
		}
	}
	if target.String != nil {
		err = goa.MergeErrors(err, goa.ValidateFormat("target.string", *target.String, goa.FormatDateTime))
	}
	if utf8.RuneCountInString(target.ZeroValueString) < 1 {
	    err = goa.MergeError(err, goa.InvalidLengthError("target.zero_value_string", target.ZeroValueString, utf8.RuneCountInString(target.ZeroValueString), 1, true))
	}
}
`

	StringPointerValidationCode = `func Validate() (err error) {
	if target.RequiredString != nil {
		err = goa.MergeErrors(err, goa.ValidatePattern("target.required_string", *target.RequiredString, "^[A-z].*[a-z]$"))
	}
	if target.RequiredString != nil {
		if utf8.RuneCountInString(*target.RequiredString) < 1 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("target.required_string", *target.RequiredString, utf8.RuneCountInString(*target.RequiredString), 1, true))
		}
	}
	if target.RequiredString != nil {
		if utf8.RuneCountInString(*target.RequiredString) > 10 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("target.required_string", *target.RequiredString, utf8.RuneCountInString(*target.RequiredString), 10, false))
		}
	}
	if target.DefaultString != nil {
		if !(*target.DefaultString == "foo" || *target.DefaultString == "bar") {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("target.default_string", *target.DefaultString, []interface{}{"foo", "bar"}))
		}
	}
	if target.String != nil {
		err = goa.MergeErrors(err, goa.ValidateFormat("target.string", *target.String, goa.FormatDateTime))
	}
	if utf8.RuneCountInString(target.ZeroValueString) < 1 {
	    err = goa.MergeError(err, goa.InvalidLengthError("target.zero_value_string", target.ZeroValueString, utf8.RuneCountInString(target.ZeroValueString), 1, true))
	}
}
`

	StringUseDefaultValidationCode = `func Validate() (err error) {
	err = goa.MergeErrors(err, goa.ValidatePattern("target.required_string", target.RequiredString, "^[A-z].*[a-z]$"))
	if utf8.RuneCountInString(target.RequiredString) < 1 {
		err = goa.MergeErrors(err, goa.InvalidLengthError("target.required_string", target.RequiredString, utf8.RuneCountInString(target.RequiredString), 1, true))
	}
	if utf8.RuneCountInString(target.RequiredString) > 10 {
		err = goa.MergeErrors(err, goa.InvalidLengthError("target.required_string", target.RequiredString, utf8.RuneCountInString(target.RequiredString), 10, false))
	}
	if !(target.DefaultString == "foo" || target.DefaultString == "bar") {
		err = goa.MergeErrors(err, goa.InvalidEnumValueError("target.default_string", target.DefaultString, []interface{}{"foo", "bar"}))
	}
	if target.String != nil {
		err = goa.MergeErrors(err, goa.ValidateFormat("target.string", *target.String, goa.FormatDateTime))
	}
	if utf8.RuneCountInString(target.ZeroValueString) < 1 {
	    err = goa.MergeError(err, goa.InvalidLengthError("target.zero_value_string", target.ZeroValueString, utf8.RuneCountInString(target.ZeroValueString), 1, true))
	}
}
`

	StringUseZeroValueValidationCode = `func Validate() (err error) {
	err = goa.MergeErrors(err, goa.ValidatePattern("target.required_string", target.RequiredString, "^[A-z].*[a-z]$"))
	if utf8.RuneCountInString(target.RequiredString) < 1 {
		err = goa.MergeErrors(err, goa.InvalidLengthError("target.required_string", target.RequiredString, utf8.RuneCountInString(target.RequiredString), 1, true))
	}
	if utf8.RuneCountInString(target.RequiredString) > 10 {
		err = goa.MergeErrors(err, goa.InvalidLengthError("target.required_string", target.RequiredString, utf8.RuneCountInString(target.RequiredString), 10, false))
	}
	if !(target.DefaultString == "foo" || target.DefaultString == "bar") {
		err = goa.MergeErrors(err, goa.InvalidEnumValueError("target.default_string", target.DefaultString, []interface{}{"foo", "bar"}))
	}
	if target.String != nil {
		err = goa.MergeErrors(err, goa.ValidateFormat("target.string", *target.String, goa.FormatDateTime))
	}
	if utf8.RuneCountInString(target.ZeroValueString) < 1 {
	    err = goa.MergeError(err, goa.InvalidLengthError("target.zero_value_string", target.ZeroValueString, utf8.RuneCountInString(target.ZeroValueString), 1, true))
	}
}
`
	UserTypeRequiredValidationCode = `func Validate() (err error) {
	if target.RequiredInteger != nil {
		if err2 := ValidateInteger(target.RequiredInteger); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if target.DefaultString != nil {
		if err2 := ValidateString(target.DefaultString); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if target.Float != nil {
		if err2 := ValidateFloat(target.Float); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if target.ZeroValueInteger != nil {
	    if err2 := ValidateInteger(target.ZeroValueInteger); err2 != nil {
		err = goa.MergeErrors(err, err2)
	    }

	}
}
`

	UserTypePointerValidationCode = `func Validate() (err error) {
	if target.RequiredInteger != nil {
		if err2 := ValidateInteger(target.RequiredInteger); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if target.DefaultString != nil {
		if err2 := ValidateString(target.DefaultString); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if target.Float != nil {
		if err2 := ValidateFloat(target.Float); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if target.ZeroValueInteger != nil {
	    if err2 := ValidateInteger(target.ZeroValueInteger); err2 != nil {
		err = goa.MergeErrors(err, err2)
	    }

	}
}
`
	UserTypeUseDefaultValidationCode = `func Validate() (err error) {
	if target.RequiredInteger != nil {
		if err2 := ValidateInteger(target.RequiredInteger); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if target.DefaultString != nil {
		if err2 := ValidateString(target.DefaultString); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if target.Float != nil {
		if err2 := ValidateFloat(target.Float); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if target.ZeroValueInteger != nil {
	    if err2 := ValidateInteger(target.ZeroValueInteger); err2 != nil {
		err = goa.MergeErrors(err, err2)
	    }

	}
}
`

	UserTypeUseZeroValueValidationCode = `func Validate() (err error) {
	if target.RequiredInteger != nil {
		if err2 := ValidateInteger(target.RequiredInteger); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if target.DefaultString != nil {
		if err2 := ValidateString(target.DefaultString); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if target.Float != nil {
		if err2 := ValidateFloat(target.Float); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if target.ZeroValueInteger != nil {
	    if err2 := ValidateInteger(target.ZeroValueInteger); err2 != nil {
		err = goa.MergeErrors(err, err2)
	    }

	}
}
`
	UserTypeArrayValidationCode = `func Validate() (err error) {
	for _, e := range target.Array {
		if e != nil {
			if err2 := ValidateFloat(e); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
	}
}
`

	ArrayRequiredValidationCode = `func Validate() (err error) {
	if len(target.RequiredArray) < 5 {
		err = goa.MergeErrors(err, goa.InvalidLengthError("target.required_array", target.RequiredArray, len(target.RequiredArray), 5, true))
	}
	if len(target.DefaultArray) > 3 {
		err = goa.MergeErrors(err, goa.InvalidLengthError("target.default_array", target.DefaultArray, len(target.DefaultArray), 3, false))
	}
	for _, e := range target.Array {
		if !(e == 0 || e == 1 || e == 1 || e == 2 || e == 3 || e == 5) {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("target.array[*]", e, []interface{}{0, 1, 1, 2, 3, 5}))
		}
	}
	if len(target.ZeroValueArray) < 1 {
	    err = goa.MergeErrors(err, goa.InvalidLengthError("target.zero_array", target.ZeroValueArray, len(target.ZeroValueArray), 1, true))
	}
}
`

	ArrayPointerValidationCode = `func Validate() (err error) {
	if len(target.RequiredArray) < 5 {
		err = goa.MergeErrors(err, goa.InvalidLengthError("target.required_array", target.RequiredArray, len(target.RequiredArray), 5, true))
	}
	if len(target.DefaultArray) > 3 {
		err = goa.MergeErrors(err, goa.InvalidLengthError("target.default_array", target.DefaultArray, len(target.DefaultArray), 3, false))
	}
	for _, e := range target.Array {
		if !(e == 0 || e == 1 || e == 1 || e == 2 || e == 3 || e == 5) {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("target.array[*]", e, []interface{}{0, 1, 1, 2, 3, 5}))
		}
	}
	if len(target.ZeroValueArray) < 1 {
	    err = goa.MergeErrors(err, goa.InvalidLengthError("target.zero_array", target.ZeroValueArray, len(target.ZeroValueArray), 1, true))
	}
}
`

	ArrayUseDefaultValidationCode = `func Validate() (err error) {
	if len(target.RequiredArray) < 5 {
		err = goa.MergeErrors(err, goa.InvalidLengthError("target.required_array", target.RequiredArray, len(target.RequiredArray), 5, true))
	}
	if len(target.DefaultArray) > 3 {
		err = goa.MergeErrors(err, goa.InvalidLengthError("target.default_array", target.DefaultArray, len(target.DefaultArray), 3, false))
	}
	for _, e := range target.Array {
		if !(e == 0 || e == 1 || e == 1 || e == 2 || e == 3 || e == 5) {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("target.array[*]", e, []interface{}{0, 1, 1, 2, 3, 5}))
		}
	}
	if len(target.ZeroValueArray) < 1 {
	    err = goa.MergeErrors(err, goa.InvalidLengthError("target.zero_array", target.ZeroValueArray, len(target.ZeroValueArray), 1, true))
	}
}
`

	ArrayUseZeroValueValidationCode = `func Validate() (err error) {
	if len(target.RequiredArray) < 5 {
		err = goa.MergeErrors(err, goa.InvalidLengthError("target.required_array", target.RequiredArray, len(target.RequiredArray), 5, true))
	}
	if len(target.DefaultArray) > 3 {
		err = goa.MergeErrors(err, goa.InvalidLengthError("target.default_array", target.DefaultArray, len(target.DefaultArray), 3, false))
	}
	for _, e := range target.Array {
		if !(e == 0 || e == 1 || e == 1 || e == 2 || e == 3 || e == 5) {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("target.array[*]", e, []interface{}{0, 1, 1, 2, 3, 5}))
		}
	}
	if len(target.ZeroValueArray) < 1 {
	    err = goa.MergeErrors(err, goa.InvalidLengthError("target.zero_array", target.ZeroValueArray, len(target.ZeroValueArray), 1, true))
	}
}
`
	MapRequiredValidationCode = `func Validate() (err error) {
	if len(target.RequiredMap) < 5 {
		err = goa.MergeErrors(err, goa.InvalidLengthError("target.required_map", target.RequiredMap, len(target.RequiredMap), 5, true))
	}
	if len(target.DefaultMap) > 3 {
		err = goa.MergeErrors(err, goa.InvalidLengthError("target.default_map", target.DefaultMap, len(target.DefaultMap), 3, false))
	}
	for k, v := range target.Map {
		err = goa.MergeErrors(err, goa.ValidatePattern("target.map.key", k, "^[A-Z]"))
		if v > 5 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("target.map[key]", v, 5, false))
		}
	}
	if len(target.ZeroValueMap) < 1 {
	    err = goa.MergeErrors(err, goa.InvalidLengthError("target.zero_value_map", target.ZeroValueMap, len(target.ZeroValueMap), 1, true))
	}
}
`

	MapPointerValidationCode = `func Validate() (err error) {
	if len(target.RequiredMap) < 5 {
		err = goa.MergeErrors(err, goa.InvalidLengthError("target.required_map", target.RequiredMap, len(target.RequiredMap), 5, true))
	}
	if len(target.DefaultMap) > 3 {
		err = goa.MergeErrors(err, goa.InvalidLengthError("target.default_map", target.DefaultMap, len(target.DefaultMap), 3, false))
	}
	for k, v := range target.Map {
		err = goa.MergeErrors(err, goa.ValidatePattern("target.map.key", k, "^[A-Z]"))
		if v > 5 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("target.map[key]", v, 5, false))
		}
	}
	if len(target.ZeroValueMap) < 1 {
	    err = goa.MergeErrors(err, goa.InvalidLengthError("target.zero_value_map", target.ZeroValueMap, len(target.ZeroValueMap), 1, true))
	}
}
`

	MapUseDefaultValidationCode = `func Validate() (err error) {
	if len(target.RequiredMap) < 5 {
		err = goa.MergeErrors(err, goa.InvalidLengthError("target.required_map", target.RequiredMap, len(target.RequiredMap), 5, true))
	}
	if len(target.DefaultMap) > 3 {
		err = goa.MergeErrors(err, goa.InvalidLengthError("target.default_map", target.DefaultMap, len(target.DefaultMap), 3, false))
	}
	for k, v := range target.Map {
		err = goa.MergeErrors(err, goa.ValidatePattern("target.map.key", k, "^[A-Z]"))
		if v > 5 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("target.map[key]", v, 5, false))
		}
	}
	if len(target.ZeroValueMap) < 1 {
	    err = goa.MergeErrors(err, goa.InvalidLengthError("target.zero_value_map", target.ZeroValueMap, len(target.ZeroValueMap), 1, true))
	}
}
`
	MapUseZeroValueValidationCode = `func Validate() (err error) {
	if len(target.RequiredMap) < 5 {
		err = goa.MergeErrors(err, goa.InvalidLengthError("target.required_map", target.RequiredMap, len(target.RequiredMap), 5, true))
	}
	if len(target.DefaultMap) > 3 {
		err = goa.MergeErrors(err, goa.InvalidLengthError("target.default_map", target.DefaultMap, len(target.DefaultMap), 3, false))
	}
	for k, v := range target.Map {
		err = goa.MergeErrors(err, goa.ValidatePattern("target.map.key", k, "^[A-Z]"))
		if v > 5 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("target.map[key]", v, 5, false))
		}
	}
	if len(target.ZeroValueMap) < 1 {
	    err = goa.MergeErrors(err, goa.InvalidLengthError("target.zero_value_map", target.ZeroValueMap, len(target.ZeroValueMap), 1, true))
	}
}
`
)
