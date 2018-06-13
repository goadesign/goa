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

// Validate runs the validations defined on the viewed result type ResultType.
func (result *ResultType) Validate() (err error) {
	switch result.View {
	case "default", "":
		err = result.Projected.Validate()
	case "tiny":
		err = result.Projected.ValidateTiny()
	default:
		err = goa.InvalidEnumValueError("view", result.View, []interface{}{"default", "tiny"})
	}
	return
}

// Validate runs the validations defined on ResultTypeView using the "default"
// view.
func (result *ResultTypeView) Validate() (err error) {
	if result.A == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
	}
	if result.B == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("b", "result"))
	}
	return
}

// ValidateTiny runs the validations defined on ResultTypeView using the "tiny"
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

// Validate runs the validations defined on the viewed result type
// ResultTypeCollection.
func (result ResultTypeCollection) Validate() (err error) {
	switch result.View {
	case "default", "":
		err = result.Projected.Validate()
	case "tiny":
		err = result.Projected.ValidateTiny()
	default:
		err = goa.InvalidEnumValueError("view", result.View, []interface{}{"default", "tiny"})
	}
	return
}

// Validate runs the validations defined on ResultTypeCollectionView using the
// "default" view.
func (result ResultTypeCollectionView) Validate() (err error) {
	for _, item := range result {
		if err2 := item.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// ValidateTiny runs the validations defined on ResultTypeCollectionView using
// the "tiny" view.
func (result ResultTypeCollectionView) ValidateTiny() (err error) {
	for _, item := range result {
		if err2 := item.ValidateTiny(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// Validate runs the validations defined on ResultTypeView using the "default"
// view.
func (result *ResultTypeView) Validate() (err error) {
	if result.A == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
	}
	if result.B == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("b", "result"))
	}
	return
}

// ValidateTiny runs the validations defined on ResultTypeView using the "tiny"
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
	A *UserTypeView
	B *string
}

// UserTypeView is a type that runs validations on a projected type.
type UserTypeView struct {
	A *string
}

// Validate runs the validations defined on the viewed result type ResultType.
func (result *ResultType) Validate() (err error) {
	switch result.View {
	case "default", "":
		err = result.Projected.Validate()
	case "tiny":
		err = result.Projected.ValidateTiny()
	default:
		err = goa.InvalidEnumValueError("view", result.View, []interface{}{"default", "tiny"})
	}
	return
}

// Validate runs the validations defined on ResultTypeView using the "default"
// view.
func (result *ResultTypeView) Validate() (err error) {
	if result.A == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
	}
	return
}

// ValidateTiny runs the validations defined on ResultTypeView using the "tiny"
// view.
func (result *ResultTypeView) ValidateTiny() (err error) {
	if result.A == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
	}
	return
}

// Validate runs the validations defined on UserTypeView.
func (result *UserTypeView) Validate() (err error) {

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

// Validate runs the validations defined on the viewed result type RT.
func (result *RT) Validate() (err error) {
	switch result.View {
	case "default", "":
		err = result.Projected.Validate()
	case "tiny":
		err = result.Projected.ValidateTiny()
	default:
		err = goa.InvalidEnumValueError("view", result.View, []interface{}{"default", "tiny"})
	}
	return
}

// Validate runs the validations defined on RTView using the "default" view.
func (result *RTView) Validate() (err error) {

	if result.B != nil {
		if err2 := result.B.ValidateExtended(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if result.C != nil {
		if err2 := result.C.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// ValidateTiny runs the validations defined on RTView using the "tiny" view.
func (result *RTView) ValidateTiny() (err error) {

	if result.B != nil {
		if err2 := result.B.ValidateTiny(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if result.C != nil {
		if err2 := result.C.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// Validate runs the validations defined on RT2View using the "default" view.
func (result *RT2View) Validate() (err error) {
	if result.C == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("c", "result"))
	}
	if result.D == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("d", "result"))
	}
	return
}

// ValidateExtended runs the validations defined on RT2View using the
// "extended" view.
func (result *RT2View) ValidateExtended() (err error) {
	if result.C == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("c", "result"))
	}
	if result.D == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("d", "result"))
	}
	return
}

// ValidateTiny runs the validations defined on RT2View using the "tiny" view.
func (result *RT2View) ValidateTiny() (err error) {
	if result.D == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("d", "result"))
	}
	return
}

// Validate runs the validations defined on UserTypeView.
func (result *UserTypeView) Validate() (err error) {

	return
}

// Validate runs the validations defined on RT3View using the "default" view.
func (result *RT3View) Validate() (err error) {
	if result.X == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("x", "result"))
	}
	if result.Y == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("y", "result"))
	}
	return
}

// ValidateTiny runs the validations defined on RT3View using the "tiny" view.
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

// Validate runs the validations defined on the viewed result type RT.
func (result *RT) Validate() (err error) {
	switch result.View {
	case "default", "":
		err = result.Projected.Validate()
	case "tiny":
		err = result.Projected.ValidateTiny()
	default:
		err = goa.InvalidEnumValueError("view", result.View, []interface{}{"default", "tiny"})
	}
	return
}

// Validate runs the validations defined on RTView using the "default" view.
func (result *RTView) Validate() (err error) {
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

// ValidateTiny runs the validations defined on RTView using the "tiny" view.
func (result *RTView) ValidateTiny() (err error) {
	if result.A == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("a", "result"))
	}
	if result.A != nil {
		if err2 := result.A.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}
`
