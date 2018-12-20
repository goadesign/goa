package testdata

const ResultWithMultipleViewsCode = `// ResultType is the viewed result type that is projected based on a view.
type ResultType struct {
	// Type to project
	Projected *ResultTypeView
	// View to render
	View string
}

// ResultTypeView is a type that runs validations on a projected type.
type ResultTypeView struct {
	A *string
	B *string
}

// ValidateResultType runs the validations defined on the viewed result type
// ResultType.
func ValidateResultType(result *ResultType) (err error) {
	switch result.View {
	case "default", "":
		err = ValidateResultTypeView(result.Projected)
	case "tiny":
		err = ValidateResultTypeViewTiny(result.Projected)
	default:
		err = goa.InvalidEnumValueError("view", result.View, []interface{}{"default", "tiny"})
	}
	return
}

// ValidateResultTypeView runs the validations defined on ResultTypeView using
// the "default" view.
func ValidateResultTypeView(result *ResultTypeView) (err error) {
	if result.A == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
	}
	if result.B == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("b", "result"))
	}
	return
}

// ValidateResultTypeViewTiny runs the validations defined on ResultTypeView
// using the "tiny" view.
func ValidateResultTypeViewTiny(result *ResultTypeView) (err error) {
	if result.A == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
	}
	return
}
`

const ResultCollectionMultipleViewsCode = `// ResultTypeCollection is the viewed result type that is projected based on a
// view.
type ResultTypeCollection struct {
	// Type to project
	Projected ResultTypeCollectionView
	// View to render
	View string
}

// ResultTypeCollectionView is a type that runs validations on a projected type.
type ResultTypeCollectionView []*ResultTypeView

// ResultTypeView is a type that runs validations on a projected type.
type ResultTypeView struct {
	A *string
	B *string
}

// ValidateResultTypeCollection runs the validations defined on the viewed
// result type ResultTypeCollection.
func ValidateResultTypeCollection(result ResultTypeCollection) (err error) {
	switch result.View {
	case "default", "":
		err = ValidateResultTypeCollectionView(result.Projected)
	case "tiny":
		err = ValidateResultTypeCollectionViewTiny(result.Projected)
	default:
		err = goa.InvalidEnumValueError("view", result.View, []interface{}{"default", "tiny"})
	}
	return
}

// ValidateResultTypeCollectionView runs the validations defined on
// ResultTypeCollectionView using the "default" view.
func ValidateResultTypeCollectionView(result ResultTypeCollectionView) (err error) {
	for _, item := range result {
		if err2 := ValidateResultTypeView(item); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// ValidateResultTypeCollectionViewTiny runs the validations defined on
// ResultTypeCollectionView using the "tiny" view.
func ValidateResultTypeCollectionViewTiny(result ResultTypeCollectionView) (err error) {
	for _, item := range result {
		if err2 := ValidateResultTypeViewTiny(item); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// ValidateResultTypeView runs the validations defined on ResultTypeView using
// the "default" view.
func ValidateResultTypeView(result *ResultTypeView) (err error) {
	if result.A == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
	}
	if result.B == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("b", "result"))
	}
	return
}

// ValidateResultTypeViewTiny runs the validations defined on ResultTypeView
// using the "tiny" view.
func ValidateResultTypeViewTiny(result *ResultTypeView) (err error) {
	if result.A == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
	}
	return
}
`

const ResultWithUserTypeCode = `// ResultType is the viewed result type that is projected based on a view.
type ResultType struct {
	// Type to project
	Projected *ResultTypeView
	// View to render
	View string
}

// ResultTypeView is a type that runs validations on a projected type.
type ResultTypeView struct {
	A *UserTypeView
	B *string
}

// UserTypeView is a type that runs validations on a projected type.
type UserTypeView struct {
	A *string
}

// ValidateResultType runs the validations defined on the viewed result type
// ResultType.
func ValidateResultType(result *ResultType) (err error) {
	switch result.View {
	case "default", "":
		err = ValidateResultTypeView(result.Projected)
	case "tiny":
		err = ValidateResultTypeViewTiny(result.Projected)
	default:
		err = goa.InvalidEnumValueError("view", result.View, []interface{}{"default", "tiny"})
	}
	return
}

// ValidateResultTypeView runs the validations defined on ResultTypeView using
// the "default" view.
func ValidateResultTypeView(result *ResultTypeView) (err error) {
	if result.A == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
	}
	return
}

// ValidateResultTypeViewTiny runs the validations defined on ResultTypeView
// using the "tiny" view.
func ValidateResultTypeViewTiny(result *ResultTypeView) (err error) {
	if result.A == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
	}
	return
}

// ValidateUserTypeView runs the validations defined on UserTypeView.
func ValidateUserTypeView(result *UserTypeView) (err error) {

	return
}
`

const ResultWithResultTypeCode = `// RT is the viewed result type that is projected based on a view.
type RT struct {
	// Type to project
	Projected *RTView
	// View to render
	View string
}

// RTView is a type that runs validations on a projected type.
type RTView struct {
	A *string
	B *RT2View
	C *RT3View
}

// RT2View is a type that runs validations on a projected type.
type RT2View struct {
	C *string
	D *UserTypeView
	E *string
}

// UserTypeView is a type that runs validations on a projected type.
type UserTypeView struct {
	P *string
}

// RT3View is a type that runs validations on a projected type.
type RT3View struct {
	X []string
	Y map[int]*UserTypeView
	Z *string
}

// ValidateRT runs the validations defined on the viewed result type RT.
func ValidateRT(result *RT) (err error) {
	switch result.View {
	case "default", "":
		err = ValidateRTView(result.Projected)
	case "tiny":
		err = ValidateRTViewTiny(result.Projected)
	default:
		err = goa.InvalidEnumValueError("view", result.View, []interface{}{"default", "tiny"})
	}
	return
}

// ValidateRTView runs the validations defined on RTView using the "default"
// view.
func ValidateRTView(result *RTView) (err error) {

	if result.B != nil {
		if err2 := ValidateRT2ViewExtended(result.B); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if result.C != nil {
		if err2 := ValidateRT3View(result.C); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// ValidateRTViewTiny runs the validations defined on RTView using the "tiny"
// view.
func ValidateRTViewTiny(result *RTView) (err error) {

	if result.B != nil {
		if err2 := ValidateRT2ViewTiny(result.B); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if result.C != nil {
		if err2 := ValidateRT3View(result.C); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// ValidateRT2View runs the validations defined on RT2View using the "default"
// view.
func ValidateRT2View(result *RT2View) (err error) {
	if result.C == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("c", "result"))
	}
	if result.D == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("d", "result"))
	}
	return
}

// ValidateRT2ViewExtended runs the validations defined on RT2View using the
// "extended" view.
func ValidateRT2ViewExtended(result *RT2View) (err error) {
	if result.C == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("c", "result"))
	}
	if result.D == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("d", "result"))
	}
	return
}

// ValidateRT2ViewTiny runs the validations defined on RT2View using the "tiny"
// view.
func ValidateRT2ViewTiny(result *RT2View) (err error) {
	if result.D == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("d", "result"))
	}
	return
}

// ValidateUserTypeView runs the validations defined on UserTypeView.
func ValidateUserTypeView(result *UserTypeView) (err error) {

	return
}

// ValidateRT3View runs the validations defined on RT3View using the "default"
// view.
func ValidateRT3View(result *RT3View) (err error) {
	if result.X == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("x", "result"))
	}
	if result.Y == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("y", "result"))
	}
	return
}

// ValidateRT3ViewTiny runs the validations defined on RT3View using the "tiny"
// view.
func ValidateRT3ViewTiny(result *RT3View) (err error) {
	if result.X == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("x", "result"))
	}
	return
}
`

const ResultWithRecursiveResultTypeCode = `// RT is the viewed result type that is projected based on a view.
type RT struct {
	// Type to project
	Projected *RTView
	// View to render
	View string
}

// RTView is a type that runs validations on a projected type.
type RTView struct {
	A *RTView
}

// ValidateRT runs the validations defined on the viewed result type RT.
func ValidateRT(result *RT) (err error) {
	switch result.View {
	case "default", "":
		err = ValidateRTView(result.Projected)
	case "tiny":
		err = ValidateRTViewTiny(result.Projected)
	default:
		err = goa.InvalidEnumValueError("view", result.View, []interface{}{"default", "tiny"})
	}
	return
}

// ValidateRTView runs the validations defined on RTView using the "default"
// view.
func ValidateRTView(result *RTView) (err error) {
	if result.A == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
	}
	if result.A != nil {
		if err2 := ValidateRTViewTiny(result.A); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// ValidateRTViewTiny runs the validations defined on RTView using the "tiny"
// view.
func ValidateRTViewTiny(result *RTView) (err error) {
	if result.A == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
	}
	if result.A != nil {
		if err2 := ValidateRTView(result.A); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}
`
