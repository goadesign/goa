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

// Validate runs the validations defined on ResultType.
func (result *ResultType) Validate() (err error) {
	switch result.View {
	case "default", "":
		err = result.Projected.ValidateDefault()
	case "tiny":
		err = result.Projected.ValidateTiny()
	default:
		err = goa.InvalidEnumValueError("view", result.View, []interface{}{"default", "tiny"})
	}
	return
}

// ValidateDefault runs the validations defined on ResultType using the
// "default" view.
func (result *ResultTypeView) ValidateDefault() (err error) {
	if result.A == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
	}
	if result.B == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("b", "result"))
	}
	return
}

// ValidateTiny runs the validations defined on ResultType using the "tiny"
// view.
func (result *ResultTypeView) ValidateTiny() (err error) {
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

// Validate runs the validations defined on ResultTypeCollection.
func (result ResultTypeCollection) Validate() (err error) {
	switch result.View {
	case "default", "":
		err = result.Projected.ValidateDefault()
	case "tiny":
		err = result.Projected.ValidateTiny()
	default:
		err = goa.InvalidEnumValueError("view", result.View, []interface{}{"default", "tiny"})
	}
	return
}

// ValidateDefault runs the validations defined on ResultTypeCollection using
// the "default" view.
func (result ResultTypeCollectionView) ValidateDefault() (err error) {
	for _, item := range result {
		if err2 := item.ValidateDefault(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// ValidateTiny runs the validations defined on ResultTypeCollection using the
// "tiny" view.
func (result ResultTypeCollectionView) ValidateTiny() (err error) {
	for _, item := range result {
		if err2 := item.ValidateTiny(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// ValidateDefault runs the validations defined on ResultType using the
// "default" view.
func (result *ResultTypeView) ValidateDefault() (err error) {
	if result.A == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
	}
	if result.B == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("b", "result"))
	}
	return
}

// ValidateTiny runs the validations defined on ResultType using the "tiny"
// view.
func (result *ResultTypeView) ValidateTiny() (err error) {
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
	A *UserType
	B *string
}

// UserType is a type that runs validations on a projected type.
type UserType struct {
	A *string
}

// Validate runs the validations defined on ResultType.
func (result *ResultType) Validate() (err error) {
	switch result.View {
	case "default", "":
		err = result.Projected.ValidateDefault()
	case "tiny":
		err = result.Projected.ValidateTiny()
	default:
		err = goa.InvalidEnumValueError("view", result.View, []interface{}{"default", "tiny"})
	}
	return
}

// ValidateDefault runs the validations defined on ResultType using the
// "default" view.
func (result *ResultTypeView) ValidateDefault() (err error) {
	if result.A == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
	}
	return
}

// ValidateTiny runs the validations defined on ResultType using the "tiny"
// view.
func (result *ResultTypeView) ValidateTiny() (err error) {
	if result.A == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
	}
	return
}

// Validate runs the validations defined on UserType.
func (result *UserType) Validate() (err error) {

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
	D *UserType
	E *string
}

// UserType is a type that runs validations on a projected type.
type UserType struct {
	P *string
}

// RT3View is a type that runs validations on a projected type.
type RT3View struct {
	X []string
	Y map[int]*UserType
	Z *string
}

// Validate runs the validations defined on RT.
func (result *RT) Validate() (err error) {
	switch result.View {
	case "default", "":
		err = result.Projected.ValidateDefault()
	case "tiny":
		err = result.Projected.ValidateTiny()
	default:
		err = goa.InvalidEnumValueError("view", result.View, []interface{}{"default", "tiny"})
	}
	return
}

// ValidateDefault runs the validations defined on RT using the "default" view.
func (result *RTView) ValidateDefault() (err error) {

	if result.B != nil {
		if err2 := result.B.ValidateExtended(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if result.C != nil {
		if err2 := result.C.ValidateDefault(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// ValidateTiny runs the validations defined on RT using the "tiny" view.
func (result *RTView) ValidateTiny() (err error) {

	if result.B != nil {
		if err2 := result.B.ValidateTiny(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if result.C != nil {
		if err2 := result.C.ValidateDefault(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// ValidateDefault runs the validations defined on RT2 using the "default" view.
func (result *RT2View) ValidateDefault() (err error) {
	if result.C == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("c", "result"))
	}
	if result.D == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("d", "result"))
	}
	return
}

// ValidateExtended runs the validations defined on RT2 using the "extended"
// view.
func (result *RT2View) ValidateExtended() (err error) {
	if result.C == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("c", "result"))
	}
	if result.D == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("d", "result"))
	}
	return
}

// ValidateTiny runs the validations defined on RT2 using the "tiny" view.
func (result *RT2View) ValidateTiny() (err error) {
	if result.D == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("d", "result"))
	}
	return
}

// Validate runs the validations defined on UserType.
func (result *UserType) Validate() (err error) {

	return
}

// ValidateDefault runs the validations defined on RT3 using the "default" view.
func (result *RT3View) ValidateDefault() (err error) {
	if result.X == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("x", "result"))
	}
	if result.Y == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("y", "result"))
	}
	return
}

// ValidateTiny runs the validations defined on RT3 using the "tiny" view.
func (result *RT3View) ValidateTiny() (err error) {
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

// Validate runs the validations defined on RT.
func (result *RT) Validate() (err error) {
	switch result.View {
	case "default", "":
		err = result.Projected.ValidateDefault()
	case "tiny":
		err = result.Projected.ValidateTiny()
	default:
		err = goa.InvalidEnumValueError("view", result.View, []interface{}{"default", "tiny"})
	}
	return
}

// ValidateDefault runs the validations defined on RT using the "default" view.
func (result *RTView) ValidateDefault() (err error) {
	if result.A == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
	}
	if result.A != nil {
		if err2 := result.A.ValidateTiny(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// ValidateTiny runs the validations defined on RT using the "tiny" view.
func (result *RTView) ValidateTiny() (err error) {
	if result.A == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
	}
	if result.A != nil {
		if err2 := result.A.ValidateDefault(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}
`
